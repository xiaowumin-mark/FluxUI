# 高级组件行为测试模板

适用于尚未公开或刚进入 Beta 的复杂组件。测试放在被测能力附近；跨包行为使用公开 API 测试。测试 fixture 可以使用真实 `internal.Runtime`，但不得因此让生产 `internal` 依赖 `widget` 或 `ui`。

## 夹具约定

- 使用固定 constraints、显式 `runtime.BeginFrame/EndFrame` 和至少两帧渲染来验证跨帧状态。
- 通过 pointer/keyboard/Gio event 注入输入；不要直接改内部字段绕过 default action。
- 测试探针只在 `_test.go` 中读取：回调序列、focused target、overlay stack、viewport、visible keys、redraw reason。
- 每个测试名称描述用户可见合同，例如 `TestComboboxReorderPreservesActiveKey`。

## 必填用例矩阵

| 类别 | 用例 | 预期断言 |
| --- | --- | --- |
| 基线 | 默认值、空项、边界值 | constraints、默认行为和 callback 次数正确。 |
| 受控状态 | 外部 value 与用户操作同帧 | 外部值优先，意图最多一次，不回写旧值。 |
| identity | 重排、插入、删除、重复 key | 状态按 key 不串位；冲突可诊断。 |
| keyboard | Tab、Enter、Space、Escape、Arrow、Home/End | 仅实现的按键有确定行为，disabled 项被跳过。 |
| pointer | press/release、outside click、滚动后命中 | 与 keyboard/default action 合同一致。 |
| lifecycle | 卸载、重挂、Ref 队列 | 命令消费、丢弃或恢复符合文档。 |
| Overlay/scroll | placement、focus restore、nested scroll | topmost 规则、viewport 与 hit refresh 正确。 |
| 回归反例 | 一个过去可能错误的输入组合 | 明确证明该错误不会复现。 |

## 测试记录

```md
## <Component> 行为测试记录

- 被测公开 API / internal helper：
- fixture 文件：
- pointer 路径：
- keyboard 路径：
- controlled/Ref 同帧路径：
- identity/重排/卸载路径：
- Overlay/scroll/platform 路径：
- 已知无法自动化项与 smoke：
- 覆盖率（有效行为分支，目标 ≥80%）：
- 回滚测试：
```

## 最低命令

```powershell
go test ./widget -run 'Test<组件或合同前缀>' -count=1
go test ./<受影响包> -count=1
```

若改动触及 runtime、并发、动画、Hook 或 Ref，补充核心包 `go test -race`，并把结果写入 PR 和 release checklist。
