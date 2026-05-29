package theme

import "image/color"

type ColorScheme struct {
	Primary            color.NRGBA
	OnPrimary          color.NRGBA
	PrimaryContainer   color.NRGBA
	OnPrimaryContainer color.NRGBA

	Secondary            color.NRGBA
	OnSecondary          color.NRGBA
	SecondaryContainer   color.NRGBA
	OnSecondaryContainer color.NRGBA

	Tertiary            color.NRGBA
	OnTertiary          color.NRGBA
	TertiaryContainer   color.NRGBA
	OnTertiaryContainer color.NRGBA

	Error            color.NRGBA
	OnError          color.NRGBA
	ErrorContainer   color.NRGBA
	OnErrorContainer color.NRGBA

	Background   color.NRGBA
	OnBackground color.NRGBA

	Surface                 color.NRGBA
	OnSurface               color.NRGBA
	SurfaceVariant          color.NRGBA
	OnSurfaceVariant        color.NRGBA
	SurfaceContainerLowest  color.NRGBA
	SurfaceContainerLow     color.NRGBA
	SurfaceContainer        color.NRGBA
	SurfaceContainerHigh    color.NRGBA
	SurfaceContainerHighest color.NRGBA

	Outline        color.NRGBA
	OutlineVariant color.NRGBA

	InverseSurface   color.NRGBA
	InverseOnSurface color.NRGBA
	InversePrimary   color.NRGBA

	Scrim  color.NRGBA
	Shadow color.NRGBA

	Success   color.NRGBA
	OnSuccess color.NRGBA
	Warning   color.NRGBA
	OnWarning color.NRGBA
	Disabled  color.NRGBA

	// SurfaceMuted is a compatibility alias for pre-MD3 FluxUI components.
	// New code should prefer SurfaceVariant or SurfaceContainer* roles.
	SurfaceMuted color.NRGBA
}

func LightColors() ColorScheme {
	return ColorScheme{
		Primary:            color.NRGBA{R: 103, G: 80, B: 164, A: 255},
		OnPrimary:          color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		PrimaryContainer:   color.NRGBA{R: 234, G: 221, B: 255, A: 255},
		OnPrimaryContainer: color.NRGBA{R: 33, G: 0, B: 93, A: 255},

		Secondary:            color.NRGBA{R: 98, G: 91, B: 113, A: 255},
		OnSecondary:          color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		SecondaryContainer:   color.NRGBA{R: 232, G: 222, B: 248, A: 255},
		OnSecondaryContainer: color.NRGBA{R: 29, G: 25, B: 43, A: 255},

		Tertiary:            color.NRGBA{R: 125, G: 82, B: 96, A: 255},
		OnTertiary:          color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		TertiaryContainer:   color.NRGBA{R: 255, G: 216, B: 228, A: 255},
		OnTertiaryContainer: color.NRGBA{R: 49, G: 17, B: 29, A: 255},

		Error:            color.NRGBA{R: 179, G: 38, B: 30, A: 255},
		OnError:          color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		ErrorContainer:   color.NRGBA{R: 249, G: 222, B: 220, A: 255},
		OnErrorContainer: color.NRGBA{R: 65, G: 14, B: 11, A: 255},

		Background:   color.NRGBA{R: 255, G: 251, B: 254, A: 255},
		OnBackground: color.NRGBA{R: 28, G: 27, B: 31, A: 255},

		Surface:                 color.NRGBA{R: 255, G: 251, B: 254, A: 255},
		OnSurface:               color.NRGBA{R: 28, G: 27, B: 31, A: 255},
		SurfaceVariant:          color.NRGBA{R: 231, G: 224, B: 236, A: 255},
		OnSurfaceVariant:        color.NRGBA{R: 73, G: 69, B: 79, A: 255},
		SurfaceContainerLowest:  color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		SurfaceContainerLow:     color.NRGBA{R: 247, G: 242, B: 250, A: 255},
		SurfaceContainer:        color.NRGBA{R: 243, G: 237, B: 247, A: 255},
		SurfaceContainerHigh:    color.NRGBA{R: 236, G: 230, B: 240, A: 255},
		SurfaceContainerHighest: color.NRGBA{R: 230, G: 224, B: 233, A: 255},

		Outline:        color.NRGBA{R: 121, G: 116, B: 126, A: 255},
		OutlineVariant: color.NRGBA{R: 202, G: 196, B: 208, A: 255},

		InverseSurface:   color.NRGBA{R: 49, G: 48, B: 51, A: 255},
		InverseOnSurface: color.NRGBA{R: 244, G: 239, B: 244, A: 255},
		InversePrimary:   color.NRGBA{R: 208, G: 188, B: 255, A: 255},

		Scrim:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Shadow: color.NRGBA{R: 0, G: 0, B: 0, A: 255},

		Success:      color.NRGBA{R: 48, G: 106, B: 71, A: 255},
		OnSuccess:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Warning:      color.NRGBA{R: 116, G: 86, B: 0, A: 255},
		OnWarning:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Disabled:     color.NRGBA{R: 121, G: 116, B: 126, A: 255},
		SurfaceMuted: color.NRGBA{R: 231, G: 224, B: 236, A: 255},
	}
}

