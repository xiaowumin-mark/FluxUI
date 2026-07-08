# Advanced Components Example

`examples/advanced_components` is a React-style `RunElement` integration example.

## Run

```sh
go run ./examples/advanced_components
```

## What It Covers

This example combines higher-level controls in one compact showcase: `AppBar`, `Tabs`, `Select`, `Image`, `ListView`, `ScrollView`, `BottomNavigation`, `Dialog`, and `Toast`. It tracks tab, navigation, select, scroll, reach-end, dialog, and toast state through `ui.UseState`.

## P7 Smoke

| Step | Operation | Expected result |
| --- | --- | --- |
| AC-01 | Switch tabs and bottom navigation items. | Active state, indicator, and visible content update without layout jumps. |
| AC-02 | Open `Select`, choose a different priority, then reopen it. | The popup closes after selection, the selected value persists, and toast state is updated once per real change. |
| AC-03 | Scroll the main content and the embedded list to the end. | `ScrollOnChange` reports a non-zero y offset and the list reach-end counter can increment. |
| AC-04 | Click the image and open/close the dialog through confirm, cancel, and mask close. | Toast/dialog lifecycle stays isolated from page layout and scroll state. |
