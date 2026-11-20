package update

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"time"
)

const usage = `Usage: astra update [check | --rollback]
       moi-cli update [check | --rollback]
		moi-cli update install --dir DIRECTORY --channel HTTPS_URL --compatibility REFERENCE --astra-url HTTPS_ORIGIN [--skill-dir DIRECTORY]
Only MOI-managed installations are updated. Running clients must exit before switching.
Metadata checks may briefly wait for another check; retry if CLI_UPDATE_CHECK_BUSY is reported.
`

func Run(ctx context.Context, args []string, out io.Writer) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		_, err := fmt.Fprint(out, usage)
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	var m *Manager
	if len(args) > 0 && args[0] == "install" {
		f := flag.NewFlagSet("update install", flag.ContinueOnError)
		f.SetOutput(io.Discard)
		root := f.String("dir", "", "installation prefix")
		channel := f.String("channel", "", "trusted channel URL")
		skills := f.String("skill-dir", "", "Skill search root")
		compatibility := f.String("compatibility", "", "accepted release compatibility reference")
		astraURL := f.String("astra-url", "", "trusted Astra service origin")
		local := f.Bool("allow-loopback-http", false, "local acceptance only")
		if f.Parse(args[1:]) != nil || f.NArg() != 0 || *root == "" || *channel == "" || *compatibility == "" || *astraURL == "" {
			return invalid("INSTALL_ARGUMENTS")
		}
		if *skills == "" {
			home, e := os.UserHomeDir()
			if e != nil {
				return e
			}
			*skills = filepath.Join(home, ".astra", "skills")
		}
		m, err = Init(*root, Installation{Channel: *channel, SkillRoot: *skills, Compatibility: *compatibility, AstraURL: *astraURL, Loopback: *local})
		if err == nil {
			err = prepareAuthDirectory()
		}
		args = nil
	} else {
		m, err = Discover(exe)
		if err == nil && m == nil {
			return invalid("NOT_MANAGED")
		}
	}
	if err != nil {
		return err
	}
	// stderr remains a terminal for curl | sh installs; stdin need not be one.
	if !(len(args) == 1 && args[0] == "refresh") && progressEnabled() {
		m.progress = &progressDisplay{out: os.Stderr}
		defer m.progress.clear()
	}
	var result Result
	switch {
	case len(args) == 0:
		result, err = m.Apply(ctx, false)
	case len(args) == 1 && args[0] == "--rollback":
		result, err = m.Apply(ctx, true)
	case len(args) == 1 && args[0] == "check":
		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		result, err = m.Check(checkCtx)
	case len(args) == 1 && args[0] == "refresh":
		// Private best-effort startup worker: no stdout/stderr, no installation.
		m.refresh()
		return nil
	default:
		return invalid("ARGUMENTS")
	}
	if err != nil {
		return err
	}
	m.progress.clear()
	return json.NewEncoder(out).Encode(result)
}

