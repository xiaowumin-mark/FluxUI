package widget

import (
	"image"
	"image/color"
	"time"

	"github.com/xiaowumin-mark/FluxUI/event"
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
	decoration      style.Decoration
	tabDecoration   style.Decoration
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

	children := make([]Widget, 0, len(t.items))
	for idx := range t.items {
		item := t.items[idx]
		active := item.Key == activeKey

		txtColor := normalText
		indicatorBar := color.NRGBA{A: 0}
		tabBg := color.NRGBA{}
		if active {
			txtColor = activeText
			indicatorBar = indicator
			tabBg = withAlpha(indicator, 20)
		}

		tabStates := withDefaultStates(t.config.tabDecoration,
			style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, indicator, style.StateLayerHoverOpacity)),
			style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, indicator, style.StateLayerPressedOpacity)),
			style.Decoration{}.WithBg(style.DisabledContainer(cs.OnSurface)),
		)
		tab := layoutWidgetFunc(func(tabCtx *internal.Context) layout.Dimensions {
			clickable := event.UseClickable(tabCtx)
			for clickable.Clicked(tabCtx) {
				activeKey = item.Key
				if t.config.onChange != nil {
					t.config.onChange(tabCtx, item.Key)
				}
			}
			activeDecoration := resolveDecorationState(tabStates, clickable.Hovered(), clickable.Pressed(), false)
			tabDecoration := componentDecoration(activeDecoration, tabBg, style.Symmetric(8, 10), 8)
			tabContent := ContainerDecoration(
				tabDecoration,
				Column(
					Text(item.Label, TextColor(txtColor), TextType(tabCtx.Theme().Types.LabelLarge)),
					Padding(
						style.Insets{Top: 4},
						ContainerDecoration(
							style.Decoration{}.WithBg(indicatorBar).WithRad(2),
							(&fixedSizeWidget{
								width:  22,
								height: 3,
								child:  Spacer(0, 0),
							}),
						),
					),
				),
			)
			size := tabCtx.LayoutRippleArea(clickable.Handle(), internal.RippleSpec{
				Color:   txtColor,
				Radius:  8,
				Opacity: style.StateLayerPressedOpacity,
			}, func(childCtx *internal.Context) image.Point {
				return tabContent.Layout(childCtx.Child(0)).Size
			})
			if clickable.Focused(tabCtx) {
				tabCtx.DrawFocusIndicator(size, internal.FocusIndicatorSpec{
					Color:  indicator,
					Radius: 8,
				})
			}
			return layout.Dimensions{Size: size}
		})
		children = append(children, Padding(style.Insets{Right: 6}, tab))
	}

	row := Row(children...)
	if hasDecorationVisual(t.config.decoration) {
		row = ContainerDecoration(t.config.decoration, row)
	}
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
	confirmText  string
	cancelText   string
	onOpenChange func(ctx *internal.Context, open bool)
	onConfirm    func(ctx *internal.Context)
	onCancel     func(ctx *internal.Context)
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
	wasOpen bool
}

