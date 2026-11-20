#!/usr/bin/env python3
"""Capture ONE native moi-cli invocation; no retries, login or business verdicts.

capture() is the supported module interface. Each invocation reserves a new private
directory. Input JSON/SQL bytes are snapshotted and bound to request.json BEFORE
spawn; the CLI reads those snapshots, not the original paths. Inline --json and
--query become their formal --json-file / --query-file counterparts. Only pass
commands whose help supports those counterparts. --validate-file is JSON input.
Registered Workflow/Knowledge JSON inputs are copied as structured UTF-8 files.
Explicit --file-path / --file / --query-image inputs are copied as opaque bytes:
this can retain private business content and requires the caller's upload/retention scope.
This helper does not inspect opaque files or claim that they are secret-free.

Evidence is exclusive-create, directories 0700 / files 0600, not WORM. A request
left in 'started' without a final record is UNKNOWN, even if spawn was not seen.
Response stdout/stderr are redacted UTF-8 text. Their digests cover ONLY saved
redacted text, never original process output or wire bytes. Redaction covers common
credential fields/patterns, not arbitrary secrets embedded in business prose.
--output / --output-file paths remain CLI-managed, unredacted external artifacts;
they are recorded but not captured, hashed, or claimed as protected evidence.

Preflight errors raise ValueError/OSError without executing a command. Once started,
timeouts, signals and interruptions are unknown; a normal nonzero CLI exit is
recorded as response_received and is NOT interpreted as business failure.
"""
from __future__ import annotations

import argparse
from collections import Counter
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import re
import signal
import stat
import subprocess
import sys
from urllib.parse import urlsplit


_STRUCTURED_FILES = {
    "--json-file": ".json",
    "--query-file": ".sql",
    "--validate-file": ".json",
    "--acceptance-file": ".json",
    "--workflow-file": ".json",
    "--runtime-fields-file": ".json",
    "--runtime-layout-file": ".json",
    "--default-values-file": ".json",
    "--values-file": ".json",
    "--config-override-file": ".json",
    "--volume-trigger-file": ".json",
    "--dynamic-service-file": ".json",
    "--files-file": ".json",
    "--tables-file": ".json",
    "--source-selections-file": ".json",
}
_OPAQUE_FILES = {"--file-path", "--file", "--query-image"}
_INLINE = {"--json": ("--json-file", ".json"), "--query": ("--query-file", ".sql")}
_OUTPUTS = {"--output", "--output-file"}
_CREDENTIAL_FILES = {"--api-key-file", "--runtime-token-file", "--actor-assertion-file", "--provisioner-token-file"}
_GLOBAL_OVERRIDES = {"endpoint", "workspaceid", "auth", "apikey", "config", "configfile", "profile", "nointeractive", "o"}
_SECRET_KEYS = {
    "password", "passwd", "pwd", "token", "accesstoken", "refreshtoken", "idtoken",
    "authtoken", "bearertoken", "apikey", "secret", "clientsecret", "privatekey",
    "authorization", "cookie", "setcookie", "pat", "credential", "credentials",
}
_KEY_PATTERN = r"(?:password|passwd|pwd|access[_-]?token|refresh[_-]?token|id[_-]?token|auth[_-]?token|bearer[_-]?token|api[_-]?key|client[_-]?secret|private[_-]?key|authorization|set[_-]?cookie|cookie|secret|token|pat)"
_TEXT_RULES = [
    ("private_key", re.compile(r"-----BEGIN (?:[A-Z ]+ )?PRIVATE KEY-----.*?(?:-----END (?:[A-Z ]+ )?PRIVATE KEY-----|\Z)", re.S)),
    ("bearer", re.compile(r"\bBearer\s+[A-Za-z0-9._~+/=-]+", re.I)),
    ("secret_token", re.compile(r"\b(?:sk-(?:uca-v\d+-)?[A-Za-z0-9_-]{12,}|github_pat_[A-Za-z0-9_]{12,}|gh[pousr]_[A-Za-z0-9]{12,}|xox[baprs]-[A-Za-z0-9-]{12,})")),
    ("jwt", re.compile(r"\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b")),
    ("credential_assignment", re.compile(r"\b" + _KEY_PATTERN + r"[\"']?\s*[:=]\s*(?:\"[^\"]*\"|'[^']*'|[^\s,;}\]]+)", re.I)),
    ("url_userinfo", re.compile(r"https?://[^\s/@]+:[^\s/@]+@", re.I)),
]


def _now() -> str:
    return datetime.now(timezone.utc).isoformat()


def _digest(raw: bytes) -> str:
    return hashlib.sha256(raw).hexdigest()


