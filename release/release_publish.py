#!/usr/bin/env python3
"""Upload immutable candidates or promote an existing OSS bundle, without builds."""
import argparse
import copy
from contextlib import contextmanager
from datetime import datetime, timedelta, timezone
import json
import logging
import os
from pathlib import Path
import time
from urllib.error import HTTPError
from release_candidate import (
    BUCKET, ENDPOINT, PUBLIC_URL, TARGETS, assemble, canonical, channel_key, channel_name,
    channel_url, digest, load_published_manifest, prod_version_from_qa, qa_bundle_version,
    regular_bytes, require_allowlisted_url, require_client_url, source_metadata, validate_manifest,
)
from release_transport import download, verify_public, verify_stream
from channel_installer import render as render_installer


LOGGER = logging.getLogger("moi.release")
STABLE_INSTALLER_KEY = "channels/stable/install.sh"


def log_oss(outcome, operation, key, **fields):
    LOGGER.info(json.dumps({"component": "release_publisher", "outcome": outcome,
                            "operation": operation, "object_key": key, **fields}, sort_keys=True))


@contextmanager
def oss_stage(operation, key, **fields):
    started = time.monotonic()
    log_oss("started", operation, key, **fields)
    try:
        yield
    except Exception as error:
        cause = getattr(error, "reason", None) or error.__cause__
        details = {"elapsed_ms": round((time.monotonic() - started) * 1000),
                   "error_type": type(error).__name__}
        for field in ("status", "code", "request_id"):
            value = getattr(error, field, None)
            if value is not None:
                details[field] = str(value)
        if cause is not None:
            details["cause_type"] = type(cause).__name__
            if getattr(cause, "errno", None) is not None:
                details["cause_errno"] = cause.errno
        log_oss("failed", operation, key, **fields, **details)
        raise
    else:
        log_oss("completed", operation, key,
                **fields, elapsed_ms=round((time.monotonic() - started) * 1000))


class OSSStore:
    def __init__(self):
        import oss2
        ak = os.environ.get("MOI_RELEASE_OSS_ACCESS_KEY_ID", "")
        sk = os.environ.get("MOI_RELEASE_OSS_ACCESS_KEY_SECRET", "")
        if not ak or not sk:
            raise ValueError("configure MOI_RELEASE_OSS_ACCESS_KEY_ID and MOI_RELEASE_OSS_ACCESS_KEY_SECRET")
        self.oss2 = oss2
        self.bucket = oss2.Bucket(oss2.AuthV4(ak, sk), ENDPOINT, BUCKET,
                                 region="cn-hangzhou", connect_timeout=60)

    def validate_storage(self):
        with oss_stage("get_bucket_versioning", ""):
            if self.bucket.get_bucket_versioning().status not in (None, ""):
                raise ValueError("release publication requires an unversioned bucket; no bucket setting was changed")

    def get(self, key):
        """Small control objects only; binaries are verified as streams."""
        with oss_stage("get_metadata", key):
            try:
                result = self.bucket.get_object(key)
            except self.oss2.exceptions.NoSuchKey:
                log_oss("missing", "get_metadata", key)
                return None
            try:
                data = result.read(1024 * 1024 + 1)
                if len(data) > 1024 * 1024:
                    raise ValueError("release metadata exceeds size limit")
                return data
            finally:
                result.close()

    def check_existing(self, key, expected_sha, expected_size):
        """One HEAD and at most one streamed GET. Never trust metadata SHA alone."""
        with oss_stage("head_object", key, expected_size=expected_size):
            try:
                head = self.bucket.head_object(key)
            except self.oss2.exceptions.NoSuchKey:
                log_oss("missing", "head_object", key)
                return False
        if head.content_length != expected_size:
            log_oss("size_mismatch", "head_object", key,
                    expected_size=expected_size, actual_size=head.content_length)
            raise ValueError("immutable release conflict: " + key)
        with oss_stage("verify_existing", key, expected_size=expected_size):
            response = self.bucket.get_object(key)
            try:
                verify_stream(response, expected_sha, expected_size)
            except ValueError:
                raise ValueError("immutable release conflict: " + key) from None
            finally:
                response.close()
        return True

    def put(self, key, data, immutable):
        content_type = "application/gzip" if key.endswith(".tar.gz") else ("text/x-shellscript" if key.endswith(".sh") else "application/json")
        headers = {"Content-Type": content_type,
                   "Cache-Control": "public, max-age=31536000, immutable" if immutable else "no-store"}
        if immutable:
            headers["x-oss-forbid-overwrite"] = "true"
        # Do not blindly retry ambiguous PUT failures; reruns recheck contents.
        with oss_stage("put_object", key, size=len(data), immutable=immutable):
            self.bucket.put_object(key, data, headers=headers)

    def verify_public(self, key, expected_sha, expected_size):
        with oss_stage("verify_public", key, expected_size=expected_size):
            verify_public(PUBLIC_URL + "/" + key, expected_sha, expected_size)


