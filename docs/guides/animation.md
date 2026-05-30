<!-- fluxui-doc-meta
{
  "id": "animation",
  "title": "Animation 动画",
  "category": "使用指南",
  "order": 720,
  "summary": "FluxUI 支持数值动画、Decoration 动画、Easing 曲线和旧 Widget 动画构造器。",
  "example": { "id": "animation_basic" },
  "apis": [
    "type Easing func(float32) float32",
    "UseAnimatedValue[T float32|float64|int](ctx *Context, target T, duration time.Duration, easing anim.Easing) T",
    "UseAnimatedDecoration(ctx *Context, target Decoration, duration time.Duration, easing anim.Easing) Decoration",
    "Linear, EaseOut, EaseInOut, EaseIn, EaseInBack, EaseOutBack, EaseInOutBack, EaseOutBounce",
    "CubicBezier(x1, y1, x2, y2 float32) Easing",
    "Animate(opts ...anim.Option) *anim.Animation",
    "Duration(duration time.Duration) anim.Option",
    "From(value float32) anim.Option",
    "To(value float32) anim.Option",
    "Ease(easing anim.Easing) anim.Option"
  ]
}
-->

# Animation 动画

## API 说明
新代码推荐使用 Hook 形式的动画：

- `UseAnimatedValue`：对数值逐帧插值。
- `UseAnimatedDecoration`：对 Decoration 中可插值字段逐帧插值。
- `CubicBezier`：创建自定义缓动曲线。

旧 API `Animate` / `Duration` / `From` / `To` / `Ease` 仍保留，但新代码应优先使用 Hook。

## 数值动画
```go
func ProgressDemo(ctx *ui.Context) ui.Element {
    target := ui.UseState(ctx, float32(20))
    value := ui.UseAnimatedValue(ctx, target.Value(), 250*time.Millisecond, ui.EaseOut)

    return ui.ColumnElement(
        ui.ProgressBarElement(value),
        ui.ButtonElement(ui.TextElement("切换进度"), ui.OnClick(func(ctx *ui.Context) {
            if target.Value() < 80 {
                target.Set(90)
            } else {
                target.Set(20)
            }
        })),
    )
}
```

## Decoration 动画
`UseAnimatedDecoration` 会平滑插值：背景色、透明度、圆角、Transform、Border、Shadow、Gradient。Padding、Margin、图片和状态装饰会瞬切。

```go
func CardDemo(ctx *ui.Context) ui.Element {
    selected := ui.UseState(ctx, false)
    th := ui.UseTheme(ctx)

    target := ui.Bg(th.Surface).WithPad(ui.All(16)).WithRad(12)
    if selected.Value() {
        target = target.Merge(ui.Elevation(3)).
            WithTransform(ui.Transform2D{ScaleX: 1.04, ScaleY: 1.04})
    }

    deco := ui.UseAnimatedDecoration(ctx, target, 180*time.Millisecond, ui.EaseOutBack)
    return ui.ClickAreaElement(
        ui.ContainerDecorationElement(deco, ui.TextElement("点击切换动画")),
        func(ctx *ui.Context) { selected.Set(!selected.Value()) },
    )
}
```

## Easing 曲线
常用选择：

- `Linear`：匀速。
- `EaseOut`：进入快、结束慢，适合大多数 UI 反馈。
- `EaseInOut`：适合来回切换。
- `EaseOutBack`：带轻微回弹。
- `EaseOutBounce`：弹跳收尾，适合少量强调动效。
- `CubicBezier(...)`：自定义 CSS 风格曲线。

文档浏览器中的示例会同时展示多条曲线：点击“触发动画”后，多个小方块会使用不同 easing 同时位移；悬浮任一方块时，会触发基于 `Decoration.Transform` 的独立 hover 缩放动画。

自定义 CSS 风格曲线：

```go
curve := ui.CubicBezier(0.2, 0.8, 0.2, 1)
v := ui.UseAnimatedValue(ctx, target, 300*time.Millisecond, curve)
```

## MD3 组件动画

如果是在为默认 MD3 组件实现 hover、pressed、focus、selected、menu、toast 或 loading 状态动画，应优先遵循 `docs/guides/material3-motion.md`。该规范定义了组件默认时长、Material 3 缓动曲线和 loading 循环节奏，避免每个组件单独发明动画参数。

## 使用建议
- 交互反馈通常 120~250ms 即可。
- 页面级过渡可使用 250~450ms。
- 不要让大量列表项同时运行复杂阴影/变换动画。
