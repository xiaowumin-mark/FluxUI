# 贡献指南

欢迎为 FluxUI 提交 Issue 或 Pull Request。

## 开发流程

```bash
git clone https://github.com/xiaowumin-mark/FluxUI
cd FluxUI
go mod tidy
make test
```

## 提交前检查

| 命令 | 用途 |
|------|------|
| `make test` | 运行全部测试 |
| `make vet` | 静态分析 |
| `make fmt` | 格式化代码 |

## 设计约束

- `ui` 是唯一对外入口，不破坏模块边界
- 新技术能力先在 `ui` 层暴露，底层按依赖方向放置
- 新增能力需补充示例和文档

## Pull Request 要求

1. 保持模块边界清晰，不跨层依赖
2. 添加对应测试
3. 补充示例（`examples/`）或文档（`docs/widgets/`）
4. 运行 `make test` 确保全部通过
