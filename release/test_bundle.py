import contextlib
import hashlib
import importlib.util
import io
import os
import subprocess
import tarfile
import tempfile
import threading
import unittest
from unittest import mock
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("bundle", HERE / "bundle.py")
bundle = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bundle)
@contextlib.contextmanager
def download_server(handler):
    with ThreadingHTTPServer(("127.0.0.1", 0), handler) as server:
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            yield server.server_port
        finally:
            server.shutdown()
            thread.join(timeout=5)


class BundleTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="moi-pair-test-")
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.prefix = self.root / "installation with spaces 中文"
        self.astra = self.root / "astra"
        self.moi = self.root / "moi-cli"
        self.license = self.root / "LICENSE"
        self.license.write_text("synthetic fixture license\n")
        self.make_binaries("1.0.0")

    def make_binaries(self, version):
        self.astra.write_text(f'#!/bin/sh\nif [ "$1" = --version ]; then echo "astra {version}"; else echo --output-fd; fi\n')
        self.moi.write_text(f'''#!/bin/sh
if [ "$1" = --version ]; then echo "moi-cli version {version}"
elif [ "$1" = skill ] && [ "$2" = export ]; then
    mkdir -p "$4" && printf '%s\\n' '---' 'name: moi-cli' 'description: Synthetic test skill' '---' 'version {version}' > "$4/SKILL.md"
elif [ "$1" = skill ]; then echo 'skill export --dir DIRECTORY'
else echo -auth; fi
''')
        self.astra.chmod(0o755)
        self.moi.chmod(0o755)

    def package(self, version):
        with contextlib.redirect_stdout(io.StringIO()):
            return bundle.bundle(self.astra, self.moi, self.license, version, self.root / "artifacts")

    def install(self, archive, digest, ok=True, env=None):
        result = subprocess.run(["sh", str(HERE / "install.sh"), "--archive", str(archive), "--sha256", digest,
            "--dir", str(self.prefix), "--preview-state"], capture_output=True, text=True, timeout=30, env=env)
        self.assertEqual(result.returncode == 0, ok, result.stderr + result.stdout)
        return result

    def test_atomic_pair_upgrade_idempotence_and_retained_old_release(self):
        first, sha1 = self.package("first")
        self.install(first, sha1)
        old = (self.prefix / "current").resolve()
        self.install(first, sha1)
        self.assertEqual((self.prefix / "current").resolve(), old)
        self.make_binaries("1.1.0")
        second, sha2 = self.package("second")
        self.install(second, sha2)
        self.assertNotEqual((self.prefix / "current").resolve(), old)
        self.assertIn("1.0.0", subprocess.check_output([str(old / "astra"), "--version"], text=True))
        for name in ("astra", "moi-cli"):
            self.assertIn("1.1.0", subprocess.check_output([str(self.prefix / "current" / name), "--version"], text=True))
        skill = self.prefix / "astra-state/skills/moi-cli"
        self.assertIn("1.1.0", (skill / "SKILL.md").read_text())
        self.assertTrue(skill.resolve().is_relative_to((self.prefix / "astra-state/skills").resolve()))
        self.assertIn("1.0.0", (old / "skills/moi-cli/SKILL.md").read_text())

    def test_user_skill_is_never_overwritten(self):
        skill = self.prefix / "astra-state/skills/moi-cli"
        skill.mkdir(parents=True)
        (skill / "SKILL.md").write_text("user-owned skill")
        archive, digest = self.package("user-conflict")
        self.install(archive, digest, ok=False)
        self.assertEqual((skill / "SKILL.md").read_text(), "user-owned skill")
        self.assertFalse((self.prefix / "current").exists())

    def test_modified_managed_skill_is_not_silently_replaced(self):
        archive, digest = self.package("managed-change")
        self.install(archive, digest)
        skill = self.prefix / "astra-state/skills/moi-cli/SKILL.md"
        skill.write_text("user modified managed content")
        self.install(archive, digest, ok=False)
        self.assertEqual(skill.read_text(), "user modified managed content")

    def test_binary_activation_failure_restores_previous_skill(self):
        archive, digest = self.package("before")
        self.install(archive, digest)
        old_pair = (self.prefix / "current").resolve()
        skill = self.prefix / "astra-state/skills/moi-cli"
        old_skill = skill.resolve()
        self.make_binaries("1.2.0")
        archive, digest = self.package("after")
        wrappers = self.root / "wrapper"
        wrappers.mkdir()
        failing_mv = wrappers / "mv"
        failing_mv.write_text('#!/bin/sh\nfor dest do :; done\ncase "$dest" in */current) exit 17;; esac\nexec /bin/mv "$@"\n')
        failing_mv.chmod(0o755)
        self.install(archive, digest, ok=False, env={**os.environ, "PATH": str(wrappers) + ":" + os.environ["PATH"]})
        self.assertEqual((self.prefix / "current").resolve(), old_pair)
        self.assertEqual(skill.resolve(), old_skill)
        self.assertIn("1.0.0", (skill / "SKILL.md").read_text())
        self.install(archive, digest)
        self.assertIn("1.2.0", (skill / "SKILL.md").read_text())

    def test_explicit_global_skill_directory_preserves_other_skills(self):
        archive, digest = self.package("global")
        skills = self.root / "user skills"
        other = skills / "other/SKILL.md"
        other.parent.mkdir(parents=True)
        other.write_text("other user skill")
        args = ["sh", str(HERE / "install.sh"), "--archive", str(archive), "--sha256", digest,
                "--dir", str(self.prefix), "--skill-dir", str(skills)]
        result = subprocess.run(args, capture_output=True, text=True, timeout=30)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(other.read_text(), "other user skill")
        self.assertTrue((skills / "moi-cli/SKILL.md").resolve().is_relative_to(skills.resolve()))
        # A second install prefix cannot take over the first installation's slot.
        args[args.index("--dir") + 1] = str(self.root / "different install")
        result = subprocess.run(args, capture_output=True, text=True, timeout=30)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("different installation", result.stderr)

    def test_astra_state_root_controls_default_global_skill_location(self):
        archive, digest = self.package("state-root")
        state = self.root / "isolated astra state"
        args = ["sh", str(HERE / "install.sh"), "--archive", str(archive), "--sha256", digest,
                "--dir", str(self.prefix)]
        result = subprocess.run(args, env={**os.environ, "ASTRA_LOCAL_STATE_ROOT": str(state)}, capture_output=True, text=True, timeout=30)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue((state / "skills/moi-cli/SKILL.md").is_file())

    def test_archive_validation_is_independent_of_customer_locale(self):
        archive, digest = self.package("locale-check")
        for locale in ("C", "en_US.UTF-8", "zh_CN.UTF-8"):
            with self.subTest(locale=locale):
                self.prefix = self.root / ("installation-" + locale)
                self.install(archive, digest, env={**os.environ, "LANG": locale, "LC_ALL": locale})
                self.assertTrue((self.prefix / "current/astra").is_file())
                self.assertTrue((self.prefix / "current/moi-cli").is_file())

    def test_hash_failure_and_activation_failure_leave_old_pair(self):
        first, sha1 = self.package("first")
        self.install(first, sha1)
        old = (self.prefix / "current").resolve()
        self.make_binaries("1.2.0")
        second, sha2 = self.package("second")
        self.install(second, "0" * 64, ok=False)
        self.assertEqual((self.prefix / "current").resolve(), old)
        wrappers = self.root / "wrappers"
        wrappers.mkdir()
        failing_mv = wrappers / "mv"
        failing_mv.write_text('#!/bin/sh\ncase "$1" in -fh|-fT) exit 17;; esac\nexec /bin/mv "$@"\n')
        failing_mv.chmod(0o755)
        self.install(second, sha2, ok=False, env={**os.environ, "PATH": str(wrappers) + ":" + os.environ["PATH"]})
        self.assertEqual((self.prefix / "current").resolve(), old)
        self.assertFalse((self.prefix / ".install-lock").exists())
        # Retry recovers the complete staged release without replacing its files.
        self.install(second, sha2)

    def test_rejects_traversal_members_before_extracting(self):
        archive = self.root / "malicious.tar.gz"
        with tarfile.open(archive, "w:gz") as output:
            info = tarfile.TarInfo("../../escaped")
            info.size = 4
            output.addfile(info, io.BytesIO(b"oops"))
        sha = hashlib.sha256(archive.read_bytes()).hexdigest()
        self.install(archive, sha, ok=False)
        self.assertFalse((self.root / "escaped").exists())
        self.assertFalse((self.prefix / "current").exists())

    def test_rejects_existing_unowned_current(self):
        archive, sha = self.package("first")
        self.prefix.mkdir()
        (self.prefix / "current").write_text("user-owned data")
        self.install(archive, sha, ok=False)
        self.assertEqual((self.prefix / "current").read_text(), "user-owned data")

    def test_packager_refuses_symlink_and_overwrite(self):
        archive, _ = self.package("first")
        original = archive.read_bytes()
        checksum = archive.with_suffix(archive.suffix + ".sha256")
        original_checksum = checksum.read_bytes()
        with self.assertRaises(FileExistsError):
            self.package("first")
        self.assertEqual(archive.read_bytes(), original)
        self.assertEqual(checksum.read_bytes(), original_checksum)
        link = self.root / "astra-link"
        link.symlink_to(self.astra)
        with self.assertRaises(ValueError):
            bundle.version_of(link, "astra ")

    def test_missing_license_leaves_version_retryable(self):
        self.license.unlink()
        with self.assertRaises(FileNotFoundError):
            self.package("retry")
        self.assertEqual(list((self.root / "artifacts").iterdir()), [])
        self.license.write_text("restored license")
        archive, digest = self.package("retry")
        self.assertEqual(hashlib.sha256(archive.read_bytes()).hexdigest(), digest)

    def test_tar_and_checksum_failures_leave_no_partial_outputs(self):
        for target in ("tarfile.TarFile.addfile", "hashlib.file_digest", "pathlib.Path.write_text"):
            with self.subTest(target=target):
                with mock.patch(target, side_effect=OSError("injected output failure")):
                    with self.assertRaises(OSError):
                        self.package("retry")
                self.assertEqual(list((self.root / "artifacts").iterdir()), [])
        self.package("retry")

    def test_existing_checksum_is_preserved_and_owned_archive_removed(self):
        artifacts = self.root / "artifacts"
        artifacts.mkdir()
        checksum = artifacts / f"moi-astra-retry-{bundle.native_target()}.tar.gz.sha256"
        checksum.write_text("existing checksum")
        with self.assertRaises(FileExistsError):
            self.package("retry")
        self.assertEqual(list(artifacts.iterdir()), [checksum])
        self.assertEqual(checksum.read_text(), "existing checksum")

    def test_http_requires_explicit_loopback_opt_in(self):
        archive, digest = self.package("download")
        class handler(BaseHTTPRequestHandler):
            def do_GET(self):
                self.send_response(200)
                self.end_headers()
                self.wfile.write(archive.read_bytes())

            def log_message(self, *args):
                pass
        with download_server(handler) as port:
            url = f"http://127.0.0.1:{port}/bundle.tar.gz"
            args = ["sh", str(HERE / "install.sh"), "--url", url, "--sha256", digest, "--dir", str(self.prefix), "--preview-state"]
            denied = subprocess.run(args, capture_output=True, text=True, timeout=30)
            self.assertNotEqual(denied.returncode, 0)
            self.assertFalse(self.prefix.exists())
            # An environment proxy must never see a local preview request.
            result = subprocess.run(args + ["--allow-loopback-http"], capture_output=True, text=True, timeout=30,
                env={**os.environ, "http_proxy": "http://127.0.0.1:1", "ALL_PROXY": "http://127.0.0.1:1", "no_proxy": ""})
            self.assertEqual(result.returncode, 0, result.stderr + result.stdout)
        self.assertTrue((self.prefix / "current/astra").is_file())

    def test_http_rejects_non_loopback_and_authority_tricks(self):
        for url in ("http://example.com/file", "http://localhost:123/file", "http://127.0.0.1.example.com:123/file",
                    "http://127.0.0.1:123@example.com/file", "http://127.0.0.1:123/file?redirect=other"):
            with self.subTest(url=url):
                result = subprocess.run(["sh", str(HERE / "install.sh"), "--url", url, "--sha256", "0" * 64,
                    "--allow-loopback-http", "--dir", str(self.prefix), "--preview-state"], capture_output=True, text=True, timeout=10)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("literal 127.0.0.1", result.stderr)
                self.assertFalse(self.prefix.exists())

    def test_http_refuses_redirect_without_following(self):
        requests = []

        class Redirect(BaseHTTPRequestHandler):
            def do_GET(self):
                requests.append(self.path)
                self.send_response(302)
                self.send_header("Location", "/do-not-follow")
                self.end_headers()

            def log_message(self, *args):
                pass

        with download_server(Redirect) as port:
            result = subprocess.run(["sh", str(HERE / "install.sh"), "--url", f"http://127.0.0.1:{port}/bundle",
                "--sha256", "0" * 64, "--allow-loopback-http", "--dir", str(self.prefix), "--preview-state"],
                capture_output=True, text=True, timeout=10)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("redirects are refused", result.stderr)
        self.assertEqual(requests, ["/bundle"])
        self.assertFalse((self.prefix / "current").exists())


if __name__ == "__main__":
    unittest.main()
