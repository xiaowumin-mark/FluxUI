//go:build darwin || linux

package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type unixFileDialogCommand struct {
	name  string
	args  []string
	parse func([]byte, error) (FileDialogResult, error)
}

type unixFileDialogProvider struct {
	name     string
	commands []string
	build    func(FileDialogMode, fileDialogOptions) (unixFileDialogCommand, error)
}

func (unixDriver) openFileDialog(ctx context.Context, mode FileDialogMode, opts fileDialogOptions) (FileDialogResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return FileDialogResult{}, err
	}
	if err := validateUnixFileDialogOptions(mode, opts); err != nil {
		return FileDialogResult{}, err
	}

	provider, err := selectUnixFileDialogProvider()
	if err != nil {
		return FileDialogResult{}, fmt.Errorf("system: %s: open: %w: %v", CapabilityFileDialog, ErrUnavailable, err)
	}
	command, err := provider.build(mode, opts)
	if err != nil {
		return FileDialogResult{}, err
	}
	result, err := runUnixFileDialog(ctx, provider, command)
	if err != nil || result.Cancelled {
		return result, err
	}
	if err := validateUnixFileDialogResult(mode, opts, result); err != nil {
		return FileDialogResult{}, err
	}
	return result, nil
}

func runUnixFileDialog(ctx context.Context, provider unixFileDialogProvider, command unixFileDialogCommand) (FileDialogResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return FileDialogResult{}, err
	}
	if command.name == "" {
		return FileDialogResult{}, fmt.Errorf("system: %s: %s command is empty: %w", CapabilityFileDialog, provider.name, ErrUnavailable)
	}

	cmd := exec.CommandContext(ctx, command.name, command.args...)
	output, err := cmd.CombinedOutput()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return FileDialogResult{}, ctxErr
	}
	if command.parse == nil {
		if err != nil {
			return FileDialogResult{}, unixFileDialogCommandError(provider.name, err, output)
		}
		return unixFileDialogResultFromOutput(output), nil
	}
	return command.parse(output, err)
}

func selectUnixFileDialogProvider() (unixFileDialogProvider, error) {
	return unixFileDialogProviderWithLookPath(exec.LookPath)
}

func unixFileDialogProviderWithLookPath(lookPath unixLookPathFunc) (unixFileDialogProvider, error) {
	candidates := unixFileDialogProviders()
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
		return unixFileDialogProvider{}, fmt.Errorf("no file dialog providers configured")
	}
	return unixFileDialogProvider{}, fmt.Errorf("file dialog commands unavailable: %s", strings.Join(uniqueStrings(missing), ", "))
}

func (p unixFileDialogProvider) commandNames() []string {
	return p.commands
}

func validateUnixFileDialogOptions(mode FileDialogMode, opts fileDialogOptions) error {
	switch mode {
	case FileDialogOpenFile, FileDialogOpenFiles, FileDialogSaveFile, FileDialogPickFolder:
	default:
		return fmt.Errorf("system: unknown file dialog mode %q", mode)
	}
	if opts.owner != 0 {
		return fmt.Errorf("system: %s: owner window is unsupported on this platform: %w", CapabilityFileDialog, ErrUnsupported)
	}
	if opts.defaultDir != "" {
		if err := validateUnixFileDialogDefaultDir(opts.defaultDir); err != nil {
			return err
		}
	}
	return nil
}

func validateUnixFileDialogDefaultDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("system: configure file dialog default dir: %w", &FileDialogError{
			Kind: FileDialogErrorDefaultDir,
			Path: path,
			Err:  err,
		})
	}
	if !info.IsDir() {
		return fmt.Errorf("system: configure file dialog default dir: %w", &FileDialogError{
			Kind: FileDialogErrorDefaultDir,
			Path: path,
			Err:  fmt.Errorf("not a directory"),
		})
	}
	return nil
}

func validateUnixFileDialogResult(mode FileDialogMode, opts fileDialogOptions, result FileDialogResult) error {
	for _, path := range result.Paths {
		if path == "" {
			continue
		}
		switch mode {
		case FileDialogOpenFile, FileDialogOpenFiles:
			if !opts.allowMissingPath {
				if err := validateUnixExistingFile(path); err != nil {
					return err
				}
			}
		case FileDialogPickFolder:
			if !opts.allowMissingPath {
				if err := validateUnixExistingDir(path); err != nil {
					return err
				}
			}
		case FileDialogSaveFile:
			if err := validateUnixSaveTarget(path, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateUnixExistingFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return &FileDialogError{Kind: FileDialogErrorSelectedPath, Path: path, Err: err}
	}
	if info.IsDir() {
		return &FileDialogError{Kind: FileDialogErrorSelectedPath, Path: path, Err: fmt.Errorf("expected file, got directory")}
	}
	return nil
}

func validateUnixExistingDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return &FileDialogError{Kind: FileDialogErrorSelectedPath, Path: path, Err: err}
	}
	if !info.IsDir() {
		return &FileDialogError{Kind: FileDialogErrorSelectedPath, Path: path, Err: fmt.Errorf("expected directory")}
	}
	return nil
}

func validateUnixSaveTarget(path string, opts fileDialogOptions) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return &FileDialogError{Kind: FileDialogErrorSelectedPath, Path: dir, Err: fmt.Errorf("parent is not a directory")}
		}
		return nil
	}
	if opts.allowCreateDirs || opts.allowMissingPath {
		return nil
	}
	return &FileDialogError{Kind: FileDialogErrorSelectedPath, Path: dir, Err: err}
}

func unixFileDialogResultFromOutput(output []byte) FileDialogResult {
	paths := unixFileDialogOutputPaths(output)
	if len(paths) == 0 {
		return FileDialogResult{Cancelled: true}
	}
	return FileDialogResult{Paths: paths}
}

func unixFileDialogOutputPaths(output []byte) []string {
	text := strings.TrimSpace(string(output))
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		paths = append(paths, path)
	}
	return paths
}

func unixFileDialogStartPath(opts fileDialogOptions) string {
	if opts.defaultDir != "" && opts.defaultName != "" {
		return filepath.Join(opts.defaultDir, opts.defaultName)
	}
	if opts.defaultDir != "" {
		return opts.defaultDir
	}
	if opts.defaultName != "" {
		return opts.defaultName
	}
	return ""
}

func unixFileDialogFilterExtensions(filters []FileFilter) []string {
	seen := map[string]bool{}
	var extensions []string
	for _, filter := range filters {
		for _, pattern := range unixNormalizeFilterPatterns(filter.Patterns) {
			pattern = strings.TrimSpace(pattern)
			pattern = strings.TrimPrefix(pattern, "*")
			pattern = strings.TrimPrefix(pattern, ".")
			if pattern == "" || strings.ContainsAny(pattern, "*?/\\") {
				continue
			}
			if seen[pattern] {
				continue
			}
			seen[pattern] = true
			extensions = append(extensions, pattern)
		}
	}
	return extensions
}

func unixNormalizeFilterPatterns(patterns []string) []string {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if !strings.ContainsAny(pattern, "*?") {
			if strings.HasPrefix(pattern, ".") {
				pattern = "*" + pattern
			} else {
				pattern = "*." + pattern
			}
		}
		normalized = append(normalized, pattern)
	}
	return normalized
}

func unixFileDialogCommandError(provider string, err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	if detail != "" {
		return fmt.Errorf("system: %s: run %s: %w: %v: %s", CapabilityFileDialog, provider, ErrUnavailable, err, detail)
	}
	return fmt.Errorf("system: %s: run %s: %w: %v", CapabilityFileDialog, provider, ErrUnavailable, err)
}
