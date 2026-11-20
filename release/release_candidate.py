#!/usr/bin/env python3
"""Build and validate per-component release candidates; never upload or sign.

The native pair's protocol/Skill probes remain owned by check_native_skill.py.
This owner adds deterministic component packaging and a complete-set manifest.
"""
import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import re
import stat
import subprocess
import sys
import tarfile
import tempfile
import tomllib
import urllib.parse

from bundle import native_target, version_of
from check_native_skill import check as check_native_skill
from channel_installer import render as render_installer, require_client_url
from release_transport import download, verify_public

TARGETS = ("darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64")
RUNNERS = {"darwin-amd64": "macos-15-intel", "darwin-arm64": "macos-15",
           "linux-amd64": "ubuntu-24.04", "linux-arm64": "ubuntu-24.04-arm"}
COMPONENTS = ("astra", "astra-edge", "moi-cli")
BUCKET = "astra-suite"
ENDPOINT = "https://oss-cn-hangzhou.aliyuncs.com"
PUBLIC_URL = "https://astra-suite.oss-cn-hangzhou.aliyuncs.com"
ASTRA_REPOSITORY = "matrixorigin/Astra"
MOI_REPOSITORY = "matrixorigin/matrixflow"
MAX_ARCHIVE = 512 * 1024 * 1024
RUST_TARGETS = {
    "linux-amd64": "x86_64-unknown-linux-musl",
    "linux-arm64": "aarch64-unknown-linux-musl",
    "darwin-amd64": "x86_64-apple-darwin",
    "darwin-arm64": "aarch64-apple-darwin",
}


def require_version(value):
    """MOI bundles only accept stable component versions.

    Astra's own release workflow may publish 0.1.5-rc1. That string is not a
    client component version: the updater contract is exactly x.y.z.
    """
    if not re.fullmatch(r"(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)", value or ""):
        raise ValueError("component version must be a stable x.y.z version without a prerelease suffix or leading v")
    return value


def require_bundle(value):
    if not isinstance(value, str) or not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}", value):
        raise ValueError("invalid bundle version")
    return value


def require_prod_bundle(value):
    value = require_bundle(value)
    if value.endswith("-qa") or value.endswith("-prod"):
        raise ValueError("Prod bundle version must not carry an environment suffix")
    return value


def qa_bundle_version(value):
    """Accept only a bundle id whose single -qa suffix is at the end."""
    value = require_bundle(value)
    if not value.endswith("-qa") or value.count("-qa") != 1:
        raise ValueError("QA bundle version must end with one -qa suffix")
    require_prod_bundle(value[:-3])
    return value


def prod_version_from_qa(value):
    qa_bundle_version(value)
    return value[:-3]


def channel_name(environment):
    if environment == "qa":
        return "qa"
    if environment == "prod":
        return "stable"
    raise ValueError("release environment must be qa or prod")


def channel_key(environment):
    return "channels/" + channel_name(environment) + ".json"


def channel_url(environment):
    return PUBLIC_URL + "/" + channel_key(environment)


def require_environment_bundle(environment, bundle_version):
    if environment == "qa":
        qa_bundle_version(bundle_version)
    elif environment == "prod":
        require_prod_bundle(bundle_version)
    else:
        raise ValueError("release environment must be qa or prod")
    return environment


def require_allowlisted_url(url, allowlist, label):
    actual = require_client_url(url)
    expected = require_client_url(allowlist)
    if actual != expected:
        raise ValueError(label + " must exactly match the environment allowlist")
    return actual


def require_sha(value):
    if not isinstance(value, str) or not re.fullmatch(r"[0-9a-f]{40}", value):
        raise ValueError("source must be a full commit SHA")
    return value


def require_compatibility(value):
    if not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9._:/@+-]{0,159}", value):
        raise ValueError("provide a non-secret compatibility contract/release reference")
    return value


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=True).encode() + b"\n"


def selected_targets(selection="all"):
    if selection == "all":
        return TARGETS
    if selection in ("darwin-arm64", "linux-arm64"):
        return (selection,)
    raise ValueError("platforms: select all, darwin-arm64 or linux-arm64")


