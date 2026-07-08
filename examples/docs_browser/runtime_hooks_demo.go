package main

import (
	"fmt"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

var docsRuntimeScopeKey = ui.NewContextKey("root")

type docsRuntimeTask struct {
	ID    string
	Label string
}

var docsRuntimeTasks = []docsRuntimeTask{
	{ID: "metadata", Label: "解析元数据"},
	{ID: "preview", Label: "渲染预览"},
	{ID: "copy", Label: "复制代码"},
}

func buildDocsHooksLifecycleDemo(demoCtx *ui.Context, count docsIntState, showChild docsBoolState, logs docsStringSliceState) ui.Element {
	th := ui.UseTheme(demoCtx)
	renderCount := ui.UseRef(demoCtx, 0)
	memoRuns := ui.UseRef(demoCtx, 0)
	alias := ui.UseState(demoCtx, "docs-browser")
	async := ui.UseAsync[string](demoCtx)
	renderCount.Current++

	doubled := ui.UseMemo(demoCtx, []any{count.Value()}, func() int {
		memoRuns.Current++
		return count.Value() * 2
	})
	increment := ui.UseCallback(demoCtx, []any{count.Value()}, func(ctx *ui.Context) {
		count.Set(count.Value() + 1)
	})

	ui.UseEffect(demoCtx, func() func() {
		return nil
	})
	ui.UseMount(demoCtx, func() func() {
		appendDemoLog(logs.Value, logs.Set, "Demo 挂载")
		return func() {
			appendDemoLog(logs.Value, logs.Set, "Demo 卸载")
		}
	})
	ui.UseEffectWithDeps(demoCtx, []any{count.Value()}, func() func() {
		appendDemoLog(logs.Value, logs.Set, fmt.Sprintf("计数变更 -> %d", count.Value()))
		return nil
	})
	ui.UseInterval(demoCtx, 30*time.Second, func() {
		appendDemoLog(logs.Value, logs.Set, "定时器滴答")
	})

	return ui.ColumnElement(
		ui.TextElement("React 风格运行时", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
		ui.VSpacerElement(8),
		docsRuntimeCountersCard(count, doubled, renderCount.Current, memoRuns.Current, increment, th),
		ui.VSpacerElement(10),
		docsRuntimeAsyncCard(async, th),
		ui.VSpacerElement(10),
		docsRuntimeInputCard(alias, th),
		ui.VSpacerElement(10),
		docsRuntimeProviderCard(alias.Value(), count.Value(), th),
		ui.VSpacerElement(10),
		docsRuntimeKeyedList(th),
		ui.VSpacerElement(10),
		docsRuntimeLifecycleCard(showChild, logs, th),
	)
}

func docsRuntimeCountersCard(
	count docsIntState,
	doubled int,
	renderCount int,
	memoRuns int,
	increment func(*ui.Context),
	th *ui.Theme,
) ui.Element {
	return docsRuntimeSection(th, ui.ColumnElement(
		ui.RowElement(
			ui.TextElement("UseState + UseMemo + UseRef + UseCallback", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
			ui.ExpandedElement(ui.SpacerElement(0, 0)),
			ui.ButtonElement(
				ui.TextElement("+1"),
				ui.ButtonPadding(ui.Symmetric(5, 10)),
				ui.OnClick(increment),
			),
		),
		ui.VSpacerElement(8),
		ui.RowElement(
			docsRuntimeStat("计数", fmt.Sprintf("%d", count.Value()), th),
			ui.HSpacerElement(8),
			docsRuntimeStat("memo 加倍", fmt.Sprintf("%d", doubled), th),
			ui.HSpacerElement(8),
			docsRuntimeStat("渲染次数", fmt.Sprintf("%d", renderCount), th),
			ui.HSpacerElement(8),
			docsRuntimeStat("memo 运行次数", fmt.Sprintf("%d", memoRuns), th),
		),
	))
}

func docsRuntimeAsyncCard(async *ui.AsyncHandle[string], th *ui.Theme) ui.Element {
	status := "空闲"
	data := ""
	if async != nil {
		status = asyncStatusLabel(async.Status())
		data = async.Data()
		if err := async.Error(); err != nil {
			data = err.Error()
		}
	}
	return docsRuntimeSection(th, ui.ColumnElement(
		ui.RowElement(
			ui.TextElement("UseAsync + UseInterval", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
			ui.ExpandedElement(ui.SpacerElement(0, 0)),
			ui.TextButtonElement(
				ui.TextElement("运行异步"),
				ui.ButtonPadding(ui.Symmetric(5, 10)),
				ui.OnClick(func(ctx *ui.Context) {
					if async == nil {
						return
					}
					async.Run(func() (string, error) {
						time.Sleep(250 * time.Millisecond)
						return "元数据已加载", nil
					})
				}),
			),
			ui.HSpacerElement(6),
			ui.TextButtonElement(
				ui.TextElement("重置"),
				ui.ButtonPadding(ui.Symmetric(5, 10)),
				ui.OnClick(func(ctx *ui.Context) {
					if async != nil {
						async.Reset()
					}
				}),
			),
		),
		ui.VSpacerElement(8),
		ui.RowElement(
			docsRuntimeStat("状态", status, th),
			ui.HSpacerElement(8),
			docsRuntimeStat("数据", dataOrDash(data), th),
		),
		ui.VSpacerElement(6),
		ui.TextElement("UseInterval 在组件保持挂载时追加生命周期日志滴答。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
	))
}

func asyncStatusLabel(status ui.AsyncStatus) string {
	switch status {
	case ui.AsyncLoading:
		return "加载中"
	case ui.AsyncSuccess:
		return "成功"
	case ui.AsyncError:
		return "错误"
	default:
		return "空闲"
	}
}

func dataOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func docsRuntimeInputCard(alias docsStringState, th *ui.Theme) ui.Element {
	return docsRuntimeSection(th, ui.ColumnElement(
		ui.TextElement("受控 TextFieldElement", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
		ui.VSpacerElement(8),
		ui.FixedWidthElement(
			320,
			ui.TextFieldElement(
				alias.Value(),
				ui.InputPlaceholder("作用域名称"),
				ui.InputSingleLine(true),
				ui.InputOnChange(func(ctx *ui.Context, value string) {
					alias.Set(value)
				}),
			),
		),
	))
}

func docsRuntimeProviderCard(alias string, count int, th *ui.Theme) ui.Element {
	scope := fmt.Sprintf("%s:%d", alias, count)
	return docsRuntimeSection(th,
		ui.Provider[string](
			docsRuntimeScopeKey,
			scope,
			ui.ComponentElement(func(ctx *ui.Context) ui.Element {
				current := ui.UseContext(ctx, docsRuntimeScopeKey)
				font := ui.DefaultFontSpec().WithWeight(ui.FontWeightSemiBold)
				return ui.ColumnElement(
					ui.TextElement("Provider + UseContext + Fragment", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
					ui.VSpacerElement(8),
					ui.WithFontElement(
						font,
						ui.TextElement("作用域 = "+current, ui.TextSize(13), ui.TextColor(th.Colors.Primary)),
					),
					ui.VSpacerElement(6),
					ui.Fragment(
						ui.TextElement("Fragment 子元素 A", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
						ui.TextElement("Fragment 子元素 B", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
					),
					ui.VSpacerElement(6),
					ui.FromWidget(ui.Text("FromWidget 桥接旧版 Widget 代码", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
				)
			}),
		),
	)
}

func docsRuntimeKeyedList(th *ui.Theme) ui.Element {
	return docsRuntimeSection(th, ui.ColumnElement(
		ui.TextElement("Key + ComponentElement 在 ListViewElement 中", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
		ui.VSpacerElement(8),
		ui.FixedHeightElement(
			118,
			ui.ListViewElement(
				len(docsRuntimeTasks),
				func(ctx *ui.Context, index int) ui.Element {
					task := docsRuntimeTasks[index]
					return ui.Key(task.ID, ui.ComponentElement(func(itemCtx *ui.Context) ui.Element {
						selected := ui.UseState(itemCtx, false)
						label := task.Label
						if selected.Value() {
							label += " 已选"
						}
						return ui.CardElement(
							ui.RowElement(
								ui.TextElement(label, ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
								ui.ExpandedElement(ui.SpacerElement(0, 0)),
								ui.TextElement(task.ID, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
							),
							ui.CardOnClick(func(ctx *ui.Context) {
								selected.Set(!selected.Value())
							}),
						)
					}))
				},
				ui.ListItemSpacing(6),
			),
		),
	))
}

func docsRuntimeLifecycleCard(showChild docsBoolState, logs docsStringSliceState, th *ui.Theme) ui.Element {
	content := []ui.Element{
		ui.RowElement(
			ui.TextElement("UseMount + UseEffectWithDeps + UseLifecycle", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
			ui.ExpandedElement(ui.SpacerElement(0, 0)),
			ui.TextButtonElement(
				ui.TextElement("切换子组件"),
				ui.ButtonPadding(ui.Symmetric(5, 10)),
				ui.OnClick(func(ctx *ui.Context) {
					showChild.Set(!showChild.Value())
				}),
			),
		),
		ui.VSpacerElement(8),
	}
	if showChild.Value() {
		content = append(content,
			ui.Key("lifecycle-child", ui.ComponentElement(func(childCtx *ui.Context) ui.Element {
				ui.UseLifecycle(childCtx, func() {
					appendDemoLog(logs.Value, logs.Set, "子组件挂载")
				}, func() {
					appendDemoLog(logs.Value, logs.Set, "子组件卸载")
				})
				return ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SecondaryContainer).WithPad(ui.All(8)).WithRad(8),
					ui.TextElement("子组件已挂载", ui.TextSize(12), ui.TextColor(th.Colors.OnSecondaryContainer)),
				)
			})),
			ui.VSpacerElement(8),
		)
	}
	content = append(content, docsRuntimeLogLines(logs.Value(), th)...)
	return docsRuntimeSection(th, ui.ColumnElement(content...))
}

func docsRuntimeLogLines(logs []string, th *ui.Theme) []ui.Element {
	if len(logs) == 0 {
		return []ui.Element{
			ui.TextElement("暂无生命周期事件。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		}
	}
	out := make([]ui.Element, 0, len(logs))
	for _, item := range logs {
		out = append(out,
			ui.PaddingElement(
				ui.Insets{Bottom: 4},
				ui.TextElement(item, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		)
	}
	return out
}

func docsRuntimeSection(th *ui.Theme, child ui.Element) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerLow).
			WithPad(ui.All(12)).
			WithRad(10).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		child,
	)
}

func docsRuntimeStat(label string, value string, th *ui.Theme) ui.Element {
	return ui.ExpandedElement(
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(8)).WithRad(8),
			ui.ColumnElement(
				ui.TextElement(label, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
				ui.VSpacerElement(4),
				ui.TextElement(value, ui.TextSize(15), ui.TextColor(th.Colors.OnSurface)),
			),
		),
	)
}
