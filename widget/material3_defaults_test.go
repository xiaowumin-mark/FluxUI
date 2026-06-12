package widget

import (
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"
	"github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestButtonMD3VariantDefaults(t *testing.T) {
	th := theme.New(theme.LightColors())

	filled := resolveButtonDefaults(buttonVariantFilled, th)
	if filled.background != th.Colors.Primary || filled.foreground != th.Colors.OnPrimary {
		t.Fatalf("filled button defaults = %#v, want primary/onPrimary", filled)
	}
	if filled.radius != th.Shapes.Full {
		t.Fatalf("filled button radius = %v, want %v", filled.radius, th.Shapes.Full)
	}
	if filled.text != th.Types.LabelLarge {
		t.Fatalf("filled button text = %#v, want LabelLarge", filled.text)
	}

	outlined := resolveButtonDefaults(buttonVariantOutlined, th)
	if outlined.background != (color.NRGBA{}) {
		t.Fatalf("outlined button background = %#v, want transparent", outlined.background)
	}
	if outlined.border.Color != th.Colors.Outline || outlined.border.Width != 1 {
		t.Fatalf("outlined button border = %#v, want outline width 1", outlined.border)
	}
	if outlined.disabledContainer {
		t.Fatalf("outlined button disabled container should remain transparent")
	}

	elevated := resolveButtonDefaults(buttonVariantElevated, th)
	if elevated.background != style.SurfaceAtElevation(th.Colors, 1) {
		t.Fatalf("elevated button background = %#v, want tonal elevation", elevated.background)
	}
	if elevated.shadow.IsZero() {
		t.Fatalf("elevated button shadow should be set")
	}
}

func TestButtonRadiusAllowsExplicitZero(t *testing.T) {
	var cfg buttonConfig
	ButtonRadius(0)(&cfg)
	if !cfg.hasRadius {
		t.Fatalf("ButtonRadius should mark radius as explicitly configured")
	}
	if cfg.radius != 0 {
		t.Fatalf("ButtonRadius(0) radius = %v, want 0", cfg.radius)
	}
}

func TestInputMD3VariantDefaults(t *testing.T) {
	th := theme.New(theme.LightColors())

	outlined := resolveInputDefaults(inputVariantOutlined, th)
	if outlined.background != th.Colors.Surface {
		t.Fatalf("outlined input background = %#v, want surface", outlined.background)
	}
	if outlined.border != th.Colors.Outline || outlined.borderWidth != 1 {
		t.Fatalf("outlined input border = %#v width %v, want outline width 1", outlined.border, outlined.borderWidth)
	}
	if outlined.placeholder != th.Colors.OnSurfaceVariant {
		t.Fatalf("outlined input placeholder = %#v, want onSurfaceVariant", outlined.placeholder)
	}
	if outlined.text != th.Types.BodyLarge {
		t.Fatalf("outlined input text = %#v, want BodyLarge", outlined.text)
	}

	filled := resolveInputDefaults(inputVariantFilled, th)
	if filled.background != th.Colors.SurfaceContainerHighest {
		t.Fatalf("filled input background = %#v, want surfaceContainerHighest", filled.background)
	}
	if filled.border.A != 0 || filled.borderWidth != 0 {
		t.Fatalf("filled input border = %#v width %v, want none", filled.border, filled.borderWidth)
	}
}

func TestSelectTriggerKeepsMD3Height(t *testing.T) {
	rt := internal.NewRuntime(theme.New(theme.LightColors()))
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops: &ops,
		Constraints: gioLayout.Constraints{
			Max: image.Pt(230, 600),
		},
	}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)

	dims := Select("medium", []SelectOptionItem[string]{
		{Label: "Low priority", Value: "low"},
		{Label: "Medium priority", Value: "medium"},
		{Label: "High priority", Value: "high"},
	}).Layout(ctx)

	wantHeight := gtx.Dp(safeDp(56))
	if dims.Size.Y != wantHeight {
		t.Fatalf("select trigger height = %d, want %d", dims.Size.Y, wantHeight)
	}
}

