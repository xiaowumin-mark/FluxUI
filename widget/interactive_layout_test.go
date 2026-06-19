package widget

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/f32"
	"gioui.org/io/input"
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

	if got, want := Slider(40).Layout(ctx).Size, image.Pt(200, 20); got != want {
		t.Fatalf("default slider size = %v, want %v", got, want)
	}

	rt.EndFrame()
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
