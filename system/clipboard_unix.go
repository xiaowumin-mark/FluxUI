//go:build darwin || linux

package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type unixClipboardCommand struct {
	name string
	args []string
}

type unixClipboardProvider struct {
	name  string
	read  unixClipboardCommand
	write unixClipboardCommand
}

type unixLookPathFunc func(string) (string, error)

func (unixDriver) readClipboardText(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	provider, err := selectUnixClipboardProvider()
	if err != nil {
		return "", fmt.Errorf("system: %s: read text: %w: %v", CapabilityClipboard, ErrUnavailable, err)
	}

	cmd := exec.CommandContext(ctx, provider.read.name, provider.read.args...)
	output, err := cmd.Output()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("system: %s: read text with %s: %w: %v", CapabilityClipboard, provider.name, ErrUnavailable, err)
	}
	return string(output), nil
}

func (unixDriver) writeClipboardText(ctx context.Context, text string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	provider, err := selectUnixClipboardProvider()
	if err != nil {
		return fmt.Errorf("system: %s: write text: %w: %v", CapabilityClipboard, ErrUnavailable, err)
	}

	cmd := exec.CommandContext(ctx, provider.write.name, provider.write.args...)
	cmd.Stdin = strings.NewReader(text)
	if output, err := cmd.CombinedOutput(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("system: %s: write text with %s: %w: %s", CapabilityClipboard, provider.name, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	return nil
}

func selectUnixClipboardProvider() (unixClipboardProvider, error) {
	return unixClipboardProviderWithLookPath(exec.LookPath)
}

func unixClipboardProviderWithLookPath(lookPath unixLookPathFunc) (unixClipboardProvider, error) {
	candidates := unixClipboardCandidates()
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
		return unixClipboardProvider{}, fmt.Errorf("no clipboard providers configured")
	}
	return unixClipboardProvider{}, fmt.Errorf("clipboard commands unavailable: %s", strings.Join(uniqueStrings(missing), ", "))
}

func (p unixClipboardProvider) commandNames() []string {
	if p.read.name == p.write.name {
		return []string{p.read.name}
	}
	return []string{p.read.name, p.write.name}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
