package widget

import (
	"image"
	"image/color"
	"time"

	"github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"
	"github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type foregroundWidget struct {
	color color.NRGBA
	child Widget
}

func withForeground(col color.NRGBA, child Widget) Widget {
	if child == nil {
		return nil
	}
	return &foregroundWidget{color: col, child: child}
}

func (f *foregroundWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if f == nil || f.child == nil {
		return layout.Dimensions{}
	}
	return f.child.Layout(ctx.WithForeground(f.color).Child(0))
}

type minSizeWidget struct {
	width  float32
	height float32
	child  Widget
}

func minSize(width, height float32, child Widget) Widget {
	if child == nil {
		return nil
	}
	return &minSizeWidget{width: width, height: height, child: child}
}

func (m *minSizeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if m == nil || m.child == nil {
		return layout.Dimensions{}
	}
	gtx := ctx.Gtx
	gtx.Constraints.Min = image.Point{}
	if m.width > 0 {
		w := gtx.Dp(unit.Dp(m.width))
		if w > gtx.Constraints.Max.X && gtx.Constraints.Max.X > 0 {
			w = gtx.Constraints.Max.X
		}
		if gtx.Constraints.Min.X < w {
			gtx.Constraints.Min.X = w
		}
	}
	if m.height > 0 {
		h := gtx.Dp(unit.Dp(m.height))
		if h > gtx.Constraints.Max.Y && gtx.Constraints.Max.Y > 0 {
			h = gtx.Constraints.Max.Y
		}
		if gtx.Constraints.Min.Y < h {
			gtx.Constraints.Min.Y = h
		}
	}
	if gtx.Constraints.Min.X > gtx.Constraints.Max.X && gtx.Constraints.Max.X > 0 {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y && gtx.Constraints.Max.Y > 0 {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	next := *ctx
	next.Gtx = gtx
	dims := m.child.Layout(next.Child(0))
	dims.Size = clampPointToConstraints(dims.Size, gtx.Constraints.Min, gtx.Constraints.Max)
	return dims
}

type md3ActionSurfaceSpec struct {
	Background     color.NRGBA
	Foreground     color.NRGBA
	Radius         float32
	Padding        style.Insets
	Border         style.Border
	Shadow         style.BoxShadow
	MinWidth       float32
	MinHeight      float32
	FillWidth      bool
	Disabled       bool
	FocusColor     color.NRGBA
	TextStyle      theme.TextStyle
	SnapBackground bool
}

func md3ActionSurface(ctx *internal.Context, clickable *event.Clickable, spec md3ActionSurfaceSpec, child Widget) layout.Dimensions {
	if child == nil {
		child = Spacer(0, 0)
	}

	cs := ctx.Theme().Colors
	bg := spec.Background
	fg := spec.Foreground
	if fg.A == 0 {
		fg = cs.OnSurface
	}
	hovered := clickable != nil && clickable.Hovered()
	pressed := clickable != nil && clickable.Pressed()
	focused := clickable != nil && clickable.Focused(ctx)
	duration, easing := md3InteractionTiming(ctx, hovered, pressed, focused, spec.Disabled)

	if spec.Disabled {
		fg = style.DisabledContent(cs.OnSurface)
		if bg.A != 0 {
			bg = style.DisabledContainer(cs.OnSurface)
		}
	} else if clickable != nil {
		if opacity := materialAnimatedStateLayerOpacity(ctx, hovered, pressed, false); opacity > 0 {
			bg = style.StateLayer(bg, fg, opacity)
		}
	}
	if !spec.SnapBackground {
		bg = md3AnimateColor(ctx, "md3-action-bg", bg, duration, easing)
	}
	fg = md3AnimateColor(ctx, "md3-action-fg", fg, duration, easing)
	border := style.Border{
		Width: md3AnimateFloat(ctx, "md3-action-border-width", spec.Border.Width, duration, easing),
		Color: md3AnimateColor(ctx, "md3-action-border-color", spec.Border.Color, duration, easing),
	}

	surfaceSpec := internal.SurfaceSpec{
		Background:  bg,
		Radius:      spec.Radius,
		Padding:     toInternalInsets(spec.Padding),
		BorderColor: border.Color,
		BorderWidth: border.Width,
	}
	shadowSpec := surfaceSpec
	if !spec.Shadow.IsZero() {
		shadowSpec.HasShadow = true
		shadowSpec.ShadowOffsetX = spec.Shadow.OffsetX
		shadowSpec.ShadowOffsetY = spec.Shadow.OffsetY
		shadowSpec.ShadowBlur = spec.Shadow.Blur
		shadowSpec.ShadowColor = spec.Shadow.Color
	}

	layoutSurface := func(surfaceCtx *internal.Context) image.Point {
		gtx := surfaceCtx.Gtx
		gtx.Constraints.Min = image.Point{}
		if spec.MinWidth > 0 {
			minW := gtx.Dp(safeDp(spec.MinWidth))
			if minW > gtx.Constraints.Max.X && gtx.Constraints.Max.X > 0 {
				minW = gtx.Constraints.Max.X
			}
			if gtx.Constraints.Min.X < minW {
				gtx.Constraints.Min.X = minW
			}
		}
		if spec.MinHeight > 0 {
			minH := gtx.Dp(safeDp(spec.MinHeight))
			if minH > gtx.Constraints.Max.Y && gtx.Constraints.Max.Y > 0 {
				minH = gtx.Constraints.Max.Y
			}
			if gtx.Constraints.Min.Y < minH {
				gtx.Constraints.Min.Y = minH
			}
		}
		if spec.FillWidth && gtx.Constraints.Max.X > 0 {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
		if gtx.Constraints.Min.X > gtx.Constraints.Max.X && gtx.Constraints.Max.X > 0 {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
		if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y && gtx.Constraints.Max.Y > 0 {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		}

		next := *surfaceCtx
		next.Gtx = gtx
		if spec.TextStyle.Size > 0 || spec.TextStyle.LineHeight > 0 {
			next = *next.WithTextStyle(spec.TextStyle)
		}
		return next.LayoutSurfaceLooseContent(surfaceSpec, func(contentCtx *internal.Context) image.Point {
			return withForeground(fg, child).Layout(contentCtx.Child(0)).Size
		})
	}

	var size image.Point
	recorded := op.Record(ctx.Gtx.Ops)
	if spec.Disabled || clickable == nil {
		size = layoutSurface(ctx.Child(0))
	} else {
		size = ctx.LayoutRippleOverlayArea(clickable.Handle(), internal.RippleSpec{
			Color:   fg,
			Radius:  spec.Radius,
			Opacity: style.StateLayerPressedOpacity,
		}, func(childCtx *internal.Context) image.Point {
			return layoutSurface(childCtx.Child(0))
		})
	}
	surfaceCall := recorded.Stop()
	ctx.DrawSurfaceShadow(size, shadowSpec)
	surfaceCall.Add(ctx.Gtx.Ops)

	focus := spec.FocusColor
	if focus.A == 0 {
		focus = cs.Primary
	}
	md3DrawFocusIndicator(ctx, size, internal.FocusIndicatorSpec{
		Color:  focus,
		Radius: spec.Radius,
	}, focused, spec.Disabled)

	return layout.Dimensions{Size: size}
}

func middleRow(children ...Widget) Widget {
	return layoutWidgetFunc(func(ctx *internal.Context) layout.Dimensions {
		items := make([]gioLayout.FlexChild, 0, len(children))
		for index, child := range children {
			idx := index
			current := child
			if current == nil {
				continue
			}
			if flexed, ok := current.(*flexedWidget); ok {
				weight := flexed.weight
				inner := flexed.child
				items = append(items, gioLayout.Flexed(weight, func(gtx gioLayout.Context) gioLayout.Dimensions {
					if inner == nil {
						return gioLayout.Dimensions{}
					}
					next := *ctx
					next.Gtx = gtx
					dims := inner.Layout(next.Child(idx))
					return gioLayout.Dimensions{Size: dims.Size}
				}))
				continue
			}
			items = append(items, gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := *ctx
				next.Gtx = gtx
				dims := current.Layout(next.Child(idx))
				return gioLayout.Dimensions{Size: dims.Size}
			}))
		}
		dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(ctx.Gtx, items...)
		return layout.Dimensions{Size: dims.Size}
	})
}

// MenuItem describes a Material 3 menu row.
type MenuItem struct {
	Key           string
	Label         string
	Leading       Widget
	Trailing      Widget
	Disabled      bool
	Selected      bool
	Divider       bool
	Children      []MenuItem
	Type          string
	Href          string
	Target        string
	KeepOpen      bool
	TypeaheadText string
}

type MenuCorner int

const (
	MenuCornerStartStart MenuCorner = iota
	MenuCornerStartEnd
	MenuCornerEndStart
	MenuCornerEndEnd
)

type MenuDefaultFocus int

const (
	MenuDefaultFocusNone MenuDefaultFocus = iota
	MenuDefaultFocusListRoot
	MenuDefaultFocusFirstItem
	MenuDefaultFocusLastItem
)

type MenuPositioning int

const (
	MenuPositioningAbsolute MenuPositioning = iota
	MenuPositioningFixed
	MenuPositioningDocument
	MenuPositioningPopover
)

type MenuOption func(*menuConfig)

type menuConfig struct {
	selectedKey            string
	onSelect               func(ctx *internal.Context, key string)
	width                  float32
	maxHeight              float32
	decoration             style.Decoration
	quick                  bool
	hasOverflow            bool
	xOffset                float32
	yOffset                float32
	anchorCorner           MenuCorner
	menuCorner             MenuCorner
	defaultFocus           MenuDefaultFocus
	positioning            MenuPositioning
	level                  int
	typeaheadDelay         time.Duration
	noHorizontalFlip       bool
	noVerticalFlip         bool
	stayOpenOnOutsideClick bool
	stayOpenOnFocusout     bool
	skipRestoreFocus       bool
	noNavigationWrap       bool
	hoverOpenDelay         time.Duration
	hoverCloseDelay        time.Duration
}

type menuWidget struct {
	items  []MenuItem
	config menuConfig
}

type menuRuntimeState struct {
	activeSubmenu string
}

func Menu(items []MenuItem, opts ...MenuOption) Widget {
	cfg := menuConfig{maxHeight: 280, anchorCorner: MenuCornerEndStart, menuCorner: MenuCornerStartStart}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &menuWidget{items: append([]MenuItem(nil), items...), config: cfg}
}

func MenuSelectedKey(key string) MenuOption {
	return func(cfg *menuConfig) { cfg.selectedKey = key }
}

func MenuOnSelect(fn func(ctx *internal.Context, key string)) MenuOption {
	return func(cfg *menuConfig) { cfg.onSelect = fn }
}

func MenuWidth(width float32) MenuOption {
	return func(cfg *menuConfig) { cfg.width = width }
}

func MenuMaxHeight(height float32) MenuOption {
	return func(cfg *menuConfig) { cfg.maxHeight = height }
}

func MenuDecoration(d style.Decoration) MenuOption {
	return func(cfg *menuConfig) { cfg.decoration = d }
}

func MenuQuick(quick bool) MenuOption {
	return func(cfg *menuConfig) { cfg.quick = quick }
}

func MenuHasOverflow(hasOverflow bool) MenuOption {
	return func(cfg *menuConfig) { cfg.hasOverflow = hasOverflow }
}

func MenuXOffset(offset float32) MenuOption {
	return func(cfg *menuConfig) { cfg.xOffset = offset }
}

func MenuYOffset(offset float32) MenuOption {
	return func(cfg *menuConfig) { cfg.yOffset = offset }
}

func MenuAnchorCorner(corner MenuCorner) MenuOption {
	return func(cfg *menuConfig) { cfg.anchorCorner = corner }
}

func MenuMenuCorner(corner MenuCorner) MenuOption {
	return func(cfg *menuConfig) { cfg.menuCorner = corner }
}

func MenuDefaultFocusOf(focus MenuDefaultFocus) MenuOption {
	return func(cfg *menuConfig) { cfg.defaultFocus = focus }
}

func MenuPositioningOf(positioning MenuPositioning) MenuOption {
	return func(cfg *menuConfig) { cfg.positioning = positioning }
}

func MenuTypeaheadDelay(delay time.Duration) MenuOption {
	return func(cfg *menuConfig) { cfg.typeaheadDelay = delay }
}

func MenuNoHorizontalFlip(disabled bool) MenuOption {
	return func(cfg *menuConfig) { cfg.noHorizontalFlip = disabled }
}

func MenuNoVerticalFlip(disabled bool) MenuOption {
	return func(cfg *menuConfig) { cfg.noVerticalFlip = disabled }
}

func MenuStayOpenOnOutsideClick(stayOpen bool) MenuOption {
	return func(cfg *menuConfig) { cfg.stayOpenOnOutsideClick = stayOpen }
}

func MenuStayOpenOnFocusout(stayOpen bool) MenuOption {
	return func(cfg *menuConfig) { cfg.stayOpenOnFocusout = stayOpen }
}

func MenuSkipRestoreFocus(skip bool) MenuOption {
	return func(cfg *menuConfig) { cfg.skipRestoreFocus = skip }
}

func MenuNoNavigationWrap(noWrap bool) MenuOption {
	return func(cfg *menuConfig) { cfg.noNavigationWrap = noWrap }
}

func MenuHoverOpenDelay(delay time.Duration) MenuOption {
	return func(cfg *menuConfig) { cfg.hoverOpenDelay = delay }
}

func MenuHoverCloseDelay(delay time.Duration) MenuOption {
	return func(cfg *menuConfig) { cfg.hoverCloseDelay = delay }
}

func (m *menuWidget) Layout(ctx *internal.Context) layout.Dimensions {
	state := menuRuntimeStateFor(ctx)
	showSelection := m.config.selectedKey != ""
	for idx := range m.items {
		if m.items[idx].Selected {
			showSelection = true
			break
		}
	}

	rowAt := func(rowCtx *internal.Context, index int) Widget {
		item := m.items[index]
		if item.Divider {
			return Padding(style.Insets{Top: 4, Bottom: 4}, Divider(DividerColor(rowCtx.Theme().Colors.OutlineVariant)))
		}
		selected := item.Selected || (m.config.selectedKey != "" && item.Key == m.config.selectedKey)
		return m.menuRow(item, selected, showSelection, state)
	}

	var body Widget
	estimatedHeight := float32(len(m.items))*densityHeight(ctx, 48, 40) + densityMetric(ctx, 16, 12)
	if m.config.maxHeight > 0 && estimatedHeight > m.config.maxHeight {
		body = FixedHeight(m.config.maxHeight, ListView(len(m.items), rowAt, ListVirtualized(true)))
	} else {
		children := make([]Widget, 0, len(m.items))
		for idx := range m.items {
			children = append(children, rowAt(ctx.Child(idx), idx))
		}
		body = Column(children...)
	}

	cs := ctx.Theme().Colors
	deco := style.Decoration{}.
		WithBg(m.config.decoration.ResolveBg(cs.SurfaceContainer)).
		WithPad(m.config.decoration.ResolvePad(densityInsets(ctx, style.All(8), style.All(6)))).
		WithRad(m.config.decoration.ResolveRad(ctx.Theme().Shapes.ExtraSmall))
	if m.config.decoration.Shadow != nil {
		deco = deco.WithShadow(*m.config.decoration.Shadow)
	} else {
		deco = deco.WithShadow(style.ElevationShadow(cs, 2))
	}
	if m.config.decoration.Border != nil {
		deco = deco.WithBorder(*m.config.decoration.Border)
	}

	root := Widget(ContainerDecoration(deco, body))
	if m.config.width > 0 {
		root = FixedWidth(m.config.width, root)
	} else {
		root = expandWidth(root)
	}
	return root.Layout(ctx.Child(0))
}

func menuRuntimeStateFor(ctx *internal.Context) *menuRuntimeState {
	value := ctx.Memo("menu-state", func() any {
		return &menuRuntimeState{}
	})
	state, ok := value.(*menuRuntimeState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: menu state type mismatch")
	}
	return state
}

func menuItemStateKey(item MenuItem) string {
	if item.Key != "" {
		return item.Key
	}
	return item.Label
}

func (m *menuWidget) menuRow(item MenuItem, selected bool, showSelection bool, state *menuRuntimeState) Widget {
	return layoutWidgetFunc(func(rowCtx *internal.Context) layout.Dimensions {
		clickable := event.UseClickable(rowCtx)
		hasSubmenu := len(item.Children) > 0
		if !item.Disabled && !hasSubmenu {
			for clickable.Clicked(rowCtx) {
				if m.config.onSelect != nil {
					m.config.onSelect(rowCtx, item.Key)
				}
			}
		}

		cs := rowCtx.Theme().Colors
		bg := color.NRGBA{}
		fg := cs.OnSurface
		if selected {
			bg = cs.SecondaryContainer
			fg = cs.OnSecondaryContainer
		}
		fg = md3AnimateColor(rowCtx, "menu-row-fg", fg, style.InteractionSelectedDuration, style.InteractionStandardEasing)

		selectionProgress := md3SelectionProgress(rowCtx, selected)
		trailing := item.Trailing
		if trailing == nil && hasSubmenu {
			trailing = Icon("chevron_right", IconSize(18))
		}
		if trailing == nil && showSelection {
			trailing = selectCheckMarkProgress(selectionProgress, cs.Primary)
		}

		rowChildren := make([]Widget, 0, 4)
		if item.Leading != nil {
			rowChildren = append(rowChildren,
				FixedWidth(28, Center(withForeground(fg, item.Leading))),
				Padding(style.Insets{Left: 12}, Spacer(0, 0)),
			)
		}
		rowChildren = append(rowChildren, Expanded(Text(item.Label, TextType(rowCtx.Theme().Types.BodyMedium))))
		if trailing != nil {
			rowChildren = append(rowChildren, Padding(style.Insets{Left: 12}, FixedWidth(24, Center(withForeground(fg, trailing)))))
		}
		content := middleRow(rowChildren...)
		snapshot := clickable.Snapshot(rowCtx, true)
		itemStateKey := menuItemStateKey(item)
		if !item.Disabled && hasSubmenu && (snapshot.Hovered || snapshot.Focused || snapshot.Pressed) && state != nil {
			state.activeSubmenu = itemStateKey
		}
		if !item.Disabled && !hasSubmenu && snapshot.Hovered && state != nil {
			state.activeSubmenu = ""
		}
		dims := md3ActionSurface(rowCtx, clickable, md3ActionSurfaceSpec{
			Background: bg,
			Foreground: fg,
			Radius:     rowCtx.Theme().Shapes.ExtraSmall,
			Padding:    densityInsets(rowCtx, style.Symmetric(8, 12), style.Symmetric(6, 12)),
			MinHeight:  densityHeight(rowCtx, 48, 40),
			FillWidth:  true,
			Disabled:   item.Disabled,
		}, content)
		if hasSubmenu && !item.Disabled && state != nil {
			openSubmenu := state.activeSubmenu == itemStateKey
			progress, visible := md3OverlayProgress(
				rowCtx.Scope("submenu"),
				"menu-submenu",
				openSubmenu,
				style.InteractionMenuEnterDuration,
				style.InteractionMenuExitDuration,
				style.InteractionEmphasizedDecelerateEasing,
				style.InteractionEmphasizedAccelerateEasing,
			)
			if visible {
				childCfg := m.config
				childCfg.level++
				childCfg.width = 0
				if childCfg.maxHeight <= 0 {
					childCfg.maxHeight = 280
				}
				submenu := &menuWidget{items: item.Children, config: childCfg}
				macro := op.Record(rowCtx.Gtx.Ops)
				subCtx := *rowCtx
				subCtx.Gtx = rowCtx.Gtx
				subCtx.Gtx.Constraints.Min = image.Point{}
				if subCtx.Gtx.Constraints.Max.X <= 0 {
					subCtx.Gtx.Constraints.Max.X = dims.Size.X
				}
				subSize := layoutMD3RevealTransition(subCtx.Child(2), progress, false, func(revealCtx *internal.Context) image.Point {
					return submenu.Layout(revealCtx.Child(0)).Size
				})
				call := macro.Stop()
				deferMacro := op.Record(rowCtx.Gtx.Ops)
				offset := op.Offset(image.Point{X: dims.Size.X + rowCtx.Gtx.Dp(safeDp(4))}).Push(rowCtx.Gtx.Ops)
				call.Add(rowCtx.Gtx.Ops)
				offset.Pop()
				deferCall := deferMacro.Stop()
				_ = subSize
				op.Defer(rowCtx.Gtx.Ops, deferCall)
			}
		}
		return dims
	})
}

type DropdownMenuOption func(*dropdownMenuConfig)

type dropdownMenuConfig struct {
	menu         menuConfig
	onOpenChange func(ctx *internal.Context, open bool)
}

type dropdownMenuWidget struct {
	open    bool
	trigger Widget
	items   []MenuItem
	config  dropdownMenuConfig
}

type dropdownMenuRuntimeState struct {
	outsideTag any
}

func DropdownMenu(open bool, trigger Widget, items []MenuItem, opts ...DropdownMenuOption) Widget {
	cfg := dropdownMenuConfig{menu: menuConfig{maxHeight: 280, anchorCorner: MenuCornerEndStart, menuCorner: MenuCornerStartStart}}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &dropdownMenuWidget{
		open:    open,
		trigger: trigger,
		items:   append([]MenuItem(nil), items...),
		config:  cfg,
	}
}

func DropdownMenuSelectedKey(key string) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.selectedKey = key }
}

