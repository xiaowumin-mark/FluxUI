package widget

import (
	"image/color"
	"strings"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"
	theme "github.com/xiaowumin-mark/FluxUI/theme"

	"gioui.org/io/key"
	gioWidget "gioui.org/widget"
)

type InputOption func(*inputConfig)

type inputVariant int

const (
	inputVariantOutlined inputVariant = iota
	inputVariantFilled
)

type inputConfig struct {
	variant        inputVariant
	disabled       bool
	padding        style.Insets
	hasPadding     bool
	radius         float32
	border         color.NRGBA
	borderFocus    color.NRGBA
	background     color.NRGBA
	foreground     color.NRGBA
	placeholder    string
	hasBorder      bool
	hasBorderFocus bool
	hasBackground  bool
	hasForeground  bool
	textSize       float32
	maxLen         int
	password       bool
	singleLine     bool
	font           theme.FontSpec
	hasFamily      bool
	hasStyle       bool
	hasWeight      bool
	onChange       func(ctx *internal.Context, value string)
	onFocus        func(ctx *internal.Context, focused bool)
	ref            *InputRef
	decoration     style.Decoration
}

type inputWidget struct {
	value  string
	config inputConfig
}

type inputState struct {
	editor      *gioWidget.Editor
	initialized bool
	focused     bool
	syncedValue string
}

func inputStateFor(ctx *internal.Context) *inputState {
	value := ctx.Memo("input", func() any {
		return &inputState{
			editor: &gioWidget.Editor{
				SingleLine: true,
			},
		}
	})

	state, ok := value.(*inputState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: input state type mismatch")
	}
	return state
}

func TextField(value string, opts ...InputOption) Widget {
	return newTextField(inputVariantOutlined, value, opts...)
}

func OutlinedTextField(value string, opts ...InputOption) Widget {
	return newTextField(inputVariantOutlined, value, opts...)
}

func FilledTextField(value string, opts ...InputOption) Widget {
	return newTextField(inputVariantFilled, value, opts...)
}

func newTextField(variant inputVariant, value string, opts ...InputOption) Widget {
	cfg := inputConfig{
		variant:     variant,
		placeholder: "Enter text...",
		maxLen:      0,
		password:    false,
		singleLine:  true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &inputWidget{
		value:  value,
		config: cfg,
	}
}

func InputPlaceholder(text string) InputOption {
	return func(cfg *inputConfig) {
		cfg.placeholder = text
	}
}

func InputPadding(insets style.Insets) InputOption {
	return func(cfg *inputConfig) {
		cfg.padding = insets
		cfg.hasPadding = true
	}
}

func InputRadius(radius float32) InputOption {
	return func(cfg *inputConfig) {
		cfg.radius = radius
	}
}

func InputBorder(color color.NRGBA) InputOption {
	return func(cfg *inputConfig) {
		cfg.border = color
		cfg.hasBorder = true
	}
}

func InputBorderFocus(color color.NRGBA) InputOption {
	return func(cfg *inputConfig) {
		cfg.borderFocus = color
		cfg.hasBorderFocus = true
	}
}

func InputBackground(color color.NRGBA) InputOption {
	return func(cfg *inputConfig) {
		cfg.background = color
		cfg.hasBackground = true
	}
}

func InputForeground(color color.NRGBA) InputOption {
	return func(cfg *inputConfig) {
		cfg.foreground = color
		cfg.hasForeground = true
	}
}

func InputTextSize(size float32) InputOption {
	return func(cfg *inputConfig) {
		cfg.textSize = size
	}
}

func InputMaxLen(maxLen int) InputOption {
	return func(cfg *inputConfig) {
		cfg.maxLen = maxLen
	}
}

func InputPassword(password bool) InputOption {
	return func(cfg *inputConfig) {
		cfg.password = password
	}
}

func InputSingleLine(singleLine bool) InputOption {
	return func(cfg *inputConfig) {
		cfg.singleLine = singleLine
	}
}

// InputFontFamily 设置输入框字体族（局部覆盖）。
func InputFontFamily(family string) InputOption {
	return func(cfg *inputConfig) {
		cfg.font.Family = strings.TrimSpace(family)
		cfg.hasFamily = true
	}
}

func InputDisabled(disabled bool) InputOption {
	return func(cfg *inputConfig) {
		cfg.disabled = disabled
	}
}

func InputOnChange(fn func(ctx *internal.Context, value string)) InputOption {
	return func(cfg *inputConfig) {
		cfg.onChange = fn
	}
}

func InputOnFocus(fn func(ctx *internal.Context, focused bool)) InputOption {
	return func(cfg *inputConfig) {
		cfg.onFocus = fn
	}
}

// InputAttachRef 绑定命令型引用，用于外部主动操作输入框。
func InputAttachRef(ref *InputRef) InputOption {
	return func(cfg *inputConfig) {
		cfg.ref = ref
	}
}

// InputDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func InputDecoration(d style.Decoration) InputOption {
	return func(cfg *inputConfig) {
		cfg.decoration = d
	}
}

