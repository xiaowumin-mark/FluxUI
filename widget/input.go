package widget

import (
	"image/color"
	"strings"

	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"
	"fluxui/theme"

	"gioui.org/io/key"
	gioWidget "gioui.org/widget"
)

type InputOption func(*inputConfig)

type inputConfig struct {
	disabled       bool
	padding        style.Insets
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
	hasValue       bool
	value          string
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
		panic("fluxui/widget: input state type mismatch")
	}
	return state
}

func TextField(value string, opts ...InputOption) Widget {
	cfg := inputConfig{
		padding:     style.Symmetric(8, 12),
		radius:      8,
		border:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		borderFocus: color.NRGBA{R: 66, G: 133, B: 244, A: 255},
		background:  color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		foreground:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		placeholder: "Enter text...",
		textSize:    16,
		maxLen:      0,
		password:    false,
		singleLine:  true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.hasValue {
		value = cfg.value
	}

	return &inputWidget{
		value:  value,
		config: cfg,
	}
}

func InputValue(value string) InputOption {
	return func(cfg *inputConfig) {
		cfg.value = value
		cfg.hasValue = true
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

// InputFont 设置输入框字体（局部覆盖）。
func InputFont(spec theme.FontSpec) InputOption {
	return func(cfg *inputConfig) {
		cfg.font = spec
		cfg.hasStyle = true
		cfg.hasWeight = true
		if strings.TrimSpace(spec.Family) != "" {
			cfg.hasFamily = true
		}
	}
}

// InputFontFamily 设置输入框字体族（局部覆盖）。
func InputFontFamily(family string) InputOption {
	return func(cfg *inputConfig) {
		cfg.font.Family = strings.TrimSpace(family)
		cfg.hasFamily = true
	}
}

// InputFontStyle 设置输入框字体样式（局部覆盖）。
func InputFontStyle(style theme.FontStyle) InputOption {
	return func(cfg *inputConfig) {
		cfg.font.Style = style
		cfg.hasStyle = true
	}
}

// InputFontWeight 设置输入框字体字重（局部覆盖）。
func InputFontWeight(weight theme.FontWeight) InputOption {
	return func(cfg *inputConfig) {
		cfg.font.Weight = weight
		cfg.hasWeight = true
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
		t.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
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

	bg := t.config.background
	if !t.config.hasBackground {
		bg = ctx.Theme().Surface
	}

	fg := t.config.foreground
	if !t.config.hasForeground {
		fg = ctx.Theme().TextColor
	}
	if t.config.disabled {
		fg = ctx.Theme().Disabled
	}

	border := t.config.border
	if focused && t.config.hasBorderFocus {
		border = t.config.borderFocus
	}
	if t.config.disabled {
		border = ctx.Theme().Disabled
	}

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
		Background:  bg,
		Foreground:  fg,
		Border:      border,
		Radius:      t.config.radius,
		Padding:     toInternalInsets(t.config.padding),
		TextSize:    t.config.textSize,
		Placeholder: t.config.placeholder,
		Password:    t.config.password,
		MaxLen:      t.config.maxLen,
		SingleLine:  t.config.singleLine,
		Font:        font,
	})

	return layout.Dimensions{Size: size}
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
