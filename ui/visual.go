//go:build visual

package ui

// VisualRootBuilder creates a stateful Element root renderer for visual
// regression captures. It is only available when building with -tags visual.
func VisualRootBuilder(root Component) func(ctx *Context) Widget {
	reconciler := newReconciler()
	return func(ctx *Context) Widget {
		return reconciler.Render(ctx, root)
	}
}