def validate_inputs(env):
    """Report all input errors before building, downloading or loading credentials."""
    errors = []

    def check(key, label, validator, hint):
        value = env.get(key, "")
        if not value or not value.strip():
            errors.append(f"{label}: required. {hint}")
            return
        try:
            validator(value)
        except ValueError:
            # Do not echo supplied values: misplaced credentials must stay private.
            errors.append(f"{label}: invalid. {hint}")

    operation = env.get("RELEASE_OPERATION", "build")
    if operation not in ("build", "publish-channel", "promote", "promote-environment"):
        errors.append("operation: select build, publish-channel, promote, or promote-environment.")
    dry_run = env.get("RELEASE_DRY_RUN", "true")
    if dry_run not in ("true", "false"):
        errors.append("dry_run: select true or false.")
    if dry_run == "false" and (env.get("GITHUB_REPOSITORY") != "matrixorigin/matrixflow"
                                or env.get("GITHUB_REF") != "refs/heads/dev"):
        errors.append("Use workflow from: OSS publication requires matrixorigin/matrixflow dev.")
    mode = env.get("ASTRA_MODE", "")
    if operation == "build":
        if env.get("RELEASE_ENVIRONMENT") != "qa":
            errors.append("release_environment: operation=build only publishes QA.")
        try:
            selected_targets(env.get("RELEASE_PLATFORMS", "all"))
        except ValueError as error:
            errors.append(str(error))
        check("BUNDLE_VERSION", "bundle_version", qa_bundle_version,
              "Use a new QA bundle ID ending in -qa, such as 2026.09.22.1-qa.")
        moi_mode = env.get("MOI_MODE", "source")
        if moi_mode == "source":
            check("MOI_VERSION", "moi_version", require_version,
                  "Use a stable x.y.z version such as 0.1.0, without v or a prerelease suffix.")
            try:
                require_source_ref(env.get("MOI_SOURCE_REF", "dev"), "MOI")
            except ValueError:
                errors.append("moi_source_ref: invalid. Use a matrixorigin/matrixflow branch, tag, or commit.")
        elif moi_mode == "reuse":
            check("MOI_REF", "moi_ref", require_bundle, "Use the published bundle containing moi-cli.")
            check("MOI_MANIFEST_SHA256", "moi_manifest_sha256", require_manifest_sha,
                  "Copy the published bundle manifest SHA256.")
        else:
            errors.append("moi_mode: select source or reuse.")
        check("ASTRA_MODE", "astra_mode", require_source_mode, "Select source, release or reuse.")
        check("COMPATIBILITY_REFERENCE", "compatibility_reference", require_compatibility,
              "Use moi-native-v1 for the first series; keep the same value for compatible updates (no spaces).")
        check("MOI_RELEASE_ASTRA_URL_QA", "repository Variable MOI_RELEASE_ASTRA_URL_QA", require_client_url,
              "Set the QA Astra HTTPS origin in Actions Variables. Required even for dry-run.")
        if env.get("MOI_RELEASE_ASTRA_URL_QA", "").strip():
            check("ASTRA_URL", "astra_url", require_client_url,
                  "Enter the QA Astra HTTPS origin. It must exactly match MOI_RELEASE_ASTRA_URL_QA.")
            if not any(item.startswith("astra_url:") or item.startswith("repository Variable MOI_RELEASE_ASTRA_URL_QA:")
                       for item in errors):
                try:
                    require_allowlisted_url(env.get("ASTRA_URL"), env.get("MOI_RELEASE_ASTRA_URL_QA"), "QA Astra URL")
                except ValueError as error:
                    errors.append(str(error))
        check("GITHUB_SHA", "workflow commit", require_sha, "GitHub must provide a full MOI commit SHA.")
        if mode in ("source", "release"):
            check("ASTRA_VERSION", "astra_version", require_version,
                  "Use a stable x.y.z Astra version with no prerelease suffix or leading v.")
        if mode == "source" and not env.get("ASTRA_REF", "").strip():
            errors.append("astra_ref: required. Use a matrixorigin/Astra branch/tag/commit, such as main.")
        if mode == "reuse":
            check("ASTRA_REF", "astra_ref", require_bundle, "Use the existing OSS bundle ID to reuse, not an Astra branch.")
    elif operation == "publish-channel":
        check("BUNDLE_VERSION", "bundle_version", qa_bundle_version,
              "Use the uploaded QA bundle ID ending in one -qa suffix.")
    elif operation == "promote":
        check("BUNDLE_VERSION", "bundle_version", require_prod_bundle,
              "Use an uploaded Prod bundle ID with no -qa or -prod suffix.")
        check("MOI_RELEASE_ASTRA_URL_PROD", "repository Variable MOI_RELEASE_ASTRA_URL_PROD", require_client_url,
              "Set the Prod Astra HTTPS origin in Actions Variables.")
    elif operation == "promote-environment":
        check("BUNDLE_VERSION", "bundle_version", qa_bundle_version,
              "Use the QA bundle ID ending in one -qa suffix. Prod version is derived by removing it.")
        check("MOI_RELEASE_ASTRA_URL_PROD", "repository Variable MOI_RELEASE_ASTRA_URL_PROD", require_client_url,
              "Set the Prod Astra HTTPS origin in Actions Variables.")
        if env.get("MOI_RELEASE_ASTRA_URL_PROD", "").strip():
            check("ASTRA_URL", "astra_url", require_client_url,
                  "Enter the Prod Astra HTTPS origin. It must exactly match MOI_RELEASE_ASTRA_URL_PROD.")
            if not any(item.startswith("astra_url:") or item.startswith("repository Variable MOI_RELEASE_ASTRA_URL_PROD:")
                       for item in errors):
                try:
                    require_allowlisted_url(env.get("ASTRA_URL"), env.get("MOI_RELEASE_ASTRA_URL_PROD"), "Prod Astra URL")
                except ValueError as error:
                    errors.append(str(error))
    if operation in ("publish-channel", "promote", "promote-environment") or (operation == "build" and mode == "reuse"):
        check("MANIFEST_SHA256", "manifest_sha256", require_manifest_sha,
              "Copy the 64-character Manifest SHA256 from the original upload Summary; not a Git commit or binary digest.")
    if errors:
        raise ValueError("Release configuration errors:\n- " + "\n- ".join(errors))


