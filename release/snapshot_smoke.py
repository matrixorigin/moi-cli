#!/usr/bin/env python3
"""Exercise the rendered snapshot installer and real native clients without login/network.

Only archive transport is replaced with copies of the already verified candidates.
The installer, archive checks, exported Skill and native executables are real.
"""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

from qa_snapshot import assemble
from release_candidate import native_target


def check(root, env):
    env = {**env, "SNAPSHOT_PLATFORMS": native_target()}
    manifest, uploads, installer = assemble(root, env)
    with tempfile.TemporaryDirectory(prefix="qa-snapshot-native-") as temporary:
        test = Path(temporary).resolve()
        home, stub, prefix = test / "home", test / "transport", test / "qa-client"
        home.mkdir()
        stub.mkdir()
        (home / ".moi").mkdir()
        sentinel = home / ".moi/auth.json"
        sentinel.write_text("isolated formal-auth sentinel; not a credential")
        formal_config = home / ".moi-cli.json"
        formal_bytes = b'{"endpoint":"https://prod.example.test/newmoi"}\n'
        formal_config.write_bytes(formal_bytes)
        fetch = stub / "curl"
        mapping = {item["url"]: str(uploads[item["key"]].resolve())
                   for platform in manifest["platforms"].values() for item in platform["components"].values()}
        fetch.write_text(f"#!{sys.executable}\n" + '''import json, pathlib, shutil, sys
mapping = json.loads(''' + repr(json.dumps(mapping)) + ''')
args = sys.argv[1:]
shutil.copyfile(mapping[args[-1]], args[args.index('-o') + 1])
sys.stdout.write('200')
''')
        fetch.chmod(0o755)
        script = test / "install.sh"
        script.write_bytes(installer)
        child_env = {**os.environ, "HOME": str(home), "PATH": str(stub) + ":" + os.environ["PATH"],
                     "MOI_CLI_CONFIG_FILE": str(formal_config),
                     "ASTRA_SERVER_URL": "https://prod-edge.example.test"}
        subprocess.run(["sh", str(script), "--dir", str(prefix), "--no-modify-path"],
                       env=child_env, check=True, timeout=120)
        destination = prefix / "snapshots" / manifest["snapshot_id"]
        # Use an empty CWD so project-local instructions/Skills cannot satisfy discovery.
        for name in ("astra", "astra-edge", "moi-cli"):
            launcher = prefix / "bin" / (name + "-qa")
            result = subprocess.run([str(launcher), "--version"], cwd=home, env=child_env,
                                    check=True, capture_output=True, text=True, timeout=60)
            expected = manifest["platforms"][native_target()]["components"][name]["version"]
            label = "moi-cli version " if name == "moi-cli" else name + " "
            if result.stdout.strip() != label + expected:
                raise ValueError("installed snapshot binary version mismatch")
        astra = str(prefix / "bin/astra-qa")
        subprocess.run([astra, "login", "--help"], cwd=home, env=child_env,
                       check=True, capture_output=True, timeout=60)
        # This is an offline configuration command, including on older reused CLIs.
        config_url = "https://qa-config.example.test/newmoi"
        subprocess.run([str(prefix / "bin/moi-cli-qa"), "-auth", "pat", "-api-key", "offline-test-placeholder",
                        "config", "set-endpoint", config_url], cwd=home, env=child_env,
                       check=True, capture_output=True, timeout=60)
        snapshot_config = destination / "moi-home/.moi-cli.json"
        if (formal_config.read_bytes() != formal_bytes
                or json.loads(snapshot_config.read_bytes()).get("endpoint") != config_url):
            raise ValueError("snapshot config write escaped its private configuration file")
        result = subprocess.run([astra, "skill", "list", "--source", "local"], cwd=home,
                                env=child_env, check=True, capture_output=True, text=True, timeout=60)
        skills = json.loads(result.stdout)["skills"]
        if not any(s.get("skill_name") == "moi-cli" and s.get("source") == "local"
                   and s.get("status") == "active" for s in skills):
            raise ValueError("snapshot Astra did not discover the paired Skill")
        if (destination / "auth").exists() or sentinel.read_text() != "isolated formal-auth sentinel; not a credential":
            raise ValueError("offline snapshot smoke changed authentication state")
        print("PASS: snapshot installer, real native launchers, isolated config write, login help and paired Skill discovery")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", required=True, type=Path)
    check(parser.parse_args().root, os.environ)
