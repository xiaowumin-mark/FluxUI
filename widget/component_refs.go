package widget

import "sync"

type commandQueue[T any] struct {
	mu          sync.Mutex
	commands    []T
	invalidator func()
}

func (q *commandQueue[T]) enqueue(cmd T) {
	q.mu.Lock()
	q.commands = append(q.commands, cmd)
	invalidator := q.invalidator
	q.mu.Unlock()
	if invalidator != nil {
		invalidator()
	}
}

func (q *commandQueue[T]) bindInvalidator(fn func()) {
	q.mu.Lock()
	q.invalidator = fn
	q.mu.Unlock()
}

func (q *commandQueue[T]) drainCommands() []T {
	q.mu.Lock()
	if len(q.commands) == 0 {
		q.mu.Unlock()
		return nil
	}
	out := make([]T, len(q.commands))
	copy(out, q.commands)
	q.commands = q.commands[:0]
	q.mu.Unlock()
	return out
}

type ButtonRef struct {
	queue commandQueue[struct{}]
}

func NewButtonRef() *ButtonRef {
	return &ButtonRef{}
}

func (r *ButtonRef) Click() {
	if r == nil {
		return
	}
	r.queue.enqueue(struct{}{})
}

func (r *ButtonRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *ButtonRef) drainCommands() []struct{} {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type ClickAreaRef struct {
	queue commandQueue[struct{}]
}

func NewClickAreaRef() *ClickAreaRef {
	return &ClickAreaRef{}
}

func (r *ClickAreaRef) Click() {
	if r == nil {
		return
	}
	r.queue.enqueue(struct{}{})
}

func (r *ClickAreaRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *ClickAreaRef) drainCommands() []struct{} {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type inputCommandKind uint8

const (
	inputCmdSetText inputCommandKind = iota + 1
	inputCmdAppend
	inputCmdClear
	inputCmdFocus
	inputCmdBlur
)

type inputCommand struct {
	kind inputCommandKind
	text string
}

type InputRef struct {
	queue commandQueue[inputCommand]
}

func NewInputRef() *InputRef {
	return &InputRef{}
}

func (r *InputRef) SetText(value string) {
	if r == nil {
		return
	}
	r.queue.enqueue(inputCommand{
		kind: inputCmdSetText,
		text: value,
	})
}

func (r *InputRef) Append(value string) {
	if r == nil || value == "" {
		return
	}
	r.queue.enqueue(inputCommand{
		kind: inputCmdAppend,
		text: value,
	})
}

func (r *InputRef) Clear() {
	if r == nil {
		return
	}
	r.queue.enqueue(inputCommand{kind: inputCmdClear})
}

func (r *InputRef) Focus() {
	if r == nil {
		return
	}
	r.queue.enqueue(inputCommand{kind: inputCmdFocus})
}

func (r *InputRef) Blur() {
	if r == nil {
		return
	}
	r.queue.enqueue(inputCommand{kind: inputCmdBlur})
}

func (r *InputRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *InputRef) drainCommands() []inputCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type boolCommandKind uint8

const (
	boolCmdSet boolCommandKind = iota + 1
	boolCmdToggle
)

type boolCommand struct {
	kind  boolCommandKind
	value bool
}

type CheckboxRef struct {
	queue commandQueue[boolCommand]
}

func NewCheckboxRef() *CheckboxRef {
	return &CheckboxRef{}
}

func (r *CheckboxRef) SetChecked(checked bool) {
	if r == nil {
		return
	}
	r.queue.enqueue(boolCommand{
		kind:  boolCmdSet,
		value: checked,
	})
}

func (r *CheckboxRef) Toggle() {
	if r == nil {
		return
	}
	r.queue.enqueue(boolCommand{kind: boolCmdToggle})
}

func (r *CheckboxRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *CheckboxRef) drainCommands() []boolCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type SwitchRef struct {
	queue commandQueue[boolCommand]
}

func NewSwitchRef() *SwitchRef {
	return &SwitchRef{}
}

func (r *SwitchRef) SetChecked(checked bool) {
	if r == nil {
		return
	}
	r.queue.enqueue(boolCommand{
		kind:  boolCmdSet,
		value: checked,
	})
}

func (r *SwitchRef) Toggle() {
	if r == nil {
		return
	}
	r.queue.enqueue(boolCommand{kind: boolCmdToggle})
}

func (r *SwitchRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *SwitchRef) drainCommands() []boolCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type sliderCommandKind uint8

const (
	sliderCmdSet sliderCommandKind = iota + 1
	sliderCmdStep
)

type sliderCommand struct {
	kind  sliderCommandKind
	value float32
	delta float32
}

type SliderRef struct {
	queue commandQueue[sliderCommand]
}

func NewSliderRef() *SliderRef {
	return &SliderRef{}
}

func (r *SliderRef) SetValue(value float32) {
	if r == nil {
		return
	}
	r.queue.enqueue(sliderCommand{
		kind:  sliderCmdSet,
		value: value,
	})
}

func (r *SliderRef) StepBy(delta float32) {
	if r == nil || delta == 0 {
		return
	}
	r.queue.enqueue(sliderCommand{
		kind:  sliderCmdStep,
		delta: delta,
	})
}

func (r *SliderRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *SliderRef) drainCommands() []sliderCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type RadioGroupRef struct {
	queue commandQueue[string]
}

func NewRadioGroupRef() *RadioGroupRef {
	return &RadioGroupRef{}
}

func (r *RadioGroupRef) SetValue(value string) {
	if r == nil {
		return
	}
	r.queue.enqueue(value)
}

func (r *RadioGroupRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *RadioGroupRef) drainCommands() []string {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type selectCommandKind uint8

const (
	selectCmdSetValue selectCommandKind = iota + 1
	selectCmdOpen
	selectCmdClose
	selectCmdToggle
)

type selectCommand[T comparable] struct {
	kind  selectCommandKind
	value T
}

type SelectRef[T comparable] struct {
	queue commandQueue[selectCommand[T]]
}

func NewSelectRef[T comparable]() *SelectRef[T] {
	return &SelectRef[T]{}
}

func (r *SelectRef[T]) SetValue(value T) {
	if r == nil {
		return
	}
	r.queue.enqueue(selectCommand[T]{
		kind:  selectCmdSetValue,
		value: value,
	})
}

func (r *SelectRef[T]) Open() {
	if r == nil {
		return
	}
	r.queue.enqueue(selectCommand[T]{kind: selectCmdOpen})
}

func (r *SelectRef[T]) Close() {
	if r == nil {
		return
	}
	r.queue.enqueue(selectCommand[T]{kind: selectCmdClose})
}

func (r *SelectRef[T]) Toggle() {
	if r == nil {
		return
	}
	r.queue.enqueue(selectCommand[T]{kind: selectCmdToggle})
}

func (r *SelectRef[T]) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *SelectRef[T]) drainCommands() []selectCommand[T] {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type TabsRef struct {
	queue commandQueue[string]
}

func NewTabsRef() *TabsRef {
	return &TabsRef{}
}

func (r *TabsRef) SetActive(key string) {
	if r == nil {
		return
	}
	r.queue.enqueue(key)
}

func (r *TabsRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *TabsRef) drainCommands() []string {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type DialogRef struct {
	queue commandQueue[boolCommand]
}

func NewDialogRef() *DialogRef {
	return &DialogRef{}
}

func (r *DialogRef) Open() {
	if r == nil {
		return
	}
	r.queue.enqueue(boolCommand{
		kind:  boolCmdSet,
		value: true,
	})
}

func (r *DialogRef) Close() {
	if r == nil {
		return
	}
	r.queue.enqueue(boolCommand{
		kind:  boolCmdSet,
		value: false,
	})
}

func (r *DialogRef) Toggle() {
	if r == nil {
		return
	}
	r.queue.enqueue(boolCommand{kind: boolCmdToggle})
}

func (r *DialogRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *DialogRef) drainCommands() []boolCommand {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

type BottomNavRef struct {
	queue commandQueue[string]
}

func NewBottomNavRef() *BottomNavRef {
	return &BottomNavRef{}
}

func (r *BottomNavRef) SetActive(key string) {
	if r == nil {
		return
	}
	r.queue.enqueue(key)
}

func (r *BottomNavRef) bindInvalidator(fn func()) {
	if r == nil {
		return
	}
	r.queue.bindInvalidator(fn)
}

func (r *BottomNavRef) drainCommands() []string {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}
