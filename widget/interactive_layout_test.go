package widget

import (
	"image"
	"image/color"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func newInteractiveLayoutTestContext(rt *internal.Runtime) *internal.Context {
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: gioLayout.Exact(image.Pt(800, 600)),
	}
	return internal.NewContext(gtx, rt)
}

type interactionFrameHarness struct {
	rt     *internal.Runtime
	ops    op.Ops
	router input.Router
	now    time.Time
	size   image.Point
	moves  int
}

func newInteractionFrameHarness(size image.Point) *interactionFrameHarness {
	return &interactionFrameHarness{
		rt:   internal.NewRuntime(nil),
		now:  time.Unix(0, 0),
		size: size,
	}
}

func (h *interactionFrameHarness) layout(w Widget) internal.InteractionFrameStats {
	h.ops.Reset()
	if pos, ok := h.rt.CoalescedPointerMove(); ok {
		h.moves++
		h.router.Queue(pointer.Event{
			Kind:      pointer.Move,
			Source:    pointer.Mouse,
			PointerID: 1,
			Time:      time.Duration(h.moves) * time.Millisecond,
			Position:  f32.Pt(float32(pos.X), float32(pos.Y)),
		})
	}
	gtx := gioLayout.Context{
		Ops:         &h.ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         h.now,
		Constraints: gioLayout.Exact(h.size),
		Source:      h.router.Source(),
	}
	h.rt.BeginFrame()
	w.Layout(internal.NewContext(gtx, h.rt))
	h.rt.EndFrame()
	h.router.Frame(&h.ops)
	h.now = h.now.Add(16 * time.Millisecond)
	return h.rt.LastInteractionStats()
}

func (h *interactionFrameHarness) move(x, y int) {
	h.rt.QueuePointerMove(image.Pt(x, y))
}

func (h *interactionFrameHarness) key(ev key.Event) {
	h.router.Queue(ev)
}

func (h *interactionFrameHarness) pointer(ev pointer.Event) {
	h.router.Queue(ev)
}

func (h *interactionFrameHarness) click(w Widget, x, y int) {
	h.layout(w)
	h.pointer(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(float32(x), float32(y)),
	})
	h.layout(w)
	h.pointer(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(float32(x), float32(y)),
	})
	h.layout(w)
}

type captureContextWidget struct {
	child Widget
	ctx   **internal.Context
}

func (w captureContextWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.ctx != nil {
		*w.ctx = ctx
	}
	if w.child == nil {
		return layout.Dimensions{}
	}
	return w.child.Layout(ctx)
}

type preventClickDefaultWidget struct {
	child Widget
}

func (w preventClickDefaultWidget) Layout(ctx *internal.Context) layout.Dimensions {
	fluxevent.On(ctx, fluxevent.Click, func(_ *internal.Context, ev *internal.Event) {
		ev.PreventDefault()
	})
	if w.child == nil {
		return layout.Dimensions{}
	}
	return w.child.Layout(ctx.Child(0))
}

type testFocusTargetWidget struct {
	id   *internal.PathID
	size image.Point
}

func (w testFocusTargetWidget) Layout(ctx *internal.Context) layout.Dimensions {
	fluxevent.RegisterFocusTarget(ctx)
	if w.id != nil {
		*w.id = ctx.PathID()
	}
	size := w.size
	if size.X <= 0 || size.Y <= 0 {
		size = image.Pt(48, 32)
	}
	return layout.Dimensions{Size: size}
}

func TestHoverCallbacksAreChangeOnlyForSameTargetMoves(t *testing.T) {
	hoverCalls := 0
	hovering := false
	w := Button(Text("Hover"), OnHover(func(_ *internal.Context, value bool) {
		hoverCalls++
		hovering = value
	}))
	h := newInteractionFrameHarness(image.Pt(240, 120))

	h.layout(w)
	h.move(20, 20)
	stats := h.layout(w)
	if hoverCalls != 1 || !hovering {
		t.Fatalf("first hover move calls=%d hovering=%t, want one enter", hoverCalls, hovering)
	}
	if stats.HoverChanged != 1 {
		t.Fatalf("first hover target changes = %d, want 1", stats.HoverChanged)
	}

	h.move(30, 22)
	stats = h.layout(w)
	if hoverCalls != 1 {
		t.Fatalf("same target move calls=%d, want unchanged", hoverCalls)
	}
	if stats.HoverChanged != 0 {
		t.Fatalf("same target hover changes = %d, want 0", stats.HoverChanged)
	}
}

func TestBlankAreaMovesDoNotTriggerHoverTarget(t *testing.T) {
	hoverCalls := 0
	w := Button(Text("Hover"), OnHover(func(_ *internal.Context, _ bool) {
		hoverCalls++
	}))
	h := newInteractionFrameHarness(image.Pt(240, 120))

	h.layout(w)
	h.move(220, 100)
	stats := h.layout(w)
	if hoverCalls != 0 {
		t.Fatalf("blank area hover calls=%d, want 0", hoverCalls)
	}
	if stats.HoverChanged != 0 || stats.HoverTarget != "" {
		t.Fatalf("blank area stats=%+v, want no hover target change", stats)
	}
}

