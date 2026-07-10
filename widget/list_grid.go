package widget

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"strconv"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioEvent "gioui.org/io/event"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	gioWidget "gioui.org/widget"
)

const nonVirtualizedLargeItemThreshold = 256

// ScrollOption 定义滚动配置。
type ScrollOption func(*scrollConfig)

type scrollConfig struct {
	vertical   bool
	horizontal bool
	barVisible bool
	onChange   func(ctx *internal.Context, x, y float32)
	autoToEnd  bool
	autoKey    any
	hasAutoKey bool
	ref        *ScrollRef
}

type scrollWidget struct {
	child  Widget
	config scrollConfig
}

type scrollState struct {
	list        gioLayout.List
	bar         gioWidget.Scrollbar
	wheelTag    any
	wheelLeft   float32
	lastFirst   int
	lastOff     int
	autoInited  bool
	lastAutoKey any
}

type wheelPolicy struct {
	axis gioLayout.Axis
}

type wheelPolicyDecision struct {
	delta   float32
	handled bool
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

// ScrollAttachRef 绑定一个命令型引用，用于外部主动控制 ScrollView。
func ScrollAttachRef(ref *ScrollRef) ScrollOption {
	return func(cfg *scrollConfig) {
		cfg.ref = ref
	}
}

// ScrollAutoToEnd 控制是否启用“自动贴底”能力。
// 启用后建议搭配 ScrollAutoToEndKey 使用，以便仅在数据新增时滚到底部。
func ScrollAutoToEnd(enabled bool) ScrollOption {
	return func(cfg *scrollConfig) {
		cfg.autoToEnd = enabled
	}
}

// ScrollAutoToEndKey 设置自动贴底触发键。
// 当 key 变化时，滚动区域会自动跳转到底部。
func ScrollAutoToEndKey(key any) ScrollOption {
	return func(cfg *scrollConfig) {
		cfg.autoToEnd = true
		cfg.autoKey = key
		cfg.hasAutoKey = true
	}
}

func (s *scrollWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if s.child == nil {
		return layout.Dimensions{}
	}

	state := scrollStateFor(ctx)
	forceToEnd := false
	pendingScrollBy := float32(0)
	if s.config.ref != nil {
		bindCommandRef(ctx, "scroll", s.config.ref, &s.config.ref.queue)
		for _, cmd := range s.config.ref.drainCommands() {
			switch cmd.kind {
			case scrollCmdToStart:
				state.list.Position.First = 0
				state.list.Position.Offset = 0
				state.list.Position.BeforeEnd = true
			case scrollCmdToEnd:
				forceToEnd = true
				state.list.Position.BeforeEnd = false
			case scrollCmdToOffset:
				state.list.Position.First = 0
				state.list.Position.Offset = cmd.offset
				state.list.Position.BeforeEnd = true
			case scrollCmdBy:
				pendingScrollBy += cmd.delta
			}
		}
	}
	state.list.ScrollToEnd = s.config.autoToEnd || forceToEnd
	if s.config.autoToEnd {
		if !state.autoInited {
			state.autoInited = true
			state.lastAutoKey = s.config.autoKey
			state.list.Position.BeforeEnd = false
		} else if s.config.hasAutoKey && !reflect.DeepEqual(state.lastAutoKey, s.config.autoKey) {
			state.lastAutoKey = s.config.autoKey
			state.list.Position.BeforeEnd = false
		}
	}
	state.list.Axis = resolveAxis(s.config.vertical, s.config.horizontal)
	s.processWheelEvents(ctx, state)

	dims := s.layoutContent(ctx, state, pendingScrollBy)

	if s.config.barVisible {
		s.drawScrollBar(ctx, state, dims.Size)
	}

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

func (s *scrollWidget) layoutContent(ctx *internal.Context, state *scrollState, pendingScrollBy float32) layout.Dimensions {
	if ctx == nil || state == nil || s.child == nil {
		return layout.Dimensions{}
	}

	axis := state.list.Axis
	childGtx := ctx.Gtx
	const scrollContentLimit = 1_000_000
	if axis == gioLayout.Horizontal {
		childGtx.Constraints.Min.X = 0
		childGtx.Constraints.Max.X = scrollContentLimit
	} else {
		childGtx.Constraints.Min.Y = 0
		childGtx.Constraints.Max.Y = scrollContentLimit
	}

	viewportMax := ctx.Gtx.Constraints.Max
	viewport := image.Rectangle{Min: ctx.Position(), Max: ctx.Position().Add(viewportMax)}
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordViewport(viewport)
	}

	layoutChild := func(offset int) (layout.Dimensions, op.CallOp) {
		macro := op.Record(ctx.Gtx.Ops)
		next := *ctx
		next.Gtx = childGtx
		next = *next.WithViewport(viewport)
		next = *next.WithPointerPassThrough(true)
		scrollOffset := image.Point{}
		if axis == gioLayout.Horizontal {
			scrollOffset.X = -offset
		} else {
			scrollOffset.Y = -offset
		}
		next = *next.WithPositionOffset(scrollOffset)
		childDims := s.child.Layout(next.Child(0))
		return childDims, macro.Stop()
	}

	initialOffset := state.list.Position.Offset
	if initialOffset < 0 {
		initialOffset = 0
	}
	childDims, call := layoutChild(initialOffset)
	contentMajor := axis.Convert(childDims.Size).X
	if contentMajor < 0 {
		contentMajor = 0
	}

	size := childDims.Size
	major := contentMajor
	maxMajor := axis.Convert(ctx.Gtx.Constraints.Max).X
	if major > maxMajor {
		major = maxMajor
	}
	sizeOnAxis := axis.Convert(size)
	sizeOnAxis.X = major
	size = axis.Convert(sizeOnAxis)
	size = clampPointToConstraints(size, ctx.Gtx.Constraints.Min, ctx.Gtx.Constraints.Max)
	viewportMajor := axis.Convert(size).X

	offset := initialOffset
	if pendingScrollBy != 0 && contentMajor > 0 {
		offset += int(math.Round(float64(pendingScrollBy * float32(contentMajor))))
		state.list.Position.BeforeEnd = true
	}
	if state.list.ScrollToEnd && !state.list.Position.BeforeEnd {
		offset = scrollViewMaxOffset(contentMajor, viewportMajor)
	}
	offset = scrollViewClampOffset(offset, contentMajor, viewportMajor)
	if offset != initialOffset {
		childDims, call = layoutChild(offset)
		contentMajor = axis.Convert(childDims.Size).X
		if contentMajor < 0 {
			contentMajor = 0
		}
		offset = scrollViewClampOffset(offset, contentMajor, viewportMajor)
	}

	scrollViewSetPosition(state, offset, contentMajor, viewportMajor)
	s.registerWheelTarget(ctx, state, size)

	area := clip.Rect(image.Rectangle{Max: size}).Push(ctx.Gtx.Ops)
	scrollOffset := image.Point{}
	if axis == gioLayout.Horizontal {
		scrollOffset.X = -offset
	} else {
		scrollOffset.Y = -offset
	}
	trans := op.Offset(scrollOffset).Push(ctx.Gtx.Ops)
	call.Add(ctx.Gtx.Ops)
	trans.Pop()
	area.Pop()

	return layout.Dimensions{Size: size}
}

func (s *scrollWidget) registerWheelTarget(ctx *internal.Context, state *scrollState, size image.Point) {
	if ctx == nil || state == nil || state.wheelTag == nil || size.X <= 0 || size.Y <= 0 {
		return
	}
	pass := pointer.PassOp{}.Push(ctx.Gtx.Ops)
	defer pass.Pop()
	area := clip.Rect(image.Rectangle{Max: size}).Push(ctx.Gtx.Ops)
	gioEvent.Op(ctx.Gtx.Ops, state.wheelTag)
	area.Pop()
}

func (s *scrollWidget) processWheelEvents(ctx *internal.Context, state *scrollState) {
	if ctx == nil || state == nil || state.wheelTag == nil {
		return
	}
	policy := wheelPolicy{axis: state.list.Axis}
	filter := policy.filter(state.wheelTag)
	for {
		raw, ok := ctx.Gtx.Event(filter)
		if !ok {
			break
		}
		pe, ok := raw.(pointer.Event)
		if !ok {
			continue
		}
		wheel := fluxevent.WheelEventFromGio(ctx, pe)
		if !fluxevent.DispatchWheelEvent(ctx, ctx.PathID(), &wheel) {
			continue
		}
		s.applyWheelDefault(ctx, state, policy, &wheel)
	}
}

func (s *scrollWidget) applyWheelDefault(ctx *internal.Context, state *scrollState, policy wheelPolicy, wheel *fluxevent.WheelEvent) {
	if ctx == nil || state == nil || wheel == nil {
		return
	}
	decision := policy.decide(wheel)
	if !decision.handled || decision.delta == 0 {
		return
	}
	state.wheelLeft += decision.delta
	pixels := int(state.wheelLeft)
	state.wheelLeft -= float32(pixels)
	if pixels == 0 {
		return
	}
	state.list.Position.Offset += pixels
	state.list.Position.BeforeEnd = true
	ctx.RequestFrameRedrawReason("input.scrollview.wheel")
}

func (p wheelPolicy) filter(target any) pointer.Filter {
	const scrollLimit = 1 << 30
	filter := pointer.Filter{
		Target: target,
		Kinds:  pointer.Scroll,
	}
	if p.axis == gioLayout.Horizontal {
		filter.ScrollX = pointer.ScrollRange{Min: -scrollLimit, Max: scrollLimit}
		filter.ScrollY = pointer.ScrollRange{Min: 0, Max: 0}
		return filter
	}
	filter.ScrollX = pointer.ScrollRange{Min: 0, Max: 0}
	filter.ScrollY = pointer.ScrollRange{Min: -scrollLimit, Max: scrollLimit}
	return filter
}

func (p wheelPolicy) decide(wheel *fluxevent.WheelEvent) wheelPolicyDecision {
	if wheel == nil {
		return wheelPolicyDecision{}
	}
	if p.axis == gioLayout.Horizontal {
		if wheel.DeltaX != 0 {
			return wheelPolicyDecision{delta: wheel.DeltaX, handled: true}
		}
		return wheelPolicyDecision{}
	}
	if wheel.DeltaY != 0 {
		return wheelPolicyDecision{delta: wheel.DeltaY, handled: true}
	}
	return wheelPolicyDecision{}
}

func scrollViewMaxOffset(contentMajor, viewportMajor int) int {
	maxOffset := contentMajor - viewportMajor
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func scrollViewClampOffset(offset, contentMajor, viewportMajor int) int {
	maxOffset := scrollViewMaxOffset(contentMajor, viewportMajor)
	if offset < 0 {
		return 0
	}
	if offset > maxOffset {
		return maxOffset
	}
	return offset
}

func scrollViewSetPosition(state *scrollState, offset, contentMajor, viewportMajor int) {
	if state == nil {
		return
	}
	offset = scrollViewClampOffset(offset, contentMajor, viewportMajor)
	pos := &state.list.Position
	pos.First = 0
	pos.Offset = offset
	pos.Count = 0
	if contentMajor > 0 {
		pos.Count = 1
	}
	pos.Length = contentMajor
	pos.OffsetLast = viewportMajor - (contentMajor - offset)
	pos.BeforeEnd = offset < scrollViewMaxOffset(contentMajor, viewportMajor)
}

func scrollViewScrollByFraction(state *scrollState, delta float32, viewportMajor int) bool {
	if state == nil || delta == 0 || state.list.Position.Length <= 0 {
		return false
	}
	pos := state.list.Position
	next := pos.Offset + int(math.Round(float64(delta*float32(pos.Length))))
	next = scrollViewClampOffset(next, pos.Length, viewportMajor)
	if next == pos.Offset {
		return false
	}
	scrollViewSetPosition(state, next, pos.Length, viewportMajor)
	return true
}

func (s *scrollWidget) drawScrollBar(ctx *internal.Context, state *scrollState, size image.Point) {
	if ctx == nil || state == nil || size.X <= 0 || size.Y <= 0 {
		return
	}

	pos := state.list.Position
	viewport := size.Y
	if state.list.Axis == gioLayout.Horizontal {
		viewport = size.X
	}
	if viewport <= 0 || pos.Length <= viewport {
		return
	}

	thickness := ctx.Gtx.Dp(unit.Dp(6))
	margin := ctx.Gtx.Dp(unit.Dp(2))
	minThumb := ctx.Gtx.Dp(unit.Dp(24))
	if thickness <= 0 {
		thickness = 6
	}
	if margin < 0 {
		margin = 0
	}
	if minThumb <= 0 {
		minThumb = 12
	}

	trackColor := setAlpha(ctx.Theme().SurfaceMuted, 120)
	thumbColor := setAlpha(ctx.Theme().TextColor, 180)
	if state.bar.TrackHovered() {
		trackColor = setAlpha(ctx.Theme().SurfaceMuted, 150)
	}
	if state.bar.IndicatorHovered() || state.bar.Dragging() {
		thumbColor = setAlpha(ctx.Theme().TextColor, 220)
	}

	viewportStart, viewportEnd := viewportFromListPosition(pos, 1, viewport)
	if viewportEnd <= viewportStart {
		return
	}

	if state.list.Axis == gioLayout.Horizontal {
		track := image.Rectangle{
			Min: image.Point{X: margin, Y: size.Y - margin - thickness},
			Max: image.Point{X: size.X - margin, Y: size.Y - margin},
		}
		s.handleScrollBarInput(ctx, state, track, viewportStart, viewportEnd)
		if delta := state.bar.ScrollDistance(); delta != 0 {
			if scrollViewScrollByFraction(state, delta, viewport) {
				viewportStart, viewportEnd = viewportFromListPosition(state.list.Position, 1, viewport)
				ctx.RequestFrameRedrawReason("input.scrollbar")
			}
		}
		if state.bar.Dragging() {
			ctx.RequestFrameRedrawReason("input.scrollbar")
		}
		drawScrollbarOnAxis(ctx, state, track, viewportStart, viewportEnd, minThumb, trackColor, thumbColor, true)
		return
	}

	track := image.Rectangle{
		Min: image.Point{X: size.X - margin - thickness, Y: margin},
		Max: image.Point{X: size.X - margin, Y: size.Y - margin},
	}
	s.handleScrollBarInput(ctx, state, track, viewportStart, viewportEnd)
	if delta := state.bar.ScrollDistance(); delta != 0 {
		if scrollViewScrollByFraction(state, delta, viewport) {
			viewportStart, viewportEnd = viewportFromListPosition(state.list.Position, 1, viewport)
			ctx.RequestFrameRedrawReason("input.scrollbar")
		}
	}
	if state.bar.Dragging() {
		ctx.RequestFrameRedrawReason("input.scrollbar")
	}
	drawScrollbarOnAxis(ctx, state, track, viewportStart, viewportEnd, minThumb, trackColor, thumbColor, false)
}

func (s *scrollWidget) handleScrollBarInput(
	ctx *internal.Context,
	state *scrollState,
	track image.Rectangle,
	viewportStart float32,
	viewportEnd float32,
) {
	if ctx == nil || state == nil || track.Dx() <= 0 || track.Dy() <= 0 {
		return
	}

	local := ctx.Gtx
	local.Constraints = gioLayout.Exact(image.Point{X: track.Dx(), Y: track.Dy()})
	state.bar.Update(local, state.list.Axis, viewportStart, viewportEnd)
}

func drawScrollbarOnAxis(
	ctx *internal.Context,
	state *scrollState,
	track image.Rectangle,
	viewportStart float32,
	viewportEnd float32,
	minThumb int,
	trackColor color.NRGBA,
	thumbColor color.NRGBA,
	horizontal bool,
) {
	if ctx == nil || state == nil || track.Dx() <= 0 || track.Dy() <= 0 {
		return
	}
	if viewportEnd <= viewportStart {
		return
	}

	trackLen := track.Dy()
	if horizontal {
		trackLen = track.Dx()
	}
	if trackLen <= 0 {
		return
	}

	thumbLen := int(math.Round(float64((viewportEnd - viewportStart) * float32(trackLen))))
	if thumbLen < minThumb {
		thumbLen = minThumb
	}
	if thumbLen > trackLen {
		thumbLen = trackLen
	}

	thumbOffset := int(math.Round(float64(viewportStart * float32(trackLen))))
	travel := trackLen - thumbLen
	if thumbOffset > travel {
		thumbOffset = travel
	}
	if thumbOffset < 0 {
		thumbOffset = 0
	}

	radius := track.Dx()
	if track.Dy() < radius {
		radius = track.Dy()
	}
	radius /= 2
	if radius < 1 {
		radius = 1
	}

	paint.FillShape(ctx.Gtx.Ops, trackColor, clip.UniformRRect(track, radius).Op(ctx.Gtx.Ops))

	thumb := track
	if horizontal {
		thumb.Min.X = track.Min.X + thumbOffset
		thumb.Max.X = thumb.Min.X + thumbLen
	} else {
		thumb.Min.Y = track.Min.Y + thumbOffset
		thumb.Max.Y = thumb.Min.Y + thumbLen
	}
	paint.FillShape(ctx.Gtx.Ops, thumbColor, clip.UniformRRect(thumb, radius).Op(ctx.Gtx.Ops))

	trackOffset := op.Offset(track.Min).Push(ctx.Gtx.Ops)
	trackArea := clip.Rect(image.Rectangle{Max: image.Point{X: track.Dx(), Y: track.Dy()}}).Push(ctx.Gtx.Ops)
	state.bar.AddDrag(ctx.Gtx.Ops)
	trackArea.Pop()

	passTrack := pointer.PassOp{}.Push(ctx.Gtx.Ops)
	trackArea = clip.Rect(image.Rectangle{Max: image.Point{X: track.Dx(), Y: track.Dy()}}).Push(ctx.Gtx.Ops)
	state.bar.AddTrack(ctx.Gtx.Ops)
	trackArea.Pop()
	passTrack.Pop()
	trackOffset.Pop()

	thumbOffsetOp := op.Offset(thumb.Min).Push(ctx.Gtx.Ops)
	thumbArea := clip.Rect(image.Rectangle{Max: image.Point{X: thumb.Dx(), Y: thumb.Dy()}}).Push(ctx.Gtx.Ops)
	passIndicator := pointer.PassOp{}.Push(ctx.Gtx.Ops)
	state.bar.AddIndicator(ctx.Gtx.Ops)
	passIndicator.Pop()
	thumbArea.Pop()
	thumbOffsetOp.Pop()
}

func viewportFromListPosition(lp gioLayout.Position, elements int, majorAxisSize int) (start, end float32) {
	if elements <= 0 || majorAxisSize <= 0 || lp.Length <= 0 {
		return 0, 0
	}

	lengthEstPx := float32(lp.Length)
	elementLenEstPx := lengthEstPx / float32(elements)

	listOffsetF := float32(lp.Offset)
	listOffsetL := float32(lp.OffsetLast)

	viewportStart := clampFloat32((float32(lp.First)*elementLenEstPx+listOffsetF)/lengthEstPx, 0, 1)
	viewportEnd := clampFloat32((float32(lp.First+lp.Count)*elementLenEstPx+listOffsetL)/lengthEstPx, 0, 1)
	viewportFraction := viewportEnd - viewportStart

	visiblePx := float32(majorAxisSize)
	visibleFraction := visiblePx / lengthEstPx

	err := visibleFraction - viewportFraction
	adjStart := viewportStart
	adjEnd := viewportEnd
	if viewportFraction < 1 {
		startShare := viewportStart / (1 - viewportFraction)
		endShare := (1 - viewportEnd) / (1 - viewportFraction)
		startErr := startShare * err
		endErr := endShare * err

		adjStart -= startErr
		adjEnd += endErr
	}

	start = clampFloat32(adjStart, 0, 1)
	end = clampFloat32(adjEnd, 0, 1)
	if end < start {
		end = start
	}
	return start, end
}

func setAlpha(col color.NRGBA, alpha uint8) color.NRGBA {
	col.A = alpha
	return col
}

func scrollStateFor(ctx *internal.Context) *scrollState {
	value := ctx.Memo("scroll", func() any {
		return &scrollState{
			list:     gioLayout.List{Axis: gioLayout.Vertical},
			wheelTag: new(int),
		}
	})
	state, ok := value.(*scrollState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: scroll state type mismatch")
	}
	if state.wheelTag == nil {
		state.wheelTag = new(int)
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
	decoration  style.Decoration
}

type listViewWidget struct {
	count   int
	builder func(ctx *internal.Context, index int) Widget
	config  listConfig
}

type listViewState struct {
	list        gioLayout.List
	viewportMaj int
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

// ListDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func ListDecoration(d style.Decoration) ListOption {
	return func(cfg *listConfig) {
		cfg.decoration = d
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
	if !l.config.virtualized {
		return l.layoutNonVirtualized(ctx)
	}

	state := listViewStateFor(ctx)
	state.list.Axis = toGioAxis(l.config.axis)

	listChild := layoutWidgetFunc(func(listCtx *internal.Context) layout.Dimensions {
		visibleItems := 0
		viewport := viewportForContext(listCtx)
		dims := state.list.Layout(listCtx.Gtx, l.count, func(gtx gioLayout.Context, index int) gioLayout.Dimensions {
			visibleItems++
			next := *listCtx
			next.Gtx = gtx
			next = *next.WithViewport(viewport)
			childCtx := next.Child(index)
			child := l.builder(childCtx, index)
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
			childDims := child.Layout(childCtx)
			return gioLayout.Dimensions{Size: childDims.Size}
		})
		state.viewportMaj = toGioAxis(l.config.axis).Convert(dims.Size).X
		if rt := listCtx.Runtime(); rt != nil {
			rt.RecordVirtualizedItems(l.count, visibleItems, viewport)
		}
		return layout.Dimensions{Size: dims.Size}
	})

	dims := l.wrapListBody(listChild).Layout(ctx.Child(0))
	l.dispatchReachEnd(ctx, state)
	return dims
}

func (l *listViewWidget) layoutNonVirtualized(ctx *internal.Context) layout.Dimensions {
	if rt := ctx.Runtime(); rt != nil {
		viewport := viewportForContext(ctx)
		rt.RecordVirtualizedItems(l.count, l.count, viewport)
		if l.count >= nonVirtualizedLargeItemThreshold {
			rt.RecordNonVirtualizedItems(l.count)
		}
	}

	children := make([]Widget, 0, l.count)
	for index := 0; index < l.count; index++ {
		idx := index
		var child Widget = layoutWidgetFunc(func(itemCtx *internal.Context) layout.Dimensions {
			item := l.builder(itemCtx, idx)
			if item == nil {
				item = emptyWidget
			}
			return item.Layout(itemCtx)
		})
		if l.config.itemSpacing > 0 && index < l.count-1 {
			if l.config.axis == Horizontal {
				child = Padding(style.Insets{Right: l.config.itemSpacing}, child)
			} else {
				child = Padding(style.Insets{Bottom: l.config.itemSpacing}, child)
			}
		}
		children = append(children, child)
	}

	var body Widget
	if l.config.axis == Horizontal {
		body = Row(children...)
	} else {
		body = Column(children...)
	}
	return l.wrapListBody(body).Layout(ctx.Child(0))
}

func (l *listViewWidget) wrapListBody(body Widget) Widget {
	var root Widget = body
	if l.config.axis == Vertical {
		root = expandWidth(root)
	}
	padding := l.config.decoration.ResolvePad(l.config.padding)
	if !padding.IsZero() {
		root = Padding(padding, root)
		if l.config.axis == Vertical {
			root = expandWidth(root)
		}
	}
	return root
}

func (l *listViewWidget) dispatchReachEnd(ctx *internal.Context, state *listViewState) {
	if l.config.onReachEnd == nil || state == nil || l.count <= 0 {
		return
	}
	pos := state.list.Position
	if pos.Count <= 0 {
		state.reachCalled = false
		return
	}

	// Gio 的 Position.BeforeEnd 会在真正触达末尾时置为 false。
	atEnd := !pos.BeforeEnd && pos.First+pos.Count >= l.count

	// 兜底：部分场景 BeforeEnd 变化会滞后，用视口比例再做一次判定。
	if !atEnd && state.viewportMaj > 0 && pos.Length > 0 {
		_, viewportEnd := viewportFromListPosition(pos, l.count, state.viewportMaj)
		atEnd = viewportEnd >= 0.999
	}

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
		panic("github.com/xiaowumin-mark/FluxUIwidget: list view state type mismatch")
	}
	return state
}

type gridViewState struct {
	list        gioLayout.List
	viewportMaj int
	reachCalled bool
}

func gridViewStateFor(ctx *internal.Context) *gridViewState {
	value := ctx.Memo("grid-view", func() any {
		return &gridViewState{
			list: gioLayout.List{Axis: gioLayout.Vertical},
		}
	})
	return value.(*gridViewState)
}

// GridOption 定义网格配置。
type GridOption func(*gridConfig)

type gridConfig struct {
	rowGap       float32
	colGap       float32
	padding      style.Insets
	minItemWidth float32
	onReachEnd   func(ctx *internal.Context)
	decoration   style.Decoration
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

var emptyWidget = layoutWidgetFunc(func(_ *internal.Context) layout.Dimensions {
	return layout.Dimensions{}
})

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

// GridDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func GridDecoration(d style.Decoration) GridOption {
	return func(cfg *gridConfig) {
		cfg.decoration = d
	}
}

func GridMinItemWidth(width float32) GridOption {
	return func(cfg *gridConfig) {
		cfg.minItemWidth = width
	}
}

func GridOnReachEnd(fn func(ctx *internal.Context)) GridOption {
	return func(cfg *gridConfig) {
		cfg.onReachEnd = fn
	}
}

func (g *gridWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if rt := ctx.Runtime(); rt != nil && len(g.children) >= nonVirtualizedLargeItemThreshold {
		rt.RecordNonVirtualizedItems(len(g.children))
	}
	cols := g.resolveColumns(ctx)
	return buildGrid(cols, g.children, g.config).Layout(ctx.Child(0))
}

func (g *gridViewWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if g.builder == nil || g.count <= 0 {
		return layout.Dimensions{}
	}

	cols := g.resolveColumns(ctx)
	rowCount := (g.count + cols - 1) / cols

	state := gridViewStateFor(ctx)
	state.list.Axis = gioLayout.Vertical

	listChild := layoutWidgetFunc(func(listCtx *internal.Context) layout.Dimensions {
		visibleItems := 0
		viewport := viewportForContext(listCtx)
		dims := state.list.Layout(listCtx.Gtx, rowCount, func(gtx gioLayout.Context, rowIndex int) gioLayout.Dimensions {
			startIdx := rowIndex * cols
			endIdx := startIdx + cols
			if endIdx > g.count {
				endIdx = g.count
			}
			visibleItems += endIdx - startIdx

			next := *listCtx
			next.Gtx = gtx
			next = *next.WithViewport(viewport)
			rowCtx := next.Scope(strconv.Itoa(rowIndex))

			rowChildren := make([]Widget, 0, cols)
			for i := startIdx; i < endIdx; i++ {
				cell := g.builder(rowCtx.Child(i-startIdx), i)
				if cell == nil {
					cell = emptyWidget
				}
				if g.config.colGap > 0 && i < endIdx-1 {
					cell = Padding(style.Insets{Right: g.config.colGap}, cell)
				}
				rowChildren = append(rowChildren, cell)
			}

			for i := endIdx - startIdx; i < cols; i++ {
				empty := emptyWidget
				if g.config.colGap > 0 && i < cols-1 {
					rowChildren = append(rowChildren, Padding(style.Insets{Right: g.config.colGap}, empty))
				} else {
					rowChildren = append(rowChildren, empty)
				}
			}

			row := Row(rowChildren...)
			if g.config.rowGap > 0 && rowIndex < rowCount-1 {
				row = Padding(style.Insets{Bottom: g.config.rowGap}, row)
			}

			childDims := row.Layout(rowCtx.Child(cols))
			return gioLayout.Dimensions{Size: childDims.Size}
		})
		state.viewportMaj = dims.Size.Y
		if rt := listCtx.Runtime(); rt != nil {
			rt.RecordVirtualizedItems(g.count, visibleItems, viewport)
		}
		return layout.Dimensions{Size: dims.Size}
	})

	var root Widget = expandWidth(listChild)
	padding := g.config.decoration.ResolvePad(g.config.padding)
	if !padding.IsZero() {
		root = Padding(padding, root)
		root = expandWidth(root)
	}

	dims := root.Layout(ctx.Child(0))
	g.dispatchReachEnd(ctx, state, rowCount)
	return dims
}

func (g *gridViewWidget) dispatchReachEnd(ctx *internal.Context, state *gridViewState, rowCount int) {
	if g.config.onReachEnd == nil || state == nil || rowCount <= 0 {
		return
	}
	pos := state.list.Position
	if pos.Count <= 0 {
		state.reachCalled = false
		return
	}
	atEnd := !pos.BeforeEnd && pos.First+pos.Count >= rowCount
	if !atEnd && state.viewportMaj > 0 && pos.Length > 0 {
		_, viewportEnd := viewportFromListPosition(pos, rowCount, state.viewportMaj)
		atEnd = viewportEnd >= 0.999
	}
	if atEnd && !state.reachCalled {
		state.reachCalled = true
		g.config.onReachEnd(ctx)
	}
	if !atEnd {
		state.reachCalled = false
	}
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
		rowChildren := make([]Widget, 0, end-i)
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

func viewportForContext(ctx *internal.Context) image.Rectangle {
	if ctx == nil {
		return image.Rectangle{}
	}
	if viewport, ok := ctx.Viewport(); ok {
		return viewport
	}
	return image.Rectangle{
		Min: ctx.Position(),
		Max: ctx.Position().Add(ctx.Gtx.Constraints.Max),
	}
}
