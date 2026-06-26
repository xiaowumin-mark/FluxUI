# FluxUI

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/xiaowumin-mark/FluxUI)](https://goreportcard.com/report/github.com/xiaowumin-mark/FluxUI)

FluxUI 是一个基于 [Gio](https://gioui.org/) 的声明式 Go UI 框架。  
它不是替代 Gio，而是在 Gio 之上提供更易用的组件化 API、状态管理与动画能力。

## 项目目标

- 声明式 UI 结构（函数树）
- 每帧重建 UI（Immediate Mode）
- 统一状态管理与生命周期能力
- 帧驱动动画（不依赖 goroutine）
- 面向工程化的组件与示例体系

## 核心能力

- React-style 声明式组件：`RunElement`、`Element`、函数组件、`Fragment`、`Key`、`Provider`
- 基础组件：`Text`、`Button`、`TextField`、`Checkbox`、`Switch`、`Slider`
- 布局组件：`Column`、`Row`、`Stack`、`Padding`、`Container`、`ScrollView`
- 导航与弹层：`Router`、`Tabs`、`BottomNavigation`、`Dialog`、`Toast`、`Popup`
- 媒体组件：`Image`、`Icon`、`Card`
- 状态管理：`State[T](ctx)` / `UseState[T](ctx, initial)`（可控状态模式）
- Hooks：`UseEffect`、`UseEffectWithDeps`、`UseMount`、`UseLifecycle`、`UseMemo`、`UseRef`、`UseCallback`
- Context：`Provider`、`UseContext`
- 路由 Hooks：`UseParams`、`UseLocation`、`UseNavigate`
- 命令式 Ref：可从外部调用组件方法（如滚动、聚焦、切换、打开/关闭）
- 多窗口能力（桌面端）
- 字体系统：系统字体发现 + 全局/局部字体覆盖
- 帧驱动动画：`Animate` + 缓动函数

## 架构说明

依赖方向遵循：

`ui -> widget -> (layout/state/anim/event/style) -> internal -> gio`

设计约束：

- `ui` 是唯一对外入口
- 不跨层调用，不破坏模块边界
- 业务逻辑不写在底层渲染层

内部包含 React-style reconciler（元素树 diff + 组件实例管理）、HookStore（组件级 Hook 槽位），以及 Rules of Hooks 校验（防止条件性 hook 调用）。

## 环境要求

- Go `1.25+`
- 支持 Gio 的桌面运行环境（Windows/macOS/Linux）

## 快速开始

```bash
go mod tidy
go run ./examples/counter
```

最小示例：

```go
package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		count := ui.State[int](ctx)
		return ui.Center(
			ui.Button(
				ui.Text("点击 +1"),
				ui.OnClick(func(ctx *ui.Context) {
					count.Set(count.Value() + 1)
				}),
			),
		)
	}, ui.Title("FluxUI Demo"), ui.Size(900, 600))
}
```

React-style 写法：

```go
package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

func main() {
	_ = ui.RunElement(func(ctx *ui.Context) ui.Element {
		count := ui.UseState(ctx, 0)
		return ui.FromWidget(
			ui.Center(
				ui.Button(
					ui.Text("点击 +1"),
					ui.OnClick(func(ctx *ui.Context) {
						count.Set(count.Value() + 1)
					}),
				),
			),
		)
	})
}
```

## 示例程序

```bash
go run ./examples/counter
go run ./examples/react_workspace
go run ./examples/basic_components
go run ./examples/advanced_components
go run ./examples/layout
go run ./examples/animation_showcase
go run ./examples/state_management
go run ./examples/form_validation
go run ./examples/textfield_demo
go run ./examples/theme_custom
go run ./examples/hooks_lifecycle
go run ./examples/multi_window
go run ./examples/vscode_layout
go run ./examples/icon_fonts
go run ./examples/docs_browser
go run ./examples/network_request
go run ./examples/router
go run ./examples/popup_demo
go run ./examples/team_workspace
go run ./examples/virtual_scroll
go run ./examples/fonts
go run ./examples/horizontal_scroll
go run ./examples/drag_drop_showcase
```

## 构建项目

> 需要先安装gogio cmd `go install gioui.org/cmd/gogio@latest`

```bash
gogio -target [platform] -o [output] [package]
```

示例：

```bash
gogio -target windows -o example.exe ./examples/counter
```


## 文档

- 组件文档目录：`docs/widgets`
- 文档系统说明：`docs/README.md`
- 示例文档浏览器：`examples/docs_browser`

## 项目结构

```text
fluxui/
├── app/        # 应用与窗口运行时入口
├── ui/         # 对外 API 门面层
├── widget/     # 组件实现
├── layout/     # 布局系统
├── state/      # 状态与生命周期
├── anim/       # 帧驱动动画
├── event/      # 输入事件封装
├── style/      # 样式系统
├── theme/      # 主题与字体
├── internal/   # 内部运行时（不对外暴露）
├── examples/   # 示例应用
└── docs/       # 框架文档
```

## 测试

```bash
go test ./...
```

## 贡献

欢迎提交 Issue / PR。  
提交前建议：

1. 保持模块边界清晰，不跨层依赖
2. 为新增能力补充示例与文档
3. 运行 `go test ./...` 确保通过
