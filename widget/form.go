package widget

import (
	"image/color"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/internal/fieldstate"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"
	"github.com/xiaowumin-mark/FluxUI/theme"
)

// FormOption configures a Form. Options only describe the synchronous form
// snapshot; validation and submission work remain owned by the host.
type FormOption func(*formConfig)

type formConfig struct {
	onSubmit func(ctx *internal.Context, event *FormSubmitEvent)
	disabled bool
	pending  bool
	ref      *FormRef
}

// FormSubmitEvent describes a submit intention observed by a Form.
//
// Input is non-nil when the intention originated from a descendant Input
// submit event. FromRef is true for a FormRef command. Calling PreventDefault
// prevents the underlying Input event when one exists; it never starts, waits
// for, or cancels host business work.
type FormSubmitEvent struct {
	Input            *fluxevent.InputEvent
	FromRef          bool
	DefaultPrevented bool
}

// PreventDefault cancels the form submit intention. For an input-originated
// submission it also marks the bubbling Input submit event as prevented.
func (e *FormSubmitEvent) PreventDefault() {
	if e == nil {
		return
	}
	e.DefaultPrevented = true
	if e.Input != nil {
		e.Input.PreventDefault()
	}
}

type formCommand struct{}

// FormRef queues imperative form commands for consumption during the next
// layout of its attached Form.
type FormRef struct {
	queue commandQueue[formCommand]
}

// NewFormRef creates an unbound FormRef.
func NewFormRef() *FormRef {
	return &FormRef{}
}

// Submit queues a host-visible submit intention. It does not submit data by
// itself and is ignored while the attached Form is disabled or pending.
func (r *FormRef) Submit() {
	if r == nil {
		return
	}
	r.queue.enqueue(formCommand{})
}

func (r *FormRef) drainCommands() []formCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type formWidget struct {
	child  Widget
	config formConfig
}

// Form groups fields and forwards descendant Input submit intentions to the
// host. It owns neither field values nor validation/submission goroutines.
func Form(child Widget, opts ...FormOption) Widget {
	cfg := formConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &formWidget{child: child, config: cfg}
}

// FormOnSubmit observes an allowed submit intention. The host owns all
// validation, asynchronous work, and resulting pending/error snapshots.
func FormOnSubmit(fn func(ctx *internal.Context, event *FormSubmitEvent)) FormOption {
	return func(cfg *formConfig) {
		cfg.onSubmit = fn
	}
}

// FormDisabled blocks form submit intentions. It does not mutate arbitrary
// descendant controls; pass their own disabled options when that is desired.
func FormDisabled(disabled bool) FormOption {
	return func(cfg *formConfig) {
		cfg.disabled = disabled
	}
}

// FormPending blocks repeat form submit intentions while host work is pending.
func FormPending(pending bool) FormOption {
	return func(cfg *formConfig) {
		cfg.pending = pending
	}
}

// FormAttachRef binds a command FormRef to the Form for the current frame.
func FormAttachRef(ref *FormRef) FormOption {
	return func(cfg *formConfig) {
		cfg.ref = ref
	}
}

func (f *formWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	formCtx := ctx.Scope("form")
	f.registerSubmitListener(formCtx)

	if f.config.ref != nil {
		bindCommandRef(formCtx, "form", f.config.ref, &f.config.ref.queue)
		for range f.config.ref.drainCommands() {
			f.dispatchSubmit(formCtx, nil, true)
		}
	}

	if f.child == nil {
		return layout.Dimensions{}
	}
	return f.child.Layout(formCtx.Child(0))
}

func (f *formWidget) registerSubmitListener(ctx *internal.Context) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	if f.config.onSubmit == nil && !f.config.disabled && !f.config.pending {
		return
	}
	fluxevent.OnInput(ctx, fluxevent.Submit, func(listenerCtx *internal.Context, input *fluxevent.InputEvent) {
		if input == nil || input.DefaultPrevented {
			return
		}
		f.dispatchSubmit(listenerCtx, input, false)
	})
}

func (f *formWidget) dispatchSubmit(ctx *internal.Context, input *fluxevent.InputEvent, fromRef bool) {
	if f == nil || f.config.disabled || f.config.pending {
		if input != nil {
			input.PreventDefault()
		}
		return
	}
	if f.config.onSubmit == nil {
		return
	}
	event := &FormSubmitEvent{Input: input, FromRef: fromRef}
	f.config.onSubmit(ctx, event)
	if input != nil && input.DefaultPrevented {
		event.DefaultPrevented = true
	}
}

// FieldStatus identifies the host-supplied validation status of a FieldState.
type FieldStatus = fieldstate.Status

