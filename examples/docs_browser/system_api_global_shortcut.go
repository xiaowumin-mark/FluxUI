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
		status := ui.UseState(sectionCtx, "注册全局快捷键，然后按 Ctrl+Alt+Shift+F12。")
		events := ui.UseState(sectionCtx, []string{"暂无快捷键事件。"})
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
				ui.ExpandedElement(docsSystemRunAsyncButton("注册快捷键", status, disabled || registered, func(ctx *ui.Context) string {
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
						return "注册快捷键失败：" + err.Error()
					}
					shortcut.Set(handle)
					return "快捷键已注册。按 Ctrl+Alt+Shift+F12 触发。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("注销快捷键", status, disabled || !registered, func(ctx *ui.Context) string {
					current := shortcut.Value()
					if current == nil {
						return "没有可注销的快捷键。"
					}
					err := current.Close()
					shortcut.Set(nil)
					if err != nil && !system.IsClosed(err) {
						return "注销快捷键失败：" + err.Error()
					}
					return "快捷键已注销。"
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("快捷键事件", events.Value(), th, 98),
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
