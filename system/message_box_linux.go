//go:build linux

package system

import (
	"fmt"
	"os/exec"
	"strings"
)

func unixMessageBoxProviders() []unixMessageBoxProvider {
	return []unixMessageBoxProvider{
		{
			name:     "zenity",
			commands: []string{"zenity"},
			build:    linuxZenityMessageBoxCommand,
		},
		{
			name:     "kdialog",
			commands: []string{"kdialog"},
			build:    linuxKdialogMessageBoxCommand,
		},
	}
}

func linuxZenityMessageBoxCommand(opts messageBoxOptions) (unixMessageBoxCommand, error) {
	args := []string{linuxZenityDialogKind(opts)}
	args = append(args, "--modal", "--title", opts.title, "--text", opts.text)

	switch opts.buttons {
	case "", MessageBoxOK:
		args = append(args, "--ok-label", "OK")
	case MessageBoxOKCancel:
		args = append(args, "--ok-label", "OK", "--cancel-label", "Cancel")
	case MessageBoxYesNo:
		args = append(args, "--ok-label", "Yes", "--cancel-label", "No")
	case MessageBoxYesNoCancel:
		args = append(args, "--ok-label", "Yes", "--cancel-label", "Cancel", "--extra-button", "No")
	case MessageBoxRetryCancel:
		args = append(args, "--ok-label", "Retry", "--cancel-label", "Cancel")
	default:
		return unixMessageBoxCommand{}, fmt.Errorf("system: unknown message box button set %q", opts.buttons)
	}

	if linuxZenityDefaultIsCancel(opts) {
		args = append(args, "--default-cancel")
	}

	return unixMessageBoxCommand{
		name: "zenity",
		args: args,
		parse: func(output []byte, err error) (MessageBoxResult, error) {
			return linuxZenityParseMessageBoxResult(opts.buttons, output, err)
		},
	}, nil
}

func linuxZenityDialogKind(opts messageBoxOptions) string {
	if opts.buttons != "" && opts.buttons != MessageBoxOK {
		return "--question"
	}
	switch opts.kind {
	case MessageBoxWarning:
		return "--warning"
	case MessageBoxError:
		return "--error"
	default:
		return "--info"
	}
}

func linuxZenityDefaultIsCancel(opts messageBoxOptions) bool {
	switch opts.defaultButton {
	case MessageBoxResultCancel:
		return true
	case MessageBoxResultNo:
		return opts.buttons == MessageBoxYesNo
	default:
		return false
	}
}

func linuxZenityParseMessageBoxResult(buttons MessageBoxButtons, output []byte, err error) (MessageBoxResult, error) {
	text := strings.TrimSpace(string(output))
	if err == nil {
		if text != "" {
			if result, ok := unixMessageBoxResultForLabel(text); ok {
				return result, nil
			}
			return "", fmt.Errorf("system: %s: unexpected zenity result %q", CapabilityMessageBox, text)
		}
		return linuxZenityPrimaryResult(buttons), nil
	}
	if exit, ok := err.(*exec.ExitError); ok {
		switch exit.ExitCode() {
		case 1:
			return linuxZenityExitOneResult(buttons), nil
		case 5:
			return MessageBoxResultCancel, nil
		}
	}
	return "", unixMessageBoxCommandError("zenity", err, output)
}

func linuxZenityPrimaryResult(buttons MessageBoxButtons) MessageBoxResult {
	switch buttons {
	case MessageBoxYesNo, MessageBoxYesNoCancel:
		return MessageBoxResultYes
	case MessageBoxRetryCancel:
		return MessageBoxResultRetry
	default:
		return MessageBoxResultOK
	}
}

func linuxZenityExitOneResult(buttons MessageBoxButtons) MessageBoxResult {
	switch buttons {
	case MessageBoxYesNo:
		return MessageBoxResultNo
	default:
		return MessageBoxResultCancel
	}
}

func linuxKdialogMessageBoxCommand(opts messageBoxOptions) (unixMessageBoxCommand, error) {
	args := []string{"--title", opts.title}
	switch opts.buttons {
	case "", MessageBoxOK:
		dialogKind, err := linuxKdialogMessageKind(opts.kind)
		if err != nil {
			return unixMessageBoxCommand{}, err
		}
		args = append(args, dialogKind, opts.text)
	case MessageBoxOKCancel:
		args = append(args, "--yes-label", "OK", "--no-label", "Cancel", "--yesno", opts.text)
	case MessageBoxYesNo:
		args = append(args, "--yesno", opts.text)
	case MessageBoxYesNoCancel:
		args = append(args, "--yesnocancel", opts.text)
	case MessageBoxRetryCancel:
		args = append(args, "--yes-label", "Retry", "--no-label", "Cancel", "--yesno", opts.text)
	default:
		return unixMessageBoxCommand{}, fmt.Errorf("system: unknown message box button set %q", opts.buttons)
	}

	return unixMessageBoxCommand{
		name: "kdialog",
		args: args,
		parse: func(output []byte, err error) (MessageBoxResult, error) {
			return linuxKdialogParseMessageBoxResult(opts.buttons, output, err)
		},
	}, nil
}

func linuxKdialogMessageKind(kind MessageBoxKind) (string, error) {
	switch kind {
	case "", MessageBoxInfo, MessageBoxQuestion:
		return "--msgbox", nil
	case MessageBoxWarning:
		return "--sorry", nil
	case MessageBoxError:
		return "--error", nil
	default:
		return "", fmt.Errorf("system: unknown message box kind %q", kind)
	}
}

func linuxKdialogParseMessageBoxResult(buttons MessageBoxButtons, output []byte, err error) (MessageBoxResult, error) {
	if err == nil {
		return linuxKdialogExitZeroResult(buttons), nil
	}
	if exit, ok := err.(*exec.ExitError); ok {
		switch exit.ExitCode() {
		case 1:
			return linuxKdialogExitOneResult(buttons), nil
		case 2:
			return MessageBoxResultCancel, nil
		}
	}
	return "", unixMessageBoxCommandError("kdialog", err, output)
}

func linuxKdialogExitZeroResult(buttons MessageBoxButtons) MessageBoxResult {
	switch buttons {
	case MessageBoxYesNo, MessageBoxYesNoCancel:
		return MessageBoxResultYes
	case MessageBoxRetryCancel:
		return MessageBoxResultRetry
	default:
		return MessageBoxResultOK
	}
}

func linuxKdialogExitOneResult(buttons MessageBoxButtons) MessageBoxResult {
	switch buttons {
	case MessageBoxYesNo, MessageBoxYesNoCancel:
		return MessageBoxResultNo
	default:
		return MessageBoxResultCancel
	}
}
