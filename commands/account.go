package commands

import (
	sdk "github.com/matrixorigin/moi-go-sdk/v5"

	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
)

// ExecuteAccountCommand runs product SDK AccountService methods via moi-cli.
func ExecuteAccountCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowAccountHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "checkin":
		return runAccountServiceCheckin(ctx, args[1:])
	case "bootstrap-native":
		return runAccountServiceBootstrapNative(ctx, args[1:])
	case "registration-invite":
		return runAccountServiceRegistrationInvite(ctx, args[1:])
	case "reward-center-events":
		return runAccountServiceRewardCenterEvents(ctx, args[1:])
	case "reward-center-invitations":
		return runAccountServiceRewardCenterInvitations(ctx, args[1:])
	case "reward-center-items":
		return runAccountServiceRewardCenterItems(ctx, args[1:])
	case "current-credit":
		return runAccountServiceCurrentCredit(ctx, args[1:])
	case "checkin-status":
		return runAccountServiceCheckinStatus(ctx, args[1:])
	case "update-preferences":
		return runAccountServiceUpdatePreferences(ctx, args[1:])
	case "preferences":
		return runAccountServicePreferences(ctx, args[1:])
	case "logout-browser-session":
		return runAccountServiceLogoutBrowserSession(ctx, args[1:])
	case "browser-session":
		return runAccountServiceBrowserSession(ctx, args[1:])
	case "complete-browser-logout":
		return runAccountServiceCompleteBrowserLogout(ctx, args[1:])
	case "complete-browser-login":
		return runAccountServiceCompleteBrowserLogin(ctx, args[1:])
	case "begin-browser-login":
		return runAccountServiceBeginBrowserLogin(ctx, args[1:])
	case "db-connection":
		return runAccountServiceDBConnection(ctx, args[1:])
	default:
		return fmt.Errorf("unknown account subcommand: %s", sub)
	}
}

// ShowAccountHelp prints help for account.
func ShowAccountHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: checkin")

	fmt.Fprintf(os.Stderr, `Usage: account <subcommand> [options]

Product SDK surface: AccountService

Subcommands:
  bootstrap-native                AccountService.BootstrapNative
  registration-invite             AccountService.RegistrationInvite
  reward-center-events             AccountService.RewardCenterEvents
  reward-center-invitations        AccountService.RewardCenterInvitations
  reward-center-items              AccountService.RewardCenterItems
  current-credit                   AccountService.CurrentCredit
  checkin-status                   AccountService.CheckinStatus
  update-preferences               AccountService.UpdatePreferences
  preferences                      AccountService.Preferences
  logout-browser-session           AccountService.LogoutBrowserSession
  browser-session                  AccountService.BrowserSession
  complete-browser-logout          AccountService.CompleteBrowserLogout
  complete-browser-login           AccountService.CompleteBrowserLogin
  begin-browser-login              AccountService.BeginBrowserLogin
  db-connection                    AccountService.DBConnection

`)
}

func runAccountServiceBootstrapNative(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account bootstrap-native", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli account bootstrap-native\n\nProduct SDK: AccountService.BootstrapNative")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	result, err := ctx.Client.Account().BootstrapNative(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceDBConnection(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account db-connection", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account db-connection [options]\n\nProduct SDK: AccountService.DBConnection\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Account().DBConnection(context.Background(), wid)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceBeginBrowserLogin(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account begin-browser-login", flag.ExitOnError)
	returnTo := fs.String("return-to", "", "return-to")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account begin-browser-login [options]\n\nProduct SDK: AccountService.BeginBrowserLogin\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Account().BeginBrowserLogin(context.Background(), *returnTo)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return printJSON(map[string]any{"status": result.StatusCode, "location": result.Header.Get("Location")})
}

func runAccountServiceCompleteBrowserLogin(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account complete-browser-login", flag.ExitOnError)
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account complete-browser-login [options]\n\nProduct SDK: AccountService.CompleteBrowserLogin\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.BrowserLoginCallbackRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Account().CompleteBrowserLogin(context.Background(), &request)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return printJSON(map[string]any{"status": result.StatusCode, "location": result.Header.Get("Location")})
}

func runAccountServiceCompleteBrowserLogout(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account complete-browser-logout", flag.ExitOnError)
	state := fs.String("state", "", "state")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account complete-browser-logout [options]\n\nProduct SDK: AccountService.CompleteBrowserLogout\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Account().CompleteBrowserLogout(context.Background(), *state)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return printJSON(map[string]any{"status": result.StatusCode, "location": result.Header.Get("Location")})
}

func runAccountServiceBrowserSession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account browser-session", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account browser-session [options]\n\nProduct SDK: AccountService.BrowserSession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Account().BrowserSession(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceLogoutBrowserSession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account logout-browser-session", flag.ExitOnError)
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account logout-browser-session [options]\n\nProduct SDK: AccountService.LogoutBrowserSession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.BrowserLogoutRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Account().LogoutBrowserSession(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServicePreferences(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account preferences", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account preferences [options]\n\nProduct SDK: AccountService.Preferences\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Account().Preferences(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceUpdatePreferences(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account update-preferences", flag.ExitOnError)
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account update-preferences [options]\n\nProduct SDK: AccountService.UpdatePreferences\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.AuthPreferencesRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Account().UpdatePreferences(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceCheckinStatus(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account checkin-status", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account checkin-status [options]\n\nProduct SDK: AccountService.CheckinStatus\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Account().CheckinStatus(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceCurrentCredit(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account current-credit", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli account current-credit [options]\n\nProduct SDK: AccountService.CurrentCredit\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Account().CurrentCredit(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceCheckin(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account checkin", flag.ExitOnError)
	setSDKCommandUsage(fs)
	source := fs.String("source", "", "source")
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Account().Checkin(context.Background(), *source)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
