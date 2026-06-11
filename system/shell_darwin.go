//go:build darwin

package system

func unixShellOpenCommand() string {
	return "open"
}

func unixShellOpenURLCommand(target string) unixShellCommand {
	return unixShellCommand{name: unixShellOpenCommand(), args: []string{target}}
}

func unixShellOpenPathCommand(path string) unixShellCommand {
	return unixShellCommand{name: unixShellOpenCommand(), args: []string{path}}
}

func unixShellRevealPathCommand(path string) unixShellCommand {
	return unixShellCommand{name: unixShellOpenCommand(), args: []string{"-R", path}}
}
