package widget

import (
	"image"
	"image/color"
	"reflect"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
)

func TestR0SelectOptionsWireTheComponentContract(t *testing.T) {
	items := []SelectOptionItem[string]{{Label: "One", Value: "one"}}
	ref := NewSelectRef[string]()
	leading := Spacer(8, 8)
	trailing := Spacer(8, 8)
	decoration := style.Decoration{}.WithRad(7)
	menuDecoration := style.Decoration{}.WithRad(11)
	changes := 0
	openChanges := 0

	w, ok := FilledSelect("one", items,
		SelectPlaceholder[string]("Choose"),
		SelectLabel[string]("Priority"),
		SelectSupportingText[string]("Required for triage"),
		SelectErrorText[string]("Pick a value"),
		SelectError[string](true),
		SelectRequired[string](true),
		SelectNoAsterisk[string](true),
		SelectDisabled[string](true),
		SelectMaxHeight[string](144),
		SelectWidth[string](192),
		SelectXOffset[string](5),
		SelectYOffset[string](9),
		SelectLeading[string](leading),
		SelectTrailing[string](trailing),
		SelectOnChange(func(*internal.Context, string) { changes++ }),
		SelectOnOpenChange[string](func(*internal.Context, bool) { openChanges++ }),
		SelectAttachRef(ref),
		SelectDecoration[string](decoration),
		SelectMenuDecoration[string](menuDecoration),
	).(*selectWidget[string])
	if !ok {
		t.Fatalf("FilledSelect type = %T, want *selectWidget[string]", FilledSelect("one", items))
	}

	items[0].Label = "mutated"
	if got := w.options[0].Label; got != "One" {
		t.Fatalf("select options were not defensively copied: got %q", got)
	}
	if w.config.variant != selectVariantFilled || w.config.placeholder != "Choose" || w.config.label != "Priority" {
		t.Fatalf("select identity config = %#v", w.config)
	}
	if w.config.supportingText != "Required for triage" || w.config.errorText != "Pick a value" {
		t.Fatalf("select supporting/error config = %#v", w.config)
	}
	if !w.config.error || !w.config.required || !w.config.noAsterisk || !w.config.disabled {
		t.Fatalf("select boolean config = %#v", w.config)
	}
	if w.config.maxHeight != 144 || w.config.width != 192 || w.config.xOffset != 5 || w.config.yOffset != 9 {
		t.Fatalf("select geometry config = %#v", w.config)
	}
	if w.config.leading == nil || w.config.trailing == nil || w.config.ref != ref {
		t.Fatalf("select widget/ref config = %#v", w.config)
	}
	if !reflect.DeepEqual(w.config.decoration, decoration) || !reflect.DeepEqual(w.config.menuDecoration, menuDecoration) {
		t.Fatalf("select decorations were not retained: %#v", w.config)
	}
	w.config.onChange(nil, "two")
	w.config.onOpen(nil, true)
	if changes != 1 || openChanges != 1 {
		t.Fatalf("select callbacks were not wired: changes=%d open=%d", changes, openChanges)
	}

	outlined := OutlinedSelect("one", nil).(*selectWidget[string])
	if outlined.config.variant != selectVariantOutlined {
		t.Fatalf("OutlinedSelect variant = %v, want outlined", outlined.config.variant)
	}

	cfg := selectConfig[string]{variant: selectVariantFilled}
	SelectFilled[string](false)(&cfg)
	if cfg.variant != selectVariantOutlined {
		t.Fatalf("SelectFilled(false) variant = %v, want outlined", cfg.variant)
	}
	SelectFilled[string](true)(&cfg)
	if cfg.variant != selectVariantFilled {
		t.Fatalf("SelectFilled(true) variant = %v, want filled", cfg.variant)
	}
}