def _key(value: str) -> str:
    return re.sub(r"[^a-z0-9]", "", value.lower())


def _redact_text(text: str) -> tuple[str, dict]:
    counts: Counter = Counter()

    def plain(value: str) -> str:
        for name, pattern in _TEXT_RULES:
            value, n = pattern.subn("[REDACTED]", value)
            counts[name] += n
        return value

    def walk(value):
        if isinstance(value, dict):
            result = {}
            for k, v in value.items():
                safe_key = plain(k)
                if _key(k) in _SECRET_KEYS and v not in (None, ""):
                    result[safe_key] = "[REDACTED]"
                    counts["credential_field"] += 1
                else:
                    result[safe_key] = walk(v)
            return result
        if isinstance(value, list):
            return [walk(v) for v in value]
        return plain(value) if isinstance(value, str) else value

    try:
        parsed = json.loads(text)
    except (ValueError, RecursionError):
        safe = plain(text)
    else:
        safe_object = walk(parsed)
        safe = json.dumps(safe_object, ensure_ascii=False, indent=2) + "\n" if any(counts.values()) else text
    return safe, {k: n for k, n in counts.items() if n}


def _reject_secret(text: str) -> None:
    _, matches = _redact_text(text)
    if matches:
        raise ValueError("Credential-like input refused; do not pass secrets to the evidence recorder")


def _absolute(path: Path) -> Path:
    return Path(os.path.abspath(os.fspath(path)))


def _dir_fd(path: Path, *, create: bool = False) -> int:
    """Walk canonical directory components with NOFOLLOW.

    Resolve directory aliases first so macOS /var and /tmp work. The final input
    file is still opened separately with NOFOLLOW and must be a regular file.
    """
    path = Path(os.path.abspath(os.fspath(path))).resolve(strict=False)
    fd = os.open(path.anchor, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        for part in path.parts[1:]:
            if create:
                try:
                    os.mkdir(part, 0o700, dir_fd=fd)
                except FileExistsError:
                    pass
            child = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=fd)
            os.close(fd)
            fd = child
        return fd
    except BaseException:
        os.close(fd)
        raise


def _read_input(path: Path) -> bytes:
    parent = _dir_fd(path.parent)
    fd = None
    try:
        fd = os.open(path.name, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=parent)
        original = os.fstat(fd)
        if not stat.S_ISREG(original.st_mode):
            raise ValueError("Input must be a regular file, never a symlink, device or pipe")
        chunks = []
        while True:
            chunk = os.read(fd, 1024 * 1024)
            if not chunk:
                break
            chunks.append(chunk)
        final = os.fstat(fd)
        raw = b"".join(chunks)
        if (original.st_size, original.st_mtime_ns, original.st_ctime_ns) != (final.st_size, final.st_mtime_ns, final.st_ctime_ns) or len(raw) != final.st_size:
            raise ValueError("Input changed while being captured; nothing was executed")
        return raw
    finally:
        if fd is not None:
            os.close(fd)
        os.close(parent)


def _write(fd: int, name: str, raw: bytes) -> None:
    out = os.open(name, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600, dir_fd=fd)
    os.fchmod(out, 0o600)
    with os.fdopen(out, "wb") as stream:
        stream.write(raw)
        stream.flush()
        os.fsync(stream.fileno())
    os.fsync(fd)


def _json_bytes(value: dict) -> bytes:
    return (json.dumps(value, ensure_ascii=False, indent=2, allow_nan=False) + "\n").encode("utf-8")


