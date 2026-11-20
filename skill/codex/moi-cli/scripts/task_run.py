#!/usr/bin/env python3
"""Private local task journal for explicit moi-cli calls. No replay or business inference."""
import argparse
from contextlib import contextmanager
from datetime import datetime, timezone
import fcntl
import hashlib
import json
import os
from pathlib import Path
import re
import sys
import tempfile

sys.dont_write_bytecode = True
ID = re.compile(r"[A-Za-z0-9][A-Za-z0-9._-]{0,63}\Z")
# A deliberately narrow set of formal observation commands. SQL run/execute and
# refresh/retry/repair are excluded even if an operator labels them 'read'.
# Each observation binds a canonical object type and the real CLI identity flag.
PROBES = {
    ("catalog", "database-info"): ("database", "database-id"),
    ("catalog", "table-info"): ("table", "table-id"),
    ("catalog", "volume-info"): ("volume", "volume-id"),
    ("catalog", "file-info"): ("file", "file-id"),
    ("catalog", "file-data-asset"): ("file", "file-id"),
    ("workflow", "get"): ("workflow", "workflow-id"),
    ("workflow", "get-run"): ("execution", "execution-id"),
    ("workflow", "get-run-result"): ("execution", "execution-id"),
    ("workflow", "get-run-node"): ("execution", "execution-id"),
    ("workflow", "catalog-file-lineage-overview"): ("file", "file-id"),
    ("sql", "query-result"): ("statement", "statement-id"),
    ("dashboard", "get"): ("dashboard", "dashboard-id"),
    ("agent", "get-task"): ("task", "task-id"),
    ("semantic-model", "get"): ("model", "model-id"),
    ("semantic-model", "get-source-document"): ("source", "source-id"),
}



def now():
    return datetime.now(timezone.utc).isoformat()


def valid_id(value):
    if not isinstance(value, str) or not ID.fullmatch(value):
        raise ValueError("IDs must be 1-64 safe filename characters")
    return value


def digest(path):
    if path.is_symlink() or not path.is_file():
        raise ValueError("evidence must be a regular file")
    return hashlib.sha256(path.read_bytes()).hexdigest()


def safe_metadata(value):
    # Metadata must never become an alternative credential store. This is a
    # known-pattern rejection, not a classifier for arbitrary private prose.
    raw = json.dumps(value, ensure_ascii=False)
    if re.search(r"sk-uca-v1-|Bearer\s+\S+|[a-z]+://[^\s/]+:[^\s/]+@", raw, re.I):
        raise ValueError("credential-like metadata rejected")
    if isinstance(value, dict):
        for key, child in value.items():
            if re.sub(r"[^a-z]", "", key.lower()) in {"password", "token", "apikey", "pat", "cookie", "secret", "authorization", "accesstoken", "refreshtoken"}:
                raise ValueError("credential fields are not task metadata")
            safe_metadata(child)
    elif isinstance(value, list):
        for child in value:
            safe_metadata(child)


def atomic_json(path, data):
    safe_metadata(data)
    if path.is_symlink():
        raise ValueError("journal symlink rejected")
    fd, name = tempfile.mkstemp(prefix=".journal-", dir=path.parent)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd, "w", encoding="utf-8") as stream:
            json.dump(data, stream, ensure_ascii=False, indent=2)
            stream.write("\n")
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(name, path)
        directory = os.open(path.parent, os.O_RDONLY)
        try:
            os.fsync(directory)
        finally:
            os.close(directory)
    finally:
        if os.path.exists(name):
            os.unlink(name)


@contextmanager
def journal(directory):
    directory = Path(directory).absolute()
    if directory.is_symlink() or not directory.is_dir():
        raise ValueError("run directory must already exist and must not be a symlink")
    directory = directory.resolve()
    if directory.stat().st_mode & 0o077:
        raise ValueError("run directory must be private (0700)")
    fd = os.open(directory / ".lock", os.O_RDWR | os.O_CREAT | os.O_NOFOLLOW, 0o600)
    try:
        fcntl.flock(fd, fcntl.LOCK_EX)
        path = directory / "run.json"
        if path.is_symlink():
            raise ValueError("journal symlink rejected")
        data = json.loads(path.read_text())
        if data.get("schema_version") != 1:
            raise ValueError("unsupported task journal schema")
        yield directory, data
    finally:
        fcntl.flock(fd, fcntl.LOCK_UN)
        os.close(fd)


