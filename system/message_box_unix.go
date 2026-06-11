//go:build darwin || linux

package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type unixMessageBoxCommand struct {
	name  string
	args  []string
	parse func([]byte, error) (MessageBoxResult, error)
}

type unixMessageBoxProvider struct {
	name     string
	commands []string
	build    func(messageBoxOptions) (unixMessageBoxCommand, error)
}

type unixMessageBoxButton struct {
	label  string
	result MessageBoxResult
	cancel bool
}

func (unixDriver) showMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	return showUnixMessageBox(ctx, opts)
}

func (unixDriver) showDetailedMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxDetailedResult, error) {
	result, err := showUnixMessageBox(ctx, opts)
	if err != nil {
		return MessageBoxDetailedResult{}, err
	}
	return MessageBoxDetailedResult{
		Result:   result,
		ButtonID: string(result),
	}, nil
}

func showUnixMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := validateUnixMessageBoxOptions(opts); err != nil {
		return "", err
	}

	provider, err := selectUnixMessageBoxProvider()
	if err != nil {
		return "", fmt.Errorf("system: %s: show: %w: %v", CapabilityMessageBox, ErrUnavailable, err)
	}
	command, err := provider.build(opts)
	if err != nil {
		return "", err
	}
	return runUnixMessageBox(ctx, provider, command)
}

func runUnixMessageBox(ctx context.Context, provider unixMessageBoxProvider, command unixMessageBoxCommand) (MessageBoxResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if command.name == "" {
		return "", fmt.Errorf("system: %s: %s command is empty: %w", CapabilityMessageBox, provider.name, ErrUnavailable)
	}

	cmd := exec.CommandContext(ctx, command.name, command.args...)
	output, err := cmd.CombinedOutput()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", ctxErr
	}
	if command.parse == nil {
		if err != nil {
			return "", unixMessageBoxCommandError(provider.name, err, output)
		}
		return MessageBoxResultOK, nil
	}
	return command.parse(output, err)
}

func selectUnixMessageBoxProvider() (unixMessageBoxProvider, error) {
	return unixMessageBoxProviderWithLookPath(exec.LookPath)
}

func unixMessageBoxProviderWithLookPath(lookPath unixLookPathFunc) (unixMessageBoxProvider, error) {
	candidates := unixMessageBoxProviders()
	var missing []string
	for _, candidate := range candidates {
		ok := true
		for _, name := range candidate.commandNames() {
			if _, err := lookPath(name); err != nil {
				missing = append(missing, name)
				ok = false
				break
			}
		}
		if ok {
			return candidate, nil
		}
	}
	if len(missing) == 0 {
		return unixMessageBoxProvider{}, fmt.Errorf("no message box providers configured")
	}
	return unixMessageBoxProvider{}, fmt.Errorf("message box commands unavailable: %s", strings.Join(uniqueStrings(missing), ", "))
}

func (p unixMessageBoxProvider) commandNames() []string {
	return p.commands
}

func validateUnixMessageBoxOptions(opts messageBoxOptions) error {
	if feature := unsupportedUnixMessageBoxOption(opts); feature != "" {
		return fmt.Errorf("system: %s: %s is unsupported on this platform: %w", CapabilityMessageBox, feature, ErrUnsupported)
	}
	if err := validateUnixMessageBoxKind(opts.kind); err != nil {
		return err
	}
	if _, err := unixMessageBoxButtons(opts.buttons); err != nil {
		return err
	}
	if opts.defaultButton != "" && !unixMessageBoxHasResult(opts.buttons, opts.defaultButton) {
		return fmt.Errorf("system: default button %q is not in button set %q", opts.defaultButton, opts.buttons)
	}
	return nil
}

func unsupportedUnixMessageBoxOption(opts messageBoxOptions) string {
	switch {
	case opts.details != "":
		return "details"
	case opts.footer != "":
		return "footer"
	case opts.verificationText != "" || opts.verificationChecked:
		return "verification"
	case len(opts.customButtons) > 0:
		return "custom buttons"
	case opts.defaultButtonID != "":
		return "default custom button"
	case opts.commandLinks || opts.commandLinksNoIcon:
		return "command links"
	case opts.expandedDetailsByDefault || opts.expandDetailsInFooterArea:
		return "expanded details"
	case opts.owner != 0:
		return "owner window"
	default:
		return ""
	}
}

