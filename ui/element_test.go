package ui

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/state"
	widgetpkg "github.com/xiaowumin-mark/FluxUI/widget"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestUseStateWithInitial(t *testing.T) {
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
