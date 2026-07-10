package widget

import (
	"bytes"
	"image"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"

	gioFont "gioui.org/font"
	"gioui.org/font/gofont"
	gioEvent "gioui.org/io/event"
	gioInput "gioui.org/io/input"
	"gioui.org/io/key"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	gioText "gioui.org/text"
	"gioui.org/unit"
	gioWidget "gioui.org/widget"
)

func TestInputPasswordDisablesAndClearsHistory(t *testing.T) {
	state := &inputState{}
	state.configureInputHistory(true, "visible", true)
	state.recordInputHistory("visible text", fluxevent.InputSourceUser)
	if len(state.valueHistory) != 2 {
		t.Fatalf("expected normal input history, got %#v", state.valueHistory)
	}

	state.configureInputHistory(false, "secret", false)
	state.recordInputHistory("secret value", fluxevent.InputSourceUser)
	if state.historyEnabled {
		t.Fatal("expected password history to remain disabled")
	}
	if len(state.valueHistory) != 0 || state.historyBytes != 0 || state.historyIndex != -1 {
		t.Fatalf("expected password history to be cleared, got %#v", state)
	}

	state.configureInputHistory(true, "visible again", false)
	if !state.historyEnabled || len(state.valueHistory) != 1 || state.valueHistory[0] != "visible again" {
		t.Fatalf("expected normal history to restart from the current value, got %#v", state.valueHistory)
	}
}

func TestInputHistoryIsBoundedByEntriesAndBytes(t *testing.T) {
	state := &inputState{}
	state.configureInputHistory(true, "", true)
	for i := 0; i < inputHistoryMaxEntries*2; i++ {
		value := strconv.Itoa(i) + strings.Repeat("x", 8<<10)
		state.recordInputHistory(value, fluxevent.InputSourceUser)
	}

	if len(state.valueHistory) > inputHistoryMaxEntries {
		t.Fatalf("history entry limit exceeded: %d", len(state.valueHistory))
	}
	if state.historyBytes > inputHistoryMaxBytes {
		t.Fatalf("history byte limit exceeded: %d", state.historyBytes)
	}
	total := 0
	for _, value := range state.valueHistory {
		total += len(value)
	}
	if total != state.historyBytes {
		t.Fatalf("history byte accounting mismatch: got %d want %d", state.historyBytes, total)
	}
	if state.historyIndex != len(state.valueHistory)-1 {
		t.Fatalf("expected history index at latest entry, got %d of %d", state.historyIndex, len(state.valueHistory))
	}

	state.recordInputHistory(strings.Repeat("z", inputHistoryMaxBytes+1), fluxevent.InputSourceUser)
	if len(state.valueHistory) != 0 || state.historyBytes != 0 {
		t.Fatal("expected an oversized snapshot to clear retained history")
	}
}

func TestInputStateDisposesWhenUnmounted(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	var ops op.Ops

	runtime.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, runtime)
	state := inputStateFor(ctx)
	state.syncedValue = "sensitive"
	state.configureInputHistory(true, "sensitive", true)
	runtime.EndFrame()

	runtime.BeginFrame()
	runtime.EndFrame()

	if state.editor != nil || state.syncedValue != "" || len(state.valueHistory) != 0 {
		t.Fatalf("expected unmounted input state to be cleared, got %#v", state)
	}
}

func TestPasswordInputClearsGioEditorUndoHistory(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	var ops op.Ops

	runtime.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, runtime).Scope("password")
	TextField("secret-value", InputPassword(true)).Layout(ctx)
	state := inputStateFor(internal.NewContext(gioLayout.Context{Ops: &ops}, runtime).Scope("password"))
	runtime.EndFrame()

	history, nextIndex, ok := editorUndoHistoryFields(state.editor)
	if !ok {
		t.Fatal("expected Gio Editor undo history adapter to be available")
	}
	if history.Len() != 0 || nextIndex.Int() != 0 {
		t.Fatalf("password editor retained undo history: entries=%d index=%d", history.Len(), nextIndex.Int())
	}
}

func TestPasswordInputUndoCannotRestoreEarlierText(t *testing.T) {
	editor := &gioWidget.Editor{SingleLine: true}
	editor.SetText("first-secret")
	editor.SetText("replacement-secret")

	input := &inputWidget{config: inputConfig{password: true}}
	input.constrainEditorHistory(editor)
	sendEditorKeys(t, editor, key.Event{
		Name:      "Z",
		Modifiers: key.ModShortcut,
		State:     key.Press,
	})

	if got := editor.Text(); got != "replacement-secret" {
		t.Fatalf("password undo restored retained text: got %q", got)
	}
}