func TestPressableHitRectUsesChildSizeUnderExactParent(t *testing.T) {
	calls := 0
	w := Pressable(Spacer(40, 30), func(*internal.Context) {
		calls++
	})
	h := newInteractionFrameHarness(image.Pt(160, 120))

	h.click(w, 100, 80)
	if calls != 0 {
		t.Fatalf("pressable outside child calls = %d, want 0", calls)
	}

	h.click(w, 20, 20)
	if calls != 1 {
		t.Fatalf("pressable inside child calls = %d, want 1", calls)
	}
}

func TestPointerMoveCoalescingUsesLatestPosition(t *testing.T) {
	hoverCalls := 0
	w := Button(Text("Hover"), OnHover(func(_ *internal.Context, _ bool) {
		hoverCalls++
	}))
	h := newInteractionFrameHarness(image.Pt(240, 120))

	h.layout(w)
	h.move(20, 20)
	h.move(220, 100)
	stats := h.layout(w)
	if hoverCalls != 0 {
		t.Fatalf("coalesced final blank move hover calls=%d, want 0", hoverCalls)
	}
	if stats.PointerMoves != 1 {
		t.Fatalf("coalesced pointer moves=%d, want 1", stats.PointerMoves)
	}
	if stats.HoverChanged != 0 || stats.HoverTarget != "" {
		t.Fatalf("coalesced stats=%+v, want latest blank target", stats)
	}
}

func TestContainerDecorationOnHoverIsChangeOnly(t *testing.T) {
	hoverCalls := 0
	w := ContainerDecoration(
		style.Decoration{}.
			WithBg(color.NRGBA{R: 20, G: 30, B: 40, A: 255}).
			WithHover(style.Decoration{}.WithBg(color.NRGBA{R: 40, G: 50, B: 60, A: 255})).
			WithPad(style.All(8)),
		Text("Hover"),
		ContainerDecorationOnHover(func(_ *internal.Context, _ bool) {
			hoverCalls++
		}),
	)
	h := newInteractionFrameHarness(image.Pt(240, 120))

	h.layout(w)
	h.move(20, 20)
	h.layout(w)
	h.move(30, 22)
	h.layout(w)
	if hoverCalls != 1 {
		t.Fatalf("container hover calls=%d, want one change-only callback", hoverCalls)
	}
}

func TestContainerDecorationMarginAndSiblingBlankDoNotClick(t *testing.T) {
	calls := 0
	deco := style.Decoration{}.
		WithBg(color.NRGBA{R: 20, G: 30, B: 40, A: 255}).
		WithPad(style.All(10)).
		WithMargin(style.All(20)).
		WithShadow(style.BoxShadow{OffsetX: 0, OffsetY: 0, Blur: 16, Color: color.NRGBA{A: 128}})
	w := Row(
		ContainerDecoration(deco, Spacer(40, 20), ContainerDecorationOnClick(func(*internal.Context) {
			calls++
		})),
		Spacer(100, 80),
	)
	h := newInteractionFrameHarness(image.Pt(240, 120))

	h.click(w, 10, 30)
	if calls != 0 {
		t.Fatalf("container margin click calls = %d, want 0", calls)
	}

	h.click(w, 90, 30)
	if calls != 0 {
		t.Fatalf("container right margin/shadow click calls = %d, want 0", calls)
	}

	h.click(w, 25, 25)
	if calls != 1 {
		t.Fatalf("container surface click calls = %d, want 1", calls)
	}
}

func TestCardRippleHitStaysInsideSurface(t *testing.T) {
	calls := 0
	w := Row(
		ElevatedCard(Spacer(40, 30), CardOnClick(func(*internal.Context) {
			calls++
		})),
		Spacer(100, 80),
	)
	h := newInteractionFrameHarness(image.Pt(240, 120))

	h.click(w, 70, 20)
	if calls != 0 {
		t.Fatalf("card shadow/sibling click calls = %d, want 0", calls)
	}

	h.click(w, 20, 20)
	if calls != 1 {
		t.Fatalf("card surface click calls = %d, want 1", calls)
	}
}

func TestTabsPointerClickDoesNotEscapeItemRow(t *testing.T) {
	active := "one"
	w := Tabs(active,
		[]TabItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}},
		TabsOnChange(func(_ *internal.Context, key string) {
			active = key
		}),
	)
	h := newInteractionFrameHarness(image.Pt(240, 120))

	h.click(w, 180, 100)
	if active != "one" {
		t.Fatalf("tabs active after row-external click = %q, want one", active)
	}

	h.click(w, 180, 24)
	if active != "two" {
		t.Fatalf("tabs active after item click = %q, want two", active)
	}
}

