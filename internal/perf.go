package internal

import (
	"fmt"
	"image"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// PerfSection identifies one coarse frame cost bucket.
type PerfSection int

const (
	PerfLayout PerfSection = iota
	PerfDraw
	PerfAnimation
	PerfState
	PerfText
	PerfInput
)

func (s PerfSection) String() string {
	switch s {
	case PerfLayout:
		return "layout"
	case PerfDraw:
		return "draw"
	case PerfAnimation:
		return "animation"
	case PerfState:
		return "state"
	case PerfText:
		return "text"
	case PerfInput:
		return "input"
	default:
		return "unknown"
	}
}

// PerfDiagnostics controls runtime frame diagnostics.
type PerfDiagnostics struct {
	Enabled          bool
	MeasureDurations bool
	LogRedrawReasons bool
	LogEvents        bool
	Writer           io.Writer
}

// FrameSectionStats contains aggregate data for one frame section.
type FrameSectionStats struct {
	Duration time.Duration
	Count    int64
}

// VirtualizationStats summarizes viewport-based list/grid clipping for one
// frame. It counts items whose widgets were actually built and laid out.
type VirtualizationStats struct {
	Containers             int64
	Viewports              int64
	TotalItems             int64
	VisibleItems           int64
	CulledItems            int64
	NonVirtualizedWarnings int64
	LastViewportWidth      int
	LastViewportHeight     int
}

// RenderCacheStats summarizes static render cache reuse for one frame.
type RenderCacheStats struct {
	TextHits          int64
	TextMisses        int64
	StaticPaintHits   int64
	StaticPaintMisses int64
	StaticTreeHits    int64
	StaticTreeMisses  int64
}

// RedrawReasonKey identifies one structured redraw source.
type RedrawReasonKey struct {
	Reason    string
	Source    string
	OwnerPath string
}

// RedrawReasonStats contains one structured redraw reason aggregate.
type RedrawReasonStats struct {
	Reason    string
	Source    string
	OwnerPath string
	Count     int
}

// EventDiagnosticsStats summarizes event dispatch work observed in one frame.
type EventDiagnosticsStats struct {
	Dispatches                  int64
	ListenerCalls               int64
	ListenerDuration            time.Duration
	TargetsRegistered           int64
	ListenersRegistered         int64
	FocusTargetsRegistered      int64
	ShortcutListenersRegistered int64
	DefaultPrevented            int64
	PropagationStopped          int64
	ImmediatePropagationStopped int64
	PassivePreventDefault       int64
	PointerEvents               int64
	WheelEvents                 int64
	KeyboardEvents              int64
	FocusEvents                 int64
	InputEvents                 int64
	DragEvents                  int64
	CustomEvents                int64
	ActivationEvents            int64
	LastType                    string
	LastTarget                  string
	LastPath                    string
	LastDefaultPrevented        bool
	LastPropagationStopped      bool
	LastImmediateStopped        bool
	LastDefaultAllowed          bool
	LastPreventDefaultTarget    string
	LastPreventDefaultPhase     string
	LastPassivePreventTarget    string
	LastPassivePreventPhase     string
	LastStopTarget              string
	LastStopPhase               string
	LastImmediateStopTarget     string
	LastImmediateStopPhase      string
	LastPathRewrite             string
}

// FrameStats is the coarse performance snapshot for one rendered frame.
type FrameStats struct {
	Frame          uint64
	StartedAt      time.Time
	Duration       time.Duration
	Interaction    InteractionFrameStats
	Events         EventDiagnosticsStats
	Virtualization VirtualizationStats
	Cache          RenderCacheStats
	Layout         FrameSectionStats
	Draw           FrameSectionStats
	Animation      FrameSectionStats
	State          FrameSectionStats
	Text           FrameSectionStats
	Input          FrameSectionStats
	Reasons        []string
	ReasonCounts   map[string]int
	RedrawReasons  []RedrawReasonStats
	RedrawCounts   map[RedrawReasonKey]int
}

type runtimePerfState struct {
	mu                 sync.Mutex
	config             PerfDiagnostics
	enabled            atomic.Bool
	measureDurations   atomic.Bool
	logRedrawReasons   atomic.Bool
	logEvents          atomic.Bool
	frameSeq           uint64
	current            FrameStats
	last               FrameStats
	pendingReasons     map[string]int
	pendingRedraw      map[RedrawReasonKey]int
	currentFrameActive bool
}

func (r *Runtime) SetPerfDiagnostics(config PerfDiagnostics) {
	if r == nil {
		return
	}
	if config.Enabled && !config.MeasureDurations {
		config.MeasureDurations = true
	}
	if config.Enabled && (config.LogRedrawReasons || config.LogEvents) && config.Writer == nil {
		config.Writer = os.Stderr
	}

	r.perf.mu.Lock()
	r.perf.config = config
	if !config.Enabled {
		r.perf.current = FrameStats{}
		r.perf.last = FrameStats{}
		r.perf.pendingReasons = nil
		r.perf.pendingRedraw = nil
		r.perf.currentFrameActive = false
	}
	r.perf.mu.Unlock()

	r.perf.enabled.Store(config.Enabled)
	r.perf.measureDurations.Store(config.Enabled && config.MeasureDurations)
	r.perf.logRedrawReasons.Store(config.Enabled && config.LogRedrawReasons)
	r.perf.logEvents.Store(config.Enabled && config.LogEvents)
}

func (r *Runtime) PerfDiagnostics() PerfDiagnostics {
	if r == nil {
		return PerfDiagnostics{}
	}
	r.perf.mu.Lock()
	config := r.perf.config
	r.perf.mu.Unlock()
	return config
}

func (r *Runtime) LastFrameStats() FrameStats {
	if r == nil || !r.perf.enabled.Load() {
		return FrameStats{}
	}
	r.perf.mu.Lock()
	stats := cloneFrameStats(r.perf.last)
	r.perf.mu.Unlock()
	return stats
}

// RecordRedrawReason records why a redraw was requested.
func (r *Runtime) RecordRedrawReason(reason string) {
	r.recordRedrawReason(reason, "runtime", 0)
}

func (r *Runtime) recordRedrawReason(reason, source string, owner PathID) {
	if r == nil || !r.perf.enabled.Load() || reason == "" {
		return
	}
	if source == "" {
		source = "runtime"
	}
	ownerPath := ""
	if owner != 0 {
		ownerPath = r.debugInteractionTarget(owner)
	}
	r.perf.mu.Lock()
	r.perf.addReasonLocked(reason, source, ownerPath)
	r.perf.mu.Unlock()
}

// RequestRedrawReason requests a redraw and records the caller's reason.
func (r *Runtime) RequestRedrawReason(reason string) {
	if reason == "" {
		reason = "RequestRedraw"
	}
	r.recordRedrawReason(reason, "runtime", 0)
	r.requestRedraw()
}

// RecordFrameSection increments the section operation count.
func (r *Runtime) RecordFrameSection(section PerfSection, count int64) {
	if r == nil || !r.perf.enabled.Load() || count == 0 {
		return
	}
	r.perf.mu.Lock()
	r.perf.addSectionCountLocked(section, count)
	r.perf.mu.Unlock()
}

// RecordViewport records that a scrollable or virtualized container exposed a
// viewport to its descendants.
func (r *Runtime) RecordViewport(viewport image.Rectangle) {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	r.perf.mu.Lock()
	r.perf.addViewportLocked(viewport)
	r.perf.mu.Unlock()
}

// RecordVirtualizedItems records one virtualized container's total and
// actually built item count.
func (r *Runtime) RecordVirtualizedItems(total, visible int, viewport image.Rectangle) {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	if total < 0 {
		total = 0
	}
	if visible < 0 {
		visible = 0
	}
	if visible > total {
		visible = total
	}
	r.perf.mu.Lock()
	r.perf.current.Virtualization.Containers++
	r.perf.current.Virtualization.TotalItems += int64(total)
	r.perf.current.Virtualization.VisibleItems += int64(visible)
	r.perf.current.Virtualization.CulledItems += int64(total - visible)
	r.perf.addViewportLocked(viewport)
	r.perf.mu.Unlock()
}

// RecordNonVirtualizedItems records a diagnostic warning for a large container
// that opted out of viewport clipping.
func (r *Runtime) RecordNonVirtualizedItems(total int) {
	if r == nil || !r.perf.enabled.Load() || total <= 0 {
		return
	}
	r.perf.mu.Lock()
	r.perf.current.Virtualization.NonVirtualizedWarnings++
	r.perf.mu.Unlock()
}

func (r *Runtime) RecordTextCache(hit bool) {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	r.perf.mu.Lock()
	if hit {
		r.perf.current.Cache.TextHits++
	} else {
		r.perf.current.Cache.TextMisses++
	}
	r.perf.mu.Unlock()
}

func (r *Runtime) RecordStaticPaintCache(hit bool) {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	r.perf.mu.Lock()
	if hit {
		r.perf.current.Cache.StaticPaintHits++
	} else {
		r.perf.current.Cache.StaticPaintMisses++
	}
	r.perf.mu.Unlock()
}

func (r *Runtime) RecordStaticSubtreeCache(hit bool) {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	r.perf.mu.Lock()
	if hit {
		r.perf.current.Cache.StaticTreeHits++
	} else {
		r.perf.current.Cache.StaticTreeMisses++
	}
	r.perf.mu.Unlock()
}

// StartFrameSection records count and returns a completion function that adds duration.
func (r *Runtime) StartFrameSection(section PerfSection, count int64) func() {
	if r == nil || !r.perf.enabled.Load() {
		return nil
	}
	if count != 0 {
		r.RecordFrameSection(section, count)
	}
	if !r.perf.measureDurations.Load() {
		return nil
	}
	start := time.Now()
	return func() {
		r.perf.mu.Lock()
		r.perf.addSectionDurationLocked(section, time.Since(start))
		r.perf.mu.Unlock()
	}
}

func (c *Context) recordFrameSection(section PerfSection, count int64) {
	if c == nil || c.runtime == nil {
		return
	}
	c.runtime.RecordFrameSection(section, count)
}

func (c *Context) startFrameSection(section PerfSection, count int64) func() {
	if c == nil || c.runtime == nil {
		return nil
	}
	return c.runtime.StartFrameSection(section, count)
}

func (r *Runtime) eventDiagnosticsEnabled() bool {
	return r != nil && r.perf.enabled.Load()
}

func (r *Runtime) eventListenerDurationEnabled() bool {
	return r != nil && r.perf.enabled.Load() && r.perf.measureDurations.Load()
}

func (r *Runtime) recordEventDispatch(event *Event, path []PathID, listenerCalls int64, listenerDuration time.Duration, allowed bool) {
	if r == nil || !r.perf.enabled.Load() || event == nil {
		return
	}

	var (
		writer io.Writer
		log    bool
		record = r.newEventLogRecord(event, path, listenerCalls, listenerDuration, allowed)
	)
	if r.perf.logEvents.Load() {
		log = true
	}

	r.perf.mu.Lock()
	stats := &r.perf.current.Events
	stats.Dispatches++
	stats.ListenerCalls += listenerCalls
	stats.ListenerDuration += listenerDuration
	if event.DefaultPrevented {
		stats.DefaultPrevented++
	}
	if event.PropagationStopped() {
		stats.PropagationStopped++
	}
	if event.ImmediatePropagationStopped() {
		stats.ImmediatePropagationStopped++
	}
	if event.passivePreventDefaultTarget != 0 {
		stats.PassivePreventDefault++
	}
	switch eventDiagnosticsKind(event.Type) {
	case "pointer":
		stats.PointerEvents++
	case "wheel":
		stats.WheelEvents++
	case "keyboard":
		stats.KeyboardEvents++
	case "focus":
		stats.FocusEvents++
	case "input":
		stats.InputEvents++
	case "drag":
		stats.DragEvents++
	case "activation":
		stats.ActivationEvents++
	default:
		stats.CustomEvents++
	}
	stats.LastType = record.Type
	stats.LastTarget = record.Target
	stats.LastPath = record.Path
	stats.LastDefaultPrevented = record.DefaultPrevented
	stats.LastPropagationStopped = record.PropagationStop
	stats.LastImmediateStopped = record.ImmediateStop
	stats.LastDefaultAllowed = record.DefaultAllowed
	stats.LastPreventDefaultTarget = record.PreventDefaultTarget
	stats.LastPreventDefaultPhase = record.PreventDefaultPhase
	stats.LastPassivePreventTarget = record.PassivePreventTarget
	stats.LastPassivePreventPhase = record.PassivePreventPhase
	stats.LastStopTarget = record.StopTarget
	stats.LastStopPhase = record.StopPhase
	stats.LastImmediateStopTarget = record.ImmediateStopTarget
	stats.LastImmediateStopPhase = record.ImmediateStopPhase
	stats.LastPathRewrite = record.PathRewrite
	if log {
		writer = r.perf.config.Writer
	}
	r.perf.mu.Unlock()

	if log && writer != nil {
		_, _ = fmt.Fprintln(writer, formatEventLogRecord(record))
	}
}

func (r *Runtime) recordEventRegistryCounts(targets, listeners, focusTargets, shortcuts int64) {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	r.perf.mu.Lock()
	r.perf.current.Events.TargetsRegistered = targets
	r.perf.current.Events.ListenersRegistered = listeners
	r.perf.current.Events.FocusTargetsRegistered = focusTargets
	r.perf.current.Events.ShortcutListenersRegistered = shortcuts
	r.perf.mu.Unlock()
}

type eventLogRecord struct {
	Type                 string
	Target               string
	Path                 string
	ListenerCalls        int64
	ListenerDuration     time.Duration
	DefaultPrevented     bool
	PropagationStop      bool
	ImmediateStop        bool
	DefaultAllowed       bool
	PreventDefaultTarget string
	PreventDefaultPhase  string
	PassivePreventTarget string
	PassivePreventPhase  string
	StopTarget           string
	StopPhase            string
	ImmediateStopTarget  string
	ImmediateStopPhase   string
	PathRewrite          string
}

func (r *Runtime) newEventLogRecord(event *Event, path []PathID, listenerCalls int64, listenerDuration time.Duration, allowed bool) eventLogRecord {
	record := eventLogRecord{
		Type:             string(event.Type),
		Target:           r.debugInteractionTarget(event.Target),
		ListenerCalls:    listenerCalls,
		ListenerDuration: listenerDuration,
		DefaultPrevented: event.DefaultPrevented,
		PropagationStop:  event.PropagationStopped(),
		ImmediateStop:    event.ImmediatePropagationStopped(),
		DefaultAllowed:   allowed,
	}
	if record.Target == "" {
		record.Target = r.debugInteractionTarget(normalizeOptionalPathID(event.Target))
	}
	if len(path) > 0 {
		parts := make([]string, 0, len(path))
		for _, id := range path {
			parts = append(parts, r.debugInteractionTarget(id))
		}
		record.Path = strings.Join(parts, ">")
	}
	record.PreventDefaultTarget = r.debugEventTarget(event.preventDefaultTarget)
	record.PreventDefaultPhase = eventPhaseDiagnostics(event.preventDefaultPhase)
	record.PassivePreventTarget = r.debugEventTarget(event.passivePreventDefaultTarget)
	record.PassivePreventPhase = eventPhaseDiagnostics(event.passivePreventDefaultPhase)
	record.StopTarget = r.debugEventTarget(event.propagationStopTarget)
	record.StopPhase = eventPhaseDiagnostics(event.propagationStopPhase)
	record.ImmediateStopTarget = r.debugEventTarget(event.immediateStopTarget)
	record.ImmediateStopPhase = eventPhaseDiagnostics(event.immediateStopPhase)
	record.PathRewrite = r.eventPathRewriteSummary(path)
	return record
}

func formatEventLogRecord(record eventLogRecord) string {
	return fmt.Sprintf(
		"event type=%q target=%q path=%q listeners=%d listener_duration=%s default_prevented=%t propagation_stopped=%t immediate_stopped=%t default_allowed=%t prevent_default_target=%q prevent_default_phase=%q passive_prevent_target=%q passive_prevent_phase=%q stop_target=%q stop_phase=%q immediate_stop_target=%q immediate_stop_phase=%q path_rewrite=%q",
		record.Type,
		record.Target,
		record.Path,
		record.ListenerCalls,
		record.ListenerDuration,
		record.DefaultPrevented,
		record.PropagationStop,
		record.ImmediateStop,
		record.DefaultAllowed,
		record.PreventDefaultTarget,
		record.PreventDefaultPhase,
		record.PassivePreventTarget,
		record.PassivePreventPhase,
		record.StopTarget,
		record.StopPhase,
		record.ImmediateStopTarget,
		record.ImmediateStopPhase,
		record.PathRewrite,
	)
}

func (r *Runtime) debugEventTarget(target PathID) string {
	if target == 0 {
		return ""
	}
	return r.debugInteractionTarget(target)
}

func eventPhaseDiagnostics(phase EventPhase) string {
	switch phase {
	case EventPhaseCapture:
		return "capture"
	case EventPhaseTarget:
		return "target"
	case EventPhaseBubble:
		return "bubble"
	default:
		return ""
	}
}

func (r *Runtime) eventPathRewriteSummary(path []PathID) string {
	if r == nil || len(path) == 0 {
		return ""
	}
	var parts []string
	for _, id := range path {
		id = normalizePathID(id)
		entry, ok := r.events.targets[id]
		if !ok {
			continue
		}
		if parent := normalizeOptionalPathID(entry.Parent); parent != 0 {
			if structural := r.structuralEventParent(id); structural != 0 && parent != structural {
				parts = append(parts, fmt.Sprintf("portal:%s->%s", r.debugInteractionTarget(id), r.debugInteractionTarget(parent)))
			}
		}
		switch entry.Boundary.Mode {
		case EventBoundaryStop:
			parts = append(parts, "boundary-stop:"+r.debugInteractionTarget(id))
		case EventBoundaryRedirect:
			redirect := normalizeOptionalPathID(entry.Boundary.Redirect)
			if redirect == 0 {
				redirect = rootPathID
			}
			parts = append(parts, fmt.Sprintf("boundary-redirect:%s->%s", r.debugInteractionTarget(id), r.debugInteractionTarget(redirect)))
		}
	}
	return strings.Join(parts, ",")
}

func (r *Runtime) structuralEventParent(target PathID) PathID {
	if r == nil {
		return 0
	}
	target = normalizePathID(target)
	if target == rootPathID {
		return 0
	}
	r.initPathTable()
	if entry := r.pathDebug[target]; entry != nil {
		return normalizeOptionalPathID(entry.parent)
	}
	return 0
}

func (r *Runtime) beginPerfFrame() {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	r.perf.mu.Lock()
	r.perf.frameSeq++
	stats := FrameStats{
		Frame:        r.perf.frameSeq,
		ReasonCounts: make(map[string]int, len(r.perf.pendingReasons)),
		RedrawCounts: make(map[RedrawReasonKey]int, len(r.perf.pendingRedraw)),
	}
	if r.perf.measureDurations.Load() {
		stats.StartedAt = time.Now()
	}
	for reason, count := range r.perf.pendingReasons {
		stats.ReasonCounts[reason] = count
	}
	for key, count := range r.perf.pendingRedraw {
		stats.RedrawCounts[key] = count
	}
	clear(r.perf.pendingReasons)
	clear(r.perf.pendingRedraw)
	r.perf.current = stats
	r.perf.currentFrameActive = true
	r.perf.mu.Unlock()
}

func (r *Runtime) endPerfFrame() {
	if r == nil || !r.perf.enabled.Load() {
		return
	}

	var (
		stats       FrameStats
		interaction = r.LastInteractionStats()
		writer      io.Writer
		log         bool
	)

	r.perf.mu.Lock()
	if r.perf.measureDurations.Load() && !r.perf.current.StartedAt.IsZero() {
		r.perf.current.Duration = time.Since(r.perf.current.StartedAt)
	}
	r.perf.current.Interaction = interaction
	r.perf.currentFrameActive = false
	r.perf.last = cloneFrameStats(r.perf.current)
	stats = cloneFrameStats(r.perf.last)
	writer = r.perf.config.Writer
	log = r.perf.logRedrawReasons.Load()
	r.perf.mu.Unlock()

	if log && writer != nil {
		_, _ = fmt.Fprintln(writer, FormatFrameStats(stats))
	}
}

func (p *runtimePerfState) addReasonLocked(reason, source, ownerPath string) {
	if p.pendingReasons == nil {
		p.pendingReasons = make(map[string]int, 4)
	}
	if p.pendingRedraw == nil {
		p.pendingRedraw = make(map[RedrawReasonKey]int, 4)
	}
	key := RedrawReasonKey{Reason: reason, Source: source, OwnerPath: ownerPath}
	p.pendingReasons[reason]++
	p.pendingRedraw[key]++
	if p.currentFrameActive {
		if p.current.ReasonCounts == nil {
			p.current.ReasonCounts = make(map[string]int, 4)
		}
		if p.current.RedrawCounts == nil {
			p.current.RedrawCounts = make(map[RedrawReasonKey]int, 4)
		}
		p.current.ReasonCounts[reason]++
		p.current.RedrawCounts[key]++
	}
}

func (p *runtimePerfState) addViewportLocked(viewport image.Rectangle) {
	if viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		return
	}
	p.current.Virtualization.Viewports++
	p.current.Virtualization.LastViewportWidth = viewport.Dx()
	p.current.Virtualization.LastViewportHeight = viewport.Dy()
}

