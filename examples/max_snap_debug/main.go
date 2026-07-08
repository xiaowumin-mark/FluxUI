package main

import (
	"fmt"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/icons/md3"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	app := func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		popupOpen := ui.UseState(ctx, false)
		frameHidden := ui.UseState(ctx, false)

		state := ui.WindowState{}
		currentID := ui.CurrentWindowID(ctx)
		if handle, ok := ui.GetWindow(currentID); ok {
			if next, ok := handle.State(); ok {
				state = next
			}
		}

		disabled := !state.Resizable || state.Minimized || state.Fullscreen ||
			(state.MaxWidth > 0 && state.MaxHeight > 0)

		frameLabel := "Win32 标准边框"
		if frameHidden.Value() {
			frameLabel = "隐藏边框 (自绘)"
		}

		return ui.ContainerDecorationElement(
			ui.Bg(color.NRGBA{R: 240, G: 245, B: 250, A: 255}).WithPad(ui.All(16)),
			ui.StackElement(
				ui.ColumnElement(
					func() ui.Element {
						if frameHidden.Value() {
							return ui.WindowDragAreaElement(
								ui.ContainerDecorationElement(
									ui.Bg(color.NRGBA{R: 30, G: 40, B: 55, A: 255}).WithPad(ui.Symmetric(8, 16)).WithRad(8),
									ui.RowElement(
										ui.TextElement("Max Snap Debug — 运行时切换边框测试",
											ui.TextSize(14), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
										ui.ExpandedElement(ui.SpacerElement(0, 0)),
										ui.TextElement("← 拖动此处移动窗口", ui.TextSize(11),
											ui.TextColor(color.NRGBA{R: 180, G: 190, B: 200, A: 255})),
									),
								),
							)
						}
						return ui.SpacerElement(0, 0)
					}(),
					ui.VSpacerElement(12),

					ui.RowElement(
						ui.OutlinedButtonElement(
							ui.TextElement("打开弹窗 (Popup)", ui.TextSize(14)),
							ui.ButtonPadding(ui.Symmetric(10, 18)),
							ui.OnClick(func(ctx *ui.Context) {
								popupOpen.Set(!popupOpen.Value())
							}),
						),
						ui.HSpacerElement(12),
						ui.FilledButtonElement(
							ui.TextElement("切换边框", ui.TextSize(14)),
							ui.ButtonPadding(ui.Symmetric(10, 18)),
							ui.OnClick(func(ctx *ui.Context) {
								next := !frameHidden.Value()
								frameHidden.Set(next)
								if next {
									ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
										Mode:   ui.WindowsFrameHidden,
										Shadow: true,
										Corner: ui.WindowsCornerRound,
										Border: ui.WindowsFrameBorderHidden,
									})
								} else {
									ui.WindowSetWindowsFrameStyle(ctx, ui.WindowsFrameStyle{
										Mode: ui.WindowsFrameDefault,
									})
								}
							}),
						),
						ui.HSpacerElement(12),
						ui.TextElement(fmt.Sprintf("当前: %s | resizable=%v maximized=%v disabled=%v",
							frameLabel, state.Resizable, state.Maximized, disabled),
							ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
					),

					ui.VSpacerElement(16),

					ui.ContainerDecorationElement(
						ui.Bg(th.Colors.SurfaceContainer).
							WithPad(ui.All(14)).
							WithRad(10).
							WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
						ui.ColumnElement(
							ui.TextElement("测试步骤", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
							ui.VSpacerElement(6),
							ui.TextElement("1. 点击「打开弹窗」", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
							ui.VSpacerElement(3),
							ui.TextElement("2. 点击「切换边框」隐藏 Win32 边框", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
							ui.VSpacerElement(3),
							ui.TextElement("3. 悬停弹窗内的 Snap Flyout 按钮", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
							ui.VSpacerElement(3),
							ui.TextElement("4. 观察能否弹出 Windows 布局小窗", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
						),
					),
				),

				buildPopup(popupOpen.Value(), func(open bool) {
					popupOpen.Set(open)
				}, disabled, th),
			),
		)
	}

	_ = ui.RunElement(app, ui.Title("Max Snap Debug"), ui.Size(900, 600))
}

func buildPopup(open bool, onClose func(bool), disabled bool, th *ui.Theme) ui.Element {
	if !open {
		return ui.PopupElement(false, ui.SpacerElement(0, 0))
	}

	content := ui.ColumnElement(
		ui.RowElement(
			ui.TextElement("弹窗 - 悬停下方的 Snap Flyout 按钮测试", ui.TextSize(16), ui.TextColor(th.Colors.OnSurface)),
			ui.ExpandedElement(ui.SpacerElement(0, 0)),
			ui.TextButtonElement(
				ui.TextElement("关闭", ui.TextSize(12)),
				ui.ButtonPadding(ui.Symmetric(5, 10)),
				ui.OnClick(func(ctx *ui.Context) { onClose(false) }),
			),
		),
		ui.VSpacerElement(12),
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(12)).WithRad(12).
				WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.FixedHeightElement(360,
				ui.ScrollViewElement(
					ui.ColumnElement(
						ui.TextElement("滚动区域内的内容", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
						ui.VSpacerElement(8),
						ui.TextElement("下方是一个 WindowMaximizeButton，包裹在 TooltipElement 中，位于 ScrollView 内部。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
						ui.VSpacerElement(12),

						ui.ContainerDecorationElement(
							ui.Bg(th.Colors.SurfaceContainerHigh).WithPad(ui.All(12)).WithRad(8).
								WithBorder(ui.Border{Width: 1, Color: th.Colors.Outline}),
							ui.RowElement(
								ui.ExpandedElement(ui.ColumnElement(
									ui.TextElement("Snap Flyout 按钮", ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
									ui.VSpacerElement(4),
									ui.TextElement("悬停此按钮触发 Windows Snap Flyout 布局小窗。", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
								)),
								ui.HSpacerElement(8),
								buildMaxButton(disabled, th),
							),
						),

						buildManySpacers(20),
					),
					ui.ScrollVertical(true),
				),
			),
		),
	)

	return ui.PopupElement(
		open,
		content,
		ui.PopupWidth(800),
		ui.PopupMaskAlpha(112),
		ui.PopupMaskClosable(true),
		ui.PopupOnOpenChange(func(ctx *ui.Context, open bool) { onClose(open) }),
	)
}

func buildMaxButton(disabled bool, th *ui.Theme) ui.Element {
	icon := "crop_square"
	return ui.TooltipElement(
		"Windows Snap Flyout (maximize)",
		ui.WindowMaximizeButtonElement(
			ui.IconButtonElement(
				ui.IconElement(icon, ui.IconSize(18), ui.IconUseFont(md3.ID)),
				ui.IconButtonSize(40),
				ui.IconButtonForeground(th.Colors.OnSurface),
				ui.IconButtonDecoration(ui.Bg(th.Colors.SurfaceContainer).
					WithRad(8).
					WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}).
					WithHover(ui.Bg(th.Colors.PrimaryContainer).WithRad(8)),
				),
				ui.IconButtonDisabled(disabled),
				ui.IconButtonOnClick(func(ctx *ui.Context) {
					fmt.Println("[CLICK] SnapFlyout button clicked")
					if !ui.WindowMaximize(ctx) {
						fmt.Println("[CLICK] WindowMaximize failed")
					}
				}),
			),
			ui.WindowMaximizeButtonDisabled(disabled),
		),
	)
}

func buildManySpacers(count int) ui.Element {
	return ui.ColumnElement(func() []ui.Element {
		items := make([]ui.Element, count)
		for i := 0; i < count; i++ {
			items[i] = ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 200, G: 210, B: 225, A: 180}).WithPad(ui.Symmetric(6, 8)).WithRad(4),
				ui.TextElement(fmt.Sprintf("Spacer #%d — 填充滚动区域", i+1), ui.TextSize(11),
					ui.TextColor(color.NRGBA{R: 80, G: 90, B: 110, A: 255})),
			)
		}
		return items
	}()...)
}
