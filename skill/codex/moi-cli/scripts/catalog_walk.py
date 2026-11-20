#!/usr/bin/env python3
"""Recursively discover Catalog volumes and files through the dev Product CLI."""
from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import subprocess


def run_cli(cli, endpoint, workspace, timeout, args):
    command = [
        cli,
        "-o",
        "-no-interactive",
        "-endpoint",
        endpoint,
        "-workspace-id",
        workspace,
        *args,
    ]
    try:
        result = subprocess.run(command, capture_output=True, text=True, timeout=timeout, check=False)
    except subprocess.TimeoutExpired as error:
        raise RuntimeError(f"CLI timed out: {' '.join(args)}") from error
    if result.returncode:
        raise RuntimeError(f"CLI failed ({result.returncode}): {result.stderr.strip()}")
    try:
        return json.loads(result.stdout)
    except json.JSONDecodeError as error:
        raise RuntimeError(f"CLI returned non-JSON output: {' '.join(args)}") from error


def walk(cli, endpoint, workspace, database, root, timeout, max_depth, max_nodes, page_size):
    volumes = []
    visited = set()
    queue = []
    entry_count = 0
    if root:
        queue.append((str(root), None, "root", 0))
    else:
        response = run_cli(cli, endpoint, workspace, timeout, ["catalog", "database-children", "--database-id", str(database)])
        roots = [item for item in response.get("list", []) if item.get("type") == "volume"]
        if len(roots) != 1:
            raise RuntimeError(f"expected one root Volume, got {len(roots)}")
        queue.append((str(roots[0]["id"]), None, roots[0].get("display_name") or roots[0].get("name") or "root", 0))
    while queue:
        volume_id, parent_id, path, depth = queue.pop(0)
        if volume_id in visited:
            continue
        if depth > max_depth:
            raise RuntimeError(f"max depth exceeded: {max_depth}")
        visited.add(volume_id)
        files = []
        children = []
        page = 1
        seen_page_ids = set()
        while True:
            response = run_cli(
                cli,
                endpoint,
                workspace,
                timeout,
                [
                    "catalog",
                    "files",
                    "--volume-id",
                    volume_id,
                    "--parent-id",
                    "",
                    "--page",
                    str(page),
                    "--page-size",
                    str(page_size),
                ],
            )
            items = response.get("list", [])
            total = int(response.get("total", len(items)))
            for item in items:
                identity = f"{item.get('file_type')}:{item.get('id')}:{item.get('parent_id') or volume_id}"
                if identity in seen_page_ids:
                    raise RuntimeError(f"duplicate entry while paginating: {identity}")
                seen_page_ids.add(identity)
                if item.get("file_type") == "folder":
                    children.append(item)
                else:
                    files.append(item)
            if len(seen_page_ids) >= total or not items:
                break
            page += 1
            if page > max_nodes:
                raise RuntimeError("pagination exceeded the node budget")
        entry_count += len(files) + len(children) + 1
        if entry_count > max_nodes:
            raise RuntimeError(f"max nodes exceeded: {max_nodes}")
        volumes.append(
            {
                "volume_id": volume_id,
                "parent_volume_id": parent_id,
                "path": path,
                "depth": depth,
                "files": files,
                "total_entries": len(files) + len(children),
            }
        )
        for child in children:
            child_id = str(child["volume_id"])
            if child_id == volume_id or (parent_id and child_id == parent_id):
                raise RuntimeError(f"cycle detected at Volume {child_id}")
            queue.append((child_id, volume_id, f"{path}/{child.get('name', child_id)}", depth + 1))
    return {"complete": True, "database_id": str(database) if database else None, "root_volume_id": root, "volumes": volumes}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--database-id")
    parser.add_argument("--root-volume-id", help="Optional known root Volume; skips database-children")
    parser.add_argument("--workspace-id", required=True)
    parser.add_argument("--endpoint", required=True)
    parser.add_argument("--cli-bin", default=os.environ.get("MOI_DEV_CLI_BIN", "~/.local/bin/moi-cli-dev"))
    parser.add_argument("--timeout", type=float, default=60)
    parser.add_argument("--max-depth", type=int, default=8)
    parser.add_argument("--max-nodes", type=int, default=500)
    parser.add_argument("--page-size", type=int, default=50)
    args = parser.parse_args()
    if not args.database_id and not args.root_volume_id:
        parser.error("--database-id or --root-volume-id is required")
    if args.max_depth < 0 or args.max_nodes < 1 or not 1 <= args.page_size <= 200 or args.timeout <= 0:
        parser.error("invalid traversal limits")
    result = walk(
        str(Path(args.cli_bin).expanduser()),
        args.endpoint,
        args.workspace_id,
        args.database_id,
        args.root_volume_id,
        args.timeout,
        args.max_depth,
        args.max_nodes,
        args.page_size,
    )
    result["root_volume_id"] = result["root_volume_id"] or result["volumes"][0]["volume_id"]
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
