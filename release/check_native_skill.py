#!/usr/bin/env python3
"""Verify paired installation with real native binaries and Astra's Skill loader.

Runs without login, a conversation or a model. All installation, credentials and
local state paths are temporary; nothing is installed into the user's profile.
"""
import argparse
import json
import os
import subprocess
import tempfile
from pathlib import Path

from bundle import bundle


def check(astra, moi, license_path):
    with tempfile.TemporaryDirectory(prefix="moi-native-skill-check-") as temporary:
        root = Path(temporary)
        archive, digest = bundle(astra, moi, license_path, "skill-check", root / "bundles")
        prefix, state = root / "client", root / "state"
        env = {
            **os.environ,
            "ASTRA_LOCAL_STATE_ROOT": str(state),
            "ASTRA_CLI_CREDENTIALS_DIR": str(root / "legacy-auth"),
            "MOI_AUTH_DIR": str(root / "native-auth"),
        }
        subprocess.run([
            "sh", str(Path(__file__).parent / "install.sh"),
            "--archive", str(archive), "--sha256", digest,
            "--dir", str(prefix), "--skill-dir", str(state / "skills"),
        ], check=True, env=env, timeout=120)
        installed = state / "skills/moi-cli/SKILL.md"
        if not installed.resolve().is_relative_to((state / "skills").resolve()):
            raise RuntimeError("Skill real path escapes Astra search root")
        result = subprocess.run([
            str(prefix / "current/astra"), "skill", "list", "--source", "local",
        ], cwd=root, env=env, check=True, capture_output=True, text=True, timeout=60)
        skills = json.loads(result.stdout)["skills"]
        if not any(s.get("skill_name") == "moi-cli" and s.get("source") == "local"
                   and s.get("status") == "active" for s in skills):
            raise RuntimeError("Astra did not discover active local moi-cli Skill")
        if (root / "legacy-auth").exists() or (root / "native-auth").exists():
            raise RuntimeError("offline installation/discovery created authentication state")
        print("PASS: real native binaries installed; Astra discovered active local moi-cli Skill")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--astra", required=True, type=Path)
    parser.add_argument("--moi", required=True, type=Path)
    parser.add_argument("--astra-license", required=True, type=Path)
    args = parser.parse_args()
    check(args.astra.resolve(), args.moi.resolve(), args.astra_license.resolve())


if __name__ == "__main__":
    main()