def event(data, kind, **fields):
    data["events"].append({"at": now(), "kind": kind, **fields})
    data["updated_at"] = now()


def init(directory, workspace_id, endpoint, goal):
    from urllib.parse import urlsplit
    u = urlsplit(endpoint)
    if u.scheme not in {"http", "https"} or not u.netloc or u.username or u.password or u.query or u.fragment:
        raise ValueError("endpoint must be an explicit HTTP(S) product endpoint without credentials/query")
    if not workspace_id or not goal.strip():
        raise ValueError("workspace and goal required")
    directory = Path(directory).absolute()
    # No takeover of an existing directory, including an interrupted init.
    safe_metadata({"workspace_id": workspace_id, "endpoint": endpoint, "goal": goal})
    directory.mkdir(mode=0o700, parents=False, exist_ok=False)
    os.chmod(directory, 0o700)
    data = {"schema_version": 1, "created_at": now(), "updated_at": now(),
            "goal": goal, "scope": {"workspace_id": workspace_id, "endpoint": endpoint.rstrip("/")},
            "steps": {}, "calls": {}, "references": {}, "events": []}
    event(data, "initialized")
    atomic_json(directory / "run.json", data)
    return {"run_dir": str(directory.resolve()), "scope": data["scope"], "state": "planned"}


def add_step(directory, step_id, capability_id, intent, effect, acceptance, dependencies):
    valid_id(step_id)
    valid_id(capability_id)
    cards = json.loads((Path(__file__).resolve().parents[1] / "references/capability-cards.json").read_text())
    if capability_id not in {c["capability_id"] for c in cards["cards"]}:
        raise ValueError("unknown capability ID; select the documented business capability")
    if effect not in {"read", "write"}:
        raise ValueError("effect must be read or write")
    if not intent.strip() or not acceptance or any(not x.strip() for x in acceptance):
        raise ValueError("intent and explicit acceptance criteria required")
    with journal(directory) as (root, data):
        if step_id in data["steps"]:
            raise ValueError("step ID already exists")
        if len(dependencies) != len(set(dependencies)) or any(d not in data["steps"] for d in dependencies):
            raise ValueError("dependencies must reference existing distinct steps")
        data["steps"][step_id] = {"capability_id": capability_id, "intent": intent, "effect": effect,
            "depends_on": dependencies, "acceptance": {f"a{i}": text for i, text in enumerate(acceptance, 1)},
            "state": "planned", "primary_invocation": None, "checks": {}}
        event(data, "step_added", step_id=step_id)
        atomic_json(root / "run.json", data)
    return {"step_id": step_id, "state": "planned"}


def inside(root, value):
    path = Path(value)
    if not path.is_absolute():
        path = root / path
    if path.is_symlink():
        raise ValueError("evidence symlink rejected")
    resolved = path.resolve(strict=True)
    if not resolved.is_relative_to(root / "evidence") or not resolved.is_file():
        raise ValueError("evidence must be a file inside this run's evidence directory")
    return resolved


def call_receipt(root, data, invocation_id):
    call = data["calls"][invocation_id]
    if call.get("transport_status") != "response_received":
        raise ValueError("this invocation has no recorded response; observe the original object")
    for field in ["record", "request", "response"]:
        item = call[field]
        p = inside(root, item["path"])
        if digest(p) != item["sha256"]:
            raise ValueError("evidence changed; do not reuse old conclusions")
    from record_cli import verify_capture
    verified = verify_capture(inside(root, call["record"]["path"]))
    if any(verified[k] != data["scope"][k] for k in ["workspace_id", "endpoint"]):
        raise ValueError("receipt scope differs from task scope")
    return call


def flag_values(args, wanted):
    values = []
    for i, arg in enumerate(args):
        flag, eq, value = arg.partition("=")
        if flag.lstrip("-") == wanted:
            if not eq and i + 1 >= len(args):
                raise ValueError("missing identity flag value")
            values.append(value if eq else args[i + 1])
    return values


