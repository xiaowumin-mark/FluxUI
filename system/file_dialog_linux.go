//go:build linux

package system

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func unixFileDialogProviders() []unixFileDialogProvider {
	return []unixFileDialogProvider{
		{
			name:     "zenity",
			commands: []string{"zenity"},
			build:    linuxZenityFileDialogCommand,
		},
		{
			name:     "kdialog",
			commands: []string{"kdialog"},
			build:    linuxKdialogFileDialogCommand,
		},
	}
}

func linuxZenityFileDialogCommand(mode FileDialogMode, opts fileDialogOptions) (unixFileDialogCommand, error) {
	args := []string{"--file-selection"}
	if opts.title != "" {
		args = append(args, "--title", opts.title)
	}
	if start := unixFileDialogStartPath(opts); start != "" {
		args = append(args, "--filename", start)
	}

	switch mode {
	case FileDialogOpenFile:
	case FileDialogOpenFiles:
		args = append(args, "--multiple", "--separator", "\n")
	case FileDialogSaveFile:
		args = append(args, "--save")
		if opts.overwritePrompt {
			args = append(args, "--confirm-overwrite")
		}
	case FileDialogPickFolder:
		args = append(args, "--directory")
	default:
		return unixFileDialogCommand{}, fmt.Errorf("system: unknown file dialog mode %q", mode)
	}

	if mode != FileDialogPickFolder {
		args = append(args, linuxZenityFileFilters(opts.filters)...)
	}

	return unixFileDialogCommand{
		name: "zenity",
		args: args,
		parse: func(output []byte, err error) (FileDialogResult, error) {
			return linuxParseFileDialogResult("zenity", output, err)
		},
	}, nil
}

func linuxZenityFileFilters(filters []FileFilter) []string {
	args := make([]string, 0, len(filters))
	for _, filter := range filters {
		patterns := unixNormalizeFilterPatterns(filter.Patterns)
		if len(patterns) == 0 {
			continue
		}
		name := strings.TrimSpace(filter.Name)
		if name == "" {
			name = strings.Join(patterns, " ")
		}
		args = append(args, "--file-filter="+name+" | "+strings.Join(patterns, " "))
	}
	return args
}

func linuxKdialogFileDialogCommand(mode FileDialogMode, opts fileDialogOptions) (unixFileDialogCommand, error) {
	args := []string{}
	if opts.title != "" {
		args = append(args, "--title", opts.title)
	}

	start := unixFileDialogStartPath(opts)
	if start == "" {
		start = "."
	}
	filter := linuxKdialogFileFilter(opts.filters)

	switch mode {
	case FileDialogOpenFile:
		args = append(args, "--getopenfilename", start, filter)
	case FileDialogOpenFiles:
		args = append(args, "--multiple", "--separate-output", "--getopenfilename", start, filter)
	case FileDialogSaveFile:
		args = append(args, "--getsavefilename", start, filter)
	case FileDialogPickFolder:
		if opts.defaultDir != "" {
			start = opts.defaultDir
		} else {
			start = filepath.Dir(start)
			if start == "" {
				start = "."
			}
		}
		args = append(args, "--getexistingdirectory", start)
	default:
		return unixFileDialogCommand{}, fmt.Errorf("system: unknown file dialog mode %q", mode)
	}

	return unixFileDialogCommand{
		name: "kdialog",
		args: args,
		parse: func(output []byte, err error) (FileDialogResult, error) {
			return linuxParseFileDialogResult("kdialog", output, err)
		},
	}, nil
}

func linuxKdialogFileFilter(filters []FileFilter) string {
	parts := make([]string, 0, len(filters))
	for _, filter := range filters {
		patterns := unixNormalizeFilterPatterns(filter.Patterns)
		if len(patterns) == 0 {
			continue
		}
		name := strings.TrimSpace(filter.Name)
		if name == "" {
			name = strings.Join(patterns, " ")
		}
		parts = append(parts, strings.Join(patterns, " ")+"|"+name)
	}
	return strings.Join(parts, "\n")
}

func linuxParseFileDialogResult(provider string, output []byte, err error) (FileDialogResult, error) {
	if err == nil {
		return unixFileDialogResultFromOutput(output), nil
	}
	if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
		return FileDialogResult{Cancelled: true}, nil
	}
	return FileDialogResult{}, unixFileDialogCommandError(provider, err, output)
}
