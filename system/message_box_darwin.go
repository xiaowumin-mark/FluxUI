//go:build darwin

package system

import (
	"fmt"
	"strings"
)

func unixMessageBoxProviders() []unixMessageBoxProvider {
	return []unixMessageBoxProvider{
		{
			name:     "osascript",
			commands: []string{"osascript"},
			build:    darwinMessageBoxCommand,
		},
	}
}

func darwinMessageBoxCommand(opts messageBoxOptions) (unixMessageBoxCommand, error) {
	icon, err := darwinMessageBoxIcon(opts.kind)
	if err != nil {
		return unixMessageBoxCommand{}, err
	}
	defaultLabel, err := unixMessageBoxDefaultLabel(opts)
	if err != nil {
		return unixMessageBoxCommand{}, err
	}
	buttons, err := unixMessageBoxButtons(opts.buttons)
	if err != nil {
		return unixMessageBoxCommand{}, err
	}

	buttonLabels := make([]string, 0, len(buttons))
	for _, button := range buttons {
		buttonLabels = append(buttonLabels, darwinAppleScriptQuote(button.label))
	}

	parts := []string{
		"display dialog " + darwinAppleScriptQuote(opts.text),
		"with title " + darwinAppleScriptQuote(opts.title),
		"buttons {" + strings.Join(buttonLabels, ", ") + "}",
		"default button " + darwinAppleScriptQuote(defaultLabel),
	}
	if cancelLabel, ok := unixMessageBoxCancelLabel(opts.buttons); ok {
		parts = append(parts, "cancel button "+darwinAppleScriptQuote(cancelLabel))
	}
	parts = append(parts, "with icon "+icon)

	return unixMessageBoxCommand{
		name: "osascript",
		args: []string{"-e", strings.Join(parts, " ")},
		parse: func(output []byte, err error) (MessageBoxResult, error) {
			return darwinParseMessageBoxResult(output, err)
		},
	}, nil
}

func darwinMessageBoxIcon(kind MessageBoxKind) (string, error) {
	switch kind {
	case "", MessageBoxInfo, MessageBoxQuestion:
		return "note", nil
	case MessageBoxWarning:
		return "caution", nil
	case MessageBoxError:
		return "stop", nil
	default:
		return "", fmt.Errorf("system: unknown message box kind %q", kind)
	}
}

func darwinParseMessageBoxResult(output []byte, err error) (MessageBoxResult, error) {
	text := strings.TrimSpace(string(output))
	if err != nil {
		if strings.Contains(strings.ToLower(text), "user canceled") || strings.Contains(text, "-128") {
			return MessageBoxResultCancel, nil
		}
		return "", unixMessageBoxCommandError("osascript", err, output)
	}

	const prefix = "button returned:"
	label := text
	if index := strings.Index(label, prefix); index >= 0 {
		label = label[index+len(prefix):]
	}
	if index := strings.Index(label, ","); index >= 0 {
		label = label[:index]
	}
	if result, ok := unixMessageBoxResultForLabel(label); ok {
		return result, nil
	}
	return "", fmt.Errorf("system: %s: unexpected osascript result %q", CapabilityMessageBox, text)
}

func darwinAppleScriptQuote(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\""
}
