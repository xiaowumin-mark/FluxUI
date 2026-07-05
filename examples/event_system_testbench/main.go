package main

import (
	"fmt"
	"strings"
	"unicode"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type benchState[T any] interface {
	Value() T
	Set(T)
}

type pointState struct {
	X float32
	Y float32
}

type dropState struct {
	Active bool
	Text   string
}

func app(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	logs := ui.UseState(ctx, []string{"就绪：请从上到下测试每个事件区域"})

	addLog := func(message string) {
		prependLog(logs, message)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.Surface).WithPad(ui.All(16)),
		ui.ScrollViewElement(
			ui.ColumnElement(
				headerPanel(th),
				ui.VSpacerElement(12),
				roadmapChecklistPanel(th),
				ui.VSpacerElement(12),
				compatPanel(th, addLog),
				ui.VSpacerElement(12),
				dispatchPanel(th, addLog),
				ui.VSpacerElement(12),
				pointerWheelPanel(th, addLog),
				ui.VSpacerElement(12),
				keyboardFocusPanel(th, addLog),
				ui.VSpacerElement(12),
				textInputPanel(th, addLog),
				ui.VSpacerElement(12),
				defaultBehaviorPanel(th, addLog),
				ui.VSpacerElement(12),
				dragDropPanel(th, addLog),
				ui.VSpacerElement(12),
				architecturePanel(th, addLog),
				ui.VSpacerElement(12),
				diagnosticsPanel(th, addLog),
				ui.VSpacerElement(12),
				logPanel(th, logs.Value()),
			),
			ui.ScrollVertical(true),
		),
	)
}

func headerPanel(th *ui.Theme) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.PrimaryContainer).WithPad(ui.All(16)).WithRad(10),
		ui.ColumnElement(
			ui.TextElement("事件系统路线图全量测试台", ui.TextSize(24), ui.TextColor(th.Colors.OnPrimaryContainer)),
			ui.VSpacerElement(6),
			ui.TextElement("独立 example：覆盖 P0 到 P7 的事件 API、默认行为、边界策略和诊断能力。所有界面文案为中文，事件 API 名保留英文以便对照代码。", ui.TextSize(13), ui.TextColor(th.Colors.OnPrimaryContainer)),
		),
	)
}

