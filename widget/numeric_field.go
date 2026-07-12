package widget

import (
	"math/big"
	"regexp"
	"strings"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
)

// NumericValue is the host-visible result of parsing a NumericField's raw
// text. Text is always the original controlled text; Value is a canonical,
// exact decimal when parsing succeeded. Invalid values retain their raw text
// so a host never loses an in-progress edit.
type NumericValue struct {
	Text  string
	Valid bool
	Value string
	Error string
}

// NumericFieldOption configures NumericField and SpinBox.
type NumericFieldOption func(*numericFieldConfig)

type numericFieldKind uint8

const (
	numericFieldKindText numericFieldKind = iota
	numericFieldKindSpinBox
)

type numericFieldConfig struct {
	min      string
	max      string
	step     string
	integer  bool
	disabled bool
	pending  bool
	error    bool
	required bool

	hasMin            bool
	hasMax            bool
	hasStep           bool
	hasDisabled       bool
	hasPending        bool
	hasError          bool
	hasRequired       bool
	hasPlaceholder    bool
	hasLabel          bool
	hasSupportingText bool
	hasErrorText      bool
	hasNoAsterisk     bool

	placeholder    string
	label          string
	supportingText string
	errorText      string
	noAsterisk     bool
	inputOpts      []InputOption

	onChange       func(ctx *internal.Context, value string)
	onParsedChange func(ctx *internal.Context, value NumericValue)
	ref            *NumericFieldRef
}

type numericFieldWidget struct {
	value  string
	kind   numericFieldKind
	config numericFieldConfig
}

// NumericField renders a single-line, controlled numeric text field. Its
// value is raw text rather than float64 so incomplete input and exact decimal
// values remain observable to the host.
func NumericField(value string, opts ...NumericFieldOption) Widget {
	return newNumericField(numericFieldKindText, value, opts...)
}

// SpinBox renders NumericField with exact increment and decrement controls.
func SpinBox(value string, opts ...NumericFieldOption) Widget {
	return newNumericField(numericFieldKindSpinBox, value, opts...)
}

func newNumericField(kind numericFieldKind, value string, opts ...NumericFieldOption) Widget {
	cfg := numericFieldConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.inputOpts = append([]InputOption(nil), cfg.inputOpts...)
	return &numericFieldWidget{value: value, kind: kind, config: cfg}
}

// NumericFieldMin sets the inclusive minimum exact decimal value.
func NumericFieldMin(value string) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.min = value
		cfg.hasMin = true
	}
}

// NumericFieldMax sets the inclusive maximum exact decimal value.
func NumericFieldMax(value string) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.max = value
		cfg.hasMax = true
	}
}

// NumericFieldStep sets the positive exact decimal increment used by SpinBox.
// A missing, malformed, or non-positive step falls back to one.
func NumericFieldStep(value string) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.step = value
		cfg.hasStep = true
	}
}

// NumericFieldInteger requires parsed values to be mathematical integers.
func NumericFieldInteger(integer bool) NumericFieldOption {
	return func(cfg *numericFieldConfig) { cfg.integer = integer }
}

// NumericFieldDisabled blocks user and NumericFieldRef mutation commands.
func NumericFieldDisabled(disabled bool) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.disabled = disabled
		cfg.hasDisabled = true
	}
}

// NumericFieldPlaceholder sets the field placeholder.
func NumericFieldPlaceholder(text string) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.placeholder = text
		cfg.hasPlaceholder = true
	}
}

// NumericFieldLabel sets the field label.
func NumericFieldLabel(text string) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.label = text
		cfg.hasLabel = true
	}
}

// NumericFieldSupportingText sets supporting text displayed below the field.
func NumericFieldSupportingText(text string) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.supportingText = text
		cfg.hasSupportingText = true
	}
}

// NumericFieldErrorText sets a host-supplied validation message. A non-empty
// message marks the field invalid; parse and range errors are used only when
// no host message is supplied.
func NumericFieldErrorText(text string) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.errorText = text
		cfg.hasErrorText = true
	}
}

// NumericFieldError controls the host-supplied invalid visual state.
func NumericFieldError(err bool) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.error = err
		cfg.hasError = true
	}
}

// NumericFieldRequired adds the required indicator without performing
// validation itself.
func NumericFieldRequired(required bool) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.required = required
		cfg.hasRequired = true
	}
}

