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
		r.config.ref.bindInvalidator(redrawInvalidator(ctx))
		commands := r.config.ref.drainCommands()
		if !r.config.disabled {
			for _, next := range commands {
				if next == currentValue {
					continue
				}
				currentValue = next
				if r.config.onChange != nil {
					r.config.onChange(ctx, next)
				}
			}
		}
	}

	cs := ctx.Theme().Colors
	mainColor := cs.Primary
	if r.config.hasColor {
		mainColor = r.config.color
	}
	mainColor = md3AnimateColor(ctx, "radio-main-color", mainColor, style.InteractionSelectedDuration, style.InteractionStandardEasing)
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
			duration, easing := md3InteractionTiming(clickable.Hovered(), clickable.Pressed(), false, r.config.disabled)
			itemLabelColor = md3AnimateColor(rowCtx, "radio-label", itemLabelColor, duration, easing)

			content := func(contentCtx *internal.Context) image.Point {
				dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx,
					gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						checkedProgress := md3SelectionProgress(&next, checked)
						size := next.LayoutRadio(nil, checked, internal.RadioSpec{
							Size:            r.config.size,
							Color:           mainColor,
							CheckedProgress: checkedProgress,
							Disabled:        r.config.disabled,
							Hovered:         clickable.Hovered(),
							Pressed:         clickable.Pressed(),
						})
						return gioLayout.Dimensions{Size: size}
					}),
					gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						next.Gtx.Constraints.Min = image.Point{}
						size := next.LayoutInset(internal.Insets{Left: 8}, func(labelCtx *internal.Context) image.Point {
							return Text(item.Label, TextType(labelCtx.Theme().Types.BodyMedium), TextColor(itemLabelColor)).Layout(labelCtx.Child(0)).Size
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
					visual := md3AnimateDecoration(targetCtx, "radio-decoration", stripStateDecoration(active), duration, easing)
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
	if s.config.disabled && state.opened {
		state.opened = false
	}
	if s.config.ref != nil {
		s.config.ref.bindInvalidator(redrawInvalidator(ctx))
		commands := s.config.ref.drainCommands()
		if !s.config.disabled {
			for _, cmd := range commands {
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
		duration, easing := md3InteractionTiming(clickable.Hovered(), clickable.Pressed(), state.opened || clickable.Focused(triggerCtx), s.config.disabled)
		textColor = md3AnimateColor(triggerCtx, "select-text", textColor, duration, easing)
		arrowColor = md3AnimateColor(triggerCtx, "select-arrow-color", arrowColor, duration, easing)
		arrowProgress := md3AnimateFloatDirectional(
			triggerCtx,
			"select-arrow-progress",
			boolProgress(state.opened),
			style.InteractionMenuEnterDuration,
			style.InteractionMenuExitDuration,
			style.InteractionEmphasizedDecelerateEasing,
			style.InteractionEmphasizedAccelerateEasing,
		)

		content := layoutWidgetFunc(func(contentCtx *internal.Context) layout.Dimensions {
			dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx,
				gioLayout.Flexed(1, func(gtx gioLayout.Context) gioLayout.Dimensions {
					next := *contentCtx
					next.Gtx = gtx
					dims := Text(label, TextType(triggerCtx.Theme().Types.BodyLarge), TextColor(textColor)).Layout(next.Child(0))
					return gioLayout.Dimensions{Size: dims.Size}
				}),
				gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
					next := *contentCtx
					next.Gtx = gtx
					next.Gtx.Constraints.Min = image.Point{}
					dims := Padding(style.Insets{Left: 8}, selectArrowProgress(arrowProgress, arrowColor)).Layout(next.Child(1))
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
			Padding:    s.config.decoration.ResolvePad(densityInsets(triggerCtx, style.Insets{Top: 8, Right: 12, Bottom: 8, Left: 16}, style.Insets{Top: 4, Right: 12, Bottom: 4, Left: 16})),
			Border:     border,
			MinHeight:  densityHeight(triggerCtx, 56, 48),
			FillWidth:  true,
			Disabled:   s.config.disabled,
		}, content)
	})

	toggleDims := toggle.Layout(ctx.Child(0))
	popupProgress, popupVisible := md3OverlayProgress(
		ctx,
		"select-popup",
		state.opened && len(s.options) > 0,
		style.InteractionMenuEnterDuration,
		style.InteractionMenuExitDuration,
		style.InteractionEmphasizedDecelerateEasing,
		style.InteractionEmphasizedAccelerateEasing,
	)
	if !popupVisible {
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
			if !s.config.disabled {
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
			}

			cs := rowCtx.Theme().Colors
			bg := color.NRGBA{}
			if isActive {
				bg = style.StateLayer(color.NRGBA{}, cs.Primary, style.StateLayerFocusOpacity)
			}
			if opacity := materialAnimatedStateLayerOpacity(rowCtx, clickable.Hovered(), clickable.Pressed(), s.config.disabled); opacity > 0 {
				bg = style.StateLayer(bg, cs.Primary, opacity)
			}
			duration, easing := md3InteractionTiming(clickable.Hovered(), clickable.Pressed(), clickable.Focused(rowCtx), s.config.disabled)
			bg = md3AnimateColor(rowCtx, "select-row-bg", bg, duration, easing)
			itemRow := layoutWidgetFunc(func(contentCtx *internal.Context) layout.Dimensions {
				dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx,
					gioLayout.Flexed(1, func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						dims := Text(itemLabel, TextColor(cs.OnSurface), TextType(rowCtx.Theme().Types.BodyMedium)).Layout(next.Child(0))
						return gioLayout.Dimensions{Size: dims.Size}
					}),
					gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						next.Gtx.Constraints.Min = image.Point{}
						progress := md3SelectionProgress(rowCtx, isActive)
						dims := Padding(style.Insets{Left: 12}, selectCheckMarkProgress(progress, cs.Primary)).Layout(next.Child(1))
						return gioLayout.Dimensions{Size: dims.Size}
					}),
				)
				return layout.Dimensions{Size: dims.Size}
			})
			itemContent := ContainerDecoration(
				style.Decoration{}.WithBg(bg).WithPad(densityInsets(rowCtx, style.Symmetric(8, 12), style.Symmetric(6, 12))).WithRad(rowCtx.Theme().Shapes.ExtraSmall),
				itemRow,
			)
			size := rowCtx.LayoutRippleArea(clickable.Handle(), internal.RippleSpec{
				Color:   cs.Primary,
				Radius:  rowCtx.Theme().Shapes.ExtraSmall,
				Opacity: style.StateLayerPressedOpacity,
			}, func(childCtx *internal.Context) image.Point {
				return itemContent.Layout(childCtx.Child(0)).Size
			})
			md3DrawFocusIndicator(rowCtx, size, internal.FocusIndicatorSpec{
				Color:  cs.Primary,
				Radius: rowCtx.Theme().Shapes.ExtraSmall,
			}, clickable.Focused(rowCtx), s.config.disabled)
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

	rowHeightPx := ctx.Gtx.Dp(safeDp(densityHeight(ctx, 40, 36)))
	rowSpacingPx := ctx.Gtx.Dp(safeDp(4))
	panelPad := s.config.decoration.ResolvePad(style.All(6))
	estimatedHeightPx := ctx.Gtx.Dp(safeDp(panelPad.Top+panelPad.Bottom)) + rowHeightPx*len(items)
	if len(items) > 1 {
		estimatedHeightPx += rowSpacingPx * (len(items) - 1)
	}
	if estimatedHeightPx <= 0 || estimatedHeightPx > maxHPx {
		estimatedHeightPx = maxHPx
	}
	gapPx := ctx.Gtx.Dp(safeDp(6))
	popupW := toggleDims.Size.X
	if popupW <= 0 {
		popupW = ctx.Gtx.Constraints.Max.X
	}
	if maxW := ctx.Gtx.Constraints.Max.X; maxW > 0 && popupW > maxW {
		popupW = maxW
	}
	if popupW <= 0 {
		popupW = 1
	}
	placement := md3PopupPlacementForAnchor(ctx, toggleDims.Size, image.Point{X: popupW}, estimatedHeightPx, gapPx)

	popupMacro := op.Record(ctx.Gtx.Ops)
	popupCtx := *ctx
	popupCtx.Gtx = ctx.Gtx
	popupCtx.Gtx.Constraints.Min = image.Point{}
	popupCtx.Gtx.Constraints.Max = image.Point{X: popupW, Y: placement.MaxHeightPx}
	popupSize := layoutMD3OverlayTransition(popupCtx.Child(1), popupProgress, placement.TransitionOffset, func(transitionCtx *internal.Context) image.Point {
		return panel.Layout(transitionCtx.Child(0)).Size
	})
	popupCall := popupMacro.Stop()
	deferMacro := op.Record(ctx.Gtx.Ops)
	offset := op.Offset(image.Point{
		X: placement.OffsetX,
		Y: md3PopupOffsetY(toggleDims.Size.Y, popupSize.Y, placement),
	}).Push(ctx.Gtx.Ops)
	popupCall.Add(ctx.Gtx.Ops)
	offset.Pop()
	deferCall := deferMacro.Stop()
	op.Defer(ctx.Gtx.Ops, deferCall)

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
	return selectArrowProgress(boolProgress(open), col)
}