func DropdownMenuOnSelect(fn func(ctx *internal.Context, key string)) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.onSelect = fn }
}

func DropdownMenuOnOpenChange(fn func(ctx *internal.Context, open bool)) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.onOpenChange = fn }
}

func DropdownMenuWidth(width float32) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.width = width }
}

func DropdownMenuMaxHeight(height float32) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.maxHeight = height }
}

func DropdownMenuDecoration(d style.Decoration) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.decoration = d }
}

func DropdownMenuQuick(quick bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.quick = quick }
}

func DropdownMenuHasOverflow(hasOverflow bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.hasOverflow = hasOverflow }
}

func DropdownMenuXOffset(offset float32) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.xOffset = offset }
}

func DropdownMenuYOffset(offset float32) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.yOffset = offset }
}

func DropdownMenuAnchorCorner(corner MenuCorner) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.anchorCorner = corner }
}

func DropdownMenuMenuCorner(corner MenuCorner) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.menuCorner = corner }
}

func DropdownMenuDefaultFocusOf(focus MenuDefaultFocus) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.defaultFocus = focus }
}

func DropdownMenuPositioningOf(positioning MenuPositioning) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.positioning = positioning }
}

func DropdownMenuTypeaheadDelay(delay time.Duration) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.typeaheadDelay = delay }
}

