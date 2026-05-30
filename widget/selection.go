package widget

import (
	"fmt"
	"image"
	"image/color"

	event "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/f32"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// RadioItem 单选项。
type RadioItem struct {
	Label string
	Value string
}

// RadioGroupOption 定义单选组配置。
type RadioGroupOption func(*radioGroupConfig)

type radioGroupConfig struct {
	direction  Axis
	disabled   bool
	onChange   func(ctx *internal.Context, value string)
	size       float32
	color      color.NRGBA
	hasColor   bool
	ref        *RadioGroupRef
	decoration style.Decoration
}

type radioGroupWidget struct {
	value  string
	items  []RadioItem
	config radioGroupConfig
}

// RadioGroup 创建单选组。
func RadioGroup(value string, items []RadioItem, opts ...RadioGroupOption) Widget {
	cfg := radioGroupConfig{
		direction: Vertical,
		size:      20,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &radioGroupWidget{
		value:  value,
		items:  append([]RadioItem(nil), items...),
		config: cfg,
	}
}

// RadioGroupDirection 设置排列方向。
func RadioGroupDirection(axis Axis) RadioGroupOption {
	return func(cfg *radioGroupConfig) {
		cfg.direction = axis
	}
}

// RadioGroupDisabled 设置禁用。
func RadioGroupDisabled(disabled bool) RadioGroupOption {
	return func(cfg *radioGroupConfig) {
		cfg.disabled = disabled
	}
}

// RadioGroupOnChange 设置值变更回调。
func RadioGroupOnChange(fn func(ctx *internal.Context, value string)) RadioGroupOption {
	return func(cfg *radioGroupConfig) {
		cfg.onChange = fn
	}
}

// RadioGroupSize 设置圆点尺寸。
func RadioGroupSize(size float32) RadioGroupOption {
	return func(cfg *radioGroupConfig) {
		cfg.size = size
	}
}

// RadioGroupColor 设置激活色。
func RadioGroupColor(col color.NRGBA) RadioGroupOption {
	return func(cfg *radioGroupConfig) {
		cfg.color = col
		cfg.hasColor = true
	}
}

// RadioGroupAttachRef 绑定命令型引用，用于外部主动设置值。
func RadioGroupAttachRef(ref *RadioGroupRef) RadioGroupOption {
	return func(cfg *radioGroupConfig) {
		cfg.ref = ref
	}
}

// RadioGroupDecoration 通过 Decoration 统一设置选项外层装饰和交互态。
func RadioGroupDecoration(d style.Decoration) RadioGroupOption {
	return func(cfg *radioGroupConfig) {
		cfg.decoration = d
	}
}

func (r *radioGroupWidget) Layout(ctx *internal.Context) layout.Dimensions {
	currentValue := r.value
	if r.config.ref != nil {
		r.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
		for _, next := range r.config.ref.drainCommands() {
			if next == currentValue {
				continue
			}
			currentValue = next
			if r.config.onChange != nil {
				r.config.onChange(ctx, next)
			}
		}
	}

	cs := ctx.Theme().Colors
	mainColor := cs.Primary
	if r.config.hasColor {
		mainColor = r.config.color
	}
	labelColor := cs.OnSurface
	if r.config.disabled {
		labelColor = style.DisabledContent(cs.OnSurface)
	}

	children := make([]Widget, 0, len(r.items))
	for idx := range r.items {
		item := r.items[idx]
		checked := item.Value == currentValue

		var row Widget = layoutWidgetFunc(func(rowCtx *internal.Context) layout.Dimensions {
			clickable := event.UseClickable(rowCtx)
			if !r.config.disabled {
				for clickable.Clicked(rowCtx) {
					if item.Value == currentValue {
						continue
					}
					currentValue = item.Value
					if r.config.onChange != nil {
						r.config.onChange(rowCtx, item.Value)
					}
				}
			}

			itemLabelColor := labelColor
			if !r.config.disabled && clickable.Hovered() {
				itemLabelColor = style.StateLayer(itemLabelColor, mainColor, style.StateLayerHoverOpacity)
			}

			content := func(contentCtx *internal.Context) image.Point {
				dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx,
					gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						size := next.LayoutRadio(nil, checked, internal.RadioSpec{
							Size:     r.config.size,
							Color:    mainColor,
							Disabled: r.config.disabled,
							Hovered:  clickable.Hovered(),
							Pressed:  clickable.Pressed(),
						})
						return gioLayout.Dimensions{Size: size}
					}),
					gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						next.Gtx.Constraints.Min = image.Point{}
						size := next.LayoutInset(internal.Insets{Left: 8}, func(labelCtx *internal.Context) image.Point {
							return labelCtx.LayoutText(internal.TextSpec{
								Content:   item.Label,
								Size:      labelCtx.Theme().Types.BodyMedium.Size,
								Color:     itemLabelColor,
								Alignment: internal.AlignStart,
							})
						})
						return gioLayout.Dimensions{Size: size}
					}),
				)
				return dims.Size
			}

			deco := withDefaultStates(r.config.decoration,
				style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, mainColor, style.StateLayerHoverOpacity)).WithRad(8),
				style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, mainColor, style.StateLayerPressedOpacity)).WithRad(8),
				style.Decoration{}.WithBg(style.DisabledContainer(cs.OnSurface)).WithRad(8),
			)
			target := func(targetCtx *internal.Context) image.Point {
				if hasAnyDecoration(r.config.decoration) {
					active := resolveDecorationState(deco, clickable.Hovered(), clickable.Pressed(), r.config.disabled)
					visual := stripStateDecoration(active)
					return layoutDecorationShell(targetCtx.Child(0), visual, content).Size
				}
				return content(targetCtx.Child(0))
			}
			return layoutStateLayerTouchTarget(rowCtx, clickable.Handle(), r.config.disabled, internal.RippleSpec{
				Color:   mainColor,
				Radius:  20,
				Opacity: style.StateLayerPressedOpacity,
			}, 48, 40, func(size image.Point) image.Point {
				controlSize := selectionControlSizePx(rowCtx, r.config.size, 20)
				return image.Pt(controlSize/2, size.Y/2)
			}, func() float32 {
				return materialStateLayerOpacity(clickable.Hovered(), clickable.Pressed())
			}, target)
		})

		if r.config.direction == Horizontal {
			children = append(children, Padding(style.Insets{Right: 12}, row))
		} else {
			children = append(children, Padding(style.Insets{Bottom: 6}, row))
		}
	}

	if r.config.direction == Horizontal {
		return Row(children...).Layout(ctx.Child(0))
	}
	return Column(children...).Layout(ctx.Child(0))
}

