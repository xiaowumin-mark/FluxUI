package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

// docsAdvancedComponentTemplateDemo is the buildable starting point for a new
// advanced-component document. Keep its state local, make the public value
// controlled, and replace the Select trial with the component under review.
func docsAdvancedComponentTemplateDemo(ctx *ui.Context) ui.Element {
	value := ui.UseState(ctx, "identity")
	status := ui.UseState(ctx, "Ready for a controlled-value test")
	showError := ui.UseState(ctx, false)
	options := []ui.SelectOptionItem[string]{
		{Label: "Collection identity", Value: "identity"},
		{Label: "Roving focus", Value: "focus"},
		{Label: "Anchored overlay", Value: "overlay"},
	}

	return ui.FixedWidthElement(520, ui.ColumnElement(
		ui.TextElement("Advanced component demo template", ui.TextSize(20)),
		ui.SpacerElement(0, 8),
		ui.TextElement("Copy this component, replace the trial control, and keep controlled value, keyboard, overlay and error states observable."),
		ui.SpacerElement(0, 16),
		ui.OutlinedSelectElement(
			value.Value(),
			options,
			ui.SelectLabel[string]("R0 contract trial"),
			ui.SelectSupportingText[string](status.Value()),
			ui.SelectError[string](showError.Value()),
			ui.SelectErrorText[string]("Template: host supplied validation error"),
			ui.SelectOnChange[string](func(ctx *ui.Context, next string) {
				value.Set(next)
				showError.Set(false)
				status.Set("OnChange: " + next)
			}),
		),
		ui.SpacerElement(0, 12),
		ui.RowElement(
			docsDemoControlButton("Reset", func(ctx *ui.Context) {
				value.Set("identity")
				showError.Set(false)
				status.Set("Reset by host state")
			}),
			ui.HSpacerElement(8),
			docsDemoControlButton("Show error", func(ctx *ui.Context) {
				showError.Set(true)
			}),
		),
	))
}
