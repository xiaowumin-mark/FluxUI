package widget

import (
	"image"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"
)

// AppBarOption 顶栏配置。
type AppBarOption func(*appBarConfig)

type appBarConfig struct {
	leading    Widget
	actions    []Widget
	height     float32
	background color.NRGBA
	hasBG      bool
	decoration style.Decoration
}

type appBarWidget struct {
	title  Widget
	config appBarConfig
}

// AppBar 创建顶部导航栏。
func AppBar(title Widget, opts ...AppBarOption) Widget {
	cfg := appBarConfig{
		height: 56,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &appBarWidget{
		title:  title,
		config: cfg,
	}
}

func AppBarLeading(leading Widget) AppBarOption {
	return func(cfg *appBarConfig) {
		cfg.leading = leading
	}
}

func AppBarActions(actions ...Widget) AppBarOption {
	return func(cfg *appBarConfig) {
		cfg.actions = append([]Widget(nil), actions...)
	}
}

func AppBarHeight(height float32) AppBarOption {
	return func(cfg *appBarConfig) {
		cfg.height = height
	}
}

func AppBarBackground(col color.NRGBA) AppBarOption {
	return func(cfg *appBarConfig) {
		cfg.background = col
		cfg.hasBG = true
	}
}

// AppBarDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func AppBarDecoration(d style.Decoration) AppBarOption {
	return func(cfg *appBarConfig) {
		cfg.decoration = d
	}
}

func (a *appBarWidget) Layout(ctx *internal.Context) layout.Dimensions {
	cs := ctx.Theme().Colors
	bg := a.config.decoration.ResolveBg(cs.SurfaceContainer)
	if a.config.hasBG {
		bg = a.config.background
	}

	left := a.config.leading
	if left == nil {
		left = Text("")
	}
	center := a.title
	if center == nil {
		center = Text("")
	}
	right := Text("")
	if len(a.config.actions) > 0 {
		right = Row(a.config.actions...)
	}

	content := Row(
		left,
		Expanded(Padding(style.Insets{Left: 12, Right: 12}, center)),
		right,
	)
	if len(a.config.actions) > 0 {
		content = Row(
			left,
			Expanded(Padding(style.Insets{Left: 12, Right: 12}, center)),
			ScrollView(
				right,
				ScrollHorizontal(true),
				ScrollVertical(false),
			),
		)
	}

	barDeco := style.Decoration{}.
		WithBg(bg).
		WithPad(a.config.decoration.ResolvePad(style.Symmetric(8, 12))).
		WithRad(a.config.decoration.ResolveRad(0))
	if a.config.decoration.Shadow != nil {
		barDeco = barDeco.WithShadow(*a.config.decoration.Shadow)
	} else {
		barDeco = barDeco.WithShadow(style.ElevationShadow(cs, 2))
	}
	if a.config.decoration.Border != nil {
		barDeco = barDeco.WithBorder(*a.config.decoration.Border)
	}
	bar := ContainerDecoration(barDeco, content)
	bar = expandWidth(bar)

	return (&fixedSizeWidget{
		height: a.config.height,
		child:  bar,
	}).Layout(ctx.Child(0))
}

// NavItem 底部导航项。
type NavItem struct {
	Key   string
	Label string
	Icon  Widget
}

// BottomNavAlignment 定义底部导航对齐方式。
type BottomNavAlignment int

const (
	BottomNavAlignStart BottomNavAlignment = iota
	BottomNavAlignCenter
	BottomNavAlignEnd
	BottomNavAlignSpaceEvenly
)

// BottomNavOption 底部导航配置。
type BottomNavOption func(*bottomNavConfig)

type bottomNavConfig struct {
	onChange      func(ctx *internal.Context, key string)
	background    color.NRGBA
	hasBG         bool
	activeColor   color.NRGBA
	hasActive     bool
	inactiveColor color.NRGBA
	hasInactive   bool
	alignment     BottomNavAlignment
	ref           *BottomNavRef
	decoration    style.Decoration
}

type bottomNavWidget struct {
	active string
	items  []NavItem
	config bottomNavConfig
}

// BottomNavigation 创建底部导航。
func BottomNavigation(active string, items []NavItem, opts ...BottomNavOption) Widget {
	cfg := bottomNavConfig{
		alignment: BottomNavAlignSpaceEvenly,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &bottomNavWidget{
		active: active,
		items:  append([]NavItem(nil), items...),
		config: cfg,
	}
}

func BottomNavOnChange(fn func(ctx *internal.Context, key string)) BottomNavOption {
	return func(cfg *bottomNavConfig) {
		cfg.onChange = fn
	}
}

func BottomNavBackground(col color.NRGBA) BottomNavOption {
	return func(cfg *bottomNavConfig) {
		cfg.background = col
		cfg.hasBG = true
	}
}

func BottomNavActiveColor(col color.NRGBA) BottomNavOption {
	return func(cfg *bottomNavConfig) {
		cfg.activeColor = col
		cfg.hasActive = true
	}
}

func BottomNavInactiveColor(col color.NRGBA) BottomNavOption {
	return func(cfg *bottomNavConfig) {
		cfg.inactiveColor = col
		cfg.hasInactive = true
	}
}

func BottomNavAlignmentOf(alignment BottomNavAlignment) BottomNavOption {
	return func(cfg *bottomNavConfig) {
		cfg.alignment = alignment
	}
}

// BottomNavAttachRef 绑定命令型引用，用于外部主动切换当前项。
func BottomNavAttachRef(ref *BottomNavRef) BottomNavOption {
	return func(cfg *bottomNavConfig) {
		cfg.ref = ref
	}
}

// BottomNavDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func BottomNavDecoration(d style.Decoration) BottomNavOption {
	return func(cfg *bottomNavConfig) {
		cfg.decoration = d
	}
}

