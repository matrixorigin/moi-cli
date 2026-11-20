#!/usr/bin/env python3
"""Generate a pinned bootstrap; installation/update stays owned by moi-cli."""
import argparse
import json
from pathlib import Path
import re
import shlex
from urllib.parse import urlparse


def require_client_url(value):
    if not isinstance(value, str) or not value or any(c.isspace() for c in value):
        raise ValueError("a trusted Astra HTTPS origin must be configured for the release")
    parsed = urlparse(value)
    if (parsed.scheme != "https" or not parsed.hostname or parsed.username or parsed.password
            or parsed.query or parsed.fragment or parsed.path not in ("", "/")):
        raise ValueError("Astra service must be an HTTPS origin without credentials or query")
    if parsed.port is not None and not 1 <= parsed.port <= 65535:
        raise ValueError("invalid Astra service port")
    return value.rstrip("/")


def render(manifest, channel):
    parsed = urlparse(channel)
    if parsed.scheme != "https" or not parsed.hostname or parsed.username or parsed.password or parsed.query or parsed.fragment:
        raise ValueError("trusted HTTPS channel required")
    client_url = require_client_url(manifest.get("client_config", {}).get("astra_url"))
    compatibility = manifest.get("compatibility_reference")
    if not isinstance(compatibility, str) or not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9._:/@+-]{0,159}", compatibility):
        raise ValueError("invalid compatibility reference")
    cases = []
    for platform, entry in sorted(manifest["platforms"].items()):
        if platform not in ("darwin-arm64", "darwin-amd64", "linux-arm64", "linux-amd64"):
            raise ValueError("invalid platform")
        item = entry["components"]["moi-cli"]
        url = urlparse(item["url"])
        if (url.scheme, url.netloc) != (parsed.scheme, parsed.netloc) or url.query or url.fragment or url.username:
            raise ValueError("cross-origin bootstrap")
        if not re.fullmatch(r"[0-9a-f]{64}", item["sha256"]):
            raise ValueError("invalid checksum")
        if type(item.get("size")) is not int or not 0 < item["size"] <= 512 * 1024 * 1024:
            raise ValueError("invalid bootstrap size")
        cases.append(f"  {platform}) binary_url={shlex.quote(item['url'])}; expected={item['sha256']}; expected_size={item['size']};;")
    template = r'''#!/bin/sh
set -eu
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }
@SETUP_FUNCTIONS@
prefix="${HOME:?HOME is required}/.local/share/moi/client"
skill_dir="${ASTRA_LOCAL_STATE_ROOT:-$HOME/.astra}/skills"
shell_profile=''
modify_path=yes
client_api_url=@CLIENT_URL@
while [ "$#" -gt 0 ]; do
  case "$1" in
    --dir|--skill-dir) [ "$#" -ge 2 ] || fail 'missing argument'
      case "$1" in --dir) prefix=$2;; --skill-dir) skill_dir=$2;; esac; shift 2;;
    --shell-profile) [ "$#" -ge 2 ] || fail 'missing argument'; shell_profile=$2; shift 2;;
    --no-modify-path) modify_path=no; shift;;
    --help|-h) printf '%s\n' 'Usage: install.sh [--dir ABSOLUTE_DIRECTORY] [--skill-dir ABSOLUTE_DIRECTORY] [--shell-profile ABSOLUTE_FILE] [--no-modify-path]'; exit 0;;
    *) fail 'unknown option';;
  esac
done
for required in curl tar mktemp uname awk wc sed grep cmp readlink chmod cp ln mkdir rm rmdir mv cat dirname; do
  command -v "$required" >/dev/null 2>&1 || fail "$required is required"
done
command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1 || fail 'sha256sum or shasum is required'
client_setup_preflight
case "$(uname -s)" in Darwin) os=darwin;; Linux) os=linux;; *) fail 'unsupported OS';; esac
case "$(uname -m)" in arm64|aarch64) arch=arm64;; x86_64|amd64) arch=amd64;; *) fail 'unsupported architecture';; esac
case "$os-$arch" in
@CASES@
  *) fail 'no release for this platform';;
esac
stage=$(mktemp -d "${TMPDIR:-/tmp}/moi-bootstrap.XXXXXX")
trap 'rm -f "$stage/component.tar.gz" "$stage/moi-cli"; rmdir "$stage"' EXIT
trap 'exit 1' HUP INT TERM
curl_progress=-s
if [ -t 2 ] && [ "${TERM:-}" != dumb ] && [ "${MOI_AGENT_CALL:-}" != 1 ] && [ -z "${CI:-}" ]; then
  printf '%s\n' 'Downloading installer (moi-cli)...' >&2
  curl_progress=--progress-bar
fi
status=$(curl @CURL@ -fS "$curl_progress" --max-filesize "$expected_size" --max-time 300 -w '%{http_code}' -o "$stage/component.tar.gz" "$binary_url") || fail 'download failed'
[ "$status" = 200 ] || fail 'direct download required; redirects refused'
[ "$(wc -c < "$stage/component.tar.gz" | awk '{print $1}')" = "$expected_size" ] || fail 'download size mismatch'
if command -v sha256sum >/dev/null 2>&1; then actual=$(sha256sum "$stage/component.tar.gz" | awk '{print $1}')
else actual=$(shasum -a 256 "$stage/component.tar.gz" | awk '{print $1}'); fi
[ "$actual" = "$expected" ] || fail 'checksum mismatch'
[ "$(tar -tzf "$stage/component.tar.gz")" = moi-cli ] || fail 'unexpected archive members'
tar -xOf "$stage/component.tar.gz" moi-cli > "$stage/moi-cli"
chmod 0755 "$stage/moi-cli"
[ "$("$stage/moi-cli" --moi-update-protocol)" = moi-client-update-v1 ] || fail 'unsupported updater'
"$stage/moi-cli" update install --dir "$prefix" --skill-dir "$skill_dir" --channel @CHANNEL@ \
  --compatibility @COMPATIBILITY@ --astra-url "$client_api_url"
client_setup
'''
    return (template.replace("@CASES@", "\n".join(cases))
            .replace("@SETUP_FUNCTIONS@", Path(__file__).with_name("client_setup.sh").read_text())
            .replace("@CLIENT_URL@", shlex.quote(client_url))
            .replace("@CHANNEL@", shlex.quote(channel))
            .replace("@COMPATIBILITY@", shlex.quote(compatibility))
            .replace("@CURL@", "--proto '=https' --tlsv1.2")).encode()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--channel", required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    data = render(json.loads(args.manifest.read_bytes()), args.channel)
    with args.output.open("xb") as output:
        output.write(data)