func roadmapChecklistPanel(th *ui.Theme) ui.Element {
	items := []string{
		"P0 兼容边界：OnClick / OnHover / InputOnChange / ScrollOnChange / ctx.Gtx.Event escape hatch",
		"P1 EventTarget：TargetID、Type、Event、Phase、capture / target / bubble、stop、preventDefault、Once、Passive、synthetic dispatch",
		"P2 Pointer / Wheel：pointerdown/up/move/enter/leave/over/out/cancel、click/dblclick/auxclick/contextmenu、wheel、pointer capture、coalesced move",
		"P3 Focus / Keyboard：FocusManager、focus/blur/focusin/focusout、keydown/keyup、repeat、key/code/modifiers、局部快捷键、Enter/Space/Escape/Tab 默认行为",
		"P4 Text / IME：beforeinput、input、change、submit、compositionstart/update/end、输入来源、程序化 SetText、best-effort 拦截",
		"P5 默认行为迁移：Button、Pressable、ClickArea、Checkbox、Switch、Radio、Select、ScrollView、Slider、Dialog、Popup、DragSource、DropTarget",
		"P6 自定义事件和边界：Detail payload、DispatchEvent 到指定 target、EventPortal、EventBoundary stop/redirect、activation event",
		"P7 性能和诊断：LogEvents、事件路径、listener 耗时、取消状态、pointer/wheel 高频日志与分配观察入口",
	}
	rows := make([]ui.Element, 0, len(items))
	for _, item := range items {
		rows = append(rows, ui.PaddingElement(
			ui.Insets{Top: 4},
			ui.TextElement("• "+item, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		))
	}
	return section(th, "路线图功能清单", "下面每个区域都对应一个或多个验收点，右侧/底部事件日志用于记录实际触发情况。", ui.ColumnElement(rows...))
}

func compatPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		input := ui.UseState(ctx, "兼容输入")
		hover := ui.UseState(ctx, false)
		scrollOffset := ui.UseState(ctx, "0,0")

		scrollItems := make([]ui.Element, 0, 18)
		for i := 1; i <= 18; i++ {
			scrollItems = append(scrollItems, ui.PaddingElement(
				ui.Insets{Top: 4},
				ui.TextElement(fmt.Sprintf("旧 ScrollOnChange 测试行 %02d", i), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			))
		}

		return section(th, "P0 兼容 API 和 escape hatch", "验证旧 API 仍可用；高级用户仍可直接使用 ctx.Gtx.Event，但推荐迁移到新事件层。", ui.ColumnElement(
			ui.RowElement(
				ui.OutlinedButtonElement(
					ui.TextElement("旧 OnClick"),
					ui.OnClick(func(ctx *ui.Context) {
						addLog("P0：旧 OnClick(ctx) 已触发")
					}),
					ui.OnHover(func(ctx *ui.Context, value bool) {
						if hover.Value() != value {
							hover.Set(value)
							addLog(fmt.Sprintf("P0：旧 OnHover(ctx, %t)", value))
						}
					}),
				),
				ui.HSpacerElement(10),
				ui.TextElement(fmt.Sprintf("hover=%t", hover.Value()), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
			ui.VSpacerElement(10),
			ui.OutlinedTextFieldElement(
				input.Value(),
				ui.InputLabel("旧 InputOnChange"),
				ui.InputSupportingText("输入任意文本，日志应显示旧 change 回调仍然触发"),
				ui.InputOnChange(func(ctx *ui.Context, value string) {
					input.Set(value)
					addLog("P0：InputOnChange value=" + value)
				}),
			),
			ui.VSpacerElement(10),
			ui.TextElement("旧 ScrollOnChange offset="+scrollOffset.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.FixedHeightElement(104, ui.ScrollViewElement(
				ui.ColumnElement(scrollItems...),
				ui.ScrollVertical(true),
				ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
					scrollOffset.Set(fmt.Sprintf("%.0f,%.0f", x, y))
					addLog(fmt.Sprintf("P0：ScrollOnChange x=%.0f y=%.0f", x, y))
				}),
			)),
			ui.VSpacerElement(8),
			ui.TextElement("escape hatch：如确需临时访问 Gio 原始输入，可在自定义 Widget 内使用 ctx.Gtx.Event(filter)。本测试台只展示推荐事件层，避免抢占事件。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		))
	})
}

func dispatchPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		preventDefault := ui.UseState(ctx, false)
		stopBubble := ui.UseState(ctx, false)
		customType := ui.EventType("event-test:dispatch")
		immediateType := ui.EventType("event-test:immediate")

		ui.OnEvent(ctx, customType, func(ctx *ui.Context, ev *ui.Event) {
			addLog(fmt.Sprintf("P1：父级 capture phase=%s path=%d", phaseLabel(ev.Phase), len(ev.ComposedPath())))
		}, ui.Capture(), ui.EventPriority(20))
		ui.OnEvent(ctx, customType, func(ctx *ui.Context, ev *ui.Event) {
			addLog(fmt.Sprintf("P1：父级 bubble defaultPrevented=%t", ev.DefaultPrevented))
		})
		ui.OnEvent(ctx, immediateType, func(ctx *ui.Context, ev *ui.Event) {
			addLog("P1：父级 immediate bubble 不应在 StopImmediate 后出现")
		})

		target := ui.ComponentElement(func(targetCtx *ui.Context) ui.Element {
			targetID := targetCtx.PathID()
			ui.OnEvent(targetCtx, customType, func(ctx *ui.Context, ev *ui.Event) {
				ok := ev.PreventDefault()
				addLog(fmt.Sprintf("P1：passive listener 调 PreventDefault 返回=%t default=%t", ok, ev.DefaultPrevented))
			}, ui.Passive(), ui.EventPriority(15))
			ui.OnEvent(targetCtx, customType, func(ctx *ui.Context, ev *ui.Event) {
				addLog("P1：once listener 本次按钮内两次派发只应出现一次")
			}, ui.Once(), ui.EventPriority(10))
			ui.OnEvent(targetCtx, customType, func(ctx *ui.Context, ev *ui.Event) {
				if preventDefault.Value() {
					ev.PreventDefault()
				}
				if stopBubble.Value() {
					ev.StopPropagation()
				}
				addLog(fmt.Sprintf("P1：目标 listener detail=%v prevent=%t stop=%t", ev.Detail, ev.DefaultPrevented, ev.PropagationStopped()))
			})
			ui.OnEvent(targetCtx, immediateType, func(ctx *ui.Context, ev *ui.Event) {
				addLog("P1：StopImmediate 第一个 listener 已执行")
				ev.StopImmediatePropagation()
			}, ui.EventPriority(10))
			ui.OnEvent(targetCtx, immediateType, func(ctx *ui.Context, ev *ui.Event) {
				addLog("P1：第二个 immediate listener 不应出现")
			})

			return ui.ContainerDecorationElement(
				panelBg(th).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
				ui.ColumnElement(
					ui.TextElement("目标 target", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
					ui.VSpacerElement(8),
					ui.RowElement(
						ui.OutlinedButtonElement(
							ui.TextElement("派发一次"),
							ui.OnClick(func(ctx *ui.Context) {
								allowed := ui.DispatchCustomEvent(ctx, targetID, customType, map[string]string{"按钮": "派发一次"}, ui.CustomCancelable(true))
								addLog(fmt.Sprintf("P1：synthetic dispatch 返回 defaultAllowed=%t", allowed))
							}),
						),
						ui.HSpacerElement(8),
						ui.OutlinedButtonElement(
							ui.TextElement("同帧派发两次"),
							ui.OnClick(func(ctx *ui.Context) {
								ui.DispatchCustomEvent(ctx, targetID, customType, "第一次", ui.CustomCancelable(true))
								ui.DispatchCustomEvent(ctx, targetID, customType, "第二次", ui.CustomCancelable(true))
							}),
						),
						ui.HSpacerElement(8),
						ui.OutlinedButtonElement(
							ui.TextElement("StopImmediate"),
							ui.OnClick(func(ctx *ui.Context) {
								ui.DispatchCustomEvent(ctx, targetID, immediateType, "immediate", ui.CustomCancelable(true))
							}),
						),
					),
				),
			)
		})

		return section(th, "P1 EventTarget 和基础分发", "测试 capture/target/bubble、StopPropagation、StopImmediatePropagation、PreventDefault、Once、Passive、synthetic dispatch。", ui.ColumnElement(
			ui.RowElement(
				toggleButton("PreventDefault="+boolText(preventDefault.Value()), func(ctx *ui.Context) {
					preventDefault.Set(!preventDefault.Value())
					addLog(fmt.Sprintf("P1：PreventDefault 开关=%t", !preventDefault.Value()))
				}),
				ui.HSpacerElement(8),
				toggleButton("StopPropagation="+boolText(stopBubble.Value()), func(ctx *ui.Context) {
					stopBubble.Set(!stopBubble.Value())
					addLog(fmt.Sprintf("P1：StopPropagation 开关=%t", !stopBubble.Value()))
				}),
			),
			ui.VSpacerElement(10),
			target,
		))
	})
}

func pointerWheelPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		pos := ui.UseState(ctx, pointState{X: 160, Y: 80})
		dragging := ui.UseState(ctx, false)
		menuOpen := ui.UseState(ctx, false)
		menuAt := ui.UseState(ctx, pointState{X: 180, Y: 60})
		moveCount := ui.UseState(ctx, 0)
		wheelCount := ui.UseState(ctx, 0)
		lastMove := ui.UseState(ctx, "尚未移动")
		lastWheel := ui.UseState(ctx, "尚未滚轮")

		surface := pointerSurface(th, pos.Value(), dragging.Value(), menuOpen.Value(), menuAt.Value(), moveCount.Value(), lastMove.Value(), lastWheel.Value())
		area := ui.PointerAreaElement(
			surface,
			ui.PointerCaptureOnPress(true),
			ui.PointerOnOver(func(ctx *ui.Context, ev *ui.PointerEvent) {
				addLog("P2：pointerover")
			}),
			ui.PointerOnEnter(func(ctx *ui.Context, ev *ui.PointerEvent) {
				addLog("P2：pointerenter")
			}),
			ui.PointerOnLeave(func(ctx *ui.Context, ev *ui.PointerEvent) {
				addLog("P2：pointerleave")
			}),
			ui.PointerOnOut(func(ctx *ui.Context, ev *ui.PointerEvent) {
				addLog("P2：pointerout")
			}),
			ui.PointerOnDown(func(ctx *ui.Context, ev *ui.PointerEvent) {
				dragging.Set(true)
				menuOpen.Set(false)
				pos.Set(pointState{X: clamp(ev.Position.X, 16, 344), Y: clamp(ev.Position.Y, 16, 184)})
				ev.SetPointerCapture(ctx)
				addLog(fmt.Sprintf("P2：pointerdown id=%d type=%s button=%d buttons=%d mods=%s x=%.0f y=%.0f capture=%t", ev.PointerID, ev.PointerType, ev.Button, ev.Buttons, modifierText(ev.Modifiers), ev.Position.X, ev.Position.Y, ev.HasPointerCapture(ctx)))
			}),
			ui.PointerOnMove(func(ctx *ui.Context, ev *ui.PointerEvent) {
				if dragging.Value() {
					pos.Set(pointState{X: clamp(ev.Position.X, 16, 344), Y: clamp(ev.Position.Y, 16, 184)})
				}
				nextCount := moveCount.Value() + 1
				moveCount.Set(nextCount)
				samples := ev.CoalescedSamples()
				lastMove.Set(fmt.Sprintf("move=%d x=%.0f y=%.0f coalesced=%d", nextCount, ev.Position.X, ev.Position.Y, len(samples)))
				if len(samples) > 1 || nextCount%12 == 0 {
					addLog(fmt.Sprintf("P2：pointermove #%d coalesced=%d last=(%.0f,%.0f)", nextCount, len(samples), ev.Position.X, ev.Position.Y))
				}
			}, ui.Passive()),
			ui.PointerOnUp(func(ctx *ui.Context, ev *ui.PointerEvent) {
				dragging.Set(false)
				released := ev.ReleasePointerCapture(ctx)
				addLog(fmt.Sprintf("P2：pointerup releasedCapture=%t hasCapture=%t", released, ev.HasPointerCapture(ctx)))
			}),
			ui.PointerOnCancel(func(ctx *ui.Context, ev *ui.PointerEvent) {
				dragging.Set(false)
				addLog("P2：pointercancel")
			}),
			ui.PointerOnClick(func(ctx *ui.Context, ev *ui.PointerEvent) {
				addLog(fmt.Sprintf("P2：click count=%d x=%.0f y=%.0f", ev.ClickCount, ev.Position.X, ev.Position.Y))
			}),
			ui.PointerOnDoubleClick(func(ctx *ui.Context, ev *ui.PointerEvent) {
				addLog(fmt.Sprintf("P2：dblclick x=%.0f y=%.0f", ev.Position.X, ev.Position.Y))
			}),
			ui.PointerOnAuxClick(func(ctx *ui.Context, ev *ui.PointerEvent) {
				addLog(fmt.Sprintf("P2：auxclick button=%d", ev.Button))
			}),
			ui.PointerOnContextMenu(func(ctx *ui.Context, ev *ui.PointerEvent) {
				menuAt.Set(pointState{X: clamp(ev.Position.X, 0, 270), Y: clamp(ev.Position.Y, 0, 150)})
				menuOpen.Set(true)
				ev.PreventDefault()
				addLog(fmt.Sprintf("P2：contextmenu 已阻止默认 x=%.0f y=%.0f", ev.Position.X, ev.Position.Y))
			}),
			ui.PointerOnWheel(func(ctx *ui.Context, ev *ui.WheelEvent) {
				ev.PreventDefault()
				next := wheelCount.Value() + 1
				wheelCount.Set(next)
				lastWheel.Set(fmt.Sprintf("wheel=%d dx=%.0f dy=%.0f mode=%d x=%.0f y=%.0f", next, ev.DeltaX, ev.DeltaY, ev.DeltaMode, ev.Position.X, ev.Position.Y))
				addLog(fmt.Sprintf("P2：wheel dx=%.0f dy=%.0f prevented=%t mods=%s", ev.DeltaX, ev.DeltaY, ev.DefaultPrevented, modifierText(ev.Modifiers)))
			}),
		)

		return section(th, "P2 Pointer 和 Wheel", "拖拽蓝点、右键、中键、双击、滚轮。区域会测试 pointer capture、右键菜单、wheel PreventDefault 和 coalesced move。", area)
	})
}

