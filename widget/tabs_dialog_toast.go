package widget

import (
	"image"
	"image/color"
	"time"

	"github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// TabItem 标签项。
type TabItem struct {
	Key      string
	Label    string
	Icon     Widget
	Disabled bool
}

// TabsOption 标签配置。
type TabsOption func(*tabsConfig)

type tabsConfig struct {
	onChange        func(ctx *internal.Context, key string)
	scrollable      bool
	secondary       bool
	autoActivate    bool
	inlineIcon      bool
	fullWidth       bool
	indicatorColor  color.NRGBA
	hasIndicator    bool
	textColor       color.NRGBA
	hasTextColor    bool
	activeTextColor color.NRGBA
	hasActiveColor  bool
	ref             *TabsRef
	decoration      style.Decoration
	tabDecoration   style.Decoration
}

type tabsWidget struct {
	active string
	items  []TabItem
	config tabsConfig
}

type tabsRuntimeState struct {
	activeIndex       int
	previousIndex     int
	initialized       bool
	lastIndicatorRect []tabsIndicatorRect
	lastActiveKey     string
}

type tabsIndicatorRect struct {
	X      int
	Width  int
	Height int
	Radius int
}

type tabsRowWidget struct {
	children          []Widget
	activeIndex       int
	indicatorColor    color.NRGBA
	secondary         bool
	scrollable        bool
	fullWidth         bool
	state             *tabsRuntimeState
	indicatorProgress float32
}

type tabsTabSurfaceWidget struct {
	child      Widget
	decoration style.Decoration
	minWidth   float32
	minHeight  float32
}

const tabsIndicatorDuration = 250 * time.Millisecond

// Tabs 创建标签栏。
func Tabs(active string, items []TabItem, opts ...TabsOption) Widget {
	cfg := tabsConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &tabsWidget{
		active: active,
		items:  append([]TabItem(nil), items...),
		config: cfg,
	}
}

func TabsOnChange(fn func(ctx *internal.Context, key string)) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.onChange = fn
	}
}

func TabsScrollable(scrollable bool) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.scrollable = scrollable
	}
}

func TabsSecondary(secondary bool) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.secondary = secondary
	}
}

func TabsPrimary(primary bool) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.secondary = !primary
	}
}

func TabsAutoActivate(autoActivate bool) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.autoActivate = autoActivate
	}
}

func TabsInlineIcon(inline bool) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.inlineIcon = inline
	}
}

func TabsFullWidth(fullWidth bool) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.fullWidth = fullWidth
	}
}

func TabsIndicatorColor(col color.NRGBA) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.indicatorColor = col
		cfg.hasIndicator = true
	}
}

func TabsTextColor(col color.NRGBA) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.textColor = col
		cfg.hasTextColor = true
	}
}

func TabsActiveTextColor(col color.NRGBA) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.activeTextColor = col
		cfg.hasActiveColor = true
	}
}

// TabsAttachRef 绑定命令型引用，用于外部主动切换标签页。
func TabsAttachRef(ref *TabsRef) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.ref = ref
	}
}

// TabsDecoration 通过 Decoration 统一设置标签栏外层装饰。
func TabsDecoration(d style.Decoration) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.decoration = d
	}
}

// TabsTabDecoration 通过 Decoration 统一设置单个标签项装饰和交互态。
func TabsTabDecoration(d style.Decoration) TabsOption {
	return func(cfg *tabsConfig) {
		cfg.tabDecoration = d
	}
}

func (t *tabsWidget) Layout(ctx *internal.Context) layout.Dimensions {
	activeKey := t.active
	if t.config.ref != nil {
		t.config.ref.bindInvalidator(redrawInvalidator(ctx))
		for _, key := range t.config.ref.drainCommands() {
			if key == activeKey {
				continue
			}
			activeKey = key
			if t.config.onChange != nil {
				t.config.onChange(ctx, key)
			}
		}
	}

	cs := ctx.Theme().Colors
	normalText := cs.OnSurfaceVariant
	if t.config.hasTextColor {
		normalText = t.config.textColor
	}
	activeText := cs.Primary
	if t.config.hasActiveColor {
		activeText = t.config.activeTextColor
	}
	indicator := cs.Primary
	if t.config.hasIndicator {
		indicator = t.config.indicatorColor
	}

	activeIndex := t.activeIndex(activeKey)
	state := tabsRuntimeStateFor(ctx)
	if !state.initialized {
		state.initialized = true
		state.activeIndex = activeIndex
		state.previousIndex = activeIndex
	} else if activeIndex >= 0 && activeIndex != state.activeIndex {
		state.previousIndex = state.activeIndex
		state.activeIndex = activeIndex
		md3FloatStateFor(ctx, "tabs-indicator-transform", 0).snap(ctx.Now(), 0, tabsIndicatorDuration, style.InteractionEmphasizedEasing)
	}
	indicatorProgress := md3FloatStateFor(ctx, "tabs-indicator-transform", 1).advance(ctx, 1, tabsIndicatorDuration, style.InteractionEmphasizedEasing)

	children := make([]Widget, 0, len(t.items))
	for idx := range t.items {
		item := t.items[idx]
		active := item.Key == activeKey

		tab := layoutWidgetFunc(func(tabCtx *internal.Context) layout.Dimensions {
			clickable := event.UseClickable(tabCtx)
			disabled := item.Disabled
			drainClickableDefaultAction(tabCtx, clickable, disabled, func(actionCtx *internal.Context) {
				if item.Key == activeKey {
					return
				}
				activeKey = item.Key
				if t.config.onChange != nil {
					t.config.onChange(actionCtx, item.Key)
				}
			})
			snapshot := clickableInteractionSnapshot(tabCtx, clickable, disabled, !disabled)
			txtColor := normalText
			if active {
				txtColor = activeText
			}
			if disabled {
				txtColor = style.DisabledContent(cs.OnSurface)
			}
			duration, easing := md3InteractionTiming(tabCtx, snapshot.Hovered, snapshot.Pressed, snapshot.Focused, disabled)
			txtColor = md3AnimateColor(tabCtx, "tab-text", txtColor, style.InteractionSelectedDuration, style.InteractionStandardEasing)
			stateOpacity := materialAnimatedStateLayerOpacity(tabCtx, snapshot.Hovered, snapshot.Pressed, disabled)
			stateColor := indicator
			if !active {
				stateColor = cs.OnSurface
			}
			tabBg := style.StateLayer(color.NRGBA{}, stateColor, stateOpacity)
			tabBg = md3AnimateColor(tabCtx, "tab-bg", tabBg, duration, easing)
			tabContent := t.layoutMaterialTabContent(item, txtColor, tabBg)
			size := tabCtx.LayoutRippleArea(clickable.Handle(), internal.RippleSpec{
				Color:   txtColor,
				Radius:  8,
				Opacity: style.StateLayerPressedOpacity,
			}, func(childCtx *internal.Context) image.Point {
				return tabContent.Layout(childCtx.Child(0)).Size
			})
			md3DrawFocusIndicator(tabCtx, size, internal.FocusIndicatorSpec{
				Color:  indicator,
				Radius: 8,
			}, snapshot.Focused, disabled)
			return layout.Dimensions{Size: size}
		})
		children = append(children, tab)
	}

	tabRow := &tabsRowWidget{
		children:          children,
		activeIndex:       activeIndex,
		indicatorColor:    indicator,
		secondary:         t.config.secondary,
		scrollable:        t.config.scrollable,
		fullWidth:         t.config.fullWidth,
		state:             state,
		indicatorProgress: indicatorProgress,
	}
	var row Widget = layoutWidgetFunc(func(rowCtx *internal.Context) layout.Dimensions {
		dims := tabRow.Layout(rowCtx.Child(0))
		line := rowCtx.Gtx.Dp(safeDp(1))
		if line < 1 {
			line = 1
		}
		paint.FillShape(rowCtx.Gtx.Ops, cs.OutlineVariant, clip.Rect(image.Rect(0, dims.Size.Y-line, dims.Size.X, dims.Size.Y)).Op())
		return dims
	})
	if hasDecorationVisual(t.config.decoration) {
		row = ContainerDecoration(t.config.decoration, row)
	}
	if t.config.scrollable {
		return ScrollView(
			row,
			ScrollHorizontal(true),
			ScrollVertical(false),
			ScrollBarVisible(false),
		).Layout(ctx.Child(0))
	}
	return row.Layout(ctx.Child(0))
}

