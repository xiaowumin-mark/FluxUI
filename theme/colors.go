package theme

import "image/color"

type ColorScheme struct {
	Primary      color.NRGBA
	OnPrimary    color.NRGBA
	Secondary    color.NRGBA
	OnSecondary  color.NRGBA
	Surface      color.NRGBA
	OnSurface    color.NRGBA
	SurfaceMuted color.NRGBA
	Background   color.NRGBA
	OnBackground color.NRGBA
	Error        color.NRGBA
	OnError      color.NRGBA
	Success      color.NRGBA
	OnSuccess    color.NRGBA
	Warning      color.NRGBA
	OnWarning    color.NRGBA
	Disabled     color.NRGBA
	Outline      color.NRGBA
}

func LightColors() ColorScheme {
	return ColorScheme{
		Primary:      color.NRGBA{R: 49, G: 107, B: 255, A: 255},
		OnPrimary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Secondary:    color.NRGBA{R: 100, G: 149, B: 237, A: 255},
		OnSecondary:  color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Surface:      color.NRGBA{R: 248, G: 250, B: 252, A: 255},
		OnSurface:    color.NRGBA{R: 15, G: 23, B: 42, A: 255},
		SurfaceMuted: color.NRGBA{R: 226, G: 232, B: 240, A: 255},
		Background:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		OnBackground: color.NRGBA{R: 15, G: 23, B: 42, A: 255},
		Error:        color.NRGBA{R: 220, G: 53, B: 69, A: 255},
		OnError:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Success:      color.NRGBA{R: 40, G: 167, B: 69, A: 255},
		OnSuccess:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Warning:      color.NRGBA{R: 255, G: 193, B: 7, A: 255},
		OnWarning:    color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Disabled:     color.NRGBA{R: 148, G: 163, B: 184, A: 255},
		Outline:      color.NRGBA{R: 203, G: 213, B: 225, A: 255},
	}
}

func DarkColors() ColorScheme {
	return ColorScheme{
		Primary:      color.NRGBA{R: 66, G: 133, B: 244, A: 255},
		OnPrimary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Secondary:    color.NRGBA{R: 138, G: 180, B: 248, A: 255},
		OnSecondary:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Surface:      color.NRGBA{R: 32, G: 33, B: 36, A: 255},
		OnSurface:    color.NRGBA{R: 232, G: 234, B: 237, A: 255},
		SurfaceMuted: color.NRGBA{R: 60, G: 64, B: 67, A: 255},
		Background:   color.NRGBA{R: 18, G: 18, B: 20, A: 255},
		OnBackground: color.NRGBA{R: 232, G: 234, B: 237, A: 255},
		Error:        color.NRGBA{R: 234, G: 67, B: 53, A: 255},
		OnError:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Success:      color.NRGBA{R: 52, G: 168, B: 83, A: 255},
		OnSuccess:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Warning:      color.NRGBA{R: 251, G: 188, B: 4, A: 255},
		OnWarning:    color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Disabled:     color.NRGBA{R: 97, G: 97, B: 97, A: 255},
		Outline:      color.NRGBA{R: 95, G: 99, B: 104, A: 255},
	}
}
