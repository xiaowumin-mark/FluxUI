package widget

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"
	theme "github.com/xiaowumin-mark/FluxUI/theme"

	gioFont "gioui.org/font"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	gioWidget "gioui.org/widget"
	"gioui.org/widget/material"
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
	label          string
	supportingText string
	errorText      string
	prefixText     string
	suffixText     string
	leading        Widget
	trailing       Widget
	error          bool
	required       bool
	noAsterisk     bool
	counter        bool
	rows           int
	minRows        int
	maxRows        int
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
	onBeforeInput  fluxevent.InputHandler
	onInputEvent   fluxevent.InputHandler
	onSubmit       fluxevent.InputHandler
	ref            *InputRef
	decoration     style.Decoration
}

type inputWidget struct {
	value  string
	config inputConfig
}

const (
	inputHistoryMaxEntries = 64
	inputHistoryMaxBytes   = 256 << 10
)

type inputState struct {
	editor         *gioWidget.Editor
	initialized    bool
	focused        bool
	syncedValue    string
	blurTag        any
	fieldSize      image.Point
	valueHistory   []string
	historyIndex   int
	historyBytes   int
	historyEnabled bool
}

func inputStateFor(ctx *internal.Context) *inputState {
	key := ctx.NextMemoryKey("input")
	value := ctx.PersistentKey(key, func() any {
		return &inputState{
			editor: &gioWidget.Editor{
				SingleLine: true,
			},
			blurTag: new(int),
		}
	})

	state, ok := value.(*inputState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: input state type mismatch")
	}
	if runtime := ctx.Runtime(); runtime != nil {
		runtime.UseEffectKey(key, true, []any{state}, func() func() {
			return state.dispose
		})
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

func InputLabel(text string) InputOption {
	return func(cfg *inputConfig) {
		cfg.label = text
	}
}

func InputSupportingText(text string) InputOption {
	return func(cfg *inputConfig) {
		cfg.supportingText = text
	}
}

func InputErrorText(text string) InputOption {
	return func(cfg *inputConfig) {
		cfg.errorText = text
	}
}

func InputError(error bool) InputOption {
	return func(cfg *inputConfig) {
		cfg.error = error
	}
}

func InputRequired(required bool) InputOption {
	return func(cfg *inputConfig) {
		cfg.required = required
	}
}

func InputNoAsterisk(noAsterisk bool) InputOption {
	return func(cfg *inputConfig) {
		cfg.noAsterisk = noAsterisk
	}
}

func InputPrefixText(text string) InputOption {
	return func(cfg *inputConfig) {
		cfg.prefixText = text
	}
}

func InputSuffixText(text string) InputOption {
	return func(cfg *inputConfig) {
		cfg.suffixText = text
	}
}

func InputLeading(leading Widget) InputOption {
	return func(cfg *inputConfig) {
		cfg.leading = leading
	}
}

func InputTrailing(trailing Widget) InputOption {
	return func(cfg *inputConfig) {
		cfg.trailing = trailing
	}
}

func InputCounter(visible bool) InputOption {
	return func(cfg *inputConfig) {
		cfg.counter = visible
	}
}

func InputRows(rows int) InputOption {
	return func(cfg *inputConfig) {
		cfg.rows = rows
		if rows > 1 {
			cfg.singleLine = false
		}
	}
}

func InputMinRows(rows int) InputOption {
	return func(cfg *inputConfig) {
		cfg.minRows = rows
		if rows > 1 {
			cfg.singleLine = false
		}
	}
}

func InputMaxRows(rows int) InputOption {
	return func(cfg *inputConfig) {
		cfg.maxRows = rows
		if rows > 1 {
			cfg.singleLine = false
		}
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

func InputOnBeforeInput(fn fluxevent.InputHandler) InputOption {
	return func(cfg *inputConfig) {
		cfg.onBeforeInput = fn
	}
}

func InputOnInputEvent(fn fluxevent.InputHandler) InputOption {
	return func(cfg *inputConfig) {
		cfg.onInputEvent = fn
	}
}

func InputOnSubmit(fn fluxevent.InputHandler) InputOption {
	return func(cfg *inputConfig) {
		cfg.onSubmit = fn
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

	t.registerInputEventListeners(ctx)

	controlled := t.config.onChange != nil
	historyReset := false

	if !state.initialized {
		editor.SetText(t.value)
		state.syncedValue = t.value
		state.initialized = true
		historyReset = true
	} else if controlled && t.value != state.syncedValue {
		if t.config.password {
			editor = t.rebuildPasswordEditor(ctx, state, editor, t.value, false)
		} else if shouldRecreateEditorForMemory(state.syncedValue, t.value) {
			state.editor = &gioWidget.Editor{
				SingleLine: t.config.singleLine,
			}
			editor = state.editor
			state.focused = false
		}
		if !t.config.password {
			editor.SetText(t.value)
		}
		state.syncedValue = t.value
		historyReset = true
	}
	state.configureInputHistory(!t.config.password, state.syncedValue, historyReset)

	editor.SingleLine = t.config.singleLine
	editor.ReadOnly = t.config.disabled
	editor.Submit = t.config.onSubmit != nil
	editor.MaxLen = t.config.maxLen
	if t.config.password {
		editor.Mask = '*'
	} else {
		editor.Mask = 0
	}
	t.constrainEditorHistory(editor)

	if t.config.ref != nil {
		bindCommandRef(ctx, "input", t.config.ref, &t.config.ref.queue)
		for _, cmd := range t.config.ref.drainCommands() {
			switch cmd.kind {
			case inputCmdSetText:
				editor = t.commitProgrammaticInput(ctx, state, editor, cmd.text, cmd.text, fluxevent.InputTypeProgrammaticSetText)
			case inputCmdAppend:
				editor = t.commitProgrammaticInput(ctx, state, editor, editor.Text()+cmd.text, cmd.text, fluxevent.InputTypeProgrammaticAppend)
			case inputCmdClear:
				editor = t.commitProgrammaticInput(ctx, state, editor, "", "", fluxevent.InputTypeProgrammaticClear)
			case inputCmdFocus:
				ctx.Gtx.Execute(key.FocusCmd{Tag: editor})
			case inputCmdBlur:
				ctx.Gtx.Execute(key.FocusCmd{Tag: nil})
			}
		}
	}
	t.constrainEditorHistory(editor)
	if t.config.password {
		t.discardPasswordUndoRedo(ctx, editor)
	}

	passwordEditorDirty := false
	for {
		ev, ok := editor.Update(ctx.Gtx)
		if !ok {
			break
		}
		switch ev := ev.(type) {
		case gioWidget.ChangeEvent:
			text := editor.Text()
			if text != state.syncedValue {
				passwordEditorDirty = passwordEditorDirty || t.config.password
				editor = t.commitUserInput(ctx, state, editor, text, ev)
			}
		case gioWidget.SubmitEvent:
			t.dispatchSubmit(ctx, state, ev.Text, ev)
		}
		t.constrainEditorHistory(editor)
	}
	if passwordEditorDirty {
		editor = t.rebuildPasswordEditor(ctx, state, editor, state.syncedValue, true)
	}

	focused := ctx.Gtx.Focused(editor)
	if focused {
		t.handleOutsidePressBlur(ctx, state, editor)
		focused = ctx.Gtx.Focused(editor)
	}
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
	duration, easing := md3InteractionTiming(ctx, false, false, focused, t.config.disabled)
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
	borderWidth := activeDecoration.ResolveBorder(style.Border{Width: defaults.borderWidth, Color: defaults.border}).Width
	if t.config.hasBorder {
		border = t.config.border
	}
	if focused && t.config.hasBorderFocus {
		border = t.config.borderFocus
	} else if focused {
		border = cs.Primary
	}
	if t.config.error {
		border = cs.Error
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
	paddingDefault := densityInsets(ctx, style.Symmetric(0, 16), style.Symmetric(0, 12))
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

	labelColor := defaults.placeholder
	if focused {
		labelColor = cs.Primary
	}
	if t.config.error {
		labelColor = cs.Error
	}
	if t.config.disabled {
		labelColor = style.DisabledContent(cs.OnSurface)
	}
	labelColor = md3AnimateColor(ctx, "input-label", labelColor, duration, easing)

	size := t.layoutMD3TextField(ctx, editor, inputRenderSpec{
		background:       bg,
		foreground:       fg,
		placeholderColor: defaults.placeholder,
		labelColor:       labelColor,
		border:           border,
		borderWidth:      borderWidth,
		radius:           radius,
		padding:          padding,
		textSize:         firstPositive(t.config.textSize, defaults.text.Size),
		lineHeight:       defaults.text.LineHeight,
		font:             font,
		focused:          focused,
	})
	state.fieldSize = size

	return layout.Dimensions{Size: size}
}

func (t *inputWidget) constrainEditorHistory(editor *gioWidget.Editor) {
	maxEntries := inputHistoryMaxEntries
	maxBytes := inputHistoryMaxBytes
	if t.config.password {
		maxEntries = 0
		maxBytes = 0
	}
	if !limitEditorUndoHistory(editor, maxEntries, maxBytes) {
		panic("FluxUI/widget: incompatible Gio Editor undo history layout")
	}
}

func (t *inputWidget) discardPasswordUndoRedo(ctx *internal.Context, editor *gioWidget.Editor) {
	if ctx == nil || editor == nil {
		return
	}
	for {
		_, ok := ctx.Gtx.Event(key.Filter{
			Focus:    editor,
			Name:     "Z",
			Required: key.ModShortcut,
			Optional: key.ModShift,
		})
		if !ok {
			return
		}
	}
}

func (t *inputWidget) rebuildPasswordEditor(ctx *internal.Context, state *inputState, editor *gioWidget.Editor, value string, preserveSelection bool) *gioWidget.Editor {
	if state == nil || editor == nil || !t.config.password {
		return editor
	}

	selectionStart, selectionEnd := 0, 0
	if preserveSelection {
		selectionStart, selectionEnd = editor.Selection()
	}
	focused := ctx != nil && ctx.Gtx.Focused(editor)
	next := &gioWidget.Editor{
		Alignment:       editor.Alignment,
		LineHeight:      editor.LineHeight,
		LineHeightScale: editor.LineHeightScale,
		SingleLine:      editor.SingleLine,
		ReadOnly:        editor.ReadOnly,
		Submit:          editor.Submit,
		Mask:            editor.Mask,
		InputHint:       editor.InputHint,
		MaxLen:          editor.MaxLen,
		Filter:          editor.Filter,
		WrapPolicy:      editor.WrapPolicy,
	}
	next.SetText(value)
	if preserveSelection {
		next.SetCaret(selectionStart, selectionEnd)
	}
	t.constrainEditorHistory(next)

	// Gio keeps deleted text in several private backing buffers. Detach all of
	// them immediately; the router may retain the old tag until the next frame.
	// Replacing the editor also terminates any in-progress IME composition, which
	// is preferable to retaining password fragments in the composition buffers.
	*editor = gioWidget.Editor{}
	state.editor = next
	if focused {
		ctx.Gtx.Execute(key.FocusCmd{Tag: next})
	}
	return next
}

func (t *inputWidget) registerInputEventListeners(ctx *internal.Context) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	if t.config.onBeforeInput != nil {
		fluxevent.OnInput(ctx, fluxevent.BeforeInput, t.config.onBeforeInput)
	}
	if t.config.onInputEvent != nil {
		fluxevent.OnInput(ctx, fluxevent.Input, t.config.onInputEvent)
	}
	if t.config.onSubmit != nil {
		fluxevent.OnInput(ctx, fluxevent.Submit, t.config.onSubmit)
	}
}

func (t *inputWidget) commitProgrammaticInput(ctx *internal.Context, state *inputState, editor *gioWidget.Editor, next string, data string, inputType string) *gioWidget.Editor {
	if state == nil || editor == nil {
		return editor
	}
	previous := state.syncedValue
	if next == previous {
		return editor
	}
	before := t.newInputEvent(ctx, fluxevent.BeforeInput, previous, next, data, inputType, fluxevent.InputSourceProgrammatic, false, nil, false)
	if !t.dispatchInputEvent(ctx, before) {
		return editor
	}
	if t.config.password {
		editor = t.rebuildPasswordEditor(ctx, state, editor, next, false)
	} else {
		editor.SetText(next)
		t.constrainEditorHistory(editor)
	}
	value := editor.Text()
	if value == previous {
		return editor
	}
	if value != next {
		_, data = inputTextDiff(previous, value)
	}
	t.finishInputMutation(ctx, state, previous, value, data, inputType, fluxevent.InputSourceProgrammatic, false, nil)
	return editor
}

func (t *inputWidget) commitUserInput(ctx *internal.Context, state *inputState, editor *gioWidget.Editor, value string, native any) *gioWidget.Editor {
	if state == nil || editor == nil {
		return editor
	}
	// Editor.Update has already applied this mutation. Drop password history
	// before invoking callbacks so re-entrant code cannot observe or undo it.
	t.constrainEditorHistory(editor)
	previous := state.syncedValue
	if value == previous {
		return editor
	}
	mutation := state.classifyUserInput(previous, value)
	before := t.newInputEvent(ctx, fluxevent.BeforeInput, previous, value, mutation.data, mutation.inputType, mutation.source, true, native, true)
	if !t.dispatchInputEvent(ctx, before) {
		editor.SetText(previous)
		t.constrainEditorHistory(editor)
		state.syncedValue = previous
		return editor
	}
	t.finishInputMutation(ctx, state, previous, value, mutation.data, mutation.inputType, mutation.source, true, native)
	return editor
}

func (t *inputWidget) finishInputMutation(ctx *internal.Context, state *inputState, previous string, value string, data string, inputType string, source fluxevent.InputSource, bestEffort bool, native any) {
	if value == previous {
		return
	}
	state.syncedValue = value
	state.recordInputHistory(value, source)

	input := t.newInputEvent(ctx, fluxevent.Input, previous, value, data, inputType, source, source != fluxevent.InputSourceProgrammatic, native, bestEffort)
	t.dispatchInputEvent(ctx, input)
	change := t.newInputEvent(ctx, fluxevent.Change, previous, value, data, inputType, source, source != fluxevent.InputSourceProgrammatic, native, bestEffort)
	t.dispatchInputEvent(ctx, change)
	if t.config.onChange != nil {
		t.config.onChange(ctx, value)
	}
}

func (t *inputWidget) dispatchSubmit(ctx *internal.Context, state *inputState, value string, native any) {
	previous := ""
	if state != nil {
		previous = state.syncedValue
	}
	ev := t.newInputEvent(ctx, fluxevent.Submit, previous, value, "", fluxevent.InputTypeInsertLineBreak, fluxevent.InputSourceUser, true, native, false)
	t.dispatchInputEvent(ctx, ev)
}

func (t *inputWidget) newInputEvent(ctx *internal.Context, eventType fluxevent.Type, previous string, value string, data string, inputType string, source fluxevent.InputSource, trusted bool, native any, bestEffort bool) *fluxevent.InputEvent {
	target := fluxevent.TargetID(0)
	if ctx != nil {
		target = ctx.PathID()
	}
	return &fluxevent.InputEvent{
		Event: fluxevent.Event{
			Type:    eventType,
			Target:  target,
			Trusted: trusted,
		},
		Data:          data,
		InputType:     inputType,
		Source:        source,
		Value:         value,
		PreviousValue: previous,
		BestEffort:    bestEffort,
		Native:        native,
	}
}

func (t *inputWidget) dispatchInputEvent(ctx *internal.Context, ev *fluxevent.InputEvent) bool {
	if ev == nil {
		return true
	}
	target := fluxevent.TargetID(0)
	if ctx != nil {
		target = ctx.PathID()
	}
	allowed := fluxevent.DispatchInputEvent(ctx, target, ev)
	if ctx != nil && ctx.Runtime() != nil {
		return allowed
	}
	switch ev.Type {
	case fluxevent.BeforeInput:
		if t.config.onBeforeInput != nil {
			t.config.onBeforeInput(ctx, ev)
		}
	case fluxevent.Input:
		if t.config.onInputEvent != nil {
			t.config.onInputEvent(ctx, ev)
		}
	case fluxevent.Submit:
		if t.config.onSubmit != nil {
			t.config.onSubmit(ctx, ev)
		}
	}
	return !(ev.Cancelable && ev.DefaultPrevented)
}

type inputMutation struct {
	data      string
	inputType string
	source    fluxevent.InputSource
}

func (s *inputState) classifyUserInput(previous, value string) inputMutation {
	removed, inserted := inputTextDiff(previous, value)
	mutation := inputMutation{
		data:      inserted,
		inputType: inputTypeForDiff(removed, inserted),
		source:    fluxevent.InputSourceUser,
	}
	switch {
	case s != nil && s.historyIndex > 0 && s.historyIndex <= len(s.valueHistory)-1 && s.valueHistory[s.historyIndex-1] == value:
		mutation.source = fluxevent.InputSourceUndo
		mutation.inputType = fluxevent.InputTypeHistoryUndo
	case s != nil && s.historyIndex >= 0 && s.historyIndex+1 < len(s.valueHistory) && s.valueHistory[s.historyIndex+1] == value:
		mutation.source = fluxevent.InputSourceRedo
		mutation.inputType = fluxevent.InputTypeHistoryRedo
	case removed != "" && inserted == "":
		mutation.source = fluxevent.InputSourceDelete
	case inserted != "" && likelyPasteData(inserted):
		mutation.source = fluxevent.InputSourcePaste
		mutation.inputType = fluxevent.InputTypeInsertFromPaste
	}
	return mutation
}

func (s *inputState) resetInputHistory(value string) {
	if s == nil {
		return
	}
	s.clearInputHistory()
	if len(value) > inputHistoryMaxBytes {
		return
	}
	s.valueHistory = append(s.valueHistory, value)
	s.historyIndex = 0
	s.historyBytes = len(value)
}

func (s *inputState) recordInputHistory(value string, source fluxevent.InputSource) {
	if s == nil || !s.historyEnabled {
		return
	}
	if len(value) > inputHistoryMaxBytes {
		s.clearInputHistory()
		return
	}
	if len(s.valueHistory) == 0 {
		s.resetInputHistory(value)
		return
	}
	switch source {
	case fluxevent.InputSourceUndo:
		if s.historyIndex > 0 && s.valueHistory[s.historyIndex-1] == value {
			s.historyIndex--
			return
		}
	case fluxevent.InputSourceRedo:
		if s.historyIndex+1 < len(s.valueHistory) && s.valueHistory[s.historyIndex+1] == value {
			s.historyIndex++
			return
		}
	}
	if s.historyIndex >= 0 && s.historyIndex < len(s.valueHistory) && s.valueHistory[s.historyIndex] == value {
		return
	}
	if s.historyIndex+1 < len(s.valueHistory) {
		for i := s.historyIndex + 1; i < len(s.valueHistory); i++ {
			s.historyBytes -= len(s.valueHistory[i])
			s.valueHistory[i] = ""
		}
		s.valueHistory = s.valueHistory[:s.historyIndex+1]
	}
	s.valueHistory = append(s.valueHistory, value)
	s.historyBytes += len(value)
	s.historyIndex = len(s.valueHistory) - 1
	s.trimInputHistory()
}

func (s *inputState) configureInputHistory(enabled bool, value string, reset bool) {
	if s == nil {
		return
	}
	if !enabled {
		s.historyEnabled = false
		s.clearInputHistory()
		return
	}
	if !s.historyEnabled || reset {
		s.historyEnabled = true
		s.resetInputHistory(value)
	}
}

func (s *inputState) trimInputHistory() {
	if s == nil {
		return
	}
	drop := 0
	for len(s.valueHistory)-drop > inputHistoryMaxEntries || s.historyBytes > inputHistoryMaxBytes {
		s.historyBytes -= len(s.valueHistory[drop])
		s.valueHistory[drop] = ""
		drop++
	}
	if drop == 0 {
		return
	}
	copy(s.valueHistory, s.valueHistory[drop:])
	newLen := len(s.valueHistory) - drop
	clear(s.valueHistory[newLen:])
	s.valueHistory = s.valueHistory[:newLen]
	s.historyIndex -= drop
	if s.historyIndex < 0 && len(s.valueHistory) > 0 {
		s.historyIndex = 0
	}
}

func (s *inputState) clearInputHistory() {
	if s == nil {
		return
	}
	clear(s.valueHistory)
	s.valueHistory = nil
	s.historyIndex = -1
	s.historyBytes = 0
}

func (s *inputState) dispose() {
	if s == nil {
		return
	}
	s.clearInputHistory()
	s.historyEnabled = false
	s.syncedValue = ""
	if s.editor != nil {
		*s.editor = gioWidget.Editor{}
	}
	s.editor = nil
	s.blurTag = nil
	s.fieldSize = image.Point{}
	s.focused = false
	s.initialized = false
}

func inputTypeForDiff(removed, inserted string) string {
	switch {
	case removed != "" && inserted == "":
		return fluxevent.InputTypeDeleteContentBackward
	case removed != "" && inserted != "":
		return fluxevent.InputTypeInsertReplacement
	case inserted != "":
		return fluxevent.InputTypeInsertText
	default:
		return fluxevent.InputTypeInsertText
	}
}

func inputTextDiff(previous, value string) (removed string, inserted string) {
	if previous == value {
		return "", ""
	}
	prevRunes := []rune(previous)
	nextRunes := []rune(value)
	start := 0
	for start < len(prevRunes) && start < len(nextRunes) && prevRunes[start] == nextRunes[start] {
		start++
	}
	end := 0
	for start+end < len(prevRunes) && start+end < len(nextRunes) &&
		prevRunes[len(prevRunes)-1-end] == nextRunes[len(nextRunes)-1-end] {
		end++
	}
	return string(prevRunes[start : len(prevRunes)-end]), string(nextRunes[start : len(nextRunes)-end])
}

func likelyPasteData(data string) bool {
	if strings.ContainsAny(data, "\n\r\t") {
		return true
	}
	return len([]rune(data)) > 1
}

func (t *inputWidget) handleOutsidePressBlur(ctx *internal.Context, state *inputState, editor *gioWidget.Editor) {
	if ctx == nil || state == nil || state.blurTag == nil || editor == nil {
		return
	}
	viewport, ok := ctx.Viewport()
	if !ok || viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		viewport = image.Rectangle{Max: ctx.Gtx.Constraints.Max}
	}
	if viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		return
	}
	position := ctx.Position()
	pass := pointer.PassOp{}.Push(ctx.Gtx.Ops)
	offset := op.Offset(image.Pt(-position.X, -position.Y)).Push(ctx.Gtx.Ops)
	clipStack := clip.Rect(viewport).Push(ctx.Gtx.Ops)
	gioEvent.Op(ctx.Gtx.Ops, state.blurTag)
	clipStack.Pop()
	offset.Pop()
	pass.Pop()

	fieldRect := image.Rectangle{
		Min: position,
		Max: position.Add(state.fieldSize),
	}
	if fieldRect.Dx() <= 0 || fieldRect.Dy() <= 0 {
		fieldRect.Max = position.Add(ctx.Gtx.Constraints.Min)
	}
	for {
		ev, ok := ctx.Gtx.Event(pointer.Filter{
			Target: state.blurTag,
			Kinds:  pointer.Press,
		})
		if !ok {
			break
		}
		if pe, ok := ev.(pointer.Event); ok {
			p := image.Pt(int(pe.Position.X+0.5), int(pe.Position.Y+0.5))
			if !p.In(fieldRect) {
				ctx.Gtx.Execute(key.FocusCmd{Tag: nil})
			}
		}
	}
}

type inputRenderSpec struct {
	background       color.NRGBA
	foreground       color.NRGBA
	placeholderColor color.NRGBA
	labelColor       color.NRGBA
	border           color.NRGBA
	borderWidth      float32
	radius           float32
	padding          style.Insets
	textSize         float32
	lineHeight       float32
	font             theme.FontSpec
	focused          bool
}

func (t *inputWidget) layoutMD3TextField(ctx *internal.Context, editor *gioWidget.Editor, spec inputRenderSpec) image.Point {
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordFrameSection(internal.PerfLayout, 1)
	}
	gtx := ctx.Gtx
	fieldHeightDp := densityHeight(ctx, 56, 48)
	rows := t.config.rows
	if rows <= 0 {
		rows = t.config.minRows
	}
	if rows <= 0 {
		rows = 1
	}
	if t.config.singleLine {
		rows = 1
	}
	lineHeight := firstPositive(spec.lineHeight, spec.textSize*1.5, 24)
	if !t.config.singleLine && rows < 2 {
		rows = 2
	}
	if t.config.maxRows > 0 && rows > t.config.maxRows {
		rows = t.config.maxRows
	}
	if !t.config.singleLine {
		fieldHeightDp = 32 + lineHeight*float32(rows)
		if fieldHeightDp < densityHeight(ctx, 56, 48) {
			fieldHeightDp = densityHeight(ctx, 56, 48)
		}
	}

	fieldHeight := gtx.Dp(safeDp(fieldHeightDp))
	width := gtx.Constraints.Min.X
	if width <= 0 && gtx.Constraints.Max.X > 0 && gtx.Constraints.Max.X < 1_000_000 {
		width = gtx.Constraints.Max.X
	}
	if width <= 0 {
		width = gtx.Dp(safeDp(280))
	}
	size := gtx.Constraints.Constrain(image.Pt(width, fieldHeight))
	if size.Y < fieldHeight && gtx.Constraints.Max.Y >= fieldHeight {
		size.Y = fieldHeight
	}

	hasValue := editor.Text() != ""
	hasLabel := strings.TrimSpace(t.config.label) != ""
	floatTarget := float32(0)
	if hasLabel && (hasValue || spec.focused) {
		floatTarget = 1
	}
	labelProgress := t.md3TextFieldLabelProgress(ctx, floatTarget, hasValue)
	if !hasLabel {
		labelProgress = 0
	}

	contentLeft := gtx.Dp(safeDp(spec.padding.Left))
	contentRight := gtx.Dp(safeDp(spec.padding.Right))
	if contentLeft <= 0 {
		contentLeft = gtx.Dp(safeDp(16))
	}
	if contentRight <= 0 {
		contentRight = gtx.Dp(safeDp(16))
	}
	if t.config.leading != nil {
		contentLeft += gtx.Dp(safeDp(32))
	}
	if t.config.trailing != nil {
		contentRight += gtx.Dp(safeDp(32))
	}

	label := t.displayLabel()
	t.drawInputContainer(ctx, size, spec, labelProgress, contentLeft, label)

	labelLineHeight := spec.lineHeight
	if labelLineHeight <= 0 {
		labelLineHeight = spec.textSize * 1.25
	}
	if labelLineHeight < spec.textSize {
		labelLineHeight = spec.textSize
	}
	labelRestY := (float32(size.Y) - labelLineHeight) / 2
	labelFloatY := float32(gtx.Dp(safeDp(8)))
	editorTopRest := gtx.Dp(safeDp(8))
	editorTopFloat := gtx.Dp(safeDp(8))
	if hasLabel {
		switch t.config.variant {
		case inputVariantFilled:
			labelRestY = (float32(size.Y) - labelLineHeight) / 2
			labelFloatY = float32(gtx.Dp(safeDp(7)))
			editorTopRest = gtx.Dp(safeDp(18))
			editorTopFloat = gtx.Dp(safeDp(24))
		default:
			labelRestY = (float32(size.Y) - labelLineHeight) / 2
			labelFloatY = -float32(gtx.Dp(safeDp(8)))
			editorTopRest = gtx.Dp(safeDp(8))
			editorTopFloat = gtx.Dp(safeDp(8))
		}
	}
	editorTop := editorTopRest + int(float32(editorTopFloat-editorTopRest)*labelProgress+0.5)
	if !t.config.singleLine {
		editorTop = gtx.Dp(safeDp(22))
		if !hasLabel {
			editorTop = gtx.Dp(safeDp(12))
		}
	}
	editorBottom := gtx.Dp(safeDp(6))
	editorHeight := size.Y - editorTop - editorBottom
	if editorHeight < gtx.Dp(safeDp(24)) {
		editorHeight = gtx.Dp(safeDp(24))
	}
	editorWidth := size.X - contentLeft - contentRight

	prefixWidth := 0
	suffixWidth := 0
	if strings.TrimSpace(t.config.prefixText) != "" {
		prefixWidth = t.measureAffix(ctx, t.config.prefixText, spec)
	}
	if strings.TrimSpace(t.config.suffixText) != "" {
		suffixWidth = t.measureAffix(ctx, t.config.suffixText, spec)
	}
	editorWidth -= prefixWidth + suffixWidth
	if editorWidth < 1 {
		editorWidth = 1
	}

	t.layoutInputSlots(ctx, size, spec)
	contentOpacity := t.contentOpacity(hasLabel, labelProgress)
	if prefixWidth > 0 {
		t.drawAffix(ctx, image.Pt(contentLeft, editorTop+(editorHeight-gtx.Dp(safeDp(24)))/2), t.config.prefixText, spec, contentOpacity)
	}
	editorX := contentLeft + prefixWidth
	t.layoutEditor(ctx, editor, image.Rect(editorX, editorTop, editorX+editorWidth, editorTop+editorHeight), spec, labelProgress, contentOpacity)
	if suffixWidth > 0 {
		t.drawAffix(ctx, image.Pt(editorX+editorWidth, editorTop+(editorHeight-gtx.Dp(safeDp(24)))/2), t.config.suffixText, spec, contentOpacity)
	}
	if hasLabel {
		t.drawFloatingLabel(ctx, label, contentLeft, labelRestY, labelFloatY, labelProgress, spec)
	}

	total := size
	supporting := t.supportingLineText()
	counter := t.counterText(editor.Text())
	if supporting != "" || counter != "" {
		rowHeight := t.layoutSupportingRow(ctx, size.X, size.Y, supporting, counter, spec)
		total.Y += rowHeight
	}
	return total
}

func (t *inputWidget) md3TextFieldLabelProgress(ctx *internal.Context, target float32, populated bool) float32 {
	if ctx == nil {
		return target
	}
	initial := float32(0)
	if populated && target > 0 {
		initial = target
	}
	state := md3FloatStateFor(ctx, "input-floating-label", initial)
	return state.advance(ctx, target, 180*time.Millisecond, style.InteractionEmphasizedDecelerateEasing)
}

func (t *inputWidget) contentOpacity(hasLabel bool, labelProgress float32) float32 {
	if !hasLabel {
		return 1
	}
	if labelProgress <= 0.45 {
		return 0
	}
	return clampFloat32((labelProgress-0.45)/0.55, 0, 1)
}

func (t *inputWidget) drawInputContainer(ctx *internal.Context, size image.Point, spec inputRenderSpec, labelProgress float32, labelLeft int, label string) {
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordFrameSection(internal.PerfDraw, 1)
	}
	gtx := ctx.Gtx
	radius := clampRRectRadiusPx(size, gtx.Dp(safeDp(spec.radius)))
	rect := image.Rectangle{Max: size}
	if spec.background.A > 0 {
		paint.FillShape(gtx.Ops, spec.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if t.config.variant == inputVariantFilled {
		lineColor := spec.border
		if lineColor.A == 0 {
			lineColor = ctx.Theme().Colors.OnSurfaceVariant
		}
		h := gtx.Dp(safeDp(1))
		if spec.focused || t.config.error {
			h = gtx.Dp(safeDp(2))
		}
		paint.FillShape(gtx.Ops, lineColor, clip.Rect(image.Rect(0, size.Y-h, size.X, size.Y)).Op())
		return
	}
	width := spec.borderWidth
	if spec.focused || t.config.error {
		width = 2
	}
	if width <= 0 || spec.border.A == 0 {
		return
	}
	bw := gtx.Dp(safeDp(width))
	if bw <= 0 {
		bw = 1
	}
	strokeRect := rect
	half := (bw + 1) / 2
	strokeRect.Min = strokeRect.Min.Add(image.Pt(half, half))
	strokeRect.Max = strokeRect.Max.Sub(image.Pt(half, half))
	if strokeRect.Dx() <= 0 || strokeRect.Dy() <= 0 {
		return
	}
	paint.FillShape(gtx.Ops, spec.border, clip.Stroke{
		Path:  clip.UniformRRect(strokeRect, clampRRectRadiusPx(strokeRect.Size(), radius)).Path(gtx.Ops),
		Width: float32(bw),
	}.Op())
	if strings.TrimSpace(label) != "" && labelProgress > 0.05 {
		labelWidth := t.labelWidth(ctx, label, spec, 0.75)
		pad := gtx.Dp(safeDp(4))
		x := labelLeft - pad
		if x < 0 {
			x = 0
		}
		notch := image.Rect(x, 0, x+labelWidth+pad*2, bw+1)
		paint.FillShape(gtx.Ops, spec.background, clip.Rect(notch).Op())
	}
}

func (t *inputWidget) layoutEditor(ctx *internal.Context, editor *gioWidget.Editor, rect image.Rectangle, spec inputRenderSpec, labelProgress float32, opacity float32) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return
	}
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordFrameSection(internal.PerfInput, 1)
	}
	placeholder := t.config.placeholder
	if strings.TrimSpace(t.config.label) != "" && labelProgress < 0.99 {
		placeholder = ""
	}
	gtx := ctx.Gtx
	editorCtx := gtx
	editorCtx.Constraints = gioLayout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	var opacityStack paint.OpacityStack
	if opacity < 1 {
		opacityStack = paint.PushOpacity(gtx.Ops, opacity)
	}
	ed := material.Editor(ctx.MaterialTheme(), editor, placeholder)
	ed.Font = gioFont.Font{
		Typeface: gioFont.Typeface(spec.font.Family),
		Style:    inputGioFontStyle(spec.font.Style),
		Weight:   gioFont.Weight(spec.font.Weight),
	}
	ed.Color = spec.foreground
	ed.HintColor = spec.placeholderColor
	ed.TextSize = unit.Sp(spec.textSize)
	if spec.lineHeight > 0 {
		ed.LineHeight = unit.Sp(spec.lineHeight)
	}
	_ = layoutMaterialInputEditor(editorCtx, ed, t.config.singleLine)
	if opacity < 1 {
		opacityStack.Pop()
	}
	stack.Pop()
}