func (r *tabsRowWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if len(r.children) == 0 {
		return layout.Dimensions{}
	}

	availableW := ctx.Gtx.Constraints.Max.X
	if availableW < 0 {
		availableW = 0
	}
	equalWidth := r.fullWidth || !r.scrollable
	tabRects := make([]image.Rectangle, len(r.children))
	indicatorRects := make([]tabsIndicatorRect, len(r.children))

	x := 0
	maxH := 0
	for idx, child := range r.children {
		if child == nil {
			continue
		}
		childCtx := *ctx
		childCtx.Gtx.Constraints.Min.X = 0
		if equalWidth && availableW > 0 {
			w := availableW / len(r.children)
			if idx == len(r.children)-1 {
				w = availableW - x
			}
			if w < 0 {
				w = 0
			}
			childCtx.Gtx.Constraints.Min.X = w
			childCtx.Gtx.Constraints.Max.X = w
		}
		childOffset := op.Offset(image.Point{X: x}).Push(ctx.Gtx.Ops)
		next := *ctx
		next.Gtx = childCtx.Gtx
		next = *next.WithPositionOffset(image.Point{X: x})
		dims := child.Layout(next.Child(idx))
		childOffset.Pop()

		w := dims.Size.X
		if equalWidth && availableW > 0 {
			w = childCtx.Gtx.Constraints.Max.X
		}
		tabRects[idx] = image.Rect(x, 0, x+w, dims.Size.Y)
		indicatorRects[idx] = r.indicatorRectFor(ctx, tabRects[idx])
		x += w
		if dims.Size.Y > maxH {
			maxH = dims.Size.Y
		}
	}

	if equalWidth && availableW > x {
		x = availableW
	}
	if maxH <= 0 {
		return layout.Dimensions{}
	}

	r.drawIndicator(ctx, maxH, indicatorRects)
	r.rememberIndicatorRects(indicatorRects)
	return layout.Dimensions{Size: image.Point{X: x, Y: maxH}}
}

func (r *tabsRowWidget) indicatorRectFor(ctx *internal.Context, tab image.Rectangle) tabsIndicatorRect {
	h := ctx.Gtx.Dp(safeDp(3))
	w := ctx.Gtx.Dp(safeDp(24))
	radius := ctx.Gtx.Dp(safeDp(3))
	if r.secondary {
		h = ctx.Gtx.Dp(safeDp(2))
		w = tab.Dx()
		radius = 0
	}
	if h < 1 {
		h = 1
	}
	if w < 1 {
		w = 1
	}
	x := tab.Min.X + (tab.Dx()-w)/2
	if r.secondary {
		x = tab.Min.X
	}
	return tabsIndicatorRect{X: x, Width: w, Height: h, Radius: radius}
}

func (r *tabsRowWidget) drawIndicator(ctx *internal.Context, rowHeight int, rects []tabsIndicatorRect) {
	if r.activeIndex < 0 || r.activeIndex >= len(rects) {
		return
	}
	to := rects[r.activeIndex]
	from := to
	if r.state != nil && r.state.previousIndex >= 0 && r.state.previousIndex < len(rects) && r.state.previousIndex != r.activeIndex {
		if r.state.previousIndex < len(r.state.lastIndicatorRect) {
			from = r.state.lastIndicatorRect[r.state.previousIndex]
		} else {
			from = rects[r.state.previousIndex]
		}
	}

	x, w := animatedTabIndicatorBounds(from, to, r.indicatorProgress)
	if w < 1 {
		w = 1
	}
	y := rowHeight - to.Height
	if y < 0 {
		y = 0
	}
	paint.FillShape(ctx.Gtx.Ops, r.indicatorColor, clip.UniformRRect(image.Rect(x, y, x+w, y+to.Height), to.Radius).Op(ctx.Gtx.Ops))
}

func (r *tabsRowWidget) rememberIndicatorRects(rects []tabsIndicatorRect) {
	if r.state == nil {
		return
	}
	if cap(r.state.lastIndicatorRect) < len(rects) {
		r.state.lastIndicatorRect = make([]tabsIndicatorRect, len(rects))
	} else {
		r.state.lastIndicatorRect = r.state.lastIndicatorRect[:len(rects)]
	}
	copy(r.state.lastIndicatorRect, rects)
}

func animatedTabIndicatorBounds(from, to tabsIndicatorRect, progress float32) (x, width int) {
	p := clampFloat32(progress, 0, 1)
	fromL := float32(from.X)
	fromR := float32(from.X + from.Width)
	toL := float32(to.X)
	toR := float32(to.X + to.Width)
	if p >= 1 || (fromL == toL && fromR == toR) {
		return to.X, to.Width
	}

	leftP := p
	rightP := p
	if toL > fromL {
		leftP = tabsIndicatorTailProgress(p)
		rightP = tabsIndicatorLeadProgress(p)
	} else if toL < fromL {
		leftP = tabsIndicatorLeadProgress(p)
		rightP = tabsIndicatorTailProgress(p)
	}
	left := fromL + (toL-fromL)*leftP
	right := fromR + (toR-fromR)*rightP
	if right < left {
		left, right = right, left
	}
	return int(left + 0.5), int(right - left + 0.5)
}

func tabsIndicatorLeadProgress(p float32) float32 {
	p = clampFloat32(p, 0, 1)
	return 1 - (1-p)*(1-p)*(1-p)
}

func tabsIndicatorTailProgress(p float32) float32 {
	p = clampFloat32(p, 0, 1)
	return p * p * (3 - 2*p)
}

func tabsRuntimeStateFor(ctx *internal.Context) *tabsRuntimeState {
	value := ctx.Memo("tabs-state", func() any {
		return &tabsRuntimeState{activeIndex: -1, previousIndex: -1}
	})
	state, ok := value.(*tabsRuntimeState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: tabs state type mismatch")
	}
	return state
}

func (t *tabsWidget) activeIndex(key string) int {
	for idx, item := range t.items {
		if item.Key == key {
			return idx
		}
	}
	if len(t.items) > 0 {
		return 0
	}
	return -1
}

func (t *tabsWidget) layoutMaterialTabContent(item TabItem, txtColor color.NRGBA, tabBg color.NRGBA) Widget {
	return layoutWidgetFunc(func(tabCtx *internal.Context) layout.Dimensions {
		minW := t.materialTabMinWidth(item)
		minH := float32(48)
		if !t.config.secondary && item.Icon != nil && item.Label != "" && !t.config.inlineIcon {
			minH = 64
		}
		if t.config.secondary {
			minH = 48
		}
		deco := style.Decoration{}.
			WithBg(t.config.tabDecoration.ResolveBg(tabBg)).
			WithPad(t.config.tabDecoration.ResolvePad(style.Insets{Left: 16, Right: 16})).
			WithRad(t.config.tabDecoration.ResolveRad(0))
		if t.config.tabDecoration.Border != nil {
			deco = deco.WithBorder(*t.config.tabDecoration.Border)
		}
		return (&tabsTabSurfaceWidget{
			child:      t.tabLabelContent(tabCtx, item, txtColor),
			decoration: deco,
			minWidth:   minW,
			minHeight:  minH,
		}).Layout(tabCtx.Child(0))
	})
}