func (t *inputWidget) Layout(ctx *internal.Context) layout.Dimensions {
	state := inputStateFor(ctx)
	editor := state.editor
	if editor == nil {
		return layout.Dimensions{}
	}

	controlled := t.config.onChange != nil

	if !state.initialized {
		editor.SetText(t.value)
		state.syncedValue = t.value
		state.initialized = true
	} else if controlled && t.value != state.syncedValue {
		if shouldRecreateEditorForMemory(state.syncedValue, t.value) {
			state.editor = &gioWidget.Editor{
				SingleLine: t.config.singleLine,
			}
			editor = state.editor
			state.focused = false
		}
		editor.SetText(t.value)
		state.syncedValue = t.value
	}

	editor.SingleLine = t.config.singleLine
	editor.ReadOnly = t.config.disabled
	editor.MaxLen = t.config.maxLen
	if t.config.password {
		editor.Mask = '*'
	} else {
		editor.Mask = 0
	}

	if t.config.ref != nil {
		t.config.ref.bindInvalidator(redrawInvalidator(ctx))
		for _, cmd := range t.config.ref.drainCommands() {
			switch cmd.kind {
			case inputCmdSetText:
				editor.SetText(cmd.text)
				state.syncedValue = editor.Text()
				if t.config.onChange != nil {
					t.config.onChange(ctx, state.syncedValue)
				}
			case inputCmdAppend:
				editor.SetText(editor.Text() + cmd.text)
				state.syncedValue = editor.Text()
				if t.config.onChange != nil {
					t.config.onChange(ctx, state.syncedValue)
				}
			case inputCmdClear:
				editor.SetText("")
				state.syncedValue = ""
				if t.config.onChange != nil {
					t.config.onChange(ctx, "")
				}
			case inputCmdFocus:
				ctx.Gtx.Execute(key.FocusCmd{Tag: editor})
			case inputCmdBlur:
				ctx.Gtx.Execute(key.FocusCmd{Tag: nil})
			}
		}
	}

	for {
		ev, ok := editor.Update(ctx.Gtx)
		if !ok {
			break
		}
		if _, changed := ev.(gioWidget.ChangeEvent); changed && t.config.onChange != nil {
			text := editor.Text()
			if text != state.syncedValue {
				state.syncedValue = text
				t.config.onChange(ctx, text)
			}
		}
	}

	focused := ctx.Gtx.Focused(editor)
	if state.focused != focused {
		state.focused = focused
		if t.config.onFocus != nil {
			t.config.onFocus(ctx, focused)
		}
	}

	activeDecoration := t.config.decoration
	if t.config.disabled {
		activeDecoration = resolveDecorationState(activeDecoration, false, false, true)
	} else if focused && activeDecoration.Focused != nil {
		activeDecoration = activeDecoration.Merge(*activeDecoration.Focused)
	}
	duration, easing := md3InteractionTiming(false, false, focused, t.config.disabled)
	activeDecoration = md3AnimateDecoration(ctx, "input-decoration", stripStateDecoration(activeDecoration), duration, easing)

	th := ctx.Theme()
	cs := th.Colors
	defaults := resolveInputDefaults(t.config.variant, th)

	bg := activeDecoration.ResolveBg(defaults.background)
	if t.config.hasBackground {
		bg = activeDecoration.ResolveBg(t.config.background)
	}

	fg := t.config.foreground
	if !t.config.hasForeground {
		fg = defaults.foreground
	}
	if t.config.disabled {
		fg = style.DisabledContent(cs.OnSurface)
	}

	border := activeDecoration.ResolveBorder(style.Border{Width: defaults.borderWidth, Color: defaults.border}).Color
	if t.config.hasBorder {
		border = t.config.border
	}
	if focused && t.config.hasBorderFocus {
		border = t.config.borderFocus
	} else if focused {
		border = cs.Primary
	}
	if t.config.disabled {
		if t.config.decoration.Disabled == nil || t.config.decoration.Disabled.Border == nil {
			border = style.DisabledContent(cs.OnSurface)
		}
	}
	bg = md3AnimateColor(ctx, "input-bg", bg, duration, easing)
	fg = md3AnimateColor(ctx, "input-fg", fg, duration, easing)
	border = md3AnimateColor(ctx, "input-border", border, duration, easing)

	radiusDefault := defaults.radius
	if t.config.radius > 0 {
		radiusDefault = t.config.radius
	}
	radius := activeDecoration.ResolveRad(radiusDefault)
	paddingDefault := densityInsets(ctx, style.Symmetric(8, 12), style.Symmetric(4, 10))
	if t.config.hasPadding {
		paddingDefault = t.config.padding
	}
	padding := activeDecoration.ResolvePad(paddingDefault)

	font := ctx.Font()
	if t.config.hasFamily && strings.TrimSpace(t.config.font.Family) != "" {
		font.Family = strings.TrimSpace(t.config.font.Family)
	}
	if t.config.hasStyle {
		font.Style = t.config.font.Style
	}
	if t.config.hasWeight {
		font.Weight = t.config.font.Weight
	}
	font = font.Normalize()

	size := ctx.LayoutInput(editor, internal.InputSpec{
		Background:       bg,
		Foreground:       fg,
		PlaceholderColor: defaults.placeholder,
		Border:           border,
		Radius:           radius,
		Padding:          toInternalInsets(padding),
		TextSize:         firstPositive(t.config.textSize, defaults.text.Size),
		LineHeight:       defaults.text.LineHeight,
		Placeholder:      t.config.placeholder,
		Password:         t.config.password,
		MaxLen:           t.config.maxLen,
		SingleLine:       t.config.singleLine,
		Font:             font,
	})

	return layout.Dimensions{Size: size}
}

type inputDefaults struct {
	background  color.NRGBA
	foreground  color.NRGBA
	placeholder color.NRGBA
	border      color.NRGBA
	borderWidth float32
	radius      float32
	text        theme.TextStyle
}

func resolveInputDefaults(variant inputVariant, th *theme.Theme) inputDefaults {
	if th == nil {
		th = theme.Default()
	}
	cs := th.Colors
	defaults := inputDefaults{
		background:  cs.Surface,
		foreground:  cs.OnSurface,
		placeholder: cs.OnSurfaceVariant,
		border:      cs.Outline,
		borderWidth: 1,
		radius:      th.Shapes.ExtraSmall,
		text:        th.Types.BodyLarge,
	}
	if variant == inputVariantFilled {
		defaults.background = cs.SurfaceContainerHighest
		defaults.border = color.NRGBA{}
		defaults.borderWidth = 0
		defaults.radius = th.Shapes.Small
	}
	return defaults
}

func firstPositive(values ...float32) float32 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func shouldRecreateEditorForMemory(prev, next string) bool {
	const heavyTextBytes = 512 * 1024
	if len(prev) < heavyTextBytes {
		return false
	}
	if len(next) == 0 {
		return true
	}
	return len(next) <= len(prev)/8
}
