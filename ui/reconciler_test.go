package ui

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	widgetpkg "github.com/xiaowumin-mark/FluxUI/widget"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestReconcilerManagesRootComponentInstance(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	values := make([]int, 0, 2)
	keys := make([]string, 0, 2)

	root := func(ctx *Context) Element {
		state := UseState(ctx, 1)
		values = append(values, state.Value())
		keys = append(keys, state.Key())
		state.Set(state.Value() + 1)
		return nil
	}

	for range 2 {
		rt.BeginFrame()
		r.Render(internal.NewContext(gtx, rt), root)
		rt.EndFrame()
	}

	if len(values) != 2 || values[0] != 1 || values[1] != 2 {
		t.Fatalf("expected reconciler-managed root state values [1 2], got %v", values)
	}
	if len(keys) != 2 || keys[0] == "root/state:0" || keys[0] == "" || keys[0] != keys[1] {
		t.Fatalf("expected stable HookSlot state key, got %v", keys)
	}
	if r.root == nil || r.root.Instance == nil {
		t.Fatal("expected root fiber to retain component instance")
	}
}

func TestReconcilerRootUnmountCleanup(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	cleanups := 0

	root := func(ctx *Context) Element {
		UseMount(ctx, func() func() {
			return func() { cleanups++ }
		})
		return nil
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	rt.BeginFrame()
	rt.EndFrame()

	if cleanups != 1 {
		t.Fatalf("expected root component cleanup after unmount, got %d", cleanups)
	}
}

func TestReconcilerRootProviderContext(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	key := ContextKey[string]{Default: "default"}
	seen := ""

	root := func(ctx *Context) Element {
		return Provider(key, "provided", captureElement(func(ctx *Context) {
			seen = UseContext(ctx, key)
		}))
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	if seen != "provided" {
		t.Fatalf("expected provider value through reconciler root render, got %q", seen)
	}
}

func TestReconcilerKeyedChildrenPreserveStateAcrossReorder(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	items := []string{"a", "b", "c"}
	seen := map[string][]int{}

	itemComponent := func(id string) Component {
		return func(ctx *Context) Element {
			state := UseState(ctx, id)
			seen[id] = append(seen[id], len(state.Value()))
			state.Set(state.Value() + "+")
			return nil
		}
	}
	root := func(ctx *Context) Element {
		children := make([]Element, 0, len(items))
		for _, id := range items {
			children = append(children, Key(id, ComponentElement(itemComponent(id))))
		}
		return Fragment(children...)
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
			t.Fatalf("expected keyed child %s to preserve state across reorder, got %v", id, values)
		}
	}
}

func TestReconcilerUnkeyedChildrenUseIndexFallback(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	items := []string{"a", "b"}
	seen := map[string]string{}

	itemComponent := func(id string) Component {
		return func(ctx *Context) Element {
			state := UseState(ctx, id)
			seen[id] = state.Value()
			return nil
		}
	}
	root := func(ctx *Context) Element {
		children := make([]Element, 0, len(items))
		for _, id := range items {
			children = append(children, ComponentElement(itemComponent(id)))
		}
		return Fragment(children...)
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	items = []string{"b", "a"}
	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	if seen["b"] != "a" || seen["a"] != "b" {
		t.Fatalf("expected unkeyed children to follow index fallback, got seen=%v", seen)
	}
}

func TestReconcilerHostProviderAndNestedComponentAdapterPath(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	contextKey := ContextKey[string]{Default: "default"}
	host := &layoutCaptureWidget{}
	seenContext := ""
	seenState := 0

	child := func(ctx *Context) Element {
		seenContext = UseContext(ctx, contextKey)
		state := UseState(ctx, 7)
		seenState = state.Value()
		state.Set(8)
		return FromWidget(host)
	}
	root := func(ctx *Context) Element {
		return Fragment(
			Provider(contextKey, "provided", Key("child", ComponentElement(child))),
			FromWidget(widgetpkg.Text("legacy-host")),
		)
	}

	rt.BeginFrame()
	w := r.Render(internal.NewContext(gtx, rt), root)
	if w == nil {
		t.Fatal("expected reconciler to return host adapter widget")
	}
	w.Layout(internal.NewContext(gtx, rt))
	rt.EndFrame()

	if seenContext != "provided" {
		t.Fatalf("expected provider value in nested component, got %q", seenContext)
	}
	if seenState != 7 {
		t.Fatalf("expected nested component hook state initial value 7, got %d", seenState)
	}
	if host.layouts != 1 {
		t.Fatalf("expected legacy host widget to layout once, got %d", host.layouts)
	}
	if r.root == nil || len(r.root.Children) != 2 {
		t.Fatalf("expected root fragment to track two children, got %#v", r.root)
	}
	providerNode := r.root.Children[0]
	if providerNode.Kind != fiberProvider || len(providerNode.Children) != 1 || providerNode.Children[0].Kind != fiberComponent {
		t.Fatalf("expected provider -> component fiber path, got %#v", providerNode)
	}
	if providerNode.Children[0].Children == nil || providerNode.Children[0].Children[0].Kind != fiberHost {
		t.Fatalf("expected nested component to end at host fiber, got %#v", providerNode.Children[0])
	}
	if r.root.Children[1].Kind != fiberHost || r.root.Children[1].Children != nil || r.root.Children[1].Instance != nil {
		t.Fatalf("expected legacy FromWidget sibling to be host leaf, got %#v", r.root.Children[1])
	}
}

func TestReconcilerUnmountCleanupOnKeyChange(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	key := "a"
	cleanups := map[string]int{}

	child := func(id string) Component {
		return func(ctx *Context) Element {
			UseMount(ctx, func() func() {
				return func() { cleanups[id]++ }
			})
			return nil
		}
	}
	root := func(ctx *Context) Element {
		return Key(key, ComponentElement(child(key)))
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	key = "b"
	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	if cleanups["a"] != 1 {
		t.Fatalf("expected old keyed child cleanup once after key change, got %v", cleanups)
	}
	if cleanups["b"] != 0 {
		t.Fatalf("expected new keyed child to remain mounted, got %v", cleanups)
	}
}

func TestReconcilerUnmountCleanupOnTypeChange(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	useAlt := false
	cleanups := map[string]int{}

	first := func(ctx *Context) Element {
		UseMount(ctx, func() func() {
			return func() { cleanups["first"]++ }
		})
		return nil
	}
	second := func(ctx *Context) Element {
		UseMount(ctx, func() func() {
			return func() { cleanups["second"]++ }
		})
		return nil
	}
	root := func(ctx *Context) Element {
		if useAlt {
			return ComponentElement(second)
		}
		return ComponentElement(first)
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	useAlt = true
	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	if cleanups["first"] != 1 {
		t.Fatalf("expected first component cleanup once after type change, got %v", cleanups)
	}
	if cleanups["second"] != 0 {
		t.Fatalf("expected second component to remain mounted, got %v", cleanups)
	}
}

func TestReconcilerUnmountCleanupOnProviderSubtreeDeletion(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	show := true
	contextKey := ContextKey[string]{Default: "default"}
	cleanups := 0

	child := func(ctx *Context) Element {
		UseMount(ctx, func() func() {
			return func() { cleanups++ }
		})
		return nil
	}
	root := func(ctx *Context) Element {
		if !show {
			return Fragment()
		}
		return Fragment(Provider(contextKey, "provided", ComponentElement(child)))
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	show = false
	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	if cleanups != 1 {
		t.Fatalf("expected provider subtree deletion cleanup once, got %d", cleanups)
	}
}

func TestReconcilerUnmountCleanupOnFragmentChildDeletion(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	items := []string{"a", "b"}
	cleanups := map[string]int{}

	child := func(id string) Component {
		return func(ctx *Context) Element {
			UseMount(ctx, func() func() {
				return func() { cleanups[id]++ }
			})
			return nil
		}
	}
	root := func(ctx *Context) Element {
		children := make([]Element, 0, len(items))
		for _, id := range items {
			children = append(children, Key(id, ComponentElement(child(id))))
		}
		return Fragment(children...)
	}

	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	items = []string{"a"}
	rt.BeginFrame()
	r.Render(internal.NewContext(gtx, rt), root)
	rt.EndFrame()

	if cleanups["a"] != 0 || cleanups["b"] != 1 {
		t.Fatalf("expected only deleted fragment child cleanup, got %v", cleanups)
	}
}

func TestReconcilerKeyedDynamicListInsertDeleteReorderKeepsStateByKey(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	items := []string{"a", "b", "c"}
	seen := map[string][]string{}
	cleanups := map[string]int{}

	child := func(id string) Component {
		return func(ctx *Context) Element {
			state := UseState(ctx, id)
			seen[id] = append(seen[id], state.Value())
			state.Set(state.Value() + "+")
			UseMount(ctx, func() func() {
				return func() { cleanups[id]++ }
			})
			return nil
		}
	}
	root := func(ctx *Context) Element {
		children := make([]Element, 0, len(items))
		for _, id := range items {
			children = append(children, Key(id, ComponentElement(child(id))))
		}
		return Fragment(children...)
	}
	render := func() {
		rt.BeginFrame()
		r.Render(internal.NewContext(gtx, rt), root)
		rt.EndFrame()
	}

	render()
	items = []string{"x", "a", "b", "c"}
	render()
	items = []string{"c", "x", "a"}
	render()

	if got := seen["a"]; len(got) != 3 || got[0] != "a" || got[1] != "a+" || got[2] != "a++" {
		t.Fatalf("expected keyed item a state to follow key, got %v", got)
	}
	if got := seen["b"]; len(got) != 2 || got[0] != "b" || got[1] != "b+" {
		t.Fatalf("expected keyed item b state until deletion, got %v", got)
	}
	if got := seen["c"]; len(got) != 3 || got[0] != "c" || got[1] != "c+" || got[2] != "c++" {
		t.Fatalf("expected keyed item c state to survive reorder, got %v", got)
	}
	if got := seen["x"]; len(got) != 2 || got[0] != "x" || got[1] != "x+" {
		t.Fatalf("expected inserted keyed item x to keep own state, got %v", got)
	}
	if cleanups["b"] != 1 {
		t.Fatalf("expected deleted keyed item b cleanup once, got %v", cleanups)
	}
	for _, id := range []string{"a", "c", "x"} {
		if cleanups[id] != 0 {
			t.Fatalf("expected kept keyed item %s to remain mounted, got cleanups=%v", id, cleanups)
		}
	}
}

func TestReconcilerUnkeyedDynamicListInsertDeleteUsesIndexFallback(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	r := newReconciler()
	items := []string{"a", "b", "c"}
	seen := map[string][]string{}

	child := func(id string) Component {
		return func(ctx *Context) Element {
			state := UseState(ctx, id)
			seen[id] = append(seen[id], state.Value())
			state.Set(state.Value() + "+")
			return nil
		}
	}
	root := func(ctx *Context) Element {
		children := make([]Element, 0, len(items))
		for _, id := range items {
			children = append(children, ComponentElement(child(id)))
		}
		return Fragment(children...)
	}
	render := func() {
		rt.BeginFrame()
		r.Render(internal.NewContext(gtx, rt), root)
		rt.EndFrame()
	}

	render()
	items = []string{"x", "a", "b", "c"}
	render()
	items = []string{"c", "x", "a"}
	render()

	if got := seen["x"]; len(got) != 2 || got[0] != "a+" || got[1] != "b++" {
		t.Fatalf("expected inserted unkeyed item x to inherit index state, got %v", got)
	}
	if got := seen["a"]; len(got) != 3 || got[0] != "a" || got[1] != "b+" || got[2] != "c++" {
		t.Fatalf("expected unkeyed item a to follow index fallback, got %v", got)
	}
	if got := seen["c"]; len(got) != 3 || got[0] != "c" || got[1] != "c" || got[2] != "a++" {
		t.Fatalf("expected unkeyed item c to follow index fallback, got %v", got)
	}
}

type layoutCaptureWidget struct {
	layouts int
}

func (w *layoutCaptureWidget) Layout(ctx *internal.Context) layout.Dimensions {
	w.layouts++
	return layout.Dimensions{}
}