func TestR0MenuOptionsWireTheComponentContract(t *testing.T) {
	items := []MenuItem{{Key: "one", Label: "One"}}
	decoration := style.Decoration{}.WithRad(9)
	selected := ""

	w, ok := Menu(items,
		MenuSelectedKey("one"),
		MenuOnSelect(func(_ *internal.Context, key string) { selected = key }),
		MenuWidth(180),
		MenuMaxHeight(224),
		MenuDecoration(decoration),
		MenuQuick(true),
		MenuHasOverflow(true),
		MenuXOffset(3),
		MenuYOffset(7),
		MenuAnchorCorner(MenuCornerEndEnd),
		MenuMenuCorner(MenuCornerStartEnd),
		MenuDefaultFocusOf(MenuDefaultFocusLastItem),
		MenuPositioningOf(MenuPositioningFixed),
		MenuTypeaheadDelay(350*time.Millisecond),
		MenuNoHorizontalFlip(true),
		MenuNoVerticalFlip(true),
		MenuStayOpenOnOutsideClick(true),
		MenuStayOpenOnFocusout(true),
		MenuSkipRestoreFocus(true),
		MenuNoNavigationWrap(true),
		MenuHoverOpenDelay(125*time.Millisecond),
		MenuHoverCloseDelay(175*time.Millisecond),
	).(*menuWidget)
	if !ok {
		t.Fatalf("Menu type = %T, want *menuWidget", Menu(items))
	}

	items[0].Label = "mutated"
	if got := w.items[0].Label; got != "One" {
		t.Fatalf("menu items were not defensively copied: got %q", got)
	}
	cfg := w.config
	if cfg.selectedKey != "one" || cfg.width != 180 || cfg.maxHeight != 224 || !reflect.DeepEqual(cfg.decoration, decoration) {
		t.Fatalf("menu identity/geometry config = %#v", cfg)
	}
	if !cfg.quick || !cfg.hasOverflow || cfg.xOffset != 3 || cfg.yOffset != 7 {
		t.Fatalf("menu display config = %#v", cfg)
	}
	if cfg.anchorCorner != MenuCornerEndEnd || cfg.menuCorner != MenuCornerStartEnd || cfg.defaultFocus != MenuDefaultFocusLastItem || cfg.positioning != MenuPositioningFixed {
		t.Fatalf("menu placement config = %#v", cfg)
	}
	if cfg.typeaheadDelay != 350*time.Millisecond || !cfg.noHorizontalFlip || !cfg.noVerticalFlip || !cfg.stayOpenOnOutsideClick || !cfg.stayOpenOnFocusout || !cfg.skipRestoreFocus || !cfg.noNavigationWrap {
		t.Fatalf("menu interaction config = %#v", cfg)
	}
	if cfg.hoverOpenDelay != 125*time.Millisecond || cfg.hoverCloseDelay != 175*time.Millisecond {
		t.Fatalf("menu hover delay config = %#v", cfg)
	}
	cfg.onSelect(nil, "two")
	if selected != "two" {
		t.Fatalf("menu onSelect received %q, want two", selected)
	}
}

