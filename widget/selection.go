package widget

import (
	"fmt"
	"image"
	"image/color"
	"time"

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
			activate := func(actionCtx *internal.Context) {
				if item.Value == currentValue {
					return
				}
				currentValue = item.Value
				if r.config.onChange != nil {
					r.config.onChange(actionCtx, item.Value)
				}
			}
			registerClickableFocusAction(rowCtx, r.config.disabled, activate)
			drainClickableDefaultAction(rowCtx, clickable, r.config.disabled, activate)
			interaction := clickableInteractionSnapshot(rowCtx, clickable, r.config.disabled, false)

			itemLabelColor := labelColor
			if !r.config.disabled && interaction.Hovered {
				itemLabelColor = style.StateLayer(itemLabelColor, mainColor, style.StateLayerHoverOpacity)
			}
			duration, easing := md3InteractionTiming(rowCtx, interaction.Hovered, interaction.Pressed, false, r.config.disabled)
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
							Hovered:         interaction.Hovered,
							Pressed:         interaction.Pressed,
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
					active := resolveDecorationState(deco, interaction.Hovered, interaction.Pressed, r.config.disabled)
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
				return materialStateLayerOpacity(rowCtx, interaction.Hovered, interaction.Pressed)
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
	Label         string
	Value         T
	Disabled      bool
	Leading       Widget
	Trailing      Widget
	TypeaheadText string
}

// SelectOption 定义下拉配置。
type SelectOption[T comparable] func(*selectConfig[T])

type selectVariant int

const (
	selectVariantOutlined selectVariant = iota
	selectVariantFilled
)

type selectConfig[T comparable] struct {
	variant        selectVariant
	placeholder    string
	label          string
	supportingText string
	errorText      string
	error          bool
	required       bool
	noAsterisk     bool
	disabled       bool
	searchable     bool
	quick          bool
	maxHeight      float32
	width          float32
	xOffset        float32
	yOffset        float32
	typeaheadDelay time.Duration
	onChange       func(ctx *internal.Context, value T)
	onOpen         func(ctx *internal.Context, opened bool)
	ref            *SelectRef[T]
	leading        Widget
	trailing       Widget
	decoration     style.Decoration
	menuDecoration style.Decoration
}

type selectWidget[T comparable] struct {
	value   T
	options []SelectOptionItem[T]
	config  selectConfig[T]
}

type selectState struct {
	opened     bool
	outsideTag any
}

// Select 创建下拉选择组件。
func Select[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget {
	return newSelect(value, options, selectVariantOutlined, opts...)
}

func OutlinedSelect[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget {
	return newSelect(value, options, selectVariantOutlined, opts...)
}

func FilledSelect[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget {
	return newSelect(value, options, selectVariantFilled, opts...)
}

func newSelect[T comparable](value T, options []SelectOptionItem[T], variant selectVariant, opts ...SelectOption[T]) Widget {
	cfg := selectConfig[T]{
		placeholder: "请选择",
		maxHeight:   240,
		variant:     variant,
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

func SelectLabel[T comparable](text string) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.label = text
	}
}

func SelectSupportingText[T comparable](text string) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.supportingText = text
	}
}

func SelectErrorText[T comparable](text string) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.errorText = text
	}
}

func SelectError[T comparable](error bool) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.error = error
	}
}

func SelectRequired[T comparable](required bool) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.required = required
	}
}

func SelectNoAsterisk[T comparable](noAsterisk bool) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.noAsterisk = noAsterisk
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

func SelectWidth[T comparable](width float32) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.width = width
	}
}

func SelectXOffset[T comparable](offset float32) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.xOffset = offset
	}
}

func SelectYOffset[T comparable](offset float32) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.yOffset = offset
	}
}

func SelectQuick[T comparable](quick bool) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.quick = quick
	}
}

func SelectTypeaheadDelay[T comparable](delay time.Duration) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.typeaheadDelay = delay
	}
}

