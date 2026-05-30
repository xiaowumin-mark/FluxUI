package ui

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	anim "github.com/xiaowumin-mark/FluxUI/anim"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/state"
	widgetpkg "github.com/xiaowumin-mark/FluxUI/widget"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestUseStateInitialValue(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	s := UseState(ctx, 7)
	if s.Value() != 7 {
		t.Fatalf("expected initial value 7, got %d", s.Value())
	}
	s.Set(9)
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	s2 := UseState(ctx2, 7)
	rt.EndFrame()
	if s2.Value() != 9 {
		t.Fatalf("expected persisted value 9, got %d", s2.Value())
	}
}

func TestUseStateFallsBackToLegacyWithoutComponentInstance(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	s := UseState(ctx, 7)
	if s.Key() != "root/state:0" {
		t.Fatalf("expected legacy path key without component instance, got %q", s.Key())
	}
	rt.EndFrame()
}

func TestUseStateUsesComponentHookSlot(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "Counter", Key: "main"}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	s := UseState(ctx.WithComponentInstance(inst), 1)
	if s.Value() != 1 {
		t.Fatalf("expected initial hook slot value 1, got %d", s.Value())
	}
	if s.Key() == "root/state:0" {
		t.Fatalf("expected hook slot key, got legacy key %q", s.Key())
	}
	s.Set(5)
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	inst2 := rt.HookStore().BeginInstance(identity)
	s2 := UseState(ctx2.WithComponentInstance(inst2), 1)
	rt.EndFrame()
	if s2.Value() != 5 {
		t.Fatalf("expected hook slot persisted value 5, got %d", s2.Value())
	}
}

func TestUseStateHookSlotTypeMismatchPanics(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "Mismatch", Key: "main"}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	UseState(ctx.WithComponentInstance(inst), 1)
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	inst2 := rt.HookStore().BeginInstance(identity)
	defer func() {
		r := recover()
		rt.EndFrame()
		if r == nil {
			t.Fatal("expected panic when hook slot state type changes")
		}
	}()
	UseState(ctx2.WithComponentInstance(inst2), "wrong")
}

func TestUseEffectUsesComponentHookSlotEveryFrame(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "EffectDemo", Key: "main"}
	runCount := 0
	cleanupCount := 0

	for range 3 {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		inst := rt.HookStore().BeginInstance(identity)
		UseEffect(ctx.WithComponentInstance(inst), func() func() {
			runCount++
			return func() { cleanupCount++ }
		})
		rt.EndFrame()
	}

	if runCount != 3 {
		t.Fatalf("expected effect to run every frame, got %d", runCount)
	}
	if cleanupCount != 2 {
		t.Fatalf("expected cleanup before reruns, got %d", cleanupCount)
	}
}

func TestUseEffectWithDepsUsesComponentHookSlot(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "DepsDemo", Key: "main"}
	runCount := 0
	cleanupCount := 0

	for range 3 {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		inst := rt.HookStore().BeginInstance(identity)
		UseEffectWithDeps(ctx.WithComponentInstance(inst), []any{42}, func() func() {
			runCount++
			return func() { cleanupCount++ }
		})
		rt.EndFrame()
	}

	if runCount != 1 {
		t.Fatalf("expected unchanged deps effect to run once, got %d", runCount)
	}
	if cleanupCount != 0 {
		t.Fatalf("expected no cleanup without rerun or unmount, got %d", cleanupCount)
	}

	for dep := 43; dep <= 44; dep++ {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		inst := rt.HookStore().BeginInstance(identity)
		UseEffectWithDeps(ctx.WithComponentInstance(inst), []any{dep}, func() func() {
			runCount++
			return func() { cleanupCount++ }
		})
		rt.EndFrame()
	}

	if runCount != 3 {
		t.Fatalf("expected effect to rerun when deps change, got %d", runCount)
	}
	if cleanupCount != 2 {
		t.Fatalf("expected cleanup before changed-deps reruns, got %d", cleanupCount)
	}
}

func TestUseEffectHookSlotUnmountCleanup(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "UnmountDemo", Key: "main"}
	cleanupCount := 0

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	UseMount(ctx.WithComponentInstance(inst), func() func() {
		return func() { cleanupCount++ }
	})
	rt.EndFrame()

	rt.BeginFrame()
	rt.EndFrame()

	if cleanupCount != 1 {
		t.Fatalf("expected cleanup on hook instance unmount, got %d", cleanupCount)
	}
}

func TestUseEffectFallsBackToLegacyWithoutComponentInstance(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	runCount := 0

	for range 3 {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		UseEffectWithDeps(ctx, []any{"legacy"}, func() func() {
			runCount++
			return nil
		})
		rt.EndFrame()
	}

	if runCount != 1 {
		t.Fatalf("expected legacy deps effect to run once, got %d", runCount)
	}
}