func (t *tabsWidget) materialTabMinWidth(item TabItem) float32 {
	minW := float32(90)
	if !t.config.scrollable || t.config.fullWidth {
		return minW
	}
	contentW := float32(0)
	if item.Icon != nil {
		contentW += 24
	}
	if item.Label != "" {
		labelW := float32(runeCount(item.Label))*8 + 2
		if t.config.secondary || t.config.inlineIcon {
			if item.Icon != nil {
				contentW += 8
			}
			contentW += labelW
		} else if labelW > contentW {
			contentW = labelW
		}
	}
	need := contentW + 32
	if need > minW {
		minW = need
	}
	return minW
}

func runeCount(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func (t *tabsWidget) tabLabelContent(ctx *internal.Context, item TabItem, txtColor color.NRGBA) Widget {
	label := Text(item.Label, TextColor(txtColor), TextType(ctx.Theme().Types.LabelLarge))
	if item.Icon == nil {
		return label
	}
	icon := FixedWidth(24, Center(withForeground(txtColor, item.Icon)))
	if item.Label == "" {
		return icon
	}
	if t.config.secondary || t.config.inlineIcon {
		return tabsInlineContent(icon, label)
	}
	return tabsStackedContent(icon, label)
}

func (t *tabsTabSurfaceWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if t == nil {
		return layout.Dimensions{}
	}
	gtx := ctx.Gtx
	size := image.Point{
		X: gtx.Dp(safeDp(t.minWidth)),
		Y: gtx.Dp(safeDp(t.minHeight)),
	}
	if size.X < gtx.Constraints.Min.X {
		size.X = gtx.Constraints.Min.X
	}
	if size.Y < gtx.Constraints.Min.Y {
		size.Y = gtx.Constraints.Min.Y
	}
	size = clampPointToConstraints(size, gtx.Constraints.Min, gtx.Constraints.Max)

	bg := t.decoration.ResolveBg(color.NRGBA{})
	radius := gtx.Dp(safeDp(t.decoration.ResolveRad(0)))
	if bg.A > 0 {
		paint.FillShape(gtx.Ops, bg, clip.UniformRRect(image.Rectangle{Max: size}, clampRRectRadiusPx(size, radius)).Op(gtx.Ops))
	}
	border := t.decoration.ResolveBorder(style.Border{})
	drawTabsSurfaceBorder(gtx, size, radius, border)

	if t.child == nil {
		return layout.Dimensions{Size: size}
	}
	pad := t.decoration.ResolvePad(style.Insets{})
	left := gtx.Dp(safeDp(pad.Left))
	top := gtx.Dp(safeDp(pad.Top))
	right := gtx.Dp(safeDp(pad.Right))
	bottom := gtx.Dp(safeDp(pad.Bottom))
	contentSize := image.Point{X: size.X - left - right, Y: size.Y - top - bottom}
	if contentSize.X < 0 {
		contentSize.X = 0
	}
	if contentSize.Y < 0 {
		contentSize.Y = 0
	}

	offset := op.Offset(image.Point{X: left, Y: top}).Push(gtx.Ops)
	contentGtx := gtx
	contentGtx.Constraints = gioLayout.Exact(contentSize)
	next := *ctx
	next.Gtx = contentGtx
	gioLayout.Center.Layout(contentGtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		centered := next
		centered.Gtx = gtx
		dims := t.child.Layout(centered.Child(0))
		return gioLayout.Dimensions{Size: dims.Size}
	})
	offset.Pop()

	return layout.Dimensions{Size: size}
}

func drawTabsSurfaceBorder(gtx gioLayout.Context, size image.Point, radius int, border style.Border) {
	if size.X <= 0 || size.Y <= 0 || border.IsZero() {
		return
	}
	width := float32(gtx.Dp(safeDp(border.Width)))
	if width <= 0 {
		width = 1
	}
	inset := width / 2
	minX := inset
	minY := inset
	maxX := float32(size.X) - inset
	maxY := float32(size.Y) - inset
	if maxX <= minX || maxY <= minY {
		return
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	rr := float32(clampRRectRadiusPx(size, radius))
	if rr <= 0 {
		path.MoveTo(f32.Pt(minX, minY))
		path.LineTo(f32.Pt(maxX, minY))
		path.LineTo(f32.Pt(maxX, maxY))
		path.LineTo(f32.Pt(minX, maxY))
		path.Close()
	} else {
		path.MoveTo(f32.Pt(minX+rr, minY))
		path.LineTo(f32.Pt(maxX-rr, minY))
		path.CubeTo(f32.Pt(maxX, minY), f32.Pt(maxX, minY), f32.Pt(maxX, minY+rr))
		path.LineTo(f32.Pt(maxX, maxY-rr))
		path.CubeTo(f32.Pt(maxX, maxY), f32.Pt(maxX, maxY), f32.Pt(maxX-rr, maxY))
		path.LineTo(f32.Pt(minX+rr, maxY))
		path.CubeTo(f32.Pt(minX, maxY), f32.Pt(minX, maxY), f32.Pt(minX, maxY-rr))
		path.LineTo(f32.Pt(minX, minY+rr))
		path.CubeTo(f32.Pt(minX, minY), f32.Pt(minX, minY), f32.Pt(minX+rr, minY))
		path.Close()
	}
	paint.FillShape(gtx.Ops, border.Color, clip.Stroke{Path: path.End(), Width: width}.Op())
}

func tabsInlineContent(icon, label Widget) Widget {
	return layoutWidgetFunc(func(ctx *internal.Context) layout.Dimensions {
		dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(ctx.Gtx,
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := *ctx
				next.Gtx = gtx
				dims := icon.Layout(next.Child(0))
				return gioLayout.Dimensions{Size: dims.Size}
			}),
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				return gioLayout.Spacer{Width: safeDp(8)}.Layout(gtx)
			}),
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := *ctx
				next.Gtx = gtx
				dims := label.Layout(next.Child(1))
				return gioLayout.Dimensions{Size: dims.Size}
			}),
		)
		return layout.Dimensions{Size: dims.Size}
	})
}

func tabsStackedContent(icon, label Widget) Widget {
	return layoutWidgetFunc(func(ctx *internal.Context) layout.Dimensions {
		dims := gioLayout.Flex{Axis: gioLayout.Vertical, Alignment: gioLayout.Middle}.Layout(ctx.Gtx,
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := *ctx
				next.Gtx = gtx
				dims := icon.Layout(next.Child(0))
				return gioLayout.Dimensions{Size: dims.Size}
			}),
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				return gioLayout.Spacer{Height: safeDp(2)}.Layout(gtx)
			}),
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := *ctx
				next.Gtx = gtx
				dims := label.Layout(next.Child(1))
				return gioLayout.Dimensions{Size: dims.Size}
			}),
		)
		return layout.Dimensions{Size: dims.Size}
	})
}

// DialogOption 对话框配置。
type DialogOption func(*dialogConfig)

// DialogType 定义对话框语义类型。
type DialogType int

const (
	DialogTypeBasic DialogType = iota
	DialogTypeAlert
	DialogTypeFullscreen
)

// DialogMount 定义对话框遮罩和面板的挂载范围。
type DialogMount int

const (
	// DialogMountGlobal 按当前窗口/根布局可用区域绘制遮罩。Dialog 应放在根 Stack
	// 的顶层以获得真正全局的 modal 行为。
	DialogMountGlobal DialogMount = iota
	// DialogMountLocal 只覆盖 Dialog 所在父布局分配到的区域。
	DialogMountLocal
)

type dialogConfig struct {
	title        string
	width        float32
	height       float32
	radius       float32
	maskClosable bool
	quick        bool
	kind         DialogType
	mount        DialogMount
	icon         Widget
	headline     Widget
	actions      []Widget
	confirmText  string
	cancelText   string
	onOpenChange func(ctx *internal.Context, open bool)
	onConfirm    func(ctx *internal.Context)
	onCancel     func(ctx *internal.Context)
	onOpen       func(ctx *internal.Context)
	onOpened     func(ctx *internal.Context)
	onClose      func(ctx *internal.Context)
	onClosed     func(ctx *internal.Context)
	onCancelOnly func(ctx *internal.Context)
	ref          *DialogRef
	decoration   style.Decoration
	maskColor    color.NRGBA
	hasMaskColor bool
}

