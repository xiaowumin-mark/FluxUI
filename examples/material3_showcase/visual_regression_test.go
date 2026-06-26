//go:build visual

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/xiaowumin-mark/FluxUI/icons/md3"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"
	ui "github.com/xiaowumin-mark/FluxUI/ui"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type visualViewport struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type visualTheme struct {
	Name  string    `json:"name"`
	Theme *ui.Theme `json:"-"`
}

type visualScreenshotResult struct {
	Name             string         `json:"name"`
	Category         string         `json:"category"`
	Theme            string         `json:"theme"`
	Viewport         visualViewport `json:"viewport"`
	Path             string         `json:"path"`
	NonZeroAlpha     int            `json:"non_zero_alpha"`
	Opaque           int            `json:"opaque"`
	QuantizedColors  int            `json:"quantized_colors"`
	LuminanceMinimum int            `json:"luminance_minimum"`
	LuminanceMaximum int            `json:"luminance_maximum"`
	GeneratedAt      string         `json:"generated_at"`
}

type visualCaptureSpec struct {
	Name     string
	Category string
	Theme    visualTheme
	Viewport visualViewport
	Root     ui.Component
	Events   []visualFrameEvents
}

type visualFrameEvents struct {
	Before []gioEvent.Event
	After  []gioEvent.Event
}

func TestMaterial3ShowcaseScreenshots(t *testing.T) {
	outputDir := visualOutputDir(t)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("create screenshot output dir: %v", err)
	}

	specs := visualCaptureSpecs()

	generatedAt := time.Now().UTC().Format(time.RFC3339)
	results := make([]visualScreenshotResult, 0, len(specs))
	for _, spec := range specs {
		path := filepath.Join(outputDir, spec.Name+".png")
		img := renderVisualScreenshot(t, spec)
		stats := assertScreenshotSmoke(t, spec.Name, img, spec.Viewport)
		writePNG(t, path, img)
		results = append(results, visualScreenshotResult{
			Name:             spec.Name,
			Category:         spec.Category,
			Theme:            spec.Theme.Name,
			Viewport:         spec.Viewport,
			Path:             path,
			NonZeroAlpha:     stats.nonZeroAlpha,
			Opaque:           stats.opaque,
			QuantizedColors:  stats.quantizedColors,
			LuminanceMinimum: stats.luminanceMin,
			LuminanceMaximum: stats.luminanceMax,
			GeneratedAt:      generatedAt,
		})
	}

	writeManifest(t, filepath.Join(outputDir, "manifest.json"), results)
	t.Logf("wrote %d Material 3 showcase screenshots to %s", len(results), outputDir)
}

func visualCaptureSpecs() []visualCaptureSpec {
	themes := []visualTheme{
		{Name: "light", Theme: ui.NewTheme(ui.LightColors())},
		{Name: "dark", Theme: ui.NewTheme(ui.DarkColors())},
	}
	fullViewports := []visualViewport{
		{Name: "desktop", Width: 1440, Height: 1000},
		{Name: "narrow", Width: 480, Height: 900},
	}

	specs := make([]visualCaptureSpec, 0, 32)
	for _, themeCase := range themes {
		for _, viewport := range fullViewports {
			specs = append(specs, visualCaptureSpec{
				Name:     fmt.Sprintf("showcase-%s-%s", themeCase.Name, viewport.Name),
				Category: "showcase",
				Theme:    themeCase,
				Viewport: viewport,
				Root:     App,
				Events:   steadyFrames(3),
			})
		}
	}

	for _, themeCase := range themes {
		specs = append(specs, componentRegionSpecs(themeCase)...)
		specs = append(specs, interactionStateSpecs(themeCase)...)
	}
	return specs
}

func componentRegionSpecs(themeCase visualTheme) []visualCaptureSpec {
	return []visualCaptureSpec{
		{
			Name:     fmt.Sprintf("region-%s-buttons", themeCase.Name),
			Category: "component-region",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "region-buttons", Width: 980, Height: 300},
			Root:     buttonsRegionRoot,
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("region-%s-inputs-cards", themeCase.Name),
			Category: "component-region",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "region-inputs-cards", Width: 980, Height: 360},
			Root:     inputsCardsRegionRoot,
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("region-%s-selection", themeCase.Name),
			Category: "component-region",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "region-selection", Width: 900, Height: 320},
			Root:     selectionRegionRoot,
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("region-%s-navigation", themeCase.Name),
			Category: "component-region",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "region-navigation", Width: 980, Height: 460},
			Root:     navigationRegionRoot,
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("region-%s-overlay", themeCase.Name),
			Category: "component-region",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "region-overlay", Width: 980, Height: 440},
			Root:     overlayRegionRoot,
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("region-%s-chips-progress", themeCase.Name),
			Category: "component-region",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "region-chips-progress", Width: 980, Height: 340},
			Root:     chipsProgressRegionRoot,
			Events:   steadyFrames(2),
		},
	}
}

