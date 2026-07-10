//go:build windows

package system

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	windowsClipboardTextPrefix  = "FLUXUI_CLIPBOARD_TEXT|"
	windowsClipboardFilePrefix  = "FLUXUI_CLIPBOARD_FILE|"
	windowsClipboardImagePrefix = "FLUXUI_CLIPBOARD_IMAGE_PNG|"

	windowsClipboardMaxTextBytes   = 16 << 20
	windowsClipboardMaxFilesBytes  = 16 << 20
	windowsClipboardMaxFileCount   = 4096
	windowsClipboardMaxImageBytes  = 32 << 20
	windowsClipboardMaxImagePixels = 16 * 1024 * 1024
	windowsClipboardMaxOutputBytes = 48 << 20
	windowsClipboardCommandTimeout = 30 * time.Second
)

func (windowsDriver) readClipboardText(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	output, err := runWindowsClipboardPowerShell(ctx, true, windowsClipboardReadPowerShellScript(), nil)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("system: %s: read text: %w: %s", CapabilityClipboard, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	encoded, ok := parseWindowsClipboardReadOutput(string(output))
	if !ok {
		return "", fmt.Errorf("system: %s: read text: %w: missing clipboard payload", CapabilityClipboard, ErrUnavailable)
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("system: %s: decode text: %w", CapabilityClipboard, err)
	}
	if len(data) > windowsClipboardMaxTextBytes {
		return "", fmt.Errorf("system: %s: read text exceeds %d bytes: %w", CapabilityClipboard, windowsClipboardMaxTextBytes, ErrUnavailable)
	}
	return string(data), nil
}

func (windowsDriver) writeClipboardText(ctx context.Context, text string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(text) > windowsClipboardMaxTextBytes {
		return fmt.Errorf("system: %s: write text exceeds %d bytes: %w", CapabilityClipboard, windowsClipboardMaxTextBytes, ErrInvalidTarget)
	}
	output, err := runWindowsClipboardPowerShell(ctx, true, windowsClipboardWritePowerShellScript(), strings.NewReader(text))
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("system: %s: write text: %w: %s", CapabilityClipboard, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	return nil
}

func (windowsDriver) readClipboardFiles(ctx context.Context) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	output, err := runWindowsClipboardPowerShell(ctx, false, windowsClipboardReadFilesPowerShellScript(), nil)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("system: %s: read files: %w: %s", CapabilityClipboard, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	paths, err := parseWindowsClipboardFilesOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("system: %s: read files: %w", CapabilityClipboard, err)
	}
	return paths, nil
}

func (windowsDriver) writeClipboardFiles(ctx context.Context, paths []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(paths) > windowsClipboardMaxFileCount {
		return fmt.Errorf("system: %s: write files exceeds %d entries: %w", CapabilityClipboard, windowsClipboardMaxFileCount, ErrInvalidTarget)
	}
	data, err := json.Marshal(paths)
	if err != nil {
		return err
	}
	if len(data) > windowsClipboardMaxFilesBytes {
		return fmt.Errorf("system: %s: write files exceeds %d bytes: %w", CapabilityClipboard, windowsClipboardMaxFilesBytes, ErrInvalidTarget)
	}
	output, err := runWindowsClipboardPowerShell(ctx, false, windowsClipboardWriteFilesPowerShellScript(), bytes.NewReader(data))
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("system: %s: write files: %w: %s", CapabilityClipboard, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	return nil
}

func (windowsDriver) readClipboardImagePNG(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	output, err := runWindowsClipboardPowerShell(ctx, true, windowsClipboardReadImagePowerShellScript(), nil)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("system: %s: read image: %w: %s", CapabilityClipboard, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	encoded, ok := parseWindowsClipboardImageOutput(string(output))
	if !ok {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("system: %s: decode image: %w", CapabilityClipboard, err)
	}
	if len(data) > windowsClipboardMaxImageBytes {
		return nil, fmt.Errorf("system: %s: read image exceeds %d bytes: %w", CapabilityClipboard, windowsClipboardMaxImageBytes, ErrUnavailable)
	}
	return data, nil
}

func (windowsDriver) writeClipboardImagePNG(ctx context.Context, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(data) > windowsClipboardMaxImageBytes {
		return fmt.Errorf("system: %s: write image exceeds %d bytes: %w", CapabilityClipboard, windowsClipboardMaxImageBytes, ErrInvalidTarget)
	}
	output, err := runWindowsClipboardPowerShell(ctx, true, windowsClipboardWriteImagePowerShellScript(), bytes.NewReader(data))
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("system: %s: write image: %w: %s", CapabilityClipboard, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	return nil
}

type windowsClipboardOutputBuffer struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	limit     int
	truncated bool
	onLimit   func()
	limitOnce sync.Once
}

func (b *windowsClipboardOutputBuffer) Write(data []byte) (int, error) {
	written := len(data)
	b.mu.Lock()
	remaining := b.limit - b.buffer.Len()
	if remaining > 0 {
		if remaining > len(data) {
			remaining = len(data)
		}
		_, _ = b.buffer.Write(data[:remaining])
	}
	triggerLimit := remaining < len(data) && !b.truncated
	if remaining < len(data) {
		b.truncated = true
	}
	b.mu.Unlock()
	if triggerLimit && b.onLimit != nil {
		b.limitOnce.Do(b.onLimit)
	}
	return written, nil
}

func (b *windowsClipboardOutputBuffer) snapshot() ([]byte, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.buffer.Bytes()...), b.truncated
}

func runWindowsClipboardPowerShell(ctx context.Context, sta bool, script string, stdin io.Reader) ([]byte, error) {
	return runWindowsClipboardPowerShellWithLimit(ctx, sta, script, stdin, windowsClipboardMaxOutputBytes)
}

func runWindowsClipboardPowerShellWithLimit(ctx context.Context, sta bool, script string, stdin io.Reader, outputLimit int) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithTimeout(ctx, windowsClipboardCommandTimeout)
	defer cancel()
	args := make([]string, 0, 8)
	if sta {
		args = append(args, "-STA")
	}
	args = append(args,
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	cmd := exec.CommandContext(runCtx, "powershell.exe", args...)
	cmd.Stdin = stdin
	output := &windowsClipboardOutputBuffer{limit: outputLimit, onLimit: cancel}
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	data, truncated := output.snapshot()
	if truncated {
		return data, fmt.Errorf("PowerShell output exceeds %d bytes: %w", outputLimit, ErrUnavailable)
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
		return data, fmt.Errorf("PowerShell clipboard operation timed out after %s: %w", windowsClipboardCommandTimeout, ErrUnavailable)
	}
	return data, err
}

func windowsClipboardReadPowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`$text=Get-Clipboard -Raw -Format Text -ErrorAction SilentlyContinue;`,
		`if($null -eq $text){$text=''};`,
		`$bytes=[Text.Encoding]::UTF8.GetBytes([string]$text);`,
		fmt.Sprintf(`if($bytes.Length -gt %d){throw 'clipboard text exceeds FluxUI limit'};`, windowsClipboardMaxTextBytes),
		`Write-Output ('` + windowsClipboardTextPrefix + `' + [Convert]::ToBase64String($bytes));`,
	}, "")
}

func windowsClipboardWritePowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`$reader=[IO.StreamReader]::new([Console]::OpenStandardInput(),[Text.UTF8Encoding]::new($false),$true);`,
		`try{$text=$reader.ReadToEnd();}finally{$reader.Dispose();}`,
		`if($text.Length -eq 0){Add-Type -AssemblyName System.Windows.Forms;[System.Windows.Forms.Clipboard]::Clear();return;}`,
		`Set-Clipboard -Value $text;`,
	}, "")
}

func windowsClipboardReadFilesPowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`Add-Type -AssemblyName System.Windows.Forms;`,
		`if(-not [System.Windows.Forms.Clipboard]::ContainsFileDropList()){return};`,
		`$files=[System.Windows.Forms.Clipboard]::GetFileDropList();`,
		fmt.Sprintf(`if($files.Count -gt %d){throw 'clipboard file count exceeds FluxUI limit'};`, windowsClipboardMaxFileCount),
		`$totalBytes=0L;`,
		`foreach($file in $files){`,
		`$path=[string]$file;`,
		`$bytes=[Text.Encoding]::UTF8.GetBytes($path);`,
		`$totalBytes+=$bytes.Length;`,
		fmt.Sprintf(`if($totalBytes -gt %d){throw 'clipboard file list exceeds FluxUI limit'};`, windowsClipboardMaxFilesBytes),
		`Write-Output ('` + windowsClipboardFilePrefix + `' + [Convert]::ToBase64String($bytes));`,
		`}`,
	}, "")
}

func windowsClipboardWriteFilesPowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`Add-Type -AssemblyName System.Windows.Forms;`,
		`$reader=[IO.StreamReader]::new([Console]::OpenStandardInput(),[Text.UTF8Encoding]::new($false),$true);`,
		`try{$json=$reader.ReadToEnd();}finally{$reader.Dispose();}`,
		`$paths=@(ConvertFrom-Json -InputObject $json);`,
		`$files=New-Object System.Collections.Specialized.StringCollection;`,
		`foreach($path in $paths){[void]$files.Add([string]$path);}`,
		`[System.Windows.Forms.Clipboard]::SetFileDropList($files);`,
	}, "")
}

func windowsClipboardReadImagePowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`Add-Type -AssemblyName System.Windows.Forms;`,
		`Add-Type -AssemblyName System.Drawing;`,
		`if(-not [System.Windows.Forms.Clipboard]::ContainsImage()){return};`,
		`$image=[System.Windows.Forms.Clipboard]::GetImage();`,
		`if($null -eq $image){return};`,
		fmt.Sprintf(`$pixels=[int64]$image.Width*[int64]$image.Height;if($pixels -gt %d){throw 'clipboard image dimensions exceed FluxUI limit'};`, windowsClipboardMaxImagePixels),
		`$stream=[System.IO.MemoryStream]::new();`,
		`try{`,
		`$image.Save($stream,[System.Drawing.Imaging.ImageFormat]::Png);`,
		fmt.Sprintf(`if($stream.Length -gt %d){throw 'clipboard image exceeds FluxUI limit'};`, windowsClipboardMaxImageBytes),
		`Write-Output ('` + windowsClipboardImagePrefix + `' + [Convert]::ToBase64String($stream.ToArray()));`,
		`}finally{`,
		`$stream.Dispose();`,
		`$image.Dispose();`,
		`}`,
	}, "")
}

func windowsClipboardWriteImagePowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`Add-Type -AssemblyName System.Windows.Forms;`,
		`Add-Type -AssemblyName System.Drawing;`,
		`$stream=[System.IO.MemoryStream]::new();`,
		`$input=[Console]::OpenStandardInput();`,
		`$input.CopyTo($stream);`,
		`$stream.Position=0;`,
		`$image=[System.Drawing.Image]::FromStream($stream);`,
		`try{[System.Windows.Forms.Clipboard]::SetImage($image);}finally{$image.Dispose();$stream.Dispose();}`,
	}, "")
}

func windowsProbeClipboardPowerShellScript() string {
	return `$ErrorActionPreference='Stop';Get-Command Get-Clipboard,Set-Clipboard | Out-Null;`
}

func parseWindowsClipboardReadOutput(output string) (string, bool) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, windowsClipboardTextPrefix) {
			return strings.TrimPrefix(line, windowsClipboardTextPrefix), true
		}
	}
	return "", false
}

func parseWindowsClipboardFilesOutput(output string) ([]string, error) {
	paths := []string{}
	totalBytes := 0
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, windowsClipboardFilePrefix) {
			continue
		}
		encoded := strings.TrimPrefix(line, windowsClipboardFilePrefix)
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, err
		}
		if len(paths) >= windowsClipboardMaxFileCount {
			return nil, fmt.Errorf("clipboard file list exceeds %d entries", windowsClipboardMaxFileCount)
		}
		totalBytes += len(data)
		if totalBytes > windowsClipboardMaxFilesBytes {
			return nil, fmt.Errorf("clipboard file list exceeds %d bytes", windowsClipboardMaxFilesBytes)
		}
		paths = append(paths, string(data))
	}
	return paths, nil
}

func parseWindowsClipboardImageOutput(output string) (string, bool) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, windowsClipboardImagePrefix) {
			return strings.TrimPrefix(line, windowsClipboardImagePrefix), true
		}
	}
	return "", false
}