func TestClickablePreventDefaultBlocksPointerDefaultActions(t *testing.T) {
	t.Run("button legacy OnClick", func(t *testing.T) {
		calls := 0
		w := preventClickDefaultWidget{child: Button(Text("Button"), OnClick(func(*internal.Context) {
			calls++
		}))}
		newInteractionFrameHarness(image.Pt(240, 120)).click(w, 24, 20)
		if calls != 0 {
			t.Fatalf("button OnClick calls = %d, want 0", calls)
		}
	})

	t.Run("pressable legacy onClick", func(t *testing.T) {
		calls := 0
		w := preventClickDefaultWidget{child: Pressable(Spacer(100, 48), func(*internal.Context) {
			calls++
		})}
		newInteractionFrameHarness(image.Pt(240, 120)).click(w, 24, 20)
		if calls != 0 {
			t.Fatalf("pressable onClick calls = %d, want 0", calls)
		}
	})

	t.Run("checkbox onChange", func(t *testing.T) {
		calls := 0
		w := preventClickDefaultWidget{child: Checkbox("Check", false, CheckboxOnChange(func(*internal.Context, bool) {
			calls++
		}))}
		newInteractionFrameHarness(image.Pt(240, 120)).click(w, 16, 16)
		if calls != 0 {
			t.Fatalf("checkbox onChange calls = %d, want 0", calls)
		}
	})

	t.Run("switch onChange", func(t *testing.T) {
		calls := 0
		w := preventClickDefaultWidget{child: Switch(false, SwitchOnChange(func(*internal.Context, bool) {
			calls++
		}))}
		newInteractionFrameHarness(image.Pt(240, 120)).click(w, 20, 16)
		if calls != 0 {
			t.Fatalf("switch onChange calls = %d, want 0", calls)
		}
	})

	t.Run("radio group onChange", func(t *testing.T) {
		calls := 0
		w := preventClickDefaultWidget{child: RadioGroup("a",
			[]RadioItem{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
			RadioGroupOnChange(func(*internal.Context, string) {
				calls++
			}),
		)}
		newInteractionFrameHarness(image.Pt(240, 160)).click(w, 16, 64)
		if calls != 0 {
			t.Fatalf("radio onChange calls = %d, want 0", calls)
		}
	})

	t.Run("select trigger open", func(t *testing.T) {
		openCalls := 0
		w := preventClickDefaultWidget{child: Select("one",
			[]SelectOptionItem[string]{{Label: "One", Value: "one"}, {Label: "Two", Value: "two"}},
			SelectOnOpenChange[string](func(*internal.Context, bool) {
				openCalls++
			}),
		)}
		newInteractionFrameHarness(image.Pt(320, 180)).click(w, 24, 24)
		if openCalls != 0 {
			t.Fatalf("select open calls = %d, want 0", openCalls)
		}
	})

	t.Run("select option onChange and close", func(t *testing.T) {
		ref := NewSelectRef[string]()
		ref.Open()
		changeCalls := 0
		closeCalls := 0
		w := preventClickDefaultWidget{child: Select("one",
			[]SelectOptionItem[string]{{Label: "One", Value: "one"}, {Label: "Two", Value: "two"}},
			SelectAttachRef(ref),
			SelectOnChange(func(*internal.Context, string) {
				changeCalls++
			}),
			SelectOnOpenChange[string](func(_ *internal.Context, opened bool) {
				if !opened {
					closeCalls++
				}
			}),
		)}
		newInteractionFrameHarness(image.Pt(320, 220)).click(w, 24, 80)
		if changeCalls != 0 {
			t.Fatalf("select option change calls = %d, want 0", changeCalls)
		}
		if closeCalls != 0 {
			t.Fatalf("select option close calls = %d, want 0", closeCalls)
		}
	})

	t.Run("tabs onChange", func(t *testing.T) {
		calls := 0
		w := preventClickDefaultWidget{child: Tabs("one",
			[]TabItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}},
			TabsOnChange(func(*internal.Context, string) {
				calls++
			}),
		)}
		newInteractionFrameHarness(image.Pt(240, 120)).click(w, 180, 24)
		if calls != 0 {
			t.Fatalf("tabs onChange calls = %d, want 0", calls)
		}
	})
}

func TestTabsPointerClickStillActivatesWithoutPreventDefault(t *testing.T) {
	active := "one"
	w := Tabs(active,
		[]TabItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}},
		TabsOnChange(func(_ *internal.Context, key string) {
			active = key
		}),
	)
	newInteractionFrameHarness(image.Pt(240, 120)).click(w, 180, 24)
	if active != "two" {
		t.Fatalf("tabs active after click = %q, want two", active)
	}
}