// NumericFieldNoAsterisk hides the required indicator's asterisk.
func NumericFieldNoAsterisk(noAsterisk bool) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.noAsterisk = noAsterisk
		cfg.hasNoAsterisk = true
	}
}

// NumericFieldPending presents a host-controlled validation-pending state.
// It never starts validation work.
func NumericFieldPending(pending bool) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.pending = pending
		cfg.hasPending = true
	}
}

// NumericFieldInputOptions forwards styling and Input event options to the
// backing TextField. NumericField's value, disabled, callbacks, and attached
// NumericFieldRef remain authoritative.
func NumericFieldInputOptions(opts ...InputOption) NumericFieldOption {
	return func(cfg *numericFieldConfig) {
		cfg.inputOpts = append(cfg.inputOpts, opts...)
	}
}

// NumericFieldOnChange runs for each accepted raw text mutation, preserving
// TextField's input, paste, IME, history, and programmatic-ref ordering.
func NumericFieldOnChange(fn func(ctx *internal.Context, value string)) NumericFieldOption {
	return func(cfg *numericFieldConfig) { cfg.onChange = fn }
}

// NumericFieldOnParsedChange runs after NumericFieldOnChange for the same raw
// text mutation. It receives invalid intermediate text as NumericValue with
// Valid set to false.
func NumericFieldOnParsedChange(fn func(ctx *internal.Context, value NumericValue)) NumericFieldOption {
	return func(cfg *numericFieldConfig) { cfg.onParsedChange = fn }
}

// NumericFieldAttachRef binds commands that operate on the backing input.
func NumericFieldAttachRef(ref *NumericFieldRef) NumericFieldOption {
	return func(cfg *numericFieldConfig) { cfg.ref = ref }
}

func (n *numericFieldWidget) Layout(ctx *internal.Context) layout.Dimensions {
	inputOpts := n.inputOptions()
	if n.config.ref != nil {
		bindCommandRef(ctx, "numeric-field", n.config.ref, &n.config.ref.queue)
		inputRef := numericFieldInputRefFor(ctx)
		n.consumeRef(inputRef)
		inputOpts = append(inputOpts, InputAttachRef(inputRef))
	}

	field := TextField(n.value, inputOpts...)
	if n.kind != numericFieldKindSpinBox {
		return field.Layout(ctx)
	}

	controls := FixedWidth(28, Column(
		IconButton(
			Text("▲", TextSize(10)),
			IconButtonSize(24),
			IconButtonDisabled(n.config.disabled),
			IconButtonOnClick(func(actionCtx *internal.Context) { n.step(actionCtx, 1) }),
		),
		IconButton(
			Text("▼", TextSize(10)),
			IconButtonSize(24),
			IconButtonDisabled(n.config.disabled),
			IconButtonOnClick(func(actionCtx *internal.Context) { n.step(actionCtx, -1) }),
		),
	))
	return Row(Expanded(field), controls).Layout(ctx)
}

func (n *numericFieldWidget) inputOptions() []InputOption {
	options := append([]InputOption(nil), n.config.inputOpts...)
	options = append(options, InputSingleLine(true))

	if n.config.hasPlaceholder {
		options = append(options, InputPlaceholder(n.config.placeholder))
	}
	if n.config.hasLabel {
		options = append(options, InputLabel(n.config.label))
	}
	if n.config.hasRequired {
		options = append(options, InputRequired(n.config.required))
	}
	if n.config.hasNoAsterisk {
		options = append(options, InputNoAsterisk(n.config.noAsterisk))
	}
	if n.config.hasDisabled {
		options = append(options, InputDisabled(n.config.disabled))
	}

	parsed := parseNumericValue(n.value, n.config)
	errorText := ""
	if n.config.hasErrorText {
		errorText = n.config.errorText
	}
	if errorText == "" {
		errorText = parsed.Error
	}
	errored := n.config.error || errorText != ""
	if n.config.hasError || errorText != "" {
		options = append(options, InputError(errored))
	}
	if errorText != "" {
		options = append(options, InputErrorText(errorText))
	}

	if n.config.pending && !errored {
		pendingText := "Validating…"
		if n.config.hasSupportingText && n.config.supportingText != "" {
			pendingText += " " + n.config.supportingText
		}
		options = append(options, InputSupportingText(pendingText))
	} else if n.config.hasSupportingText {
		options = append(options, InputSupportingText(n.config.supportingText))
	}
	if n.config.pending && errored {
		// Pending is a separate host snapshot, not a replacement for an error.
		// Keep InputErrorText visible and surface progress independently.
		options = append(options, InputTrailing(Text("Validating…")))
	}

	options = append(options, InputOnChange(func(ctx *internal.Context, value string) {
		n.dispatchChange(ctx, value)
	}))
	return options
}

