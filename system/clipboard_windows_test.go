//go:build windows

package system

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestWindowsClipboardPowerShellScripts(t *testing.T) {
	readScript := windowsClipboardReadPowerShellScript()
	if !strings.Contains(readScript, "Get-Clipboard") ||
		!strings.Contains(readScript, "ToBase64String") ||
		!strings.Contains(readScript, windowsClipboardTextPrefix) {
		t.Fatalf("unexpected read script: %s", readScript)
	}

	text := "hello ' clipboard"
	writeScript := windowsClipboardWritePowerShellScript()
	if !strings.Contains(writeScript, "Set-Clipboard") {
		t.Fatalf("unexpected write script: %s", writeScript)
	}
	if !strings.Contains(writeScript, "Clipboard]::Clear()") {
		t.Fatalf("write script should handle empty text by clearing clipboard: %s", writeScript)
	}
	if !strings.Contains(writeScript, "OpenStandardInput") || !strings.Contains(writeScript, "ReadToEnd") {
		t.Fatal("write script should read clipboard text from standard input")
	}
	if strings.Contains(writeScript, text) || strings.Contains(writeScript, base64.StdEncoding.EncodeToString([]byte(text))) {
		t.Fatal("write script must not embed clipboard text in the command line")
	}

	emptyWriteScript := windowsClipboardWritePowerShellScript()
	if !strings.Contains(emptyWriteScript, "Clipboard]::Clear()") ||
		!strings.Contains(emptyWriteScript, "return;") {
		t.Fatalf("empty write script should clear clipboard: %s", emptyWriteScript)
	}

	probeScript := windowsProbeClipboardPowerShellScript()
	if !strings.Contains(probeScript, "Get-Command Get-Clipboard,Set-Clipboard") {
		t.Fatalf("unexpected probe script: %s", probeScript)
	}

	readFilesScript := windowsClipboardReadFilesPowerShellScript()
	if !strings.Contains(readFilesScript, "FileDropList") ||
		!strings.Contains(readFilesScript, windowsClipboardFilePrefix) {
		t.Fatalf("unexpected read files script: %s", readFilesScript)
	}

	paths := []string{`C:\tmp\a.txt`, `C:\tmp\b.txt`}
	writeFilesScript := windowsClipboardWriteFilesPowerShellScript()
	if !strings.Contains(writeFilesScript, "SetFileDropList") {
		t.Fatalf("unexpected write files script: %s", writeFilesScript)
	}
	if !strings.Contains(writeFilesScript, "OpenStandardInput") || strings.Contains(writeFilesScript, paths[0]) {
		t.Fatal("write files script should read paths from standard input")
	}

	readImageScript := windowsClipboardReadImagePowerShellScript()
	if !strings.Contains(readImageScript, "ContainsImage") ||
		!strings.Contains(readImageScript, "ImageFormat") ||
		!strings.Contains(readImageScript, windowsClipboardImagePrefix) {
		t.Fatalf("unexpected read image script: %s", readImageScript)
	}

	imageData := []byte{1, 2, 3, 4}
	writeImageScript := windowsClipboardWriteImagePowerShellScript()
	if !strings.Contains(writeImageScript, "SetImage") {
		t.Fatalf("unexpected write image script: %s", writeImageScript)
	}
	if strings.Contains(writeImageScript, string(imageData)) {
		t.Fatal("write image script should embed image data as base64, not raw bytes")
	}
	if !strings.Contains(writeImageScript, "OpenStandardInput") || strings.Contains(writeImageScript, base64.StdEncoding.EncodeToString(imageData)) {
		t.Fatal("write image script should read image data from standard input")
	}
}

func TestWindowsClipboardOutputBufferIsBounded(t *testing.T) {
	limitHit := make(chan struct{}, 1)
	buffer := &windowsClipboardOutputBuffer{
		limit: 4,
		onLimit: func() {
			limitHit <- struct{}{}
		},
	}
	written, err := buffer.Write([]byte("abcdefgh"))
	if err != nil || written != 8 {
		t.Fatalf("unexpected bounded write result: written=%d err=%v", written, err)
	}
	data, truncated := buffer.snapshot()
	if string(data) != "abcd" || !truncated {
		t.Fatalf("expected bounded output and truncation marker, got %q truncated=%v", data, truncated)
	}
	select {
	case <-limitHit:
	default:
		t.Fatal("expected output limit callback")
	}
}

func TestWindowsClipboardPowerShellStopsAfterOutputLimit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	_, err := runWindowsClipboardPowerShellWithLimit(
		ctx,
		false,
		`[Console]::Out.Write(('x'*4096));Start-Sleep -Seconds 30;`,
		nil,
		64,
	)
	if err == nil || !strings.Contains(err.Error(), "output exceeds") {
		t.Fatalf("expected output limit error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 5*time.Second {
		t.Fatalf("PowerShell was not stopped after exceeding output limit: %s", elapsed)
	}
}

func TestParseWindowsClipboardReadOutput(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("hello"))
	value, ok := parseWindowsClipboardReadOutput("noise\n" + windowsClipboardTextPrefix + encoded + "\n")
	if !ok || value != encoded {
		t.Fatalf("unexpected parsed output: ok=%v value=%q", ok, value)
	}
	if _, ok := parseWindowsClipboardReadOutput("noise only"); ok {
		t.Fatal("expected missing prefix to fail")
	}
}

func TestParseWindowsClipboardFilesOutput(t *testing.T) {
	first := base64.StdEncoding.EncodeToString([]byte(`C:\tmp\a.txt`))
	second := base64.StdEncoding.EncodeToString([]byte(`C:\tmp\b.txt`))
	paths, err := parseWindowsClipboardFilesOutput(
		"noise\n" +
			windowsClipboardFilePrefix + first + "\n" +
			windowsClipboardFilePrefix + second + "\n",
	)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(paths) != 2 || paths[0] != `C:\tmp\a.txt` || paths[1] != `C:\tmp\b.txt` {
		t.Fatalf("unexpected parsed file paths: %#v", paths)
	}

	paths, err = parseWindowsClipboardFilesOutput("noise only")
	if err != nil {
		t.Fatalf("unexpected empty parse error: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no file paths, got %#v", paths)
	}

	if _, err := parseWindowsClipboardFilesOutput(windowsClipboardFilePrefix + "invalid"); err == nil {
		t.Fatal("expected invalid base64 to fail")
	}
}

func TestParseWindowsClipboardImageOutput(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("png"))
	value, ok := parseWindowsClipboardImageOutput("noise\n" + windowsClipboardImagePrefix + encoded + "\n")
	if !ok || value != encoded {
		t.Fatalf("unexpected parsed image output: ok=%v value=%q", ok, value)
	}
	if _, ok := parseWindowsClipboardImageOutput("noise only"); ok {
		t.Fatal("expected missing image prefix to fail")
	}
}
