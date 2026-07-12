package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

// docsAdvancedFormsDemo keeps every R1 snapshot in host state. It deliberately
// does not start asynchronous work: the pending switch is a host-provided
// snapshot that makes the ownership boundary visible in the Docs Browser.
func docsAdvancedFormsDemo(ctx *ui.Context) ui.Element {
	searchValue := ui.UseState(ctx, "apple")
	searchQuery := ui.UseState(ctx, "")
	searchOpened := ui.UseState(ctx, false)
	comboboxQuery := ui.UseState(ctx, "")
	comboboxOpened := ui.UseState(ctx, false)
	tagKeys := ui.UseState(ctx, []string{"go"})
	tagQuery := ui.UseState(ctx, "")
	tagOpened := ui.UseState(ctx, false)
	amount := ui.UseState(ctx, "12.50")
	pending := ui.UseState(ctx, false)
	showErrors := ui.UseState(ctx, false)
	status := ui.UseState(ctx, "Ready: all values are controlled by this component.")
	formRef := ui.UseState(ctx, ui.NewFormRef())
	searchFieldRef := ui.UseState(ctx, ui.NewFormFieldRef())
	tagsFieldRef := ui.UseState(ctx, ui.NewFormFieldRef())
	amountFieldRef := ui.UseState(ctx, ui.NewFormFieldRef())

	searchOptions := []ui.SearchSelectItem[string]{
		{Key: "apple", Label: "Apple", Value: "apple"},
		{Key: "apricot", Label: "Apricot", Value: "apricot"},
		{Key: "pear", Label: "Pear", Value: "pear"},
	}
	comboboxOptions := []ui.ComboboxItem[string]{
		{Key: "north", Label: "North campus", Value: "north"},
		{Key: "south", Label: "South campus", Value: "south"},
		{Key: "remote", Label: "Remote", Value: "remote"},
	}
	tagOptions := []ui.TagOptionItem{
		{Key: "go", Label: "Go"},
		{Key: "ui", Label: "UI"},
		{Key: "accessibility", Label: "Accessibility"},
		{Key: "testing", Label: "Testing"},
	}

	amountValue := ui.ParseNumericValue(amount.Value())
	fruitOK := searchValue.Value() != ""
	tagsOK := len(tagKeys.Value()) > 0
	amountOK := amountValue.Valid

	fruitField := ui.FieldState{
		Key:            "fruit",
		Label:          "SearchSelect",
		SupportingText: "Searchable, but selection-only.",
		Required:       true,
	}
	tagsField := ui.FieldState{
		Key:            "tags",
		Label:          "TagPicker",
		SupportingText: "Choose one or more stable tag keys.",
		Required:       true,
	}
	amountField := ui.FieldState{
		Key:            "amount",
		Label:          "NumericField",
		SupportingText: "Exact decimal text; host validates the snapshot.",
		Required:       true,
	}
	if pending.Value() {
		fruitField.Status = ui.FieldPending
		fruitField.PendingText = "Host validation pending"
		tagsField.Status = ui.FieldPending
		tagsField.PendingText = "Host validation pending"
		amountField.Status = ui.FieldPending
		amountField.PendingText = "Host validation pending"
	}
	if showErrors.Value() && !fruitOK {
		fruitField.Status = ui.FieldInvalid
		fruitField.ErrorText = "Choose a fruit from the supplied options."
	}
	if showErrors.Value() && !tagsOK {
		tagsField.Status = ui.FieldInvalid
		tagsField.ErrorText = "Choose at least one tag."
	}
	if showErrors.Value() && !amountOK {
		amountField.Status = ui.FieldInvalid
		amountField.ErrorText = "Enter a valid exact decimal amount."
	}

	invalidFields := make([]ui.FieldState, 0, 3)
	if fruitField.Status == ui.FieldInvalid {
		invalidFields = append(invalidFields, fruitField)
	}
	if tagsField.Status == ui.FieldInvalid {
		invalidFields = append(invalidFields, tagsField)
	}
	if amountField.Status == ui.FieldInvalid {
		invalidFields = append(invalidFields, amountField)
	}

	return ui.FixedWidthElement(560, ui.ColumnElement(
		ui.TextElement("R1 advanced forms", ui.TextSize(20)),
		ui.SpacerElement(0, 6),
		ui.TextElement("SearchSelect only selects options; Combobox also accepts custom text. Use Clear required + Submit to see synchronous cancellation and ValidationSummary focus routing."),
		ui.SpacerElement(0, 8),
		ui.TextElement(status.Value(), ui.TextSize(12)),
		ui.SpacerElement(0, 12),
		ui.FormElement(
			ui.ColumnElement(
				ui.ValidationSummaryElement(
					invalidFields,
					ui.ValidationSummaryEmptyText(""),
					ui.ValidationSummaryOnFocus(func(ctx *ui.Context, key string) {
						switch key {
						case "fruit":
							searchFieldRef.Value().Focus()
						case "tags":
							tagsFieldRef.Value().Focus()
						case "amount":
							amountFieldRef.Value().Focus()
						}
					}),
				),
				ui.FormFieldElement(
					"fruit",
					ui.SearchSelectElement(
						searchValue.Value(),
						searchQuery.Value(),
						searchOptions,
						ui.SearchSelectOpened[string](searchOpened.Value()),
						ui.SearchSelectPlaceholder[string]("Filter fruit"),
						ui.SearchSelectOnQueryChange[string](func(ctx *ui.Context, query string) {
							searchQuery.Set(query)
						}),
						ui.SearchSelectOnOpenChange[string](func(ctx *ui.Context, opened bool) {
							searchOpened.Set(opened)
						}),
						ui.SearchSelectOnChange[string](func(ctx *ui.Context, value string) {
							searchValue.Set(value)
							showErrors.Set(false)
						}),
					),
					ui.FormFieldState(fruitField),
					ui.FormFieldAttachRef(searchFieldRef.Value()),
				),
				ui.SpacerElement(0, 10),
				ui.ComboboxElement(
					comboboxQuery.Value(),
					comboboxOptions,
					ui.ComboboxOpened[string](comboboxOpened.Value()),
					ui.ComboboxLabel[string]("Combobox"),
					ui.ComboboxSupportingText[string]("Enter can submit an enabled custom value."),
					ui.ComboboxOnQueryChange[string](func(ctx *ui.Context, query string) {
						comboboxQuery.Set(query)
					}),
					ui.ComboboxOnOpenChange[string](func(ctx *ui.Context, opened bool) {
						comboboxOpened.Set(opened)
					}),
					ui.ComboboxOnSelect[string](func(ctx *ui.Context, item ui.ComboboxItem[string]) {
						comboboxQuery.Set(item.Label)
						status.Set("Combobox selected: " + item.Label)
					}),
					ui.ComboboxOnCustomValue[string](func(ctx *ui.Context, value string) {
						comboboxQuery.Set(value)
						status.Set("Combobox custom value: " + value)
					}),
				),
				ui.SpacerElement(0, 10),
				ui.FormFieldElement(
					"tags",
					ui.TagPickerElement(
						tagKeys.Value(),
						tagQuery.Value(),
						tagOptions,
						ui.TagPickerOpened(tagOpened.Value()),
						ui.TagPickerPlaceholder("Filter tags"),
						ui.TagPickerOnQueryChange(func(ctx *ui.Context, query string) {
							tagQuery.Set(query)
						}),
						ui.TagPickerOnOpenChange(func(ctx *ui.Context, opened bool) {
							tagOpened.Set(opened)
						}),
						ui.TagPickerOnChange(func(ctx *ui.Context, keys []string) {
							tagKeys.Set(keys)
							showErrors.Set(false)
						}),
					),
					ui.FormFieldState(tagsField),
					ui.FormFieldAttachRef(tagsFieldRef.Value()),
				),
				ui.SpacerElement(0, 10),
				ui.FormFieldElement(
					"amount",
					ui.NumericFieldElement(
						amount.Value(),
						ui.NumericFieldPlaceholder("0.00"),
						ui.NumericFieldMin("0"),
						ui.NumericFieldStep("0.25"),
						ui.NumericFieldOnChange(func(ctx *ui.Context, value string) {
							amount.Set(value)
							showErrors.Set(false)
						}),
					),
					ui.FormFieldState(amountField),
					ui.FormFieldAttachRef(amountFieldRef.Value()),
				),
				ui.SpacerElement(0, 12),
				ui.RowElement(
					ui.FilledButtonElement(
						ui.TextElement("Submit"),
						ui.OnClick(func(ctx *ui.Context) {
							formRef.Value().Submit()
						}),
					),
					ui.HSpacerElement(8),
					ui.OutlinedButtonElement(
						ui.TextElement("Clear required"),
						ui.OnClick(func(ctx *ui.Context) {
							searchValue.Set("")
							tagKeys.Set(nil)
							amount.Set("")
							showErrors.Set(false)
							status.Set("Required values cleared; submit to exercise cancellation.")
						}),
					),
					ui.HSpacerElement(8),
					ui.TextButtonElement(
						ui.TextElement("Toggle pending"),
						ui.OnClick(func(ctx *ui.Context) {
							pending.Set(!pending.Value())
							status.Set("Pending is a host snapshot; no background work was started.")
						}),
					),
				),
			),
			ui.FormPending(pending.Value()),
			ui.FormAttachRef(formRef.Value()),
			ui.FormOnSubmit(func(ctx *ui.Context, event *ui.FormSubmitEvent) {
				if !fruitOK || !tagsOK || !amountOK {
					showErrors.Set(true)
					status.Set("Submit cancelled by synchronous host validation.")
					event.PreventDefault()
					return
				}
				showErrors.Set(false)
				status.Set("Host accepted the submit intention; business submission remains external.")
			}),
		),
	))
}
