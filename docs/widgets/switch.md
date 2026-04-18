<!-- fluxui-doc-meta
{
  "id": "switch",
  "title": "Switch 开关",
  "category": "输入交互",
  "order": 230,
  "summary": "Switch 用于即时开关型布尔状态。",
  "example": { "id": "switch_basic" },
  "apis": [
    "Switch(checked bool, opts ...SwitchOption) Widget",
    "SwitchOnChange(fn func(ctx *Context, checked bool)) SwitchOption",
    "SwitchDisabled(disabled bool) SwitchOption",
    "SwitchWidth(width float32) SwitchOption",
    "SwitchHeight(height float32) SwitchOption",
    "SwitchColor(color color.NRGBA) SwitchOption",
    "SwitchTrackColor(color color.NRGBA) SwitchOption",
    "SwitchThumbColor(color color.NRGBA) SwitchOption",
    "NewSwitchRef() *SwitchRef",
    "SwitchAttachRef(ref *SwitchRef) SwitchOption",
    "(*SwitchRef).SetChecked(checked bool)",
    "(*SwitchRef).Toggle()"
  ]
}
-->

# Switch 开关

## 组件说明
Switch 常用于“立即生效”的开关项，比如通知开关、实验特性开关。

## 使用方法
- 通过 `checked` 传值。
- 变化回调用 `SwitchOnChange`。
- 可单独配置轨道颜色与拇指颜色。
- 需要外部主动切换时，使用 `SwitchAttachRef` 与 `SwitchRef` 方法。

## 使用示例
```go
enabled := ui.State[bool](ctx)
ui.Switch(
    enabled.Value(),
    ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
        enabled.Set(checked)
    }),
)
```