func interactionStateSpecs(themeCase visualTheme) []visualCaptureSpec {
	return []visualCaptureSpec{
		{
			Name:     fmt.Sprintf("state-%s-hover", themeCase.Name),
			Category: "interaction-state",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "state-hover", Width: 760, Height: 260},
			Root:     pointerButtonStateRoot("Hover state"),
			Events: []visualFrameEvents{
				{},
				{Before: []gioEvent.Event{pointerMove(110, 100)}},
			},
		},
		{
			Name:     fmt.Sprintf("state-%s-pressed-ripple", themeCase.Name),
			Category: "interaction-state",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "state-pressed-ripple", Width: 760, Height: 260},
			Root:     pointerButtonStateRoot("Pressed + ripple state"),
			Events: []visualFrameEvents{
				{},
				{Before: []gioEvent.Event{pointerMove(110, 100), pointerPress(110, 100, 16*time.Millisecond)}},
				{},
			},
		},
		{
			Name:     fmt.Sprintf("state-%s-focus", themeCase.Name),
			Category: "interaction-state",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "state-focus", Width: 760, Height: 260},
			Root:     focusedInputStateRoot(),
			Events:   steadyFrames(3),
		},
		{
			Name:     fmt.Sprintf("state-%s-selected", themeCase.Name),
			Category: "interaction-state",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "state-selected", Width: 820, Height: 340},
			Root:     selectedStateRoot,
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("state-%s-expanded", themeCase.Name),
			Category: "interaction-state",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "state-expanded", Width: 760, Height: 440},
			Root:     expandedStateRoot,
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("state-%s-toast-enter", themeCase.Name),
			Category: "interaction-state",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "state-toast-enter", Width: 760, Height: 300},
			Root:     toastFrameStateRoot("Toast enter frame", 500*time.Millisecond),
			Events:   steadyFrames(2),
		},
		{
			Name:     fmt.Sprintf("state-%s-toast-exit", themeCase.Name),
			Category: "interaction-state",
			Theme:    themeCase,
			Viewport: visualViewport{Name: "state-toast-exit", Width: 760, Height: 300},
			Root:     toastFrameStateRoot("Toast exit frame", 150*time.Millisecond),
			Events: []visualFrameEvents{
				{},
				{},
				{},
				{},
				{},
				{},
				{},
				{},
				{},
				{},
				{},
				{},
			},
		},
	}
}

func steadyFrames(count int) []visualFrameEvents {
	if count <= 0 {
		count = 1
	}
	return make([]visualFrameEvents, count)
}

func pointerMove(x, y float32) pointer.Event {
	return pointer.Event{
		Kind:     pointer.Move,
		Source:   pointer.Mouse,
		Position: f32.Pt(x, y),
	}
}

func pointerPress(x, y float32, at time.Duration) pointer.Event {
	return pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Time:      at,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(x, y),
	}
}

func renderVisualScreenshot(t *testing.T, spec visualCaptureSpec) *image.RGBA {
	t.Helper()

	if spec.Root == nil {
		t.Fatalf("%s: visual capture root is nil", spec.Name)
	}

	window, err := headless.NewWindow(spec.Viewport.Width, spec.Viewport.Height)
	if err != nil {
		t.Fatalf("create headless window: %v", err)
	}
	defer window.Release()

	runtime := internal.NewRuntime(spec.Theme.Theme)
	defer runtime.Dispose()
	runtime.SetInvalidator(func() {})

	root := ui.VisualRootBuilder(spec.Root)
	baseTime := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	events := spec.Events
	if len(events) == 0 {
		events = steadyFrames(3)
	}

	var ops op.Ops
	var router input.Router
	for frame := range events {
		for _, ev := range events[frame].Before {
			router.Queue(ev)
		}

		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(spec.Viewport.Width, spec.Viewport.Height)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}

		runtime.BeginFrame()
		ctx := runtime.Frame(gtx)
		if widget := root(ctx.Scope("build")); widget != nil {
			widget.Layout(ctx.Scope("tree"))
		}
		runtime.EndFrame()

		router.Frame(&ops)
		if err := window.Frame(&ops); err != nil {
			t.Fatalf("render headless frame %d: %v", frame, err)
		}
		for _, ev := range events[frame].After {
			router.Queue(ev)
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, spec.Viewport.Width, spec.Viewport.Height))
	if err := window.Screenshot(img); err != nil {
		t.Fatalf("capture screenshot: %v", err)
	}
	return img
}