func (m *Manager) refresh() {
	var cache Cache
	if readJSON(m.path("cache.json"), &cache) == nil {
		if age := time.Since(cache.Checked); age >= 0 && age < time.Hour {
			return
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _, _, _ = m.check(ctx, false)
}
func MachineCall(args []string) bool {
	if os.Getenv("MOI_AGENT_CALL") == "1" {
		return true
	}
	for _, a := range args {
		switch a {
		case "-o", "--json", "-no-interactive", "--no-interactive", "--version", "-version", "--help", "-help", "-h", "schema", "skill", "config", "auth", "--moi-update-protocol":
			return true
		}
	}
	return false
}
func Startup(exe string, args []string) {
	if MachineCall(args) {
		return
	}
	if !terminal(os.Stderr.Fd()) || !terminal(os.Stdin.Fd()) || os.Getenv("TERM") == "dumb" {
		return
	}
	m, err := Discover(exe)
	if err != nil || m == nil {
		return
	}
	var c Cache
	s, err := m.state()
	if err == nil && readJSON(m.path("cache.json"), &c) == nil {
		if notice := cachedNotice(c, s, time.Now()); notice != "" {
			fmt.Fprintln(os.Stderr, notice)
		}
	}
	// Short CLI requests do not await this goroutine or keep a daemon alive.
	go m.refresh()
}

func cachedNotice(c Cache, s State, now time.Time) string {
	if !idPattern.MatchString(c.Channel.Bundle) || !c.Channel.Expires.After(now) || c.Channel.Hash == s.Current {
		return ""
	}
	if c.Reason == invalid("COMPATIBILITY_CHANGE").Error() {
		return fmt.Sprintf("MOI new release: %s. %s", c.Channel.Bundle, updateMessage(invalid("COMPATIBILITY_CHANGE")))
	}
	if c.Available {
		return fmt.Sprintf("MOI update available: %s. Run astra update.", c.Channel.Bundle)
	}
	return ""
}

// Stable reason-to-display mapping for local installation errors. Remote API
// errors are not classified here. Never put raw paths or OS causes in messages.
func updateMessage(err error) string {
	switch err.Error() {
	case "CLI_UPDATE_PLATFORM_UNSUPPORTED":
		return "This release is not available for your operating system and architecture. Keep your current installation and retry when a compatible release is available."
	case "CLI_UPDATE_AUTH_DIRECTORY":
		return "The authentication directory could not be prepared. Check MOI_AUTH_DIR (default: ~/.moi): use an absolute path without dot segments to a directory owned by you, not a symbolic link, with permissions 0700. No credentials were changed."
	case "CLI_UPDATE_COMPATIBILITY_CHANGE":
		return "This release requires a separate installation. Use the official installer with a new --dir and --skill-dir; keep the current installation and login data."
	case "CLI_UPDATE_CHECK_BUSY":
		return "Another update check is in progress. Wait a few seconds and retry."
	case "CLI_UPDATE_CHECK_CANCELLED":
		return "The update check was cancelled or timed out while waiting. Retry the command."
	case "CLI_UPDATE_BUSY":
		return "This installation is in use. Exit its Astra, moi-cli and local astra-edge processes, then retry."
	case "CLI_UPDATE_SKILL_COPY_MISSING":
		return "The managed Skill copy is missing. Run astra update to restore it from the verified local release."
	case "CLI_UPDATE_LOCAL_SKILL_MODIFIED":
		return "The managed Skill copy differs from its verified source and was not overwritten. Preserve local edits before reinstalling."
	case "CLI_UPDATE_POINTER_STATE_MISMATCH":
		return "The local MOI client files do not match the saved installation state. Close Astra, astra-edge and moi-cli, move the client installation directory aside (default: ~/.local/share/moi/client), then run the official installer again. Your sign-in data does not need to be removed."
	case "CLI_UPDATE_INSTALLATION_CONFLICT":
		return "This directory belongs to a MOI client installed with different settings. To keep it, install the new environment with separate --dir and --skill-dir and --no-modify-path."
	case "CLI_UPDATE_INSTALLATION_NOT_EMPTY", "CLI_UPDATE_DEDICATED_DIRECTORY_REQUIRED":
		return "Choose an empty, dedicated directory for the MOI client installation. Check the existing files before moving or deleting them."
	case "CLI_UPDATE_SKILL_NOT_OWNED", "CLI_UPDATE_UNOWNED_POINTER":
		return "An existing MOI client file or Skill belongs to another installation and was not changed. Inspect it before removing it, or choose separate --dir and --skill-dir."
	case "CLI_UPDATE_STATE_CONTRACT", "CLI_UPDATE_TRANSACTION_CONTRACT", "CLI_UPDATE_INSTALLATION_CONTRACT", "CLI_UPDATE_LOCAL_RELEASE_MODIFIED", "CLI_UPDATE_LOCAL_MANIFEST_CHECKSUM":
		return "The local MOI client installation is incomplete or has changed. Close Astra, astra-edge and moi-cli, move the client installation directory aside, then rerun the official installer with the same options. Your sign-in data does not need to be removed."
	case "CLI_UPDATE_RECOVERY_REQUIRED":
		return "A previous client update was interrupted. Run astra update to recover it before using the client."
	case "CLI_UPDATE_RESTART_REQUIRED":
		return "This client belongs to an older installation. Exit it and start Astra or moi-cli again."
	case "CLI_UPDATE_NOT_MANAGED":
		return "This client was not installed by the MOI installer. Use the official installer to enable managed updates."
	case "CLI_UPDATE_NO_PREVIOUS_RELEASE":
		return "There is no previous client release to restore. The current installation was not changed."
	case "CLI_UPDATE_DOWNLOAD_UNAVAILABLE":
		return "The MOI client could not reach the release server. Check your network connection and retry."
	case "CLI_UPDATE_CHANNEL_EXPIRED":
		return "The published client release information has expired. Retry later or contact support with this error code."
	case "CLI_UPDATE_CHANNEL_REPLAY", "CLI_UPDATE_VERSION_CONFLICT", "CLI_UPDATE_DOWNGRADE_REJECTED":
		return "This release cannot safely replace the installed version. Keep the current installation and contact support with this error code."
	case "CLI_UPDATE_INSTALL_ARGUMENTS", "CLI_UPDATE_ARGUMENTS":
		return "The update command is incomplete or invalid. Run moi-cli update --help to see the supported options."
	default:
		return ""
	}
}

// Keep raw I/O/network causes out of the public CLI error surface.
func PublicError(err error) string {
	var probeErr *versionProbeError
	if errors.As(err, &probeErr) && slices.Contains(names, probeErr.component) {
		return "CLI_UPDATE_VERSION_PROBE: " + probeErr.component + " could not start. The new release was not activated and the current installation is unchanged. Retry the update, or install a newer bundle. The loader diagnostic is in the installation probe log."
	}
	var coded updateError
	if errors.As(err, &coded) {
		message := updateMessage(coded)
		if message == "" {
			message = "The MOI client could not complete the installation or update. Contact support with this error code if the problem persists."
		}
		return coded.Error() + ": " + message
	}
	if errors.Is(err, os.ErrPermission) {
		return "CLI_UPDATE_LOCAL_PERMISSION: installation or Skill files are not accessible; check directory permissions"
	}
	if errors.Is(err, syscall.ENOSPC) {
		return "CLI_UPDATE_LOCAL_DISK_FULL: insufficient disk space; free space and retry without deleting the current or previous release"
	}
	if isMissing(err) {
		return "CLI_UPDATE_LOCAL_FILE_MISSING: required installation files are missing; preserve login data and reinstall using the official installer"
	}
	return "CLI_UPDATE_LOCAL_IO: installation files could not be read or written"
}
