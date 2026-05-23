# Changelog

## v0.1.0 (2026-05-23)

### 新增

- React-style 声明式组件：`RunElement`、`Element`、函数组件、`Fragment`、`Key`、`Provider`
- 新增 Hooks：`UseMemo[T]`、`UseRef[T]`、`UseCallback[T]`
- 新增 Context API：`Provider[T]`、`UseContext[T]`
- 新增路由 Hooks：`UseParams`、`UseLocation`、`UseNavigate`
- 新增 Popup 组件（`PopupRef` 独立于 `DialogRef`）
- 新增 `OnHover` 事件支持

### 变更

- `RouteParams` 废弃，统一使用 `UseParams`
- `StateWithInitial` 移除，使用 `UseState(ctx, initial)`
- 字体便捷 API 精简，仅保留 `TextFont(spec)`、`TextFontWeight(weight)`
- `PopupAttachRef` 参数改为 `*PopupRef`（不再使用 `*DialogRef`）
- 移除未使用的 `event/keyboard`、`layout/constraints`、`SelectValue`、`InputValue` 等 API

### 示例

- 新增：`react_workspace`、`router`、`popup_demo`、`team_workspace`、`virtual_scroll`、`fonts`、`horizontal_scroll`
- 移除冗余示例：`animation`、`react_counter`、`router_element`
