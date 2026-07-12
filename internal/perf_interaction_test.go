package internal

import (
	"bytes"
	"image"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
)

func TestPerfSectionString(t *testing.T) {
	tests := []struct {
		section PerfSection
		want    string
	}{
		{PerfLayout, "layout"},
		{PerfDraw, "draw"},
		{PerfAnimation, "animation"},
		{PerfState, "state"},
		{PerfText, "text"},
		{PerfInput, "input"},
		{PerfSection(99), "unknown"},
	}
	for _, test := range tests {
		if got := test.section.String(); got != test.want {
			t.Errorf("PerfSection(%d).String() = %q, want %q", test.section, got, test.want)
		}
	}
}

func TestPerfDiagnosticsCollectsFrameMetrics(t *testing.T) {
	runtime := NewRuntime(nil)
	var output bytes.Buffer
	invalidations := 0
	runtime.SetInvalidator(func() { invalidations++ })
	runtime.SetPerfDiagnostics(PerfDiagnostics{
		Enabled:          true,
		LogRedrawReasons: true,
		Writer:           &output,
	})
	if config := runtime.PerfDiagnostics(); !config.Enabled || !config.MeasureDurations || config.Writer != &output {
		t.Fatalf("unexpected perf diagnostics configuration: %#v", config)
	}

	// A reason recorded before BeginFrame must be carried into the next frame.
	runtime.RecordRedrawReason("before-frame")
	runtime.BeginFrame()
	child := runtime.childPath(rootPathID, 3)
	runtime.RecordRedrawReason("during-frame")
	runtime.RecordRedrawReason("during-frame")
	runtime.recordRedrawReason("owned", "", child)
	runtime.RequestRedrawReason("")

	runtime.RecordFrameSection(PerfLayout, 3)
	runtime.RecordFrameSection(PerfAnimation, 4)
	runtime.RecordFrameSection(PerfState, 5)
	runtime.RecordFrameSection(PerfText, 6)
	runtime.RecordFrameSection(PerfInput, 7)
	runtime.RecordFrameSection(PerfSection(99), 100)
	runtime.RecordFrameSection(PerfDraw, 0)
	done := runtime.StartFrameSection(PerfDraw, 2)
	if done == nil {
		t.Fatal("enabled duration measurement did not return a section completion")
	}
	done()

	runtime.RecordViewport(image.Rect(0, 0, 20, 10))
	runtime.RecordViewport(image.Rectangle{})
	runtime.RecordVirtualizedItems(-2, -1, image.Rectangle{})
	runtime.RecordVirtualizedItems(10, 12, image.Rect(1, 2, 11, 22))
	runtime.RecordVirtualizedItems(5, 2, image.Rect(2, 4, 8, 13))
	runtime.RecordNonVirtualizedItems(0)
	runtime.RecordNonVirtualizedItems(20)

	runtime.RecordTextCache(true)
	runtime.RecordTextCache(false)
	runtime.RecordStaticPaintCache(true)
	runtime.RecordStaticPaintCache(false)
	runtime.RecordStaticSubtreeCache(true)
	runtime.RecordStaticSubtreeCache(false)
	runtime.EndFrame()

	stats := runtime.LastFrameStats()
	if stats.Frame != 1 || stats.StartedAt.IsZero() {
		t.Fatalf("unexpected frame timing statistics: %#v", stats)
	}
	if invalidations != 1 {
		t.Fatalf("RequestRedrawReason invalidated %d times, want 1", invalidations)
	}
	if stats.ReasonCounts["before-frame"] != 1 || stats.ReasonCounts["during-frame"] != 2 || stats.ReasonCounts["owned"] != 1 || stats.ReasonCounts["RequestRedraw"] != 1 {
		t.Fatalf("unexpected reason counts: %#v", stats.ReasonCounts)
	}
	ownedKey := RedrawReasonKey{Reason: "owned", Source: "runtime", OwnerPath: "root/3"}
	if stats.RedrawCounts[ownedKey] != 1 {
		t.Fatalf("owned redraw reason = %#v, want one %#v", stats.RedrawCounts, ownedKey)
	}
	if len(stats.Reasons) != 4 || stats.Reasons[0] != "RequestRedraw" || len(stats.RedrawReasons) != 4 {
		t.Fatalf("reason summaries were not cloned and sorted: reasons=%#v redraw=%#v", stats.Reasons, stats.RedrawReasons)
	}
	if stats.Layout.Count != 3 || stats.Draw.Count != 2 || stats.Animation.Count != 4 || stats.State.Count != 5 || stats.Text.Count != 6 || stats.Input.Count != 7 {
		t.Fatalf("unexpected section statistics: %#v", stats)
	}
	if got := stats.Virtualization; got.Containers != 3 || got.TotalItems != 15 || got.VisibleItems != 12 || got.CulledItems != 3 || got.Viewports != 3 || got.LastViewportWidth != 6 || got.LastViewportHeight != 9 || got.NonVirtualizedWarnings != 1 {
		t.Fatalf("unexpected virtualization statistics: %#v", got)
	}
	if got := stats.Cache; got.TextHits != 1 || got.TextMisses != 1 || got.StaticPaintHits != 1 || got.StaticPaintMisses != 1 || got.StaticTreeHits != 1 || got.StaticTreeMisses != 1 {
		t.Fatalf("unexpected cache statistics: %#v", got)
	}
	if !strings.Contains(output.String(), "frame=1") || !strings.Contains(output.String(), "redraw=") {
		t.Fatalf("redraw diagnostics log = %q", output.String())
	}

	// LastFrameStats must not expose the runtime's maps or sorted-record slice.
	stats.ReasonCounts["during-frame"] = 99
	stats.RedrawCounts[ownedKey] = 99
	stats.RedrawReasons[0].Count = 99
	again := runtime.LastFrameStats()
	if again.ReasonCounts["during-frame"] != 2 || again.RedrawCounts[ownedKey] != 1 || again.RedrawReasons[0].Count == 99 {
		t.Fatalf("LastFrameStats returned mutable runtime state: %#v", again)
	}
}

