package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xiaowumin-mark/FluxUI/ui"
)

type docMeta struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Order    int      `json:"order"`
	Summary  string   `json:"summary"`
	Example  docDemo  `json:"example"`
	APIs     []string `json:"apis"`
}

type docDemo struct {
	ID    string            `json:"id"`
	Props map[string]string `json:"props"`
}

type widgetDoc struct {
	Meta    docMeta
	Content string
	Path    string
}

type menuEntry struct {
	IsCategory bool
	Category   string
	Doc        *widgetDoc
}

type remoteLoadResult struct {
	Docs []widgetDoc
	Err  error
}

type githubContentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

const (
	githubWidgetsAPIURL = "https://api.github.com/repos/xiaowumin-mark/FluxUI/contents/docs/widgets?ref=main"
	githubWidgetsRawURL = "https://raw.githubusercontent.com/xiaowumin-mark/FluxUI/main/docs/widgets/"
)

func main() {
	docs, loadErr := loadWidgetDocs()
	docsSource := "local"
	onlineLoading := false
	var onlineErr error
	onlineResultCh := make(chan remoteLoadResult, 1)
	if len(docs) == 0 {
		docsSource = "online"
		onlineLoading = true
		docs = []widgetDoc{buildOnlineLoadingDoc(loadErr)}
		go func() {
			remoteDocs, err := loadWidgetDocsFromGitHub()
			onlineResultCh <- remoteLoadResult{Docs: remoteDocs, Err: err}
		}()
	}

	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		if onlineLoading {
			select {
			case result := <-onlineResultCh:
				onlineLoading = false
				onlineErr = result.Err
				if len(result.Docs) > 0 {
					docs = result.Docs
					loadErr = nil
				} else {
					docs = []widgetDoc{buildOnlineLoadFailedDoc(loadErr, onlineErr)}
					if loadErr != nil && onlineErr != nil {
						loadErr = fmt.Errorf("本地加载失败: %v；在线加载失败: %v", loadErr, onlineErr)
					} else if onlineErr != nil {
						loadErr = fmt.Errorf("在线加载失败: %w", onlineErr)
					}
				}
			default:
				// 在线请求在后台进行；加载期间持续请求下一帧，确保结果到达后立即刷新 UI。
				ctx.RequestRedraw()
			}
		}

		th := ui.UseTheme(ctx)

		selectedDocID := ui.State[string](ctx)
		searchKeyword := ui.State[string](ctx)
		demoInit := ui.State[bool](ctx)

		buttonCount := ui.State[int](ctx)
		inputValue := ui.State[string](ctx)
		checkboxValue := ui.State[bool](ctx)
		switchValue := ui.State[bool](ctx)
		sliderValue := ui.State[float32](ctx)
		radioValue := ui.State[string](ctx)
		selectValue := ui.State[string](ctx)
		tabValue := ui.State[string](ctx)
		dialogOpen := ui.State[bool](ctx)
		toastMessage := ui.State[string](ctx)
		bottomNavValue := ui.State[string](ctx)
		clickCount := ui.State[int](ctx)
		appbarActionCount := ui.State[int](ctx)
		listReachEndCount := ui.State[int](ctx)
		hookDemoCount := ui.State[int](ctx)
		hookDemoShowChild := ui.State[bool](ctx)
		hookDemoLogs := ui.State[[]string](ctx)

		if !demoInit.Value() {
			inputValue.Set("FluxUI")
			checkboxValue.Set(true)
			switchValue.Set(true)
			sliderValue.Set(40)
			radioValue.Set("layout")
			selectValue.Set("medium")
			tabValue.Set("overview")
			bottomNavValue.Set("home")
			demoInit.Set(true)
		}

		if selectedDocID.Value() == "" && len(docs) > 0 {
			selectedDocID.Set(docs[0].Meta.ID)
		}

		filteredDocs := filterDocs(docs, searchKeyword.Value())
		currentDoc := findDocByID(docs, selectedDocID.Value())
		if currentDoc == nil && len(docs) > 0 {
			selectedDocID.Set(docs[0].Meta.ID)
			currentDoc = &docs[0]
		}

		buildDemo := func(doc *widgetDoc) ui.Widget {
			if doc == nil {
				return ui.Text("暂无示例")
			}

			demoID := doc.Meta.Example.ID
			if demoID == "" {
				demoID = doc.Meta.ID
			}

			switch demoID {
			case "row_basic":
				return ui.Row(
					ui.Container(
						ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(10), Radius: 6},
						ui.Text("A", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
					),
					ui.Padding(
						ui.Insets{Left: 8},
						ui.Container(
							ui.Style{Background: ui.NRGBA(67, 160, 71, 255), Padding: ui.All(10), Radius: 6},
							ui.Text("B", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						),
					),
					ui.Padding(
						ui.Insets{Left: 8},
						ui.Container(
							ui.Style{Background: ui.NRGBA(245, 124, 0, 255), Padding: ui.All(10), Radius: 6},
							ui.Text("C", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						),
					),
				)
			case "column_basic":
				return ui.Column(
					ui.Container(
						ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(8), Radius: 6},
						ui.Text("第一行", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Container(
							ui.Style{Background: ui.NRGBA(67, 160, 71, 255), Padding: ui.All(8), Radius: 6},
							ui.Text("第二行", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						),
					),
				)
			case "stack_basic":
				return ui.FixedHeight(
					120,
					ui.Fill(
						ui.Stack(
							ui.Fill(
								ui.Container(
									ui.Style{Background: ui.NRGBA(234, 239, 245, 255), Radius: 8},
									ui.Spacer(0, 0),
								),
							),
							ui.Padding(
								ui.Insets{Left: 12, Top: 12},
								ui.Container(
									ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(6), Radius: 6},
									ui.Text("Layer 1", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
								),
							),
							ui.Center(
								ui.Text("Center Layer", ui.TextColor(ui.NRGBA(15, 23, 42, 255)), ui.TextSize(14)),
							),
						),
					),
				)
			case "center_basic":
				return ui.FixedHeight(
					120,
					ui.Fill(
						ui.Container(
							ui.Style{Background: ui.NRGBA(240, 244, 248, 255), Radius: 8},
							ui.Center(ui.Text("居中内容", ui.TextSize(14))),
						),
					),
				)
			case "container_basic":
				return ui.Container(
					ui.Style{
						Background: ui.NRGBA(30, 136, 229, 255),
						Padding:    ui.All(16),
						Radius:     10,
					},
					ui.Text("Container: 背景 + 内边距 + 圆角", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				)
			case "padding_basic":
				return ui.Container(
					ui.Style{
						Background: ui.NRGBA(229, 236, 246, 255),
						Radius:     8,
					},
					ui.Padding(
						ui.All(16),
						ui.Text("Padding: 这里有 16dp 内边距"),
					),
				)
			case "spacer_basic":
				return ui.Row(
					ui.Text("左"),
					ui.HSpacer(24),
					ui.Text("右"),
					ui.HSpacer(24),
					ui.Column(
						ui.Text("上"),
						ui.VSpacer(8),
						ui.Text("下"),
					),
				)
			case "divider_basic":
				return ui.Column(
					ui.Text("第一段内容"),
					ui.Divider(ui.DividerThickness(1), ui.DividerColor(ui.NRGBA(176, 190, 197, 255)), ui.DividerMargin(ui.Insets{Top: 8, Bottom: 8})),
					ui.Text("第二段内容"),
				)
			case "sizing_basic":
				return ui.Column(
					ui.Row(
						ui.FixedWidth(
							110,
							ui.Container(
								ui.Style{Background: ui.NRGBA(3, 169, 244, 255), Padding: ui.All(8), Radius: 6},
								ui.Text("FixedWidth", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
							),
						),
						ui.Padding(
							ui.Insets{Left: 8},
							ui.Expanded(
								ui.Container(
									ui.Style{Background: ui.NRGBA(76, 175, 80, 255), Padding: ui.All(8), Radius: 6},
									ui.Text("Expanded / Fill", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
								),
							),
						),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.FixedHeight(
							48,
							ui.FillWidth(
								ui.Container(
									ui.Style{Background: ui.NRGBA(255, 152, 0, 255), Padding: ui.All(8), Radius: 6},
									ui.Text("FixedHeight", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
								),
							),
						),
					),
				)
			case "click_area_basic":
				return ui.Column(
					ui.Text(fmt.Sprintf("点击次数: %d", clickCount.Value())),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.ClickArea(
							ui.FillWidth(
								ui.Container(
									ui.Style{
										Background: ui.NRGBA(227, 242, 253, 255),
										Padding:    ui.All(14),
										Radius:     8,
									},
									ui.Text("这是 ClickArea（无默认按钮动画）"),
								),
							),
							func(ctx *ui.Context) {
								clickCount.Set(clickCount.Value() + 1)
							},
						),
					),
				)
			case "text_basic":
				return ui.Column(
					ui.Text("默认文本"),
					ui.Padding(ui.Insets{Top: 6}, ui.Text("大字号文本", ui.TextSize(20))),
					ui.Padding(ui.Insets{Top: 6}, ui.Text("强调色文本", ui.TextColor(th.Primary))),
				)
			case "button_basic":
				return ui.Row(
					ui.Button(
						ui.Text("点击 +1"),
						ui.OnClick(func(ctx *ui.Context) {
							buttonCount.Set(buttonCount.Value() + 1)
						}),
					),
					ui.Padding(
						ui.Insets{Left: 10, Top: 8},
						ui.Text(fmt.Sprintf("count = %d", buttonCount.Value())),
					),
				)
			case "textfield_basic":
				return ui.Column(
					ui.TextField(
						inputValue.Value(),
						ui.InputPlaceholder("请输入内容"),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							inputValue.Set(value)
						}),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text("当前输入: "+inputValue.Value(), ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
					),
				)
			case "checkbox_basic":
				return ui.Checkbox(
					"启用功能",
					checkboxValue.Value(),
					ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
						checkboxValue.Set(checked)
					}),
				)
			case "switch_basic":
				return ui.Row(
					ui.Switch(
						switchValue.Value(),
						ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
							switchValue.Set(checked)
						}),
					),
					ui.Padding(
						ui.Insets{Left: 10, Top: 5},
						ui.Text(fmt.Sprintf("状态: %v", switchValue.Value()), ui.TextSize(13)),
					),
				)
			case "slider_basic":
				return ui.Column(
					ui.Slider(
						sliderValue.Value(),
						ui.SliderMin(0),
						ui.SliderMax(100),
						ui.SliderOnChange(func(ctx *ui.Context, value float32) {
							sliderValue.Set(value)
						}),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text(fmt.Sprintf("value = %.1f", sliderValue.Value()), ui.TextSize(13)),
					),
				)
			case "image_basic":
				return ui.Row(
					ui.Image(
						ui.ImageSource{Path: "examples/assets/sample.png", Label: "sample.png"},
						ui.ImageWidth(150),
						ui.ImageHeight(90),
						ui.ImageFitMode(ui.ImageFitContain),
						ui.ImageRadius(8),
					),
					ui.Padding(
						ui.Insets{Left: 10},
						ui.Image(
							ui.ImageSource{Path: "examples/assets/sample.png", Label: "sample.png"},
							ui.ImageWidth(150),
							ui.ImageHeight(90),
							ui.ImageFitMode(ui.ImageFitCover),
							ui.ImageRadius(8),
						),
					),
				)
			case "icon_basic":
				return ui.Row(
					ui.Icon("H", ui.IconSize(20), ui.IconColor(ui.NRGBA(30, 136, 229, 255))),
					ui.Padding(ui.Insets{Left: 12}, ui.Icon("S", ui.IconSize(20), ui.IconColor(ui.NRGBA(67, 160, 71, 255)))),
					ui.Padding(ui.Insets{Left: 12}, ui.Icon("G", ui.IconSize(20), ui.IconColor(ui.NRGBA(245, 124, 0, 255)))),
				)
			case "card_basic":
				return ui.Card(
					ui.Column(
						ui.Text("Card 卡片", ui.TextSize(15)),
						ui.Padding(
							ui.Insets{Top: 6},
							ui.Text("点击卡片会增加计数。", ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
						),
						ui.Padding(
							ui.Insets{Top: 8},
							ui.Text(fmt.Sprintf("点击次数: %d", buttonCount.Value()), ui.TextSize(13)),
						),
					),
					ui.CardOnClick(func(ctx *ui.Context) {
						buttonCount.Set(buttonCount.Value() + 1)
					}),
				)
			case "radio_group_basic":
				return ui.RadioGroup(
					radioValue.Value(),
					[]ui.RadioItem{
						{Label: "布局", Value: "layout"},
						{Label: "输入", Value: "input"},
						{Label: "反馈", Value: "feedback"},
					},
					ui.RadioGroupOnChange(func(ctx *ui.Context, value string) {
						radioValue.Set(value)
					}),
				)
			case "select_basic":
				return ui.Select(
					selectValue.Value(),
					[]ui.SelectOptionItem[string]{
						{Label: "低优先级", Value: "low"},
						{Label: "中优先级", Value: "medium"},
						{Label: "高优先级", Value: "high"},
					},
					ui.SelectPlaceholder[string]("请选择优先级"),
					ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
						selectValue.Set(value)
					}),
				)
			case "progress_bar_basic":
				return ui.Column(
					ui.Slider(
						sliderValue.Value(),
						ui.SliderMin(0),
						ui.SliderMax(100),
						ui.SliderOnChange(func(ctx *ui.Context, value float32) {
							sliderValue.Set(value)
						}),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.ProgressBar(
							sliderValue.Value(),
							ui.ProgressMin(0),
							ui.ProgressMax(100),
							ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
							ui.ProgressFillColor(th.Primary),
						),
					),
				)
			case "circular_progress_basic":
				return ui.Column(
					ui.Slider(
						sliderValue.Value(),
						ui.SliderMin(0),
						ui.SliderMax(100),
						ui.SliderOnChange(func(ctx *ui.Context, value float32) {
							sliderValue.Set(value)
						}),
					),
					ui.Padding(
						ui.Insets{Top: 12},
						ui.CircularProgress(
							sliderValue.Value(),
							ui.ProgressMin(0),
							ui.ProgressMax(100),
							ui.ProgressSize(80),
							ui.ProgressThickness(8),
							ui.ProgressFillColor(th.Primary),
							ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
						),
					),
				)
			case "tabs_basic":
				return ui.Column(
					ui.Tabs(
						tabValue.Value(),
						[]ui.TabItem{
							{Key: "overview", Label: "Overview"},
							{Key: "api", Label: "API"},
							{Key: "example", Label: "Example"},
						},
						ui.TabsOnChange(func(ctx *ui.Context, key string) {
							tabValue.Set(key)
						}),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text("当前标签: "+tabValue.Value(), ui.TextSize(13)),
					),
				)
			case "dialog_basic":
				return ui.Stack(
					ui.FillWidth(
						ui.Column(
							ui.Button(
								ui.Text("打开对话框"),
								ui.OnClick(func(ctx *ui.Context) {
									dialogOpen.Set(true)
								}),
							),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Text("Dialog 演示：支持遮罩关闭、确认、取消。", ui.TextSize(13)),
							),
						),
					),
					ui.Dialog(
						dialogOpen.Value(),
						ui.Text("这是一个组件文档中的 Dialog 示例。"),
						ui.DialogTitle("Dialog 示例"),
						ui.DialogWidth(320),
						ui.DialogMaskClosable(true),
						ui.DialogOnOpenChange(func(ctx *ui.Context, open bool) {
							dialogOpen.Set(open)
						}),
						ui.DialogOnCancel(func(ctx *ui.Context) {
							dialogOpen.Set(false)
						}),
						ui.DialogOnConfirm(func(ctx *ui.Context) {
							dialogOpen.Set(false)
						}),
					),
				)
			case "toast_basic":
				var layers []ui.Widget
				layers = append(layers,
					ui.FillWidth(
						ui.Button(
							ui.Text("显示 Toast"),
							ui.OnClick(func(ctx *ui.Context) {
								toastMessage.Set("这是一条 Toast 消息")
							}),
						),
					),
				)
				if toastMessage.Value() != "" {
					layers = append(layers,
						ui.Toast(
							toastMessage.Value(),
							ui.ToastTypeOf(ui.ToastSuccess),
							ui.ToastPositionOf(ui.ToastBottom),
							ui.ToastDuration(1600*time.Millisecond),
							ui.ToastOnClose(func(ctx *ui.Context) {
								toastMessage.Set("")
							}),
						),
					)
				}
				return ui.Stack(layers...)
			case "scroll_view_basic":
				lines := make([]ui.Widget, 0, 24)
				for i := 1; i <= 24; i++ {
					lines = append(lines,
						ui.Padding(
							ui.Insets{Bottom: 6},
							ui.Container(
								ui.Style{Background: ui.NRGBA(241, 245, 249, 255), Padding: ui.All(8), Radius: 6},
								ui.Text(fmt.Sprintf("滚动项 %02d", i), ui.TextSize(13)),
							),
						),
					)
				}
				return ui.FixedHeight(
					180,
					ui.ScrollView(
						ui.Column(lines...),
						ui.ScrollVertical(true),
					),
				)
			case "list_view_basic":
				return ui.FixedHeight(
					200,
					ui.ListView(
						80,
						func(ctx *ui.Context, index int) ui.Widget {
							return ui.Container(
								ui.Style{
									Background: rowColor(index),
									Padding:    ui.Symmetric(8, 10),
									Radius:     6,
								},
								ui.Text(fmt.Sprintf("List Item #%d", index), ui.TextSize(13)),
							)
						},
						ui.ListItemSpacing(6),
						ui.ListOnReachEnd(func(ctx *ui.Context) {
							listReachEndCount.Set(listReachEndCount.Value() + 1)
						}),
					),
				)
			case "grid_basic":
				cells := make([]ui.Widget, 0, 9)
				for i := 1; i <= 9; i++ {
					cells = append(cells,
						ui.Container(
							ui.Style{
								Background: ui.NRGBA(227, 242, 253, 255),
								Padding:    ui.All(10),
								Radius:     6,
							},
							ui.Center(ui.Text(fmt.Sprintf("Cell %d", i), ui.TextSize(12))),
						),
					)
				}
				return ui.Grid(
					3,
					cells...,
				)
			case "app_bar_basic":
				return ui.Column(
					ui.AppBar(
						ui.Text("文档示例 AppBar", ui.TextSize(14)),
						ui.AppBarActions(
							ui.Button(
								ui.Text("Action"),
								ui.ButtonPadding(ui.Symmetric(4, 8)),
								ui.OnClick(func(ctx *ui.Context) {
									appbarActionCount.Set(appbarActionCount.Value() + 1)
								}),
							),
						),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text(fmt.Sprintf("Action 点击次数: %d", appbarActionCount.Value()), ui.TextSize(13)),
					),
				)
			case "bottom_navigation_basic":
				return ui.FixedHeight(
					180,
					ui.Column(
						ui.Expanded(
							ui.Center(
								ui.Text("当前页面: "+bottomNavValue.Value(), ui.TextSize(14)),
							),
						),
						ui.BottomNavigation(
							bottomNavValue.Value(),
							[]ui.NavItem{
								{Key: "home", Label: "首页", Icon: ui.Text("H", ui.TextSize(12))},
								{Key: "docs", Label: "文档", Icon: ui.Text("D", ui.TextSize(12))},
								{Key: "profile", Label: "我的", Icon: ui.Text("P", ui.TextSize(12))},
							},
							ui.BottomNavAlignmentOf(ui.BottomNavAlignSpaceEvenly),
							ui.BottomNavOnChange(func(ctx *ui.Context, key string) {
								bottomNavValue.Set(key)
							}),
						),
					),
				)
			case "hooks_lifecycle_basic":
				hookScope := ctx.Scope("docs-hooks-demo")
				ui.UseMount(hookScope, func() func() {
					appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Demo mount")
					return func() {
						appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Demo unmount")
					}
				})
				ui.UseEffectWithDeps(hookScope, []any{hookDemoCount.Value()}, func() func() {
					appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, fmt.Sprintf("count changed -> %d", hookDemoCount.Value()))
					return nil
				})

				content := []ui.Widget{
					ui.Text(fmt.Sprintf("count = %d", hookDemoCount.Value())),
					ui.Row(
						ui.Padding(ui.All(4), ui.Button(ui.Text("+1"), ui.OnClick(func(ctx *ui.Context) {
							hookDemoCount.Set(hookDemoCount.Value() + 1)
						}))),
						ui.Padding(ui.All(4), ui.Button(ui.Text("切换子组件"), ui.OnClick(func(ctx *ui.Context) {
							hookDemoShowChild.Set(!hookDemoShowChild.Value())
						}))),
					),
				}

				if hookDemoShowChild.Value() {
					childScope := hookScope.Scope("child")
					ui.UseLifecycle(childScope, func() {
						appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Child mount")
					}, func() {
						appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Child unmount")
					})
					content = append(content,
						ui.Container(
							ui.Style{
								Background: ui.NRGBA(226, 232, 240, 255),
								Padding:    ui.All(8),
								Radius:     6,
							},
							ui.Text("子组件已挂载"),
						),
					)
				}

				logItems := hookDemoLogs.Value()
				if len(logItems) == 0 {
					content = append(content, ui.Text("(暂无日志)", ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255))))
				} else {
					for _, item := range logItems {
						content = append(content, ui.Text(item, ui.TextSize(12), ui.TextColor(ui.NRGBA(51, 65, 85, 255))))
					}
				}

				return ui.Column(content...)
			default:
				return ui.Text("该文档未配置可执行示例。")
			}
		}

		menuEntries := buildMenuEntries(filteredDocs)
		leftMenuItems := make([]ui.Widget, 0, len(menuEntries)+1)
		for idx := range menuEntries {
			entry := menuEntries[idx]
			if entry.IsCategory {
				leftMenuItems = append(leftMenuItems,
					ui.Padding(
						ui.Insets{Top: 10, Bottom: 4},
						ui.Text(entry.Category, ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255))),
					),
				)
				continue
			}
			doc := entry.Doc
			if doc == nil {
				continue
			}

			selected := currentDoc != nil && doc.Meta.ID == currentDoc.Meta.ID
			bg := ui.NRGBA(0, 0, 0, 0)
			textColor := th.TextColor
			if selected {
				bg = ui.NRGBA(226, 232, 240, 255)
				textColor = th.Primary
			}

			leftMenuItems = append(leftMenuItems,
				ui.Padding(
					ui.Insets{Bottom: 6},
					ui.FillWidth(
						ui.Button(
							ui.FillWidth(
								ui.Container(
									ui.Style{
										Background: bg,
										Padding:    ui.Symmetric(8, 10),
										Radius:     6,
									},
									ui.Text(doc.Meta.Title, ui.TextSize(13), ui.TextColor(textColor)),
								),
							),
							ui.ButtonBackground(ui.NRGBA(0, 0, 0, 0)),
							ui.ButtonPadding(ui.All(0)),
							ui.OnClick(func(ctx *ui.Context) {
								selectedDocID.Set(doc.Meta.ID)
							}),
						),
					),
				),
			)
		}

		docCountText := fmt.Sprintf("已加载 %d 个控件文档（%s）", len(docs), map[string]string{
			"online": "在线",
			"local":  "本地",
		}[docsSource])
		if onlineLoading {
			docCountText = "本地文档不可用，正在异步加载在线文档..."
		}

		leftPanel := ui.FixedWidth(
			300,
			ui.Container(
				ui.Style{
					Background: ui.NRGBA(248, 250, 252, 255),
					Padding:    ui.All(12),
				},
				ui.Column(
					ui.Text("FluxUI 控件文档", ui.TextSize(18)),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text(
							docCountText,
							ui.TextSize(12),
							ui.TextColor(ui.NRGBA(100, 116, 139, 255)),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.TextField(
							searchKeyword.Value(),
							ui.InputPlaceholder("搜索控件 / 分类"),
							ui.InputOnChange(func(ctx *ui.Context, value string) {
								searchKeyword.Set(value)
							}),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Divider(ui.DividerColor(ui.NRGBA(203, 213, 225, 255))),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Expanded(
							ui.ScrollView(
								ui.Column(leftMenuItems...),
								ui.ScrollVertical(true),
							),
						),
					),
					func() ui.Widget {
						if onlineLoading {
							msg := "本地文档读取失败，正在加载 GitHub 在线文档..."
							if loadErr != nil {
								msg += " 原因: " + loadErr.Error()
							}
							return ui.Padding(
								ui.Insets{Top: 8},
								ui.Text(
									msg,
									ui.TextSize(11),
									ui.TextColor(ui.NRGBA(180, 83, 9, 255)),
								),
							)
						}
						if loadErr == nil {
							return ui.Spacer(0, 0)
						}
						return ui.Padding(
							ui.Insets{Top: 8},
							ui.Text(
								"文档加载警告: "+loadErr.Error(),
								ui.TextSize(11),
								ui.TextColor(ui.NRGBA(185, 28, 28, 255)),
							),
						)
					}(),
				),
			),
		)

		rightPanelContent := []ui.Widget{}
		if currentDoc != nil {
			rightPanelContent = append(rightPanelContent,
				ui.Text(currentDoc.Meta.Title, ui.TextSize(26)),
			)
			rightPanelContent = append(rightPanelContent,
				ui.Padding(
					ui.Insets{Top: 6},
					ui.Text(
						fmt.Sprintf("组件ID: %s  |  分类: %s", currentDoc.Meta.ID, currentDoc.Meta.Category),
						ui.TextSize(12),
						ui.TextColor(ui.NRGBA(100, 116, 139, 255)),
					),
				),
			)

			if strings.TrimSpace(currentDoc.Meta.Summary) != "" {
				rightPanelContent = append(rightPanelContent,
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Text(currentDoc.Meta.Summary, ui.TextSize(13), ui.TextColor(ui.NRGBA(51, 65, 85, 255))),
					),
				)
			}

			rightPanelContent = append(rightPanelContent,
				ui.Padding(
					ui.Insets{Top: 16},
					ui.Text("组件示例", ui.TextSize(17)),
				),
			)
			rightPanelContent = append(rightPanelContent,
				ui.Padding(
					ui.Insets{Top: 8},
					ui.Container(
						ui.Style{
							Background: ui.NRGBA(248, 250, 252, 255),
							Padding:    ui.All(12),
							Radius:     10,
						},
						ui.FixedHeight(
							230,
							ui.Fill(buildDemo(currentDoc)),
						),
					),
				),
			)

			if len(currentDoc.Meta.APIs) > 0 {
				apiWidgets := make([]ui.Widget, 0, len(currentDoc.Meta.APIs))
				for i := range currentDoc.Meta.APIs {
					apiWidgets = append(apiWidgets,
						ui.Padding(
							ui.Insets{Bottom: 6},
							ui.Container(
								ui.Style{
									Background: ui.NRGBA(241, 245, 249, 255),
									Padding:    ui.Symmetric(6, 8),
									Radius:     6,
								},
								ui.Text(currentDoc.Meta.APIs[i], ui.TextSize(12), ui.TextColor(ui.NRGBA(30, 41, 59, 255))),
							),
						),
					)
				}
				rightPanelContent = append(rightPanelContent,
					ui.Padding(
						ui.Insets{Top: 14},
						ui.Text("API 索引", ui.TextSize(17)),
					),
				)
				rightPanelContent = append(rightPanelContent,
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Column(apiWidgets...),
					),
				)
			}

			rightPanelContent = append(rightPanelContent,
				ui.Padding(
					ui.Insets{Top: 14},
					ui.Text("文档正文", ui.TextSize(17)),
				),
			)
			rightPanelContent = append(rightPanelContent,
				ui.Padding(
					ui.Insets{Top: 8},
					ui.Column(renderMarkdownWidgets(currentDoc.Content)...),
				),
			)
		}

		rightPanel := ui.Expanded(
			ui.Container(
				ui.Style{
					Background: th.Surface,
					Padding:    ui.All(16),
				},
				ui.ScrollView(
					ui.Column(rightPanelContent...),
					ui.ScrollVertical(true),
				),
			),
		)

		return ui.Container(
			ui.Style{Background: th.Surface},
			ui.Row(
				leftPanel,
				rightPanel,
			),
		)
	}, ui.Title("FluxUI Docs Browser"), ui.Size(1360, 880))
}

