package main

import (
	"fmt"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemWindowSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "Window controls act on the current docs browser window.")
		titleSeq := ui.UseState(sectionCtx, 0)
		hiddenPolicy := ui.UseState(sectionCtx, ui.WindowHiddenMemoryReleaseTransient)

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
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemWindowSummary(state ui.WindowState, nativeHandle uintptr, nativeOK bool) string {
	native := "native=unavailable"
	if nativeOK {
		native = fmt.Sprintf("native=0x%X", nativeHandle)
	}
	return fmt.Sprintf(
		"ID=%d title=%q size=%dx%d min=%dx%d max=%dx%d scale=%.2f dpi=%d visible=%v topmost=%v hiddenMemory=%s minimized=%v maximized=%v fullscreen=%v focused=%v alive=%v %s",
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
		state.Minimized,
		state.Maximized,
		state.Fullscreen,
		state.Focused,
		state.Alive,
		native,
	)
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