func DropdownMenuNoHorizontalFlip(disabled bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.noHorizontalFlip = disabled }
}

func DropdownMenuNoVerticalFlip(disabled bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.noVerticalFlip = disabled }
}

func DropdownMenuStayOpenOnOutsideClick(stayOpen bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.stayOpenOnOutsideClick = stayOpen }
}

func DropdownMenuStayOpenOnFocusout(stayOpen bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.stayOpenOnFocusout = stayOpen }
}

func DropdownMenuSkipRestoreFocus(skip bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.skipRestoreFocus = skip }
}

func DropdownMenuNoNavigationWrap(noWrap bool) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.noNavigationWrap = noWrap }
}

func DropdownMenuHoverOpenDelay(delay time.Duration) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.hoverOpenDelay = delay }
}

func DropdownMenuHoverCloseDelay(delay time.Duration) DropdownMenuOption {
	return func(cfg *dropdownMenuConfig) { cfg.menu.hoverCloseDelay = delay }
}

func isEndAlignedMenu(anchorCorner, menuCorner MenuCorner) bool {
	anchorEnd := anchorCorner == MenuCornerStartEnd || anchorCorner == MenuCornerEndEnd
	menuEnd := menuCorner == MenuCornerStartEnd || menuCorner == MenuCornerEndEnd
	return anchorEnd && menuEnd
}

func menuItemKeepOpen(items []MenuItem, key string) bool {
	for _, item := range items {
		if item.Key == key {
			return item.KeepOpen
		}
		if len(item.Children) > 0 && menuItemKeepOpen(item.Children, key) {
			return true
		}
	}
	return false
}