type dialogWidget struct {
	open   bool
	child  Widget
	config dialogConfig
}

type dialogState struct {
	wasOpen         bool
	visibleWasOpen  bool
	restoreFocus    internal.PathID
	hasRestoreFocus bool
}

// Dialog 创建对话框。
func Dialog(open bool, child Widget, opts ...DialogOption) Widget {
	cfg := dialogConfig{
		maskClosable: true,
		confirmText:  "确定",
		cancelText:   "取消",
		mount:        DialogMountGlobal,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &dialogWidget{
		open:   open,
		child:  child,
		config: cfg,
	}
}

func DialogTitle(title string) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.title = title
	}
}

func DialogWidth(width float32) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.width = width
	}
}

// DialogHeight 设置对话框固定高度。0 表示根据内容自适应并受窗口最大高度约束。
func DialogHeight(height float32) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.height = height
	}
}

func DialogRadius(radius float32) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.radius = radius
	}
}

func DialogMaskClosable(maskClosable bool) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.maskClosable = maskClosable
	}
}

// DialogQuick 跳过打开/关闭动画，对齐 Material Web 的 quick 属性。
func DialogQuick(quick bool) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.quick = quick
	}
}

// DialogTypeOf 设置 basic/alert/fullscreen 语义。
func DialogTypeOf(kind DialogType) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.kind = kind
	}
}

// DialogMountMode 设置遮罩和面板覆盖范围。
func DialogMountMode(mount DialogMount) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.mount = mount
	}
}

// DialogGlobalOverlay 控制是否按根窗口区域绘制。默认 true。
func DialogGlobalOverlay(global bool) DialogOption {
	return func(cfg *dialogConfig) {
		if global {
			cfg.mount = DialogMountGlobal
		} else {
			cfg.mount = DialogMountLocal
		}
	}
}

// DialogIcon 设置标题上方图标。
func DialogIcon(icon Widget) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.icon = icon
	}
}

// DialogHeadline 设置自定义 headline 内容；优先级高于 DialogTitle。
func DialogHeadline(headline Widget) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.headline = headline
	}
}

// DialogActions 设置底部 action 区域；设置后不会自动生成确认/取消按钮。
func DialogActions(actions ...Widget) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.actions = append([]Widget(nil), actions...)
	}
}

func DialogOnOpenChange(fn func(ctx *internal.Context, open bool)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onOpenChange = fn
	}
}

func DialogOnConfirm(fn func(ctx *internal.Context)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onConfirm = fn
	}
}

func DialogOnCancel(fn func(ctx *internal.Context)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onCancel = fn
	}
}

func DialogOnOpen(fn func(ctx *internal.Context)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onOpen = fn
	}
}

func DialogOnOpened(fn func(ctx *internal.Context)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onOpened = fn
	}
}

func DialogOnClose(fn func(ctx *internal.Context)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onClose = fn
	}
}

func DialogOnClosed(fn func(ctx *internal.Context)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onClosed = fn
	}
}

// DialogOnCancelOnly 在 scrim/Escape 取消时调用，不替代 DialogOnCancel 的按钮回调。
func DialogOnCancelOnly(fn func(ctx *internal.Context)) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.onCancelOnly = fn
	}
}

// DialogConfirmText 自定义确认按钮文案。
func DialogConfirmText(text string) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.confirmText = text
	}
}

// DialogCancelText 自定义取消按钮文案。
func DialogCancelText(text string) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.cancelText = text
	}
}

// DialogAttachRef 绑定命令型引用，用于外部主动打开/关闭对话框。
func DialogAttachRef(ref *DialogRef) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.ref = ref
	}
}

// DialogDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func DialogDecoration(d style.Decoration) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.decoration = d
	}
}

// DialogMaskColor 设置对话框遮罩颜色。
func DialogMaskColor(col color.NRGBA) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.maskColor = col
		cfg.hasMaskColor = true
	}
}

// DialogMaskAlpha 设置对话框遮罩透明度。
func DialogMaskAlpha(alpha uint8) DialogOption {
	return DialogMaskColor(color.NRGBA{A: alpha})
}

func dialogRememberRestoreFocus(ctx *internal.Context, state *dialogState) {
	if ctx == nil || ctx.Runtime() == nil || state == nil {
		return
	}
	state.restoreFocus = ctx.Runtime().FocusedTarget()
	state.hasRestoreFocus = true
}

func dialogRestoreFocus(ctx *internal.Context, state *dialogState) {
	if ctx == nil || ctx.Runtime() == nil || state == nil || !state.hasRestoreFocus {
		return
	}
	target := state.restoreFocus
	state.restoreFocus = 0
	state.hasRestoreFocus = false
	if target != 0 && ctx.Runtime().RequestFocus(ctx, target) {
		return
	}
	ctx.Runtime().BlurFocus(ctx, 0)
}

