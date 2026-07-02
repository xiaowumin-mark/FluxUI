package widget

import (
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"
	"github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestMD3InactiveStateLayerSkipsMotionState(t *testing.T) {
	t.Setenv(theme.InteractionQualityEnv, "")
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: time.Now()}, rt)
	key := md3MotionKey(ctx, "idle")

	rt.BeginFrame()
	if got := md3AnimateStateLayerOpacity(ctx, "idle", 0); got != 0 {
		t.Fatalf("expected idle state layer opacity 0, got %v", got)
	}
	if _, ok := ctx.PersistentValueKey(key); ok {
		t.Fatal("expected idle state layer not to create motion state")
	}
	rt.EndFrame()
}

func TestMD3StateLayerReleasesAfterExit(t *testing.T) {
	t.Setenv(theme.InteractionQualityEnv, "")
	rt := internal.NewRuntime(nil)
	base := time.Now()
	var key internal.MemoryKey

	var ops op.Ops
	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: base}, rt)
	key = md3MotionKey(ctx, "hover")
	if got := md3AnimateStateLayerOpacity(ctx, "hover", style.StateLayerHoverOpacity); got != 0 {
		t.Fatalf("expected hover animation to start at 0, got %v", got)
	}
	if _, ok := ctx.PersistentValueKey(key); !ok {
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
	if _, ok := ctx.PersistentValueKey(key); ok {
		t.Fatal("expected completed state layer motion state to be released")
	}
	rt.EndFrame()
}

func TestMD3BalancedStateLayerSnapsHover(t *testing.T) {
	rt := internal.NewRuntime(theme.Default(theme.WithInteractionQuality(theme.InteractionQualityBalanced)))
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: time.Now()}, rt)
	key := md3MotionKey(ctx, "hover")

	rt.BeginFrame()
	if got := md3AnimateStateLayerOpacity(ctx, "hover", style.StateLayerHoverOpacity); got != style.StateLayerHoverOpacity {
		t.Fatalf("expected balanced hover to snap to target, got %v", got)
	}
	if _, ok := ctx.PersistentValueKey(key); ok {
		t.Fatal("expected balanced hover not to create motion state")
	}
	rt.EndFrame()
}

func TestMD3LowCPUStateLayerDisablesHoverKeepsPressed(t *testing.T) {
	rt := internal.NewRuntime(theme.Default(theme.WithInteractionQuality(theme.InteractionQualityLowCPU)))
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: time.Now()}, rt)
	hoverKey := md3MotionKey(ctx, "hover")
	pressedKey := md3MotionKey(ctx, "pressed")

	rt.BeginFrame()
	if got := md3AnimateStateLayerOpacity(ctx, "hover", style.StateLayerHoverOpacity); got != 0 {
		t.Fatalf("expected low CPU hover opacity 0, got %v", got)
	}
	if _, ok := ctx.PersistentValueKey(hoverKey); ok {
		t.Fatal("expected low CPU hover not to create motion state")
	}
	if got := md3AnimateStateLayerOpacity(ctx, "pressed", style.StateLayerPressedOpacity); got != 0 {
		t.Fatalf("expected pressed animation to start at 0, got %v", got)
	}
	if _, ok := ctx.PersistentValueKey(pressedKey); !ok {
		t.Fatal("expected low CPU pressed feedback to keep motion state")
	}
	rt.EndFrame()
}

func TestMD3SwitchProgressAnimatesFirstToggleAfterIdleFrame(t *testing.T) {
	rt := internal.NewRuntime(nil)
	base := time.Now()
	var ops op.Ops

	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: base}, rt)
	if got := md3SwitchSelectionProgress(ctx, false); got != 0 {
		t.Fatalf("initial unchecked switch progress = %v, want 0", got)
	}
	rt.EndFrame()

	ops.Reset()
	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: base.Add(16 * time.Millisecond)}, rt)
	if got := md3SwitchSelectionProgress(ctx, true); got != 0 {
		t.Fatalf("first checked switch progress = %v, want animation to start from 0", got)
	}
	rt.EndFrame()

	ops.Reset()
	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: base.Add(140 * time.Millisecond)}, rt)
	got := md3SwitchSelectionProgress(ctx, true)
	if got <= 0 || got >= 1 {
		t.Fatalf("mid-flight checked switch progress = %v, want between 0 and 1", got)
	}
	rt.EndFrame()
}

func TestMD3SwitchPressedAnimatesFirstPressAfterIdleFrame(t *testing.T) {
	rt := internal.NewRuntime(nil)
	base := time.Now()
	var ops op.Ops

	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: base}, rt)
	if got := md3SwitchPressedProgress(ctx, false); got != 0 {
		t.Fatalf("initial switch press progress = %v, want 0", got)
	}
	rt.EndFrame()

	ops.Reset()
	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: base.Add(16 * time.Millisecond)}, rt)
	if got := md3SwitchPressedProgress(ctx, true); got != 0 {
		t.Fatalf("first switch press progress = %v, want animation to start from 0", got)
	}
	rt.EndFrame()
}

func TestMD3SwitchRetainedAnimationsIdleWithoutBookkeeping(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{Enabled: true, MeasureDurations: true})
	base := time.Now()
	var ops op.Ops

	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: base}, rt)
	_ = md3SwitchSelectionProgress(ctx, false)
	_ = md3SwitchPressedProgress(ctx, false)
	_ = md3AnimateColorRetained(ctx, "switch-idle-color", style.NRGBA(1, 2, 3, 255), 150*time.Millisecond, style.InteractionStandardEasing)
	rt.EndFrame()

	ops.Reset()
	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: base.Add(16 * time.Millisecond)}, rt)
	_ = md3SwitchSelectionProgress(ctx, false)
	_ = md3SwitchPressedProgress(ctx, false)
	_ = md3AnimateColorRetained(ctx, "switch-idle-color", style.NRGBA(1, 2, 3, 255), 150*time.Millisecond, style.InteractionStandardEasing)
	rt.EndFrame()

	stats := rt.LastFrameStats()
	if stats.Animation.Count != 0 {
		t.Fatalf("idle retained switch animations recorded animation count=%d, want 0", stats.Animation.Count)
	}
	if len(stats.Reasons) != 0 {
		t.Fatalf("idle retained switch animations requested redraw reasons=%v, want none", stats.Reasons)
	}
}