func TestPerfDiagnosticsConfigurationAndDisabledCalls(t *testing.T) {
	var nilRuntime *Runtime
	nilRuntime.SetPerfDiagnostics(PerfDiagnostics{Enabled: true})
	if config := nilRuntime.PerfDiagnostics(); config.Enabled || config.Writer != nil {
		t.Fatalf("nil runtime diagnostics = %#v", config)
	}
	if stats := nilRuntime.LastFrameStats(); stats.Frame != 0 {
		t.Fatalf("nil runtime frame stats = %#v", stats)
	}
	if nilRuntime.StartFrameSection(PerfLayout, 1) != nil {
		t.Fatal("nil runtime returned a frame-section completion")
	}

	runtime := NewRuntime(nil)
	runtime.RecordRedrawReason("ignored")
	runtime.RecordFrameSection(PerfLayout, 1)
	runtime.RecordViewport(image.Rect(0, 0, 1, 1))
	runtime.RecordVirtualizedItems(1, 1, image.Rect(0, 0, 1, 1))
	runtime.RecordNonVirtualizedItems(1)
	runtime.RecordTextCache(true)
	runtime.RecordStaticPaintCache(true)
	runtime.RecordStaticSubtreeCache(true)
	if runtime.eventDiagnosticsEnabled() || runtime.eventListenerDurationEnabled() {
		t.Fatal("disabled perf diagnostics reported as enabled")
	}
	if runtime.StartFrameSection(PerfLayout, 1) != nil {
		t.Fatal("disabled perf diagnostics returned a frame-section completion")
	}

	runtime.SetPerfDiagnostics(PerfDiagnostics{Enabled: true, LogEvents: true})
	if config := runtime.PerfDiagnostics(); !config.Enabled || !config.MeasureDurations || config.Writer == nil {
		t.Fatalf("enabled logging diagnostics were not normalized: %#v", config)
	}
	if !runtime.eventDiagnosticsEnabled() || !runtime.eventListenerDurationEnabled() {
		t.Fatal("enabled perf diagnostics were not reported to event helpers")
	}
	runtime.BeginFrame()
	runtime.RecordRedrawReason("will-reset")
	runtime.EndFrame()
	if runtime.LastFrameStats().Frame != 1 {
		t.Fatal("enabled diagnostics did not retain the completed frame")
	}

	runtime.SetPerfDiagnostics(PerfDiagnostics{})
	if config := runtime.PerfDiagnostics(); config.Enabled || config.MeasureDurations || config.Writer != nil {
		t.Fatalf("disabled diagnostics were not reset: %#v", config)
	}
	if stats := runtime.LastFrameStats(); stats.Frame != 0 || stats.ReasonCounts != nil {
		t.Fatalf("disabled diagnostics retained frame state: %#v", stats)
	}
}

