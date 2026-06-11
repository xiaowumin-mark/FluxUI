package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

func docsBrowserApp(ctx *ui.Context, runtimeState *docsRuntimeState) ui.Element {
	applyRemoteDocsResult(runtimeState)
	docs := runtimeState.Docs
	loadErr := runtimeState.LoadErr
	docsSource := runtimeState.Source
	onlineLoading := runtimeState.Loading

	selectedDocID := ui.UseState(ctx, "")
	searchKeyword := ui.UseState(ctx, "")
	categoryFilter := ui.UseState(ctx, docsAllCategories)
	themeSeed := ui.UseState(ctx, defaultDocsThemeSeed)
	themeDark := ui.UseState(ctx, false)
	th := docsBrowserTheme(themeSeed.Value(), themeDark.Value())
	markdownCopyStatus := ui.UseState(ctx, "")
	apiCopyStatus := ui.UseState(ctx, "")
	systemDemoStatus := ui.UseState(ctx, "Ready. System actions run only when you click a button.")
	buttonCount := ui.UseState(ctx, 0)
	inputValue := ui.UseState(ctx, "FluxUI")
	checkboxValue := ui.UseState(ctx, true)
	switchValue := ui.UseState(ctx, true)
	sliderValue := ui.UseState(ctx, float32(40))
	styleDemoActive := ui.UseState(ctx, false)
	animationDemoActive := ui.UseState(ctx, false)
	themeDemoDark := ui.UseState(ctx, false)
	radioValue := ui.UseState(ctx, "layout")
	selectValue := ui.UseState(ctx, "medium")
	tabValue := ui.UseState(ctx, "overview")
	dialogOpen := ui.UseState(ctx, false)
	popupOpen := ui.UseState(ctx, false)
	examplePopupOpen := ui.UseState(ctx, false)
	toastMessage := ui.UseState(ctx, "")
	snackbarMessage := ui.UseState(ctx, "")
	snackbarSerial := ui.UseState(ctx, 0)
	snackbarActionCount := ui.UseState(ctx, 0)
	chipSelected := ui.UseState(ctx, true)
	searchBarValue := ui.UseState(ctx, "")
	bottomNavValue := ui.UseState(ctx, "home")
	menuOpen := ui.UseState(ctx, false)
	menuValue := ui.UseState(ctx, "copy")
	listItemSelected := ui.UseState(ctx, "inbox")
	iconButtonSelected := ui.UseState(ctx, true)
	fabCount := ui.UseState(ctx, 0)
	railValue := ui.UseState(ctx, "home")
	drawerValue := ui.UseState(ctx, "inbox")
	clickCount := ui.UseState(ctx, 0)
	appbarActionCount := ui.UseState(ctx, 0)
	listReachEndCount := ui.UseState(ctx, 0)
	hookDemoCount := ui.UseState(ctx, 0)
	hookDemoShowChild := ui.UseState(ctx, false)
	hookDemoLogs := ui.UseState(ctx, []string{})
	routerDemoAllowSettings := ui.UseState(ctx, true)
	routerDemoUserID := ui.UseState(ctx, "u1001")
	routerDemoLog := ui.UseState(ctx, "Router demo ready")
	dragDropDemoState := ui.UseState(ctx, defaultDocsDragDropState())

	filteredDocs := filterDocsByCategory(filterDocs(docs, searchKeyword.Value()), categoryFilter.Value())
	if selectedDocID.Value() == "" && len(filteredDocs) > 0 {
		selectedDocID.Set(filteredDocs[0].Meta.ID)
	}

	currentDoc := findDocByID(filteredDocs, selectedDocID.Value())
	if currentDoc == nil && len(filteredDocs) > 0 {
		selectedDocID.Set(filteredDocs[0].Meta.ID)
		currentDoc = &filteredDocs[0]
	}

	demoState := docsDemoState{
		Theme:                   th,
		SystemDemoStatus:        systemDemoStatus,
		ButtonCount:             buttonCount,
		InputValue:              inputValue,
		CheckboxValue:           checkboxValue,
		SwitchValue:             switchValue,
		SliderValue:             sliderValue,
		StyleDemoActive:         styleDemoActive,
		AnimationDemoActive:     animationDemoActive,
		ThemeDemoDark:           themeDemoDark,
		RadioValue:              radioValue,
		SelectValue:             selectValue,
		TabValue:                tabValue,
		DialogOpen:              dialogOpen,
		PopupOpen:               popupOpen,
		ToastMessage:            toastMessage,
		SnackbarMessage:         snackbarMessage,
		SnackbarSerial:          snackbarSerial,
		SnackbarActionCount:     snackbarActionCount,
		ChipSelected:            chipSelected,
		SearchBarValue:          searchBarValue,
		BottomNavValue:          bottomNavValue,
		MenuOpen:                menuOpen,
		MenuValue:               menuValue,
		ListItemSelected:        listItemSelected,
		IconButtonSelected:      iconButtonSelected,
		FabCount:                fabCount,
		RailValue:               railValue,
		DrawerValue:             drawerValue,
		ClickCount:              clickCount,
		AppbarActionCount:       appbarActionCount,
		ListReachEndCount:       listReachEndCount,
		HookDemoCount:           hookDemoCount,
		HookDemoShowChild:       hookDemoShowChild,
		HookDemoLogs:            hookDemoLogs,
		RouterDemoAllowSettings: routerDemoAllowSettings,
		RouterDemoUserID:        routerDemoUserID,
		RouterDemoLog:           routerDemoLog,
		DragDropDemoState:       dragDropDemoState,
	}
	buildDemo := func(doc *widgetDoc) ui.Element {
		return buildDocsDemo(ctx, doc, demoState)
	}
	leftPanel := docsLeftPanel(
		docs,
		filteredDocs,
		currentDoc,
		selectedDocID,
		searchKeyword,
		categoryFilter,
		themeSeed,
		themeDark,
		th,
		docsSource,
		onlineLoading,
		loadErr,
	)
	rightPanelContent := docsRightPanelContent(
		currentDoc,
		th,
		examplePopupOpen,
		apiCopyStatus,
		markdownCopyStatus,
		buildDemo,
	)

	rightPanel := ui.ExpandedElement(
		ui.ContainerElement(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(16),
			},
			ui.ScrollViewElement(
				ui.ColumnElement(rightPanelContent...),
				ui.ScrollVertical(true),
			),
		),
	)

	return ui.ThemeProviderElement(
		th,
		ui.ContainerElement(
			ui.Style{Background: th.Surface},
			ui.RowElement(
				leftPanel,
				rightPanel,
			),
		),
	)

}
