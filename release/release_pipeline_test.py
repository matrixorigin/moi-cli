#!/usr/bin/env python3
"""Offline publication tests. Uses synthetic binaries and an in-memory OSS.

Run with publish-requirements.txt installed. Native candidate execution is a
separate four-platform workflow gate, not simulated by these tests.
"""
from datetime import datetime, timezone
import copy
import io
import gzip
import json
import os
from pathlib import Path
import subprocess
import sys
import tarfile
import tempfile
import unittest
import urllib.error
from types import SimpleNamespace
from unittest.mock import patch
import yaml

import release_candidate as candidate
import release_publish as publisher
import release_transport as transport


class InputTests(unittest.TestCase):
    def test_source_preflight_prepares_assets_before_skill_checks(self):
        # Execute each real workflow shell entry with isolated make owners.
        # This checks asset production/consumption; Go test/vet are stubbed.
        root = Path(__file__).resolve().parents[3]
        for filename, step_name in (("release-moi-cli.yml", "Test CLI and embedded Skill"),
                                    ("astra-qa-snapshot.yml", "Verify source CLI and embedded Skill")):
            workflow = yaml.safe_load((root / ".github/workflows" / filename).read_text())
            step = next(step for job in workflow["jobs"].values() for step in job.get("steps", [])
                        if step.get("name") == step_name)
            for fail_assets in (False, True):
                with self.subTest(workflow=filename, fail_assets=fail_assets), tempfile.TemporaryDirectory() as temp:
                    checkout = Path(temp)
                    sdk = checkout / step["working-directory"]
                    (sdk / "cli").mkdir(parents=True)
                    moi = sdk.parent / "moi"
                    moi.mkdir()
                    (moi / "Makefile").write_text(
                        "workspace-assets:\n"
                        "\t@test \"$(FAIL_ASSETS)\" != 1\n"
                        "\t@mkdir -p pkg/model\n"
                        "\t@\"$(PYTHON)\" -c \"import zipfile; "
                        "z=zipfile.ZipFile('pkg/model/workspace_assets.zip','w'); "
                        "z.writestr('session-workspace/README.md','selected MOI source'); z.close()\"\n")
                    (sdk / "Makefile").write_text(
                        "cli-skill-check:\n"
                        "\t@\"$(PYTHON)\" -c \"import pathlib,zipfile; "
                        "z=zipfile.ZipFile('../moi/pkg/model/workspace_assets.zip'); "
                        "assert z.read('session-workspace/README.md') == b'selected MOI source'; "
                        "pathlib.Path('skill-check.ok').write_text('validated selected assets')\"\n"
                        "cli-release-test:\n\t@test -f skill-check.ok\n")
                    tools = checkout / "tools"
                    tools.mkdir()
                    go = tools / "go"
                    go.write_text('#!/bin/sh\ncase "$1" in test|vet) exit 0;; *) exit 99;; esac\n')
                    go.chmod(0o755)
                    env = {**os.environ, "PYTHON": sys.executable, "FAIL_ASSETS": "1" if fail_assets else "0",
                           "PATH": str(tools) + os.pathsep + os.environ["PATH"]}
                    result = subprocess.run(["bash", "-c", step["run"]], cwd=sdk, env=env,
                                            capture_output=True, text=True, timeout=30)
                    if fail_assets:
                        self.assertNotEqual(0, result.returncode)
                        self.assertFalse((sdk / "skill-check.ok").exists())
                    else:
                        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
                        self.assertEqual("validated selected assets", (sdk / "skill-check.ok").read_text())

    def build_env(self):
        return dict(RELEASE_OPERATION="build", RELEASE_ENVIRONMENT="qa", BUNDLE_VERSION="2026.09.20.1-qa",
                    MOI_VERSION="0.1.0", ASTRA_MODE="source", ASTRA_VERSION="0.2.7",
                    MOI_SOURCE_REF="dev",
                    ASTRA_REF="main", COMPATIBILITY_REFERENCE="moi-native-v1",
                    ASTRA_URL="https://astra.example.com",
                    MOI_RELEASE_ASTRA_URL_QA="https://astra.example.com",
                    MOI_RELEASE_ASTRA_URL_PROD="https://astra-prod.example.com",
                    GITHUB_SHA="a" * 40, RELEASE_DRY_RUN="true")

    def test_mode_specific_requirements(self):
        source = self.build_env()
        release = dict(source, ASTRA_MODE="release", ASTRA_REF="")
        reuse = dict(source, ASTRA_MODE="reuse", ASTRA_VERSION="",
                     ASTRA_REF="2026.09.19.1", MANIFEST_SHA256="b" * 64)
        promote = dict(RELEASE_OPERATION="promote", BUNDLE_VERSION="2026.09.20.1",
                       MANIFEST_SHA256="b" * 64,
                       MOI_RELEASE_ASTRA_URL_PROD="https://astra-prod.example.com")
        promoted = dict(RELEASE_OPERATION="promote-environment", BUNDLE_VERSION="2026.09.20.1-qa",
                        MANIFEST_SHA256="c" * 64, ASTRA_URL="https://astra-prod.example.com",
                        MOI_RELEASE_ASTRA_URL_PROD="https://astra-prod.example.com")
        renew_qa = dict(RELEASE_OPERATION="publish-channel", BUNDLE_VERSION="2026.09.20.1-qa",
                        MANIFEST_SHA256="d" * 64)
        for env in (source, release, reuse, promote, promoted, renew_qa):
            with self.subTest(env=env):
                candidate.validate_inputs(env)

    def test_missing_build_fields_reported_together_without_traceback(self):
        with patch.dict(os.environ, {"ASTRA_MODE": "source"}, clear=True), \
                patch.object(sys, "argv", ["release_candidate.py", "validate-inputs"]), \
                patch("sys.stderr", new_callable=io.StringIO) as stderr:
            with self.assertRaises(SystemExit) as raised:
                candidate.main()
        self.assertEqual(1, raised.exception.code)
        for field in ("bundle_version", "moi_version", "astra_version", "astra_ref",
                      "compatibility_reference", "MOI_RELEASE_ASTRA_URL_QA", "release_environment",
                      "workflow commit"):
            self.assertIn(field, stderr.getvalue())
        self.assertNotIn("Traceback", stderr.getvalue())

    def test_moi_reuse_inputs_are_explicit_and_do_not_require_a_new_version(self):
        env = dict(self.build_env(), MOI_MODE="reuse", MOI_VERSION="",
                   MOI_REF="2026.09.28.1", MOI_MANIFEST_SHA256="b" * 64)
        candidate.validate_inputs(env)
        for key in ("MOI_REF", "MOI_MANIFEST_SHA256"):
            with self.subTest(key=key), self.assertRaisesRegex(ValueError, key.lower()):
                candidate.validate_inputs(dict(env, **{key: ""}))
        with self.assertRaisesRegex(ValueError, "moi_mode"):
            candidate.validate_inputs(dict(env, MOI_MODE="unknown"))

    def test_moi_source_ref_validation_is_separate_from_the_workflow_ref(self):
        env = dict(self.build_env(), MOI_SOURCE_REF="5.2.2-dev", RELEASE_DRY_RUN="false",
                   GITHUB_REPOSITORY="matrixorigin/matrixflow", GITHUB_REF="refs/heads/dev")
        candidate.validate_inputs(env)
        with self.assertRaises(ValueError):
            candidate.validate_inputs(dict(env, GITHUB_REF="refs/heads/5.2.2-dev"))
        for ref in ("", " ", "../dev", "feature/../dev", "dev?token=private-value", "-dev", "dev\nother"):
            with self.subTest(ref=ref), self.assertRaises(ValueError) as raised:
                candidate.validate_inputs(dict(env, MOI_SOURCE_REF=ref))
            self.assertIn("moi_source_ref", str(raised.exception))
            self.assertNotIn("private-value", str(raised.exception))
        # Reuse consumes the published bundle identity, not a source ref.
        candidate.validate_inputs(dict(env, MOI_MODE="reuse", MOI_VERSION="", MOI_SOURCE_REF="not a ref",
                                       MOI_REF="previous", MOI_MANIFEST_SHA256="b" * 64))

    def test_workflow_builds_pinned_moi_source_and_keeps_publication_on_the_workflow_commit(self):
        workflow = yaml.safe_load((Path(__file__).resolve().parents[3] /
                                   ".github/workflows/release-moi-cli.yml").read_text())
        inputs = workflow[True]["workflow_dispatch"]["inputs"]
        self.assertEqual("dev", inputs["moi_source_ref"]["default"])
        steps = workflow["jobs"]["build"]["steps"]
        checkout = next(step for step in steps if step.get("with", {}).get("path") == "moi-source")
        self.assertEqual("matrixorigin/matrixflow", checkout["with"]["repository"])
        self.assertEqual("${{ needs.resolve.outputs.moi_source_sha }}", checkout["with"]["ref"])
        self.assertFalse(checkout["with"]["persist-credentials"])
        self.assertEqual("inputs.moi_mode != 'reuse'", checkout["if"])
        for name, directory in (("Test CLI and embedded Skill", "moi-source/sdk"),
                                ("Build versioned native moi-cli", "moi-source/sdk/cli")):
            step = next(step for step in steps if step.get("name") == name)
            self.assertEqual(directory, step["working-directory"])
        go = next(step for step in steps if step.get("uses") == "actions/setup-go@v5")
        self.assertEqual("moi-source/sdk/cli/go.mod", go["with"]["go-version-file"])
        for job in ("resolve", "verify", "publish", "publish-qa-channel", "promote-environment"):
            checkout = next(step for step in workflow["jobs"][job]["steps"] if "checkout@" in step.get("uses", ""))
            self.assertEqual("${{ github.sha }}", checkout["with"]["ref"])
            self.assertFalse(checkout["with"]["persist-credentials"])

    def test_workflow_skips_moi_build_but_keeps_native_pair_probes_for_reuse(self):
        workflow = yaml.safe_load((Path(__file__).resolve().parents[3] /
                                   ".github/workflows/release-moi-cli.yml").read_text())
        inputs = workflow[True]["workflow_dispatch"]["inputs"]
        self.assertEqual("source", inputs["moi_mode"]["default"])
        steps = workflow["jobs"]["build"]["steps"]
        for name in ("Test CLI and embedded Skill", "Build versioned native moi-cli"):
            matches = [step for step in steps if step.get("name") == name]
            self.assertEqual(1, len(matches), f"expected exactly one {name} step")
            self.assertEqual("inputs.moi_mode != 'reuse'", matches[0]["if"])
        go_setup = [step for step in steps if step.get("uses") == "actions/setup-go@v5"]
        self.assertEqual(1, len(go_setup), "expected exactly one Go setup step")
        self.assertEqual("inputs.moi_mode != 'reuse'", go_setup[0]["if"])
        probe = next(step for step in steps if step.get("name") == "Verify real Skill discovery and package components")
        self.assertNotIn("if", probe)
        for name in ("build", "verify", "publish"):
            env = workflow["jobs"][name]["env"]
            self.assertEqual("${{ needs.resolve.outputs.moi_version }}", env["MOI_VERSION"])
            self.assertEqual("${{ needs.resolve.outputs.moi_source_sha }}", env["MOI_SOURCE_SHA"])

    def test_digest_required_only_for_promote_and_reuse(self):
        for env in (dict(self.build_env(), ASTRA_MODE="reuse"),
                    dict(RELEASE_OPERATION="promote", BUNDLE_VERSION="2026.09.20.1",
                         MOI_RELEASE_ASTRA_URL_PROD="https://astra-prod.example.com"),
                    dict(RELEASE_OPERATION="promote-environment", BUNDLE_VERSION="2026.09.20.1-qa",
                         ASTRA_URL="https://astra-prod.example.com",
                         MOI_RELEASE_ASTRA_URL_PROD="https://astra-prod.example.com"),
                    dict(RELEASE_OPERATION="publish-channel", BUNDLE_VERSION="2026.09.20.1-qa")):
            for digest in ("", "a" * 40, "not-a-digest"):
                with self.subTest(env=env, digest=digest):
                    with self.assertRaisesRegex(ValueError, "manifest_sha256"):
                        candidate.validate_inputs(dict(env, MANIFEST_SHA256=digest))

    def test_invalid_values_not_echoed_and_publication_guard_preserved(self):
        secret = "https://user:private-password@astra.example.com"
        with self.assertRaises(ValueError) as raised:
            candidate.validate_inputs(dict(self.build_env(), ASTRA_URL=secret,
                                           RELEASE_DRY_RUN="false"))
        self.assertIn("astra_url", str(raised.exception))
        self.assertIn("matrixorigin/matrixflow dev", str(raised.exception))
        self.assertNotIn("private-password", str(raised.exception))
        candidate.validate_inputs(dict(self.build_env(), RELEASE_DRY_RUN="false",
                                       GITHUB_REPOSITORY="matrixorigin/matrixflow",
                                       GITHUB_REF="refs/heads/dev"))

    def test_workflow_defaults_and_promotion_preflight(self):
        path = Path(__file__).resolve().parents[3] / ".github/workflows/release-moi-cli.yml"
        workflow = yaml.safe_load(path.read_text())
        inputs = workflow[True]["workflow_dispatch"]["inputs"]
        self.assertEqual("moi-native-v1", inputs["compatibility_reference"]["default"])
        self.assertIn("matrixorigin/Astra", inputs["astra_ref"]["description"])
        for job_name in ("promote", "promote-environment", "publish-channel", "publish-qa-channel"):
            steps = workflow["jobs"][job_name]["steps"]
            preflight = next(i for i, step in enumerate(steps)
                             if "validate-inputs" in step.get("run", ""))
            self.assertTrue(all("secrets." not in json.dumps(step) for step in steps[:preflight + 1]), job_name)
            self.assertEqual(job_name if job_name != "publish-qa-channel" else "build",
                             steps[preflight]["env"]["RELEASE_OPERATION"])
        for name, job in workflow["jobs"].items():
            self.assertNotIn("environment", job, name)
        self.assertEqual(["qa"], inputs["release_environment"]["options"])
        self.assertIn("publish-channel", inputs["operation"]["options"])
        self.assertIn("promote-environment", inputs["operation"]["options"])
        self.assertNotIn("promote_stable", inputs)

    def test_build_rejects_direct_prod_and_prerelease_components(self):
        with self.assertRaisesRegex(ValueError, "only publishes QA"):
            candidate.validate_inputs(dict(self.build_env(), RELEASE_ENVIRONMENT="prod"))
        for key in ("ASTRA_VERSION", "MOI_VERSION"):
            with self.subTest(key=key), self.assertRaisesRegex(ValueError, key.lower()):
                candidate.validate_inputs(dict(self.build_env(), **{key: "0.1.5-rc1" if key == "ASTRA_VERSION" else "0.1.4-rc1"}))
        with self.assertRaisesRegex(ValueError, "-qa"):
            candidate.validate_inputs(dict(self.build_env(), BUNDLE_VERSION="2026.09.22.1"))
        with self.assertRaisesRegex(ValueError, "allowlist"):
            candidate.validate_inputs(dict(self.build_env(), ASTRA_URL="https://other.example.com"))

    def test_platform_selection_is_explicit_and_defaults_to_full_set(self):
        self.assertEqual(candidate.TARGETS, candidate.selected_targets())
        for target in ("darwin-arm64", "linux-arm64"):
            self.assertEqual((target,), candidate.selected_targets(target))
            candidate.validate_inputs(dict(self.build_env(), RELEASE_PLATFORMS=target))
        for selection in ("", "arm", "darwin-amd64", "all darwin-arm64", "../darwin-arm64"):
            with self.subTest(selection=selection), self.assertRaisesRegex(ValueError, "platforms"):
                candidate.validate_inputs(dict(self.build_env(), RELEASE_PLATFORMS=selection))