func (d *dropdownMenuWidget) Layout(ctx *internal.Context) layout.Dimensions {
	state := dropdownMenuRuntimeStateFor(ctx)
	trigger := d.trigger
	if trigger == nil {
		trigger = Text("Menu")
	}
	open := d.open
	clickable := event.UseClickable(ctx)
	for clickable.Clicked(ctx) {
		open = !open
		if d.config.onOpenChange != nil {
			d.config.onOpenChange(ctx, open)
		}
	}
	triggerDims := ctx.LayoutRippleArea(clickable.Handle(), internal.RippleSpec{
		Color:   ctx.Theme().Colors.Primary,
		Radius:  ctx.Theme().Shapes.ExtraSmall,
		Opacity: style.StateLayerPressedOpacity,
	}, func(childCtx *internal.Context) image.Point {
		return trigger.Layout(childCtx.Child(0)).Size
	})
	md3DrawFocusIndicator(ctx, triggerDims, internal.FocusIndicatorSpec{
		Color:  ctx.Theme().Colors.Primary,
		Radius: ctx.Theme().Shapes.ExtraSmall,
	}, clickable.Focused(ctx), false)

	enterDuration := style.InteractionMenuEnterDuration
	exitDuration := style.InteractionMenuExitDuration
	if d.config.menu.quick {
		enterDuration = 0
		exitDuration = 0
	}
	popupProgress, popupVisible := md3OverlayProgress(ctx, "dropdown-menu-popup", open && len(d.items) > 0, enterDuration, exitDuration, style.InteractionEmphasizedDecelerateEasing, style.InteractionEmphasizedAccelerateEasing)
	if !popupVisible {
		return layout.Dimensions{Size: triggerDims}
	}

	menuCfg := d.config.menu
	pxPerDp := float32(ctx.Gtx.Metric.PxPerDp)
	if pxPerDp <= 0 {
		pxPerDp = 1
	}
	if menuCfg.width <= 0 {
		menuCfg.width = float32(triggerDims.X) / pxPerDp
	}
	popupW := ctx.Gtx.Dp(safeDp(menuCfg.width))
	if popupW <= 0 {
		popupW = triggerDims.X
	}
	if maxW := ctx.Gtx.Constraints.Max.X; maxW > 0 && popupW > maxW {
		popupW = maxW
		menuCfg.width = float32(popupW) / pxPerDp
	}
	if popupW <= 0 {
		popupW = 1
	}

	estimatedHeight := float32(len(d.items))*densityHeight(ctx, 48, 40) + densityMetric(ctx, 16, 12)
	preferredHeightPx := ctx.Gtx.Dp(safeDp(estimatedHeight))
	if preferredHeightPx <= 0 {
		preferredHeightPx = 1
	}
	if menuCfg.maxHeight > 0 {
		maxHeightPx := ctx.Gtx.Dp(safeDp(menuCfg.maxHeight))
		if maxHeightPx > 0 && preferredHeightPx > maxHeightPx {
			preferredHeightPx = maxHeightPx
		}
	}
	gapPx := ctx.Gtx.Dp(safeDp(4))
	placement := md3PopupPlacementForAnchor(ctx, triggerDims, image.Point{X: popupW}, preferredHeightPx, gapPx)
	menuCfg.maxHeight = float32(placement.MaxHeightPx) / pxPerDp

	originalSelect := menuCfg.onSelect
	menuCfg.onSelect = func(selectCtx *internal.Context, key string) {
		if originalSelect != nil {
			originalSelect(selectCtx, key)
		}
		if menuItemKeepOpen(d.items, key) {
			return
		}
		if d.config.onOpenChange != nil {
			d.config.onOpenChange(selectCtx, false)
		}
	}
	menu := (&menuWidget{items: d.items, config: menuCfg})

	recordPopup := func(p md3PopupPlacement) (image.Point, op.CallOp) {
		popupMacro := op.Record(ctx.Gtx.Ops)
		popupCtx := *ctx
		popupCtx.Gtx = ctx.Gtx
		popupCtx.Gtx.Constraints.Min = image.Point{}
		popupCtx.Gtx.Constraints.Max = image.Point{X: popupW, Y: p.MaxHeightPx}
		popupSize := layoutMD3RevealTransition(popupCtx.Child(1), popupProgress, p.Direction == md3PopupUp, func(transitionCtx *internal.Context) image.Point {
			return menu.Layout(transitionCtx.Child(0)).Size
		})
		return popupSize, popupMacro.Stop()
	}
	popupSize, _ := recordPopup(placement)
	placement = md3PopupPlacementForMeasuredPopup(ctx, triggerDims, popupSize, placement, ctx.Gtx.Dp(safeDp(menuCfg.yOffset)))
	popupSize, popupCall := recordPopup(placement)
	deferMacro := op.Record(ctx.Gtx.Ops)
	offsetX := placement.OffsetX + ctx.Gtx.Dp(safeDp(menuCfg.xOffset))
	if isEndAlignedMenu(menuCfg.anchorCorner, menuCfg.menuCorner) {
		offsetX += triggerDims.X - popupSize.X
	}
	offsetY := md3PopupOffsetY(triggerDims.Y, popupSize.Y, placement) + ctx.Gtx.Dp(safeDp(menuCfg.yOffset))
	if open && d.config.onOpenChange != nil && !menuCfg.stayOpenOnOutsideClick {
		origin := ctx.Position()
		triggerRect := image.Rectangle{Min: origin, Max: origin.Add(triggerDims)}
		popupOrigin := origin.Add(image.Point{X: offsetX, Y: offsetY})
		popupRect := image.Rectangle{Min: popupOrigin, Max: popupOrigin.Add(popupSize)}
		md3DismissOnOutsidePress(ctx, state.outsideTag, []image.Rectangle{triggerRect, popupRect}, func(dismissCtx *internal.Context) {
			d.config.onOpenChange(dismissCtx, false)
		})
	}
	offset := op.Offset(image.Point{
		X: offsetX,
		Y: offsetY,
	}).Push(ctx.Gtx.Ops)
	popupCall.Add(ctx.Gtx.Ops)
	offset.Pop()
	deferCall := deferMacro.Stop()
	op.Defer(ctx.Gtx.Ops, deferCall)
	return layout.Dimensions{Size: triggerDims}
}

func dropdownMenuRuntimeStateFor(ctx *internal.Context) *dropdownMenuRuntimeState {
	value := ctx.Memo("dropdown-menu-state", func() any {
		return &dropdownMenuRuntimeState{outsideTag: new(int)}
	})
	state, ok := value.(*dropdownMenuRuntimeState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: dropdown menu state type mismatch")
	}
	if state.outsideTag == nil {
		state.outsideTag = new(int)
	}
	return state
}

type ListItemOption func(*listItemConfig)

type listItemConfig struct {
	selected   bool
	disabled   bool
	onClick    func(ctx *internal.Context)
	minHeight  float32
	decoration style.Decoration
}

type listItemWidget struct {
	headline   Widget
	supporting Widget
	leading    Widget
	trailing   Widget
	config     listItemConfig
}

func ListItem(headline string, opts ...ListItemOption) Widget {
	return ListItemWithSlots(Text(headline), nil, nil, nil, opts...)
}

func ListItemWithSlots(headline, supporting, leading, trailing Widget, opts ...ListItemOption) Widget {
	cfg := listItemConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &listItemWidget{
		headline:   headline,
		supporting: supporting,
		leading:    leading,
		trailing:   trailing,
		config:     cfg,
	}
}

func ListItemSelected(selected bool) ListItemOption {
	return func(cfg *listItemConfig) { cfg.selected = selected }
}

func ListItemDisabled(disabled bool) ListItemOption {
	return func(cfg *listItemConfig) { cfg.disabled = disabled }
}

func ListItemOnClick(fn func(ctx *internal.Context)) ListItemOption {
	return func(cfg *listItemConfig) { cfg.onClick = fn }
}

func ListItemMinHeight(height float32) ListItemOption {
	return func(cfg *listItemConfig) { cfg.minHeight = height }
}

func ListItemDecoration(d style.Decoration) ListItemOption {
	return func(cfg *listItemConfig) { cfg.decoration = d }
}

func (l *listItemWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if !l.config.disabled {
		for clickable.Clicked(ctx) {
			if l.config.onClick != nil {
				l.config.onClick(ctx)
			}
		}
	}

	cs := ctx.Theme().Colors
	bg := l.config.decoration.ResolveBg(color.NRGBA{})
	fg := cs.OnSurface
	supportingColor := cs.OnSurfaceVariant
	if l.config.selected {
		bg = l.config.decoration.ResolveBg(cs.SecondaryContainer)
		fg = cs.OnSecondaryContainer
		supportingColor = cs.OnSecondaryContainer
	}
	fg = md3AnimateColor(ctx, "list-item-fg", fg, style.InteractionSelectedDuration, style.InteractionStandardEasing)
	supportingColor = md3AnimateColor(ctx, "list-item-supporting", supportingColor, style.InteractionSelectedDuration, style.InteractionStandardEasing)

	headline := l.headline
	if headline == nil {
		headline = Text("")
	}
	contentChildren := []Widget{
		withTextStyle(ctx.Theme().Types.BodyLarge, withForeground(fg, headline)),
	}
	if l.supporting != nil {
		contentChildren = append(contentChildren, Padding(style.Insets{Top: 2}, withTextStyle(ctx.Theme().Types.BodyMedium, withForeground(supportingColor, l.supporting))))
	}
	content := Widget(Column(contentChildren...))

	rowChildren := make([]Widget, 0, 5)
	if l.leading != nil {
		rowChildren = append(rowChildren,
			FixedWidth(32, Center(withForeground(supportingColor, l.leading))),
			Padding(style.Insets{Left: 14}, Spacer(0, 0)),
		)
	}
	rowChildren = append(rowChildren, Expanded(content))
	if l.trailing != nil {
		rowChildren = append(rowChildren,
			Padding(style.Insets{Left: 14}, FixedWidth(32, Center(withForeground(supportingColor, l.trailing)))),
		)
	}

	minHeight := densityHeight(ctx, 56, 48)
	if l.config.minHeight > 0 {
		minHeight = l.config.minHeight
	}

	border := style.Border{}
	if l.config.decoration.Border != nil {
		border = *l.config.decoration.Border
	}
	shadow := style.BoxShadow{}
	if l.config.decoration.Shadow != nil {
		shadow = *l.config.decoration.Shadow
	}

	return md3ActionSurface(ctx, clickable, md3ActionSurfaceSpec{
		Background: bg,
		Foreground: fg,
		Radius:     l.config.decoration.ResolveRad(ctx.Theme().Shapes.ExtraSmall),
		Padding:    l.config.decoration.ResolvePad(densityInsets(ctx, style.Symmetric(8, 16), style.Symmetric(6, 16))),
		Border:     border,
		Shadow:     shadow,
		MinHeight:  minHeight,
		FillWidth:  true,
		Disabled:   l.config.disabled,
	}, middleRow(rowChildren...))
}

type IconButtonOption func(*iconButtonConfig)

type iconButtonVariant int

const (
	iconButtonVariantStandard iconButtonVariant = iota
	iconButtonVariantFilled
	iconButtonVariantFilledTonal
	iconButtonVariantOutlined
)

type iconButtonConfig struct {
	variant       iconButtonVariant
	disabled      bool
	selected      bool
	loading       bool
	size          float32
	onClick       func(ctx *internal.Context)
	background    color.NRGBA
	hasBackground bool
	foreground    color.NRGBA
	hasForeground bool
	loadingIcon   Widget
	decoration    style.Decoration
}

type iconButtonWidget struct {
	child  Widget
	config iconButtonConfig
}

func IconButton(child Widget, opts ...IconButtonOption) Widget {
	return newIconButton(iconButtonVariantStandard, child, opts...)
}

func FilledIconButton(child Widget, opts ...IconButtonOption) Widget {
	return newIconButton(iconButtonVariantFilled, child, opts...)
}

func FilledTonalIconButton(child Widget, opts ...IconButtonOption) Widget {
	return newIconButton(iconButtonVariantFilledTonal, child, opts...)
}

func OutlinedIconButton(child Widget, opts ...IconButtonOption) Widget {
	return newIconButton(iconButtonVariantOutlined, child, opts...)
}