func (n *numericFieldWidget) consumeRef(inputRef *InputRef) {
	if n == nil || n.config.ref == nil || inputRef == nil {
		return
	}
	commands := n.config.ref.drainCommands()
	if n.config.disabled {
		return
	}
	for _, command := range commands {
		switch command.kind {
		case numericFieldRefSetText:
			inputRef.SetText(command.text)
		case numericFieldRefAppend:
			inputRef.Append(command.text)
		case numericFieldRefClear:
			inputRef.Clear()
		case numericFieldRefFocus:
			inputRef.Focus()
		case numericFieldRefBlur:
			inputRef.Blur()
		}
	}
}

func (n *numericFieldWidget) dispatchChange(ctx *internal.Context, value string) {
	if n == nil {
		return
	}
	if n.config.onChange != nil {
		n.config.onChange(ctx, value)
	}
	if n.config.onParsedChange != nil {
		n.config.onParsedChange(ctx, parseNumericValue(value, n.config))
	}
}

func (n *numericFieldWidget) step(ctx *internal.Context, direction int) {
	if n == nil || n.config.disabled {
		return
	}
	next, changed := numericSteppedText(n.value, n.config, direction)
	if changed {
		n.dispatchChange(ctx, next)
	}
}

// ParseNumericValue parses text with no component-specific bounds. It is useful
// to hosts that need the same exact decimal representation as NumericField.
func ParseNumericValue(text string) NumericValue {
	return parseNumericValue(text, numericFieldConfig{})
}

var numericDecimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)

func parseNumericValue(text string, cfg numericFieldConfig) NumericValue {
	result := NumericValue{Text: text}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return result
	}

	value, ok := parseNumericRat(trimmed)
	if !ok {
		result.Error = "Enter a valid number"
		return result
	}
	result.Value = canonicalNumericRat(value)
	if cfg.integer && !value.IsInt() {
		result.Error = "Enter a whole number"
		return result
	}

	min, hasMin := numericBound(cfg.min, cfg.hasMin)
	max, hasMax := numericBound(cfg.max, cfg.hasMax)
	if hasMin && hasMax && min.Cmp(max) > 0 {
		hasMin = false
		hasMax = false
	}
	if hasMin && value.Cmp(min) < 0 {
		result.Error = "Value must be at least " + canonicalNumericRat(min)
		return result
	}
	if hasMax && value.Cmp(max) > 0 {
		result.Error = "Value must be at most " + canonicalNumericRat(max)
		return result
	}
	result.Valid = true
	return result
}

func parseNumericRat(text string) (*big.Rat, bool) {
	if !numericDecimalPattern.MatchString(text) {
		return nil, false
	}
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		return nil, false
	}
	return value, true
}

func numericBound(text string, declared bool) (*big.Rat, bool) {
	if !declared {
		return nil, false
	}
	return parseNumericRat(strings.TrimSpace(text))
}

func numericSteppedText(text string, cfg numericFieldConfig, direction int) (string, bool) {
	if direction == 0 {
		return text, false
	}

	min, hasMin := numericBound(cfg.min, cfg.hasMin)
	max, hasMax := numericBound(cfg.max, cfg.hasMax)
	if hasMin && hasMax && min.Cmp(max) > 0 {
		hasMin = false
		hasMax = false
	}

	current, validCurrent := parseNumericRat(strings.TrimSpace(text))
	if validCurrent && cfg.integer && !current.IsInt() {
		validCurrent = false
	}
	if !validCurrent {
		if direction > 0 && hasMin {
			next := canonicalNumericRat(min)
			return next, next != text
		}
		if direction < 0 && hasMax {
			next := canonicalNumericRat(max)
			return next, next != text
		}
		current = new(big.Rat)
	}

	step := numericStepRat(cfg)
	next := new(big.Rat).Set(current)
	if direction > 0 {
		next.Add(next, step)
	} else {
		next.Sub(next, step)
	}
	if hasMin && next.Cmp(min) < 0 {
		next.Set(min)
	}
	if hasMax && next.Cmp(max) > 0 {
		next.Set(max)
	}
	result := canonicalNumericRat(next)
	return result, result != text
}