func pointerSurface(th *ui.Theme, pos pointState, dragging bool, menuOpen bool, menuAt pointState, moveCount int, lastMove, lastWheel string) ui.Element {
	dotColor := ui.NRGBA(37, 99, 235, 255)
	if dragging {
		dotColor = ui.NRGBA(220, 38, 38, 255)
	}
	items := []ui.Element{
		ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.FixedSizeElement(372, 212, ui.SpacerElement(372, 212)),
		),
		ui.PaddingElement(
			ui.Insets{Left: 12, Top: 10},
			ui.ColumnElement(
				ui.TextElement("在此区域测试指针事件", ui.TextSize(13), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
				ui.TextElement(lastMove, ui.TextSize(11), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
				ui.TextElement(lastWheel, ui.TextSize(11), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
			),
		),
		ui.PaddingElement(
			ui.Insets{Left: clamp(pos.X-12, 0, 348), Top: clamp(pos.Y-12, 0, 188)},
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
					ui.TextElement("contextmenu 已捕获", ui.TextSize(10), ui.TextColor(ui.NRGBA(203, 213, 225, 255))),
				),
			),
		))
	}
	_ = moveCount
	return ui.StackElement(items...)
}

func keyboardFocusPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		preventTab := ui.UseState(ctx, false)
		shortcutCount := ui.UseState(ctx, 0)
		activeLabel := ui.UseState(ctx, "无")

		scope := ui.KeyboardScopeElement(
			ui.ContainerDecorationElement(
				ui.Bg(ui.NRGBA(240, 253, 244, 255)).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(34, 197, 94, 255)}),
				ui.ColumnElement(
					ui.TextElement("键盘作用域：点击焦点块后测试 Tab、Shift+Tab、Ctrl+K、Escape、Enter、Space", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
					ui.VSpacerElement(8),
					ui.RowElement(
						focusProbe(th, "焦点 A", 0, false, false, activeLabel, addLog),
						ui.HSpacerElement(8),
						focusProbe(th, "禁用 B", 1, true, false, activeLabel, addLog),
						ui.HSpacerElement(8),
						focusProbe(th, "隐藏 C", 2, false, true, activeLabel, addLog),
					),
					ui.VSpacerElement(8),
					ui.RowElement(
						ui.OutlinedButtonElement(
							ui.TextElement("MoveFocus 下一项"),
							ui.OnClick(func(ctx *ui.Context) {
								ok := ui.FocusManagerFor(ctx).Move(ui.FocusForward)
								addLog(fmt.Sprintf("P3：FocusManager.Move forward=%t", ok))
							}),
						),
						ui.HSpacerElement(8),
						ui.OutlinedButtonElement(
							ui.TextElement("MoveFocus 上一项"),
							ui.OnClick(func(ctx *ui.Context) {
								ok := ui.FocusManagerFor(ctx).Move(ui.FocusBackward)
								addLog(fmt.Sprintf("P3：FocusManager.Move backward=%t", ok))
							}),
						),
						ui.HSpacerElement(8),
						ui.OutlinedButtonElement(
							ui.TextElement("Blur 当前焦点"),
							ui.OnClick(func(ctx *ui.Context) {
								ok := ui.FocusManagerFor(ctx).Blur()
								addLog(fmt.Sprintf("P3：FocusManager.Blur=%t", ok))
							}),
						),
					),
					ui.VSpacerElement(8),
					ui.RowElement(
						toggleButton("阻止 Tab 默认="+boolText(preventTab.Value()), func(ctx *ui.Context) {
							preventTab.Set(!preventTab.Value())
						}),
						ui.HSpacerElement(8),
						ui.TextElement(fmt.Sprintf("局部快捷键次数=%d 当前焦点=%s", shortcutCount.Value(), activeLabel.Value()), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
					),
				),
			),
			ui.KeyboardScopeFocusable(true),
			ui.KeyboardScopeAutoFocus(true),
			ui.FocusOnFocus(func(ctx *ui.Context, ev *ui.FocusEvent) {
				addLog("P3：KeyboardScope focus")
			}),
			ui.FocusOnBlur(func(ctx *ui.Context, ev *ui.FocusEvent) {
				addLog("P3：KeyboardScope blur")
			}),
			ui.FocusOnIn(func(ctx *ui.Context, ev *ui.FocusEvent) {
				addLog(fmt.Sprintf("P3：focusin related=%d", ev.RelatedTarget))
			}),
			ui.FocusOnOut(func(ctx *ui.Context, ev *ui.FocusEvent) {
				addLog(fmt.Sprintf("P3：focusout related=%d", ev.RelatedTarget))
			}),
			ui.KeyOnDown(func(ctx *ui.Context, ev *ui.KeyboardEvent) {
				addLog(fmt.Sprintf("P3：keydown key=%q code=%q repeat=%t mods=%s composing=%t", ev.Key, ev.Code, ev.Repeat, modifierText(ev.Modifiers), ev.IsComposing))
				if ev.Key == "Escape" {
					ev.StopPropagation()
					addLog("P3：Escape 已在局部作用域停止传播")
				}
				if ev.Key == "Tab" && preventTab.Value() {
					ev.PreventDefault()
					addLog("P3：Tab 默认焦点移动已取消")
				}
			}),
			ui.KeyOnUp(func(ctx *ui.Context, ev *ui.KeyboardEvent) {
				addLog(fmt.Sprintf("P3：keyup key=%q code=%q", ev.Key, ev.Code))
			}),
			ui.ShortcutOn(ui.ShortcutKey("k", ui.Modifiers{Ctrl: true}), func(ctx *ui.Context, ev *ui.KeyboardEvent) {
				shortcutCount.Set(shortcutCount.Value() + 1)
				ev.PreventDefault()
				addLog("P3：局部快捷键 Ctrl+K 已触发并阻止默认行为")
			}),
		)

		return section(th, "P3 Focus、Keyboard 和局部快捷键", "组件树内键盘事件与 system global shortcut 分离；本区域只在局部 focus scope 内响应。", scope)
	})
}

