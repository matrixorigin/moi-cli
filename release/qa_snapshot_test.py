"""Behavior tests for publication isolation and the actual rendered shell installer."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from bundle import native_target
import qa_snapshot as snapshot
import release_candidate as candidate
from release_candidate import canonical, digest, package_component, require_version, validate_manifest
from snapshot_installer import render


class Store:
    def __init__(self):
        self.objects = {"channels/stable.json": b"stable", "channels/qa.json": b"qa",
                        "components/astra/0.2.13/source.json": b"formal source"}
        self.writes = []

    def validate_storage(self):
        pass

    def check_existing(self, key, sha, size):
        if key not in self.objects:
            return False
        data = self.objects[key]
        if len(data) != size or digest(data) != sha:
            raise ValueError("immutable conflict")
        return True

    def put(self, key, data, immutable):
        if not immutable or key in self.objects:
            raise AssertionError("overwrite attempted")
        self.objects[key] = data
        self.writes.append(key)

    def verify_public(self, key, sha, size):
        assert len(self.objects[key]) == size and digest(self.objects[key]) == sha


class SnapshotTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="qa-snapshot-test-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()
        self.target = native_target()
        self.env = {
            "ASTRA_SOURCE_SHA": "a" * 40, "ASTRA_VERSION": "0.2.13",
            "MOI_SOURCE_SHA": "b" * 40, "MOI_VERSION": "", "MOI_MODE": "source",
            "GITHUB_SHA": "b" * 40, "GITHUB_RUN_ID": "1234", "GITHUB_RUN_ATTEMPT": "1",
            "SNAPSHOT_RUN_ID": "1234", "SNAPSHOT_RUN_ATTEMPT": "1",
            "MOI_RELEASE_ASTRA_URL_QA": "https://qa.example.test", "SNAPSHOT_PLATFORMS": self.target,
        }
        self.env["SNAPSHOT_ID"] = snapshot.snapshot_id("a" * 40, "1234", "1")
        self.candidates = self.root / "candidates"
        folder = self.candidates / self.target
        folder.mkdir(parents=True)
        self.binary_script = '''#!/usr/bin/env python3
import json, os, pathlib, sys
name = pathlib.Path(sys.argv[0]).name
args = sys.argv[1:]
if name == "moi-cli" and args[:2] == ["skill", "export"]:
    dest = pathlib.Path(args[args.index("--dir") + 1]); dest.mkdir(parents=True)
    (dest / "SKILL.md").write_text("---\\nname: moi-cli\\n---\\nPaired Skill")
    sys.exit(0)
if name == "astra" and "login" in args:
    auth = pathlib.Path(os.environ["MOI_AUTH_DIR"]); auth.mkdir(mode=0o700)
    (auth / "test-login").write_text("QA-only marker; not a credential")
if name == "moi-cli" and args == ["helper"]:
    os.execv(os.environ["MOI_ASTRA_BINARY"], [os.environ["MOI_ASTRA_BINARY"], "helper"])
print(json.dumps({"binary": name, "args": args, "env": {k: os.environ.get(k) for k in
    ["ASTRA_API_URL", "ASTRA_SERVER_URL", "MOI_CLI_CONFIG_FILE", "ASTRA_LOCAL_STATE_ROOT", "MOI_AUTH_DIR", "MOI_ASTRA_BINARY", "ASTRA_ACCESS_TOKEN", "HOME", "PATH"]}}))
'''
        components = {}
        for name in ("astra", "astra-edge", "moi-cli"):
            binary = self.root / name
            binary.write_text(self.binary_script)
            version = "0.1.0" if name == "moi-cli" else self.env["ASTRA_VERSION"]
            data = package_component(binary, None if name == "moi-cli" else b"License")
            filename = f"{name}-{version}-{self.target}.tar.gz"
            (folder / filename).write_bytes(data)
            components[name] = {"version": version, "filename": filename, "sha256": digest(data), "size": len(data)}
        self.candidate_path = folder / "candidate.json"
        self.candidate = {"schema": "moi-snapshot-candidate-v1", "platform": self.target,
                          "moi_source_sha": "b" * 40,
                          "astra_source": {"mode": "source", "repository": "matrixorigin/Astra",
                                           "commit_sha": "a" * 40, "prerelease": False}, "components": components}
        self.candidate_path.write_bytes(canonical(self.candidate))

    def assembled(self):
        return snapshot.assemble(self.candidates, self.env)

    def install(self, *, identity=None, modify_path=False, extra=(), downloads=None):
        manifest, uploads, script = self.assembled()
        if identity:
            manifest["snapshot_id"] = identity
            for platform, entry in manifest["platforms"].items():
                for name, item in entry["components"].items():
                    old = item["key"]
                    item["key"] = snapshot.prefix(identity) + platform + "/" + name + ".tar.gz"
                    item["url"] = snapshot.PUBLIC_URL + "/" + item["key"]
                    uploads[item["key"]] = uploads.pop(old)
            script = render(manifest)
        installer = self.root / "install.sh"
        installer.write_bytes(script)
        stubs = self.root / "stubs"
        stubs.mkdir(exist_ok=True)
        fetch = stubs / "curl"
        fetch.write_text('''#!/usr/bin/env python3
import json, os, pathlib, shutil, sys
args = sys.argv[1:]
url = args[-1]
mapping = json.loads(os.environ["SNAPSHOT_TEST_DOWNLOADS"])
shutil.copyfile(mapping[url], args[args.index("-o") + 1])
print("200", end="")
''')
        fetch.chmod(0o755)
        home = self.root / "home"
        home.mkdir(exist_ok=True)
        config = home / ".moi-cli.json"
        config.write_text("formal config sentinel")
        formal = home / ".moi"
        formal.mkdir(exist_ok=True)
        (formal / "auth.json").write_text("formal auth sentinel")
        command = ["sh", str(installer), "--dir", str(self.root / "qa-client")]
        if not modify_path:
            command.append("--no-modify-path")
        command.extend(extra)
        env = {**os.environ, "HOME": str(home), "PATH": str(stubs) + ":" + os.environ["PATH"],
               "ASTRA_API_URL": "https://prod.example.test", "ASTRA_SERVER_URL": "https://prod-edge.example.test",
               "MOI_CLI_CONFIG_FILE": str(config), "ASTRA_ACCESS_TOKEN": "synthetic-sentinel",
               "MOI_AUTH_DIR": str(formal),
               "SNAPSHOT_TEST_DOWNLOADS": json.dumps(downloads or {
                   snapshot.PUBLIC_URL + "/" + key: str(path) for key, path in uploads.items()})}
        result = subprocess.run(command, env=env, capture_output=True, text=True, timeout=30)
        self.assertEqual(config.read_text(), "formal config sentinel")
        self.assertEqual((formal / "auth.json").read_text(), "formal auth sentinel")
        return result, env

    def test_same_source_version_different_commits_publish_separately(self):
        store = Store()
        formal = dict(store.objects)
        first, _ = snapshot.publish(self.candidates, self.env, store)
        self.env["ASTRA_SOURCE_SHA"] = "c" * 40
        self.env["SNAPSHOT_ID"] = snapshot.snapshot_id("c" * 40, "1234", "1")
        self.candidate["astra_source"]["commit_sha"] = "c" * 40
        self.candidate_path.write_bytes(canonical(self.candidate))
        second, _ = snapshot.publish(self.candidates, self.env, store)
        self.assertNotEqual(first["snapshot_id"], second["snapshot_id"])
        for manifest in (first, second):
            self.assertEqual(manifest["platforms"][self.target]["components"]["astra"]["version"], "0.2.13")
            self.assertIn(snapshot.prefix(manifest["snapshot_id"]) + "manifest.json", store.objects)
        for key, value in formal.items():
            self.assertEqual(store.objects[key], value)
        self.assertTrue(all(key.startswith("qa-snapshots/") for key in store.writes))

    def test_snapshot_cannot_enter_formal_manifest_contract(self):
        manifest, _, _ = self.assembled()
        with self.assertRaises((ValueError, KeyError)):
            validate_manifest(manifest, "2026.10.30.1-qa")
        for value in ("0.2.13-dev.1", "0.2.13+sha.abc"):
            snapshot.require_source_version(value)
            with self.assertRaises(ValueError):
                require_version(value)

    def test_invalid_identity_or_candidate_is_rejected_without_writes(self):
        for key, bad in [("ASTRA_SOURCE_SHA", "d" * 40), ("MOI_SOURCE_SHA", "d" * 40),
                         ("SNAPSHOT_PLATFORMS", "all"), ("SNAPSHOT_ID", "../../components")]:
            with self.subTest(key=key):
                env = {**self.env, key: bad}
                store = Store()
                with self.assertRaises(ValueError):
                    snapshot.publish(self.candidates, env, store)
                self.assertEqual(store.writes, [])

    def test_repeated_publish_reuses_bytes_and_conflicts_precede_writes(self):
        store = Store()
        snapshot.publish(self.candidates, self.env, store)
        store.writes.clear()
        snapshot.publish(self.candidates, self.env, store)
        self.assertEqual(store.writes, [])
        key = snapshot.prefix(self.env["SNAPSHOT_ID"]) + "manifest.json"
        store.objects[key] = b"conflicting metadata"
        with self.assertRaises(ValueError):
            snapshot.publish(self.candidates, self.env, store)
        self.assertEqual(store.writes, [])

    def test_store_cannot_write_formal_or_other_snapshot_objects(self):
        store = Store()
        bounded = snapshot.SnapshotStore(store, self.env["SNAPSHOT_ID"])
        for key in ["channels/qa.json", "components/astra/0.2.13/x.tar.gz", "bundles/test/manifest.json",
                    "qa-snapshots/other/install.sh", bounded.prefix + "../install.sh"]:
            with self.subTest(key=key), self.assertRaises(ValueError):
                bounded.put(key, b"unwanted", immutable=True)
        with self.assertRaises(ValueError):
            bounded.put(bounded.prefix + "manifest.json", b"unwanted", immutable=False)
        self.assertEqual(store.writes, [])

    def test_installed_login_and_credential_helper_stay_in_snapshot(self):
        result, env = self.install(modify_path=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        root = self.root / "qa-client"
        dest = root / "snapshots" / self.env["SNAPSHOT_ID"]
        command = self.root / "home/.local/bin/astra-qa"
        output = json.loads(subprocess.check_output([str(command), "login", "--device-code"], env=env, text=True))
        self.assertEqual(output["args"], ["--api-url", "https://qa.example.test", "login", "--device-code"])
        self.assertEqual(output["env"]["MOI_AUTH_DIR"], str(dest / "auth"))
        self.assertEqual(output["env"]["ASTRA_LOCAL_STATE_ROOT"], str(dest / "state"))
        self.assertIsNone(output["env"]["ASTRA_ACCESS_TOKEN"])
        self.assertTrue((dest / "auth/test-login").is_file())
        self.assertTrue((dest / "state/skills/moi-cli/SKILL.md").is_file())
        helper = root / "bin/moi-cli-qa"
        output = json.loads(subprocess.check_output([str(helper), "helper"], env=env, text=True))
        self.assertEqual(output["binary"], "astra")
        self.assertEqual(output["args"], ["--api-url", "https://qa.example.test", "helper"])
        self.assertEqual(output["env"]["HOME"], str(dest / "moi-home"))
        self.assertEqual((self.root / "home/.moi/auth.json").read_text(), "formal auth sentinel")

    def test_inherited_explicit_config_path_cannot_escape_snapshot(self):
        result, env = self.install()
        self.assertEqual(result.returncode, 0, result.stderr)
        root = self.root / "qa-client"
        expected = root / "snapshots" / self.env["SNAPSHOT_ID"] / "moi-home/.moi-cli.json"
        for name in ("astra", "astra-edge", "moi-cli"):
            with self.subTest(launcher=name):
                out = json.loads(subprocess.check_output(
                    [str(root / "bin" / (name + "-qa")), "--version"], env=env, text=True))
                self.assertEqual(out["env"]["MOI_CLI_CONFIG_FILE"], str(expected))
        # The shared launcher environment must also protect helper subprocesses.
        helper = json.loads(subprocess.check_output(
            [str(root / "bin/moi-cli-qa"), "helper"], env=env, text=True))
        self.assertEqual(helper["env"]["MOI_CLI_CONFIG_FILE"], str(expected))
        self.assertEqual((self.root / "home/.moi-cli.json").read_text(), "formal config sentinel")

    def test_inherited_edge_server_url_is_replaced_with_qa(self):
        result, env = self.install()
        self.assertEqual(result.returncode, 0, result.stderr)
        root = self.root / "qa-client"
        for name in ("astra", "astra-edge", "moi-cli"):
            with self.subTest(launcher=name):
                out = json.loads(subprocess.check_output(
                    [str(root / "bin" / (name + "-qa")), "--version"], env=env, text=True))
                self.assertEqual(out["env"]["ASTRA_SERVER_URL"], "https://qa.example.test")
                self.assertEqual(out["env"]["ASTRA_API_URL"], "https://qa.example.test")

    def test_new_snapshot_switches_commands_but_retains_old_files_and_state(self):
        result, env = self.install()
        self.assertEqual(result.returncode, 0, result.stderr)
        root = self.root / "qa-client"
        old = root / "snapshots" / self.env["SNAPSHOT_ID"]
        old_command = old / "launchers/astra-qa"
        subprocess.check_output([str(old_command), "login"], env=env)
        new_id = snapshot.snapshot_id("a" * 40, "1234", "2")
        result, env = self.install(identity=new_id)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual((root / "active").resolve(), root / "snapshots" / new_id)
        self.assertTrue((old / "auth/test-login").is_file())
        output = json.loads(subprocess.check_output([str(old_command), "--version"], env=env, text=True))
        self.assertEqual(output["env"]["MOI_AUTH_DIR"], str(old / "auth"))

    def test_corrupt_download_does_not_activate(self):
        _, uploads, _ = self.assembled()
        corrupt = self.root / "corrupt"
        corrupt.write_bytes(b"corrupt archive")
        downloads = {snapshot.PUBLIC_URL + "/" + key: str(corrupt) for key in uploads}
        result, _ = self.install(downloads=downloads)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("mismatch", result.stderr)
        self.assertFalse((self.root / "qa-client/active").exists())
        self.assertFalse((self.root / "qa-client/.install-lock").exists())

    def test_existing_non_snapshot_install_and_commands_are_not_overwritten(self):
        root = self.root / "qa-client"
        root.mkdir()
        (root / "installation.json").write_text("formal sentinel")
        result, _ = self.install()
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual((root / "installation.json").read_text(), "formal sentinel")
        self.assertFalse((root / ".qa-snapshot-root").exists())

    def test_existing_global_command_conflict_preserves_it_and_does_not_activate(self):
        commands = self.root / "home/.local/bin"
        commands.mkdir(parents=True)
        command = commands / "astra-qa"
        command.write_text("user-owned command sentinel")
        result, _ = self.install(modify_path=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("another installation", result.stderr)
        self.assertEqual(command.read_text(), "user-owned command sentinel")
        self.assertFalse((self.root / "qa-client/active").exists())

    def test_symlink_installation_ancestor_is_rejected_without_writes(self):
        elsewhere = self.root / "elsewhere"
        elsewhere.mkdir()
        (self.root / "qa-client").symlink_to(elsewhere, target_is_directory=True)
        result, _ = self.install()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("symbolic links", result.stderr)
        self.assertEqual(list(elsewhere.iterdir()), [])

    def test_shell_metacharacters_in_install_path_are_not_executed(self):
        custom = self.root / "QA's snapshot $(false)"
        result, env = self.install(extra=("--dir", str(custom)))
        self.assertEqual(result.returncode, 0, result.stderr)
        output = json.loads(subprocess.check_output([str(custom / "bin/astra-qa"), "--version"], env=env, text=True))
        self.assertEqual(output["env"]["MOI_AUTH_DIR"], str(custom / "snapshots" / self.env["SNAPSHOT_ID"] / "auth"))

    def test_candidate_archive_digest_failure_prevents_any_oss_write(self):
        filename = self.candidate["components"]["astra"]["filename"]
        (self.candidate_path.parent / filename).write_bytes(b"changed after probing")
        store = Store()
        with self.assertRaisesRegex(ValueError, "digest/size"):
            snapshot.publish(self.candidates, self.env, store)
        self.assertEqual(store.writes, [])

    def test_reuse_resolution_preserves_published_moi_identity(self):
        env = {**self.env, "MOI_MODE": "reuse", "ASTRA_REF": "main",
               "MOI_REF": "2026.10.23.1-qa", "MOI_MANIFEST_SHA256": "d" * 64}
        with patch.object(snapshot, "resolve_source_revision", return_value=("a" * 40, "0.2.13")), \
                patch.object(snapshot, "resolve_moi", return_value=("0.1.8", "c" * 40)):
            result = snapshot.resolve(env)
        self.assertEqual(result["moi_version"], "0.1.8")
        self.assertEqual(result["moi_source_sha"], "c" * 40)

    def test_failed_job_retry_preserves_resolved_snapshot_identity(self):
        env = {**self.env, "GITHUB_RUN_ATTEMPT": "2"}
        before, _, _ = self.assembled()
        after, _, _ = snapshot.assemble(self.candidates, env)
        self.assertEqual(canonical(before), canonical(after))

    def test_snapshot_resolve_needs_no_version_inputs(self):
        env = {**self.env, "ASTRA_REF": "main"}
        with patch.object(snapshot, "resolve_source_revision", return_value=("a" * 40, "0.2.13-dev.1")):
            result = snapshot.resolve(env)
        self.assertEqual(result["astra_version"], "0.2.13-dev.1")
        self.assertEqual(result["moi_version"], "")
        self.assertEqual(result["snapshot_id"], "aaaaaaaaaaaa-1234-1")

    def test_prepare_accepts_original_prerelease_but_produces_snapshot_only_candidate(self):
        source = self.root / "source"
        version = "0.2.13-dev.1"
        binaries = source / "target" / candidate.RUST_TARGETS[self.target] / "release"
        binaries.mkdir(parents=True)
        for name in ("astra", "astra-edge"):
            path = binaries / name
            path.write_text(f'''#!/bin/sh
case "$1" in
  --version) printf '%s\\n' '{name} {version}';;
  --moi-update-protocol) printf '%s\\n' moi-client-update-v1;;
  --help) printf '%s\\n' help;;
  *) exit 1;;
esac
''')
            path.chmod(0o755)
        moi = self.root / "prepared-moi-cli"
        moi.write_text('''#!/bin/sh
case "$1" in
  --version) printf '%s\\n' 'moi-cli version 0.1.0';;
  --moi-update-protocol) printf '%s\\n' moi-client-update-v1;;
  *) exit 1;;
esac
''')
        # Packaging uses the actual component basename.
        moi = moi.rename(self.root / "moi-cli")
        moi.chmod(0o755)
        (source / "LICENSE").write_text("License")
        (source / "Cargo.toml").write_text(f'[workspace.package]\nversion = "{version}"\n')
        for name in ("astra-cli", "astra-edge"):
            directory = source / "crates" / name
            directory.mkdir(parents=True)
            (directory / "Cargo.toml").write_text('[package]\nversion.workspace = true\n')
        subprocess.run(["git", "init", "--quiet", str(source)], check=True)
        subprocess.run(["git", "add", "Cargo.toml", "crates", "LICENSE"], cwd=source, check=True)
        git_env = {**os.environ, "GIT_AUTHOR_NAME": "Fixture", "GIT_AUTHOR_EMAIL": "fixture@example.test",
                   "GIT_COMMITTER_NAME": "Fixture", "GIT_COMMITTER_EMAIL": "fixture@example.test"}
        subprocess.run(["git", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture"],
                       cwd=source, env=git_env, check=True)
        sha = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=source, text=True).strip()
        output = self.root / "prepared"
        # Native Skill/Mach-O behavior is covered separately by the real-binary smoke.
        with patch.object(candidate, "check_native_skill"), patch.object(candidate, "reject_host_openssl"):
            candidate.prepare(moi, "0.1.0", version, "b" * 40, output,
                              astra_root=source, astra_sha=sha, snapshot=True)
        prepared = json.loads((output / "candidate.json").read_bytes())
        self.assertEqual(prepared["schema"], "moi-snapshot-candidate-v1")
        self.assertEqual(prepared["astra_source"]["commit_sha"], sha)
        self.assertEqual(prepared["components"]["astra"]["version"], version)
        with self.assertRaises(ValueError):
            candidate.prepare(moi, "0.1.0", version, "b" * 40, self.root / "formal",
                              astra_root=source, astra_sha=sha)
        self.assertFalse((self.root / "formal").exists())


if __name__ == "__main__":
    unittest.main()
