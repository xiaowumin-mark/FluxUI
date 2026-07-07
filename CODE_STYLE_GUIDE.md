# FluxUI 代码编写风格指南

本文约束 FluxUI 仓库中的 Go 代码、文档和测试风格。优先遵循现有代码风格；本文用于补齐跨组件开发时必须保持一致的规则。

## Go 基础风格

- 所有 Go 代码必须通过 `gofmt`。
- 包内命名保持简短但明确，对外 API 使用完整语义。
- 不为局部简单逻辑添加抽象；只有当抽象能消除真实重复或统一行为边界时才添加 helper。
- 不在函数中混合 layout、event、style、state 四类职责；复杂组件应拆成窄 helper。
- 错误返回和回调触发路径必须显式，不依赖隐式零值副作用。

## 文件组织

- 公共 API 靠近组件入口，内部 helper 靠近使用处。
- 组件 option、config、widget struct、Layout/helper、Ref 按稳定顺序组织。
- 测试文件与被测能力同包放置；跨包行为使用公开 API 测试。
- 新增文档使用 UTF-8 编码。

## 命名约定

- 对外组件：`Button`、`ScrollView`、`Dialog`。
- 对外 option：`ButtonDisabled`、`ScrollOnChange`、`DialogMaskClosable`。
- 内部配置：`buttonConfig`、`scrollViewConfig`。
- 内部状态：`buttonState`、`scrollState`。
- 内部 helper：使用动词或职责名，例如 `dispatchClickDefault`、`resolveDecorationState`。
- Ref：`XxxRef`、`NewXxxRef`、`XxxAttachRef`。

## Option 和 Callback 风格

- option 函数只写入 config，不执行 layout、event drain、redraw 或 state mutation。
- callback 参数需要携带 `ctx` 时，统一使用 `func(ctx *Context, ...)` 或现有组件约定。
- `OnChange` 类 callback 默认只在真实值变化时触发；例外必须写入文档。
- 旧 callback 与新 event 并存时，新 event 先分发，旧 callback 作为 default action 或兼容回调。

## Helper 抽象规则

可以抽 helper 的情况：

- 多个组件共享同一 default action gate。
- 多个组件共享同一 focus target 注册规则。
- 多个组件共享同一 decoration 合并或 interaction snapshot。
- 多个组件共享同一 scroll/overlay region 计算。

不应抽 helper 的情况：

- 只是两处代码长得像，但语义不同。
- helper 需要大量 bool 参数才能覆盖差异。
- 会把 Button surface、Tabs indicator、Select field、Switch thumb/track 等真实视觉差异压平。
- 会改变旧 API 或旧 callback 顺序。

## Event 代码规则

- event 类型、bubbles、cancelable、default action gate 必须成对设计。
- 新增可取消事件时，必须有调用方读取 allowed。
- listener 顺序、capture/target/bubble、once/passive/priority 语义不能在 widget 层私自重写。
- portal/boundary 只改变 event path，不应改变 layout path 或 state path，除非文档明确。

## Layout 代码规则

- 先计算 constraints，再 layout child，再注册命中和绘制。
- 不在 layout 中无条件读取窗口全局尺寸。
- 不用视觉 shadow/ripple 的外扩尺寸反推 layout 尺寸。
- 横向滚动和纵向滚动的 axis 决策必须集中，不在多个组件里重复手写。

## State 和 Ref 代码规则

- 状态必须能解释 path owner。
- Ref 命令必须能解释消费时机和失效条件。
- controlled value、内部 state、Ref 写入、用户交互写入必须有明确优先级。
- 不允许同一帧中多个来源互相覆盖后仍触发多个不一致 callback。

## Diagnostics 和 Perf 规则

- diagnostics 默认关闭，不得让默认路径承担 registry dump 或 event history 成本。
- 高频 pointer/wheel/hover/scroll 路径避免不必要分配。
- redraw reason 应使用稳定字符串或结构化字段，不随意新增同义 reason。
- benchmark 应区分 layout 成本、paint 成本、event registration 成本和 input dispatch 成本。

## 测试规则

- 行为修复优先补最小单元测试。
- 跨组件边界修复必须补集成测试或手动 smoke 步骤。
- 旧 API 兼容修复必须包含旧 callback 路径。
- default action 修复必须包含 `PreventDefault` 生效和不生效两个方向。
- 滚动修复必须覆盖 axis、父子传递、offset 更新和命中刷新。

## 文档规则

- docs browser 可加载的文档必须包含 `fluxui-doc-meta`。
- 根目录规范文档不需要 `fluxui-doc-meta`，用于开发约束。
- 公开 API 文档必须说明默认值、受控/非受控语义、Ref 行为和回调触发条件。
- 示例 README 必须和当前入口一致，不能继续描述已经迁移的 legacy 入口。
