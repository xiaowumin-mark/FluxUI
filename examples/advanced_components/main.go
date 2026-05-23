package main

import (
	"fmt"
	"image/color"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func rowColor(index int) color.NRGBA {
	if index%2 == 0 {
		return ui.NRGBA(245, 247, 250, 255)
	}
	return ui.NRGBA(235, 240, 246, 255)
}

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)

	activeTab := ui.UseState(ctx, "feed")
	activeNav := ui.UseState(ctx, "home")
	selectVal := ui.UseState(ctx, "low")
	showDialog := ui.UseState(ctx, false)
	toastMsg := ui.UseState(ctx, "")
	scrollTip := ui.UseState(ctx, "")
	reachCount := ui.UseState(ctx, 0)

	levelOptions := []ui.SelectOptionItem[string]{
		{Label: "低优先级", Value: "low"},
		{Label: "中优先级", Value: "medium"},
		{Label: "高优先级", Value: "high"},
	}

	tabs := []ui.TabItem{
		{Key: "feed", Label: "动态流"},
		{Key: "tasks", Label: "任务"},
		{Key: "media", Label: "媒体"},
		{Key: "settings", Label: "设置"},
	}

	navs := []ui.ElementNavItem{
		{Key: "home", Label: "首页", Icon: ui.TextElement("H", ui.TextColor(th.TextColor))},
		{Key: "discover", Label: "发现", Icon: ui.TextElement("D", ui.TextColor(th.TextColor))},
		{Key: "profile", Label: "我的", Icon: ui.TextElement("P", ui.TextColor(th.TextColor))},
	}

	header := ui.ContainerDecorationElement(
		ui.Bg(th.Primary).WithPad(ui.Symmetric(10, 14)).WithRad(10),
		ui.RowElement(
			ui.TextElement("FluxUI 高级能力示例", ui.TextSize(16), ui.TextColor(th.TextOnPrimary)),
			ui.ExpandedElement(ui.SpacerElement(0, 0)),
			ui.ButtonElement(
				ui.TextElement("弹窗"),
				ui.ButtonPadding(ui.Symmetric(6, 10)),
				ui.ButtonBackground(ui.NRGBA(255, 255, 255, 30)),
				ui.ButtonForeground(th.TextOnPrimary),
				ui.OnClick(func(ctx *ui.Context) { showDialog.Set(true) }),
			),
		),
	)

	tabBar := ui.TabsElement(
		activeTab.Value(),
		tabs,
		ui.TabsScrollable(true),
		ui.TabsOnChange(func(ctx *ui.Context, key string) {
			activeTab.Set(key)
			toastMsg.Set("切换标签: " + key)
		}),
	)

	selectRow := ui.CardElement(
		ui.ColumnElement(
			ui.TextElement("下拉选择（真实展开）", ui.TextSize(14)),
			ui.SpacerElement(0, 8),
			ui.SelectElement(
				selectVal.Value(),
				levelOptions,
				ui.SelectPlaceholder[string]("请选择优先级"),
				ui.SelectMaxHeight[string](180),
				ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
					selectVal.Set(value)
					toastMsg.Set("优先级切换为: " + value)
				}),
			),
		),
	)

	imageDemo := ui.CardElement(
		ui.ColumnElement(
			ui.TextElement("图片渲染（含 fit/radius/click）", ui.TextSize(14)),
			ui.SpacerElement(0, 8),
			ui.RowElement(
				ui.PaddingElement(
					ui.Insets{Right: 8},
					ui.ImageElement(
						ui.ImageSource{Path: "examples/assets/sample.png"},
						ui.ImageWidth(120),
						ui.ImageHeight(80),
						ui.ImageFitMode(ui.ImageFitContain),
						ui.ImageRadius(8),
						ui.ImageOnClick(func(ctx *ui.Context) { toastMsg.Set("点击了 Contain 图片") }),
					),
				),
				ui.ImageElement(
					ui.ImageSource{Path: "examples/assets/sample.png"},
					ui.ImageWidth(120),
					ui.ImageHeight(80),
					ui.ImageFitMode(ui.ImageFitCover),
					ui.ImageRadius(18),
				),
			),
		),
	)

	longList := ui.CardElement(
		ui.ColumnElement(
			ui.TextElement("滚动 + 虚拟列表", ui.TextSize(14)),
			ui.SpacerElement(0, 8),
			ui.ListViewElement(
				120,
				func(ctx *ui.Context, index int) ui.Element {
					return ui.ContainerDecorationElement(
						ui.Bg(rowColor(index)).WithPad(ui.Symmetric(8, 10)).WithRad(6),
						ui.RowElement(
							ui.TextElement(fmt.Sprintf("#%03d", index), ui.TextSize(12)),
							ui.SpacerElement(10, 0),
							ui.TextElement("列表项内容"),
						),
					)
				},
				ui.ListItemSpacing(6),
				ui.ListOnReachEnd(func(ctx *ui.Context) { reachCount.Set(reachCount.Value() + 1) }),
			),
			ui.SpacerElement(0, 8),
			ui.TextElement(
				fmt.Sprintf("触底回调次数: %d | %s", reachCount.Value(), scrollTip.Value()),
				ui.TextSize(12),
				ui.TextColor(th.SurfaceMuted),
			),
		),
	)

	homeContent := ui.ScrollViewElement(
		ui.ColumnElement(
			ui.PaddingElement(ui.Insets{Bottom: 10}, tabBar),
			ui.PaddingElement(ui.Insets{Bottom: 10}, selectRow),
			ui.PaddingElement(ui.Insets{Bottom: 10}, imageDemo),
			ui.PaddingElement(ui.Insets{Bottom: 10}, longList),
			ui.SpacerElement(0, 24),
		),
		ui.ScrollVertical(true),
		ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) { scrollTip.Set(fmt.Sprintf("滚动偏移 y=%.2f", y)) }),
	)

	discoverContent := ui.ScrollViewElement(
		ui.ColumnElement(
			ui.CardElement(
				ui.ColumnElement(
					ui.TextElement("发现页", ui.TextSize(15)),
					ui.SpacerElement(0, 8),
					ui.TextElement("这里用于展示推荐内容、热点卡片和探索能力。"),
				),
			),
			ui.SpacerElement(0, 10),
			ui.CardElement(
				ui.ColumnElement(
					ui.TextElement("当前标签", ui.TextSize(14)),
					ui.SpacerElement(0, 6),
					ui.TextElement(activeTab.Value(), ui.TextColor(th.Primary)),
				),
			),
			ui.SpacerElement(0, 20),
		),
		ui.ScrollVertical(true),
	)

	profileContent := ui.ScrollViewElement(
		ui.ColumnElement(
			ui.CardElement(
				ui.ColumnElement(
					ui.TextElement("个人中心", ui.TextSize(15)),
					ui.SpacerElement(0, 8),
					ui.TextElement("用于展示账号信息、偏好设置和统计数据。"),
				),
			),
			ui.SpacerElement(0, 10),
			ui.CardElement(
				ui.ColumnElement(
					ui.TextElement("当前优先级", ui.TextSize(14)),
					ui.SpacerElement(0, 6),
					ui.TextElement(selectVal.Value(), ui.TextColor(th.Primary)),
				),
			),
			ui.SpacerElement(0, 20),
		),
		ui.ScrollVertical(true),
	)

	content := homeContent
	switch activeNav.Value() {
	case "discover":
		content = discoverContent
	case "profile":
		content = profileContent
	}

	page := ui.ColumnElement(
		header,
		ui.ExpandedElement(
			ui.PaddingElement(
				ui.Insets{Left: 12, Right: 12, Top: 10},
				content,
			),
		),
		ui.BottomNavigationElement(
			activeNav.Value(),
			navs,
			ui.BottomNavAlignmentOf(ui.BottomNavAlignSpaceEvenly),
			ui.BottomNavOnChange(func(ctx *ui.Context, key string) {
				activeNav.Set(key)
				toastMsg.Set("切换到底部导航: " + key)
			}),
		),
	)

	dialog := ui.DialogElement(
		showDialog.Value(),
		ui.TextElement("这是一个带遮罩的对话框，支持点击蒙层关闭。"),
		ui.DialogTitle("操作确认"),
		ui.DialogWidth(320),
		ui.DialogMaskClosable(true),
		ui.DialogOnOpenChange(func(ctx *ui.Context, open bool) { showDialog.Set(open) }),
		ui.DialogOnCancel(func(ctx *ui.Context) { showDialog.Set(false) }),
		ui.DialogOnConfirm(func(ctx *ui.Context) {
			showDialog.Set(false)
			toastMsg.Set("你点击了确定")
		}),
	)

	var layers []ui.Element
	layers = append(layers, page)
	if showDialog.Value() {
		layers = append(layers, dialog)
	}
	if toastMsg.Value() != "" {
		layers = append(layers, ui.ToastElement(
			toastMsg.Value(),
			ui.ToastTypeOf(ui.ToastSuccess),
			ui.ToastPositionOf(ui.ToastBottom),
			ui.ToastDuration(2200*time.Millisecond),
			ui.ToastOnClose(func(ctx *ui.Context) { toastMsg.Set("") }),
		))
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface),
		ui.StackElement(layers...),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Advanced Components"), ui.Size(760, 920))
}
