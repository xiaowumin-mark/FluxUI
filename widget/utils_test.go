package widget

import (
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestMD3InactiveStateLayerSkipsMotionState(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: time.Now()}, rt)
	key := md3MotionKey(ctx, "idle")

	rt.BeginFrame()
	if got := md3AnimateStateLayerOpacity(ctx, "idle", 0); got != 0 {
		t.Fatalf("expected idle state layer opacity 0, got %v", got)
	}
	if _, ok := ctx.PersistentValue(key); ok {
		t.Fatal("expected idle state layer not to create motion state")
	}
	rt.EndFrame()
}

func TestMD3StateLayerReleasesAfterExit(t *testing.T) {
	rt := internal.NewRuntime(nil)
	base := time.Now()
	key := ""

	var ops op.Ops
	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: base}, rt)
	key = md3MotionKey(ctx, "hover")
	if got := md3AnimateStateLayerOpacity(ctx, "hover", style.StateLayerHoverOpacity); got != 0 {
		t.Fatalf("expected hover animation to start at 0, got %v", got)
	}
	if _, ok := ctx.PersistentValue(key); !ok {
		t.Fatal("expected active state layer to create motion state")
	}
	rt.EndFrame()

	ops.Reset()
	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: base.Add(style.InteractionHoverEnterDuration + time.Millisecond)}, rt)
	if got := md3AnimateStateLayerOpacity(ctx, "hover", 0); got <= 0 {
		t.Fatalf("expected exit animation to retain positive opacity, got %v", got)
	}
	rt.EndFrame()

	ops.Reset()
	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: base.Add(style.InteractionHoverEnterDuration + style.InteractionPressedExitDuration + 2*time.Millisecond)}, rt)
	if got := md3AnimateStateLayerOpacity(ctx, "hover", 0); got != 0 {
		t.Fatalf("expected completed exit opacity 0, got %v", got)
	}
	if _, ok := ctx.PersistentValue(key); ok {
		t.Fatal("expected completed state layer motion state to be released")
	}
	rt.EndFrame()
}