def publish(store, manifest, uploads):
    """Precheck every collision, then write only absent objects. Never promote."""
    validate_manifest(manifest, manifest["bundle_version"])
    store.validate_storage()
    components = {item["key"]: item for platform in manifest["platforms"].values()
                  for item in platform["components"].values()}
    if set(uploads) != set(components):
        raise ValueError("upload set does not match manifest")
    source_key, source_data = source_metadata(manifest)
    bundle_key = f"bundles/{manifest['bundle_version']}/manifest.json"
    manifest_data = canonical(manifest)
    installer_key = f"bundles/{manifest['bundle_version']}/install.sh"
    environment = manifest.get("release_environment")
    if environment not in ("qa", "prod"):
        raise ValueError("new manifests require release_environment")
    metadata = {source_key: source_data, bundle_key: manifest_data,
                installer_key: render_installer(manifest, channel_url(environment))}
    if (digest(metadata[installer_key]) != manifest["installer"]["sha256"]
            or len(metadata[installer_key]) != manifest["installer"]["size"]):
        raise ValueError("local installer changed after assembly")
    return publish_immutable_files(store, uploads, metadata, components, manifest_data)


def publish_immutable_files(store, uploads, metadata, components, manifest_data):
    """Shared collision preflight and readback for validated release/snapshot sets."""
    existing = {}
    for key, path in uploads.items():
        item = components[key]
        data = regular_bytes(path)
        if digest(data) != item["sha256"] or len(data) != item["size"]:
            raise ValueError("local artifact changed after validation")
        existing[key] = store.check_existing(key, item["sha256"], item["size"])
    for key, data in metadata.items():
        existing[key] = store.check_existing(key, digest(data), len(data))
    # Retain existence flags, not hundreds of MB of downloaded objects. Public
    # readback is the post-upload integrity check; no duplicate SDK GET.
    for key in [*uploads, *metadata]:
        data = regular_bytes(uploads[key]) if key in uploads else metadata[key]
        if key in components and (digest(data) != components[key]["sha256"] or len(data) != components[key]["size"]):
            raise ValueError("local artifact changed during publication")
        if not existing[key]:
            store.put(key, data, immutable=True)
        store.verify_public(key, digest(data), len(data))
        print("reused" if existing[key] else "uploaded", key)
    return digest(manifest_data)


def stable_version(value):
    """Return an ordering key for a component version admitted to a channel."""
    if not isinstance(value, str) or "-" in value:
        raise ValueError("prerelease components cannot be published")
    return tuple(int(part) for part in value.split("."))


def read_channel(store, key):
    # The legacy promote/publish-channel dry-run does not inspect the mutable
    # channel. Prod environment promotion explicitly reads its public channel.
    if store is None:
        return None
    return store.get(key)


def read_public_channel(key):
    """Read the current channel anonymously for an accurate Prod dry-run."""
    try:
        return download(PUBLIC_URL + "/" + key, 1024 * 1024)
    except HTTPError as error:
        if error.code == 404:
            return None
        raise


