package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
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

type docsRuntimeState struct {
	Docs       []widgetDoc
	Source     string
	LoadErr    error
	OnlineErr  error
	Loading    bool
	ResultCh   <-chan remoteLoadResult
	ResultDone bool
}

type githubContentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

const (
	githubDocsAPIBase = "https://api.github.com/repos/xiaowumin-mark/FluxUI/contents/docs"
	githubDocsRawBase = "https://raw.githubusercontent.com/xiaowumin-mark/FluxUI/main/docs"
)

var docsSubdirs = []string{"widgets", "style", "theme", "guides"}

func main() {
	initialDocs, loadErr := loadWidgetDocs()
	runtimeState := &docsRuntimeState{
		Docs:    initialDocs,
		Source:  "local",
		LoadErr: loadErr,
	}
	if len(runtimeState.Docs) == 0 {
		resultCh := make(chan remoteLoadResult, 1)
		runtimeState.Source = "online"
		runtimeState.Loading = true
		runtimeState.ResultCh = resultCh
		runtimeState.Docs = []widgetDoc{buildOnlineLoadingDoc(loadErr)}
		go func() {
			remoteDocs, err := loadWidgetDocsFromGitHub()
			resultCh <- remoteLoadResult{Docs: remoteDocs, Err: err}
		}()
	}

	app := func(ctx *ui.Context) ui.Element {
		applyRemoteDocsResult(runtimeState)
		docs := runtimeState.Docs
		loadErr := runtimeState.LoadErr
		docsSource := runtimeState.Source
		onlineLoading := runtimeState.Loading

		th := ui.UseTheme(ctx)

		selectedDocID := ui.UseState(ctx, "")
		searchKeyword := ui.UseState(ctx, "")
		buttonCount := ui.UseState(ctx, 0)
		inputValue := ui.UseState(ctx, "FluxUI")
		checkboxValue := ui.UseState(ctx, true)
		switchValue := ui.UseState(ctx, true)
		sliderValue := ui.UseState(ctx, float32(40))
		styleDemoActive := ui.UseState(ctx, false)
		animationDemoActive := ui.UseState(ctx, false)
		themeDemoDark := ui.UseState(ctx, false)
		radioValue := ui.UseState(ctx, "layout")
		selectValue := ui.UseState(ctx, "medium")
		tabValue := ui.UseState(ctx, "overview")
		dialogOpen := ui.UseState(ctx, false)
		popupOpen := ui.UseState(ctx, false)
		toastMessage := ui.UseState(ctx, "")
		snackbarMessage := ui.UseState(ctx, "")
		snackbarSerial := ui.UseState(ctx, 0)
		snackbarActionCount := ui.UseState(ctx, 0)
		chipSelected := ui.UseState(ctx, true)
		searchBarValue := ui.UseState(ctx, "")
		bottomNavValue := ui.UseState(ctx, "home")
		menuOpen := ui.UseState(ctx, false)
		menuValue := ui.UseState(ctx, "copy")
		listItemSelected := ui.UseState(ctx, "inbox")
		iconButtonSelected := ui.UseState(ctx, true)
		fabCount := ui.UseState(ctx, 0)
		railValue := ui.UseState(ctx, "home")
		drawerValue := ui.UseState(ctx, "inbox")
		clickCount := ui.UseState(ctx, 0)
		appbarActionCount := ui.UseState(ctx, 0)
		listReachEndCount := ui.UseState(ctx, 0)
		hookDemoCount := ui.UseState(ctx, 0)
		hookDemoShowChild := ui.UseState(ctx, false)
		hookDemoLogs := ui.UseState(ctx, []string{})
		routerDemoAllowSettings := ui.UseState(ctx, true)
		routerDemoUserID := ui.UseState(ctx, "u1001")
		routerDemoLog := ui.UseState(ctx, "Router demo ready")

		if selectedDocID.Value() == "" && len(docs) > 0 {
			selectedDocID.Set(docs[0].Meta.ID)
		}

		filteredDocs := filterDocs(docs, searchKeyword.Value())
		currentDoc := findDocByID(docs, selectedDocID.Value())
		if currentDoc == nil && len(docs) > 0 {
			selectedDocID.Set(docs[0].Meta.ID)
			currentDoc = &docs[0]
		}

		animationDemo := func(demoCtx *ui.Context) ui.Element {
			const (
				trackWidth  = float32(156)
				squareSize  = float32(28)
				trackHeight = float32(36)
			)
			target := float32(0)
			if animationDemoActive.Value() {
				target = trackWidth - squareSize
			}

			type curveDemo struct {
				name   string
				short  string
				color  color.NRGBA
				easing ui.Easing
			}

			curves := []curveDemo{
				{name: "EaseOut", short: "EO", color: ui.NRGBA(59, 130, 246, 255), easing: ui.EaseOut},
				{name: "EaseInOut", short: "IO", color: ui.NRGBA(34, 197, 94, 255), easing: ui.EaseInOut},
				{name: "OutBack", short: "BK", color: ui.NRGBA(251, 146, 60, 255), easing: ui.EaseOutBack},
				{name: "Elastic", short: "EL", color: ui.NRGBA(168, 85, 247, 255), easing: easeOutElastic},
				{name: "Bounce", short: "BO", color: ui.NRGBA(239, 68, 68, 255), easing: ui.EaseOutBounce},
				{name: "Linear", short: "LN", color: ui.NRGBA(100, 116, 139, 255), easing: ui.Linear},
			}

			curveCard := func(curve curveDemo, index int) ui.Element {
				return ui.Key(curve.name, ui.ComponentElement(func(squareCtx *ui.Context) ui.Element {
					hovered := ui.UseState(squareCtx, false)
					x := ui.UseAnimatedValue(squareCtx, target, 640*time.Millisecond, curve.easing)
					scale := ui.UseAnimatedValue(squareCtx, hoverScale(hovered.Value()), 150*time.Millisecond, ui.EaseOut)

					square := ui.ContainerDecorationElement(
						ui.Bg(curve.color).
							WithRad(8).
							Merge(ui.Shadow(0, 6, 14, color.NRGBA{R: curve.color.R, G: curve.color.G, B: curve.color.B, A: 56})).
							WithTransform(ui.Transform2D{ScaleX: scale, ScaleY: scale, Origin: ui.TransformCenter}),
						ui.FixedSizeElement(
							28,
							28,
							ui.CenterElement(ui.TextElement(curve.short, ui.TextSize(10), ui.TextColor(ui.NRGBA(255, 255, 255, 255)))),
						),
						ui.OnDecoHover(func(ctx *ui.Context, hovering bool) {
							hovered.Set(hovering)
						}),
					)

					track := ui.FixedSizeElement(
						trackWidth,
						trackHeight,
						ui.StackElement(
							ui.ContainerDecorationElement(
								ui.Bg(ui.NRGBA(226, 232, 240, 255)).WithRad(12),
								ui.SpacerElement(trackWidth, trackHeight),
							),
							ui.PaddingElement(
								ui.Insets{Left: x, Top: (trackHeight - squareSize) / 2},
								square,
							),
						),
					)

					return ui.ContainerDecorationElement(
						ui.Bg(ui.NRGBA(248, 250, 252, 255)).
							WithPad(ui.All(8)).
							WithRad(12).
							WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(226, 232, 240, 255)}),
						ui.ColumnElement(
							ui.RowElement(
								ui.FixedWidthElement(86, ui.TextElement(curve.name, ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255)))),
								ui.TextElement(fmt.Sprintf("#%d", index+1), ui.TextSize(11), ui.TextColor(ui.NRGBA(148, 163, 184, 255))),
							),
							ui.PaddingElement(ui.Insets{Top: 6}, track),
						),
					)
				}))
			}

			rows := make([]ui.Element, 0, 3)
			for i := 0; i < len(curves); i += 2 {
				rowItems := []ui.Element{
					ui.FixedWidthElement(245, curveCard(curves[i], i)),
				}
				if i+1 < len(curves) {
					rowItems = append(rowItems,
						ui.PaddingElement(ui.Insets{Left: 10}, ui.FixedWidthElement(245, curveCard(curves[i+1], i+1))),
					)
				}
				rows = append(rows, ui.PaddingElement(ui.Insets{Bottom: 8}, ui.RowElement(rowItems...)))
			}

			return ui.ColumnElement(
				ui.TextElement("Easing 曲线对比", ui.TextSize(15), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
				ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("点击触发位移动画；悬浮方块查看 transform hover 缩放。", ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255)))),
				ui.PaddingElement(ui.Insets{Top: 10}, ui.ColumnElement(rows...)),
				ui.PaddingElement(
					ui.Insets{Top: 8, Bottom: 8},
					ui.FixedWidthElement(
						160,
						ui.ButtonElement(
							ui.TextElement("触发动画"),
							ui.ButtonDecoration(
								ui.Bg(ui.NRGBA(15, 23, 42, 255)).
									WithPad(ui.Symmetric(8, 12)).
									WithRad(10).
									WithHover(ui.Bg(ui.NRGBA(30, 41, 59, 255))).
									WithPressed(ui.Bg(ui.NRGBA(51, 65, 85, 255))),
							),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								animationDemoActive.Set(!animationDemoActive.Value())
							}),
						),
					),
				),
			)
		}

		hooksLifecycleDemo := func(demoCtx *ui.Context) ui.Element {
			ui.UseMount(demoCtx, func() func() {
				appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Demo mount")
				return func() {
					appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Demo unmount")
				}
			})
			ui.UseEffectWithDeps(demoCtx, []any{hookDemoCount.Value()}, func() func() {
				appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, fmt.Sprintf("count changed -> %d", hookDemoCount.Value()))
				return nil
			})

			content := []ui.Element{
				ui.TextElement(fmt.Sprintf("count = %d", hookDemoCount.Value())),
				ui.RowElement(
					ui.PaddingElement(ui.All(4), ui.ButtonElement(ui.TextElement("+1"), ui.OnClick(func(ctx *ui.Context) {
						hookDemoCount.Set(hookDemoCount.Value() + 1)
					}))),
					ui.PaddingElement(ui.All(4), ui.ButtonElement(ui.TextElement("切换子组件"), ui.OnClick(func(ctx *ui.Context) {
						hookDemoShowChild.Set(!hookDemoShowChild.Value())
					}))),
				),
			}

			if hookDemoShowChild.Value() {
				content = append(content,
					ui.ComponentElement(func(childCtx *ui.Context) ui.Element {
						ui.UseLifecycle(childCtx, func() {
							appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Child mount")
						}, func() {
							appendDemoLog(hookDemoLogs.Value, hookDemoLogs.Set, "Child unmount")
						})
						return ui.ContainerElement(
							ui.Style{
								Background: ui.NRGBA(226, 232, 240, 255),
								Padding:    ui.All(8),
								Radius:     6,
							},
							ui.TextElement("子组件已挂载"),
						)
					}),
				)
			}

			logItems := hookDemoLogs.Value()
			if len(logItems) == 0 {
				content = append(content, ui.TextElement("(暂无日志)", ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255))))
			} else {
				for _, item := range logItems {
					content = append(content, ui.TextElement(item, ui.TextSize(12), ui.TextColor(ui.NRGBA(51, 65, 85, 255))))
				}
			}

			return ui.ColumnElement(content...)
		}

		buildDemo := func(doc *widgetDoc) ui.Element {
			if doc == nil {
				return ui.TextElement("暂无示例")
			}

			demoID := doc.Meta.Example.ID
			if demoID == "" {
				demoID = doc.Meta.ID
			}

			switch demoID {
			case "row_basic":
				return ui.RowElement(
					ui.ContainerElement(
						ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(10), Radius: 6},
						ui.TextElement("A", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
					),
					ui.PaddingElement(
						ui.Insets{Left: 8},
						ui.ContainerElement(
							ui.Style{Background: ui.NRGBA(67, 160, 71, 255), Padding: ui.All(10), Radius: 6},
							ui.TextElement("B", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						),
					),
					ui.PaddingElement(
						ui.Insets{Left: 8},
						ui.ContainerElement(
							ui.Style{Background: ui.NRGBA(245, 124, 0, 255), Padding: ui.All(10), Radius: 6},
							ui.TextElement("C", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						),
					),
				)
			case "column_basic":
				return ui.ColumnElement(
					ui.ContainerElement(
						ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(8), Radius: 6},
						ui.TextElement("第一行", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.ContainerElement(
							ui.Style{Background: ui.NRGBA(67, 160, 71, 255), Padding: ui.All(8), Radius: 6},
							ui.TextElement("第二行", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						),
					),
				)
			case "stack_basic":
				return ui.FixedHeightElement(
					120,
					ui.FillElement(
						ui.StackElement(
							ui.FillElement(
								ui.ContainerElement(
									ui.Style{Background: ui.NRGBA(234, 239, 245, 255), Radius: 8},
									ui.SpacerElement(0, 0),
								),
							),
							ui.PaddingElement(
								ui.Insets{Left: 12, Top: 12},
								ui.ContainerElement(
									ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(6), Radius: 6},
									ui.TextElement("Layer 1", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
								),
							),
							ui.CenterElement(
								ui.TextElement("Center Layer", ui.TextColor(ui.NRGBA(15, 23, 42, 255)), ui.TextSize(14)),
							),
						),
					),
				)
			case "center_basic":
				return ui.FixedHeightElement(
					120,
					ui.FillElement(
						ui.ContainerElement(
							ui.Style{Background: ui.NRGBA(240, 244, 248, 255), Radius: 8},
							ui.CenterElement(ui.TextElement("居中内容", ui.TextSize(14))),
						),
					),
				)
			case "container_basic":
				return ui.ContainerElement(
					ui.Style{
						Background: ui.NRGBA(30, 136, 229, 255),
						Padding:    ui.All(16),
						Radius:     10,
					},
					ui.TextElement("Container: 背景 + 内边距 + 圆角", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				)
			case "decoration_basic", "decoration_guide_basic":
				base := ui.Bg(ui.NRGBA(255, 255, 255, 255)).
					WithPad(ui.All(16)).
					WithMargin(ui.All(8)).
					WithRad(16).
					WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(203, 213, 225, 255)}).
					WithHover(ui.Bg(ui.NRGBA(239, 246, 255, 255))).
					WithPressed(ui.Bg(ui.NRGBA(219, 234, 254, 255)))
				if styleDemoActive.Value() {
					base = base.Merge(ui.Elevation(2)).WithBorder(ui.Border{Width: 1, Color: th.Primary})
				}
				return ui.PaddingElement(
					ui.Symmetric(8, 12),
					ui.ClickAreaElement(
						ui.FixedWidthElement(
							340,
							ui.ContainerDecorationElement(
								base,
								ui.ColumnElement(
									ui.TextElement("Decoration 卡片", ui.TextSize(16), ui.TextColor(th.TextColor)),
									ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("点击切换边框和阴影", ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255)))),
								),
							),
						),
						func(ctx *ui.Context) { styleDemoActive.Set(!styleDemoActive.Value()) },
					),
				)
			case "insets_basic":
				return ui.ContainerDecorationElement(
					ui.Bg(ui.NRGBA(239, 246, 255, 255)).WithRad(12),
					ui.PaddingElement(
						ui.Symmetric(10, 18),
						ui.TextElement("Symmetric(10, 18): 上下 10 / 左右 18", ui.TextColor(ui.NRGBA(30, 64, 175, 255))),
					),
				)
			case "border_basic":
				return ui.ContainerDecorationElement(
					ui.Bg(ui.NRGBA(255, 255, 255, 255)).
						WithPad(ui.All(14)).
						WithRad(12).
						WithBorder(ui.Border{Width: 2, Color: th.Primary}),
					ui.TextElement("Border: width + color", ui.TextColor(th.Primary)),
				)
			case "shadow_basic":
				return ui.PaddingElement(
					ui.Symmetric(8, 12),
					ui.FixedWidthElement(
						280,
						ui.ContainerDecorationElement(
							ui.Bg(ui.NRGBA(255, 255, 255, 255)).
								WithPad(ui.All(16)).
								WithMargin(ui.All(8)).
								WithRad(16).
								Merge(ui.Elevation(3)),
							ui.TextElement("Elevation(3) 阴影卡片"),
						),
					),
				)
			case "gradient_basic":
				return ui.ContainerDecorationElement(
					ui.LinearGrad(
						image.Point{X: 0, Y: 0},
						image.Point{X: 260, Y: 120},
						ui.NRGBA(14, 165, 233, 255),
						ui.NRGBA(34, 197, 94, 255),
					).WithPad(ui.All(18)).WithRad(18),
					ui.TextElement("LinearGradient 渐变背景", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(16)),
				)
			case "transform_basic":
				return ui.PaddingElement(
					ui.All(10),
					ui.ContainerDecorationElement(
						ui.Bg(ui.NRGBA(254, 243, 199, 255)).
							WithPad(ui.All(14)).
							WithRad(12).
							Merge(ui.Rotate(-4)),
						ui.TextElement("Rotate(-4)"),
					),
				)
			case "image_fill_basic":
				return ui.ContainerDecorationElement(
					ui.Bg(ui.NRGBA(15, 23, 42, 255)).WithPad(ui.All(16)).WithRad(14),
					ui.ColumnElement(
						ui.TextElement("ImageBg(img, ImageFillCover)", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("背景图建议在 effect/async 中加载后传入", ui.TextSize(12), ui.TextColor(ui.NRGBA(203, 213, 225, 255)))),
					),
				)
			case "theme_basic":
				localTheme := ui.NewTheme(ui.LightColors())
				label := "局部浅色主题"
				if themeDemoDark.Value() {
					localTheme = ui.NewTheme(ui.DarkColors())
					label = "局部深色主题"
				}
				return ui.ColumnElement(
					ui.ButtonElement(ui.TextElement("切换局部主题"), ui.OnClick(func(ctx *ui.Context) {
						themeDemoDark.Set(!themeDemoDark.Value())
					})),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.ThemeProviderElement(
							localTheme,
							ui.ContainerDecorationElement(
								ui.Bg(localTheme.Surface).WithPad(ui.All(14)).WithRad(12),
								ui.TextElement(label, ui.TextColor(localTheme.TextColor)),
							),
						),
					),
				)
			case "color_scheme_basic":
				return ui.RowElement(
					ui.ContainerDecorationElement(ui.Bg(th.Primary).WithPad(ui.All(10)).WithRad(8), ui.TextElement("Primary", ui.TextColor(th.TextOnPrimary))),
					ui.PaddingElement(ui.Insets{Left: 8}, ui.ContainerDecorationElement(ui.Bg(th.Colors.Warning).WithPad(ui.All(10)).WithRad(8), ui.TextElement("Warning", ui.TextColor(th.Colors.OnWarning)))),
					ui.PaddingElement(ui.Insets{Left: 8}, ui.ContainerDecorationElement(ui.Bg(th.Colors.Success).WithPad(ui.All(10)).WithRad(8), ui.TextElement("Success", ui.TextColor(th.Colors.OnSuccess)))),
				)
			case "getting_started_basic":
				return ui.ColumnElement(
					ui.TextElement("Hello FluxUI", ui.TextSize(18)),
					ui.PaddingElement(ui.Insets{Top: 8}, ui.ButtonElement(ui.TextElement("count +1"), ui.OnClick(func(ctx *ui.Context) {
						buttonCount.Set(buttonCount.Value() + 1)
					}))),
					ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement(fmt.Sprintf("count = %d", buttonCount.Value()), ui.TextColor(th.Primary))),
				)
			case "animation_basic":
				return ui.Key("docs-demo-animation", ui.ComponentElement(animationDemo))
			case "padding_basic":
				return ui.ContainerElement(
					ui.Style{
						Background: ui.NRGBA(229, 236, 246, 255),
						Radius:     8,
					},
					ui.PaddingElement(
						ui.All(16),
						ui.TextElement("Padding: 这里有 16dp 内边距"),
					),
				)
			case "spacer_basic":
				return ui.RowElement(
					ui.TextElement("左"),
					ui.HSpacerElement(24),
					ui.TextElement("右"),
					ui.HSpacerElement(24),
					ui.ColumnElement(
						ui.TextElement("上"),
						ui.VSpacerElement(8),
						ui.TextElement("下"),
					),
				)
			case "divider_basic":
				return ui.ColumnElement(
					ui.TextElement("第一段内容"),
					ui.DividerElement(ui.DividerThickness(1), ui.DividerColor(ui.NRGBA(176, 190, 197, 255)), ui.DividerMargin(ui.Insets{Top: 8, Bottom: 8})),
					ui.TextElement("第二段内容"),
				)
			case "sizing_basic":
				return ui.ColumnElement(
					ui.RowElement(
						ui.FixedWidthElement(
							110,
							ui.ContainerElement(
								ui.Style{Background: ui.NRGBA(3, 169, 244, 255), Padding: ui.All(8), Radius: 6},
								ui.TextElement("FixedWidth", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
							),
						),
						ui.PaddingElement(
							ui.Insets{Left: 8},
							ui.ExpandedElement(
								ui.ContainerElement(
									ui.Style{Background: ui.NRGBA(76, 175, 80, 255), Padding: ui.All(8), Radius: 6},
									ui.TextElement("Expanded / Fill", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
								),
							),
						),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.FixedHeightElement(
							48,
							ui.FillWidthElement(
								ui.ContainerElement(
									ui.Style{Background: ui.NRGBA(255, 152, 0, 255), Padding: ui.All(8), Radius: 6},
									ui.TextElement("FixedHeight", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
								),
							),
						),
					),
				)
			case "pressable_basic", "click_area_basic":
				return ui.ColumnElement(
					ui.TextElement(fmt.Sprintf("点击次数: %d", clickCount.Value())),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.PressableElement(
							ui.FillWidthElement(
								ui.ContainerElement(
									ui.Style{
										Background: ui.NRGBA(227, 242, 253, 255),
										Padding:    ui.All(14),
										Radius:     8,
									},
									ui.TextElement("这是 Pressable（无固定视觉样式）"),
								),
							),
							func(ctx *ui.Context) {
								clickCount.Set(clickCount.Value() + 1)
							},
						),
					),
				)
			case "text_basic":
				return ui.ColumnElement(
					ui.TextElement("默认文本"),
					ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("大字号文本", ui.TextSize(20))),
					ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("强调色文本", ui.TextColor(th.Primary))),
				)
			case "button_basic":
				return ui.RowElement(
					ui.ButtonElement(
						ui.TextElement("点击 +1"),
						ui.OnClick(func(ctx *ui.Context) {
							buttonCount.Set(buttonCount.Value() + 1)
						}),
					),
					ui.PaddingElement(
						ui.Insets{Left: 10, Top: 8},
						ui.TextElement(fmt.Sprintf("count = %d", buttonCount.Value())),
					),
				)
			case "textfield_basic":
				return ui.ColumnElement(
					ui.TextFieldElement(
						inputValue.Value(),
						ui.InputPlaceholder("请输入内容"),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							inputValue.Set(value)
						}),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement("当前输入: "+inputValue.Value(), ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
					),
				)
			case "checkbox_basic":
				return ui.CheckboxElement(
					"启用功能",
					checkboxValue.Value(),
					ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
						checkboxValue.Set(checked)
					}),
				)
			case "switch_basic":
				return ui.RowElement(
					ui.SwitchElement(
						switchValue.Value(),
						ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
							switchValue.Set(checked)
						}),
					),
					ui.PaddingElement(
						ui.Insets{Left: 10, Top: 5},
						ui.TextElement(fmt.Sprintf("状态: %v", switchValue.Value()), ui.TextSize(13)),
					),
				)
			case "slider_basic":
				return ui.ColumnElement(
					ui.SliderElement(
						sliderValue.Value(),
						ui.SliderMin(0),
						ui.SliderMax(100),
						ui.SliderOnChange(func(ctx *ui.Context, value float32) {
							sliderValue.Set(value)
						}),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement(fmt.Sprintf("value = %.1f", sliderValue.Value()), ui.TextSize(13)),
					),
				)
			case "image_basic":
				return ui.RowElement(
					ui.ImageElement(
						ui.ImageSource{Path: "examples/assets/sample.png", Label: "sample.png"},
						ui.ImageWidth(150),
						ui.ImageHeight(90),
						ui.ImageFitMode(ui.ImageFitContain),
						ui.ImageRadius(8),
					),
					ui.PaddingElement(
						ui.Insets{Left: 10},
						ui.ImageElement(
							ui.ImageSource{Path: "examples/assets/sample.png", Label: "sample.png"},
							ui.ImageWidth(150),
							ui.ImageHeight(90),
							ui.ImageFitMode(ui.ImageFitCover),
							ui.ImageRadius(8),
						),
					),
				)
			case "icon_basic":
				return ui.RowElement(
					ui.IconElement("H", ui.IconSize(20), ui.IconColor(ui.NRGBA(30, 136, 229, 255))),
					ui.PaddingElement(ui.Insets{Left: 12}, ui.IconElement("S", ui.IconSize(20), ui.IconColor(ui.NRGBA(67, 160, 71, 255)))),
					ui.PaddingElement(ui.Insets{Left: 12}, ui.IconElement("G", ui.IconSize(20), ui.IconColor(ui.NRGBA(245, 124, 0, 255)))),
				)
			case "card_basic":
				return ui.CardElement(
					ui.ColumnElement(
						ui.TextElement("Card 卡片", ui.TextSize(15)),
						ui.PaddingElement(
							ui.Insets{Top: 6},
							ui.TextElement("点击卡片会增加计数。", ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
						),
						ui.PaddingElement(
							ui.Insets{Top: 8},
							ui.TextElement(fmt.Sprintf("点击次数: %d", buttonCount.Value()), ui.TextSize(13)),
						),
					),
					ui.CardOnClick(func(ctx *ui.Context) {
						buttonCount.Set(buttonCount.Value() + 1)
					}),
				)
			case "radio_group_basic":
				return ui.RadioGroupElement(
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
				return ui.SelectElement(
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
			case "menu_basic":
				return ui.FixedWidthElement(
					260,
					ui.DropdownMenuElement(
						menuOpen.Value(),
						ui.ContainerDecorationElement(
							ui.Bg(th.Colors.Surface).WithPad(ui.Symmetric(10, 16)).WithRad(th.Shapes.ExtraSmall).WithBorder(ui.Border{Width: 1, Color: th.Colors.Outline}),
							ui.TextElement("Open menu", ui.TextColor(th.Colors.OnSurface)),
						),
						[]ui.MenuItem{
							{Key: "copy", Label: "Copy"},
							{Key: "share", Label: "Share"},
							{Key: "archive", Label: "Archive"},
							{Key: "delete", Label: "Delete", Disabled: true},
						},
						ui.DropdownMenuSelectedKey(menuValue.Value()),
						ui.DropdownMenuOnOpenChange(func(ctx *ui.Context, open bool) {
							menuOpen.Set(open)
						}),
						ui.DropdownMenuOnSelect(func(ctx *ui.Context, key string) {
							menuValue.Set(key)
							menuOpen.Set(false)
						}),
					),
				)
			case "list_item_basic":
				return ui.FixedWidthElement(
					360,
					ui.ColumnElement(
						ui.ListItemElementWithSlots(
							ui.TextElement("Inbox"),
							ui.TextElement("12 unread messages"),
							ui.IconElement("I"),
							ui.TextElement("12"),
							ui.ListItemSelected(listItemSelected.Value() == "inbox"),
							ui.ListItemOnClick(func(ctx *ui.Context) { listItemSelected.Set("inbox") }),
						),
						ui.ListItemElementWithSlots(
							ui.TextElement("Archive"),
							ui.TextElement("Older conversations"),
							ui.IconElement("A"),
							nil,
							ui.ListItemSelected(listItemSelected.Value() == "archive"),
							ui.ListItemOnClick(func(ctx *ui.Context) { listItemSelected.Set("archive") }),
						),
						ui.ListItemElementWithSlots(
							ui.TextElement("Disabled"),
							ui.TextElement("Unavailable item"),
							ui.IconElement("D"),
							nil,
							ui.ListItemDisabled(true),
						),
					),
				)
			case "icon_button_basic":
				return ui.RowElement(
					ui.PaddingElement(ui.Insets{Right: 10}, ui.IconButtonElement(ui.IconElement("S"), ui.IconButtonSelected(iconButtonSelected.Value()), ui.IconButtonOnClick(func(ctx *ui.Context) {
						iconButtonSelected.Set(!iconButtonSelected.Value())
					}))),
					ui.PaddingElement(ui.Insets{Right: 10}, ui.FilledIconButtonElement(ui.IconElement("F"), ui.IconButtonSelected(true))),
					ui.PaddingElement(ui.Insets{Right: 10}, ui.FilledTonalIconButtonElement(ui.IconElement("T"))),
					ui.PaddingElement(ui.Insets{Right: 10}, ui.OutlinedIconButtonElement(ui.IconElement("O"))),
					ui.OutlinedIconButtonElement(ui.IconElement("D"), ui.IconButtonDisabled(true)),
				)
			case "floating_action_button_basic":
				return ui.ColumnElement(
					ui.RowElement(
						ui.PaddingElement(ui.Insets{Right: 12}, ui.SmallFloatingActionButtonElement(ui.IconElement("+"), ui.FloatingActionButtonOnClick(func(ctx *ui.Context) {
							fabCount.Set(fabCount.Value() + 1)
						}))),
						ui.PaddingElement(ui.Insets{Right: 12}, ui.FloatingActionButtonElement(ui.IconElement("+"), ui.FloatingActionButtonOnClick(func(ctx *ui.Context) {
							fabCount.Set(fabCount.Value() + 1)
						}))),
						ui.PaddingElement(ui.Insets{Right: 12}, ui.LargeFloatingActionButtonElement(ui.IconElement("+"), ui.FloatingActionButtonOnClick(func(ctx *ui.Context) {
							fabCount.Set(fabCount.Value() + 1)
						}))),
						ui.ExtendedFloatingActionButtonElement(ui.IconElement("+"), ui.TextElement("Create"), ui.FloatingActionButtonOnClick(func(ctx *ui.Context) {
							fabCount.Set(fabCount.Value() + 1)
						})),
					),
					ui.PaddingElement(ui.Insets{Top: 12}, ui.TextElement(fmt.Sprintf("FAB clicks: %d", fabCount.Value()), ui.TextSize(13))),
				)
			case "progress_bar_basic":
				return ui.ColumnElement(
					ui.SliderElement(
						sliderValue.Value(),
						ui.SliderMin(0),
						ui.SliderMax(100),
						ui.SliderOnChange(func(ctx *ui.Context, value float32) {
							sliderValue.Set(value)
						}),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.ProgressBarElement(
							sliderValue.Value(),
							ui.ProgressMin(0),
							ui.ProgressMax(100),
							ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
							ui.ProgressFillColor(th.Primary),
						),
					),
				)
			case "circular_progress_basic":
				return ui.ColumnElement(
					ui.SliderElement(
						sliderValue.Value(),
						ui.SliderMin(0),
						ui.SliderMax(100),
						ui.SliderOnChange(func(ctx *ui.Context, value float32) {
							sliderValue.Set(value)
						}),
					),
					ui.PaddingElement(
						ui.Insets{Top: 12},
						ui.CircularProgressElement(
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
				return ui.ColumnElement(
					ui.TabsElement(
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
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement("当前标签: "+tabValue.Value(), ui.TextSize(13)),
					),
				)
			case "dialog_basic":
				return ui.StackElement(
					ui.FillWidthElement(
						ui.ColumnElement(
							ui.ButtonElement(
								ui.TextElement("打开对话框"),
								ui.OnClick(func(ctx *ui.Context) {
									dialogOpen.Set(true)
								}),
							),
							ui.PaddingElement(
								ui.Insets{Top: 8},
								ui.TextElement("Dialog 演示：支持遮罩关闭、确认、取消。", ui.TextSize(13)),
							),
						),
					),
					ui.DialogElement(
						dialogOpen.Value(),
						ui.TextElement("这是一个组件文档中的 Dialog 示例。"),
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
			case "popup_basic":
				return ui.StackElement(
					ui.FillWidthElement(
						ui.ColumnElement(
							ui.ButtonElement(
								ui.TextElement("打开 Popup"),
								ui.OnClick(func(ctx *ui.Context) {
									popupOpen.Set(true)
								}),
							),
							ui.PaddingElement(
								ui.Insets{Top: 8},
								ui.TextElement("Popup 演示：弹窗内容完全自定义，无预置标题和按钮。", ui.TextSize(13)),
							),
						),
					),
					ui.PopupElement(
						popupOpen.Value(),
						ui.ColumnElement(
							ui.TextElement("自定义弹窗内容", ui.TextSize(16)),
							ui.PaddingElement(
								ui.Insets{Top: 8},
								ui.TextElement("这里可以放置任意组件。", ui.TextSize(13)),
							),
							ui.PaddingElement(
								ui.Insets{Top: 12},
								ui.ButtonElement(
									ui.TextElement("关闭"),
									ui.OnClick(func(ctx *ui.Context) {
										popupOpen.Set(false)
									}),
								),
							),
						),
						ui.PopupWidth(300),
						ui.PopupPadding(ui.All(16)),
						ui.PopupRadius(12),
						ui.PopupMaskClosable(true),
						ui.PopupOnOpenChange(func(ctx *ui.Context, open bool) {
							popupOpen.Set(open)
						}),
					),
				)
			case "toast_basic":
				var layers []ui.Element
				layers = append(layers,
					ui.FillWidthElement(
						ui.ButtonElement(
							ui.TextElement("显示 Toast"),
							ui.OnClick(func(ctx *ui.Context) {
								toastMessage.Set("这是一条 Toast 消息")
							}),
						),
					),
				)
				if toastMessage.Value() != "" {
					layers = append(layers,
						ui.ToastElement(
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
				return ui.StackElement(layers...)
			case "snackbar_basic":
				layers := []ui.Element{
					ui.FixedHeightElement(
						160,
						ui.ContainerDecorationElement(
							ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(16)).WithRad(th.Shapes.Medium),
							ui.ColumnElement(
								ui.FilledButtonElement(
									ui.TextElement("Show snackbar"),
									ui.OnClick(func(ctx *ui.Context) {
										snackbarSerial.Set(snackbarSerial.Value() + 1)
										snackbarMessage.Set("Draft archived")
									}),
								),
								ui.PaddingElement(
									ui.Insets{Top: 10},
									ui.TextElement(fmt.Sprintf("Action clicks: %d", snackbarActionCount.Value()), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
								),
							),
						),
					),
				}
				if snackbarMessage.Value() != "" {
					layers = append(layers,
						ui.Key(
							fmt.Sprintf("snackbar-%d", snackbarSerial.Value()),
							ui.SnackbarElement(
								snackbarMessage.Value(),
								ui.SnackbarAction("Undo", func(ctx *ui.Context) {
									snackbarActionCount.Set(snackbarActionCount.Value() + 1)
									snackbarMessage.Set("")
								}),
								ui.ToastDuration(0),
							),
						),
					)
				}
				return ui.StackElement(layers...)
			case "tooltip_basic":
				return ui.TooltipElement(
					"Tooltip text",
					ui.FilledTonalButtonElement(ui.TextElement("Hover me")),
				)
			case "badge_basic":
				return ui.RowElement(
					ui.PaddingElement(
						ui.Insets{Right: 24},
						ui.BadgeElement(
							ui.IconButtonElement(ui.IconElement("M")),
							"3",
						),
					),
					ui.BadgeElement(
						ui.IconButtonElement(ui.IconElement("N")),
						"",
						ui.BadgeVisible(true),
					),
				)
			case "chip_basic":
				return ui.RowElement(
					ui.PaddingElement(
						ui.Insets{Right: 8},
						ui.AssistChipElement("Assist", ui.ChipLeading(ui.Icon("i", ui.IconSize(16)))),
					),
					ui.PaddingElement(
						ui.Insets{Right: 8},
						ui.FilterChipElement(
							"Filter",
							ui.ChipSelected(chipSelected.Value()),
							ui.ChipOnClick(func(ctx *ui.Context) {
								chipSelected.Set(!chipSelected.Value())
							}),
						),
					),
					ui.PaddingElement(
						ui.Insets{Right: 8},
						ui.InputChipElement("Input", ui.ChipTrailing(ui.Icon("x", ui.IconSize(14)))),
					),
					ui.SuggestionChipElement("Suggestion"),
				)
			case "search_bar_basic":
				return ui.FixedWidthElement(
					420,
					ui.ColumnElement(
						ui.SearchBarElement(
							searchBarValue.Value(),
							ui.SearchBarPlaceholder("Search docs"),
							ui.SearchBarLeading(ui.Icon("S", ui.IconSize(18))),
							ui.SearchBarOnChange(func(ctx *ui.Context, value string) {
								searchBarValue.Set(value)
							}),
						),
						ui.PaddingElement(
							ui.Insets{Top: 10},
							ui.TextElement("value = "+searchBarValue.Value(), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
						),
					),
				)
			case "progress_indicators_basic":
				return ui.FixedWidthElement(
					360,
					ui.ColumnElement(
						ui.SliderElement(
							sliderValue.Value(),
							ui.SliderMin(0),
							ui.SliderMax(100),
							ui.SliderOnChange(func(ctx *ui.Context, value float32) {
								sliderValue.Set(value)
							}),
						),
						ui.PaddingElement(
							ui.Insets{Top: 12},
							ui.LinearProgressIndicatorElement(
								sliderValue.Value(),
								ui.ProgressMin(0),
								ui.ProgressMax(100),
								ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
								ui.ProgressFillColor(th.Primary),
							),
						),
						ui.PaddingElement(
							ui.Insets{Top: 16},
							ui.CircularProgressIndicatorElement(
								sliderValue.Value(),
								ui.ProgressMin(0),
								ui.ProgressMax(100),
								ui.ProgressSize(72),
								ui.ProgressFillColor(th.Primary),
								ui.ProgressLabelVisible(true),
							),
						),
					),
				)
			case "scroll_view_basic":
				lines := make([]ui.Element, 0, 24)
				for i := 1; i <= 24; i++ {
					lines = append(lines,
						ui.PaddingElement(
							ui.Insets{Bottom: 6},
							ui.ContainerElement(
								ui.Style{Background: ui.NRGBA(241, 245, 249, 255), Padding: ui.All(8), Radius: 6},
								ui.TextElement(fmt.Sprintf("滚动项 %02d", i), ui.TextSize(13)),
							),
						),
					)
				}
				return ui.FixedHeightElement(
					180,
					ui.ScrollViewElement(
						ui.ColumnElement(lines...),
						ui.ScrollVertical(true),
					),
				)
			case "list_view_basic":
				return ui.FixedHeightElement(
					200,
					ui.ListViewElement(
						80,
						func(ctx *ui.Context, index int) ui.Element {
							return ui.ContainerElement(
								ui.Style{
									Background: rowColor(index),
									Padding:    ui.Symmetric(8, 10),
									Radius:     6,
								},
								ui.TextElement(fmt.Sprintf("List Item #%d", index), ui.TextSize(13)),
							)
						},
						ui.ListItemSpacing(6),
						ui.ListOnReachEnd(func(ctx *ui.Context) {
							listReachEndCount.Set(listReachEndCount.Value() + 1)
						}),
					),
				)
			case "grid_basic":
				cells := make([]ui.Element, 0, 9)
				for i := 1; i <= 9; i++ {
					cells = append(cells,
						ui.ContainerElement(
							ui.Style{
								Background: ui.NRGBA(227, 242, 253, 255),
								Padding:    ui.All(10),
								Radius:     6,
							},
							ui.CenterElement(ui.TextElement(fmt.Sprintf("Cell %d", i), ui.TextSize(12))),
						),
					)
				}
				return ui.GridElement(
					3,
					cells...,
				)
			case "app_bar_basic":
				return ui.ColumnElement(
					ui.AppBarElementWithSlots(
						ui.TextElement("文档示例 AppBar", ui.TextSize(14)),
						nil,
						[]ui.Element{
							ui.ButtonElement(
								ui.TextElement("Action"),
								ui.ButtonPadding(ui.Symmetric(4, 8)),
								ui.OnClick(func(ctx *ui.Context) {
									appbarActionCount.Set(appbarActionCount.Value() + 1)
								}),
							),
						},
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement(fmt.Sprintf("Action 点击次数: %d", appbarActionCount.Value()), ui.TextSize(13)),
					),
				)
			case "bottom_navigation_basic":
				return ui.FixedHeightElement(
					180,
					ui.ColumnElement(
						ui.ExpandedElement(
							ui.CenterElement(
								ui.TextElement("当前页面: "+bottomNavValue.Value(), ui.TextSize(14)),
							),
						),
						ui.BottomNavigationElement(
							bottomNavValue.Value(),
							[]ui.ElementNavItem{
								{Key: "home", Label: "首页", Icon: ui.TextElement("H", ui.TextSize(12))},
								{Key: "docs", Label: "文档", Icon: ui.TextElement("D", ui.TextSize(12))},
								{Key: "profile", Label: "我的", Icon: ui.TextElement("P", ui.TextSize(12))},
							},
							ui.BottomNavAlignmentOf(ui.BottomNavAlignSpaceEvenly),
							ui.BottomNavOnChange(func(ctx *ui.Context, key string) {
								bottomNavValue.Set(key)
							}),
						),
					),
				)
			case "navigation_rail_basic":
				return ui.FixedHeightElement(
					240,
					ui.RowElement(
						ui.NavigationRailElement(
							railValue.Value(),
							[]ui.ElementNavItem{
								{Key: "home", Label: "Home", Icon: ui.IconElement("H")},
								{Key: "search", Label: "Search", Icon: ui.IconElement("S")},
								{Key: "settings", Label: "Settings", Icon: ui.IconElement("G")},
							},
							ui.NavigationRailOnChange(func(ctx *ui.Context, key string) {
								railValue.Set(key)
							}),
						),
						ui.ExpandedElement(
							ui.CenterElement(ui.TextElement("Rail page: "+railValue.Value(), ui.TextSize(14))),
						),
					),
				)
			case "navigation_drawer_basic":
				return ui.FixedHeightElement(
					240,
					ui.RowElement(
						ui.NavigationDrawerElement(
							drawerValue.Value(),
							[]ui.ElementNavItem{
								{Key: "inbox", Label: "Inbox", Icon: ui.IconElement("I")},
								{Key: "sent", Label: "Sent", Icon: ui.IconElement("S")},
								{Key: "drafts", Label: "Drafts", Icon: ui.IconElement("D")},
							},
							ui.NavigationDrawerWidth(280),
							ui.NavigationDrawerOnChange(func(ctx *ui.Context, key string) {
								drawerValue.Set(key)
							}),
						),
						ui.ExpandedElement(
							ui.CenterElement(ui.TextElement("Drawer page: "+drawerValue.Value(), ui.TextSize(14))),
						),
					),
				)
			case "router_basic":
				routerUsers := []struct {
					ID   string
					Name string
				}{
					{ID: "u1001", Name: "Ava"},
					{ID: "u1002", Name: "Noah"},
					{ID: "u1003", Name: "Mia"},
				}
				findUserName := func(id string) string {
					for _, item := range routerUsers {
						if item.ID == id {
							return item.Name
						}
					}
					return "unknown"
				}

				routes := []ui.Route{
					{
						Path: "/",
						Builder: func(routeCtx *ui.Context) ui.Widget {
							return ui.RenderElement(ui.ColumnElement(
								ui.TextElement("Home", ui.TextSize(14)),
								ui.PaddingElement(
									ui.Insets{Top: 6},
									ui.TextFieldElement(
										routerDemoUserID.Value(),
										ui.InputPlaceholder("user id, e.g. u1002"),
										ui.InputOnChange(func(ctx *ui.Context, value string) {
											routerDemoUserID.Set(value)
										}),
									),
								),
								ui.PaddingElement(
									ui.Insets{Top: 6},
									ui.RowElement(
										ui.ButtonElement(
											ui.TextElement("Detail"),
											ui.ButtonPadding(ui.Symmetric(4, 8)),
											ui.OnClick(func(ctx *ui.Context) {
												id := strings.TrimSpace(routerDemoUserID.Value())
												if id == "" {
													routerDemoLog.Set("empty user id")
													return
												}
												ui.Navigate(ctx, "/user/"+id+"?tab=profile")
											}),
										),
										ui.PaddingElement(
											ui.Insets{Left: 6},
											ui.ButtonElement(
												ui.TextElement("Settings"),
												ui.ButtonPadding(ui.Symmetric(4, 8)),
												ui.OnClick(func(ctx *ui.Context) {
													ui.Navigate(ctx, "/settings")
												}),
											),
										),
										ui.PaddingElement(
											ui.Insets{Left: 6},
											ui.ButtonElement(
												ui.TextElement("404"),
												ui.ButtonPadding(ui.Symmetric(4, 8)),
												ui.OnClick(func(ctx *ui.Context) {
													ui.Navigate(ctx, "/not-found")
												}),
											),
										),
									),
								),
							))
						},
					},
					{
						Path: "/user/:id",
						Builder: func(routeCtx *ui.Context) ui.Widget {
							params := ui.UseParams(routeCtx)
							id := params.Path("id")
							tab := params.Query("tab")
							if tab == "" {
								tab = "overview"
							}
							return ui.RenderElement(ui.ColumnElement(
								ui.TextElement("User Detail", ui.TextSize(14)),
								ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("id: "+id)),
								ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("name: "+findUserName(id))),
								ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("tab: "+tab)),
								ui.PaddingElement(
									ui.Insets{Top: 8},
									ui.RowElement(
										ui.ButtonElement(
											ui.TextElement("Replace tab=activity"),
											ui.ButtonPadding(ui.Symmetric(4, 8)),
											ui.OnClick(func(ctx *ui.Context) {
												ui.NavigateReplace(ctx, "/user/"+id+"?tab=activity", ui.WithNavTransition(ui.TransitionFade))
											}),
										),
										ui.PaddingElement(
											ui.Insets{Left: 6},
											ui.ButtonElement(
												ui.TextElement("Back"),
												ui.ButtonPadding(ui.Symmetric(4, 8)),
												ui.OnClick(func(ctx *ui.Context) {
													ui.NavigateBack(ctx)
												}),
											),
										),
									),
								),
							))
						},
					},
					{
						Path: "/settings",
						Builder: func(routeCtx *ui.Context) ui.Widget {
							return ui.RenderElement(ui.ColumnElement(
								ui.TextElement("Settings", ui.TextSize(14)),
								ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("guard 通过后才可进入")),
								ui.PaddingElement(
									ui.Insets{Top: 8},
									ui.ButtonElement(
										ui.TextElement("Back"),
										ui.ButtonPadding(ui.Symmetric(4, 8)),
										ui.OnClick(func(ctx *ui.Context) {
											ui.NavigateBack(ctx)
										}),
									),
								),
							))
						},
					},
				}

				routerDemo := ui.Router(
					ctx,
					routes,
					ui.RouterTransition(ui.TransitionSlideLeft),
					ui.RouterTransitionDuration(220*time.Millisecond),
					ui.RouterBeforeEach(func(ctx *ui.Context, from, to string) bool {
						if to == "/settings" && !routerDemoAllowSettings.Value() {
							routerDemoLog.Set("guard blocked: " + from + " -> " + to)
							return false
						}
						routerDemoLog.Set("navigated: " + from + " -> " + to)
						return true
					}),
					ui.RouterNotFound(func(routeCtx *ui.Context) ui.Widget {
						return ui.RenderElement(ui.ColumnElement(
							ui.TextElement("404", ui.TextSize(16), ui.TextColor(ui.NRGBA(220, 38, 38, 255))),
							ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("path: "+ui.CurrentPath(routeCtx), ui.TextSize(12))),
							ui.PaddingElement(
								ui.Insets{Top: 8},
								ui.ButtonElement(
									ui.TextElement("Go Home"),
									ui.ButtonPadding(ui.Symmetric(4, 8)),
									ui.OnClick(func(ctx *ui.Context) {
										ui.NavigateReplace(ctx, "/")
									}),
								),
							),
						))
					}),
				)

				current := ui.CurrentPath(ctx)
				if current == "" {
					current = "/"
				}

				return ui.ColumnElement(
					ui.TextElement(fmt.Sprintf("path=%s | depth=%d", current, ui.StackDepth(ctx)), ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
					ui.PaddingElement(
						ui.Insets{Top: 6},
						ui.CheckboxElement(
							"allow settings",
							routerDemoAllowSettings.Value(),
							ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
								routerDemoAllowSettings.Set(checked)
							}),
						),
					),
					ui.PaddingElement(
						ui.Insets{Top: 6},
						ui.TextElement(routerDemoLog.Value(), ui.TextSize(12), ui.TextColor(ui.NRGBA(51, 65, 85, 255))),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.ExpandedElement(
							ui.ContainerDecorationElement(
								ui.Bg(ui.NRGBA(241, 245, 249, 255)).WithPad(ui.All(8)).WithRad(6),
								ui.FromWidget(routerDemo),
							),
						),
					),
				)
			case "hooks_lifecycle_basic":
				return ui.Key("docs-demo-hooks", ui.ComponentElement(hooksLifecycleDemo))
			default:
				return ui.TextElement("该文档未配置可执行示例。")
			}
		}

		menuEntries := buildMenuEntries(filteredDocs)
		leftMenuItems := make([]ui.Element, 0, len(menuEntries)+1)
		for idx := range menuEntries {
			entry := menuEntries[idx]
			if entry.IsCategory {
				leftMenuItems = append(leftMenuItems,
					ui.PaddingElement(
						ui.Insets{Top: 10, Bottom: 4},
						ui.TextElement(entry.Category, ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255))),
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
				ui.PaddingElement(
					ui.Insets{Bottom: 6},
					ui.FillWidthElement(
						ui.ButtonElement(
							ui.FillWidthElement(
								ui.ContainerElement(
									ui.Style{
										Background: bg,
										Padding:    ui.Symmetric(8, 10),
										Radius:     6,
									},
									ui.TextElement(doc.Meta.Title, ui.TextSize(13), ui.TextColor(textColor)),
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

		docCountText := fmt.Sprintf("已加载 %d 篇文档（%s）", len(docs), map[string]string{
			"online": "在线",
			"local":  "本地",
		}[docsSource])
		if onlineLoading {
			docCountText = "本地文档不可用，正在异步加载在线文档..."
		}

		leftPanel := ui.FixedWidthElement(
			300,
			ui.ContainerElement(
				ui.Style{
					Background: ui.NRGBA(248, 250, 252, 255),
					Padding:    ui.All(12),
				},
				ui.ColumnElement(
					ui.TextElement("FluxUI 文档", ui.TextSize(18)),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement(
							docCountText,
							ui.TextSize(12),
							ui.TextColor(ui.NRGBA(100, 116, 139, 255)),
						),
					),
					ui.PaddingElement(
						ui.Insets{Top: 10},
						ui.TextFieldElement(
							searchKeyword.Value(),
							ui.InputPlaceholder("搜索文档 / 分类"),
							ui.InputOnChange(func(ctx *ui.Context, value string) {
								searchKeyword.Set(value)
							}),
						),
					),
					ui.PaddingElement(
						ui.Insets{Top: 10},
						ui.DividerElement(ui.DividerColor(ui.NRGBA(203, 213, 225, 255))),
					),
					ui.PaddingElement(
						ui.Insets{Top: 10},
						ui.ExpandedElement(
							ui.ScrollViewElement(
								ui.ColumnElement(leftMenuItems...),
								ui.ScrollVertical(true),
							),
						),
					),
					func() ui.Element {
						if onlineLoading {
							msg := "本地文档读取失败，正在加载 GitHub 在线文档..."
							if loadErr != nil {
								msg += " 原因: " + loadErr.Error()
							}
							return ui.PaddingElement(
								ui.Insets{Top: 8},
								ui.TextElement(
									msg,
									ui.TextSize(11),
									ui.TextColor(ui.NRGBA(180, 83, 9, 255)),
								),
							)
						}
						if loadErr == nil {
							return ui.SpacerElement(0, 0)
						}
						return ui.PaddingElement(
							ui.Insets{Top: 8},
							ui.TextElement(
								"文档加载警告: "+loadErr.Error(),
								ui.TextSize(11),
								ui.TextColor(ui.NRGBA(185, 28, 28, 255)),
							),
						)
					}(),
				),
			),
		)

		rightPanelContent := []ui.Element{}
		if currentDoc != nil {
			demoHeight := float32(230)
			if currentDoc.Meta.Example.ID == "animation_basic" {
				demoHeight = 300
			} else if currentDoc.Meta.Example.ID == "router_basic" {
				demoHeight = 320
			} else if currentDoc.Meta.Example.ID == "navigation_rail_basic" || currentDoc.Meta.Example.ID == "navigation_drawer_basic" {
				demoHeight = 280
			}
			demoContent := buildDemo(currentDoc)
			demoViewport := ui.FillElement(demoContent)
			if currentDoc.Meta.Example.ID == "animation_basic" {
				demoViewport = ui.FillElement(ui.ScrollViewElement(demoContent, ui.ScrollVertical(true)))
			} else if shouldCenterDemo(currentDoc.Meta.Example.ID) {
				demoViewport = ui.CenterElement(demoContent)
			}

			rightPanelContent = append(rightPanelContent,
				ui.TextElement(currentDoc.Meta.Title, ui.TextSize(26)),
			)
			rightPanelContent = append(rightPanelContent,
				ui.PaddingElement(
					ui.Insets{Top: 6},
					ui.TextElement(
						fmt.Sprintf("组件ID: %s  |  分类: %s", currentDoc.Meta.ID, currentDoc.Meta.Category),
						ui.TextSize(12),
						ui.TextColor(ui.NRGBA(100, 116, 139, 255)),
					),
				),
			)

			if strings.TrimSpace(currentDoc.Meta.Summary) != "" {
				rightPanelContent = append(rightPanelContent,
					ui.PaddingElement(
						ui.Insets{Top: 10},
						ui.TextElement(currentDoc.Meta.Summary, ui.TextSize(13), ui.TextColor(ui.NRGBA(51, 65, 85, 255))),
					),
				)
			}

			rightPanelContent = append(rightPanelContent,
				ui.PaddingElement(
					ui.Insets{Top: 16},
					ui.TextElement("组件示例", ui.TextSize(17)),
				),
			)
			rightPanelContent = append(rightPanelContent,
				ui.PaddingElement(
					ui.Insets{Top: 8},
					ui.ContainerElement(
						ui.Style{
							Background: ui.NRGBA(248, 250, 252, 255),
							Padding:    ui.All(12),
							Radius:     10,
						},
						ui.FixedHeightElement(
							demoHeight,
							demoViewport,
						),
					),
				),
			)

			if len(currentDoc.Meta.APIs) > 0 {
				apiWidgets := make([]ui.Element, 0, len(currentDoc.Meta.APIs))
				for i := range currentDoc.Meta.APIs {
					apiWidgets = append(apiWidgets,
						ui.PaddingElement(
							ui.Insets{Bottom: 6},
							ui.ContainerElement(
								ui.Style{
									Background: ui.NRGBA(241, 245, 249, 255),
									Padding:    ui.Symmetric(6, 8),
									Radius:     6,
								},
								ui.TextElement(currentDoc.Meta.APIs[i], ui.TextSize(12), ui.TextColor(ui.NRGBA(30, 41, 59, 255))),
							),
						),
					)
				}
				rightPanelContent = append(rightPanelContent,
					ui.PaddingElement(
						ui.Insets{Top: 14},
						ui.TextElement("API 索引", ui.TextSize(17)),
					),
				)
				rightPanelContent = append(rightPanelContent,
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.ColumnElement(apiWidgets...),
					),
				)
			}

			rightPanelContent = append(rightPanelContent,
				ui.PaddingElement(
					ui.Insets{Top: 14},
					ui.TextElement("文档正文", ui.TextSize(17)),
				),
			)
			rightPanelContent = append(rightPanelContent,
				ui.PaddingElement(
					ui.Insets{Top: 8},
					ui.ColumnElement(renderMarkdownWidgets(currentDoc.Content)...),
				),
			)
		}

		rightPanel := ui.ExpandedElement(
			ui.ContainerElement(
				ui.Style{
					Background: th.Surface,
					Padding:    ui.All(16),
				},
				ui.ScrollViewElement(
					ui.ColumnElement(rightPanelContent...),
					ui.ScrollVertical(true),
				),
			),
		)

		return ui.ContainerElement(
			ui.Style{Background: th.Surface},
			ui.RowElement(
				leftPanel,
				rightPanel,
			),
		)

	}

	_ = ui.RunElement(app, ui.Title("FluxUI Docs Browser"), ui.Size(1360, 880))
}

func shouldCenterDemo(exampleID string) bool {
	switch exampleID {
	case "button_basic",
		"checkbox_basic",
		"switch_basic",
		"slider_basic",
		"menu_basic",
		"image_basic",
		"icon_basic",
		"icon_button_basic",
		"floating_action_button_basic",
		"card_basic",
		"list_item_basic",
		"tooltip_basic",
		"badge_basic",
		"chip_basic",
		"radio_group_basic",
		"progress_bar_basic",
		"circular_progress_basic",
		"progress_indicators_basic",
		"tabs_basic":
		return true
	default:
		return false
	}
}

func applyRemoteDocsResult(state *docsRuntimeState) {
	if state == nil || !state.Loading || state.ResultDone || state.ResultCh == nil {
		return
	}
	select {
	case result := <-state.ResultCh:
		state.ResultDone = true
		state.Loading = false
		state.OnlineErr = result.Err
		if len(result.Docs) > 0 {
			state.Docs = result.Docs
			state.LoadErr = nil
			return
		}
		state.Docs = []widgetDoc{buildOnlineLoadFailedDoc(state.LoadErr, state.OnlineErr)}
		if state.LoadErr != nil && state.OnlineErr != nil {
			state.LoadErr = fmt.Errorf("本地加载失败: %v；在线加载失败: %v", state.LoadErr, state.OnlineErr)
		} else if state.OnlineErr != nil {
			state.LoadErr = fmt.Errorf("在线加载失败: %w", state.OnlineErr)
		}
	default:
	}
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

func renderMarkdownWidgets(content string) []ui.Element {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	widgets := make([]ui.Element, 0, len(lines)+8)
	inCode := false
	codeLines := make([]string, 0, 12)

	flushCode := func() {
		if len(codeLines) == 0 {
			return
		}
		text := strings.Join(codeLines, "\n")
		widgets = append(widgets,
			ui.ContainerElement(
				ui.Style{
					Background: ui.NRGBA(15, 23, 42, 255),
					Padding:    ui.All(10),
					Radius:     8,
				},
				ui.TextElement(text, ui.TextSize(12), ui.TextColor(ui.NRGBA(226, 232, 240, 255))),
			),
		)
		widgets = append(widgets, ui.VSpacerElement(10))
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
			widgets = append(widgets, ui.VSpacerElement(8))
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "### "):
			widgets = append(widgets, ui.TextElement(strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")), ui.TextSize(16)))
		case strings.HasPrefix(trimmed, "## "):
			widgets = append(widgets, ui.TextElement(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")), ui.TextSize(19)))
		case strings.HasPrefix(trimmed, "# "):
			widgets = append(widgets, ui.TextElement(strings.TrimSpace(strings.TrimPrefix(trimmed, "# ")), ui.TextSize(23)))
		case strings.HasPrefix(trimmed, "- "):
			widgets = append(widgets, ui.TextElement("• "+strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), ui.TextSize(13)))
		default:
			widgets = append(widgets, ui.TextElement(line, ui.TextSize(13), ui.TextColor(ui.NRGBA(51, 65, 85, 255))))
		}
	}

	if inCode {
		flushCode()
	}
	return widgets
}

func loadWidgetDocs() ([]widgetDoc, error) {
	docsRoot, err := resolveDocsRootDir()
	if err != nil {
		return nil, err
	}

	docs := make([]widgetDoc, 0, 48)
	for _, subdir := range docsSubdirs {
		dir := filepath.Join(docsRoot, subdir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

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
	}

	if len(docs) == 0 {
		return nil, errors.New("docs 下没有可解析的 Markdown 文档")
	}

	sortDocs(docs)

	return docs, nil
}

func loadWidgetDocsFromGitHub() ([]widgetDoc, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	docs := make([]widgetDoc, 0, 48)
	var firstErr error
	for _, subdir := range docsSubdirs {
		entries, err := fetchGitHubDocsEntries(client, subdir)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		for _, entry := range entries {
			if entry.Type != "file" {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(entry.Name), ".md") {
				continue
			}

			url := strings.TrimSpace(entry.DownloadURL)
			if url == "" {
				url = githubDocsRawBase + "/" + subdir + "/" + entry.Name
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
	}

	if len(docs) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, errors.New("GitHub docs 下没有可解析的 Markdown 文档")
	}

	sortDocs(docs)

	return docs, nil
}

func sortDocs(docs []widgetDoc) {
	sort.Slice(docs, func(i, j int) bool {
		a := docs[i].Meta
		b := docs[j].Meta
		aRank := categoryRank(a.Category)
		bRank := categoryRank(b.Category)
		if aRank != bRank {
			return aRank < bRank
		}
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return a.Title < b.Title
	})
}

func categoryRank(category string) int {
	switch category {
	case "使用指南":
		return 10
	case "样式系统":
		return 20
	case "主题系统":
		return 30
	case "状态与副作用":
		return 40
	case "布局系统":
		return 50
	case "基础显示":
		return 60
	case "输入交互":
		return 70
	case "反馈组件":
		return 80
	case "导航组件":
		return 90
	case "系统":
		return 100
	default:
		return 999
	}
}

func fetchGitHubDocsEntries(client *http.Client, subdir string) ([]githubContentEntry, error) {
	body, err := fetchHTTPBytes(client, githubDocsAPIBase+"/"+subdir+"?ref=main", true)
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
		"本地 `docs` 文档目录不可用，正在从 GitHub 拉取在线 Markdown 文档。",
		"",
		"在线来源：",
		"- https://github.com/xiaowumin-mark/FluxUI/tree/main/docs",
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
		Path:    githubDocsAPIBase,
	}
}

func buildOnlineLoadFailedDoc(localErr, onlineErr error) widgetDoc {
	lines := []string{
		"# 文档加载失败",
		"",
		"本地与在线文档都未能加载，请检查：",
		"- 本地 `docs` 目录是否存在且可读",
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
		Path:    githubDocsAPIBase,
	}
}

func resolveDocsRootDir() (string, error) {
	candidates := make([]string, 0, 12)
	candidates = append(candidates,
		"docs",
		filepath.Join("..", "docs"),
		filepath.Join("..", "..", "docs"),
	)

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "docs"),
			filepath.Join(cwd, "..", "docs"),
			filepath.Join(cwd, "..", "..", "docs"),
		)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs"),
			filepath.Join(exeDir, "..", "docs"),
			filepath.Join(exeDir, "..", "..", "docs"),
			filepath.Join(exeDir, "..", "..", "..", "docs"),
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
	return "", errors.New("未找到 docs 文档目录")
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

func hoverScale(hovered bool) float32 {
	if hovered {
		return 1.18
	}
	return 1
}

func easeOutElastic(v float32) float32 {
	v = clamp01(v)
	if v == 0 || v == 1 {
		return v
	}
	return float32(math.Pow(2, -10*float64(v)))*float32(math.Sin(float64(v*10-0.75)*2*math.Pi/3)) + 1
}

func clamp01(v float32) float32 {
	if math.IsNaN(float64(v)) || v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
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
