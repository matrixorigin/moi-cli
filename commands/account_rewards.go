package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
)

func runAccountServiceRegistrationInvite(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account registration-invite", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli account registration-invite [options]\n\nProduct SDK: AccountService.RegistrationInvite\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	result, err := ctx.Client.Account().RegistrationInvite(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceRewardCenterItems(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account reward-center-items", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli account reward-center-items [options]\n\nProduct SDK: AccountService.RewardCenterItems\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	result, err := ctx.Client.Account().RewardCenterItems(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceRewardCenterInvitations(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account reward-center-invitations", flag.ContinueOnError)
	page := fs.String("page", "", "Page number")
	pageSize := fs.String("page-size", "", "Page size")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli account reward-center-invitations [options]\n\nProduct SDK: AccountService.RewardCenterInvitations\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	result, err := ctx.Client.Account().RewardCenterInvitations(context.Background(), *page, *pageSize)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAccountServiceRewardCenterEvents(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("account reward-center-events", flag.ContinueOnError)
	eventType := fs.String("event-type", "", "Reward event type filter")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli account reward-center-events [options]\n\nProduct SDK: AccountService.RewardCenterEvents\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	result, err := ctx.Client.Account().RewardCenterEvents(context.Background(), *eventType)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