func (p *runtimePerfState) addSectionCountLocked(section PerfSection, count int64) {
	stats := p.sectionLocked(section)
	if stats == nil {
		return
	}
	stats.Count += count
}

func (p *runtimePerfState) addSectionDurationLocked(section PerfSection, duration time.Duration) {
	if duration <= 0 {
		return
	}
	stats := p.sectionLocked(section)
	if stats == nil {
		return
	}
	stats.Duration += duration
}

func (p *runtimePerfState) sectionLocked(section PerfSection) *FrameSectionStats {
	switch section {
	case PerfLayout:
		return &p.current.Layout
	case PerfDraw:
		return &p.current.Draw
	case PerfAnimation:
		return &p.current.Animation
	case PerfState:
		return &p.current.State
	case PerfText:
		return &p.current.Text
	case PerfInput:
		return &p.current.Input
	default:
		return nil
	}
}

func eventDiagnosticsKind(eventType EventType) string {
	switch eventType {
	case "pointerdown", "pointerup", "pointermove", "pointerenter", "pointerleave", "pointerover", "pointerout", "pointercancel", "click", "dblclick", "auxclick", "contextmenu":
		return "pointer"
	case "wheel":
		return "wheel"
	case EventTypeKeyDown, EventTypeKeyUp:
		return "keyboard"
	case EventTypeFocus, EventTypeBlur, EventTypeFocusIn, EventTypeFocusOut:
		return "focus"
	case EventTypeBeforeInput, EventTypeInput, EventTypeChange, EventTypeSubmit, EventTypeCompositionStart, EventTypeCompositionUpdate, EventTypeCompositionEnd:
		return "input"
	case "dragstart", "drag", "dragenter", "dragover", "dragleave", "drop", "dragend":
		return "drag"
	case EventTypeActivate:
		return "activation"
	default:
		return "custom"
	}
}