def require_manifest_sha(value):
    if not re.fullmatch(r"[0-9a-f]{64}", value):
        raise ValueError("provide the published manifest SHA256 from the release summary")
    return value


def digest(data):
    return hashlib.sha256(data).hexdigest()


def regular_bytes(path, limit=MAX_ARCHIVE):
    info = path.lstat()
    if not stat.S_ISREG(info.st_mode) or not 0 < info.st_size <= limit:
        raise ValueError("release input must be a bounded regular file")
    return path.read_bytes()


def archive_members(data, expected):
    """Validate the entire archive before extraction; never use extractall."""
    with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as archive:
        result, total = {}, 0
        for member in archive:
            if member.name not in expected or member.name in result:
                raise ValueError("unexpected or duplicate archive members")
            total += member.size
            if total > MAX_ARCHIVE:
                raise ValueError("archive expands beyond release size limit")
            if not member.isfile() or not 0 < member.size <= MAX_ARCHIVE or member.mode & 0o7000:
                raise ValueError("archive member is not a safe regular file")
            result[member.name] = archive.extractfile(member).read()
        if set(result) != set(expected):
            raise ValueError("missing archive members")
        return result


def package_component(binary, license_data=None):
    entries = {binary.name: regular_bytes(binary)}
    if license_data is not None:
        entries["LICENSE"] = license_data
    output = io.BytesIO()
    with gzip.GzipFile(filename="", fileobj=output, mode="wb", mtime=0) as compressed:
        with tarfile.open(fileobj=compressed, mode="w", format=tarfile.USTAR_FORMAT) as archive:
            for name, data in sorted(entries.items()):
                member = tarfile.TarInfo(name)
                member.mode = 0o644 if name == "LICENSE" else 0o755
                member.size = len(data)
                archive.addfile(member, io.BytesIO(data))
    return output.getvalue()


def require_source_mode(mode):
    if mode not in ("release", "source", "reuse"):
        raise ValueError("Astra source mode must be release, source or reuse")
    return mode


def require_source_version(value):
    """A source build may retain its Cargo SemVer, including a prerelease."""
    if not isinstance(value, str) or not re.fullmatch(
            r"(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?", value):
        raise ValueError("invalid source version")
    return value


def require_source_ref(ref, component):
    if not isinstance(ref, str) or not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9._/-]{0,199}", ref) or ".." in ref:
        raise ValueError(f"provide a {component} branch, tag, or commit")
    return ref


def resolve_commit(ref, repository, component):
    require_source_ref(ref, component)
    encoded = urllib.parse.quote(ref, safe="")
    return require_sha(download(f"https://api.github.com/repos/{repository}/commits/{encoded}", 128,
                                github_api=True, accept="application/vnd.github.sha").decode().strip())


def resolve_source_revision(ref, targets=TARGETS):
    """Resolve source and its reported version once, before native fan-out."""
    sha = resolve_commit(ref, ASTRA_REPOSITORY, "Astra")
    manifest = tomllib.loads(download(
        f"https://raw.githubusercontent.com/{ASTRA_REPOSITORY}/{sha}/Cargo.toml", 1024 * 1024).decode())
    version = require_source_version(manifest["workspace"]["package"]["version"])
    # Fail before native runners start, not after four independent CLI builds.
    workflow = download(f"https://raw.githubusercontent.com/{ASTRA_REPOSITORY}/{sha}/.github/workflows/release-binaries.yml",
                        256 * 1024)
    for target in targets:
        parse_upstream_build_steps(workflow, target)
    return sha, version


def resolve_source(ref, version, targets=TARGETS):
    sha, reported = resolve_source_revision(ref, targets)
    if reported != require_version(version):
        raise ValueError("Astra source Cargo.toml version differs from astra_version; select matching source/version")
    return sha


def verify_checkout(root, sha, component):
    require_sha(sha)
    actual = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
    if actual != sha:
        raise ValueError(f"{component} checkout does not match resolved commit")
    dirty = subprocess.check_output(["git", "status", "--porcelain", "--untracked-files=no"], cwd=root, text=True)
    if dirty.strip():
        raise ValueError(f"{component} source checkout contains tracked changes")


def source_identity(root, version, sha, *, snapshot=False):
    (require_source_version if snapshot else require_version)(version)
    verify_checkout(root, sha, "Astra")
    manifest = tomllib.loads(regular_bytes(root / "Cargo.toml", 1024 * 1024).decode())
    if manifest["workspace"]["package"]["version"] != version:
        raise ValueError("Astra source version mismatch")
    for name in ("astra-cli", "astra-edge"):
        package = tomllib.loads(regular_bytes(root / "crates" / name / "Cargo.toml", 1024 * 1024).decode())
        if package["package"]["version"] != {"workspace": True}:
            raise ValueError("Astra client must inherit workspace version")
    return {"mode": "source", "repository": ASTRA_REPOSITORY,
            "commit_sha": sha, "prerelease": "-" in version}


