<!-- fluxui-doc-meta
{
  "id": "material3_motion",
  "title": "Material 3 组件动画规范",
  "category": "使用指南",
  "order": 122,
  "summary": "定义 FluxUI MD3 默认组件动画的时长、缓动曲线和状态映射，用作后续组件接入动画的基础规范。",
  "example": { "id": "material3_showcase" },
  "apis": [
    "style.InteractionHoverDuration",
    "style.InteractionPressedDuration",
    "style.InteractionFocusDuration",
    "style.InteractionSelectedDuration",
    "style.InteractionMenuEnterDuration",
    "style.InteractionToastEnterDuration",
    "style.InteractionLoadingValueDuration",
    "style.InteractionStandardEasing(v float32) float32",
    "style.InteractionEmphasizedDecelerateEasing(v float32) float32"
  ]
}
-->

# Material 3 组件动画规范

本文档定义 FluxUI 默认 MD3 组件动画的第一版 motion token。它是 Button、Tabs、Navigation、Menu、Toast、ProgressIndicators 等默认组件接入动画的基础规范；组件实现应优先复用这里的时长、缓动曲线和状态映射。

当前实现已将默认动画接入主要 MD3 组件路径：Button、TextField、Selection controls、Switch、Slider、Tabs、BottomNavigation、Menu、Select、DropdownMenu、ListItem、IconButton、FAB、NavigationRail、NavigationDrawer、Chip、Tooltip、Dialog、Popup、Toast/Snackbar、ProgressIndicators 和交互 Decoration。Pressable/ClickArea 继续保持无固定视觉样式，不默认附加 state layer 或 ripple。

## 基本原则

- 动画服务于状态感知，不做装饰性长动画。
- 桌面工具界面优先短时长、低位移、低缩放，避免拖慢操作节奏。
- hover、pressed、focus、selected 默认只动画颜色、透明度、边框、阴影和 indicator；不应改变 layout 尺寸。
- menu、toast 这类 overlay 可以使用轻微 opacity + transform，但 transform 位移应控制在 4~8dp，scale 控制在 0.98~1.00。
- loading 可以持续动画，但必须在隐藏、完成或不可见时停止请求 frame。
- disabled 状态不绘制 hover、pressed、focus 动画，也不触发 ripple。

## 缓动曲线

FluxUI 使用 Material 3 motion token 作为默认曲线来源。后续组件实现时应优先引用 `style.Interaction*Easing`，或使用 `ui.CubicBezier(...)` 创建相同曲线。

| 名称 | 曲线 | 默认用途 |
| --- | --- | --- |
| Linear | `linear` | loading 循环、持续旋转、shimmer 位移。 |
| Standard | `cubic-bezier(0.2, 0, 0, 1)` | selected、indicator、一般状态切换。 |
| Standard decelerate | `cubic-bezier(0, 0, 0, 1)` | hover/focus/pressed 进入，轻量状态出现。 |
| Standard accelerate | `cubic-bezier(0.3, 0, 1, 1)` | hover/focus/pressed 退出，轻量状态消失。 |
| Emphasized | `cubic-bezier(0.2, 0, 0, 1)` | 中等强调的尺寸、位置或容器变化。 |
| Emphasized decelerate | `cubic-bezier(0.05, 0.7, 0.1, 1)` | menu、toast、dialog 类 overlay 进入。 |
| Emphasized accelerate | `cubic-bezier(0.3, 0, 0.8, 0.15)` | menu、toast、dialog 类 overlay 退出。 |

## 默认状态映射

| 状态 | 默认时长 | 默认曲线 | 动画对象 |
| --- | --- | --- | --- |
| Hover enter | 100ms | Standard decelerate | state layer opacity、背景色、边框色、图标/文字色。 |
| Hover exit | 100ms | Standard accelerate | state layer opacity、背景色、边框色、图标/文字色。 |
| Pressed enter | 50ms | Standard decelerate | pressed state layer、按钮/列表项轻量色彩反馈。 |
| Pressed exit | 100ms | Standard accelerate | pressed state layer fade-out；ripple fade 单独使用 ripple token。 |
| Focus enter | 100ms | Standard decelerate | focus ring opacity、outline color、少量 outline width。 |
| Focus exit | 100ms | Standard accelerate | focus ring fade-out、outline color。 |
| Selected | 150ms | Standard | selected container、label/icon color、check mark opacity。 |
| Selected indicator | 200ms | Standard | Tabs、NavigationBar、NavigationRail、NavigationDrawer indicator 位置和尺寸。 |
| Menu enter | 150ms | Emphasized decelerate | opacity + y offset 4dp -> 0，或 scale 0.98 -> 1.00。 |
| Menu exit | 100ms | Emphasized accelerate | opacity + y offset 0 -> 4dp，或 scale 1.00 -> 0.98。 |
| Toast enter | 200ms | Emphasized decelerate | opacity + y offset 8dp -> 0。 |
| Toast exit | 150ms | Emphasized accelerate | opacity + y offset 0 -> 8dp。 |
| Loading value | 250ms | Standard | determinate progress 数值变化。 |
| Loading linear cycle | 1200ms | Linear | indeterminate linear progress。 |
| Loading circular cycle | 1400ms | Linear | indeterminate circular progress / spinner rotation。 |
| Loading shimmer cycle | 1200ms | Linear | skeleton 或 loading placeholder shimmer。 |

Ripple 保持独立节奏：expand 使用 450ms，fade 使用 550ms。Pressed 的 50/100ms 只控制 state layer 和非 ripple 视觉反馈，不覆盖 ripple primitive 的扩散与淡出。

## 组件接入规则

Button、IconButton、Chip、ListItem、MenuItem、Navigation item 等高频交互组件应先接入 hover、pressed、focus、selected 的颜色/透明度过渡，再考虑阴影或 transform。按钮按压不默认缩放，避免密集工具栏在连续点击时显得跳动。

Tabs、BottomNavigation、NavigationRail、NavigationDrawer 的 selected indicator 应使用 200ms Standard。颜色切换可以使用 150ms Standard，indicator 位移和尺寸变化允许稍慢，但不能引起父布局重新测量。

Menu、Select popup、DropdownMenu 和 Tooltip 默认使用 opacity + 轻微 y offset。打开方向与实际弹出方向相反时，y offset 的方向应跟随 anchor，避免菜单看起来从错误方向滑入。

Toast 和 Snackbar 的自动关闭时长不属于 motion token；`ToastDuration(...)` 只控制停留时间。出入场动画应在停留计时之外单独处理，action hover/pressed 继续复用 Button/Pressable token。

ProgressIndicators 的 determinate 值变化使用 250ms Standard。indeterminate loading 使用线性循环；循环动画必须只在组件可见且状态仍为 loading 时请求下一帧。

## 验收要求

后续组件接入动画时，至少验证：

- 动画不改变 layout 尺寸，不造成文本、图标或控件抖动。
- Light/Dark 下 state layer、focus ring、selected indicator 都有足够可见度。
- disabled 状态不绘制 hover/pressed/ripple。
- Menu、Toast 退出后不再命中点击区域。
- Loading 隐藏或完成后停止持续 invalidation。
- showcase 交互状态截图覆盖 hover、pressed/ripple、focus、selected、expanded、toast enter/exit。

参考来源：

- Material 3 motion easing and duration: `https://m3.material.io/styles/motion/easing-and-duration/overview`
- Material 3 motion token specs: `https://m3.material.io/styles/motion/easing-and-duration/tokens-specs`
- Material Components motion theming: `https://github.com/material-components/material-components-android/blob/master/docs/theming/Motion.md`