func cloneFrameStats(stats FrameStats) FrameStats {
	if len(stats.ReasonCounts) > 0 {
		counts := make(map[string]int, len(stats.ReasonCounts))
		reasons := make([]string, 0, len(stats.ReasonCounts))
		for reason, count := range stats.ReasonCounts {
			counts[reason] = count
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)
		stats.ReasonCounts = counts
		stats.Reasons = reasons
	} else {
		stats.ReasonCounts = nil
		stats.Reasons = nil
	}
	if len(stats.RedrawCounts) > 0 {
		counts := make(map[RedrawReasonKey]int, len(stats.RedrawCounts))
		records := make([]RedrawReasonStats, 0, len(stats.RedrawCounts))
		for key, count := range stats.RedrawCounts {
			counts[key] = count
			records = append(records, RedrawReasonStats{
				Reason:    key.Reason,
				Source:    key.Source,
				OwnerPath: key.OwnerPath,
				Count:     count,
			})
		}
		sort.Slice(records, func(i, j int) bool {
			if records[i].Reason != records[j].Reason {
				return records[i].Reason < records[j].Reason
			}
			if records[i].Source != records[j].Source {
				return records[i].Source < records[j].Source
			}
			return records[i].OwnerPath < records[j].OwnerPath
		})
		stats.RedrawCounts = counts
		stats.RedrawReasons = records
	} else {
		stats.RedrawCounts = nil
		stats.RedrawReasons = nil
	}
	return stats
}

