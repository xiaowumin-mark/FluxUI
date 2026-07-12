# 高级组件视觉状态矩阵模板

本表用于 Docs Browser demo、Component Lab 或专项 example 的视觉验收。截图/人工记录都必须说明主题、窗口尺寸、缩放、平台和构建 commit；不可自动化的格子不能留空。

## 组件与环境

| 项目 | 值 |
| --- | --- |
| 组件/API | `<Component>` |
| demo/example | `<路径与 example id>` |
| commit / Go / Gio | `<...>` |
| OS / display server / DPI | `<...>` |
| viewport | `<窄/宽具体尺寸>` |
| 截图 artifact 或人工记录 | `<链接>` |

## 必测状态

| 维度 | 必测值 | 结果/链接 | 备注或豁免 |
| --- | --- | --- | --- |
| 主题 | Light、Dark |  |  |
| 宽度 | 窄、宽 |  | 长标签和横向溢出也要覆盖。 |
| 缩放 | 100%、125%、150%、200% |  | 记录实际 scale。 |
| 交互 | idle、hover、pressed、focus-visible |  | 仅在组件支持时；不支持要写明。 |
| 可用性 | enabled、disabled、read-only（适用时） |  |  |
| 业务状态 | empty、loading、error、长文本 |  |  |
| 集合/选择 | selected、active、重排/过滤后（适用时） |  |  |
| Overlay | 打开、边缘 placement、outside click 后（适用时） |  |  |
| 滚动 | 初始、滚动后不移动鼠标（适用时） |  | 命中/hover 必须刷新。 |

## 视觉通过条件

- 文本、icon、focus ring、error/supporting text 不被裁切或与背景混淆。
- 状态层不改变承诺的 layout/hit area；长文本和窄宽度有可解释的 overflow 策略。
- Overlay 位于可见 viewport 内，必要时翻转、平移或内部滚动；不遮蔽错误的普通文档流。
- 截图差异更新须经明确批准，视觉 smoke artifact 可追溯到 commit。

## 原生平台补充

如组件涉及文件、拖放、窗口、IME、剪贴板、通知或托盘，另列出 mock 覆盖与 Windows/macOS/Linux-X11/Linux-Wayland 的实际 smoke；headless 截图不替代此项。
