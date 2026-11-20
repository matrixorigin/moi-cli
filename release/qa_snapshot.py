#!/usr/bin/env python3
"""Pinned QA snapshots. Never write release components, bundles, or channels."""
import argparse
import json
import os
from pathlib import Path
import re

from release_candidate import (
    COMPONENTS, PUBLIC_URL, RUNNERS, TARGETS, archive_members, build_source,
    canonical, digest, load_published_manifest, prepare, regular_bytes,
    require_client_url, require_manifest_sha, require_sha, require_source_version,
    resolve_moi, resolve_source_revision,
)
from snapshot_installer import render
from bundle import version_of


def targets(selection):
    if selection == "all":
        return TARGETS
    if selection in TARGETS:
        return (selection,)
    raise ValueError("select all or a supported native platform")


def snapshot_id(sha, run, attempt):
    require_sha(sha)
    for value in (run, attempt):
        if not re.fullmatch(r"[1-9][0-9]{0,19}", str(value)):
            raise ValueError("invalid snapshot run identity")
    return f"{sha[:12]}-{run}-{attempt}"


def prefix(identity):
    if not re.fullmatch(r"[0-9a-f]{12}-[1-9][0-9]{0,19}-[1-9][0-9]{0,19}", identity):
        raise ValueError("invalid snapshot identity")
    return "qa-snapshots/" + identity + "/"


def qa_url(env):
    # There is no arbitrary target URL or Prod option on the snapshot workflow.
    return require_client_url(env.get("MOI_RELEASE_ASTRA_URL_QA"))


def resolve(env):
    selected = targets(env.get("SNAPSHOT_PLATFORMS", "darwin-arm64"))
    qa_url(env)
    mode = env.get("MOI_MODE")
    if mode not in ("source", "reuse"):
        raise ValueError("moi_mode must be source or reuse")
    source = require_sha(env["GITHUB_SHA"])
    # Both modes pin MOI to this workflow's trusted source or an exact manifest.
    if mode == "reuse":
        require_manifest_sha(env.get("MOI_MANIFEST_SHA256", ""))
        moi_version, source = resolve_moi(env, selected)
    else:
        # Keep the source binary's own version; snapshot identity owns uniqueness.
        moi_version = ""
    sha, version = resolve_source_revision(env.get("ASTRA_REF", "main"), selected)
    identity = snapshot_id(sha, env["GITHUB_RUN_ID"], env["GITHUB_RUN_ATTEMPT"])
    return {
        "snapshot_id": identity, "snapshot_run_id": str(env["GITHUB_RUN_ID"]),
        "snapshot_attempt": str(env["GITHUB_RUN_ATTEMPT"]), "astra_sha": sha, "astra_version": version,
        "moi_version": moi_version, "moi_source_sha": source,
        "matrix": json.dumps({"include": [{"target": t, "runner": RUNNERS[t]} for t in selected]},
                             separators=(",", ":")),
    }


def assemble(root, env):
    sha = require_sha(env["ASTRA_SOURCE_SHA"])
    identity = env["SNAPSHOT_ID"]
    if identity != snapshot_id(sha, env["SNAPSHOT_RUN_ID"], env["SNAPSHOT_RUN_ATTEMPT"]):
        raise ValueError("snapshot does not match the pinned source/run")
    base = prefix(identity)
    selected = targets(env["SNAPSHOT_PLATFORMS"])
    versions = {"astra": require_source_version(env["ASTRA_VERSION"]),
                "astra-edge": require_source_version(env["ASTRA_VERSION"]),
                "moi-cli": require_source_version(env["MOI_VERSION"]) if env["MOI_MODE"] == "reuse" else None}
    moi_sha = require_sha(env["MOI_SOURCE_SHA"])
    if root.is_symlink() or {p.name for p in root.iterdir()} != set(selected):
        raise ValueError("snapshot candidate platforms differ from selection")
    platforms, uploads = {}, {}
    for target in selected:
        folder = root / target
        if folder.is_symlink() or not folder.is_dir():
            raise ValueError("snapshot platform must be a real directory")
        c = json.loads(regular_bytes(folder / "candidate.json", 65536))
        expected_source = {"mode": "source", "repository": "matrixorigin/Astra",
                           "commit_sha": sha, "prerelease": "-" in versions["astra"]}
        if (c.get("schema") != "moi-snapshot-candidate-v1" or c.get("platform") != target
                or c.get("moi_source_sha") != moi_sha or c.get("astra_source") != expected_source
                or set(c.get("components", {})) != set(COMPONENTS)):
            raise ValueError("snapshot candidate identity mismatch")
        reported_moi = require_source_version(c["components"]["moi-cli"]["version"])
        if versions["moi-cli"] is None:
            versions["moi-cli"] = reported_moi
        elif versions["moi-cli"] != reported_moi:
            raise ValueError("snapshot moi-cli versions differ across platforms")
        entries, allowed = {}, {"candidate.json"}
        for name in COMPONENTS:
            item = c["components"][name]
            filename = f"{name}-{versions[name]}-{target}.tar.gz"
            if item.get("filename") != filename or item.get("version") != versions[name]:
                raise ValueError("snapshot component identity mismatch")
            data = regular_bytes(folder / filename)
            if digest(data) != item.get("sha256") or len(data) != item.get("size"):
                raise ValueError("snapshot artifact digest/size mismatch")
            archive_members(data, (name,) if name == "moi-cli" else (name, "LICENSE"))
            key = base + target + "/" + name + ".tar.gz"
            entries[name] = {"version": versions[name], "key": key,
                             "url": PUBLIC_URL + "/" + key, "sha256": digest(data), "size": len(data)}
            uploads[key] = folder / filename
            allowed.add(filename)
        if {p.name for p in folder.iterdir()} != allowed:
            raise ValueError("unexpected snapshot candidate files")
        platforms[target] = {"components": entries}
    manifest = {
        "schema": "moi-client-snapshot-v1", "snapshot_id": identity,
        "astra_source_sha": sha, "moi_source_sha": moi_sha,
        "publisher_source_sha": require_sha(env["GITHUB_SHA"]),
        "run_id": str(env["SNAPSHOT_RUN_ID"]), "run_attempt": str(env["SNAPSHOT_RUN_ATTEMPT"]),
        "astra_url": qa_url(env), "platforms": platforms,
        "skill": {"owner": "moi-cli", "delivery": "embedded-export"},
    }
    installer = render(manifest)
    manifest["installer"] = {"key": base + "install.sh", "sha256": digest(installer), "size": len(installer)}
    return manifest, uploads, installer


