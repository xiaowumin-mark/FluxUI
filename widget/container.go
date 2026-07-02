package widget

import (
	"image"
	"image/color"

	event "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/f32"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type containerWidget struct {
	style style.Style
	child Widget
}

type ContainerDecorationOption func(*decorationContainerConfig)

type decorationContainerConfig struct {
	disabled     bool
	onHoverEnter func(ctx *internal.Context)
	onHoverLeave func(ctx *internal.Context)
	onHover      func(ctx *internal.Context, hovering bool)
	onClick      func(ctx *internal.Context)
	onPressed    func(ctx *internal.Context, pressed bool)
}

type decorationContainerWidget struct {
	decoration style.Decoration
	child      Widget
	config     decorationContainerConfig
	wasHovered bool
	wasPressed bool
}

type paddingWidget struct {
	insets style.Insets
	child  Widget
}

// ContainerDecorationDisabled 设置容器禁用状态。
func ContainerDecorationDisabled(disabled bool) ContainerDecorationOption {
	return func(cfg *decorationContainerConfig) { cfg.disabled = disabled }
}

// ContainerDecorationOnClick 设置点击回调。
func ContainerDecorationOnClick(fn func(ctx *internal.Context)) ContainerDecorationOption {
	return func(cfg *decorationContainerConfig) { cfg.onClick = fn }
}

// ContainerDecorationOnHoverEnter 设置鼠标进入回调。
func ContainerDecorationOnHoverEnter(fn func(ctx *internal.Context)) ContainerDecorationOption {
	return func(cfg *decorationContainerConfig) { cfg.onHoverEnter = fn }
}

// ContainerDecorationOnHoverLeave 设置鼠标离开回调。
func ContainerDecorationOnHoverLeave(fn func(ctx *internal.Context)) ContainerDecorationOption {
	return func(cfg *decorationContainerConfig) { cfg.onHoverLeave = fn }
}

// ContainerDecorationOnHover 设置悬浮变化回调。
func ContainerDecorationOnHover(fn func(ctx *internal.Context, hovering bool)) ContainerDecorationOption {
	return func(cfg *decorationContainerConfig) { cfg.onHover = fn }
}

// ContainerDecorationOnPressed 设置按下/释放回调。
func ContainerDecorationOnPressed(fn func(ctx *internal.Context, pressed bool)) ContainerDecorationOption {
	return func(cfg *decorationContainerConfig) { cfg.onPressed = fn }
}

// Deprecated: Container 已被 ContainerDecoration 取代。请使用 ContainerDecoration + Decoration 构造器链。
func Container(st style.Style, child Widget) Widget {
	return &containerWidget{
		style: st,
		child: child,
	}
}

// ContainerDecoration 基于 Decoration 创建带背景、内/外边距、圆角和边框装饰的容器。
//
// 可选参数 opts 支持交互状态和事件回调：
//
//	ContainerDecoration(decoration, child,
//	    ContainerDecorationOnClick(func(ctx) { ... }),
//	    ContainerDecorationOnHoverEnter(func(ctx) { ... }),
//	)
func ContainerDecoration(d style.Decoration, child Widget, opts ...ContainerDecorationOption) Widget {
	w := &decorationContainerWidget{
		decoration: d,
		child:      child,
	}
	for _, opt := range opts {
		opt(&w.config)
	}
	return w
}

// Padding 创建仅带内边距的容器。
func Padding(insets style.Insets, child Widget) Widget {
	return &paddingWidget{
		insets: insets,
		child:  child,
	}
}

func (c *containerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	size := ctx.LayoutInset(toInternalInsets(c.style.Margin), func(marginCtx *internal.Context) image.Point {
		return marginCtx.LayoutSurface(internal.SurfaceSpec{
			Background: c.style.Background,
			Radius:     c.style.Radius,
			Padding:    toInternalInsets(c.style.Padding),
		}, func(contentCtx *internal.Context) image.Point {
			if c.child == nil {
				return image.Point{}
			}
			return c.child.Layout(contentCtx.Child(0)).Size
		})
	})
	return layout.Dimensions{Size: size}
}

func (c *decorationContainerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	hasInteraction := c.decoration.Hover != nil || c.decoration.Pressed != nil ||
		c.decoration.Disabled != nil ||
		c.config.onHoverEnter != nil || c.config.onHoverLeave != nil ||
		c.config.onHover != nil || c.config.onClick != nil || c.config.onPressed != nil

	if hasInteraction {
		return c.layoutInteractive(ctx)
	}
	return c.layoutPassive(ctx)
}

func (c *decorationContainerWidget) layoutPassive(ctx *internal.Context) layout.Dimensions {
	spec := c.buildSpec(ctx)
	margin := c.decoration.ResolveMargin(style.Insets{})

	size := ctx.LayoutInset(toInternalInsets(margin), func(marginCtx *internal.Context) image.Point {
		return layoutSurfaceWithTransform(marginCtx, spec, c.decoration.ResolveTransform(), func(contentCtx *internal.Context) image.Point {
			if c.child == nil {
				return image.Point{}
			}
			return c.child.Layout(contentCtx.Child(0)).Size
		})
	})
	return layout.Dimensions{Size: size}
}

func (c *decorationContainerWidget) layoutInteractive(ctx *internal.Context) layout.Dimensions {
	var clickable *event.Clickable
	interaction := event.InteractionSnapshot{}

	if !c.config.disabled {
		clickable = event.UseClickable(ctx)

		for clickable.Clicked(ctx) {
			if c.config.onClick != nil {
				c.config.onClick(ctx)
			}
		}

		interaction = clickable.Snapshot(ctx, false)
		if changed, hovering := clickable.HoverChangedWithSnapshot(ctx, interaction.Hovered); changed {
			c.wasHovered = hovering
			if hovering && c.config.onHoverEnter != nil {
				c.config.onHoverEnter(ctx)
			}
			if !hovering && c.config.onHoverLeave != nil {
				c.config.onHoverLeave(ctx)
			}
			if c.config.onHover != nil {
				c.config.onHover(ctx, hovering)
			}
		}
	}

	pressed := interaction.Pressed
	if c.wasPressed != pressed {
		c.wasPressed = pressed
		if c.config.onPressed != nil {
			c.config.onPressed(ctx, pressed)
		}
	}

	activeDeco := c.decoration
	if c.config.disabled {
		if c.decoration.Disabled != nil {
			activeDeco = c.decoration.Merge(*c.decoration.Disabled)
		}
	} else if pressed {
		if c.decoration.Pressed != nil {
			activeDeco = c.decoration.Merge(*c.decoration.Pressed)
		}
	} else if interaction.Hovered {
		if c.decoration.Hover != nil {
			activeDeco = c.decoration.Merge(*c.decoration.Hover)
		}
	}

	visualDeco := stripStateDecoration(activeDeco)
	if !c.config.disabled {
		duration, easing := md3InteractionTiming(ctx, interaction.Hovered, pressed, false, false)
		visualDeco = md3AnimateDecoration(ctx, "container-decoration", visualDeco, duration, easing)
	}
	spec := decorationSurfaceSpec(visualDeco, ctx, ctx.Theme().Surface)
	baseMargin := c.decoration.ResolveMargin(style.Insets{})
	transform := visualDeco.ResolveTransform()

	visualMacro := op.Record(ctx.Gtx.Ops)
	dims := ctx.LayoutInset(toInternalInsets(baseMargin), func(marginCtx *internal.Context) image.Point {
		return layoutSurfaceWithTransform(marginCtx, spec, transform, func(contentCtx *internal.Context) image.Point {
			if c.child == nil {
				return image.Point{}
			}
			return c.child.Layout(contentCtx.Child(0)).Size
		})
	})
	visualCall := visualMacro.Stop()

	var handle *internal.ClickableState
	if clickable != nil {
		handle = clickable.Handle()
	}
	layoutTransformedClickArea(ctx, handle, baseMargin, transform, dims)
	visualCall.Add(ctx.Gtx.Ops)

	return layout.Dimensions{Size: dims}
}

func layoutSurfaceWithTransform(ctx *internal.Context, spec internal.SurfaceSpec, transform *style.Transform2D, child func(*internal.Context) image.Point) image.Point {
	if transform == nil {
		return ctx.LayoutSurface(spec, child)
	}

	macro := op.Record(ctx.Gtx.Ops)
	size := ctx.LayoutSurface(spec, child)
	call := macro.Stop()

	stack := op.Affine(decorationTransformMatrix(ctx.Gtx, *transform, size)).Push(ctx.Gtx.Ops)
	call.Add(ctx.Gtx.Ops)
	stack.Pop()

	return size
}

func decorationTransformMatrix(gtx gioLayout.Context, transform style.Transform2D, size image.Point) f32.Affine2D {
	transform.TranslateX = float32(gtx.Dp(unit.Dp(transform.TranslateX)))
	transform.TranslateY = float32(gtx.Dp(unit.Dp(transform.TranslateY)))
	return transform.Affine(size)
}

func (c *decorationContainerWidget) buildSpec(ctx *internal.Context) internal.SurfaceSpec {
	return decorationSurfaceSpec(c.decoration, ctx, ctx.Theme().Surface)
}

func decorationSurfaceSpec(d style.Decoration, ctx *internal.Context, defaultBg color.NRGBA) internal.SurfaceSpec {
	bg := d.ResolveBg(defaultBg)
	rad := d.ResolveRad(0)
	border := d.ResolveBorder(style.Border{})
	opacity := d.ResolveOpacity()
	grad := d.ResolveGradient()
	shadow := d.ResolveShadow()

	spec := internal.SurfaceSpec{
		Background:  bg,
		Radius:      rad,
		Padding:     toInternalInsets(d.ResolvePad(style.Insets{})),
		BorderColor: border.Color,
		BorderWidth: border.Width,
		Opacity:     opacity,
		CircleClip:  d.CircleClip,
	}

	if grad != nil {
		spec.HasGradient = true
		spec.GradientStart = f32.Pt(float32(grad.Start.X), float32(grad.Start.Y))
		spec.GradientEnd = f32.Pt(float32(grad.End.X), float32(grad.End.Y))
		spec.GradientFrom = grad.From
		spec.GradientTo = grad.To
	}

	if shadow != nil {
		spec.HasShadow = true
		spec.ShadowOffsetX = shadow.OffsetX
		spec.ShadowOffsetY = shadow.OffsetY
		spec.ShadowBlur = shadow.Blur
		spec.ShadowColor = shadow.Color
	}

	if img := d.ResolveImage(); img != nil && img.Src != nil {
		spec.HasImage = true
		spec.ImageOp = paint.NewImageOp(img.Src)
		spec.ImageFit = int(img.Fit)
	}

	return spec
}

func (p *paddingWidget) Layout(ctx *internal.Context) layout.Dimensions {
	size := ctx.LayoutInset(toInternalInsets(p.insets), func(contentCtx *internal.Context) image.Point {
		if p.child == nil {
			return image.Point{}
		}
		return p.child.Layout(contentCtx.Child(0)).Size
	})
	return layout.Dimensions{Size: size}
}