func (b *bottomNavWidget) Layout(ctx *internal.Context) layout.Dimensions {
	activeKey := b.active
	if b.config.ref != nil {
		b.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
		for _, key := range b.config.ref.drainCommands() {
			if key == activeKey {
				continue
			}
			activeKey = key
			if b.config.onChange != nil {
				b.config.onChange(ctx, key)
			}
		}
	}

	cs := ctx.Theme().Colors
	bg := b.config.decoration.ResolveBg(cs.SurfaceContainer)
	if b.config.hasBG {
		bg = b.config.background
	}
	activeColor := cs.OnSecondaryContainer
	if b.config.hasActive {
		activeColor = b.config.activeColor
	}
	inactiveColor := cs.OnSurfaceVariant
	if b.config.hasInactive {
		inactiveColor = b.config.inactiveColor
	}

	tabs := make([]Widget, 0, len(b.items))
	for idx := range b.items {
		item := b.items[idx]
		isActive := item.Key == activeKey
		col := inactiveColor
		tabBg := color.NRGBA{}
		if isActive {
			col = activeColor
			tabBg = cs.SecondaryContainer
		}

		icon := item.Icon
		if icon == nil {
			icon = Text("•", TextColor(col))
		}
		tab := layoutWidgetFunc(func(tabCtx *internal.Context) layout.Dimensions {
			clickable := event.UseClickable(tabCtx)
			for clickable.Clicked(tabCtx) {
				activeKey = item.Key
				if b.config.onChange != nil {
					b.config.onChange(tabCtx, item.Key)
				}
			}
			bg := tabBg
			if clickable.Pressed() {
				bg = style.StateLayer(bg, col, style.StateLayerPressedOpacity)
			} else if clickable.Hovered() {
				bg = style.StateLayer(bg, col, style.StateLayerHoverOpacity)
			}
			itemContent := ContainerDecoration(
				style.Decoration{}.WithBg(bg).WithPad(style.Symmetric(6, 10)).WithRad(tabCtx.Theme().Shapes.Full),
				Column(
					icon,
					Padding(style.Insets{Top: 4}, Text(item.Label, TextColor(col), TextType(tabCtx.Theme().Types.LabelMedium))),
				),
			)
			size := tabCtx.LayoutRippleArea(clickable.Handle(), internal.RippleSpec{
				Color:   col,
				Radius:  tabCtx.Theme().Shapes.Full,
				Opacity: style.StateLayerPressedOpacity,
			}, func(childCtx *internal.Context) image.Point {
				return itemContent.Layout(childCtx.Child(0)).Size
			})
			if clickable.Focused(tabCtx) {
				tabCtx.DrawFocusIndicator(size, internal.FocusIndicatorSpec{
					Color:  cs.Primary,
					Radius: tabCtx.Theme().Shapes.Full,
				})
			}
			return layout.Dimensions{Size: size}
		})
		tabs = append(tabs, tab)
	}

	content := b.layoutTabs(tabs)
	if len(tabs) > 6 {
		content = ScrollView(
			content,
			ScrollHorizontal(true),
			ScrollVertical(false),
		)
	}

	return expandWidth(ContainerDecoration(
		style.Decoration{}.WithBg(bg).WithPad(style.Symmetric(8, 8)),
		content,
	)).Layout(ctx.Child(0))
}

func (b *bottomNavWidget) layoutTabs(tabs []Widget) Widget {
	if len(tabs) == 0 {
		return Row()
	}

	switch b.config.alignment {
	case BottomNavAlignStart:
		row := make([]Widget, 0, len(tabs)+1)
		row = append(row, tabs...)
		row = append(row, Expanded(Spacer(0, 0)))
		return Row(row...)
	case BottomNavAlignEnd:
		row := make([]Widget, 0, len(tabs)+1)
		row = append(row, Expanded(Spacer(0, 0)))
		row = append(row, tabs...)
		return Row(row...)
	case BottomNavAlignCenter:
		row := make([]Widget, 0, len(tabs)+2)
		row = append(row, Expanded(Spacer(0, 0)))
		row = append(row, tabs...)
		row = append(row, Expanded(Spacer(0, 0)))
		return Row(row...)
	default:
		if len(tabs) == 1 {
			return Row(Expanded(tabs[0]))
		}
		children := make([]Widget, 0, len(tabs)*2+1)
		children = append(children, Expanded(Spacer(0, 0)))
		for i, tab := range tabs {
			children = append(children, tab)
			if i < len(tabs)-1 {
				children = append(children, Expanded(Spacer(0, 0)))
			}
		}
		children = append(children, Expanded(Spacer(0, 0)))
		return Row(children...)
	}
}
