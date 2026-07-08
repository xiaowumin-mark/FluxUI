package main

import (
	"fmt"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/icons/md3"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func app(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	status := ui.UseState(ctx, "Windows frame controls are ready.")

	state := ui.WindowState{}
	currentID := ui.CurrentWindowID(ctx)
	if handle, ok := ui.GetWindow(currentID); ok {
		if next, ok := handle.State(); ok {
			state = next
		}
	}

	return ui.ContainerDecorationElement(
		windowShellDecoration(state),
		ui.ColumnElement(
			titleBar(status.Value(), state),
			ui.ScrollViewElement(
				ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 244, G: 247, B: 251, A: 255}).WithPad(ui.All(20)),
					ui.ColumnElement(
						ui.TextElement("Windows Chrome Showcase", ui.TextSize(22), ui.TextColor(th.Colors.OnSurface)),
						ui.VSpacerElement(6),
						ui.TextElement(chromeSummary(state, ui.ProbeWindowsChrome()), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
						ui.VSpacerElement(16),
						ui.RowElement(
							ui.ExpandedElement(framePanel(state, status)),
							ui.HSpacerElement(14),
							ui.ExpandedElement(dragPanel()),
						),
						ui.VSpacerElement(14),
						snapFlyoutPanel(state, status),
						ui.VSpacerElement(14),
						ui.RowElement(
							ui.ExpandedElement(runtimePanel(status)),
							ui.HSpacerElement(14),
							ui.ExpandedElement(notesPanel()),
						),
						ui.VSpacerElement(14),
						ui.ContainerDecorationElement(
							ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).WithPad(ui.All(12)).WithRad(8),
							ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
						),
					),
				),
				ui.ScrollVertical(true),
			),
		),
	)
}

func titleBar(status string, state ui.WindowState) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(color.NRGBA{R: 26, G: 33, B: 45, A: 255}).WithPad(ui.Symmetric(8, 10)),
		ui.RowElement(
			ui.ExpandedElement(ui.WindowDragAreaElement(
				ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 18}).WithPad(ui.Symmetric(7, 12)).WithRad(6),
					ui.RowElement(
						ui.TextElement("FluxUI", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.HSpacerElement(10),
						ui.TextElement(status, ui.TextSize(12), ui.TextColor(color.NRGBA{R: 220, G: 229, B: 238, A: 230})),
					),
				),
			)),
			ui.HSpacerElement(8),
			captionButton("minimize", false, func(ctx *ui.Context) {
				ui.WindowMinimize(ctx)
			}),
			ui.HSpacerElement(6),
			maximizeCaptionButton(state),
			ui.HSpacerElement(6),
			captionButton("close", false, func(ctx *ui.Context) {
				ui.WindowClose(ctx)
			}),
		),
	)
}

func maximizeCaptionButton(state ui.WindowState) ui.Element {
	disabled := !captionMaximizeAvailable(state) || state.Fullscreen || state.Minimized
	icon := "crop_square"
	onClick := func(ctx *ui.Context) {
		ui.WindowMaximize(ctx)
	}
	if state.Maximized {
		icon = "filter_none"
		onClick = func(ctx *ui.Context) {
			ui.WindowRestore(ctx)
		}
	}
	return ui.WindowMaximizeButtonElement(
		captionButton(icon, disabled, onClick),
		ui.WindowMaximizeButtonDisabled(disabled),
	)
}

func captionButton(icon string, disabled bool, onClick func(*ui.Context)) ui.Element {
	return ui.IconButtonElement(
		ui.IconElement(icon, ui.IconSize(17), ui.IconUseFont(md3.ID)),
		ui.IconButtonSize(32),
		ui.IconButtonForeground(captionButtonForeground(disabled)),
		ui.IconButtonDecoration(captionButtonDecoration()),
		ui.IconButtonDisabled(disabled),
		ui.IconButtonOnClick(onClick),
	)
}

func captionButtonDecoration() ui.Decoration {
	return ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 18}).
		WithRad(6).
		WithHover(ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 34}).WithRad(6)).
		WithPressed(ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 46}).WithRad(6)).
		WithDisabled(ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 10}).WithRad(6))
}

func captionButtonForeground(disabled bool) color.NRGBA {
	if disabled {
		return color.NRGBA{R: 220, G: 229, B: 238, A: 95}
	}
	return color.NRGBA{R: 255, G: 255, B: 255, A: 245}
}

func captionMaximizeAvailable(state ui.WindowState) bool {
	return state.Resizable && !(state.MaxWidth > 0 && state.MaxHeight > 0)
}

