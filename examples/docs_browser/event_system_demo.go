package main

import (
	"fmt"
	"image/color"
	"strings"
	"unicode"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsEventState[T any] interface {
	Value() T
	Set(T)
}

type docsEventPoint struct {
	X float32
	Y float32
}

type docsEventDropState struct {
	Active bool
	Text   string
}

func buildDocsEventSystemDemo(th *ui.Theme) ui.Element {
	return ui.Key("docs-event-system-demo", ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		if th == nil {
			th = ui.UseTheme(ctx)
		}
		logs := ui.UseState(ctx, []string{"就绪：移动、点击、右键、输入、快捷键、拖拽"})
		pos := ui.UseState(ctx, docsEventPoint{X: 136, Y: 72})
		dragging := ui.UseState(ctx, false)
		menuOpen := ui.UseState(ctx, false)
		menuAt := ui.UseState(ctx, docsEventPoint{X: 180, Y: 60})
		numberValue := ui.UseState(ctx, "128")
		shortcutCount := ui.UseState(ctx, 0)
		dropState := ui.UseState(ctx, docsEventDropState{Text: "将 text/json/custom 载荷拖放到这里"})

		addLog := func(message string) {
			next := append([]string{message}, logs.Value()...)
			if len(next) > 9 {
				next = next[:9]
			}
			logs.Set(next)
		}

		ui.OnEvent(ctx, ui.EventType("docs:event-system:synthetic"), func(ctx *ui.Context, ev *ui.Event) {
			addLog(fmt.Sprintf("捕获自定义事件 阶段=%s 目标=%s", docsEventPhaseLabel(ev.Phase), ctx.TreePath()))
		}, ui.Capture())
		ui.OnEvent(ctx, ui.EventType("docs:event-system:synthetic"), func(ctx *ui.Context, ev *ui.Event) {
			addLog(fmt.Sprintf("冒泡自定义事件 detail=%v 已阻止默认=%t", ev.Detail, ev.DefaultPrevented))
		})

		return ui.ScrollViewElement(
			ui.ColumnElement(
				ui.RowElement(
					ui.ExpandedElement(docsEventCanvasPanel(th, pos, dragging, menuOpen, menuAt, addLog)),
					ui.HSpacerElement(12),
					ui.ExpandedElement(docsEventKeyboardAndInputPanel(th, numberValue, shortcutCount, addLog)),
				),
				ui.VSpacerElement(12),
				ui.RowElement(
					ui.ExpandedElement(docsEventSyntheticPanel(th, addLog)),
					ui.HSpacerElement(12),
					ui.ExpandedElement(docsEventDragDropPanel(th, dropState, addLog)),
				),
				ui.VSpacerElement(12),
				docsEventLogPanel(th, logs.Value()),
			),
			ui.ScrollVertical(true),
		)
	}))
}

