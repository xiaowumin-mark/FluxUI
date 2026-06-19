package main

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
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
	}
	ui.UseEffectWithDeps(ctx, []any{currentID, hasHandle}, func() func() {
		if !hasHandle {
			return nil
		}
		sub, ok := handle.SubscribeEvents()
		if !ok {
			return nil
		}
		done := make(chan struct{})
		go func() {
			for {
				select {
				case event, ok := <-sub.Events():
					if !ok {
						return
					}
					eventLog.Set(appendEventLog(eventLog.Value(), []ui.WindowEvent{event}))
					handle.Invalidate()
				case <-done:
					return
				}
			}
		}()
		return func() {
			close(done)
			sub.Close()
		}
	})
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
			ui.PaddingElement(ui.Insets{Top: 12}, windowsChromeControls(state)),
			ui.PaddingElement(ui.Insets{Top: 16}, ui.TextElement("事件订阅", ui.TextSize(16))),
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

func windowsChromeControls(state ui.WindowState) ui.Element {
	probe := ui.ProbeWindowsChrome()
	return ui.ContainerDecorationElement(
		ui.Bg(ui.NRGBA(245, 247, 250, 255)).WithPad(ui.All(12)).WithRad(8),
		ui.ColumnElement(
			ui.TextElement("Windows chrome", ui.TextSize(16)),
			ui.VSpacerElement(6),
			ui.TextElement(formatChromeProbe(probe), ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
			ui.VSpacerElement(8),
			ui.WindowDragAreaElement(
				ui.ContainerDecorationElement(
					ui.Bg(ui.NRGBA(226, 232, 240, 255)).WithPad(ui.Symmetric(8, 10)).WithRad(6),
					ui.TextElement("Drag this strip to move a hidden-frame window", ui.TextSize(12), ui.TextColor(ui.NRGBA(30, 41, 59, 255))),
				),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ButtonElement(ui.TextElement("Modern frame"), ui.OnClick(func(ctx *ui.Context) {
					ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
						Mode:   ui.WindowsFrameHidden,
						Shadow: true,
						Corner: ui.WindowsCornerRound,
						Border: ui.WindowsFrameBorderHidden,
					})
				})),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("Modern border"), ui.OnClick(func(ctx *ui.Context) {
					ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
						Mode:        ui.WindowsFrameHidden,
						Shadow:      true,
						Corner:      ui.WindowsCornerRound,
						Border:      ui.WindowsFrameBorderColor,
						BorderColor: ui.NRGBA(59, 130, 246, 255),
					})
				})),
				ui.HSpacerElement(8),
				ui.TextElement(formatChromeState(state), ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
			),
		),
	)
}

func formatChromeProbe(probe ui.WindowsChromeAvailability) string {
	return fmt.Sprintf(
		"probe supported=%v frame=%v drag=%v",
		probe.Supported,
		probe.FrameStyle,
		probe.DragMove,
	)
}

func formatChromeState(state ui.WindowState) string {
	return fmt.Sprintf("frame=%d corner=%d border=%d", state.WindowsFrameStyle.Mode, state.WindowsFrameStyle.Corner, state.WindowsFrameStyle.Border)
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
	var closeAllowed atomic.Bool
	var closePrompting atomic.Bool

	_ = ui.RunElementMulti(
		ui.WindowElement(
			mainWindow,
			ui.Title("Window Showcase"),
			ui.Size(680, 520),
			ui.MinSize(520, 360),
			ui.OnCloseRequested(func(request ui.WindowCloseRequest) bool {
				if closeAllowed.Load() {
					return true
				}
				if closePrompting.CompareAndSwap(false, true) {
					go confirmWindowClose(request.Window, &closeAllowed, &closePrompting)
				} else {
					request.Window.SetTitle("关闭确认已打开")
					request.Window.Invalidate()
				}
				return false
			}),
		),
		ui.WindowElement(
			toolWindow,
			ui.Title("Tool Window"),
			ui.Size(420, 260),
			ui.MinSize(320, 220),
		),
	)
}

func confirmWindowClose(handle ui.WindowHandle, closeAllowed *atomic.Bool, closePrompting *atomic.Bool) {
	handle.SetTitle("等待关闭确认")
	handle.Invalidate()

	opts := []system.MessageBoxOption{
		system.MessageBoxTitle("保存更改"),
		system.MessageBoxText("关闭前是否保存当前文档？"),
		system.MessageBoxStyle(system.MessageBoxQuestion),
		system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
		system.MessageBoxDefaultButton(system.MessageBoxResultCancel),
	}
	if owner, ok := handle.NativeHandle(); ok {
		opts = append(opts, system.MessageBoxOwner(owner))
	}

	result, err := system.ShowMessageBox(context.Background(), opts...)
	if err != nil {
		handle.SetTitle(fmt.Sprintf("关闭确认失败: %v", err))
		handle.Invalidate()
		closePrompting.Store(false)
		return
	}

	switch result {
	case system.MessageBoxResultYes:
		handle.SetTitle("已保存，正在关闭")
		closeAllowed.Store(true)
		handle.Close()
	case system.MessageBoxResultNo:
		handle.SetTitle("不保存，正在关闭")
		closeAllowed.Store(true)
		handle.Close()
	default:
		handle.SetTitle("关闭已取消")
		handle.Invalidate()
		closePrompting.Store(false)
	}
}
