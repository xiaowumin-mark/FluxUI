# FluxUI Docs Browser

`examples/docs_browser` is the interactive documentation host for FluxUI. It runs with `ui.RunElement`, loads Markdown docs from `docs/`, parses `fluxui-doc-meta`, and renders a live preview for every documented example id.

## Structure

- `main.go`: startup only; it wires docs loading into `ui.RunElement`.
- `docs_sources.go`: local docs discovery, GitHub fallback, HTTP fetching, and load-error documents.
- `docs_metadata.go`, `docs_overview.go`: `fluxui-doc-meta` parsing, sorting, generated overview document, menu entry construction, and doc lookup helpers.
- `docs_browser_app.go`: top-level state wiring and page assembly.
- `demo_registry.go`: the only place that maps docs `example.id` values to executable demos.
- `left_panel.go`, `search.go`, `api_search.go`, `category_filter.go`: navigation list, multi-term document/API search, API quick-reference matches, counted category filters, theme controls, and load status.
- `right_panel.go`, `example_viewer.go`: document body, API index, inline demo, and popup demo.
- `markdown_renderer.go`, `api_index.go`: Markdown H1-H6 headings, nested/task/numbered lists, blockquotes, tables, code blocks, copy buttons, and API list rendering.
- `controls_core_demo.go`, `controls_feedback_demo.go`, `controls_selection_demo.go`, `controls_media_demo.go`, `controls_overlay_demo.go`, `controls_progress_demo.go`: focused control demos split by responsibility.
- `*_demo.go`: focused demo modules for layout/style, layout/grid, navigation/router, drag/drop, Material 3, animation, runtime/hooks, and System API.

## Coverage Rules

Every Markdown document with an `example.id` must have a matching case in `demo_registry.go`. The tests in `docs_metadata_test.go` enforce:

- all loaded docs reference known example ids;
- the generated `docs_browser_overview` page loads first and exposes a live overview demo;
- every referenced example id has a registry case;
- every loaded document can build its demo and right panel content;
- metadata, API signatures, and Markdown content are searchable;
- theme seed and dark mode changes produce distinct MD3 color schemes;
- key UI sections such as Markdown rendering, API index, popup demos, Material 3, drag/drop, runtime/hooks, grid, Router, and System API build without nil results;
- representative React-style APIs stay visible in the docs browser demo source, including `GridViewElement`, `ScrollHorizontal`, `ScrollAutoToEndKey`, `ListVirtualized`, hooks/context/router APIs, command refs such as `ButtonAttachRef`, `InputAttachRef`, `TabsAttachRef`, `DialogAttachRef`, and common MD3 variants/options such as filled/tonal/elevated buttons, outlined/filled text fields, filled/elevated/outlined cards, chip slots, menu sizing, navigation colors/alignment, search/badge/tooltip options, decoration events, drag/drop lifecycle constants, typography options, progress indeterminate/size states, and System API owner/window-event/notification-backend/tray-update paths.

When adding a new docs example, prefer adding a small focused `*_demo.go` module and then registering it in `demo_registry.go`. Avoid putting new demo implementations directly into `docs_browser_app.go`.
