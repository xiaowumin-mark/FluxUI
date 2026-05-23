# Style System Unification Plan

当前 `style.Style` 仅被 `Container` 消费。其他 9 个 widget 各自重复实现 `bg/padding/radius` 存储字段和 option 函数（共 17 个重复函数）。本计划引入统一的 `Decoration` 类型，实现一次性定义、跨组件复用。

## 目标

- 用 `Decoration` 统一替换 9 个 widget 的重复装饰代码
- 保持 100% 向后兼容（旧 option 函数不删除）
- 新增 `Bg` / `Pad` / `Rad` 全局便捷构造函数

## 新增 API

```go
// style/decoration.go
type Decoration struct {
    Background *color.NRGBA  // nil → 组件默认（来自 Theme）
    Padding    *Insets       // nil → 组件默认
    Radius     *float32      // nil → 组件默认
}

func (d Decoration) WithBg(color.NRGBA) Decoration
func (d Decoration) WithPad(Insets)   Decoration
func (d Decoration) WithRad(float32)  Decoration
func (d Decoration) Merge(Decoration) Decoration
func (d Decoration) ResolveBg(color.NRGBA) color.NRGBA
func (d Decoration) ResolvePad(Insets)   Insets
func (d Decoration) ResolveRad(float32)  float32

// ui/ 层全局便捷函数
func Bg(color.NRGBA)  style.Decoration
func Pad(style.Insets) style.Decoration
func Rad(float32)      style.Decoration
```

## 迁移策略：分阶段

### Phase 1：Foundation

- 新建 `style/decoration.go`
- 在 `ui/ui.go` 添加 `Bg` / `Pad` / `Rad` 构造函数
- 零 widget 改动，纯新增

### Phase 2：Widget 批量迁移

每个 widget 加 `XxxDecoration(d Decoration) XxxOption`，渲染时优先读取 `decoration`，旧 option 函数不变。

| Widget | 新增 option | 旧 option 保留 |
|--------|------------|---------------|
| Button | `ButtonDecoration(d)` | `ButtonBackground` / `ButtonPadding` / `ButtonRadius` |
| TextField | `InputDecoration(d)` | `InputBackground` / `InputPadding` / `InputRadius` |
| Card | `CardDecoration(d)` | `CardBackground` / `CardPadding` / `CardRadius` |
| Image | `ImageDecoration(d)` | `ImageBackground` / `ImageRadius` |
| Popup | `PopupDecoration(d)` | `PopupBackground` / `PopupPadding` / `PopupRadius` |
| AppBar | `AppBarDecoration(d)` | `AppBarBackground` |
| BottomNav | `BottomNavDecoration(d)` | `BottomNavBackground` |
| ListView | `ListDecoration(d)` | `ListPadding` |
| GridView | `GridDecoration(d)` | `GridPadding` |

### Phase 3：硬编码替换

Toast / Dialog / Select 当前在 Layout 中对 bg/pad/rad 使用硬编码字面量。改为使用 `Decoration` 解析，允许调用方通过 `Decoration` option 覆盖。

### Phase 4：示例与文档

- 更新 `examples/theme_custom` 等使用新 API
- 更新 `docs/widgets/` 中对应的 API 列表

### Phase 5：Cleanup（可选）

- Checkbox 的 `padding` / `background` 字段存在但无 public setter（死代码），移除