func DarkColors() ColorScheme {
	return ColorScheme{
		Primary:            color.NRGBA{R: 208, G: 188, B: 255, A: 255},
		OnPrimary:          color.NRGBA{R: 56, G: 30, B: 114, A: 255},
		PrimaryContainer:   color.NRGBA{R: 79, G: 55, B: 139, A: 255},
		OnPrimaryContainer: color.NRGBA{R: 234, G: 221, B: 255, A: 255},

		Secondary:            color.NRGBA{R: 204, G: 194, B: 220, A: 255},
		OnSecondary:          color.NRGBA{R: 51, G: 45, B: 65, A: 255},
		SecondaryContainer:   color.NRGBA{R: 74, G: 68, B: 88, A: 255},
		OnSecondaryContainer: color.NRGBA{R: 232, G: 222, B: 248, A: 255},

		Tertiary:            color.NRGBA{R: 239, G: 184, B: 200, A: 255},
		OnTertiary:          color.NRGBA{R: 73, G: 37, B: 50, A: 255},
		TertiaryContainer:   color.NRGBA{R: 99, G: 59, B: 72, A: 255},
		OnTertiaryContainer: color.NRGBA{R: 255, G: 216, B: 228, A: 255},

		Error:            color.NRGBA{R: 242, G: 184, B: 181, A: 255},
		OnError:          color.NRGBA{R: 96, G: 20, B: 16, A: 255},
		ErrorContainer:   color.NRGBA{R: 140, G: 29, B: 24, A: 255},
		OnErrorContainer: color.NRGBA{R: 249, G: 222, B: 220, A: 255},

		Background:   color.NRGBA{R: 28, G: 27, B: 31, A: 255},
		OnBackground: color.NRGBA{R: 230, G: 225, B: 229, A: 255},

		Surface:                 color.NRGBA{R: 20, G: 18, B: 24, A: 255},
		OnSurface:               color.NRGBA{R: 230, G: 225, B: 229, A: 255},
		SurfaceVariant:          color.NRGBA{R: 73, G: 69, B: 79, A: 255},
		OnSurfaceVariant:        color.NRGBA{R: 202, G: 196, B: 208, A: 255},
		SurfaceContainerLowest:  color.NRGBA{R: 15, G: 13, B: 19, A: 255},
		SurfaceContainerLow:     color.NRGBA{R: 29, G: 27, B: 32, A: 255},
		SurfaceContainer:        color.NRGBA{R: 33, G: 31, B: 38, A: 255},
		SurfaceContainerHigh:    color.NRGBA{R: 43, G: 41, B: 48, A: 255},
		SurfaceContainerHighest: color.NRGBA{R: 54, G: 52, B: 59, A: 255},

		Outline:        color.NRGBA{R: 147, G: 143, B: 153, A: 255},
		OutlineVariant: color.NRGBA{R: 73, G: 69, B: 79, A: 255},

		InverseSurface:   color.NRGBA{R: 230, G: 225, B: 229, A: 255},
		InverseOnSurface: color.NRGBA{R: 49, G: 48, B: 51, A: 255},
		InversePrimary:   color.NRGBA{R: 103, G: 80, B: 164, A: 255},

		Scrim:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Shadow: color.NRGBA{R: 0, G: 0, B: 0, A: 255},

		Success:      color.NRGBA{R: 170, G: 210, B: 183, A: 255},
		OnSuccess:    color.NRGBA{R: 24, G: 55, B: 36, A: 255},
		Warning:      color.NRGBA{R: 236, G: 194, B: 91, A: 255},
		OnWarning:    color.NRGBA{R: 62, G: 47, B: 0, A: 255},
		Disabled:     color.NRGBA{R: 147, G: 143, B: 153, A: 255},
		SurfaceMuted: color.NRGBA{R: 73, G: 69, B: 79, A: 255},
	}
}

