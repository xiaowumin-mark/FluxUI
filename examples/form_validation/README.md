# R1 Form Validation 示例

`examples/form_validation` 是一个可运行的 React-style `RunElement` 表单示例。它使用 `Form`、`FormField`、`ValidationSummary` 与 `FormRef`，把字段值、同步校验、宿主 pending/error 快照和提交意图全部保留在宿主状态中；示例不启动 goroutine，也不发起网络请求。

## 运行

```sh
go run ./examples/form_validation
```

## 覆盖的行为

- 受控的用户名、邮箱、密码与确认密码 `TextFieldElement` 输入。
- 同步校验：用户名至少 3 个字符、邮箱格式、密码至少 8 个字符以及确认密码一致性。
- `FormRef.Submit()` 触发的 submit intention；按钮并不直接提交业务数据。
- `FormPending(true)` 表示宿主异步阶段，并阻止重复 submit intention。
- `FormSubmitEvent.PreventDefault()` 取消“下次提交取消”或同步校验失败时的本次提交意图。
- `ValidationSummary` 只展示 invalid 字段，并通过稳定 key 调用相应 `FormFieldRef.Focus()`。
- 宿主完成 pending 后可清除错误，或通过“返回宿主异步错误”把邮箱错误快照写回字段与摘要。

## P7 Smoke

| 步骤 | 操作 | 预期结果 |
| --- | --- | --- |
| FV-01 | 直接点击“提交”。 | 同步校验失败；摘要列出必填/格式错误，点击摘要项定位相应字段。 |
| FV-02 | 填写有效用户名、邮箱、密码和确认密码，再点击“提交”。 | 状态显示宿主异步校验 pending；提交按钮显示 loading，重复提交被 `FormPending` 阻止。 |
| FV-03 | 在 pending 后点击“完成异步校验”。 | pending 清除，邮箱没有宿主异步错误。 |
| FV-04 | 在 pending 后点击“返回宿主异步错误”。 | pending 清除；邮箱显示“已注册”错误，ValidationSummary 可定位邮箱字段。 |
| FV-05 | 点击“下次提交取消”，再点击“提交”。 | 宿主调用 `PreventDefault()`，状态显示提交意图已取消，不开始新的 pending。 |
| FV-06 | 在已有同步或异步错误后编辑任一字段。 | 宿主清除旧校验快照、pending 和异步邮箱错误，受控文本与最新输入一致。 |

这个示例刻意把异步完成、错误返回和取消做成可点击的宿主状态转换，便于重复验证而不依赖外部服务。
