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
		status := ui.UseState(sectionCtx, "剪贴板和 Shell 助手使用当前操作系统。")
		clipboardDisabled := !system.Supports(system.CapabilityClipboard)
		shellDisabled := !system.Supports(system.CapabilityShell)
		samplePath := docsDragSampleFile()

		return docsSystemSection("Clipboard / Shell API", ui.ColumnElement(
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("写入剪贴板", status, clipboardDisabled, func(ctx *ui.Context) string {
					if err := system.WriteClipboardText(context.Background(), "Copied from FluxUI docs browser"); err != nil {
						return "写入剪贴板失败：" + err.Error()
					}
					return "已写入剪贴板文本。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("读取剪贴板", status, clipboardDisabled, func(ctx *ui.Context) string {
					text, err := system.ReadClipboardText(context.Background())
					if err != nil {
						return "读取剪贴板失败：" + err.Error()
					}
					if text == "" {
						return "剪贴板文本为空。"
					}
					return fmt.Sprintf("Clipboard text: %q", text)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("打开 URL", status, shellDisabled, func(ctx *ui.Context) string {
					if err := system.OpenURL(context.Background(), "https://github.com/xiaowumin-mark/FluxUI"); err != nil {
						return "打开 URL 失败：" + err.Error()
					}
					return "已请求浏览器打开 FluxUI。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("写入文件", status, clipboardDisabled, func(ctx *ui.Context) string {
					if err := system.WriteClipboardFiles(context.Background(), []string{samplePath}); err != nil {
						return "写入文件失败：" + err.Error()
					}
					return "已写入剪贴板文件列表。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("读取文件", status, clipboardDisabled, func(ctx *ui.Context) string {
					files, err := system.ReadClipboardFiles(context.Background())
					if err != nil {
						return "读取文件失败：" + err.Error()
					}
					if len(files) == 0 {
						return "剪贴板文件列表为空。"
					}
					return "Clipboard files: " + fmt.Sprint(files)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("写入图像", status, clipboardDisabled, func(ctx *ui.Context) string {
					data, err := os.ReadFile("examples/assets/sample.png")
					if err != nil {
						return "读取示例图像失败：" + err.Error()
					}
					if err := system.WriteClipboardImagePNG(context.Background(), data); err != nil {
						return "写入图像失败：" + err.Error()
					}
					return "已将 PNG 图像写入剪贴板。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("打开路径", status, shellDisabled, func(ctx *ui.Context) string {
					if err := system.OpenPath(context.Background(), samplePath); err != nil {
						return "打开路径失败：" + err.Error()
					}
					return "已请求默认应用打开：" + samplePath
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("在资源管理器中显示", status, shellDisabled, func(ctx *ui.Context) string {
					if err := system.RevealPath(context.Background(), samplePath); err != nil {
						return "在资源管理器中显示失败：" + err.Error()
					}
					return "已请求在资源管理器中显示：" + samplePath
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("读取图像", status, clipboardDisabled, func(ctx *ui.Context) string {
					data, err := system.ReadClipboardImagePNG(context.Background())
					if err != nil {
						return "读取图像失败：" + err.Error()
					}
					return fmt.Sprintf("Clipboard PNG bytes: %d", len(data))
				})),
			),
			ui.VSpacerElement(8),
			ui.TextElement("示例路径："+samplePath, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(4),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}