func TestPasswordInputSameBatchEditAndUndoKeepsNewValue(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	harness := newPasswordInputHarness(runtime)

	state := harness.render(t, 0, "first-password", nil)
	harness.router.Source().Execute(key.FocusCmd{Tag: state.editor})
	state = harness.render(t, 1, "first-password", nil)
	oldEditor := state.editor

	state = harness.render(t, 2, "first-password", nil,
		key.EditEvent{
			Range: key.Range{Start: 0, End: len("first-password")},
			Text:  "replacement-password",
		},
		key.SelectionEvent{Start: 4, End: 11},
		key.Event{
			Name:      "Z",
			Modifiers: key.ModShortcut,
			State:     key.Press,
		},
	)

	if got := state.editor.Text(); got != "replacement-password" {
		t.Fatalf("same-batch undo restored the previous password: got %q", got)
	}
	if state.editor == oldEditor {
		t.Fatal("password edit reused the Gio Editor with sensitive backing storage")
	}
	if start, end := state.editor.Selection(); start != 4 || end != 11 {
		t.Fatalf("password editor replacement lost selection: got (%d, %d)", start, end)
	}
	if !harness.router.Source().Focused(state.editor) {
		t.Fatal("password editor replacement did not preserve focus")
	}
}

func TestPasswordInputRebuildPreservesPendingSubmit(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	harness := newPasswordInputHarness(runtime)
	var submitted []string
	options := []InputOption{InputOnSubmit(func(_ *internal.Context, ev *fluxevent.InputEvent) {
		submitted = append(submitted, ev.Value)
	})}

	state := harness.render(t, 0, "old-password", options)
	harness.router.Source().Execute(key.FocusCmd{Tag: state.editor})
	state = harness.render(t, 1, "old-password", options)
	oldEditor := state.editor
	state = harness.render(t, 2, "old-password", options, key.EditEvent{
		Range: key.Range{Start: 0, End: len("old-password")},
		Text:  "submitted-password\n",
	})

	if len(submitted) != 1 || submitted[0] != "submitted-password" {
		t.Fatalf("password submit events = %v, want [submitted-password]", submitted)
	}
	if state.editor == oldEditor {
		t.Fatal("password submit mutation did not rebuild the Gio Editor")
	}
}

func TestPasswordInputRebuildsEditorWithoutRetainingReplacedText(t *testing.T) {
	t.Run("programmatic clear", func(t *testing.T) {
		runtime := internal.NewRuntime(nil)
		defer runtime.Dispose()
		harness := newPasswordInputHarness(runtime)
		ref := NewInputRef()
		secret := "programmatic-secret-marker"

		state := harness.render(t, 0, secret, []InputOption{InputAttachRef(ref)})
		oldEditor := state.editor
		_ = oldEditor.Text()
		if !editorRetainsSensitiveBytes(oldEditor, secret) {
			t.Fatal("test setup did not place the password in Gio backing storage")
		}

		ref.Clear()
		state = harness.render(t, 1, secret, []InputOption{InputAttachRef(ref)})
		assertPasswordEditorRebuiltWithout(t, state.editor, oldEditor, secret)
		if got := state.editor.Text(); got != "" {
			t.Fatalf("password clear left value %q", got)
		}
	})

	t.Run("controlled replacement", func(t *testing.T) {
		runtime := internal.NewRuntime(nil)
		defer runtime.Dispose()
		harness := newPasswordInputHarness(runtime)
		oldSecret := "controlled-old-secret-marker"
		options := []InputOption{InputOnChange(func(*internal.Context, string) {})}

		state := harness.render(t, 0, oldSecret, options)
		oldEditor := state.editor
		_ = oldEditor.Text()
		state = harness.render(t, 1, "controlled-new-secret", options)

		assertPasswordEditorRebuiltWithout(t, state.editor, oldEditor, oldSecret)
	})

	t.Run("canceled beforeinput", func(t *testing.T) {
		runtime := internal.NewRuntime(nil)
		defer runtime.Dispose()
		harness := newPasswordInputHarness(runtime)
		previous := "accepted-password"
		rejected := "rejected-secret-marker"
		options := []InputOption{InputOnBeforeInput(func(_ *internal.Context, ev *fluxevent.InputEvent) {
			ev.PreventDefault()
		})}

		state := harness.render(t, 0, previous, options)
		harness.router.Source().Execute(key.FocusCmd{Tag: state.editor})
		state = harness.render(t, 1, previous, options)
		oldEditor := state.editor
		state = harness.render(t, 2, previous, options, key.EditEvent{
			Range: key.Range{Start: 0, End: len(previous)},
			Text:  rejected,
		})

		assertPasswordEditorRebuiltWithout(t, state.editor, oldEditor, rejected)
		if got := state.editor.Text(); got != previous {
			t.Fatalf("canceled beforeinput left value %q, want %q", got, previous)
		}
	})

	t.Run("dispose", func(t *testing.T) {
		secret := "disposed-secret-marker"
		editor := &gioWidget.Editor{}
		editor.SetText(secret)
		_ = editor.Text()
		if !editorRetainsSensitiveBytes(editor, secret) {
			t.Fatal("test setup did not place the password in Gio backing storage")
		}
		state := &inputState{editor: editor, syncedValue: secret}
		state.dispose()
		if state.editor != nil || state.syncedValue != "" {
			t.Fatalf("disposed input state still retains editor metadata: %#v", state)
		}
		if editorRetainsSensitiveBytes(editor, secret) {
			t.Fatal("disposed Gio Editor still references password bytes")
		}
	})
}