def previous_channel(previous_channel_bytes, sequence, environment):
    if previous_channel_bytes is None:
        return None
    name = channel_name(environment)
    try:
        previous = json.loads(previous_channel_bytes)
    except (TypeError, ValueError):
        raise ValueError("channel promotion is stale or the previous channel is invalid") from None
    bundle = previous.get("bundle_version") if isinstance(previous, dict) else None
    manifest_sha = previous.get("manifest_sha256") if isinstance(previous, dict) else None
    if (not isinstance(previous, dict) or previous.get("schema") != "moi-client-channel-v1"
            or previous.get("channel") != name or type(previous.get("sequence")) is not int
            or sequence <= previous["sequence"]
            or previous.get("manifest_url") != f"{PUBLIC_URL}/bundles/{bundle}/manifest.json"
            or not isinstance(manifest_sha, str) or len(manifest_sha) != 64):
        raise ValueError("channel promotion is stale or the previous channel is invalid")
    return load_published_manifest(bundle, manifest_sha)[0]


def validate_platform_coverage(manifest, previous_manifest):
    selected = set(manifest["platforms"])
    if previous_manifest is None:
        if selected != set(TARGETS):
            raise ValueError("first Prod release requires all four platforms")
        return
    missing = set(previous_manifest["platforms"]) - selected
    if missing:
        raise ValueError("stable promotion would drop platforms")


def validate_component_versions(manifest, previous_manifest):
    """Reject a lower component version on the channel the client already follows.

    QA compares only with the previous QA bundle. Prod compares only with the
    current stable bundle. Versions are uniform across platforms.
    """
    if previous_manifest is None:
        return
    previous_components = next(iter(previous_manifest["platforms"].values()))["components"]
    for platform in manifest["platforms"].values():
        for name, item in platform["components"].items():
            previous = previous_components[name]
            if stable_version(item["version"]) < stable_version(previous["version"]):
                raise ValueError(f"channel promotion would downgrade {name}")


def validate_stable(manifest, sequence, previous_manifest):
    if type(sequence) is not int or sequence <= 0:
        raise ValueError("promotion requires a positive sequence")
    if any(p["astra_source"]["prerelease"] for p in manifest["platforms"].values()):
        raise ValueError("stable cannot reference an Astra prerelease")
    if any("-" in c["version"] for p in manifest["platforms"].values() for c in p["components"].values()):
        raise ValueError("stable cannot reference prerelease components")
    validate_platform_coverage(manifest, previous_manifest)
    if previous_manifest is None:
        return
    if manifest["client_config"]["astra_url"] != previous_manifest["client_config"]["astra_url"]:
        raise ValueError("Prod Astra URL must match the current stable bundle")
    if manifest["compatibility_reference"] != previous_manifest["compatibility_reference"]:
        raise ValueError("Prod compatibility reference must match the current stable bundle")
    validate_component_versions(manifest, previous_manifest)


def assert_prod_manifest(manifest, allowlist):
    """Old promote may renew a Prod bundle. It cannot convert QA into Prod."""
    allowlist = require_client_url(allowlist)
    environment = manifest.get("release_environment")
    version = manifest["bundle_version"]
    if environment == "qa" or version.endswith("-qa"):
        raise ValueError("QA bundle cannot be promoted to stable")
    if version.endswith("-prod"):
        raise ValueError("Prod bundle version must not carry an environment suffix")
    url = manifest["client_config"]["astra_url"]
    if environment is None:
        if url != allowlist:
            raise ValueError("historical bundle environment cannot be determined")
        return
    if environment != "prod" or url != allowlist:
        raise ValueError("Prod bundle Astra URL is not on the Prod allowlist")


def verify_published_objects(store, manifest):
    source_key, source_data = source_metadata(manifest)
    if store is None:
        verify_public(PUBLIC_URL + "/" + source_key, digest(source_data), len(source_data))
    else:
        store.verify_public(source_key, digest(source_data), len(source_data))
    for platform in manifest["platforms"].values():
        for item in platform["components"].values():
            if store is None:
                verify_public(item["url"], item["sha256"], item["size"])
            else:
                store.verify_public(item["key"], item["sha256"], item["size"])
    installer = manifest["installer"]
    if store is None:
        verify_public(PUBLIC_URL + "/" + installer["key"], installer["sha256"], installer["size"])
    else:
        store.verify_public(installer["key"], installer["sha256"], installer["size"])