class MoiSourceTests(unittest.TestCase):
    def test_source_resolution_defaults_to_dev_and_pins_the_selected_upstream_ref(self):
        for ref, encoded in ((None, "dev"), ("5.2.2-dev", "5.2.2-dev"),
                             ("feature/cli", "feature%2Fcli"), ("b" * 40, "b" * 40)):
            env = {"MOI_VERSION": "1.2.3", "GITHUB_SHA": "a" * 40}
            if ref is not None:
                env["MOI_SOURCE_REF"] = ref
            with self.subTest(ref=ref), patch.object(candidate, "download", return_value=("c" * 40).encode()) as fetch:
                self.assertEqual(("1.2.3", "c" * 40), candidate.resolve_moi(env, candidate.TARGETS))
            fetch.assert_called_once_with("https://api.github.com/repos/matrixorigin/matrixflow/commits/" + encoded,
                                          128, github_api=True, accept="application/vnd.github.sha")

    def test_source_resolution_does_not_fall_back_to_the_workflow_commit(self):
        env = {"MOI_VERSION": "1.2.3", "GITHUB_SHA": "a" * 40, "MOI_SOURCE_REF": "missing"}
        for failure in (ValueError("synthetic missing ref"), b"not-a-commit"):
            with self.subTest(failure=failure), patch.object(candidate, "download", side_effect=[failure]):
                with self.assertRaises(ValueError):
                    candidate.resolve_moi(env, candidate.TARGETS)

    def test_moi_checkout_verification_rejects_a_different_commit_or_tracked_changes(self):
        with tempfile.TemporaryDirectory(prefix="moi-source-test-") as temp:
            root = Path(temp)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            source = root / "source.txt"
            source.write_text("selected source\n")
            subprocess.run(["git", "add", "source.txt"], cwd=root, check=True)
            subprocess.run(["git", "-c", "user.name=Test", "-c", "user.email=test@example.test",
                            "commit", "-qm", "synthetic source"], cwd=root, check=True)
            sha = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
            with patch.dict(os.environ, {"MOI_SOURCE_SHA": sha}, clear=True), \
                    patch.object(sys, "argv", ["release_candidate.py", "verify-moi-source", "--moi-root", str(root)]), \
                    patch("builtins.print"):
                candidate.main()
            with self.assertRaises(ValueError):
                candidate.verify_checkout(root, "a" * 40, "MOI")
            source.write_text("changed source\n")
            with self.assertRaises(ValueError):
                candidate.verify_checkout(root, sha, "MOI")
            output = root / "candidate"
            env = dict(InputTests().build_env(), MOI_SOURCE_SHA=sha, ASTRA_MODE="release")
            with patch.dict(os.environ, env, clear=True), \
                    patch.object(sys, "argv", ["release_candidate.py", "prepare", "--moi-root", str(root),
                                               "--moi", str(root / "moi-cli"), "--output", str(output),
                                               "--astra-release-root", str(root)]), \
                    self.assertRaises(ValueError):
                candidate.main()
            self.assertFalse(output.exists())


class MemoryOSS:
    def __init__(self):
        self.objects = {}
        self.writes = []
        self.fail_key = None
        self.invalid_storage = False
        self.private_key = None
        self.checks = []
        self.public_reads = []

    def validate_storage(self):
        if self.invalid_storage:
            raise ValueError("invalid storage")

    def get(self, key):
        return self.objects.get(key)

    def put(self, key, data, immutable):
        if key == self.fail_key:
            raise OSError("synthetic interrupted upload")
        if immutable and key in self.objects:
            raise ValueError("forbid overwrite")
        self.objects[key] = data
        self.writes.append((key, immutable))

    def check_existing(self, key, expected_sha, expected_size):
        self.checks.append(key)
        data = self.get(key)
        if data is None:
            return False
        if candidate.digest(data) != expected_sha or len(data) != expected_size:
            raise ValueError("immutable release conflict: " + key)
        return True

    def verify_public(self, key, expected_sha, expected_size):
        self.public_reads.append(key)
        data = self.get(key)
        if self.private_key == key or data is None:
            raise ValueError("anonymous download unavailable")
        transport.verify_stream(io.BytesIO(data), expected_sha, expected_size)


