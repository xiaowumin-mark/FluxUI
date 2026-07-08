package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemMessageBoxSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "消息框是原生且绑定到所有者的。")
		disabled := !system.Supports(system.CapabilityMessageBox)
		owner, ownerOK := ui.CurrentWindowNativeHandle(sectionCtx)

		return docsSystemSection("MessageBox API", ui.ColumnElement(
			ui.TextElement(docsSystemMessageBoxOwnerLabel(owner, ownerOK), ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("信息", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("FluxUI 文档"),
						system.MessageBoxText("这是一条信息消息。"),
						system.MessageBoxStyle(system.MessageBoxInfo),
						system.MessageBoxButtonSet(system.MessageBoxOK),
					)
					return formatDocsSystemMessageBoxResult("Info", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("询问", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("保存更改"),
						system.MessageBoxText("要在关闭前保存吗？"),
						system.MessageBoxStyle(system.MessageBoxQuestion),
						system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
						system.MessageBoxDefaultButton(system.MessageBoxResultCancel),
					)
					return formatDocsSystemMessageBoxResult("Question", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("警告", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("重试操作"),
						system.MessageBoxText("操作失败。重试？"),
						system.MessageBoxStyle(system.MessageBoxWarning),
						system.MessageBoxButtonSet(system.MessageBoxRetryCancel),
						system.MessageBoxDefaultButton(system.MessageBoxResultRetry),
					)
					return formatDocsSystemMessageBoxResult("Warning", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("错误", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("Error"),
						system.MessageBoxText("这是一条错误消息。"),
						system.MessageBoxStyle(system.MessageBoxError),
						system.MessageBoxButtonSet(system.MessageBoxOKCancel),
					)
					return formatDocsSystemMessageBoxResult("Error", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("详细", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxDetailedContext(ctx, context.Background(),
						system.MessageBoxTitle("详细对话框"),
						system.MessageBoxText("此对话框使用了可展开详情、验证复选框和自定义按钮。"),
						system.MessageBoxDetails("仅当平台支持任务对话框样式时，详情才会显示。"),
						system.MessageBoxFooter("页脚文本也是详细变体的一部分。"),
						system.MessageBoxVerification("记住我的选择", false),
						system.MessageBoxCustomButtons(
							system.MessageBoxButton{ID: "save", Label: "保存并关闭", Result: system.MessageBoxResultCustom},
							system.MessageBoxButton{ID: "discard", Label: "放弃更改", Result: system.MessageBoxResultCustom},
							system.MessageBoxButton{ID: "cancel", Label: "取消", Result: system.MessageBoxResultCancel},
						),
						system.MessageBoxDefaultButtonID("cancel"),
						system.MessageBoxCommandLinks(true),
						system.MessageBoxExpandedDetailsByDefault(true),
						system.MessageBoxExpandDetailsInFooterArea(false),
					)
					return formatDocsSystemMessageBoxDetailedResult("Detailed", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("自动取消", status, disabled, func(ctx *ui.Context) string {
					callCtx, cancel := context.WithCancel(context.Background())
					timer := time.AfterFunc(1500*time.Millisecond, cancel)
					defer timer.Stop()
					defer cancel()
					result, err := ui.ShowMessageBoxContext(ctx, callCtx,
						system.MessageBoxTitle("自动取消"),
						system.MessageBoxText("当上下文被取消时，此消息框将关闭。"),
						system.MessageBoxStyle(system.MessageBoxQuestion),
						system.MessageBoxButtonSet(system.MessageBoxOKCancel),
					)
					return formatDocsSystemMessageBoxResult("Auto cancel", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("异步信息", status, disabled, func(ctx *ui.Context) string {
					response := <-ui.ShowMessageBoxAsyncContext(ctx, context.Background(),
						system.MessageBoxTitle("异步消息"),
						system.MessageBoxText("使用 ui.ShowMessageBoxAsyncContext 并通过响应通道报告。"),
						system.MessageBoxStyle(system.MessageBoxInfo),
						system.MessageBoxButtonSet(system.MessageBoxOKCancel),
					)
					return formatDocsSystemMessageBoxResult("Async info", response.Result, response.Err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("异步详细", status, disabled, func(ctx *ui.Context) string {
					response := <-ui.ShowMessageBoxDetailedAsyncContext(ctx, context.Background(),
						system.MessageBoxTitle("异步详细"),
						system.MessageBoxText("使用详细的异步包装器。"),
						system.MessageBoxDetails("当平台支持时，结果通道携带按钮 ID 和验证状态。"),
						system.MessageBoxVerification("保留此设置", true),
						system.MessageBoxCustomButtons(
							system.MessageBoxButton{ID: "apply", Label: "应用", Result: system.MessageBoxResultCustom},
							system.MessageBoxButton{ID: "cancel", Label: "取消", Result: system.MessageBoxResultCancel},
						),
						system.MessageBoxDefaultButtonID("apply"),
						system.MessageBoxCommandLinksNoIcon(true),
					)
					return formatDocsSystemMessageBoxDetailedResult("Async detailed", response.Result, response.Err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("系统所有者", status, disabled, func(ctx *ui.Context) string {
					owner, _ := ui.CurrentWindowNativeHandle(ctx)
					result, err := system.ShowMessageBox(context.Background(),
						system.MessageBoxOwner(owner),
						system.MessageBoxTitle("显式所有者"),
						system.MessageBoxText("此直接 system.ShowMessageBox 调用显式接收 MessageBoxOwner。"),
						system.MessageBoxButtonSet(system.MessageBoxOK),
					)
					return formatDocsSystemMessageBoxResult("System owner", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemMessageBoxOwnerLabel(owner uintptr, ok bool) string {
	if !ok || owner == 0 {
		return "原生所有者：不可用；UI 包装器将回退到无所有者消息框。"
	}
	return fmt.Sprintf("原生所有者：0x%X；UI 包装器自动注入此项，system.* 调用可显式传递 MessageBoxOwner。", owner)
}

func formatDocsSystemMessageBoxResult(label string, result system.MessageBoxResult, err error) string {
	if err != nil {
		return fmt.Sprintf("%s 失败：%v", label, err)
	}
	return fmt.Sprintf("%s 结果：%s", label, result)
}

func formatDocsSystemMessageBoxDetailedResult(label string, result system.MessageBoxDetailedResult, err error) string {
	if err != nil {
		return fmt.Sprintf("%s 失败：%v", label, err)
	}
	return fmt.Sprintf("%s 结果：%s 按钮=%q 验证=%v", label, result.Result, result.ButtonID, result.VerificationChecked)
}