func docsEventCanvasPanel(
	th *ui.Theme,
	pos docsEventState[docsEventPoint],
	dragging docsEventState[bool],
	menuOpen docsEventState[bool],
	menuAt docsEventState[docsEventPoint],
	addLog func(string),
) ui.Element {
	canvas := ui.PointerAreaElement(
		docsEventCanvasSurface(th, pos.Value(), dragging.Value(), menuOpen.Value(), menuAt.Value()),
		ui.PointerCaptureOnPress(true),
		ui.PointerOnDown(func(ctx *ui.Context, ev *ui.PointerEvent) {
			dragging.Set(true)
			menuOpen.Set(false)
			docsEventMovePoint(pos, ev.Position.X, ev.Position.Y)
			ev.SetPointerCapture(ctx)
			addLog(fmt.Sprintf("指针按下 id=%d 按钮=%d x=%.0f y=%.0f", ev.PointerID, ev.Button, ev.Position.X, ev.Position.Y))
		}),
		ui.PointerOnMove(func(ctx *ui.Context, ev *ui.PointerEvent) {
			if dragging.Value() {
				docsEventMovePoint(pos, ev.Position.X, ev.Position.Y)
			}
			samples := ev.CoalescedSamples()
			if len(samples) > 1 {
				addLog(fmt.Sprintf("指针移动 合并=%d 最后=(%.0f,%.0f)", len(samples), ev.Position.X, ev.Position.Y))
			}
		}, ui.Passive()),
		ui.PointerOnUp(func(ctx *ui.Context, ev *ui.PointerEvent) {
			dragging.Set(false)
			_ = ev.ReleasePointerCapture(ctx)
			addLog(fmt.Sprintf("指针抬起 捕获=%t", ev.HasPointerCapture(ctx)))
		}),
		ui.PointerOnContextMenu(func(ctx *ui.Context, ev *ui.PointerEvent) {
			menuAt.Set(docsEventPoint{X: docsEventClamp(ev.Position.X, 0, 250), Y: docsEventClamp(ev.Position.Y, 0, 125)})
			menuOpen.Set(true)
			ev.PreventDefault()
			addLog(fmt.Sprintf("右键菜单已阻止默认 %.0f,%.0f", ev.Position.X, ev.Position.Y))
		}),
		ui.PointerOnWheel(func(ctx *ui.Context, ev *ui.WheelEvent) {
			ev.PreventDefault()
			addLog(fmt.Sprintf("滚轮 delta=(%.0f, %.0f) 已阻止=%t", ev.DeltaX, ev.DeltaY, ev.DefaultPrevented))
		}),
	)

	return docsEventSection(th, "指针、滚轮和右键菜单", ui.ColumnElement(
		canvas,
		ui.VSpacerElement(8),
		ui.TextElement("PointerAreaElement + 指针捕获 + wheel PreventDefault", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
	))
}

func docsEventCanvasSurface(th *ui.Theme, pos docsEventPoint, dragging bool, menuOpen bool, menuAt docsEventPoint) ui.Element {
	border := ui.Border{Width: 1, Color: th.Colors.OutlineVariant}
	dotColor := ui.NRGBA(37, 99, 235, 255)
	if dragging {
		dotColor = ui.NRGBA(220, 38, 38, 255)
	}

	items := []ui.Element{
		ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithRad(8).WithBorder(border),
			ui.FixedSizeElement(320, 178, ui.SpacerElement(320, 178)),
		),
		ui.PaddingElement(
			ui.Insets{Left: docsEventClamp(pos.X-12, 0, 296), Top: docsEventClamp(pos.Y-12, 0, 154)},
			ui.ContainerDecorationElement(
				ui.Bg(dotColor).WithRad(999),
				ui.FixedSizeElement(24, 24, ui.SpacerElement(24, 24)),
			),
		),
	}
	if menuOpen {
		items = append(items, ui.PaddingElement(
			ui.Insets{Left: menuAt.X, Top: menuAt.Y},
			ui.ContainerDecorationElement(
				ui.Bg(ui.NRGBA(15, 23, 42, 245)).WithPad(ui.Symmetric(8, 10)).WithRad(8),
				ui.ColumnElement(
					ui.TextElement("右键菜单", ui.TextSize(12), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
					ui.TextElement("已捕获右键点击", ui.TextSize(10), ui.TextColor(ui.NRGBA(203, 213, 225, 255))),
				),
			),
		))
	}
	return ui.StackElement(items...)
}

func docsEventKeyboardAndInputPanel(
	th *ui.Theme,
	numberValue docsEventState[string],
	shortcutCount docsEventState[int],
	addLog func(string),
) ui.Element {
	scope := ui.KeyboardScopeElement(
		ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(240, 253, 244, 255)).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(34, 197, 94, 255)}),
			ui.ColumnElement(
				ui.TextElement("本地作用域：Ctrl+K / Escape / Tab", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(6),
				ui.TextElement(fmt.Sprintf("快捷键次数 = %d", shortcutCount.Value()), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		),
		ui.KeyboardScopeFocusable(true),
		ui.KeyboardScopeAutoFocus(true),
		ui.FocusOnFocus(func(ctx *ui.Context, ev *ui.FocusEvent) {
			addLog("焦点作用域获得焦点")
		}),
		ui.FocusOnBlur(func(ctx *ui.Context, ev *ui.FocusEvent) {
			addLog("焦点作用域失去焦点")
		}),
		ui.KeyOnDown(func(ctx *ui.Context, ev *ui.KeyboardEvent) {
			if ev.Key == "Escape" {
				ev.StopPropagation()
				addLog("keydown Escape 已在作用域内停止传播")
			}
		}),
		ui.ShortcutOn(ui.ShortcutKey("k", ui.Modifiers{Ctrl: true}), func(ctx *ui.Context, ev *ui.KeyboardEvent) {
			shortcutCount.Set(shortcutCount.Value() + 1)
			ev.PreventDefault()
			addLog("本地快捷键 Ctrl+K 已阻止默认行为")
		}),
	)

	input := ui.OutlinedTextFieldElement(
		numberValue.Value(),
		ui.InputLabel("只允许数字"),
		ui.InputSingleLine(true),
		ui.InputSupportingText("beforeinput 拦截非数字；Enter 提交"),
		ui.InputOnBeforeInput(func(ctx *ui.Context, ev *ui.InputEvent) {
			if docsEventRejectInput(ev) {
				ev.PreventDefault()
				addLog(fmt.Sprintf("beforeinput 已拦截 data=%q type=%s", ev.Data, ev.InputType))
			}
		}),
		ui.InputOnInputEvent(func(ctx *ui.Context, ev *ui.InputEvent) {
			addLog(fmt.Sprintf("input 值=%q 来源=%s bestEffort=%t", ev.Value, ev.Source, ev.BestEffort))
		}),
		ui.InputOnSubmit(func(ctx *ui.Context, ev *ui.InputEvent) {
			addLog("提交数字=" + ev.Value)
		}),
		ui.InputOnChange(func(ctx *ui.Context, value string) {
			numberValue.Set(value)
		}),
	)

	return docsEventSection(th, "焦点、键盘和富输入", ui.ColumnElement(
		scope,
		ui.VSpacerElement(10),
		input,
	))
}

func docsEventSyntheticPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return docsEventSection(th, "捕获、冒泡和自定义事件", ui.ColumnElement(
		ui.TextElement("合成派发会沿组件事件路径冒泡。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.VSpacerElement(8),
		ui.OutlinedButtonElement(
			ui.TextElement("派发自定义事件"),
			ui.OnClick(func(ctx *ui.Context) {
				allowed := ui.DispatchCustomEvent(
					ctx,
					ctx.PathID(),
					ui.EventType("docs:event-system:synthetic"),
					map[string]string{"source": "docs_browser"},
					ui.CustomCancelable(true),
				)
				addLog(fmt.Sprintf("合成派发 defaultAllowed=%t", allowed))
			}),
		),
		ui.VSpacerElement(8),
		ui.TextElement("OnEvent(ctx, type, handler, Capture()) 可在冒泡前观察捕获阶段。", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
	))
}

func docsEventDragDropPanel(th *ui.Theme, dropState docsEventState[docsEventDropState], addLog func(string)) ui.Element {
	active := dropState.Value().Active
	targetColor := ui.NRGBA(255, 251, 235, 255)
	borderColor := ui.NRGBA(245, 158, 11, 255)
	if active {
		targetColor = ui.NRGBA(254, 243, 199, 255)
		borderColor = ui.NRGBA(217, 119, 6, 255)
	}

	source := ui.DragSourceElement(
		docsEventMiniCard("拖拽 JSON 载荷", "dragstart / drag / dragend", ui.NRGBA(239, 246, 255, 255), ui.NRGBA(37, 99, 235, 255)),
		ui.DragSourcePayloads(
			ui.DragPayload{Type: "application/json", Data: []byte(`{"kind":"event-system","source":"docs_browser"}`)},
			ui.DragPayload{Type: "text/plain", Data: []byte("FluxUI event-system payload")},
		),
		ui.DragSourceOperations(ui.DragOperationCopy, ui.DragOperationMove),
		ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
			addLog(fmt.Sprintf("拖拽源 %s op=%s type=%s", docsDragSourceEventLabel(event.Kind), event.Operation, event.Type))
		}),
	)

	target := ui.DropTargetElement(
		docsEventMiniCard("拖放目标", dropState.Value().Text, targetColor, borderColor),
		func(ctx *ui.Context, event ui.DropEvent) {
			next := dropState.Value()
			next.Active = false
			next.Text = fmt.Sprintf("%s bytes=%d op=%s", event.Type, len(event.Data), event.Operation)
			if event.Text != "" {
				next.Text = compactDocsPayloadText(event.Text)
			}
			dropState.Set(next)
			addLog("拖放 " + next.Text)
		},
		ui.DropTargetTypes("application/json", "application/x-fluxui-doc", "text/plain", "text/plain;charset=utf-8"),
		ui.DropTargetOperation(ui.DragOperationCopy),
		ui.DropTargetOnActiveChange(func(ctx *ui.Context, event ui.DropTargetStateEvent) {
			next := dropState.Value()
			next.Active = event.Active
			dropState.Set(next)
			addLog("拖放目标 active=" + fmt.Sprint(event.Active))
		}),
		ui.DropTargetOnError(func(ctx *ui.Context, event ui.DropEvent) {
			if event.Err != nil {
				addLog("拖放错误 " + event.Err.Error())
			} else {
				addLog("拖放错误")
			}
		}),
	)

	return docsEventSection(th, "拖放事件流", ui.RowElement(
		ui.ExpandedElement(source),
		ui.HSpacerElement(10),
		ui.ExpandedElement(target),
	))
}