func TestGioEditorUndoHistoryIsBounded(t *testing.T) {
	editor := &gioWidget.Editor{}
	for i := 0; i < inputHistoryMaxEntries*2; i++ {
		editor.SetText(strconv.Itoa(i) + strings.Repeat("x", 8<<10))
		if !limitEditorUndoHistory(editor, inputHistoryMaxEntries, inputHistoryMaxBytes) {
			t.Fatal("expected Gio Editor undo history adapter to be available")
		}
	}

	history, _, ok := editorUndoHistoryFields(editor)
	if !ok {
		t.Fatal("expected Gio Editor undo history fields")
	}
	totalBytes := 0
	for i := 0; i < history.Len(); i++ {
		totalBytes += editorModificationBytes(history.Index(i))
	}
	if history.Len() > inputHistoryMaxEntries || totalBytes > inputHistoryMaxBytes {
		t.Fatalf("Gio Editor history exceeded limits: entries=%d bytes=%d", history.Len(), totalBytes)
	}
	if history.Cap() > inputHistoryMaxEntries {
		t.Fatalf("Gio Editor history capacity exceeded limit: %d", history.Cap())
	}
}

func TestGioEditorHistoryTrimPreservesRedoChain(t *testing.T) {
	editor := &gioWidget.Editor{SingleLine: true}
	for _, value := range []string{"a", "ab", "abc", "abcd"} {
		editor.SetText(value)
	}

	undo := key.Event{Name: "Z", Modifiers: key.ModShortcut, State: key.Press}
	redo := key.Event{Name: "Z", Modifiers: key.ModShortcut | key.ModShift, State: key.Press}
	sendEditorKeys(t, editor, undo, undo)
	if got := editor.Text(); got != "ab" {
		t.Fatalf("unexpected text before trim: got %q want %q", got, "ab")
	}

	if !limitEditorUndoHistory(editor, 2, inputHistoryMaxBytes) {
		t.Fatal("expected Gio Editor undo history adapter to be available")
	}
	history, nextIndex, ok := editorUndoHistoryFields(editor)
	if !ok {
		t.Fatal("expected Gio Editor undo history fields")
	}
	if history.Len() != 2 || nextIndex.Int() != 0 {
		t.Fatalf("unexpected trimmed redo window: entries=%d index=%d", history.Len(), nextIndex.Int())
	}

	sendEditorKeys(t, editor, redo, redo)
	if got := editor.Text(); got != "abcd" {
		t.Fatalf("trimmed redo chain is invalid: got %q want %q", got, "abcd")
	}
	sendEditorKeys(t, editor, undo, undo, undo)
	if got := editor.Text(); got != "ab" {
		t.Fatalf("trimmed undo boundary is invalid: got %q want %q", got, "ab")
	}
}

func TestGioEditorHistoryClearsDiscardedRedoBacking(t *testing.T) {
	editor := &gioWidget.Editor{SingleLine: true}
	for _, value := range []string{"a", "ab", "abc", "abcd"} {
		editor.SetText(value)
	}
	undo := key.Event{Name: "Z", Modifiers: key.ModShortcut, State: key.Press}
	sendEditorKeys(t, editor, undo, undo)
	editor.SetText("branch")

	history, _, ok := editorUndoHistoryFields(editor)
	if !ok {
		t.Fatal("expected Gio Editor undo history fields")
	}
	if history.Cap() <= history.Len() || editorHistoryTailBytes(history) == 0 {
		t.Fatal("test setup did not create a retained redo tail")
	}
	if !limitEditorUndoHistory(editor, inputHistoryMaxEntries, inputHistoryMaxBytes) {
		t.Fatal("expected Gio Editor undo history adapter to be available")
	}
	if tailBytes := editorHistoryTailBytes(history); tailBytes != 0 {
		t.Fatalf("discarded redo data remains in history backing array: %d bytes", tailBytes)
	}
}

