package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/xiaowumin-mark/FluxUI/system"
)

func maybeRunDocsBrowserToastActivator() bool {
	if !hasDocsBrowserArg("--fluxui-docs-toast-activator") {
		return false
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := system.RunToastActivator(ctx, docsSystemToastActivatorCLSID, func(event system.ToastActivationEvent) {
		fmt.Fprintf(os.Stdout, "FluxUI docs Toast activation: appID=%s arguments=%s input=%v\n", event.AppID, event.Arguments, event.UserInput)
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "FluxUI docs Toast activator stopped:", err)
	}
	return true
}

func hasDocsBrowserArg(target string) bool {
	for _, arg := range os.Args[1:] {
		if arg == target {
			return true
		}
	}
	return false
}
