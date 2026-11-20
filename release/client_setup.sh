# Embedded by channel_installer.py for production installs.
# This configures launchers, not another installer/updater or credential owner.
quote_shell() { printf "'"; printf '%s' "$1" | sed "s/'/'\\\\''/g"; printf "'"; }

launcher_conflict() {
  fail "CLI_INSTALL_LAUNCHER_CONFLICT: The existing MOI client command at $1 differs from this installer and was not overwritten. If you are switching environments, remove the old MOI client installation before retrying. To keep it, use --dir and --skill-dir with --no-modify-path."
}

user_command_conflict() {
  fail "CLI_INSTALL_COMMAND_CONFLICT: A command already exists at $1 and was not overwritten. To keep it, rerun with --no-modify-path and use the full command paths printed after installation. Otherwise, inspect the existing command before removing it."
}

profile_conflict() {
  fail "CLI_INSTALL_PROFILE_CONFLICT: The MOI PATH block in $1 differs from this installation and was not overwritten. Rerun with --no-modify-path to leave this file unchanged, or choose a different regular file with --shell-profile."
}

client_command_dir() {
  # A subprocess cannot change its parent's PATH. Reuse only conventional,
  # already-searchable user bin directories; never write a system bin directory.
  for candidate in "$HOME/.local/bin" "$HOME/bin"; do
    case ":$PATH:" in
      *":$candidate:"*)
        if [ -d "$candidate" ] && [ -w "$candidate" ] && [ ! -L "$candidate" ]; then
          printf '%s\n' "$candidate"
          return
        fi;;
    esac
  done
}

client_setup_preflight() {
  modify_path=${modify_path:-yes}
  case "$prefix" in /*) ;; *) fail 'installation directory must be absolute';; esac
  case "$prefix" in *'
'*|*''*) fail 'invalid installation directory';; esac
  case "$prefix" in *:*) fail 'installation directory must not contain a PATH separator';; esac
  case "$skill_dir" in
    /*/skills) astra_state_root=${skill_dir%/skills};;
    *) fail 'Skill directory must be an absolute Astra state directory ending in /skills';;
  esac
  [ -n "$astra_state_root" ] && [ "$astra_state_root" != / ] || fail 'invalid Astra state directory'
  case "$astra_state_root" in *'