func focusProbe(th *ui.Theme, label string, tabIndex int, disabled bool, hidden bool, activeLabel benchState[string], addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		targetID := ctx.PathID()
		ui.RegisterFocusTarget(ctx,
			ui.FocusTabIndex(tabIndex),
			ui.FocusDisabled(disabled),
			ui.FocusHidden(hidden),
			ui.FocusActivate(func(ctx *ui.Context) {
				addLog("P3：Enter/Space 默认激活 " + label)
			}),
		)
		ui.OnEvent(ctx, ui.EventFocus, func(ctx *ui.Context, ev *ui.Event) {
			activeLabel.Set(label)
			addLog("P3：" + label + " focus")
		})
		ui.OnEvent(ctx, ui.EventBlur, func(ctx *ui.Context, ev *ui.Event) {
			addLog("P3：" + label + " blur")
		})

		bg := ui.NRGBA(255, 255, 255, 255)
		border := th.Colors.OutlineVariant
		if ui.IsFocused(ctx) {
			bg = ui.NRGBA(219, 234, 254, 255)
			border = ui.NRGBA(37, 99, 235, 255)
		}
		if disabled || hidden {
			bg = ui.NRGBA(241, 245, 249, 255)
		}
		status := "可聚焦"
		if disabled {
			status = "disabled"
		}
		if hidden {
			status = "hidden"
		}
		return ui.PointerAreaElement(
			ui.ContainerDecorationElement(
				ui.Bg(bg).WithPad(ui.All(10)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: border}),
				ui.FixedSizeElement(136, 58, ui.ColumnElement(
					ui.TextElement(label, ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
					ui.TextElement(status, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
				)),
			),
			ui.PointerOnClick(func(ctx *ui.Context, ev *ui.PointerEvent) {
				ok := ui.FocusManagerFor(ctx).Request(targetID)
				addLog(fmt.Sprintf("P3：点击请求焦点 %s => %t", label, ok))
			}),
		)
	})
}

func textInputPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		digits := ui.UseState(ctx, "123")
		freeText := ui.UseState(ctx, "FluxUI")
		compositionText := ui.UseState(ctx, "尚未派发 composition")
		inputRef := ui.UseRef(ctx, ui.NewInputRef())
		if inputRef.Current == nil {
			inputRef.Current = ui.NewInputRef()
		}

		compositionTarget := ui.ComponentElement(func(targetCtx *ui.Context) ui.Element {
			targetID := targetCtx.PathID()
			fluxevent.OnComposition(targetCtx, fluxevent.CompositionStart, func(ctx *ui.Context, ev *fluxevent.CompositionEvent) {
				compositionText.Set("compositionstart: " + ev.Data)
				addLog("P4：compositionstart data=" + ev.Data)
			})
			fluxevent.OnComposition(targetCtx, fluxevent.CompositionUpdate, func(ctx *ui.Context, ev *fluxevent.CompositionEvent) {
				compositionText.Set("compositionupdate: " + ev.Data)
				addLog("P4：compositionupdate data=" + ev.Data)
			})
			fluxevent.OnComposition(targetCtx, fluxevent.CompositionEnd, func(ctx *ui.Context, ev *fluxevent.CompositionEvent) {
				compositionText.Set("compositionend: " + ev.Data)
				addLog("P4：compositionend data=" + ev.Data)
			})
			return ui.ContainerDecorationElement(
				panelBg(th).WithPad(ui.All(10)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
				ui.ColumnElement(
					ui.TextElement("IME composition synthetic 入口", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
					ui.TextElement(compositionText.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
					ui.VSpacerElement(8),
					ui.RowElement(
						ui.OutlinedButtonElement(
							ui.TextElement("派发 start/update/end"),
							ui.OnClick(func(ctx *ui.Context) {
								fluxevent.DispatchCompositionEvent(ctx, targetID, &fluxevent.CompositionEvent{Event: fluxevent.Event{Type: fluxevent.CompositionStart}, Data: "拼"})
								fluxevent.DispatchCompositionEvent(ctx, targetID, &fluxevent.CompositionEvent{Event: fluxevent.Event{Type: fluxevent.CompositionUpdate}, Data: "拼音"})
								fluxevent.DispatchCompositionEvent(ctx, targetID, &fluxevent.CompositionEvent{Event: fluxevent.Event{Type: fluxevent.CompositionEnd}, Data: "拼音"})
							}),
						),
						ui.HSpacerElement(8),
						ui.TextElement("真实 IME 生命周期取决于 Gio 后端；这里测试事件层入口。", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
					),
				),
			)
		})

		return section(th, "P4 Text Input、IME 和编辑事件", "测试 beforeinput 可取消、input/change/submit、输入来源、程序化 ref 命令和 composition 事件入口。", ui.ColumnElement(
			ui.RowElement(
				ui.ExpandedElement(ui.OutlinedTextFieldElement(
					digits.Value(),
					ui.InputLabel("只允许数字"),
					ui.InputSingleLine(true),
					ui.InputSupportingText("beforeinput 会拦截非数字；Enter 提交"),
					ui.InputOnBeforeInput(func(ctx *ui.Context, ev *ui.InputEvent) {
						if rejectNonDigit(ev) {
							ev.PreventDefault()
							addLog(fmt.Sprintf("P4：beforeinput 拦截 data=%q type=%s source=%s bestEffort=%t", ev.Data, ev.InputType, ev.Source, ev.BestEffort))
						}
					}),
					ui.InputOnInputEvent(func(ctx *ui.Context, ev *ui.InputEvent) {
						addLog(fmt.Sprintf("P4：input value=%q data=%q type=%s source=%s bestEffort=%t", ev.Value, ev.Data, ev.InputType, ev.Source, ev.BestEffort))
					}),
					ui.InputOnSubmit(func(ctx *ui.Context, ev *ui.InputEvent) {
						addLog("P4：submit digits=" + ev.Value)
					}),
					ui.InputOnChange(func(ctx *ui.Context, value string) {
						digits.Set(value)
						addLog("P4：change digits=" + value)
					}),
				)),
				ui.HSpacerElement(12),
				ui.ExpandedElement(ui.OutlinedTextFieldElement(
					freeText.Value(),
					ui.InputAttachRef(inputRef.Current),
					ui.InputLabel("程序化输入来源"),
					ui.InputSingleLine(true),
					ui.InputSupportingText("SetText / Append / Clear 应标记 programmatic 来源"),
					ui.InputOnInputEvent(func(ctx *ui.Context, ev *ui.InputEvent) {
						addLog(fmt.Sprintf("P4：programmatic input value=%q type=%s source=%s", ev.Value, ev.InputType, ev.Source))
					}),
					ui.InputOnChange(func(ctx *ui.Context, value string) {
						freeText.Set(value)
					}),
				)),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.OutlinedButtonElement(ui.TextElement("SetText"), ui.OnClick(func(ctx *ui.Context) {
					inputRef.Current.SetText("程序化设置")
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("Append"), ui.OnClick(func(ctx *ui.Context) {
					inputRef.Current.Append("+追加")
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("Clear"), ui.OnClick(func(ctx *ui.Context) {
					inputRef.Current.Clear()
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("Focus"), ui.OnClick(func(ctx *ui.Context) {
					inputRef.Current.Focus()
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("Blur"), ui.OnClick(func(ctx *ui.Context) {
					inputRef.Current.Blur()
				})),
			),
			ui.VSpacerElement(10),
			compositionTarget,
		))
	})
}

func defaultBehaviorPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		blockClick := ui.UseState(ctx, false)
		blockWheel := ui.UseState(ctx, false)
		checkbox := ui.UseState(ctx, true)
		switchValue := ui.UseState(ctx, false)
		radio := ui.UseState(ctx, "alpha")
		selectValue := ui.UseState(ctx, "copy")
		slider := ui.UseState(ctx, float32(42))
		dialogOpen := ui.UseState(ctx, false)
		popupOpen := ui.UseState(ctx, false)

		zone := ui.ComponentElement(func(zoneCtx *ui.Context) ui.Element {
			ui.OnEvent(zoneCtx, ui.EventClick, func(ctx *ui.Context, ev *ui.Event) {
				addLog(fmt.Sprintf("P5：父容器 capture click target=%d default=%t", ev.Target, ev.DefaultPrevented))
				if blockClick.Value() {
					ev.PreventDefault()
					addLog("P5：父容器阻止 click 默认行为，子组件 onChange/onClick 不应执行")
				}
			}, ui.Capture())
			ui.OnEvent(zoneCtx, ui.EventWheel, func(ctx *ui.Context, ev *ui.Event) {
				if blockWheel.Value() {
					ev.PreventDefault()
					addLog("P5：父容器阻止 wheel 默认滚动")
				}
			}, ui.Capture())

			scrollRows := make([]ui.Element, 0, 22)
			for i := 1; i <= 22; i++ {
				scrollRows = append(scrollRows, ui.TextElement(fmt.Sprintf("ScrollView 默认 wheel 行 %02d", i), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)))
			}

			return ui.ContainerDecorationElement(
				panelBg(th).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
				ui.ColumnElement(
					ui.RowElement(
						ui.OutlinedButtonElement(ui.TextElement("Button"), ui.OnClick(func(ctx *ui.Context) {
							addLog("P5：Button OnClick 默认行为执行")
						})),
						ui.HSpacerElement(8),
						ui.PressableElement(
							ui.ContainerDecorationElement(ui.Bg(ui.NRGBA(224, 242, 254, 255)).WithPad(ui.Symmetric(8, 12)).WithRad(8), ui.TextElement("Pressable")),
							func(ctx *ui.Context) {
								addLog("P5：Pressable 点击默认行为执行")
							},
						),
						ui.HSpacerElement(8),
						ui.ClickAreaElement(
							ui.ContainerDecorationElement(ui.Bg(ui.NRGBA(254, 243, 199, 255)).WithPad(ui.Symmetric(8, 12)).WithRad(8), ui.TextElement("ClickArea")),
							func(ctx *ui.Context) {
								addLog("P5：ClickArea 点击默认行为执行")
							},
						),
					),
					ui.VSpacerElement(10),
					ui.RowElement(
						ui.ExpandedElement(ui.CheckboxElement("Checkbox", checkbox.Value(), ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
							checkbox.Set(checked)
							addLog(fmt.Sprintf("P5：Checkbox onChange=%t", checked))
						}))),
						ui.HSpacerElement(8),
						ui.ExpandedElement(ui.RowElement(
							ui.SwitchElement(switchValue.Value(), ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
								switchValue.Set(checked)
								addLog(fmt.Sprintf("P5：Switch onChange=%t", checked))
							})),
							ui.HSpacerElement(8),
							ui.TextElement("Switch", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
						)),
					),
					ui.VSpacerElement(10),
					ui.RowElement(
						ui.ExpandedElement(ui.RadioGroupElement(
							radio.Value(),
							[]ui.RadioItem{{Label: "Alpha", Value: "alpha"}, {Label: "Beta", Value: "beta"}},
							ui.RadioGroupOnChange(func(ctx *ui.Context, value string) {
								radio.Set(value)
								addLog("P5：Radio onChange=" + value)
							}),
						)),
						ui.HSpacerElement(12),
						ui.ExpandedElement(ui.OutlinedSelectElement[string](
							selectValue.Value(),
							[]ui.SelectOptionItem[string]{{Label: "复制", Value: "copy"}, {Label: "移动", Value: "move"}, {Label: "链接", Value: "link"}},
							ui.SelectLabel[string]("Select"),
							ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
								selectValue.Set(value)
								addLog("P5：Select onChange=" + value)
							}),
						)),
					),
					ui.VSpacerElement(10),
					ui.TextElement(fmt.Sprintf("Slider value=%.0f", slider.Value()), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
					ui.SliderElement(slider.Value(), ui.SliderMin(0), ui.SliderMax(100), ui.SliderStep(1), ui.SliderOnChange(func(ctx *ui.Context, value float32) {
						slider.Set(value)
						addLog(fmt.Sprintf("P5：Slider pointer 默认行为 value=%.0f", value))
					})),
					ui.VSpacerElement(10),
					ui.RowElement(
						ui.OutlinedButtonElement(ui.TextElement("打开 Dialog"), ui.OnClick(func(ctx *ui.Context) {
							dialogOpen.Set(true)
							addLog("P5：Dialog open")
						})),
						ui.HSpacerElement(8),
						ui.OutlinedButtonElement(ui.TextElement("打开 Popup"), ui.OnClick(func(ctx *ui.Context) {
							popupOpen.Set(true)
							addLog("P5：Popup open")
						})),
						ui.HSpacerElement(8),
						ui.FixedHeightElement(96, ui.ScrollViewElement(ui.ColumnElement(scrollRows...), ui.ScrollVertical(true), ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
							addLog(fmt.Sprintf("P5：ScrollView 默认滚动 y=%.0f", y))
						}))),
					),
				),
			)
		})

		return section(th, "P5 组件默认行为统一迁移", "打开阻止开关后，父容器 capture listener 会取消 click 或 wheel 默认行为，用来验证核心组件是否经由事件系统表达。", ui.StackElement(
			ui.ColumnElement(
				ui.RowElement(
					toggleButton("阻止 click 默认="+boolText(blockClick.Value()), func(ctx *ui.Context) {
						blockClick.Set(!blockClick.Value())
					}),
					ui.HSpacerElement(8),
					toggleButton("阻止 wheel 默认="+boolText(blockWheel.Value()), func(ctx *ui.Context) {
						blockWheel.Set(!blockWheel.Value())
					}),
				),
				ui.VSpacerElement(10),
				zone,
			),
			ui.DialogElement(
				dialogOpen.Value(),
				ui.TextElement("点击弹窗内部按钮或空白区域不应误关闭；Escape 和遮罩关闭会走 Dialog 默认事件边界。", ui.TextColor(th.Colors.OnSurfaceVariant)),
				ui.DialogHeadlineElement(ui.TextElement("Dialog 事件边界测试")),
				ui.DialogWidth(420),
				ui.DialogMaskClosable(true),
				ui.DialogGlobalOverlay(true),
				ui.DialogActions(
					ui.TextButton(ui.Text("关闭"), ui.OnClick(func(ctx *ui.Context) {
						dialogOpen.Set(false)
						addLog("P5：Dialog 内部按钮关闭")
					})),
				),
				ui.DialogOnOpenChange(func(ctx *ui.Context, open bool) {
					dialogOpen.Set(open)
					addLog(fmt.Sprintf("P5：Dialog openChange=%t", open))
				}),
			),
			ui.PopupElement(
				popupOpen.Value(),
				ui.ColumnElement(
					ui.TextElement("Popup 事件边界测试", ui.TextSize(16), ui.TextColor(th.Colors.OnSurface)),
					ui.VSpacerElement(8),
					ui.TextElement("点击内部空白不应误关闭；点击遮罩可关闭。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
					ui.VSpacerElement(10),
					ui.OutlinedButtonElement(ui.TextElement("关闭 Popup"), ui.OnClick(func(ctx *ui.Context) {
						popupOpen.Set(false)
						addLog("P5：Popup 内部按钮关闭")
					})),
				),
				ui.PopupWidth(360),
				ui.PopupMaskClosable(true),
				ui.PopupOnOpenChange(func(ctx *ui.Context, open bool) {
					popupOpen.Set(open)
					addLog(fmt.Sprintf("P5：Popup openChange=%t", open))
				}),
			),
		))
	})
}

func dragDropPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		drop := ui.UseState(ctx, dropState{Text: "把文本、JSON 或外部文件拖到这里"})
		syntheticType := ui.EventDragOver

		syntheticTarget := ui.ComponentElement(func(targetCtx *ui.Context) ui.Element {
			targetID := targetCtx.PathID()
			ui.OnDrag(targetCtx, syntheticType, func(ctx *ui.Context, ev *ui.DragEvent) {
				ev.PreventDefault()
				addLog(fmt.Sprintf("P5/P6：typed dragover MIME=%s text=%q prevented=%t", ev.MIMEType, ev.Text, ev.DefaultPrevented))
			})
			return ui.RowElement(
				ui.OutlinedButtonElement(ui.TextElement("合成 dragover"), ui.OnClick(func(ctx *ui.Context) {
					allowed := ui.DispatchDragEvent(ctx, targetID, &ui.DragEvent{
						Event:    ui.Event{Type: ui.EventDragOver},
						MIMEType: "application/json",
						Text:     `{"synthetic":true}`,
						Types:    []string{"application/json", "text/plain"},
					})
					addLog(fmt.Sprintf("P5/P6：DispatchDragEvent defaultAllowed=%t", allowed))
				})),
				ui.HSpacerElement(8),
				ui.TextElement("typed dragover 会 PreventDefault，因此返回 false。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			)
		})

		source := ui.DragSourceElement(
			miniBox("拖拽 JSON 载荷", "DragSource: dragstart / drag / dragend"),
			ui.DragSourcePayloads(
				ui.DragPayload{Type: "application/json", Data: []byte(`{"kind":"event-system-testbench"}`)},
				ui.DragPayload{Type: "text/plain", Data: []byte("FluxUI 事件系统测试载荷")},
			),
			ui.DragSourceOperations(ui.DragOperationCopy, ui.DragOperationMove),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				addLog(fmt.Sprintf("P5：DragSource event=%s op=%s type=%s err=%v", event.Kind, event.Operation, event.Type, event.Err))
			}),
		)

		targetColor := ui.NRGBA(255, 251, 235, 255)
		if drop.Value().Active {
			targetColor = ui.NRGBA(254, 243, 199, 255)
		}
		target := ui.DropTargetElement(
			ui.ContainerDecorationElement(
				ui.Bg(targetColor).WithPad(ui.All(14)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(245, 158, 11, 255)}),
				ui.FixedHeightElement(92, ui.CenterElement(ui.TextElement(drop.Value().Text, ui.TextSize(13), ui.TextColor(ui.NRGBA(120, 53, 15, 255))))),
			),
			func(ctx *ui.Context, event ui.DropEvent) {
				next := drop.Value()
				next.Active = false
				next.Text = fmt.Sprintf("%s bytes=%d op=%s", event.Type, len(event.Data), event.Operation)
				if event.Text != "" {
					next.Text = trimText(event.Text, 80)
				}
				if len(event.Paths) > 0 {
					next.Text = strings.Join(event.Paths, ", ")
				}
				drop.Set(next)
				addLog("P5：drop " + next.Text)
			},
			ui.DropTargetTypes("application/json", "application/x-fluxui-doc", "text/plain", "text/plain;charset=utf-8", "text/uri-list"),
			ui.DropTargetOperation(ui.DragOperationCopy),
			ui.DropTargetOnActiveChange(func(ctx *ui.Context, event ui.DropTargetStateEvent) {
				next := drop.Value()
				next.Active = event.Active
				drop.Set(next)
				addLog(fmt.Sprintf("P5：DropTarget active=%t types=%s", event.Active, strings.Join(event.Types, ",")))
			}),
			ui.DropTargetOnError(func(ctx *ui.Context, event ui.DropEvent) {
				if event.Err != nil {
					addLog("P5：DropTarget error " + event.Err.Error())
				} else {
					addLog("P5：DropTarget error")
				}
			}),
		)

		return section(th, "P5 DragSource / DropTarget 和 drag* 事件流", "测试真实拖放组件，以及 typed DragEvent 的合成派发和 PreventDefault。", ui.ColumnElement(
			ui.RowElement(ui.ExpandedElement(source), ui.HSpacerElement(12), ui.ExpandedElement(target)),
			ui.VSpacerElement(10),
			syntheticTarget,
		))
	})
}