func (d *dialogWidget) Layout(ctx *internal.Context) layout.Dimensions {
	open := d.open
	if d.config.ref != nil {
		d.config.ref.bindInvalidator(redrawInvalidator(ctx))
		for _, cmd := range d.config.ref.drainCommands() {
			next := open
			switch cmd.kind {
			case boolCmdSet:
				next = cmd.value
			case boolCmdToggle:
				next = !open
			}
			open = next
		}
	}

	state := dialogStateFor(ctx)
	if state.wasOpen != open {
		if open {
			dialogRememberRestoreFocus(ctx, state)
			if d.config.onOpen != nil {
				d.config.onOpen(ctx)
			}
		}
		if !open && d.config.onClose != nil {
			d.config.onClose(ctx)
		}
		if !open {
			dialogRestoreFocus(ctx, state)
		}
		state.wasOpen = open
		if d.config.onOpenChange != nil {
			d.config.onOpenChange(ctx, open)
		}
	}
	enterDuration := 500 * time.Millisecond
	exitDuration := 150 * time.Millisecond
	if d.config.quick {
		enterDuration = 0
		exitDuration = 0
	}
	overlayProgress, overlayVisible := md3OverlayProgress(
		ctx,
		"dialog-overlay",
		open,
		enterDuration,
		exitDuration,
		style.InteractionEmphasizedEasing,
		style.InteractionEmphasizedAccelerateEasing,
	)
	if !overlayVisible {
		if state.visibleWasOpen && !open && d.config.onClosed != nil {
			d.config.onClosed(ctx)
		}
		state.visibleWasOpen = false
		return layout.Dimensions{}
	}
	if open && overlayProgress >= 1 && !state.visibleWasOpen {
		state.visibleWasOpen = true
		if d.config.onOpened != nil {
			d.config.onOpened(ctx)
		}
	}

	cs := ctx.Theme().Colors
	maskColor := withAlpha(cs.Scrim, 82)
	if d.config.hasMaskColor {
		maskColor = d.config.maskColor
	}
	maskColor = withAlphaFactor(maskColor, overlayProgress)
	mask := fillWidget(func(maskCtx *internal.Context, size image.Point) {
		if size.X <= 0 || size.Y <= 0 {
			return
		}
		paint.FillShape(maskCtx.Gtx.Ops, maskColor, clip.Rect(image.Rectangle{Max: size}).Op())
	}, open && d.config.maskClosable && d.config.onOpenChange != nil, func(maskCtx *internal.Context) {
		if d.config.onCancelOnly != nil {
			d.config.onCancelOnly(maskCtx)
		}
		d.config.onOpenChange(maskCtx, false)
	})

	sections := make([]dialogSection, 0, 4)
	if d.config.icon != nil {
		sections = append(sections, dialogSection{
			kind:  dialogSectionIcon,
			child: Center(FixedSize(24, 24, withForeground(cs.Secondary, d.config.icon))),
		})
	}
	headline := d.config.headline
	if headline == nil && d.config.title != "" {
		headline = Text(d.config.title, TextType(ctx.Theme().Types.HeadlineSmall), TextColor(cs.OnSurface))
	}
	if headline != nil {
		if d.config.icon != nil {
			headline = Center(headline)
		}
		sections = append(sections, dialogSection{kind: dialogSectionHeadline, child: headline})
	}
	if d.child != nil {
		sections = append(sections, dialogSection{kind: dialogSectionContent, child: withTextStyle(ctx.Theme().Types.BodyMedium, d.child)})
	}

	actions := d.config.actions
	if len(actions) == 0 {
		actions = make([]Widget, 0, 2)
		if d.config.onCancel != nil {
			actions = append(actions, TextButton(Text(d.config.cancelText, TextColor(cs.Primary), TextType(ctx.Theme().Types.LabelLarge)), ButtonForeground(cs.Primary), ButtonPadding(style.Symmetric(10, 12)), OnClick(d.config.onCancel)))
		}
		if d.config.onConfirm != nil {
			actions = append(actions, TextButton(Text(d.config.confirmText, TextColor(cs.Primary), TextType(ctx.Theme().Types.LabelLarge)), ButtonForeground(cs.Primary), ButtonPadding(style.Symmetric(10, 12)), OnClick(d.config.onConfirm)))
		}
	}
	if len(actions) > 0 {
		actionChildren := make([]Widget, 0, len(actions)+1)
		actionChildren = append(actionChildren, Expanded(Spacer(0, 0)))
		for index, action := range actions {
			if index > 0 {
				actionChildren = append(actionChildren, HSpacer(8))
			}
			actionChildren = append(actionChildren, action)
		}
		sections = append(sections, dialogSection{kind: dialogSectionActions, child: Row(actionChildren...)})
	}

	radiusDefault := md3DefaultSquircleRadius(d.config.decoration, ctx.Theme().Shapes.ExtraLarge, 36)
	if d.config.radius > 0 {
		radiusDefault = d.config.radius
	}
	panelDeco := materialDialogSurfaceDecoration(ctx, d.config.decoration, cs.SurfaceContainerHigh, style.All(24), radiusDefault)

	owner := ctx.PathID()
	content := anchoredOverlayWidget(layoutWidgetFunc(func(overlayCtx *internal.Context) layout.Dimensions {
		portalCtx := overlayCtx.Scope("dialog-portal")
		event.RegisterPortal(portalCtx, owner)
		event.RegisterBoundary(portalCtx, event.BoundaryStopPropagation())
		if open {
			event.OnKeyDown(portalCtx, func(keyCtx *internal.Context, ev *event.KeyboardEvent) {
				if keyCtx == nil || keyCtx.Runtime() == nil || ev == nil || ev.DefaultPrevented {
					return
				}
				switch ev.Key {
				case "Tab":
					ev.PreventDefault()
					direction := internal.FocusForward
					if ev.Modifiers.Shift {
						direction = internal.FocusBackward
					}
					keyCtx.Runtime().MoveFocusWithin(keyCtx, portalCtx.PathID(), direction)
				case "Escape":
					ev.PreventDefault()
					if d.config.onOpenChange != nil {
						if d.config.onCancelOnly != nil {
							d.config.onCancelOnly(keyCtx)
						}
						d.config.onOpenChange(keyCtx, false)
					}
				}
			})
		}
		panel := modalPanelInputShield(d.materialPanel(panelDeco, sections, overlayProgress))
		size := layoutMaterialDialogTransition(portalCtx, overlayProgress, d.config.quick, func(transitionCtx *internal.Context) image.Point {
			return panel.Layout(transitionCtx.Child(0)).Size
		})
		if open {
			md3EnsureFocusWithin(portalCtx)
			md3ProcessOverlayKeyboardEvents(portalCtx)
		}
		return layout.Dimensions{Size: size}
	}), gioLayout.Center)

	return Stack(
		mask,
		content,
	).Layout(ctx.Child(0))
}

type dialogSectionKind uint8

const (
	dialogSectionIcon dialogSectionKind = iota
	dialogSectionHeadline
	dialogSectionContent
	dialogSectionActions
)

type dialogSection struct {
	kind  dialogSectionKind
	child Widget
}

func (d *dialogWidget) materialPanel(deco style.Decoration, sections []dialogSection, progress float32) Widget {
	return layoutWidgetFunc(func(panelCtx *internal.Context) layout.Dimensions {
		gtx := panelCtx.Gtx
		max := gtx.Constraints.Max
		margin := gtx.Dp(safeDp(24))
		if margin < 0 {
			margin = 0
		}
		minW := gtx.Dp(safeDp(280))
		maxW := gtx.Dp(safeDp(560))
		if d.config.kind == DialogTypeFullscreen {
			minW = max.X - margin*2
			maxW = minW
		}
		if d.config.width > 0 {
			minW = gtx.Dp(safeDp(d.config.width))
			maxW = minW
		}
		if limit := max.X - margin*2; limit > 0 && maxW > limit {
			maxW = limit
		}
		if minW > maxW {
			minW = maxW
		}
		maxH := gtx.Dp(safeDp(560))
		if limit := max.Y - margin*2; limit > 0 && maxH > limit {
			maxH = limit
		}
		if d.config.kind == DialogTypeFullscreen {
			maxH = max.Y - margin*2
		}
		if d.config.height > 0 {
			maxH = gtx.Dp(safeDp(d.config.height))
			if limit := max.Y - margin*2; limit > 0 && maxH > limit {
				maxH = limit
			}
		}
		if maxH < gtx.Dp(safeDp(140)) {
			maxH = gtx.Dp(safeDp(140))
		}

		body := &dialogBodyWidget{sections: sections, progress: progress}
		panel := ContainerDecoration(deco, body)
		inner := gtx
		inner.Constraints.Min.X = minW
		inner.Constraints.Max.X = maxW
		inner.Constraints.Max.Y = maxH
		if d.config.height > 0 || d.config.kind == DialogTypeFullscreen {
			inner.Constraints.Min.Y = maxH
		}
		next := *panelCtx
		next.Gtx = inner
		return panel.Layout(next.Child(0))
	})
}

type dialogBodyWidget struct {
	sections []dialogSection
	progress float32
}

func (b *dialogBodyWidget) Layout(ctx *internal.Context) layout.Dimensions {
	children := make([]Widget, 0, len(b.sections))
	for _, section := range b.sections {
		if section.child == nil {
			continue
		}
		pad := materialDialogSectionPadding(section.kind, b.hasIcon(), b.hasHeadline(), b.hasActions())
		child := section.child
		if !pad.IsZero() {
			child = Padding(pad, child)
		}
		children = append(children, dialogSectionFade(section.kind, b.progress, child))
	}
	return Column(children...).Layout(ctx.Child(0))
}

func (b *dialogBodyWidget) hasHeadline() bool {
	for _, section := range b.sections {
		if section.kind == dialogSectionHeadline {
			return true
		}
	}
	return false
}

func (b *dialogBodyWidget) hasIcon() bool {
	for _, section := range b.sections {
		if section.kind == dialogSectionIcon {
			return true
		}
	}
	return false
}

func (b *dialogBodyWidget) hasActions() bool {
	for _, section := range b.sections {
		if section.kind == dialogSectionActions {
			return true
		}
	}
	return false
}

func materialDialogSectionPadding(kind dialogSectionKind, hasIcon, hasHeadline, hasActions bool) style.Insets {
	switch kind {
	case dialogSectionIcon:
		return style.Insets{Top: 24, Left: 24, Right: 24}
	case dialogSectionHeadline:
		top := float32(24)
		if hasIcon {
			top = 16
		}
		return style.Insets{Top: top, Left: 24, Right: 24}
	case dialogSectionContent:
		top := float32(24)
		bottom := float32(24)
		if hasHeadline {
			top = 16
		}
		if hasActions {
			bottom = 8
		}
		return style.Insets{Top: top, Left: 24, Right: 24, Bottom: bottom}
	case dialogSectionActions:
		return style.Insets{Top: 16, Right: 24, Bottom: 24, Left: 24}
	default:
		return style.All(0)
	}
}