func docsEventSection(th *ui.Theme, title string, child ui.Element) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(15), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			child,
		),
	)
}

func docsEventMiniCard(title, subtitle string, bg color.NRGBA, accent color.NRGBA) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(bg).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: accent}),
		ui.FixedHeightElement(76, ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(13), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
			ui.VSpacerElement(5),
			ui.TextElement(subtitle, ui.TextSize(11), ui.TextColor(accent)),
		)),
	)
}

func docsEventLogPanel(th *ui.Theme, logs []string) ui.Element {
	items := []ui.Element{
		ui.TextElement("事件日志", ui.TextSize(15), ui.TextColor(th.Colors.OnSurface)),
	}
	for _, item := range logs {
		items = append(items, ui.PaddingElement(
			ui.Insets{Top: 4},
			ui.TextElement(item, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
		))
	}
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(10)).WithRad(8),
		ui.FixedHeightElement(132, ui.ScrollViewElement(ui.ColumnElement(items...), ui.ScrollVertical(true))),
	)
}

func docsEventMovePoint(pos docsEventState[docsEventPoint], x, y float32) {
	pos.Set(docsEventPoint{
		X: docsEventClamp(x, 16, 304),
		Y: docsEventClamp(y, 16, 162),
	})
}

func docsEventRejectInput(ev *ui.InputEvent) bool {
	if ev == nil || ev.Data == "" {
		return false
	}
	if !strings.HasPrefix(ev.InputType, "insert") {
		return false
	}
	for _, r := range ev.Data {
		if !unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func docsEventClamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func docsEventPhaseLabel(phase ui.EventPhase) string {
	switch phase {
	case ui.EventPhaseCapture:
		return "捕获"
	case ui.EventPhaseTarget:
		return "目标"
	case ui.EventPhaseBubble:
		return "冒泡"
	default:
		return "无"
	}
}