func buildMenuEntries(docs []widgetDoc) []menuEntry {
	entries := make([]menuEntry, 0, len(docs)+8)
	lastCategory := ""
	for i := range docs {
		doc := &docs[i]
		category := doc.Meta.Category
		if category == "" {
			category = "未分类"
		}
		if category != lastCategory {
			entries = append(entries, menuEntry{
				IsCategory: true,
				Category:   category,
			})
			lastCategory = category
		}
		entries = append(entries, menuEntry{
			Doc: doc,
		})
	}
	return entries
}

func filterDocs(docs []widgetDoc, keyword string) []widgetDoc {
	keyword = strings.TrimSpace(strings.ToLower(keyword))
	if keyword == "" {
		return docs
	}

	out := make([]widgetDoc, 0, len(docs))
	for i := range docs {
		doc := docs[i]
		haystack := strings.ToLower(doc.Meta.ID + " " + doc.Meta.Title + " " + doc.Meta.Category + " " + doc.Meta.Summary)
		if strings.Contains(haystack, keyword) {
			out = append(out, doc)
		}
	}
	return out
}

func findDocByID(docs []widgetDoc, id string) *widgetDoc {
	for i := range docs {
		if docs[i].Meta.ID == id {
			return &docs[i]
		}
	}
	return nil
}