func dialogSectionFade(kind dialogSectionKind, progress float32, child Widget) Widget {
	return layoutWidgetFunc(func(ctx *internal.Context) layout.Dimensions {
		p := clampFloat32(progress, 0, 1)
		start := float32(0.2)
		if kind == dialogSectionActions {
			start = 0.5
		}
		alpha := float32(1)
		if p < start {
			alpha = 0
		} else {
			alpha = clampFloat32((p-start)/(1-start), 0, 1)
		}
		if alpha < 1 {
			defer paint.PushOpacity(ctx.Gtx.Ops, alpha).Pop()
		}
		return child.Layout(ctx.Child(0))
	})
}

func layoutMaterialDialogTransition(ctx *internal.Context, progress float32, quick bool, child func(*internal.Context) image.Point) image.Point {
	if child == nil {
		return image.Point{}
	}
	if quick {
		return child(ctx.Child(0))
	}
	p := clampFloat32(progress, 0, 1)
	offsetY := -ctx.Gtx.Dp(safeDp(50 * (1 - p)))
	stack := op.Offset(image.Point{Y: offsetY}).Push(ctx.Gtx.Ops)
	defer stack.Pop()
	return child(ctx.Child(0))
}

func materialDialogSurfaceDecoration(ctx *internal.Context, d style.Decoration, defaultBg color.NRGBA, defaultPad style.Insets, defaultRadius float32) style.Decoration {
	cs := ctx.Theme().Colors
	deco := style.Decoration{}.
		WithBg(d.ResolveBg(defaultBg)).
		WithPad(d.ResolvePad(defaultPad)).
		WithRad(d.ResolveRad(defaultRadius))
	deco = md3ApplyOuterCornerShape(ctx, deco, d, d.ResolveRad(defaultRadius))
	if d.Shadow != nil {
		deco = deco.WithShadow(*d.Shadow)
	} else {
		deco = deco.WithShadow(style.ElevationShadow(cs, 3))
	}
	if d.Border != nil {
		deco = deco.WithBorder(*d.Border)
	}
	return deco
}

func dialogStateFor(ctx *internal.Context) *dialogState {
	value := ctx.Memo("dialog", func() any {
		return &dialogState{}
	})
	state, ok := value.(*dialogState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: dialog state type mismatch")
	}
	return state
}

// ToastType 吐司类型。
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastPosition 吐司位置。
type ToastPosition int

const (
	ToastTop ToastPosition = iota
	ToastCenter
	ToastBottom
)

// ToastOption 吐司配置。
type ToastOption func(*toastConfig)

type toastConfig struct {
	kind         ToastType
	duration     time.Duration
	position     ToastPosition
	onClose      func(ctx *internal.Context)
	actionLabel  string
	onAction     func(ctx *internal.Context)
	decoration   style.Decoration
	textColor    color.NRGBA
	hasTextColor bool
}

type toastWidget struct {
	message string
	config  toastConfig
}

type toastState struct {
	lastMessage string
	startAt     time.Time
	closed      bool
}

// Toast 创建吐司。
func Toast(message string, opts ...ToastOption) Widget {
	cfg := toastConfig{
		kind:     ToastInfo,
		duration: 2500 * time.Millisecond,
		position: ToastBottom,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &toastWidget{
		message: message,
		config:  cfg,
	}
}

func ToastTypeOf(t ToastType) ToastOption {
	return func(cfg *toastConfig) {
		cfg.kind = t
	}
}

func ToastDuration(duration time.Duration) ToastOption {
	return func(cfg *toastConfig) {
		cfg.duration = duration
	}
}

func ToastPositionOf(p ToastPosition) ToastOption {
	return func(cfg *toastConfig) {
		cfg.position = p
	}
}

func ToastOnClose(fn func(ctx *internal.Context)) ToastOption {
	return func(cfg *toastConfig) { cfg.onClose = fn }
}

func ToastAction(label string, fn func(ctx *internal.Context)) ToastOption {
	return func(cfg *toastConfig) {
		cfg.actionLabel = label
		cfg.onAction = fn
	}
}

// ToastDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func ToastDecoration(d style.Decoration) ToastOption {
	return func(cfg *toastConfig) {
		cfg.decoration = d
	}
}

// ToastTextColor 设置吐司文本颜色。
func ToastTextColor(col color.NRGBA) ToastOption {
	return func(cfg *toastConfig) {
		cfg.textColor = col
		cfg.hasTextColor = true
	}
}

func (t *toastWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if t.message == "" {
		return layout.Dimensions{}
	}

	state := toastStateFor(ctx)
	if state.lastMessage != t.message {
		state.lastMessage = t.message
		state.startAt = ctx.Now()
		state.closed = false
	}

	if t.config.duration > 0 && !state.closed {
		if ctx.Now().Sub(state.startAt) >= t.config.duration {
			state.closed = true
			if t.config.onClose != nil {
				t.config.onClose(ctx)
			}
		} else {
			ctx.RequestFrameRedrawReason("animation.toast")
		}
	}

	toastProgress, toastVisible := md3OverlayProgress(
		ctx,
		"toast-overlay",
		!state.closed,
		style.InteractionToastEnterDuration,
		style.InteractionToastExitDuration,
		style.InteractionEmphasizedDecelerateEasing,
		style.InteractionEmphasizedAccelerateEasing,
	)
	if !toastVisible {
		return layout.Dimensions{}
	}

	cs := ctx.Theme().Colors
	bg := t.config.decoration.ResolveBg(cs.InverseSurface)
	fg := cs.InverseOnSurface
	switch t.config.kind {
	case ToastSuccess:
		bg = t.config.decoration.ResolveBg(cs.Success)
		fg = cs.OnSuccess
	case ToastWarning:
		bg = t.config.decoration.ResolveBg(cs.Warning)
		fg = cs.OnWarning
	case ToastError:
		bg = t.config.decoration.ResolveBg(cs.Error)
		fg = cs.OnError
	case ToastInfo:
		fg = cs.InverseOnSurface
	}
	if t.config.hasTextColor {
		fg = t.config.textColor
	}

	content := Widget(Text(t.message, TextColor(fg), TextType(ctx.Theme().Types.BodyMedium)))
	if t.config.actionLabel != "" {
		action := TextButton(
			Text(t.config.actionLabel, TextColor(cs.InversePrimary), TextType(ctx.Theme().Types.LabelLarge)),
			ButtonForeground(cs.InversePrimary),
			ButtonPadding(style.Symmetric(6, 8)),
			OnClick(func(actionCtx *internal.Context) {
				if t.config.onAction != nil {
					t.config.onAction(actionCtx)
				}
				state.closed = true
				if t.config.onClose != nil {
					t.config.onClose(actionCtx)
				}
			}),
		)
		content = middleRow(
			Expanded(content),
			Padding(style.Insets{Left: 12}, action),
		)
	}

	toastRadius := md3DefaultSquircleRadius(t.config.decoration, ctx.Theme().Shapes.ExtraSmall, 14)
	body := ContainerDecoration(
		md3ApplyOuterCornerShape(ctx, style.Decoration{}.
			WithBg(bg).
			WithPad(t.config.decoration.ResolvePad(style.Symmetric(8, 16))).
			WithRad(toastRadius), t.config.decoration, toastRadius),
		content,
	)

	anchor := gioLayout.S
	offset := float32(8)
	switch t.config.position {
	case ToastTop:
		anchor = gioLayout.N
		offset = -8
	case ToastCenter:
		anchor = gioLayout.Center
		offset = 0
	}

	return anchoredOverlayWidget(
		layoutWidgetFunc(func(overlayCtx *internal.Context) layout.Dimensions {
			size := layoutMD3OverlayTransition(overlayCtx, toastProgress, offset, func(transitionCtx *internal.Context) image.Point {
				return Padding(style.Insets{Top: 8, Bottom: 8, Left: 8, Right: 8}, body).Layout(transitionCtx.Child(0)).Size
			})
			return layout.Dimensions{Size: size}
		}),
		anchor,
	).Layout(ctx.Child(0))
}

