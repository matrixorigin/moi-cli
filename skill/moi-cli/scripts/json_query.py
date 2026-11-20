#!/usr/bin/env python3
"""Filter, sort and page an already captured moi-cli JSON result locally.

This helper never calls MOI. Its output is client-side post-processing evidence,
not proof of a server-side filter, a product UI control or a new business result.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import sys
from typing import Any

from record_cli import _redact_text


_AUTO_LISTS = ("items", "list", "query_list", "models", "connectors", "files")


def _load(path: Path) -> Any:
    value = json.loads(path.read_text(encoding="utf-8"))
    if isinstance(value, dict) and isinstance(value.get("stdout"), str):
        stdout = value["stdout"].strip()
        if not stdout:
            raise ValueError("captured CLI stdout is empty")
        value = json.loads(stdout)
    return value


def _at(value: Any, path: str) -> Any:
    current = value
    if not path:
        return current
    for part in path.split("."):
        if not isinstance(current, dict) or part not in current:
            return None
        current = current[part]
    return current


def _items(document: Any, path: str) -> tuple[list[Any], str]:
    if path:
        value = _at(document, path)
        if not isinstance(value, list):
            raise ValueError(f"items path is not a list: {path}")
        return value, path
    if isinstance(document, list):
        return document, ""
    if isinstance(document, dict):
        for key in _AUTO_LISTS:
            if isinstance(document.get(key), list):
                return document[key], key
    raise ValueError("no list found; pass --items-path explicitly")


def _text(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value
    return json.dumps(value, ensure_ascii=False, sort_keys=True)


def _matches(item: Any, args: argparse.Namespace) -> bool:
    if args.search_text:
        if args.search_field:
            haystack = " ".join(_text(_at(item, field)) for field in args.search_field)
        else:
            haystack = _text(item)
        if args.search_text.casefold() not in haystack.casefold():
            return False
    for expression in args.where:
        field, expected = expression.split("=", 1)
        actual = _at(item, field)
        if actual is None or _text(actual).casefold() != expected.casefold():
            return False
    return True


def _sort_key(field: str):
    def key(item: Any):
        value = _at(item, field)
        if value is None:
            return (2, "", "")
        if isinstance(value, (int, float)) and not isinstance(value, bool):
            return (0, float(value), "")
        return (1, 0, _text(value).casefold())

    return key


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("input", type=Path, help="CLI JSON or a recorder response.json containing stdout")
    parser.add_argument("--items-path", default="", help="Dot path to the result list; common lists are detected")
    parser.add_argument("--search-text", default="", help="Case-insensitive client-side substring filter")
    parser.add_argument("--search-field", action="append", default=[], help="Field used by --search-text; repeatable")
    parser.add_argument("--where", action="append", default=[], metavar="FIELD=VALUE", help="Exact client-side field filter; repeatable")
    parser.add_argument("--sort", dest="sort_field", default="", help="Dot path used for a stable local sort")
    parser.add_argument("--desc", action="store_true", help="Sort descending instead of ascending")
    parser.add_argument("--select", action="append", default=[], metavar="FIELD", help="Output only this field; repeatable")
    parser.add_argument("--offset", type=int, default=0)
    parser.add_argument("--limit", type=int, default=50)
    args = parser.parse_args()
    if args.offset < 0 or args.limit < 1:
        parser.error("--offset must be >= 0 and --limit must be >= 1")
    for expression in args.where:
        if "=" not in expression or not expression.split("=", 1)[0]:
            parser.error("--where requires FIELD=VALUE")

    document = _load(args.input)
    source_items, detected_path = _items(document, args.items_path)
    filtered = [item for item in source_items if _matches(item, args)]
    if args.sort_field:
        filtered.sort(key=_sort_key(args.sort_field), reverse=args.desc)
    total = len(filtered)
    page = filtered[args.offset:args.offset + args.limit]
    if args.select:
        page = [{field: _at(item, field) for field in args.select} for item in page]

    result = {
        "schema_version": 1,
        "mode": "client_side_post_processing",
        "server_filter_claim": False,
        "product_ui_claim": False,
        "items_path": detected_path,
        "source_count": len(source_items),
        "total": total,
        "returned": len(page),
        "offset": args.offset,
        "limit": args.limit,
        "items": page,
    }
    safe, redactions = _redact_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    if redactions:
        redacted_result = json.loads(safe)
        redacted_result["redaction_note"] = "Common credential patterns were redacted; arbitrary business secrets are not guaranteed detectable."
        redacted_result["redactions"] = redactions
        safe = json.dumps(redacted_result, ensure_ascii=False, indent=2) + "\n"
    sys.stdout.write(safe)


if __name__ == "__main__":
    main()