func (t *inputWidget) layoutInputSlots(ctx *internal.Context, size image.Point, spec inputRenderSpec) {
	slotSize := ctx.Gtx.Dp(safeDp(24))
	if slotSize <= 0 {
		slotSize = 24
	}
	y := (size.Y - slotSize) / 2
	if t.config.leading != nil {
		t.layoutInputSlot(ctx, t.config.leading, image.Rect(ctx.Gtx.Dp(safeDp(12)), y, ctx.Gtx.Dp(safeDp(12))+slotSize, y+slotSize), spec)
	}
	if t.config.trailing != nil {
		x := size.X - ctx.Gtx.Dp(safeDp(12)) - slotSize
		t.layoutInputSlot(ctx, t.config.trailing, image.Rect(x, y, x+slotSize, y+slotSize), spec)
	}
}

func (t *inputWidget) layoutInputSlot(ctx *internal.Context, child Widget, rect image.Rectangle, spec inputRenderSpec) {
	if child == nil {
		return
	}
	gtx := ctx.Gtx
	slotCtx := gtx
	slotCtx.Constraints = gioLayout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	next := *ctx.WithForeground(spec.placeholderColor)
	next.Gtx = slotCtx
	gioLayout.Center.Layout(slotCtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		centered := next
		centered.Gtx = gtx
		dims := child.Layout(centered.Child(0))
		return gioLayout.Dimensions{Size: dims.Size}
	})
	stack.Pop()
}