func framePanel(state ui.WindowState, status interface {
	Set(string)
}) ui.Element {
	frame := state.WindowsFrameStyle
	return section("Frame", ui.ColumnElement(
		ui.TextElement("Modern chrome keeps the Windows frame hidden and draws the title bar/border in FluxUI.", ui.TextSize(12)),
		ui.VSpacerElement(10),
		ui.RowElement(
			chromeButton("Modern frame", frame.Mode == ui.WindowsFrameHidden && frame.Border == ui.WindowsFrameBorderHidden, false, func(ctx *ui.Context) {
				applyFrameStyle(ctx, status, ui.WindowsFrameStyle{
					Mode:   ui.WindowsFrameHidden,
					Shadow: true,
					Corner: ui.WindowsCornerRound,
					Border: ui.WindowsFrameBorderHidden,
				})
			}),
			ui.HSpacerElement(8),
			chromeButton("Modern border", frame.Mode == ui.WindowsFrameHidden && frame.Border == ui.WindowsFrameBorderColor, false, func(ctx *ui.Context) {
				applyFrameStyle(ctx, status, ui.WindowsFrameStyle{
					Mode:        ui.WindowsFrameHidden,
					Shadow:      true,
					Corner:      ui.WindowsCornerRound,
					Border:      ui.WindowsFrameBorderColor,
					BorderColor: color.NRGBA{R: 59, G: 130, B: 246, A: 255},
				})
			}),
		),
		ui.VSpacerElement(8),
		ui.RowElement(
			chromeButton("Round", frame.Corner == ui.WindowsCornerRound, false, func(ctx *ui.Context) {
				frame.Mode = ui.WindowsFrameHidden
				frame.Corner = ui.WindowsCornerRound
				frame.Shadow = true
				applyFrameStyle(ctx, status, frame)
			}),
			ui.HSpacerElement(8),
			chromeButton("Small round", frame.Corner == ui.WindowsCornerRoundSmall, false, func(ctx *ui.Context) {
				frame.Mode = ui.WindowsFrameHidden
				frame.Corner = ui.WindowsCornerRoundSmall
				frame.Shadow = true
				applyFrameStyle(ctx, status, frame)
			}),
			ui.HSpacerElement(8),
			chromeButton("Square", frame.Corner == ui.WindowsCornerDoNotRound, false, func(ctx *ui.Context) {
				frame.Mode = ui.WindowsFrameHidden
				frame.Corner = ui.WindowsCornerDoNotRound
				frame.Shadow = true
				applyFrameStyle(ctx, status, frame)
			}),
		),
		ui.VSpacerElement(8),
		ui.RowElement(
			chromeButton("Border hidden", frame.Border == ui.WindowsFrameBorderHidden, false, func(ctx *ui.Context) {
				frame.Mode = ui.WindowsFrameHidden
				frame.Border = ui.WindowsFrameBorderHidden
				applyFrameStyle(ctx, status, frame)
			}),
			ui.HSpacerElement(8),
			chromeButton("Border color", frame.Border == ui.WindowsFrameBorderColor, false, func(ctx *ui.Context) {
				frame.Mode = ui.WindowsFrameHidden
				frame.Shadow = true
				frame.Corner = ui.WindowsCornerRound
				frame.Border = ui.WindowsFrameBorderColor
				frame.BorderColor = color.NRGBA{R: 59, G: 130, B: 246, A: 255}
				applyFrameStyle(ctx, status, frame)
			}),
			ui.HSpacerElement(8),
			chromeButton("Shadow off", !frame.Shadow, false, func(ctx *ui.Context) {
				frame.Mode = ui.WindowsFrameHidden
				frame.Shadow = false
				applyFrameStyle(ctx, status, frame)
			}),
		),
	))
}

