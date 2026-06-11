package widget

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

type fixedTestWidget struct {
	size image.Point
}

func (w fixedTestWidget) Layout(ctx *internal.Context) layout.Dimensions {
	return layout.Dimensions{Size: ctx.Gtx.Constraints.Constrain(w.size)}
}

type pointerSinkTestWidget struct {
	size image.Point
	tag  any
}

func (w pointerSinkTestWidget) Layout(ctx *internal.Context) layout.Dimensions {
	dims := ctx.Gtx.Constraints.Constrain(w.size)
	stack := clip.Rect(image.Rectangle{Max: dims}).Push(ctx.Gtx.Ops)
	gioEvent.Op(ctx.Gtx.Ops, w.tag)
	stack.Pop()
	for {
		if _, ok := ctx.Gtx.Event(pointer.Filter{
			Target: w.tag,
			Kinds:  pointer.Press | pointer.Drag | pointer.Release,
		}); !ok {
			break
		}
	}
	return layout.Dimensions{Size: dims}
}

func TestDragSourceToDropTargetTransfer(t *testing.T) {
	runDragSourceToDropTargetTransfer(t, false, fixedTestWidget{size: image.Pt(20, 20)}, fixedTestWidget{size: image.Pt(20, 20)})
}

func TestDragSourceToDropTargetTransferWithPassThroughHitTesting(t *testing.T) {
	runDragSourceToDropTargetTransfer(t, true, fixedTestWidget{size: image.Pt(20, 20)}, fixedTestWidget{size: image.Pt(20, 20)})
}

func TestDragSourceToInteractiveDropTargetTransfer(t *testing.T) {
	runDragSourceToDropTargetTransfer(t, false,
		pointerSinkTestWidget{size: image.Pt(20, 20), tag: new(int)},
		pointerSinkTestWidget{size: image.Pt(20, 20), tag: new(int)},
	)
}

func runDragSourceToDropTargetTransfer(t *testing.T, passThrough bool, sourceChild, targetChild Widget) {
	t.Helper()

	var router input.Router
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Constraints: gioLayout.Exact(image.Pt(80, 24)),
		Source:      router.Source(),
	}
	runtime := internal.NewRuntime(nil)

	var sourceEvents []DragSourceEvent
	var drops []DropEvent
	dragSource := DragSource(
		sourceChild,
		DragSourceText("hello world"),
		DragSourceOnEvent(func(ctx *internal.Context, event DragSourceEvent) {
			sourceEvents = append(sourceEvents, event)
		}),
	)
	dropTarget := DropTarget(
		targetChild,
		func(ctx *internal.Context, event DropEvent) {
			drops = append(drops, event)
		},
		DropTargetTypes("text/plain"),
	)
	root := dragDropPairTestWidget{
		source:      dragSource,
		target:      dropTarget,
		passThrough: passThrough,
	}

	layoutDragDropTestFrame(t, runtime, gtx, root)
	router.Frame(gtx.Ops)
	router.Queue(
		pointer.Event{
			Source:   pointer.Mouse,
			Buttons:  pointer.ButtonPrimary,
			Kind:     pointer.Press,
			Position: f32.Pt(10, 10),
		},
		// The first move must still hit the source so Gio records the data source.
		pointer.Event{
			Source:   pointer.Mouse,
			Buttons:  pointer.ButtonPrimary,
			Kind:     pointer.Move,
			Position: f32.Pt(10, 10),
		},
	)

	layoutDragDropTestFrame(t, runtime, gtx, root)
	router.Frame(gtx.Ops)
	router.Queue(pointer.Event{
		Source:   pointer.Mouse,
		Kind:     pointer.Release,
		Position: f32.Pt(50, 10),
	})
	layoutDragDropTestFrame(t, runtime, gtx, root)
	router.Frame(gtx.Ops)
	layoutDragDropTestFrame(t, runtime, gtx, root)

	if len(drops) != 1 {
		t.Fatalf("expected one drop event, got drops=%#v sourceEvents=%#v", drops, sourceEvents)
	}
	if drops[0].Err != nil {
		t.Fatalf("unexpected drop error: %v", drops[0].Err)
	}
	if drops[0].Type != "text/plain" || drops[0].Text != "hello world" || string(drops[0].Data) != "hello world" {
		t.Fatalf("unexpected drop payload: %#v", drops[0])
	}
	if drops[0].Operation != DragOperationCopy {
		t.Fatalf("expected default copy operation, got %q", drops[0].Operation)
	}

	if !containsDragSourceEvent(sourceEvents, DragSourceEventStarted) {
		t.Fatalf("expected started source event, got %#v", sourceEvents)
	}
	if !containsDragSourceEvent(sourceEvents, DragSourceEventRequested) {
		t.Fatalf("expected requested source event, got %#v", sourceEvents)
	}
	if !containsDragSourceEvent(sourceEvents, DragSourceEventCompleted) {
		t.Fatalf("expected completed source event, got %#v", sourceEvents)
	}
}