func TestSelectionControlsDoNotReportUnchangedCurrentItem(t *testing.T) {
	t.Run("select current option activation", func(t *testing.T) {
		changeCalls := 0
		closeCalls := 0
		current := "one"
		state := &selectState{opened: true}
		w := &selectWidget[string]{
			value: "one",
			options: []SelectOptionItem[string]{
				{Label: "One", Value: "one"},
				{Label: "Two", Value: "two"},
			},
			config: selectConfig[string]{
				onChange: func(*internal.Context, string) {
					changeCalls++
				},
				onOpen: func(_ *internal.Context, opened bool) {
					if !opened {
						closeCalls++
					}
				},
			},
		}

		w.activateOptionValue(nil, state, nil, &current, "one")
		if changeCalls != 0 {
			t.Fatalf("select current option change calls = %d, want 0", changeCalls)
		}
		if closeCalls != 1 {
			t.Fatalf("select current option close calls = %d, want 1", closeCalls)
		}
		if current != "one" || state.opened {
			t.Fatalf("select current=%q opened=%t, want one and closed", current, state.opened)
		}

		state.opened = true
		w.activateOptionValue(nil, state, nil, &current, "two")
		if changeCalls != 1 {
			t.Fatalf("select changed option change calls = %d, want 1", changeCalls)
		}
		if closeCalls != 2 {
			t.Fatalf("select changed option close calls = %d, want 2", closeCalls)
		}
		if current != "two" || state.opened {
			t.Fatalf("select current=%q opened=%t, want two and closed", current, state.opened)
		}
	})

	t.Run("select current option click", func(t *testing.T) {
		ref := NewSelectRef[string]()
		ref.Open()
		changeCalls := 0
		w := Select("one",
			[]SelectOptionItem[string]{{Label: "One", Value: "one"}, {Label: "Two", Value: "two"}},
			SelectAttachRef(ref),
			SelectOnChange(func(*internal.Context, string) {
				changeCalls++
			}),
		)
		newInteractionFrameHarness(image.Pt(320, 220)).click(w, 24, 80)
		if changeCalls != 0 {
			t.Fatalf("select current option change calls = %d, want 0", changeCalls)
		}
	})

	t.Run("tabs current tab", func(t *testing.T) {
		calls := 0
		active := "one"
		w := Tabs(active,
			[]TabItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}},
			TabsOnChange(func(_ *internal.Context, key string) {
				calls++
				active = key
			}),
		)
		newInteractionFrameHarness(image.Pt(240, 120)).click(w, 24, 24)
		if calls != 0 {
			t.Fatalf("tabs current tab change calls = %d, want 0", calls)
		}
		if active != "one" {
			t.Fatalf("tabs active after current click = %q, want one", active)
		}
	})
}

func TestSameFrameRefCommandsAccumulate(t *testing.T) {
	t.Run("checkbox toggles", func(t *testing.T) {
		ref := NewCheckboxRef()
		ref.Toggle()
		ref.Toggle()
		var changes []bool

		rt := internal.NewRuntime(nil)
		rt.BeginFrame()
		ctx := newInteractiveLayoutTestContext(rt)
		Checkbox("ok", false,
			CheckboxAttachRef(ref),
			CheckboxOnChange(func(_ *internal.Context, checked bool) {
				changes = append(changes, checked)
			}),
		).Layout(ctx)
		rt.EndFrame()

		if len(changes) != 2 || changes[0] != true || changes[1] != false {
			t.Fatalf("checkbox ref changes = %v, want [true false]", changes)
		}
	})

	t.Run("switch toggles", func(t *testing.T) {
		ref := NewSwitchRef()
		ref.Toggle()
		ref.Toggle()
		var changes []bool

		rt := internal.NewRuntime(nil)
		rt.BeginFrame()
		ctx := newInteractiveLayoutTestContext(rt)
		Switch(false,
			SwitchAttachRef(ref),
			SwitchOnChange(func(_ *internal.Context, checked bool) {
				changes = append(changes, checked)
			}),
		).Layout(ctx)
		rt.EndFrame()

		if len(changes) != 2 || changes[0] != true || changes[1] != false {
			t.Fatalf("switch ref changes = %v, want [true false]", changes)
		}
	})

	t.Run("slider set and steps", func(t *testing.T) {
		ref := NewSliderRef()
		ref.SetValue(10)
		ref.StepBy(5)
		ref.StepBy(-2)
		changeCalls := 0

		rt := internal.NewRuntime(nil)
		rt.BeginFrame()
		ctx := newInteractiveLayoutTestContext(rt).Scope("slider")
		Slider(0,
			SliderAttachRef(ref),
			SliderOnChange(func(*internal.Context, float32) {
				changeCalls++
			}),
		).Layout(ctx)
		value, ok := ctx.PersistentValueKey(internal.MemoryKey{Path: ctx.PathID(), Namespace: "slider", Slot: 0})
		rt.EndFrame()
		if !ok {
			t.Fatal("slider state was not recorded")
		}
		state, ok := value.(*sliderState)
		if !ok {
			t.Fatalf("slider state type = %T, want *sliderState", value)
		}

		got := applySliderStep(100*state.end, 0, 100, 1)
		if got != 13 {
			t.Fatalf("slider ref value = %v, want 13", got)
		}
		if changeCalls != 0 {
			t.Fatalf("slider ref change calls = %d, want 0", changeCalls)
		}
	})
}

