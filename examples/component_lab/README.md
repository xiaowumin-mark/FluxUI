# FluxUI Component Lab

Single-page visual and interaction lab for FluxUI components.

Run:

```sh
go run ./examples/component_lab
```

The example keeps all major Element API components in one scrollable surface and includes:

- Light/dark theme switching, seed color switching, and compact density switching.
- Controlled input, selection, navigation, overlay, list, grid, drag/drop, and router states.
- Ref-driven controls for buttons, pressables, inputs, selection, tabs, dialogs, popups, scroll, and bottom navigation.
- Slider/RangeSlider cursor clipping, hover/pressed state, dialog/popup overlay, and idle redraw smoke coverage.

## P7 Smoke

| Step | Operation | Expected result |
| --- | --- | --- |
| CL-01 | Switch palette, dark mode, and compact density. | Text remains readable and spacing changes without overlapping controls. |
| CL-02 | Scroll from the first section to drag/drop and router, then back. | Top-level virtualization does not produce blank, duplicated, or stale sections. |
| CL-03 | Hover and click decoration, buttons, pressables, chips, and cards. | Hover/pressed/click state stays within the visual target and clears when the pointer leaves. |
| CL-04 | Move across Slider and RangeSlider tracks, then outside them. | Pointer cursor is clipped to the track area and does not stay active for the whole window. |
| CL-05 | Open Select, Menu, Dialog, Popup, Toast, Snackbar, and Tooltip entries. | Overlays do not push layout, and closing them leaves no stale hover/cursor/focus state. |
| CL-06 | Use ScrollRef/ListView/GridView controls. | Inner scroll areas update logs/reach-end state without breaking parent page scrolling. |
