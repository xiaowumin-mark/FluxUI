package ui

import (
	"strings"
	"testing"
	"time"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	fluxlayout "github.com/xiaowumin-mark/FluxUI/layout"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

// r0ContractWidget gives the Element and reconciler contract tests a widget
// that does not depend on a platform renderer.
type r0ContractWidget struct {
	name string
}

func (r0ContractWidget) Layout(*internal.Context) fluxlayout.Dimensions {
	return fluxlayout.Dimensions{}
}

type r0ContextProbeElement struct {
	widget    widget.Widget
	identity_ ElementIdentity
	onContext func(*Context)
}

func (e r0ContextProbeElement) render() widget.Widget {
	return e.widget
}

func (e r0ContextProbeElement) identity() ElementIdentity {
	return e.identity_
}

func (e r0ContextProbeElement) renderWithContext(ctx *Context) widget.Widget {
	if e.onContext != nil {
		e.onContext(ctx)
	}
	return e.widget
}

type r0PlainElement struct {
	widget widget.Widget
}

func (e r0PlainElement) render() widget.Widget {
	return e.widget
}

func r0ContractContext(runtime *internal.Runtime) *Context {
	return internal.NewContext(gioLayout.Context{Ops: new(op.Ops)}, runtime)
}

func TestR0ElementIdentityFacadeContracts(t *testing.T) {
	leafWidget := r0ContractWidget{name: "leaf"}
	leaf := r0ContextProbeElement{
		widget:    leafWidget,
		identity_: ElementIdentity{Kind: "leaf", Key: "child-key"},
	}
	provider := providerElement[string]{
		key:   NewContextKey("fallback"),
		value: "provided",
		child: leaf,
	}

	if got := provider.render(); got != leafWidget {
		t.Fatalf("provider render = %#v, want child widget", got)
	}
	if got, want := provider.identity(), (ElementIdentity{Kind: "provider", Key: "child-key", ChildCount: 1}); got != want {
		t.Fatalf("provider identity = %#v, want %#v", got, want)
	}
	if got := (providerElement[string]{}).render(); got != nil {
		t.Fatalf("empty provider render = %#v, want nil", got)
	}
	if got := (providerElement[string]{child: r0PlainElement{widget: leafWidget}}).identity().Key; got != "" {
		t.Fatalf("provider copied a key from a non-identifiable child: %q", got)
	}

	keyed := keyElement{key: "explicit-key", child: leaf}
	if got := keyed.render(); got != leafWidget {
		t.Fatalf("keyed render = %#v, want child widget", got)
	}
	if got := (keyElement{}).render(); got != nil {
		t.Fatalf("empty keyed render = %#v, want nil", got)
	}
	if got, want := keyed.identity(), (ElementIdentity{Kind: "leaf", Key: "explicit-key", ChildCount: 1}); got != want {
		t.Fatalf("keyed identity = %#v, want %#v", got, want)
	}

	component := componentElement{component: func(*Context) Element { return FromWidget(leafWidget) }}
	if got := component.render(); got != nil {
		t.Fatalf("component placeholder render = %#v, want nil", got)
	}
	if got, want := component.identity(), (ElementIdentity{Kind: "component", ChildCount: 1}); got != want {
		t.Fatalf("component identity = %#v, want %#v", got, want)
	}
	if component.Component() == nil {
		t.Fatal("component accessor lost its function")
	}

	if got := elementFromWidget(nil); got != nil {
		t.Fatalf("elementFromWidget(nil) = %#v, want nil", got)
	}
	if got := elementFromWidget(leafWidget); ElementInfo(got).Kind != "host" {
		t.Fatalf("elementFromWidget identity = %#v, want host", ElementInfo(got))
	}
	if got := componentTypeID(nil); got != "component" {
		t.Fatalf("nil component type ID = %q, want component", got)
	}
	if got := componentTypeID(component.Component()); !strings.Contains(got, "@") {
		t.Fatalf("component type ID %q does not include a function identity", got)
	}
}

func TestR0ElementContextRenderingContracts(t *testing.T) {
	key := NewContextKey("fallback")
	seenValue := ""
	leafWidget := r0ContractWidget{name: "context-leaf"}
	probe := r0ContextProbeElement{
		widget: leafWidget,
		onContext: func(ctx *Context) {
			seenValue = UseContext(ctx, key)
		},
	}
	ctx := r0ContractContext(nil)

	if got := renderElementWithContext(ctx, providerElement[string]{key: key, value: "provided", child: probe}); got != leafWidget {
		t.Fatalf("provider context render = %#v, want child widget", got)
	}
	if seenValue != "provided" {
		t.Fatalf("provider context value = %q, want provided", seenValue)
	}

	seenValue = ""
	if got := (keyElement{key: "item", child: probe}).renderWithContext(ctx); got != leafWidget {
		t.Fatalf("keyed context render = %#v, want child widget", got)
	}
	if seenValue != "fallback" {
		t.Fatalf("keyed context value = %q, want fallback", seenValue)
	}
	if got := (fragmentElement{children: []Element{probe, nil}}).renderWithContext(ctx); got == nil {
		t.Fatal("fragment context render returned nil for a non-empty fragment")
	}

	if got := renderComponent(nil, "root", 0, "", func(*Context) Element { return FromWidget(leafWidget) }); got != nil {
		t.Fatalf("renderComponent with nil context = %#v, want nil", got)
	}
	plainCtx := r0ContractContext(nil)
	if got := renderComponent(plainCtx, "root", 1, "plain", func(componentCtx *Context) Element {
		if componentCtx != plainCtx {
			t.Fatal("component without a runtime should retain its context")
		}
		return FromWidget(leafWidget)
	}); got == nil {
		t.Fatal("renderComponent without a runtime returned nil")
	}

	runtime := internal.NewRuntime(nil)
	runtime.BeginFrame()
	instanceBound := false
	if got := renderComponent(r0ContractContext(runtime), "root", 2, "runtime", func(componentCtx *Context) Element {
		instanceBound = componentCtx.ComponentInstance() != nil
		return FromWidget(leafWidget)
	}); got == nil {
		t.Fatal("renderComponent with a runtime returned nil")
	}
	runtime.EndFrame()
	if !instanceBound {
		t.Fatal("renderComponent did not bind a component instance")
	}

	builder := ElementRootBuilder(func(*Context) Element { return FromWidget(leafWidget) })
	if got := builder(nil); got != nil {
		t.Fatalf("ElementRootBuilder(nil) = %#v, want nil", got)
	}
	runtime.BeginFrame()
	if got := builder(r0ContractContext(runtime)); got == nil {
		t.Fatal("ElementRootBuilder returned nil for a host component")
	}
	runtime.EndFrame()
}

func TestR0ReconcilerKeyedElementDispatchContract(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	runtime.BeginFrame()
	defer runtime.EndFrame()

	r := newReconciler()
	ctx := r0ContractContext(runtime)
	host := FromWidget(r0ContractWidget{name: "host"})
	componentCalls := 0
	component := Component(func(*Context) Element {
		componentCalls++
		return host
	})
	providerKey := NewContextKey("")
	cases := []struct {
		name  string
		child Element
	}{
		{name: "component", child: ComponentElement(component)},
		{name: "fragment", child: Fragment(host)},
		{name: "provider", child: Provider(providerKey, "value", host)},
		{name: "host", child: host},
	}
	for position, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parent := &fiberNode{ID: "parent"}
			if got := r.renderKeyedElement(ctx, parent, keyElement{key: tc.name, child: tc.child}, position); got == nil {
				t.Fatalf("keyed %s child rendered nil", tc.name)
			}
		})
	}
	if componentCalls != 1 {
		t.Fatalf("component calls = %d, want 1", componentCalls)
	}
	if got := r.renderKeyedElement(ctx, &fiberNode{}, keyElement{key: "empty"}, 0); got != nil {
		t.Fatalf("keyed empty child = %#v, want nil", got)
	}
}