func TestPressableKeyboardActivationUsesCancelableClick(t *testing.T) {
	calls := 0
	var pressableCtx *internal.Context
	w := preventClickDefaultWidget{child: captureContextWidget{
		ctx: &pressableCtx,
		child: Pressable(Spacer(100, 48), func(*internal.Context) {
			calls++
		}),
	}}
	h := newInteractionFrameHarness(image.Pt(240, 120))
	h.layout(w)
	if pressableCtx == nil {
		t.Fatal("pressable context was not captured")
	}
	if !h.rt.RequestFocus(pressableCtx, pressableCtx.PathID()) {
		t.Fatal("failed to focus pressable")
	}
	h.layout(w)
	h.key(key.Event{Name: key.NameEnter, State: key.Press})
	h.layout(w)
	if calls != 0 {
		t.Fatalf("pressable keyboard activation calls = %d, want 0 after click PreventDefault", calls)
	}
}

func TestSliderAndSwitchCanReusePathWithoutHookCountPanic(t *testing.T) {
	rt := internal.NewRuntime(nil)

	rt.BeginFrame()
	_ = Slider(40).Layout(newInteractiveLayoutTestContext(rt))
	rt.EndFrame()

	rt.BeginFrame()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected switching from Slider to Switch at same path not to panic, got %v", r)
		}
	}()
	_ = Switch(true).Layout(newInteractiveLayoutTestContext(rt))
	rt.EndFrame()
}

func TestDefaultInteractiveWidgetsDoNotAddDecorationPadding(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)

	if got, want := Switch(true).Layout(ctx).Size, image.Pt(52, 32); got != want {
		t.Fatalf("default switch size = %v, want %v", got, want)
	}

	rt.EndFrame()
	rt.BeginFrame()
	ctx = newInteractiveLayoutTestContext(rt)

	if got, want := Slider(40).Layout(ctx).Size, image.Pt(200, 40); got != want {
		t.Fatalf("default slider size = %v, want %v", got, want)
	}

	rt.EndFrame()
}

func TestSwitchWithIconsPreservesDefaultSize(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)
	got := Switch(true,
		SwitchIcons("check", "close"),
		SwitchIconFontFamily("Material Symbols Outlined"),
	).Layout(ctx).Size
	rt.EndFrame()

	if want := image.Pt(52, 32); got != want {
		t.Fatalf("switch with icons size = %v, want %v", got, want)
	}
}

func TestSwitchThumbCenterTracksAnimatedProgress(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)
	size := image.Pt(52, 32)

	off := switchThumbCenter(ctx, size, 0)
	on := switchThumbCenter(ctx, size, 1)
	rt.EndFrame()

	if off.X >= on.X {
		t.Fatalf("switch thumb center off=%v on=%v, want left-to-right motion", off, on)
	}
	if off != image.Pt(16, 16) || on != image.Pt(36, 16) {
		t.Fatalf("switch thumb centers off=%v on=%v, want (16,16) and (36,16)", off, on)
	}
}

func TestWindowDragAreaPreservesChildSize(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(800, 600)}
	dims := WindowDragArea(Spacer(180, 32)).Layout(ctx)
	rt.EndFrame()

	if got, want := dims.Size, image.Pt(180, 32); got != want {
		t.Fatalf("WindowDragArea size = %v, want %v", got, want)
	}
}

func TestWindowDragAreaRegistersMoveAction(t *testing.T) {
	var ops op.Ops
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	gtx := gioLayout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: gioLayout.Exact(image.Pt(800, 600)),
	}
	ctx := internal.NewContext(gtx, rt)
	WindowDragArea(Spacer(180, 32)).Layout(ctx)
	rt.EndFrame()

	var router input.Router
	router.Frame(&ops)
	action, ok := router.ActionAt(f32.Pt(90, 16))
	if !ok {
		t.Fatal("expected WindowDragArea to register a system action")
	}
	if action != system.ActionMove {
		t.Fatalf("WindowDragArea action = %v, want %v", action, system.ActionMove)
	}
	if !rt.WindowDragAreaActive() {
		t.Fatal("expected WindowDragArea to mark the frame as containing a native drag area")
	}
}

func TestWindowDragAreaDisabledDoesNotRegisterNativeDragArea(t *testing.T) {
	var ops op.Ops
	rt := internal.NewRuntime(nil)
	gtx := gioLayout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: gioLayout.Exact(image.Pt(800, 600)),
	}
	w := WindowDragArea(Spacer(180, 32), WindowDragAreaDisabled(true))

	layoutWindowDragAreaTestFrame(rt, gtx, w)

	if rt.WindowDragAreaActive() {
		t.Fatal("disabled WindowDragArea should not mark a native drag area")
	}
}

func layoutWindowDragAreaTestFrame(rt *internal.Runtime, gtx gioLayout.Context, w Widget) {
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	w.Layout(ctx)
	rt.EndFrame()
}