func SelectFilled[T comparable](filled bool) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		if filled {
			cfg.variant = selectVariantFilled
		} else {
			cfg.variant = selectVariantOutlined
		}
	}
}

func SelectLeading[T comparable](leading Widget) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.leading = leading
	}
}

func SelectTrailing[T comparable](trailing Widget) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.trailing = trailing
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

func SelectMenuDecoration[T comparable](d style.Decoration) SelectOption[T] {
	return func(cfg *selectConfig[T]) {
		cfg.menuDecoration = d
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
	fieldCtx := ctx.Child(0)
	toggleDims := s.layoutSelectField(fieldCtx, label, currentIndex >= 0, state)
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

	optionRow := func(_ *internal.Context, index int) Widget {
		item := s.options[index]
		itemLabel := item.Label
		if itemLabel == "" {
			itemLabel = fmt.Sprintf("%v", item.Value)
		}
		isActive := index == currentIndex
		var row Widget = layoutWidgetFunc(func(rowCtx *internal.Context) layout.Dimensions {
			clickable := event.UseClickable(rowCtx)
			disabled := s.config.disabled || item.Disabled
			activate := func(actionCtx *internal.Context) {
				s.activateOptionValue(actionCtx, state, fieldCtx, &currentValue, item.Value)
			}
			registerClickableFocusAction(rowCtx, disabled, activate)
			drainClickableDefaultAction(rowCtx, clickable, disabled, activate)
			interaction := clickableInteractionSnapshot(rowCtx, clickable, disabled, true)

			cs := rowCtx.Theme().Colors
			bg := color.NRGBA{}
			if isActive {
				bg = cs.SecondaryContainer
			}
			if opacity := materialAnimatedStateLayerOpacity(rowCtx, interaction.Hovered, interaction.Pressed, disabled); opacity > 0 {
				bg = style.StateLayer(bg, cs.Primary, opacity)
			}
			duration, easing := md3InteractionTiming(rowCtx, interaction.Hovered, interaction.Pressed, interaction.Focused, disabled)
			bg = md3AnimateColor(rowCtx, "select-row-bg", bg, duration, easing)
			fg := cs.OnSurface
			if isActive {
				fg = cs.OnSecondaryContainer
			}
			if disabled {
				fg = style.DisabledContent(cs.OnSurface)
			}
			fg = md3AnimateColor(rowCtx, "select-row-fg", fg, duration, easing)
			itemRow := layoutWidgetFunc(func(contentCtx *internal.Context) layout.Dimensions {
				children := make([]gioLayout.FlexChild, 0, 5)
				if item.Leading != nil {
					children = append(children, gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						next.Gtx.Constraints.Min = image.Point{}
						dims := FixedWidth(24, Center(withForeground(fg, item.Leading))).Layout(next.Child(0))
						return gioLayout.Dimensions{Size: dims.Size}
					}), gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						return gioLayout.Dimensions{Size: image.Point{X: contentCtx.Gtx.Dp(safeDp(12))}}
					}))
				}
				children = append(children, gioLayout.Flexed(1, func(gtx gioLayout.Context) gioLayout.Dimensions {
					next := *contentCtx
					next.Gtx = gtx
					dims := Text(itemLabel, TextColor(fg), TextType(rowCtx.Theme().Types.BodyLarge)).Layout(next.Child(1))
					return gioLayout.Dimensions{Size: dims.Size}
				}))
				trailing := item.Trailing
				if trailing == nil && isActive {
					trailing = selectCheckMarkProgress(md3SelectionProgress(rowCtx, isActive), cs.Primary)
				}
				if trailing != nil {
					children = append(children, gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
						next := *contentCtx
						next.Gtx = gtx
						next.Gtx.Constraints.Min = image.Point{}
						dims := Padding(style.Insets{Left: 12}, FixedWidth(24, Center(withForeground(fg, trailing)))).Layout(next.Child(2))
						return gioLayout.Dimensions{Size: dims.Size}
					}))
				}
				dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx, children...)
				return layout.Dimensions{Size: dims.Size}
			})
			itemContent := ContainerDecoration(
				style.Decoration{}.WithBg(bg).WithPad(densityInsets(rowCtx, style.Symmetric(12, 16), style.Symmetric(8, 12))).WithRad(0),
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
			}, interaction.Focused, disabled)
			return layout.Dimensions{Size: size}
		})
		row = expandWidth(row)
		return row
	}

	list := ListView(
		len(s.options),
		func(ctx *internal.Context, index int) Widget {
			return optionRow(ctx, index)
		},
		ListItemSpacing(0),
		ListVirtualized(true),
	)
	panel := expandWidth(
		ContainerDecoration(
			style.Decoration{}.
				WithBg(s.config.menuDecoration.ResolveBg(ctx.Theme().Colors.SurfaceContainer)).
				WithPad(s.config.menuDecoration.ResolvePad(style.Insets{})).
				WithRad(s.config.menuDecoration.ResolveRad(ctx.Theme().Shapes.ExtraSmall)).
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

	rowHeightPx := ctx.Gtx.Dp(safeDp(densityHeight(ctx, 48, 40)))
	rowSpacingPx := 0
	panelPad := s.config.menuDecoration.ResolvePad(style.Insets{})
	estimatedHeightPx := ctx.Gtx.Dp(safeDp(panelPad.Top+panelPad.Bottom)) + rowHeightPx*len(s.options)
	if len(s.options) > 1 {
		estimatedHeightPx += rowSpacingPx * (len(s.options) - 1)
	}
	if estimatedHeightPx <= 0 || estimatedHeightPx > maxHPx {
		estimatedHeightPx = maxHPx
	}
	gapPx := ctx.Gtx.Dp(safeDp(0))
	popupW := toggleDims.Size.X
	if s.config.width > 0 {
		popupW = ctx.Gtx.Dp(safeDp(s.config.width))
	}
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

	recordPopup := func(p md3PopupPlacement) (image.Point, op.CallOp) {
		popupMacro := op.Record(ctx.Gtx.Ops)
		popupCtx := *ctx
		popupCtx.Gtx = ctx.Gtx
		popupCtx.Gtx.Constraints.Min = image.Point{}
		popupCtx.Gtx.Constraints.Max = image.Point{X: popupW, Y: p.MaxHeightPx}
		popupSize := layoutMD3RevealTransition(popupCtx.Child(1), popupProgress, p.Direction == md3PopupUp, func(transitionCtx *internal.Context) image.Point {
			return panel.Layout(transitionCtx.Child(0)).Size
		})
		return popupSize, popupMacro.Stop()
	}
	popupSize, _ := recordPopup(placement)
	placement = md3PopupPlacementForMeasuredPopup(ctx, toggleDims.Size, popupSize, placement, ctx.Gtx.Dp(safeDp(s.config.yOffset)))
	popupSize, popupCall := recordPopup(placement)
	deferMacro := op.Record(ctx.Gtx.Ops)
	offsetPoint := image.Point{
		X: placement.OffsetX + ctx.Gtx.Dp(safeDp(s.config.xOffset)),
		Y: md3PopupOffsetY(toggleDims.Size.Y, popupSize.Y, placement) + ctx.Gtx.Dp(safeDp(s.config.yOffset)),
	}
	if state.opened {
		origin := ctx.Position()
		fieldRect := image.Rectangle{Min: origin, Max: origin.Add(toggleDims.Size)}
		popupOrigin := origin.Add(offsetPoint)
		popupRect := image.Rectangle{Min: popupOrigin, Max: popupOrigin.Add(popupSize)}
		md3DismissOnOutsidePress(ctx, state.outsideTag, []image.Rectangle{fieldRect, popupRect}, func(dismissCtx *internal.Context) {
			if state.opened {
				state.opened = false
				if s.config.onOpen != nil {
					s.config.onOpen(dismissCtx, false)
				}
				md3ClearFocusIfInside(dismissCtx, ctx.PathID())
			}
		})
		md3RegisterEscapeClose(ctx, func(escapeCtx *internal.Context) {
			if state.opened {
				state.opened = false
				if s.config.onOpen != nil {
					s.config.onOpen(escapeCtx, false)
				}
				event.RequestFocus(fieldCtx)
			}
		})
		md3ProcessOverlayKeyboardEvents(ctx)
	}
	offset := op.Offset(offsetPoint).Push(ctx.Gtx.Ops)
	popupCall.Add(ctx.Gtx.Ops)
	offset.Pop()
	deferCall := deferMacro.Stop()
	op.Defer(ctx.Gtx.Ops, deferCall)

	return toggleDims
}

func (s *selectWidget[T]) activateOptionValue(actionCtx *internal.Context, state *selectState, fieldCtx *internal.Context, currentValue *T, next T) {
	if currentValue == nil {
		return
	}
	if next != *currentValue {
		*currentValue = next
		if s.config.onChange != nil {
			s.config.onChange(actionCtx, next)
		}
	}
	if state != nil && state.opened {
		state.opened = false
		if s.config.onOpen != nil {
			s.config.onOpen(actionCtx, false)
		}
		if fieldCtx != nil {
			event.RequestFocus(fieldCtx)
		}
	}
}

func (s *selectWidget[T]) layoutSelectField(ctx *internal.Context, valueLabel string, hasValue bool, state *selectState) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	activate := func(actionCtx *internal.Context) {
		event.RequestFocus(ctx)
		state.opened = !state.opened
		if s.config.onOpen != nil {
			s.config.onOpen(actionCtx, state.opened)
		}
	}
	registerClickableFocusAction(ctx, s.config.disabled, activate)
	drainClickableDefaultAction(ctx, clickable, s.config.disabled, activate)
	interaction := clickableInteractionSnapshot(ctx, clickable, s.config.disabled, true)

	cs := ctx.Theme().Colors
	focused := state.opened || interaction.Focused
	errored := s.config.error
	textColor := cs.OnSurface
	supportColor := cs.OnSurfaceVariant
	labelColor := cs.OnSurfaceVariant
	arrowColor := cs.OnSurfaceVariant
	outlineColor := cs.Outline
	indicatorColor := cs.OnSurfaceVariant
	if focused {
		labelColor = cs.Primary
		outlineColor = cs.Primary
		indicatorColor = cs.Primary
		arrowColor = cs.Primary
	}
	if errored {
		labelColor = cs.Error
		outlineColor = cs.Error
		indicatorColor = cs.Error
		supportColor = cs.Error
	}
	if !hasValue {
		textColor = cs.OnSurfaceVariant
	}
	if s.config.disabled {
		textColor = style.DisabledContent(cs.OnSurface)
		supportColor = style.DisabledContent(cs.OnSurface)
		labelColor = style.DisabledContent(cs.OnSurface)
		arrowColor = style.DisabledContent(cs.OnSurface)
		outlineColor = style.DisabledContent(cs.OnSurface)
		indicatorColor = style.DisabledContent(cs.OnSurface)
	}

	duration, easing := md3InteractionTiming(ctx, interaction.Hovered, interaction.Pressed, focused, s.config.disabled)
	textColor = md3AnimateColor(ctx, "select-text", textColor, duration, easing)
	labelColor = md3AnimateColor(ctx, "select-label", labelColor, duration, easing)
	arrowColor = md3AnimateColor(ctx, "select-arrow-color", arrowColor, duration, easing)
	outlineColor = md3AnimateColor(ctx, "select-outline", outlineColor, duration, easing)
	indicatorColor = md3AnimateColor(ctx, "select-indicator", indicatorColor, duration, easing)
	arrowProgress := md3AnimateFloatDirectional(
		ctx,
		"select-arrow-progress",
		boolProgress(state.opened),
		style.InteractionMenuEnterDuration,
		style.InteractionMenuExitDuration,
		style.InteractionEmphasizedDecelerateEasing,
		style.InteractionEmphasizedAccelerateEasing,
	)

	fieldLabel := selectRequiredLabel(s.config.label, s.config.required, s.config.noAsterisk)
	displayText := valueLabel
	if !hasValue && s.config.label != "" {
		displayText = s.config.placeholder
	}
	if displayText == "" {
		displayText = s.config.placeholder
	}
	content := layoutWidgetFunc(func(contentCtx *internal.Context) layout.Dimensions {
		availableW := contentCtx.Gtx.Constraints.Max.X
		if availableW <= 0 {
			availableW = contentCtx.Gtx.Dp(safeDp(240))
		}
		gapPx := contentCtx.Gtx.Dp(safeDp(12))
		leadingW := 0
		var leadingCall op.CallOp
		if s.config.leading != nil {
			macro := op.Record(contentCtx.Gtx.Ops)
			next := *contentCtx
			next.Gtx.Constraints.Min = image.Point{}
			next.Gtx.Constraints.Max.X = minInt(24+gapPx, availableW)
			leadingSize := FixedWidth(24, Center(withForeground(labelColor, s.config.leading))).Layout(next.Child(0)).Size
			leadingCall = macro.Stop()
			leadingW = leadingSize.X + gapPx
		}
		trailing := s.config.trailing
		if trailing == nil {
			trailing = selectArrowProgress(arrowProgress, arrowColor)
		}
		trailingMacro := op.Record(contentCtx.Gtx.Ops)
		trailingCtx := *contentCtx
		trailingCtx.Gtx.Constraints.Min = image.Point{}
		trailingSize := FixedWidth(24, Center(withForeground(arrowColor, trailing))).Layout(trailingCtx.Child(2)).Size
		trailingCall := trailingMacro.Stop()
		trailingW := trailingSize.X
		if trailingW > 0 {
			trailingW += contentCtx.Gtx.Dp(safeDp(8))
		}

		bodyMaxW := availableW - leadingW - trailingW
		if bodyMaxW < 1 {
			bodyMaxW = 1
		}
		bodyMacro := op.Record(contentCtx.Gtx.Ops)
		bodyCtx := *contentCtx
		bodyCtx.Gtx.Constraints.Min = image.Point{}
		bodyCtx.Gtx.Constraints.Max.X = bodyMaxW
		var body Widget
		valueText := Text(displayText, TextType(ctx.Theme().Types.BodyLarge), TextColor(textColor))
		if fieldLabel != "" {
			body = Column(
				Text(fieldLabel, TextType(ctx.Theme().Types.LabelMedium), TextColor(labelColor)),
				Padding(style.Insets{Top: 2}, valueText),
			)
		} else {
			body = valueText
		}
		bodySize := body.Layout(bodyCtx.Child(1)).Size
		bodyCall := bodyMacro.Stop()

		height := bodySize.Y
		if leadingW > 0 && height < contentCtx.Gtx.Dp(safeDp(24)) {
			height = contentCtx.Gtx.Dp(safeDp(24))
		}
		if height < trailingSize.Y {
			height = trailingSize.Y
		}
		if height < 1 {
			height = 1
		}
		if leadingW > 0 {
			stack := op.Offset(image.Point{Y: (height - contentCtx.Gtx.Dp(safeDp(24))) / 2}).Push(contentCtx.Gtx.Ops)
			leadingCall.Add(contentCtx.Gtx.Ops)
			stack.Pop()
		}
		bodyY := (height - bodySize.Y) / 2
		if bodyY < 0 {
			bodyY = 0
		}
		bodyStack := op.Offset(image.Point{X: leadingW, Y: bodyY}).Push(contentCtx.Gtx.Ops)
		bodyCall.Add(contentCtx.Gtx.Ops)
		bodyStack.Pop()
		trailingX := availableW - trailingSize.X
		if trailingX < leadingW+bodySize.X {
			trailingX = leadingW + bodySize.X
		}
		trailingY := (height - trailingSize.Y) / 2
		if trailingY < 0 {
			trailingY = 0
		}
		trailingStack := op.Offset(image.Point{X: trailingX, Y: trailingY}).Push(contentCtx.Gtx.Ops)
		trailingCall.Add(contentCtx.Gtx.Ops)
		trailingStack.Pop()
		return layout.Dimensions{Size: image.Point{X: availableW, Y: height}}
	})

	bg := cs.Surface
	border := style.Border{Width: 1, Color: outlineColor}
	radius := ctx.Theme().Shapes.ExtraSmall
	if s.config.variant == selectVariantFilled {
		bg = cs.SurfaceContainerHighest
		border = style.Border{}
		radius = ctx.Theme().Shapes.ExtraSmall
	}
	if s.config.disabled && s.config.variant == selectVariantFilled {
		bg = style.DisabledContainer(cs.OnSurface)
	}
	bg = s.config.decoration.ResolveBg(bg)
	padding := s.config.decoration.ResolvePad(densityInsets(ctx, style.Insets{Top: 8, Right: 12, Bottom: 8, Left: 16}, style.Insets{Top: 5, Right: 12, Bottom: 5, Left: 12}))
	border.Width = md3AnimateFloat(ctx, "select-outline-width", selectBorderWidth(s.config.variant, focused, s.config.disabled), duration, easing)

	field := layoutWidgetFunc(func(fieldCtx *internal.Context) layout.Dimensions {
		dims := md3ActionSurface(fieldCtx, clickable, md3ActionSurfaceSpec{
			Background: bg,
			Foreground: textColor,
			Radius:     s.config.decoration.ResolveRad(radius),
			Padding:    padding,
			Border:     border,
			MinHeight:  densityHeight(fieldCtx, 56, 48),
			FillWidth:  true,
			Disabled:   s.config.disabled,
		}, content)
		if s.config.variant == selectVariantFilled {
			thickness := fieldCtx.Gtx.Dp(safeDp(1))
			if focused {
				thickness = fieldCtx.Gtx.Dp(safeDp(2))
			}
			if thickness < 1 {
				thickness = 1
			}
			rect := image.Rect(0, dims.Size.Y-thickness, dims.Size.X, dims.Size.Y)
			paint.FillShape(fieldCtx.Gtx.Ops, indicatorColor, clip.Rect(rect).Op())
		}
		return dims
	})

	var root Widget = field
	supporting := s.config.supportingText
	if errored && s.config.errorText != "" {
		supporting = s.config.errorText
	}
	if supporting != "" {
		root = Column(
			field,
			Padding(style.Insets{Top: 4, Left: 16, Right: 16}, Text(supporting, TextType(ctx.Theme().Types.BodySmall), TextColor(supportColor))),
		)
	}
	if s.config.width > 0 {
		root = FixedWidth(s.config.width, root)
	}
	return root.Layout(ctx.Child(0))
}

func selectBorderWidth(variant selectVariant, focused bool, disabled bool) float32 {
	if variant == selectVariantFilled {
		return 0
	}
	if focused && !disabled {
		return 2
	}
	return 1
}

func selectRequiredLabel(label string, required bool, noAsterisk bool) string {
	if label == "" || !required || noAsterisk {
		return label
	}
	return label + " *"
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
		return &selectState{outsideTag: new(int)}
	})
	state, ok := value.(*selectState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: select state type mismatch")
	}
	if state.outsideTag == nil {
		state.outsideTag = new(int)
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
