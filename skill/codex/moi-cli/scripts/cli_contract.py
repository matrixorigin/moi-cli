#!/usr/bin/env python3
"""Inspect registered file inputs and formal CLI contract gaps without product calls."""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import os
import subprocess
import sys

import record_cli


FLAG_COMMANDS = {
    "--acceptance-file": [("workflow", "deploy-candidate")],
    "--workflow-file": [("workflow", "deploy")],
    "--runtime-fields-file": [("workflow", "deploy"), ("workflow", "update")],
    "--runtime-layout-file": [("workflow", "update")],
    "--default-values-file": [("workflow", "deploy"), ("workflow", "update")],
    "--values-file": [("workflow", "run"), ("workflow", "retry-run"), ("workflow", "create-run-node-rerun")],
    "--config-override-file": [("workflow", "create-run-node-rerun")],
    "--volume-trigger-file": [("workflow", "deploy"), ("workflow", "update")],
    "--dynamic-service-file": [("workflow", "deploy"), ("workflow", "update")],
    "--files-file": [("semantic-model", "create-with-sources"), ("semantic-model", "create-with-spec")],
    "--tables-file": [("semantic-model", "create-with-spec")],
    "--source-selections-file": [("semantic-model", "create-with-sources")],
    "--query-image": [("semantic-model", "search-drawings")],
}

CREDENTIAL_FLAG_COMMANDS = {
    "--api-key-file": ("model", "probe-llm-models"),
    "--runtime-token-file": ("runtime-gateway", "invoke-model"),
    "--actor-assertion-file": ("system", "service-account-workspaces"),
    "--provisioner-token-file": ("system", "service-account-workspaces"),
}


def load_gaps(root: Path) -> list[dict]:
    document = json.loads((root / "references/cli-contract-gaps.json").read_text())
    if document.get("schema_version") != 1 or not isinstance(document.get("entries"), list) or not document["entries"]:
        raise ValueError("invalid CLI gap registry")
    seen = set()
    for entry in document["entries"]:
        feature_id = entry.get("feature_id", "")
        if not feature_id or feature_id in seen:
            raise ValueError("missing or duplicate gap feature ID")
        seen.add(feature_id)
        if entry.get("scope") not in {"full_cli_missing", "partial_cli_gap"} or entry.get("status") not in {"cli_missing", "partial"}:
            raise ValueError(f"invalid gap scope/status: {feature_id}")
        for field in ("feature_name", "current_evidence", "missing_contract", "formal_next_action"):
            if not isinstance(entry.get(field), str) or not entry[field].strip():
                raise ValueError(f"gap {field} missing: {feature_id}")
        forbidden = entry.get("forbidden_substitutes")
        if not isinstance(forbidden, list) or not forbidden or any(not isinstance(x, str) or not x.strip() for x in forbidden):
            raise ValueError(f"forbidden substitutes missing: {feature_id}")
        if not isinstance(entry.get("evidence"), list) or not entry["evidence"]:
            raise ValueError(f"gap evidence missing: {feature_id}")
    return document["entries"]


def validate(root: Path, cli_bin: Path) -> dict:
    gaps = load_gaps(root)
    checked = []
    for flag, commands in FLAG_COMMANDS.items():
        if flag not in record_cli._STRUCTURED_FILES and flag not in record_cli._OPAQUE_FILES and flag not in record_cli._OUTPUTS:
            raise ValueError(f"file flag lacks recorder classification: {flag}")
        for domain, command in commands:
            result = subprocess.run([str(cli_bin), domain, command, "--help"], capture_output=True, text=True)
            if result.returncode or "-" + flag.lstrip("-") not in result.stdout + result.stderr:
                raise ValueError(f"registered flag is absent from formal help: {domain} {command} {flag}")
            checked.append({"flag": flag, "domain": domain, "command": command})
    for flag, (domain, command) in CREDENTIAL_FLAG_COMMANDS.items():
        if flag in record_cli._STRUCTURED_FILES or flag in record_cli._OPAQUE_FILES:
            raise ValueError(f"credential file must not be captured: {flag}")
        result = subprocess.run([str(cli_bin), domain, command, "--help"], capture_output=True, text=True)
        if result.returncode or "-" + flag.lstrip("-") not in result.stdout + result.stderr:
            raise ValueError(f"credential flag lacks formal help: {domain} {command} {flag}")
        checked.append({"flag": flag, "domain": domain, "command": command, "credential_refused": True})
    return {
        "valid": True,
        "gap_count": len(gaps),
        "full_cli_missing": sum(item["scope"] == "full_cli_missing" for item in gaps),
        "partial_cli_gap": sum(item["scope"] == "partial_cli_gap" for item in gaps),
        "structured_file_flags": len(record_cli._STRUCTURED_FILES),
        "opaque_file_flags": len(record_cli._OPAQUE_FILES),
        "output_flags": len(record_cli._OUTPUTS),
        "credential_file_flags_refused": len(record_cli._CREDENTIAL_FILES),
        "checked_flag_commands": len(checked),
        "cli_bin": str(cli_bin),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--cli-bin", type=Path, default=Path(os.environ.get("MOI_DEV_CLI_BIN", str(Path.home() / ".local/bin/moi-cli-dev"))))
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("file-inputs")
    sub.add_parser("validate")
    gaps = sub.add_parser("gaps")
    gaps.add_argument("--scope", choices=["full_cli_missing", "partial_cli_gap"])
    args = parser.parse_args()
    try:
        root = args.root.resolve()
        if args.command == "file-inputs":
            result = {
                "structured": record_cli._STRUCTURED_FILES,
                "opaque": sorted(record_cli._OPAQUE_FILES),
                "outputs": sorted(record_cli._OUTPUTS),
                "credential_files_refused": sorted(record_cli._CREDENTIAL_FILES),
                "default_cli_bin": str(Path.home() / ".local/bin/moi-cli-dev"),
            }
        elif args.command == "validate":
            result = validate(root, args.cli_bin.resolve())
        else:
            entries = load_gaps(root)
            result = [item for item in entries if not args.scope or item["scope"] == args.scope]
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0
    except (ValueError, OSError, KeyError, TypeError, json.JSONDecodeError) as error:
        print(json.dumps({"valid": False, "error_type": type(error).__name__, "message": str(error) if isinstance(error, ValueError) else "CLI contract files or help could not be read"}, ensure_ascii=False))
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