// FormatFrameStats formats a compact one-line frame diagnostics record.
func FormatFrameStats(stats FrameStats) string {
	reason := "none"
	if len(stats.Reasons) > 0 {
		parts := make([]string, 0, len(stats.Reasons))
		for _, r := range stats.Reasons {
			count := stats.ReasonCounts[r]
			if count > 1 {
				parts = append(parts, fmt.Sprintf("%s:%d", r, count))
			} else {
				parts = append(parts, r)
			}
		}
		reason = strings.Join(parts, ",")
	}
	redraw := "none"
	if len(stats.RedrawReasons) > 0 {
		parts := make([]string, 0, len(stats.RedrawReasons))
		for _, r := range stats.RedrawReasons {
			label := r.Reason
			if r.Source != "" {
				label += "@" + r.Source
			}
			if r.OwnerPath != "" {
				label += "@" + r.OwnerPath
			}
			if r.Count > 1 {
				label = fmt.Sprintf("%s:%d", label, r.Count)
			}
			parts = append(parts, label)
		}
		redraw = strings.Join(parts, ",")
	}

	return fmt.Sprintf(
		"frame=%d duration=%s reason=%s redraw=%s pointer_moves=%d pointer_events=%d wheel_events=%d keyboard_events=%d focus_events=%d hover_changes=%d pressed_changes=%d focus_changes=%d hover_target=%q event_dispatches=%d event_listeners=%d event_listener_duration=%s event_targets=%d event_registered_listeners=%d event_focus_targets=%d event_shortcuts=%d event_default_prevented=%d event_propagation_stopped=%d event_immediate_stopped=%d event_passive_prevent_default=%d event_pointer=%d event_wheel=%d event_keyboard=%d event_focus=%d event_input=%d event_drag=%d event_custom=%d event_activation=%d event_last_type=%q event_last_target=%q event_last_path=%q event_last_default_prevented=%t event_last_propagation_stopped=%t event_last_immediate_stopped=%t event_last_default_allowed=%t event_last_prevent_default_target=%q event_last_prevent_default_phase=%q event_last_passive_prevent_target=%q event_last_passive_prevent_phase=%q event_last_stop_target=%q event_last_stop_phase=%q event_last_immediate_stop_target=%q event_last_immediate_stop_phase=%q event_last_path_rewrite=%q virtual_items=%d/%d virtual_culled=%d virtual_containers=%d viewports=%d viewport=%dx%d nonvirtual_warnings=%d text_cache=%d/%d static_paint_cache=%d/%d static_tree_cache=%d/%d layout=%s draw=%s animation=%s state=%s text=%s input=%s layout_ops=%d draw_ops=%d animations=%d state_ops=%d text_ops=%d input_ops=%d",
		stats.Frame,
		stats.Duration,
		reason,
		redraw,
		stats.Interaction.PointerMoves,
		stats.Interaction.PointerEvents,
		stats.Interaction.WheelEvents,
		stats.Interaction.KeyboardEvents,
		stats.Interaction.FocusEvents,
		stats.Interaction.HoverChanged,
		stats.Interaction.PressedChanged,
		stats.Interaction.FocusChanged,
		stats.Interaction.HoverTarget,
		stats.Events.Dispatches,
		stats.Events.ListenerCalls,
		stats.Events.ListenerDuration,
		stats.Events.TargetsRegistered,
		stats.Events.ListenersRegistered,
		stats.Events.FocusTargetsRegistered,
		stats.Events.ShortcutListenersRegistered,
		stats.Events.DefaultPrevented,
		stats.Events.PropagationStopped,
		stats.Events.ImmediatePropagationStopped,
		stats.Events.PassivePreventDefault,
		stats.Events.PointerEvents,
		stats.Events.WheelEvents,
		stats.Events.KeyboardEvents,
		stats.Events.FocusEvents,
		stats.Events.InputEvents,
		stats.Events.DragEvents,
		stats.Events.CustomEvents,
		stats.Events.ActivationEvents,
		stats.Events.LastType,
		stats.Events.LastTarget,
		stats.Events.LastPath,
		stats.Events.LastDefaultPrevented,
		stats.Events.LastPropagationStopped,
		stats.Events.LastImmediateStopped,
		stats.Events.LastDefaultAllowed,
		stats.Events.LastPreventDefaultTarget,
		stats.Events.LastPreventDefaultPhase,
		stats.Events.LastPassivePreventTarget,
		stats.Events.LastPassivePreventPhase,
		stats.Events.LastStopTarget,
		stats.Events.LastStopPhase,
		stats.Events.LastImmediateStopTarget,
		stats.Events.LastImmediateStopPhase,
		stats.Events.LastPathRewrite,
		stats.Virtualization.VisibleItems,
		stats.Virtualization.TotalItems,
		stats.Virtualization.CulledItems,
		stats.Virtualization.Containers,
		stats.Virtualization.Viewports,
		stats.Virtualization.LastViewportWidth,
		stats.Virtualization.LastViewportHeight,
		stats.Virtualization.NonVirtualizedWarnings,
		stats.Cache.TextHits,
		stats.Cache.TextHits+stats.Cache.TextMisses,
		stats.Cache.StaticPaintHits,
		stats.Cache.StaticPaintHits+stats.Cache.StaticPaintMisses,
		stats.Cache.StaticTreeHits,
		stats.Cache.StaticTreeHits+stats.Cache.StaticTreeMisses,
		stats.Layout.Duration,
		stats.Draw.Duration,
		stats.Animation.Duration,
		stats.State.Duration,
		stats.Text.Duration,
		stats.Input.Duration,
		stats.Layout.Count,
		stats.Draw.Count,
		stats.Animation.Count,
		stats.State.Count,
		stats.Text.Count,
		stats.Input.Count,
	)
}
