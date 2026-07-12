# 高级组件 Docs Browser 可构建示例模板

本模板位于 `docs/templates/`，不会被 Docs Browser 扫描。复制到 `docs/widgets/` 或适当的已扫描目录后，必须替换所有占位符并把元数据置于文件开头。

## 交付清单

- [ ] 文档开头有唯一的 `fluxui-doc-meta`，`example.id` 已在 `examples/docs_browser` 注册。
- [ ] Widget 与 Element 入口、默认值、受控/非受控语义、disabled/loading、callback、Ref（若有）均真实描述。
- [ ] 键盘表、Overlay/滚动行为、平台 capability 和已知限制已写出。
- [ ] 示例以当前公开 API 编译；不得把预留 option、mock capability 或计划功能写成已完成特性。
- [ ] 执行 `go test ./examples/docs_browser`，并在需要时运行 visual smoke。

## 可复制的文档骨架

````md
<!-- fluxui-doc-meta
{
  "id": "<component_id>",
  "title": "<Component> <中文名称>",
  "category": "<分类>",
  "order": <排序号>,
  "summary": "<一句真实能力描述，不宣传预留功能。>",
  "example": { "id": "<docs_browser_example_id>" },
  "apis": [
    "<Component>(...)",
    "<主要 option / ref API>"
  ]
}
-->

# <Component> <中文名称>

## 何时使用

<使用场景、非目标，以及与相邻组件的区别。>

## 最小示例

```go
// 使用已注册的 docs browser example id 所对应的真实公开 API。
```

## 状态与回调合同

| 输入/动作 | 默认值或优先级 | 回调/副作用 |
| --- | --- | --- |
| 受控 value | <外部 value 优先级> | <何时 OnChange，是否只在真实变化时> |
| disabled/loading | <入口阻断行为> | <Ref 是否受限> |
| Ref | <挂载、队列、卸载行为> | <是否触发用户回调> |

## 键盘与焦点

| 按键 | 行为 |
| --- | --- |
| Tab / Shift+Tab | <行为> |
| Enter / Space | <行为> |
| Escape / Arrow / Home / End | <行为或“不支持”> |

## Overlay、滚动与平台

<placement、outside click、focus restore、可见 viewport、native capability 或“不适用”。>

## 可访问性与视觉状态

<已提供的 Role/Name/Value/State；不要把基础 semantic 输出描述为完整读屏支持。>

## 限制与迁移

<未实现能力、弃用项、替代组件、回滚/兼容说明。>
````

## 示例注册记录

| 项目 | 填写值 |
| --- | --- |
| 文档文件 | `docs/widgets/<component>.md` |
| metadata id | `<component_id>` |
| docs browser example id | `<docs_browser_example_id>` |
| 注册源文件/函数 | `<examples/docs_browser/...>` |
| build/test 命令与结果 | `go test ./examples/docs_browser`：`<结果>` |
| visual artifact/手工 smoke | `<链接或步骤>` |
