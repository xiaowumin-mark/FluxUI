package anim

import (
	"fmt"
	"time"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
)

// Option 定义动画配置项。
type Option func(*Animation)

// Animation 是 frame 驱动的 tween 动画。
type Animation struct {
	from     float32
	to       float32
	duration time.Duration
	easing   Easing
}

type track struct {
	startedAt time.Time
	from      float32
	to        float32
	duration  time.Duration
}

// New 创建动画定义。
func New(opts ...Option) *Animation {
	anim := &Animation{
		from:     0,
		to:       1,
		duration: 300 * time.Millisecond,
		easing:   Linear,
	}
	for _, opt := range opts {
		opt(anim)
	}
	return anim
}

// Duration 设置持续时间。
func Duration(duration time.Duration) Option {
	return func(animation *Animation) {
		animation.duration = duration
	}
}

// From 设置起始值。
func From(value float32) Option {
	return func(animation *Animation) {
		animation.from = value
	}
}

// To 设置结束值。
func To(value float32) Option {
	return func(animation *Animation) {
		animation.to = value
	}
}

// Ease 设置缓动函数。
func Ease(easing Easing) Option {
	return func(animation *Animation) {
		if easing != nil {
			animation.easing = easing
		}
	}
}

// Value 返回当前 frame 的动画值。
func (a *Animation) Value(ctx *internal.Context) float32 {
	if a == nil {
		return 0
	}

	value := ctx.Memo("animation", func() any {
		return &track{}
	})

	timeline, ok := value.(*track)
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUIanim: key %q 的动画轨道类型错误", ctx.TreePath()))
	}

	now := ctx.Now()
	if timeline.startedAt.IsZero() || timeline.from != a.from || timeline.to != a.to || timeline.duration != a.duration {
		timeline.startedAt = now
		timeline.from = a.from
		timeline.to = a.to
		timeline.duration = a.duration
	}

	if a.duration <= 0 {
		return a.to
	}

	elapsed := now.Sub(timeline.startedAt)
	if elapsed < a.duration {
		ctx.RequestRedraw()
	}

	progress := clamp01(float32(elapsed) / float32(a.duration))
	return lerp(a.from, a.to, a.easing(progress))
}
