<!-- fluxui-doc-meta
{
  "id": "material3",
  "title": "Material Design 3 默认样式",
  "category": "使用指南",
  "order": 120,
  "summary": "说明 FluxUI 默认样式如何按 Material Design 3 token 和组件变体落地。",
  "example": { "id": "material3_showcase" },
  "apis": [
    "LightColors() ColorScheme",
    "DarkColors() ColorScheme",
    "FilledButton(child Widget, opts ...ButtonOption) Widget",
    "OutlinedTextField(value string, opts ...InputOption) Widget",
    "ElevatedCard(child Widget, opts ...CardOption) Widget"
  ]
}
-->

# Material Design 3 默认样式

FluxUI 默认主题现在以 Material Design 3 为目标：颜色、圆角、字体层级、状态层和 elevation 都从主题 token 派生。

## 推荐入口

- 完整实施计划：`docs/material3-design-plan.md`
- 长期推进路线：`docs/material3-roadmap.md`
- 人工视觉回归：`examples/material3_showcase/main.go`
- 主题 token：`docs/theme/color_scheme.md` 和 `docs/theme/theme.md`

## 组件变体

- Button: `FilledButton`、`FilledTonalButton`、`OutlinedButton`、`TextButton`、`ElevatedButton`。
- TextField: `OutlinedTextField`、`FilledTextField`。
- Card: `FilledCard`、`ElevatedCard`、`OutlinedCard`。

旧的 `Button`、`TextField`、`Card` 继续可用，并映射到稳定的 MD3 默认变体。

## 验证

默认样式调整后应运行：

```sh
go test ./...
go vet ./...
go test ./examples/material3_showcase
```
