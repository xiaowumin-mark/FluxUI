package ui

import (
	"reflect"
	"strconv"

	fluxapp "github.com/xiaowumin-mark/FluxUI/app"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	state "github.com/xiaowumin-mark/FluxUI/state"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

// Element 是 React 风格实验 API 的声明式节点。
// 当前阶段它已经从 Widget 容器拆分为独立树结构，但最终仍会渲染为现有 Widget。
type Element interface {
	render() widget.Widget
}

// ElementIdentity 描述实验 Element 的身份信息。
type ElementIdentity struct {
	Kind       string
	Key        string
	ChildCount int
}

// Component 是实验阶段的函数组件签名。
type Component func(ctx *Context) Element

type hostElement struct {
	child Widget
}

type fragmentElement struct {
	children []Element
}

type providerElement[T any] struct {
	value T
	child Element
}

type keyElement struct {
	key   string
	child Element
}

type componentElement struct {
	component Component
}

type keyed interface {
	Key() string
}

type identifiable interface {
	identity() ElementIdentity
}

func (e hostElement) render() widget.Widget {
	return e.child
}

func (e hostElement) identity() ElementIdentity {
	return ElementIdentity{Kind: "host", ChildCount: 1}
}

func (e fragmentElement) render() widget.Widget {
	children := make([]Widget, 0, len(e.children))
	for _, child := range e.children {
		if child == nil {
			continue
		}
		if w := child.render(); w != nil {
			children = append(children, w)
		}
	}
	if len(children) == 0 {
		return nil
	}
	return widget.Column(children...)
}

func (e fragmentElement) identity() ElementIdentity {
	return ElementIdentity{Kind: "fragment", ChildCount: len(e.children)}
}

func (e fragmentElement) Children() []Element {
	return append([]Element(nil), e.children...)
}

func (e providerElement[T]) render() widget.Widget {
	if e.child == nil {
		return nil
	}
	return e.child.render()
}

func (e providerElement[T]) identity() ElementIdentity {
	ident := ElementIdentity{Kind: "provider", ChildCount: 1}
	if child, ok := e.child.(identifiable); ok {
		childIdent := child.identity()
		if childIdent.Key != "" {
			ident.Key = childIdent.Key
		}
	}
	return ident
}

func (e providerElement[T]) Child() Element {
	return e.child
}

func (e keyElement) render() widget.Widget {
	if e.child == nil {
		return nil
	}
	return e.child.render()
}

func (e keyElement) identity() ElementIdentity {
	ident := ElementIdentity{Kind: "key", Key: e.key, ChildCount: 1}
	if child, ok := e.child.(identifiable); ok {
		childIdent := child.identity()
		if childIdent.Kind != "" {
			ident.Kind = childIdent.Kind
		}
		if childIdent.ChildCount > 0 {
			ident.ChildCount = childIdent.ChildCount
		}
	}
	ident.Key = e.key
	return ident
}

func (e keyElement) Key() string {
	return e.key
}

func (e keyElement) Child() Element {
	return e.child
}

func (e componentElement) render() widget.Widget {
	return nil
}

func (e componentElement) identity() ElementIdentity {
	return ElementIdentity{Kind: "component", ChildCount: 1}
}

func (e componentElement) Component() Component {
	return e.component
}

func (e providerElement[T]) providerContext(ctx *Context) *Context {
	return internal.WithProviderValue[T](ctx, e.value)
}

// RunElement 启动 React 风格实验入口。
// 当前实现仍复用既有 Widget 运行链路，但 Element 与 Widget 的结构已拆分。
func RunElement(root Component, opts ...AppOption) error {
	reconciler := newReconciler()
	return fluxapp.Run(func(ctx *internal.Context) widget.Widget {
		return reconciler.Render(ctx, root)
	}, opts...)
}

// UseState 创建或读取带初始值的状态。
func UseState[T any](ctx *Context, initial T) *state.State[T] {
	if hook := ctx.NextHookSlot(internal.HookState); hook != nil {
		key := ""
		if inst := ctx.ComponentInstance(); inst != nil {
			key = inst.ID + "/state:" + strconv.Itoa(inst.HookCount()-1)
		}
		return state.FromHookSlot[T](ctx, key, hook, initial)
	}
	return state.UseWithInitial[T](ctx, initial)
}

// FromWidget 将旧 Widget 包装为 Element host 节点。
func FromWidget(w Widget) Element {
	if w == nil {
		return nil
	}
	return hostElement{child: w}
}

// Key 为子树附加稳定 key。
// 当前阶段先作为树节点信息保留，后续会进入 reconciler 身份匹配。
func Key(key string, child Element) Element {
	if child == nil {
		return nil
	}
	return keyElement{key: key, child: child}
}

// Fragment 组合多个子节点。
func Fragment(children ...Element) Element {
	return fragmentElement{children: append([]Element(nil), children...)}
}

// ComponentElement wraps a function component so it can appear below another Element.
func ComponentElement(component Component) Element {
	if component == nil {
		return nil
	}
	return componentElement{component: component}
}

// ContextKey identifies a typed provider value.
type ContextKey[T any] struct {
	Default T
}

// Provider overrides a typed context value for a child subtree.
func Provider[T any](key ContextKey[T], value T, child Element) Element {
	if child == nil {
		return nil
	}
	return providerElement[T]{value: value, child: child}
}

// UseContext reads the nearest typed provider value or the key default.
func UseContext[T any](ctx *Context, key ContextKey[T]) T {
	return internal.ProviderValue[T](ctx, key.Default)
}

// RenderElement 将实验 Element 统一渲染为底层 Widget。
// 这是 Phase 1.3 的唯一转换出口，后续 reconciler 也会从这里接管。
func RenderElement(el Element) Widget {
	return renderElement(el)
}

// ElementKey 提取 Element 的 key 信息，供后续 reconciler 使用。
func ElementKey(el Element) string {
	if el == nil {
		return ""
	}
	if k, ok := el.(keyed); ok {
		return k.Key()
	}
	return ""
}

// ElementInfo 返回 Element 的身份信息，供后续 reconciler 使用。
func ElementInfo(el Element) ElementIdentity {
	if el == nil {
		return ElementIdentity{}
	}
	if ident, ok := el.(identifiable); ok {
		return ident.identity()
	}
	return ElementIdentity{Kind: "unknown"}
}

func renderElement(el Element) widget.Widget {
	if el == nil {
		return nil
	}
	return el.render()
}

func renderElementWithContext(ctx *Context, el Element) widget.Widget {
	if el == nil {
		return nil
	}
	if renderable, ok := el.(contextRenderable); ok {
		return renderable.renderWithContext(ctx)
	}
	return el.render()
}

type contextRenderable interface {
	renderWithContext(ctx *Context) widget.Widget
}

func (e fragmentElement) renderWithContext(ctx *Context) widget.Widget {
	children := make([]Widget, 0, len(e.children))
	for _, child := range e.children {
		if child == nil {
			continue
		}
		if w := renderElementWithContext(ctx, child); w != nil {
			children = append(children, w)
		}
	}
	if len(children) == 0 {
		return nil
	}
	return widget.Column(children...)
}

func (e keyElement) renderWithContext(ctx *Context) widget.Widget {
	return renderElementWithContext(ctx, e.child)
}

func (e providerElement[T]) renderWithContext(ctx *Context) widget.Widget {
	return renderElementWithContext(internal.WithProviderValue[T](ctx, e.value), e.child)
}

func useEffect(ctx *Context, hasDeps bool, deps []any, effect Effect) {
	if ctx == nil || effect == nil {
		return
	}
	if hook := ctx.NextHookSlot(internal.HookEffect); hook != nil {
		rt := ctx.Runtime()
		if rt == nil || rt.HookStore() == nil {
			return
		}
		rt.HookStore().UseEffect(hook, hasDeps, deps, internal.EffectSetup(effect))
		return
	}
	if hasDeps {
		state.UseEffectWithDeps(ctx, deps, effect)
		return
	}
	state.UseEffect(ctx, effect)
}

func renderComponent(ctx *Context, parentID string, position int, key string, component Component) Element {
	if ctx == nil || component == nil {
		return nil
	}
	identity := internal.ComponentIdentity{
		ParentID: parentID,
		TypeID:   componentTypeID(component),
		Key:      key,
		Position: position,
	}
	inst := beginComponentInstance(ctx, identity)
	if inst == nil {
		return component(ctx)
	}
	return component(ctx.WithComponentInstance(inst))
}

func componentTypeID(component Component) string {
	if component == nil {
		return "component"
	}
	return reflect.TypeOf(component).String() + "@" + strconv.FormatUint(uint64(reflect.ValueOf(component).Pointer()), 16)
}
