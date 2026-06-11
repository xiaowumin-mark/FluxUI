package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func shouldCenterDemo(exampleID string) bool {
	switch exampleID {
	case "button_basic",
		"checkbox_basic",
		"switch_basic",
		"slider_basic",
		"menu_basic",
		"image_basic",
		"icon_basic",
		"icon_button_basic",
		"floating_action_button_basic",
		"card_basic",
		"list_item_basic",
		"tooltip_basic",
		"badge_basic",
		"chip_basic",
		"radio_group_basic",
		"progress_bar_basic",
		"circular_progress_basic",
		"progress_indicators_basic",
		"tabs_basic":
		return true
	default:
		return false
	}
}

func rowColor(index int) color.NRGBA {
	if index%2 == 0 {
		return ui.NRGBA(241, 245, 249, 255)
	}
	return ui.NRGBA(226, 232, 240, 255)
}

func hoverScale(hovered bool) float32 {
	if hovered {
		return 1.18
	}
	return 1
}

func easeOutElastic(v float32) float32 {
	v = clamp01(v)
	if v == 0 || v == 1 {
		return v
	}
	return float32(math.Pow(2, -10*float64(v)))*float32(math.Sin(float64(v*10-0.75)*2*math.Pi/3)) + 1
}

func clamp01(v float32) float32 {
	if math.IsNaN(float64(v)) || v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func appendDemoLog(getLogs func() []string, setLogs func([]string), message string) {
	if getLogs == nil || setLogs == nil {
		return
	}
	items := append([]string{}, getLogs()...)
	items = append(items, fmt.Sprintf("%s  %s", time.Now().Format("15:04:05"), message))
	if len(items) > 8 {
		items = items[len(items)-8:]
	}
	setLogs(items)
}
