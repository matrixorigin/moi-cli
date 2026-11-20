#!/usr/bin/env python3
"""Real loopback updater install/update/rollback test; no login or external I/O."""
import argparse
from collections import Counter
from datetime import datetime, timedelta, timezone
import fcntl
from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import threading

from bundle import native_target
from release_candidate import package_component, canonical, digest


def check(astra, edge, moi_before, moi_after, license_path):
    with tempfile.TemporaryDirectory(prefix="moi-native-update-") as temporary:
        # macOS /var is an alias of /private/var; compare canonical paths when
        # proving destructive test fixtures remain inside this temporary root.
        root = Path(temporary).resolve()
        public = root / "public"
        public.mkdir()
        requests = Counter()

        class Handler(SimpleHTTPRequestHandler):
            def do_GET(self):
                requests[self.path] += 1
                super().do_GET()
            def log_message(self, *_):
                pass

        server = ThreadingHTTPServer(("127.0.0.1", 0), partial(Handler, directory=str(public)))
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            base = f"http://127.0.0.1:{server.server_port}"
            channel_url = base + "/channels/stable.json"
            prefix, state, auth = root / "installed", root / "astra-state", root / "auth"
            auth.mkdir(mode=0o700)
            sentinel = auth / "preservation-sentinel"
            sentinel.write_text("synthetic existing login data: must not be read or changed")
            env = {**os.environ, "ASTRA_LOCAL_STATE_ROOT": str(state), "MOI_AUTH_DIR": str(auth),
                   "ASTRA_CLI_CREDENTIALS_DIR": str(root / "legacy"), "MOI_AGENT_CALL": "1",
                   "HOME": str(root / "home"), "SHELL": "/bin/sh"}
            (root / "home").mkdir()
            platform = native_target()
            cached = {}

            def run(args, expected=0, timeout=300):
                result = subprocess.run([str(a) for a in args], env=env, cwd=root,
                                        capture_output=True, text=True, timeout=timeout)
                if result.returncode != expected:
                    raise RuntimeError(f"{Path(str(args[0])).name} {args[1:]}: exit {result.returncode}: {result.stderr[-1000:]}")
                return result

            def publish(bundle, moi, sequence, compatibility="native-update-acceptance-v1"):
                components = {}
                for name, path in (("astra", astra), ("astra-edge", edge), ("moi-cli", moi)):
                    version = run([path, "--version"]).stdout.strip().split()[-1]
                    key = f"components/{name}/{version}/{platform}.tar.gz"
                    if key not in cached:
                        data = package_component(path, None if name == "moi-cli" else license_path.read_bytes())
                        cached[key] = (digest(data), len(data))
                        destination = public / key
                        destination.parent.mkdir(parents=True, exist_ok=True)
                        destination.write_bytes(data)
                    sha, size = cached[key]
                    components[name] = {"version": version, "key": key, "url": base+"/"+key,
                                        "sha256": sha, "size": size}
                manifest = {"schema": "moi-client-bundle-v1", "bundle_version": bundle,
                            "compatibility_reference": compatibility,
                            "platforms": {platform: {"components": components}},
                            "skill": {"owner": "moi-cli", "delivery": "embedded-export"},
                            "client_config": {"astra_url": "https://astra.example.test"}}
                data = canonical(manifest)
                path = public / f"bundles/{bundle}/manifest.json"
                path.parent.mkdir(parents=True)
                path.write_bytes(data)
                now = datetime.now(timezone.utc)
                channel = {"schema": "moi-client-channel-v1", "bundle_version": bundle,
                           "manifest_url": base+f"/bundles/{bundle}/manifest.json", "manifest_sha256": digest(data),
                           "sequence": sequence, "issued_at": now.isoformat(), "expires_at": (now+timedelta(days=1)).isoformat()}
                channel_file = public / "channels/stable.json"
                channel_file.parent.mkdir(exist_ok=True)
                channel_file.write_bytes(canonical(channel))
                return manifest

            def install():
                # The public shell bootstrap requires HTTPS. This fixture owns
                # only an explicit loopback updater installation; bootstrap_test
                # separately exercises the generated HTTPS shell entrypoint.
                run([moi_before, "update", "install", "--dir", prefix,
                     "--skill-dir", state / "skills", "--channel", channel_url,
                     "--compatibility", "native-update-acceptance-v1",
                     "--astra-url", "https://astra.example.test", "--allow-loopback-http"])

            first = publish("native-update-1", moi_before, 1)
            install()
            installed_astra = prefix / "current/astra"
            print("PASS real loopback updater installation and three native binary probes", flush=True)
            run([installed_astra, "update", "check"])
            skills = json.loads(run([installed_astra, "skill", "list", "--source", "local"], timeout=90).stdout)
            if not any(s.get("skill_name")=="moi-cli" and s.get("status")=="active" for s in skills["skills"]):
                raise RuntimeError("installed Skill not active")
            auth_snapshot = {p.name: p.read_bytes() for p in auth.iterdir() if p.is_file()}
            print("PASS real Astra local Skill discovery", flush=True)
            before_requests = requests.copy()
            publish("native-update-2", moi_after, 2)
            result = json.loads(run([installed_astra, "update", "check"]).stdout)
            if result["changed"] != ["moi-cli"]:
                raise RuntimeError("wrong component diff")
            with (prefix / "runtime.lock").open("r+") as lease:
                fcntl.flock(lease, fcntl.LOCK_SH | fcntl.LOCK_NB)
                result = run([installed_astra, "update"], expected=1)
                if "CLI_UPDATE_BUSY" not in result.stderr:
                    raise RuntimeError("busy install was not rejected")
                fcntl.flock(lease, fcntl.LOCK_EX | fcntl.LOCK_NB)
                for binary in ("astra", "astra-edge", "moi-cli"):
                    result = run([prefix/"current"/binary, "--version"], expected=1)
                    if "CLI_UPDATE_BUSY" not in result.stderr:
                        raise RuntimeError(f"{binary} does not share the updater lock")
            old_astra = installed_astra.resolve()
            run([installed_astra, "update"])
            for name in ("astra", "astra-edge"):
                key = "/" + first["platforms"][platform]["components"][name]["key"]
                if requests[key] != before_requests[key]:
                    raise RuntimeError("unchanged component downloaded again")
            result = run([old_astra, "--version"], expected=1)
            if "RESTART_REQUIRED" not in result.stderr:
                raise RuntimeError("old executable admitted")
            print("PASS Astra delegates update; unchanged Astra/Edge reused; busy and old-process checks", flush=True)
            run([installed_astra, "update", "--rollback"])
            expected = run([moi_before,"--version"]).stdout
            if run([prefix/"current/moi-cli","--version"]).stdout != expected:
                raise RuntimeError("rollback version mismatch")
            # Destructive fixtures are confined to this TemporaryDirectory.
            # Return to the channel tip to exercise no-change local repair.
            run([installed_astra, "update"])
            skill_link = state / "skills/moi-cli"
            skill_copy = skill_link.resolve(strict=True)
            if not skill_copy.is_relative_to(root):
                raise RuntimeError("test Skill path escaped temporary root")
            shutil.rmtree(skill_copy)
            run([installed_astra, "update"])
            if not (skill_copy / "SKILL.md").is_file():
                raise RuntimeError("missing current Skill copy was not restored")
            print("PASS real updater restores missing derived Skill copy", flush=True)
            shutil.rmtree(prefix)
            install()
            run([installed_astra, "update", "check"])
            if not (skill_link / "SKILL.md").is_file():
                raise RuntimeError("same-prefix reinstall did not restore managed Skill")
            print("PASS same-prefix reinstall with residual global Skill ownership", flush=True)
            selected = installed_astra.resolve()
            publish("native-update-3", moi_after, 3, compatibility="native-update-acceptance-v2")
            result = json.loads(run([installed_astra, "update", "check"]).stdout)
            if result.get("status") != "reinstall_required" or result.get("reason") != "CLI_UPDATE_COMPATIBILITY_CHANGE":
                raise RuntimeError("cross-contract release was not reported")
            result = run([installed_astra, "update"], expected=1)
            if "separate installation" not in result.stderr or installed_astra.resolve() != selected:
                raise RuntimeError("cross-contract release lacked guidance or changed current")
            print("PASS incompatible release is visible but cannot activate", flush=True)
            if sentinel.read_text() != "synthetic existing login data: must not be read or changed":
                raise RuntimeError("authentication data changed")
            if {p.name: p.read_bytes() for p in auth.iterdir() if p.is_file()} != auth_snapshot or (root/"legacy").exists():
                raise RuntimeError("updater touched login state")
            print("PASS rollback and authentication/configuration isolation", flush=True)
        finally:
            server.shutdown()
            server.server_close()
            thread.join(timeout=5)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    for flag in ("astra", "edge", "moi-before", "moi-after", "astra-license"):
        parser.add_argument("--"+flag, required=True, type=Path)
    args = parser.parse_args()
    check(args.astra.resolve(), args.edge.resolve(), args.moi_before.resolve(),
          args.moi_after.resolve(), args.astra_license.resolve())