func TestUseMemoUsesComponentHookSlot(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "MemoDemo", Key: "main"}
	runs := 0

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	first := UseMemo(ctx.WithComponentInstance(inst), []any{1}, func() int {
		runs++
		return 10
	})
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	inst2 := rt.HookStore().BeginInstance(identity)
	second := UseMemo(ctx2.WithComponentInstance(inst2), []any{1}, func() int {
		runs++
		return 20
	})
	rt.EndFrame()

	if first != 10 || second != 10 || runs != 1 {
		t.Fatalf("expected memoized value to be reused, first=%d second=%d runs=%d", first, second, runs)
	}

	rt.BeginFrame()
	ctx3 := internal.NewContext(gtx, rt)
	inst3 := rt.HookStore().BeginInstance(identity)
	third := UseMemo(ctx3.WithComponentInstance(inst3), []any{2}, func() int {
		runs++
		return 30
	})
	rt.EndFrame()

	if third != 30 || runs != 2 {
		t.Fatalf("expected deps change to recompute, third=%d runs=%d", third, runs)
	}
}

func TestUseRefPersistsComponentHookSlot(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "RefDemo", Key: "main"}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	ref := UseRef(ctx.WithComponentInstance(inst), 1)
	ref.Current = 42
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	inst2 := rt.HookStore().BeginInstance(identity)
	ref2 := UseRef(ctx2.WithComponentInstance(inst2), 1)
	rt.EndFrame()

	if ref2 != ref || ref2.Current != 42 {
		t.Fatalf("expected ref to persist, got %#v", ref2)
	}
}

func TestUseCallbackUsesMemoIdentity(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "CallbackDemo", Key: "main"}

	fn1 := func() int { return 1 }
	fn2 := func() int { return 2 }

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	cb1 := UseCallback(ctx.WithComponentInstance(inst), []any{"same"}, fn1)
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	inst2 := rt.HookStore().BeginInstance(identity)
	cb2 := UseCallback(ctx2.WithComponentInstance(inst2), []any{"same"}, fn2)
	rt.EndFrame()

	if cb1() != 1 || cb2() != 1 {
		t.Fatalf("expected callback to be memoized while deps are unchanged")
	}
}

func TestUseContextReturnsDefaultWithoutProvider(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	key := ContextKey[string]{Default: "fallback"}

	if got := UseContext(ctx, key); got != "fallback" {
		t.Fatalf("expected default context value, got %q", got)
	}
}

func TestProviderOverridesContextForSubtree(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	key := ContextKey[string]{Default: "fallback"}
	seen := ""
	child := captureElement(func(ctx *Context) {
		seen = UseContext(ctx, key)
	})

	renderElementWithContext(ctx, Provider(key, "provided", child))

	if seen != "provided" {
		t.Fatalf("expected provided context value, got %q", seen)
	}
	if got := UseContext(ctx, key); got != "fallback" {
		t.Fatalf("provider should not mutate parent context, got %q", got)
	}
}

func TestProviderNearestValueWins(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	key := ContextKey[int]{Default: 1}
	seen := 0
	child := captureElement(func(ctx *Context) {
		seen = UseContext(ctx, key)
	})

	renderElementWithContext(ctx, Provider(key, 2, Provider(key, 3, child)))

	if seen != 3 {
		t.Fatalf("expected nearest provider value 3, got %d", seen)
	}
}

func TestProviderSiblingIsolation(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	key := ContextKey[string]{Default: "default"}
	seenA := ""
	seenB := ""

	frag := Fragment(
		Provider(key, "a", captureElement(func(ctx *Context) {
			seenA = UseContext(ctx, key)
		})),
		captureElement(func(ctx *Context) {
			seenB = UseContext(ctx, key)
		}),
	)
	renderElementWithContext(ctx, frag)

	if seenA != "a" {
		t.Fatalf("expected first sibling provider value a, got %q", seenA)
	}
	if seenB != "default" {
		t.Fatalf("expected second sibling default value, got %q", seenB)
	}
}

func TestProviderPreservesLegacyStateFallback(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	key := ContextKey[string]{Default: "fallback"}
	stateKey := ""
	seenContext := ""

	rt.BeginFrame()
	renderElementWithContext(ctx, Provider(key, "provided", captureElement(func(ctx *Context) {
		seenContext = UseContext(ctx, key)
		stateKey = UseState(ctx, 10).Key()
	})))
	rt.EndFrame()

	if seenContext != "provided" {
		t.Fatalf("expected provided context value, got %q", seenContext)
	}
	if stateKey != "root/state:0" {
		t.Fatalf("expected legacy state path key inside provider, got %q", stateKey)
	}
}

