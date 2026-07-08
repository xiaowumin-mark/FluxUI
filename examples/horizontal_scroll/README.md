# Horizontal Scroll Example

`examples/horizontal_scroll` is a React-style `RunElement` example for horizontal-only scroll behavior.

## Run

```sh
go run ./examples/horizontal_scroll
```

## What It Covers

The example contains two `ScrollViewElement` cases with `ScrollHorizontal(true)` and `ScrollVertical(false)`: a single long row and multi-line long content. The offset label reports the current x/y values from `ScrollOnChange`.

## P7 Smoke

| Step | Operation | Expected result |
| --- | --- | --- |
| HS-01 | Use a touchpad or horizontal wheel over the card row. | The x offset increases and the card row moves horizontally. |
| HS-02 | Use a normal vertical wheel over the horizontal row. | The horizontal x offset should not change because vertical wheel is not converted to horizontal scroll. |
| HS-03 | Scroll the multi-line long content horizontally. | All long lines move together without stretching or tearing. |
| HS-04 | Observe the offset label while scrolling. | x changes for horizontal input; y remains unchanged for the horizontal-only cases. |
