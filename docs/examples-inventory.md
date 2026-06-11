# Examples Inventory

This document records the current role of each example after the React-style runtime and docs/examples rollout. It is the cleanup reference for deciding which examples to keep, merge, migrate, or remove.

## Cleanup policy

- Keep legacy examples that exercise compatibility, host state, complex controls, or integration workflows.
- Keep React-style examples that demonstrate broad `RunElement`, hooks, router APIs, or Element composition coverage.
- Delete narrow examples when a documented, broader example supersedes them and the compact legacy smoke demo remains.
- Prefer adding a parallel React-style showcase before replacing an existing integration example.
- Remove only examples that are empty, unreachable, or explicitly superseded by a documented replacement.

## Current inventory

| Example | Role | Runtime | Decision | Notes |
| --- | --- | --- | --- | --- |
| `examples/react_workspace` | Canonical React-style runtime showcase | React-style `RunElement` | Keep | Demonstrates hooks, context, router, keyed identity, fragments, components, stable Element wrappers, transitions, and `FromWidget` bridging. |
| `examples/counter` | Minimal counter smoke demo | Legacy `Run` / `Widget` | Keep | Compact legacy state example. |
| `examples/router` | Router compatibility showcase | Legacy `Run` / `Widget` | Keep | Comprehensive router demo and docs browser `router_basic` counterpart. |
| `examples/docs_browser` | Docs runtime host | React-style `RunElement` | Keep | Owns Markdown loading, metadata parsing, search/category filters, MD3 theme switching, API index, inline and popup previews, runtime/grid coverage, System API demos, and online fallback. |
| `examples/drag_drop_showcase` | Drag-and-drop API showcase | React-style `RunElement` | Keep | Demonstrates `DragSourceElement`, `DropTargetElement`, MIME payloads, lifecycle events, active/error callbacks, and conservative system probe output. |
| `examples/form_validation` | Complex input workflow | Legacy `Run` / `Widget` | Keep | Compatibility note exists; wait for stable input host-state strategy. |
| `examples/virtual_scroll` | Virtual list/grid performance reference | Legacy `Run` / `Widget` | Keep | Compatibility note exists; wait for list/grid lifecycle and identity strategy. |
| `examples/team_workspace` | Business dashboard integration showcase | Legacy `Run` / `Widget` | Keep | Compatibility note exists; migrate only through a parallel React-style dashboard. |
| `examples/advanced_components` | Higher-level component integration showcase | Legacy `Run` / `Widget` | Keep | Compatibility note exists; candidate for a future parallel React-style showcase. |
| `examples/animation_showcase` | Animation lifecycle showcase | Legacy `Run` / `Widget` | Keep | Lifecycle note exists; wait for React-style animation lifecycle rules. |
| `examples/vscode_layout` | IDE-style layout and menu showcase | Legacy `Run` / `Widget` | Keep | Compatibility note exists; wait for menu overlay and input host-state strategy. |
| `examples/hooks_lifecycle` | Hook and lifecycle behavior demo | Legacy `Run` / `Widget` | Keep | Useful for runtime/lifecycle smoke testing. |
| `examples/network_request` | Async/network workflow demo | Legacy `Run` / `Widget` | Keep | Useful for async state and request behavior. |
| `examples/multi_window` | Multi-window feature demo | Legacy `Run` / `Widget` | Keep | Element multi-window entry points are not planned yet. |
| `examples/fonts` | Font rendering demo | Legacy `Run` / `Widget` | Keep | Feature-specific smoke demo. |
| `examples/theme_custom` | Theme customization demo | Legacy `Run` / `Widget` | Keep | Feature-specific smoke demo. |
| `examples/horizontal_scroll` | Scroll behavior demo | Legacy `Run` / `Widget` | Keep | Useful scroll compatibility reference. |
| `examples/basic_components` | Broad basic widget demo | Legacy `Run` / `Widget` | Review later | Merge candidate; overlaps with widget docs and docs browser examples. |
| `examples/layout` | Basic layout demo | Legacy `Run` / `Widget` | Review later | Merge candidate; overlaps with Batch 1 layout docs. |
| `examples/state_management` | State demo | Legacy `Run` / `Widget` | Review later | Merge candidate or future React-style state comparison candidate. |
| `examples/textfield_demo` | Text input demo | Legacy `Run` / `Widget` | Review later | Merge candidate; keep until input host-state docs are stable. |
| `examples/popup_demo` | Popup demo | Legacy `Run` / `Widget` | Review later | Merge candidate; keep until overlay lifecycle docs are stable. |
| `examples/assets` | Shared assets | N/A | Keep | Contains shared sample assets used by examples. |
| `examples/router_demo` | Empty directory | N/A | Remove | Empty cleanup candidate removed during cleanup. |

## Review-later candidates

These examples are not removed in the cleanup pass because they still compile, demonstrate specific behavior, or may be useful as compact smoke demos:

- `examples/basic_components`
- `examples/layout`
- `examples/state_management`
- `examples/textfield_demo`
- `examples/popup_demo`

Future work should decide whether each one should be merged into a richer showcase, kept as a focused smoke demo, or replaced by a parallel React-style example.

## Removed examples

- `examples/router_demo` was an empty directory and has been removed as a low-risk cleanup item.
- `examples/react_counter` was removed because `examples/react_workspace` covers the React-style `UseState` path and `examples/counter` remains the compact counter smoke demo.
- `examples/router_element` was removed because `examples/react_workspace` covers `RouterElement` and `examples/router` remains the comprehensive router showcase.
- `examples/animation` was removed because `examples/animation_showcase` covers the same animation helpers and broader lifecycle scenarios.
