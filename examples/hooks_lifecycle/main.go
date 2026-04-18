package main

import (
	"fmt"
	"time"

	"fluxui/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		count := ui.State[int](ctx)
		showChild := ui.State[bool](ctx)
		events := ui.State[[]string](ctx)
		scrollRefState := ui.State[*ui.ScrollRef](ctx)
		if scrollRefState.Value() == nil {
			scrollRefState.Set(ui.NewScrollRef())
		}
		scrollRef := scrollRefState.Value()

		// 挂载时初始化一次
		ui.UseMount(ctx, func() func() {
			appendLog(events.Value, events.Set, "App mount")
			return func() {
				appendLog(events.Value, events.Set, "App unmount")
			}
		})

		// 依赖变化时触发
		ui.UseEffectWithDeps(ctx, []any{count.Value()}, func() func() {
			appendLog(events.Value, events.Set, fmt.Sprintf("count changed -> %d", count.Value()))
			return nil
		})

		body := []ui.Widget{
			ui.Text("Hooks 与生命周期示例", ui.TextSize(24)),
			ui.Text(fmt.Sprintf("count = %d", count.Value()), ui.TextSize(16)),
			ui.Row(
				ui.Padding(ui.All(4), ui.Button(ui.Text("+1"), ui.OnClick(func(ctx *ui.Context) {
					count.Set(count.Value() + 1)
				}))),
				ui.Padding(ui.All(4), ui.Button(ui.Text("-1"), ui.OnClick(func(ctx *ui.Context) {
					count.Set(count.Value() - 1)
				}))),
				ui.Padding(ui.All(4), ui.Button(ui.Text("切换子组件"), ui.OnClick(func(ctx *ui.Context) {
					showChild.Set(!showChild.Value())
				}))),
				ui.Padding(ui.All(4), ui.Button(ui.Text("批量写入日志"), ui.OnClick(func(ctx *ui.Context) {
					for i := 0; i < 40; i++ {
						appendLog(events.Value, events.Set, fmt.Sprintf("bulk log #%d", i+1))
					}
				}))),
				ui.Padding(ui.All(4), ui.Button(ui.Text("滚动到底部"), ui.OnClick(func(ctx *ui.Context) {
					if scrollRef != nil {
						scrollRef.ScrollToBottom()
					}
				}))),
				ui.Padding(ui.All(4), ui.Button(ui.Text("滚动到顶部"), ui.OnClick(func(ctx *ui.Context) {
					if scrollRef != nil {
						scrollRef.ScrollToTop()
					}
				}))),
			),
		}

		if showChild.Value() {
			body = append(body, ui.Padding(ui.All(8), childPanel(ctx, events.Value, events.Set)))
		}

		body = append(body,
			ui.Divider(),
			ui.Text("生命周期日志（可滚动）", ui.TextSize(16)),
			ui.Expanded(
				ui.ScrollView(
					ui.Column(buildLogWidgets(events.Value())...),
					ui.ScrollAutoToEndKey(len(events.Value())),
					ui.ScrollAttachRef(scrollRef),
				),
			),
		)

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(16),
			},
			ui.Column(body...),
		)
	}, ui.Title("FluxUI Hooks Lifecycle"), ui.Size(700, 560))
}

func childPanel(ctx *ui.Context, getLogs func() []string, setLogs func([]string)) ui.Widget {
	scoped := ctx.Scope("child-panel")

	ui.UseLifecycle(scoped,
		func() {
			appendLog(getLogs, setLogs, "Child mount")
		},
		func() {
			appendLog(getLogs, setLogs, "Child unmount")
		},
	)

	return ui.Container(
		ui.Style{
			Background: ui.NRGBA(33, 43, 63, 255),
			Padding:    ui.All(12),
			Radius:     8,
		},
		ui.Text("这是可切换的子组件，切换时会触发 mount/unmount", ui.TextColor(ui.NRGBA(232, 239, 255, 255))),
	)
}

func appendLog(getLogs func() []string, setLogs func([]string), message string) {
	if getLogs == nil || setLogs == nil {
		return
	}
	items := append([]string{}, getLogs()...)
	items = append(items, fmt.Sprintf("%s  %s", time.Now().Format("15:04:05"), message))
	if len(items) > 200 {
		items = items[len(items)-200:]
	}
	setLogs(items)
}

func buildLogWidgets(items []string) []ui.Widget {
	if len(items) == 0 {
		return []ui.Widget{ui.Text("(暂无)")}
	}

	out := make([]ui.Widget, 0, len(items))
	for _, item := range items {
		out = append(out, ui.Padding(ui.All(2), ui.Text(item)))
	}
	return out
}