func TestLegacyStateAndHookSlotStateCoexist(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "MixedState", Key: "main"}
	var legacyKey string
	var hookKey string

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	legacy := state.UseWithInitial(ctx, 10)
	inst := rt.HookStore().BeginInstance(identity)
	hook := UseState(ctx.WithComponentInstance(inst), 20)
	legacy.Set(11)
	hook.Set(22)
	legacyKey = legacy.Key()
	hookKey = hook.Key()
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	legacy2 := state.UseWithInitial(ctx2, 10)
	inst2 := rt.HookStore().BeginInstance(identity)
	hook2 := UseState(ctx2.WithComponentInstance(inst2), 20)
	rt.EndFrame()

	if legacyKey != "root/state:0" {
		t.Fatalf("expected legacy state to keep path key, got %q", legacyKey)
	}
	if hookKey == legacyKey || hookKey == "" {
		t.Fatalf("expected hook slot state to use independent key, got %q", hookKey)
	}
	if legacy2.Value() != 11 {
		t.Fatalf("expected legacy state to persist 11, got %d", legacy2.Value())
	}
	if hook2.Value() != 22 {
		t.Fatalf("expected hook slot state to persist 22, got %d", hook2.Value())
	}
}

func TestLegacyEffectAndHookSlotEffectCoexist(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "MixedEffect", Key: "main"}
	legacyRuns := 0
	hookRuns := 0
	legacyCleanups := 0
	hookCleanups := 0

	for range 2 {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		state.UseEffect(ctx, func() func() {
			legacyRuns++
			return func() { legacyCleanups++ }
		})
		inst := rt.HookStore().BeginInstance(identity)
		UseEffect(ctx.WithComponentInstance(inst), func() func() {
			hookRuns++
			return func() { hookCleanups++ }
		})
		rt.EndFrame()
	}

	if legacyRuns != 2 || hookRuns != 2 {
		t.Fatalf("expected both effects to run twice, legacy=%d hook=%d", legacyRuns, hookRuns)
	}
	if legacyCleanups != 1 || hookCleanups != 1 {
		t.Fatalf("expected both effects to cleanup before rerun once, legacy=%d hook=%d", legacyCleanups, hookCleanups)
	}

	rt.BeginFrame()
	rt.EndFrame()
	if legacyCleanups != 2 || hookCleanups != 2 {
		t.Fatalf("expected both effects to cleanup on unmount, legacy=%d hook=%d", legacyCleanups, hookCleanups)
	}
}

func TestProviderKeepsLegacyMemoAndPersistentStable(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	key := ContextKey[string]{Default: "fallback"}
	memoCalls := 0
	persistentCalls := 0
	seen := ""
	memoValue := 0
	persistentValue := ""
	nextKey := ""

	rt.BeginFrame()
	renderElementWithContext(ctx, Provider(key, "provided", captureElement(func(ctx *Context) {
		seen = UseContext(ctx, key)
		memoValue = ctx.Memo("memo", func() any {
			memoCalls++
			return 42
		}).(int)
		persistentValue = ctx.Persistent("legacy-persistent", func() any {
			persistentCalls++
			return "stable"
		}).(string)
		nextKey = ctx.NextKey("state")
	})))
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	renderElementWithContext(ctx2, Provider(key, "provided", captureElement(func(ctx *Context) {
		_ = ctx.Memo("memo", func() any {
			memoCalls++
			return 99
		}).(int)
		_ = ctx.Persistent("legacy-persistent", func() any {
			persistentCalls++
			return "changed"
		}).(string)
		_ = ctx.NextKey("state")
	})))
	rt.EndFrame()

	if seen != "provided" {
		t.Fatalf("expected provider value, got %q", seen)
	}
	if memoValue != 42 || memoCalls != 1 {
		t.Fatalf("expected stable legacy memo value/calls, value=%d calls=%d", memoValue, memoCalls)
	}
	if persistentValue != "stable" || persistentCalls != 1 {
		t.Fatalf("expected stable persistent value/calls, value=%q calls=%d", persistentValue, persistentCalls)
	}
	if nextKey != "root/state:1" {
		t.Fatalf("expected provider not to disturb legacy hook index, got %q", nextKey)
	}
}