func TestMD3SelectionNavigationAndOverlayDefaultsLayout(t *testing.T) {
	rt := internal.NewRuntime(nil)

	cases := []struct {
		name string
		w    Widget
	}{
		{name: "checkbox disabled", w: Checkbox("Check", true, CheckboxDisabled(true))},
		{name: "switch disabled", w: Switch(true, SwitchDisabled(true))},
		{name: "slider disabled", w: Slider(40, SliderDisabled(true))},
		{name: "radio disabled", w: RadioGroup("a", []RadioItem{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}}, RadioGroupDisabled(true))},
		{name: "appbar", w: AppBar(Text("Title"))},
		{name: "bottom navigation", w: BottomNavigation("home", []NavItem{{Key: "home", Label: "Home", Icon: Text("H")}, {Key: "settings", Label: "Settings", Icon: Text("S")}})},
		{name: "tabs", w: Tabs("one", []TabItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}})},
		{name: "dialog", w: Dialog(true, Text("Dialog"), DialogTitle("Title"))},
		{name: "popup", w: Popup(true, Text("Popup"), PopupPadding(style.All(12)))},
		{name: "toast", w: Toast("Toast", ToastDuration(0))},
	}

	for _, tc := range cases {
		rt.BeginFrame()
		dims := tc.w.Layout(newInteractiveLayoutTestContext(rt))
		rt.EndFrame()
		if dims.Size.X <= 0 || dims.Size.Y <= 0 {
			t.Fatalf("%s returned empty size %v", tc.name, dims.Size)
		}
	}
}

func TestModalInternalPressDoesNotCloseMaskClosableOverlay(t *testing.T) {
	tests := []struct {
		name string
		root func(onOpenChange func(*internal.Context, bool)) Widget
	}{
		{
			name: "popup",
			root: func(onOpenChange func(*internal.Context, bool)) Widget {
				return Popup(
					true,
					Spacer(260, 180),
					PopupWidth(320),
					PopupHeight(220),
					PopupQuick(true),
					PopupMaskClosable(true),
					PopupOnOpenChange(onOpenChange),
				)
			},
		},
		{
			name: "dialog",
			root: func(onOpenChange func(*internal.Context, bool)) Widget {
				return Dialog(
					true,
					Spacer(240, 120),
					DialogWidth(320),
					DialogHeight(220),
					DialogQuick(true),
					DialogMaskClosable(true),
					DialogOnOpenChange(onOpenChange),
				)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rt := internal.NewRuntime(nil)
			defer rt.Dispose()

			var router input.Router
			var ops op.Ops
			now := time.Unix(0, 0)
			closeCalls := 0
			w := tc.root(func(_ *internal.Context, open bool) {
				if !open {
					closeCalls++
				}
			})

			renderPointer := func(frame int, events ...pointer.Event) {
				for _, ev := range events {
					router.Queue(ev)
				}
				ops.Reset()
				gtx := gioLayout.Context{
					Ops:         &ops,
					Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
					Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
					Constraints: gioLayout.Exact(image.Pt(640, 480)),
					Source:      router.Source(),
				}
				rt.BeginFrame()
				w.Layout(internal.NewContext(gtx, rt))
				rt.EndFrame()
				router.Frame(&ops)
			}

			renderPointer(0)
			renderPointer(1, pointer.Event{
				Kind:      pointer.Press,
				Source:    pointer.Mouse,
				PointerID: 1,
				Buttons:   pointer.ButtonPrimary,
				Position:  f32.Pt(320, 240),
			})
			renderPointer(2, pointer.Event{
				Kind:      pointer.Release,
				Source:    pointer.Mouse,
				PointerID: 1,
				Position:  f32.Pt(320, 240),
			})
			if closeCalls != 0 {
				t.Fatalf("internal press closed %s, closeCalls=%d", tc.name, closeCalls)
			}

			renderPointer(3, pointer.Event{
				Kind:      pointer.Press,
				Source:    pointer.Mouse,
				PointerID: 2,
				Buttons:   pointer.ButtonPrimary,
				Position:  f32.Pt(20, 20),
			})
			renderPointer(4, pointer.Event{
				Kind:      pointer.Release,
				Source:    pointer.Mouse,
				PointerID: 2,
				Position:  f32.Pt(20, 20),
			})
			if closeCalls == 0 {
				t.Fatalf("outside mask press did not close %s", tc.name)
			}
		})
	}
}