func TestPerfEventDiagnosticsRecordsAndLogs(t *testing.T) {
	runtime := NewRuntime(nil)
	var output bytes.Buffer
	runtime.SetPerfDiagnostics(PerfDiagnostics{Enabled: true, LogEvents: true, Writer: &output})
	runtime.BeginFrame()
	child := runtime.childPath(rootPathID, 8)
	event := &Event{
		Type:                        "click",
		Target:                      child,
		DefaultPrevented:            true,
		propagationStopped:          true,
		immediatePropagationStopped: true,
		preventDefaultTarget:        child,
		preventDefaultPhase:         EventPhaseTarget,
		passivePreventDefaultTarget: child,
		passivePreventDefaultPhase:  EventPhaseCapture,
		propagationStopTarget:       rootPathID,
		propagationStopPhase:        EventPhaseBubble,
		immediateStopTarget:         child,
		immediateStopPhase:          EventPhaseTarget,
	}
	runtime.recordEventDispatch(nil, nil, 0, 0, true)
	runtime.recordEventRegistryCounts(3, 4, 5, 6)
	if current := runtime.perf.current.Events; current.TargetsRegistered != 3 || current.ListenersRegistered != 4 || current.FocusTargetsRegistered != 5 || current.ShortcutListenersRegistered != 6 {
		t.Fatalf("event registry counts were not recorded: %#v", current)
	}
	runtime.recordEventDispatch(event, []PathID{child, rootPathID}, 2, 3*time.Millisecond, false)
	runtime.EndFrame()

	stats := runtime.LastFrameStats().Events
	if stats.Dispatches != 1 || stats.ListenerCalls != 2 || stats.ListenerDuration != 3*time.Millisecond || stats.DefaultPrevented != 1 || stats.PropagationStopped != 1 || stats.ImmediatePropagationStopped != 1 || stats.PassivePreventDefault != 1 || stats.PointerEvents != 1 {
		t.Fatalf("unexpected event diagnostics: %#v", stats)
	}
	if stats.LastType != "click" || stats.LastTarget != "root/8" || stats.LastPath != "root/8>root" || !stats.LastDefaultPrevented || !stats.LastPropagationStopped || !stats.LastImmediateStopped || stats.LastDefaultAllowed {
		t.Fatalf("unexpected last-event diagnostics: %#v", stats)
	}
	if stats.LastPreventDefaultTarget != "root/8" || stats.LastPreventDefaultPhase != "target" || stats.LastPassivePreventTarget != "root/8" || stats.LastPassivePreventPhase != "capture" || stats.LastStopTarget != "root" || stats.LastStopPhase != "bubble" || stats.LastImmediateStopTarget != "root/8" || stats.LastImmediateStopPhase != "target" {
		t.Fatalf("event provenance diagnostics were incomplete: %#v", stats)
	}
	if !strings.Contains(output.String(), `event type="click"`) || !strings.Contains(output.String(), `listeners=2`) {
		t.Fatalf("event diagnostics log = %q", output.String())
	}
}

