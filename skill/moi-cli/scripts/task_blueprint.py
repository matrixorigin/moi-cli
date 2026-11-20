#!/usr/bin/env python3
"""Validate task blueprints, initialize task_run plans and assess safe recovery.

This helper never invokes moi-cli. Initialization writes a local task journal;
recovery only inspects that journal and never submits, retries or repairs anything.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import sys

import task_run


NAME = re.compile(r"[a-z0-9]+(?:-[a-z0-9]+)*\Z")


def root_dir(value: Path | None) -> Path:
    return (value or Path(__file__).resolve().parents[1]).resolve()


def blueprint_paths(root: Path) -> list[Path]:
    return sorted((root / "assets/task-blueprints").glob("*.json"))


def read_blueprint(path: Path, capabilities: set[str]) -> dict:
    data = json.loads(path.read_text())
    if data.get("schema_version") != 1 or not NAME.fullmatch(data.get("blueprint_id", "")):
        raise ValueError(f"invalid blueprint identity: {path.name}")
    if not re.fullmatch(r"\d+\.\d+\.\d+", data.get("version", "")) or not data.get("title") or not data.get("goal"):
        raise ValueError(f"invalid blueprint metadata: {path.name}")
    inputs = data.get("required_inputs")
    if not isinstance(inputs, list) or not inputs or len(inputs) != len(set(inputs)) or any(not NAME.fullmatch(x.replace("_", "-")) for x in inputs):
        raise ValueError(f"invalid required inputs: {path.name}")
    steps = data.get("steps")
    if not isinstance(steps, list) or not steps:
        raise ValueError(f"blueprint has no steps: {path.name}")
    seen = set()
    for step in steps:
        step_id = step.get("step_id", "")
        if not NAME.fullmatch(step_id) or step_id in seen:
            raise ValueError(f"invalid or duplicate step: {path.name}/{step_id}")
        if step.get("capability_id") not in capabilities or step.get("effect") not in {"read", "write"}:
            raise ValueError(f"invalid capability/effect: {path.name}/{step_id}")
        if not step.get("intent") or not isinstance(step.get("acceptance"), list) or not step["acceptance"] or any(not isinstance(x, str) or not x.strip() for x in step["acceptance"]):
            raise ValueError(f"step intent/acceptance missing: {path.name}/{step_id}")
        dependencies = step.get("depends_on")
        if not isinstance(dependencies, list) or len(dependencies) != len(set(dependencies)) or any(dep not in seen for dep in dependencies):
            raise ValueError(f"dependencies must point to earlier unique steps: {path.name}/{step_id}")
        seen.add(step_id)
    cleanup = data.get("cleanup", {})
    if cleanup.get("requires_explicit_confirmation") is not True or not cleanup.get("objects") or not cleanup.get("verification"):
        raise ValueError(f"cleanup confirmation/verification missing: {path.name}")
    return data


def load_all(root: Path) -> dict[str, dict]:
    cards = json.loads((root / "references/capability-cards.json").read_text())["cards"]
    capabilities = {card["capability_id"] for card in cards}
    paths = blueprint_paths(root)
    if not paths:
        raise ValueError("no task blueprints found")
    result = {}
    for path in paths:
        blueprint = read_blueprint(path, capabilities)
        if blueprint["blueprint_id"] in result:
            raise ValueError("duplicate blueprint ID")
        result[blueprint["blueprint_id"]] = blueprint
    return result


def initialize(root: Path, blueprint_id: str, run_dir: Path, workspace_id: str, endpoint: str, inputs: list[str]) -> dict:
    blueprint = load_all(root).get(blueprint_id)
    if blueprint is None:
        raise ValueError("unknown blueprint ID")
    values = {}
    for item in inputs:
        key, separator, value = item.partition("=")
        if not separator or key not in blueprint["required_inputs"] or key in values or not value.strip():
            raise ValueError("inputs must be unique nonempty KEY=VALUE entries declared by the blueprint")
        values[key] = value
    if set(values) != set(blueprint["required_inputs"]):
        raise ValueError("all blueprint required inputs must be supplied explicitly")
    task_run.safe_metadata(values)
    created = task_run.init(run_dir, workspace_id, endpoint, f"{blueprint['goal']} [blueprint:{blueprint_id}@{blueprint['version']}]")
    for step in blueprint["steps"]:
        task_run.add_step(run_dir, step["step_id"], step["capability_id"], step["intent"], step["effect"], step["acceptance"], step["depends_on"])
    with task_run.journal(run_dir) as (root_path, data):
        data["blueprint"] = {"id": blueprint_id, "version": blueprint["version"], "inputs": values, "cleanup": blueprint["cleanup"]}
        task_run.event(data, "blueprint_bound", blueprint_id=blueprint_id, blueprint_version=blueprint["version"])
        task_run.atomic_json(root_path / "run.json", data)
    return {**created, "blueprint_id": blueprint_id, "blueprint_version": blueprint["version"], "step_count": len(blueprint["steps"]), "next": recover(run_dir)}


def recover(run_dir: Path) -> dict:
    status = task_run.status(run_dir)
    with task_run.journal(run_dir) as (_, data):
        primary_owners = {}
        for step_id, step in data["steps"].items():
            primary = step.get("primary_invocation")
            if primary is None:
                continue
            if primary in primary_owners:
                raise ValueError("one primary invocation cannot belong to multiple steps")
            primary_owners[primary] = step_id
        uncertain = [step_id for step_id, step in data["steps"].items() if step["state"] in {"unknown", "submitted"}]
        if uncertain:
            observations = []
            for step_id in uncertain:
                primary = data["steps"][step_id]["primary_invocation"]
                call = data["calls"].get(primary, {})
                observations.append({"step_id": step_id, "invocation_id": primary, "target_reference": call.get("target_reference")})
            return {"state": "blocked_unknown_submission", "allowed_new_primary": False, "next_action": "observe_original_object_only", "uncertain": observations, "note": "Unknown/submitted writes cannot be replayed. Use saved typed references and formal read commands; without the original ID keep unresolved."}
        failed = []
        for step_id, step in data["steps"].items():
            if any(history and not history[-1]["passed"] for history in step["checks"].values()):
                failed.append(step_id)
        if failed:
            return {"state": "acceptance_failed", "allowed_new_primary": False, "failed_steps": failed, "next_action": "preserve_failure_and_create_explicit_authorized_recovery_step", "note": "Do not edit the old expectation or replay the original call."}
        observed = [step_id for step_id, step in data["steps"].items() if step["state"] == "observed"]
        if observed:
            return {"state": "acceptance_incomplete", "allowed_new_primary": False, "observed_steps": observed, "next_action": "complete_original_acceptance_checks", "note": "A response exists; finish its original acceptance evidence instead of submitting another primary call."}
        steps = status["steps"]
        if steps and all(step["state"] == "verified" for step in steps.values()):
            return {"state": "completed", "allowed_new_primary": False, "next_action": "review_results_and_apply_separately_authorized_cleanup"}
        ready = [step_id for step_id, step in steps.items() if step["state"] == "planned" and all(steps[dep]["state"] == "verified" for dep in step["depends_on"])]
        if ready:
            selected = ready[0]
            return {"state": "ready_for_single_execution", "allowed_new_primary": True, "next_step_id": selected, "next_action": "execute_one_explicitly_authorized_formal_command", "note": "Only this step is ready. Its one primary invocation must use a new invocation ID and the saved task scope."}
        return {"state": "waiting_dependencies", "allowed_new_primary": False, "next_action": "finish_upstream_acceptance_first", "steps": {step_id: step["state"] for step_id, step in steps.items()}}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("list")
    sub.add_parser("validate")
    show = sub.add_parser("show")
    show.add_argument("blueprint_id")
    init = sub.add_parser("init")
    init.add_argument("blueprint_id")
    init.add_argument("--run-dir", type=Path, required=True)
    init.add_argument("--workspace-id", required=True)
    init.add_argument("--endpoint", required=True)
    init.add_argument("--input", action="append", default=[])
    resume = sub.add_parser("recover")
    resume.add_argument("--run-dir", type=Path, required=True)
    args = parser.parse_args()
    try:
        root = root_dir(args.root)
        if args.command == "list":
            result = [{"blueprint_id": key, "version": value["version"], "title": value["title"], "steps": len(value["steps"]), "required_inputs": value["required_inputs"]} for key, value in load_all(root).items()]
        elif args.command == "validate":
            blueprints = load_all(root)
            result = {"valid": True, "blueprints": len(blueprints), "steps": sum(len(item["steps"]) for item in blueprints.values())}
        elif args.command == "show":
            blueprint = load_all(root).get(args.blueprint_id)
            if blueprint is None:
                raise ValueError("unknown blueprint ID")
            result = blueprint
        elif args.command == "init":
            result = initialize(root, args.blueprint_id, args.run_dir, args.workspace_id, args.endpoint, args.input)
        else:
            result = recover(args.run_dir)
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0
    except (ValueError, OSError, KeyError, TypeError, json.JSONDecodeError) as error:
        print(json.dumps({"valid": False, "error_type": type(error).__name__, "message": str(error) if isinstance(error, ValueError) else "Task blueprint or journal could not be read"}, ensure_ascii=False))
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