class OSSLoggingTests(unittest.TestCase):
    def test_public_stable_channel_distinguishes_absent_from_unreadable(self):
        url = candidate.PUBLIC_URL + "/channels/stable.json"
        with patch.object(publisher, "download", side_effect=urllib.error.HTTPError(url, 404, "Not Found", {}, None)):
            self.assertIsNone(publisher.read_public_channel("channels/stable.json"))
        with patch.object(publisher, "download", side_effect=urllib.error.HTTPError(url, 403, "Forbidden", {}, None)):
            with self.assertRaises(urllib.error.HTTPError):
                publisher.read_public_channel("channels/stable.json")

    def test_failed_put_identifies_the_object_stage_and_error_without_credentials(self):
        class FailingBucket:
            def put_object(self, key, data, headers):
                raise TimeoutError("synthetic timeout")

        store = object.__new__(publisher.OSSStore)
        store.bucket = FailingBucket()
        with self.assertLogs("moi.release", level="INFO") as captured, self.assertRaises(TimeoutError):
            store.put("components/moi-cli/1.0.0/linux-amd64.tar.gz", b"data", immutable=True)
        events = [json.loads(record.getMessage()) for record in captured.records]
        self.assertEqual(["started", "failed"], [event["outcome"] for event in events])
        self.assertTrue(all(event["operation"] == "put_object" for event in events))
        self.assertTrue(all(event["object_key"].endswith("linux-amd64.tar.gz") for event in events))
        self.assertEqual(4, events[0]["size"])
        self.assertEqual("TimeoutError", events[1]["error_type"])
        self.assertGreaterEqual(events[1]["elapsed_ms"], 0)
        self.assertNotIn("synthetic timeout", str(events))


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="moi-publish-test-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.store = MemoryOSS()
        self.sha = "a" * 40
        self.now = datetime(2026, 9, 17, tzinfo=timezone.utc)
        self.candidates = self.make_candidates("first", "1.0.0")
        self.manifest, self.uploads = self.assemble(self.candidates)

    def make_candidates(self, directory, moi_version, platforms="all"):
        root = self.root / directory
        root.mkdir()
        for target in candidate.selected_targets(platforms):
            folder = root / target
            folder.mkdir()
            entries = {}
            for name in candidate.COMPONENTS:
                version = moi_version if name == "moi-cli" else "0.2.4"
                binary = self.root / name
                binary.write_bytes(f"synthetic {name} {version} {target}".encode())
                data = candidate.package_component(binary, None if name == "moi-cli" else b"license")
                filename = f"{name}-{version}-{target}.tar.gz"
                (folder / filename).write_bytes(data)
                entries[name] = {"version": version, "filename": filename,
                                 "sha256": candidate.digest(data), "size": len(data)}
            (folder / "candidate.json").write_bytes(candidate.canonical({
                "schema": "moi-component-candidate-v1", "platform": target, "moi_source_sha": self.sha,
                "astra_source": {"mode": "release", "repository": "matrixorigin/Astra", "tag": "v0.2.4", "release_id": 123,
                                 "asset_id": 456, "archive_sha256": "b" * 64, "prerelease": False},
                "components": entries,
            }))
        return root

    def assemble(self, root, moi="1.0.0", bundle="2026.09.17.1",
                 compatibility="server-contract-v1", astra_url="https://astra.example.test"):
        return candidate.assemble(root, bundle, moi, "0.2.4", self.sha, compatibility,
                                  astra_url=astra_url)

    def publish(self, promote=False, sequence=1001):
        with patch("builtins.print"):
            publisher.publish(self.store, self.manifest, self.uploads)
            if promote:
                self.promote(sequence)

    def download_published(self, url, limit, **kwargs):
        prefix = candidate.PUBLIC_URL + "/"
        self.assertTrue(url.startswith(prefix))
        key = url[len(prefix):]
        data = self.store.get(key)
        if data is None or self.store.private_key == key:
            raise ValueError("requires an uploaded object")
        if len(data) > limit:
            raise ValueError("download exceeds size limit")
        return data

    def promote(self, sequence=1001, expected_sha=None):
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            publisher.promote_existing(self.store, self.manifest["bundle_version"],
                                       expected_sha or candidate.digest(candidate.canonical(self.manifest)), sequence,
                                       self.manifest["client_config"]["astra_url"], self.now)

    def test_complete_set_contains_twelve_objects_and_embedded_skill(self):
        self.assertEqual(12, len(self.uploads))
        self.assertEqual("embedded-export", self.manifest["skill"]["delivery"])
        self.assertTrue(all(k.startswith("components/") for k in self.uploads))
        for platform in self.manifest["platforms"].values():
            self.assertEqual(set(candidate.COMPONENTS), set(platform["components"]))

    def test_single_arm_bundle_upload_promote_and_expand_without_rebuilding(self):
        full_manifest = self.manifest
        for target in ("darwin-arm64", "linux-arm64"):
            with self.subTest(target=target):
                self.store = MemoryOSS()
                root = self.make_candidates(target, "1.0.0", target)
                self.manifest, self.uploads = candidate.assemble(
                    root, "single-" + target, "1.0.0", "0.2.4", self.sha, "server-contract-v1",
                    astra_url="https://astra.example.test", platforms=target)
                self.assertEqual({target}, set(self.manifest["platforms"]))
                self.assertEqual(3, len(self.uploads))
                self.publish()
                self.assertEqual(6, len(self.store.writes))
                with self.assertRaisesRegex(ValueError, "four platforms"):
                    self.promote()
                self.assertNotIn("channels/stable.json", self.store.objects)
                installer = self.store.get(f"bundles/single-{target}/install.sh").decode()
                self.assertIn(target + ") binary_url=", installer)
                self.assertNotIn("amd64) binary_url=", installer)
                # A later full bundle may cover the single-platform release. The reverse drops platforms.
                publisher.validate_stable(full_manifest, 2001, self.manifest)
                with self.assertRaisesRegex(ValueError, "drop platforms"):
                    publisher.validate_stable(self.manifest, 2001, full_manifest)
                downgrade = copy.deepcopy(full_manifest)
                for platform in downgrade["platforms"].values():
                    platform["components"]["moi-cli"]["version"] = "0.9.0"
                with self.assertRaisesRegex(ValueError, "downgrade"):
                    publisher.validate_stable(downgrade, 2001, self.manifest)
                # Default all must NOT silently assemble an incomplete artifact set.
                with self.assertRaisesRegex(ValueError, "selected platforms"):
                    self.assemble(root)
                with self.assertRaisesRegex(ValueError, "selected platforms"):
                    candidate.assemble(self.candidates, "extra", "1.0.0", "0.2.4", self.sha,
                                       "server-contract-v1", astra_url="https://astra.example.test", platforms=target)

    def test_candidate_upload_does_not_promote_and_is_idempotent(self):
        self.publish()
        self.assertEqual(15, len(self.store.writes))  # components, source, manifest, installer
        self.assertNotIn("channels/stable.json", self.store.objects)
        self.assertTrue(all(immutable for _, immutable in self.store.writes))
        self.publish()
        self.assertEqual(15, len(self.store.writes))

    def test_bootstrap_is_pinned_and_must_be_public_before_promotion(self):
        self.publish()
        item = self.manifest["installer"]
        script = self.store.objects[item["key"]]
        self.assertEqual(candidate.digest(script), item["sha256"])
        self.assertEqual(len(script), item["size"])
        for platform in self.manifest["platforms"].values():
            self.assertIn(platform["components"]["moi-cli"]["sha256"].encode(), script)
        syntax = subprocess.run(["sh", "-n"], input=script, capture_output=True)
        self.assertEqual(0, syntax.returncode, syntax.stderr)
        del self.store.objects[item["key"]]
        self.store.writes.clear()
        with self.assertRaises(ValueError):
            self.promote()
        self.assertEqual([], self.store.writes)

    def test_channel_is_plain_json_and_written_last(self):
        self.publish(True)
        self.assertEqual(("channels/stable.json", False), self.store.writes[-1])
        self.assertEqual((publisher.STABLE_INSTALLER_KEY, False), self.store.writes[-2])
        channel = json.loads(self.store.objects["channels/stable.json"])
        manifest_bytes = self.store.objects["bundles/2026.09.17.1/manifest.json"]
        self.assertEqual(candidate.digest(manifest_bytes), channel["manifest_sha256"])
        self.assertEqual(self.manifest, json.loads(manifest_bytes))
        self.assertEqual("2026-10-17T00:00:00+00:00", channel["expires_at"])
        installer = self.manifest["installer"]
        self.assertEqual(self.store.objects[installer["key"]], self.store.objects[publisher.STABLE_INSTALLER_KEY])
        self.assertIn(publisher.STABLE_INSTALLER_KEY, self.store.public_reads)

    def test_only_moi_change_reuses_astra_artifacts(self):
        self.publish(True)
        self.store.writes.clear()
        root = self.make_candidates("second", "1.0.1")
        self.manifest, self.uploads = self.assemble(root, "1.0.1", "2026.09.17.2")
        self.publish(True, 2001)
        components = [key for key, _ in self.store.writes if key.startswith("components/")]
        self.assertEqual(4, len(components))
        self.assertTrue(all(key.startswith("components/moi-cli/1.0.1/") for key in components))

    def test_conflicting_component_is_rejected_before_any_write(self):
        self.store.objects[next(iter(self.uploads))] = b"different bytes"
        with self.assertRaisesRegex(ValueError, "immutable release conflict"):
            self.publish(True)
        self.assertEqual([], self.store.writes)

    def test_conflicting_bundle_is_rejected_before_any_write(self):
        self.store.objects["bundles/2026.09.17.1/manifest.json"] = b"different manifest"
        with self.assertRaisesRegex(ValueError, "immutable release conflict"):
            self.publish(True)
        self.assertEqual([], self.store.writes)

    def test_partial_upload_never_promotes_and_retry_completes(self):
        self.store.fail_key = list(self.uploads)[2]
        with self.assertRaises(OSError):
            self.publish(True)
        self.assertNotIn("channels/stable.json", self.store.objects)
        self.assertNotIn("bundles/2026.09.17.1/manifest.json", self.store.objects)
        self.store.fail_key = None
        self.publish(True)
        self.assertEqual(17, len(self.store.writes))  # complete upload plus stable installer and channel

    def test_bad_readback_never_promotes(self):
        original = self.store.put

        def corrupt(key, data, immutable):
            original(key, b"corrupted", immutable)

        self.store.put = corrupt
        with self.assertRaisesRegex(ValueError, "readback mismatch"):
            self.publish(True)
        self.assertNotIn("channels/stable.json", self.store.objects)

    def test_new_release_failure_preserves_existing_stable(self):
        self.publish(True)
        previous = self.store.objects["channels/stable.json"]
        root = self.make_candidates("new", "1.0.1")
        self.manifest, self.uploads = self.assemble(root, "1.0.1", "2026.09.17.2")
        self.store.fail_key = "bundles/2026.09.17.2/manifest.json"
        with self.assertRaises(OSError):
            self.publish(True, 2001)
        self.assertEqual(previous, self.store.objects["channels/stable.json"])
        self.store.fail_key = None
        self.publish(True, 2002)
        channel = json.loads(self.store.objects["channels/stable.json"])
        self.assertEqual("2026.09.17.2", channel["bundle_version"])

    def test_stable_installer_failure_does_not_advance_channel_and_retry_repairs_it(self):
        self.publish(True)
        old_channel = self.store.objects["channels/stable.json"]
        old_installer = self.store.objects[publisher.STABLE_INSTALLER_KEY]
        root = self.make_candidates("new", "1.0.1")
        self.manifest, self.uploads = self.assemble(root, "1.0.1", "2026.09.17.2")
        self.publish()
        self.store.writes.clear()
        self.store.fail_key = publisher.STABLE_INSTALLER_KEY
        with self.assertRaises(OSError):
            self.promote(2001)
        self.assertEqual(old_channel, self.store.objects["channels/stable.json"])
        self.assertEqual(old_installer, self.store.objects[publisher.STABLE_INSTALLER_KEY])
        self.assertEqual([], self.store.writes)
        self.store.fail_key = None
        self.promote(2001)
        self.assertEqual(self.store.objects[self.manifest["installer"]["key"]],
                         self.store.objects[publisher.STABLE_INSTALLER_KEY])
        self.assertEqual([(publisher.STABLE_INSTALLER_KEY, False), ("channels/stable.json", False)],
                         self.store.writes)

    def test_channel_failure_after_installer_write_can_retry_same_sequence(self):
        self.publish(True)
        old_channel = self.store.objects["channels/stable.json"]
        root = self.make_candidates("new", "1.0.1")
        self.manifest, self.uploads = self.assemble(root, "1.0.1", "2026.09.17.2")
        self.publish()
        self.store.writes.clear()
        self.store.fail_key = "channels/stable.json"
        with self.assertRaises(OSError):
            self.promote(2001)
        self.assertEqual(old_channel, self.store.objects["channels/stable.json"])
        self.assertEqual(self.store.objects[self.manifest["installer"]["key"]],
                         self.store.objects[publisher.STABLE_INSTALLER_KEY])
        self.assertEqual([(publisher.STABLE_INSTALLER_KEY, False)], self.store.writes)
        self.store.fail_key = None
        self.promote(2001)
        self.assertEqual("2026.09.17.2", json.loads(self.store.objects["channels/stable.json"])["bundle_version"])
        self.assertEqual([(publisher.STABLE_INSTALLER_KEY, False), ("channels/stable.json", False)],
                         self.store.writes)

    def test_changed_channel_during_verification_refuses_promotion(self):
        self.publish()
        original_verify = self.store.verify_public

        def concurrent_writer(key, sha, size):
            original_verify(key, sha, size)
            self.store.objects["channels/stable.json"] = b"external update"

        self.store.verify_public = concurrent_writer
        with self.assertRaisesRegex(ValueError, "changed during publication"):
            self.promote()
        self.assertEqual(b"external update", self.store.objects["channels/stable.json"])
        self.assertTrue(all(key != "channels/stable.json" for key, _ in self.store.writes))

    def test_stale_promotion_is_rejected(self):
        self.publish(True, 2001)
        self.store.writes.clear()
        for sequence in (1001, 2001):
            with self.subTest(sequence=sequence), self.assertRaisesRegex(ValueError, "stale"):
                self.publish(True, sequence)
        self.assertEqual([], self.store.writes)

    def test_historical_bundle_cannot_downgrade_a_component(self):
        old_manifest = self.manifest
        old_sha = candidate.digest(candidate.canonical(old_manifest))
        self.publish(True, 1001)
        root = self.make_candidates("newer", "1.0.1")
        self.manifest, self.uploads = self.assemble(root, "1.0.1", "2026.09.17.2")
        self.publish(True, 2001)
        stable = self.store.objects["channels/stable.json"]
        self.store.writes.clear()
        self.manifest = old_manifest
        with self.assertRaisesRegex(ValueError, "downgrade moi-cli"):
            self.promote(3001, old_sha)
        self.assertEqual(stable, self.store.objects["channels/stable.json"])
        self.assertEqual([], self.store.writes)

    def test_service_change_requires_new_compatibility_reference(self):
        self.publish(True, 1001)
        root = self.make_candidates("service-change", "1.0.1")
        self.manifest, self.uploads = self.assemble(
            root, "1.0.1", "2026.09.17.2", astra_url="https://astra-new.example.test")
        self.publish()
        stable = self.store.objects["channels/stable.json"]
        self.store.writes.clear()
        with self.assertRaisesRegex(ValueError, "Prod Astra URL"):
            self.promote(2001)
        self.assertEqual(stable, self.store.objects["channels/stable.json"])
        self.assertEqual([], self.store.writes)

    def test_service_change_with_new_compatibility_still_cannot_promote(self):
        self.publish(True, 1001)
        root = self.make_candidates("new-contract", "1.0.1")
        self.manifest, self.uploads = self.assemble(
            root, "1.0.1", "2026.09.17.2", compatibility="server-contract-v2",
            astra_url="https://astra-new.example.test")
        self.publish()
        stable = self.store.objects["channels/stable.json"]
        self.store.writes.clear()
        with self.assertRaisesRegex(ValueError, "Prod Astra URL"):
            self.promote(2001)
        self.assertEqual(stable, self.store.objects["channels/stable.json"])
        self.assertEqual([], self.store.writes)

    def test_prerelease_cannot_promote(self):
        for platform in self.manifest["platforms"].values():
            platform["astra_source"]["prerelease"] = True
        with self.assertRaisesRegex(ValueError, "prerelease"):
            self.publish()
        self.assertEqual([], self.store.writes)
        self.assertNotIn("channels/stable.json", self.store.objects)

    def test_invalid_previous_channel_refuses_writes(self):
        self.publish()
        self.store.writes.clear()
        for previous in (b"not-json", b"[]", b"null", b"{}", candidate.canonical({
                "schema": "moi-client-channel-v1", "channel": "stable", "sequence": "1000"})):
            with self.subTest(previous=previous):
                self.store.objects["channels/stable.json"] = previous
                with self.assertRaises(ValueError):
                    self.promote()
                self.assertEqual([], self.store.writes)

    def test_publication_entrypoint_requires_only_oss_credentials(self):
        summary = self.root / "summary.md"
        env = {"BUNDLE_VERSION": "2026.09.22.1-qa", "RELEASE_ENVIRONMENT": "qa",
               "MOI_VERSION": "1.0.0", "ASTRA_VERSION": "0.2.4",
               "ASTRA_MODE": "release", "ASTRA_URL": "https://astra-qa.example.test",
               "MOI_RELEASE_ASTRA_URL_QA": "https://astra-qa.example.test",
               "GITHUB_SHA": self.sha, "COMPATIBILITY_REFERENCE": "server-contract-v1",
               "GITHUB_RUN_NUMBER": "1", "GITHUB_RUN_ATTEMPT": "1",
               "GITHUB_STEP_SUMMARY": str(summary),
               "MOI_RELEASE_OSS_ACCESS_KEY_ID": "synthetic", "MOI_RELEASE_OSS_ACCESS_KEY_SECRET": "synthetic"}
        with patch.dict(os.environ, env, clear=True), patch.object(publisher, "OSSStore", return_value=self.store), \
                patch.object(sys, "argv", ["release_publish.py", "upload", "--root", str(self.candidates)]), \
                patch("builtins.print"):
            publisher.main()
            uploaded = summary.read_text()
            self.assertIn("Channel: not updated", uploaded)
            self.assertNotIn("channels/qa.json", uploaded)
            manifest_sha = candidate.digest(self.store.objects["bundles/2026.09.22.1-qa/manifest.json"])
            with patch.object(candidate, "download", side_effect=self.download_published), \
                    patch.object(sys, "argv", ["release_publish.py", "publish-channel",
                                              "--bundle-version", "2026.09.22.1-qa",
                                              "--manifest-sha256", manifest_sha]):
                publisher.main()
        self.assertIn("channels/qa.json", self.store.objects)
        self.assertNotIn("channels/stable.json", self.store.objects)

    def test_missing_platform_or_extra_file_fails(self):
        folder = self.candidates / "linux-arm64"
        moved = self.root / "missing"
        folder.rename(moved)
        with self.assertRaisesRegex(ValueError, "selected platforms"):
            self.assemble(self.candidates)
        moved.rename(folder)
        (folder / ".env").write_text("synthetic must never be published")
        with self.assertRaisesRegex(ValueError, "unexpected files"):
            self.assemble(self.candidates)

    def test_tampered_archive_or_wrong_source_fails(self):
        archive = next(iter(self.uploads.values()))
        original = archive.read_bytes()
        archive.write_bytes(b"tampered")
        with self.assertRaisesRegex(ValueError, "digest/size mismatch"):
            self.assemble(self.candidates)
        archive.write_bytes(original)
        with self.assertRaisesRegex(ValueError, "identity mismatch"):
            candidate.assemble(self.candidates, "one", "1.0.0", "0.2.4", "c" * 40, "contract-v1")

    def test_versions_reject_path_traversal_and_shell_injection(self):
        for value in ("../x", "v1.0.0", "1.0.0\n", "1.0.0;touch", "$(id)", "1.0", "01.0.0", "0.1.5-rc1", "1.2.3-rc1"):
            with self.subTest(value=value), self.assertRaises(ValueError):
                candidate.require_version(value)
        for value in ("../x", "x/y", "x\ny", "-bad"):
            with self.subTest(value=value), self.assertRaises(ValueError):
                candidate.require_bundle(value)

    def test_deterministic_packaging_and_symlink_refusal(self):
        binary = self.root / "moi-cli"
        first = candidate.package_component(binary)
        os.utime(binary, (100, 100))
        self.assertEqual(first, candidate.package_component(binary))
        link = self.root / "link"
        link.symlink_to(binary)
        with self.assertRaisesRegex(ValueError, "regular file"):
            candidate.package_component(link)

    def test_archive_traversal_links_duplicates_and_special_modes_rejected(self):
        for kind in ("traversal", "symlink", "duplicate", "suid"):
            with self.subTest(kind=kind):
                buffer = io.BytesIO()
                with tarfile.open(fileobj=buffer, mode="w:gz") as archive:
                    member = tarfile.TarInfo("../escape" if kind == "traversal" else "moi-cli")
                    member.size = 1
                    if kind == "symlink":
                        member.type, member.linkname = tarfile.SYMTYPE, "/tmp/escape"
                    if kind == "suid":
                        member.mode = 0o4755
                    archive.addfile(member, io.BytesIO(b"x"))
                    if kind == "duplicate":
                        archive.addfile(member, io.BytesIO(b"x"))
                with self.assertRaises(ValueError):
                    candidate.archive_members(buffer.getvalue(), ("moi-cli",))

    def test_fetch_requires_exact_public_release_and_digest(self):
        release = {"tag_name": "v0.2.4", "draft": False, "assets": []}
        with patch.object(candidate, "download", return_value=candidate.canonical(release)):
            with self.assertRaisesRegex(ValueError, "digest unavailable"):
                candidate.fetch_astra("0.2.4", "linux-amd64", self.root / "download")
        release["draft"] = True
        with patch.object(candidate, "download", return_value=candidate.canonical(release)):
            with self.assertRaisesRegex(ValueError, "published exact tag"):
                candidate.fetch_astra("0.2.4", "linux-amd64", self.root / "download")

    def test_fetch_verifies_digest_before_extracting(self):
        release = {"tag_name": "v0.2.4", "draft": False, "id": 123, "assets": [{
            "name": "astra-v0.2.4-linux-amd64.tar.gz", "id": 456, "state": "uploaded",
            "browser_download_url": "https://github.com/matrixorigin/Astra/releases/download/v0.2.4/astra-v0.2.4-linux-amd64.tar.gz",
            "size": 3, "digest": "sha256:" + "a" * 64,
        }]}
        with patch.object(candidate, "download", side_effect=[candidate.canonical(release), b"bad"]):
            with self.assertRaisesRegex(ValueError, "digest mismatch"):
                candidate.fetch_astra("0.2.4", "linux-amd64", self.root / "download")
        self.assertFalse((self.root / "download").exists())

    def test_non_owner_publication_request_fails_preflight_without_secrets(self):
        script = Path(candidate.__file__)
        env = {**os.environ, "RELEASE_OPERATION": "build", "RELEASE_ENVIRONMENT": "qa",
               "BUNDLE_VERSION": "2026.09.22.1-qa", "MOI_VERSION": "1.0.0", "ASTRA_VERSION": "0.2.4",
               "ASTRA_MODE": "release", "ASTRA_URL": "https://astra.example.test",
               "MOI_RELEASE_ASTRA_URL_QA": "https://astra.example.test",
               "COMPATIBILITY_REFERENCE": "contract-v1",
               "GITHUB_SHA": self.sha, "GITHUB_REPOSITORY": "someone/fork", "GITHUB_REF": "refs/heads/dev"}
        for dry, expected in (("true", 0), ("false", 1)):
            result = subprocess.run([sys.executable, str(script), "validate-inputs"],
                                    env={**env, "RELEASE_DRY_RUN": dry}, capture_output=True, timeout=10)
            self.assertEqual(expected, result.returncode)
    def test_invalid_storage_refuses_all_writes(self):
        self.store.invalid_storage = True
        with self.assertRaises(ValueError):
            self.publish(True)
        self.assertEqual([], self.store.writes)

    def test_private_object_cannot_be_promoted(self):
        self.store.private_key = next(iter(self.uploads))
        with self.assertRaisesRegex(ValueError, "anonymous download"):
            self.publish(True)
        self.assertNotIn("channels/stable.json", self.store.objects)

    def test_public_read_uses_fixed_origin_without_authorization(self):
        store = publisher.OSSStore.__new__(publisher.OSSStore)
        key = "components/moi-cli/1.0.0/linux-arm64.tar.gz"
        with patch("urllib.request.build_opener") as factory:
            fetch = factory.return_value.open
            response = fetch.return_value.__enter__.return_value
            response.url = candidate.PUBLIC_URL + "/" + key
            response.read.side_effect = io.BytesIO(b"data").read
            store.verify_public(key, candidate.digest(b"data"), 4)
            request = fetch.call_args.args[0]
            self.assertEqual(candidate.PUBLIC_URL + "/" + key, request.full_url)
            self.assertNotIn("Authorization", request.headers)
            response.read.side_effect = io.BytesIO(b"bad").read
            with self.assertRaisesRegex(ValueError, "readback mismatch"):
                store.verify_public(key, candidate.digest(b"data"), 4)

    def test_oss_adapter_never_sets_acl_or_overwrites_immutable_objects(self):
        with patch.dict(os.environ, {"MOI_RELEASE_OSS_ACCESS_KEY_ID": "synthetic",
                                     "MOI_RELEASE_OSS_ACCESS_KEY_SECRET": "synthetic"}):
            with patch("oss2.Bucket") as factory:
                store = publisher.OSSStore()
                self.assertEqual("astra-suite", factory.call_args.args[2])
                for status in ("Enabled", "Suspended"):
                    factory.return_value.get_bucket_versioning.return_value.status = status
                    with self.assertRaisesRegex(ValueError, "unversioned"):
                        store.validate_storage()
                factory.return_value.get_bucket_versioning.return_value.status = ""
                store.validate_storage()
                store.put("components/test", b"data", True)
                headers = factory.return_value.put_object.call_args.kwargs["headers"]
                self.assertEqual("true", headers["x-oss-forbid-overwrite"])
                self.assertNotIn("x-oss-object-acl", headers)
                store.put("channels/stable.json", b"data", False)
                headers = factory.return_value.put_object.call_args.kwargs["headers"]
                self.assertEqual("no-store", headers["Cache-Control"])
                self.assertNotIn("x-oss-object-acl", headers)
                store.put(publisher.STABLE_INSTALLER_KEY, b"#!/bin/sh\n", False)
                headers = factory.return_value.put_object.call_args.kwargs["headers"]
                self.assertEqual("text/x-shellscript", headers["Content-Type"])
                self.assertEqual("no-store", headers["Cache-Control"])
                self.assertNotIn("x-oss-forbid-overwrite", headers)

    def source_candidates(self):
        for target in candidate.TARGETS:
            path = self.candidates / target / "candidate.json"
            data = json.loads(path.read_bytes())
            data["astra_source"] = {"mode": "source", "repository": candidate.ASTRA_REPOSITORY,
                                    "commit_sha": "d" * 40, "prerelease": False}
            path.write_bytes(candidate.canonical(data))
        return candidate.assemble(self.candidates, "source-bundle", "1.0.0", "0.2.4", self.sha,
                                  "contract-v1", "source", "d" * 40, astra_url="https://astra.example.test")

    def test_source_manifest_preserves_commit_and_same_component_contract(self):
        self.manifest, self.uploads = self.source_candidates()
        self.assertEqual(12, len(self.uploads))
        self.publish()
        self.assertNotIn("channels/stable.json", self.store.objects)
        for platform in self.manifest["platforms"].values():
            self.assertEqual("d" * 40, platform["astra_source"]["commit_sha"])
        identity = json.loads(self.store.objects["components/astra/0.2.4/source.json"])
        self.assertEqual("d" * 40, identity["commit_sha"])

    def test_source_provenance_mismatch_refuses_assembly(self):
        self.source_candidates()
        for field, value in (("commit_sha", "e" * 40), ("mode", "release"),
                             ("repository", "unknown/repository"), ("prerelease", True)):
            path = self.candidates / "linux-arm64" / "candidate.json"
            original = path.read_bytes()
            data = json.loads(original)
            data["astra_source"][field] = value
            path.write_bytes(candidate.canonical(data))
            with self.subTest(field=field), self.assertRaises(ValueError):
                candidate.assemble(self.candidates, "source-bundle", "1.0.0", "0.2.4", self.sha,
                                   "contract-v1", "source", "d" * 40)
            path.write_bytes(original)

    def test_same_version_different_source_rejected_even_with_identical_binaries(self):
        self.manifest, self.uploads = self.source_candidates()
        self.publish()
        self.store.writes.clear()
        self.manifest["bundle_version"] = "another-bundle"
        self.manifest["installer"]["key"] = "bundles/another-bundle/install.sh"
        for platform in self.manifest["platforms"].values():
            platform["astra_source"]["commit_sha"] = "e" * 40
        with self.assertRaisesRegex(ValueError, "immutable release conflict.*source.json"):
            self.publish()
        self.assertEqual([], self.store.writes)

    def test_promotion_only_never_uploads_missing_objects(self):
        with self.assertRaisesRegex(ValueError, "requires an uploaded object"):
            self.promote()
        self.assertEqual([], self.store.writes)
        self.publish()
        self.store.writes.clear()
        self.promote()
        self.assertEqual([(publisher.STABLE_INSTALLER_KEY, False), ("channels/stable.json", False)], self.store.writes)

    def test_promotion_only_requires_uploaded_manifest_and_identity(self):
        self.publish()
        for key in ("bundles/2026.09.17.1/manifest.json", "components/astra/0.2.4/source.json"):
            existing = self.store.objects.pop(key)
            self.store.writes.clear()
            with self.subTest(key=key), self.assertRaises(ValueError):
                self.promote()
            self.assertEqual([], self.store.writes)
            self.store.objects[key] = existing

    def test_resolve_source_pins_commit_and_checks_version_before_build(self):
        path, _ = self.upstream_fixture()
        responses = [b"d" * 40, b'[workspace.package]\nversion = "0.2.4"\n', path.read_bytes()]
        with patch.object(candidate, "download", side_effect=responses) as fetch:
            self.assertEqual("d" * 40, candidate.resolve_source("feature/moi", "0.2.4"))
            self.assertTrue(fetch.call_args_list[0].args[0].endswith("/commits/feature%2Fmoi"))
            self.assertIn("/" + "d" * 40 + "/Cargo.toml", fetch.call_args_list[1].args[0])
            self.assertEqual("application/vnd.github.sha", fetch.call_args_list[0].kwargs["accept"])
            self.assertIn("/" + "d" * 40 + "/.github/workflows/release-binaries.yml", fetch.call_args_list[2].args[0])
        with patch.object(candidate, "download", side_effect=responses):
            with self.assertRaisesRegex(ValueError, "version differs"):
                candidate.resolve_source("main", "0.2.5")
        for ref in ("", "--help", "../main", "main\nsha=evil", "$(id)", "https://evil"):
            with self.subTest(ref=ref), patch.object(candidate, "download") as fetch:
                with self.assertRaises(ValueError):
                    candidate.resolve_source(ref, "0.2.4")
                fetch.assert_not_called()

    def test_source_checkout_must_be_clean_pinned_and_match_workspace_version(self):
        root = self.root / "astra-source"
        root.mkdir()
        (root / "Cargo.toml").write_text('[workspace.package]\nversion = "0.2.4"\n')
        for name in ("astra-cli", "astra-edge"):
            folder = root / "crates" / name
            folder.mkdir(parents=True)
            (folder / "Cargo.toml").write_text('[package]\nversion.workspace = true\n')
        with patch.object(candidate.subprocess, "check_output", side_effect=["d" * 40, ""]):
            self.assertEqual("d" * 40, candidate.source_identity(root, "0.2.4", "d" * 40)["commit_sha"])
        for results, message in ((["e" * 40], "resolved commit"), (["d" * 40, " M Cargo.toml"], "tracked changes")):
            with patch.object(candidate.subprocess, "check_output", side_effect=results):
                with self.assertRaisesRegex(ValueError, message):
                    candidate.source_identity(root, "0.2.4", "d" * 40)
        with patch.object(candidate.subprocess, "check_output", side_effect=["d" * 40, ""]):
            with self.assertRaisesRegex(ValueError, "source version mismatch"):
                candidate.source_identity(root, "0.2.5", "d" * 40)

    def test_source_build_uses_locked_client_only_native_release_contract(self):
        (self.root / "rust-toolchain.toml").write_text('[toolchain]\nchannel = "1.97.0"\n')
        for target, rust_target in candidate.RUST_TARGETS.items():
            with self.subTest(target=target), patch.object(candidate, "native_target", return_value=target), \
                    patch.object(candidate, "source_identity") as identity, patch.object(candidate.subprocess, "run") as run, \
                    patch.object(candidate, "upstream_build_steps", return_value=[("upstream build", {"FEATURES": "test"})]) as steps:
                candidate.build_source(self.root, "0.2.4", "d" * 40)
                self.assertEqual(2, identity.call_count)
                steps.assert_called_once_with(self.root, target)
                self.assertEqual(["bash", "-c", "upstream build"], run.call_args.args[0])
                self.assertEqual("test", run.call_args.kwargs["env"]["FEATURES"])
                self.assertEqual("0.2.4", run.call_args.kwargs["env"]["RELEASE_VERSION"])

    def test_source_build_failure_propagates_without_packaging(self):
        (self.root / "rust-toolchain.toml").write_text('[toolchain]\nchannel = "1.97.0"\n')
        with patch.object(candidate, "source_identity"), patch.object(candidate.subprocess, "run",
                side_effect=subprocess.CalledProcessError(1, "cargo")), \
                patch.object(candidate, "upstream_build_steps", return_value=[("build", {})]):
            with self.assertRaises(subprocess.CalledProcessError):
                candidate.build_source(self.root, "0.2.4", "d" * 40)

    def test_source_build_rejects_unpinned_toolchain_before_cargo(self):
        (self.root / "rust-toolchain.toml").write_text('[toolchain]\nchannel = "stable"\n')
        with patch.object(candidate, "source_identity"), patch.object(candidate.subprocess, "run") as run:
            with self.assertRaisesRegex(ValueError, "pin a numeric Rust toolchain"):
                candidate.build_source(self.root, "0.2.4", "d" * 40)
            run.assert_not_called()

    def test_prepare_source_packages_local_binaries_and_runs_shared_skill_probe(self):
        source = self.root / "source"
        binary_root = source / "target" / candidate.RUST_TARGETS[candidate.native_target()] / "release"
        binary_root.mkdir(parents=True)
        for name in ("astra", "astra-edge"):
            (binary_root / name).write_bytes(b"synthetic binary")
        (source / "LICENSE").write_bytes(b"license")
        moi = self.root / "moi-cli"
        output = self.root / "prepared"
        provenance = {"mode": "source", "repository": candidate.ASTRA_REPOSITORY,
                      "commit_sha": "d" * 40, "prerelease": False}
        with patch.object(candidate, "source_identity", return_value=provenance), \
                patch.object(candidate, "fetch_astra") as fetch, \
                patch.object(candidate, "version_of", side_effect=["moi-cli version 1.0.0", "astra 0.2.4", "astra-edge 0.2.4"]), \
                patch.object(candidate.subprocess, "run", return_value=SimpleNamespace(stdout="moi-client-update-v1\n")) as run, \
                patch.object(candidate, "reject_host_openssl") as macho, \
                patch.object(candidate, "check_native_skill") as probe:
            candidate.prepare(moi, "1.0.0", "0.2.4", self.sha, output, source, "d" * 40)
            fetch.assert_not_called()
            self.assertEqual(5, run.call_count)  # help plus all three occupancy probes
            self.assertEqual([binary_root / "astra", binary_root / "astra-edge"],
                             [call.args[0] for call in macho.call_args_list])
            probe.assert_called_once_with(binary_root / "astra", moi, source / "LICENSE")
        data = json.loads((output / "candidate.json").read_bytes())
        self.assertEqual(provenance, data["astra_source"])
        self.assertEqual(set(candidate.COMPONENTS), set(data["components"]))

    def test_new_astra_reuses_exact_moi_archives_and_original_source_on_all_platforms(self):
        self.publish()
        original = copy.deepcopy(self.manifest)
        self.store.writes.clear()
        source = self.root / "new-astra"
        source.mkdir()
        (source / "LICENSE").write_bytes(b"license")
        prepared = self.root / "reused-moi"
        prepared.mkdir()
        provenance = {"mode": "source", "repository": candidate.ASTRA_REPOSITORY,
                      "commit_sha": "d" * 40, "prerelease": False}
        for target in candidate.TARGETS:
            binary_root = source / "target" / candidate.RUST_TARGETS[target] / "release"
            binary_root.mkdir(parents=True)
            for name in ("astra", "astra-edge"):
                (binary_root / name).write_bytes(b"new synthetic binary")
            with patch.object(candidate, "native_target", return_value=target), \
                    patch.object(candidate, "download", side_effect=self.download_published), \
                    patch.object(candidate, "source_identity", return_value=provenance), \
                    patch.object(candidate, "version_of", side_effect=["moi-cli version 1.0.0", "astra 0.2.5", "astra-edge 0.2.5"]), \
                    patch.object(candidate.subprocess, "run", return_value=SimpleNamespace(stdout="moi-client-update-v1\n")), \
                    patch.object(candidate, "reject_host_openssl"), \
                    patch.object(candidate, "check_native_skill") as probe:
                candidate.prepare(self.root / "missing-moi", "1.0.0", "0.2.5", self.sha,
                                  prepared / target, source, "d" * 40, moi_reuse_manifest=original)
                self.assertEqual("moi-cli", probe.call_args.args[1].name)
        manifest, uploads = candidate.assemble(prepared, "2026.10.10.1-qa", "1.0.0", "0.2.5",
                                               self.sha, "server-contract-v1", "source", "d" * 40,
                                               astra_url="https://astra.example.test", release_environment="qa")
        self.assertEqual(original["moi_source_sha"], manifest["moi_source_sha"])
        for target in candidate.TARGETS:
            self.assertEqual(original["platforms"][target]["components"]["moi-cli"],
                             manifest["platforms"][target]["components"]["moi-cli"])
        publisher.publish(self.store, manifest, uploads)
        self.assertFalse(any(key.startswith("components/moi-cli/") for key, _ in self.store.writes))

    def test_moi_reuse_resolution_pins_old_source_and_rejects_missing_platforms(self):
        self.publish()
        env = dict(InputTests().build_env(), MOI_MODE="reuse", MOI_VERSION="",
                   MOI_REF=self.manifest["bundle_version"],
                   MOI_MANIFEST_SHA256=candidate.digest(candidate.canonical(self.manifest)),
                   MOI_SOURCE_REF="ignored source ref",
                   GITHUB_SHA="e" * 40, GITHUB_OUTPUT=str(self.root / "moi.outputs"))
        with patch.dict(os.environ, env, clear=True), \
                patch.object(candidate, "download", side_effect=self.download_published), \
                patch.object(candidate, "resolve_source", return_value="d" * 40), \
                patch.object(sys, "argv", ["release_candidate.py", "resolve-source"]), patch("builtins.print"):
            candidate.main()
        fields = dict(line.split("=", 1) for line in (self.root / "moi.outputs").read_text().splitlines())
        self.assertEqual("1.0.0", fields["moi_version"])
        self.assertEqual(self.sha, fields["moi_source_sha"])
        with patch.object(candidate, "download", side_effect=self.download_published), \
                self.assertRaisesRegex(ValueError, "manifest digest mismatch"):
            candidate.resolve_moi(dict(env, MOI_MANIFEST_SHA256="f" * 64), candidate.TARGETS)
        partial = copy.deepcopy(self.manifest)
        partial["platforms"] = {"darwin-arm64": partial["platforms"]["darwin-arm64"]}
        with patch.object(candidate, "load_published_manifest", return_value=(partial, b"")), \
                self.assertRaisesRegex(ValueError, "all selected platforms"):
            candidate.resolve_moi(env, candidate.TARGETS)

    def test_moi_version_failure_leaves_output_retryable(self):
        output = self.root / "version-failure"
        with patch.object(candidate, "native_target", return_value="darwin-arm64"), \
                patch.object(candidate, "version_of", return_value="moi-cli version 9.9.9"):
            for attempt in range(2):
                with self.subTest(attempt=attempt), self.assertRaisesRegex(ValueError, "binary version"):
                    candidate.prepare(self.root / "moi-cli", "1.0.0", "0.2.5", self.sha, output,
                                      astra_root=self.root, astra_sha="d" * 40)
                self.assertFalse(output.exists())

    def test_reused_moi_digest_source_and_version_checked_before_execution(self):
        self.publish()
        for version, sha, expected in (("1.1.0", self.sha, "version/source"),
                                       ("1.0.0", "e" * 40, "version/source"),
                                       ("1.0.0", self.sha, "digest/size")):
            with tempfile.TemporaryDirectory(dir=self.root) as temp, \
                    patch.object(candidate, "native_target", return_value="darwin-arm64"), \
                    patch.object(candidate, "download", return_value=b"bad"), \
                    patch.object(candidate, "version_of") as execute:
                output = Path(temp) / "out"
                with self.assertRaisesRegex(ValueError, expected):
                    candidate.prepare(self.root / "missing", version, "0.2.5", sha, output,
                                      astra_root=self.root, astra_sha="d" * 40, moi_reuse_manifest=self.manifest)
                self.assertFalse(output.exists())
            execute.assert_not_called()

    def test_darwin_client_rejects_external_openssl_load_commands(self):
        portable = subprocess.CompletedProcess(
            args=["otool"], returncode=0,
            stdout="/tmp/astra:\n\t/usr/lib/libSystem.B.dylib (compatibility version 1.0.0)\n")
        linked = subprocess.CompletedProcess(
            args=["otool"], returncode=0,
            stdout="/tmp/astra:\n\t/opt/homebrew/opt/openssl@3/lib/libssl.3.dylib (compatibility version 3.0.0)\n")
        bundled = subprocess.CompletedProcess(
            args=["otool"], returncode=0,
            stdout="/tmp/astra:\n\t@rpath/libssl.3.dylib (compatibility version 3.0.0)\n")
        stand_in = subprocess.CompletedProcess(
            args=["otool"], returncode=1, stdout="", stderr="error: /tmp/astra: is not an object file")
        with patch.object(candidate.sys, "platform", "darwin"), \
                patch.object(candidate.subprocess, "run", return_value=portable):
            candidate.reject_host_openssl(Path("astra"))
        for bad in (linked, bundled):
            with patch.object(candidate.sys, "platform", "darwin"), \
                    patch.object(candidate.subprocess, "run", return_value=bad), \
                    self.assertRaisesRegex(ValueError, "OpenSSL outside the component archive"):
                candidate.reject_host_openssl(Path("astra"))
        with patch.object(candidate.sys, "platform", "darwin"), \
                patch.object(candidate.subprocess, "run", return_value=stand_in):
            candidate.reject_host_openssl(Path("astra"))
        with patch.object(candidate.sys, "platform", "linux"), \
                patch.object(candidate.subprocess, "run") as run:
            candidate.reject_host_openssl(Path("astra"))
            run.assert_not_called()

    def test_prepare_rejects_binary_version_before_skill_or_artifact_publication(self):
        for versions in (["moi-cli version 0.0.0"], ["moi-cli version 1.0.0", "astra 0.2.3"]):
            with self.subTest(versions=versions), tempfile.TemporaryDirectory(dir=self.root) as temp, \
                    patch.object(candidate, "source_identity"), \
                    patch.object(candidate, "version_of", side_effect=versions), \
                    patch.object(candidate, "check_native_skill") as probe:
                output = Path(temp) / "candidate"
                with self.assertRaisesRegex(ValueError, "version"):
                    candidate.prepare(self.root / "moi-cli", "1.0.0", "0.2.4", self.sha, output,
                                      self.root, "d" * 40)
                probe.assert_not_called()
                self.assertFalse((output / "candidate.json").exists())

    def test_resolve_source_entrypoint_records_one_sha_and_release_never_resolves_ref(self):
        env = {"BUNDLE_VERSION": "one", "MOI_VERSION": "1.0.0", "ASTRA_VERSION": "0.2.4",
               "GITHUB_SHA": self.sha, "COMPATIBILITY_REFERENCE": "contract-v1", "ASTRA_REF": "main"}
        for mode in ("source", "release"):
            output = self.root / f"{mode}.outputs"
            with patch.dict(os.environ, {**env, "ASTRA_MODE": mode, "GITHUB_OUTPUT": str(output)}, clear=True), \
                    patch.object(sys, "argv", ["release_candidate.py", "resolve-source"]), \
                    patch.object(candidate, "resolve_commit", return_value="c" * 40), \
                    patch.object(candidate, "resolve_source", return_value="d" * 40) as resolve, patch("builtins.print"):
                candidate.main()
            self.assertEqual(1 if mode == "source" else 0, resolve.call_count)
            fields = dict(line.split("=", 1) for line in output.read_text().splitlines())
            self.assertEqual("d" * 40 if mode == "source" else "", fields["astra_sha"])
            self.assertEqual("0.2.4", fields["astra_version"])
            self.assertEqual(mode, fields["origin_mode"])
            self.assertEqual("c" * 40, fields["moi_source_sha"])
            self.assertEqual(list(candidate.TARGETS), fields["targets"].split())
            self.assertEqual([{"target": t, "runner": candidate.RUNNERS[t]} for t in candidate.TARGETS],
                             json.loads(fields["matrix"])["include"])

    def test_reuse_resolves_original_version_and_source_without_astra_version(self):
        self.manifest, self.uploads = self.source_candidates()
        self.publish()
        output = self.root / "reuse.outputs"
        env = {"MOI_VERSION": "1.1.0", "GITHUB_SHA": "e" * 40, "BUNDLE_VERSION": "next-bundle",
               "COMPATIBILITY_REFERENCE": "contract-v1", "ASTRA_MODE": "reuse", "ASTRA_REF": self.manifest["bundle_version"],
               "MANIFEST_SHA256": candidate.digest(candidate.canonical(self.manifest)),
               "GITHUB_OUTPUT": str(output)}
        with patch.dict(os.environ, env, clear=True), \
                patch.object(candidate, "download", side_effect=self.download_published), \
                patch.object(candidate, "resolve_commit", return_value="c" * 40), \
                patch.object(candidate, "resolve_source", side_effect=AssertionError("must not resolve moving ref")), \
                patch.object(sys, "argv", ["release_candidate.py", "resolve-source"]), patch("builtins.print"):
            candidate.main()
        fields = dict(line.split("=", 1) for line in output.read_text().splitlines())
        self.assertEqual("d" * 40, fields["astra_sha"])
        self.assertEqual("0.2.4", fields["astra_version"])
        self.assertEqual("source", fields["origin_mode"])
        self.assertEqual(list(candidate.TARGETS), fields["targets"].split())

    def test_arm_resolve_outputs_one_runner_and_reuse_rejects_missing_platforms(self):
        output = self.root / "arm.outputs"
        env = dict(InputTests().build_env(), GITHUB_OUTPUT=str(output), RELEASE_PLATFORMS="darwin-arm64")
        with patch.dict(os.environ, env, clear=True), patch("builtins.print"), \
                patch.object(sys, "argv", ["release_candidate.py", "resolve-source"]), \
                patch.object(candidate, "resolve_commit", return_value="c" * 40), \
                patch.object(candidate, "resolve_source", return_value="d" * 40) as resolve:
            candidate.main()
        resolve.assert_called_once_with("main", "0.2.7", ("darwin-arm64",))
        fields = dict(line.split("=", 1) for line in output.read_text().splitlines())
        self.assertEqual("darwin-arm64", fields["targets"])
        self.assertEqual({"include": [{"target": "darwin-arm64", "runner": "macos-15"}]}, json.loads(fields["matrix"]))
        partial = copy.deepcopy(self.manifest)
        partial["platforms"] = {"darwin-arm64": partial["platforms"]["darwin-arm64"]}
        for selection in ("all", "linux-arm64"):
            env.update(ASTRA_MODE="reuse", ASTRA_REF="existing", MANIFEST_SHA256="a" * 64, RELEASE_PLATFORMS=selection)
            with patch.dict(os.environ, env, clear=True), \
                    patch.object(sys, "argv", ["release_candidate.py", "resolve-source"]), \
                    patch.object(candidate, "resolve_commit", return_value="c" * 40), \
                    patch.object(candidate, "load_published_manifest", return_value=(partial, b"unused")), \
                    self.assertRaisesRegex(ValueError, "does not contain all selected platforms"):
                candidate.main()

    def test_manifest_rejects_empty_unknown_and_unoffered_platform_sets(self):
        for platforms in ({}, {"windows-arm64": next(iter(self.manifest["platforms"].values()))},
                          {"darwin-amd64": self.manifest["platforms"]["darwin-amd64"]}):
            invalid = dict(self.manifest, platforms=platforms)
            with self.assertRaisesRegex(ValueError, "platform set"):
                candidate.validate_manifest(invalid, invalid["bundle_version"])

    def test_prepare_requires_exactly_one_origin_before_executing_binaries(self):
        for kwargs in ({}, {"astra_root": self.root, "release_root": self.root},
                       {"astra_root": self.root, "reuse_manifest": self.manifest}):
            with patch.object(candidate, "version_of") as version, self.assertRaisesRegex(ValueError, "exactly one"):
                candidate.prepare(self.root / "moi-cli", "1.0.0", "0.2.4", self.sha, self.root / "out", **kwargs)
            version.assert_not_called()

    def test_existing_bundle_promotes_and_renews_after_moi_head_changes_without_candidates(self):
        self.manifest, self.uploads = self.source_candidates()
        self.publish()
        # No local artifacts or build inputs are available to the promotion CLI.
        for path in self.uploads.values():
            path.unlink()
        self.store.writes.clear()
        args = ["release_publish.py", "promote", "--bundle-version", "source-bundle",
                "--manifest-sha256", candidate.digest(candidate.canonical(self.manifest))]
        with patch.object(publisher, "OSSStore", return_value=self.store), \
                patch.object(candidate, "download", side_effect=self.download_published), \
                patch.object(publisher, "assemble", side_effect=AssertionError("must not rebuild manifest")), \
                patch.object(sys, "argv", args), patch("builtins.print"):
            for run in (2, 3):
                with patch.dict(os.environ, {"GITHUB_SHA": "e" * 40, "GITHUB_RUN_NUMBER": str(run),
                                            "GITHUB_RUN_ATTEMPT": "1",
                                            "MOI_RELEASE_ASTRA_URL_PROD": "https://astra.example.test"}, clear=True):
                    publisher.main()
                self.assertEqual(run * 1000 + 1, json.loads(self.store.objects["channels/stable.json"])["sequence"])
        self.assertEqual([(publisher.STABLE_INSTALLER_KEY, False),
                          ("channels/stable.json", False), ("channels/stable.json", False)], self.store.writes)

    def test_promotion_preserves_original_manifest_bytes_including_whitespace(self):
        self.publish()
        raw = json.dumps(self.manifest, indent=3).encode() + b"\n\n"
        key = f"bundles/{self.manifest['bundle_version']}/manifest.json"
        self.store.objects[key] = raw
        self.promote(expected_sha=candidate.digest(raw))
        self.assertEqual(candidate.digest(raw), json.loads(self.store.objects["channels/stable.json"])["manifest_sha256"])
        self.assertEqual(raw, self.store.objects[key])

    def test_promotion_rejects_wrong_manifest_digest_and_missing_component_without_writes(self):
        self.publish()
        self.store.writes.clear()
        with self.assertRaisesRegex(ValueError, "manifest digest mismatch"):
            self.promote(expected_sha="f" * 64)
        del self.store.objects[next(iter(self.uploads))]
        with self.assertRaisesRegex(ValueError, "anonymous download"):
            self.promote()
        self.assertEqual([], self.store.writes)

    def test_public_manifest_rejects_untrusted_urls_keys_and_incomplete_platforms(self):
        for mutate in (
                lambda m: m["platforms"].pop("linux-arm64"),
                lambda m: m["platforms"]["linux-amd64"]["components"]["astra"].update(url="https://evil.invalid/asset"),
                lambda m: m["platforms"]["linux-amd64"]["components"]["astra"].update(key="channels/stable.json"),
                lambda m: m["platforms"]["linux-amd64"]["components"]["astra"].update(size=candidate.MAX_ARCHIVE + 1),
                lambda m: m.update(bundle_version="../escape"),
                lambda m: m["platforms"]["linux-amd64"]["components"].pop("astra-edge")):
            manifest = json.loads(candidate.canonical(self.manifest))
            mutate(manifest)
            data = candidate.canonical(manifest)
            with patch.object(candidate, "download", return_value=data), self.assertRaises(ValueError):
                candidate.load_published_manifest(self.manifest["bundle_version"], candidate.digest(data))

    def test_promotion_dry_run_uses_no_oss_credentials_or_writes(self):
        self.publish()
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), \
                patch.object(publisher, "OSSStore", side_effect=AssertionError("no credentials in dry-run")), \
                patch.object(publisher, "verify_public", side_effect=lambda url, sha, size:
                             self.store.verify_public(url.removeprefix(candidate.PUBLIC_URL + "/"), sha, size)), \
                patch.object(sys, "argv", ["release_publish.py", "promote", "--bundle-version", self.manifest["bundle_version"],
                                          "--manifest-sha256", candidate.digest(candidate.canonical(self.manifest)), "--dry-run"]), \
                patch.dict(os.environ, {"GITHUB_RUN_NUMBER": "2", "GITHUB_RUN_ATTEMPT": "1",
                                       "MOI_RELEASE_ASTRA_URL_PROD": "https://astra.example.test"}, clear=True), patch("builtins.print"):
            publisher.main()
        self.assertEqual([], self.store.writes)

    def test_promote_environment_dry_run_summary_is_marked(self):
        qa_url = "https://astra-qa.example.test"
        prod_url = "https://astra.example.test"
        self.manifest, self.uploads = candidate.assemble(
            self.candidates, "2026.09.22.1-qa", "1.0.0", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        summary = self.root / "prod-summary.md"
        sha = candidate.digest(candidate.canonical(self.manifest))
        with patch.object(candidate, "download", side_effect=self.download_published), \
                patch.object(publisher, "read_public_channel", return_value=None), \
                patch.object(publisher, "OSSStore", side_effect=AssertionError("no credentials in dry-run")), \
                patch.object(publisher, "verify_public", side_effect=lambda url, digest_value, size:
                             self.store.verify_public(url.removeprefix(candidate.PUBLIC_URL + "/"), digest_value, size)), \
                patch.object(sys, "argv", ["release_publish.py", "promote-environment", "--dry-run",
                                          "--bundle-version", "2026.09.22.1-qa", "--manifest-sha256", sha]), \
                patch.dict(os.environ, {"GITHUB_RUN_NUMBER": "2", "GITHUB_RUN_ATTEMPT": "1",
                                       "ASTRA_URL": prod_url, "MOI_RELEASE_ASTRA_URL_PROD": prod_url,
                                       "GITHUB_STEP_SUMMARY": str(summary)}, clear=True), patch("builtins.print"):
            publisher.main()
        text = summary.read_text()
        self.assertIn("Dry run: no OSS objects were written.", text)
        self.assertNotIn("channels/stable.json", self.store.objects)

    def test_reuse_source_archives_without_rebuilding_or_repacking_astra(self):
        self.manifest, self.uploads = self.source_candidates()
        # Model a non-canonical Rust release archive: reuse must preserve its
        # exact bytes, not rely on rebuilding or deterministic re-packaging.
        for platform in self.manifest["platforms"].values():
            for name in ("astra", "astra-edge"):
                item = platform["components"][name]
                path = self.uploads[item["key"]]
                data = gzip.compress(gzip.decompress(path.read_bytes()), mtime=12345)
                path.write_bytes(data)
                item.update(sha256=candidate.digest(data), size=len(data))
        self.publish()
        baseline = self.manifest
        self.store.writes.clear()
        resolved = self.root / "moi-only.outputs"
        env = dict(InputTests().build_env(), MOI_VERSION="1.0.1", MOI_SOURCE_REF="5.2.2-dev",
                   ASTRA_MODE="reuse", ASTRA_REF=baseline["bundle_version"],
                   MANIFEST_SHA256=candidate.digest(candidate.canonical(baseline)),
                   GITHUB_SHA="e" * 40, GITHUB_OUTPUT=str(resolved))

        def download(url, limit, **kwargs):
            if url == "https://api.github.com/repos/matrixorigin/matrixflow/commits/5.2.2-dev":
                return ("f" * 40).encode()
            return self.download_published(url, limit, **kwargs)

        with patch.dict(os.environ, env, clear=True), patch.object(candidate, "download", side_effect=download), \
                patch.object(sys, "argv", ["release_candidate.py", "resolve-source"]), patch("builtins.print"):
            candidate.main()
        fields = dict(line.split("=", 1) for line in resolved.read_text().splitlines())
        self.assertEqual("f" * 40, fields["moi_source_sha"])
        self.assertEqual("d" * 40, fields["astra_sha"])
        output = self.root / "reused-candidates"
        moi = self.root / "moi-cli"
        moi.write_bytes(b"new moi-cli 1.0.1")
        original_pack = candidate.package_component
        def package(binary, license_data=None):
            self.assertEqual("moi-cli", binary.name)
            return original_pack(binary, license_data)
        for target in candidate.TARGETS:
            with patch.object(candidate, "native_target", return_value=target), \
                    patch.object(candidate, "download", side_effect=self.download_published), \
                    patch.object(candidate, "verify_public", side_effect=lambda url, sha, size:
                                 self.store.verify_public(url.removeprefix(candidate.PUBLIC_URL + "/"), sha, size)), \
                    patch.object(candidate, "version_of", side_effect=["moi-cli version 1.0.1", "astra 0.2.4", "astra-edge 0.2.4"]), \
                    patch.object(candidate.subprocess, "run", return_value=SimpleNamespace(stdout="moi-client-update-v1\n")) as run, patch.object(candidate, "check_native_skill"), \
                    patch.object(candidate, "reject_host_openssl") as macho, \
                    patch.object(candidate, "build_source", side_effect=AssertionError("must not build")), \
                    patch.object(candidate, "package_component", side_effect=package):
                candidate.prepare(moi, fields["moi_version"], fields["astra_version"], fields["moi_source_sha"],
                                  output / target, reuse_manifest=baseline)
                self.assertEqual(3, sum(call.args[0][-1] == "--moi-update-protocol" for call in run.call_args_list))
                self.assertTrue(all(call.args[0][-1] in ("--help", "--moi-update-protocol") for call in run.call_args_list))
                self.assertEqual(2, macho.call_count)
        self.manifest, self.uploads = candidate.assemble(output, "new-moi-only", "1.0.1", "0.2.4", fields["moi_source_sha"],
                                                       "contract-v2", "source", "d" * 40, astra_url="https://astra.example.test")
        self.publish()
        self.assertEqual(fields["moi_source_sha"], self.manifest["moi_source_sha"])
        self.assertEqual(6, len(self.store.writes))  # four moi binaries + manifest + installer
        for target in candidate.TARGETS:
            for name in ("astra", "astra-edge"):
                self.assertEqual(baseline["platforms"][target]["components"][name],
                                 self.manifest["platforms"][target]["components"][name])

    def test_publication_reads_each_object_only_once_per_phase(self):
        self.publish()
        self.assertEqual(15, len(self.store.checks))
        self.assertEqual(15, len(self.store.public_reads))
        self.assertEqual(15, len(set(self.store.checks)))
        self.store.checks.clear()
        self.store.public_reads.clear()
        self.promote()
        self.assertEqual([], self.store.checks)  # no private component GET in promotion
        self.assertEqual(16, len(self.store.public_reads))  # twelve components + source + both installers + channel

    def test_oss_size_conflict_short_circuits_get_and_matching_object_streams_once(self):
        with patch.dict(os.environ, {"MOI_RELEASE_OSS_ACCESS_KEY_ID": "synthetic", "MOI_RELEASE_OSS_ACCESS_KEY_SECRET": "synthetic"}), \
                patch("oss2.Bucket") as factory:
            store = publisher.OSSStore()
            factory.return_value.head_object.return_value.content_length = 5
            with self.assertRaisesRegex(ValueError, "immutable release conflict"):
                store.check_existing("components/x", candidate.digest(b"data"), 4)
            factory.return_value.get_object.assert_not_called()
            factory.return_value.head_object.return_value.content_length = 4
            response = factory.return_value.get_object.return_value
            response.read.side_effect = io.BytesIO(b"data").read
            self.assertTrue(store.check_existing("components/x", candidate.digest(b"data"), 4))
            factory.return_value.get_object.assert_called_once()
            response.close.assert_called_once()

    def test_transport_credentials_only_reach_nonredirecting_github_api(self):
        for url, api in (("https://api.github.com/repos/matrixorigin/Astra/commits/main", True),
                         ("https://github.com/matrixorigin/Astra/releases/download/v1/file", False),
                         (candidate.PUBLIC_URL + "/bundles/one/manifest.json", False)):
            with self.subTest(url=url), patch.dict(os.environ, {"MOI_RELEASE_GITHUB_TOKEN": "synthetic-token"}), \
                    patch("urllib.request.build_opener") as factory:
                response = factory.return_value.open.return_value.__enter__.return_value
                response.url, response.read.return_value = url, b"sha"
                self.assertEqual(b"sha", transport.download(url, 128, github_api=api))
                request = factory.return_value.open.call_args.args[0]
                self.assertEqual(api, "Authorization" in request.headers)
                self.assertIsInstance(factory.call_args.args[0], transport.NoRedirect)
        with self.assertRaises(ValueError):
            transport.download("https://github.com/release", 128, github_api=True)
        with self.assertRaises(ValueError):
            transport.download("https://api.github.com/release", 128, github_api=True, redirects=True)
        self.assertIsNone(transport.NoRedirect().redirect_request(None, None, 302, None, None, "https://evil.invalid"))

    def test_transport_retries_only_bounded_transient_reads(self):
        for status, attempts in ((429, 3), (503, 3), (403, 1), (404, 1)):
            with self.subTest(status=status), patch.object(transport.time, "sleep") as sleep:
                error = urllib.error.HTTPError("https://api.github.com", status, "synthetic", {}, None)
                with patch("urllib.request.build_opener") as factory:
                    factory.return_value.open.side_effect = error
                    with self.assertRaises(urllib.error.HTTPError):
                        transport.download("https://api.github.com", 128, github_api=True)
                    self.assertEqual(attempts, factory.return_value.open.call_count)
                    self.assertEqual(attempts - 1, sleep.call_count)
        with patch.object(transport.time, "sleep") as sleep:
            with self.assertRaises(ValueError):
                transport.retry_read(lambda: (_ for _ in ()).throw(ValueError("integrity failure")))
            sleep.assert_not_called()

    def test_stream_verification_is_bounded_and_rejects_truncation_and_oversize(self):
        content = b"x" * (3 * 1024 * 1024)
        class BoundedStream(io.BytesIO):
            def read(self, size=-1):
                if not 0 < size <= 1024 * 1024:
                    raise AssertionError("unbounded read")
                return super().read(size)
        transport.verify_stream(BoundedStream(content), candidate.digest(content), len(content))
        for bad in (content[:-1], content + b"extra", b"y" * len(content)):
            with self.assertRaises(ValueError):
                transport.verify_stream(BoundedStream(bad), candidate.digest(content), len(content))

    def test_transport_retries_wrapped_timeouts_but_not_tls_errors(self):
        import ssl
        for reason, attempts in ((TimeoutError(), 3), (ConnectionResetError(), 3),
                                 (ssl.SSLCertVerificationError(), 1)):
            with patch.object(transport.time, "sleep"), patch("urllib.request.build_opener") as factory:
                factory.return_value.open.side_effect = urllib.error.URLError(reason)
                with self.assertRaises(urllib.error.URLError):
                    transport.download(candidate.PUBLIC_URL + "/bundles/example/manifest.json", 1024)
                self.assertEqual(attempts, factory.return_value.open.call_count)

    def upstream_fixture(self):
        steps = [
            {"name": "Install pinned Rust target", "shell": "bash", "run": "rustup target add '${{ matrix.target }}'"},
            {"name": "Install musl build dependencies", "shell": "bash",
             "if": "${{ contains(matrix.target, 'unknown-linux-musl') }}", "run": "echo install-musl"},
            {"name": "Build client candidates", "shell": "bash", "run": "echo upstream-build '${{ matrix.target }}'",
             "env": {"RELEASE_FEATURES": "${{ matrix.features }}"}},
            {"name": "Probe Edge TLS from the joint build", "shell": "bash",
             "run": "python3 scripts/ci/probe_edge_tls.py target/${{ matrix.target }}/release/astra-edge"},
            {"name": "Execute client candidates", "shell": "bash", "run": "echo upstream-version-check"},
            {"name": "Verify Darwin session execution lease contract", "shell": "bash",
             "if": "${{ contains(matrix.target, 'apple-darwin') }}", "run": "echo upstream-lease-tests"},
            {"name": "Package client archive", "run": "must-not-run-upstream-packaging"},
        ]
        workflow = {"jobs": {"build": {"strategy": {"matrix": {"include": [
            {"suffix": target, "target": rust, "features": "vendored" if target.startswith("linux") else ""}
            for target, rust in candidate.RUST_TARGETS.items()]}}, "steps": steps}}}
        path = self.root / ".github/workflows/release-binaries.yml"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(yaml.safe_dump(workflow))
        return path, workflow

    def test_upstream_build_adapter_executes_owned_gates_for_all_platforms(self):
        self.upstream_fixture()
        for target, rust in candidate.RUST_TARGETS.items():
            steps = candidate.upstream_build_steps(self.root, target)
            scripts = "\n".join(script for script, _ in steps)
            self.assertIn(rust, scripts)
            self.assertIn("upstream-version-check", scripts)
            self.assertIn(f"probe_edge_tls.py target/{rust}/release/astra-edge", scripts)
            self.assertLess(scripts.index("upstream-build"), scripts.index("probe_edge_tls.py"))
            self.assertLess(scripts.index("probe_edge_tls.py"), scripts.index("upstream-version-check"))
            self.assertEqual(target.startswith("darwin"), "upstream-lease-tests" in scripts)
            self.assertEqual(target.startswith("linux"), "install-musl" in scripts)
            self.assertNotIn("packaging", scripts)

    def test_upstream_build_adapter_refuses_silently_skipping_new_or_missing_gates(self):
        for mutation in (lambda w: w["jobs"]["build"]["steps"].append({"name": "new mandatory test", "run": "test"}),
                         lambda w: w["jobs"]["build"]["steps"].pop(3),
                         lambda w: w["jobs"]["build"]["steps"][2].update(run="echo '${{ secrets.UNKNOWN }}'")):
            path, workflow = self.upstream_fixture()
            mutation(workflow)
            path.write_text(yaml.safe_dump(workflow))
            with self.assertRaises(ValueError):
                candidate.upstream_build_steps(self.root, "darwin-arm64")

    def test_real_astra_tls_workflow_and_drift_fail_before_fanout(self):
        # Exact upstream workflow from Astra c26f96b6 (#848 merge), which failed
        # in MOI run 35564045151. Not a reconstruction of the adapter's whitelist.
        data = (Path(__file__).parent / "fixtures/astra-c26f96b6-release-binaries.yml").read_bytes()
        for target, rust in candidate.RUST_TARGETS.items():
            scripts = candidate.parse_upstream_build_steps(data, target)
            self.assertTrue(any(f'probe_edge_tls.py "target/{rust}/release/astra-edge"' in script
                                for script, _ in scripts))
        workflow = yaml.safe_load(data)
        workflow["jobs"]["build"]["steps"].append({"name": "new upstream gate", "shell": "bash", "run": "false"})
        with patch.object(candidate, "download", side_effect=[b"d" * 40,
                          b'[workspace.package]\nversion="0.2.7"', yaml.safe_dump(workflow).encode()]), \
                self.assertRaisesRegex(ValueError, "new upstream gate"):
            candidate.resolve_source("main", "0.2.7", ("darwin-arm64",))

    def test_workflow_consumes_selected_matrix_and_platforms_in_every_phase(self):
        path = Path(__file__).resolve().parents[3] / ".github/workflows/release-moi-cli.yml"
        workflow = yaml.safe_load(path.read_text())
        inputs = workflow[True]["workflow_dispatch"]["inputs"]
        self.assertEqual("all", inputs["platforms"]["default"])
        self.assertEqual(["all", "darwin-arm64", "linux-arm64"], inputs["platforms"]["options"])
        self.assertEqual("${{ inputs.platforms }}", workflow["env"]["RELEASE_PLATFORMS"])
        self.assertEqual("${{ fromJSON(needs.resolve.outputs.matrix) }}", workflow["jobs"]["build"]["strategy"]["matrix"])
        arrange = next(s for s in workflow["jobs"]["verify"]["steps"] if s.get("name") == "Arrange exact platform directories")
        self.assertEqual("${{ needs.resolve.outputs.targets }}", arrange["env"]["RELEASE_TARGETS"])
        self.assertIn("for target in $RELEASE_TARGETS", arrange["run"])

    def test_qa_bundle_promotes_to_prod_without_reuploading_components(self):
        qa_url = "https://astra-qa.example.test"
        prod_url = "https://astra.example.test"
        root = self.make_candidates("qa-build", "1.0.2")
        self.manifest, self.uploads = candidate.assemble(
            root, "2026.09.22.1-qa", "1.0.2", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        qa_manifest = self.store.objects["bundles/2026.09.22.1-qa/manifest.json"]
        qa_installer = self.store.objects["bundles/2026.09.22.1-qa/install.sh"]
        self.assertIn(b"channels/qa.json", qa_installer)
        self.assertNotIn(b"channels/stable.json", qa_installer)
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            publisher.publish_qa_channel(self.store, "2026.09.22.1-qa", candidate.digest(qa_manifest), 1001, self.now)
        self.assertEqual([("channels/qa.json", False)], self.store.writes)
        self.assertNotIn("channels/stable.json", self.store.objects)
        self.assertNotIn(publisher.STABLE_INSTALLER_KEY, self.store.objects)
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            publisher.promote_environment(self.store, "2026.09.22.1-qa", candidate.digest(qa_manifest),
                                          prod_url, prod_url, 1001, self.now)
        self.assertEqual(qa_manifest, self.store.objects["bundles/2026.09.22.1-qa/manifest.json"])
        written = [key for key, _ in self.store.writes]
        self.assertEqual(["bundles/2026.09.22.1/manifest.json", "bundles/2026.09.22.1/install.sh",
                          publisher.STABLE_INSTALLER_KEY, "channels/stable.json"], written)
        prod = json.loads(self.store.objects["bundles/2026.09.22.1/manifest.json"])
        self.assertEqual("prod", prod["release_environment"])
        self.assertEqual(prod_url, prod["client_config"]["astra_url"])
        self.assertEqual(self.manifest["platforms"], prod["platforms"])
        prod_installer = self.store.objects["bundles/2026.09.22.1/install.sh"]
        self.assertEqual(candidate.digest(prod_installer), prod["installer"]["sha256"])
        self.assertEqual(len(prod_installer), prod["installer"]["size"])
        self.assertEqual(prod_installer, self.store.objects[publisher.STABLE_INSTALLER_KEY])
        self.assertIn(b"channels/stable.json", prod_installer)
        self.assertNotIn(b"channels/qa.json", prod_installer)
        # Same sequence is valid on the other channel; repeating it on one channel is not.
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            with self.assertRaisesRegex(ValueError, "stale"):
                publisher.publish_qa_channel(self.store, "2026.09.22.1-qa", candidate.digest(qa_manifest), 1001, self.now)
            with self.assertRaisesRegex(ValueError, "stale"):
                publisher.promote_environment(self.store, "2026.09.22.1-qa", candidate.digest(qa_manifest),
                                              prod_url, prod_url, 1001, self.now)

    def test_existing_prod_contract_survives_qa_promotion(self):
        prod_url = "https://astra.example.test"
        qa_url = "https://astra-qa.example.test"
        self.manifest, self.uploads = self.assemble(
            self.candidates, compatibility="prod-contract-v1", astra_url=prod_url)
        self.publish(promote=True)
        old_prod_manifest = self.store.objects["bundles/2026.09.17.1/manifest.json"]

        root = self.make_candidates("qa-build-with-other-contract", "1.0.2")
        self.manifest, self.uploads = candidate.assemble(
            root, "2026.09.22.1-qa", "1.0.2", "0.2.4", self.sha, "qa-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        qa_bytes = self.store.objects["bundles/2026.09.22.1-qa/manifest.json"]
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), \
                patch.object(publisher, "download", side_effect=self.download_published), \
                patch.object(publisher, "verify_public", side_effect=lambda url, sha, size:
                             self.store.verify_public(url.removeprefix(candidate.PUBLIC_URL + "/"), sha, size)), \
                patch("builtins.print"):
            dry_run_sha = publisher.promote_environment(None, "2026.09.22.1-qa", candidate.digest(qa_bytes),
                                                        prod_url, prod_url, 2001, self.now)
        self.assertEqual([], self.store.writes)
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            published_sha = publisher.promote_environment(self.store, "2026.09.22.1-qa", candidate.digest(qa_bytes),
                                                         prod_url, prod_url, 2001, self.now)
        self.assertEqual(dry_run_sha, published_sha)

        promoted = json.loads(self.store.objects["bundles/2026.09.22.1/manifest.json"])
        self.assertEqual("prod-contract-v1", promoted["compatibility_reference"])
        self.assertEqual(self.manifest["platforms"], promoted["platforms"])
        self.assertEqual(qa_bytes, self.store.objects["bundles/2026.09.22.1-qa/manifest.json"])
        self.assertEqual(old_prod_manifest, self.store.objects["bundles/2026.09.17.1/manifest.json"])
        installer = self.store.objects["bundles/2026.09.22.1/install.sh"]
        self.assertIn(b"--compatibility prod-contract-v1", installer)
        self.assertIn(b"channels/stable.json", installer)
        self.assertNotIn(b"qa-contract-v1", installer)
        self.assertEqual(["bundles/2026.09.22.1/manifest.json", "bundles/2026.09.22.1/install.sh",
                          publisher.STABLE_INSTALLER_KEY, "channels/stable.json"],
                         [key for key, _ in self.store.writes])

        mismatched = copy.deepcopy(promoted)
        mismatched["compatibility_reference"] = "another-contract"
        with self.assertRaisesRegex(ValueError, "Prod compatibility reference"):
            publisher.validate_stable(mismatched, 3001, json.loads(old_prod_manifest))

    def test_qa_channel_rejects_downgrade_without_applying_prod_platform_rules(self):
        qa_url = "https://astra-qa.example.test"
        root = self.make_candidates("qa-current", "1.0.2")
        self.manifest, self.uploads = candidate.assemble(
            root, "2026.09.22.1-qa", "1.0.2", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        current_sha = candidate.digest(candidate.canonical(self.manifest))
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            publisher.publish_qa_channel(self.store, "2026.09.22.1-qa", current_sha, 1001, self.now)
        qa_channel = self.store.objects["channels/qa.json"]
        lower_root = self.make_candidates("qa-lower", "1.0.0")
        self.manifest, self.uploads = candidate.assemble(
            lower_root, "2026.09.21.1-qa", "1.0.0", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        lower_sha = candidate.digest(candidate.canonical(self.manifest))
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "downgrade moi-cli"):
            publisher.publish_qa_channel(self.store, "2026.09.21.1-qa", lower_sha, 2001, self.now)
        self.assertEqual(qa_channel, self.store.objects["channels/qa.json"])
        self.assertEqual([], self.store.writes)
        self.assertNotIn("channels/stable.json", self.store.objects)
        arm_root = self.make_candidates("qa-arm", "1.0.3", "darwin-arm64")
        self.manifest, self.uploads = candidate.assemble(
            arm_root, "2026.09.22.2-qa", "1.0.3", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, platforms="darwin-arm64", release_environment="qa")
        self.publish()
        arm_sha = candidate.digest(candidate.canonical(self.manifest))
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            publisher.publish_qa_channel(self.store, "2026.09.22.2-qa", arm_sha, 2001, self.now)
        self.assertEqual("2026.09.22.2-qa", json.loads(self.store.objects["channels/qa.json"])["bundle_version"])
        self.assertEqual([("channels/qa.json", False)], self.store.writes)
        self.assertNotIn("channels/stable.json", self.store.objects)

    def test_environment_promotion_rejects_bad_digest_downgrade_shrink_and_old_promote(self):
        qa_url = "https://astra-qa.example.test"
        prod_url = "https://astra.example.test"
        self.manifest, self.uploads = candidate.assemble(
            self.candidates, "2026.09.22.1-qa", "1.0.0", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        qa_sha = candidate.digest(candidate.canonical(self.manifest))
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "digest mismatch"):
            publisher.promote_environment(self.store, "2026.09.22.1-qa", "f" * 64, prod_url, prod_url, 1001, self.now)
        self.assertEqual([], self.store.writes)
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "QA bundle cannot be promoted"):
            publisher.promote_existing(self.store, "2026.09.22.1-qa", qa_sha, 1001, prod_url, self.now)
        self.assertEqual([], self.store.writes)
        historical = copy.deepcopy(self.manifest)
        del historical["release_environment"]
        historical["bundle_version"] = "2026.09.21.1"
        historical["installer"]["key"] = "bundles/2026.09.21.1/install.sh"
        historical["client_config"]["astra_url"] = "https://unknown.example.test"
        self.store.objects["bundles/2026.09.21.1/manifest.json"] = candidate.canonical(historical)
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "cannot be determined"):
            publisher.promote_existing(self.store, "2026.09.21.1",
                                       candidate.digest(candidate.canonical(historical)), 1001, prod_url, self.now)
        self.store.fail_key = "channels/qa.json"
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaises(OSError):
            publisher.publish_qa_channel(self.store, "2026.09.22.1-qa", qa_sha, 1001, self.now)
        self.assertNotIn("channels/qa.json", self.store.objects)
        self.assertNotIn("channels/stable.json", self.store.objects)
        self.assertEqual(qa_sha, candidate.digest(self.store.objects["bundles/2026.09.22.1-qa/manifest.json"]))
        self.store.fail_key = None
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            publisher.promote_environment(self.store, "2026.09.22.1-qa", qa_sha, prod_url, prod_url, 2001, self.now)
        stable = self.store.objects["channels/stable.json"]
        lower_root = self.make_candidates("lower", "0.9.0")
        self.manifest, self.uploads = candidate.assemble(
            lower_root, "2026.09.23.1-qa", "0.9.0", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        lower_sha = candidate.digest(candidate.canonical(self.manifest))
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "downgrade"):
            publisher.promote_environment(self.store, "2026.09.23.1-qa", lower_sha, prod_url, prod_url, 3001, self.now)
        self.assertEqual(stable, self.store.objects["channels/stable.json"])
        self.assertNotIn("bundles/2026.09.23.1/manifest.json", self.store.objects)
        arm_root = self.make_candidates("arm-only", "1.0.1", "darwin-arm64")
        self.manifest, self.uploads = candidate.assemble(
            arm_root, "2026.09.24.1-qa", "1.0.1", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, platforms="darwin-arm64", release_environment="qa")
        self.publish()
        arm_sha = candidate.digest(candidate.canonical(self.manifest))
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "drop platforms"):
            publisher.promote_environment(self.store, "2026.09.24.1-qa", arm_sha, prod_url, prod_url, 3001, self.now)
        self.assertEqual(stable, self.store.objects["channels/stable.json"])
        self.assertEqual([], self.store.writes)

    def test_environment_promotion_rejects_template_drift_and_qa_reread_before_stable_write(self):
        qa_url = "https://astra-qa.example.test"
        prod_url = "https://astra.example.test"
        qa_key = "bundles/2026.09.22.1-qa/manifest.json"
        self.manifest, self.uploads = candidate.assemble(
            self.candidates, "2026.09.22.1-qa", "1.0.0", "0.2.4", self.sha, "server-contract-v1",
            astra_url=qa_url, release_environment="qa")
        self.publish()
        qa_sha = candidate.digest(self.store.objects[qa_key])
        self.store.writes.clear()
        real_render = publisher.render_installer

        def drifted(manifest, channel):
            return real_render(manifest, channel) + b"# drifted\n"

        with patch.object(publisher, "render_installer", side_effect=drifted), \
                patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "installer template"):
            publisher.promote_environment(self.store, "2026.09.22.1-qa", qa_sha, prod_url, prod_url, 1001, self.now)
        self.assertEqual([], self.store.writes)
        self.assertNotIn("channels/stable.json", self.store.objects)
        self.assertNotIn("bundles/2026.09.22.1/manifest.json", self.store.objects)
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"):
            publisher.promote_environment(self.store, "2026.09.22.1-qa", qa_sha, prod_url, prod_url, 1001, self.now)
        stable = self.store.objects["channels/stable.json"]
        original = self.store.get

        def reread(key, mutate):
            if key == qa_key:
                reread.count += 1
                if reread.count > 1:
                    return mutate(original(key))
            return original(key)

        def timeout_on_qa_manifest(key):
            def fail(data):
                raise TimeoutError("synthetic metadata read timeout")
            return reread(key, fail)

        reread.count = 0
        self.store.get = timeout_on_qa_manifest
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaises(TimeoutError):
            publisher.promote_environment(self.store, "2026.09.22.1-qa", qa_sha, prod_url, prod_url, 2001, self.now)
        self.assertEqual(stable, self.store.objects["channels/stable.json"])
        self.assertEqual([], self.store.writes)
        self.assertGreater(reread.count, 1)

        def changed_qa_manifest(key):
            return reread(key, lambda data: data + b" ")

        reread.count = 0
        self.store.get = changed_qa_manifest
        self.store.writes.clear()
        with patch.object(candidate, "download", side_effect=self.download_published), patch("builtins.print"), \
                self.assertRaisesRegex(ValueError, "QA manifest changed"):
            publisher.promote_environment(self.store, "2026.09.22.1-qa", qa_sha, prod_url, prod_url, 2001, self.now)
        self.assertEqual(stable, self.store.objects["channels/stable.json"])
        self.assertEqual([], self.store.writes)

    def test_pr_gate_runs_full_suite_with_locked_dependencies(self):
        root = Path(__file__).resolve().parents[3]
        makefile = (root / "sdk/Makefile").read_text()
        static = next(line for line in makefile.splitlines() if line.startswith("test-static:"))
        self.assertIn("cli-publish-test", static)
        self.assertIn("--require-hashes -r $(CLI_DIR)/release/publish-requirements.txt", makefile)
        # Self-hosted PR runners may default to an incomplete package mirror.
        # Keep the public release lock on its authoritative index without
        # changing the runner's global pip configuration or relaxing hashes.
        install = next(line for line in makefile.splitlines() if "release-venv/bin/python -m pip install" in line)
        self.assertIn("--index-url https://pypi.org/simple --require-hashes", install)
        self.assertIn("release_pipeline_test.py", makefile)


if __name__ == "__main__":
    unittest.main(verbosity=2)
