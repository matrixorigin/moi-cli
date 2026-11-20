#!/usr/bin/env python3
"""Read-only MOI readiness checks through the formal CLI; never read credentials."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
from urllib.parse import urlsplit


def run(argv, timeout):
    """Keep subprocess output in memory only. Reports contain allowlisted facts."""
    try:
        proc = subprocess.Popen(argv, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                text=True, errors="replace", start_new_session=True)
    except OSError:
        return {"reason": "process_start_failed"}, "", ""
    try:
        stdout, stderr = proc.communicate(timeout=timeout)
    except subprocess.TimeoutExpired:
        try:
            os.killpg(proc.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
        proc.communicate()
        return {"reason": "timeout"}, "", ""
    return {"exit_code": proc.returncode}, stdout, stderr


def failure(meta, stdout, stderr, auth):
    # Client-observed categories, not invented server business reasons.
    if "reason" in meta:
        return meta["reason"]
    text = (stdout + "\n" + stderr).lower()
    if "api-key is required" in text:
        return "pat_missing"
    if "credential unavailable" in text:
        return "native_credential_unavailable" if auth == "moi" else "credential_unavailable"
    if "unmarshal product response" in text or "unknown field" in text:
        return "cli_response_decode_failed"
    if "unknown command" in text or "flag provided but not defined" in text:
        return "cli_contract_missing"
    for code, reason in (("errunauthorized", "authentication_rejected"),
                         ("errforbidden", "permission_denied"),
                         ("errnotfound", "endpoint_or_resource_not_found"),
                         ("errapikeyrequired", "pat_missing"),
                         ("errcurrentroleinvalid", "role_configuration_invalid"),
                         ("errworkspacecurrentroleconflict", "role_configuration_conflict")):
        if "[" + code + "]" in text:
            return reason
    for number, reason in ((401, "authentication_rejected"), (403, "permission_denied"),
                           (404, "endpoint_or_resource_not_found"), (405, "method_rejected")):
        if re.search(r"\b(?:http(?: status)?|status(?: code)?)\s*[:=]?\s*" + str(number) + r"\b", text):
            return reason
    if re.search(r"\b(?:http(?: status)?|status(?: code)?)\s*[:=]?\s*5[0-9]{2}\b", text):
        return "service_error_response"
    if any(term in text for term in ("connection refused", "no such host", "certificate verify", "tls handshake")):
        return "network_or_tls_failed"
    return "cli_call_failed"


def executable(path):
    target = Path(path).expanduser().absolute()
    if not target.is_file():
        return str(target), "missing"
    if not os.access(target, os.X_OK):
        return str(target), "not_executable"
    return str(target), "passed"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--online", action="store_true", help="Make read-only Product API calls")
    parser.add_argument("--endpoint", help="Explicit target; never infer from credential/config files")
    parser.add_argument("--workspace-id", help="Previously selected real workspace ID")
    parser.add_argument("--auth", choices=("native", "pat", "both"), default="native")
    parser.add_argument("--timeout", type=float, default=90, help="Per-call seconds, 0 < value <= 120; native refresh may need 75 seconds")
    parser.add_argument("--skill-dir", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--expected-version", help="Package wrapper checks VERSION against this Skill")
    args = parser.parse_args()
    if not 0 < args.timeout <= 120:
        parser.error("timeout must be greater than zero and at most 120 seconds")
    checks = []

    def add(name, status, reason=None, **facts):
        check = {"check": name, "status": status, **facts}
        if reason:
            check["reason"] = reason
        checks.append(check)

    skill_dir = args.skill_dir.expanduser().absolute()
    version = None
    try:
        doc = (skill_dir / "SKILL.md").read_text()
        frontmatter = re.match(r"^---\n(.*?)\n---", doc, re.DOTALL)
        match = re.search(r"^\s+skill-version:\s*([0-9]+\.[0-9]+\.[0-9]+)\s*$",
                          frontmatter.group(1) if frontmatter else "", re.MULTILINE)
        version = match.group(1) if match else None
        refs = skill_dir / "references"
        if not version or not refs.is_dir():
            add("skill", "failed", "skill_metadata_or_references_missing")
        elif args.expected_version and version != args.expected_version:
            add("skill", "failed", "skill_version_mismatch")
        else:
            digest = hashlib.sha256()
            for path in sorted(p for p in refs.rglob("*") if p.is_file()):
                digest.update(path.relative_to(refs).as_posix().encode() + b"\0")
                digest.update(path.read_bytes() + b"\0")
            add("skill", "passed", reference_fingerprint=digest.hexdigest()[:16])
    except (OSError, UnicodeError):
        add("skill", "failed", "skill_unreadable")

    cli, cli_status = executable(os.environ.get("MOI_DEV_CLI_BIN") or "~/.local/bin/moi-cli-dev")
    add("cli", cli_status, None if cli_status == "passed" else "cli_" + cli_status, path=cli)
    if cli_status == "passed":
        meta, out, err = run([cli, "-version"], min(args.timeout, 5))
        match = re.fullmatch(r"moi-cli version ([0-9A-Za-z.+_-]+)\s*", out)
        ok = meta.get("exit_code") == 0 and match is not None
        add("cli_version", "passed" if ok else "failed", None if ok else "cli_version_unavailable",
            version=match.group(1) if ok else None)
        meta, out, err = run([cli, "-help"], min(args.timeout, 5))
        required = ("-endpoint", "-auth", "-workspace-id", "-o", "-no-interactive")
        ok = meta.get("exit_code") == 0 and all(re.search(r"(?m)^\s*" + re.escape(flag) + r"(?:\s|$)", out + err) for flag in required)
        add("cli_global_options", "passed" if ok else "failed", None if ok else "cli_contract_missing")

    # Native moi-cli resolves MOI_ASTRA_BINARY (or astra on PATH), not ASTRA_BIN.
    helper = os.environ.get("MOI_ASTRA_BINARY") or "astra"
    helper_path = shutil.which(helper)
    astra, astra_status = executable(helper_path or helper)
    if args.auth in ("native", "both"):
        if helper_path and not Path(helper_path).is_absolute():
            add("astra_helper_alignment", "failed", "astra_helper_path_not_absolute")
        requested = os.environ.get("ASTRA_BIN")
        if requested:
            requested_path = Path(requested).expanduser().absolute()
            if not helper_path or requested_path.resolve() != Path(helper_path).resolve():
                add("astra_helper_alignment", "failed", "astra_helper_path_mismatch")
        add("astra", astra_status, None if astra_status == "passed" else "astra_" + astra_status, path=astra)
        if astra_status == "passed":
            meta, out, err = run([astra, "auth", "status"], min(args.timeout, 5))
            text = (out + "\n" + err).lower()
            if meta.get("exit_code") == 0 and "signed in to moi" in text:
                add("native_login_hint", "passed", "local_status_only")
            elif "needs to be renewed" in text or "run astra login" in text:
                add("native_login_hint", "failed", "login_required_or_expired")
            else:
                add("native_login_hint", "unknown", meta.get("reason", "login_status_unrecognized"))
    else:
        add("astra", "not_required", "pat_does_not_require_native_login")

    # Login hints are advisory: only a real request proves the selected API path.
    core_ready = all(c["status"] == "passed" for c in checks if c["check"] not in ("native_login_hint", "astra", "astra_helper_alignment"))
    native_tools_ready = astra_status == "passed" and not any(c["check"] == "astra_helper_alignment" and c["status"] != "passed" for c in checks)
    local_ready = core_ready and (args.auth == "pat" or native_tools_ready)
    layers = {"local": "passed" if local_ready else "failed", "native_api": "not_checked",
              "pat_api": "not_checked", "native_workspace": "not_checked", "pat_workspace": "not_checked",
              "native_permissions": "not_checked", "pat_permissions": "not_checked"}
    target = {}
    if args.online:
        valid_endpoint = False
        if args.endpoint:
            try:
                url = urlsplit(args.endpoint)
                valid_endpoint = (url.scheme in ("http", "https") and bool(url.hostname) and
                                  url.username is None and url.password is None and
                                  not url.query and not url.fragment and
                                  not re.search(r"[\s\x00-\x1f]", args.endpoint))
                _ = url.port
            except ValueError:
                valid_endpoint = False
        add("endpoint", "passed" if valid_endpoint else "blocked",
            None if valid_endpoint else ("endpoint_required" if not args.endpoint else "endpoint_invalid"))
        if valid_endpoint:
            target["endpoint"] = args.endpoint
        if args.workspace_id and re.fullmatch(r"[A-Za-z0-9_-]{1,128}", args.workspace_id):
            target["workspace_id"] = args.workspace_id
        elif args.workspace_id:
            add("workspace_selection", "blocked", "workspace_id_invalid")
        else:
            add("workspace_selection", "blocked", "workspace_required_no_auto_selection")

        if core_ready and valid_endpoint:
            for label in (("native", "pat") if args.auth == "both" else (args.auth,)):
                if label == "native" and not native_tools_ready:
                    add("native_api", "blocked", "astra_unavailable")
                    layers["native_api"] = "blocked"
                    continue
                auth = "moi" if label == "native" else "pat"
                base = [cli, "-endpoint", args.endpoint, "-auth", auth, "-o", "-no-interactive"]

                def probe(name, command):
                    meta, out, err = run(base + command, args.timeout)
                    if meta.get("exit_code") != 0:
                        add(name, "failed", failure(meta, out, err, auth), exit_code=meta.get("exit_code"))
                        return None
                    try:
                        value = json.loads(out)
                    except ValueError:
                        add(name, "failed", "cli_response_not_json", **meta)
                        return None
                    if value is None:
                        add(name, "failed", "cli_response_empty", **meta)
                    return value

                # workspace get is an alias; use the real list response to prove discovery.
                data = probe(label + "_api", ["workspace", "list"])
                if data is None:
                    layers[label + "_api"] = "failed"
                    continue
                workspaces = data.get("workspaces") if isinstance(data, dict) else None
                if not isinstance(workspaces, list) or not all(isinstance(w, dict) and isinstance(w.get("id"), str) for w in workspaces):
                    add(label + "_api", "failed", "workspace_response_shape_changed")
                    layers[label + "_api"] = "failed"
                    continue
                layers[label + "_api"] = "passed"
                add(label + "_api", "passed", visible_workspace_count=len(workspaces))
                wid = target.get("workspace_id")
                if not wid:
                    continue
                matches = [w for w in workspaces if w["id"] == wid]
                if len(matches) != 1:
                    add(label + "_workspace", "blocked", "workspace_not_in_visible_list")
                    layers[label + "_workspace"] = "blocked"
                    continue
                access = matches[0].get("access_status")
                state = matches[0].get("status")
                if access not in (None, "", "accepted") or state not in (None, "", "ACTIVE"):
                    add(label + "_workspace", "blocked", "workspace_not_active_or_access_not_accepted")
                    layers[label + "_workspace"] = "blocked"
                    continue
                add(label + "_workspace", "passed", "visible_membership_only")
                layers[label + "_workspace"] = "passed"
                caps = probe(label + "_permissions", ["-workspace-id", wid, "iam", "current-principal-capabilities"])
                if caps is None:
                    layers[label + "_permissions"] = "failed"
                elif (not isinstance(caps, dict) or
                      not isinstance(caps.get("effective_role_id"), str) or not caps["effective_role_id"] or
                      not isinstance(caps.get("capability_version"), str) or not caps["capability_version"] or
                      not isinstance(caps.get("items", []), list)):
                    add(label + "_permissions", "failed", "capability_response_shape_changed")
                    layers[label + "_permissions"] = "failed"
                else:
                    # Do not infer write grants or expose identity/policy payloads.
                    add(label + "_permissions", "passed", "permission_endpoint_readable_only")
                    layers[label + "_permissions"] = "passed"

    problems = [c for c in checks if c["status"] in ("failed", "blocked", "missing", "not_executable", "unknown") and c["check"] != "native_login_hint"]
    login_hint_problem = any(c["check"] == "native_login_hint" and c["status"] != "passed" for c in checks)
    status = ("ready_for_read_checks" if not problems else "needs_attention") if args.online else ("local_ready" if local_ready and not login_hint_problem else "needs_attention")
    result = {"schema_version": 1, "status": status, "skill_version": version,
              "online_requested": args.online, "auth": args.auth, "target": target,
              "layers": layers, "checks": checks,
              "limits": ["只证明列出的读取入口；业务动作及其权限仍按目标领域核验。",
                         "未检查的认证通道不推断可用；PAT 检查不证明与 Astra 为同一账号。",
                         "不读取凭据文件，不输出原始响应，不自动登录、选工作区或修改配置。"]}
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if status in ("local_ready", "ready_for_read_checks") else 1


if __name__ == "__main__":
    raise SystemExit(main())
