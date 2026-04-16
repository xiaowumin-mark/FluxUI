package widget

import (
	"fmt"
	"image/color"

	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"
)

// RadioItem 单选项。
type RadioItem struct {
	Label string
	Value string
}

// RadioGroupOption 定义单选组配置。
type RadioGroupOption func(*radioGroupConfig)

type radioGroupConfig struct {
	direction Axis
	disabled  bool
	onChange  func(ctx *internal.Context, value string)
	size      float32
	color     color.NRGBA
	hasColor  bool
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
		size:      18,
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

func (r *radioGroupWidget) Layout(ctx *internal.Context) layout.Dimensions {
	mainColor := ctx.Theme().Primary
	if r.config.hasColor {
		mainColor = r.config.color
	}

	children := make([]Widget, 0, len(r.items))
	for idx := range r.items {
		item := r.items[idx]
		checked := item.Value == r.value

		row := Row(
			Checkbox(
				"",
				checked,
				CheckboxSize(r.config.size),
				CheckboxColor(mainColor),
				CheckboxDisabled(r.config.disabled),
				CheckboxOnChange(func(ctx *internal.Context, checked bool) {
					if r.config.disabled || !checked {
						return
					}
					if r.config.onChange != nil {
						r.config.onChange(ctx, item.Value)
					}
				}),
			),
			Padding(style.Insets{Left: 6}, Text(item.Label)),
		)

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

func (s *selectWidget[T]) Layout(ctx *internal.Context) layout.Dimensions {
	state := selectStateFor(ctx)
	label, currentIndex := s.resolveCurrentLabel()

	arrow := "▼"
	if state.opened {
		arrow = "▲"
	}

	toggle := Button(
		Row(
			Text(label),
			Padding(style.Insets{Left: 8}, Text(arrow, TextColor(ctx.Theme().SurfaceMuted))),
		),
		Disabled(s.config.disabled),
		OnClick(func(ctx *internal.Context) {
			if s.config.disabled {
				return
			}
			state.opened = !state.opened
			if s.config.onOpen != nil {
				s.config.onOpen(ctx, state.opened)
			}
		}),
	)
	toggle = expandWidth(toggle)

	optionsBody := layoutWidgetFunc(func(ctx *internal.Context) layout.Dimensions {
		if !state.opened || len(s.options) == 0 {
			return layout.Dimensions{}
		}

		items := make([]Widget, 0, len(s.options))
		for idx := range s.options {
			item := s.options[idx]
			itemLabel := item.Label
			if itemLabel == "" {
				itemLabel = fmt.Sprintf("%v", item.Value)
			}
			isActive := idx == currentIndex
			bg := color.NRGBA{}
			if isActive {
				p := ctx.Theme().Primary
				bg = color.NRGBA{R: p.R, G: p.G, B: p.B, A: 30}
			}

			row := Button(
				Row(
					Text(itemLabel),
					Padding(style.Insets{Left: 8}, Text(selectMark(isActive), TextColor(ctx.Theme().Primary))),
				),
				ButtonBackground(bg),
				ButtonForeground(ctx.Theme().TextColor),
				ButtonPadding(style.Symmetric(8, 10)),
				ButtonRadius(6),
				OnClick(func(ctx *internal.Context) {
					if s.config.onChange != nil {
						s.config.onChange(ctx, item.Value)
					}
					if state.opened {
						state.opened = false
						if s.config.onOpen != nil {
							s.config.onOpen(ctx, false)
						}
					}
				}),
			)
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

		maxH := s.config.maxHeight
		if maxH <= 0 {
			maxH = 240
		}
		return (&fixedSizeWidget{
			height: maxH,
			child: expandWidth(
				Container(
					style.Style{
						Background: ctx.Theme().Surface,
						Padding:    style.All(6),
						Radius:     8,
					},
					list,
				),
			),
		}).Layout(ctx.Child(0))
	})

	return Column(
		toggle,
		Padding(style.Insets{Top: 6}, optionsBody),
	).Layout(ctx.Child(0))
}

func (s *selectWidget[T]) resolveCurrentLabel() (string, int) {
	label := s.config.placeholder
	currentIndex := -1
	for idx := range s.options {
		if s.options[idx].Value == s.value {
			label = s.options[idx].Label
			currentIndex = idx
			break
		}
	}
	if label == "" {
		label = fmt.Sprintf("%v", s.value)
	}
	return label, currentIndex
}

func selectStateFor(ctx *internal.Context) *selectState {
	value := ctx.Memo("select", func() any {
		return &selectState{}
	})
	state, ok := value.(*selectState)
	if !ok {
		panic("fluxui/widget: select state type mismatch")
	}
	return state
}

func selectMark(active bool) string {
	if active {
		return "●"
	}
	return "○"
}