const (
	// FieldValid means the host has no current validation error.
	FieldValid FieldStatus = fieldstate.Valid
	// FieldInvalid prioritizes ErrorText over SupportingText.
	FieldInvalid FieldStatus = fieldstate.Invalid
	// FieldPending prioritizes PendingText over SupportingText.
	FieldPending FieldStatus = fieldstate.Pending

	// FieldStatusValid is an explicit-name alias for FieldValid.
	FieldStatusValid = FieldValid
	// FieldStatusInvalid is an explicit-name alias for FieldInvalid.
	FieldStatusInvalid = FieldInvalid
	// FieldStatusPending is an explicit-name alias for FieldPending.
	FieldStatusPending = FieldPending
)

// FieldState is the host-controlled, presentation-only snapshot for one form
// field. Key is stable identity for summaries and focus routing; it never
// stores the field's business value or runs validation.
type FieldState struct {
	Key            string
	Label          string
	SupportingText string
	ErrorText      string
	PendingText    string
	Required       bool
	Disabled       bool
	ReadOnly       bool
	// Pending is independent from Status so a host can retain an invalid error
	// while an asynchronous follow-up validation is in flight. StatusPending is
	// retained as a compatibility shorthand for Pending=true when no invalid
	// status is present.
	Pending bool
	Status  FieldStatus
}

// Normalized maps an unknown status to FieldValid.
func (s FieldState) Normalized() FieldState {
	pending := s.Pending || s.Status == FieldPending
	presentation := s.presentation()
	s.Label = presentation.Label
	s.SupportingText = presentation.SupportingText
	s.ErrorText = presentation.ErrorText
	s.PendingText = presentation.PendingText
	s.Required = presentation.Required
	s.Disabled = presentation.Disabled
	s.ReadOnly = presentation.ReadOnly
	s.Status = presentation.Status
	s.Pending = pending
	return s
}

// IsInvalid reports whether this normalized field state is invalid.
func (s FieldState) IsInvalid() bool {
	return s.presentation().IsInvalid()
}

// IsPending reports whether this normalized field state is pending validation.
func (s FieldState) IsPending() bool {
	s = s.Normalized()
	return s.Pending
}

// Message returns the primary message selected by the field-state precedence
// rule. An invalid field keeps its error/supporting message even when Pending
// is also true; use PendingMessage to render its separate progress text.
func (s FieldState) Message() string {
	s = s.Normalized()
	if s.IsInvalid() {
		if s.ErrorText != "" {
			return s.ErrorText
		}
		return s.SupportingText
	}
	if s.IsPending() && s.PendingText != "" {
		return s.PendingText
	}
	return s.SupportingText
}

// PendingMessage returns pending text when validation is in progress. When a
// field is invalid as well, this text is intentionally separate from Message
// so an error cannot be hidden by a pending indication.
func (s FieldState) PendingMessage() string {
	if !s.IsPending() {
		return ""
	}
	return s.Normalized().PendingText
}

func (s FieldState) presentation() fieldstate.State {
	return fieldstate.State{
		Label:          s.Label,
		SupportingText: s.SupportingText,
		ErrorText:      s.ErrorText,
		PendingText:    s.PendingText,
		Required:       s.Required,
		Disabled:       s.Disabled,
		ReadOnly:       s.ReadOnly,
		Status:         s.Status,
	}.Normalized()
}

// FormFieldOption configures a FormField presentation snapshot.
type FormFieldOption func(*formFieldConfig)

type formFieldConfig struct {
	state FieldState
	ref   *FormFieldRef
}

type formFieldCommand struct{}

// FormFieldRef queues imperative field commands for an attached FormField.
type FormFieldRef struct {
	queue commandQueue[formFieldCommand]
}

// NewFormFieldRef creates an unbound FormFieldRef.
func NewFormFieldRef() *FormFieldRef {
	return &FormFieldRef{}
}

// Focus queues a programmatic focus request. FormField is intentionally not a
// tab stop; this target exists for summary and host-directed focus routing.
func (r *FormFieldRef) Focus() {
	if r == nil {
		return
	}
	r.queue.enqueue(formFieldCommand{})
}

func (r *FormFieldRef) drainCommands() []formFieldCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type formFieldWidget struct {
	child  Widget
	config formFieldConfig
}

// FormField adds label, required, supporting/error/pending text, and a stable
// field identity around child. It does not mutate the child value or start
// validation; disabled/read-only remain presentation state unless the child is
// configured with matching options by its host.
func FormField(key string, child Widget, opts ...FormFieldOption) Widget {
	cfg := formFieldConfig{state: FieldState{Key: key}}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	// The constructor key is the identity authority. A state snapshot may carry
	// a duplicate key for use in summaries, but cannot retarget this field.
	cfg.state.Key = key
	return &formFieldWidget{child: child, config: cfg}
}

