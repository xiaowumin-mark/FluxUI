package anim

// Easing 定义插值曲线。
type Easing func(float32) float32

// Linear 线性缓动。
func Linear(v float32) float32 {
	return clamp01(v)
}

// EaseOut 二次缓出。
func EaseOut(v float32) float32 {
	v = clamp01(v)
	return 1 - (1-v)*(1-v)
}

// EaseInOut 平滑缓入缓出。
func EaseInOut(v float32) float32 {
	v = clamp01(v)
	if v < 0.5 {
		return 2 * v * v
	}
	return 1 - ((-2*v + 2) * (-2*v + 2) / 2)
}