func selectArrowProgress(progress float32, col color.NRGBA) Widget {
	return &selectArrowWidget{progress: clampFloat32(progress, 0, 1), color: col}
}

type selectArrowWidget struct {
	progress float32
	color    color.NRGBA
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
	leftY := float32(size) * (0.35 + 0.30*s.progress)
	midY := float32(size) * (0.75 - 0.50*s.progress)
	rightY := leftY
	path.MoveTo(f32.Pt(1, leftY))
	path.LineTo(f32.Pt(float32(size)*0.5, midY))
	path.LineTo(f32.Pt(float32(size)-1, rightY))
	paint.FillShape(ctx.Gtx.Ops, s.color, clip.Stroke{Path: path.End(), Width: 2}.Op())
	stack.Pop()
	return layout.Dimensions{Size: image.Point{X: box, Y: box}}
}

func selectCheckMark(active bool, col color.NRGBA) Widget {
	return selectCheckMarkProgress(boolProgress(active), col)
}

func selectCheckMarkProgress(progress float32, col color.NRGBA) Widget {
	return &selectCheckMarkWidget{progress: clampFloat32(progress, 0, 1), color: col}
}

type selectCheckMarkWidget struct {
	progress float32
	color    color.NRGBA
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
	if s.progress > 0 {
		offset := image.Pt((box-mark)/2, (box-mark)/2)
		stack := op.Offset(offset).Push(ctx.Gtx.Ops)
		col := s.color
		col.A = uint8(float32(col.A)*s.progress + 0.5)
		internal.DrawCheckMark(ctx.Gtx, mark, col)
		stack.Pop()
	}
	return layout.Dimensions{Size: image.Point{X: box, Y: box}}
}