def _prepare(args: list[str]) -> tuple[list, list[dict], list[dict]]:
    if not isinstance(args, list) or not args or not all(isinstance(v, str) and "\0" not in v for v in args):
        raise ValueError("args must be a nonempty list of strings without NUL")
    if len(args) >= 2 and args[0] == "config" and args[1] == "set-api-key":
        raise ValueError("Credential configuration commands cannot be captured")
    plans, inputs, outputs = [], [], []
    i = 0
    while i < len(args):
        arg = args[i]
        flag, eq, attached = arg.partition("=")
        if flag.startswith("-") and (_key(flag) in _GLOBAL_OVERRIDES or _key(flag) in _SECRET_KEYS):
            raise ValueError("Scope, auth, config and credential flags cannot override recorder globals")
        if flag in _CREDENTIAL_FILES:
            raise ValueError("Credential file flags cannot be captured; use the formal secret-specific flow")
        _reject_secret(arg)
        if flag in _STRUCTURED_FILES or flag in _OPAQUE_FILES or flag in _INLINE or flag in _OUTPUTS:
            if eq:
                value = attached
            else:
                i += 1
                if i >= len(args):
                    raise ValueError("Missing value for file or inline input flag")
                value = args[i]
            if not value or "\0" in value:
                raise ValueError("Empty file or inline input is not supported")
            _reject_secret(value)
            if flag in _OUTPUTS:
                path = _absolute(Path(value))
                parent = _dir_fd(path.parent)
                try:
                    try:
                        os.stat(path.name, dir_fd=parent, follow_symlinks=False)
                    except FileNotFoundError:
                        pass
                    else:
                        raise FileExistsError("External output exists; refusing overwrite")
                finally:
                    os.close(parent)
                outputs.append({"flag": flag, "path": str(path), "managed": False, "redacted": False, "captured_or_hashed": False, "note": "CLI-managed external artifact; no exclusive creation or content protection guarantee"})
                plans.extend([flag, value])
            else:
                if flag in _INLINE:
                    executed_flag, extension = _INLINE[flag]
                    raw, source, opaque = value.encode("utf-8"), {"kind": "inline", "flag": flag}, False
                else:
                    path = _absolute(Path(value))
                    raw = _read_input(path)
                    executed_flag, opaque = flag, flag in _OPAQUE_FILES
                    extension = path.suffix if opaque and re.fullmatch(r"\.[A-Za-z0-9]{1,12}", path.suffix) else _STRUCTURED_FILES.get(flag, ".bin")
                    source = {"kind": "file", "path": str(path), "flag": flag}
                if not opaque:
                    try:
                        text = raw.decode("utf-8")
                    except UnicodeDecodeError:
                        raise ValueError("JSON/SQL input must be UTF-8") from None
                    _reject_secret(text)
                n = len(inputs)
                inputs.append({"source": source, "flag": executed_flag, "extension": extension, "bytes": raw, "opaque": opaque})
                plans.extend([executed_flag, n])
        else:
            if flag.startswith("-") and (flag.endswith("-file") or flag.endswith("-file-path")):
                raise ValueError("Unsupported file flag; its input would not be captured")
            plans.append(arg)
        i += 1
    return plans, inputs, outputs