func TestDialogFocusTrapEscapeAndRestoreFocus(t *testing.T) {
	h := newInteractionFrameHarness(image.Pt(640, 480))
	open := false
	closeCalls := 0
	var outsideID, firstID, secondID internal.PathID

	render := func() Widget {
		return Stack(
			testFocusTargetWidget{id: &outsideID, size: image.Pt(80, 40)},
			Dialog(
				open,
				Row(
					testFocusTargetWidget{id: &firstID, size: image.Pt(80, 40)},
					testFocusTargetWidget{id: &secondID, size: image.Pt(80, 40)},
				),
				DialogQuick(true),
				DialogOnOpenChange(func(_ *internal.Context, next bool) {
					open = next
					if !next {
						closeCalls++
					}
				}),
			),
		)
	}

	h.layout(render())
	if !h.rt.RequestFocus(nil, outsideID) {
		t.Fatal("failed to focus outside target before opening dialog")
	}
	open = true
	h.layout(render())
	if got := h.rt.FocusedTarget(); got != firstID {
		t.Fatalf("dialog initial focus = %v, want first panel target %v", got, firstID)
	}

	h.key(key.Event{Name: key.NameTab, State: key.Press})
	h.layout(render())
	if got := h.rt.FocusedTarget(); got != secondID {
		t.Fatalf("dialog Tab focus = %v, want second panel target %v", got, secondID)
	}

	h.key(key.Event{Name: key.NameTab, State: key.Press})
	h.layout(render())
	if got := h.rt.FocusedTarget(); got != firstID {
		t.Fatalf("dialog Tab wrap focus = %v, want first panel target %v", got, firstID)
	}

	h.key(key.Event{Name: key.NameTab, State: key.Press, Modifiers: key.ModShift})
	h.layout(render())
	if got := h.rt.FocusedTarget(); got != secondID {
		t.Fatalf("dialog Shift+Tab wrap focus = %v, want second panel target %v", got, secondID)
	}

	h.key(key.Event{Name: key.NameEscape, State: key.Press})
	h.layout(render())
	if open {
		t.Fatal("dialog Escape did not request close")
	}
	if closeCalls != 1 {
		t.Fatalf("dialog close calls = %d, want 1", closeCalls)
	}
	h.layout(render())
	if got := h.rt.FocusedTarget(); got != outsideID {
		t.Fatalf("dialog restored focus = %v, want outside target %v", got, outsideID)
	}
}

func TestSelectEscapeClosesAndRestoresFieldFocus(t *testing.T) {
	h := newInteractionFrameHarness(image.Pt(320, 240))
	ref := NewSelectRef[string]()
	options := []SelectOptionItem[string]{
		{Label: "One", Value: "one"},
		{Label: "Two", Value: "two"},
	}
	var selectCtx *internal.Context
	openChanges := []bool{}
	render := func() Widget {
		return captureContextWidget{
			ctx: &selectCtx,
			child: Select(
				"one",
				options,
				SelectAttachRef(ref),
				SelectOnOpenChange[string](func(_ *internal.Context, opened bool) {
					openChanges = append(openChanges, opened)
				}),
			),
		}
	}

	ref.Open()
	h.layout(render())
	fieldID := selectCtx.Child(0).PathID()
	if !h.rt.RequestFocus(selectCtx, fieldID) {
		t.Fatal("failed to focus select field")
	}
	h.layout(render())
	h.key(key.Event{Name: key.NameEscape, State: key.Press})
	h.layout(render())

	if len(openChanges) == 0 || openChanges[len(openChanges)-1] {
		t.Fatalf("select open changes = %v, want final false", openChanges)
	}
	if got := h.rt.FocusedTarget(); got != fieldID {
		t.Fatalf("select Escape restored focus = %v, want field %v", got, fieldID)
	}
}

func TestDropdownMenuEscapeClosesAndRestoresTriggerFocus(t *testing.T) {
	h := newInteractionFrameHarness(image.Pt(320, 240))
	open := true
	closeCalls := 0
	var menuCtx *internal.Context
	render := func() Widget {
		return captureContextWidget{
			ctx: &menuCtx,
			child: DropdownMenu(
				open,
				Text("Menu"),
				[]MenuItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}},
				DropdownMenuOnOpenChange(func(_ *internal.Context, next bool) {
					open = next
					if !next {
						closeCalls++
					}
				}),
			),
		}
	}

	h.layout(render())
	triggerID := menuCtx.PathID()
	if !h.rt.RequestFocus(menuCtx, triggerID) {
		t.Fatal("failed to focus dropdown trigger")
	}
	h.layout(render())
	h.key(key.Event{Name: key.NameEscape, State: key.Press})
	h.layout(render())

	if open {
		t.Fatal("dropdown Escape did not request close")
	}
	if closeCalls != 1 {
		t.Fatalf("dropdown close calls = %d, want 1", closeCalls)
	}
	if got := h.rt.FocusedTarget(); got != triggerID {
		t.Fatalf("dropdown Escape restored focus = %v, want trigger %v", got, triggerID)
	}
}

func TestMD3PopupPlacementFlipsWhenBelowSpaceIsInsufficient(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(200, 90)}
	ctx = ctx.WithViewport(image.Rect(0, 0, 200, 90))
	ctx = ctx.WithPositionOffset(image.Pt(0, 60))

	placement := md3PopupVerticalPlacement(ctx, 48, 80, 6)
	rt.EndFrame()

	if placement.Direction != md3PopupUp {
		t.Fatalf("popup direction = %v, want up", placement.Direction)
	}
	if placement.TransitionOffset >= 0 {
		t.Fatalf("popup transition offset = %v, want negative for upward popup", placement.TransitionOffset)
	}
	if got, want := md3PopupOffsetY(48, 80, placement), -86; got != want {
		t.Fatalf("popup y offset = %d, want %d", got, want)
	}
}

