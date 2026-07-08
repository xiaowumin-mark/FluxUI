package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/xiaowumin-mark/FluxUI/icons/md3"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemWindowSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "窗口控件作用于当前文档浏览器窗口。")
		titleSeq := ui.UseState(sectionCtx, 0)
		hiddenPolicy := ui.UseState(sectionCtx, ui.WindowHiddenMemoryReleaseTransient)
		closeGuard := ui.UseState(sectionCtx, false)
		eventLog := ui.UseState(sectionCtx, []string{"尚未观察到窗口事件。"})
		windowSub := ui.UseState[*ui.WindowEventSubscription](sectionCtx, nil)

		currentSub := windowSub.Value()
		ui.UseEffectWithDeps(sectionCtx, []any{currentSub}, func() func() {
			sub := currentSub
			if sub == nil {
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
						eventLog.Set(docsSystemPrependLog(eventLog.Value(), formatDocsWindowEvent("sub", event), 8))
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

		currentID := ui.CurrentWindowID(sectionCtx)
		handle, hasHandle := ui.GetWindow(currentID)
		windowState := ui.WindowState{ID: currentID}
		nativeHandle, nativeOK := uintptr(0), false
		if hasHandle {
			if next, ok := handle.State(); ok {
				windowState = next
			}
			nativeHandle, nativeOK = handle.NativeHandle()
		}

		disabled := !hasHandle
		summary := docsSystemWindowSummary(windowState, nativeHandle, nativeOK)
		hiddenLabel := "隐藏内存：释放瞬态"
		if hiddenPolicy.Value() == ui.WindowHiddenMemoryKeepRenderingState {
			hiddenLabel = "隐藏内存：保持渲染状态"
		}
		subscriptionActive := currentSub != nil

		button := func(label string, onClick func(*ui.Context)) ui.Element {
			return ui.OutlinedButtonElement(
				ui.TextElement(label, ui.TextSize(12)),
				ui.Disabled(disabled),
				ui.ButtonPadding(ui.Symmetric(6, 10)),
				ui.OnClick(onClick),
			)
		}

		return docsSystemSection("Window API", ui.ColumnElement(
			ui.TextElement(summary, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("标题 +1", func(ctx *ui.Context) {
					next := titleSeq.Value() + 1
					titleSeq.Set(next)
					if !ui.WindowSetTitle(ctx, fmt.Sprintf("FluxUI Docs #%d", next)) {
						status.Set("窗口标题更新失败。")
						return
					}
					status.Set(fmt.Sprintf("标题已更新为 #%d。", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("移动到 48,48", func(ctx *ui.Context) {
					if !ui.WindowSetPosition(ctx, 48, 48) {
						status.Set("窗口位置更新失败。")
						return
					}
					status.Set("窗口已移动到 48,48。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("调整为 760x520", func(ctx *ui.Context) {
					if !ui.WindowSetSize(ctx, 760, 520) {
						status.Set("窗口大小更新失败。")
						return
					}
					status.Set("窗口已调整为 760x520。")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("居中", func(ctx *ui.Context) {
					if !ui.WindowCenter(ctx) {
						status.Set("窗口居中失败。")
						return
					}
					status.Set("窗口已居中。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("聚焦", func(ctx *ui.Context) {
					if !ui.WindowRequestFocus(ctx) {
						status.Set("窗口聚焦请求失败。")
						return
					}
					status.Set("已请求窗口聚焦。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("提升", func(ctx *ui.Context) {
					if !ui.WindowRaise(ctx) {
						status.Set("窗口提升失败。")
						return
					}
					status.Set("已请求窗口提升。")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button(topmostLabel(windowState.AlwaysOnTop), func(ctx *ui.Context) {
					next := !windowState.AlwaysOnTop
					if !ui.WindowSetAlwaysOnTop(ctx, next) {
						status.Set("窗口置顶更新失败。")
						return
					}
					status.Set(fmt.Sprintf("置顶已设置为 %v。", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("隐藏 3 秒", func(ctx *ui.Context) {
					if !ui.WindowHide(ctx) {
						status.Set("窗口隐藏失败。")
						return
					}
					status.Set("窗口已隐藏，持续 3 秒。")
					go func(h ui.WindowHandle) {
						time.Sleep(3 * time.Second)
						_ = h.Show()
						status.Set("窗口已恢复显示。")
					}(handle)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("全屏", func(ctx *ui.Context) {
					if !ui.WindowFullscreen(ctx) {
						status.Set("窗口全屏失败。")
						return
					}
					status.Set("已请求窗口全屏。")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button(resizableLabel(windowState.Resizable), func(ctx *ui.Context) {
					next := !windowState.Resizable
					if !ui.WindowSetResizable(ctx, next) {
						status.Set("窗口调整大小更新失败。")
						return
					}
					status.Set(fmt.Sprintf("调整大小已设置为 %v。", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(decoratedLabel(windowState.Decorated), func(ctx *ui.Context) {
					next := !windowState.Decorated
					if !ui.WindowSetDecorated(ctx, next) {
						status.Set("窗口边框装饰更新失败。")
						return
					}
					status.Set(fmt.Sprintf("边框装饰已设置为 %v。", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(docsSystemWindowMinMaxLabel(windowState), func(ctx *ui.Context) {
					if docsSystemWindowHasMaxSize(windowState) {
						if !ui.WindowSetMaxSize(ctx, 0, 0) {
							status.Set("清除窗口最大尺寸失败。")
							return
						}
						status.Set("窗口最大尺寸已清除；最大化与 Snap Flyout 恢复可用。")
						return
					}
					if !ui.WindowSetMinSize(ctx, 640, 420) {
						status.Set("窗口最小尺寸更新失败。")
						return
					}
					if !ui.WindowSetMaxSize(ctx, 1200, 900) {
						status.Set("窗口最大尺寸更新失败。")
						return
					}
					status.Set("窗口最小/最大已设置为 640x420 / 1200x900。")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("Modern 边框", func(ctx *ui.Context) {
					if !ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
						Mode:   ui.WindowsFrameHidden,
						Shadow: true,
						Corner: ui.WindowsCornerRound,
						Border: ui.WindowsFrameBorderHidden,
					}) {
						status.Set("Windows 边框更新失败。")
						return
					}
					status.Set("Windows 边框已隐藏；拖拽条已激活。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Modern 描边", func(ctx *ui.Context) {
					if !ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
						Mode:        ui.WindowsFrameHidden,
						Shadow:      true,
						Corner:      ui.WindowsCornerRound,
						Border:      ui.WindowsFrameBorderColor,
						BorderColor: ui.NRGBA(59, 130, 246, 255),
					}) {
						status.Set("Windows 边框描边更新失败。")
						return
					}
					status.Set("Modern 描边已启用，同时不恢复原生 Win32 边框。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("探测 Chrome", func(ctx *ui.Context) {
					status.Set(formatDocsWindowsChromeAvailability(ui.ProbeWindowsChrome()))
				})),
			),
			ui.VSpacerElement(8),
			ui.WindowDragAreaElement(
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.Symmetric(8, 10)).WithRad(8),
					ui.TextElement("隐藏 Windows 边框后可拖拽此区域。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
				),
			),
			ui.VSpacerElement(8),
			docsSystemSnapFlyoutRow(windowState, status, th),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("开始拖拽", func(ctx *ui.Context) {
					if !ui.WindowStartDragMove(ctx) {
						status.Set("StartDragMove 失败；请从指针按下或拖拽条使用。")
						return
					}
					status.Set("已请求 StartDragMove。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(ui.TextElement("背景材质、原生透明度和窗口背景色稍后添加。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button(closeGuardLabel(closeGuard.Value()), func(ctx *ui.Context) {
					if !hasHandle {
						status.Set("窗口句柄不可用。")
						return
					}
					if closeGuard.Value() {
						if !handle.SetCloseRequestedHandler(nil) {
							status.Set("清除关闭守卫失败。")
							return
						}
						closeGuard.Set(false)
						status.Set("关闭守卫已清除。")
						return
					}
					if !handle.SetCloseRequestedHandler(func(request ui.WindowCloseRequest) bool {
						eventLog.Set(docsSystemPrependLog(eventLog.Value(), "关闭请求已被文档守卫取消", 8))
						return false
					}) {
						status.Set("安装关闭守卫失败。")
						return
					}
					closeGuard.Set(true)
					status.Set("关闭守卫已安装；再次点击以清除，然后才能关闭。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("轮询事件", func(ctx *ui.Context) {
					if !hasHandle {
						status.Set("窗口句柄不可用。")
						return
					}
					events := handle.PollEvents()
					if len(events) == 0 {
						status.Set("没有排队的窗口事件。")
						return
					}
					lines := eventLog.Value()
					for _, event := range events {
						lines = docsSystemPrependLog(lines, formatDocsWindowEvent("poll", event), 8)
					}
					eventLog.Set(lines)
					status.Set(fmt.Sprintf("已轮询 %d 个窗口事件。", len(events)))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(windowSubscribeLabel(subscriptionActive), func(ctx *ui.Context) {
					if !hasHandle {
						status.Set("窗口句柄不可用。")
						return
					}
					if current := windowSub.Value(); current != nil {
						_ = current.Close()
						windowSub.Set(nil)
						status.Set("窗口事件订阅已停止。")
						return
					}
					sub, ok := handle.SubscribeEvents(
						ui.WindowEventSizeChanged,
						ui.WindowEventScaleChanged,
						ui.WindowEventFocusChanged,
						ui.WindowEventStateChanged,
						ui.WindowEventCloseRequested,
						ui.WindowEventClosed,
					)
					if !ok {
						status.Set("窗口事件订阅失败。")
						return
					}
					windowSub.Set(sub)
					status.Set("窗口事件订阅已开始。")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("还原", func(ctx *ui.Context) {
					if !ui.WindowRestore(ctx) {
						status.Set("窗口还原失败。")
						return
					}
					status.Set("已请求窗口还原。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("最小化", func(ctx *ui.Context) {
					if !ui.WindowMinimize(ctx) {
						status.Set("窗口最小化失败。")
						return
					}
					status.Set("已请求窗口最小化。")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(hiddenLabel, func(ctx *ui.Context) {
					next := toggleHiddenMemoryPolicy(hiddenPolicy.Value())
					if !ui.WindowSetHiddenMemoryPolicy(ctx, next) {
						status.Set("隐藏内存策略更新失败。")
						return
					}
					hiddenPolicy.Set(next)
					status.Set("隐藏内存策略已更新。")
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("窗口事件", eventLog.Value(), th, 92),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemSnapFlyoutRow(state ui.WindowState, status docsStringState, th *ui.Theme) ui.Element {
	return ui.RowElement(
		ui.ExpandedElement(ui.ColumnElement(
			ui.TextElement("Windows Snap Flyout 按钮", ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(4),
			ui.TextElement("在 Windows 11 上将鼠标悬停于此图标上可让 OS 显示 Snap Layouts；点击切换最大化/还原。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		)),
		ui.HSpacerElement(8),
		docsSystemWindowMaximizeButton(state, status, th),
	)
}

func docsSystemWindowMaximizeButton(state ui.WindowState, status docsStringState, th *ui.Theme) ui.Element {
	disabled := !docsSystemWindowMaximizeAvailable(state)
	icon := "crop_square"
	action := "maximize"
	if state.Maximized {
		icon = "filter_none"
		action = "restore"
	}
	foreground := th.Colors.OnSurface
	if disabled {
		foreground = th.Colors.OnSurfaceVariant
	}
	return ui.TooltipElement(
		fmt.Sprintf("Windows Snap Flyout (%s)", action),
		ui.WindowMaximizeButtonElement(
			ui.IconButtonElement(
				ui.IconElement(icon, ui.IconSize(18), ui.IconUseFont(md3.ID)),
				ui.IconButtonSize(40),
				ui.IconButtonForeground(foreground),
				ui.IconButtonDecoration(docsSystemWindowMaximizeButtonDecoration(th)),
				ui.IconButtonDisabled(disabled),
				ui.IconButtonOnClick(func(ctx *ui.Context) {
					if state.Maximized {
						if !ui.WindowRestore(ctx) {
							status.Set("窗口还原失败。")
							return
						}
						status.Set("已通过 Snap Flyout 按钮请求窗口还原。")
						return
					}
					if !ui.WindowMaximize(ctx) {
						status.Set("窗口最大化失败。")
						return
					}
					status.Set("已通过 Snap Flyout 按钮请求窗口最大化。")
				}),
			),
			ui.WindowMaximizeButtonDisabled(disabled),
		),
	)
}

func docsSystemWindowMaximizeButtonDecoration(th *ui.Theme) ui.Decoration {
	return ui.Bg(th.Colors.SurfaceContainerHigh).
		WithRad(8).
		WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}).
		WithHover(ui.Bg(th.Colors.PrimaryContainer).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.Primary})).
		WithPressed(ui.Bg(th.Colors.SecondaryContainer).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.Secondary})).
		WithDisabled(ui.Bg(th.Colors.SurfaceContainerLow).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}))
}

func docsSystemWindowMaximizeAvailable(state ui.WindowState) bool {
	return state.Resizable &&
		!state.Minimized &&
		!state.Fullscreen &&
		!docsSystemWindowHasMaxSize(state)
}

func docsSystemWindowHasMaxSize(state ui.WindowState) bool {
	return state.MaxWidth > 0 && state.MaxHeight > 0
}

func docsSystemWindowMinMaxLabel(state ui.WindowState) string {
	if docsSystemWindowHasMaxSize(state) {
		return "清除最大尺寸"
	}
	return "设置最小/最大"
}

func docsSystemWindowSummary(state ui.WindowState, nativeHandle uintptr, nativeOK bool) string {
	native := "native=unavailable"
	if nativeOK {
		native = fmt.Sprintf("native=0x%X", nativeHandle)
	}
	return fmt.Sprintf(
		"ID=%d title=%q size=%dx%d min=%dx%d max=%dx%d scale=%.2f dpi=%d visible=%v topmost=%v hiddenMemory=%s renderSuspended=%v decorated=%v resizable=%v minimized=%v maximized=%v fullscreen=%v focused=%v alive=%v %s",
		state.ID,
		state.Title,
		state.Width,
		state.Height,
		state.MinWidth,
		state.MinHeight,
		state.MaxWidth,
		state.MaxHeight,
		state.Scale,
		state.DPI,
		state.Visible,
		state.AlwaysOnTop,
		windowHiddenMemoryPolicyName(state.HiddenMemoryPolicy),
		state.RenderSuspended,
		state.Decorated,
		state.Resizable,
		state.Minimized,
		state.Maximized,
		state.Fullscreen,
		state.Focused,
		state.Alive,
		native,
	)
}

func formatDocsWindowsChromeAvailability(value ui.WindowsChromeAvailability) string {
	return fmt.Sprintf(
		"Windows chrome supported=%v frame=%v drag=%v",
		value.Supported,
		value.FrameStyle,
		value.DragMove,
	)
}

func formatDocsWindowEvent(source string, event ui.WindowEvent) string {
	parts := []string{
		source,
		string(event.Kind),
		fmt.Sprintf("size=%dx%d", event.State.Width, event.State.Height),
		fmt.Sprintf("focused=%v", event.State.Focused),
	}
	if event.State.Scale > 0 {
		parts = append(parts, fmt.Sprintf("scale=%.2f", event.State.Scale))
	}
	return strings.Join(parts, " ")
}

func windowHiddenMemoryPolicyName(policy ui.WindowHiddenMemoryPolicy) string {
	if policy == ui.WindowHiddenMemoryKeepRenderingState {
		return "keep"
	}
	return "release"
}

func toggleHiddenMemoryPolicy(policy ui.WindowHiddenMemoryPolicy) ui.WindowHiddenMemoryPolicy {
	if policy == ui.WindowHiddenMemoryKeepRenderingState {
		return ui.WindowHiddenMemoryReleaseTransient
	}
	return ui.WindowHiddenMemoryKeepRenderingState
}

func topmostLabel(always bool) string {
	if always {
		return "关闭置顶"
	}
	return "开启置顶"
}

func resizableLabel(resizable bool) string {
	if resizable {
		return "禁用调整大小"
	}
	return "启用调整大小"
}

func decoratedLabel(decorated bool) string {
	if decorated {
		return "隐藏边框"
	}
	return "显示边框"
}

func closeGuardLabel(enabled bool) string {
	if enabled {
		return "清除关闭守卫"
	}
	return "安装关闭守卫"
}

func windowSubscribeLabel(active bool) string {
	if active {
		return "停止订阅"
	}
	return "订阅事件"
}
