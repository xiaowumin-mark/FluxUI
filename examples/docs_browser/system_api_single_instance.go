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
		status := ui.UseState(sectionCtx, "获取主实例，然后模拟第二次启动。")
		events := ui.UseState(sectionCtx, []string{"暂无二次启动。"})
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
		summary := "主实例尚未获取。"
		if acquired {
			summary = "主实例已激活：" + docsSystemSingleInstanceID + "。"
		}

		return docsSystemSection("Single Instance API", ui.ColumnElement(
			ui.TextElement(summary, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("获取主实例", status, disabled || acquired, func(ctx *ui.Context) string {
					acquiredInstance, err := system.AcquireSingleInstance(context.Background(),
						system.SingleInstanceID(docsSystemSingleInstanceID),
						system.SingleInstanceOnSecondLaunch(func(event system.SingleInstanceEvent) {
							events.Set(docsSystemPrependLog(events.Value(), formatDocsSystemSingleInstanceEvent(event), 8))
						}),
					)
					if err != nil {
						if system.IsAlreadyRunning(err) {
							return "另一个主实例已在运行。"
						}
						return "获取主实例失败：" + err.Error()
					}
					instance.Set(acquiredInstance)
					return "主实例已获取。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("模拟第二次启动", status, disabled || !acquired, func(ctx *ui.Context) string {
					secondary, err := system.AcquireSingleInstance(context.Background(),
						system.SingleInstanceID(docsSystemSingleInstanceID),
						system.SingleInstanceArgs("--from=docs-browser", "--second-launch"),
						system.SingleInstancePayload(fmt.Sprintf("docs-browser://secondary/%d", time.Now().UnixNano())),
					)
					if secondary != nil {
						_ = secondary.Close()
					}
					if system.IsAlreadyRunning(err) {
						return "二次启动已转发到主实例。"
					}
					if err != nil {
						return "二次启动失败：" + err.Error()
					}
					return "从二次模拟中意外获取了主实例。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("释放主实例", status, disabled || !acquired, func(ctx *ui.Context) string {
					current := instance.Value()
					if current == nil {
						return "没有可释放的主实例。"
					}
					err := current.Close()
					instance.Set(nil)
					if err != nil && !system.IsClosed(err) {
						return "释放主实例失败：" + err.Error()
					}
					return "主实例已释放。"
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("二次启动事件", events.Value(), th, 112),
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