func architecturePanel(th *ui.Theme, addLog func(string)) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		eventType := ui.EventType("event-test:architecture")
		ownerID := ctx.PathID()

		ui.OnEvent(ctx, eventType, func(ctx *ui.Context, ev *ui.Event) {
			addLog(fmt.Sprintf("P6：owner capture detail=%v path=%d", ev.Detail, len(ev.ComposedPath())))
		}, ui.Capture())
		ui.OnEvent(ctx, eventType, func(ctx *ui.Context, ev *ui.Event) {
			addLog(fmt.Sprintf("P6：owner bubble detail=%v default=%t", ev.Detail, ev.DefaultPrevented))
		})
		ui.OnActivate(ctx, func(ctx *ui.Context, ev *ui.ActivationEvent) {
			addLog(fmt.Sprintf("P6：activation source=%s keyEquivalent=%s", ev.Source, ev.KeyboardEquivalent))
		})

		return section(th, "P6 自定义事件、Portal 和边界策略", "验证 Detail payload、指定 target 派发、普通冒泡、modal stop boundary、redirect boundary、portal owner path 和 activation event。", ui.ColumnElement(
			ui.RowElement(
				ui.ExpandedElement(architectureProbe(th, "普通子树", "应冒泡到 owner", eventType, addLog, nil)),
				ui.HSpacerElement(10),
				ui.ExpandedElement(ui.EventBoundaryElement(
					architectureProbe(th, "Stop Boundary", "应被边界截断，owner 不应收到", eventType, addLog, nil),
					ui.EventBoundaryStopPropagation(),
				)),
			),
			ui.VSpacerElement(10),
			ui.RowElement(
				ui.ExpandedElement(ui.EventBoundaryElement(
					architectureProbe(th, "Redirect Boundary", "应重定向回 owner", eventType, addLog, nil),
					ui.EventBoundaryRedirectTo(ownerID),
				)),
				ui.HSpacerElement(10),
				ui.ExpandedElement(ui.EventPortalElement(
					architectureProbe(th, "EventPortal", "视觉位置不重要，逻辑路径到 owner", eventType, addLog, nil),
					ownerID,
				)),
			),
			ui.VSpacerElement(10),
			ui.RowElement(
				ui.OutlinedButtonElement(ui.TextElement("派发 activation event"), ui.OnClick(func(ctx *ui.Context) {
					allowed := ui.DispatchActivationEvent(ctx, ownerID, &ui.ActivationEvent{
						Source:             ui.ActivationSourceProgrammatic,
						KeyboardEquivalent: "Enter/Space",
					})
					addLog(fmt.Sprintf("P6：DispatchActivationEvent defaultAllowed=%t", allowed))
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("DispatchEvent 指定 owner"), ui.OnClick(func(ctx *ui.Context) {
					ev := ui.Event{Type: eventType, Bubbles: true, Cancelable: true, Detail: "直接指定 owner target"}
					allowed := ui.DispatchEvent(ctx, ownerID, ev)
					addLog(fmt.Sprintf("P6：DispatchEvent(owner) defaultAllowed=%t", allowed))
				})),
			),
		))
	})
}

