# Examples Inventory

This document records the current role of each example after the React-style runtime and docs/examples rollout. It is the cleanup reference for deciding which examples to keep, merge, migrate, or remove.

## Cleanup policy

- Keep legacy examples that exercise compatibility, host state, complex controls, or integration workflows.
- Keep React-style examples that demonstrate `RunElement`, hooks, router APIs, or Element composition.
- Do not delete an example only because a React-style counterpart exists.
- Prefer adding a parallel React-style showcase before replacing an existing integration example.
- Remove only examples that are empty, unreachable, or explicitly superseded by a documented replacement.

## Current inventory

| Example | Role | Runtime | Decision | Notes |
| --- | --- | --- | --- | --- |
| `examples/react_counter` | Canonical React-style counter | React-style `RunElement` | Keep | Minimal state example for docs snippets. |
| `examples/router_element` | Canonical React-style router | React-style `RunElement` | Keep | Demonstrates `RouterElement`, route hooks, params, and navigation. |
| `examples/react_workspace` | Full React-style runtime showcase | React-style `RunElement` | Keep | Demonstrates hooks, context, router, keyed identity, fragments, components, and `FromWidget` bridging. |
| `examples/counter` | Legacy counter counterpart | Legacy `Run` / `Widget` | Keep | Compatibility comparison for `examples/react_counter`. |
| `examples/router` | Legacy router counterpart | Legacy `Run` / `Widget` | Keep | Compatibility comparison for `examples/router_element` and docs browser `router_basic`. |
| `examples/docs_browser` | Docs runtime host | Legacy `Run` / `Widget` | Keep | Owns Markdown loading, metadata parsing, legacy example id mapping, previews, and online fallback. |
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
| `examples/animation` | Smaller animation demo | Legacy `Run` / `Widget` | Review later | Merge candidate with `animation_showcase` or keep as compact smoke demo. |
| `examples/basic_components` | Broad basic widget demo | Legacy `Run` / `Widget` | Review later | Merge candidate; overlaps with widget docs and docs browser examples. |
| `examples/layout` | Basic layout demo | Legacy `Run` / `Widget` | Review later | Merge candidate; overlaps with Batch 1 layout docs. |
| `examples/state_management` | State demo | Legacy `Run` / `Widget` | Review later | Merge candidate or future React-style state comparison candidate. |
| `examples/textfield_demo` | Text input demo | Legacy `Run` / `Widget` | Review later | Merge candidate; keep until input host-state docs are stable. |
| `examples/popup_demo` | Popup demo | Legacy `Run` / `Widget` | Review later | Merge candidate; keep until overlay lifecycle docs are stable. |
| `examples/assets` | Shared assets | N/A | Keep | Contains shared sample assets used by examples. |
| `examples/router_demo` | Empty directory | N/A | Remove | Empty cleanup candidate removed during cleanup. |

## Review-later candidates

These examples are not removed in the cleanup pass because they still compile, demonstrate specific behavior, or may be useful as compact smoke demos:

- `examples/animation`
- `examples/basic_components`
- `examples/layout`
- `examples/state_management`
- `examples/textfield_demo`
- `examples/popup_demo`

Future work should decide whether each one should be merged into a richer showcase, kept as a focused smoke demo, or replaced by a parallel React-style example.

## Removed examples

- `examples/router_demo` was an empty directory and has been removed as a low-risk cleanup item.