func renderMarkdownWidgets(content string) []ui.Widget {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	widgets := make([]ui.Widget, 0, len(lines)+8)
	inCode := false
	codeLines := make([]string, 0, 12)

	flushCode := func() {
		if len(codeLines) == 0 {
			return
		}
		text := strings.Join(codeLines, "\n")
		widgets = append(widgets,
			ui.Container(
				ui.Style{
					Background: ui.NRGBA(15, 23, 42, 255),
					Padding:    ui.All(10),
					Radius:     8,
				},
				ui.Text(text, ui.TextSize(12), ui.TextColor(ui.NRGBA(226, 232, 240, 255))),
			),
		)
		widgets = append(widgets, ui.VSpacer(10))
		codeLines = codeLines[:0]
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				inCode = false
				flushCode()
			} else {
				inCode = true
				codeLines = codeLines[:0]
			}
			continue
		}

		if inCode {
			codeLines = append(codeLines, line)
			continue
		}

		if trimmed == "" {
			widgets = append(widgets, ui.VSpacer(8))
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "### "):
			widgets = append(widgets, ui.Text(strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")), ui.TextSize(16)))
		case strings.HasPrefix(trimmed, "## "):
			widgets = append(widgets, ui.Text(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")), ui.TextSize(19)))
		case strings.HasPrefix(trimmed, "# "):
			widgets = append(widgets, ui.Text(strings.TrimSpace(strings.TrimPrefix(trimmed, "# ")), ui.TextSize(23)))
		case strings.HasPrefix(trimmed, "- "):
			widgets = append(widgets, ui.Text("• "+strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), ui.TextSize(13)))
		default:
			widgets = append(widgets, ui.Text(line, ui.TextSize(13), ui.TextColor(ui.NRGBA(51, 65, 85, 255))))
		}
	}

	if inCode {
		flushCode()
	}
	return widgets
}

