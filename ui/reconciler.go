package ui

import (
	"strconv"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

type fiberKind string

const (
	fiberComponent fiberKind = "component"
	fiberFragment  fiberKind = "fragment"
	fiberProvider  fiberKind = "provider"
	fiberHost      fiberKind = "host"
)

type fiberNode struct {
	ID       string
	Kind     fiberKind
	TypeID   string
	Key      string
	Position int
	Parent   *fiberNode
	Children []*fiberNode
	Instance *internal.ComponentInstance
	Element  Element
}

type reconciler struct {
	root *fiberNode
}

func newReconciler() *reconciler {
	return &reconciler{}
}

func (r *reconciler) Render(ctx *Context, root Component) widget.Widget {
	if ctx == nil || root == nil {
		return nil
	}
	return r.renderRootComponent(ctx, root)
}

func (r *reconciler) renderRootComponent(ctx *Context, root Component) widget.Widget {
	if ctx == nil || root == nil {
		return nil
	}
	identity := internal.ComponentIdentity{
		ParentID: "root",
		TypeID:   componentTypeID(root),
		Position: 0,
	}
	inst := beginComponentInstance(ctx, identity)
	if inst == nil {
		return renderElementWithContext(ctx, root(ctx))
	}
	node := r.root
	if node == nil || node.ID != inst.ID {
		node = &fiberNode{
			ID:       inst.ID,
			Kind:     fiberComponent,
			TypeID:   identity.TypeID,
			Position: identity.Position,
			Instance: inst,
		}
		r.root = node
	} else {
		node.Instance = inst
	}
	node.Element = root(ctx.WithComponentInstance(inst))
	return r.renderComponentOutput(ctx.WithComponentInstance(inst), node, node.Element)
}

func (r *reconciler) renderElement(ctx *Context, parent *fiberNode, el Element) widget.Widget {
	if el == nil {
		if parent != nil {
			parent.Children = nil
		}
		return nil
	}
	switch e := el.(type) {
	case keyElement:
		return r.renderElement(ctx, parent, e.Child())
	case fragmentElement:
		return r.renderFragment(ctx, parent, e)
	case componentElement:
		if parent != nil {
			parent.Kind = fiberComponent
			parent.TypeID = componentTypeID(e.component)
			parent.Element = el
		}
		return r.renderComponentNode(ctx, parent, e.component)
	case providerScoped:
		if parent != nil {
			parent.Kind = fiberProvider
			parent.TypeID = "provider"
			parent.Element = el
		}
		return r.renderProvider(ctx, parent, e)
	default:
		return r.renderHost(ctx, parent, el)
	}
}

func (r *reconciler) renderKeyedElement(ctx *Context, parent *fiberNode, el keyElement, position int) widget.Widget {
	child := el.Child()
	if child == nil {
		return nil
	}
	switch childEl := child.(type) {
	case componentElement:
		node := reuseChild(parent, fiberComponent, componentTypeID(childEl.component), el.Key(), position, child)
		return r.renderComponentNode(ctx, node, childEl.component)
	case fragmentElement:
		node := reuseChild(parent, fiberFragment, "fragment", el.Key(), position, child)
		return r.renderFragment(ctx, node, childEl)
	case providerScoped:
		node := reuseChild(parent, fiberProvider, "provider", el.Key(), position, child)
		return r.renderProvider(ctx, node, childEl)
	default:
		node := reuseChild(parent, fiberHost, ElementInfo(child).Kind, el.Key(), position, child)
		return r.renderHost(ctx, node, child)
	}
}

func (r *reconciler) renderProvider(ctx *Context, node *fiberNode, provider providerScoped) widget.Widget {
	if provider == nil {
		return nil
	}
	child := provider.Child()
	if child == nil {
		if node != nil {
			node.Children = nil
		}
		return nil
	}
	if node != nil {
		node.Kind = fiberProvider
		node.TypeID = "provider"
		node.Children = reconcileChildren(node, []Element{child})
		return r.renderElement(provider.providerContext(ctx), node.Children[0], child)
	}
	return r.renderElement(provider.providerContext(ctx), nil, child)
}

func (r *reconciler) renderHost(ctx *Context, node *fiberNode, el Element) widget.Widget {
	if node != nil {
		node.Kind = fiberHost
		node.TypeID = ElementInfo(el).Kind
		node.Element = el
		node.Instance = nil
		node.Children = nil
	}
	return renderElementWithContext(ctx, el)
}

func (r *reconciler) renderFragment(ctx *Context, parent *fiberNode, el fragmentElement) widget.Widget {
	children := el.Children()
	if parent != nil {
		parent.Children = reconcileChildren(parent, children)
	}
	widgets := make([]Widget, 0, len(children))
	for idx, child := range children {
		if child == nil {
			continue
		}
		var node *fiberNode
		if parent != nil && idx < len(parent.Children) {
			node = parent.Children[idx]
		}
		if w := r.renderElement(ctx, node, child); w != nil {
			widgets = append(widgets, w)
		}
	}
	if len(widgets) == 0 {
		return nil
	}
	return widget.Column(widgets...)
}

func (r *reconciler) renderComponentNode(ctx *Context, node *fiberNode, component Component) widget.Widget {
	if ctx == nil || node == nil || component == nil {
		return nil
	}
	identity := internal.ComponentIdentity{
		ParentID: nodeParentID(node),
		TypeID:   node.TypeID,
		Key:      node.Key,
		Position: node.Position,
	}
	inst := beginComponentInstance(ctx, identity)
	if inst == nil {
		return renderElementWithContext(ctx, component(ctx))
	}
	node.ID = inst.ID
	node.Instance = inst
	componentCtx := ctx.WithComponentInstance(inst)
	node.Element = component(componentCtx)
	return r.renderComponentOutput(componentCtx, node, node.Element)
}

func (r *reconciler) renderComponentOutput(ctx *Context, node *fiberNode, el Element) widget.Widget {
	if node == nil || el == nil {
		if node != nil {
			node.Children = nil
		}
		return nil
	}
	if fragment, ok := el.(fragmentElement); ok {
		return r.renderFragment(ctx, node, fragment)
	}
	node.Children = reconcileChildren(node, []Element{el})
	return r.renderElement(ctx, node.Children[0], el)
}

func reuseChild(parent *fiberNode, kind fiberKind, typeID string, key string, position int, el Element) *fiberNode {
	if parent == nil {
		return &fiberNode{Kind: kind, TypeID: typeID, Key: key, Position: position, Element: el}
	}
	if position < len(parent.Children) {
		child := parent.Children[position]
		if child != nil && child.Kind == kind && child.TypeID == typeID && child.Key == key {
			return child
		}
	}
	return &fiberNode{Kind: kind, TypeID: typeID, Key: key, Position: position, Parent: parent, Element: el}
}

func reconcileChildren(parent *fiberNode, elements []Element) []*fiberNode {
	if parent == nil {
		return nil
	}
	previous := parent.Children
	var byKey map[string]*fiberNode
	if containsKeyedElement(elements) {
		byKey = make(map[string]*fiberNode)
		for _, child := range previous {
			if child != nil && child.Key != "" {
				byKey[childMatchKey(child.Kind, child.TypeID, child.Key)] = child
			}
		}
	}
	next := previous[:0]
	for idx, el := range elements {
		kind, typeID, key := elementFiberIdentity(el)
		var node *fiberNode
		if key != "" && byKey != nil {
			node = byKey[childMatchKey(kind, typeID, key)]
		} else if idx < len(previous) {
			prev := previous[idx]
			if prev != nil && prev.Key == "" && prev.Kind == kind && prev.TypeID == typeID {
				node = prev
			}
		}
		if node == nil {
			node = &fiberNode{Parent: parent}
		}
		node.Kind = kind
		node.TypeID = typeID
		node.Key = key
		node.Position = idx
		node.Element = el
		node.ID = fiberStableID(parent, kind, typeID, key, idx)
		next = append(next, node)
	}
	if len(next) < len(previous) {
		clear(previous[len(next):])
	}
	return next
}

func containsKeyedElement(elements []Element) bool {
	for _, el := range elements {
		if elementExplicitKey(el) != "" {
			return true
		}
	}
	return false
}

func elementExplicitKey(el Element) string {
	if el == nil {
		return ""
	}
	if keyed, ok := el.(keyElement); ok {
		return keyed.Key()
	}
	return ElementKey(el)
}

func elementFiberIdentity(el Element) (fiberKind, string, string) {
	if el == nil {
		return fiberHost, "nil", ""
	}
	if keyed, ok := el.(keyElement); ok {
		kind, typeID, _ := elementFiberIdentity(keyed.Child())
		return kind, typeID, keyed.Key()
	}
	switch e := el.(type) {
	case componentElement:
		return fiberComponent, componentTypeID(e.component), ""
	case fragmentElement:
		return fiberFragment, "fragment", ""
	case providerScoped:
		return fiberProvider, "provider", ElementKey(el)
	default:
		return fiberHost, ElementInfo(el).Kind, ElementKey(el)
	}
}

func childMatchKey(kind fiberKind, typeID string, key string) string {
	return string(kind) + ":" + typeID + "#" + key
}

func nodeParentID(node *fiberNode) string {
	if node == nil || node.Parent == nil || node.Parent.ID == "" {
		return "root"
	}
	return node.Parent.ID
}

func fiberStableID(parent *fiberNode, kind fiberKind, typeID string, key string, position int) string {
	parentID := "root"
	if parent != nil && parent.ID != "" {
		parentID = parent.ID
	}
	if kind == fiberComponent {
		return internal.ComponentIdentity{
			ParentID: parentID,
			TypeID:   typeID,
			Key:      key,
			Position: position,
		}.StableID()
	}
	if key != "" {
		return parentID + "/" + string(kind) + ":" + typeID + "#" + strconv.Quote(key)
	}
	return parentID + "/" + string(kind) + ":" + typeID + "@" + strconv.Itoa(position)
}

func beginComponentInstance(ctx *Context, identity internal.ComponentIdentity) *internal.ComponentInstance {
	if ctx == nil || ctx.Runtime() == nil || ctx.Runtime().HookStore() == nil {
		return nil
	}
	return ctx.Runtime().HookStore().BeginInstance(identity)
}

type providerScoped interface {
	Child() Element
	providerContext(ctx *Context) *Context
}