func TestPerfDiagnosticHelpers(t *testing.T) {
	for _, test := range []struct {
		eventType EventType
		want      string
	}{
		{"pointerdown", "pointer"},
		{"wheel", "wheel"},
		{EventTypeKeyDown, "keyboard"},
		{EventTypeFocus, "focus"},
		{EventTypeInput, "input"},
		{"dragstart", "drag"},
		{EventTypeActivate, "activation"},
		{"custom", "custom"},
	} {
		if got := eventDiagnosticsKind(test.eventType); got != test.want {
			t.Errorf("eventDiagnosticsKind(%q) = %q, want %q", test.eventType, got, test.want)
		}
	}
	for _, test := range []struct {
		phase EventPhase
		want  string
	}{
		{EventPhaseCapture, "capture"},
		{EventPhaseTarget, "target"},
		{EventPhaseBubble, "bubble"},
		{EventPhaseNone, ""},
	} {
		if got := eventPhaseDiagnostics(test.phase); got != test.want {
			t.Errorf("eventPhaseDiagnostics(%d) = %q, want %q", test.phase, got, test.want)
		}
	}

	runtime := NewRuntime(nil)
	if runtime.eventPathRewriteSummary(nil) != "" || runtime.debugEventTarget(0) != "" || runtime.structuralEventParent(rootPathID) != 0 {
		t.Fatal("empty diagnostic helpers returned a value")
	}
	runtime.BeginFrame()
	portal := runtime.childPath(rootPathID, 1)
	owner := runtime.scopePath(rootPathID, "owner")
	stop := runtime.childPath(rootPathID, 2)
	redirect := runtime.childPath(rootPathID, 3)
	runtime.RegisterEventTarget(portal, owner)
	runtime.RegisterEventTarget(stop, rootPathID)
	runtime.RegisterEventTargetOptions(stop, rootPathID, EventTargetOptions{Boundary: EventBoundaryPolicy{Mode: EventBoundaryStop}})
	runtime.RegisterEventTarget(redirect, rootPathID)
	runtime.RegisterEventTargetOptions(redirect, rootPathID, EventTargetOptions{Boundary: EventBoundaryPolicy{Mode: EventBoundaryRedirect, Redirect: owner}})

	summary := runtime.eventPathRewriteSummary([]PathID{portal, stop, redirect})
	for _, want := range []string{
		"portal:root/1->root/owner",
		"boundary-stop:root/2",
		"boundary-redirect:root/3->root/owner",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("path rewrite summary %q does not contain %q", summary, want)
		}
	}
	if runtime.debugEventTarget(portal) != "root/1" || runtime.structuralEventParent(portal) != rootPathID {
		t.Fatalf("unexpected event path debug data: target=%q parent=%d", runtime.debugEventTarget(portal), runtime.structuralEventParent(portal))
	}
	runtime.EndFrame()

	formatted := formatEventLogRecord(eventLogRecord{Type: "input", Target: "root", Path: "root", ListenerCalls: 2, DefaultAllowed: true})
	if !strings.Contains(formatted, `event type="input"`) || !strings.Contains(formatted, "default_allowed=true") {
		t.Fatalf("formatted event log = %q", formatted)
	}
}

