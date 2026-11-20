#!/usr/bin/env python3
"""Validate required horizontal/vertical/comparison reports for one iteration."""

import json
from pathlib import Path
import re
import sys


REQUIRED = ["横向纵向功能报告.md", "原版与dev版功能对比报告.md", "总体进展报告.md"]


def main():
    if len(sys.argv) != 3:
        print("usage: iteration_report_check.py EVIDENCE_DIR PACKAGE_ROOT", file=sys.stderr)
        return 2
    evidence = Path(sys.argv[1]).resolve()
    package = Path(sys.argv[2]).resolve()
    registry = json.loads((package / "references/capability-cards.json").read_text())
    skill_text = (package / "SKILL.md").read_text()
    version_match = re.search(r"(?m)^\s+skill-version:\s*[\"']?([0-9]+\.[0-9]+\.[0-9]+)[\"']?\s*$", skill_text)
    if not version_match:
        raise ValueError("SKILL.md metadata.skill-version is missing or invalid")
    version = version_match.group(1)
    errors = []
    for name in REQUIRED:
        if not (evidence / name).is_file():
            errors.append(f"missing report: {name}")
    manifest_path = evidence / "report-manifest.json"
    if not manifest_path.is_file():
        errors.append("missing report-manifest.json")
        manifest = {}
    else:
        manifest = json.loads(manifest_path.read_text())
    if manifest.get("skill_version") != version:
        errors.append(f"manifest skill_version {manifest.get('skill_version')} != {version}")
    if not manifest.get("cli_version"):
        errors.append("cli_version missing")
    if not re.fullmatch(r"[0-9a-f]{40}", manifest.get("source_commit", "")):
        errors.append("source_commit must be the full 40-character commit")
    if not manifest.get("environment"):
        errors.append("environment missing")
    if not manifest.get("limitations"):
        errors.append("limitations must be a non-empty list")
    if not manifest.get("evidence"):
        errors.append("evidence links must be non-empty")
    if manifest.get("reports") != REQUIRED:
        errors.append("reports must list the three required filenames in order")

    horizontal = evidence / REQUIRED[0]
    comparison = evidence / REQUIRED[1]
    overall = evidence / REQUIRED[2]
    if horizontal.is_file():
        text = horizontal.read_text()
        for menu in registry["menus"]:
            if menu["menu"] not in text:
                errors.append(f"horizontal report missing menu: {menu['menu']}")
        for domain in ("Workflow", "Catalog", "Knowledge"):
            if domain not in text:
                errors.append(f"horizontal report missing vertical domain: {domain}")
    if comparison.is_file() and "原版" not in comparison.read_text():
        errors.append("comparison report lacks original-version comparison")
    if overall.is_file() and "限制" not in overall.read_text():
        errors.append("overall report lacks limitations")

    for doc in [horizontal, comparison, overall]:
        if not doc.is_file():
            continue
        for target in re.findall(r"\[[^]]+\]\(([^)]+)\)", doc.read_text()):
            target = target.split("#", 1)[0].strip()
            if not target or target.startswith(("http://", "https://", "mailto:", "<")):
                continue
            if not (doc.parent / target).resolve().exists():
                errors.append(f"broken report link: {doc.name} -> {target}")
    result = {"valid": not errors, "skill_version": version, "reports": REQUIRED, "errors": errors}
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if not errors else 1


if __name__ == "__main__":
    raise SystemExit(main())