func buttonsRegionRoot(ctx *ui.Context) ui.Element {
	return fixtureShell(ctx, "Buttons and FABs",
		ui.RowElement(
			paddedFixture(ui.FilledButtonElement(ui.TextElement("Filled"))),
			paddedFixture(ui.FilledTonalButtonElement(ui.TextElement("Filled tonal"))),
			paddedFixture(ui.OutlinedButtonElement(ui.TextElement("Outlined"))),
			paddedFixture(ui.TextButtonElement(ui.TextElement("Text"))),
			paddedFixture(ui.ElevatedButtonElement(ui.TextElement("Elevated"))),
			paddedFixture(ui.FilledButtonElement(ui.TextElement("Disabled"), ui.Disabled(true))),
		),
		ui.SpacerElement(0, 18),
		ui.RowElement(
			paddedFixture(ui.IconButtonElement(ui.IconElement("search"), ui.IconButtonSelected(true))),
			paddedFixture(ui.FilledIconButtonElement(ui.IconElement("favorite"), ui.IconButtonSelected(true))),
			paddedFixture(ui.FilledTonalIconButtonElement(ui.IconElement("tune"))),
			paddedFixture(ui.OutlinedIconButtonElement(ui.IconElement("radio_button_unchecked"))),
			ui.SpacerElement(16, 0),
			paddedFixture(ui.SmallFloatingActionButtonElement(ui.IconElement("add"))),
			paddedFixture(ui.FloatingActionButtonElement(ui.IconElement("add"))),
			paddedFixture(ui.ExtendedFloatingActionButtonElement(ui.IconElement("add"), ui.TextElement("Create"))),
		),
	)
}

func inputsCardsRegionRoot(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	return fixtureShell(ctx, "TextField, Select, Cards",
		ui.RowElement(
			ui.FixedWidthElement(240, ui.OutlinedTextFieldElement("Outlined value", ui.InputPlaceholder("Outlined"))),
			ui.SpacerElement(14, 0),
			ui.FixedWidthElement(240, ui.FilledTextFieldElement("Filled value", ui.InputPlaceholder("Filled"))),
			ui.SpacerElement(14, 0),
			ui.FixedWidthElement(220, ui.SelectElement("medium", []ui.SelectOptionItem[string]{
				{Label: "Low priority", Value: "low"},
				{Label: "Medium priority", Value: "medium"},
				{Label: "High priority", Value: "high"},
			})),
		),
		ui.SpacerElement(0, 18),
		ui.RowElement(
			fixtureCard("Filled", th, ui.FilledCardElement),
			ui.SpacerElement(14, 0),
			fixtureCard("Elevated", th, ui.ElevatedCardElement),
			ui.SpacerElement(14, 0),
			fixtureCard("Outlined", th, ui.OutlinedCardElement),
		),
	)
}

func selectionRegionRoot(ctx *ui.Context) ui.Element {
	return fixtureShell(ctx, "Selection controls",
		ui.RowElement(
			ui.CheckboxElement("Checked", true),
			ui.SpacerElement(24, 0),
			ui.CheckboxElement("Unchecked", false),
			ui.SpacerElement(24, 0),
			ui.SwitchElement(true),
			ui.SpacerElement(24, 0),
			ui.SwitchElement(false),
		),
		ui.SpacerElement(0, 18),
		ui.RowElement(
			ui.FixedWidthElement(260, ui.SliderElement(62, ui.SliderMin(0), ui.SliderMax(100))),
			ui.SpacerElement(24, 0),
			ui.RadioGroupElement("b", []ui.RadioItem{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
				{Label: "C", Value: "c"},
			}, ui.RadioGroupDirection(ui.Horizontal)),
		),
	)
}

