package main

import ui "github.com/xiaowumin-mark/FluxUI/ui"

type docsDemoPresentation struct {
	Height float32
	Center bool
	Scroll bool
}

func docsDemoPresentationFor(exampleID string) docsDemoPresentation {
	presentation := docsDemoPresentation{Height: 230}
	switch exampleID {
	case "docs_overview":
		presentation.Height = 310
	case "animation_basic":
		presentation.Height = 300
		presentation.Scroll = true
	case "router_basic":
		presentation.Height = 320
	case "drop_target_basic", "drag_source_basic":
		presentation.Height = 430
	case "grid_basic":
		presentation.Height = 380
		presentation.Scroll = true
	case "scroll_view_basic", "list_view_basic":
		presentation.Height = 390
		presentation.Scroll = true
	case "material3_showcase":
		presentation.Height = 440
		presentation.Scroll = true
	case "system_api_basic":
		presentation.Height = 360
		presentation.Scroll = true
	case "hooks_lifecycle_basic":
		presentation.Height = 430
		presentation.Scroll = true
	case "navigation_rail_basic", "navigation_drawer_basic":
		presentation.Height = 280
	}
	presentation.Center = shouldCenterDemo(exampleID)
	return presentation
}

func docsDemoViewport(exampleID string, content ui.Element) ui.Element {
	presentation := docsDemoPresentationFor(exampleID)
	viewport := ui.FillElement(content)
	if presentation.Scroll {
		return ui.FillElement(ui.ScrollViewElement(content, ui.ScrollVertical(true)))
	}
	if presentation.Center {
		return ui.CenterElement(content)
	}
	return viewport
}

func isDocsDemoKnown(exampleID string) bool {
	switch exampleID {
	case "docs_overview",
		"row_basic",
		"column_basic",
		"stack_basic",
		"center_basic",
		"container_basic",
		"decoration_basic",
		"decoration_guide_basic",
		"insets_basic",
		"border_basic",
		"shadow_basic",
		"gradient_basic",
		"transform_basic",
		"image_fill_basic",
		"theme_basic",
		"color_scheme_basic",
		"getting_started_basic",
		"animation_basic",
		"padding_basic",
		"spacer_basic",
		"divider_basic",
		"sizing_basic",
		"pressable_basic",
		"click_area_basic",
		"text_basic",
		"button_basic",
		"drop_target_basic",
		"drag_source_basic",
		"system_api_basic",
		"material3_showcase",
		"textfield_basic",
		"checkbox_basic",
		"switch_basic",
		"slider_basic",
		"image_basic",
		"icon_basic",
		"card_basic",
		"radio_group_basic",
		"select_basic",
		"menu_basic",
		"list_item_basic",
		"icon_button_basic",
		"floating_action_button_basic",
		"progress_bar_basic",
		"circular_progress_basic",
		"tabs_basic",
		"dialog_basic",
		"popup_basic",
		"toast_basic",
		"snackbar_basic",
		"tooltip_basic",
		"badge_basic",
		"chip_basic",
		"search_bar_basic",
		"progress_indicators_basic",
		"scroll_view_basic",
		"list_view_basic",
		"grid_basic",
		"app_bar_basic",
		"bottom_navigation_basic",
		"navigation_rail_basic",
		"navigation_drawer_basic",
		"router_basic",
		"hooks_lifecycle_basic":
		return true
	default:
		return false
	}
}
