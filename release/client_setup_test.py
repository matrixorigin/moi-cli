"""Production client setup uses only disposable homes, prefixes and shell files."""
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import tempfile
import unittest

from channel_installer import render


HERE = Path(__file__).resolve().parent


class ClientSetupTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="moi-setup-test.")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()
        self.home = self.root / "home"
        self.home.mkdir()
        self.env = dict(os.environ, HOME=str(self.home),
                        XDG_CONFIG_HOME=str(self.home / ".config"),
                        ZDOTDIR=str(self.home), MOI_AUTH_DIR=str(self.home / ".moi"),
                        MOI_CLI_CONFIG_FILE=str(self.home / ".moi-cli.json"),
                        ASTRA_LOCAL_STATE_ROOT=str(self.home / ".astra"),
                        ASTRA_CLI_CREDENTIALS_DIR=str(self.home / ".astra-legacy"))
        for key in ("BASH_ENV", "ENV", "CDPATH", "CURL_HOME", "MOI_ASTRA_BINARY"):
            self.env.pop(key, None)
        # Quoting must work with paths often found on user machines.
        self.prefix = self.root / "client's space"
        self.profile = self.root / "shell rc"
        self.command_dir = self.root / "command-bin"
        self.command_dir.mkdir()
        self.profile.write_text("# User configuration\nexport USER_SENTINEL=keep\n")
        self.prefix.mkdir()
        (self.prefix / "current").mkdir()
        for name in ("astra", "astra-edge", "moi-cli"):
            fake = self.prefix / "current" / name
            fake.write_text("#!/bin/sh\nprintf '%s\\n' \"$MOI_AUTH_DIR\" \"$MOI_CLI_CONFIG_FILE\" \"$MOI_ASTRA_BINARY\" \"$ASTRA_API_URL\" \"$ASTRA_LOCAL_STATE_ROOT\" \"$ASTRA_CLI_CREDENTIALS_DIR\" \"$@\"\n")
            fake.chmod(0o755)

    def setup_client(self, *, prefix=None, profile=None, auto_profile=False, command_dir=True, shell="/bin/sh", modify_path="yes", api_url="https://astra.example.test"):
        selected_prefix = prefix or self.prefix
        stage = Path(tempfile.mkdtemp(dir=self.root, prefix="stage."))
        script = self.root / "setup.sh"
        quoted_profile = "''" if auto_profile else shlex.quote(str(profile or self.profile))
        script.write_text("#!/bin/sh\nset -eu\nfail() { printf '%s\\n' \"$*\" >&2; exit 1; }\n" +
                          (HERE / "client_setup.sh").read_text() +
                          "\nclient_command_dir() { printf '%s\\n' " + shlex.quote(str(self.command_dir) if command_dir else "") + "; }\n" +
                          f"\nprefix={shlex.quote(str(selected_prefix))}\n" +
                          f"shell_profile={quoted_profile}\n" +
                          f"modify_path={modify_path}\nclient_api_url={shlex.quote(api_url)}\n" +
                          f"skill_dir={shlex.quote(str(selected_prefix / 'astra-state/skills'))}\n" +
                          f"stage={shlex.quote(str(stage))}\nclient_setup_preflight\nclient_setup\n")
        env = dict(self.env, PATH=f"{self.command_dir}:{os.environ['PATH']}", SHELL=shell)
        return subprocess.run(["sh", str(script)], text=True, capture_output=True, timeout=10, env=env)

    def test_setup_launchers_and_profile_are_idempotent_and_isolated(self):
        before = self.profile.read_bytes()
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertIn("ready in your current PATH", result.stdout)
        self.assertIn("Astra and moi-cli are installed.", result.stdout)
        self.assertIn("Sign in once with astra login to use both Astra and moi-cli.", result.stdout)
        self.assertIn("\nastra login\nastra\n", result.stdout)
        self.assertIn("In Astra, try: List my MOI workspaces", result.stdout)
        self.assertTrue(result.stdout.isascii())
        self.assertIn("\nmoi-cli workspace list\n", result.stdout)
        self.assertNotIn("astra-edge", result.stdout)
        self.assertEqual(str(self.command_dir), (self.prefix / "setup/command-dir").read_text().strip())
        for name in ("astra", "astra-edge", "moi-cli"):
            self.assertEqual(self.prefix / "bin" / name, (self.command_dir / name).readlink())
        self.assertEqual(before, (self.prefix / "setup/shell-1.before").read_bytes())
        rc = self.profile.read_bytes()
        self.assertTrue(rc.startswith(before))
        self.assertIn(b"MOI client", rc)
        self.assertFalse((self.prefix / "workspace").exists())
        self.assertNotIn("\\ncd ", result.stdout)
        for name in ("client-auth", "astra-legacy", "moi-cli.json"):
            self.assertFalse((self.prefix / name).exists())
        env = dict(self.env, ASTRA_API_URL="https://wrong.example")
        for name in ("astra", "astra-edge", "moi-cli"):
            output = subprocess.check_output([str(self.prefix / "bin" / name), "--version", "space argument", "quote'"], env=env, text=True).splitlines()
            self.assertEqual([env["MOI_AUTH_DIR"], env["MOI_CLI_CONFIG_FILE"], str(self.prefix / "bin/astra"),
                              "https://astra.example.test", str(self.prefix / "astra-state"),
                              env["ASTRA_CLI_CREDENTIALS_DIR"], "--version", "space argument", "quote'"], output)
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual(rc, self.profile.read_bytes())
        self.assertFalse((self.prefix / "activate.sh").exists())

    def test_other_installation_profile_is_rejected_before_modification(self):
        self.assertEqual(0, self.setup_client().returncode)
        before = self.profile.read_bytes()
        other = self.root / "other"
        result = self.setup_client(prefix=other, command_dir=False)
        self.assertNotEqual(0, result.returncode)
        self.assertIn("CLI_INSTALL_PROFILE_CONFLICT", result.stderr)
        self.assertIn(str(self.profile), result.stderr)
        self.assertIn("--no-modify-path", result.stderr)
        self.assertFalse(other.exists())
        self.assertEqual(before, self.profile.read_bytes())

    def test_launchers_preserve_users_working_directory(self):
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        for name in ("astra", "astra-edge", "moi-cli"):
            (self.prefix / "current" / name).write_text("#!/bin/sh\npwd -P\n")
        for directory_name in ("project one", "project's two"):
            directory = self.root / directory_name
            directory.mkdir()
            for name in ("astra", "astra-edge", "moi-cli"):
                with self.subTest(directory=directory_name, binary=name):
                    output = subprocess.check_output([str(self.command_dir / name)], cwd=directory, env=self.env, text=True)
                    self.assertEqual(str(directory), output.strip())

    def test_symlink_profile_is_left_unchanged_and_install_continues(self):
        link = self.root / "linked rc"
        link.symlink_to(self.profile)
        before = self.profile.read_bytes()
        result = self.setup_client(profile=link, command_dir=False)
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual(before, self.profile.read_bytes())
        self.assertTrue(link.is_symlink())
        self.assertTrue((self.prefix / "bin/astra").is_file())
        self.assertIn("Shell profile is a symlink and was left unchanged:", result.stdout)
        self.assertIn(str(link), result.stdout)
        self.assertIn("--shell-profile", result.stdout)
        command = result.stdout.split("Run now without changing PATH: ", 1)[1].splitlines()[0]
        self.assertEqual([str(self.prefix / "bin/astra"), "login"], shlex.split(command))
        self.assertNotIn("# >>> MOI client >>>", self.profile.read_text())

    def test_bash_symlink_profiles_stay_unchanged(self):
        target = self.home / "dotfiles-bashrc"
        target.write_text("# managed bashrc\n")
        bashrc = self.home / ".bashrc"
        bashrc.symlink_to(target)
        login_target = self.home / "dotfiles-profile"
        login_target.write_text("# managed profile\n")
        login = self.home / ".profile"
        login.symlink_to(login_target)
        before_rc = target.read_bytes()
        before_login = login_target.read_bytes()
        result = self.setup_client(auto_profile=True, shell="/bin/bash", command_dir=False)
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual(before_rc, target.read_bytes())
        self.assertEqual(before_login, login_target.read_bytes())
        self.assertTrue(bashrc.is_symlink())
        self.assertTrue(login.is_symlink())
        self.assertFalse((self.home / ".bash_profile").exists())
        self.assertIn(str(bashrc), result.stdout)
        self.assertIn(str(login), result.stdout)
        self.assertIn("--shell-profile", result.stdout)
        command = result.stdout.split("Run now without changing PATH: ", 1)[1].splitlines()[0]
        self.assertEqual([str(self.prefix / "bin/astra"), "login"], shlex.split(command))

    def test_bash_writes_regular_profile_and_skips_symlink(self):
        bashrc = self.home / ".bashrc"
        bashrc.write_text("# user bashrc\n")
        login_target = self.home / "dotfiles-profile"
        login_target.write_text("# managed profile\n")
        login = self.home / ".bash_profile"
        login.symlink_to(login_target)
        before_login = login_target.read_bytes()
        result = self.setup_client(auto_profile=True, shell="/bin/bash")
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertIn("# >>> MOI client >>>", bashrc.read_text())
        self.assertEqual(before_login, login_target.read_bytes())
        self.assertIn(str(login), result.stdout)
        self.assertIn("ready in your current PATH", result.stdout)

    def test_zsh_symlink_profile_prints_path_followup(self):
        target = self.home / "dotfiles-zshrc"
        target.write_text("# managed zshrc\n")
        zshrc = self.home / ".zshrc"
        zshrc.symlink_to(target)
        before = target.read_bytes()
        result = self.setup_client(auto_profile=True, shell="/bin/zsh", command_dir=False)
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual(before, target.read_bytes())
        self.assertTrue(zshrc.is_symlink())
        self.assertIn("Shell profile is a symlink and was left unchanged:", result.stdout)
        self.assertIn(str(zshrc), result.stdout)
        self.assertIn("--shell-profile", result.stdout)
        hint = result.stdout.split("To use this terminal, copy and run:\n", 1)[1].splitlines()[0]
        self.assertTrue(hint.startswith("export PATH="))
        self.assertIn("space/bin", hint)
        command = result.stdout.split("Run now without changing PATH: ", 1)[1].splitlines()[0]
        self.assertEqual([str(self.prefix / "bin/astra"), "login"], shlex.split(command))

    def test_existing_launcher_is_not_overwritten(self):
        (self.prefix / "bin").mkdir()
        launcher = self.prefix / "bin/astra"
        launcher.write_text("user owned")
        before = self.profile.read_bytes()
        result = self.setup_client()
        self.assertNotEqual(0, result.returncode)
        self.assertIn("CLI_INSTALL_LAUNCHER_CONFLICT", result.stderr)
        self.assertIn(str(launcher), result.stderr)
        self.assertIn("switching environments", result.stderr)
        self.assertIn("--dir and --skill-dir", result.stderr)
        self.assertEqual("user owned", launcher.read_text())
        self.assertEqual(before, self.profile.read_bytes())

    def test_qa_launcher_is_not_replaced_by_prod_launcher(self):
        qa = self.setup_client(api_url="https://astra-qa.example.test")
        self.assertEqual(0, qa.returncode, qa.stderr)
        launcher = self.prefix / "bin/astra"
        before = launcher.read_bytes()
        profile_before = self.profile.read_bytes()

        prod = self.setup_client()
        self.assertNotEqual(0, prod.returncode)
        self.assertIn("CLI_INSTALL_LAUNCHER_CONFLICT", prod.stderr)
        self.assertIn(str(launcher), prod.stderr)
        self.assertIn("switching environments", prod.stderr)
        self.assertEqual(before, launcher.read_bytes())
        self.assertEqual(profile_before, self.profile.read_bytes())

    def test_missing_profile_created(self):
        self.profile.unlink()
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertIn("export PATH=", self.profile.read_text())

    def test_same_installation_can_configure_another_shell_without_losing_backups(self):
        first_before = self.profile.read_bytes()
        self.assertEqual(0, self.setup_client().returncode)
        second = self.root / "another shell rc"
        second.write_text("# Other shell settings\n")
        second_before = second.read_bytes()
        result = self.setup_client(profile=second)
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual(first_before, (self.prefix / "setup/shell-1.before").read_bytes())
        self.assertEqual(second_before, (self.prefix / "setup/shell-2.before").read_bytes())
        self.assertEqual(str(self.profile), (self.prefix / "setup/shell-1.path").read_text().strip())
        self.assertEqual(str(second), (self.prefix / "setup/shell-2.path").read_text().strip())
        self.assertEqual(0, self.setup_client(profile=second).returncode)
        self.assertEqual(1, second.read_text().count("# >>> MOI client >>>"))

    def test_current_shell_works_without_reloading_profile(self):
        env = dict(self.env, PATH=f"{self.command_dir}:{os.environ['PATH']}")
        # Start the same shell BEFORE install; do not source rc or change PATH.
        for shell_name in ("sh", "bash", "zsh"):
            shell = shutil.which(shell_name)
            if shell is None:
                continue
            with self.subTest(shell=shell):
                process = subprocess.Popen([shell, "-f"], stdin=subprocess.PIPE,
                                           stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                           text=True, env=env)
                try:
                    process.stdin.write("printf 'shell-ready\\n'\n")
                    process.stdin.flush()
                    self.assertEqual("shell-ready\n", process.stdout.readline())
                    result = self.setup_client()
                    self.assertEqual(0, result.returncode, result.stderr)
                    output, errors = process.communicate("astra --version\n", timeout=10)
                    self.assertEqual(0, process.returncode, errors)
                    self.assertIn(str(self.home / ".moi"), output)
                    self.assertIn("--version", output)
                finally:
                    if process.poll() is None:
                        process.kill()
                        process.communicate()
                for name in ("astra", "astra-edge", "moi-cli"):
                    (self.command_dir / name).unlink()

    def test_no_searchable_user_bin_gives_immediate_absolute_command(self):
        result = self.setup_client(command_dir=False)
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertIn("Run now without changing PATH:", result.stdout)
        self.assertNotIn("ready in your current PATH", result.stdout)
        command = result.stdout.split("Or use moi-cli directly after signing in:\n", 1)[1].splitlines()[0]
        self.assertEqual([str(self.prefix / "bin/moi-cli"), "workspace", "list"], shlex.split(command))
        self.assertFalse((self.prefix / "setup/command-dir").exists())

    def test_current_terminal_path_hint_resolves_custom_prefix(self):
        for shell in ("sh", "bash", "zsh", "fish"):
            with self.subTest(shell=shell):
                # Separate profiles avoid mixing POSIX and fish persistent blocks.
                profile = self.root / (shell + "rc")
                result = self.setup_client(command_dir=False, shell="/usr/bin/" + shell, profile=profile)
                self.assertEqual(0, result.returncode, result.stderr)
                hint = result.stdout.split("To use this terminal, copy and run:\n", 1)[1].splitlines()[0]
                self.assertTrue(hint.startswith("set -gx PATH " if shell == "fish" else "export PATH="))
                executable = shutil.which(shell)
                if executable:
                    check = "command -s astra" if shell == "fish" else "command -v astra"
                    resolved = subprocess.run([executable, "-c", hint + "\n" + check],
                                              env=self.env, text=True, capture_output=True, timeout=10)
                    self.assertEqual(0, resolved.returncode, resolved.stderr)
                    self.assertEqual(str(self.prefix / "bin/astra"), resolved.stdout.strip())

    def test_unknown_shell_does_not_print_an_inapplicable_export(self):
        result = self.setup_client(command_dir=False, shell="/bin/unknown")
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertNotIn("To use this terminal, copy and run:", result.stdout)
        self.assertIn("Run now without changing PATH:", result.stdout)

    def test_existing_user_command_is_never_overwritten(self):
        for name in ("astra", "astra-edge", "moi-cli"):
            with self.subTest(name=name):
                existing = self.command_dir / name
                existing.write_text("existing tool")
                result = self.setup_client()
                self.assertNotEqual(0, result.returncode)
                self.assertIn("CLI_INSTALL_COMMAND_CONFLICT", result.stderr)
                self.assertIn(str(existing), result.stderr)
                self.assertIn("--no-modify-path", result.stderr)
                self.assertEqual("existing tool", existing.read_text())
                self.assertFalse((self.prefix / "bin").exists())
                existing.unlink()

    def test_stale_setup_lock_is_recovered_but_live_owner_is_preserved(self):
        setup = self.prefix / "setup"
        setup.mkdir()
        lock = setup / "setup.lock"
        lock.write_text("99999999\n")
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertFalse(lock.exists())

        lock.write_text(f"{os.getpid()}\n")
        before = self.profile.read_bytes()
        result = self.setup_client()
        self.assertNotEqual(0, result.returncode)
        self.assertIn("already running", result.stderr)
        self.assertEqual(before, self.profile.read_bytes())
        self.assertEqual(f"{os.getpid()}\n", lock.read_text())
        self.assertFalse(list(setup.glob(".setup-lock.*")))

    def test_profile_publication_leaves_no_partial_stage(self):
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertFalse(list(self.profile.parent.glob(".moi-profile.*")))
        self.assertEqual(1, self.profile.read_text().count("# >>> MOI client >>>"))

    def test_complete_backup_from_interrupted_setup_is_reused(self):
        setup = self.prefix / "setup"
        setup.mkdir()
        before = self.profile.read_bytes()
        (setup / "shell-1.before").write_bytes(before)
        (setup / "shell-1.path").write_text(str(self.profile) + "\n")
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual(before, (setup / "shell-1.before").read_bytes())
        self.assertEqual(1, self.profile.read_text().count("# >>> MOI client >>>"))

    def test_production_renderer_rejects_local_channel(self):
        with self.assertRaisesRegex(ValueError, "HTTPS"):
            render({"platforms": {}, "compatibility_reference": "contract-v1"},
                   "http://127.0.0.1:28191/channels/stable.json")

    def test_production_renderer_includes_shared_setup_and_trusted_service(self):
        data = render({"platforms": {}, "compatibility_reference": "server-contract-v1", "client_config": {"astra_url": "https://astra.example.test"}},
                      "https://example.com/channels/stable.json")
        self.assertIn(b"client_setup_preflight", data)
        self.assertNotIn(b"127.0.0.1", data)
        self.assertNotIn(b"client-preview", data)
        self.assertIn(b"client_api_url=https://astra.example.test", data)
        self.assertIn(b"--shell-profile", data)

    def test_production_does_not_create_preview_credentials_or_override_user_state(self):
        result = self.setup_client()
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertIn("Sign in once with astra login to use both Astra and moi-cli.", result.stdout)
        self.assertIn("In Astra, try: List my MOI workspaces", result.stdout)
        self.assertTrue(result.stdout.isascii())
        self.assertIn("\nmoi-cli workspace list\n", result.stdout)
        self.assertNotIn("astra-edge", result.stdout)
        for name in ("client-auth", "astra-state", "astra-legacy", "moi-cli.json"):
            self.assertFalse((self.prefix / name).exists())
        env = dict(self.env, MOI_AUTH_DIR="/explicit/auth", MOI_CLI_CONFIG_FILE="/explicit/config",
                   ASTRA_LOCAL_STATE_ROOT="/explicit/state", ASTRA_CLI_CREDENTIALS_DIR="/explicit/legacy")
        output = subprocess.check_output([str(self.command_dir / "astra"), "login"], env=env, text=True).splitlines()
        self.assertEqual(["/explicit/auth", "/explicit/config", str(self.prefix / "bin/astra"),
                          "https://astra.example.test", str(self.prefix / "astra-state"), "/explicit/legacy", "login"], output)

    def test_path_block_is_safe_and_idempotent_in_posix_shells(self):
        self.assertEqual(0, self.setup_client().returncode)
        for shell in ("sh", "bash", "zsh"):
            if not shutil.which(shell):
                continue
            result = subprocess.run([shell, "-c", '. "$1"; . "$1"; command -v astra; printf "%s\\n" "$PATH"',
                                     "test", str(self.profile)], env=self.env, text=True, capture_output=True)
            self.assertEqual(0, result.returncode, result.stderr)
            self.assertEqual(str(self.prefix / "bin/astra"), result.stdout.splitlines()[0])
            self.assertEqual(1, result.stdout.splitlines()[1].split(":").count(str(self.prefix / "bin")))

    def test_fish_uses_fish_syntax(self):
        result = self.setup_client(shell="/usr/bin/fish")
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertIn("set -gx PATH", self.profile.read_text())
        self.assertNotIn("export PATH", self.profile.read_text())
        fish = shutil.which("fish")
        if fish:
            result = subprocess.run([fish, "-c", 'source $argv[1]; source $argv[1]; command -s astra',
                                     str(self.profile)], env=self.env, text=True, capture_output=True)
            self.assertEqual(0, result.returncode, result.stderr)
            self.assertEqual(str(self.prefix / "bin/astra"), result.stdout.strip())

    def test_no_modify_path_preserves_profile(self):
        before = self.profile.read_bytes()
        result = self.setup_client(modify_path="no", command_dir=False)
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual(before, self.profile.read_bytes())
        self.assertIn("Run now without changing PATH:", result.stdout)
        command = result.stdout.split("Or use moi-cli directly after signing in:\n", 1)[1].splitlines()[0]
        self.assertEqual([str(self.prefix / "bin/moi-cli"), "workspace", "list"], shlex.split(command))

    def test_no_modify_path_preserves_existing_commands(self):
        sentinel = self.command_dir / "astra"
        sentinel.write_text("existing independent Astra\n")
        result = self.setup_client(modify_path="no")
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual("existing independent Astra\n", sentinel.read_text())
        self.assertFalse((self.command_dir / "moi-cli").exists())
        self.assertFalse((self.prefix / "setup/command-dir").exists())

    def test_production_requires_explicit_safe_service_url(self):
        for url in (None, "", "http://127.0.0.1:17001", "https://user:secret@example.com", "https://example.com?token=x", "https://example.com/path", "https://example.com\n"):
            with self.subTest(url=url), self.assertRaises(ValueError):
                render({"platforms": {}, "compatibility_reference": "server-contract-v1", "client_config": {"astra_url": url}}, "https://example.com/channels/stable.json")


if __name__ == "__main__":
    unittest.main()