func TestMD3PopupPlacementStaysDownWhenBelowSpaceFits(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(200, 200)}
	ctx = ctx.WithViewport(image.Rect(0, 0, 200, 200))

	placement := md3PopupVerticalPlacement(ctx, 48, 80, 6)
	rt.EndFrame()

	if placement.Direction != md3PopupDown {
		t.Fatalf("popup direction = %v, want down", placement.Direction)
	}
	if placement.TransitionOffset <= 0 {
		t.Fatalf("popup transition offset = %v, want positive for downward popup", placement.TransitionOffset)
	}
	if got, want := md3PopupOffsetY(48, 80, placement), 54; got != want {
		t.Fatalf("popup y offset = %d, want %d", got, want)
	}
}

func TestMD3PopupMeasuredPlacementFlipsInsideViewport(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(240, 160)}
	ctx = ctx.WithViewport(image.Rect(0, 0, 240, 160))
	ctx = ctx.WithPositionOffset(image.Pt(0, 70))

	placement := md3PopupPlacementForAnchor(ctx, image.Pt(120, 40), image.Pt(120, 70), 40, 4)
	if placement.Direction != md3PopupDown {
		t.Fatalf("initial popup direction = %v, want down", placement.Direction)
	}
	placement = md3PopupPlacementForMeasuredPopup(ctx, image.Pt(120, 40), image.Pt(120, 90), placement, 0)
	rt.EndFrame()

	if placement.Direction != md3PopupUp {
		t.Fatalf("measured popup direction = %v, want up", placement.Direction)
	}
	if placement.TransitionOffset >= 0 {
		t.Fatalf("measured popup transition offset = %v, want negative for upward popup", placement.TransitionOffset)
	}
	if got, want := placement.MaxHeightPx, 40; got != want {
		t.Fatalf("measured popup max height = %d, want %d", got, want)
	}
}

func TestDropdownAndSelectPopupLayoutWithConstrainedSpace(t *testing.T) {
	items := []MenuItem{
		{Key: "one", Label: "One"},
		{Key: "two", Label: "Two"},
		{Key: "three", Label: "Three"},
		{Key: "four", Label: "Four"},
	}

	rt := internal.NewRuntime(nil)
	ctx := newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(180, 90)}

	rt.BeginFrame()
	dropdownDims := DropdownMenu(true, Text("Menu"), items, DropdownMenuMaxHeight(180)).Layout(ctx)
	rt.EndFrame()
	if dropdownDims.Size.X <= 0 || dropdownDims.Size.Y <= 0 {
		t.Fatalf("dropdown constrained layout returned empty size %v", dropdownDims.Size)
	}

	selectRef := NewSelectRef[string]()
	selectRef.Open()
	options := []SelectOptionItem[string]{
		{Label: "One", Value: "one"},
		{Label: "Two", Value: "two"},
		{Label: "Three", Value: "three"},
		{Label: "Four", Value: "four"},
	}

	rt = internal.NewRuntime(nil)
	ctx = newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(180, 90)}

	rt.BeginFrame()
	selectDims := Select("one", options, SelectAttachRef(selectRef), SelectMaxHeight[string](180)).Layout(ctx)
	rt.EndFrame()
	if selectDims.Size.X <= 0 || selectDims.Size.Y <= 0 {
		t.Fatalf("select constrained layout returned empty size %v", selectDims.Size)
	}
}

func TestDropdownAndSelectPopupLayoutInsideScrollView(t *testing.T) {
	items := []MenuItem{
		{Key: "one", Label: "One"},
		{Key: "two", Label: "Two"},
		{Key: "three", Label: "Three"},
		{Key: "four", Label: "Four"},
	}

	scrollRef := NewScrollRef()
	scrollRef.ScrollToOffset(120)
	rt := internal.NewRuntime(nil)
	ctx := newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(220, 120)}
	rt.BeginFrame()
	dims := ScrollView(
		Column(
			Spacer(1, 180),
			DropdownMenu(true, Text("Menu"), items, DropdownMenuMaxHeight(180)),
			Spacer(1, 180),
		),
		ScrollAttachRef(scrollRef),
	).Layout(ctx)
	rt.EndFrame()
	if dims.Size.X <= 0 || dims.Size.Y <= 0 {
		t.Fatalf("scroll dropdown layout returned empty size %v", dims.Size)
	}

	selectRef := NewSelectRef[string]()
	selectRef.Open()
	options := []SelectOptionItem[string]{
		{Label: "One", Value: "one"},
		{Label: "Two", Value: "two"},
		{Label: "Three", Value: "three"},
		{Label: "Four", Value: "four"},
	}
	scrollRef = NewScrollRef()
	scrollRef.ScrollToOffset(120)
	rt = internal.NewRuntime(nil)
	ctx = newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(220, 120)}
	rt.BeginFrame()
	dims = ScrollView(
		Column(
			Spacer(1, 180),
			Select("one", options, SelectAttachRef(selectRef), SelectMaxHeight[string](180)),
			Spacer(1, 180),
		),
		ScrollAttachRef(scrollRef),
	).Layout(ctx)
	rt.EndFrame()
	if dims.Size.X <= 0 || dims.Size.Y <= 0 {
		t.Fatalf("scroll select layout returned empty size %v", dims.Size)
	}
}