func (t *inputWidget) drawFloatingLabel(ctx *internal.Context, label string, left int, restY, floatY, progress float32, spec inputRenderSpec) {
	scale := float32(1) - 0.25*progress
	y := restY + (floatY-restY)*progress
	size := spec.textSize * scale
	if size < 11 {
		size = 11
	}
	stack := op.Offset(image.Pt(left, int(y+0.5))).Push(ctx.Gtx.Ops)
	ctx.LayoutText(internal.TextSpec{
		Content:    label,
		Size:       size,
		LineHeight: size * 1.2,
		Color:      spec.labelColor,
		Alignment:  internal.AlignStart,
		Font:       spec.font,
		FontReady:  true,
	})
	stack.Pop()
}

func (t *inputWidget) measureAffix(ctx *internal.Context, text string, spec inputRenderSpec) int {
	if strings.TrimSpace(text) == "" {
		return 0
	}
	return t.textSize(ctx, text, spec.textSize, spec.lineHeight, spec.font).X + ctx.Gtx.Dp(safeDp(6))
}

func (t *inputWidget) drawAffix(ctx *internal.Context, pos image.Point, text string, spec inputRenderSpec, opacity float32) {
	if opacity <= 0 {
		return
	}
	stack := op.Offset(pos).Push(ctx.Gtx.Ops)
	var opacityStack paint.OpacityStack
	if opacity < 1 {
		opacityStack = paint.PushOpacity(ctx.Gtx.Ops, opacity)
	}
	ctx.LayoutText(internal.TextSpec{
		Content:    text,
		Size:       spec.textSize,
		LineHeight: spec.lineHeight,
		Color:      spec.placeholderColor,
		Alignment:  internal.AlignStart,
		Font:       spec.font,
		FontReady:  true,
	})
	if opacity < 1 {
		opacityStack.Pop()
	}
	stack.Pop()
}

