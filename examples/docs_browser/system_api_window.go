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
		status := ui.UseState(sectionCtx, "Window controls act on the current docs browser window.")
		titleSeq := ui.UseState(sectionCtx, 0)
		hiddenPolicy := ui.UseState(sectionCtx, ui.WindowHiddenMemoryReleaseTransient)
		closeGuard := ui.UseState(sectionCtx, false)
		eventLog := ui.UseState(sectionCtx, []string{"No window events observed yet."})
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
		hiddenLabel := "Hidden memory: release transient"
		if hiddenPolicy.Value() == ui.WindowHiddenMemoryKeepRenderingState {
			hiddenLabel = "Hidden memory: keep rendering state"
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
				ui.ExpandedElement(button("Title +1", func(ctx *ui.Context) {
					next := titleSeq.Value() + 1
					titleSeq.Set(next)
					if !ui.WindowSetTitle(ctx, fmt.Sprintf("FluxUI Docs #%d", next)) {
						status.Set("Window title update failed.")
						return
					}
					status.Set(fmt.Sprintf("Title updated to #%d.", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Move to 48,48", func(ctx *ui.Context) {
					if !ui.WindowSetPosition(ctx, 48, 48) {
						status.Set("Window position update failed.")
						return
					}
					status.Set("Window moved to 48,48.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Resize 760x520", func(ctx *ui.Context) {
					if !ui.WindowSetSize(ctx, 760, 520) {
						status.Set("Window size update failed.")
						return
					}
					status.Set("Window resized to 760x520.")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("Center", func(ctx *ui.Context) {
					if !ui.WindowCenter(ctx) {
						status.Set("Window center failed.")
						return
					}
					status.Set("Window centered.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Focus", func(ctx *ui.Context) {
					if !ui.WindowRequestFocus(ctx) {
						status.Set("Window focus request failed.")
						return
					}
					status.Set("Requested window focus.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Raise", func(ctx *ui.Context) {
					if !ui.WindowRaise(ctx) {
						status.Set("Window raise failed.")
						return
					}
					status.Set("Requested window raise.")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button(topmostLabel(windowState.AlwaysOnTop), func(ctx *ui.Context) {
					next := !windowState.AlwaysOnTop
					if !ui.WindowSetAlwaysOnTop(ctx, next) {
						status.Set("Window topmost update failed.")
						return
					}
					status.Set(fmt.Sprintf("Always-on-top set to %v.", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Hide 3s", func(ctx *ui.Context) {
					if !ui.WindowHide(ctx) {
						status.Set("Window hide failed.")
						return
					}
					status.Set("Window hidden for 3 seconds.")
					go func(h ui.WindowHandle) {
						time.Sleep(3 * time.Second)
						_ = h.Show()
						status.Set("Window restored after hide.")
					}(handle)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Fullscreen", func(ctx *ui.Context) {
					if !ui.WindowFullscreen(ctx) {
						status.Set("Window fullscreen failed.")
						return
					}
					status.Set("Window fullscreen requested.")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button(resizableLabel(windowState.Resizable), func(ctx *ui.Context) {
					next := !windowState.Resizable
					if !ui.WindowSetResizable(ctx, next) {
						status.Set("Window resizable update failed.")
						return
					}
					status.Set(fmt.Sprintf("Resizable set to %v.", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(decoratedLabel(windowState.Decorated), func(ctx *ui.Context) {
					next := !windowState.Decorated
					if !ui.WindowSetDecorated(ctx, next) {
						status.Set("Window decoration update failed.")
						return
					}
					status.Set(fmt.Sprintf("Decorated set to %v.", next))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(docsSystemWindowMinMaxLabel(windowState), func(ctx *ui.Context) {
					if docsSystemWindowHasMaxSize(windowState) {
						if !ui.WindowSetMaxSize(ctx, 0, 0) {
							status.Set("Window max size clear failed.")
							return
						}
						status.Set("Window max size cleared; maximize and Snap Flyout are available again.")
						return
					}
					if !ui.WindowSetMinSize(ctx, 640, 420) {
						status.Set("Window min size update failed.")
						return
					}
					if !ui.WindowSetMaxSize(ctx, 1200, 900) {
						status.Set("Window max size update failed.")
						return
					}
					status.Set("Window min/max set to 640x420 / 1200x900.")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("Modern frame", func(ctx *ui.Context) {
					if !ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
						Mode:   ui.WindowsFrameHidden,
						Shadow: true,
						Corner: ui.WindowsCornerRound,
						Border: ui.WindowsFrameBorderHidden,
					}) {
						status.Set("Windows frame update failed.")
						return
					}
					status.Set("Windows frame hidden; drag strip is active.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Modern border", func(ctx *ui.Context) {
					if !ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
						Mode:        ui.WindowsFrameHidden,
						Shadow:      true,
						Corner:      ui.WindowsCornerRound,
						Border:      ui.WindowsFrameBorderColor,
						BorderColor: ui.NRGBA(59, 130, 246, 255),
					}) {
						status.Set("Windows frame border update failed.")
						return
					}
					status.Set("Modern border enabled without restoring the native Win32 frame.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Probe chrome", func(ctx *ui.Context) {
					status.Set(formatDocsWindowsChromeAvailability(ui.ProbeWindowsChrome()))
				})),
			),
			ui.VSpacerElement(8),
			ui.WindowDragAreaElement(
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.Symmetric(8, 10)).WithRad(8),
					ui.TextElement("Drag this strip after hiding the Windows frame.", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
				),
			),
			ui.VSpacerElement(8),
			docsSystemSnapFlyoutRow(windowState, status, th),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("Start drag", func(ctx *ui.Context) {
					if !ui.WindowStartDragMove(ctx) {
						status.Set("StartDragMove failed; use it from a pointer press or drag strip.")
						return
					}
					status.Set("StartDragMove requested.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(ui.TextElement("Background material, native transparency, and window background color are deferred.", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button(closeGuardLabel(closeGuard.Value()), func(ctx *ui.Context) {
					if !hasHandle {
						status.Set("Window handle unavailable.")
						return
					}
					if closeGuard.Value() {
						if !handle.SetCloseRequestedHandler(nil) {
							status.Set("Clear close guard failed.")
							return
						}
						closeGuard.Set(false)
						status.Set("Close guard cleared.")
						return
					}
					if !handle.SetCloseRequestedHandler(func(request ui.WindowCloseRequest) bool {
						eventLog.Set(docsSystemPrependLog(eventLog.Value(), "close requested and cancelled by docs guard", 8))
						return false
					}) {
						status.Set("Install close guard failed.")
						return
					}
					closeGuard.Set(true)
					status.Set("Close guard installed; click again to clear before closing.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Poll events", func(ctx *ui.Context) {
					if !hasHandle {
						status.Set("Window handle unavailable.")
						return
					}
					events := handle.PollEvents()
					if len(events) == 0 {
						status.Set("No queued window events.")
						return
					}
					lines := eventLog.Value()
					for _, event := range events {
						lines = docsSystemPrependLog(lines, formatDocsWindowEvent("poll", event), 8)
					}
					eventLog.Set(lines)
					status.Set(fmt.Sprintf("Polled %d window events.", len(events)))
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(windowSubscribeLabel(subscriptionActive), func(ctx *ui.Context) {
					if !hasHandle {
						status.Set("Window handle unavailable.")
						return
					}
					if current := windowSub.Value(); current != nil {
						_ = current.Close()
						windowSub.Set(nil)
						status.Set("Window event subscription stopped.")
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
						status.Set("Window event subscription failed.")
						return
					}
					windowSub.Set(sub)
					status.Set("Window event subscription started.")
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(button("Restore", func(ctx *ui.Context) {
					if !ui.WindowRestore(ctx) {
						status.Set("Window restore failed.")
						return
					}
					status.Set("Window restore requested.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button("Minimize", func(ctx *ui.Context) {
					if !ui.WindowMinimize(ctx) {
						status.Set("Window minimize failed.")
						return
					}
					status.Set("Window minimize requested.")
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(button(hiddenLabel, func(ctx *ui.Context) {
					next := toggleHiddenMemoryPolicy(hiddenPolicy.Value())
					if !ui.WindowSetHiddenMemoryPolicy(ctx, next) {
						status.Set("Hidden memory policy update failed.")
						return
					}
					hiddenPolicy.Set(next)
					status.Set("Hidden memory policy updated.")
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("Window events", eventLog.Value(), th, 92),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemSnapFlyoutRow(state ui.WindowState, status docsStringState, th *ui.Theme) ui.Element {
	return ui.RowElement(
		ui.ExpandedElement(ui.ColumnElement(
			ui.TextElement("Windows Snap Flyout button", ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(4),
			ui.TextElement("Hover this icon on Windows 11 to let the OS show Snap Layouts; click toggles maximize/restore.", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
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
							status.Set("Window restore failed.")
							return
						}
						status.Set("Window restore requested from Snap Flyout button.")
						return
					}
					if !ui.WindowMaximize(ctx) {
						status.Set("Window maximize failed.")
						return
					}
					status.Set("Window maximize requested from Snap Flyout button.")
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
		return "Clear max size"
	}
	return "Set min/max"
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
		return "Turn off topmost"
	}
	return "Turn on topmost"
}

func resizableLabel(resizable bool) string {
	if resizable {
		return "Disable resize"
	}
	return "Enable resize"
}

func decoratedLabel(decorated bool) string {
	if decorated {
		return "Hide frame"
	}
	return "Show frame"
}

func closeGuardLabel(enabled bool) string {
	if enabled {
		return "Clear close guard"
	}
	return "Install close guard"
}

func windowSubscribeLabel(active bool) string {
	if active {
		return "Stop subscription"
	}
	return "Subscribe events"
}
