package event

import "github.com/xiaowumin-mark/FluxUI/internal"

const (
	DragStart Type = "dragstart"
	Drag      Type = "drag"
	DragEnter Type = "dragenter"
	DragOver  Type = "dragover"
	DragLeave Type = "dragleave"
	Drop      Type = "drop"
	DragEnd   Type = "dragend"
)

// DragEvent carries drag-and-drop data through the generic event system.
type DragEvent struct {
	Event

	MIMEType  string
	Data      []byte
	Text      string
	Paths     []string
	Operation string
	Types     []string
	Active    bool
	Err       error
	Native    any
}

// DragHandler handles a typed drag event.
type DragHandler func(ctx *internal.Context, event *DragEvent)

// OnDrag registers a typed drag listener for the current context target.
func OnDrag(ctx *internal.Context, eventType Type, handler DragHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, eventType, func(ctx *internal.Context, ev *Event) {
		if ev == nil {
			return
		}
		if dragEvent, ok := ev.Detail.(*DragEvent); ok {
			handler(ctx, dragEvent)
		}
	}, opts...)
}

// DispatchDragEvent dispatches a typed drag event.
func DispatchDragEvent(ctx *internal.Context, target TargetID, ev *DragEvent) bool {
	if ev == nil {
		return true
	}
	applyDragDefaults(ev)
	ev.Event.Detail = ev
	return DispatchEvent(ctx, target, &ev.Event)
}

func applyDragDefaults(ev *DragEvent) {
	if ev.Type == "" {
		ev.Type = Drag
	}
	ev.Bubbles = true
	switch ev.Type {
	case DragEnd, DragLeave:
		ev.Cancelable = false
	default:
		ev.Cancelable = true
	}
}