def write_channel(store, environment, bundle_version, manifest_sha, sequence, previous, now=None):
    current_key = channel_key(environment)
    assert_channel_unchanged(store, environment, previous)
    now = now or datetime.now(timezone.utc)
    channel = {"schema": "moi-client-channel-v1", "channel": channel_name(environment), "sequence": sequence,
               "bundle_version": bundle_version,
               "manifest_url": f"{PUBLIC_URL}/bundles/{bundle_version}/manifest.json",
               "manifest_sha256": manifest_sha,
               "issued_at": now.isoformat(), "expires_at": (now + timedelta(days=30)).isoformat()}
    data = canonical(channel)
    store.put(current_key, data, immutable=False)
    if store.get(current_key) != data:
        raise ValueError(channel_name(environment) + " readback mismatch; inspect publication before retrying")
    store.verify_public(current_key, digest(data), len(data))
    return data


def assert_channel_unchanged(store, environment, previous):
    if store.get(channel_key(environment)) != previous:
        raise ValueError(channel_name(environment) + " changed during publication; refusing promotion")


def write_stable_installer(store, manifest, data):
    """Publish the verified Prod bootstrap before moving the stable pointer."""
    installer = manifest["installer"]
    if digest(data) != installer["sha256"] or len(data) != installer["size"]:
        raise ValueError("published Prod installer does not match its manifest")
    if store.get(STABLE_INSTALLER_KEY) != data:
        store.put(STABLE_INSTALLER_KEY, data, immutable=False)
    store.verify_public(STABLE_INSTALLER_KEY, installer["sha256"], installer["size"])


def promote_existing(store, bundle_version, manifest_sha, sequence, prod_allowlist, now=None):
    """Read the exact published manifest; only stable may be written.

    store=None is a public read-only dry-run. No current GITHUB_SHA, local
    candidates, recompilation or manifest reserialization affects its digest.
    """
    if type(sequence) is not int or sequence <= 0:
        raise ValueError("promotion requires a positive sequence")
    manifest, manifest_data = load_published_manifest(bundle_version, manifest_sha)
    assert_prod_manifest(manifest, prod_allowlist)
    current_key = channel_key("prod")
    if store is not None:
        store.validate_storage()
    previous = read_channel(store, current_key)
    previous_manifest = previous_channel(previous, sequence, "prod")
    validate_stable(manifest, sequence, previous_manifest)
    verify_published_objects(store, manifest)
    if store is None:
        print("verified existing bundle (dry-run)", bundle_version, manifest_sha)
        return
    installer_data = store.get(manifest["installer"]["key"])
    if installer_data is None:
        raise ValueError("published Prod installer is missing")
    assert_channel_unchanged(store, "prod", previous)
    write_stable_installer(store, manifest, installer_data)
    write_channel(store, "prod", bundle_version, digest(manifest_data), sequence, previous, now)
    print("promoted", bundle_version, manifest_sha)


def publish_qa_channel(store, bundle_version, manifest_sha, sequence, now=None):
    """Point channels/qa.json at an already uploaded QA bundle. Never touch stable."""
    qa_bundle_version(bundle_version)
    if type(sequence) is not int or sequence <= 0:
        raise ValueError("promotion requires a positive sequence")
    manifest, manifest_data = load_published_manifest(bundle_version, manifest_sha)
    if manifest.get("release_environment") != "qa":
        raise ValueError("QA channel requires a QA manifest")
    current_key = channel_key("qa")
    if store is not None:
        store.validate_storage()
    previous = read_channel(store, current_key)
    validate_component_versions(manifest, previous_channel(previous, sequence, "qa"))
    verify_published_objects(store, manifest)
    if store is None:
        print("verified QA bundle channel (dry-run)", bundle_version, manifest_sha)
        return
    write_channel(store, "qa", bundle_version, digest(manifest_data), sequence, previous, now)
    print("published qa channel", bundle_version, manifest_sha)