func TestR0ReconcilerIdentityHelperContracts(t *testing.T) {
	host := FromWidget(r0ContractWidget{name: "host"})
	keyedHost := keyElement{key: "stable", child: host}

	if got := reuseChild(nil, fiberHost, "host", "stable", 2, host); got.Parent != nil || got.Kind != fiberHost || got.Key != "stable" {
		t.Fatalf("nil-parent child = %#v, want detached host node", got)
	}
	parent := &fiberNode{ID: "parent"}
	existing := &fiberNode{Kind: fiberHost, TypeID: "host", Key: "stable"}
	parent.Children = []*fiberNode{existing}
	if got := reuseChild(parent, fiberHost, "host", "stable", 0, host); got != existing {
		t.Fatal("reuseChild did not preserve a matching positioned node")
	}
	if got := reuseChild(parent, fiberHost, "other", "stable", 0, host); got == existing || got.Parent != parent {
		t.Fatalf("reuseChild did not allocate a mismatched node: %#v", got)
	}
	if got := reuseChild(parent, fiberHost, "host", "stable", 1, host); got.Parent != parent {
		t.Fatalf("reuseChild did not attach an out-of-range child: %#v", got)
	}

	if kind, typeID, key := elementFiberIdentity(nil); kind != fiberHost || typeID != "nil" || key != "" {
		t.Fatalf("nil fiber identity = (%q, %q, %q)", kind, typeID, key)
	}
	if kind, _, key := elementFiberIdentity(keyedHost); kind != fiberHost || key != "stable" {
		t.Fatalf("keyed host identity = (%q, %q)", kind, key)
	}
	if kind, _, _ := elementFiberIdentity(ComponentElement(func(*Context) Element { return host })); kind != fiberComponent {
		t.Fatalf("component fiber kind = %q, want %q", kind, fiberComponent)
	}
	if kind, typeID, _ := elementFiberIdentity(Fragment(host)); kind != fiberFragment || typeID != "fragment" {
		t.Fatalf("fragment fiber identity = (%q, %q)", kind, typeID)
	}
	if kind, _, _ := elementFiberIdentity(Provider(NewContextKey(""), "value", host)); kind != fiberProvider {
		t.Fatalf("provider fiber kind = %q, want %q", kind, fiberProvider)
	}

	if got := elementExplicitKey(nil); got != "" {
		t.Fatalf("nil explicit key = %q, want empty", got)
	}
	if got := elementExplicitKey(keyedHost); got != "stable" {
		t.Fatalf("keyed explicit key = %q, want stable", got)
	}
	if containsKeyedElement([]Element{host}) {
		t.Fatal("unkeyed host reported as keyed")
	}
	if !containsKeyedElement([]Element{host, keyedHost}) {
		t.Fatal("keyed child was not detected")
	}
	if got, want := childMatchKey(fiberHost, "host", "stable"), "host:host#stable"; got != want {
		t.Fatalf("child match key = %q, want %q", got, want)
	}

	if got, want := nodeParentID(nil), "root"; got != want {
		t.Fatalf("nil node parent ID = %q, want %q", got, want)
	}
	if got, want := nodeParentID(&fiberNode{Parent: parent}), "parent"; got != want {
		t.Fatalf("node parent ID = %q, want %q", got, want)
	}
	if got, want := fiberStableID(nil, fiberHost, "host", "", 2), "root/host:host@2"; got != want {
		t.Fatalf("unkeyed host stable ID = %q, want %q", got, want)
	}
	if got, want := fiberStableID(parent, fiberHost, "host", "stable", 2), "parent/host:host#\"stable\""; got != want {
		t.Fatalf("keyed host stable ID = %q, want %q", got, want)
	}
	componentID := fiberStableID(parent, fiberComponent, "component-type", "stable", 2)
	wantComponentID := internal.ComponentIdentity{ParentID: "parent", TypeID: "component-type", Key: "stable", Position: 2}.StableID()
	if componentID != wantComponentID {
		t.Fatalf("component stable ID = %q, want %q", componentID, wantComponentID)
	}

	identity := internal.ComponentIdentity{ParentID: "root", TypeID: "test", Position: 0}
	if got := beginComponentInstance(nil, identity); got != nil {
		t.Fatalf("nil context component instance = %#v, want nil", got)
	}
	if got := beginComponentInstance(r0ContractContext(nil), identity); got != nil {
		t.Fatalf("nil runtime component instance = %#v, want nil", got)
	}
	runtime := internal.NewRuntime(nil)
	runtime.BeginFrame()
	if got := beginComponentInstance(r0ContractContext(runtime), identity); got == nil {
		t.Fatal("runtime component instance = nil")
	}
	runtime.EndFrame()
}