func editorHistoryTailBytes(history reflect.Value) int {
	if history.Cap() == history.Len() {
		return 0
	}
	full := history.Slice(0, history.Cap())
	total := 0
	for i := history.Len(); i < full.Len(); i++ {
		total += editorModificationBytes(full.Index(i))
	}
	return total
}

type passwordInputHarness struct {
	runtime *internal.Runtime
	router  gioInput.Router
	ops     op.Ops
}

func newPasswordInputHarness(runtime *internal.Runtime) *passwordInputHarness {
	return &passwordInputHarness{runtime: runtime}
}

func (h *passwordInputHarness) render(t *testing.T, frame int, value string, options []InputOption, events ...gioEvent.Event) *inputState {
	t.Helper()
	for _, event := range events {
		h.router.Queue(event)
	}

	h.ops.Reset()
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(image.Pt(320, 56)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC).Add(time.Duration(frame) * 16 * time.Millisecond),
		Source:      h.router.Source(),
		Ops:         &h.ops,
	}
	h.runtime.BeginFrame()
	ctx := internal.NewContext(gtx, h.runtime).Scope("password-input")
	allOptions := make([]InputOption, 0, len(options)+1)
	allOptions = append(allOptions, InputPassword(true))
	allOptions = append(allOptions, options...)
	TextField(value, allOptions...).Layout(ctx)
	state := inputStateFor(internal.NewContext(gtx, h.runtime).Scope("password-input"))
	h.runtime.EndFrame()
	h.router.Frame(&h.ops)
	return state
}

func assertPasswordEditorRebuiltWithout(t *testing.T, editor, oldEditor *gioWidget.Editor, secret string) {
	t.Helper()
	if editor == oldEditor {
		t.Fatal("password mutation reused the Gio Editor")
	}
	if editorRetainsSensitiveBytes(oldEditor, secret) {
		t.Fatal("detached Gio Editor still references replaced password bytes")
	}
	if editorRetainsSensitiveBytes(editor, secret) {
		t.Fatal("replacement Gio Editor retained the replaced password bytes")
	}
}

func editorRetainsSensitiveBytes(editor *gioWidget.Editor, secret string) bool {
	if editor == nil || secret == "" {
		return false
	}
	needle := []byte(secret)
	value := reflect.ValueOf(editor).Elem()

	for _, field := range []reflect.Value{
		value.FieldByName("scratch"),
		value.FieldByName("ime").FieldByName("scratch"),
	} {
		if reflectedByteSliceContains(field, needle) {
			return true
		}
	}

	buffer := writableEditorField(value.FieldByName("buffer"))
	if !buffer.IsNil() && reflectedByteSliceContains(buffer.Elem().FieldByName("text"), needle) {
		return true
	}

	history, _, ok := editorUndoHistoryFields(editor)
	if ok {
		for i := 0; i < history.Len(); i++ {
			entry := history.Index(i)
			for _, name := range [...]string{"ApplyContent", "ReverseContent"} {
				if strings.Contains(entry.FieldByName(name).String(), secret) {
					return true
				}
			}
		}
	}
	return false
}

func reflectedByteSliceContains(field reflect.Value, needle []byte) bool {
	if !field.IsValid() || field.Kind() != reflect.Slice || field.Type().Elem().Kind() != reflect.Uint8 || field.Cap() == 0 {
		return false
	}
	field = writableEditorField(field)
	return bytes.Contains(field.Slice(0, field.Cap()).Bytes(), needle)
}

func sendEditorKeys(t *testing.T, editor *gioWidget.Editor, events ...key.Event) {
	t.Helper()
	router := new(gioInput.Router)
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Constraints: gioLayout.Exact(image.Pt(320, 48)),
		Source:      router.Source(),
	}
	shaper := gioText.NewShaper(gioText.NoSystemFonts(), gioText.WithCollection(gofont.Collection()))
	layoutEditor := func() {
		editor.Layout(gtx, shaper, gioFont.Font{}, unit.Sp(14), op.CallOp{}, op.CallOp{})
	}

	gtx.Execute(key.FocusCmd{Tag: editor})
	layoutEditor()
	router.Frame(&ops)
	ops.Reset()
	layoutEditor()
	router.Frame(&ops)
	ops.Reset()
	for _, event := range events {
		router.Queue(event)
	}
	layoutEditor()
}
