package widget

import (
	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"

	gioLayout "gioui.org/layout"
)

// ScrollOption 定义滚动配置。
type ScrollOption func(*scrollConfig)

type scrollConfig struct {
	vertical   bool
	horizontal bool
	barVisible bool
	onChange   func(ctx *internal.Context, x, y float32)
}

type scrollWidget struct {
	child  Widget
	config scrollConfig
}

type scrollState struct {
	list      gioLayout.List
	lastFirst int
	lastOff   int
}

// ScrollView 创建滚动容器。
func ScrollView(child Widget, opts ...ScrollOption) Widget {
	cfg := scrollConfig{
		vertical:   true,
		horizontal: false,
		barVisible: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &scrollWidget{
		child:  child,
		config: cfg,
	}
}

func ScrollVertical(vertical bool) ScrollOption {
	return func(cfg *scrollConfig) {
		cfg.vertical = vertical
	}
}

func ScrollHorizontal(horizontal bool) ScrollOption {
	return func(cfg *scrollConfig) {
		cfg.horizontal = horizontal
	}
}

func ScrollBarVisible(visible bool) ScrollOption {
	return func(cfg *scrollConfig) {
		cfg.barVisible = visible
	}
}

func ScrollOnChange(fn func(ctx *internal.Context, x, y float32)) ScrollOption {
	return func(cfg *scrollConfig) {
		cfg.onChange = fn
	}
}

func (s *scrollWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if s.child == nil {
		return layout.Dimensions{}
	}

	state := scrollStateFor(ctx)
	state.list.Axis = resolveAxis(s.config.vertical, s.config.horizontal)

	dims := state.list.Layout(ctx.Gtx, 1, func(gtx gioLayout.Context, index int) gioLayout.Dimensions {
		next := *ctx
		next.Gtx = gtx
		childDims := s.child.Layout(next.Child(index))
		return gioLayout.Dimensions{Size: childDims.Size}
	})

	if s.config.onChange != nil {
		first := state.list.Position.First
		off := state.list.Position.Offset
		if first != state.lastFirst || off != state.lastOff {
			state.lastFirst = first
			state.lastOff = off
			if state.list.Axis == gioLayout.Horizontal {
				s.config.onChange(ctx, float32(first)+float32(off)/1024, 0)
			} else {
				s.config.onChange(ctx, 0, float32(first)+float32(off)/1024)
			}
		}
	}

	return layout.Dimensions{Size: dims.Size}
}

func scrollStateFor(ctx *internal.Context) *scrollState {
	value := ctx.Memo("scroll", func() any {
		return &scrollState{
			list: gioLayout.List{Axis: gioLayout.Vertical},
		}
	})
	state, ok := value.(*scrollState)
	if !ok {
		panic("fluxui/widget: scroll state type mismatch")
	}
	return state
}

// ListOption 定义列表配置。
type ListOption func(*listConfig)

type listConfig struct {
	axis        Axis
	virtualized bool
	itemSpacing float32
	padding     style.Insets
	onReachEnd  func(ctx *internal.Context)
}

type listViewWidget struct {
	count   int
	builder func(ctx *internal.Context, index int) Widget
	config  listConfig
}

type listViewState struct {
	list        gioLayout.List
	reachCalled bool
}

// ListView 创建列表组件。
func ListView(count int, itemBuilder func(ctx *internal.Context, index int) Widget, opts ...ListOption) Widget {
	cfg := listConfig{
		axis:        Vertical,
		virtualized: true,
		itemSpacing: 0,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &listViewWidget{
		count:   count,
		builder: itemBuilder,
		config:  cfg,
	}
}

func ListAxis(axis Axis) ListOption {
	return func(cfg *listConfig) {
		cfg.axis = axis
	}
}

func ListVirtualized(virtualized bool) ListOption {
	return func(cfg *listConfig) {
		cfg.virtualized = virtualized
	}
}

func ListItemSpacing(spacing float32) ListOption {
	return func(cfg *listConfig) {
		cfg.itemSpacing = spacing
	}
}

func ListPadding(insets style.Insets) ListOption {
	return func(cfg *listConfig) {
		cfg.padding = insets
	}
}

func ListOnReachEnd(fn func(ctx *internal.Context)) ListOption {
	return func(cfg *listConfig) {
		cfg.onReachEnd = fn
	}
}

func (l *listViewWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if l.builder == nil || l.count <= 0 {
		return layout.Dimensions{}
	}

	state := listViewStateFor(ctx)
	state.list.Axis = toGioAxis(l.config.axis)

	listChild := layoutWidgetFunc(func(listCtx *internal.Context) layout.Dimensions {
		dims := state.list.Layout(listCtx.Gtx, l.count, func(gtx gioLayout.Context, index int) gioLayout.Dimensions {
			next := *listCtx
			next.Gtx = gtx
			child := l.builder(next.Child(index), index)
			if child == nil {
				return gioLayout.Dimensions{}
			}
			if l.config.itemSpacing > 0 && index < l.count-1 {
				if l.config.axis == Horizontal {
					child = Padding(style.Insets{Right: l.config.itemSpacing}, child)
				} else {
					child = Padding(style.Insets{Bottom: l.config.itemSpacing}, child)
				}
			}
			childDims := child.Layout(next.Child(index))
			return gioLayout.Dimensions{Size: childDims.Size}
		})
		return layout.Dimensions{Size: dims.Size}
	})

	var root Widget = listChild
	if l.config.axis == Vertical {
		root = expandWidth(root)
	}
	if !l.config.padding.IsZero() {
		root = Padding(l.config.padding, root)
		if l.config.axis == Vertical {
			root = expandWidth(root)
		}
	}

	dims := root.Layout(ctx.Child(0))
	l.dispatchReachEnd(ctx, state)
	return dims
}

func (l *listViewWidget) dispatchReachEnd(ctx *internal.Context, state *listViewState) {
	if l.config.onReachEnd == nil || state == nil || l.count <= 0 {
		return
	}
	pos := state.list.Position
	atEnd := pos.First+pos.Count >= l.count && pos.OffsetLast >= 0
	if atEnd && !state.reachCalled {
		state.reachCalled = true
		l.config.onReachEnd(ctx)
	}
	if !atEnd {
		state.reachCalled = false
	}
}

func listViewStateFor(ctx *internal.Context) *listViewState {
	value := ctx.Memo("list-view", func() any {
		return &listViewState{
			list: gioLayout.List{Axis: gioLayout.Vertical},
		}
	})
	state, ok := value.(*listViewState)
	if !ok {
		panic("fluxui/widget: list view state type mismatch")
	}
	return state
}

// GridOption 定义网格配置。
type GridOption func(*gridConfig)

type gridConfig struct {
	rowGap       float32
	colGap       float32
	padding      style.Insets
	minItemWidth float32
}

type gridWidget struct {
	columns  int
	children []Widget
	config   gridConfig
}

type gridViewWidget struct {
	count   int
	columns int
	builder func(ctx *internal.Context, index int) Widget
	config  gridConfig
}

// Grid 创建网格布局。
func Grid(columns int, children ...Widget) Widget {
	if columns <= 0 {
		columns = 1
	}
	return &gridWidget{
		columns:  columns,
		children: append([]Widget(nil), children...),
		config:   gridConfig{},
	}
}

// GridView 创建网格列表。
func GridView(count int, columns int, itemBuilder func(ctx *internal.Context, index int) Widget, opts ...GridOption) Widget {
	if columns <= 0 {
		columns = 1
	}
	cfg := gridConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &gridViewWidget{
		count:   count,
		columns: columns,
		builder: itemBuilder,
		config:  cfg,
	}
}

func GridGap(rowGap, colGap float32) GridOption {
	return func(cfg *gridConfig) {
		cfg.rowGap = rowGap
		cfg.colGap = colGap
	}
}

func GridPadding(insets style.Insets) GridOption {
	return func(cfg *gridConfig) {
		cfg.padding = insets
	}
}

func GridMinItemWidth(width float32) GridOption {
	return func(cfg *gridConfig) {
		cfg.minItemWidth = width
	}
}

func (g *gridWidget) Layout(ctx *internal.Context) layout.Dimensions {
	cols := g.resolveColumns(ctx)
	return buildGrid(cols, g.children, g.config).Layout(ctx.Child(0))
}

func (g *gridViewWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if g.builder == nil || g.count <= 0 {
		return layout.Dimensions{}
	}
	items := make([]Widget, 0, g.count)
	for i := 0; i < g.count; i++ {
		w := g.builder(ctx.Child(i), i)
		if w != nil {
			items = append(items, w)
		}
	}
	cols := g.resolveColumns(ctx)
	return buildGrid(cols, items, g.config).Layout(ctx.Child(0))
}

func (g *gridWidget) resolveColumns(ctx *internal.Context) int {
	return resolveGridColumns(ctx, g.columns, g.config)
}

func (g *gridViewWidget) resolveColumns(ctx *internal.Context) int {
	return resolveGridColumns(ctx, g.columns, g.config)
}

func resolveGridColumns(ctx *internal.Context, base int, cfg gridConfig) int {
	columns := base
	if columns <= 0 {
		columns = 1
	}
	if cfg.minItemWidth <= 0 {
		return columns
	}

	maxW := ctx.MaxConstraints().X
	if maxW <= 0 {
		return columns
	}
	contentW := maxW - insetHorizontalPx(ctx, cfg.padding)
	if contentW <= 0 {
		return 1
	}
	minW := ctx.Gtx.Dp(safeDp(cfg.minItemWidth))
	if minW <= 0 {
		return columns
	}
	colGap := ctx.Gtx.Dp(safeDp(cfg.colGap))

	best := 1
	for c := 1; c <= columns; c++ {
		need := c*minW + (c-1)*colGap
		if need <= contentW {
			best = c
		} else {
			break
		}
	}
	return best
}

func buildGrid(columns int, children []Widget, cfg gridConfig) Widget {
	if columns <= 0 {
		columns = 1
	}
	rows := make([]Widget, 0, (len(children)+columns-1)/columns)
	for i := 0; i < len(children); i += columns {
		end := i + columns
		if end > len(children) {
			end = len(children)
		}
		rowChildren := make([]Widget, 0, columns)
		for j := i; j < end; j++ {
			cell := children[j]
			if cfg.colGap > 0 && j < end-1 {
				rowChildren = append(rowChildren, Padding(style.Insets{Right: cfg.colGap}, cell))
			} else {
				rowChildren = append(rowChildren, cell)
			}
		}
		row := Row(rowChildren...)
		if cfg.rowGap > 0 && end < len(children) {
			row = Padding(style.Insets{Bottom: cfg.rowGap}, row)
		}
		rows = append(rows, row)
	}
	body := Column(rows...)
	if !cfg.padding.IsZero() {
		body = Padding(cfg.padding, body)
	}
	return body
}

type layoutWidgetFunc func(ctx *internal.Context) layout.Dimensions

func (f layoutWidgetFunc) Layout(ctx *internal.Context) layout.Dimensions {
	if f == nil {
		return layout.Dimensions{}
	}
	return f(ctx)
}

func resolveAxis(vertical, horizontal bool) gioLayout.Axis {
	if horizontal && !vertical {
		return gioLayout.Horizontal
	}
	return gioLayout.Vertical
}

func toGioAxis(axis Axis) gioLayout.Axis {
	if axis == Horizontal {
		return gioLayout.Horizontal
	}
	return gioLayout.Vertical
}

func insetHorizontalPx(ctx *internal.Context, insets style.Insets) int {
	return ctx.Gtx.Dp(safeDp(insets.Left)) + ctx.Gtx.Dp(safeDp(insets.Right))
}
