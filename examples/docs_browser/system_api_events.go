package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemEventsSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		watching := ui.UseState(sectionCtx, false)
		log := ui.UseState(sectionCtx, []string{"系统事件为尽力而为且取决于驱动。"})
		disabled := !system.Supports(system.CapabilitySystemEvents)

		ui.UseEffectWithDeps(sectionCtx, []any{watching.Value()}, func() func() {
			if !watching.Value() || disabled {
				return nil
			}

			sub, err := system.SubscribeSystemEvents(context.Background(),
				system.SystemEventDisplayChanged,
				system.SystemEventDPIChanged,
				system.SystemEventThemeChanged,
				system.SystemEventSettingsChanged,
				system.SystemEventPowerChanged,
				system.SystemEventSessionChanged,
			)
			if err != nil {
				log.Set([]string{"Subscribe failed: " + err.Error()})
				return nil
			}

			done := make(chan struct{})
			go func() {
				for {
					select {
					case ev, ok := <-sub.Events():
						if !ok {
							return
						}
						next := append([]string{formatDocsSystemEvent(ev)}, log.Value()...)
						if len(next) > 8 {
							next = next[:8]
						}
						log.Set(next)
					case <-done:
						return
					}
				}
			}()

			return func() {
				close(done)
				_ = sub.Close()
			}
		})

		return docsSystemSection("System Events API", ui.ColumnElement(
			ui.RowElement(
				ui.OutlinedButtonElement(
					ui.TextElement(eventListenLabel(watching.Value()), ui.TextSize(12)),
					ui.Disabled(disabled),
					ui.ButtonPadding(ui.Symmetric(6, 10)),
					ui.OnClick(func(ctx *ui.Context) {
						next := !watching.Value()
						watching.Set(next)
						if !next {
							log.Set([]string{"监听已停止。"})
						} else {
							log.Set([]string{"监听已开始。"})
						}
					}),
				),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(
					ui.TextElement("清除日志", ui.TextSize(12)),
					ui.Disabled(disabled),
					ui.ButtonPadding(ui.Symmetric(6, 10)),
					ui.OnClick(func(ctx *ui.Context) {
						log.Set([]string{"日志已清除。"})
					}),
				),
			),
			ui.VSpacerElement(8),
			ui.TextElement("类型：显示、DPI、主题、设置、电源、会话。", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(10)).WithRad(8),
				ui.FixedHeightElement(
					140,
					ui.ScrollViewElement(
						ui.ColumnElement(renderDocsSystemEventLog(log.Value(), th)...),
						ui.ScrollVertical(true),
					),
				),
			),
		), th)
	})
}

func formatDocsSystemEvent(ev system.SystemEvent) string {
	detail := strings.TrimSpace(ev.Detail)
	if detail != "" {
		return fmt.Sprintf("%s %s - %s", ev.Time.Format("15:04:05"), ev.Kind, detail)
	}
	return fmt.Sprintf("%s %s", ev.Time.Format("15:04:05"), ev.Kind)
}

func eventListenLabel(watching bool) string {
	if watching {
		return "停止监听"
	}
	return "开始监听"
}

func renderDocsSystemEventLog(lines []string, th *ui.Theme) []ui.Element {
	if len(lines) == 0 {
		lines = []string{"暂无事件。"}
	}
	els := make([]ui.Element, 0, len(lines))
	for _, line := range lines {
		els = append(els, ui.TextElement(line, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)))
	}
	return els
}