def upstream_build_steps(root, target):
    return parse_upstream_build_steps(regular_bytes(root / ".github/workflows/release-binaries.yml", 256 * 1024), target)


def parse_upstream_build_steps(data, target):
    """Consume the pinned upstream client build owner, not a copied Cargo recipe.

    Only the current client build/test steps are supported. Unknown run steps
    or template expressions fail closed so new upstream gates aren't skipped.
    Packaging/upload remain MOI-owned and are deliberately not executed here.
    """
    import yaml
    workflow = yaml.safe_load(data)
    build = workflow["jobs"]["build"]
    matches = [entry for entry in build["strategy"]["matrix"]["include"] if entry.get("suffix") == target]
    if len(matches) != 1 or matches[0]["target"] != RUST_TARGETS[target]:
        raise ValueError("upstream native target contract changed")
    matrix = matches[0]
    names = {
        "Install pinned Rust target": None,
        "Install musl build dependencies": "${{ contains(matrix.target, 'unknown-linux-musl') }}",
        "Build client candidates": None,
        "Probe Edge TLS from the joint build": None,
        "Execute client candidates": None,
        "Verify Darwin session execution lease contract": "${{ contains(matrix.target, 'apple-darwin') }}",
    }
    selected, seen = [], set()
    for step in build["steps"]:
        if "run" not in step or step.get("name") == "Package client archive":
            continue
        name = step.get("name")
        if name not in names or name in seen or step.get("if") != names[name] or step.get("shell") != "bash":
            raise ValueError(f"upstream client build step contract changed: {name!r}; update the adapter contract")
        seen.add(name)
        if name == "Install musl build dependencies" and not target.startswith("linux-"):
            continue
        if name == "Verify Darwin session execution lease contract" and not target.startswith("darwin-"):
            continue
        def render(value):
            for key in ("target", "features"):
                value = value.replace("${{ matrix." + key + " }}", matrix[key])
            if "${{" in value:
                raise ValueError("unsupported upstream build expression")
            return value
        selected.append((render(step["run"]), {key: render(value) for key, value in step.get("env", {}).items()}))
    if seen != set(names):
        raise ValueError("upstream client build/test gate missing")
    return selected


def build_source(root, version, sha, *, snapshot=False):
    """Execute only the pinned upstream client's build + tests, never its release."""
    source_identity(root, version, sha, snapshot=snapshot)
    toolchain = tomllib.loads(regular_bytes(root / "rust-toolchain.toml", 65536).decode())
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", toolchain["toolchain"]["channel"]):
        raise ValueError("Astra source must pin a numeric Rust toolchain")
    env = {**os.environ, "RELEASE_VERSION": version, "RELEASE_SOURCE_SHA": sha}
    for script, step_env in upstream_build_steps(root, native_target()):
        subprocess.run(["bash", "-c", script], cwd=root, env={**env, **step_env}, check=True)
    source_identity(root, version, sha, snapshot=snapshot)


def fetch_astra(version, target, destination):
    require_version(version)
    if target not in TARGETS:
        raise ValueError("unsupported target")
    base = f"https://api.github.com/repos/{ASTRA_REPOSITORY}/releases/tags/v{version}"
    release = json.loads(download(base, 4 * 1024 * 1024, github_api=True))
    if release.get("draft") or release.get("tag_name") != f"v{version}":
        raise ValueError("Astra release is not a published exact tag")
    name = f"astra-v{version}-{target}.tar.gz"
    assets = [a for a in release.get("assets", []) if a.get("name") == name]
    if len(assets) != 1 or not re.fullmatch(r"sha256:[0-9a-f]{64}", assets[0].get("digest") or ""):
        raise ValueError("Astra asset missing or GitHub SHA256 digest unavailable")
    asset = assets[0]
    url = f"https://github.com/{ASTRA_REPOSITORY}/releases/download/v{version}/{name}"
    if asset.get("browser_download_url") != url or asset.get("state") != "uploaded":
        raise ValueError("unexpected Astra release asset origin/state")
    data = download(url, MAX_ARCHIVE, redirects=True)
    if len(data) != asset.get("size") or "sha256:" + digest(data) != asset["digest"]:
        raise ValueError("Astra release asset size or digest mismatch")
    members = archive_members(data, ("astra", "astra-edge", "LICENSE"))
    destination.mkdir()
    for name, content in members.items():
        path = destination / name
        path.write_bytes(content)
        path.chmod(0o644 if name == "LICENSE" else 0o755)
    return {"mode": "release", "repository": ASTRA_REPOSITORY, "tag": f"v{version}",
            "release_id": release["id"], "asset_id": asset["id"],
            "archive_sha256": digest(data), "prerelease": bool(release.get("prerelease"))}


