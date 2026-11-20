#!/usr/bin/env sh
# Install one explicitly pinned, matched pair; never fetch unrelated "latest"s.
set -eu
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }
archive=''; url=''; expected=''; loopback_http=no; prefix="${HOME:?HOME is required}/.local/share/moi/cli"
skill_dir=''; preview_state=no
while [ "$#" -gt 0 ]; do
    case "$1" in
        --archive|--url|--sha256|--dir|--skill-dir)
            [ "$#" -ge 2 ] || fail "$1 needs a value"
            case "$1" in
                --archive) archive=$2;; --url) url=$2;; --sha256) expected=$2;; --dir) prefix=$2;; --skill-dir) skill_dir=$2;;
            esac
            shift 2;;
        --allow-loopback-http) loopback_http=yes; shift;;
        --preview-state) preview_state=yes; shift;;
        --help|-h)
            printf '%s\n' 'Usage: install.sh (--archive FILE | --url HTTPS_URL) --sha256 HASH [--dir DIRECTORY] [--skill-dir DIRECTORY]' \
                'Skill defaults to ASTRA_LOCAL_STATE_ROOT/skills or ~/.astra/skills; --preview-state uses the install directory/astra-state/skills.' \
                'Local download previews only: --allow-loopback-http permits http://127.0.0.1:PORT/path without redirects.'
            exit 0;;
        *) fail "unknown option: $1";;
    esac
