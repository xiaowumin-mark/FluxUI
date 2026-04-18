package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		activeTab := ui.State[string](ctx)
		if activeTab.Value() == "" {
			activeTab.Set("feed")
		}

		activeNav := ui.State[string](ctx)
		if activeNav.Value() == "" {
			activeNav.Set("home")
		}

		selectVal := ui.State[string](ctx)
		if selectVal.Value() == "" {
			selectVal.Set("low")
		}

		showDialog := ui.State[bool](ctx)
		toastMsg := ui.State[string](ctx)
		scrollTip := ui.State[string](ctx)
		reachCount := ui.State[int](ctx)

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

		navs := []ui.NavItem{
			{Key: "home", Label: "首页", Icon: ui.Text("H", ui.TextColor(th.TextColor))},
			{Key: "discover", Label: "发现", Icon: ui.Text("D", ui.TextColor(th.TextColor))},
			{Key: "profile", Label: "我的", Icon: ui.Text("P", ui.TextColor(th.TextColor))},
		}

		header := ui.AppBar(
			ui.Text("FluxUI 高级能力示例", ui.TextSize(16)),
			ui.AppBarActions(
				ui.Button(
					ui.Text("弹窗"),
					ui.ButtonPadding(ui.Symmetric(6, 10)),
					ui.OnClick(func(ctx *ui.Context) {
						showDialog.Set(true)
					}),
				),
			),
		)

		tabBar := ui.Tabs(
			activeTab.Value(),
			tabs,
			ui.TabsScrollable(true),
			ui.TabsOnChange(func(ctx *ui.Context, key string) {
				activeTab.Set(key)
				toastMsg.Set("切换标签: " + key)
			}),
		)

		selectRow := ui.Card(
			ui.Column(
				ui.Text("下拉选择（真实展开）", ui.TextSize(14)),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.Select(
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
			),
		)

		imageDemo := ui.Card(
			ui.Column(
				ui.Text("图片渲染（含 fit/radius/click）", ui.TextSize(14)),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.Row(
						ui.Padding(
							ui.Insets{Right: 8},
							ui.Image(
								ui.ImageSource{Path: "examples/assets/sample.png"},
								ui.ImageWidth(120),
								ui.ImageHeight(80),
								ui.ImageFitMode(ui.ImageFitContain),
								ui.ImageRadius(8),
								ui.ImageOnClick(func(ctx *ui.Context) {
									toastMsg.Set("点击了 Contain 图片")
								}),
							),
						),
						ui.Image(
							ui.ImageSource{Path: "examples/assets/sample.png"},
							ui.ImageWidth(120),
							ui.ImageHeight(80),
							ui.ImageFitMode(ui.ImageFitCover),
							ui.ImageRadius(18),
						),
					),
				),
			),
		)

		longList := ui.Card(
			ui.Column(
				ui.Text("滚动 + 虚拟列表", ui.TextSize(14)),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.ListView(
						120,
						func(ctx *ui.Context, index int) ui.Widget {
							return ui.Container(
								ui.Style{
									Background: rowColor(index),
									Padding:    ui.Symmetric(8, 10),
									Radius:     6,
								},
								ui.Row(
									ui.Text(fmt.Sprintf("#%03d", index), ui.TextSize(12)),
									ui.Padding(ui.Insets{Left: 10}, ui.Text("列表项内容")),
								),
							)
						},
						ui.ListItemSpacing(6),
						ui.ListOnReachEnd(func(ctx *ui.Context) {
							reachCount.Set(reachCount.Value() + 1)
						}),
					),
				),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.Text(
						fmt.Sprintf("触底回调次数: %d | %s", reachCount.Value(), scrollTip.Value()),
						ui.TextSize(12),
						ui.TextColor(th.SurfaceMuted),
					),
				),
			),
		)

		homeContent := ui.ScrollView(
			ui.Column(
				ui.Padding(ui.Insets{Bottom: 10}, tabBar),
				ui.Padding(ui.Insets{Bottom: 10}, selectRow),
				ui.Padding(ui.Insets{Bottom: 10}, imageDemo),
				ui.Padding(ui.Insets{Bottom: 10}, longList),
				ui.VSpacer(24),
			),
			ui.ScrollVertical(true),
			ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
				scrollTip.Set(fmt.Sprintf("滚动偏移 y=%.2f", y))
			}),
		)

		discoverContent := ui.ScrollView(
			ui.Column(
				ui.Card(
					ui.Column(
						ui.Text("发现页", ui.TextSize(15)),
						ui.Padding(ui.Insets{Top: 8}, ui.Text("这里用于展示推荐内容、热点卡片和探索能力。")),
					),
				),
				ui.Padding(
					ui.Insets{Top: 10},
					ui.Card(
						ui.Column(
							ui.Text("当前标签", ui.TextSize(14)),
							ui.Padding(ui.Insets{Top: 6}, ui.Text(activeTab.Value(), ui.TextColor(th.Primary))),
						),
					),
				),
				ui.VSpacer(20),
			),
			ui.ScrollVertical(true),
		)

		profileContent := ui.ScrollView(
			ui.Column(
				ui.Card(
					ui.Column(
						ui.Text("个人中心", ui.TextSize(15)),
						ui.Padding(ui.Insets{Top: 8}, ui.Text("用于展示账号信息、偏好设置和统计数据。")),
					),
				),
				ui.Padding(
					ui.Insets{Top: 10},
					ui.Card(
						ui.Column(
							ui.Text("当前优先级", ui.TextSize(14)),
							ui.Padding(ui.Insets{Top: 6}, ui.Text(selectVal.Value(), ui.TextColor(th.Primary))),
						),
					),
				),
				ui.VSpacer(20),
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

		page := ui.Column(
			header,
			ui.Expanded(
				ui.Padding(
					ui.Insets{Left: 12, Right: 12, Top: 10},
					content,
				),
			),
			ui.BottomNavigation(
				activeNav.Value(),
				navs,
				ui.BottomNavAlignmentOf(ui.BottomNavAlignSpaceEvenly),
				ui.BottomNavOnChange(func(ctx *ui.Context, key string) {
					activeNav.Set(key)
					toastMsg.Set("切换到底部导航: " + key)
				}),
			),
		)

		dialog := ui.Dialog(
			showDialog.Value(),
			ui.Text("这是一个带遮罩的对话框，支持点击蒙层关闭。"),
			ui.DialogTitle("操作确认"),
			ui.DialogWidth(320),
			ui.DialogMaskClosable(true),
			ui.DialogOnOpenChange(func(ctx *ui.Context, open bool) {
				showDialog.Set(open)
			}),
			ui.DialogOnCancel(func(ctx *ui.Context) {
				showDialog.Set(false)
			}),
			ui.DialogOnConfirm(func(ctx *ui.Context) {
				showDialog.Set(false)
				toastMsg.Set("你点击了确定")
			}),
		)

		var layers []ui.Widget
		layers = append(layers, page)
		if showDialog.Value() {
			layers = append(layers, dialog)
		}
		if toastMsg.Value() != "" {
			layers = append(layers, ui.Toast(
				toastMsg.Value(),
				ui.ToastTypeOf(ui.ToastSuccess),
				ui.ToastPositionOf(ui.ToastBottom),
				ui.ToastDuration(2200*time.Millisecond),
				ui.ToastOnClose(func(ctx *ui.Context) {
					toastMsg.Set("")
				}),
			))
		}

		return ui.Container(
			ui.Style{Background: th.Surface},
			ui.Stack(layers...),
		)
	}, ui.Title("FluxUI Advanced Components"), ui.Size(760, 920))
}

func rowColor(index int) color.NRGBA {
	if index%2 == 0 {
		return ui.NRGBA(245, 247, 250, 255)
	}
	return ui.NRGBA(235, 240, 246, 255)
}