// FormFieldLabel sets the visible field label.
func FormFieldLabel(label string) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.Label = label
	}
}

// FormFieldSupportingText sets the normal supporting message.
func FormFieldSupportingText(text string) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.SupportingText = text
	}
}

// FormFieldErrorText sets the error message selected when Status is invalid.
func FormFieldErrorText(text string) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.ErrorText = text
	}
}

// FormFieldPendingText sets the pending message selected when Status is pending.
func FormFieldPendingText(text string) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.PendingText = text
	}
}

// FormFieldRequired controls required presentation.
func FormFieldRequired(required bool) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.Required = required
	}
}

// FormFieldDisabled controls disabled presentation and FormFieldRef focus.
func FormFieldDisabled(disabled bool) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.Disabled = disabled
	}
}

// FormFieldReadOnly controls read-only presentation.
func FormFieldReadOnly(readOnly bool) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.ReadOnly = readOnly
	}
}

// FormFieldPending controls pending presentation independently from invalid
// status so a host can show both an error and ongoing validation progress.
func FormFieldPending(pending bool) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state.Pending = pending
	}
}

// FormFieldState replaces the presentation snapshot, except that the key
// passed to FormField remains the field's stable identity.
func FormFieldState(state FieldState) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.state = state
	}
}

// FormFieldAttachRef binds a command FormFieldRef to the field.
func FormFieldAttachRef(ref *FormFieldRef) FormFieldOption {
	return func(cfg *formFieldConfig) {
		cfg.ref = ref
	}
}

func (f *formFieldWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	field := f.config.state.Normalized()
	fieldCtx := ctx.Scope("form-field:" + field.Key)
	fluxevent.RegisterFocusTarget(fieldCtx, fluxevent.FocusTabIndex(-1), fluxevent.FocusDisabled(field.Disabled))
	if f.config.ref != nil {
		bindCommandRef(fieldCtx, "form-field", f.config.ref, &f.config.ref.queue)
		if !field.Disabled {
			for range f.config.ref.drainCommands() {
				fluxevent.RequestFocus(fieldCtx)
			}
		} else {
			_ = f.config.ref.drainCommands()
		}
	}

	th := fieldCtx.Theme()
	labelColor := th.Colors.OnSurface
	if field.IsInvalid() {
		labelColor = th.Colors.Error
	}
	if field.Disabled {
		labelColor = style.DisabledContent(th.Colors.OnSurface)
	}
	messageColor := th.Colors.OnSurfaceVariant
	if field.IsInvalid() {
		messageColor = th.Colors.Error
	} else if field.IsPending() {
		messageColor = th.Colors.Primary
	}
	if field.Disabled {
		messageColor = style.DisabledContent(th.Colors.OnSurface)
	}

	label := field.Label
	if label != "" && field.Required {
		label += " *"
	}
	message, pendingMessage := formFieldMessages(field)
	// Keep these slots in a fixed order. In particular, an external validation
	// update which adds/removes a message must never shift child from slot 1.
	return Column(
		formFieldTextSlot{content: label, color: labelColor, textStyle: th.Types.LabelLarge, trailingGap: 4},
		formFieldChildSlot{child: f.child},
		formFieldTextSlot{content: message, color: messageColor, textStyle: th.Types.BodySmall, leadingGap: 4},
		formFieldTextSlot{content: pendingMessage, color: pendingMessageColor(fieldCtx, field), textStyle: th.Types.BodySmall, leadingGap: 2},
	).Layout(fieldCtx)
}

func formFieldMessages(field FieldState) (message, pendingMessage string) {
	field = field.Normalized()
	message = field.Message()
	if field.IsInvalid() && field.IsPending() {
		pendingMessage = field.PendingMessage()
	}
	return message, pendingMessage
}

func pendingMessageColor(ctx *internal.Context, field FieldState) color.NRGBA {
	th := ctx.Theme()
	if field.Disabled {
		return style.DisabledContent(th.Colors.OnSurface)
	}
	return th.Colors.Primary
}

type formFieldChildSlot struct {
	child Widget
}

func (s formFieldChildSlot) Layout(ctx *internal.Context) layout.Dimensions {
	if s.child == nil {
		return layout.Dimensions{}
	}
	return s.child.Layout(ctx.Child(0))
}

type formFieldTextSlot struct {
	content     string
	color       color.NRGBA
	textStyle   theme.TextStyle
	leadingGap  float32
	trailingGap float32
}

