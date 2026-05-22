# Docs Browser Compatibility Note

`examples/docs_browser` remains a legacy `Run` / `Widget` integration example during Batch 6 of the React-style docs rollout.

This example is the docs runtime host, not only a visual showcase. It loads widget Markdown files, parses metadata, maps legacy example ids to preview widgets, renders interactive demos, and falls back to online docs when local files are unavailable.

Do not rewrite this example to `RunElement` during the docs rollout, and do not change its runtime example mapping as part of widget documentation updates.

Future migration should be designed as a separate docs browser runtime project. That design needs explicit rules for preserving legacy example ids, hosting React-style preview snippets side by side, loading docs from local and remote sources, and keeping preview state stable across navigation and search.

The current legacy path remains the compatibility reference for validating that migrated Markdown docs still render in the existing browser.
