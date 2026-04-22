package router

import "time"

// Transition 定义页面切换动画类型。
type Transition int

const (
	// TransitionNone 无动画，直接切换。
	TransitionNone Transition = iota
	// TransitionFade 淡入淡出。
	TransitionFade
	// TransitionSlideLeft 从右向左滑入（前进）。
	TransitionSlideLeft
	// TransitionSlideRight 从左向右滑入（后退）。
	TransitionSlideRight
)

// transitionState 保存过渡动画状态。
type transitionState struct {
	active     bool
	from       string // 前一个路径
	to         string // 目标路径
	transition Transition
	startTime  time.Time
	duration   time.Duration
	progress   float32 // 0→1
}

// reverseTransition 返回过渡的反向版本（用于返回操作）。
func reverseTransition(t Transition) Transition {
	switch t {
	case TransitionSlideLeft:
		return TransitionSlideRight
	case TransitionSlideRight:
		return TransitionSlideLeft
	default:
		return t
	}
}

// defaultTransitionDuration 默认过渡动画时长。
const defaultTransitionDuration = 300 * time.Millisecond
