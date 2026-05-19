# Deprecation 与版本节奏

## 目标

制定 legacy API 的 deprecation 文案和版本节奏，确保旧项目不会因为 React-style API 收敛而被迫迁移。

## 策略结论

- 当前阶段不在代码中添加 Go `Deprecated:` 注释。
- 当前阶段只做文档级推荐和迁移说明。
- `Widget` / `Run` / `App` / `Window` / `RunMulti` 继续作为稳定兼容入口。
- 旧 Router API 继续作为稳定兼容入口。
- `FromWidget` 不 deprecated，长期作为 escape hatch 保留。

## 版本节奏

| 阶段 | 动作 | 说明 |
| --- | --- | --- |
| Current | 文档级推荐 | 新项目推荐 `RunElement` / `Element` / `Component`，旧 API 保持稳定。 |
| Next minor | 示例分批迁移 | docs 默认示例逐步切到 React-style，保留 legacy 对照和回退。 |
| Later minor | 兼容层说明强化 | 如果新 API 足够稳定，可在文档中更明确标注 legacy compatibility API。 |
| Future major | 重新评估代码级 deprecation | 只有在替代 API、迁移指南和版本窗口都清楚时才考虑。 |

## 代码注释规则

- 暂不为 `Run` / `App` / `Window` / `RunMulti` 添加 `// Deprecated:`。
- 暂不为旧 Widget 控件添加 `// Deprecated:`。
- 暂不为旧 Router API 添加 `// Deprecated:`。
- 可以在文档中使用 “legacy compatibility API” 表述。
- 未来如果添加 `Deprecated:` 注释，必须同时提供替代 API、迁移示例和版本说明。

## 旧项目保护原则

- 不修改旧 API 签名。
- 不删除 legacy examples。
- 不破坏 docs_browser 的旧示例展示能力。
- 每批 docs/examples 迁移都必须能回退到 legacy 示例。
- 小版本升级不应强迫用户迁移 root API。

## 暂不 deprecated 的 API

- `Widget`
- `Run`
- `App`
- `Window`
- `RunMulti`
- `Router`
- `Route`
- `Navigate`
- `RouteParams`
- `FromWidget`

## 后续重新评估条件

- React-style root API 被实际 examples 和 docs 验证稳定。
- 复杂控件 host-state 生命周期完成迁移或有明确兼容策略。
- Router React-style API 已有完整文档与示例。
- 用户迁移路径和回退路径足够清晰。
- 项目准备进入明确的 major version 窗口。
