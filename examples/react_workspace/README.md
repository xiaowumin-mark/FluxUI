# React Workspace Example

`examples/react_workspace` is the canonical full React-style runtime showcase for FluxUI.

It intentionally uses `RunElement` and stable Element APIs instead of rewriting an existing legacy showcase. The example is a small workspace/task app that demonstrates how the completed React-style runtime fits together in one program.

Covered runtime features:

- `RunElement` as the application root.
- Function components returning `Element`.
- `UseState` for component-owned HookSlot state.
- `UseMount` and `UseEffectWithDeps` for lifecycle and dependency-based effects.
- `Provider` and `UseContext` for typed context values.
- `RouterElement`, `RouteElement`, `UseNavigate`, `UseLocation`, and `UseParams` for routing.
- `Fragment`, `ComponentElement`, and `Key` for composition and keyed component identity.
- `FromWidget` for bridging legacy widgets such as `Card` and `ProgressBar` into an Element tree.

The task list page includes a rotate-order action. Each task card has local expanded state and is wrapped with `Key(task.ID, ComponentElement(...))`, so the example can be used to verify that component state follows stable business ids instead of list indexes.

Route pages receive the live workspace snapshot through route component factories because `RouterElement` still bridges through the legacy router host. `Provider` / `UseContext` is demonstrated inside the keyed task-card subtree, where the provider and consumer stay within the same Element render pass.

This example does not depend on unfrozen wrappers such as `TextFieldElement`, `TabsElement`, `BottomNavigationElement`, `CardElement`, or `ProgressBarElement`. Legacy widgets are shown through `FromWidget` where the Element wrapper is not stable yet.

Run it with:

```sh
go run ./examples/react_workspace
```