func loadWidgetDocs() ([]widgetDoc, error) {
	dir, err := resolveDocsWidgetsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	docs := make([]widgetDoc, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}

		doc, parseErr := parseWidgetDoc(path, string(data))
		if parseErr != nil {
			continue
		}
		docs = append(docs, doc)
	}

	if len(docs) == 0 {
		return nil, errors.New("docs/widgets 下没有可解析的组件文档")
	}

	sort.Slice(docs, func(i, j int) bool {
		a := docs[i].Meta
		b := docs[j].Meta
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return a.Title < b.Title
	})

	return docs, nil
}

func loadWidgetDocsFromGitHub() ([]widgetDoc, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	entries, err := fetchGitHubDocsEntries(client)
	if err != nil {
		return nil, err
	}

	docs := make([]widgetDoc, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != "file" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name), ".md") {
			continue
		}

		url := strings.TrimSpace(entry.DownloadURL)
		if url == "" {
			url = githubWidgetsRawURL + entry.Name
		}

		text, err := fetchHTTPText(client, url)
		if err != nil {
			continue
		}
		doc, err := parseWidgetDoc(url, text)
		if err != nil {
			continue
		}
		docs = append(docs, doc)
	}

	if len(docs) == 0 {
		return nil, errors.New("GitHub docs/widgets 下没有可解析的组件文档")
	}

	sort.Slice(docs, func(i, j int) bool {
		a := docs[i].Meta
		b := docs[j].Meta
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return a.Title < b.Title
	})

	return docs, nil
}