func navigationRegionRoot(ctx *ui.Context) ui.Element {
	navItems := []ui.ElementNavItem{
		{Key: "home", Label: "Home", Icon: ui.IconElement("home")},
		{Key: "search", Label: "Search", Icon: ui.IconElement("search")},
		{Key: "settings", Label: "Settings", Icon: ui.IconElement("settings")},
	}
	return fixtureShell(ctx, "Navigation",
		ui.TabsElement("components", []ui.TabItem{
			{Key: "overview", Label: "Overview"},
			{Key: "components", Label: "Components"},
			{Key: "tokens", Label: "Tokens"},
		}),
		ui.SpacerElement(0, 12),
		ui.BottomNavigationElement("search", navItems),
		ui.SpacerElement(0, 16),
		ui.FixedHeightElement(230, ui.RowElement(
			ui.NavigationRailElement("search", navItems),
			ui.SpacerElement(14, 0),
			ui.NavigationDrawerElement("settings", navItems, ui.NavigationDrawerWidth(260)),
		)),
	)
}

func overlayRegionRoot(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	return fixtureShell(ctx, "Overlay surfaces",
		ui.FixedHeightElement(320, ui.StackElement(
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(16)).WithRad(th.Shapes.Medium),
				ui.ColumnElement(
					ui.TextElement("Base content behind overlays", ui.TextType(th.Types.BodyLarge), ui.TextColor(th.Colors.OnSurface)),
					ui.SpacerElement(0, 12),
					ui.PopupElement(true,
						ui.TextElement("Popup surface", ui.TextColor(th.Colors.OnSurface)),
						ui.PopupPadding(ui.All(16)),
						ui.PopupWidth(280),
					),
					ui.ToastElement("Toast surface", ui.ToastDuration(0)),
					ui.SnackbarElement("Snackbar with action", ui.SnackbarAction("Undo", func(ctx *ui.Context) {}), ui.ToastDuration(0)),
				),
			),
			ui.DialogElement(true,
				ui.TextElement("Dialog content uses elevated surface tokens.", ui.TextColor(th.Colors.OnSurface)),
				ui.DialogTitle("Dialog"),
				ui.DialogWidth(380),
			),
		)),
	)
}

func chipsProgressRegionRoot(ctx *ui.Context) ui.Element {
	return fixtureShell(ctx, "Chips, Search, Progress",
		ui.RowElement(
			paddedFixture(ui.AssistChipElement("Assist")),
			paddedFixture(ui.FilterChipElement("Filter", ui.ChipSelected(true))),
			paddedFixture(ui.InputChipElement("Input")),
			paddedFixture(ui.SuggestionChipElement("Suggestion")),
		),
		ui.SpacerElement(0, 18),
		ui.RowElement(
			ui.FixedWidthElement(340, ui.SearchBarElement("material", ui.SearchBarPlaceholder("Search components"))),
			ui.SpacerElement(24, 0),
			ui.FixedWidthElement(260, ui.ColumnElement(
				ui.LinearProgressIndicatorElement(0.62),
				ui.SpacerElement(0, 18),
				ui.LinearProgressIndicatorElement(0.28),
			)),
			ui.SpacerElement(24, 0),
			ui.CircularProgressIndicatorElement(0.72, ui.ProgressSize(72), ui.ProgressLabelVisible(true)),
		),
	)
}

func pointerButtonStateRoot(title string) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return stateShell(ctx, title,
			ui.RowElement(
				ui.FilledButtonElement(ui.TextElement("Interactive target")),
				ui.SpacerElement(16, 0),
				stateSwatch(ctx, "Hover layer", style.StateLayer(th.Colors.PrimaryContainer, th.Colors.OnPrimaryContainer, style.StateLayerHoverOpacity), th.Colors.OnPrimaryContainer),
				ui.SpacerElement(12, 0),
				stateSwatch(ctx, "Pressed layer", style.StateLayer(th.Colors.PrimaryContainer, th.Colors.OnPrimaryContainer, style.StateLayerPressedOpacity), th.Colors.OnPrimaryContainer),
			),
		)
	}
}

