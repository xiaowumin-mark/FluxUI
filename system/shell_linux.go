//go:build linux

package system

import (
	"os"
	"path/filepath"
)

func unixShellOpenCommand() string {
	return "xdg-open"
}

func unixShellOpenURLCommand(target string) unixShellCommand {
	return unixShellCommand{name: unixShellOpenCommand(), args: []string{target}}
}

func unixShellOpenPathCommand(path string) unixShellCommand {
	return unixShellCommand{name: unixShellOpenCommand(), args: []string{path}}
}

func unixShellRevealPathCommand(path string) unixShellCommand {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return unixShellCommand{name: unixShellOpenCommand(), args: []string{path}}
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		dir = path
	}
	return unixShellCommand{name: unixShellOpenCommand(), args: []string{dir}}
}