func TestPhase2HookSlotProviderAndLegacyIntegration(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "Phase2Integration", Key: "main"}
	contextKey := ContextKey[string]{Default: "default"}
	legacyEffects := 0
	hookEffects := 0
	hookCleanups := 0
	legacyKey := ""
	hookKey := ""
	seenContext := ""

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	renderElementWithContext(ctx, Provider(contextKey, "provided", captureElement(func(ctx *Context) {
		seenContext = UseContext(ctx, contextKey)
		legacy := state.UseWithInitial(ctx, 100)
		legacy.Set(101)
		legacyKey = legacy.Key()
		state.UseEffectWithDeps(ctx, []any{"legacy"}, func() func() {
			legacyEffects++
			return nil
		})

		inst := rt.HookStore().BeginInstance(identity)
		hookCtx := ctx.WithComponentInstance(inst)
		hookState := UseState(hookCtx, 200)
		hookState.Set(202)
		hookKey = hookState.Key()
		UseMount(hookCtx, func() func() {
			hookEffects++
			return func() { hookCleanups++ }
		})
	})))
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	renderElementWithContext(ctx2, Provider(contextKey, "provided", captureElement(func(ctx *Context) {
		legacy := state.UseWithInitial(ctx, 100)
		if legacy.Value() != 101 {
			t.Fatalf("expected legacy state 101, got %d", legacy.Value())
		}
		state.UseEffectWithDeps(ctx, []any{"legacy"}, func() func() {
			legacyEffects++
			return nil
		})

		inst := rt.HookStore().BeginInstance(identity)
		hookCtx := ctx.WithComponentInstance(inst)
		hookState := UseState(hookCtx, 200)
		if hookState.Value() != 202 {
			t.Fatalf("expected hook slot state 202, got %d", hookState.Value())
		}
		UseMount(hookCtx, func() func() {
			hookEffects++
			return func() { hookCleanups++ }
		})
	})))
	rt.EndFrame()

	rt.BeginFrame()
	rt.EndFrame()

	if seenContext != "provided" {
		t.Fatalf("expected provided context, got %q", seenContext)
	}
	if legacyKey != "root/state:0" {
		t.Fatalf("expected legacy state path key, got %q", legacyKey)
	}
	if hookKey == "" || hookKey == legacyKey {
		t.Fatalf("expected independent hook slot state key, got %q", hookKey)
	}
	if legacyEffects != 1 {
		t.Fatalf("expected legacy deps effect to run once, got %d", legacyEffects)
	}
	if hookEffects != 1 {
		t.Fatalf("expected hook mount effect to run once, got %d", hookEffects)
	}
	if hookCleanups != 1 {
		t.Fatalf("expected hook cleanup after instance unmount, got %d", hookCleanups)
	}
}

func TestFromWidgetAndKeyPreserveTreeStructure(t *testing.T) {
	w := widgetpkg.Text("hello")
	el := FromWidget(w)
	if renderElement(el) == nil {
		t.Fatal("expected element to retain wrapped widget")
	}
	if renderElement(Key("demo", el)) == nil {
		t.Fatal("expected key wrapper to preserve child widget in phase 1")
	}
}

func TestFragmentCollectsChildren(t *testing.T) {
	frag := Fragment(FromWidget(widgetpkg.Text("a")), FromWidget(widgetpkg.Text("b")))
	if RenderElement(frag) == nil {
		t.Fatal("expected fragment to produce a widget for phase 1")
	}
}

func TestElementKeyExtraction(t *testing.T) {
	child := FromWidget(widgetpkg.Text("a"))
	if got := ElementKey(child); got != "" {
		t.Fatalf("expected empty key for host element, got %q", got)
	}

	keyed := Key("todo-1", child)
	if got := ElementKey(keyed); got != "todo-1" {
		t.Fatalf("expected extracted key todo-1, got %q", got)
	}
}

func TestElementInfoReflectsIdentity(t *testing.T) {
	host := FromWidget(widgetpkg.Text("hello"))
	hostInfo := ElementInfo(host)
	if hostInfo.Kind != "host" {
		t.Fatalf("expected host kind, got %q", hostInfo.Kind)
	}
	if hostInfo.Key != "" {
		t.Fatalf("expected empty host key, got %q", hostInfo.Key)
	}

	frag := Fragment(FromWidget(widgetpkg.Text("a")), FromWidget(widgetpkg.Text("b")))
	fragInfo := ElementInfo(frag)
	if fragInfo.Kind != "fragment" {
		t.Fatalf("expected fragment kind, got %q", fragInfo.Kind)
	}
	if fragInfo.ChildCount != 2 {
		t.Fatalf("expected fragment child count 2, got %d", fragInfo.ChildCount)
	}

	keyed := Key("todo-1", host)
	keyedInfo := ElementInfo(keyed)
	if keyedInfo.Key != "todo-1" {
		t.Fatalf("expected keyed element key todo-1, got %q", keyedInfo.Key)
	}
	if keyedInfo.Kind != "host" {
		t.Fatalf("expected keyed child kind host, got %q", keyedInfo.Kind)
	}
}

func TestElementBoundarySafety(t *testing.T) {
	if RenderElement(nil) != nil {
		t.Fatal("expected nil element to render as nil")
	}
	if info := ElementInfo(nil); info != (ElementIdentity{}) {
		t.Fatalf("expected zero identity for nil element, got %#v", info)
	}
	if got := ElementKey(nil); got != "" {
		t.Fatalf("expected empty key for nil element, got %q", got)
	}
}