def invoke(directory, step_id, invocation_id, args, timeout, probe=False, expected_workspace=None, expected_endpoint=None, reference=None):
    from record_cli import capture
    valid_id(invocation_id)
    if args and args[0] == "--":
        args = args[1:]
    if len(args) < 2:
        raise ValueError("one formal domain command required")
    if probe and tuple(args[:2]) not in PROBES:
        raise ValueError("probe only supports listed observation commands; it cannot submit/retry SQL or writes")
    with journal(directory) as (root, data):
        validate_evidence(root, data)
        scope = data["scope"]
        if reference is not None:
            ref = data["references"][reference]
            if ref["workspace_id"] != scope["workspace_id"]:
                raise ValueError("reference workspace mismatch")
        if probe:
            if reference is None:
                raise ValueError("probe requires a saved typed reference")
            object_type, identity_flag = PROBES[tuple(args[:2])]
            if ref["object_type"] != object_type or flag_values(args[2:], identity_flag) != [str(ref["id"])]:
                raise ValueError("probe object type or identity differs from saved reference")
        elif reference is not None:
            # For an uncertain mutation of an existing object, freeze the object
            # before sending. No name guessing or post-hoc replacement.
            if not any(flag_values(args[2:], f) == [str(ref["id"])] for f in {v[1] for v in PROBES.values() if v[0] == ref["object_type"]}):
                raise ValueError("primary command does not explicitly target saved reference")
        if expected_workspace and expected_workspace != scope["workspace_id"]:
            raise ValueError("workspace differs from saved task scope")
        if expected_endpoint and expected_endpoint.rstrip("/") != scope["endpoint"]:
            raise ValueError("endpoint differs from saved task scope")
        step = data["steps"][step_id]
        if step["state"] == "verified":
            raise ValueError("verified step is frozen; later observations require a new explicit step")
        if invocation_id in data["calls"]:
            raise ValueError("invocation already exists; inspect its original record")
        if probe:
            if not step["primary_invocation"]:
                raise ValueError("probe requires an original invocation")
        else:
            if step["primary_invocation"] or step["state"] != "planned":
                raise ValueError("step already submitted; no automatic replay, even after timeout/error")
            if any(data["steps"][d]["state"] != "verified" for d in step["depends_on"]):
                raise ValueError("dependencies not verified")
            step["primary_invocation"] = invocation_id
            step["state"] = "submitted"
        data["calls"][invocation_id] = {"step_id": step_id, "kind": "observation" if probe else "primary", "transport_status": "unknown", "reserved_at": now(), "target_reference": reference}
        event(data, "invocation_reserved", step_id=step_id, invocation_id=invocation_id)
        atomic_json(root / "run.json", data)
    # Reservation is durable before subprocess. A second process cannot submit
    # the same step while the CLI is running. A crash never resets it to planned.
    try:
        result = capture(evidence_dir=root / "evidence", invocation_id=invocation_id,
                         workspace_id=scope["workspace_id"], endpoint=scope["endpoint"], args=args, timeout=timeout)
    except BaseException:
        # May have failed before or after remote acceptance. Keep uncertainty.
        with journal(root) as (root, data):
            if not probe:
                data["steps"][step_id]["state"] = "unknown"
            event(data, "capture_interrupted", invocation_id=invocation_id)
            atomic_json(root / "run.json", data)
        raise
    with journal(root) as (root, data):
        call = data["calls"][invocation_id]
        for key, result_key in [("record", "record_path"), ("request", "request_path"), ("response", "response_path")]:
            p = inside(root, result[result_key])
            call[key] = {"path": str(p.relative_to(root)), "sha256": digest(p)}
        call.update(transport_status=result["transport_status"], exit_code=result["exit_code"])
        if not probe:
            data["steps"][step_id]["state"] = "observed" if result["transport_status"] == "response_received" else "unknown"
        event(data, "invocation_recorded", invocation_id=invocation_id)
        atomic_json(root / "run.json", data)
    return {"invocation_id": invocation_id, "transport_status": result["transport_status"], "exit_code": result["exit_code"], "business_state": "not_inferred"}


def pointer(document, path):
    if path == "":
        return document
    if not path.startswith("/"):
        raise ValueError("use a JSON pointer into the recorded CLI response")
    value = document
    for token in path[1:].split("/"):
        token = token.replace("~1", "/").replace("~0", "~")
        value = value[int(token)] if isinstance(value, list) else value[token]
    return value


def response_value(root, data, invocation_id, stream, path):
    call = call_receipt(root, data, invocation_id)
    response = json.loads(inside(root, call["response"]["path"]).read_text())
    return pointer(json.loads(response[stream]), path)


