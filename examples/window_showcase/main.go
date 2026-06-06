package main

import (
	"fmt"
	"strings"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func mainWindow(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	titleSeq := ui.UseState(ctx, 0)
	eventLog := ui.UseState(ctx, []string{})

	currentID := ui.CurrentWindowID(ctx)
	handle, hasHandle := ui.GetWindow(currentID)
	state := ui.WindowState{ID: currentID}
	if hasHandle {
		if next, ok := handle.State(); ok {
			state = next
		}
		if events := handle.PollEvents(); len(events) > 0 {
			eventLog.Set(appendEventLog(eventLog.Value(), events))
		}
	}
	maxLimited := hasWindowMaxSize(state)

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("Window API Showcase", ui.TextSize(22)),
			ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement(formatState(state), ui.TextSize(13), ui.TextColor(th.Primary))),
			ui.PaddingElement(ui.Insets{Top: 12}, ui.RowElement(
				ui.FilledButtonElement(ui.TextElement("改标题"), ui.OnClick(func(ctx *ui.Context) {
					next := titleSeq.Value() + 1
					titleSeq.Set(next)
					ui.WindowSetTitle(ctx, fmt.Sprintf("Window Showcase #%d", next))
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("改尺寸"), ui.OnClick(func(ctx *ui.Context) {
					ui.WindowSetSize(ctx, 760, 520)
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("限制尺寸"), ui.OnClick(func(ctx *ui.Context) {
					ui.WindowSetMinSize(ctx, 560, 360)
					ui.WindowSetMaxSize(ctx, 1200, 860)
				})),
			)),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.RowElement(
				ui.ButtonElement(ui.TextElement("最小化"), ui.OnClick(func(ctx *ui.Context) { ui.WindowMinimize(ctx) })),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement(maximizeLabel(maxLimited)), ui.Disabled(maxLimited), ui.OnClick(func(ctx *ui.Context) { ui.WindowMaximize(ctx) })),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("还原"), ui.OnClick(func(ctx *ui.Context) { ui.WindowRestore(ctx) })),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("全屏"), ui.OnClick(func(ctx *ui.Context) { ui.WindowFullscreen(ctx) })),
			)),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.RowElement(
				ui.ButtonElement(ui.TextElement("居中"), ui.OnClick(func(ctx *ui.Context) { ui.WindowCenter(ctx) })),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("前置请求"), ui.OnClick(func(ctx *ui.Context) { ui.WindowRaise(ctx) })),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement(topmostLabel(state.AlwaysOnTop)), ui.OnClick(func(ctx *ui.Context) {
					ui.WindowSetAlwaysOnTop(ctx, !state.AlwaysOnTop)
				})),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("隐藏 10 秒"), ui.OnClick(func(ctx *ui.Context) {
					if !hasHandle {
						return
					}
					ui.WindowHide(ctx)
					go func(handle ui.WindowHandle) {
						time.Sleep(10 * time.Second)
						handle.Show()
						handle.Invalidate()
					}(handle)
				})),
			)),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.RowElement(
				ui.ButtonElement(ui.TextElement(hiddenMemoryPolicyLabel(state.HiddenMemoryPolicy)), ui.OnClick(func(ctx *ui.Context) {
					ui.WindowSetHiddenMemoryPolicy(ctx, toggleHiddenMemoryPolicy(state.HiddenMemoryPolicy))
				})),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("显示当前"), ui.OnClick(func(ctx *ui.Context) { ui.WindowShow(ctx) })),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("关闭工具窗口"), ui.OnClick(func(ctx *ui.Context) {
					for _, other := range ui.ListWindows() {
						if other.ID() != currentID {
							other.Close()
						}
					}
				})),
			)),
			ui.PaddingElement(ui.Insets{Top: 16}, ui.TextElement("事件", ui.TextSize(16))),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement(formatEvents(eventLog.Value()), ui.TextSize(12), ui.TextColor(th.Primary))),
		),
	)
}