func TestStatelessDisplayElementsWrapLegacyWidgets(t *testing.T) {
	if el := TextElement("hello"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected TextElement to wrap a legacy text widget")
	}
	if el := SpacerElement(8, 12); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SpacerElement to wrap a legacy spacer widget")
	}
	if el := DividerElement(); el == nil || RenderElement(el) == nil {
		t.Fatal("expected DividerElement to wrap a legacy divider widget")
	}
}

func TestLayoutElementsWrapLegacyWidgets(t *testing.T) {
	if el := ColumnElement(TextElement("a"), TextElement("b")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ColumnElement to wrap a legacy column widget")
	}
	if el := RowElement(TextElement("a"), TextElement("b")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected RowElement to wrap a legacy row widget")
	}
	if el := StackElement(TextElement("a"), TextElement("b")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected StackElement to wrap a legacy stack widget")
	}
	if el := CenterElement(TextElement("a")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected CenterElement to wrap a legacy center widget")
	}
	if el := PaddingElement(All(4), TextElement("a")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected PaddingElement to wrap a legacy padding widget")
	}
	if el := ContainerElement(Style{}, TextElement("a")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ContainerElement to wrap a legacy container widget")
	}
}

func TestInteractiveElementsWrapLegacyWidgets(t *testing.T) {
	if el := ButtonElement(TextElement("go")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ButtonElement to wrap a legacy button widget")
	}
	if el := ClickAreaElement(TextElement("tap"), nil); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ClickAreaElement to wrap a legacy click area widget")
	}
	if el := CheckboxElement("check", true); el == nil || RenderElement(el) == nil {
		t.Fatal("expected CheckboxElement to wrap a legacy checkbox widget")
	}
	if el := SwitchElement(true); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SwitchElement to wrap a legacy switch widget")
	}
}

