package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemClipboardShellSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "Clipboard and shell helpers use the active OS.")
		clipboardDisabled := !system.Supports(system.CapabilityClipboard)
		shellDisabled := !system.Supports(system.CapabilityShell)
		samplePath := docsDragSampleFile()

		return docsSystemSection("Clipboard / Shell API", ui.ColumnElement(
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Write clipboard", status, clipboardDisabled, func(ctx *ui.Context) string {
					if err := system.WriteClipboardText(context.Background(), "Copied from FluxUI docs browser"); err != nil {
						return "Write clipboard failed: " + err.Error()
					}
					return "Wrote clipboard text."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Read clipboard", status, clipboardDisabled, func(ctx *ui.Context) string {
					text, err := system.ReadClipboardText(context.Background())
					if err != nil {
						return "Read clipboard failed: " + err.Error()
					}
					if text == "" {
						return "Clipboard text is empty."
					}
					return fmt.Sprintf("Clipboard text: %q", text)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Open URL", status, shellDisabled, func(ctx *ui.Context) string {
					if err := system.OpenURL(context.Background(), "https://github.com/xiaowumin-mark/FluxUI"); err != nil {
						return "Open URL failed: " + err.Error()
					}
					return "Requested browser open for FluxUI."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Write files", status, clipboardDisabled, func(ctx *ui.Context) string {
					if err := system.WriteClipboardFiles(context.Background(), []string{samplePath}); err != nil {
						return "Write files failed: " + err.Error()
					}
					return "Wrote clipboard file list."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Read files", status, clipboardDisabled, func(ctx *ui.Context) string {
					files, err := system.ReadClipboardFiles(context.Background())
					if err != nil {
						return "Read files failed: " + err.Error()
					}
					if len(files) == 0 {
						return "Clipboard file list is empty."
					}
					return "Clipboard files: " + fmt.Sprint(files)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Write image", status, clipboardDisabled, func(ctx *ui.Context) string {
					data, err := os.ReadFile("examples/assets/sample.png")
					if err != nil {
						return "Read sample image failed: " + err.Error()
					}
					if err := system.WriteClipboardImagePNG(context.Background(), data); err != nil {
						return "Write image failed: " + err.Error()
					}
					return "Wrote PNG image to clipboard."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Open path", status, shellDisabled, func(ctx *ui.Context) string {
					if err := system.OpenPath(context.Background(), samplePath); err != nil {
						return "Open path failed: " + err.Error()
					}
					return "Requested default app open for: " + samplePath
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Reveal path", status, shellDisabled, func(ctx *ui.Context) string {
					if err := system.RevealPath(context.Background(), samplePath); err != nil {
						return "Reveal path failed: " + err.Error()
					}
					return "Requested reveal for: " + samplePath
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Read image", status, clipboardDisabled, func(ctx *ui.Context) string {
					data, err := system.ReadClipboardImagePNG(context.Background())
					if err != nil {
						return "Read image failed: " + err.Error()
					}
					return fmt.Sprintf("Clipboard PNG bytes: %d", len(data))
				})),
			),
			ui.VSpacerElement(8),
			ui.TextElement("Sample path: "+samplePath, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(4),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}
