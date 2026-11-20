#!/usr/bin/env python3
"""Package a tested native Astra/moi-cli pair on its actual target platform.

This does not build, upload, sign, or declare an untested cross-platform release.
Use Astra's existing four-runner build matrix to provide its native binary.
"""
import argparse
import hashlib
import io
import os
import platform
import re
import stat
import subprocess
import tarfile
import tempfile
from pathlib import Path


def native_target():
    systems = {"Darwin": "darwin", "Linux": "linux"}
    machines = {"arm64": "arm64", "aarch64": "arm64", "x86_64": "amd64", "amd64": "amd64"}
    try:
        return systems[platform.system()] + "-" + machines[platform.machine()]
    except KeyError as exc:
        raise ValueError("only macOS/Linux arm64/amd64 are supported") from exc


def version_of(binary, prefix):
    if not stat.S_ISREG(binary.lstat().st_mode):
        raise ValueError("binary input must be a regular file, not a symlink")
    result = subprocess.run([str(binary.resolve()), "--version"], check=True, capture_output=True, text=True, timeout=15)
    version = result.stdout.strip()
    if not re.fullmatch(re.escape(prefix) + r"[0-9]+\.[0-9]+\.[0-9]+(?:[-+][A-Za-z0-9.-]+)?", version):
        raise ValueError("unexpected binary version response")
    return version


def bundle(astra, moi, license_path, version, destination):
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]{0,79}", version):
        raise ValueError("invalid bundle version")
    target = native_target()
    # Version/help probes alone are not sufficient: require the new wire owner
    # and both consumers to be present in this explicitly selected pair.
    astra_version = version_of(astra, "astra ")
    moi_version = version_of(moi, "moi-cli version ")
    for binary, args, marker in ((astra, ["auth", "credential", "--help"], "--output-fd"), (moi, ["-h"], "-auth"), (moi, ["skill", "--help"], "skill export")):
        result = subprocess.run([str(binary.resolve()), *args], capture_output=True, text=True, timeout=15)
        if result.returncode != 0 or marker not in result.stdout + result.stderr:
            raise ValueError("binary does not implement the native login bundle contract")
    manifest = f"moi-astra-native-v2\n{target}\n{version}\n{astra_version}\n{moi_version}\n".encode()
    destination.mkdir(parents=True, exist_ok=True)
    archive_path = destination / f"moi-astra-{version}-{target}.tar.gz"
    checksum = archive_path.with_suffix(archive_path.suffix + ".sha256")
    # Stage both outputs on the destination filesystem. Failed reads/writes leave
    # no versioned artifact, and hard links publish without replacing existing files.
    with tempfile.TemporaryDirectory(prefix=".bundle-", dir=destination) as staging:
        staged_archive = Path(staging) / archive_path.name
        staged_checksum = Path(staging) / checksum.name
        with staged_archive.open("xb") as output, tarfile.open(fileobj=output, mode="w:gz") as archive:
            for name, source, mode in (("astra", astra, 0o755), ("moi-cli", moi, 0o755), ("ASTRA-LICENSE", license_path, 0o644)):
                if not stat.S_ISREG(source.lstat().st_mode):
                    raise ValueError("archive members must be regular files")
                info = tarfile.TarInfo(name)
                info.size, info.mode = source.stat().st_size, mode
                with source.open("rb") as stream:
                    archive.addfile(info, stream)
            info = tarfile.TarInfo("bundle.txt")
            info.size, info.mode = len(manifest), 0o644
            archive.addfile(info, io.BytesIO(manifest))
        with staged_archive.open("rb") as stream:
            digest = hashlib.file_digest(stream, "sha256").hexdigest()
        staged_checksum.write_text(f"{digest}  {archive_path.name}\n")
        os.link(staged_archive, archive_path)
        try:
            os.link(staged_checksum, checksum)
        except BaseException:
            # Only the archive created by this attempt is ours to remove.
            archive_path.unlink()
            raise
    print(archive_path)
    print("SHA256 " + digest)
    return archive_path, digest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--astra", required=True, type=Path)
    parser.add_argument("--moi", required=True, type=Path)
    parser.add_argument("--astra-license", required=True, type=Path)
    parser.add_argument("--version", required=True)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    bundle(args.astra, args.moi, args.astra_license, args.version, args.output)


if __name__ == "__main__":
    main()
