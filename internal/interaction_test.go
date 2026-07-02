package internal

import "testing"

func TestInteractionPressedChangeRequestsRedraw(t *testing.T) {
	rt := NewRuntime(nil)
	redraws := 0
	rt.SetInvalidator(func() {
		redraws++
	})

	rt.BeginFrame()
	rt.ObserveInteractionSnapshot(1, ClickableSnapshot{}, ClickableSnapshot{}, false)
	rt.EndFrame()
	if redraws != 0 {
		t.Fatalf("initial interaction snapshot requested %d redraws, want 0", redraws)
	}

	rt.BeginFrame()
	rt.ObserveInteractionSnapshot(1, ClickableSnapshot{}, ClickableSnapshot{Pressed: true}, true)
	rt.EndFrame()
	if redraws == 0 {
		t.Fatal("pressed interaction change did not request redraw")
	}
}

func TestInteractionHoverTargetChangeRequestsRedraw(t *testing.T) {
	rt := NewRuntime(nil)
	redraws := 0
	rt.SetInvalidator(func() {
		redraws++
	})

	rt.BeginFrame()
	rt.ObserveInteractionSnapshot(1, ClickableSnapshot{}, ClickableSnapshot{Hovered: true}, false)
	rt.EndFrame()
	if redraws == 0 {
		t.Fatal("hover target change did not request redraw")
	}
}