func TestR0DropdownMenuOptionsWireTheComponentContract(t *testing.T) {
	items := []MenuItem{{Key: "one", Label: "One"}}
	decoration := style.Decoration{}.WithRad(12)
	selected := ""
	openChanges := 0

	w, ok := DropdownMenu(true, Spacer(80, 32), items,
		DropdownMenuSelectedKey("one"),
		DropdownMenuOnSelect(func(_ *internal.Context, key string) { selected = key }),
		DropdownMenuOnOpenChange(func(*internal.Context, bool) { openChanges++ }),
		DropdownMenuWidth(180),
		DropdownMenuMaxHeight(224),
		DropdownMenuDecoration(decoration),
		DropdownMenuQuick(true),
		DropdownMenuHasOverflow(true),
		DropdownMenuXOffset(3),
		DropdownMenuYOffset(7),
		DropdownMenuAnchorCorner(MenuCornerEndEnd),
		DropdownMenuMenuCorner(MenuCornerStartEnd),
		DropdownMenuDefaultFocusOf(MenuDefaultFocusLastItem),
		DropdownMenuPositioningOf(MenuPositioningFixed),
		DropdownMenuTypeaheadDelay(350*time.Millisecond),
		DropdownMenuNoHorizontalFlip(true),
		DropdownMenuNoVerticalFlip(true),
		DropdownMenuStayOpenOnOutsideClick(true),
		DropdownMenuStayOpenOnFocusout(true),
		DropdownMenuSkipRestoreFocus(true),
		DropdownMenuNoNavigationWrap(true),
		DropdownMenuHoverOpenDelay(125*time.Millisecond),
		DropdownMenuHoverCloseDelay(175*time.Millisecond),
	).(*dropdownMenuWidget)
	if !ok {
		t.Fatalf("DropdownMenu type = %T, want *dropdownMenuWidget", DropdownMenu(false, nil, items))
	}

	items[0].Label = "mutated"
	if got := w.items[0].Label; got != "One" {
		t.Fatalf("dropdown items were not defensively copied: got %q", got)
	}
	cfg := w.config.menu
	if !w.open || w.trigger == nil || cfg.selectedKey != "one" || cfg.width != 180 || cfg.maxHeight != 224 || !reflect.DeepEqual(cfg.decoration, decoration) {
		t.Fatalf("dropdown identity/geometry config = %#v", w.config)
	}
	if !cfg.quick || !cfg.hasOverflow || cfg.xOffset != 3 || cfg.yOffset != 7 {
		t.Fatalf("dropdown display config = %#v", cfg)
	}
	if cfg.anchorCorner != MenuCornerEndEnd || cfg.menuCorner != MenuCornerStartEnd || cfg.defaultFocus != MenuDefaultFocusLastItem || cfg.positioning != MenuPositioningFixed {
		t.Fatalf("dropdown placement config = %#v", cfg)
	}
	if cfg.typeaheadDelay != 350*time.Millisecond || !cfg.noHorizontalFlip || !cfg.noVerticalFlip || !cfg.stayOpenOnOutsideClick || !cfg.stayOpenOnFocusout || !cfg.skipRestoreFocus || !cfg.noNavigationWrap {
		t.Fatalf("dropdown interaction config = %#v", cfg)
	}
	if cfg.hoverOpenDelay != 125*time.Millisecond || cfg.hoverCloseDelay != 175*time.Millisecond {
		t.Fatalf("dropdown hover delay config = %#v", cfg)
	}
	cfg.onSelect(nil, "two")
	w.config.onOpenChange(nil, false)
	if selected != "two" || openChanges != 1 {
		t.Fatalf("dropdown callbacks were not wired: selected=%q open=%d", selected, openChanges)
	}
}

func TestR0DropdownKeepOpenPolicyWalksNestedItems(t *testing.T) {
	items := []MenuItem{
		{Key: "close", Label: "Close"},
		{Key: "parent", Label: "Parent", Children: []MenuItem{{Key: "keep", Label: "Keep", KeepOpen: true}}},
	}
	if menuItemKeepOpen(items, "close") {
		t.Fatal("non-keep-open item should close its dropdown")
	}
	if !menuItemKeepOpen(items, "keep") {
		t.Fatal("nested keep-open item should keep its dropdown open")
	}
	if menuItemKeepOpen(items, "missing") {
		t.Fatal("missing item should not be treated as keep-open")
	}
}

func TestR0DividerOptionsConfigureAndLayOutBothAxes(t *testing.T) {
	lineColor := color.NRGBA{R: 12, G: 34, B: 56, A: 255}
	margin := style.All(3)
	decoration := style.Decoration{}.WithRad(4)
	vertical := Divider(
		DividerVertical(true),
		DividerThickness(2),
		DividerColor(lineColor),
		DividerLength(36),
		DividerMargin(margin),
		DividerDecoration(decoration),
	).(*dividerWidget)

	if !vertical.config.vertical || vertical.config.thickness != 2 || vertical.config.color != lineColor || !vertical.config.hasColor || vertical.config.length != 36 {
		t.Fatalf("vertical divider config = %#v", vertical.config)
	}
	if vertical.config.margin != margin || !reflect.DeepEqual(vertical.config.decoration, decoration) {
		t.Fatalf("vertical divider spacing/decoration config = %#v", vertical.config)
	}

	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(160, 80)}
	if dims := vertical.Layout(ctx); dims.Size.X <= 0 || dims.Size.Y <= 0 {
		t.Fatalf("vertical divider dimensions = %v, want non-empty", dims.Size)
	}
	rt.EndFrame()

	rt.BeginFrame()
	ctx = newInteractiveLayoutTestContext(rt)
	ctx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(160, 80)}
	horizontal := Divider(DividerLength(72), DividerThickness(3)).(*dividerWidget)
	if dims := horizontal.Layout(ctx); dims.Size.X <= 0 || dims.Size.Y <= 0 {
		t.Fatalf("horizontal divider dimensions = %v, want non-empty", dims.Size)
	}
	rt.EndFrame()
}