def validate_provenance(upstream, version):
    if (not isinstance(upstream, dict) or upstream.get("repository") != ASTRA_REPOSITORY
            or type(upstream.get("prerelease")) is not bool):
        raise ValueError("invalid Astra provenance")
    if upstream.get("mode") == "source":
        require_sha(upstream.get("commit_sha"))
        if upstream["prerelease"] != ("-" in version):
            raise ValueError("invalid source prerelease marker")
        fields = ("mode", "repository", "commit_sha")
    elif upstream.get("mode") == "release":
        if (upstream.get("tag") != "v" + version
                or not re.fullmatch(r"[0-9a-f]{64}", upstream.get("archive_sha256", ""))
                or type(upstream.get("release_id")) is not int or upstream["release_id"] <= 0
                or type(upstream.get("asset_id")) is not int or upstream["asset_id"] <= 0):
            raise ValueError("invalid Astra release provenance")
        fields = ("mode", "repository", "tag", "release_id")
    else:
        raise ValueError("invalid Astra provenance mode")
    return {field: upstream[field] for field in fields}


def validate_manifest(manifest, bundle_version):
    """Validate untrusted published metadata before deriving any download/write key."""
    require_bundle(bundle_version)
    if (not isinstance(manifest, dict) or manifest.get("schema") != "moi-client-bundle-v1"
            or manifest.get("bundle_version") != bundle_version
            or manifest.get("skill") != {"owner": "moi-cli", "delivery": "embedded-export"}
            or not isinstance(manifest.get("platforms"), dict)
            or set(manifest["platforms"]) not in (set(TARGETS), {"darwin-arm64"}, {"linux-arm64"})):
        raise ValueError("invalid published bundle identity/platform set")
    require_sha(manifest.get("moi_source_sha"))
    require_compatibility(manifest.get("compatibility_reference", ""))
    require_client_url(manifest.get("client_config", {}).get("astra_url"))
    if "release_environment" in manifest:
        require_environment_bundle(manifest.get("release_environment"), bundle_version)
    installer = manifest.get("installer", {})
    if (installer.get("key") != f"bundles/{bundle_version}/install.sh"
            or not re.fullmatch(r"[0-9a-f]{64}", installer.get("sha256", ""))
            or type(installer.get("size")) is not int or not 0 < installer["size"] < 65536):
        raise ValueError("invalid bundle installer reference")
    shared = None
    for platform in manifest["platforms"].values():
        if not isinstance(platform, dict) or set(platform.get("components", {})) != set(COMPONENTS):
            raise ValueError("invalid published component set")
    for target, platform in manifest["platforms"].items():
        for name, item in platform["components"].items():
            version = require_version(item.get("version", ""))
            key = f"components/{name}/{version}/{target}.tar.gz"
            if (item.get("key") != key or item.get("url") != PUBLIC_URL + "/" + key
                    or not re.fullmatch(r"[0-9a-f]{64}", item.get("sha256", ""))
                    or type(item.get("size")) is not int or not 0 < item["size"] <= MAX_ARCHIVE):
                raise ValueError("invalid published component origin/digest/size")
        astra_version = platform["components"]["astra"]["version"]
        if platform["components"]["astra-edge"]["version"] != astra_version:
            raise ValueError("Astra/Edge version mismatch")
        if platform.get("astra_source", {}).get("prerelease") is not False:
            raise ValueError("prerelease components cannot be published")
        identity = validate_provenance(platform.get("astra_source"), astra_version)
        versions = tuple(platform["components"][name]["version"] for name in COMPONENTS)
        current = (versions, identity, platform["astra_source"]["prerelease"])
        if shared is not None and shared != current:
            raise ValueError("inconsistent published platform sources/versions")
        shared = current
    return manifest


def load_published_manifest(bundle_version, expected_sha):
    require_bundle(bundle_version)
    require_manifest_sha(expected_sha)
    data = download(f"{PUBLIC_URL}/bundles/{bundle_version}/manifest.json", 1024 * 1024)
    if digest(data) != expected_sha:
        raise ValueError("published manifest digest mismatch")
    return validate_manifest(json.loads(data), bundle_version), data


def source_metadata(manifest):
    platform = next(iter(manifest["platforms"].values()))
    version = platform["components"]["astra"]["version"]
    return (f"components/astra/{version}/source.json",
            canonical(validate_provenance(platform["astra_source"], version)))


def resolve_moi(env, targets):
    """Pin the binary version and its original source, not the publisher revision."""
    if env.get("MOI_MODE", "source") == "source":
        return (require_version(env["MOI_VERSION"]),
                resolve_commit(env.get("MOI_SOURCE_REF", "dev"), MOI_REPOSITORY, "MOI"))
    if env.get("MOI_MODE") != "reuse":
        raise ValueError("moi_mode must be source or reuse")
    manifest, _ = load_published_manifest(env["MOI_REF"], env["MOI_MANIFEST_SHA256"])
    if not set(targets).issubset(manifest["platforms"]):
        raise ValueError("moi reuse bundle does not contain all selected platforms")
    return (next(iter(manifest["platforms"].values()))["components"]["moi-cli"]["version"],
            manifest["moi_source_sha"])