func validateUnixMessageBoxKind(kind MessageBoxKind) error {
	switch kind {
	case "", MessageBoxInfo, MessageBoxWarning, MessageBoxError, MessageBoxQuestion:
		return nil
	default:
		return fmt.Errorf("system: unknown message box kind %q", kind)
	}
}

func unixMessageBoxButtons(buttons MessageBoxButtons) ([]unixMessageBoxButton, error) {
	switch buttons {
	case "", MessageBoxOK:
		return []unixMessageBoxButton{
			{label: "OK", result: MessageBoxResultOK},
		}, nil
	case MessageBoxOKCancel:
		return []unixMessageBoxButton{
			{label: "Cancel", result: MessageBoxResultCancel, cancel: true},
			{label: "OK", result: MessageBoxResultOK},
		}, nil
	case MessageBoxYesNo:
		return []unixMessageBoxButton{
			{label: "No", result: MessageBoxResultNo},
			{label: "Yes", result: MessageBoxResultYes},
		}, nil
	case MessageBoxYesNoCancel:
		return []unixMessageBoxButton{
			{label: "Cancel", result: MessageBoxResultCancel, cancel: true},
			{label: "No", result: MessageBoxResultNo},
			{label: "Yes", result: MessageBoxResultYes},
		}, nil
	case MessageBoxRetryCancel:
		return []unixMessageBoxButton{
			{label: "Cancel", result: MessageBoxResultCancel, cancel: true},
			{label: "Retry", result: MessageBoxResultRetry},
		}, nil
	default:
		return nil, fmt.Errorf("system: unknown message box button set %q", buttons)
	}
}

func unixMessageBoxHasResult(buttons MessageBoxButtons, result MessageBoxResult) bool {
	_, ok := unixMessageBoxLabelForResult(buttons, result)
	return ok
}

func unixMessageBoxLabelForResult(buttons MessageBoxButtons, result MessageBoxResult) (string, bool) {
	specs, err := unixMessageBoxButtons(buttons)
	if err != nil {
		return "", false
	}
	for _, spec := range specs {
		if spec.result == result {
			return spec.label, true
		}
	}
	return "", false
}

func unixMessageBoxResultForLabel(label string) (MessageBoxResult, bool) {
	switch strings.TrimSpace(label) {
	case "OK":
		return MessageBoxResultOK, true
	case "Cancel":
		return MessageBoxResultCancel, true
	case "Yes":
		return MessageBoxResultYes, true
	case "No":
		return MessageBoxResultNo, true
	case "Retry":
		return MessageBoxResultRetry, true
	default:
		return "", false
	}
}

func unixMessageBoxDefaultLabel(opts messageBoxOptions) (string, error) {
	label, ok := unixMessageBoxLabelForResult(opts.buttons, opts.defaultButton)
	if ok {
		return label, nil
	}
	buttons, err := unixMessageBoxButtons(opts.buttons)
	if err != nil {
		return "", err
	}
	if len(buttons) == 0 {
		return "", fmt.Errorf("system: message box button set %q is empty", opts.buttons)
	}
	return buttons[len(buttons)-1].label, nil
}

func unixMessageBoxCancelLabel(buttons MessageBoxButtons) (string, bool) {
	specs, err := unixMessageBoxButtons(buttons)
	if err != nil {
		return "", false
	}
	for _, spec := range specs {
		if spec.cancel {
			return spec.label, true
		}
	}
	return "", false
}

func unixMessageBoxCommandError(provider string, err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	if detail != "" {
		return fmt.Errorf("system: %s: run %s: %w: %v: %s", CapabilityMessageBox, provider, ErrUnavailable, err, detail)
	}
	return fmt.Errorf("system: %s: run %s: %w: %v", CapabilityMessageBox, provider, ErrUnavailable, err)
}
