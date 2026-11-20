#!/usr/bin/env python3
"""Inspect saved CLI list-runs JSON. No network access or writes."""
import argparse
import json
import sys
from collections import Counter
from pathlib import Path


def inspect(data, dispatch_id, expected_files, version_id):
    if isinstance(data, dict) and "response" in data:
        if data.get("exit_code") != 0:
            raise ValueError("CLI receipt is not successful")
        data = data["response"]
    if not isinstance(data, dict) or not isinstance(data.get("executions"), list):
        raise ValueError("list-runs response must contain an executions array")
    executions = data["executions"]
    if any(not isinstance(row, dict) for row in executions):
        raise ValueError("execution records must be objects")
    rows = [row for row in executions if row.get("dispatch_job_id") == dispatch_id]
    if any(not isinstance(row.get("file_id"), str) or not row["file_id"] for row in rows):
        raise ValueError("matched child records must identify their input file")
    files = Counter(row.get("file_id") for row in rows)
    expected = set(expected_files)
    errors = []
    unexpected = [file_id for file_id in files if file_id not in expected]
    if unexpected:
        errors.append("unexpected input files in this dispatch")
    if any(count > 1 for count in files.values()):
        errors.append("multiple executions for an input file; inspect attempts explicitly")
    ids = [row.get("execution_id") for row in rows]
    if any(not isinstance(value, str) or not value or value == dispatch_id for value in ids):
        errors.append("missing or invalid child execution ID")
    if len({value for value in ids if isinstance(value, str)}) != len(ids):
        errors.append("duplicate child execution ID")
    if version_id and any(row.get("moi_workflow_version_id") != version_id for row in rows):
        errors.append("workflow version mismatch")
    missing = sorted(expected - set(files))
    complete_listing = type(data.get("total")) is int and data["total"] == len(executions)
    failed = any(row.get("status") in {"failed", "cancelled", "rejected"} for row in rows)
    if errors:
        state, code = "mismatch", 2
    elif failed:
        state, code = "failed", 3
    elif not complete_listing or missing or not rows or any(row.get("status") != "completed" for row in rows):
        state, code = "pending_or_incomplete_observation", 4
    else:
        state, code = "execution_set_completed", 0
    return {
        "state": state,
        "dispatch_id": dispatch_id,
        "listing_complete": complete_listing,
        "missing_file_ids": missing,
        "errors": errors,
        "executions": [
            {key: row.get(key) for key in (
                "file_id", "execution_id", "status", "moi_case_id", "moi_workflow_version_id"
            )}
            for row in rows
        ],
        "scope": "Execution set only; verify results, artifacts and Knowledge citations separately.",
    }, code


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("runs_json", type=Path)
    parser.add_argument("--dispatch-id", required=True)
    parser.add_argument("--expected-file-id", action="append", required=True)
    parser.add_argument("--version-id")
    args = parser.parse_args()
    try:
        report, code = inspect(
            json.loads(args.runs_json.read_text()), args.dispatch_id,
            args.expected_file_id, args.version_id,
        )
    except (ValueError, OSError) as error:
        print(json.dumps({"state": "invalid_input", "error": str(error)}, ensure_ascii=False))
        return 2
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return code


if __name__ == "__main__":
    sys.exit(main())