func toastStateFor(ctx *internal.Context) *toastState {
	value := ctx.Memo("toast-state", func() any {
		return &toastState{}
	})
	state, ok := value.(*toastState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: toast state type mismatch")
	}
	return state
}

type SnackbarOption = ToastOption

func Snackbar(message string, opts ...SnackbarOption) Widget {
	return Toast(message, opts...)
}

func SnackbarAction(label string, fn func(ctx *internal.Context)) SnackbarOption {
	return ToastAction(label, fn)
}

type overlayAnchorWidget struct {
	child  Widget
	anchor gioLayout.Direction
}

func anchoredOverlayWidget(child Widget, anchor gioLayout.Direction) Widget {
	return &overlayAnchorWidget{
		child:  child,
		anchor: anchor,
	}
}

func (o *overlayAnchorWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if o.child == nil {
		return layout.Dimensions{}
	}
	gtx := ctx.Gtx
	size := gtx.Constraints.Max
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = gioLayout.Exact(size)

	dims := o.anchor.Layout(gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := *ctx
		next.Gtx = gtx
		childDims := o.child.Layout(next.Child(0))
		return gioLayout.Dimensions{Size: childDims.Size}
	})
	_ = dims
	return layout.Dimensions{Size: size}
}

type fillWidgetDef struct {
	draw    func(ctx *internal.Context, size image.Point)
	click   bool
	onClick func(ctx *internal.Context)
}

func fillWidget(draw func(ctx *internal.Context, size image.Point), clickable bool, onClick func(ctx *internal.Context)) Widget {
	return &fillWidgetDef{
		draw:    draw,
		click:   clickable,
		onClick: onClick,
	}
}

func (f *fillWidgetDef) Layout(ctx *internal.Context) layout.Dimensions {
	gtx := ctx.Gtx
	size := gtx.Constraints.Max
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{}
	}

	layoutFill := func(current *internal.Context) layout.Dimensions {
		inner := current.Gtx
		inner.Constraints = gioLayout.Exact(size)
		next := *current
		next.Gtx = inner
		if f.draw != nil {
			f.draw(next.Child(0), size)
		}
		return layout.Dimensions{Size: size}
	}

	if !f.click || f.onClick == nil {
		return layoutFill(ctx.Child(0))
	}

	return Pressable(
		layoutWidgetFunc(func(btnCtx *internal.Context) layout.Dimensions {
			return layoutFill(btnCtx.Child(0))
		}),
		f.onClick,
	).Layout(ctx.Child(0))
}

// PopupOption 弹窗配置。
type PopupOption func(*popupConfig)

type modalPanelInputShieldWidget struct {
	child Widget
}

func modalPanelInputShield(child Widget) Widget {
	if child == nil {
		return nil
	}
	return &modalPanelInputShieldWidget{child: child}
}

func (w *modalPanelInputShieldWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	tag := ctx.Memo("modal-panel-input-shield", func() any {
		return new(int)
	})
	for {
		_, ok := ctx.Gtx.Event(pointer.Filter{
			Target:  tag,
			Kinds:   pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Enter | pointer.Leave | pointer.Cancel | pointer.Scroll,
			ScrollX: pointer.ScrollRange{Min: -1 << 30, Max: 1 << 30},
			ScrollY: pointer.ScrollRange{Min: -1 << 30, Max: 1 << 30},
		})
		if !ok {
			break
		}
	}

	macro := op.Record(ctx.Gtx.Ops)
	dims := w.child.Layout(ctx.Child(0))
	call := macro.Stop()
	if dims.Size.X > 0 && dims.Size.Y > 0 {
		area := clip.Rect(image.Rectangle{Max: dims.Size}).Push(ctx.Gtx.Ops)
		gioEvent.Op(ctx.Gtx.Ops, tag)
		area.Pop()
	}
	call.Add(ctx.Gtx.Ops)
	return dims
}

type popupConfig struct {
	width         float32
	height        float32
	radius        float32
	maskClosable  bool
	kind          DialogType
	mount         DialogMount
	background    color.NRGBA
	hasBackground bool
	padding       style.Insets
	hasPadding    bool
	onOpenChange  func(ctx *internal.Context, open bool)
	onOpen        func(ctx *internal.Context)
	onOpened      func(ctx *internal.Context)
	onClose       func(ctx *internal.Context)
	onClosed      func(ctx *internal.Context)
	onCancelOnly  func(ctx *internal.Context)
	ref           *PopupRef
	decoration    style.Decoration
	maskColor     color.NRGBA
	hasMaskColor  bool
	quick         bool
}

type popupWidget struct {
	open   bool
	child  Widget
	config popupConfig
}

// Popup 创建自定义内容弹窗，内部内容完全由用户定义。
func Popup(open bool, child Widget, opts ...PopupOption) Widget {
	cfg := popupConfig{
		maskClosable: true,
		mount:        DialogMountGlobal,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &popupWidget{
		open:   open,
		child:  child,
		config: cfg,
	}
}

// PopupWidth 设置弹窗宽度。
func PopupWidth(width float32) PopupOption {
	return func(cfg *popupConfig) {
		cfg.width = width
	}
}

// PopupHeight 设置弹窗固定高度。0 表示根据内容自适应并受窗口最大高度约束。
func PopupHeight(height float32) PopupOption {
	return func(cfg *popupConfig) {
		cfg.height = height
	}
}

// PopupRadius 设置弹窗圆角。
func PopupRadius(radius float32) PopupOption {
	return func(cfg *popupConfig) {
		cfg.radius = radius
	}
}

// PopupMaskClosable 设置点击遮罩是否关闭弹窗。
func PopupMaskClosable(maskClosable bool) PopupOption {
	return func(cfg *popupConfig) {
		cfg.maskClosable = maskClosable
	}
}

// PopupBackground 设置弹窗背景色。
func PopupBackground(bg color.NRGBA) PopupOption {
	return func(cfg *popupConfig) {
		cfg.background = bg
		cfg.hasBackground = true
	}
}

// PopupPadding 设置弹窗内边距。
func PopupPadding(insets style.Insets) PopupOption {
	return func(cfg *popupConfig) {
		cfg.padding = insets
		cfg.hasPadding = true
	}
}

// PopupQuick 跳过打开/关闭动画，与 DialogQuick 保持一致。
func PopupQuick(quick bool) PopupOption {
	return func(cfg *popupConfig) {
		cfg.quick = quick
	}
}

// PopupTypeOf 设置 basic/alert/fullscreen 语义，尺寸策略与 Dialog 保持一致。
func PopupTypeOf(kind DialogType) PopupOption {
	return func(cfg *popupConfig) {
		cfg.kind = kind
	}
}

// PopupMountMode 设置遮罩和面板覆盖范围。
func PopupMountMode(mount DialogMount) PopupOption {
	return func(cfg *popupConfig) {
		cfg.mount = mount
	}
}

// PopupGlobalOverlay 控制是否按根窗口区域绘制。默认 true。
func PopupGlobalOverlay(global bool) PopupOption {
	return func(cfg *popupConfig) {
		if global {
			cfg.mount = DialogMountGlobal
		} else {
			cfg.mount = DialogMountLocal
		}
	}
}

// PopupOnOpenChange 监听弹窗打开/关闭。
func PopupOnOpenChange(fn func(ctx *internal.Context, open bool)) PopupOption {
	return func(cfg *popupConfig) {
		cfg.onOpenChange = fn
	}
}

func PopupOnOpen(fn func(ctx *internal.Context)) PopupOption {
	return func(cfg *popupConfig) {
		cfg.onOpen = fn
	}
}

func PopupOnOpened(fn func(ctx *internal.Context)) PopupOption {
	return func(cfg *popupConfig) {
		cfg.onOpened = fn
	}
}

func PopupOnClose(fn func(ctx *internal.Context)) PopupOption {
	return func(cfg *popupConfig) {
		cfg.onClose = fn
	}
}

func PopupOnClosed(fn func(ctx *internal.Context)) PopupOption {
	return func(cfg *popupConfig) {
		cfg.onClosed = fn
	}
}

// PopupOnCancelOnly 在 scrim/Escape 取消时调用。
func PopupOnCancelOnly(fn func(ctx *internal.Context)) PopupOption {
	return func(cfg *popupConfig) {
		cfg.onCancelOnly = fn
	}
}

// PopupAttachRef 绑定命令型引用，用于外部主动打开/关闭弹窗。
func PopupAttachRef(ref *PopupRef) PopupOption {
	return func(cfg *popupConfig) {
		cfg.ref = ref
	}
}

// PopupDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func PopupDecoration(d style.Decoration) PopupOption {
	return func(cfg *popupConfig) {
		cfg.decoration = d
	}
}

