package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const docsSystemSingleInstanceID = "com.fluxui.docs_browser.demo"

func docsSystemSingleInstanceSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "Acquire primary, then simulate a second launch.")
		events := ui.UseState(sectionCtx, []string{"No secondary launches yet."})
		instance := ui.UseState[*system.SingleInstance](sectionCtx, nil)
		disabled := !system.Supports(system.CapabilitySingleInstance)

		current := instance.Value()
		ui.UseEffectWithDeps(sectionCtx, []any{current}, func() func() {
			handle := current
			if handle == nil {
				return nil
			}
			return func() {
				_ = handle.Close()
			}
		})

		acquired := current != nil
		summary := "Primary instance is not acquired."
		if acquired {
			summary = "Primary instance is active for " + docsSystemSingleInstanceID + "."
		}

		return docsSystemSection("Single Instance API", ui.ColumnElement(
			ui.TextElement(summary, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Acquire primary", status, disabled || acquired, func(ctx *ui.Context) string {
					acquiredInstance, err := system.AcquireSingleInstance(context.Background(),
						system.SingleInstanceID(docsSystemSingleInstanceID),
						system.SingleInstanceOnSecondLaunch(func(event system.SingleInstanceEvent) {
							events.Set(docsSystemPrependLog(events.Value(), formatDocsSystemSingleInstanceEvent(event), 8))
						}),
					)
					if err != nil {
						if system.IsAlreadyRunning(err) {
							return "Another primary instance is already running."
						}
						return "Acquire primary failed: " + err.Error()
					}
					instance.Set(acquiredInstance)
					return "Primary instance acquired."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Simulate second launch", status, disabled || !acquired, func(ctx *ui.Context) string {
					secondary, err := system.AcquireSingleInstance(context.Background(),
						system.SingleInstanceID(docsSystemSingleInstanceID),
						system.SingleInstanceArgs("--from=docs-browser", "--second-launch"),
						system.SingleInstancePayload(fmt.Sprintf("docs-browser://secondary/%d", time.Now().UnixNano())),
					)
					if secondary != nil {
						_ = secondary.Close()
					}
					if system.IsAlreadyRunning(err) {
						return "Secondary launch forwarded to the primary instance."
					}
					if err != nil {
						return "Secondary launch failed: " + err.Error()
					}
					return "Unexpectedly acquired primary from secondary simulation."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Release primary", status, disabled || !acquired, func(ctx *ui.Context) string {
					current := instance.Value()
					if current == nil {
						return "No primary instance to release."
					}
					err := current.Close()
					instance.Set(nil)
					if err != nil && !system.IsClosed(err) {
						return "Release primary failed: " + err.Error()
					}
					return "Primary instance released."
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("Second launch events", events.Value(), th, 112),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func formatDocsSystemSingleInstanceEvent(event system.SingleInstanceEvent) string {
	args := strings.Join(event.Args, " ")
	if args == "" {
		args = "(none)"
	}
	payload := event.Payload
	if payload == "" {
		payload = "(empty)"
	}
	return fmt.Sprintf("%s id=%s args=%s payload=%s", time.Now().Format("15:04:05"), event.ID, args, payload)
}
