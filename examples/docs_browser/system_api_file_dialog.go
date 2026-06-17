package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemFileDialogSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "File dialogs bind to the current window owner.")
		disabled := !system.Supports(system.CapabilityFileDialog)
		owner, ownerOK := ui.CurrentWindowNativeHandle(sectionCtx)
		ownerText := docsSystemNativeOwnerLabel(owner, ownerOK)
		defaultDir := docsSystemDefaultDialogDir()

		return docsSystemSection("File Dialog API", ui.ColumnElement(
			ui.TextElement(ownerText, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Open file", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.OpenFileDialogContext(ctx, context.Background(),
						system.FileDialogTitle("Open file"),
						system.FileDialogDefaultDir(defaultDir),
						system.FileDialogRememberDir("docs-browser-open"),
						system.FileDialogFilters(
							system.FileFilter{Name: "Text", Patterns: []string{"txt", "md", "json"}},
							system.FileFilter{Name: "All files", Patterns: []string{"*.*"}},
						),
					)
					return formatDocsSystemFileDialogResult("Open file", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Open files", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.OpenFilesDialogContext(ctx, context.Background(),
						system.FileDialogTitle("Open multiple files"),
						system.FileDialogRememberDir("docs-browser-open"),
						system.FileDialogFilters(
							system.FileFilter{Name: "Text", Patterns: []string{"txt", "md", "json"}},
							system.FileFilter{Name: "All files", Patterns: []string{"*.*"}},
						),
					)
					return formatDocsSystemFileDialogResult("Open files", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Save file", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.SaveFileDialogContext(ctx, context.Background(),
						system.FileDialogTitle("Save file"),
						system.FileDialogDefaultDir(defaultDir),
						system.FileDialogDefaultName("fluxui-docs"),
						system.FileDialogDefaultExtension("txt"),
						system.FileDialogAllowCreateDirs(true),
						system.FileDialogAllowMissingPath(true),
						system.FileDialogOverwritePrompt(true),
						system.FileDialogRememberDir("docs-browser-save"),
						system.FileDialogFilters(
							system.FileFilter{Name: "Text", Patterns: []string{"txt"}},
							system.FileFilter{Name: "All files", Patterns: []string{"*.*"}},
						),
					)
					return formatDocsSystemFileDialogResult("Save file", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Pick folder", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.PickFolderDialogContext(ctx, context.Background(),
						system.FileDialogTitle("Pick folder"),
						system.FileDialogDefaultDir(defaultDir),
						system.FileDialogRememberDir("docs-browser-folder"),
					)
					return formatDocsSystemFileDialogResult("Pick folder", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Auto cancel", status, disabled, func(ctx *ui.Context) string {
					callCtx, cancel := context.WithCancel(context.Background())
					timer := time.AfterFunc(1500*time.Millisecond, cancel)
					defer timer.Stop()
					defer cancel()
					result, err := ui.OpenFileDialogContext(ctx, callCtx,
						system.FileDialogTitle("Auto cancel after 1.5s"),
						system.FileDialogRememberDir("docs-browser-open"),
						system.FileDialogFilters(system.FileFilter{Name: "All files", Patterns: []string{"*.*"}}),
					)
					return formatDocsSystemFileDialogResult("Auto cancel", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("System owner", status, disabled, func(ctx *ui.Context) string {
					owner, _ := ui.CurrentWindowNativeHandle(ctx)
					result, err := system.OpenFileDialog(context.Background(),
						system.FileDialogTitle("Open with explicit owner"),
						system.FileDialogOwner(owner),
						system.FileDialogDefaultDir(defaultDir),
						system.FileDialogAllowMissingPath(false),
						system.FileDialogFilters(system.FileFilter{Name: "Markdown", Patterns: []string{".md"}}),
					)
					return formatDocsSystemFileDialogResult("System owner", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemDefaultDialogDir() string {
	return cachedDocsSystemDefaultDialogDir()
}

func docsSystemNativeOwnerLabel(owner uintptr, ok bool) string {
	if !ok || owner == 0 {
		return "Native owner: unavailable; ui wrappers will fall back to ownerless dialogs."
	}
	return fmt.Sprintf("Native owner: 0x%X; ui wrappers inject this automatically, system.* calls can pass FileDialogOwner explicitly.", owner)
}

func formatDocsSystemFileDialogResult(label string, result system.FileDialogResult, err error) string {
	if err != nil {
		return fmt.Sprintf("%s failed: %v", label, err)
	}
	if result.Cancelled {
		return fmt.Sprintf("%s cancelled.", label)
	}
	if len(result.Paths) == 0 {
		return fmt.Sprintf("%s completed with no paths.", label)
	}
	return fmt.Sprintf("%s selected: %s", label, strings.Join(result.Paths, ", "))
}