func newIconButton(variant iconButtonVariant, child Widget, opts ...IconButtonOption) Widget {
	cfg := iconButtonConfig{variant: variant, size: 40}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &iconButtonWidget{child: child, config: cfg}
}

func IconButtonOnClick(fn func(ctx *internal.Context)) IconButtonOption {
	return func(cfg *iconButtonConfig) { cfg.onClick = fn }
}

func IconButtonDisabled(disabled bool) IconButtonOption {
	return func(cfg *iconButtonConfig) { cfg.disabled = disabled }
}

func IconButtonSelected(selected bool) IconButtonOption {
	return func(cfg *iconButtonConfig) { cfg.selected = selected }
}

func IconButtonLoading(loading bool) IconButtonOption {
	return func(cfg *iconButtonConfig) { cfg.loading = loading }
}

func IconButtonLoadingIndicator(indicator Widget) IconButtonOption {
	return func(cfg *iconButtonConfig) { cfg.loadingIcon = indicator }
}

func IconButtonSize(size float32) IconButtonOption {
	return func(cfg *iconButtonConfig) { cfg.size = size }
}

func IconButtonBackground(col color.NRGBA) IconButtonOption {
	return func(cfg *iconButtonConfig) {
		cfg.background = col
		cfg.hasBackground = true
	}
}

func IconButtonForeground(col color.NRGBA) IconButtonOption {
	return func(cfg *iconButtonConfig) {
		cfg.foreground = col
		cfg.hasForeground = true
	}
}

func IconButtonDecoration(d style.Decoration) IconButtonOption {
	return func(cfg *iconButtonConfig) { cfg.decoration = d }
}

func (i *iconButtonWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if !i.config.disabled && !i.config.loading {
		for clickable.Clicked(ctx) {
			if i.config.onClick != nil {
				i.config.onClick(ctx)
			}
		}
	}

	cs := ctx.Theme().Colors
	bg := color.NRGBA{}
	fg := cs.OnSurfaceVariant
	border := style.Border{}
	switch i.config.variant {
	case iconButtonVariantFilled:
		bg = cs.Primary
		fg = cs.OnPrimary
		if !i.config.selected {
			bg = cs.SurfaceContainerHighest
			fg = cs.Primary
		}
	case iconButtonVariantFilledTonal:
		bg = cs.SecondaryContainer
		fg = cs.OnSecondaryContainer
	case iconButtonVariantOutlined:
		fg = cs.OnSurfaceVariant
		border = style.Border{Width: 1, Color: cs.Outline}
	default:
		if i.config.selected {
			fg = cs.Primary
		}
	}
	if i.config.selected && i.config.variant == iconButtonVariantOutlined {
		bg = cs.InverseSurface
		fg = cs.InverseOnSurface
	}
	if i.config.hasBackground {
		bg = i.config.background
	}
	if i.config.hasForeground {
		fg = i.config.foreground
	}
	if i.config.decoration.Border != nil {
		border = *i.config.decoration.Border
	}

	size := i.config.size
	if size <= 0 {
		size = 40
	}
	radius := i.config.decoration.ResolveRad(ctx.Theme().Shapes.Full)
	child := i.child
	if i.config.loading {
		child = i.config.loadingIcon
		if child == nil {
			child = LoadingIndicator(ProgressSize(20), ProgressThickness(3), ProgressFillColor(fg), ProgressTrackColor(color.NRGBA{}))
		}
	}
	if child == nil {
		child = Icon("icon", IconSize(20))
	}

	return md3ActionSurface(ctx, clickable, md3ActionSurfaceSpec{
		Background: bg,
		Foreground: fg,
		Radius:     radius,
		Padding:    i.config.decoration.ResolvePad(style.All(0)),
		Border:     border,
		MinWidth:   size,
		MinHeight:  size,
		Disabled:   i.config.disabled,
	}, Center(child))
}

type FloatingActionButtonOption func(*fabConfig)

type fabVariant int

const (
	fabVariantRegular fabVariant = iota
	fabVariantSmall
	fabVariantLarge
	fabVariantExtended
)

type fabConfig struct {
	variant       fabVariant
	disabled      bool
	onClick       func(ctx *internal.Context)
	background    color.NRGBA
	hasBackground bool
	foreground    color.NRGBA
	hasForeground bool
	decoration    style.Decoration
}

type fabWidget struct {
	icon   Widget
	label  Widget
	config fabConfig
}

func FloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget {
	return newFAB(fabVariantRegular, icon, nil, opts...)
}

func SmallFloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget {
	return newFAB(fabVariantSmall, icon, nil, opts...)
}

func LargeFloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget {
	return newFAB(fabVariantLarge, icon, nil, opts...)
}

func ExtendedFloatingActionButton(icon, label Widget, opts ...FloatingActionButtonOption) Widget {
	return newFAB(fabVariantExtended, icon, label, opts...)
}

func newFAB(variant fabVariant, icon, label Widget, opts ...FloatingActionButtonOption) Widget {
	cfg := fabConfig{variant: variant}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &fabWidget{icon: icon, label: label, config: cfg}
}

func FloatingActionButtonOnClick(fn func(ctx *internal.Context)) FloatingActionButtonOption {
	return func(cfg *fabConfig) { cfg.onClick = fn }
}

func FloatingActionButtonDisabled(disabled bool) FloatingActionButtonOption {
	return func(cfg *fabConfig) { cfg.disabled = disabled }
}

func FloatingActionButtonBackground(col color.NRGBA) FloatingActionButtonOption {
	return func(cfg *fabConfig) {
		cfg.background = col
		cfg.hasBackground = true
	}
}

func FloatingActionButtonForeground(col color.NRGBA) FloatingActionButtonOption {
	return func(cfg *fabConfig) {
		cfg.foreground = col
		cfg.hasForeground = true
	}
}

func FloatingActionButtonDecoration(d style.Decoration) FloatingActionButtonOption {
	return func(cfg *fabConfig) { cfg.decoration = d }
}

func (f *fabWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if !f.config.disabled {
		for clickable.Clicked(ctx) {
			if f.config.onClick != nil {
				f.config.onClick(ctx)
			}
		}
	}

	cs := ctx.Theme().Colors
	bg := cs.PrimaryContainer
	fg := cs.OnPrimaryContainer
	if f.config.hasBackground {
		bg = f.config.background
	}
	if f.config.hasForeground {
		fg = f.config.foreground
	}

	minW, minH := float32(56), float32(56)
	radius := float32(16)
	padding := style.All(0)
	content := f.icon
	if content == nil {
		content = Icon("add", IconSize(24))
	}
	switch f.config.variant {
	case fabVariantSmall:
		minW, minH = 40, 40
		radius = 12
	case fabVariantLarge:
		minW, minH = 96, 96
		radius = 28
	case fabVariantExtended:
		minW, minH = 80, 56
		radius = 16
		padding = style.Symmetric(0, 16)
		label := f.label
		if label == nil {
			label = Text("")
		}
		if f.icon != nil {
			content = Row(f.icon, Padding(style.Insets{Left: 12}, label))
		} else {
			content = label
		}
	}

	if f.config.decoration.Shadow != nil {
		// handled below through the resolved shadow
	}
	radius = f.config.decoration.ResolveRad(radius)
	shadow := style.ElevationShadow(cs, 3)
	if f.config.decoration.Shadow != nil {
		shadow = *f.config.decoration.Shadow
	}

	return md3ActionSurface(ctx, clickable, md3ActionSurfaceSpec{
		Background: bg,
		Foreground: fg,
		Radius:     radius,
		Padding:    f.config.decoration.ResolvePad(padding),
		Shadow:     shadow,
		MinWidth:   minW,
		MinHeight:  minH,
		Disabled:   f.config.disabled,
	}, Center(content))
}

type NavigationRailOption func(*navigationRailConfig)

type navigationRailConfig struct {
	onChange      func(ctx *internal.Context, key string)
	width         float32
	header        Widget
	footer        Widget
	activeColor   color.NRGBA
	hasActive     bool
	inactiveColor color.NRGBA
	hasInactive   bool
	decoration    style.Decoration
}

type navigationRailWidget struct {
	active string
	items  []NavItem
	config navigationRailConfig
}

func NavigationRail(active string, items []NavItem, opts ...NavigationRailOption) Widget {
	cfg := navigationRailConfig{width: 80}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &navigationRailWidget{active: active, items: append([]NavItem(nil), items...), config: cfg}
}

func NavigationRailOnChange(fn func(ctx *internal.Context, key string)) NavigationRailOption {
	return func(cfg *navigationRailConfig) { cfg.onChange = fn }
}

func NavigationRailWidth(width float32) NavigationRailOption {
	return func(cfg *navigationRailConfig) { cfg.width = width }
}

func NavigationRailHeader(header Widget) NavigationRailOption {
	return func(cfg *navigationRailConfig) { cfg.header = header }
}

func NavigationRailFooter(footer Widget) NavigationRailOption {
	return func(cfg *navigationRailConfig) { cfg.footer = footer }
}

func NavigationRailActiveColor(col color.NRGBA) NavigationRailOption {
	return func(cfg *navigationRailConfig) {
		cfg.activeColor = col
		cfg.hasActive = true
	}
}

