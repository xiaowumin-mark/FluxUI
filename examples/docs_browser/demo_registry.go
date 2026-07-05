package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

type docsDemoState struct {
	Theme                   *ui.Theme
	SystemDemoStatus        docsStringState
	ButtonCount             docsIntState
	InputValue              docsStringState
	CheckboxValue           docsBoolState
	SwitchValue             docsBoolState
	SliderValue             docsFloat32State
	StyleDemoActive         docsBoolState
	AnimationDemoActive     docsBoolState
	ThemeDemoDark           docsBoolState
	RadioValue              docsStringState
	SelectValue             docsStringState
	TabValue                docsStringState
	DialogOpen              docsBoolState
	PopupOpen               docsBoolState
	ToastMessage            docsStringState
	SnackbarMessage         docsStringState
	SnackbarSerial          docsIntState
	SnackbarActionCount     docsIntState
	ChipSelected            docsBoolState
	SearchBarValue          docsStringState
	BottomNavValue          docsStringState
	MenuOpen                docsBoolState
	MenuValue               docsStringState
	ListItemSelected        docsStringState
	IconButtonSelected      docsBoolState
	FabCount                docsIntState
	RailValue               docsStringState
	DrawerValue             docsStringState
	ClickCount              docsIntState
	AppbarActionCount       docsIntState
	ListReachEndCount       docsIntState
	HookDemoCount           docsIntState
	HookDemoShowChild       docsBoolState
	HookDemoLogs            docsStringSliceState
	RouterDemoAllowSettings docsBoolState
	RouterDemoUserID        docsStringState
	RouterDemoLog           docsStringState
	DragDropDemoState       docsDragDropStateHandle
}

func buildDocsDemo(ctx *ui.Context, doc *widgetDoc, state docsDemoState) ui.Element {
	if doc == nil {
		return ui.TextElement("No demo available")
	}

	demoID := doc.Meta.Example.ID
	if demoID == "" {
		demoID = doc.Meta.ID
	}

	th := state.Theme
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}

	switch demoID {
	case "docs_overview":
		return buildDocsOverviewDemo(th)
	case "row_basic":
		return docsRowDemo()
	case "column_basic":
		return docsColumnDemo()
	case "stack_basic":
		return docsStackDemo()
	case "center_basic":
		return docsCenterDemo()
	case "container_basic":
		return docsContainerDemo()
	case "decoration_basic", "decoration_guide_basic":
		return docsDecorationDemo(state.StyleDemoActive, th)
	case "style_showcase":
		return docsStyleShowcaseDemo(th)
	case "insets_basic":
		return docsInsetsDemo()
	case "border_basic":
		return docsBorderDemo(th)
	case "shadow_basic":
		return docsShadowDemo()
	case "gradient_basic":
		return docsGradientDemo()
	case "transform_basic":
		return docsTransformDemo()
	case "image_fill_basic":
		return docsImageFillDemo()
	case "theme_basic":
		return docsThemeDemo(state.ThemeDemoDark)
	case "color_scheme_basic":
		return docsColorSchemeDemo(th)
	case "getting_started_basic":
		return docsGettingStartedDemo(state.ButtonCount, th)
	case "animation_basic":
		return ui.Key("docs-demo-animation", ui.ComponentElement(func(demoCtx *ui.Context) ui.Element {
			return buildDocsAnimationDemo(demoCtx, state.AnimationDemoActive)
		}))
	case "padding_basic":
		return docsPaddingDemo()
	case "spacer_basic":
		return docsSpacerDemo()
	case "divider_basic":
		return docsDividerDemo()
	case "sizing_basic":
		return docsSizingDemo()
	case "pressable_basic", "click_area_basic":
		return docsPressableDemo(state.ClickCount)
	case "text_basic":
		return docsTextDemo(th)
	case "button_basic":
		return docsButtonDemo(state.ButtonCount)
	case "drop_target_basic", "drag_source_basic":
		return buildDocsDragDropDemo(ctx, state.DragDropDemoState, th)
	case "system_api_basic":
		return buildDocsSystemAPIDemo(ctx, state.SystemDemoStatus, th)
	case "event_system_basic":
		return buildDocsEventSystemDemo(th)
	case "material3_showcase":
		return buildDocsMaterial3Showcase(th)
	case "textfield_basic":
		return docsTextFieldDemo(state.InputValue)
	case "checkbox_basic":
		return docsCheckboxDemo(state.CheckboxValue)
	case "switch_basic":
		return docsSwitchDemo(state.SwitchValue)
	case "slider_basic":
		return docsSliderDemo(state.SliderValue)
	case "image_basic":
		return docsImageDemo()
	case "icon_basic":
		return docsIconDemo()
	case "icon_fonts":
		return docsIconFontsDemo()
	case "card_basic":
		return docsCardDemo(state.ButtonCount)
	case "radio_group_basic":
		return docsRadioGroupDemo(state.RadioValue)
	case "select_basic":
		return docsSelectDemo(state.SelectValue)
	case "menu_basic":
		return docsMenuDemo(state.MenuOpen, state.MenuValue, th)
	case "list_item_basic":
		return docsListItemDemo(state.ListItemSelected)
	case "icon_button_basic":
		return docsIconButtonDemo(state.IconButtonSelected)
	case "floating_action_button_basic":
		return docsFloatingActionButtonDemo(state.FabCount)
	case "progress_bar_basic":
		return docsProgressBarDemo(state.SliderValue, th)
	case "circular_progress_basic":
		return docsCircularProgressDemo(state.SliderValue, th)
	case "tabs_basic":
		return docsTabsDemo(state.TabValue)
	case "dialog_basic":
		return docsDialogDemo(state.DialogOpen)
	case "popup_basic":
		return docsPopupDemo(state.PopupOpen)
	case "toast_basic":
		return docsToastDemo(state.ToastMessage)
	case "snackbar_basic":
		return docsSnackbarDemo(state.SnackbarSerial, state.SnackbarMessage, state.SnackbarActionCount, th)
	case "tooltip_basic":
		return docsTooltipDemo()
	case "badge_basic":
		return docsBadgeDemo()
	case "chip_basic":
		return docsChipDemo(state.ChipSelected)
	case "search_bar_basic":
		return docsSearchBarDemo(state.SearchBarValue, th)
	case "progress_indicators_basic":
		return docsProgressIndicatorsDemo(state.SliderValue, th)
	case "scroll_view_basic":
		return docsScrollViewDemo()
	case "list_view_basic":
		return docsListViewDemo(state.ListReachEndCount)
	case "grid_basic":
		return docsGridDemo(th)
	case "app_bar_basic":
		return docsAppBarDemo(state.AppbarActionCount)
	case "bottom_navigation_basic":
		return docsBottomNavigationDemo(state.BottomNavValue)
	case "navigation_rail_basic":
		return docsNavigationRailDemo(state.RailValue)
	case "navigation_drawer_basic":
		return docsNavigationDrawerDemo(state.DrawerValue)
	case "router_basic":
		return docsRouterDemo(ctx, state.RouterDemoAllowSettings, state.RouterDemoUserID, state.RouterDemoLog)
	case "hooks_lifecycle_basic":
		return ui.Key("docs-demo-hooks", ui.ComponentElement(func(demoCtx *ui.Context) ui.Element {
			return buildDocsHooksLifecycleDemo(demoCtx, state.HookDemoCount, state.HookDemoShowChild, state.HookDemoLogs)
		}))
	default:
		return ui.TextElement("This document has no executable demo configured.")
	}
}