func (t *inputWidget) layoutSupportingRow(ctx *internal.Context, width int, fieldHeight int, supporting, counter string, spec inputRenderSpec) int {
	rowGap := ctx.Gtx.Dp(safeDp(4))
	rowTop := fieldHeight + rowGap
	left := ctx.Gtx.Dp(safeDp(16))
	right := ctx.Gtx.Dp(safeDp(16))
	rowHeight := ctx.Gtx.Dp(safeDp(20))
	counterWidth := 0
	if counter != "" {
		counterWidth = t.textSize(ctx, counter, 12, 16, spec.font).X
	}
	if supporting != "" {
		col := ctx.Theme().Colors.OnSurfaceVariant
		if t.config.error {
			col = ctx.Theme().Colors.Error
		}
		maxX := width - left - right
		if counterWidth > 0 {
			maxX -= counterWidth + ctx.Gtx.Dp(safeDp(12))
		}
		if maxX < 0 {
			maxX = 0
		}
		stack := op.Offset(image.Pt(left, rowTop)).Push(ctx.Gtx.Ops)
		clipStack := clip.Rect(image.Rect(0, 0, maxX, rowHeight)).Push(ctx.Gtx.Ops)
		supportCtx := *ctx
		supportCtx.Gtx.Constraints = gioLayout.Exact(image.Pt(maxX, rowHeight))
		supportCtx.LayoutText(internal.TextSpec{Content: supporting, Size: 12, LineHeight: 16, Color: col, Alignment: internal.AlignStart, Font: spec.font, FontReady: true})
		clipStack.Pop()
		stack.Pop()
	}
	if counter != "" {
		col := ctx.Theme().Colors.OnSurfaceVariant
		x := width - right - counterWidth
		if x < left {
			x = left
		}
		stack := op.Offset(image.Pt(x, rowTop)).Push(ctx.Gtx.Ops)
		ctx.LayoutText(internal.TextSpec{Content: counter, Size: 12, LineHeight: 16, Color: col, Alignment: internal.AlignStart, Font: spec.font, FontReady: true})
		stack.Pop()
	}
	return rowHeight + rowGap
}