func NavigationRailInactiveColor(col color.NRGBA) NavigationRailOption {
	return func(cfg *navigationRailConfig) {
		cfg.inactiveColor = col
		cfg.hasInactive = true
	}
}

func NavigationRailDecoration(d style.Decoration) NavigationRailOption {
	return func(cfg *navigationRailConfig) { cfg.decoration = d }
}

func (n *navigationRailWidget) Layout(ctx *internal.Context) layout.Dimensions {
	cs := ctx.Theme().Colors
	children := make([]Widget, 0, len(n.items)+3)
	if n.config.header != nil {
		children = append(children, Padding(style.Insets{Bottom: 16}, Center(n.config.header)))
	}
	for idx := range n.items {
		item := n.items[idx]
		children = append(children, Padding(style.Insets{Bottom: 8}, n.railItem(item)))
	}
	children = append(children, Expanded(Spacer(0, 0)))
	if n.config.footer != nil {
		children = append(children, Center(n.config.footer))
	}

	bg := n.config.decoration.ResolveBg(cs.SurfaceContainer)
	deco := style.Decoration{}.
		WithBg(bg).
		WithPad(n.config.decoration.ResolvePad(style.Symmetric(12, 8))).
		WithRad(n.config.decoration.ResolveRad(0))
	if n.config.decoration.Border != nil {
		deco = deco.WithBorder(*n.config.decoration.Border)
	}
	if n.config.decoration.Shadow != nil {
		deco = deco.WithShadow(*n.config.decoration.Shadow)
	}
	width := n.config.width
	if width <= 0 {
		width = 80
	}
	return FixedWidth(width, ContainerDecoration(deco, Column(children...))).Layout(ctx.Child(0))
}

func (n *navigationRailWidget) railItem(item NavItem) Widget {
	return layoutWidgetFunc(func(itemCtx *internal.Context) layout.Dimensions {
		clickable := event.UseClickable(itemCtx)
		for clickable.Clicked(itemCtx) {
			if n.config.onChange != nil {
				n.config.onChange(itemCtx, item.Key)
			}
		}

		cs := itemCtx.Theme().Colors
		active := item.Key == n.active
		fg := cs.OnSurfaceVariant
		if n.config.hasInactive {
			fg = n.config.inactiveColor
		}
		indicatorBg := color.NRGBA{}
		if active {
			indicatorBg = cs.SecondaryContainer
			fg = cs.OnSecondaryContainer
			if n.config.hasActive {
				fg = n.config.activeColor
			}
		}
		fg = md3AnimateColor(itemCtx, "rail-fg", fg, style.InteractionSelectedDuration, style.InteractionStandardEasing)
		icon := item.Icon
		if icon == nil {
			icon = Text(firstLabelRune(item.Label))
		}
		indicator := FixedSize(56, 32, ContainerDecoration(
			style.Decoration{}.WithBg(indicatorBg).WithRad(itemCtx.Theme().Shapes.Full),
			Center(withForeground(fg, icon)),
		))
		content := FixedWidth(56, Column(
			Center(indicator),
			Padding(style.Insets{Top: 4}, Center(Text(item.Label, TextType(itemCtx.Theme().Types.LabelMedium)))),
		))
		return md3ActionSurface(itemCtx, clickable, md3ActionSurfaceSpec{
			Background:     color.NRGBA{},
			Foreground:     fg,
			Radius:         itemCtx.Theme().Shapes.Large,
			Padding:        style.Symmetric(4, 0),
			MinHeight:      64,
			FillWidth:      true,
			SnapBackground: true,
		}, content)
	})
}

func firstLabelRune(label string) string {
	for _, r := range label {
		return string(r)
	}
	return ""
}

type NavigationDrawerOption func(*navigationDrawerConfig)

type navigationDrawerConfig struct {
	onChange      func(ctx *internal.Context, key string)
	width         float32
	header        Widget
	footer        Widget
	activeColor   color.NRGBA
	hasActive     bool
	inactiveColor color.NRGBA
	hasInactive   bool
	decoration    style.Decoration
}

type navigationDrawerWidget struct {
	active string
	items  []NavItem
	config navigationDrawerConfig
}

func NavigationDrawer(active string, items []NavItem, opts ...NavigationDrawerOption) Widget {
	cfg := navigationDrawerConfig{width: 360}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &navigationDrawerWidget{active: active, items: append([]NavItem(nil), items...), config: cfg}
}

func NavigationDrawerOnChange(fn func(ctx *internal.Context, key string)) NavigationDrawerOption {
	return func(cfg *navigationDrawerConfig) { cfg.onChange = fn }
}

func NavigationDrawerWidth(width float32) NavigationDrawerOption {
	return func(cfg *navigationDrawerConfig) { cfg.width = width }
}

func NavigationDrawerHeader(header Widget) NavigationDrawerOption {
	return func(cfg *navigationDrawerConfig) { cfg.header = header }
}

func NavigationDrawerFooter(footer Widget) NavigationDrawerOption {
	return func(cfg *navigationDrawerConfig) { cfg.footer = footer }
}

func NavigationDrawerActiveColor(col color.NRGBA) NavigationDrawerOption {
	return func(cfg *navigationDrawerConfig) {
		cfg.activeColor = col
		cfg.hasActive = true
	}
}

func NavigationDrawerInactiveColor(col color.NRGBA) NavigationDrawerOption {
	return func(cfg *navigationDrawerConfig) {
		cfg.inactiveColor = col
		cfg.hasInactive = true
	}
}

func NavigationDrawerDecoration(d style.Decoration) NavigationDrawerOption {
	return func(cfg *navigationDrawerConfig) { cfg.decoration = d }
}

func (n *navigationDrawerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	cs := ctx.Theme().Colors
	children := make([]Widget, 0, len(n.items)+3)
	if n.config.header != nil {
		children = append(children, Padding(style.Insets{Bottom: 12}, n.config.header))
	}
	for idx := range n.items {
		item := n.items[idx]
		children = append(children, Padding(style.Insets{Bottom: 4}, n.drawerItem(item)))
	}
	children = append(children, Expanded(Spacer(0, 0)))
	if n.config.footer != nil {
		children = append(children, Padding(style.Insets{Top: 12}, n.config.footer))
	}

	bg := n.config.decoration.ResolveBg(cs.SurfaceContainerLow)
	deco := style.Decoration{}.
		WithBg(bg).
		WithPad(n.config.decoration.ResolvePad(densityInsets(ctx, style.All(12), style.All(8)))).
		WithRad(n.config.decoration.ResolveRad(0))
	if n.config.decoration.Border != nil {
		deco = deco.WithBorder(*n.config.decoration.Border)
	}
	if n.config.decoration.Shadow != nil {
		deco = deco.WithShadow(*n.config.decoration.Shadow)
	}
	width := n.config.width
	if width <= 0 {
		width = 360
	}
	return FixedWidth(width, ContainerDecoration(deco, Column(children...))).Layout(ctx.Child(0))
}

func (n *navigationDrawerWidget) drawerItem(item NavItem) Widget {
	return layoutWidgetFunc(func(itemCtx *internal.Context) layout.Dimensions {
		clickable := event.UseClickable(itemCtx)
		for clickable.Clicked(itemCtx) {
			if n.config.onChange != nil {
				n.config.onChange(itemCtx, item.Key)
			}
		}

		cs := itemCtx.Theme().Colors
		active := item.Key == n.active
		bg := color.NRGBA{}
		fg := cs.OnSurfaceVariant
		if n.config.hasInactive {
			fg = n.config.inactiveColor
		}
		if active {
			bg = cs.SecondaryContainer
			fg = cs.OnSecondaryContainer
			if n.config.hasActive {
				fg = n.config.activeColor
			}
		}
		fg = md3AnimateColor(itemCtx, "drawer-fg", fg, style.InteractionSelectedDuration, style.InteractionStandardEasing)
		icon := item.Icon
		if icon == nil {
			icon = Text("")
		}
		content := middleRow(
			FixedWidth(32, Center(withForeground(fg, icon))),
			Padding(style.Insets{Left: 12}, Spacer(0, 0)),
			Expanded(Text(item.Label, TextType(itemCtx.Theme().Types.BodyLarge))),
		)
		return md3ActionSurface(itemCtx, clickable, md3ActionSurfaceSpec{
			Background:     bg,
			Foreground:     fg,
			Radius:         itemCtx.Theme().Shapes.Full,
			Padding:        densityInsets(itemCtx, style.Symmetric(4, 16), style.Symmetric(2, 16)),
			MinHeight:      densityHeight(itemCtx, 48, 40),
			FillWidth:      true,
			SnapBackground: true,
		}, content)
	})
}

type TooltipOption func(*tooltipConfig)

type tooltipConfig struct {
	disabled   bool
	offset     float32
	decoration style.Decoration
	textColor  color.NRGBA
	hasText    bool
}

type tooltipWidget struct {
	label  string
	child  Widget
	config tooltipConfig
}

func Tooltip(label string, child Widget, opts ...TooltipOption) Widget {
	cfg := tooltipConfig{offset: 6}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &tooltipWidget{label: label, child: child, config: cfg}
}

func TooltipDisabled(disabled bool) TooltipOption {
	return func(cfg *tooltipConfig) { cfg.disabled = disabled }
}