def promote_environment(store, source_version, source_sha, astra_url, prod_allowlist, sequence, now=None):
    """Reuse a QA bundle's binaries and publish a Prod manifest. No rebuild, no component upload."""
    prod_version = prod_version_from_qa(source_version)
    astra_url = require_allowlisted_url(astra_url, prod_allowlist, "Prod Astra URL")
    if type(sequence) is not int or sequence <= 0:
        raise ValueError("promotion requires a positive sequence")
    source, source_data = load_published_manifest(source_version, source_sha)
    if source.get("release_environment") != "qa":
        raise ValueError("promote-environment requires a QA manifest")
    current_key = channel_key("prod")
    if store is not None:
        store.validate_storage()
    previous = read_public_channel(current_key) if store is None else read_channel(store, current_key)
    previous_manifest = previous_channel(previous, sequence, "prod")
    manifest = copy.deepcopy(source)
    manifest["bundle_version"] = prod_version
    manifest["release_environment"] = "prod"
    manifest["client_config"] = {"astra_url": astra_url}
    if previous_manifest is not None:
        # Keep the existing Prod installation contract while reusing QA binaries.
        manifest["compatibility_reference"] = previous_manifest["compatibility_reference"]
    installer_bytes = render_installer(manifest, channel_url("prod"))
    manifest["installer"] = {"key": f"bundles/{prod_version}/install.sh",
                             "sha256": digest(installer_bytes), "size": len(installer_bytes)}
    validate_manifest(manifest, prod_version)
    for target, platform in source["platforms"].items():
        for name, item in platform["components"].items():
            if manifest["platforms"][target]["components"][name] != item:
                raise ValueError("component identity changed during promotion")
        if manifest["platforms"][target]["astra_source"] != platform["astra_source"]:
            raise ValueError("Astra provenance changed during promotion")
    if manifest["moi_source_sha"] != source["moi_source_sha"]:
        raise ValueError("source identity changed during promotion")
    validate_stable(manifest, sequence, previous_manifest)
    # The Prod installer is rendered from this checkout. Reject when that template
    # no longer reproduces the installer QA already published and accepted.
    rendered_qa = render_installer(source, channel_url("qa"))
    if (digest(rendered_qa) != source["installer"]["sha256"]
            or len(rendered_qa) != source["installer"]["size"]):
        raise ValueError("installer template does not match the published QA installer")
    verify_published_objects(store, source)
    manifest_data = canonical(manifest)
    metadata = {f"bundles/{prod_version}/manifest.json": manifest_data,
                manifest["installer"]["key"]: installer_bytes}
    if store is None:
        print("verified QA bundle for Prod promotion (dry-run)", prod_version, digest(manifest_data))
        return digest(manifest_data)
    if store.get(f"bundles/{source_version}/manifest.json") != source_data:
        raise ValueError("QA manifest changed during Prod promotion")
    for key, data in metadata.items():
        if not store.check_existing(key, digest(data), len(data)):
            store.put(key, data, immutable=True)
        store.verify_public(key, digest(data), len(data))
    assert_channel_unchanged(store, "prod", previous)
    write_stable_installer(store, manifest, installer_bytes)
    write_channel(store, "prod", prod_version, digest(manifest_data), sequence, previous, now)
    print("promoted environment", source_version, prod_version, digest(manifest_data))
    return digest(manifest_data)