func TestPerfStatisticsHelpersAndFormatting(t *testing.T) {
	var state runtimePerfState
	state.addViewportLocked(image.Rectangle{})
	state.addViewportLocked(image.Rect(0, 0, 4, 5))
	state.addSectionCountLocked(PerfLayout, 2)
	state.addSectionCountLocked(PerfSection(99), 5)
	state.addSectionDurationLocked(PerfDraw, 0)
	state.addSectionDurationLocked(PerfDraw, time.Millisecond)
	state.addSectionDurationLocked(PerfSection(99), time.Millisecond)
	if state.current.Virtualization.Viewports != 1 || state.current.Virtualization.LastViewportWidth != 4 || state.current.Virtualization.LastViewportHeight != 5 || state.current.Layout.Count != 2 || state.current.Draw.Duration != time.Millisecond || state.sectionLocked(PerfSection(99)) != nil {
		t.Fatalf("unexpected perf-state helpers: %#v", state.current)
	}

	base := FrameStats{
		Frame: 7,
		ReasonCounts: map[string]int{
			"zeta":  1,
			"alpha": 2,
		},
		RedrawCounts: map[RedrawReasonKey]int{
			{Reason: "zeta", Source: "runtime"}:                       1,
			{Reason: "alpha", Source: "context", OwnerPath: "root/1"}: 2,
		},
		Interaction: InteractionFrameStats{PointerMoves: 1},
		Events:      EventDiagnosticsStats{Dispatches: 1},
		Layout:      FrameSectionStats{Count: 3},
	}
	cloned := cloneFrameStats(base)
	if len(cloned.Reasons) != 2 || cloned.Reasons[0] != "alpha" || len(cloned.RedrawReasons) != 2 || cloned.RedrawReasons[0].Reason != "alpha" {
		t.Fatalf("cloned statistics were not sorted: %#v", cloned)
	}
	base.ReasonCounts["alpha"] = 99
	for key := range base.RedrawCounts {
		base.RedrawCounts[key] = 99
	}
	if cloned.ReasonCounts["alpha"] != 2 || cloned.RedrawReasons[0].Count != 2 {
		t.Fatalf("cloned statistics share source maps: %#v", cloned)
	}
	if empty := cloneFrameStats(FrameStats{}); empty.ReasonCounts != nil || empty.Reasons != nil || empty.RedrawCounts != nil || empty.RedrawReasons != nil {
		t.Fatalf("empty statistics clone retained empty collections: %#v", empty)
	}
	if formatted := FormatFrameStats(FrameStats{}); !strings.Contains(formatted, "reason=none") || !strings.Contains(formatted, "redraw=none") {
		t.Fatalf("empty frame formatting = %q", formatted)
	}
	if formatted := FormatFrameStats(cloned); !strings.Contains(formatted, "reason=alpha:2,zeta") || !strings.Contains(formatted, "redraw=alpha@context@root/1:2,zeta@runtime") || !strings.Contains(formatted, "layout_ops=3") {
		t.Fatalf("frame formatting = %q", formatted)
	}
}

func TestInteractionPointerQueueAndCoalescing(t *testing.T) {
	var nilRuntime *Runtime
	nilRuntime.QueuePointerMove(image.Pt(1, 1))
	nilRuntime.QueuePointerMoveF32(f32.Point{X: 1, Y: 1})
	if position, ok := nilRuntime.CoalescedPointerMove(); position != (image.Point{}) || ok {
		t.Fatalf("nil runtime coalesced pointer = %v, %t", position, ok)
	}

	runtime := NewRuntime(nil)
	runtime.QueuePointerMove(image.Pt(1, 2))
	runtime.QueuePointerMove(image.Pt(3, 4))
	if position, ok := runtime.CoalescedPointerMove(); !ok || position != image.Pt(3, 4) {
		t.Fatalf("coalesced pointer = %v, %t", position, ok)
	}
	runtime.QueuePointerMoveF32(f32.Point{X: 5.2, Y: 6.7})
	if position, ok := runtime.CoalescedPointerMove(); !ok || position != image.Pt(5, 7) {
		t.Fatalf("f32 coalesced pointer = %v, %t", position, ok)
	}
	if _, ok := runtime.CoalescedPointerMove(); ok {
		t.Fatal("coalesced pointer move was not cleared")
	}

	runtime.BeginFrame()
	runtime.QueuePointerMove(image.Pt(9, 10))
	if position, ok := runtime.CoalescedPointerMove(); !ok || position != image.Pt(9, 10) {
		t.Fatalf("frame pointer = %v, %t", position, ok)
	}
	runtime.EndFrame()
	if stats := runtime.LastInteractionStats(); stats.Frame != 1 || stats.PointerMoves != 3 {
		t.Fatalf("unexpected pointer frame statistics: %#v", stats)
	}
}

