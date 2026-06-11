# React-style Runtime and Docs Status

This document is the current status page for the React-style runtime and docs/examples rollout. The detailed phase plans are complete and now serve as historical records.

## Current status

- The React-style runtime work is complete: `RunElement`, `Element`, function components, hook slots, effects, providers, reconciler identity, keyed reuse, unmount cleanup, router hooks, stable Element wrappers, and `FromWidget` are implemented and tested.
- Legacy `Run` / `Widget` remains a stable compatibility path. It is not deprecated in code and should not be removed during docs/example cleanup.
- `FromWidget` remains a long-term escape hatch for mixing legacy widgets into Element trees.
- Docs/examples rollout Batches 1-6 are complete. The rollout updated low-risk widget docs with React-style snippets, consolidated redundant standalone examples into the full React-style workspace, and recorded compatibility notes for complex examples.
- Widget docs have been audited for React runtime coverage. Stable wrappers are documented with React-style snippets; host-state/lifecycle-heavy controls still document which state remains owned by the underlying widget host.
- `examples/docs_browser` now runs with `ui.RunElement` and owns the integrated documentation experience: Markdown rendering, metadata-driven example ids, API index, inline previews, popup previews, theme controls, runtime/grid coverage, and System API demos.

## Canonical examples

- `examples/react_workspace` is the canonical React-style runtime showcase for hooks, context, router, keyed identity, fragments, components, and `FromWidget` bridging.
- `examples/counter` remains the compact legacy counter smoke demo.
- `examples/router` remains the comprehensive router compatibility showcase.
- Removed redundant standalone React-style examples: `examples/react_counter` and `examples/router_element`.

## Completed rollout batches

| Batch | Scope | Outcome |
| --- | --- | --- |
| Batch 1 | Stateless display and basic layout docs | React-style snippets added, legacy examples retained. |
| Batch 2 | Simple interaction docs | React-style snippets added for low-risk controlled state examples. |
| Batch 3 | Router docs and example | RouterElement coverage now lives in `examples/react_workspace`; legacy router retained. |
| Batch 4 | Complex input and form strategy | Strategy notes added; complex Element API names not frozen. |
| Batch 5 | Scroll, list, grid, overlays, toast, virtual scroll | Compatibility notes added; lifecycle-heavy examples kept legacy. |
| Batch 6 | Integration showcases | Compatibility and lifecycle notes added; showcase code kept legacy. |

## Widget docs coverage

- Stable Element wrappers are now available across display, layout, sizing, interaction, input, selection, media/card, progress, overlay, scroll/list/grid, navigation, and router APIs.
- Complex widgets keep their existing host-state ownership: text editing, slider drag state, select popup state, scroll offsets, list/grid viewport state, overlay open-change dedupe, toast timers, and ref command queues remain inside the underlying widget path.
- Docs may still use `FromWidget(w Widget) Element` as the long-term escape hatch for custom or legacy-only compositions.
- New React-style docs can reference `TextFieldElement`, `SliderElement`, `SelectElement`, `RadioGroupElement`, `CardElement`, `IconElement`, `ImageElement`, `AppBarElement`, `TabsElement`, `BottomNavigationElement`, `ProgressBarElement`, `CircularProgressElement`, `DialogElement`, `PopupElement`, `ToastElement`, `ScrollViewElement`, `ListViewElement`, `GridElement`, and `GridViewElement`.

## Final wrapper pass

- `ButtonElement`, layout wrappers, `CardElement`, and other composite wrappers render their Element children with context, so nested `ComponentElement`, `Provider`, and `Key` continue to work inside host wrappers.
- Dynamic wrappers such as `ListViewElement` and `GridViewElement` accept item builders returning `Element`; item identity still needs stable business keys when order changes.
- Navigation wrappers include `AppBarElement`, `AppBarElementWithSlots`, and `BottomNavigationElement`; bottom navigation uses `ElementNavItem` so icons can be declared as Elements.

## Full runtime showcase

- `examples/react_workspace` demonstrates the completed React-style mode in one standalone app.
- It covers `RunElement`, function components, `UseState`, `UseMount`, `UseEffectWithDeps`, `Provider`, `UseContext`, `RouterElement`, route hooks, `Fragment`, `ComponentElement`, `Key`, stable Element wrappers, and `FromWidget`.
- Its route pages receive live app settings through route component factories because `RouterElement` still bridges through the legacy router host; `Provider` / `UseContext` coverage stays inside the keyed task-card subtree.
- It keeps `FromWidget` coverage as an escape-hatch example while stable wrappers are available for the full public widget set.
- `examples/docs_browser` is the user-facing docs showcase. It mirrors the same runtime primitives in documented pages, adds popup/inline examples for every docs `example.id`, and includes live System API examples.

## Active strategy documents

- `docs/legacy-api-positioning.md` records the stable legacy compatibility stance.
- `docs/escape-hatch-strategy.md` records the long-term `FromWidget` bridge strategy.
- `docs/deprecation-and-versioning.md` records why Go `Deprecated:` comments are not added yet.
- `docs/examples-inventory.md` records each example's current role and cleanup status.

## Historical planning records

- `docs/react-style-refactor-plan.md` is the detailed runtime and docs rollout history.
- `docs/docs-example-migration-plan.md` is the detailed Batch 1-6 docs/examples rollout history.
- `docs/phase6-root-element-plan.md` is the completed root Element planning record.
- `docs/phase7-legacy-api-plan.md` is the completed legacy API positioning planning record.

These files should be treated as historical records unless a future migration phase explicitly reopens them.

## Cleanup guidance

- Prefer adding concise status or inventory documents over deleting historical plans immediately.
- Do not delete legacy examples only because a React-style counterpart exists.
- Only remove examples that are empty or explicitly superseded by an inventory decision.
- Complex integration examples should be kept until parallel React-style showcases are designed and validated.
