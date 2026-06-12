package widget

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"

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
