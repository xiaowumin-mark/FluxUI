//go:build darwin

package system

import (
	"fmt"
	"strings"
)

func unixFileDialogProviders() []unixFileDialogProvider {
	return []unixFileDialogProvider{
		{
			name:     "osascript",
			commands: []string{"osascript"},
			build:    darwinFileDialogCommand,
		},
	}
}

func darwinFileDialogCommand(mode FileDialogMode, opts fileDialogOptions) (unixFileDialogCommand, error) {
	script, err := darwinFileDialogScript(mode, opts)
	if err != nil {
		return unixFileDialogCommand{}, err
	}
	return unixFileDialogCommand{
		name: "osascript",
		args: []string{"-e", script},
		parse: func(output []byte, err error) (FileDialogResult, error) {
			return darwinParseFileDialogResult(output, err)
		},
	}, nil
}

func darwinFileDialogScript(mode FileDialogMode, opts fileDialogOptions) (string, error) {
	switch mode {
	case FileDialogOpenFile:
		return darwinChooseFileScript(opts, false), nil
	case FileDialogOpenFiles:
		return darwinChooseFileScript(opts, true), nil
	case FileDialogSaveFile:
		return darwinChooseFileNameScript(opts), nil
	case FileDialogPickFolder:
		return darwinChooseFolderScript(opts), nil
	default:
		return "", fmt.Errorf("system: unknown file dialog mode %q", mode)
	}
}

func darwinChooseFileScript(opts fileDialogOptions, multiple bool) string {
	parts := []string{"choose file"}
	if opts.title != "" {
		parts = append(parts, "with prompt "+darwinAppleScriptQuote(opts.title))
	}
	if opts.defaultDir != "" {
		parts = append(parts, "default location POSIX file "+darwinAppleScriptQuote(opts.defaultDir))
	}
	if extensions := unixFileDialogFilterExtensions(opts.filters); len(extensions) > 0 {
		quoted := make([]string, 0, len(extensions))
		for _, extension := range extensions {
			quoted = append(quoted, darwinAppleScriptQuote(extension))
		}
		parts = append(parts, "of type {"+strings.Join(quoted, ", ")+"}")
	}
	if multiple {
		parts = append(parts, "with multiple selections allowed")
	}

	choose := strings.Join(parts, " ")
	if !multiple {
		return "set chosenFile to " + choose + "\nreturn POSIX path of chosenFile"
	}
	return "set chosenFiles to " + choose + "\nset outputPaths to \"\"\nrepeat with chosenFile in chosenFiles\nset outputPaths to outputPaths & POSIX path of chosenFile & linefeed\nend repeat\nreturn outputPaths"
}

func darwinChooseFileNameScript(opts fileDialogOptions) string {
	parts := []string{"choose file name"}
	if opts.title != "" {
		parts = append(parts, "with prompt "+darwinAppleScriptQuote(opts.title))
	}
	if opts.defaultDir != "" {
		parts = append(parts, "default location POSIX file "+darwinAppleScriptQuote(opts.defaultDir))
	}
	if opts.defaultName != "" {
		parts = append(parts, "default name "+darwinAppleScriptQuote(opts.defaultName))
	}
	return "set chosenFile to " + strings.Join(parts, " ") + "\nreturn POSIX path of chosenFile"
}

func darwinChooseFolderScript(opts fileDialogOptions) string {
	parts := []string{"choose folder"}
	if opts.title != "" {
		parts = append(parts, "with prompt "+darwinAppleScriptQuote(opts.title))
	}
	if opts.defaultDir != "" {
		parts = append(parts, "default location POSIX file "+darwinAppleScriptQuote(opts.defaultDir))
	}
	return "set chosenFolder to " + strings.Join(parts, " ") + "\nreturn POSIX path of chosenFolder"
}

func darwinParseFileDialogResult(output []byte, err error) (FileDialogResult, error) {
	text := strings.TrimSpace(string(output))
	if err != nil {
		lower := strings.ToLower(text)
		if strings.Contains(lower, "user canceled") || strings.Contains(text, "-128") {
			return FileDialogResult{Cancelled: true}, nil
		}
		return FileDialogResult{}, unixFileDialogCommandError("osascript", err, output)
	}
	return unixFileDialogResultFromOutput(output), nil
}
