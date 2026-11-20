"""Execute the generated production bootstrap with an isolated HOME.

Tar, checksums, shell setup and launchers are real. A test-only curl adapter
maps the synthetic HTTPS origin to an ephemeral loopback HTTP server. The tiny
updater is a test double; native behavior is checked by check_native_update.
"""
from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
import io
import os
import shlex
import shutil
import sys
from pathlib import Path
import subprocess
import tarfile
import tempfile
import threading
import unittest

from channel_installer import render
from release_candidate import digest


class BootstrapTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="moi-bootstrap-test.")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()
        self.home = self.root / "home"
        self.home.mkdir()
        self.tools = self.root / "tools"
        self.tools.mkdir()
        self.public = self.root / "public"
        self.public.mkdir()

        class Handler(SimpleHTTPRequestHandler):
            def log_message(self, *_):
                pass

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), partial(Handler, directory=str(self.public)))
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()
        self.addCleanup(self.stop_server)
        self.origin = "https://downloads.example.test"
        curl = shutil.which("curl")
        self.assertIsNotNone(curl)
        adapter = self.tools / "curl"
        adapter.write_text(
            "#!/bin/sh\nexec " + shlex.quote(sys.executable) + " -c " + shlex.quote(
                "import os, sys\n"
                "args = sys.argv[1:]\n"
                "assert args[-1] == 'https://downloads.example.test/cli.tar.gz'\n"
                "index = args.index('--proto') + 1\n"
                "assert args[index] == '=https'\n"
                "args[index] = '=http'\n"
                f"args[-1] = 'http://127.0.0.1:{self.server.server_port}/cli.tar.gz'\n"
                f"os.execv({curl!r}, [{curl!r}, '--noproxy', '*', *args])\n"
            ) + ' "$@"\n'
        )
        adapter.chmod(0o755)
        self.env = {**os.environ, "HOME": str(self.home), "SHELL": "/bin/sh",
                    "PATH": str(self.tools) + ":" + os.environ["PATH"], "MOI_AGENT_CALL": "1"}
        # Parent environment must not select the developer's real shell/state.
        for key in ("ZDOTDIR", "XDG_CONFIG_HOME", "ASTRA_LOCAL_STATE_ROOT", "MOI_AUTH_DIR",
                    "ASTRA_CLI_CREDENTIALS_DIR", "MOI_CLI_CONFIG_FILE", "MOI_ASTRA_BINARY", "ASTRA_API_URL",
                    "BASH_ENV", "ENV", "CDPATH", "CURL_HOME"):
            self.env.pop(key, None)

    def stop_server(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=5)

    def script(self, corrupt=None, platforms=("darwin-arm64", "darwin-amd64", "linux-arm64", "linux-amd64")):
        fake = b'''#!/bin/sh
set -eu
if [ "$1" = --moi-update-protocol ]; then echo moi-client-update-v1; exit; fi
[ "$1" = update ] && [ "$2" = install ] && [ "$3" = --dir ]
prefix=$4
mkdir -p "$prefix/current"
for binary in astra astra-edge moi-cli; do
  printf '#!/bin/sh\nprintf "%%s\\n" "$ASTRA_API_URL" "$PWD" "$@"\n' > "$prefix/current/$binary"
  chmod 755 "$prefix/current/$binary"
done
'''
        archive = io.BytesIO()
        with tarfile.open(fileobj=archive, mode="w:gz") as output:
            entry = tarfile.TarInfo("moi-cli")
            entry.size, entry.mode = len(fake), 0o755
            output.addfile(entry, io.BytesIO(fake))
        payload = archive.getvalue()
        (self.public / "cli.tar.gz").write_bytes(payload)
        item = {"url": self.origin + "/cli.tar.gz", "sha256": digest(payload), "size": len(payload)}
        if corrupt:
            item.update(corrupt)
        manifest = {"compatibility_reference": "server-contract-v1",
                    "client_config": {"astra_url": "https://astra.example.test"},
                    "platforms": {p: {"components": {"moi-cli": item}} for p in platforms}}
        script = self.root / "install.sh"
        script.write_bytes(render(manifest, self.origin + "/channels/stable.json"))
        return script

    def test_production_install_repeat_and_commands_from_any_directory(self):
        script = self.script()
        for system, machine in (("Darwin", "arm64"), ("Darwin", "x86_64"),
                                ("Linux", "aarch64"), ("Linux", "x86_64")):
            with self.subTest(system=system, machine=machine):
                uname = self.tools / "uname"
                uname.write_text(f'#!/bin/sh\ncase "$1" in -s) echo {system};; -m) echo {machine};; esac\n')
                uname.chmod(0o755)
                for _ in range(2):
                    result = subprocess.run(["sh", str(script)], env=self.env, capture_output=True, text=True, timeout=15)
                    self.assertEqual(0, result.returncode, result.stderr)
                prefix = self.home / ".local/share/moi/client"
                profile = self.home / ".profile"
                self.assertEqual(1, profile.read_text().count("# >>> MOI client >>>"))
                for name in ("astra", "astra-edge", "moi-cli"):
                    result = subprocess.run(["sh", "-c", '. "$HOME/.profile"; exec ' + name + ' "space argument"'],
                                            env=self.env, cwd=self.root, text=True, capture_output=True, timeout=10)
                    self.assertEqual(0, result.returncode, result.stderr)
                    self.assertEqual(["https://astra.example.test", str(self.root), "space argument"], result.stdout.splitlines())
                self.assertFalse((prefix / "client-auth").exists())
                self.assertFalse((self.home / ".moi").exists())

    def test_download_failure_does_not_touch_installation_or_profile(self):
        for corrupt in ({"sha256": "0" * 64}, {"size": 1}):
            script = self.script(corrupt)
            result = subprocess.run(["sh", str(script)], env=self.env, text=True, capture_output=True, timeout=15)
            self.assertNotEqual(0, result.returncode)
            self.assertFalse((self.home / ".local").exists())
            self.assertFalse((self.home / ".profile").exists())

    def test_mac_arm_only_installer_rejects_other_platforms_without_installing(self):
        script = self.script(platforms=("darwin-arm64",))
        uname = self.tools / "uname"
        for system, machine in (("Linux", "aarch64"), ("Darwin", "x86_64"), ("Darwin", "arm64")):
            with self.subTest(system=system, machine=machine):
                uname.write_text(f'#!/bin/sh\ncase "$1" in -s) echo {system};; -m) echo {machine};; esac\n')
                uname.chmod(0o755)
                result = subprocess.run(["sh", str(script)], env=self.env, capture_output=True, text=True, timeout=15)
                if (system, machine) == ("Darwin", "arm64"):
                    self.assertEqual(0, result.returncode, result.stderr)
                    self.assertTrue((self.home / ".local/share/moi/client/bin/astra").is_file())
                else:
                    self.assertNotEqual(0, result.returncode)
                    self.assertIn("no release for this platform", result.stderr)
                    self.assertFalse((self.home / ".local").exists())
                    self.assertFalse((self.home / ".profile").exists())

    def test_unknown_shell_installs_without_writing_invalid_rc(self):
        result = subprocess.run(["sh", str(self.script())], env={**self.env, "SHELL": "/bin/tcsh"},
                                text=True, capture_output=True, timeout=15)
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertFalse((self.home / ".profile").exists())
        self.assertIn("Unknown login shell", result.stderr)
        self.assertIn("Run now without changing PATH", result.stdout)
        self.assertTrue((self.home / ".local/share/moi/client/bin/astra").is_file())

    def test_existing_launcher_conflict_is_detected_before_updater_runs(self):
        prefix = self.home / ".local/share/moi/client"
        launcher = prefix / "bin/astra"
        launcher.parent.mkdir(parents=True)
        launcher.write_text("user-owned launcher\n")
        result = subprocess.run(["sh", str(self.script())], env=self.env,
                                text=True, capture_output=True, timeout=15)
        self.assertNotEqual(0, result.returncode)
        self.assertIn("CLI_INSTALL_LAUNCHER_CONFLICT", result.stderr)
        self.assertIn(str(launcher), result.stderr)
        self.assertIn("switching environments", result.stderr)
        self.assertEqual("user-owned launcher\n", launcher.read_text())
        self.assertFalse((prefix / "current").exists())
        self.assertFalse((self.home / ".profile").exists())


if __name__ == "__main__":
    unittest.main()