// SelectOptionItem 下拉选项。
type SelectOptionItem[T comparable] struct {
	Label string
	Value T
}

// SelectOption 定义下拉配置。
type SelectOption[T comparable] func(*selectConfig[T])

type selectConfig[T comparable] struct {
	placeholder string
	disabled    bool
	searchable  bool
	maxHeight   float32
	onChange    func(ctx *internal.Context, value T)
	onOpen      func(ctx *internal.Context, opened bool)
	ref         *SelectRef[T]
	decoration  style.Decoration
}

type selectWidget[T comparable] struct {
	value   T
	options []SelectOptionItem[T]
	config  selectConfig[T]
}

type selectState struct {
	opened bool
}

// Select 创建下拉选择组件。
func Select[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget {
	cfg := selectConfig[T]{
		placeholder: "请选择",
		maxHeight:   240,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &selectWidget[T]{
		value:   value,
		options: append([]SelectOptionItem[T](nil), options...),
		config:  cfg,
	}
}

// SelectPlaceholder 设置占位文案。
func SelectPlaceholder[T comparable](text string) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.placeholder = text
	}
}

// SelectDisabled 设置禁用。
func SelectDisabled[T comparable](disabled bool) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.disabled = disabled
	}
}

// SelectSearchable 设置可搜索（预留参数）。
func SelectSearchable[T comparable](searchable bool) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.searchable = searchable
	}
}

// SelectMaxHeight 设置下拉面板最大高度。
func SelectMaxHeight[T comparable](height float32) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.maxHeight = height
	}
}