func TestInteractionSnapshotsEventsAndRedrawReasons(t *testing.T) {
	runtime := NewRuntime(nil)
	runtime.SetPerfDiagnostics(PerfDiagnostics{Enabled: true})
	redraws := 0
	runtime.SetInvalidator(func() { redraws++ })
	child := runtime.childPath(rootPathID, 4)
	if stats := runtime.ObserveInteractionSnapshot(0, ClickableSnapshot{}, ClickableSnapshot{}, true); stats != (InteractionFrameStats{}) {
		t.Fatalf("zero target interaction snapshot = %#v", stats)
	}
	runtime.ObserveEventDispatch("")

	runtime.BeginFrame()
	initial := runtime.ObserveInteractionSnapshot(child, ClickableSnapshot{}, ClickableSnapshot{Hovered: true}, false)
	if initial.Frame != 1 || initial.PressedChanged != 0 || initial.FocusChanged != 0 {
		t.Fatalf("initial interaction snapshot = %#v", initial)
	}
	changed := runtime.ObserveInteractionSnapshot(child, ClickableSnapshot{}, ClickableSnapshot{Hovered: true, Pressed: true, Focused: true}, true)
	if changed.PressedChanged != 1 || changed.FocusChanged != 1 {
		t.Fatalf("interaction changes were not recorded: %#v", changed)
	}
	runtime.ObserveEventDispatch("pointerdown")
	runtime.ObserveEventDispatch("wheel")
	runtime.ObserveEventDispatch(EventTypeKeyDown)
	runtime.ObserveEventDispatch(EventTypeFocus)
	runtime.ObserveEventDispatch("custom")
	runtime.EndFrame()

	stats := runtime.LastInteractionStats()
	if stats.PointerEvents != 1 || stats.WheelEvents != 1 || stats.KeyboardEvents != 1 || stats.FocusEvents != 1 || stats.HoverChanged != 1 || stats.PressedChanged != 1 || stats.FocusChanged != 1 || stats.HoverTarget != "root/4" || stats.HoverTargetPrev != "" {
		t.Fatalf("unexpected interaction frame statistics: %#v", stats)
	}
	frame := runtime.LastFrameStats()
	if frame.ReasonCounts[string(InteractionPressedChanged)] != 1 || frame.ReasonCounts[string(InteractionFocusChanged)] != 1 || frame.ReasonCounts[string(InteractionHoverChanged)] != 1 || redraws != 3 {
		t.Fatalf("interaction redraw diagnostics = reasons %#v redraws %d", frame.ReasonCounts, redraws)
	}

	runtime.BeginFrame()
	runtime.ObserveInteractionSnapshot(child, ClickableSnapshot{}, ClickableSnapshot{Hovered: true}, false)
	runtime.EndFrame()
	if stats := runtime.LastInteractionStats(); stats.HoverChanged != 0 || stats.HoverTarget != "root/4" {
		t.Fatalf("stable hover target changed: %#v", stats)
	}

	runtime.BeginFrame()
	runtime.EndFrame()
	if stats := runtime.LastInteractionStats(); stats.HoverChanged != 1 || stats.HoverTarget != "" || stats.HoverTargetPrev != "root/4" {
		t.Fatalf("cleared hover target statistics = %#v", stats)
	}
}

func TestInteractionHelpers(t *testing.T) {
	for _, test := range []struct {
		eventType EventType
		want      string
	}{
		{"pointermove", "pointer"},
		{"wheel", "wheel"},
		{EventTypeKeyUp, "keyboard"},
		{EventTypeFocusOut, "focus"},
		{"custom", ""},
	} {
		if got := eventInteractionKind(test.eventType); got != test.want {
			t.Errorf("eventInteractionKind(%q) = %q, want %q", test.eventType, got, test.want)
		}
	}

	var nilRuntime *Runtime
	if nilRuntime.LastInteractionStats() != (InteractionFrameStats{}) || nilRuntime.debugInteractionTarget(rootPathID) != "" {
		t.Fatal("nil interaction helpers returned state")
	}
	nilRuntime.beginInteractionFrame()
	nilRuntime.endInteractionFrame()
	nilRuntime.ObserveEventDispatch(EventTypeFocus)

	runtime := NewRuntime(nil)
	if runtime.debugInteractionTarget(0) != "" || runtime.debugInteractionTarget(rootPathID) != "root" || runtime.debugInteractionTarget(PathID(99)) != "path#99" {
		t.Fatalf("unexpected interaction target debug strings: zero=%q root=%q unknown=%q", runtime.debugInteractionTarget(0), runtime.debugInteractionTarget(rootPathID), runtime.debugInteractionTarget(PathID(99)))
	}
}
