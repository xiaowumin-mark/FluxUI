# Advanced Components Compatibility Note

`examples/advanced_components` remains a legacy `Run` / `Widget` integration example during Batch 6 of the React-style docs rollout.

This example combines several higher-level controls in one compact showcase: `AppBar`, `Tabs`, `Select`, `Image`, `ListView`, `ScrollView`, `BottomNavigation`, `Dialog`, and `Toast`. It also tracks tab, navigation, select, scroll, reach-end, dialog, and toast state through the legacy `ui.State` path.

Do not rewrite this example in-place to `RunElement` during the docs rollout. Several controls used here still depend on host-state-heavy or lifecycle-heavy behavior, including select popup state, list reach-end callbacks, scroll offset updates, dialog mount/open state, and toast cleanup.

Future migration should start as a parallel React-style advanced components showcase after these controls have stable Element API boundaries. The current legacy example should remain available as the compatibility reference for validating the existing integrated control path.
