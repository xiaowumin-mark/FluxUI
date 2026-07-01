package internal

import (
	"fmt"
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
	Writer           io.Writer
}

// FrameSectionStats contains aggregate data for one frame section.
type FrameSectionStats struct {
	Duration time.Duration
	Count    int64
}

// FrameStats is the coarse performance snapshot for one rendered frame.
type FrameStats struct {
	Frame        uint64
	StartedAt    time.Time
	Duration     time.Duration
	Layout       FrameSectionStats
	Draw         FrameSectionStats
	Animation    FrameSectionStats
	State        FrameSectionStats
	Text         FrameSectionStats
	Input        FrameSectionStats
	Reasons      []string
	ReasonCounts map[string]int
}

type runtimePerfState struct {
	mu                 sync.Mutex
	config             PerfDiagnostics
	enabled            atomic.Bool
	measureDurations   atomic.Bool
	logRedrawReasons   atomic.Bool
	frameSeq           uint64
	current            FrameStats
	last               FrameStats
	pendingReasons     map[string]int
	currentFrameActive bool
}

func (r *Runtime) SetPerfDiagnostics(config PerfDiagnostics) {
	if r == nil {
		return
	}
	if config.Enabled && !config.MeasureDurations {
		config.MeasureDurations = true
	}
	if config.Enabled && config.LogRedrawReasons && config.Writer == nil {
		config.Writer = os.Stderr
	}

	r.perf.mu.Lock()
	r.perf.config = config
	if !config.Enabled {
		r.perf.current = FrameStats{}
		r.perf.last = FrameStats{}
		r.perf.pendingReasons = nil
		r.perf.currentFrameActive = false
	}
	r.perf.mu.Unlock()

	r.perf.enabled.Store(config.Enabled)
	r.perf.measureDurations.Store(config.Enabled && config.MeasureDurations)
	r.perf.logRedrawReasons.Store(config.Enabled && config.LogRedrawReasons)
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
	if r == nil || !r.perf.enabled.Load() || reason == "" {
		return
	}
	r.perf.mu.Lock()
	r.perf.addReasonLocked(reason)
	r.perf.mu.Unlock()
}

// RequestRedrawReason requests a redraw and records the caller's reason.
func (r *Runtime) RequestRedrawReason(reason string) {
	if reason == "" {
		reason = "RequestRedraw"
	}
	r.RecordRedrawReason(reason)
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

func (r *Runtime) beginPerfFrame() {
	if r == nil || !r.perf.enabled.Load() {
		return
	}
	r.perf.mu.Lock()
	r.perf.frameSeq++
	stats := FrameStats{
		Frame:        r.perf.frameSeq,
		ReasonCounts: make(map[string]int, len(r.perf.pendingReasons)),
	}
	if r.perf.measureDurations.Load() {
		stats.StartedAt = time.Now()
	}
	for reason, count := range r.perf.pendingReasons {
		stats.ReasonCounts[reason] = count
	}
	clear(r.perf.pendingReasons)
	r.perf.current = stats
	r.perf.currentFrameActive = true
	r.perf.mu.Unlock()
}

func (r *Runtime) endPerfFrame() {
	if r == nil || !r.perf.enabled.Load() {
		return
	}

	var (
		stats  FrameStats
		writer io.Writer
		log    bool
	)

	r.perf.mu.Lock()
	if r.perf.measureDurations.Load() && !r.perf.current.StartedAt.IsZero() {
		r.perf.current.Duration = time.Since(r.perf.current.StartedAt)
	}
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

func (p *runtimePerfState) addReasonLocked(reason string) {
	if p.pendingReasons == nil {
		p.pendingReasons = make(map[string]int, 4)
	}
	p.pendingReasons[reason]++
	if p.currentFrameActive {
		if p.current.ReasonCounts == nil {
			p.current.ReasonCounts = make(map[string]int, 4)
		}
		p.current.ReasonCounts[reason]++
	}
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

	return fmt.Sprintf(
		"frame=%d duration=%s reason=%s layout=%s draw=%s animation=%s state=%s text=%s input=%s layout_ops=%d draw_ops=%d animations=%d state_ops=%d text_ops=%d input_ops=%d",
		stats.Frame,
		stats.Duration,
		reason,
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
