# Virtual Scroll Example

`examples/virtual_scroll` is a React-style `RunElement` example for high-volume list and grid rendering.

## Run

```sh
go run ./examples/virtual_scroll
```

## What It Covers

The example exercises high-volume `ListViewElement` and `GridViewElement` behavior with 50,000 list rows and 100,000 grid cells. It is the manual smoke entry for virtual window stability, item identity, tab switching, and grid sizing.

## P7 Smoke

| Step | Operation | Expected result |
| --- | --- | --- |
| VS-01 | Scroll the ListView tab quickly through a large range. | Visible item numbers remain continuous, with no blank or duplicated rows. |
| VS-02 | Switch to the GridView tab and scroll. | Four columns remain stable, gap spacing stays consistent, and cell labels match their index. |
| VS-03 | Switch between ListView and GridView repeatedly. | Active tab state changes cleanly and the app stays responsive. |
| VS-04 | Resize the window modestly and repeat VS-01/VS-02. | Content remains accessible through scrolling without layout overflow. |