func (s formFieldTextSlot) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil || s.content == "" {
		return layout.Dimensions{}
	}
	children := make([]Widget, 0, 3)
	if s.leadingGap > 0 {
		children = append(children, VSpacer(s.leadingGap))
	}
	children = append(children, Text(s.content, TextColor(s.color), TextType(s.textStyle)))
	if s.trailingGap > 0 {
		children = append(children, VSpacer(s.trailingGap))
	}
	return Column(children...).Layout(ctx)
}

// ValidationSummaryOption configures a ValidationSummary.
type ValidationSummaryOption func(*validationSummaryConfig)

type validationSummaryConfig struct {
	onFocus   func(ctx *internal.Context, key string)
	title     string
	emptyText string
	disabled  bool
}

type validationSummaryWidget struct {
	fields []FieldState
	config validationSummaryConfig
}

// ValidationSummary renders host-supplied invalid field snapshots. Activating
// an error invokes OnFocus with its stable key, leaving actual focus/scroll
// routing to the host.
func ValidationSummary(fields []FieldState, opts ...ValidationSummaryOption) Widget {
	cfg := validationSummaryConfig{title: "Please correct the following fields"}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &validationSummaryWidget{
		fields: append([]FieldState(nil), fields...),
		config: cfg,
	}
}

// ValidationSummaryOnFocus is called when the user activates an invalid field
// entry. The host can focus and scroll the corresponding field by key.
func ValidationSummaryOnFocus(fn func(ctx *internal.Context, key string)) ValidationSummaryOption {
	return func(cfg *validationSummaryConfig) {
		cfg.onFocus = fn
	}
}

// ValidationSummaryTitle sets the title shown when invalid fields are present.
func ValidationSummaryTitle(title string) ValidationSummaryOption {
	return func(cfg *validationSummaryConfig) {
		cfg.title = title
	}
}

// ValidationSummaryEmptyText sets the text shown when no invalid fields exist.
func ValidationSummaryEmptyText(text string) ValidationSummaryOption {
	return func(cfg *validationSummaryConfig) {
		cfg.emptyText = text
	}
}

// ValidationSummaryDisabled disables error-entry activation.
func ValidationSummaryDisabled(disabled bool) ValidationSummaryOption {
	return func(cfg *validationSummaryConfig) {
		cfg.disabled = disabled
	}
}

func (s *validationSummaryWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	summaryCtx := ctx.Scope("validation-summary")
	invalid := validationSummaryInvalidFields(s.fields)
	if len(invalid) == 0 {
		if s.config.emptyText == "" {
			return layout.Dimensions{}
		}
		color := summaryTextColor(summaryCtx, s.config.disabled, false)
		return Text(s.config.emptyText, TextColor(color), TextType(summaryCtx.Theme().Types.BodySmall)).Layout(summaryCtx.Child(0))
	}

	children := make([]Widget, 0, len(invalid)+1)
	if s.config.title != "" {
		children = append(children, formFieldTextSlot{
			content:     s.config.title,
			color:       summaryTextColor(summaryCtx, s.config.disabled, true),
			textStyle:   summaryCtx.Theme().Types.TitleSmall,
			trailingGap: 4,
		})
	}
	for _, field := range invalid {
		children = append(children, validationSummaryItemWidget{
			field:    field,
			disabled: s.config.disabled,
			onFocus:  s.config.onFocus,
		})
	}
	return Column(children...).Layout(summaryCtx)
}

func validationSummaryInvalidFields(fields []FieldState) []FieldState {
	invalid := make([]FieldState, 0, len(fields))
	for _, field := range fields {
		field = field.Normalized()
		if !field.IsInvalid() || field.Message() == "" {
			continue
		}
		invalid = append(invalid, field)
	}
	return invalid
}

func summaryTextColor(ctx *internal.Context, disabled, title bool) color.NRGBA {
	th := ctx.Theme()
	if disabled {
		return style.DisabledContent(th.Colors.OnSurface)
	}
	if title {
		return th.Colors.Error
	}
	return th.Colors.Error
}

type validationSummaryItemWidget struct {
	field    FieldState
	disabled bool
	onFocus  func(ctx *internal.Context, key string)
}

func (s validationSummaryItemWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	itemCtx := ctx.Scope("validation-summary-item:" + s.field.Key)
	message := s.field.Message()
	if message == "" {
		return layout.Dimensions{}
	}
	content := message
	if s.field.Label != "" {
		content = s.field.Label + ": " + message
	}
	color := summaryTextColor(itemCtx, s.disabled, false)
	text := Text(content, TextColor(color), TextType(itemCtx.Theme().Types.BodySmall))
	if s.disabled || s.onFocus == nil {
		return text.Layout(itemCtx.Child(0))
	}
	return Pressable(text, func(actionCtx *internal.Context) {
		s.onFocus(actionCtx, s.field.Key)
	}).Layout(itemCtx.Child(0))
}