class SnapshotStore:
    """Constrain even the shared publisher to this snapshot's object namespace."""
    def __init__(self, store, identity):
        self.store, self.prefix = store, prefix(identity)

    def validate_storage(self):
        self.store.validate_storage()

    def check_existing(self, key, sha, size):
        self.check_key(key)
        return self.store.check_existing(key, sha, size)

    def check_key(self, key):
        if (not key.startswith(self.prefix) or ".." in key or "//" in key
                or not re.fullmatch(r"[A-Za-z0-9./_-]+", key)):
            raise ValueError("snapshot write escaped its namespace")

    def put(self, key, data, immutable):
        self.check_key(key)
        if immutable is not True:
            raise ValueError("snapshot objects must be immutable")
        self.store.put(key, data, immutable=True)

    def verify_public(self, key, sha, size):
        self.check_key(key)
        self.store.verify_public(key, sha, size)


def publish(root, env, store):
    # Reassemble from data only: the credential-bearing publisher executes no binaries.
    from release_publish import publish_immutable_files
    manifest, uploads, installer = assemble(root, env)
    store = SnapshotStore(store, manifest["snapshot_id"])
    store.validate_storage()
    components = {c["key"]: c for p in manifest["platforms"].values() for c in p["components"].values()}
    data = canonical(manifest)
    metadata = {prefix(manifest["snapshot_id"]) + "manifest.json": data,
                manifest["installer"]["key"]: installer}
    return manifest, publish_immutable_files(store, uploads, metadata, components, data)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("resolve", "build", "prepare", "verify", "publish"))
    parser.add_argument("--astra-root", type=Path)
    parser.add_argument("--moi", type=Path)
    parser.add_argument("--root", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    env = os.environ
    if args.command == "resolve":
        values = resolve(env)
        with open(env["GITHUB_OUTPUT"], "a") as output:
            for key, value in values.items():
                output.write(f"{key}={value}\n")
    elif args.command == "build":
        build_source(args.astra_root.resolve(), env["ASTRA_VERSION"], env["ASTRA_SOURCE_SHA"], snapshot=True)
    elif args.command == "prepare":
        reused = (load_published_manifest(env["MOI_REF"], env["MOI_MANIFEST_SHA256"])[0]
                  if env["MOI_MODE"] == "reuse" else None)
        moi_version = (env["MOI_VERSION"] if reused else
                       version_of(args.moi.resolve(), "moi-cli version ").removeprefix("moi-cli version "))
        prepare(args.moi.resolve(), moi_version, env["ASTRA_VERSION"], env["MOI_SOURCE_SHA"],
                args.output, astra_root=args.astra_root.resolve(), astra_sha=env["ASTRA_SOURCE_SHA"],
                moi_reuse_manifest=reused, snapshot=True)
    elif args.command == "verify":
        manifest, _, _ = assemble(args.root, env)
        print("Verified QA snapshot", manifest["snapshot_id"], digest(canonical(manifest)))
    else:
        if env.get("GITHUB_REPOSITORY") != "matrixorigin/matrixflow" or env.get("GITHUB_REF") != "refs/heads/dev":
            raise ValueError("snapshot publication requires matrixorigin/matrixflow dev")
        from release_publish import OSSStore
        manifest, sha = publish(args.root, env, OSSStore())
        with open(env["GITHUB_STEP_SUMMARY"], "a") as output:
            output.write(f"QA snapshot: `{manifest['snapshot_id']}`\n\nManifest SHA256: `{sha}`\n\n")
            output.write(f"Installer: {PUBLIC_URL}/{manifest['installer']['key']}\n\n")
            output.write("Run the installer, then `astra-qa login` and `astra-qa`. No release channels were changed.\n")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        # Contract errors are deliberately value-free; transport/SDK exceptions may contain credentials.
        detail = str(error) if isinstance(error, ValueError) else "inspect the failed stage"
        print("QA snapshot failed:", type(error).__name__, detail)
        raise SystemExit(1) from None
