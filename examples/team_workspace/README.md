# Team Workspace Compatibility Note

`examples/team_workspace` remains a legacy `Run` / `Widget` integration example during Batch 6 of the React-style docs rollout.

This example is the broadest business-style showcase in the current rollout. It combines shared task data, selected-item state, search and filters, sorting, tabs, responsive layout branches, multiple scroll regions, virtual lists, dialogs, toasts, app bar actions, bottom navigation, and settings controls.

Do not rewrite this example to `RunElement` in-place during the docs rollout. It depends on several host-state-heavy controls whose React-style boundaries are still being documented, including `TextField`, `Select`, `RadioGroup`, `Slider`, `ScrollView`, `ListView`, `Dialog`, and `Toast`.

Future migration should start with a parallel React-style dashboard example after the complex input, list identity, overlay lifecycle, toast cleanup, and shared state strategy is stable. The existing legacy example should remain available as the integration and compatibility reference until that parallel version is proven equivalent.

The current example is useful for regression checks because it exercises many controls together in one app-like workflow without depending on the experimental Element runtime.
