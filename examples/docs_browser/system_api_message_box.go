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
		status := ui.UseState(sectionCtx, "Message boxes are native and owner-bound.")
		disabled := !system.Supports(system.CapabilityMessageBox)

		return docsSystemSection("MessageBox API", ui.ColumnElement(
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Info", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("FluxUI Docs"),
						system.MessageBoxText("This is an info message box."),
						system.MessageBoxStyle(system.MessageBoxInfo),
						system.MessageBoxButtonSet(system.MessageBoxOK),
					)
					return formatDocsSystemMessageBoxResult("Info", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Question", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("Save changes"),
						system.MessageBoxText("Do you want to save before closing?"),
						system.MessageBoxStyle(system.MessageBoxQuestion),
						system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
						system.MessageBoxDefaultButton(system.MessageBoxResultCancel),
					)
					return formatDocsSystemMessageBoxResult("Question", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Warning", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("Retry operation"),
						system.MessageBoxText("The operation failed. Try again?"),
						system.MessageBoxStyle(system.MessageBoxWarning),
						system.MessageBoxButtonSet(system.MessageBoxRetryCancel),
						system.MessageBoxDefaultButton(system.MessageBoxResultRetry),
					)
					return formatDocsSystemMessageBoxResult("Warning", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Error", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
						system.MessageBoxTitle("Error"),
						system.MessageBoxText("This is an error message box."),
						system.MessageBoxStyle(system.MessageBoxError),
						system.MessageBoxButtonSet(system.MessageBoxOKCancel),
					)
					return formatDocsSystemMessageBoxResult("Error", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Detailed", status, disabled, func(ctx *ui.Context) string {
					result, err := ui.ShowMessageBoxDetailedContext(ctx, context.Background(),
						system.MessageBoxTitle("Detailed dialog"),
						system.MessageBoxText("This dialog uses expandable details, verification, and custom buttons."),
						system.MessageBoxDetails("Details are shown only when the platform supports task-dialog style dialogs."),
						system.MessageBoxFooter("Footer text is also part of the detailed variant."),
						system.MessageBoxVerification("Remember my selection", false),
						system.MessageBoxCustomButtons(
							system.MessageBoxButton{ID: "save", Label: "Save and close", Result: system.MessageBoxResultCustom},
							system.MessageBoxButton{ID: "discard", Label: "Discard changes", Result: system.MessageBoxResultCustom},
							system.MessageBoxButton{ID: "cancel", Label: "Cancel", Result: system.MessageBoxResultCancel},
						),
						system.MessageBoxDefaultButtonID("cancel"),
						system.MessageBoxCommandLinks(true),
					)
					return formatDocsSystemMessageBoxDetailedResult("Detailed", result, err)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Auto cancel", status, disabled, func(ctx *ui.Context) string {
					callCtx, cancel := context.WithCancel(context.Background())
					timer := time.AfterFunc(1500*time.Millisecond, cancel)
					defer timer.Stop()
					defer cancel()
					result, err := ui.ShowMessageBoxContext(ctx, callCtx,
						system.MessageBoxTitle("Auto cancel"),
						system.MessageBoxText("This message box will close when the context is canceled."),
						system.MessageBoxStyle(system.MessageBoxQuestion),
						system.MessageBoxButtonSet(system.MessageBoxOKCancel),
					)
					return formatDocsSystemMessageBoxResult("Auto cancel", result, err)
				})),
			),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func formatDocsSystemMessageBoxResult(label string, result system.MessageBoxResult, err error) string {
	if err != nil {
		return fmt.Sprintf("%s failed: %v", label, err)
	}
	return fmt.Sprintf("%s result: %s", label, result)
}

func formatDocsSystemMessageBoxDetailedResult(label string, result system.MessageBoxDetailedResult, err error) string {
	if err != nil {
		return fmt.Sprintf("%s failed: %v", label, err)
	}
	return fmt.Sprintf("%s result: %s button=%q verification=%v", label, result.Result, result.ButtonID, result.VerificationChecked)
}
