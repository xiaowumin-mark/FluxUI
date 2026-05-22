# VSCode Layout Compatibility Note

`examples/vscode_layout` remains a legacy `Run` / `Widget` integration example during Batch 6 of the React-style docs rollout.

This example demonstrates an IDE-style layout with top menus, nested menu panels, sidebar tools, a file list, editor-like `TextField` state, status bar state, and layered click-away handling through `Stack` and `ClickArea`.

Do not rewrite this example in-place to `RunElement` during the docs rollout. Although it is lower risk than `docs_browser` or `team_workspace`, it still combines menu overlay state, list selection identity, editor host state, command-driven updates, and layered event handling.

Future migration should start as a parallel React-style IDE layout example after text input host-state, list identity, menu overlay lifecycle, and click-away behavior are stable. The existing legacy example should remain available as the compatibility reference for validating complex layout and menu interactions.
