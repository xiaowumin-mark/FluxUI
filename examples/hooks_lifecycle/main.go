package main

import (
	"fmt"
	"time"

	statepkg "github.com/xiaowumin-mark/FluxUI/state"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	count := ui.UseState(ctx, 0)
	showChild := ui.UseState(ctx, false)
	events := ui.UseState(ctx, []string{})
	scrollRef := ui.NewScrollRef()

	ui.UseMount(ctx, func() func() {
		appendLog(events, "App mount")
		return func() {
			appendLog(events, "App unmount")
		}
	})

	ui.UseEffectWithDeps(ctx, []any{count.Value()}, func() func() {
		appendLog(events, fmt.Sprintf("count changed -> %d", count.Value()))
		return nil
	})

	children := []ui.Element{
		ui.TextElement("Hooks 与生命周期示例", ui.TextSize(24)),
		ui.SpacerElement(0, 4),
		ui.TextElement(fmt.Sprintf("count = %d", count.Value()), ui.TextSize(16)),
		ui.RowElement(
			ui.PaddingElement(ui.All(4), ui.ButtonElement(
				ui.TextElement("+1"),
				ui.OnClick(func(ctx *ui.Context) { count.Set(count.Value() + 1) }),
			)),
			ui.PaddingElement(ui.All(4), ui.ButtonElement(
				ui.TextElement("-1"),
				ui.OnClick(func(ctx *ui.Context) { count.Set(count.Value() - 1) }),
			)),
			ui.PaddingElement(ui.All(4), ui.ButtonElement(
				ui.TextElement("切换子组件"),
				ui.OnClick(func(ctx *ui.Context) { showChild.Set(!showChild.Value()) }),
			)),
			ui.PaddingElement(ui.All(4), ui.ButtonElement(
				ui.TextElement("批量写入日志"),
				ui.OnClick(func(ctx *ui.Context) {
					for i := 0; i < 40; i++ {
						appendLog(events, fmt.Sprintf("bulk log #%d", i+1))
					}
				}),
			)),
			ui.PaddingElement(ui.All(4), ui.ButtonElement(
				ui.TextElement("滚动到底部"),
				ui.OnClick(func(ctx *ui.Context) { scrollRef.ScrollToBottom() }),
			)),
			ui.PaddingElement(ui.All(4), ui.ButtonElement(
				ui.TextElement("滚动到顶部"),
				ui.OnClick(func(ctx *ui.Context) { scrollRef.ScrollToTop() }),
			)),
		),
	}

	if showChild.Value() {
		children = append(children, ui.PaddingElement(ui.All(8), ui.ComponentElement(ChildPanel(events))))
	}

	children = append(children,
		ui.DividerElement(),
		ui.TextElement("生命周期日志（可滚动）", ui.TextSize(16)),
		ui.ExpandedElement(
			ui.ScrollViewElement(
				ui.ColumnElement(buildLogElements(events.Value())...),
				ui.ScrollAutoToEndKey(len(events.Value())),
				ui.ScrollAttachRef(scrollRef),
			),
		),
	)

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(children...),
	)
}

func ChildPanel(events *statepkg.State[[]string]) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		ui.UseMount(ctx, func() func() {
			appendLog(events, "Child mount")
			return func() {
				appendLog(events, "Child unmount")
			}
		})

		return ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(33, 43, 63, 255)).WithPad(ui.All(12)).WithRad(8),
			ui.TextElement("这是可切换的子组件，切换时会触发 mount/unmount", ui.TextColor(ui.NRGBA(232, 239, 255, 255))),
		)
	}
}

func appendLog(events *statepkg.State[[]string], message string) {
	items := append([]string{}, events.Value()...)
	items = append(items, fmt.Sprintf("%s  %s", time.Now().Format("15:04:05"), message))
	if len(items) > 200 {
		items = items[len(items)-200:]
	}
	events.Set(items)
}

func buildLogElements(items []string) []ui.Element {
	if len(items) == 0 {
		return []ui.Element{ui.TextElement("(暂无)")}
	}
	out := make([]ui.Element, 0, len(items))
	for _, item := range items {
		out = append(out, ui.PaddingElement(ui.All(2), ui.TextElement(item)))
	}
	return out
}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Hooks Lifecycle"), ui.Size(700, 560))
}
