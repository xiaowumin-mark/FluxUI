# FluxUI 功能开发关联性检查表

本文用于完成某个功能或组件开发后，检查它与其他组件和底层系统的关联性。任何涉及 runtime、event、layout、widget、style、examples 的 PR，都应按本文逐项确认。

## 使用方式

在 PR 描述或修复记录中填写：

- 修改范围：
- 关联审查任务：
- 影响组件：
- 已执行检查项：
- 未覆盖风险：
- 回滚方式：

## 1. 包边界

- 是否新增了跨层依赖？
- 是否让 `internal` 依赖了 `widget`、`ui` 或示例代码？
- 是否让 `event` 写入了具体组件语义？
- 是否让 `style/theme` 注册事件或改变状态？
- 是否需要在 `ui` 层暴露 API，而不是只改 `widget`？

## 2. Runtime 和 Frame 生命周期

- 是否新增跨帧状态？它绑定哪个 PathID？
- 每帧必须重建的 registry 是否仍会清理？
- 是否影响 redraw reason、invalidator 或 frame end 清理？
- portal/overlay 中的 owner path、event path、state path 是否明确？
- 条件渲染或列表重排后，状态是否可能错配？

## 3. Layout 和 Constraints

- 是否尊重父级 constraints？
- 是否在有限/无限约束下都能解释尺寸？
- 是否改变了 layout size、visual size 或 hit size？
- 是否可能把横向内容撑成无限宽？
- overlay 是否误占普通文档流？
- 滚动容器是否正确区分 content size 和 viewport size？

## 4. Event 和 Default Action

- 是否新增事件？它的 `bubbles`、`cancelable`、default action 是什么？
- 是否读取 dispatch allowed，确保 `PreventDefault` 生效？
- passive listener 中 `PreventDefault` 是否仍不生效？
- `StopPropagation` 是否只影响传播，不影响必要清理？
- pointer、keyboard、wheel、drag/drop 是否有一致的 target path？
- synthetic event 是否不会重复触发真实 default action？

## 5. 旧 API 兼容

- 是否改变了 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange` 的签名？
- 是否改变旧 callback 与新 event 的顺序？
- 是否会让旧 callback 执行两次或漏执行？
- 旧 API 的 disabled/loading/controlled 行为是否保持？
- 文档是否标明新旧 API 的边界？

## 6. Focus、Keyboard 和 Shortcut

- 新组件是否需要 focus target？
- disabled/hidden/tabStop 是否影响 focus 注册？
- Enter/Space/Escape/Arrow/Tab 默认行为是否明确？
- 局部 shortcut 是否只在 scope 内触发？
- overlay 打开后 focus 是否进入、保持、恢复或清空？
- modal overlay 是否阻止 Tab 越界？

## 7. Scroll、Wheel 和 Hit Refresh

- 组件是否注册 wheel listener？
- 子控件是否会无条件截停父级滚动？
- 横向滚动是否只由明确横向策略触发？
- 是否处理 touchpad 双轴 delta 或明确暂不支持？
- scroll 后不移动鼠标，hover/click 是否命中新视觉目标？
- `ScrollOnChange` 是否只在真实 offset 变化时触发？

## 8. Overlay、Portal 和 Outside Click

- overlay 是 modal、portal-only、local deferred、hover 还是 timed？
- outside click 的 protected rect 是否覆盖 trigger 和 popup？
- 内部点击、内部滚动、打开点击、遮罩点击是否分清？
- Escape 关闭是否只影响 topmost overlay？
- Popup 是否仍按 owner path 冒泡？
- Dialog modal 是否按 boundary 截断？

## 9. Controlled Value、Ref 和 OnChange

- 外部 value、内部 state、Ref 命令、用户交互谁优先？
- 同一帧多个写入是否会互相覆盖？
- `OnChange` 是否只在真实变化时触发？
- Ref 命令在 disabled/loading/未挂载时如何处理？
- 程序设置是否触发用户语义 callback？文档是否说明？

## 10. Text、IME 和 Submit

- 用户输入、粘贴、删除、撤销/重做、程序设置是否可区分？
- `beforeinput` 取消后是否不会触发 `input/change/OnChange`？
- IME composition 中 Enter 是否不会提前 submit？
- submit 是否可取消？取消后阻止什么默认行为？
- SearchBar 是否继承 TextField 的文本事件语义？

## 11. Style、Decoration 和 Animation

- 用户 decoration 是否会丢失 disabled/hover/focus/selected/error 表达？
- hover/pressed/focused 是否只有一个权威来源？
- ripple/state layer 是否不改变 layout 和 hit area？
- cursor 是否只在当前命中区域生效，不污染整窗？
- 动画是否只影响视觉或明确 overlay 位置？
- 动画结束后是否停止 redraw？

## 12. Docs、Examples 和 Tests

- 是否需要更新 docs browser 文档？
- 是否需要新增或更新 example？
- 是否需要更新 component_lab 或 event_system_testbench？
- 是否有自动测试覆盖核心行为？
- GUI 行为无法自动测试时，是否记录了手动 smoke 步骤？
- README 和实际入口是否一致？

## 13. Perf 和 Diagnostics

- 高频路径是否新增分配？
- 是否影响 pointer move、wheel、hover、scroll？
- diagnostics 默认关闭时是否无额外明显成本？
- 是否能定位谁注册事件、谁取消默认行为、谁触发 redraw？
- 大组件树下 layout、paint、event registration 成本是否可区分？

## 14. 回滚策略

- 是否能按组件回滚？
- 是否能按 adapter/helper 回滚？
- 是否保留旧路径作为 fallback？
- 是否避免把多个系统耦合在一个不可拆回的提交里？

## 常见关联矩阵

| 修改对象 | 必查关联 |
| --- | --- |
| Button/Pressable/ClickArea | click default action、hover snapshot、focus activation、disabled/loading、ripple、旧 OnClick/OnHover |
| Checkbox/Switch/RadioGroup | controlled value、OnChange、keyboard activation、state layer、touch target、Ref |
| Input/TextField/SearchBar | beforeinput/input/change、IME、submit、focus、programmatic source、父级滚动 |
| ScrollView/ListView/GridView | wheel policy、axis、offset、ScrollOnChange、hit refresh、virtualization、docs browser |
| Tabs/chips/code/table 横向区域 | horizontal scroll、纵向父滚动、hit refresh、最大宽度 |
| Dialog/Popup/Menu/Select/Tooltip/Toast | portal、outside click、focus restore、Escape、z-order、animation、scrollable protected rect |
| Slider/RangeSlider | pointer drag、capture、keyboard step、value clamp、父级纵向滚动 |
| DragSource/DropTarget | payload、dragover/drop default action、scroll conflict、platform capability |
| Container/Card/Stack/Row/Column/Padding | constraints、decoration、event area、state layer、overflow |

## PR 完成声明模板

```md
## 关联性检查

- 修改范围：
- 关联审查任务：
- 影响组件：
- 自动测试：
- 覆盖率 / race：
- 视觉矩阵 / benchmark：
- 手动 smoke：
- API snapshot / 弃用记录：
- 未覆盖风险：
- 回滚策略：
```
