# Style System Unification Plan

当前 `style.Style` 仅被 `Container` 消费。其他 9 个 widget 各自重复实现 `bg/padding/radius` 存储字段和 option 函数（共 17 个重复函数）。本计划引入统一的 `Decoration` 类型，实现一次性定义、跨组件复用。

## Status: 已完成

所有 Phases 1-4 已实现。`Style` / `Container` / `ContainerElement` 均标记为 Deprecated。

## 新增 API

```go
// style/decoration.go
type Decoration struct {
    Background *color.NRGBA  // nil → 组件默认（来自 Theme）
    Padding    *Insets       // nil → 组件默认
    Margin     *Insets       // nil → 无边距（NEW）
    Radius     *float32      // nil → 组件默认
    Border     *Border       // nil → 无边框（NEW）
}

func (d Decoration) WithBg(color.NRGBA) Decoration
func (d Decoration) WithPad(Insets)   Decoration
func (d Decoration) WithMargin(Insets) Decoration   // NEW
func (d Decoration) WithRad(float32)  Decoration
func (d Decoration) WithBorder(Border) Decoration    // NEW
func (d Decoration) Merge(Decoration) Decoration
func (d Decoration) ResolveBg(color.NRGBA) color.NRGBA
func (d Decoration) ResolvePad(Insets)   Insets
func (d Decoration) ResolveMargin(Insets) Insets     // NEW
func (d Decoration) ResolveRad(float32)  float32
func (d Decoration) ResolveBorder(Border) Border     // NEW

// style/border.go（NEW）
type Border struct {
    Width float32
    Color color.NRGBA
}
func (b Border) IsZero() bool

// style/insets.go
func Only(top, right, bottom, left float32) Insets   // NEW
func Horizontal(v float32) Insets                     // NEW
func Vertical(v float32) Insets                       // NEW

// ui/ 层全局便捷函数
func Bg(color.NRGBA)   style.Decoration
func Pad(style.Insets)  style.Decoration
func Margin(style.Insets) style.Decoration            // NEW
func Rad(float32)       style.Decoration
func BorderDeco(width float32, color.NRGBA) style.Decoration  // NEW
func Only(top,right,bottom,left float32) Insets       // NEW
func LeftRight(v float32) Insets                      // NEW（别名 style.Horizontal，避免与 Axis 冲突）
func TopBottom(v float32) Insets                      // NEW（别名 style.Vertical，避免与 Axis 冲突）

// Deprecated
type Style       // 不再推荐，使用 Decoration
func Container(st Style, child Widget)  // 使用 ContainerDecoration
func ContainerElement(st Style, child Element)  // 使用 ContainerDecorationElement
```

## 迁移策略：分阶段

### Phase 1：Foundation ✅ 完成

- 新建 `style/decoration.go`
- 新建 `style/border.go`（Border 类型）
- 扩展 `style/insets.go`（Only / Horizontal / Vertical）
- 扩展 `style/decoration.go`（Margin / Border 字段 + 方法）
- 标记 `style.Style` 为 Deprecated
- 在 `ui/ui.go` 添加 `Bg` / `Pad` / `Rad` / `Margin` / `BorderDeco` / `Only` / `LeftRight` / `TopBottom`

### Phase 2：Widget 迁移 ✅ 完成

| Widget | 新 option | 旧 option 保留 |
|--------|----------|---------------|
| Button | `ButtonDecoration(d)` | `ButtonBackground` / `ButtonPadding` / `ButtonRadius` |
| TextField | `InputDecoration(d)` | `InputBackground` / `InputPadding` / `InputRadius` |
| Card | `CardDecoration(d)` | `CardBackground` / `CardPadding` / `CardRadius` / `CardBorder` |
| Image | `ImageDecoration(d)` | `ImageBackground` / `ImageRadius` |
| Popup | `PopupDecoration(d)` | `PopupBackground` / `PopupPadding` / `PopupRadius` |
| AppBar | `AppBarDecoration(d)` | `AppBarBackground` |
| BottomNav | `BottomNavDecoration(d)` | `BottomNavBackground` |
| ListView | `ListDecoration(d)` | `ListPadding` |
| GridView | `GridDecoration(d)` | `GridPadding` |
| Toast | `ToastDecoration(d)` | — |
| Dialog | `DialogDecoration(d)` | `DialogRadius` |
| Select | `SelectDecoration(d)` | — |
| Checkbox | `CheckboxDecoration(d)` | — |

### Phase 3：渲染打通 ✅ 完成

- `internal.SurfaceSpec` 新增 `BorderColor` / `BorderWidth` 字段
- `LayoutSurface` 渲染边框（Gio clip.Stroke）
- `widget.ContainerDecoration(d Decoration, child Widget)` 实现

### Phase 4：硬编码替换 + 边框迁移 ✅ 完成

- Toast / Dialog / Select / Popup / AppBar / BottomNav 内部全部从 `Container(Style{})` 切换到 `ContainerDecoration`
- Card 的 ad-hoc 嵌套边框技巧迁移到 `Decoration.Border` 单层渲染
- Tabs 的 indicator bar 切换到 `ContainerDecoration`

### Phase 5：API 统一（可选）

- Checkbox 的 `padding` / `background` 字段（之前已移除）
- 旧的 `ButtonBackground` / `ButtonPadding` / `ButtonRadius` 等单属性 setter 可标记 Deprecated
