package style

import (
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/theme"
)

const (
	StateLayerHoverOpacity   float32 = 0.08
	StateLayerFocusOpacity   float32 = 0.12
	StateLayerPressedOpacity float32 = 0.12
	StateLayerDraggedOpacity float32 = 0.16
)

func StateLayer(container, onColor color.NRGBA, opacity float32) color.NRGBA {
	return MixNRGBA(onColor, container, opacity)
}

func DisabledContainer(onSurface color.NRGBA) color.NRGBA {
	return withAlpha(onSurface, 31)
}

func DisabledContent(onSurface color.NRGBA) color.NRGBA {
	return withAlpha(onSurface, 97)
}

func SurfaceAtElevation(cs theme.ColorScheme, level int) color.NRGBA {
	if level <= 0 {
		return cs.Surface
	}
	if level > 5 {
		level = 5
	}
	return MixNRGBA(cs.Primary, cs.Surface, tonalElevationOverlay(level))
}

func ElevationShadow(cs theme.ColorScheme, level int) BoxShadow {
	if level <= 0 {
		return BoxShadow{}
	}
	shadow := ElevationBoxShadow(level)
	if cs.Shadow.A != 0 {
		for i := range shadow.Layers {
			alpha := shadow.Layers[i].Color.A
			shadow.Layers[i].Color = cs.Shadow
			shadow.Layers[i].Color.A = alpha
		}
		shadow.Color = cs.Shadow
		shadow.Color.A = 38
	}
	return shadow
}

func MixNRGBA(fg, bg color.NRGBA, amount float32) color.NRGBA {
	amount = clamp01(amount)
	inv := 1 - amount
	return color.NRGBA{
		R: uint8(float32(bg.R)*inv + float32(fg.R)*amount + 0.5),
		G: uint8(float32(bg.G)*inv + float32(fg.G)*amount + 0.5),
		B: uint8(float32(bg.B)*inv + float32(fg.B)*amount + 0.5),
		A: uint8(float32(bg.A)*inv + float32(fg.A)*amount + 0.5),
	}
}

func tonalElevationOverlay(level int) float32 {
	switch level {
	case 1:
		return 0.05
	case 2:
		return 0.08
	case 3:
		return 0.11
	case 4:
		return 0.12
	default:
		return 0.14
	}
}

func withAlpha(col color.NRGBA, alpha uint8) color.NRGBA {
	col.A = alpha
	return col
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