func TestDragSourceCancelEventWhenDroppedWithoutTarget(t *testing.T) {
	var router input.Router
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Constraints: gioLayout.Exact(image.Pt(80, 24)),
		Source:      router.Source(),
	}
	runtime := internal.NewRuntime(nil)

	var sourceEvents []DragSourceEvent
	root := dragDropPairTestWidget{
		source: DragSource(
			fixedTestWidget{size: image.Pt(20, 20)},
			DragSourceText("hello world"),
			DragSourceOnEvent(func(ctx *internal.Context, event DragSourceEvent) {
				sourceEvents = append(sourceEvents, event)
			}),
		),
	}

	layoutDragDropTestFrame(t, runtime, gtx, root)
	router.Frame(gtx.Ops)
	router.Queue(
		pointer.Event{
			Source:   pointer.Mouse,
			Buttons:  pointer.ButtonPrimary,
			Kind:     pointer.Press,
			Position: f32.Pt(10, 10),
		},
		pointer.Event{
			Source:   pointer.Mouse,
			Buttons:  pointer.ButtonPrimary,
			Kind:     pointer.Move,
			Position: f32.Pt(10, 10),
		},
	)

	layoutDragDropTestFrame(t, runtime, gtx, root)
	router.Frame(gtx.Ops)
	router.Queue(pointer.Event{
		Source:   pointer.Mouse,
		Kind:     pointer.Release,
		Position: f32.Pt(50, 10),
	})
	layoutDragDropTestFrame(t, runtime, gtx, root)
	router.Frame(gtx.Ops)
	layoutDragDropTestFrame(t, runtime, gtx, root)

	if !containsDragSourceEvent(sourceEvents, DragSourceEventStarted) {
		t.Fatalf("expected started source event, got %#v", sourceEvents)
	}
	if !containsDragSourceEvent(sourceEvents, DragSourceEventCancelled) {
		t.Fatalf("expected cancelled source event, got %#v", sourceEvents)
	}
	if containsDragSourceEvent(sourceEvents, DragSourceEventRequested) || containsDragSourceEvent(sourceEvents, DragSourceEventCompleted) {
		t.Fatalf("unexpected transfer completion events for cancelled drag: %#v", sourceEvents)
	}
}

type dragDropPairTestWidget struct {
	source      Widget
	target      Widget
	passThrough bool
}

func (w dragDropPairTestWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.passThrough {
		pass := pointer.PassOp{}.Push(ctx.Gtx.Ops)
		defer pass.Pop()
	}

	if w.source != nil {
		sourceCtx := ctx.Scope("source")
		sourceCtx.Gtx.Constraints = gioLayout.Exact(image.Pt(20, 20))
		w.source.Layout(sourceCtx)
	}
	if w.target != nil {
		stack := op.Offset(image.Pt(40, 0)).Push(ctx.Gtx.Ops)
		targetCtx := ctx.Scope("target")
		targetCtx.Gtx.Constraints = gioLayout.Exact(image.Pt(20, 20))
		w.target.Layout(targetCtx)
		stack.Pop()
	}
	return layout.Dimensions{Size: image.Pt(80, 24)}
}

func layoutDragDropTestFrame(t *testing.T, runtime *internal.Runtime, gtx gioLayout.Context, root Widget) {
	t.Helper()
	if gtx.Ops != nil {
		gtx.Ops.Reset()
	}
	runtime.BeginFrame()
	root.Layout(internal.NewContext(gtx, runtime))
	runtime.EndFrame()
}

func containsDragSourceEvent(events []DragSourceEvent, kind DragSourceEventKind) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}