// SelectOnChange 设置值变更回调。
func SelectOnChange[T comparable](fn func(ctx *internal.Context, value T)) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.onChange = fn
	}
}

// SelectOnOpenChange 设置展开状态回调。
func SelectOnOpenChange[T comparable](fn func(ctx *internal.Context, opened bool)) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.onOpen = fn
	}
}

// SelectAttachRef 绑定命令型引用，用于外部主动控制值与展开状态。
func SelectAttachRef[T comparable](ref *SelectRef[T]) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.ref = ref
	}
}

// SelectDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func SelectDecoration[T comparable](d style.Decoration) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.decoration = d
	}
}

func (s *selectWidget[T]) Layout(ctx *internal.Context) layout.Dimensions {
	state := selectStateFor(ctx)
	currentValue := s.value
	if s.config.ref != nil {
		s.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
		for _, cmd := range s.config.ref.drainCommands() {
			switch cmd.kind {
			case selectCmdSetValue:
				if cmd.value != currentValue && s.config.onChange != nil {
					s.config.onChange(ctx, cmd.value)
				}
				currentValue = cmd.value
			case selectCmdOpen:
				if !state.opened {
					state.opened = true
					if s.config.onOpen != nil {
						s.config.onOpen(ctx, true)
					}
				}
			case selectCmdClose:
				if state.opened {
					state.opened = false
					if s.config.onOpen != nil {
						s.config.onOpen(ctx, false)
					}
				}
			case selectCmdToggle:
				state.opened = !state.opened
				if s.config.onOpen != nil {
					s.config.onOpen(ctx, state.opened)
				}
			}
		}
	}

	label, currentIndex := s.resolveCurrentLabel(currentValue)

	toggle := layoutWidgetFunc(func(triggerCtx *internal.Context) layout.Dimensions {
		clickable := event.UseClickable(triggerCtx)
		if !s.config.disabled {
			for clickable.Clicked(triggerCtx) {
				state.opened = !state.opened
				if s.config.onOpen != nil {
					s.config.onOpen(triggerCtx, state.opened)
				}
			}
		}

		cs := triggerCtx.Theme().Colors
		textColor := cs.OnSurface
		if currentIndex < 0 {
			textColor = cs.OnSurfaceVariant
		}
		arrowColor := cs.OnSurfaceVariant
		border := style.Border{Width: 1, Color: cs.Outline}
		if state.opened || clickable.Focused(triggerCtx) {
			border = style.Border{Width: 2, Color: cs.Primary}
		}
		if s.config.disabled {
			textColor = style.DisabledContent(cs.OnSurface)
			arrowColor = style.DisabledContent(cs.OnSurface)
			border = style.Border{Width: 1, Color: style.DisabledContent(cs.OnSurface)}
		}

		content := layoutWidgetFunc(func(contentCtx *internal.Context) layout.Dimensions {
			dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx,
				gioLayout.Flexed(1, func(gtx gioLayout.Context) gioLayout.Dimensions {
					next := *contentCtx
					next.Gtx = gtx
					dims := Text(label, TextSize(triggerCtx.Theme().Types.BodyLarge.Size), TextColor(textColor)).Layout(next.Child(0))
					return gioLayout.Dimensions{Size: dims.Size}
				}),
				gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
					next := *contentCtx
					next.Gtx = gtx
					next.Gtx.Constraints.Min = image.Point{}
					dims := Padding(style.Insets{Left: 8}, selectArrow(state.opened, arrowColor)).Layout(next.Child(1))
					return gioLayout.Dimensions{Size: dims.Size}
				}),
			)
			return layout.Dimensions{Size: dims.Size}
		})
		bg := s.config.decoration.ResolveBg(cs.Surface)
		return md3ActionSurface(triggerCtx, clickable, md3ActionSurfaceSpec{
			Background: bg,
			Foreground: textColor,
			Radius:     s.config.decoration.ResolveRad(triggerCtx.Theme().Shapes.ExtraSmall),
			Padding:    s.config.decoration.ResolvePad(style.Insets{Top: 8, Right: 12, Bottom: 8, Left: 16}),
			Border:     border,
			MinHeight:  56,
			FillWidth:  true,
			Disabled:   s.config.disabled,
		}, content)
	})

	toggleDims := toggle.Layout(ctx.Child(0))
	if !state.opened || len(s.options) == 0 {
		return toggleDims
	}

	items := make([]Widget, 0, len(s.options))
	for idx := range s.options {
		item := s.options[idx]
		itemLabel := item.Label
		if itemLabel == "" {
			itemLabel = fmt.Sprintf("%v", item.Value)
		}
		isActive := idx == currentIndex
		var row Widget = layoutWidgetFunc(func(rowCtx *internal.Context) layout.Dimensions {
			clickable := event.UseClickable(rowCtx)
			for clickable.Clicked(rowCtx) {
				currentValue = item.Value
				if s.config.onChange != nil {
					s.config.onChange(rowCtx, item.Value)
				}
				if state.opened {
					state.opened = false
					if s.config.onOpen != nil {
						s.config.onOpen(rowCtx, false)
					}
				}
			}

			cs := rowCtx.Theme().Colors
			bg := color.NRGBA{}
			if isActive {
				bg = style.StateLayer(color.NRGBA{}, cs.Primary, style.StateLayerFocusOpacity)
			}
			if clickable.Pressed() {
				bg = style.StateLayer(bg, cs.Primary, style.StateLayerPressedOpacity)
			} else if clickable.Hovered() {
				bg = style.StateLayer(bg, cs.Primary, style.StateLayerHoverOpacity)
			}
			itemRow := layoutWidgetFunc(func(contentCtx *internal.Context) layout.Dimensions {
				dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx,
					gioLayout.Flexed(1, func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						dims := Text(itemLabel, TextColor(cs.OnSurface), TextSize(rowCtx.Theme().Types.BodyMedium.Size)).Layout(next.Child(0))
						return gioLayout.Dimensions{Size: dims.Size}
					}),
					gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						next.Gtx.Constraints.Min = image.Point{}
						dims := Padding(style.Insets{Left: 12}, selectCheckMark(isActive, cs.Primary)).Layout(next.Child(1))
						return gioLayout.Dimensions{Size: dims.Size}
					}),
				)
				return layout.Dimensions{Size: dims.Size}
			})
			itemContent := ContainerDecoration(
				style.Decoration{}.WithBg(bg).WithPad(style.Symmetric(8, 12)).WithRad(rowCtx.Theme().Shapes.ExtraSmall),
				itemRow,
			)
			size := rowCtx.LayoutRippleArea(clickable.Handle(), internal.RippleSpec{
				Color:   cs.Primary,
				Radius:  rowCtx.Theme().Shapes.ExtraSmall,
				Opacity: style.StateLayerPressedOpacity,
			}, func(childCtx *internal.Context) image.Point {
				return itemContent.Layout(childCtx.Child(0)).Size
			})
			if clickable.Focused(rowCtx) {
				rowCtx.DrawFocusIndicator(size, internal.FocusIndicatorSpec{
					Color:  cs.Primary,
					Radius: rowCtx.Theme().Shapes.ExtraSmall,
				})
			}
			return layout.Dimensions{Size: size}
		})
		row = expandWidth(row)
		items = append(items, row)
	}

	list := ListView(
		len(items),
		func(ctx *internal.Context, index int) Widget {
			return items[index]
		},
		ListItemSpacing(4),
		ListVirtualized(true),
	)
	panel := expandWidth(
		ContainerDecoration(
			style.Decoration{}.
				WithBg(s.config.decoration.ResolveBg(ctx.Theme().Colors.SurfaceContainer)).
				WithPad(s.config.decoration.ResolvePad(style.All(6))).
				WithRad(s.config.decoration.ResolveRad(ctx.Theme().Shapes.ExtraSmall)).
				WithShadow(style.ElevationShadow(ctx.Theme().Colors, 2)),
			list,
		),
	)

	maxH := s.config.maxHeight
	if maxH <= 0 {
		maxH = 240
	}
	maxHPx := ctx.Gtx.Dp(safeDp(maxH))
	if maxHPx <= 0 {
		maxHPx = 1
	}

	popupYOffset := toggleDims.Size.Y + ctx.Gtx.Dp(safeDp(6))
	availableY := ctx.Gtx.Constraints.Max.Y - popupYOffset
	if availableY <= 0 {
		availableY = maxHPx
	}
	if availableY > maxHPx {
		availableY = maxHPx
	}
	popupW := toggleDims.Size.X
	if popupW <= 0 {
		popupW = ctx.Gtx.Constraints.Max.X
	}
	if popupW <= 0 {
		popupW = 1
	}

	popupMacro := op.Record(ctx.Gtx.Ops)
	offset := op.Offset(image.Point{Y: popupYOffset}).Push(ctx.Gtx.Ops)
	popupCtx := *ctx
	popupCtx.Gtx = ctx.Gtx
	popupCtx.Gtx.Constraints.Min = image.Point{}
	popupCtx.Gtx.Constraints.Max = image.Point{X: popupW, Y: availableY}
	_ = panel.Layout(popupCtx.Child(1))
	offset.Pop()
	popupCall := popupMacro.Stop()
	op.Defer(ctx.Gtx.Ops, popupCall)

	return toggleDims
}