func architectureProbe(th *ui.Theme, title, subtitle string, eventType ui.EventType, addLog func(string), extra ui.Element) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		targetID := ctx.PathID()
		ui.OnEvent(ctx, eventType, func(ctx *ui.Context, ev *ui.Event) {
			addLog(fmt.Sprintf("P6：%s target listener phase=%s detail=%v", title, phaseLabel(ev.Phase), ev.Detail))
		})
		children := []ui.Element{
			ui.TextElement(title, ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			ui.TextElement(subtitle, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.OutlinedButtonElement(ui.TextElement("派发自定义事件"), ui.OnClick(func(ctx *ui.Context) {
				allowed := ui.DispatchCustomEvent(ctx, targetID, eventType, map[string]string{"来源": title}, ui.CustomCancelable(true))
				addLog(fmt.Sprintf("P6：%s dispatch defaultAllowed=%t", title, allowed))
			})),
		}
		if extra != nil {
			children = append(children, ui.VSpacerElement(8), extra)
		}
		return ui.ContainerDecorationElement(
			panelBg(th).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.FixedHeightElement(116, ui.ColumnElement(children...)),
		)
	})
}

func diagnosticsPanel(th *ui.Theme, addLog func(string)) ui.Element {
	return section(th, "P7 性能、诊断和文档教学入口", "本 example 在 main 中启用了 EnablePerfDiagnostics(true) 和 LogEvents(true)。GUI 日志显示业务侧观察结果，控制台输出底层事件路径、listener 耗时、取消状态和默认行为结果。", ui.ColumnElement(
		ui.TextElement("测试建议：", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
		ui.TextElement("• 在 Pointer 区域高速拖动，观察 coalesced move 与控制台事件诊断。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.TextElement("• 在 P1 区域开启 PreventDefault / StopPropagation，对比日志顺序和 defaultAllowed。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.TextElement("• 在 P5 区域开启阻止 click/wheel，验证核心组件默认行为是否可取消。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.VSpacerElement(8),
		ui.OutlinedButtonElement(ui.TextElement("写入一条诊断标记"), ui.OnClick(func(ctx *ui.Context) {
			addLog("P7：手动诊断标记，控制台应同时有 click 事件诊断")
		})),
	))
}

func logPanel(th *ui.Theme, logs []string) ui.Element {
	items := []ui.Element{
		ui.TextElement("事件日志", ui.TextSize(16), ui.TextColor(th.Colors.OnSurface)),
	}
	for _, item := range logs {
		items = append(items, ui.PaddingElement(
			ui.Insets{Top: 4},
			ui.TextElement(item, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
		))
	}
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(12)).WithRad(8),
		ui.FixedHeightElement(210, ui.ScrollViewElement(ui.ColumnElement(items...), ui.ScrollVertical(true))),
	)
}

func section(th *ui.Theme, title, subtitle string, child ui.Element) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(14)).WithRad(10).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(17), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(4),
			ui.TextElement(subtitle, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(12),
			child,
		),
	)
}

func miniBox(title, subtitle string) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(ui.NRGBA(239, 246, 255, 255)).WithPad(ui.All(14)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(37, 99, 235, 255)}),
		ui.FixedHeightElement(92, ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(13), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
			ui.VSpacerElement(6),
			ui.TextElement(subtitle, ui.TextSize(11), ui.TextColor(ui.NRGBA(37, 99, 235, 255))),
		)),
	)
}

func toggleButton(label string, fn func(ctx *ui.Context)) ui.Element {
	return ui.OutlinedButtonElement(ui.TextElement(label), ui.OnClick(fn))
}

func panelBg(th *ui.Theme) ui.Decoration {
	return ui.Bg(th.Colors.SurfaceContainer)
}

func prependLog(logs benchState[[]string], message string) {
	next := append([]string{message}, logs.Value()...)
	if len(next) > 28 {
		next = next[:28]
	}
	logs.Set(next)
}

func rejectNonDigit(ev *ui.InputEvent) bool {
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

func boolText(v bool) string {
	if v {
		return "开"
	}
	return "关"
}

func modifierText(mods ui.Modifiers) string {
	parts := make([]string, 0, 5)
	if mods.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if mods.Shift {
		parts = append(parts, "Shift")
	}
	if mods.Alt {
		parts = append(parts, "Alt")
	}
	if mods.Meta {
		parts = append(parts, "Meta")
	}
	if mods.Shortcut {
		parts = append(parts, "Shortcut")
	}
	if len(parts) == 0 {
		return "无"
	}
	return strings.Join(parts, "+")
}

func phaseLabel(phase ui.EventPhase) string {
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

func trimText(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len([]rune(text)) <= limit {
		return text
	}
	runes := []rune(text)
	return string(runes[:limit]) + "..."
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func main() {
	_ = ui.RunElement(
		app,
		ui.Title("FluxUI 事件系统全量测试台"),
		ui.Size(1180, 860),
		ui.EnablePerfDiagnostics(true),
		ui.LogEvents(true),
	)
}
