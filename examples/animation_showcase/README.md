# Animation Showcase

`examples/animation_showcase` 使用 **React Element API** (`RunElement` + `Component`) 全面展示 FluxUI Phase 10 动画系统。

## 内容

| Tab | 功能 | 关键 API |
|-----|------|----------|
| 缓动对比 | 6 条缓动曲线同步播放 | `UseAnimatedValue` + 31 种 `anim.Easing` |
| 值动画 | 控制方块宽高从 100→300 动画 | `UseAnimatedValue[float32]` |
| 装饰动画 | 点击同时动画背景色/圆角/缩放/阴影 | `UseAnimatedDecoration` |
| 悬停动画 | 鼠标悬浮时用 CubicBezier 平滑过渡 | `UseAnimatedDecoration` + `anim.CubicBezier(...)` |
| 脉冲动画 | 自动循环 0→1→0 的大小/透明度联动 | `UseAnimatedValue` + 状态切换 |

## 组件架构

每个 tab 通过 `ui.ComponentElement(sectionFunc)` 包装为独立组件，确保 hook 顺序不受 tab 切换影响：

```go
var content ui.Element
switch tab.Value() {
case "animValue":
    content = ui.ComponentElement(animValueSection)
case "animDeco":
    content = ui.ComponentElement(animDecoSection)
// ...
}
```

## 可用的缓动曲线

`ui` 包内置 8 个常用别名（`Linear`, `EaseOut`, `EaseInOut`, `EaseIn`, `EaseInBack`, `EaseOutBack`, `EaseInOutBack`, `EaseOutBounce`），完整 31 种 Penner 曲线 + 自定义贝塞尔通过 `import anim` 引用：

```go
import anim "github.com/xiaowumin-mark/FluxUI/anim"

// 任意 Penner 曲线
anim.EaseInOutCubic
anim.EaseOutElastic
anim.EaseInOutBounce

// 自定义贝塞尔
anim.CubicBezier(0.25, 0.1, 0.25, 1.0)  // CSS ease
```