func (s *selectWidget[T]) resolveCurrentLabel(value T) (string, int) {
	label := s.config.placeholder
	currentIndex := -1
	for idx := range s.options {
		if s.options[idx].Value == value {
			label = s.options[idx].Label
			currentIndex = idx
			break
		}
	}
	if label == "" {
		label = fmt.Sprintf("%v", value)
	}
	return label, currentIndex
}

func selectStateFor(ctx *internal.Context) *selectState {
	value := ctx.Memo("select", func() any {
		return &selectState{}
	})
	state, ok := value.(*selectState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: select state type mismatch")
	}
	return state
}

func selectArrow(open bool, col color.NRGBA) Widget {
	return &selectArrowWidget{open: open, color: col}
}

type selectArrowWidget struct {
	open  bool
	color color.NRGBA
}

func (s *selectArrowWidget) Layout(ctx *internal.Context) layout.Dimensions {
	box := ctx.Gtx.Dp(safeDp(24))
	if box < 18 {
		box = 18
	}
	size := ctx.Gtx.Dp(safeDp(10))
	if size < 8 {
		size = 8
	}
	offset := image.Pt((box-size)/2, (box-size)/2)
	stack := op.Offset(offset).Push(ctx.Gtx.Ops)
	var path clip.Path
	path.Begin(ctx.Gtx.Ops)
	if s.open {
		path.MoveTo(f32.Pt(1, float32(size)*0.65))
		path.LineTo(f32.Pt(float32(size)*0.5, float32(size)*0.25))
		path.LineTo(f32.Pt(float32(size)-1, float32(size)*0.65))
	} else {
		path.MoveTo(f32.Pt(1, float32(size)*0.35))
		path.LineTo(f32.Pt(float32(size)*0.5, float32(size)*0.75))
		path.LineTo(f32.Pt(float32(size)-1, float32(size)*0.35))
	}
	paint.FillShape(ctx.Gtx.Ops, s.color, clip.Stroke{Path: path.End(), Width: 2}.Op())
	stack.Pop()
	return layout.Dimensions{Size: image.Point{X: box, Y: box}}
}

func selectCheckMark(active bool, col color.NRGBA) Widget {
	return &selectCheckMarkWidget{active: active, color: col}
}

type selectCheckMarkWidget struct {
	active bool
	color  color.NRGBA
}

func (s *selectCheckMarkWidget) Layout(ctx *internal.Context) layout.Dimensions {
	box := ctx.Gtx.Dp(safeDp(24))
	if box < 18 {
		box = 18
	}
	mark := ctx.Gtx.Dp(safeDp(14))
	if mark < 10 {
		mark = 10
	}
	if s.active {
		offset := image.Pt((box-mark)/2, (box-mark)/2)
		stack := op.Offset(offset).Push(ctx.Gtx.Ops)
		internal.DrawCheckMark(ctx.Gtx, mark, s.color)
		stack.Pop()
	}
	return layout.Dimensions{Size: image.Point{X: box, Y: box}}
}