'*) fail 'invalid Astra state directory';; esac
  [ ! -L "$prefix" ] || fail 'installation directory must not be a symlink'
  if [ -d "$prefix" ]; then prefix=$(cd "$prefix" && pwd -P); fi
  for directory in bin setup; do
    target="$prefix/$directory"
    [ ! -L "$target" ] || fail 'setup directory must not be a symlink'
    [ ! -e "$target" ] || [ -d "$target" ] || fail 'setup path is not a directory'
  done
  for binary in astra astra-edge moi-cli; do
    target="$prefix/bin/$binary"
    if [ -e "$target" ] || [ -L "$target" ]; then
      [ ! -L "$target" ] && [ -f "$target" ] && client_launcher "$binary" | cmp -s - "$target" || launcher_conflict "$target"
    fi
  done
  command_dir=''
  if [ "$modify_path" != no ]; then command_dir=$(client_command_dir); fi
  if [ -n "$command_dir" ]; then
    for binary in astra astra-edge moi-cli; do
      target="$command_dir/$binary"
      if [ -e "$target" ] || [ -L "$target" ]; then
        [ -L "$target" ] && [ "$(readlink "$target")" = "$prefix/bin/$binary" ] || user_command_conflict "$target"
      fi
    done
  fi
  profile_extra=''
  profile_skipped=''
  profile_syntax=posix
  case "${SHELL:-}" in */fish) profile_syntax=fish;; esac
  if [ "$modify_path" = no ]; then shell_profile=''; return; fi
  if [ -z "$shell_profile" ]; then
    case "${SHELL:-}" in
      */zsh) shell_profile="${ZDOTDIR:-$HOME}/.zshrc";;
      */bash)
        shell_profile="$HOME/.bashrc"
        # Do not create a higher-priority login file that hides an existing one.
        profile_extra="$HOME/.bash_profile"
        for login_file in "$HOME/.bash_profile" "$HOME/.bash_login" "$HOME/.profile"; do
          if [ -e "$login_file" ] || [ -L "$login_file" ]; then profile_extra=$login_file; break; fi
        done;;
      */fish) shell_profile="${XDG_CONFIG_HOME:-$HOME/.config}/fish/conf.d/moi-client.fish";;
      */sh|*/dash|*/ksh) shell_profile="$HOME/.profile";;
      *)
        printf '%s\n' 'Unknown login shell: leaving shell configuration unchanged; use the absolute commands printed below.' >&2
        return;;
    esac
  fi
  # A symlink is left untouched. Skipping it must not abort the install.
  profile_skipped=''
  kept_profile=''
  kept_extra=''
  for rc in "$shell_profile" "$profile_extra"; do
    [ -n "$rc" ] || continue
    case "$rc" in /*) ;; *) fail 'shell profile must be absolute';; esac
    case "$rc" in *'
'*) fail 'invalid shell profile path';; esac
    if [ -L "$rc" ]; then
      profile_skipped="${profile_skipped}${rc}
"
      continue
    fi
    [ ! -e "$rc" ] || [ -f "$rc" ] || fail 'shell profile is not a regular file'
    if [ "$profile_syntax" != fish ]; then
      [ -d "$(dirname "$rc")" ] || fail 'shell profile parent is missing'
    fi
    if [ -f "$rc" ] && grep -qF '# >>> MOI client >>>' "$rc"; then
      expected=$(client_path_block)
      actual=$(sed -n '/^# >>> MOI client >>>$/,/^# <<< MOI client <<<$/{p;}' "$rc")
      [ "$actual" = "$expected" ] || profile_conflict "$rc"
    fi
    if [ -z "$kept_profile" ]; then kept_profile=$rc; else kept_extra=$rc; fi
  done
  shell_profile=$kept_profile
  profile_extra=$kept_extra
}

client_path_block() {
  printf '%s\n' '# >>> MOI client >>>'
  if [ "$profile_syntax" = fish ]; then
    printf 'if not contains -- %s $PATH\n  set -gx PATH %s $PATH\nend\n' "$(quote_shell "$prefix/bin")" "$(quote_shell "$prefix/bin")"
  else
    printf 'case ":$PATH:" in *:%s:*) ;; *) export PATH=%s:"$PATH";; esac\n' "$(quote_shell "$prefix/bin")" "$(quote_shell "$prefix/bin")"
  fi
  printf '%s\n' '# <<< MOI client <<<'
}

client_launcher() {
  printf '%s\n' '#!/bin/sh' '# MOI client launcher. No credentials are embedded.' 'set -eu'
  printf 'client_root=%s\n' "$(quote_shell "$prefix")"
  printf 'export ASTRA_API_URL=%s\n' "$(quote_shell "$client_api_url")"
  printf 'export ASTRA_LOCAL_STATE_ROOT=%s\n' "$(quote_shell "$astra_state_root")"
  printf '%s\n' 'export MOI_ASTRA_BINARY="$client_root/bin/astra"'
  printf '%s\n' 'export PATH="$client_root/bin:$client_root/current:$PATH"'
  printf 'exec "$client_root/current/%s" "$@"\n' "$1"
}

client_setup() {
  # The updater already validated the prefix and owns all release mutations.
  prefix=$(cd "$prefix" && pwd -P)
  umask 077
  directories='bin setup'
  for directory in $directories; do
    target="$prefix/$directory"
    [ ! -L "$target" ] || fail 'setup directory must not be a symlink'
    if [ -e "$target" ]; then
      [ -d "$target" ] || fail 'setup path is not a directory'
    else
      mkdir "$target"
    fi
  done
  # Publish a complete PID before atomically linking the lock. A dead owner is
  # recoverable; a live owner is never displaced.
  setup_lock="$prefix/setup/setup.lock"
  setup_lock_candidate="$prefix/setup/.setup-lock.$$"
  profile_stage=''
  trap 'rm -f "$setup_lock_candidate" "$profile_stage" "$stage/component.tar.gz" "$stage/moi-cli" "$stage/launcher" "$stage/path-block"; rmdir "$stage"' EXIT
  rm -f "$setup_lock_candidate"
  printf '%s\n' "$$" > "$setup_lock_candidate"
  if ! ln "$setup_lock_candidate" "$setup_lock" 2>/dev/null; then
    [ ! -L "$setup_lock" ] && [ -f "$setup_lock" ] || fail 'unsafe client setup lock'
    setup_owner=$(cat "$setup_lock")
    case "$setup_owner" in ''|*[!0-9]*) fail 'invalid client setup lock';; esac
    if kill -0 "$setup_owner" 2>/dev/null; then fail 'client setup is already running'; fi
    rm -f "$setup_lock"
    ln "$setup_lock_candidate" "$setup_lock" 2>/dev/null || fail 'client setup lock changed; retry'
  fi
  rm -f "$setup_lock_candidate"
  trap '[ ! -f "$setup_lock" ] || [ "$(cat "$setup_lock")" != "$$" ] || rm -f "$setup_lock"; rm -f "$setup_lock_candidate" "$profile_stage" "$stage/component.tar.gz" "$stage/moi-cli" "$stage/launcher" "$stage/path-block"; rmdir "$stage"' EXIT
  for binary in astra astra-edge moi-cli; do
    client_launcher "$binary" > "$stage/launcher"
    chmod 0755 "$stage/launcher"
    target="$prefix/bin/$binary"
    if [ -e "$target" ] || [ -L "$target" ]; then
      [ ! -L "$target" ] && [ -f "$target" ] && cmp -s "$stage/launcher" "$target" || launcher_conflict "$target"
    else
      # stage can live on another filesystem; install through a prefix-local file.
      cp "$stage/launcher" "$prefix/setup/launcher.new"
      ln "$prefix/setup/launcher.new" "$target" || fail 'launcher target appeared during setup'
      rm "$prefix/setup/launcher.new"
    fi
  done
  client_path_block > "$stage/path-block"
  index=0
  for rc in "$shell_profile" "$profile_extra"; do
    [ -n "$rc" ] || continue
    index=$((index + 1))
    [ ! -L "$rc" ] || fail 'shell profile changed to a symlink'
    if [ -f "$rc" ] && grep -qF '# >>> MOI client >>>' "$rc"; then
      sed -n '/^# >>> MOI client >>>$/,/^# <<< MOI client <<<$/{p;}' "$rc" > "$prefix/setup/block.current"
      cmp -s "$stage/path-block" "$prefix/setup/block.current" || profile_conflict "$rc"
      rm "$prefix/setup/block.current"
      continue
    fi
    mkdir -p "$(dirname "$rc")"
    backup="$prefix/setup/shell-$index.before"
    # A later invocation can configure a different shell. Allocate another
    # slot under setup.lock, preserving every existing profile's backup. A
    # complete backup/path pair for this profile can be reused after an
    # interruption that happened before the atomic profile publication.
    reuse_backup=false
    while [ -e "$backup" ] || [ -L "$backup" ] || [ -e "$prefix/setup/shell-$index.path" ] || [ -L "$prefix/setup/shell-$index.path" ]; do
      saved_path="$prefix/setup/shell-$index.path"
      [ ! -L "$backup" ] && [ -f "$backup" ] && [ ! -L "$saved_path" ] && [ -f "$saved_path" ] || fail 'incomplete shell profile backup; inspect previous interrupted setup'
      if [ "$(cat "$saved_path")" = "$rc" ]; then reuse_backup=true; break; fi
      index=$((index + 1))
      backup="$prefix/setup/shell-$index.before"
    done
    if [ "$reuse_backup" = false ]; then
      if [ -f "$rc" ]; then cp -p "$rc" "$backup"; else : > "$backup"; fi
      printf '%s\n' "$rc" > "$prefix/setup/shell-$index.path"
    fi
    # Publish the complete profile on the profile filesystem so interruption
    # cannot leave a partial PATH block.
    profile_stage=$(mktemp "$(dirname "$rc")/.moi-profile.XXXXXX")
    if [ -f "$rc" ]; then cp -p "$rc" "$profile_stage"; fi
    { printf '\n'; cat "$stage/path-block"; } >> "$profile_stage"
    mv "$profile_stage" "$rc"
    profile_stage=''
  done
  if [ -n "$command_dir" ]; then
    printf '%s\n' "$command_dir" > "$prefix/setup/command-dir"
    for binary in astra astra-edge moi-cli; do
      target="$command_dir/$binary"
      if [ -e "$target" ] || [ -L "$target" ]; then
        [ -L "$target" ] && [ "$(readlink "$target")" = "$prefix/bin/$binary" ] || user_command_conflict "$target"
      else
        ln -s "$prefix/bin/$binary" "$target" || fail 'could not publish command entry'
      fi
    done
  fi
  ready=true
  for binary in astra astra-edge moi-cli; do
    resolved=$(command -v "$binary" || true)
    if [ "$resolved" != "$prefix/bin/$binary" ] && { [ -z "$command_dir" ] || [ "$resolved" != "$command_dir/$binary" ]; }; then ready=false; fi
  done
  printf '\nAstra and moi-cli are installed.\n'
  printf '%s\n' 'Sign in once with astra login to use both Astra and moi-cli.'
  if [ "$ready" = true ]; then
    printf '\nMOI customer client is ready in your current PATH. Run:\n'
  else
    printf '\nInstalled. This terminal does not yet resolve the new commands.\n'
    case "${SHELL:-}" in
      */sh|*/bash|*/zsh|*/dash|*/ksh)
        printf 'To use this terminal, copy and run:\n'
        printf 'export PATH=%s:"$PATH"\n' "$(quote_shell "$prefix/bin")"
        printf 'astra login\nastra\n';;
      */fish)
        printf 'To use this terminal, copy and run:\n'
        printf 'set -gx PATH %s $PATH\n' "$(quote_shell "$prefix/bin")"
        printf 'astra login\nastra\n';;
    esac
    printf 'Run now without changing PATH: %s login\n' "$(quote_shell "$prefix/bin/astra")"
    if [ -n "$shell_profile" ]; then
      printf 'Or open a NEW terminal, then run:\n'
    else
      printf 'Then start a conversation with: %s\n' "$(quote_shell "$prefix/bin/astra")"
    fi
  fi
  if [ -n "$profile_skipped" ]; then
    printf '%s\n' 'Shell profile is a symlink and was left unchanged:'
    printf '%s' "$profile_skipped"
    printf '%s\n' 'It was not followed or modified. Pass --shell-profile with a regular file to configure PATH there.'
  fi
  printf '%s\n' 'Use your current directory or any project directory you choose.'
  if [ "$ready" = true ] || [ -n "$shell_profile" ]; then printf '%s\n' 'astra login' 'astra'; fi
  printf '\nIn Astra, try: List my MOI workspaces\n'
  printf '%s\n' 'Astra can use moi-cli for MOI tasks; no additional setup is needed.'
  printf '\nOr use moi-cli directly after signing in:\n'
  if [ "$ready" = true ]; then
    printf '%s\n' 'moi-cli workspace list'
  else
    printf '%s workspace list\n' "$(quote_shell "$prefix/bin/moi-cli")"
  fi
  printf '\nClient files: %s\n' "$prefix"
  printf '%s\n' 'Login state is managed by Astra in ~/.moi (or your explicit MOI_AUTH_DIR).'
  if [ -n "$shell_profile" ]; then printf '%s\n' 'The shell PATH block has a backup under setup/.'; fi
  printf '%s\n' 'Existing global credentials/config were not changed.'
}
