#!/bin/sh
# Rendered by snapshot_installer.py. This is not the formal bundle installer.
set -eu
umask 077
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }
quote_shell() { printf "'"; printf '%s' "$1" | sed "s/'/'\\\\''/g"; printf "'"; }
safe_path() {
  case "$1" in /*) ;; *) fail 'directory must be absolute';; esac
  case "$1" in *'/../'*|*'/./'*|*/..|*/.|*//*|*:*|*'
'*|*'
'*) fail 'directory must be canonical';; esac
  probe=$1
  while [ "$probe" != / ]; do
    [ ! -L "$probe" ] || fail 'directory must not contain symbolic links'
    probe=$(dirname "$probe")
  done
}
prefix="${HOME:?HOME is required}/.local/share/moi/qa-client"
modify_path=yes
while [ "$#" -gt 0 ]; do
  case "$1" in
    --dir) [ "$#" -ge 2 ] || fail 'missing directory'; prefix=$2; shift 2;;
    --no-modify-path) modify_path=no; shift;;
    --help|-h) printf '%s\n' 'Usage: install.sh [--dir ABSOLUTE_QA_DIRECTORY] [--no-modify-path]'; exit 0;;
    *) fail 'unknown option';;
  esac
done
safe_path "$prefix"
[ "$prefix" != / ] && [ "$prefix" != "$HOME" ] || fail 'choose a dedicated QA directory'
for required in curl tar mktemp uname awk wc sed grep sort cmp readlink chmod ln mkdir rm rmdir mv cat dirname; do
  command -v "$required" >/dev/null 2>&1 || fail "$required is required"
done
command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1 || fail 'SHA256 utility required'
case "$(uname -s)" in Darwin) os=darwin;; Linux) os=linux;; *) fail 'unsupported OS';; esac
case "$(uname -m)" in arm64|aarch64) arch=arm64;; x86_64|amd64) arch=amd64;; *) fail 'unsupported architecture';; esac
if [ -e "$prefix" ]; then
  [ -d "$prefix" ] && [ -f "$prefix/.qa-snapshot-root" ] && [ ! -L "$prefix/.qa-snapshot-root" ] || fail 'existing directory is not a QA snapshot installation'
  [ "$(cat "$prefix/.qa-snapshot-root")" = moi-client-snapshot-v1 ] || fail 'unknown QA installation'
else
  mkdir -p "$prefix"
  printf '%s\n' moi-client-snapshot-v1 > "$prefix/.qa-snapshot-root"
fi
for directory in snapshots bin; do
  safe_path "$prefix/$directory"
  mkdir -p "$prefix/$directory"
done
identity=@IDENTITY@
destination="$prefix/snapshots/$identity"
[ ! -e "$destination" ] && [ ! -L "$destination" ] || fail 'snapshot is already installed; use its existing commands or a new --dir'
mkdir "$prefix/.install-lock" 2>/dev/null || fail 'QA installation is busy; inspect an abandoned lock before retrying'
stage=''; next=''
cleanup() {
  [ -z "$next" ] || rm -f "$next"
  [ -z "$stage" ] || rm -rf "$stage"
  rmdir "$prefix/.install-lock"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM
stage=$(mktemp -d "$prefix/snapshots/.install.XXXXXX")
mkdir "$stage/bin" "$stage/launchers" "$stage/state" "$stage/moi-home"

component() {
  name=$1; url=$2; expected=$3; size=$4; members=$5
  status=$(curl --proto '=https' --tlsv1.2 -fsS --max-time 300 --max-filesize "$size" -w '%{http_code}' -o "$stage/archive" "$url") || fail 'snapshot download failed'
  [ "$status" = 200 ] || fail 'snapshot download requires direct 200; redirects refused'
  [ "$(wc -c < "$stage/archive" | awk '{print $1}')" = "$size" ] || fail 'snapshot download size mismatch'
  if command -v sha256sum >/dev/null 2>&1; then actual=$(sha256sum "$stage/archive" | awk '{print $1}')
  else actual=$(shasum -a 256 "$stage/archive" | awk '{print $1}'); fi
  [ "$actual" = "$expected" ] || fail 'snapshot checksum mismatch'
  [ "$(tar -tzf "$stage/archive" | LC_ALL=C sort)" = "$members" ] || fail 'unexpected snapshot archive members'
  # Never extract an archive-controlled path or symbolic link into the filesystem.
  tar -xOf "$stage/archive" "$name" > "$stage/bin/$name"
  chmod 0755 "$stage/bin/$name"
  if [ "$name" != moi-cli ]; then tar -xOf "$stage/archive" LICENSE > "$stage/$name-LICENSE"; fi
  rm "$stage/archive"
}
case "$os-$arch" in
@CASES@
  *) fail 'snapshot was not built for this platform';;
