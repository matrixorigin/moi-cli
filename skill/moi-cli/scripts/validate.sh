#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
PACKAGE_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
EVIDENCE_DIR="${MOI_ITERATION_EVIDENCE_DIR:-$PACKAGE_ROOT/references/iterations/5.2.3}"

exec python3 "$SCRIPT_DIR/iteration_report_check.py" "$EVIDENCE_DIR" "$PACKAGE_ROOT"
