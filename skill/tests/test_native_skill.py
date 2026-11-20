from __future__ import annotations

import filecmp
import hashlib
import copy
import importlib.util
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[1]
SKILL_DIR = ROOT / "moi-cli"
SCRIPTS = SKILL_DIR / "scripts"
CLI = Path(os.environ.get("MOI_CLI_BIN", str(ROOT.parents[1] / "dist/bin/moi-cli")))


class NativeSkillTests(unittest.TestCase):
    def setUp(self):
        self.env = {**os.environ, "MOI_CLI_BIN": str(CLI), "MOI_DEV_CLI_BIN": str(CLI), "PYTHONDONTWRITEBYTECODE": "1"}
        self.temp = tempfile.TemporaryDirectory()

    def tearDown(self):
        self.temp.cleanup()

    def test_latest_baseline_is_release_523_without_local_artifacts(self):
        text = (SKILL_DIR / "SKILL.md").read_text()
        self.assertIn("name: moi-cli", text)
        self.assertIn("skill-version: 5.2.3", text)
        for path in SKILL_DIR.rglob("*"):
            self.assertFalse(path.is_symlink(), path)
            self.assertNotIn("__pycache__", path.parts)
            if path.is_file() and path.suffix in {".md", ".json", ".py", ".yaml", ".sh"}:
                content = path.read_text()
                self.assertNotIn("/Users/", content, path)
                self.assertNotIn("../../../../reference/", content, path)
                self.assertIsNone(re.search(r"sk-uca-v\d+-[A-Za-z0-9_-]{12,}", content), path)

    def test_cli_export_matches_latest_skill(self):
        target = Path(self.temp.name) / "exported"
        subprocess.run([str(CLI), "skill", "export", "--dir", str(target)], check=True, env=self.env)
        comparisons = [filecmp.dircmp(SKILL_DIR, target)]
        while comparisons:
            comparison = comparisons.pop()
            self.assertEqual(comparison.left_only, [])
            self.assertEqual(comparison.right_only, [])
            self.assertEqual(comparison.diff_files, [])
            self.assertEqual(comparison.funny_files, [])
            comparisons.extend(comparison.subdirs.values())

    def test_latest_templates_capabilities_and_cli_contract(self):
        commands = [
            [sys.executable, str(SCRIPTS / "task_blueprint.py"), "validate"],
            [sys.executable, str(SCRIPTS / "capability_status.py"), "render", "--check"],
            [sys.executable, str(SCRIPTS / "cli_contract.py"), "--cli-bin", str(CLI), "validate"],
        ]
        for command in commands:
            result = subprocess.run(command, cwd=SKILL_DIR, capture_output=True, text=True, env=self.env)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_iteration_manifest_points_to_complete_human_reports(self):
        iteration = SKILL_DIR / "references/iterations/5.2.3"
        manifest = json.loads((iteration / "report-manifest.json").read_text())
        self.assertEqual(manifest["skill_version"], "5.2.3")
        self.assertFalse(manifest["new_online_acceptance_in_this_repository_change"])
        for name in ("横向纵向功能报告.md", "原版与dev版功能对比报告.md", "总体进展报告.md"):
            self.assertIn(name, manifest["reports"])
            self.assertTrue((iteration / name).is_file(), name)
        self.assertIn("失败原因分类.md", manifest["supplementary_reports"])

    def test_sanitized_packaged_references_match_snapshot_hashes(self):
        document = json.loads((SKILL_DIR / "references/capability-cards.json").read_text())
        for source in document["sources"].values():
            if source.get("kind") == "packaged_reference" and "snapshot_sha256" in source:
                actual = hashlib.sha256((SKILL_DIR / "references" / source["path"]).read_bytes()).hexdigest()
                self.assertEqual(actual, source["snapshot_sha256"], source["path"])

    def test_iteration_checker_runs_in_installed_layout(self):
        installed = Path(self.temp.name) / "moi-cli"
        shutil.copytree(SKILL_DIR, installed)
        env = dict(self.env)
        env.pop("MOI_ITERATION_EVIDENCE_DIR", None)
        result = subprocess.run(["bash", str(installed / "scripts/validate.sh")], cwd=installed, capture_output=True, text=True, env=env)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertTrue(json.loads(result.stdout)["valid"])

    def test_capability_registry_rejects_substituted_product_entries(self):
        spec = importlib.util.spec_from_file_location("capability_status_under_test", SCRIPTS / "capability_status.py")
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        document = json.loads((SKILL_DIR / "references/capability-cards.json").read_text())
        bad_menus = copy.deepcopy(document)
        bad_menus["menus"][0]["menu"] = "非产品入口"
        with self.assertRaisesRegex(ValueError, "required menu identities"):
            module.validate(bad_menus, SKILL_DIR)
        bad_capabilities = copy.deepcopy(document)
        bad_capabilities["cards"][0]["capability_id"] = "not-a-product-capability"
        with self.assertRaisesRegex(ValueError, "required capability identities"):
            module.validate(bad_capabilities, SKILL_DIR)

    def test_source_map_points_to_native_skill_root(self):
        repository = ROOT.parents[2]
        source_map = json.loads((repository / "moi/assets/session-workspace/reference/source-map.json").read_text())
        cli_modules = [module for module in source_map["modules"] if module["path"].startswith("sdk/cli/skill")]
        self.assertEqual([module["path"] for module in cli_modules], ["sdk/cli/skill/moi-cli"])
        for module in cli_modules:
            self.assertTrue((repository / module["path"]).is_dir(), module["path"])


if __name__ == "__main__":
    unittest.main()