// PopupMaskColor 设置弹窗遮罩颜色。
func PopupMaskColor(col color.NRGBA) PopupOption {
	return func(cfg *popupConfig) {
		cfg.maskColor = col
		cfg.hasMaskColor = true
	}
}

// PopupMaskAlpha 设置弹窗遮罩透明度。
func PopupMaskAlpha(alpha uint8) PopupOption {
	return PopupMaskColor(color.NRGBA{A: alpha})
}

func (p *popupWidget) Layout(ctx *internal.Context) layout.Dimensions {
	open := p.open
	if p.config.ref != nil {
		p.config.ref.bindInvalidator(redrawInvalidator(ctx))
		for _, cmd := range p.config.ref.drainCommands() {
			next := open
			switch cmd.kind {
			case boolCmdSet:
				next = cmd.value
			case boolCmdToggle:
				next = !open
			}
			open = next
		}
	}

	state := popupStateFor(ctx)
	if p.config.onOpenChange != nil && state.wasOpen != open {
		if open && p.config.onOpen != nil {
			p.config.onOpen(ctx)
		}
		if !open && p.config.onClose != nil {
			p.config.onClose(ctx)
		}
		state.wasOpen = open
		p.config.onOpenChange(ctx, open)
	}
	if p.config.onOpenChange == nil && state.wasOpen != open {
		if open && p.config.onOpen != nil {
			p.config.onOpen(ctx)
		}
		if !open && p.config.onClose != nil {
			p.config.onClose(ctx)
		}
		state.wasOpen = open
	}
	enterDuration := 500 * time.Millisecond
	exitDuration := 150 * time.Millisecond
	if p.config.quick {
		enterDuration = 0
		exitDuration = 0
	}
	overlayProgress, overlayVisible := md3OverlayProgress(
		ctx,
		"popup-overlay",
		open,
		enterDuration,
		exitDuration,
		style.InteractionEmphasizedEasing,
		style.InteractionEmphasizedAccelerateEasing,
	)
	if !overlayVisible {
		if state.visibleWasOpen && !open && p.config.onClosed != nil {
			p.config.onClosed(ctx)
		}
		state.visibleWasOpen = false
		return layout.Dimensions{}
	}
	if open && overlayProgress >= 1 && !state.visibleWasOpen {
		state.visibleWasOpen = true
		if p.config.onOpened != nil {
			p.config.onOpened(ctx)
		}
	}

	cs := ctx.Theme().Colors
	maskColor := withAlpha(cs.Scrim, 82)
	if p.config.hasMaskColor {
		maskColor = p.config.maskColor
	}
	maskColor = withAlphaFactor(maskColor, overlayProgress)
	mask := fillWidget(func(maskCtx *internal.Context, size image.Point) {
		if size.X <= 0 || size.Y <= 0 {
			return
		}
		paint.FillShape(maskCtx.Gtx.Ops, maskColor, clip.Rect(image.Rectangle{Max: size}).Op())
	}, open && p.config.maskClosable && p.config.onOpenChange != nil, func(maskCtx *internal.Context) {
		if p.config.onCancelOnly != nil {
			p.config.onCancelOnly(maskCtx)
		}
		p.config.onOpenChange(maskCtx, false)
	})

	bg := p.config.decoration.ResolveBg(cs.SurfaceContainerHigh)
	if p.config.hasBackground {
		bg = p.config.background
	}
	padding := p.config.decoration.ResolvePad(style.All(24))
	if p.config.hasPadding {
		padding = p.config.padding
	}
	radiusDefault := md3DefaultSquircleRadius(p.config.decoration, ctx.Theme().Shapes.ExtraLarge, 36)
	if p.config.radius > 0 {
		radiusDefault = p.config.radius
	}
	panelDeco := materialDialogSurfaceDecoration(ctx, p.config.decoration, bg, padding, radiusDefault)
	fullscreen := p.config.kind == DialogTypeFullscreen
	panelChild := p.child
	if panelChild != nil {
		panelChild = dialogSectionFade(dialogSectionContent, overlayProgress, panelChild)
	}
	panel := modalPanelInputShield(ContainerDecoration(panelDeco, panelChild))

	owner := ctx.PathID()
	content := anchoredOverlayWidget(layoutWidgetFunc(func(overlayCtx *internal.Context) layout.Dimensions {
		portalCtx := overlayCtx.Scope("popup-portal")
		event.RegisterPortal(portalCtx, owner)
		size := layoutMaterialDialogTransition(portalCtx, overlayProgress, p.config.quick, func(transitionCtx *internal.Context) image.Point {
			return layoutMaterialModalPanel(transitionCtx, p.config.width, p.config.height, fullscreen, panel)
		})
		return layout.Dimensions{Size: size}
	}), gioLayout.Center)

	return Stack(
		mask,
		content,
	).Layout(ctx.Child(0))
}

func layoutMaterialModalPanel(ctx *internal.Context, width float32, height float32, fullscreen bool, child Widget) image.Point {
	if child == nil {
		return image.Point{}
	}
	gtx := ctx.Gtx
	max := gtx.Constraints.Max
	margin := gtx.Dp(safeDp(24))
	if margin < 0 {
		margin = 0
	}
	minW := gtx.Dp(safeDp(280))
	maxW := gtx.Dp(safeDp(560))
	if fullscreen {
		minW = max.X - margin*2
		maxW = minW
	}
	if width > 0 {
		minW = gtx.Dp(safeDp(width))
		maxW = minW
	}
	if limit := max.X - margin*2; limit > 0 && maxW > limit {
		maxW = limit
	}
	if minW > maxW {
		minW = maxW
	}
	maxH := gtx.Dp(safeDp(560))
	if limit := max.Y - margin*2; limit > 0 && maxH > limit {
		maxH = limit
	}
	if fullscreen {
		maxH = max.Y - margin*2
	}
	if height > 0 {
		maxH = gtx.Dp(safeDp(height))
		if limit := max.Y - margin*2; limit > 0 && maxH > limit {
			maxH = limit
		}
	}
	if maxH < gtx.Dp(safeDp(140)) {
		maxH = gtx.Dp(safeDp(140))
	}

	inner := gtx
	inner.Constraints.Min.X = minW
	inner.Constraints.Max.X = maxW
	inner.Constraints.Max.Y = maxH
	if height > 0 || fullscreen {
		inner.Constraints.Min.Y = maxH
	}
	next := *ctx
	next.Gtx = inner
	return child.Layout(next.Child(0)).Size
}

type popupState struct {
	wasOpen        bool
	visibleWasOpen bool
}

func popupStateFor(ctx *internal.Context) *popupState {
	value := ctx.Memo("popup", func() any {
		return &popupState{}
	})
	state, ok := value.(*popupState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: popup state type mismatch")
	}
	return state
}
