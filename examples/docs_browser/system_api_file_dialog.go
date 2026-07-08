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
		status := ui.UseState(sectionCtx, "文件对话框绑定到当前窗口所有者。")
		disabled := !system.Supports(system.CapabilityFileDialog)
		owner, ownerOK := ui.CurrentWindowNativeHandle(sectionCtx)
		ownerText := docsSystemNativeOwnerLabel(owner, ownerOK)
		defaultDir := docsSystemDefaultDialogDir()

		return docsSystemSection("File Dialog API", ui.ColumnElement(
			ui.TextElement(ownerText, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("打开文件", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.OpenFileDialogContext(ctx, context.Background(),
						system.FileDialogTitle("打开文件"),
						system.FileDialogDefaultDir(defaultDir),
						system.FileDialogRememberDir("docs-browser-open"),
						system.FileDialogFilters(
							system.FileFilter{Name: "Text", Patterns: []string{"txt", "md", "json"}},
							system.FileFilter{Name: "所有文件", Patterns: []string{"*.*"}},
						),
					)
					return formatDocsSystemFileDialogResult("Open file", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("打开多个文件", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.OpenFilesDialogContext(ctx, context.Background(),
						system.FileDialogTitle("打开多个文件"),
						system.FileDialogRememberDir("docs-browser-open"),
						system.FileDialogFilters(
							system.FileFilter{Name: "Text", Patterns: []string{"txt", "md", "json"}},
							system.FileFilter{Name: "所有文件", Patterns: []string{"*.*"}},
						),
					)
					return formatDocsSystemFileDialogResult("Open files", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("保存文件", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.SaveFileDialogContext(ctx, context.Background(),
						system.FileDialogTitle("保存文件"),
						system.FileDialogDefaultDir(defaultDir),
						system.FileDialogDefaultName("fluxui-docs"),
						system.FileDialogDefaultExtension("txt"),
						system.FileDialogAllowCreateDirs(true),
						system.FileDialogAllowMissingPath(true),
						system.FileDialogOverwritePrompt(true),
						system.FileDialogRememberDir("docs-browser-save"),
						system.FileDialogFilters(
							system.FileFilter{Name: "Text", Patterns: []string{"txt"}},
							system.FileFilter{Name: "所有文件", Patterns: []string{"*.*"}},
						),
					)
					return formatDocsSystemFileDialogResult("Save file", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("选择文件夹", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.PickFolderDialogContext(ctx, context.Background(),
						system.FileDialogTitle("选择文件夹"),
						system.FileDialogDefaultDir(defaultDir),
						system.FileDialogRememberDir("docs-browser-folder"),
					)
					return formatDocsSystemFileDialogResult("Pick folder", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("自动取消", status, disabled, func(ctx *ui.Context) string {
					callCtx, cancel := context.WithCancel(context.Background())
					timer := time.AfterFunc(1500*time.Millisecond, cancel)
					defer timer.Stop()
					defer cancel()
					result, err := ui.OpenFileDialogContext(ctx, callCtx,
						system.FileDialogTitle("1.5 秒后自动取消"),
						system.FileDialogRememberDir("docs-browser-open"),
						system.FileDialogFilters(system.FileFilter{Name: "所有文件", Patterns: []string{"*.*"}}),
					)
					return formatDocsSystemFileDialogResult("Auto cancel", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("系统所有者", status, disabled, func(ctx *ui.Context) string {
					owner, _ := ui.CurrentWindowNativeHandle(ctx)
					result, err := system.OpenFileDialog(context.Background(),
						system.FileDialogTitle("以显式所有者打开"),
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
		return "原生所有者：不可用；UI 包装器将回退到无所有者对话框。"
	}
	return fmt.Sprintf("原生所有者：0x%X；UI 包装器自动注入此项，system.* 调用可显式传递 FileDialogOwner。", owner)
}

func formatDocsSystemFileDialogResult(label string, result system.FileDialogResult, err error) string {
	if err != nil {
		return fmt.Sprintf("%s 失败：%v", label, err)
	}
	if result.Cancelled {
		return fmt.Sprintf("%s 已取消。", label)
	}
	if len(result.Paths) == 0 {
		return fmt.Sprintf("%s 完成，但未选择路径。", label)
	}
	return fmt.Sprintf("%s 已选择：%s", label, strings.Join(result.Paths, ", "))
}
