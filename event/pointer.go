package event

import (
	"fmt"

	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/f32"
)

type ClickHandler func(ctx *internal.Context)

type HoverHandler func(ctx *internal.Context, hovering bool)

type InteractionSnapshot struct {
	Hovered bool
	Pressed bool
	Focused bool
}

type Clickable struct {
	handle       *internal.ClickableState
	runtime      *internal.Runtime
	target       internal.PathID
	hovered      bool
	pressed      bool
	focused      bool
	initialized  bool
	snapshotted  bool
	lastSnapshot InteractionSnapshot
}

func UseClickable(ctx *internal.Context) *Clickable {
	value := ctx.Memo("clickable", func() any {
		return &Clickable{handle: internal.NewClickableState()}
	})

	clickable, ok := value.(*Clickable)
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUI/event: key %q clickable state type mismatch", ctx.TreePath()))
	}
	if clickable.handle != nil {
		clickable.handle.BindRuntime(ctx.Runtime())
	}
	clickable.runtime = ctx.Runtime()
	clickable.target = ctx.PathID()

	return clickable
}

func (c *Clickable) Clicked(ctx *internal.Context) bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Clicked(ctx)
}

func (c *Clickable) ClickedEvent(ctx *internal.Context) (*PointerEvent, bool) {
	if c == nil || c.handle == nil {
		return nil, false
	}
	click, ok := c.handle.ClickedEvent(ctx)
	if !ok {
		return nil, false
	}
	button := ButtonNone
	pointerType := PointerOther
	var position f32.Point
	if click.HasPosition {
		button = ButtonPrimary
		pointerType = PointerMouse
		position = f32.Pt(float32(click.Position.X), float32(click.Position.Y))
	}
	return &PointerEvent{
		Event: Event{
			Type:       Click,
			Time:       eventTime(ctx),
			Bubbles:    true,
			Cancelable: true,
			Trusted:    true,
		},
		PointerType: pointerType,
		IsPrimary:   true,
		Button:      button,
		Position:    position,
		Modifiers:   ModifiersFromGio(click.Modifiers),
		ClickCount:  click.NumClicks,
	}, true
}

func (c *Clickable) Hovered() bool {
	if c == nil || c.handle == nil {
		return false
	}
	hovered := c.handle.Hovered()
	c.observeSnapshot(nil, InteractionSnapshot{
		Hovered: hovered,
		Pressed: c.lastSnapshot.Pressed,
		Focused: c.lastSnapshot.Focused,
	})
	return hovered
}

func (c *Clickable) Pressed() bool {
	if c == nil || c.handle == nil {
		return false
	}
	pressed := c.handle.Pressed()
	c.observeSnapshot(nil, InteractionSnapshot{
		Hovered: c.lastSnapshot.Hovered,
		Pressed: pressed,
		Focused: c.lastSnapshot.Focused,
	})
	return pressed
}

func (c *Clickable) Focused(ctx *internal.Context) bool {
	if c == nil || c.handle == nil {
		return false
	}
	focused := c.handle.Focused(ctx)
	c.observeSnapshot(ctx, InteractionSnapshot{
		Hovered: c.lastSnapshot.Hovered,
		Pressed: c.lastSnapshot.Pressed,
		Focused: focused,
	})
	return focused
}

func (c *Clickable) Snapshot(ctx *internal.Context, includeFocus bool) InteractionSnapshot {
	if c == nil || c.handle == nil {
		return InteractionSnapshot{}
	}
	snapshot := c.handle.Snapshot(ctx, includeFocus)
	interaction := InteractionSnapshot{
		Hovered: snapshot.Hovered,
		Pressed: snapshot.Pressed,
		Focused: snapshot.Focused,
	}
	c.observeSnapshot(ctx, interaction)
	return interaction
}

func (c *Clickable) HoverChanged() (changed bool, hovering bool) {
	return c.hoverChanged(nil)
}

func (c *Clickable) HoverChangedWithContext(ctx *internal.Context) (changed bool, hovering bool) {
	return c.hoverChanged(ctx)
}

func (c *Clickable) HoverChangedWithSnapshot(ctx *internal.Context, hovering bool) (changed bool, hover bool) {
	return c.hoverChangedValue(ctx, hovering)
}

func (c *Clickable) hoverChanged(ctx *internal.Context) (changed bool, hovering bool) {
	if c == nil {
		return false, false
	}
	hovering = c.Hovered()
	return c.hoverChangedValue(ctx, hovering)
}

func (c *Clickable) hoverChangedValue(_ *internal.Context, hovering bool) (changed bool, hover bool) {
	if c == nil {
		return false, false
	}
	c.snapshotted = true
	c.lastSnapshot.Hovered = hovering
	changed = (!c.initialized && hovering) || (c.initialized && c.hovered != hovering)
	c.hovered = hovering
	c.initialized = true
	return changed, hovering
}

func (c *Clickable) observeSnapshot(ctx *internal.Context, snapshot InteractionSnapshot) {
	if c == nil {
		return
	}
	rt := c.runtime
	if ctx != nil && ctx.Runtime() != nil {
		rt = ctx.Runtime()
		c.runtime = rt
	}
	if rt == nil {
		return
	}
	initialized := c.initialized || c.snapshotted
	if !initialized && !snapshot.Hovered && !snapshot.Pressed && !snapshot.Focused {
		c.pressed = snapshot.Pressed
		c.focused = snapshot.Focused
		c.lastSnapshot = snapshot
		c.snapshotted = true
		return
	}
	if initialized &&
		c.hovered == snapshot.Hovered &&
		c.pressed == snapshot.Pressed &&
		c.focused == snapshot.Focused &&
		!snapshot.Hovered && !snapshot.Pressed && !snapshot.Focused {
		c.lastSnapshot = snapshot
		c.snapshotted = true
		return
	}
	target := c.target
	if target == 0 && ctx != nil {
		target = ctx.PathID()
		c.target = target
	}
	if target == 0 {
		return
	}
	previous := internal.ClickableSnapshot{
		Hovered: c.hovered,
		Pressed: c.pressed,
		Focused: c.focused,
	}
	current := internal.ClickableSnapshot{
		Hovered: snapshot.Hovered,
		Pressed: snapshot.Pressed,
		Focused: snapshot.Focused,
	}
	rt.ObserveInteractionSnapshot(target, previous, current, initialized)
	c.pressed = snapshot.Pressed
	c.focused = snapshot.Focused
	c.lastSnapshot = snapshot
	c.snapshotted = true
}

func (c *Clickable) Handle() *internal.ClickableState {
	if c == nil {
		return nil
	}
	return c.handle
}