func TooltipOffset(offset float32) TooltipOption {
	return func(cfg *tooltipConfig) { cfg.offset = offset }
}

func TooltipDecoration(d style.Decoration) TooltipOption {
	return func(cfg *tooltipConfig) { cfg.decoration = d }
}

func TooltipTextColor(col color.NRGBA) TooltipOption {
	return func(cfg *tooltipConfig) {
		cfg.textColor = col
		cfg.hasText = true
	}
}

func (t *tooltipWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if t.child == nil {
		return layout.Dimensions{}
	}

	clickable := event.UseClickable(ctx)
	size := ctx.LayoutClickArea(clickable.Handle(), func(childCtx *internal.Context) image.Point {
		return t.child.Layout(childCtx.Child(0)).Size
	})
	tooltipProgress, tooltipVisible := md3OverlayProgress(
		ctx,
		"tooltip-popup",
		!t.config.disabled && t.label != "" && (clickable.Hovered() || clickable.Focused(ctx)),
		style.InteractionHoverEnterDuration,
		style.InteractionHoverExitDuration,
		style.InteractionStandardDecelerateEasing,
		style.InteractionStandardAccelerateEasing,
	)
	if !tooltipVisible {
		return layout.Dimensions{Size: size}
	}

	cs := ctx.Theme().Colors
	fg := cs.InverseOnSurface
	if t.config.hasText {
		fg = t.config.textColor
	}
	body := ContainerDecoration(
		style.Decoration{}.
			WithBg(t.config.decoration.ResolveBg(cs.InverseSurface)).
			WithPad(t.config.decoration.ResolvePad(style.Symmetric(6, 8))).
			WithRad(t.config.decoration.ResolveRad(ctx.Theme().Shapes.ExtraSmall)),
		Text(t.label, TextColor(fg), TextType(ctx.Theme().Types.BodySmall)),
	)

	macro := op.Record(ctx.Gtx.Ops)
	tooltipCtx := *ctx
	tooltipCtx.Gtx = ctx.Gtx
	tooltipCtx.Gtx.Constraints.Min = image.Point{}
	tooltipSize := body.Layout(tooltipCtx.Child(1)).Size
	call := macro.Stop()
	offset := image.Point{
		X: (size.X - tooltipSize.X) / 2,
		Y: size.Y + ctx.Gtx.Dp(safeDp(t.config.offset)),
	}
	deferMacro := op.Record(ctx.Gtx.Ops)
	stack := op.Offset(offset).Push(ctx.Gtx.Ops)
	_ = layoutMD3OverlayTransition(ctx.Child(2), tooltipProgress, 4, func(*internal.Context) image.Point {
		call.Add(ctx.Gtx.Ops)
		return tooltipSize
	})
	stack.Pop()
	deferCall := deferMacro.Stop()
	op.Defer(ctx.Gtx.Ops, deferCall)

	return layout.Dimensions{Size: size}
}

type BadgeOption func(*badgeConfig)

type badgeConfig struct {
	visible       bool
	visibleSet    bool
	background    color.NRGBA
	foreground    color.NRGBA
	hasBackground bool
	hasForeground bool
	decoration    style.Decoration
	offset        image.Point
}

type badgeWidget struct {
	child  Widget
	label  string
	config badgeConfig
}

func Badge(child Widget, label string, opts ...BadgeOption) Widget {
	cfg := badgeConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &badgeWidget{child: child, label: label, config: cfg}
}

func BadgeVisible(visible bool) BadgeOption {
	return func(cfg *badgeConfig) {
		cfg.visible = visible
		cfg.visibleSet = true
	}
}

func BadgeBackground(col color.NRGBA) BadgeOption {
	return func(cfg *badgeConfig) {
		cfg.background = col
		cfg.hasBackground = true
	}
}

func BadgeForeground(col color.NRGBA) BadgeOption {
	return func(cfg *badgeConfig) {
		cfg.foreground = col
		cfg.hasForeground = true
	}
}

func BadgeDecoration(d style.Decoration) BadgeOption {
	return func(cfg *badgeConfig) { cfg.decoration = d }
}

func BadgeOffset(x, y int) BadgeOption {
	return func(cfg *badgeConfig) { cfg.offset = image.Point{X: x, Y: y} }
}

func (b *badgeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if b.child == nil {
		return layout.Dimensions{}
	}

	childMacro := op.Record(ctx.Gtx.Ops)
	childSize := b.child.Layout(ctx.Child(0)).Size
	childCall := childMacro.Stop()
	childCall.Add(ctx.Gtx.Ops)

	visible := b.label != ""
	if b.config.visibleSet {
		visible = b.config.visible
	}
	if !visible {
		return layout.Dimensions{Size: childSize}
	}

	cs := ctx.Theme().Colors
	bg := cs.Error
	fg := cs.OnError
	if b.config.hasBackground {
		bg = b.config.background
	}
	if b.config.hasForeground {
		fg = b.config.foreground
	}

	var badge Widget
	if b.label == "" {
		badge = FixedSize(8, 8, ContainerDecoration(
			style.Decoration{}.WithBg(bg).WithRad(ctx.Theme().Shapes.Full),
			Spacer(8, 8),
		))
	} else {
		badge = minSize(16, 16, ContainerDecoration(
			style.Decoration{}.
				WithBg(b.config.decoration.ResolveBg(bg)).
				WithPad(b.config.decoration.ResolvePad(style.Symmetric(1, 5))).
				WithRad(b.config.decoration.ResolveRad(ctx.Theme().Shapes.Full)),
			Center(Text(b.label, TextColor(fg), TextType(ctx.Theme().Types.LabelSmall))),
		))
	}

	badgeMacro := op.Record(ctx.Gtx.Ops)
	badgeCtx := *ctx
	badgeCtx.Gtx = ctx.Gtx
	badgeCtx.Gtx.Constraints.Min = image.Point{}
	badgeSize := badge.Layout(badgeCtx.Child(1)).Size
	badgeCall := badgeMacro.Stop()
	offset := image.Point{
		X: childSize.X - badgeSize.X/2 + b.config.offset.X,
		Y: -badgeSize.Y/2 + b.config.offset.Y,
	}
	stack := op.Offset(offset).Push(ctx.Gtx.Ops)
	badgeCall.Add(ctx.Gtx.Ops)
	stack.Pop()
	return layout.Dimensions{Size: childSize}
}

type ChipOption func(*chipConfig)

type chipVariant int

const (
	chipVariantAssist chipVariant = iota
	chipVariantFilter
	chipVariantInput
	chipVariantSuggestion
)

type chipConfig struct {
	variant       chipVariant
	selected      bool
	disabled      bool
	softDisabled  bool
	elevated      bool
	removable     bool
	onClick       func(ctx *internal.Context)
	onRemove      func(ctx *internal.Context)
	leading       Widget
	trailing      Widget
	background    color.NRGBA
	foreground    color.NRGBA
	hasBackground bool
	hasForeground bool
	decoration    style.Decoration
}

type chipWidget struct {
	label  Widget
	config chipConfig
}

func AssistChip(label string, opts ...ChipOption) Widget {
	return newChip(chipVariantAssist, Text(label), opts...)
}

func FilterChip(label string, opts ...ChipOption) Widget {
	return newChip(chipVariantFilter, Text(label), opts...)
}

func InputChip(label string, opts ...ChipOption) Widget {
	return newChip(chipVariantInput, Text(label), opts...)
}

func SuggestionChip(label string, opts ...ChipOption) Widget {
	return newChip(chipVariantSuggestion, Text(label), opts...)
}

func AssistChipWithSlots(label Widget, opts ...ChipOption) Widget {
	return newChip(chipVariantAssist, label, opts...)
}

func FilterChipWithSlots(label Widget, opts ...ChipOption) Widget {
	return newChip(chipVariantFilter, label, opts...)
}

func InputChipWithSlots(label Widget, opts ...ChipOption) Widget {
	return newChip(chipVariantInput, label, opts...)
}

func SuggestionChipWithSlots(label Widget, opts ...ChipOption) Widget {
	return newChip(chipVariantSuggestion, label, opts...)
}

func ChipWithSlots(label Widget, opts ...ChipOption) Widget {
	return newChip(chipVariantAssist, label, opts...)
}

func newChip(variant chipVariant, label Widget, opts ...ChipOption) Widget {
	cfg := chipConfig{variant: variant}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &chipWidget{label: label, config: cfg}
}

func ChipSelected(selected bool) ChipOption {
	return func(cfg *chipConfig) { cfg.selected = selected }
}

func ChipDisabled(disabled bool) ChipOption {
	return func(cfg *chipConfig) { cfg.disabled = disabled }
}

func ChipSoftDisabled(disabled bool) ChipOption {
	return func(cfg *chipConfig) { cfg.softDisabled = disabled }
}

func ChipElevated(elevated bool) ChipOption {
	return func(cfg *chipConfig) { cfg.elevated = elevated }
}

func ChipRemovable(removable bool) ChipOption {
	return func(cfg *chipConfig) { cfg.removable = removable }
}

func ChipOnClick(fn func(ctx *internal.Context)) ChipOption {
	return func(cfg *chipConfig) { cfg.onClick = fn }
}