func TestFinalComponentElementsWrapLegacyWidgets(t *testing.T) {
	if el := TextFieldElement("hello"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected TextFieldElement to wrap a legacy text field widget")
	}
	if el := SliderElement(42); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SliderElement to wrap a legacy slider widget")
	}
	if el := RadioGroupElement("a", []RadioItem{{Label: "A", Value: "a"}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected RadioGroupElement to wrap a legacy radio group widget")
	}
	if el := SelectElement("a", []SelectOptionItem[string]{{Label: "A", Value: "a"}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SelectElement to wrap a legacy select widget")
	}
	if el := MenuElement([]MenuItem{{Key: "a", Label: "A"}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected MenuElement to wrap a menu widget")
	}
	if el := DropdownMenuElement(true, TextElement("open"), []MenuItem{{Key: "a", Label: "A"}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected DropdownMenuElement to wrap a dropdown menu widget")
	}
	if el := ListItemElement("Headline"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ListItemElement to wrap a list item widget")
	}
	if el := ListItemElementWithSlots(TextElement("Headline"), TextElement("Supporting"), IconElement("I"), TextElement("T")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ListItemElementWithSlots to wrap a list item widget")
	}
	if el := IconButtonElement(IconElement("I")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected IconButtonElement to wrap an icon button widget")
	}
	if el := FilledIconButtonElement(IconElement("I")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected FilledIconButtonElement to wrap an icon button widget")
	}
	if el := FilledTonalIconButtonElement(IconElement("I")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected FilledTonalIconButtonElement to wrap an icon button widget")
	}
	if el := OutlinedIconButtonElement(IconElement("I")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected OutlinedIconButtonElement to wrap an icon button widget")
	}
	if el := FloatingActionButtonElement(IconElement("+")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected FloatingActionButtonElement to wrap a FAB widget")
	}
	if el := SmallFloatingActionButtonElement(IconElement("+")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SmallFloatingActionButtonElement to wrap a FAB widget")
	}
	if el := LargeFloatingActionButtonElement(IconElement("+")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected LargeFloatingActionButtonElement to wrap a FAB widget")
	}
	if el := ExtendedFloatingActionButtonElement(IconElement("+"), TextElement("Create")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ExtendedFloatingActionButtonElement to wrap a FAB widget")
	}
	if el := ProgressBarElement(25); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ProgressBarElement to wrap a legacy progress bar widget")
	}
	if el := CircularProgressElement(25); el == nil || RenderElement(el) == nil {
		t.Fatal("expected CircularProgressElement to wrap a legacy circular progress widget")
	}
	if el := ImageElement(ImageSource{Label: "placeholder"}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ImageElement to wrap a legacy image widget")
	}
	if el := IconElement("star"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected IconElement to wrap a legacy icon widget")
	}
	if el := CardElement(TextElement("card")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected CardElement to wrap a legacy card widget")
	}
	if el := TabsElement("home", []TabItem{{Key: "home", Label: "Home"}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected TabsElement to wrap a legacy tabs widget")
	}
	if el := DialogElement(true, TextElement("dialog")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected DialogElement to wrap a legacy dialog widget")
	}
	if el := PopupElement(true, TextElement("popup")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected PopupElement to wrap a legacy popup widget")
	}
	if el := ToastElement("saved"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ToastElement to wrap a legacy toast widget")
	}
	if el := SnackbarElement("saved", SnackbarAction("Undo", func(ctx *Context) {})); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SnackbarElement to wrap a snackbar widget")
	}
	if el := TooltipElement("tip", TextElement("hover")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected TooltipElement to wrap a tooltip widget")
	}
	if el := BadgeElement(IconButtonElement(IconElement("M")), "3"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected BadgeElement to wrap a badge widget")
	}
	if el := AssistChipElement("Assist"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected AssistChipElement to wrap a chip widget")
	}
	if el := FilterChipElement("Filter", ChipSelected(true)); el == nil || RenderElement(el) == nil {
		t.Fatal("expected FilterChipElement to wrap a chip widget")
	}
	if el := InputChipElement("Input"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected InputChipElement to wrap a chip widget")
	}
	if el := SuggestionChipElement("Suggestion"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SuggestionChipElement to wrap a chip widget")
	}
	if el := ChipElementWithSlots(TextElement("Custom")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ChipElementWithSlots to wrap a chip widget")
	}
	if el := SearchBarElement("query"); el == nil || RenderElement(el) == nil {
		t.Fatal("expected SearchBarElement to wrap a search bar widget")
	}
	if el := LinearProgressIndicatorElement(25); el == nil || RenderElement(el) == nil {
		t.Fatal("expected LinearProgressIndicatorElement to wrap a progress indicator widget")
	}
	if el := CircularProgressIndicatorElement(25); el == nil || RenderElement(el) == nil {
		t.Fatal("expected CircularProgressIndicatorElement to wrap a progress indicator widget")
	}
	if el := ScrollViewElement(TextElement("scroll")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ScrollViewElement to wrap a legacy scroll view widget")
	}
	if el := ListViewElement(1, func(ctx *Context, index int) Element { return TextElement("item") }); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ListViewElement to wrap a legacy list view widget")
	}
	if el := GridElement(2, TextElement("a"), TextElement("b")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected GridElement to wrap a legacy grid widget")
	}
	if el := GridViewElement(1, 2, func(ctx *Context, index int) Element { return TextElement("cell") }); el == nil || RenderElement(el) == nil {
		t.Fatal("expected GridViewElement to wrap a legacy grid view widget")
	}
	if el := AppBarElement(
		TextElement("Title"),
		AppBarHeight(48),
	); el == nil || RenderElement(el) == nil {
		t.Fatal("expected AppBarElement to wrap a legacy app bar widget")
	}
	if el := AppBarElementWithSlots(
		TextElement("Title"),
		TextElement("Menu"),
		[]Element{TextElement("Save")},
		AppBarHeight(48),
	); el == nil || RenderElement(el) == nil {
		t.Fatal("expected AppBarElementWithSlots to wrap a legacy app bar widget")
	}
	if el := BottomNavigationElement("home", []ElementNavItem{{Key: "home", Label: "Home", Icon: IconElement("home")}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected BottomNavigationElement to wrap a legacy bottom navigation widget")
	}
	if el := NavigationRailElement("home", []ElementNavItem{{Key: "home", Label: "Home", Icon: IconElement("home")}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected NavigationRailElement to wrap a navigation rail widget")
	}
	if el := NavigationDrawerElement("home", []ElementNavItem{{Key: "home", Label: "Home", Icon: IconElement("home")}}); el == nil || RenderElement(el) == nil {
		t.Fatal("expected NavigationDrawerElement to wrap a navigation drawer widget")
	}
	if el := WithFontElement(DefaultFontSpec(), TextElement("font")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected WithFontElement to wrap a legacy font scope widget")
	}
	if el := FlexedElement(2, TextElement("flex")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected FlexedElement to wrap a legacy flexed widget")
	}
	if el := ExpandedElement(TextElement("expanded")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected ExpandedElement to wrap a legacy expanded widget")
	}
	if el := FixedSizeElement(20, 10, TextElement("fixed")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected FixedSizeElement to wrap a legacy fixed-size widget")
	}
	if el := FillElement(TextElement("fill")); el == nil || RenderElement(el) == nil {
		t.Fatal("expected FillElement to wrap a legacy fill widget")
	}
}

func TestCompositeElementChildrenCanUseHooksAndContext(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	contextKey := ContextKey[string]{Default: "default"}
	seenContext := make([]string, 0, 2)
	seenState := make([]int, 0, 2)

	child := func(ctx *Context) Element {
		state := UseState(ctx, 1)
		seenContext = append(seenContext, UseContext(ctx, contextKey))
		seenState = append(seenState, state.Value())
		state.Set(state.Value() + 1)
		return TextElement("child")
	}
	root := func(ctx *Context) Element {
		return Provider(contextKey, "provided", RowElement(
			ButtonElement(ComponentElement(child)),
		))
	}

	for range 2 {
		rt.BeginFrame()
		r.Render(internal.NewContext(gtx, rt), root)
		rt.EndFrame()
	}

	if len(seenContext) != 2 || seenContext[0] != "provided" || seenContext[1] != "provided" {
		t.Fatalf("expected provider context inside composite children, got %v", seenContext)
	}
	if len(seenState) != 2 || seenState[0] != 1 || seenState[1] != 2 {
		t.Fatalf("expected nested component state to persist, got %v", seenState)
	}
}

func TestTabbedAnimatedSectionsKeepIndependentHookSlots(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	gtx := gioLayout.Context{Ops: &ops, Now: now}
	r := newReconciler()
	tab := "easing"

	root := func(ctx *Context) Element {
		var content Element
		switch tab {
		case "value":
			content = ComponentElement(func(ctx *Context) Element {
				target := UseState(ctx, float32(200))
				animated := UseAnimatedValue(ctx, target.Value(), 100*time.Millisecond, EaseOutBack)
				return FixedWidthElement(animated, TextElement("value"))
			})
		case "deco":
			content = ComponentElement(func(ctx *Context) Element {
				toggle := UseState(ctx, true)
				bg := color.NRGBA{R: 59, G: 130, B: 246, A: 255}
				if !toggle.Value() {
					bg = color.NRGBA{R: 234, G: 88, B: 12, A: 255}
				}
				deco := UseAnimatedDecoration(ctx, Bg(bg).WithRad(8), 100*time.Millisecond, EaseInOut)
				return ContainerDecorationElement(deco, TextElement("deco"))
			})
		case "pulse":
			content = ComponentElement(func(ctx *Context) Element {
				active := UseState(ctx, false)
				target := float32(0)
				if active.Value() {
					target = 1
				}
				pulse := UseAnimatedValue(ctx, target, 100*time.Millisecond, EaseInOut)
				return TextElement(fmt.Sprintf("%.2f", pulse))
			})
		default:
			content = ComponentElement(func(ctx *Context) Element {
				playing := UseState(ctx, false)
				target := float32(0)
				if playing.Value() {
					target = 1
				}
				for range 6 {
					UseAnimatedValue(ctx, target, 100*time.Millisecond, anim.EaseOutElastic)
				}
				return TextElement("easing")
			})
		}
		return ColumnElement(content)
	}

	for idx, nextTab := range []string{"easing", "value", "deco", "pulse", "easing", "deco"} {
		tab = nextTab
		rt.BeginFrame()
		gtx.Now = now.Add(time.Duration(idx) * 16 * time.Millisecond)
		r.Render(internal.NewContext(gtx, rt), root)
		rt.EndFrame()
	}
}

func TestUseAnimatedValueZeroDurationResetsHookState(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	gtx := gioLayout.Context{Ops: &ops, Now: now}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "AnimValue", Key: "main"}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	if got := UseAnimatedValue(ctx.WithComponentInstance(inst), float32(10), 100*time.Millisecond, EaseOut); got != 10 {
		t.Fatalf("initial animated value = %v, want 10", got)
	}
	rt.EndFrame()

	rt.BeginFrame()
	gtx.Now = now.Add(50 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	if got := UseAnimatedValue(ctx.WithComponentInstance(inst), float32(30), 0, EaseOut); got != 30 {
		t.Fatalf("zero duration animated value = %v, want 30", got)
	}
	rt.EndFrame()

	rt.BeginFrame()
	gtx.Now = now.Add(60 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	if got := UseAnimatedValue(ctx.WithComponentInstance(inst), float32(40), 100*time.Millisecond, EaseOut); got != 30 {
		t.Fatalf("animation after zero-duration snap should start at 30, got %v", got)
	}
	rt.EndFrame()
}

func TestUseAnimatedDecorationZeroDurationResetsHookState(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	gtx := gioLayout.Context{Ops: &ops, Now: now}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "AnimDeco", Key: "main"}
	blue := Bg(color.NRGBA{R: 59, G: 130, B: 246, A: 255}).WithRad(8)
	orange := Bg(color.NRGBA{R: 234, G: 88, B: 12, A: 255}).WithRad(24)
	green := Bg(color.NRGBA{R: 22, G: 163, B: 74, A: 255}).WithRad(40)

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	_ = UseAnimatedDecoration(ctx.WithComponentInstance(inst), blue, 100*time.Millisecond, EaseOut)
	rt.EndFrame()

	rt.BeginFrame()
	gtx.Now = now.Add(50 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	if got := UseAnimatedDecoration(ctx.WithComponentInstance(inst), orange, 0, EaseOut); !anim.DecorationEqual(got, orange) {
		t.Fatalf("zero duration decoration should return target")
	}
	rt.EndFrame()

	rt.BeginFrame()
	gtx.Now = now.Add(60 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	got := UseAnimatedDecoration(ctx.WithComponentInstance(inst), green, 100*time.Millisecond, EaseOut)
	if got.Background == nil || *got.Background != *orange.Background {
		t.Fatalf("animation after zero-duration snap should start from orange, got %#v", got.Background)
	}
	rt.EndFrame()
}

func TestUseAnimatedValueTargetChangeMidAnimation(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	gtx := gioLayout.Context{Ops: &ops, Now: now}
	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "AnimMid", Key: "main"}

	// Frame 1: start animating 0→100 over 100ms
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	inst := rt.HookStore().BeginInstance(identity)
	_ = UseAnimatedValue(ctx.WithComponentInstance(inst), float32(0), 100*time.Millisecond, EaseOut)
	rt.EndFrame()

	// Frame 2: 50ms elapsed, value should be 0 (still same target)
	rt.BeginFrame()
	gtx.Now = now.Add(50 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	v := UseAnimatedValue(ctx.WithComponentInstance(inst), float32(0), 100*time.Millisecond, EaseOut)
	if v != 0 {
		t.Fatalf("at t=50ms same target, expected 0, got %v", v)
	}
	rt.EndFrame()

	// Frame 3: change target to 100, animation starts
	rt.BeginFrame()
	gtx.Now = now.Add(60 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	v60 := UseAnimatedValue(ctx.WithComponentInstance(inst), float32(100), 100*time.Millisecond, EaseOut)
	if v60 > 5 {
		t.Fatalf("at start of new animation, expected near 0, got %v", v60)
	}
	rt.EndFrame()

	// Frame 4: 50ms after target change (t=110ms total), easing progress
	rt.BeginFrame()
	gtx.Now = now.Add(110 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	v110 := UseAnimatedValue(ctx.WithComponentInstance(inst), float32(100), 100*time.Millisecond, EaseOut)
	if v110 < 30 || v110 > 80 {
		t.Fatalf("at t=110ms (50ms after switch), expected 30-80, got %v", v110)
	}
	rt.EndFrame()

	// Frame 5: change target to 200 mid-animation, from should be current value
	rt.BeginFrame()
	gtx.Now = now.Add(130 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	_ = UseAnimatedValue(ctx.WithComponentInstance(inst), float32(200), 100*time.Millisecond, EaseOut)
	rt.EndFrame()

	// Frame 6: 40ms after second target change, should progress toward 200 from mid-point
	rt.BeginFrame()
	gtx.Now = now.Add(170 * time.Millisecond)
	ctx = internal.NewContext(gtx, rt)
	inst = rt.HookStore().BeginInstance(identity)
	v170 := UseAnimatedValue(ctx.WithComponentInstance(inst), float32(200), 100*time.Millisecond, EaseOut)
	if v170 < 70 || v170 > 170 {
		t.Fatalf("40ms after second target change, expected 70-170, got %v", v170)
	}
	rt.EndFrame()
}

func TestKeyedComponentsInsideCompositeElementsPreserveStateAcrossReorder(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	items := []string{"a", "b", "c"}
	seen := map[string][]int{}

	child := func(id string) Component {
		return func(ctx *Context) Element {
			state := UseState(ctx, 1)
			seen[id] = append(seen[id], state.Value())
			state.Set(state.Value() + 1)
			return TextElement(id)
		}
	}
	root := func(ctx *Context) Element {
		children := make([]Element, 0, len(items))
		for _, id := range items {
			children = append(children, Key(id, ButtonElement(ComponentElement(child(id)))))
		}
		return RowElement(children...)
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	items = []string{"c", "a", "b"}
	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	for _, id := range []string{"a", "b", "c"} {
		values := seen[id]
		if len(values) != 2 || values[0] != 1 || values[1] != 2 {
			t.Fatalf("expected keyed composite child %s to preserve state across reorder, got %v", id, values)
		}
	}
}

func TestRouterElementWrapsComponentRoutes(t *testing.T) {
	if el := RouterElement(
		RouteElement("/", func(ctx *Context) Element { return TextElement("home") }),
	); el == nil {
		t.Fatal("expected RouterElement to return an element")
	}
}

func TestRouteElementKeyOption(t *testing.T) {
	route := RouteElement("/users/:id", func(ctx *Context) Element { return nil }, RouteKey("user-route"))
	if route.Key != "user-route" {
		t.Fatalf("expected explicit route key, got %q", route.Key)
	}
}

type captureElement func(ctx *Context)

func (e captureElement) render() widgetpkg.Widget {
	return nil
}

func (e captureElement) renderWithContext(ctx *Context) widgetpkg.Widget {
	if e != nil {
		e(ctx)
	}
	return nil
}