func focusedInputStateRoot() ui.Component {
	ref := ui.NewInputRef()
	queued := false
	return func(ctx *ui.Context) ui.Element {
		if !queued {
			ref.Focus()
			queued = true
		}
		th := ui.UseTheme(ctx)
		return stateShell(ctx, "Focus state",
			ui.RowElement(
				ui.FixedWidthElement(320, ui.OutlinedTextFieldElement("Focused value", ui.InputAttachRef(ref), ui.InputBorderFocus(th.Colors.Primary))),
				ui.SpacerElement(18, 0),
				stateSwatch(ctx, "Focus ring", color.NRGBA{R: th.Colors.Primary.R, G: th.Colors.Primary.G, B: th.Colors.Primary.B, A: 32}, th.Colors.Primary),
			),
		)
	}
}

func selectedStateRoot(ctx *ui.Context) ui.Element {
	return stateShell(ctx, "Selected state",
		ui.TabsElement("components", []ui.TabItem{
			{Key: "overview", Label: "Overview"},
			{Key: "components", Label: "Components"},
			{Key: "tokens", Label: "Tokens"},
		}),
		ui.SpacerElement(0, 14),
		ui.BottomNavigationElement("search", []ui.ElementNavItem{
			{Key: "home", Label: "Home", Icon: ui.IconElement("home")},
			{Key: "search", Label: "Search", Icon: ui.IconElement("search")},
			{Key: "settings", Label: "Settings", Icon: ui.IconElement("settings")},
		}),
		ui.SpacerElement(0, 14),
		ui.RowElement(
			paddedFixture(ui.FilterChipElement("Selected chip", ui.ChipSelected(true))),
			paddedFixture(ui.FilterChipElement("Unselected chip")),
		),
	)
}

func expandedStateRoot(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	return stateShell(ctx, "Expanded state",
		ui.RowElement(
			ui.FixedWidthElement(260, ui.DropdownMenuElement(
				true,
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.Symmetric(10, 16)).WithRad(th.Shapes.ExtraSmall).WithBorder(ui.Border{Width: 1, Color: th.Colors.Outline}),
					ui.TextElement("Open menu", ui.TextColor(th.Colors.OnSurface)),
				),
				[]ui.MenuItem{
					{Key: "copy", Label: "Copy"},
					{Key: "share", Label: "Share"},
					{Key: "archive", Label: "Archive"},
				},
				ui.DropdownMenuSelectedKey("share"),
			)),
			ui.SpacerElement(24, 0),
			ui.FixedWidthElement(280, ui.SelectElement("medium", []ui.SelectOptionItem[string]{
				{Label: "Low priority", Value: "low"},
				{Label: "Medium priority", Value: "medium"},
				{Label: "High priority", Value: "high"},
			})),
		),
	)
}

func toastFrameStateRoot(title string, duration time.Duration) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return stateShell(ctx, title,
			ui.FixedHeightElement(180, ui.StackElement(
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(16)).WithRad(th.Shapes.Medium),
					ui.TextElement("Toast frame capture surface", ui.TextType(th.Types.BodyLarge), ui.TextColor(th.Colors.OnSurface)),
				),
				ui.ToastElement("Draft archived", ui.ToastDuration(duration)),
			)),
		)
	}
}

func fixtureShell(ctx *ui.Context, title string, children ...ui.Element) ui.Element {
	th := ui.UseTheme(ctx)
	content := make([]ui.Element, 0, len(children)+2)
	content = append(content,
		ui.TextElement(title, ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
		ui.SpacerElement(0, 12),
	)
	content = append(content, children...)
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.Surface).WithPad(ui.All(18)),
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(16)).WithRad(th.Shapes.Medium),
			ui.ColumnElement(content...),
		),
	)
}

func stateShell(ctx *ui.Context, title string, children ...ui.Element) ui.Element {
	th := ui.UseTheme(ctx)
	content := make([]ui.Element, 0, len(children)+3)
	content = append(content,
		ui.TextElement(title, ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
		ui.TextElement("Interaction keyframe specimen", ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.SpacerElement(0, 16),
	)
	content = append(content, children...)
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.Surface).WithPad(ui.All(24)),
		ui.ColumnElement(content...),
	)
}

