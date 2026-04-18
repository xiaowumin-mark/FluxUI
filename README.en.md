# FluxUI

FluxUI is a declarative Go UI framework built on top of [Gio](https://gioui.org/).  
It does not replace Gio. Instead, it provides a higher-level API layer with components, state handling, and frame-based animation.

## Goals

- Declarative UI tree
- Per-frame UI rebuild (immediate mode style)
- Centralized state handling
- Frame-tick animation (no goroutine-driven animation loop)
- Production-oriented component and example system

## Core Features

- Basic widgets: `Text`, `Button`, `TextField`, `Checkbox`, `Switch`, `Slider`
- Layout widgets: `Column`, `Row`, `Stack`, `Padding`, `Container`, `ScrollView`
- Navigation & overlays: `Tabs`, `BottomNavigation`, `Dialog`, `Toast`
- Media widgets: `Image`, `Icon`, `Card`
- State: `ui.State[T](ctx)` (controlled component pattern)
- Hooks: `UseEffect`, `UseMount`, `UseLifecycle`
- Command-style refs for external control (scroll, focus, toggle, open/close)
- Multi-window support (desktop)
- Font system: system font discovery + global/local font override

## Architecture

Dependency direction:

`ui -> widget -> (layout/state/anim/event/style) -> internal -> gio`

Key constraints:

- `ui` is the public entry point
- No cross-layer violations
- Keep business logic out of low-level rendering internals

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

import "fluxui/ui"

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

## Examples

```bash
go run ./examples/basic_components
go run ./examples/advanced_components
go run ./examples/layout
go run ./examples/animation
go run ./examples/state_management
go run ./examples/form_validation
go run ./examples/textfield_demo
go run ./examples/theme_custom
go run ./examples/hooks_lifecycle
go run ./examples/multi_window
go run ./examples/vscode_layout
go run ./examples/docs_browser
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

