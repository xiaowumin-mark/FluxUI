package widget

import (
	"reflect"
	"unsafe"

	gioWidget "gioui.org/widget"
)

// Gio v0.9.0 does not expose an API for clearing or bounding Editor undo
// history. Keep the version-specific access isolated here and covered by tests.
func limitEditorUndoHistory(editor *gioWidget.Editor, maxEntries, maxBytes int) bool {
	if maxEntries < 0 || maxBytes < 0 {
		return false
	}
	history, nextIndex, ok := editorUndoHistoryFields(editor)
	if !ok {
		return false
	}
	index := int(nextIndex.Int())
	if index < 0 || index > history.Len() {
		return false
	}

	totalBytes := 0
	for i := 0; i < history.Len(); i++ {
		totalBytes += editorModificationBytes(history.Index(i))
	}

	// A Gio history is a linear chain and nextHistoryIdx is the current point
	// in that chain. Retain a contiguous window around that point: dropping a
	// modification between the cursor and a later redo would make the redo
	// suffix invalid for the editor's current contents.
	start, end := 0, history.Len()
	overLimit := func() bool {
		return end-start > maxEntries || totalBytes > maxBytes
	}
	for overLimit() && start < index {
		totalBytes -= editorModificationBytes(history.Index(start))
		start++
	}
	for overLimit() && end > index {
		end--
		totalBytes -= editorModificationBytes(history.Index(end))
	}
	if overLimit() {
		return false
	}

	retainedLen := end - start
	newIndex := index - start
	if retainedLen == 0 {
		clearEditorHistoryBacking(history)
		history.Set(reflect.Zero(history.Type()))
		nextIndex.SetInt(0)
		return true
	}

	// Gio truncates the visible slice after an undo followed by a new edit, but
	// the discarded redo strings can remain reachable through the slice's spare
	// capacity. Rebuild when trimming or when capacity exceeds the entry limit;
	// otherwise explicitly clear every unused slot.
	if start != 0 || end != history.Len() || history.Cap() > maxEntries {
		compacted := reflect.MakeSlice(history.Type(), retainedLen, retainedLen)
		reflect.Copy(compacted, history.Slice(start, end))
		clearEditorHistoryBacking(history)
		history.Set(compacted)
	} else {
		clearEditorHistoryTail(history)
	}
	nextIndex.SetInt(int64(newIndex))
	return true
}

func editorUndoHistoryFields(editor *gioWidget.Editor) (reflect.Value, reflect.Value, bool) {
	if editor == nil {
		return reflect.Value{}, reflect.Value{}, false
	}
	value := reflect.ValueOf(editor)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return reflect.Value{}, reflect.Value{}, false
	}
	value = value.Elem()
	if value.Type().PkgPath() != "gioui.org/widget" || value.Type().Name() != "Editor" {
		return reflect.Value{}, reflect.Value{}, false
	}
	history := value.FieldByName("history")
	nextIndex := value.FieldByName("nextHistoryIdx")
	if !history.IsValid() || history.Kind() != reflect.Slice || !history.CanAddr() ||
		!nextIndex.IsValid() || nextIndex.Kind() != reflect.Int || !nextIndex.CanAddr() {
		return reflect.Value{}, reflect.Value{}, false
	}
	modification := history.Type().Elem()
	if modification.Kind() != reflect.Struct || modification.PkgPath() != "gioui.org/widget" || modification.Name() != "modification" ||
		!editorModificationLayoutMatches(modification) {
		return reflect.Value{}, reflect.Value{}, false
	}
	return writableEditorField(history), writableEditorField(nextIndex), true
}

func editorModificationLayoutMatches(modification reflect.Type) bool {
	startRune, ok := modification.FieldByName("StartRune")
	if !ok || startRune.Type.Kind() != reflect.Int {
		return false
	}
	for _, name := range [...]string{"ApplyContent", "ReverseContent"} {
		field, ok := modification.FieldByName(name)
		if !ok || field.Type.Kind() != reflect.String {
			return false
		}
	}
	return true
}

func writableEditorField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

func clearEditorHistoryBacking(history reflect.Value) {
	if history.Cap() == 0 {
		return
	}
	full := history.Slice(0, history.Cap())
	for i := 0; i < full.Len(); i++ {
		full.Index(i).SetZero()
	}
}

func clearEditorHistoryTail(history reflect.Value) {
	if history.Len() == history.Cap() {
		return
	}
	full := history.Slice(0, history.Cap())
	for i := history.Len(); i < full.Len(); i++ {
		full.Index(i).SetZero()
	}
}

func editorModificationBytes(modification reflect.Value) int {
	if modification.Kind() != reflect.Struct {
		return 0
	}
	total := 0
	for _, name := range [...]string{"ApplyContent", "ReverseContent"} {
		field := modification.FieldByName(name)
		if field.IsValid() && field.Kind() == reflect.String {
			total += len(field.String())
		}
	}
	return total
}