def record_reference(directory, name, invocation_id, object_type, path, version_path=None):
    valid_id(name)
    if not object_type.strip():
        raise ValueError("object_type required")
    with journal(directory) as (root, data):
        value = response_value(root, data, invocation_id, "stdout", path)
        if type(value) not in {str, int} or value == "":
            raise ValueError("reference must be a returned nonempty string or integer ID")
        version = response_value(root, data, invocation_id, "stdout", version_path) if version_path is not None else None
        ref = {"workspace_id": data["scope"]["workspace_id"], "object_type": object_type, "id": value,
               "version": version, "source": {"invocation_id": invocation_id, "pointer": path, "version_pointer": version_path}}
        if name in data["references"]:
            raise ValueError("reference name already exists; changed objects/versions require a new explicit name")
        data["references"][name] = ref
        event(data, "reference_recorded", name=name)
        atomic_json(root / "run.json", data)
    return {"name": name, "reference": ref}


def check(directory, step_id, criterion, invocation_id, stream, path, expected):
    safe_metadata(expected)
    with journal(directory) as (root, data):
        step = data["steps"][step_id]
        if step["state"] == "verified":
            raise ValueError("verified step is frozen; later checks require a new explicit step")
        if criterion not in step["acceptance"]:
            raise ValueError("unknown acceptance criterion")
        history = step["checks"].get(criterion, [])
        if any(c["invocation_id"] == invocation_id for c in history):
            raise ValueError("this response was already checked; preserve it")
        if history and (type(expected) is not type(history[0]["expected"]) or expected != history[0]["expected"]):
            raise ValueError("acceptance expectation changed; use a new explicit task step")
        if data["calls"][invocation_id]["step_id"] != step_id:
            raise ValueError("evidence belongs to another step")
        actual = response_value(root, data, invocation_id, stream, path)
        passed = type(actual) is type(expected) and actual == expected
        step["checks"].setdefault(criterion, []).append({"invocation_id": invocation_id, "stream": stream, "pointer": path,
                                    "expected": expected, "passed": passed})
        event(data, "criterion_checked", step_id=step_id, criterion=criterion, passed=passed)
        atomic_json(root / "run.json", data)
    return {"criterion": criterion, "passed": passed}


def validate_evidence(root, data):
    for invocation, call in data["calls"].items():
        if call.get("record"):
            # Unknown responses can still have persisted timeout evidence.
            for item in [call[k] for k in ["record", "request", "response"]]:
                if digest(inside(root, item["path"])) != item["sha256"]:
                    raise ValueError("recorded evidence was changed")
            from record_cli import verify_capture
            verified = verify_capture(inside(root, call["record"]["path"]))
            if verified["invocation_id"] != invocation or any(verified[k] != data["scope"][k] for k in ["workspace_id", "endpoint"]):
                raise ValueError("receipt scope or invocation differs from task")
    for ref in data["references"].values():
        source = ref["source"]
        value = response_value(root, data, source["invocation_id"], "stdout", source["pointer"])
        if ref["workspace_id"] != data["scope"]["workspace_id"] or type(value) is not type(ref["id"]) or value != ref["id"]:
            raise ValueError("reference does not match recorded source or scope")
        if source["version_pointer"] is not None:
            version = response_value(root, data, source["invocation_id"], "stdout", source["version_pointer"])
            if type(version) is not type(ref["version"]) or version != ref["version"]:
                raise ValueError("reference version does not match recorded source")
    for step_id, step in data["steps"].items():
        for history in step["checks"].values():
            for criterion in history:
                if data["calls"][criterion["invocation_id"]]["step_id"] != step_id:
                    raise ValueError("stored check uses evidence of another step")
                actual = response_value(root, data, criterion["invocation_id"], criterion["stream"], criterion["pointer"])
                if criterion["passed"] != (type(actual) is type(criterion["expected"]) and actual == criterion["expected"]):
                    raise ValueError("stored check differs from source evidence")


def finish(directory, step_id):
    with journal(directory) as (root, data):
        validate_evidence(root, data)
        step = data["steps"][step_id]
        if not step["primary_invocation"] or not step["checks"] or set(step["checks"]) != set(step["acceptance"]):
            raise ValueError("every declared acceptance criterion needs actual response evidence")
        if not all(history[-1]["passed"] for history in step["checks"].values()):
            raise ValueError("failed acceptance evidence cannot be marked verified")
        if any(data["steps"][d]["state"] != "verified" for d in step["depends_on"]):
            raise ValueError("dependencies not verified")
        # Unknown primary writes require a later observation. A task journal
        # cannot make an uncertain submission safe just by changing its label.
        primary = data["calls"][step["primary_invocation"]]
        if primary["transport_status"] == "unknown":
            target = primary.get("target_reference")
            if target is None or not all(data["calls"][history[-1]["invocation_id"]]["kind"] == "observation" and data["calls"][history[-1]["invocation_id"]].get("target_reference") == target for history in step["checks"].values()):
                raise ValueError("unknown submission requires latest evidence for its saved original object; no original ID means unresolved")
        step["state"] = "verified"
        event(data, "step_verified", step_id=step_id)
        atomic_json(root / "run.json", data)
    return {"step_id": step_id, "state": "verified", "scope": "Only the explicit saved acceptance criteria; not a general business correctness guarantee."}


