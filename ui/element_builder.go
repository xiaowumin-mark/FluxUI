package ui

import widget "github.com/xiaowumin-mark/FluxUI/widget"

func elementFromWidget(w widget.Widget) Element {
	if w == nil {
		return nil
	}
	return hostElement{child: w}
}