// Dialog 创建对话框。
func Dialog(open bool, child Widget, opts ...DialogOption) Widget {
	cfg := dialogConfig{
		maskClosable: true,
		confirmText:  "确定",
		cancelText:   "取消",
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

	cs := ctx.Theme().Colors
	maskColor := withAlpha(cs.Scrim, 120)
	if d.config.hasMaskColor {
		maskColor = d.config.maskColor
	}
	mask := fillWidget(func(maskCtx *internal.Context, size image.Point) {
		if size.X <= 0 || size.Y <= 0 {
			return
		}
		paint.FillShape(maskCtx.Gtx.Ops, maskColor, clip.Rect(image.Rectangle{Max: size}).Op())
	}, d.config.maskClosable && d.config.onOpenChange != nil, func(maskCtx *internal.Context) {
		d.config.onOpenChange(maskCtx, false)
	})

	parts := make([]Widget, 0, 3)
	if d.config.title != "" {
		parts = append(parts, Padding(style.Insets{Bottom: 8}, Text(d.config.title, TextType(ctx.Theme().Types.HeadlineSmall), TextColor(cs.OnSurface))))
	}
	if d.child != nil {
		parts = append(parts, withTextStyle(ctx.Theme().Types.BodyMedium, d.child))
	}

	actions := make([]Widget, 0, 2)
	if d.config.onCancel != nil {
		actions = append(actions, Button(Text(d.config.cancelText), OnClick(d.config.onCancel)))
	}
	if d.config.onConfirm != nil {
		actions = append(actions, Padding(style.Insets{Left: 8}, Button(Text(d.config.confirmText), OnClick(d.config.onConfirm))))
	}
	if len(actions) > 0 {
		parts = append(parts, Padding(style.Insets{Top: 12}, Row(actions...)))
	}

	radiusDefault := ctx.Theme().Shapes.ExtraLarge
	if d.config.radius > 0 {
		radiusDefault = d.config.radius
	}
	panelDeco := style.Decoration{}.
		WithBg(d.config.decoration.ResolveBg(cs.SurfaceContainerHigh)).
		WithPad(d.config.decoration.ResolvePad(style.All(24))).
		WithRad(d.config.decoration.ResolveRad(radiusDefault))
	if d.config.decoration.Shadow != nil {
		panelDeco = panelDeco.WithShadow(*d.config.decoration.Shadow)
	} else {
		panelDeco = panelDeco.WithShadow(style.ElevationShadow(cs, 3))
	}
	if d.config.decoration.Border != nil {
		panelDeco = panelDeco.WithBorder(*d.config.decoration.Border)
	}
	panel := ContainerDecoration(panelDeco, Column(parts...))
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
			ctx.RequestFrameRedraw()
		}
	}

	if state.closed {
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

	body := ContainerDecoration(
		style.Decoration{}.WithBg(bg).WithPad(t.config.decoration.ResolvePad(style.Symmetric(8, 16))).WithRad(t.config.decoration.ResolveRad(ctx.Theme().Shapes.ExtraSmall)),
		content,
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

type popupConfig struct {
	width         float32
	radius        float32
	maskClosable  bool
	background    color.NRGBA
	hasBackground bool
	padding       style.Insets
	hasPadding    bool
	onOpenChange  func(ctx *internal.Context, open bool)
	ref           *PopupRef
	decoration    style.Decoration
	maskColor     color.NRGBA
	hasMaskColor  bool
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

// PopupOnOpenChange 监听弹窗打开/关闭。
func PopupOnOpenChange(fn func(ctx *internal.Context, open bool)) PopupOption {
	return func(cfg *popupConfig) {
		cfg.onOpenChange = fn
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
		p.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
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
		state.wasOpen = open
		p.config.onOpenChange(ctx, open)
	}
	if !open {
		return layout.Dimensions{}
	}

	cs := ctx.Theme().Colors
	maskColor := withAlpha(cs.Scrim, 120)
	if p.config.hasMaskColor {
		maskColor = p.config.maskColor
	}
	mask := fillWidget(func(maskCtx *internal.Context, size image.Point) {
		if size.X <= 0 || size.Y <= 0 {
			return
		}
		paint.FillShape(maskCtx.Gtx.Ops, maskColor, clip.Rect(image.Rectangle{Max: size}).Op())
	}, p.config.maskClosable && p.config.onOpenChange != nil, func(maskCtx *internal.Context) {
		p.config.onOpenChange(maskCtx, false)
	})

	bg := p.config.decoration.ResolveBg(cs.SurfaceContainer)
	if p.config.hasBackground {
		bg = p.config.background
	}
	padding := p.config.decoration.ResolvePad(style.All(0))
	if p.config.hasPadding {
		padding = p.config.padding
	}
	radiusDefault := ctx.Theme().Shapes.Small
	if p.config.radius > 0 {
		radiusDefault = p.config.radius
	}
	radius := p.config.decoration.ResolveRad(radiusDefault)

	var panel Widget
	panelDeco := style.Decoration{}.WithBg(bg).WithPad(padding).WithRad(radius)
	if p.config.decoration.Shadow != nil {
		panelDeco = panelDeco.WithShadow(*p.config.decoration.Shadow)
	} else {
		panelDeco = panelDeco.WithShadow(style.ElevationShadow(cs, 2))
	}
	if p.config.decoration.Border != nil {
		panelDeco = panelDeco.WithBorder(*p.config.decoration.Border)
	}
	panel = ContainerDecoration(panelDeco, p.child)
	if p.config.width > 0 {
		panel = &fixedSizeWidget{
			width: p.config.width,
			child: panel,
		}
	}

	content := anchoredOverlayWidget(panel, gioLayout.Center)

	return Stack(
		mask,
		content,
	).Layout(ctx.Child(0))
}

type popupState struct {
	wasOpen bool
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
