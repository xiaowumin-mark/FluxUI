# FluxUI

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/xiaowumin-mark/FluxUI)](https://goreportcard.com/report/github.com/xiaowumin-mark/FluxUI)

FluxUI is a declarative Go UI framework built on top of [Gio](https://gioui.org/).  
It does not replace Gio. Instead, it provides a higher-level API layer with components, state handling, and frame-based animation.

## Goals

- Declarative UI tree
- Per-frame UI rebuild (immediate mode style)
- Centralized state handling
- Frame-tick animation (no goroutine-driven animation loop)
- Production-oriented component and example system

## Core Features

- React-style declarative components: `RunElement`, `Element`, function components, `Fragment`, `Key`, `Provider`
- Basic widgets: `Text`, `Button`, `TextField`, `Checkbox`, `Switch`, `Slider`
- Layout widgets: `Column`, `Row`, `Stack`, `Padding`, `Container`, `ScrollView`
- Navigation & overlays: `Router`, `Tabs`, `BottomNavigation`, `Dialog`, `Toast`, `Popup`
- Media widgets: `Image`, `Icon`, `Card`
- State: `State[T](ctx)` / `UseState[T](ctx, initial)` (controlled component pattern)
- Hooks: `UseEffect`, `UseEffectWithDeps`, `UseMount`, `UseLifecycle`, `UseMemo`, `UseRef`, `UseCallback`
- Context: `Provider`, `UseContext`
- Router hooks: `UseParams`, `UseLocation`, `UseNavigate`
- Command-style refs for external control (scroll, focus, toggle, open/close)
- Multi-window support (desktop)
- Font system: system font discovery + global/local font override
- Frame-driven animation: `Animate` + easing functions

## Architecture

Dependency direction:

`ui -> widget -> (layout/state/anim/event/style) -> internal -> gio`

Key constraints:

- `ui` is the public entry point
- No cross-layer violations
- Keep business logic out of low-level rendering internals

Internals include a React-style reconciler (element tree diff + component instance management),
a HookStore (component-level hook slots), and Rules-of-Hooks validation (preventing conditional hook calls).

## Requirements

- Go `1.25+`
- Desktop environments supported by Gio (Windows/macOS/Linux)

## Quick Start

```bash
go mod tidy
go run ./examples/counter
```

Minimal app:

```go
package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		count := ui.State[int](ctx)
		return ui.Center(
			ui.Button(
				ui.Text("Click +1"),
				ui.OnClick(func(ctx *ui.Context) {
					count.Set(count.Value() + 1)
				}),
			),
		)
	}, ui.Title("FluxUI Demo"), ui.Size(900, 600))
}
```

React-style alternative:

```go
package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

func main() {
	_ = ui.RunElement(func(ctx *ui.Context) ui.Element {
		count := ui.UseState(ctx, 0)
		return ui.FromWidget(
			ui.Center(
				ui.Button(
					ui.Text("Click +1"),
					ui.OnClick(func(ctx *ui.Context) {
						count.Set(count.Value() + 1)
					}),
				),
			),
		)
	})
}
```

## Examples

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
go run ./examples/docs_browser
go run ./examples/network_request
go run ./examples/router
go run ./examples/popup_demo
go run ./examples/team_workspace
go run ./examples/virtual_scroll
go run ./examples/fonts
go run ./examples/horizontal_scroll
```

## Docs

- Widget docs: `docs/widgets`
- Docs format and conventions: `docs/README.md`
- In-app docs browser example: `examples/docs_browser`

## Project Structure

```text
fluxui/
├── app/        # app and window runtime entry
├── ui/         # public facade API
├── widget/     # widget implementations
├── layout/     # layout system
├── state/      # state and lifecycle
├── anim/       # frame-based animation
├── event/      # input event wrappers
├── style/      # style system
├── theme/      # theme and fonts
├── internal/   # internal runtime (not public)
├── examples/   # sample apps
└── docs/       # framework docs
```

## Build

> Before Building, please install `gogio cmd` `go install gioui.org/cmd/gogio@latest`

```bash
gogio -target [platform] -o [output] [package]
```

Example:

```bash
gogio -target windows -o example.exe ./examples/counter
```

## Test

```bash
go test ./...
```

## Contributing

Issues and PRs are welcome.

Before submitting:

1. Keep module boundaries clean
2. Add docs and examples for new features
3. Run `go test ./...` and ensure it passes