func dragPanel() ui.Element {
	return section("Drag areas", ui.ColumnElement(
		ui.TextElement("Wrap non-interactive custom title bar regions with WindowDragAreaElement.", ui.TextSize(12)),
		ui.VSpacerElement(4),
		ui.TextElement("Double click a drag strip to maximize or restore when the window state allows it.", ui.TextSize(12)),
		ui.VSpacerElement(10),
		ui.WindowDragAreaElement(
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 226, G: 232, B: 240, A: 255}).WithPad(ui.Symmetric(10, 12)).WithRad(8),
				ui.TextElement("Drag this strip", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 30, G: 41, B: 59, A: 255})),
			),
		),
		ui.VSpacerElement(8),
		ui.WindowDragAreaElement(
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 219, G: 234, B: 254, A: 255}).WithPad(ui.Symmetric(10, 12)).WithRad(8),
				ui.TextElement("Another draggable region", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 30, G: 64, B: 175, A: 255})),
			),
		),
		ui.VSpacerElement(8),
		ui.WindowDragAreaElement(
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 241, G: 245, B: 249, A: 255}).WithPad(ui.Symmetric(10, 12)).WithRad(8),
				ui.TextElement("Disabled drag region", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 71, G: 85, B: 105, A: 255})),
			),
			ui.WindowDragAreaDisabled(true),
		),
	))
}

func snapFlyoutPanel(state ui.WindowState, status interface {
	Set(string)
}) ui.Element {
	return section("Snap flyout", ui.RowElement(
		ui.ExpandedElement(ui.ColumnElement(
			ui.TextElement("Page-level native maximize region", ui.TextSize(12)),
			ui.VSpacerElement(4),
			ui.TextElement("This control is not part of the title bar, but still participates in Windows hit testing.", ui.TextSize(12)),
		)),
		ui.HSpacerElement(14),
		pageMaximizeButton(state, status),
	))
}

func pageMaximizeButton(state ui.WindowState, status interface {
	Set(string)
}) ui.Element {
	disabled := !captionMaximizeAvailable(state) || state.Fullscreen || state.Minimized
	icon := "crop_square"
	onClick := func(ctx *ui.Context) {
		if !ui.WindowMaximize(ctx) {
			status.Set("Page maximize request failed.")
			return
		}
		status.Set("Page maximize requested.")
	}
	if state.Maximized {
		icon = "filter_none"
		onClick = func(ctx *ui.Context) {
			if !ui.WindowRestore(ctx) {
				status.Set("Page restore request failed.")
				return
			}
			status.Set("Page restore requested.")
		}
	}
	return ui.WindowMaximizeButtonElement(
		ui.IconButtonElement(
			ui.IconElement(icon, ui.IconSize(20), ui.IconUseFont(md3.ID)),
			ui.IconButtonSize(40),
			ui.IconButtonForeground(pageMaximizeButtonForeground(disabled)),
			ui.IconButtonDecoration(pageMaximizeButtonDecoration()),
			ui.IconButtonDisabled(disabled),
			ui.IconButtonOnClick(onClick),
		),
		ui.WindowMaximizeButtonDisabled(disabled),
	)
}

func pageMaximizeButtonDecoration() ui.Decoration {
	return ui.Bg(color.NRGBA{R: 226, G: 232, B: 240, A: 255}).
		WithRad(8).
		WithHover(ui.Bg(color.NRGBA{R: 203, G: 213, B: 225, A: 255}).WithRad(8)).
		WithPressed(ui.Bg(color.NRGBA{R: 148, G: 163, B: 184, A: 255}).WithRad(8)).
		WithDisabled(ui.Bg(color.NRGBA{R: 241, G: 245, B: 249, A: 255}).WithRad(8))
}

func pageMaximizeButtonForeground(disabled bool) color.NRGBA {
	if disabled {
		return color.NRGBA{R: 148, G: 163, B: 184, A: 255}
	}
	return color.NRGBA{R: 15, G: 23, B: 42, A: 255}
}

func runtimePanel(status interface {
	Set(string)
}) ui.Element {
	return section("Runtime", ui.ColumnElement(
		ui.TextElement(availabilityText(ui.ProbeWindowsChrome()), ui.TextSize(12)),
		ui.VSpacerElement(10),
		ui.RowElement(
			chromeButton("Probe", false, false, func(ctx *ui.Context) {
				status.Set(availabilityText(ui.ProbeWindowsChrome()))
			}),
			ui.HSpacerElement(8),
			chromeButton("Focus", false, false, func(ctx *ui.Context) {
				if !ui.WindowRequestFocus(ctx) {
					status.Set("Focus request failed.")
					return
				}
				status.Set("Focus requested.")
			}),
			ui.HSpacerElement(8),
			chromeButton("Center", false, false, func(ctx *ui.Context) {
				if !ui.WindowCenter(ctx) {
					status.Set("Center request failed.")
					return
				}
				status.Set("Window centered.")
			}),
		),
		ui.VSpacerElement(8),
		ui.TextElement("Drag the custom title area at the top to move the whole window.", ui.TextSize(12)),
	))
}