func fetchGitHubDocsEntries(client *http.Client) ([]githubContentEntry, error) {
	body, err := fetchHTTPBytes(client, githubWidgetsAPIURL, true)
	if err != nil {
		return nil, err
	}
	var entries []githubContentEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("解析 GitHub 文档目录失败: %w", err)
	}
	return entries, nil
}

func fetchHTTPText(client *http.Client, url string) (string, error) {
	data, err := fetchHTTPBytes(client, url, false)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fetchHTTPBytes(client *http.Client, url string, api bool) ([]byte, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FluxUI-DocsBrowser")
	if api {
		req.Header.Set("Accept", "application/vnd.github+json")
	} else {
		req.Header.Set("Accept", "text/plain, */*")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(data))
		if len(snippet) > 160 {
			snippet = snippet[:160] + "..."
		}
		if snippet != "" {
			return nil, fmt.Errorf("请求 %s 失败: HTTP %d, %s", url, resp.StatusCode, snippet)
		}
		return nil, fmt.Errorf("请求 %s 失败: HTTP %d", url, resp.StatusCode)
	}
	return data, nil
}

func buildOnlineLoadingDoc(localErr error) widgetDoc {
	lines := []string{
		"# 正在加载在线文档",
		"",
		"本地 `docs/widgets` 不可用，正在从 GitHub 拉取在线 Markdown 文档。",
		"",
		"在线来源：",
		"- https://github.com/xiaowumin-mark/FluxUI/tree/main/docs/widgets",
	}
	if localErr != nil {
		lines = append(lines, "", "本地错误："+localErr.Error())
	}
	return widgetDoc{
		Meta: docMeta{
			ID:       "loading_online_docs",
			Title:    "在线文档加载中",
			Category: "系统",
			Order:    1,
			Summary:  "本地文档不可用，正在异步加载在线文档。",
			Example:  docDemo{ID: "fallback"},
		},
		Content: strings.Join(lines, "\n"),
		Path:    githubWidgetsAPIURL,
	}
}