func ChipOnRemove(fn func(ctx *internal.Context)) ChipOption {
	return func(cfg *chipConfig) {
		cfg.onRemove = fn
		cfg.removable = fn != nil
	}
}

func ChipLeading(leading Widget) ChipOption {
	return func(cfg *chipConfig) { cfg.leading = leading }
}

func ChipTrailing(trailing Widget) ChipOption {
	return func(cfg *chipConfig) { cfg.trailing = trailing }
}

func ChipBackground(col color.NRGBA) ChipOption {
	return func(cfg *chipConfig) {
		cfg.background = col
		cfg.hasBackground = true
	}
}

func ChipForeground(col color.NRGBA) ChipOption {
	return func(cfg *chipConfig) {
		cfg.foreground = col
		cfg.hasForeground = true
	}
}

func ChipDecoration(d style.Decoration) ChipOption {
	return func(cfg *chipConfig) { cfg.decoration = d }
}

func (c *chipWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	disabled := c.config.disabled || c.config.softDisabled
	if !disabled {
		for clickable.Clicked(ctx) {
			if c.config.onClick != nil {
				c.config.onClick(ctx)
			} else if c.config.onRemove != nil {
				c.config.onRemove(ctx)
			}
		}
	} else if c.config.softDisabled {
		for clickable.Clicked(ctx) {
		}
	}

	cs := ctx.Theme().Colors
	bg := cs.Surface
	fg := cs.OnSurfaceVariant
	border := style.Border{Width: 1, Color: cs.Outline}
	shadow := style.BoxShadow{}
	selected := c.config.selected
	outlined := true

	if c.config.selected {
		bg = cs.SecondaryContainer
		fg = cs.OnSecondaryContainer
		border = style.Border{}
		outlined = false
	}
	if c.config.elevated {
		if !selected {
			bg = cs.SurfaceContainerLow
		}
		border = style.Border{}
		outlined = false
		if !disabled {
			shadow = style.ElevationShadow(cs, 1)
		}
	}
	if c.config.variant == chipVariantInput && !selected && !c.config.elevated {
		bg = cs.Surface
	}
	if c.config.hasBackground {
		bg = c.config.background
		outlined = false
	}
	if c.config.hasForeground {
		fg = c.config.foreground
	}
	if c.config.decoration.Border != nil {
		border = *c.config.decoration.Border
		outlined = border.Width > 0
	}
	if c.config.decoration.Shadow != nil {
		shadow = *c.config.decoration.Shadow
	}
	if disabled {
		fg = style.DisabledContent(cs.OnSurface)
		if selected || c.config.elevated || c.config.hasBackground {
			bg = style.DisabledContainer(cs.OnSurface)
			border = style.Border{}
			outlined = false
		} else {
			bg = color.NRGBA{}
			border = style.Border{Width: 1, Color: style.DisabledContent(cs.OnSurface)}
			outlined = true
		}
		shadow = style.BoxShadow{}
	}
	fg = md3AnimateColor(ctx, "chip-fg", fg, style.InteractionSelectedDuration, style.InteractionStandardEasing)

	label := c.label
	if label == nil {
		label = Text("")
	}
	selectionProgress := md3SelectionProgress(ctx, c.config.selected)
	leading := c.config.leading
	if leading == nil && c.config.variant == chipVariantFilter && (c.config.selected || selectionProgress > 0.001) {
		leading = selectCheckMarkProgress(selectionProgress, fg)
	}
	trailing := c.config.trailing
	removable := c.config.removable || c.config.onRemove != nil || c.config.variant == chipVariantInput
	if trailing == nil && removable {
		trailing = Icon("close", IconSize(18))
	}

	rowChildren := make([]Widget, 0, 5)
	startPad := float32(12)
	endPad := float32(12)
	if leading != nil {
		startPad = 8
		rowChildren = append(rowChildren,
			FixedWidth(18, Center(withForeground(fg, leading))),
			Padding(style.Insets{Left: 8}, Spacer(0, 0)),
		)
	}
	rowChildren = append(rowChildren, withTextStyle(ctx.Theme().Types.LabelLarge, withForeground(fg, label)))
	if trailing != nil {
		endPad = 8
		trailingChild := FixedWidth(18, Center(withForeground(fg, trailing)))
		if c.config.onRemove != nil && !disabled {
			removeClickable := event.UseClickable(ctx.Scope("remove"))
			for removeClickable.Clicked(ctx) {
				c.config.onRemove(ctx)
			}
			removeIcon := trailingChild
			trailingChild = layoutWidgetFunc(func(removeCtx *internal.Context) layout.Dimensions {
				return md3ActionSurface(removeCtx, removeClickable, md3ActionSurfaceSpec{
					Background:     color.NRGBA{},
					Foreground:     fg,
					Radius:         9,
					Padding:        style.Insets{},
					MinWidth:       18,
					MinHeight:      18,
					SnapBackground: true,
				}, removeIcon)
			})
		}
		rowChildren = append(rowChildren,
			Padding(style.Insets{Left: 8}, trailingChild),
		)
	}
	padding := c.config.decoration.ResolvePad(style.Insets{Top: 6, Bottom: 6, Left: startPad, Right: endPad})
	if outlined && !disabled && bg == cs.Surface {
		bg = color.NRGBA{}
	}

	return md3ActionSurface(ctx, clickable, md3ActionSurfaceSpec{
		Background: bg,
		Foreground: fg,
		Radius:     c.config.decoration.ResolveRad(8),
		Padding:    padding,
		Border:     border,
		Shadow:     shadow,
		MinHeight:  32,
		Disabled:   c.config.disabled,
	}, middleRow(rowChildren...))
}

type SearchBarOption func(*searchBarConfig)

type searchBarConfig struct {
	placeholder string
	disabled    bool
	onChange    func(ctx *internal.Context, value string)
	leading     Widget
	trailing    Widget
	decoration  style.Decoration
	inputOpts   []InputOption
}

type searchBarWidget struct {
	value  string
	config searchBarConfig
}

func SearchBar(value string, opts ...SearchBarOption) Widget {
	cfg := searchBarConfig{placeholder: "Search"}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &searchBarWidget{value: value, config: cfg}
}

func SearchBarPlaceholder(text string) SearchBarOption {
	return func(cfg *searchBarConfig) { cfg.placeholder = text }
}

func SearchBarDisabled(disabled bool) SearchBarOption {
	return func(cfg *searchBarConfig) { cfg.disabled = disabled }
}

func SearchBarOnChange(fn func(ctx *internal.Context, value string)) SearchBarOption {
	return func(cfg *searchBarConfig) { cfg.onChange = fn }
}

func SearchBarLeading(leading Widget) SearchBarOption {
	return func(cfg *searchBarConfig) { cfg.leading = leading }
}

func SearchBarTrailing(trailing Widget) SearchBarOption {
	return func(cfg *searchBarConfig) { cfg.trailing = trailing }
}

func SearchBarDecoration(d style.Decoration) SearchBarOption {
	return func(cfg *searchBarConfig) { cfg.decoration = d }
}

func SearchBarInputOptions(opts ...InputOption) SearchBarOption {
	return func(cfg *searchBarConfig) { cfg.inputOpts = append(cfg.inputOpts, opts...) }
}

func (s *searchBarWidget) Layout(ctx *internal.Context) layout.Dimensions {
	cs := ctx.Theme().Colors
	fg := cs.OnSurface
	if s.config.disabled {
		fg = style.DisabledContent(cs.OnSurface)
	}
	iconColor := cs.OnSurface
	if s.config.disabled {
		iconColor = style.DisabledContent(cs.OnSurface)
	}

	leading := s.config.leading
	if leading == nil {
		leading = Icon("search", IconSize(18))
	}

	inputOpts := append([]InputOption{}, s.config.inputOpts...)
	inputOpts = append(inputOpts,
		InputPlaceholder(s.config.placeholder),
		InputDisabled(s.config.disabled),
		InputBackground(color.NRGBA{}),
		InputForeground(fg),
		InputBorder(color.NRGBA{}),
		InputBorderFocus(color.NRGBA{}),
		InputPadding(style.Symmetric(4, 0)),
		InputDecoration(style.Decoration{}.WithBg(color.NRGBA{}).WithPad(style.All(0)).WithRad(0)),
	)
	if s.config.onChange != nil {
		inputOpts = append(inputOpts, InputOnChange(s.config.onChange))
	}

	rowChildren := []Widget{
		FixedWidth(24, Center(withForeground(iconColor, leading))),
		Padding(style.Insets{Left: 12}, Spacer(0, 0)),
		Expanded(TextField(s.value, inputOpts...)),
	}
	if s.config.trailing != nil {
		rowChildren = append(rowChildren,
			Padding(style.Insets{Left: 12}, FixedWidth(24, Center(withForeground(iconColor, s.config.trailing)))),
		)
	}

	body := minSize(0, 56, ContainerDecoration(
		style.Decoration{}.
			WithBg(s.config.decoration.ResolveBg(cs.SurfaceContainerHigh)).
			WithPad(s.config.decoration.ResolvePad(style.Symmetric(0, 16))).
			WithRad(s.config.decoration.ResolveRad(ctx.Theme().Shapes.Full)),
		withForeground(fg, middleRow(rowChildren...)),
	))
	return body.Layout(ctx.Child(0))
}
