//go:build windows

package system

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const windowsClipboardTextPrefix = "FLUXUI_CLIPBOARD_TEXT|"
const windowsClipboardFilePrefix = "FLUXUI_CLIPBOARD_FILE|"
const windowsClipboardImagePrefix = "FLUXUI_CLIPBOARD_IMAGE_PNG|"

func (windowsDriver) readClipboardText(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	script := windowsClipboardReadPowerShellScript()
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-STA",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	output, err := cmd.CombinedOutput()
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
	return string(data), nil
}

func (windowsDriver) writeClipboardText(ctx context.Context, text string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	script := windowsClipboardWritePowerShellScript(text)
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-STA",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
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

	script := windowsClipboardReadFilesPowerShellScript()
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	output, err := cmd.CombinedOutput()
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

	script, err := windowsClipboardWriteFilesPowerShellScript(paths)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
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

	script := windowsClipboardReadImagePowerShellScript()
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-STA",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	output, err := cmd.CombinedOutput()
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
	return data, nil
}

func (windowsDriver) writeClipboardImagePNG(ctx context.Context, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	script := windowsClipboardWriteImagePowerShellScript(data)
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-STA",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("system: %s: write image: %w: %s", CapabilityClipboard, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	return nil
}

func windowsClipboardReadPowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`$text=Get-Clipboard -Raw -Format Text -ErrorAction SilentlyContinue;`,
		`if($null -eq $text){$text=''};`,
		`Write-Output ('` + windowsClipboardTextPrefix + `' + [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes([string]$text)));`,
	}, "")
}

func windowsClipboardWritePowerShellScript(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	var b strings.Builder
	b.WriteString(`$ErrorActionPreference='Stop';`)
	b.WriteString(`$text=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('`)
	b.WriteString(encoded)
	b.WriteString(`'));`)
	b.WriteString(`if($text.Length -eq 0){Add-Type -AssemblyName System.Windows.Forms;[System.Windows.Forms.Clipboard]::Clear();return;}`)
	b.WriteString(`Set-Clipboard -Value $text;`)
	return b.String()
}

func windowsClipboardReadFilesPowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`Add-Type -AssemblyName System.Windows.Forms;`,
		`if(-not [System.Windows.Forms.Clipboard]::ContainsFileDropList()){return};`,
		`$files=[System.Windows.Forms.Clipboard]::GetFileDropList();`,
		`foreach($file in $files){`,
		`$path=[string]$file;`,
		`Write-Output ('` + windowsClipboardFilePrefix + `' + [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($path)));`,
		`}`,
	}, "")
}

func windowsClipboardWriteFilesPowerShellScript(paths []string) (string, error) {
	data, err := json.Marshal(paths)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	var b strings.Builder
	b.WriteString(`$ErrorActionPreference='Stop';`)
	b.WriteString(`Add-Type -AssemblyName System.Windows.Forms;`)
	b.WriteString(`$json=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('`)
	b.WriteString(encoded)
	b.WriteString(`'));`)
	b.WriteString(`$paths=@(ConvertFrom-Json -InputObject $json);`)
	b.WriteString(`$files=New-Object System.Collections.Specialized.StringCollection;`)
	b.WriteString(`foreach($path in $paths){[void]$files.Add([string]$path);}`)
	b.WriteString(`[System.Windows.Forms.Clipboard]::SetFileDropList($files);`)
	return b.String(), nil
}

func windowsClipboardReadImagePowerShellScript() string {
	return strings.Join([]string{
		`$ErrorActionPreference='Stop';`,
		`Add-Type -AssemblyName System.Windows.Forms;`,
		`Add-Type -AssemblyName System.Drawing;`,
		`if(-not [System.Windows.Forms.Clipboard]::ContainsImage()){return};`,
		`$image=[System.Windows.Forms.Clipboard]::GetImage();`,
		`if($null -eq $image){return};`,
		`$stream=[System.IO.MemoryStream]::new();`,
		`try{`,
		`$image.Save($stream,[System.Drawing.Imaging.ImageFormat]::Png);`,
		`Write-Output ('` + windowsClipboardImagePrefix + `' + [Convert]::ToBase64String($stream.ToArray()));`,
		`}finally{`,
		`$stream.Dispose();`,
		`$image.Dispose();`,
		`}`,
	}, "")
}

func windowsClipboardWriteImagePowerShellScript(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var b strings.Builder
	b.WriteString(`$ErrorActionPreference='Stop';`)
	b.WriteString(`Add-Type -AssemblyName System.Windows.Forms;`)
	b.WriteString(`Add-Type -AssemblyName System.Drawing;`)
	b.WriteString(`$bytes=[Convert]::FromBase64String('`)
	b.WriteString(encoded)
	b.WriteString(`');`)
	b.WriteString(`$stream=[System.IO.MemoryStream]::new($bytes);`)
	b.WriteString(`$image=[System.Drawing.Image]::FromStream($stream);`)
	b.WriteString(`try{[System.Windows.Forms.Clipboard]::SetImage($image);}finally{$image.Dispose();$stream.Dispose();}`)
	return b.String()
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