func buildOnlineLoadFailedDoc(localErr, onlineErr error) widgetDoc {
	lines := []string{
		"# 文档加载失败",
		"",
		"本地与在线文档都未能加载，请检查：",
		"- 本地 `docs/widgets` 目录是否存在且可读",
		"- 网络连接是否可访问 GitHub",
		"- 文档文件是否包含 `fluxui-doc-meta` 元数据块",
	}
	if localErr != nil {
		lines = append(lines, "", "本地错误："+localErr.Error())
	}
	if onlineErr != nil {
		lines = append(lines, "在线错误："+onlineErr.Error())
	}
	return widgetDoc{
		Meta: docMeta{
			ID:       "load_error",
			Title:    "文档加载失败",
			Category: "系统",
			Order:    1,
			Summary:  "本地与在线文档均不可用。",
			Example:  docDemo{ID: "fallback"},
		},
		Content: strings.Join(lines, "\n"),
		Path:    githubWidgetsAPIURL,
	}
}

func resolveDocsWidgetsDir() (string, error) {
	candidates := make([]string, 0, 12)
	candidates = append(candidates,
		filepath.Join("docs", "widgets"),
		filepath.Join("..", "docs", "widgets"),
		filepath.Join("..", "..", "docs", "widgets"),
	)

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "docs", "widgets"),
			filepath.Join(cwd, "..", "docs", "widgets"),
			filepath.Join(cwd, "..", "..", "docs", "widgets"),
		)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs", "widgets"),
			filepath.Join(exeDir, "..", "docs", "widgets"),
			filepath.Join(exeDir, "..", "..", "docs", "widgets"),
			filepath.Join(exeDir, "..", "..", "..", "docs", "widgets"),
		)
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		info, err := os.Stat(cleaned)
		if err != nil {
			continue
		}
		if info.IsDir() {
			return cleaned, nil
		}
	}
	return "", errors.New("未找到 docs/widgets 文档目录")
}