func (t *inputWidget) supportingLineText() string {
	if t.config.error && strings.TrimSpace(t.config.errorText) != "" {
		return t.config.errorText
	}
	return t.config.supportingText
}

func (t *inputWidget) displayLabel() string {
	label := t.config.label
	if t.config.required && !t.config.noAsterisk {
		label += " *"
	}
	return label
}

func (t *inputWidget) counterText(value string) string {
	if !t.config.counter && t.config.maxLen <= 0 {
		return ""
	}
	if t.config.maxLen > 0 {
		return fmt.Sprintf("%d / %d", len([]rune(value)), t.config.maxLen)
	}
	return fmt.Sprintf("%d", len([]rune(value)))
}

func (t *inputWidget) labelWidth(ctx *internal.Context, text string, spec inputRenderSpec, scale float32) int {
	return t.textSize(ctx, text, spec.textSize*scale, spec.lineHeight*scale, spec.font).X
}

func (t *inputWidget) textSize(ctx *internal.Context, text string, size, lineHeight float32, font theme.FontSpec) image.Point {
	record := op.Record(ctx.Gtx.Ops)
	measureCtx := *ctx
	measureCtx.Gtx.Constraints.Min = image.Point{}
	dims := measureCtx.LayoutText(internal.TextSpec{Content: text, Size: size, LineHeight: lineHeight, Color: color.NRGBA{}, Alignment: internal.AlignStart, Font: font, FontReady: true})
	call := record.Stop()
	_ = call
	return dims
}

func inputGioFontStyle(style theme.FontStyle) gioFont.Style {
	if style == theme.FontStyleItalic {
		return gioFont.Italic
	}
	return gioFont.Regular
}

func layoutMaterialInputEditor(gtx gioLayout.Context, ed material.EditorStyle, singleLine bool) gioLayout.Dimensions {
	if !singleLine {
		return ed.Layout(gtx)
	}

	minY := gtx.Constraints.Min.Y
	editCtx := gtx
	editCtx.Constraints.Min.Y = 0

	macro := op.Record(gtx.Ops)
	dims := ed.Layout(editCtx)
	call := macro.Stop()

	size := dims.Size
	if size.X < gtx.Constraints.Min.X {
		size.X = gtx.Constraints.Min.X
	}
	if size.Y < minY {
		size.Y = minY
	}
	size = gtx.Constraints.Constrain(size)

	offsetY := (size.Y - dims.Size.Y) / 2
	if offsetY < 0 {
		offsetY = 0
	}
	offset := op.Offset(image.Pt(0, offsetY)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()

	return gioLayout.Dimensions{
		Size:     size,
		Baseline: dims.Baseline + offsetY,
	}
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