done
case "$prefix" in /*) ;; *) fail 'installation directory must be absolute';; esac
[ "$prefix" != / ] && [ "$prefix" != "$HOME" ] || fail 'choose a dedicated installation directory'
[ ! -L "$prefix" ] || fail 'installation directory must not be a symlink'
if [ -z "$skill_dir" ]; then
    if [ "$preview_state" = yes ]; then skill_dir="$prefix/astra-state/skills"
    elif [ -n "${ASTRA_LOCAL_STATE_ROOT:-}" ]; then skill_dir="$ASTRA_LOCAL_STATE_ROOT/skills"
    else skill_dir="$HOME/.astra/skills"; fi
fi
case "$skill_dir" in /*) ;; *) fail 'skill directory must be absolute';; esac
[ ! -L "$skill_dir" ] || fail 'skill directory must not be a symlink'
skill_link="$skill_dir/moi-cli"
skill_store="$skill_dir/.moi-cli-releases"
skill_target="$skill_store/$expected"
old_skill_target=''
if [ -e "$skill_store" ] || [ -L "$skill_store" ]; then
    [ -d "$skill_store" ] && [ ! -L "$skill_store" ] && [ ! -L "$skill_store/owner" ] && [ -f "$skill_store/owner" ] || fail 'existing Skill store is not owned by this installer'
    [ "$(cat "$skill_store/owner")" = "$prefix" ] || fail 'Skill store belongs to a different installation'
fi
if [ -e "$skill_link" ] || [ -L "$skill_link" ]; then
    [ -L "$skill_link" ] && [ -f "$skill_store/owner" ] || fail 'existing moi-cli Skill is not owned by this installation; nothing was overwritten'
    old_skill_target=$(readlink "$skill_link")
    case "$old_skill_target" in "$skill_store"/*) ;; *) fail 'existing moi-cli Skill has an unowned target';; esac
    old_skill_hash=${old_skill_target#"$skill_store/"}
    printf '%s\n' "$old_skill_hash" | grep -Eq '^[0-9a-f]{64}$' || fail 'invalid managed Skill target'
fi
printf '%s\n' "$expected" | grep -Eq '^[0-9a-f]{64}$' || fail 'an explicit SHA256 is required'
if [ -n "$archive" ]; then
    [ -z "$url" ] || fail 'choose archive or URL, not both'
    [ -f "$archive" ] || fail 'archive not found'
else
    case "$url" in
        https://*) ;;
        http://*)
            [ "$loopback_http" = yes ] || fail 'download requires HTTPS unless explicitly using a loopback preview'
            printf '%s\n' "$url" | grep -Eq '^http://127[.]0[.]0[.]1:[0-9]+/[A-Za-z0-9._/-]+$' || fail 'HTTP preview must use literal 127.0.0.1 and an explicit port'
            ;;
        *) fail 'download requires an HTTPS URL';;
    esac
fi
case "$(uname -s)" in Darwin) os=darwin;; Linux) os=linux;; *) fail 'unsupported operating system';; esac
case "$(uname -m)" in arm64|aarch64) arch=arm64;; x86_64|amd64) arch=amd64;; *) fail 'unsupported architecture';; esac
for required in tar mktemp awk grep sort cmp diff mv ln readlink cp; do
    command -v "$required" >/dev/null 2>&1 || fail "$required is required"
done
mkdir -p "$prefix/releases"
[ ! -L "$prefix/releases" ] || fail 'releases must not be a symlink'
[ ! -e "$prefix/current" ] || [ -L "$prefix/current" ] || fail 'current must be a symlink managed by this installer'
mkdir "$prefix/.install-lock" 2>/dev/null || fail 'another installation is active (or its lock needs inspection)'
stage=''; next=''; skill_next=''; skill_stage=''; skill_changed=no; activated=no
cleanup() {
    # Both paths are minted by this process inside the dedicated prefix.
    if [ -n "$next" ]; then rm -f "$next"; fi
    if [ -n "$stage" ]; then rm -rf "$stage"; fi
    if [ -n "$skill_next" ]; then rm -f "$skill_next"; fi
    if [ -n "$skill_stage" ]; then rm -rf "$skill_stage"; fi
    if [ "$skill_changed" = yes ] && [ "$activated" = no ] && [ -L "$skill_link" ] && [ "$(readlink "$skill_link")" = "$skill_target" ]; then
        if [ -n "$old_skill_target" ]; then
            rollback="$skill_dir/.moi-cli-rollback-$expected"
            ln -s "$old_skill_target" "$rollback" && {
                if [ "$os" = darwin ]; then /bin/mv -fh "$rollback" "$skill_link"; else /bin/mv -fT "$rollback" "$skill_link"; fi
            }
        else rm "$skill_link"; fi
    fi
    rmdir "$prefix/.install-lock"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM
stage=$(mktemp -d "$prefix/releases/.install.XXXXXX")
if [ -n "$url" ]; then
    if [ "$loopback_http" = yes ] && [ "${url#http://}" != "$url" ]; then
        # Never follow a redirect or send this explicitly local download to an
        # environment-configured proxy. The regular HTTPS contract is unchanged.
        status=$(curl --noproxy '*' --proto '=http' -fsS --max-time 300 -w '%{http_code}' -o "$stage/download.tar.gz" "$url") || fail 'download failed'
        [ "$status" = 200 ] || fail 'HTTP preview requires a direct 200 response; redirects are refused'
    else
        curl --proto '=https' --proto-redir '=https' --tlsv1.2 -fsSL --max-time 300 -o "$stage/download.tar.gz" "$url" || fail 'download failed'
    fi
    archive="$stage/download.tar.gz"
fi
if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$archive" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$archive" | awk '{print $1}')
else fail 'sha256sum or shasum is required'; fi
[ "$actual" = "$expected" ] || fail 'checksum mismatch; current version unchanged'
# The manifest allowlist uses byte order, not the customer's locale collation.
members=$(tar -tzf "$archive" | LC_ALL=C sort)
[ "$members" = "$(printf 'ASTRA-LICENSE\nastra\nbundle.txt\nmoi-cli')" ] || fail 'unexpected archive members'
mkdir "$stage/pair"
# Never extract paths/symlinks from an archive into the filesystem. Copy only
# the four allowlisted contents to paths owned by this invocation.
for name in ASTRA-LICENSE astra bundle.txt moi-cli; do
    tar -xOf "$archive" "$name" > "$stage/pair/$name" || fail 'cannot read archive member'
done
{
    IFS= read -r format
    IFS= read -r target
    IFS= read -r version
    IFS= read -r astra_version
    IFS= read -r moi_version
    if IFS= read -r extra; then fail 'unexpected manifest fields'; fi
} < "$stage/pair/bundle.txt"
[ "$format" = moi-astra-native-v2 ] && [ "$target" = "$os-$arch" ] || fail 'incompatible protocol or target'
chmod 0755 "$stage/pair/astra" "$stage/pair/moi-cli"
[ "$("$stage/pair/astra" --version)" = "$astra_version" ] || fail 'Astra version does not match manifest'
[ "$("$stage/pair/moi-cli" --version)" = "$moi_version" ] || fail 'moi-cli version does not match manifest'
"$stage/pair/astra" auth credential --help | grep -q -- --output-fd || fail 'Astra native credential owner is missing'
"$stage/pair/moi-cli" -h 2>&1 | grep -q -- -auth || fail 'moi-cli native consumer is missing'
"$stage/pair/moi-cli" skill export --dir "$stage/pair/skills/moi-cli" || fail 'paired Skill export failed'
[ -f "$stage/pair/skills/moi-cli/SKILL.md" ] || fail 'paired Skill entry is missing'
release="$prefix/releases/$expected"
if [ -e "$release" ] || [ -L "$release" ]; then
    [ -d "$release" ] && [ ! -L "$release" ] || fail 'existing release is not an owned directory'
    for name in ASTRA-LICENSE astra bundle.txt moi-cli; do
        [ ! -L "$release/$name" ] && cmp -s "$stage/pair/$name" "$release/$name" || fail 'existing release was modified'
    done
    [ ! -L "$release/skills" ] && diff -r "$stage/pair/skills" "$release/skills" >/dev/null || fail 'existing release Skill was modified'
else
    mv "$stage/pair" "$release"
fi
mkdir -p "$skill_dir"
if [ ! -e "$skill_store" ]; then
    mkdir "$skill_store" || fail 'cannot create managed Skill store'
    printf '%s\n' "$prefix" > "$skill_store/owner"
fi
# Astra confines discovered Skill paths to their search root. Keep immutable
# copies inside that root instead of linking out to the binary install prefix.
if [ -e "$skill_target" ] || [ -L "$skill_target" ]; then
    [ -d "$skill_target" ] && [ ! -L "$skill_target" ] && diff -r "$release/skills/moi-cli" "$skill_target" >/dev/null || fail 'managed Skill was modified'
else
    skill_stage=$(mktemp -d "$skill_store/.install.XXXXXX")
    cp -R "$release/skills/moi-cli" "$skill_stage/moi-cli" || fail 'cannot stage global Skill'
    mv "$skill_stage/moi-cli" "$skill_target" || fail 'cannot publish global Skill'
    rmdir "$skill_stage"
    skill_stage=''
fi
if [ -e "$skill_link" ] || [ -L "$skill_link" ]; then
    [ -L "$skill_link" ] && [ "$(readlink "$skill_link")" = "$old_skill_target" ] || fail 'Skill target changed during installation'
elif [ -n "$old_skill_target" ]; then fail 'Skill target disappeared during installation'; fi
skill_next="$skill_dir/.moi-cli-next-$expected"
[ ! -e "$skill_next" ] && [ ! -L "$skill_next" ] || fail 'unexpected Skill staging pointer'
ln -s "$skill_target" "$skill_next"
if [ "$os" = darwin ]; then mv -fh "$skill_next" "$skill_link";
else mv -fT "$skill_next" "$skill_link"; fi
skill_next=''; skill_changed=yes
next="$prefix/.current-$expected"
[ ! -e "$next" ] && [ ! -L "$next" ] || fail 'unexpected staging pointer'
ln -s "releases/$expected" "$next"
# Rename, do not follow an existing directory symlink. Pair activation is one
# atomic operation; interruption cannot expose a mix of the two versions.
if [ "$os" = darwin ]; then mv -fh "$next" "$prefix/current";
else mv -fT "$next" "$prefix/current"; fi
next=''
activated=yes
printf 'Installed %s (%s). Previous releases were retained.\n' "$version" "$target"
printf 'MOI Skill: %s (paired with this release).\n' "$skill_link"
printf 'Add %s/current to PATH. Run astra --api-url <your trusted Astra URL> login once.\n' "$prefix"
printf '%s\n' 'Later Astra and moi-cli commands reuse that login. No credentials or existing Astra profiles were modified.'