func TestCompactDensityShrinksCoreDefaults(t *testing.T) {
	th := theme.New(theme.LightColors())
	th.SetDensity(theme.CompactDensityScale())
	rt := internal.NewRuntime(th)
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops: &ops,
		Constraints: gioLayout.Constraints{
			Max: image.Pt(230, 600),
		},
	}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	selectDims := Select("medium", []SelectOptionItem[string]{
		{Label: "Low priority", Value: "low"},
		{Label: "Medium priority", Value: "medium"},
	}).Layout(ctx)
	rt.EndFrame()

	wantSelectHeight := gtx.Dp(safeDp(48))
	if selectDims.Size.Y != wantSelectHeight {
		t.Fatalf("compact select trigger height = %d, want %d", selectDims.Size.Y, wantSelectHeight)
	}

	rt.BeginFrame()
	ctx = internal.NewContext(gtx, rt)
	drawerDims := (&navigationDrawerWidget{active: "home"}).drawerItem(NavItem{Key: "home", Label: "Home"}).Layout(ctx)
	rt.EndFrame()

	wantDrawerItemHeight := gtx.Dp(safeDp(40))
	if drawerDims.Size.Y != wantDrawerItemHeight {
		t.Fatalf("compact drawer item height = %d, want %d", drawerDims.Size.Y, wantDrawerItemHeight)
	}
}

func TestDisabledSelectionRefsDoNotDispatch(t *testing.T) {
	rt := internal.NewRuntime(theme.New(theme.LightColors()))
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops: &ops,
		Constraints: gioLayout.Constraints{
			Max: image.Pt(230, 600),
		},
	}

	radioRef := NewRadioGroupRef()
	radioRef.SetValue("b")
	radioChanges := 0
	rt.BeginFrame()
	RadioGroup("a", []RadioItem{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
		RadioGroupDisabled(true),
		RadioGroupAttachRef(radioRef),
		RadioGroupOnChange(func(ctx *internal.Context, value string) { radioChanges++ }),
	).Layout(internal.NewContext(gtx, rt))
	rt.EndFrame()
	if radioChanges != 0 {
		t.Fatalf("disabled radio ref dispatched %d changes, want 0", radioChanges)
	}

	selectRef := NewSelectRef[string]()
	selectRef.Open()
	selectRef.SetValue("b")
	selectChanges := 0
	openChanges := 0
	rt.BeginFrame()
	Select("a", []SelectOptionItem[string]{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
		SelectDisabled[string](true),
		SelectAttachRef(selectRef),
		SelectOnChange(func(ctx *internal.Context, value string) { selectChanges++ }),
		SelectOnOpenChange[string](func(ctx *internal.Context, opened bool) { openChanges++ }),
	).Layout(internal.NewContext(gtx, rt))
	rt.EndFrame()
	if selectChanges != 0 || openChanges != 0 {
		t.Fatalf("disabled select ref dispatched change=%d open=%d, want 0/0", selectChanges, openChanges)
	}
}

func TestCardMD3VariantDefaults(t *testing.T) {
	th := theme.New(theme.LightColors())

	filled := resolveCardDefaults(cardVariantFilled, th)
	if filled.background != th.Colors.SurfaceContainerHighest {
		t.Fatalf("filled card background = %#v, want surfaceContainerHighest", filled.background)
	}
	if filled.radius != th.Shapes.Medium {
		t.Fatalf("filled card radius = %v, want %v", filled.radius, th.Shapes.Medium)
	}

	outlined := resolveCardDefaults(cardVariantOutlined, th)
	if outlined.background != th.Colors.Surface {
		t.Fatalf("outlined card background = %#v, want surface", outlined.background)
	}
	if outlined.border.Color != th.Colors.OutlineVariant || outlined.border.Width != 1 {
		t.Fatalf("outlined card border = %#v, want outlineVariant width 1", outlined.border)
	}

	elevated := resolveCardDefaults(cardVariantElevated, th)
	if elevated.background != style.SurfaceAtElevation(th.Colors, 1) {
		t.Fatalf("elevated card background = %#v, want tonal elevation", elevated.background)
	}
	if elevated.shadow.IsZero() {
		t.Fatalf("elevated card shadow should be set")
	}
}

func TestClickableCardDoesNotUseButtonWrapper(t *testing.T) {
	data, err := os.ReadFile("media_card.go")
	if err != nil {
		t.Fatalf("read media_card.go: %v", err)
	}
	text := string(data)
	start := strings.Index(text, "func (c *cardWidget) Layout")
	end := strings.Index(text, "type cardDefaults struct")
	if start < 0 || end <= start {
		t.Fatal("could not locate cardWidget.Layout source")
	}
	cardLayout := text[start:end]
	if strings.Contains(cardLayout, "Button(") || strings.Contains(cardLayout, "ButtonAttachRef") {
		t.Fatal("clickable Card should register card-bounded input directly, not wrap the surface in Button")
	}
	if !strings.Contains(cardLayout, "LayoutRippleOverlayArea") {
		t.Fatal("clickable Card should keep pressed feedback bounded to the card surface")
	}
}
