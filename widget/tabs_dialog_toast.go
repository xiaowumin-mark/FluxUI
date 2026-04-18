package widget

import (
	"image"
	"image/color"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// TabItem 标签项。
type TabItem struct {
	Key   string
	Label string
}

// TabsOption 标签配置。
type TabsOption func(*tabsConfig)

type tabsConfig struct {
	onChange        func(ctx *internal.Context, key string)
	scrollable      bool
	indicatorColor  color.NRGBA
	hasIndicator    bool
	textColor       color.NRGBA
	hasTextColor    bool
	activeTextColor color.NRGBA
	hasActiveColor  bool
	ref             *TabsRef
}

type tabsWidget struct {
	active string
	items  []TabItem
	config tabsConfig
}

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

func (t *tabsWidget) Layout(ctx *internal.Context) layout.Dimensions {
	activeKey := t.active
	if t.config.ref != nil {
		t.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
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

	normalText := ctx.Theme().TextColor
	if t.config.hasTextColor {
		normalText = t.config.textColor
	}
	activeText := ctx.Theme().Primary
	if t.config.hasActiveColor {
		activeText = t.config.activeTextColor
	}
	indicator := ctx.Theme().Primary
	if t.config.hasIndicator {
		indicator = t.config.indicatorColor
	}

	children := make([]Widget, 0, len(t.items))
	for idx := range t.items {
		item := t.items[idx]
		active := item.Key == activeKey

		txtColor := normalText
		indicatorBar := color.NRGBA{A: 0}
		if active {
			txtColor = activeText
			indicatorBar = indicator
		}

		tab := Button(
			Column(
				Text(item.Label, TextColor(txtColor)),
				Padding(
					style.Insets{Top: 4},
					Container(
						style.Style{
							Background: indicatorBar,
							Radius:     2,
						},
						(&fixedSizeWidget{
							width:  22,
							height: 3,
							child:  Spacer(0, 0),
						}),
					),
				),
			),
			ButtonBackground(color.NRGBA{}),
			ButtonForeground(txtColor),
			ButtonRadius(8),
			ButtonPadding(style.Symmetric(8, 10)),
			OnClick(func(ctx *internal.Context) {
				activeKey = item.Key
				if t.config.onChange != nil {
					t.config.onChange(ctx, item.Key)
				}
			}),
		)
		children = append(children, Padding(style.Insets{Right: 6}, tab))
	}

	row := Row(children...)
	if t.config.scrollable {
		return ScrollView(
			row,
			ScrollHorizontal(true),
			ScrollVertical(false),
		).Layout(ctx.Child(0))
	}
	return row.Layout(ctx.Child(0))
}

// DialogOption 对话框配置。
type DialogOption func(*dialogConfig)

type dialogConfig struct {
	title        string
	width        float32
	radius       float32
	maskClosable bool
	onOpenChange func(ctx *internal.Context, open bool)
	onConfirm    func(ctx *internal.Context)
	onCancel     func(ctx *internal.Context)
	ref          *DialogRef
}

type dialogWidget struct {
	open   bool
	child  Widget
	config dialogConfig
}

type dialogState struct {
	wasOpen bool
}

// Dialog 创建对话框。
func Dialog(open bool, child Widget, opts ...DialogOption) Widget {
	cfg := dialogConfig{
		radius:       12,
		maskClosable: true,
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

// DialogAttachRef 绑定命令型引用，用于外部主动打开/关闭对话框。
func DialogAttachRef(ref *DialogRef) DialogOption {
	return func(cfg *dialogConfig) {
		cfg.ref = ref
	}
}

func (d *dialogWidget) Layout(ctx *internal.Context) layout.Dimensions {
	open := d.open
	if d.config.ref != nil {
		d.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
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
	if d.config.onOpenChange != nil && state.wasOpen != open {
		state.wasOpen = open
		d.config.onOpenChange(ctx, open)
	}
	if !open {
		return layout.Dimensions{}
	}

	mask := fillWidget(func(maskCtx *internal.Context, size image.Point) {
		if size.X <= 0 || size.Y <= 0 {
			return
		}
		paint.FillShape(maskCtx.Gtx.Ops, color.NRGBA{A: 120}, clip.Rect(image.Rectangle{Max: size}).Op())
	}, d.config.maskClosable && d.config.onOpenChange != nil, func(maskCtx *internal.Context) {
		d.config.onOpenChange(maskCtx, false)
	})

	parts := make([]Widget, 0, 3)
	if d.config.title != "" {
		parts = append(parts, Padding(style.Insets{Bottom: 8}, Text(d.config.title, TextSize(18))))
	}
	if d.child != nil {
		parts = append(parts, d.child)
	}

	actions := make([]Widget, 0, 2)
	if d.config.onCancel != nil {
		actions = append(actions, Button(Text("取消"), OnClick(d.config.onCancel)))
	}
	if d.config.onConfirm != nil {
		actions = append(actions, Padding(style.Insets{Left: 8}, Button(Text("确定"), OnClick(d.config.onConfirm))))
	}
	if len(actions) > 0 {
		parts = append(parts, Padding(style.Insets{Top: 12}, Row(actions...)))
	}

	panel := Container(
		style.Style{
			Background: ctx.Theme().Surface,
			Padding:    style.All(12),
			Radius:     d.config.radius,
		},
		Column(parts...),
	)
	if d.config.width > 0 {
		panel = &fixedSizeWidget{
			width: d.config.width,
			child: panel,
		}
	}

	content := anchoredOverlayWidget(panel, gioLayout.Center)

	return Stack(
		mask,
		content,
	).Layout(ctx.Child(0))
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
	kind     ToastType
	duration time.Duration
	position ToastPosition
	onClose  func(ctx *internal.Context)
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
	return func(cfg *toastConfig) {
		cfg.onClose = fn
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
			ctx.RequestRedraw()
		}
	}

	if state.closed {
		return layout.Dimensions{}
	}

	bg := color.NRGBA{R: 60, G: 60, B: 60, A: 220}
	switch t.config.kind {
	case ToastSuccess:
		bg = color.NRGBA{R: 40, G: 167, B: 69, A: 220}
	case ToastWarning:
		bg = color.NRGBA{R: 255, G: 193, B: 7, A: 230}
	case ToastError:
		bg = color.NRGBA{R: 220, G: 53, B: 69, A: 230}
	}

	body := Container(
		style.Style{
			Background: bg,
			Padding:    style.Symmetric(8, 12),
			Radius:     8,
		},
		Text(t.message, TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
	)

	anchor := gioLayout.S
	switch t.config.position {
	case ToastTop:
		anchor = gioLayout.N
	case ToastCenter:
		anchor = gioLayout.Center
	}

	return anchoredOverlayWidget(
		Padding(style.Insets{Top: 8, Bottom: 8, Left: 8, Right: 8}, body),
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

	return ClickArea(
		layoutWidgetFunc(func(btnCtx *internal.Context) layout.Dimensions {
			return layoutFill(btnCtx.Child(0))
		}),
		f.onClick,
	).Layout(ctx.Child(0))
}