func toolWindow(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	currentID := ui.CurrentWindowID(ctx)
	handle, hasHandle := ui.GetWindow(currentID)
	state := ui.WindowState{ID: currentID}
	if hasHandle {
		if next, ok := handle.State(); ok {
			state = next
		}
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("Tool Window", ui.TextSize(20)),
			ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement(formatState(state), ui.TextSize(13), ui.TextColor(th.Primary))),
			ui.PaddingElement(ui.Insets{Top: 12}, ui.RowElement(
				ui.ButtonElement(ui.TextElement("改标题"), ui.OnClick(func(ctx *ui.Context) {
					ui.WindowSetTitle(ctx, "Tool Window Active")
				})),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("关闭"), ui.OnClick(func(ctx *ui.Context) { ui.WindowClose(ctx) })),
			)),
		),
	)
}

func formatState(state ui.WindowState) string {
	return fmt.Sprintf(
		"ID=%d title=%q size=%dx%d min=%dx%d max=%dx%d visible=%v topmost=%v renderSuspended=%v hiddenMemory=%s minimized=%v maximized=%v fullscreen=%v focused=%v alive=%v",
		state.ID,
		state.Title,
		state.Width,
		state.Height,
		state.MinWidth,
		state.MinHeight,
		state.MaxWidth,
		state.MaxHeight,
		state.Visible,
		state.AlwaysOnTop,
		state.RenderSuspended,
		hiddenMemoryPolicyName(state.HiddenMemoryPolicy),
		state.Minimized,
		state.Maximized,
		state.Fullscreen,
		state.Focused,
		state.Alive,
	)
}

func hasWindowMaxSize(state ui.WindowState) bool {
	return state.MaxWidth > 0 && state.MaxHeight > 0
}

func maximizeLabel(limited bool) string {
	if limited {
		return "最大化不可用"
	}
	return "最大化"
}

func topmostLabel(always bool) string {
	if always {
		return "取消持续置顶"
	}
	return "持续置顶"
}

func hiddenMemoryPolicyLabel(policy ui.WindowHiddenMemoryPolicy) string {
	if policy == ui.WindowHiddenMemoryKeepRenderingState {
		return "隐藏释放临时内存"
	}
	return "隐藏保留渲染状态"
}

func toggleHiddenMemoryPolicy(policy ui.WindowHiddenMemoryPolicy) ui.WindowHiddenMemoryPolicy {
	if policy == ui.WindowHiddenMemoryKeepRenderingState {
		return ui.WindowHiddenMemoryReleaseTransient
	}
	return ui.WindowHiddenMemoryKeepRenderingState
}

func hiddenMemoryPolicyName(policy ui.WindowHiddenMemoryPolicy) string {
	if policy == ui.WindowHiddenMemoryKeepRenderingState {
		return "keep"
	}
	return "release"
}

func appendEventLog(log []string, events []ui.WindowEvent) []string {
	next := append([]string{}, log...)
	for _, event := range events {
		next = append(next, fmt.Sprintf("%s: %dx%d %q", event.Kind, event.State.Width, event.State.Height, event.State.Title))
	}
	if len(next) > 8 {
		next = next[len(next)-8:]
	}
	return next
}

func formatEvents(events []string) string {
	if len(events) == 0 {
		return "暂无窗口事件。调整尺寸、切换焦点或修改窗口状态后会显示最近事件。"
	}
	return strings.Join(events, "\n")
}

func main() {
	_ = ui.RunElementMulti(
		ui.WindowElement(
			mainWindow,
			ui.Title("Window Showcase"),
			ui.Size(680, 520),
			ui.MinSize(520, 360),
		),
		ui.WindowElement(
			toolWindow,
			ui.Title("Tool Window"),
			ui.Size(420, 260),
			ui.MinSize(320, 220),
		),
	)
}