def main():
    logging.basicConfig(level=logging.INFO, format="%(message)s")
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    upload = commands.add_parser("upload")
    upload.add_argument("--root", type=Path, required=True)
    for name in ("promote", "publish-channel", "promote-environment"):
        command = commands.add_parser(name)
        command.add_argument("--bundle-version", required=True)
        command.add_argument("--manifest-sha256", required=True)
        command.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    def run_identity():
        run, attempt = int(os.environ["GITHUB_RUN_NUMBER"]), int(os.environ["GITHUB_RUN_ATTEMPT"])
        if run < 1 or not 1 <= attempt < 1000:
            raise ValueError("invalid GitHub run identity")
        return run * 1000 + attempt

    def write_summary(manifest, manifest_sha, channel=None):
        print("Published manifest SHA256:", manifest_sha)
        if os.environ.get("GITHUB_OUTPUT"):
            with open(os.environ["GITHUB_OUTPUT"], "a") as output:
                output.write(f"manifest_sha256={manifest_sha}\n")
        if os.environ.get("GITHUB_STEP_SUMMARY"):
            with open(os.environ["GITHUB_STEP_SUMMARY"], "a") as summary:
                summary.write(f"Environment: `{manifest.get('release_environment', 'historical')}`\n\n")
                summary.write(f"Astra URL: `{manifest['client_config']['astra_url']}`\n\n")
                summary.write(f"Bundle: `{manifest['bundle_version']}`\n\nManifest SHA256: `{manifest_sha}`\n")
                if channel:
                    summary.write(f"\nChannel: `{channel}`\n")
                else:
                    summary.write("\nChannel: not updated\n")
                summary.write("\nPlatforms: " + ", ".join(manifest["platforms"]) + "\n")
                summary.write(f"\nInstaller: {PUBLIC_URL}/bundles/{manifest['bundle_version']}/install.sh\n")

    if args.command == "upload":
        if os.environ.get("RELEASE_ENVIRONMENT") != "qa":
            raise ValueError("operation=build only publishes QA bundles")
        astra_url = require_allowlisted_url(os.environ.get("ASTRA_URL"), os.environ.get("MOI_RELEASE_ASTRA_URL_QA"), "QA Astra URL")
        manifest, uploads = assemble(args.root, os.environ["BUNDLE_VERSION"], os.environ["MOI_VERSION"],
                                     os.environ["ASTRA_VERSION"], os.environ.get("MOI_SOURCE_SHA", os.environ["GITHUB_SHA"]),
                                     os.environ["COMPATIBILITY_REFERENCE"], os.environ["ASTRA_MODE"],
                                     os.environ.get("ASTRA_SOURCE_SHA"), astra_url=astra_url,
                                     platforms=os.environ.get("RELEASE_PLATFORMS", "all"),
                                     release_environment="qa")
        manifest_sha = publish(OSSStore(), manifest, uploads)
        write_summary(manifest, manifest_sha)
    else:
        sequence = run_identity()
        store = None if args.dry_run else OSSStore()
        if args.command == "promote":
            promote_existing(store, args.bundle_version, args.manifest_sha256, sequence,
                             os.environ["MOI_RELEASE_ASTRA_URL_PROD"])
        elif args.command == "publish-channel":
            publish_qa_channel(store, args.bundle_version, args.manifest_sha256, sequence)
            if os.environ.get("GITHUB_STEP_SUMMARY"):
                with open(os.environ["GITHUB_STEP_SUMMARY"], "a") as summary:
                    if args.dry_run:
                        summary.write("Dry run: no OSS objects were written.\n\n")
                    summary.write(f"Channel: `{channel_key('qa')}`\n\nBundle: `{args.bundle_version}`\n")
        else:
            prod_sha = promote_environment(store, args.bundle_version, args.manifest_sha256,
                                           os.environ.get("ASTRA_URL"), os.environ["MOI_RELEASE_ASTRA_URL_PROD"], sequence)
            if prod_sha and os.environ.get("GITHUB_STEP_SUMMARY"):
                with open(os.environ["GITHUB_STEP_SUMMARY"], "a") as summary:
                    if args.dry_run:
                        summary.write("Dry run: no OSS objects were written.\n\n")
                    summary.write(f"Environment: `prod`\n\nAstra URL: `{os.environ.get('ASTRA_URL', '')}`\n\n")
                    summary.write(f"Bundle: `{prod_version_from_qa(args.bundle_version)}`\n\n")
                    summary.write(f"Manifest SHA256: `{prod_sha}`\n\nChannel: `{channel_key('prod')}`\n")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        if isinstance(error, (ValueError, KeyError, FileNotFoundError)):
            print("Release failed:", str(error))
        else:
            print("Release failed:", type(error).__name__, "(details suppressed; check release configuration)")
        raise SystemExit(1) from None
