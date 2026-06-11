package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const (
	docsSystemGlobalShortcutID  = "docs-browser-focus"
	docsSystemGlobalShortcutKey = "F12"
)

const docsSystemGlobalShortcutModifiers = system.GlobalShortcutControl | system.GlobalShortcutAlt | system.GlobalShortcutShift

func docsSystemGlobalShortcutSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "Register a global shortcut, then press Ctrl+Alt+Shift+F12.")
		events := ui.UseState(sectionCtx, []string{"No shortcut events yet."})
		shortcut := ui.UseState[*system.GlobalShortcut](sectionCtx, nil)
		disabled := !system.Supports(system.CapabilityGlobalShortcut)

		currentShortcut := shortcut.Value()
		ui.UseEffectWithDeps(sectionCtx, []any{currentShortcut}, func() func() {
			handle := currentShortcut
			if handle == nil {
				return nil
			}
			return func() {
				_ = handle.Close()
			}
		})

		currentWindowID := ui.CurrentWindowID(sectionCtx)
		currentWindow, hasWindow := ui.GetWindow(currentWindowID)
		registered := currentShortcut != nil
		summary := fmt.Sprintf("Shortcut: Ctrl+Alt+Shift+%s, registered=%v", docsSystemGlobalShortcutKey, registered)

		return docsSystemSection("Global Shortcut API", ui.ColumnElement(
			ui.TextElement(summary, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Register shortcut", status, disabled || registered, func(ctx *ui.Context) string {
					spec := system.GlobalShortcutSpec{
						ID:        docsSystemGlobalShortcutID,
						Key:       docsSystemGlobalShortcutKey,
						Modifiers: docsSystemGlobalShortcutModifiers,
					}
					handle, err := system.RegisterGlobalShortcut(context.Background(), spec, func(event system.GlobalShortcutEvent) {
						if hasWindow {
							_ = currentWindow.Show()
							_ = currentWindow.RequestFocus()
						}
						events.Set(docsSystemPrependLog(events.Value(), formatDocsSystemGlobalShortcutEvent(event), 8))
					})
					if err != nil {
						return "Register shortcut failed: " + err.Error()
					}
					shortcut.Set(handle)
					return "Shortcut registered. Press Ctrl+Alt+Shift+F12 to trigger it."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Unregister shortcut", status, disabled || !registered, func(ctx *ui.Context) string {
					current := shortcut.Value()
					if current == nil {
						return "No shortcut to unregister."
					}
					err := current.Close()
					shortcut.Set(nil)
					if err != nil && !system.IsClosed(err) {
						return "Unregister shortcut failed: " + err.Error()
					}
					return "Shortcut unregistered."
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("Shortcut events", events.Value(), th, 98),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func formatDocsSystemGlobalShortcutEvent(event system.GlobalShortcutEvent) string {
	return fmt.Sprintf("%s id=%s key=%s modifiers=%s", time.Now().Format("15:04:05"), event.ID, event.Key, docsSystemGlobalShortcutModifierText(event.Modifiers))
}

func docsSystemGlobalShortcutModifierText(modifiers system.GlobalShortcutModifiers) string {
	parts := make([]string, 0, 4)
	if modifiers&system.GlobalShortcutControl != 0 {
		parts = append(parts, "Ctrl")
	}
	if modifiers&system.GlobalShortcutAlt != 0 {
		parts = append(parts, "Alt")
	}
	if modifiers&system.GlobalShortcutShift != 0 {
		parts = append(parts, "Shift")
	}
	if modifiers&system.GlobalShortcutMeta != 0 {
		parts = append(parts, "Meta")
	}
	if len(parts) == 0 {
		return "(none)"
	}
	return strings.Join(parts, "+")
}
