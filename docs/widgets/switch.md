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
    "SwitchElement(checked bool, opts ...SwitchOption) Element",
    "SwitchOnChange(fn func(ctx *Context, checked bool)) SwitchOption",
    "SwitchDisabled(disabled bool) SwitchOption",
    "SwitchWidth(width float32) SwitchOption",
    "SwitchHeight(height float32) SwitchOption",
    "SwitchColor(color color.NRGBA) SwitchOption",
    "SwitchTrackColor(color color.NRGBA) SwitchOption",
    "SwitchThumbColor(color color.NRGBA) SwitchOption",
    "SwitchDecoration(d Decoration) SwitchOption",
    "SwitchIcons(checkedIcon, uncheckedIcon string) SwitchOption",
    "SwitchCheckedIcon(name string) SwitchOption",
    "SwitchUncheckedIcon(name string) SwitchOption",
    "SwitchIconUseFont(id string) SwitchOption",
    "SwitchIconFontFamily(family string) SwitchOption",
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

## MD3 默认样式

- checked track 使用 `Primary`，thumb 使用 `OnPrimary`。
- unchecked track 使用 `SurfaceVariant`，thumb 使用 `Outline`。
- unchecked track 使用 2dp `Outline`，thumb 默认 16dp；checked thumb 为 24dp，pressed thumb 为 28dp。
- hover/pressed 使用以 thumb 为中心的 40dp state layer。
- checked/unchecked 切换使用带轻微回弹的位移动画，pressed/hover 之间 thumb size 会平滑变化。
- disabled 使用 `OnSurface` 12% / 38%。

## 使用方法
- 通过 `checked` 传值。
- 变化回调用 `SwitchOnChange`。
- 可单独配置轨道颜色与拇指颜色。
- 可通过 `SwitchIcons` / `SwitchCheckedIcon` / `SwitchUncheckedIcon` 给拇指增加 Material Symbols 图标。
- 需要外部主动切换时，使用 `SwitchAttachRef` 与 `SwitchRef` 方法。

## 使用示例

### React-style Element
新代码可在 `RunElement` root 下返回 `SwitchElement`：

```go
func FeatureToggle(ctx *ui.Context) ui.Element {
    enabled := ui.UseState(ctx, true)
    return ui.RowElement(
        ui.SwitchElement(
            enabled.Value(),
            ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
                enabled.Set(checked)
            }),
        ),
        ui.SpacerElement(8, 0),
        ui.TextElement(fmt.Sprintf("功能开关: %v", enabled.Value())),
    )
}
```

### Legacy Widget
旧 `ui.Switch` / `Widget` 写法继续可用：

```go
enabled := ui.State[bool](ctx)
ui.Switch(
    enabled.Value(),
    ui.SwitchIcons("check", "close"),
    ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
        enabled.Set(checked)
    }),
)
```