def capture(*, evidence_dir: Path, invocation_id: str, workspace_id: str, endpoint: str,
            args: list[str], timeout: float = 60.0, cli_bin: Path | None = None) -> dict:
    """Return paths/digests, exit_code and transport_status; never a business verdict.

    response_path JSON always contains stdout/stderr strings (redacted), exit_code
    and transport_status. record_path binds request_path/SHA and response_path/SHA.
    Use a NEW ID for an authorized new intention, never to retry unknown submissions.
    """
    if not isinstance(invocation_id, str) or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_-]{0,95}", invocation_id):
        raise ValueError("invocation_id must be 1..96 safe ASCII identifier characters")
    if not isinstance(workspace_id, str) or not workspace_id.strip() or workspace_id != workspace_id.strip() or any(c.isspace() for c in workspace_id):
        raise ValueError("Explicit workspace_id is required")
    if not isinstance(endpoint, str):
        raise ValueError("Explicit HTTP(S) endpoint is required")
    parsed = urlsplit(endpoint)
    if parsed.scheme not in {"http", "https"} or not parsed.hostname or parsed.username is not None or parsed.password is not None or parsed.query or parsed.fragment or any(c.isspace() for c in endpoint):
        raise ValueError("Endpoint must be an explicit HTTP(S) URL without credentials/query/fragment")
    if isinstance(timeout, bool) or not isinstance(timeout, (int, float)) or not math.isfinite(timeout) or timeout <= 0:
        raise ValueError("timeout must be positive and finite")
    _reject_secret(workspace_id)
    _reject_secret(endpoint)
    plans, inputs, outputs = _prepare(args)
    cli = _absolute(cli_bin if cli_bin is not None else Path(os.environ.get("MOI_DEV_CLI_BIN", str(Path.home() / ".local/bin/moi-cli-dev"))))
    _reject_secret(str(cli))
    root = _absolute(evidence_dir)
    root_fd = _dir_fd(root, create=True)
    fd = None
    try:
        os.fchmod(root_fd, 0o700)
        os.mkdir(invocation_id, 0o700, dir_fd=root_fd)  # never reuse even an incomplete prior invocation
        fd = os.open(invocation_id, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=root_fd)
        directory = root / invocation_id
        frozen = []
        for n, item in enumerate(inputs):
            name = f"input-{n:03d}{item['extension']}"
            _write(fd, name, item["bytes"])
            frozen.append({"source": item["source"], "flag": item["flag"], "snapshot_path": str(directory / name), "sha256": _digest(item["bytes"]), "byte_count": len(item["bytes"]), "opaque": item["opaque"]})
        final_args = [str(directory / f"input-{v:03d}{inputs[v]['extension']}") if isinstance(v, int) else v for v in plans]
        argv = [str(cli), "-o", "-no-interactive", "-endpoint", endpoint, "-workspace-id", workspace_id] + final_args
        request = {"schema_version": 1, "invocation_id": invocation_id, "captured_at": _now(), "status": "started", "workspace_id": workspace_id, "endpoint": endpoint, "auth": "native_default_no_override", "argv": argv, "inputs": frozen, "external_outputs": outputs, "submission_note": "Persisted before local spawn. No final record means unknown; do not resubmit.", "integrity_scope": "Saved input bytes and argv; exclusive-create private files, not WORM or captured HTTP wire", "input_capture_scope": "Registered structured JSON/SQL and opaque file inputs only; other arguments are argv, not recursive directory/config snapshots; credential files are refused"}
        request_raw = _json_bytes(request)
        request_sha = _digest(request_raw)
        _write(fd, "request.json", request_raw)
        started_at = _now()
        stdout, stderr, exit_code, reason = b"", b"", None, None
        proc = None
        try:
            proc = subprocess.Popen(argv, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=False, start_new_session=True)
            stdout, stderr = proc.communicate(timeout=timeout)
            exit_code = proc.returncode
            if exit_code < 0:
                reason = "local_process_signal"
        except (subprocess.TimeoutExpired, KeyboardInterrupt) as error:
            reason = "local_timeout" if isinstance(error, subprocess.TimeoutExpired) else "local_interruption"
            if proc is not None:
                try:
                    os.killpg(proc.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                try:
                    stdout, stderr = proc.communicate(timeout=1.0)
                except subprocess.TimeoutExpired as cleanup_error:
                    stdout = cleanup_error.output or b""
                    stderr = cleanup_error.stderr or b""
                    # Escaped descendants may still hold pipes; never wait forever.
                    if proc.stdout is not None:
                        proc.stdout.close()
                    if proc.stderr is not None:
                        proc.stderr.close()
                exit_code = proc.returncode
        except OSError as error:
            reason = "local_spawn_or_io_error"
            stderr = str(error).encode("utf-8", errors="replace")
        transport = "unknown" if reason is not None else "response_received"
        safe_stdout, stdout_rules = _redact_text(stdout.decode("utf-8", errors="replace"))
        safe_stderr, stderr_rules = _redact_text(stderr.decode("utf-8", errors="replace"))
        response = {"stdout": safe_stdout, "stderr": safe_stderr, "exit_code": exit_code, "transport_status": transport, "local_reason": reason, "transport_scope": "Local CLI completion observation only; response_received does not prove remote delivery, HTTP success or business completion", "redaction": {"stdout": stdout_rules, "stderr": stderr_rules, "scope": "Common credential fields and token/assignment patterns; arbitrary secrets in prose are not guaranteed detectable", "encoding": "UTF-8 with replacement for undecodable output bytes", "hash_semantics": "Hashes cover saved redacted UTF-8 text/JSON only, NOT raw stdout/stderr or HTTP wire", "stdout_original_byte_count": len(stdout), "stderr_original_byte_count": len(stderr)}, "stdout_redacted_sha256": _digest(safe_stdout.encode("utf-8")), "stderr_redacted_sha256": _digest(safe_stderr.encode("utf-8"))}
        response_raw = _json_bytes(response)
        _write(fd, "response.json", response_raw)
        record = {"schema_version": 1, "invocation_id": invocation_id, "workspace_id": workspace_id, "endpoint": endpoint, "started_at": started_at, "finished_at": _now(), "request_path": str(directory / "request.json"), "request_sha256": request_sha, "response_path": str(directory / "response.json"), "response_sha256": _digest(response_raw), "exit_code": exit_code, "transport_status": transport, "local_reason": reason, "business_status": "not_interpreted", "external_outputs": outputs}
        record_raw = _json_bytes(record)
        _write(fd, "record.json", record_raw)
        return dict(record, record_path=str(directory / "record.json"))
    finally:
        if fd is not None:
            os.close(fd)
        os.close(root_fd)


def verify_capture(record_path: Path) -> dict:
    """Offline consistency check; not a signature or protection from owner edits.

    Verify saved request/snapshot/response bytes, scope and invocation associations.
    Missing final record is not repaired: the caller must keep submission unknown.
    """
    try:
        path = _absolute(record_path)
        if path.name != "record.json":
            raise ValueError("Expected record.json")
        record = json.loads(_read_input(path))
        directory = path.parent
        request_path = directory / "request.json"
        response_path = directory / "response.json"
        if record["request_path"] != str(request_path) or record["response_path"] != str(response_path):
            raise ValueError("Record paths must stay within its invocation directory")
        request_raw, response_raw = _read_input(request_path), _read_input(response_path)
        if _digest(request_raw) != record["request_sha256"] or _digest(response_raw) != record["response_sha256"]:
            raise ValueError("Request or saved redacted response digest mismatch")
        request, response = json.loads(request_raw), json.loads(response_raw)
        if any(record.get(k) != request.get(k) for k in ("schema_version", "invocation_id", "workspace_id", "endpoint")) or record["schema_version"] != 1 or record["invocation_id"] != directory.name or request["status"] != "started":
            raise ValueError("Record/request invocation or scope mismatch")
        argv = request["argv"]
        if argv[1:7] != ["-o", "-no-interactive", "-endpoint", record["endpoint"], "-workspace-id", record["workspace_id"]]:
            raise ValueError("Executed argv does not bind the recorded scope")
        expected_pairs = []
        expected_paths = []
        for item in request["inputs"]:
            snapshot = Path(item["snapshot_path"])
            if snapshot.parent != directory or not re.fullmatch(r"input-\d{3,}\.[A-Za-z0-9]{1,12}", snapshot.name):
                raise ValueError("Input snapshot path escaped its invocation")
            raw = _read_input(snapshot)
            if _digest(raw) != item["sha256"] or len(raw) != item["byte_count"]:
                raise ValueError("Saved input snapshot digest or length mismatch")
            expected_pairs.append((item["flag"], str(snapshot)))
            expected_paths.append(snapshot)
        actual_pairs = [(v, argv[i + 1]) for i, v in enumerate(argv[:-1]) if v in _STRUCTURED_FILES or v in _OPAQUE_FILES]
        if Counter(actual_pairs) != Counter(expected_pairs) or len(set(expected_paths)) != len(expected_paths):
            raise ValueError("Executed file inputs do not match all captured snapshots")
        if not isinstance(response["stdout"], str) or not isinstance(response["stderr"], str):
            raise ValueError("Response stdout/stderr must be saved redacted strings")
        for name in ("stdout", "stderr"):
            if _digest(response[name].encode("utf-8")) != response[name + "_redacted_sha256"]:
                raise ValueError("Redacted stream digest mismatch")
        if record["transport_status"] not in {"response_received", "unknown"} or any(record.get(k) != response.get(k) for k in ("exit_code", "transport_status", "local_reason")):
            raise ValueError("Record/response transport observations mismatch")
        if stat.S_IMODE(directory.stat().st_mode) != 0o700 or stat.S_IMODE(directory.parent.stat().st_mode) != 0o700 or any(stat.S_IMODE(p.stat().st_mode) != 0o600 for p in [path, request_path, response_path] + expected_paths):
            raise ValueError("Evidence directory/file permissions are not private")
        return {"verified": True, "record_path": str(path), "request_path": str(request_path), "request_sha256": record["request_sha256"], "response_path": str(response_path), "response_sha256": record["response_sha256"], "workspace_id": record["workspace_id"], "endpoint": record["endpoint"], "invocation_id": record["invocation_id"], "input_count": len(expected_paths), "transport_status": record["transport_status"], "scope": "Local saved-byte/scope consistency only; no signature, WORM, wire capture or business success proof; external outputs excluded"}
    except (OSError, KeyError, TypeError, IndexError, json.JSONDecodeError) as error:
        raise ValueError("Incomplete or invalid capture evidence") from error


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--evidence-dir", type=Path, required=True)
    parser.add_argument("--invocation-id", required=True)
    parser.add_argument("--workspace-id", required=True)
    parser.add_argument("--endpoint", required=True)
    parser.add_argument("--timeout", type=float, default=60.0)
    parser.add_argument("--cli-bin", type=Path)
    parser.add_argument("args", nargs=argparse.REMAINDER, help="One CLI command after --")
    options = vars(parser.parse_args())
    if options["args"][:1] == ["--"]:
        options["args"] = options["args"][1:]
    try:
        result = capture(**options)
    except (ValueError, OSError) as error:
        safe_error, _ = _redact_text(str(error))
        print(json.dumps({"local_error": safe_error, "note": "Inspect existing invocation evidence; never automatically resubmit"}, ensure_ascii=False), file=sys.stderr)
        return 2
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if result["transport_status"] == "response_received" else 3


if __name__ == "__main__":
    sys.exit(main())
