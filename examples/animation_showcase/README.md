# Animation Showcase Lifecycle Note

`examples/animation_showcase` remains a legacy `Run` / `Widget` integration example during Batch 6 of the React-style docs rollout.

This example exercises time-based animation behavior across easing comparisons, pulse transitions, staggered entrance, color interpolation, and progress indicators. Each panel drives values through `anim.New(...).Value(ctx)` while legacy `ui.State` controls the target state and tab selection.

Do not rewrite this example in-place to `RunElement` during the docs rollout. Animation state depends on frame time, redraw scheduling, animation track identity, conditional playback state, and tab-scoped context paths such as `ctx.Scope("easing")` and `ctx.Scope("pulse")`.

Future migration should wait for explicit React-style animation lifecycle rules: how animation tracks map to component instances, how `UseEffect` cleanup stops or resets animations, how remounting by key affects in-flight animations, and how redraw requests are scheduled after animation state changes.

The current legacy example should remain available as the compatibility reference for validating that the frame-driven animation path keeps working while React-style lifecycle design is finalized.