def reused_component(manifest, target, name, root):
    item = manifest["platforms"][target]["components"][name]
    data = download(item["url"], item["size"])
    if digest(data) != item["sha256"] or len(data) != item["size"]:
        raise ValueError("reused component digest/size mismatch")
    members = archive_members(data, (name,) if name == "moi-cli" else (name, "LICENSE"))
    binary = root / name
    binary.write_bytes(members[name])
    binary.chmod(0o755)
    if "LICENSE" in members:
        (root / "LICENSE").write_bytes(members["LICENSE"])
    return binary, data


def reject_host_openssl(path):
    """Reject a Mach-O client that loads OpenSSL from outside its archive.

    Synthetic non-Mach-O stand-ins used by offline tests are ignored. A real
    Darwin binary must not depend on the build host's Homebrew OpenSSL.
    """
    if sys.platform != "darwin":
        return
    result = subprocess.run(["otool", "-L", os.fspath(path)], capture_output=True, text=True)
    if result.returncode != 0:
        if "not an object file" in (result.stderr or ""):
            return
        raise ValueError("otool could not inspect the macOS client")
    for line in result.stdout.splitlines()[1:]:
        dep = line.strip().split(" (", 1)[0]
        base = dep.rsplit("/", 1)[-1]
        if base.startswith(("libssl", "libcrypto")):
            raise ValueError("macOS client links OpenSSL outside the component archive")


def prepare(moi, moi_version, astra_version, source_sha, output, astra_root=None, astra_sha=None,
            release_root=None, reuse_manifest=None, moi_reuse_manifest=None, *, snapshot=False):
    if sum(value is not None for value in (astra_root, release_root, reuse_manifest)) != 1:
        raise ValueError("provide exactly one Astra source, release directory or reuse manifest")
    version_validator = require_source_version if snapshot else require_version
    version_validator(moi_version)
    version_validator(astra_version)
    if snapshot and astra_root is None:
        raise ValueError("snapshots require pinned Astra source")
    require_sha(source_sha)
    target = native_target()
    with tempfile.TemporaryDirectory(prefix="moi-release-probe-") as temp:
        root = Path(temp)
        reused = {}
        if moi_reuse_manifest is not None:
            validate_manifest(moi_reuse_manifest, moi_reuse_manifest["bundle_version"])
            if (moi_reuse_manifest["moi_source_sha"] != source_sha
                    or moi_reuse_manifest["platforms"][target]["components"]["moi-cli"]["version"] != moi_version):
                raise ValueError("reused moi-cli version/source mismatch")
            moi, reused["moi-cli"] = reused_component(moi_reuse_manifest, target, "moi-cli", root)
        if version_of(moi, "moi-cli version ") != "moi-cli version " + moi_version:
            raise ValueError("moi-cli binary version does not match release input")
        output.mkdir(parents=True, exist_ok=False)
        if reuse_manifest is not None:
            validate_manifest(reuse_manifest, reuse_manifest["bundle_version"])
            platform = reuse_manifest["platforms"][target]
            upstream = platform["astra_source"]
            binary_root = root / "reused"
            binary_root.mkdir()
            license_path = binary_root / "LICENSE"
            source_key, source_data = source_metadata(reuse_manifest)
            verify_public(PUBLIC_URL + "/" + source_key, digest(source_data), len(source_data))
            for name in ("astra", "astra-edge"):
                item = platform["components"][name]
                if item["version"] != astra_version:
                    raise ValueError("reused Astra version mismatch")
                _, reused[name] = reused_component(reuse_manifest, target, name, binary_root)
        elif astra_root is not None:
            upstream = source_identity(astra_root, astra_version, astra_sha, snapshot=snapshot)
            binary_root = astra_root / "target" / RUST_TARGETS[target] / "release"
            license_path = astra_root / "LICENSE"
        elif release_root is not None:
            # Fetch was a separate token-bearing step. Never pass that token to binaries.
            upstream = json.loads(regular_bytes(release_root / "provenance.json", 65536))
            validate_provenance(upstream, astra_version)
            if upstream["mode"] != "release":
                raise ValueError("expected release provenance")
            binary_root, license_path = release_root, release_root / "LICENSE"
        else:
            raise ValueError("provide exactly one Astra source, release directory or reuse manifest")
        binaries = {name: binary_root / name for name in ("astra", "astra-edge")}
        binaries["moi-cli"] = moi
        for name in ("astra", "astra-edge"):
            if version_of(binaries[name], name + " ") != name + " " + astra_version:
                raise ValueError("Astra binary version mismatch")
            subprocess.run([str(binaries[name]), "--help"], check=True, stdout=subprocess.DEVNULL, timeout=30)
            reject_host_openssl(binaries[name])
        check_native_skill(binaries["astra"], moi, license_path)
        for binary in binaries.values():
            result = subprocess.run([str(binary), "--moi-update-protocol"], check=True,
                                    stdout=subprocess.PIPE, text=True, timeout=30)
            if result.stdout.strip() != "moi-client-update-v1":
                raise ValueError("paired binary lacks MOI update occupancy protocol")
        files = {}
        for name, binary in binaries.items():
            version = moi_version if name == "moi-cli" else astra_version
            filename = f"{name}-{version}-{target}.tar.gz"
            license_data = None if name == "moi-cli" else regular_bytes(license_path)
            data = reused[name] if name in reused else package_component(binary, license_data)
            (output / filename).write_bytes(data)
            files[name] = {"version": version, "filename": filename,
                           "sha256": digest(data), "size": len(data)}
        (output / "candidate.json").write_bytes(canonical({
            "schema": "moi-snapshot-candidate-v1" if snapshot else "moi-component-candidate-v1", "platform": target,
            "moi_source_sha": source_sha, "astra_source": upstream, "components": files,
        }))