esac
# Offline export uses only private staging state. Login remains Astra-owned.
HOME="$stage/moi-home" MOI_CLI_CONFIG_FILE="$stage/moi-home/.moi-cli.json" \
  ASTRA_LOCAL_STATE_ROOT="$stage/state" MOI_AUTH_DIR="$stage/auth" \
  "$stage/bin/moi-cli" skill export --dir "$stage/state/skills/moi-cli"
[ -f "$stage/state/skills/moi-cli/SKILL.md" ] || fail 'snapshot Skill export failed'

launcher() {
  printf '%s\n' '#!/bin/sh' '# MOI QA snapshot launcher' 'set -eu'
  printf 'snapshot=%s\n' "$(quote_shell "$destination")"
  printf 'export ASTRA_API_URL=%s\n' "$(quote_shell @ASTRA_URL@)"
  printf '%s\n' 'export ASTRA_SERVER_URL="$ASTRA_API_URL"' \
    'export MOI_CLI_CONFIG_FILE="$snapshot/moi-home/.moi-cli.json"' \
    'export ASTRA_LOCAL_STATE_ROOT="$snapshot/state"' \
    'export MOI_AUTH_DIR="$snapshot/auth"' \
    'export ASTRA_CLI_CREDENTIALS_DIR="$snapshot/legacy-auth"' \
    'export MOI_ASTRA_BINARY="$snapshot/launchers/astra"' \
    'export PATH="$snapshot/launchers:$snapshot/bin:$PATH"' \
    'unset ASTRA_ACCESS_TOKEN'
  if [ "$1" = astra ]; then
    printf '%s\n' 'exec "$snapshot/bin/astra" --api-url "$ASTRA_API_URL" "$@"'
  else
    if [ "$1" = moi-cli ]; then printf '%s\n' 'export HOME="$snapshot/moi-home"'; fi
    printf 'exec "$snapshot/bin/%s" "$@"\n' "$1"
  fi
}
for name in astra astra-edge moi-cli; do
  launcher "$name" > "$stage/launchers/$name-qa"
  chmod 0755 "$stage/launchers/$name-qa"
  ln -s "$name-qa" "$stage/launchers/$name"
  # Refuse replacing a command belonging to anything except this QA root.
  target="$prefix/bin/$name-qa"
  if [ -e "$target" ] || [ -L "$target" ]; then
    [ -L "$target" ] || fail 'existing QA command is not managed here'
    [ "$(readlink "$target")" = "../active/launchers/$name-qa" ] || fail 'existing QA command belongs to another installation'
  fi
done
command_dir="$HOME/.local/bin"
if [ "$modify_path" = yes ]; then
  safe_path "$command_dir"
  mkdir -p "$command_dir"
  for name in astra astra-edge moi-cli; do
    target="$command_dir/$name-qa"
    if [ -e "$target" ] || [ -L "$target" ]; then
      [ -L "$target" ] && [ "$(readlink "$target")" = "$prefix/bin/$name-qa" ] || fail 'an existing QA command belongs to another installation; use --no-modify-path'
    fi
  done
fi
[ ! -e "$prefix/active" ] || [ -L "$prefix/active" ] || fail 'QA active pointer is not managed here'
[ ! -L "$prefix/active" ] || case "$(readlink "$prefix/active")" in snapshots/*) ;; *) fail 'QA active pointer is not managed here';; esac
printf '%s\n' "$identity" > "$stage/snapshot-id"
mv "$stage" "$destination"
stage=''
for name in astra astra-edge moi-cli; do
  if [ ! -L "$prefix/bin/$name-qa" ]; then
    ln -s "../active/launchers/$name-qa" "$prefix/bin/$name-qa"
  fi
  if [ "$modify_path" = yes ] && [ ! -L "$command_dir/$name-qa" ]; then
    ln -s "$prefix/bin/$name-qa" "$command_dir/$name-qa"
  fi
done
# One pointer activates the entire pair, Edge, Skill and snapshot-local state.
next="$prefix/.active-$$"
ln -s "snapshots/$identity" "$next"
if [ "$os" = darwin ]; then mv -fh "$next" "$prefix/active"
else mv -fT "$next" "$prefix/active"; fi
next=''
printf 'Installed QA snapshot %s. Run: %s/bin/astra-qa login\n' "$identity" "$prefix"
if [ "$modify_path" = yes ]; then
  printf 'If astra-qa is not on PATH, run: export PATH=%s:"$PATH"\n' "$(quote_shell "$command_dir")"
fi
printf '%s\n' 'Snapshots have separate login/session/Skill state. Install the next snapshot explicitly; astra update does not manage snapshots.'