func TestR0ReconcilerFallbackAndProviderEmptyContracts(t *testing.T) {
	r := newReconciler()
	leaf := FromWidget(r0ContractWidget{name: "leaf"})
	if got := r.Render(nil, func(*Context) Element { return leaf }); got != nil {
		t.Fatalf("Render(nil, root) = %#v, want nil", got)
	}
	if got := r.Render(r0ContractContext(nil), nil); got != nil {
		t.Fatalf("Render(context, nil) = %#v, want nil", got)
	}
	if got := r.Render(r0ContractContext(nil), func(*Context) Element { return leaf }); got == nil {
		t.Fatal("Render without a runtime should render a host element")
	}

	node := &fiberNode{Children: []*fiberNode{{}}}
	if got := r.renderProvider(r0ContractContext(nil), node, nil); got != nil {
		t.Fatalf("nil provider = %#v, want nil", got)
	}
	if got := r.renderProvider(r0ContractContext(nil), node, providerElement[string]{}); got != nil {
		t.Fatalf("empty provider = %#v, want nil", got)
	}
	if node.Children != nil {
		t.Fatalf("empty provider retained children: %#v", node.Children)
	}
	if got := r.renderProvider(r0ContractContext(nil), nil, providerElement[string]{
		key:   NewContextKey(""),
		value: "value",
		child: leaf,
	}); got == nil {
		t.Fatal("provider without a fiber node did not render its child")
	}
}

