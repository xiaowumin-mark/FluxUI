# Virtual Scroll Compatibility Note

`examples/virtual_scroll` remains a legacy `Run` / `Widget` integration example during Batch 5 of the React-style docs rollout.

This example intentionally stays on the legacy path because it exercises high-volume `ListView` and `GridView` behavior where scroll position, virtual row/window state, reach-end behavior, item identity, and column changes are still owned by the host/list widget layer.

Do not rewrite this example to `RunElement` until the Element APIs for virtual lists and grids have explicit lifecycle rules for host-state reuse, keyed item identity, and cleanup.

The example is still useful as the compatibility reference for validating that the existing virtual scrolling path remains stable while docs add React-style strategy notes elsewhere.
