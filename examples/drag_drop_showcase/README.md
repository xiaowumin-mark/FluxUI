# Drag & Drop Showcase

`examples/drag_drop_showcase` demonstrates the production drag-and-drop API from
the React-style `RunElement` entry point.

## Run

```bash
go run ./examples/drag_drop_showcase
```

## What It Covers

- `ui.DragSourceElement` with text, file URI list, and custom JSON MIME payloads.
- `ui.DropTargetElement` accepting `text/uri-list`, `text/plain`, and `application/json`.
- `DragSourceOnEvent` lifecycle updates for started, requested, completed, and cancelled transfers.
- `DropTargetOnActiveChange` visual feedback while a compatible payload is dragged over the target.
- `DropEvent.Text`, `DropEvent.Paths`, raw bytes, errors, and application-level operation reporting.
- `system.ProbeDragAndDrop` status display, including the conservative `SupportsExternalDragOut` flag.

## Notes

The example validates FluxUI widget-to-widget transfer through Gio `io/transfer`.
External drag-in depends on the operating system and desktop backend. External
drag-out is not enabled by default; only expose that workflow when
`system.ProbeDragAndDrop(ctx).SupportsExternalDragOut` reports true.