func TestR0LegacyHookFallbackContracts(t *testing.T) {
	if got := UseMemo[int](nil, nil, nil); got != 0 {
		t.Fatalf("nil-factory memo = %d, want zero", got)
	}
	if got := UseMemo(nil, nil, func() int { return 7 }); got != 7 {
		t.Fatalf("nil-context memo = %d, want 7", got)
	}
	if ref := UseRef[int](nil, 3); ref == nil || ref.Current != 3 {
		t.Fatalf("nil-context ref = %#v, want initial value", ref)
	}
	if handle := UseAsync[int](nil); handle == nil || handle.Status() != AsyncIdle {
		t.Fatalf("nil-context async handle = %#v, want idle handle", handle)
	}
	if got := UseAnimatedValue[float32](nil, 4, time.Second, nil); got != 4 {
		t.Fatalf("nil-context animated value = %v, want 4", got)
	}
	if got := UseAnimatedDecoration(nil, Decoration{}, time.Second, nil); got != (Decoration{}) {
		t.Fatalf("nil-context animated decoration = %#v, want target", got)
	}
	UseLifecycle(nil, nil, nil)
	UseInterval(nil, 0, nil)

	runtime := internal.NewRuntime(nil)
	memoRuns := 0
	for _, dependency := range []int{1, 1, 2} {
		runtime.BeginFrame()
		got := UseMemo(r0ContractContext(runtime), []any{dependency}, func() int {
			memoRuns++
			return dependency * 10
		})
		runtime.EndFrame()
		if got != dependency*10 {
			t.Fatalf("legacy memo for dependency %d = %d", dependency, got)
		}
	}
	if memoRuns != 2 {
		t.Fatalf("legacy memo factory runs = %d, want 2", memoRuns)
	}

	runtime.BeginFrame()
	firstRef := UseRef(r0ContractContext(runtime), "initial")
	firstRef.Current = "retained"
	_ = UseAnimatedValue(r0ContractContext(runtime), float32(2), 0, nil)
	_ = UseAnimatedDecoration(r0ContractContext(runtime), Decoration{}, 0, nil)
	runtime.EndFrame()
	runtime.BeginFrame()
	secondRef := UseRef(r0ContractContext(runtime), "new")
	if secondRef != firstRef || secondRef.Current != "retained" {
		runtime.EndFrame()
		t.Fatalf("legacy ref = %#v, want retained ref %#v", secondRef, firstRef)
	}
	if got := UseAnimatedValue(r0ContractContext(runtime), float32(3), 0, nil); got != 3 {
		runtime.EndFrame()
		t.Fatalf("legacy animated value = %v, want 3", got)
	}
	_ = UseAnimatedDecoration(r0ContractContext(runtime), Decoration{}, 0, nil)
	runtime.EndFrame()

	assertLegacyHookTypeMismatch(t, func(ctx *Context) {
		_ = UseMemo(ctx, nil, func() int { return 1 })
	}, func(ctx *Context) {
		_ = UseMemo(ctx, nil, func() string { return "wrong" })
	})
	assertLegacyHookTypeMismatch(t, func(ctx *Context) {
		_ = UseRef(ctx, 1)
	}, func(ctx *Context) {
		_ = UseRef(ctx, "wrong")
	})
}

func assertLegacyHookTypeMismatch(t *testing.T, establish, mismatch func(*Context)) {
	t.Helper()
	runtime := internal.NewRuntime(nil)
	runtime.BeginFrame()
	establish(r0ContractContext(runtime))
	runtime.EndFrame()

	runtime.BeginFrame()
	defer runtime.EndFrame()
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("legacy hook type mismatch did not panic")
		}
	}()
	mismatch(r0ContractContext(runtime))
}