func normalizeColorScheme(cs ColorScheme) ColorScheme {
	if cs.SurfaceVariant.A == 0 {
		cs.SurfaceVariant = firstColor(cs.SurfaceMuted, cs.Surface)
	}
	if cs.SurfaceMuted.A == 0 {
		cs.SurfaceMuted = cs.SurfaceVariant
	}
	if cs.OnSurfaceVariant.A == 0 {
		cs.OnSurfaceVariant = cs.OnSurface
	}
	if cs.SurfaceContainerLowest.A == 0 {
		cs.SurfaceContainerLowest = cs.Surface
	}
	if cs.SurfaceContainerLow.A == 0 {
		cs.SurfaceContainerLow = cs.Surface
	}
	if cs.SurfaceContainer.A == 0 {
		cs.SurfaceContainer = cs.Surface
	}
	if cs.SurfaceContainerHigh.A == 0 {
		cs.SurfaceContainerHigh = cs.SurfaceVariant
	}
	if cs.SurfaceContainerHighest.A == 0 {
		cs.SurfaceContainerHighest = cs.SurfaceVariant
	}
	if cs.PrimaryContainer.A == 0 {
		cs.PrimaryContainer = cs.Primary
	}
	if cs.OnPrimaryContainer.A == 0 {
		cs.OnPrimaryContainer = cs.OnPrimary
	}
	if cs.SecondaryContainer.A == 0 {
		cs.SecondaryContainer = cs.Secondary
	}
	if cs.OnSecondaryContainer.A == 0 {
		cs.OnSecondaryContainer = cs.OnSecondary
	}
	if cs.Tertiary.A == 0 {
		cs.Tertiary = cs.Secondary
	}
	if cs.OnTertiary.A == 0 {
		cs.OnTertiary = cs.OnSecondary
	}
	if cs.TertiaryContainer.A == 0 {
		cs.TertiaryContainer = cs.Tertiary
	}
	if cs.OnTertiaryContainer.A == 0 {
		cs.OnTertiaryContainer = cs.OnTertiary
	}
	if cs.ErrorContainer.A == 0 {
		cs.ErrorContainer = cs.Error
	}
	if cs.OnErrorContainer.A == 0 {
		cs.OnErrorContainer = cs.OnError
	}
	if cs.OutlineVariant.A == 0 {
		cs.OutlineVariant = cs.Outline
	}
	if cs.InverseSurface.A == 0 {
		cs.InverseSurface = cs.OnSurface
	}
	if cs.InverseOnSurface.A == 0 {
		cs.InverseOnSurface = cs.Surface
	}
	if cs.InversePrimary.A == 0 {
		cs.InversePrimary = cs.Primary
	}
	if cs.Scrim.A == 0 {
		cs.Scrim = color.NRGBA{A: 255}
	}
	if cs.Shadow.A == 0 {
		cs.Shadow = color.NRGBA{A: 255}
	}
	return cs
}

func firstColor(primary, fallback color.NRGBA) color.NRGBA {
	if primary.A != 0 {
		return primary
	}
	return fallback
}