func fixtureCard(title string, th *ui.Theme, factory func(ui.Element, ...ui.CardOption) ui.Element) ui.Element {
	return ui.FixedWidthElement(220, factory(
		ui.ColumnElement(
			ui.TextElement(title, ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
			ui.SpacerElement(0, 6),
			ui.TextElement("MD3 card variant", ui.TextType(th.Types.BodyMedium), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
	))
}

func stateSwatch(ctx *ui.Context, label string, bg color.NRGBA, fg color.NRGBA) ui.Element {
	th := ui.UseTheme(ctx)
	return ui.ContainerDecorationElement(
		ui.Bg(bg).WithPad(ui.Symmetric(10, 14)).WithRad(th.Shapes.ExtraSmall).WithBorder(ui.Border{Width: 1, Color: color.NRGBA{R: fg.R, G: fg.G, B: fg.B, A: 80}}),
		ui.TextElement(label, ui.TextType(th.Types.LabelLarge), ui.TextColor(fg)),
	)
}

func paddedFixture(el ui.Element) ui.Element {
	return ui.PaddingElement(ui.Insets{Right: 8, Bottom: 8}, el)
}

type visualSmokeStats struct {
	nonZeroAlpha    int
	opaque          int
	quantizedColors int
	luminanceMin    int
	luminanceMax    int
}

func assertScreenshotSmoke(t *testing.T, name string, img *image.RGBA, viewport visualViewport) visualSmokeStats {
	t.Helper()
	if img == nil {
		t.Fatalf("%s: screenshot image is nil", name)
	}
	bounds := img.Bounds()
	if bounds.Dx() != viewport.Width || bounds.Dy() != viewport.Height {
		t.Fatalf("%s: screenshot bounds = %v, want %dx%d", name, bounds, viewport.Width, viewport.Height)
	}

	stats := analyzeScreenshot(img)
	totalPixels := viewport.Width * viewport.Height
	if stats.nonZeroAlpha < totalPixels*95/100 {
		t.Fatalf("%s: only %d/%d pixels have alpha, screenshot looks blank or transparent", name, stats.nonZeroAlpha, totalPixels)
	}
	if stats.opaque < totalPixels*90/100 {
		t.Fatalf("%s: only %d/%d pixels are opaque, screenshot looks partially empty", name, stats.opaque, totalPixels)
	}
	if stats.quantizedColors < 24 {
		t.Fatalf("%s: only %d quantized colors, screenshot looks flat or blank", name, stats.quantizedColors)
	}
	if stats.luminanceMax-stats.luminanceMin < 24 {
		t.Fatalf("%s: luminance range %d-%d is too narrow, screenshot looks flat", name, stats.luminanceMin, stats.luminanceMax)
	}
	return stats
}

func analyzeScreenshot(img *image.RGBA) visualSmokeStats {
	bounds := img.Bounds()
	colors := make(map[uint32]struct{}, 128)
	stats := visualSmokeStats{luminanceMin: 255}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		offset := img.PixOffset(bounds.Min.X, y)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r := img.Pix[offset]
			g := img.Pix[offset+1]
			b := img.Pix[offset+2]
			a := img.Pix[offset+3]
			offset += 4

			if a > 0 {
				stats.nonZeroAlpha++
			}
			if a == 255 {
				stats.opaque++
			}
			luminance := (int(r)*299 + int(g)*587 + int(b)*114) / 1000
			if luminance < stats.luminanceMin {
				stats.luminanceMin = luminance
			}
			if luminance > stats.luminanceMax {
				stats.luminanceMax = luminance
			}
			key := uint32(r>>4)<<12 | uint32(g>>4)<<8 | uint32(b>>4)<<4 | uint32(a>>4)
			colors[key] = struct{}{}
		}
	}
	stats.quantizedColors = len(colors)
	return stats
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create screenshot %s: %v", path, err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode screenshot %s: %v", path, err)
	}
}

func writeManifest(t *testing.T, path string, results []visualScreenshotResult) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create manifest %s: %v", path, err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Command string                   `json:"command"`
		Results []visualScreenshotResult `json:"results"`
	}{
		Command: "go test -tags visual ./examples/material3_showcase -run TestMaterial3ShowcaseScreenshots -count=1",
		Results: results,
	}); err != nil {
		t.Fatalf("write manifest %s: %v", path, err)
	}
}

func visualOutputDir(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("FLUXUI_VISUAL_OUTPUT"); dir != "" {
		if !filepath.IsAbs(dir) {
			return filepath.Join(repoRoot(t), dir)
		}
		return dir
	}
	return filepath.Join(repoRoot(t), "out", "material3-screenshots")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repo root containing go.mod")
		}
		dir = parent
	}
}