def assemble(root, bundle_version, moi_version, astra_version, source_sha, compatibility,
             astra_mode="release", astra_sha=None, *, astra_url=None, platforms="all",
             release_environment="prod"):
    """No binary execution here: publisher treats build artifacts as data only."""
    require_environment_bundle(release_environment, bundle_version)
    require_bundle(bundle_version)
    require_version(moi_version)
    require_version(astra_version)
    require_sha(source_sha)
    require_source_mode(astra_mode)
    if astra_mode == "source":
        require_sha(astra_sha or "")
    require_compatibility(compatibility)
    targets = selected_targets(platforms)
    if {p.name for p in root.iterdir()} != set(targets):
        raise ValueError("candidate set must contain exactly the selected platforms")
    platforms, uploads = {}, {}
    upstream_release = None
    for target in targets:
        folder = root / target
        if folder.is_symlink() or not folder.is_dir():
            raise ValueError("candidate platform must be a real directory")
        candidate = json.loads(regular_bytes(folder / "candidate.json", 65536))
        if (candidate.get("schema") != "moi-component-candidate-v1"
                or candidate.get("platform") != target or candidate.get("moi_source_sha") != source_sha
                or set(candidate.get("components", {})) != set(COMPONENTS)):
            raise ValueError("candidate identity mismatch")
        upstream = candidate["astra_source"]
        identity = validate_provenance(upstream, astra_version)
        if upstream["mode"] != astra_mode:
            raise ValueError("Astra provenance mode mismatch")
        if astra_mode == "source":
            if upstream.get("commit_sha") != astra_sha or upstream["prerelease"] != ("-" in astra_version):
                raise ValueError("Astra source commit/prerelease mismatch")
        else:
            identity = (identity, upstream["prerelease"])
            if upstream_release is not None and upstream_release != identity:
                raise ValueError("platforms came from different Astra releases")
            upstream_release = identity
        entries, allowed = {}, {"candidate.json"}
        for name in COMPONENTS:
            version = moi_version if name == "moi-cli" else astra_version
            item = candidate["components"][name]
            filename = f"{name}-{version}-{target}.tar.gz"
            if item.get("version") != version or item.get("filename") != filename:
                raise ValueError("component identity mismatch")
            data = regular_bytes(folder / filename)
            if digest(data) != item.get("sha256") or len(data) != item.get("size"):
                raise ValueError("component artifact digest/size mismatch")
            archive_members(data, (name,) if name == "moi-cli" else (name, "LICENSE"))
            key = f"components/{name}/{version}/{target}.tar.gz"
            uploads[key] = folder / filename
            entries[name] = {"version": version, "key": key, "url": PUBLIC_URL + "/" + key,
                             "sha256": digest(data), "size": len(data)}
            allowed.add(filename)
        if {p.name for p in folder.iterdir()} != allowed:
            raise ValueError("unexpected files in candidate directory")
        platforms[target] = {"components": entries, "astra_source": upstream}
    manifest = {"schema": "moi-client-bundle-v1", "bundle_version": bundle_version,
                "release_environment": release_environment,
                "moi_source_sha": source_sha, "compatibility_reference": compatibility,
                "platforms": platforms, "skill": {"owner": "moi-cli", "delivery": "embedded-export"},
                "client_config": {"astra_url": require_client_url(astra_url)}}
    installer = render_installer(manifest, channel_url(release_environment))
    manifest["installer"] = {"key": f"bundles/{bundle_version}/install.sh",
                             "sha256": digest(installer), "size": len(installer)}
    return validate_manifest(manifest, bundle_version), uploads


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("validate-inputs", "resolve-source", "verify-moi-source", "fetch-release", "build-source", "prepare", "assemble"))
    parser.add_argument("--moi", type=Path)
    parser.add_argument("--moi-root", type=Path)
    parser.add_argument("--root", type=Path)
    parser.add_argument("--output", type=Path)
    parser.add_argument("--astra-root", type=Path)
    parser.add_argument("--astra-release-root", type=Path)
    args = parser.parse_args()
    if args.command == "validate-inputs":
        try:
            validate_inputs(os.environ)
        except ValueError as error:
            parser.exit(1, f"{error}\n")
        print("Release inputs validated. No binaries built or OSS objects written.")
        return
    if args.command == "verify-moi-source":
        if args.moi_root is None:
            parser.error("--moi-root is required for source builds")
        verify_checkout(args.moi_root.resolve(), os.environ["MOI_SOURCE_SHA"], "MOI")
        print("MOI checkout matches the pinned source commit.")
        return
    moi_version = os.environ.get("MOI_VERSION", "")
    mode = require_source_mode(os.environ["ASTRA_MODE"])
    astra_version = os.environ.get("ASTRA_VERSION", "")
    if mode != "reuse" or args.command not in ("validate-inputs", "resolve-source"):
        require_version(astra_version)
    source_sha = require_sha(os.environ.get("MOI_SOURCE_SHA", os.environ["GITHUB_SHA"]))
    bundle_version = require_bundle(os.environ["BUNDLE_VERSION"])
    compatibility = require_compatibility(os.environ["COMPATIBILITY_REFERENCE"])
    environment = os.environ.get("RELEASE_ENVIRONMENT", "prod")
    astra_sha = os.environ.get("ASTRA_SOURCE_SHA")
    if args.command == "resolve-source":
        targets = selected_targets(os.environ.get("RELEASE_PLATFORMS", "all"))
        moi_version, source_sha = resolve_moi(os.environ, targets)
        origin = mode
        if mode == "reuse":
            manifest, _ = load_published_manifest(os.environ["ASTRA_REF"], os.environ["MANIFEST_SHA256"])
            if not set(targets).issubset(manifest["platforms"]):
                raise ValueError("reuse bundle does not contain all selected platforms")
            platform = next(iter(manifest["platforms"].values()))
            astra_version = platform["components"]["astra"]["version"]
            origin = platform["astra_source"]["mode"]
            sha = platform["astra_source"].get("commit_sha", "")
        else:
            sha = resolve_source(os.environ["ASTRA_REF"], astra_version, targets) if mode == "source" else ""
        matrix = {"include": [{"target": target, "runner": RUNNERS[target]} for target in targets]}
        with open(os.environ["GITHUB_OUTPUT"], "a") as output:
            output.write(f"astra_sha={sha}\nastra_version={astra_version}\norigin_mode={origin}\n")
            output.write(f"moi_version={moi_version}\nmoi_source_sha={source_sha}\n")
            output.write("matrix=" + json.dumps(matrix, separators=(",", ":")) + "\n")
            output.write("targets=" + " ".join(targets) + "\n")
        print("Astra:", mode, astra_version, sha)
        print("MOI:", os.environ.get("MOI_MODE", "source"), moi_version, source_sha)
    elif args.command == "fetch-release":
        if mode != "release":
            raise ValueError("fetch-release requires release mode")
        provenance = fetch_astra(astra_version, native_target(), args.output)
        (args.output / "provenance.json").write_bytes(canonical(provenance))
    elif args.command == "build-source":
        if mode != "source":
            raise ValueError("build-source requires source mode")
        build_source(args.astra_root.resolve(), astra_version, astra_sha)
    elif args.command == "prepare":
        if os.environ.get("MOI_MODE", "source") == "source":
            if args.moi_root is None:
                parser.error("--moi-root is required for source builds")
            verify_checkout(args.moi_root.resolve(), source_sha, "MOI")
        if ((mode == "source") != (args.astra_root is not None)
                or (mode == "release") != (args.astra_release_root is not None)):
            raise ValueError("provide the directory appropriate to the selected Astra mode")
        reused = load_published_manifest(os.environ["ASTRA_REF"], os.environ["MANIFEST_SHA256"])[0] if mode == "reuse" else None
        moi_reused = (load_published_manifest(os.environ["MOI_REF"], os.environ["MOI_MANIFEST_SHA256"])[0]
                      if os.environ.get("MOI_MODE", "source") == "reuse" else None)
        prepare(args.moi.resolve(), moi_version, astra_version, source_sha, args.output,
                args.astra_root.resolve() if args.astra_root else None, astra_sha,
                args.astra_release_root.resolve() if args.astra_release_root else None, reused, moi_reused)
    elif args.command == "assemble":
        if environment == "qa":
            astra_url = require_allowlisted_url(
                os.environ.get("ASTRA_URL"), os.environ.get("MOI_RELEASE_ASTRA_URL_QA"), "QA Astra URL")
        elif environment == "prod":
            astra_url = require_allowlisted_url(
                os.environ.get("ASTRA_URL"), os.environ.get("MOI_RELEASE_ASTRA_URL_PROD"), "Prod Astra URL")
        else:
            raise ValueError("release environment must be qa or prod")
        manifest, _ = assemble(args.root, bundle_version, moi_version, astra_version, source_sha,
                               compatibility, mode, astra_sha, astra_url=astra_url,
                               platforms=os.environ.get("RELEASE_PLATFORMS", "all"),
                               release_environment=environment)
        args.output.write_bytes(canonical(manifest))


if __name__ == "__main__":
    main()