func numericStepRat(cfg numericFieldConfig) *big.Rat {
	step, ok := numericBound(cfg.step, cfg.hasStep)
	if !ok || step.Sign() <= 0 || (cfg.integer && !step.IsInt()) {
		return big.NewRat(1, 1)
	}
	return step
}

func canonicalNumericRat(value *big.Rat) string {
	if value == nil || value.Sign() == 0 {
		return "0"
	}
	precision, terminating := decimalPrecision(value.Denom())
	if !terminating {
		return value.RatString()
	}
	text := value.FloatString(precision)
	if strings.Contains(text, ".") {
		text = strings.TrimRight(text, "0")
		text = strings.TrimRight(text, ".")
	}
	if text == "-0" || text == "" {
		return "0"
	}
	return text
}

func decimalPrecision(denominator *big.Int) (int, bool) {
	if denominator == nil || denominator.Sign() <= 0 {
		return 0, false
	}
	remaining := new(big.Int).Set(denominator)
	twos := 0
	for remaining.Bit(0) == 0 {
		remaining.Rsh(remaining, 1)
		twos++
	}

	fives := 0
	five := big.NewInt(5)
	remainder := new(big.Int)
	for {
		remainder.Mod(remaining, five)
		if remainder.Sign() != 0 {
			break
		}
		remaining.Quo(remaining, five)
		fives++
	}
	if remaining.Cmp(big.NewInt(1)) != 0 {
		return 0, false
	}
	if twos > fives {
		return twos, true
	}
	return fives, true
}

type numericFieldInputState struct {
	ref *InputRef
}

func numericFieldInputRefFor(ctx *internal.Context) *InputRef {
	if ctx == nil {
		return NewInputRef()
	}
	value := ctx.Memo("numeric-field-input-ref", func() any {
		return &numericFieldInputState{ref: NewInputRef()}
	})
	state, ok := value.(*numericFieldInputState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: numeric field state type mismatch")
	}
	if state.ref == nil {
		state.ref = NewInputRef()
	}
	return state.ref
}

type numericFieldRefCommandKind uint8

const (
	numericFieldRefSetText numericFieldRefCommandKind = iota + 1
	numericFieldRefAppend
	numericFieldRefClear
	numericFieldRefFocus
	numericFieldRefBlur
)

type numericFieldRefCommand struct {
	kind numericFieldRefCommandKind
	text string
}

func (c numericFieldRefCommand) pendingRefBytes() int {
	return len(c.text) + 1
}

// NumericFieldRef queues programmatic input commands for NumericField. Commands
// are consumed during the next layout of the attached field and are ignored
// while NumericFieldDisabled(true) is active.
type NumericFieldRef struct {
	queue commandQueue[numericFieldRefCommand]
}

// NewNumericFieldRef creates a command reference for NumericField or SpinBox.
func NewNumericFieldRef() *NumericFieldRef {
	return &NumericFieldRef{}
}

// SetValue programmatically replaces the raw numeric text.
func (r *NumericFieldRef) SetValue(value string) {
	r.SetText(value)
}

// SetText programmatically replaces the raw numeric text.
func (r *NumericFieldRef) SetText(value string) {
	if r == nil {
		return
	}
	r.queue.enqueue(numericFieldRefCommand{kind: numericFieldRefSetText, text: value})
}

// Append appends raw text using the backing Input command semantics.
func (r *NumericFieldRef) Append(value string) {
	if r == nil || value == "" {
		return
	}
	r.queue.enqueue(numericFieldRefCommand{kind: numericFieldRefAppend, text: value})
}

// Clear clears the raw numeric text.
func (r *NumericFieldRef) Clear() {
	if r == nil {
		return
	}
	r.queue.enqueue(numericFieldRefCommand{kind: numericFieldRefClear})
}

// Focus requests focus for the backing input.
func (r *NumericFieldRef) Focus() {
	if r == nil {
		return
	}
	r.queue.enqueue(numericFieldRefCommand{kind: numericFieldRefFocus})
}

// Blur clears focus from the backing input.
func (r *NumericFieldRef) Blur() {
	if r == nil {
		return
	}
	r.queue.enqueue(numericFieldRefCommand{kind: numericFieldRefBlur})
}

func (r *NumericFieldRef) drainCommands() []numericFieldRefCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}