def status(directory):
    with journal(directory) as (root, data):
        validate_evidence(root, data)
        uncertain = [key for key, value in data["steps"].items() if value["state"] in {"unknown", "submitted"}]
        complete = bool(data["steps"]) and all(s["state"] == "verified" for s in data["steps"].values())
        return {"scope": data["scope"], "goal": data["goal"], "state": "completed" if complete else "incomplete",
                "steps": {key: {"state": s["state"], "primary_invocation": s["primary_invocation"], "depends_on": s["depends_on"], "criteria": s["acceptance"], "checks": s["checks"]} for key, s in data["steps"].items()},
                "references": data["references"], "requires_original_object_observation": uncertain,
                "resume_policy": "No CLI calls here. Inspect saved receipts and typed references; observe uncertain original objects through formal read commands. Never resubmit an existing step automatically."}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    p = sub.add_parser("init")
    p.add_argument("--run-dir", required=True)
    p.add_argument("--workspace-id", required=True)
    p.add_argument("--endpoint", required=True)
    p.add_argument("--goal", required=True)
    for name in ["add-step", "execute", "probe", "reference", "check", "finish", "resume", "status"]:
        p = sub.add_parser(name)
        p.add_argument("--run-dir", required=True)
        if name in {"add-step", "execute", "probe", "check", "finish"}:
            p.add_argument("--step-id", required=True)
        if name == "add-step":
            p.add_argument("--capability-id", required=True)
            p.add_argument("--intent", required=True)
            p.add_argument("--effect", choices=["read", "write"], required=True)
            p.add_argument("--acceptance", action="append", required=True)
            p.add_argument("--depends-on", action="append", default=[])
        if name in {"execute", "probe"}:
            p.add_argument("--invocation-id", required=True)
            p.add_argument("--timeout", type=float, default=60)
            p.add_argument("--expect-workspace")
            p.add_argument("--expect-endpoint")
            p.add_argument("--reference", help="Saved typed reference; required for probe, optional existing target for primary")
            p.add_argument("args", nargs=argparse.REMAINDER)
        if name == "reference":
            p.add_argument("--name", required=True)
            p.add_argument("--invocation-id", required=True)
            p.add_argument("--object-type", required=True)
            p.add_argument("--pointer", required=True)
            p.add_argument("--version-pointer")
        if name == "check":
            p.add_argument("--criterion", required=True)
            p.add_argument("--invocation-id", required=True)
            p.add_argument("--stream", choices=["stdout", "stderr"], default="stdout")
            p.add_argument("--pointer", required=True)
            p.add_argument("--equals-json", required=True)
    a = parser.parse_args()
    try:
        if a.command == "init":
            result = init(a.run_dir, a.workspace_id, a.endpoint, a.goal)
        elif a.command == "add-step":
            result = add_step(a.run_dir, a.step_id, a.capability_id, a.intent, a.effect, a.acceptance, a.depends_on)
        elif a.command in {"execute", "probe"}:
            result = invoke(a.run_dir, a.step_id, a.invocation_id, a.args, a.timeout, a.command == "probe", a.expect_workspace, a.expect_endpoint, a.reference)
        elif a.command == "reference":
            result = record_reference(a.run_dir, a.name, a.invocation_id, a.object_type, a.pointer, a.version_pointer)
        elif a.command == "check":
            result = check(a.run_dir, a.step_id, a.criterion, a.invocation_id, a.stream, a.pointer, json.loads(a.equals_json))
        elif a.command == "finish":
            result = finish(a.run_dir, a.step_id)
        else:
            result = status(a.run_dir)
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 2 if result.get("passed") is False else 0
    except (ValueError, OSError, KeyError, IndexError, TypeError) as error:
        # Do not echo raw input or arbitrary CLI output in local error messages.
        print(json.dumps({"state": "local_error", "error_type": type(error).__name__, "message": str(error) if isinstance(error, ValueError) else "Task file, reference or evidence could not be read; inspect local records."}, ensure_ascii=False))
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
