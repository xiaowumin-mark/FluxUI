<!-- fluxui-doc-meta
{
  "id": "slider",
  "title": "Slider 滑块",
  "category": "输入交互",
  "order": 240,
  "summary": "Slider 用于区间连续值输入。",
  "example": { "id": "slider_basic" },
  "apis": [
    "Slider(value float32, opts ...SliderOption) Widget",
    "SliderOnChange(fn func(ctx *Context, value float32)) SliderOption",
    "SliderDisabled(disabled bool) SliderOption",
    "SliderMin(min float32) SliderOption",
    "SliderMax(max float32) SliderOption",
    "SliderStep(step float32) SliderOption",
    "SliderWidth(width float32) SliderOption",
    "SliderTrackColor(color color.NRGBA) SliderOption",
    "SliderThumbColor(color color.NRGBA) SliderOption",
    "SliderProgressColor(color color.NRGBA) SliderOption",
    "NewSliderRef() *SliderRef",
    "SliderAttachRef(ref *SliderRef) SliderOption",
    "(*SliderRef).SetValue(value float32)",
    "(*SliderRef).StepBy(delta float32)"
  ]
}
-->

# Slider 滑块

## 组件说明
Slider 适用于音量、进度、阈值等连续数值场景。

## 使用方法
- 设定取值范围：`SliderMin` + `SliderMax`。
- 离散步进：`SliderStep`。
- 与进度条联动时建议统一状态源。
- 外部程序调整值可使用 `SliderAttachRef`，通过 `SetValue/StepBy` 下发命令。

## Host-state / React-style 说明

- 当前仍以 legacy `Widget` + `SliderRef` 作为兼容实现。
- `Slider` 的 Element wrapper 名称与 host-state 边界尚未冻结，不在本批次文档中作为稳定公开 API 推荐。
- 在 Batch 4 期间，文档保持 legacy-first，只保留受控 value、range、step 和 ref 操作说明，不引入新的 React-style snippet。
- 与其他复杂输入一起迁移时，先保证命令式 ref 行为和旧示例兼容。

## 使用示例
```go
value := ui.State[float32](ctx)
ui.Slider(
    value.Value(),
    ui.SliderMin(0),
    ui.SliderMax(100),
    ui.SliderOnChange(func(ctx *ui.Context, v float32) {
        value.Set(v)
    }),
)
```