func notesPanel() ui.Element {
	return section("Deferred", ui.ColumnElement(
		ui.TextElement("Windows background material, native transparency, and window background color are deferred. They need a renderer-level design before becoming stable API.", ui.TextSize(12)),
		ui.VSpacerElement(6),
		ui.TextElement("WindowsFrameDefault is kept as a compatibility escape hatch, but this demo avoids it because Windows can render that native Win32 frame with an older visual style.", ui.TextSize(12)),
	))
}

func windowShellDecoration(state ui.WindowState) ui.Decoration {
	deco := ui.Bg(color.NRGBA{R: 244, G: 247, B: 251, A: 255}).WithPad(ui.All(0))
	if showModernFrameBorder(state) {
		deco = deco.WithBorder(ui.Border{Width: 1, Color: modernFrameBorderColor(state.WindowsFrameStyle)})
	}
	return deco
}

func showModernFrameBorder(state ui.WindowState) bool {
	return state.WindowsFrameStyle.Mode == ui.WindowsFrameHidden &&
		state.WindowsFrameStyle.Border == ui.WindowsFrameBorderColor &&
		!state.Maximized &&
		!state.Fullscreen
}

func modernFrameBorderColor(style ui.WindowsFrameStyle) color.NRGBA {
	if style.BorderColor.A > 0 {
		return style.BorderColor
	}
	return color.NRGBA{R: 59, G: 130, B: 246, A: 255}
}

func section(title string, child ui.Element) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).
			WithPad(ui.All(14)).
			WithRad(8).
			WithBorder(ui.Border{Width: 1, Color: color.NRGBA{R: 210, G: 218, B: 226, A: 255}}),
		ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(16)),
			ui.VSpacerElement(8),
			child,
		),
	)
}

func chromeButton(label string, active bool, disabled bool, onClick func(*ui.Context)) ui.Element {
	opts := []ui.ButtonOption{
		ui.ButtonPadding(ui.Symmetric(6, 10)),
		ui.Disabled(disabled),
		ui.OnClick(onClick),
	}
	if active {
		return ui.FilledTonalButtonElement(ui.TextElement(label, ui.TextSize(12)), opts...)
	}
	return ui.OutlinedButtonElement(ui.TextElement(label, ui.TextSize(12)), opts...)
}

func applyFrameStyle(ctx *ui.Context, status interface {
	Set(string)
}, frame ui.WindowsFrameStyle) {
	if !ui.WindowSetWindowsFrameStyle(ctx, frame) {
		status.Set("Frame style update failed. This requires Windows and an initialized HWND.")
		return
	}
	status.Set("Frame style updated: " + frameModeName(frame.Mode) + ".")
}

func chromeSummary(state ui.WindowState, availability ui.WindowsChromeAvailability) string {
	return fmt.Sprintf(
		"frame=%s corner=%s border=%s supported=%v",
		frameModeName(state.WindowsFrameStyle.Mode),
		cornerName(state.WindowsFrameStyle.Corner),
		borderName(state.WindowsFrameStyle.Border),
		availability.Supported,
	)
}

func availabilityText(value ui.WindowsChromeAvailability) string {
	return fmt.Sprintf(
		"Windows chrome supported=%v frame=%v drag=%v",
		value.Supported,
		value.FrameStyle,
		value.DragMove,
	)
}

func frameModeName(mode ui.WindowsFrameMode) string {
	if mode == ui.WindowsFrameHidden {
		return "hidden"
	}
	return "default"
}

func cornerName(corner ui.WindowsCornerPreference) string {
	switch corner {
	case ui.WindowsCornerDoNotRound:
		return "square"
	case ui.WindowsCornerRound:
		return "round"
	case ui.WindowsCornerRoundSmall:
		return "small"
	default:
		return "default"
	}
}

func borderName(border ui.WindowsFrameBorderPolicy) string {
	switch border {
	case ui.WindowsFrameBorderHidden:
		return "hidden"
	case ui.WindowsFrameBorderColor:
		return "color"
	default:
		return "default"
	}
}

func main() {
	_ = ui.RunElement(
		app,
		ui.Title("FluxUI Windows Chrome Showcase"),
		ui.Size(860, 560),
		ui.MinSize(720, 460),
		ui.WindowsFrame(ui.WindowsFrameStyle{
			Mode:   ui.WindowsFrameHidden,
			Shadow: true,
			Corner: ui.WindowsCornerRound,
			Border: ui.WindowsFrameBorderHidden,
		}),
	)
}