func parseWidgetDoc(path string, raw string) (widgetDoc, error) {
	content := strings.TrimPrefix(raw, "\uFEFF")
	startMarker := "<!-- fluxui-doc-meta"
	start := strings.Index(content, startMarker)
	if start < 0 {
		return widgetDoc{}, fmt.Errorf("文档 %s 缺少 fluxui-doc-meta", filepath.Base(path))
	}

	rest := content[start+len(startMarker):]
	endRel := strings.Index(rest, "-->")
	if endRel < 0 {
		return widgetDoc{}, fmt.Errorf("文档 %s 的元数据注释未闭合", filepath.Base(path))
	}

	metaText := strings.TrimSpace(rest[:endRel])
	var meta docMeta
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		return widgetDoc{}, fmt.Errorf("文档 %s 的元数据 JSON 解析失败: %w", filepath.Base(path), err)
	}

	if meta.ID == "" {
		meta.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if meta.Title == "" {
		meta.Title = meta.ID
	}
	if meta.Category == "" {
		meta.Category = "未分类"
	}
	if meta.Order == 0 {
		meta.Order = 9999
	}
	if meta.Example.ID == "" {
		meta.Example.ID = meta.ID
	}

	body := strings.TrimSpace(content[:start] + rest[endRel+3:])
	return widgetDoc{
		Meta:    meta,
		Content: body,
		Path:    path,
	}, nil
}

func rowColor(index int) color.NRGBA {
	if index%2 == 0 {
		return ui.NRGBA(241, 245, 249, 255)
	}
	return ui.NRGBA(226, 232, 240, 255)
}

func appendDemoLog(getLogs func() []string, setLogs func([]string), message string) {
	if getLogs == nil || setLogs == nil {
		return
	}
	items := append([]string{}, getLogs()...)
	items = append(items, fmt.Sprintf("%s  %s", time.Now().Format("15:04:05"), message))
	if len(items) > 8 {
		items = items[len(items)-8:]
	}
	setLogs(items)
}
