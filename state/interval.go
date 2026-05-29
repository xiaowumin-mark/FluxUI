package state

import (
	"sync"
	"time"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
)

// UseInterval 在组件挂载期间按固定间隔执行 fn，并在卸载时停止。
// fn 不接收 frame Context，避免 goroutine 持有过期的渲染上下文。
func UseInterval(ctx *internal.Context, interval time.Duration, fn func()) {
	if ctx == nil || interval <= 0 || fn == nil {
		return
	}
	rt := ctx.Runtime()
	UseMount(ctx, func() func() {
		stop := make(chan struct{})
		done := make(chan struct{})
		var once sync.Once
		go func() {
			defer close(done)
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					select {
					case <-stop:
						return
					default:
					}
					fn()
					if rt != nil {
						rt.RequestRedraw()
					}
				case <-stop:
					return
				}
			}
		}()
		return func() {
			once.Do(func() {
				close(stop)
				<-done
			})
		}
	})
}
