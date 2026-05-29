package theme

type TextStyle struct {
	Size       float32
	LineHeight float32
	Weight     FontWeight
}

type TypeScale struct {
	DisplayLarge   TextStyle
	DisplayMedium  TextStyle
	DisplaySmall   TextStyle
	HeadlineLarge  TextStyle
	HeadlineMedium TextStyle
	HeadlineSmall  TextStyle
	TitleLarge     TextStyle
	TitleMedium    TextStyle
	TitleSmall     TextStyle
	BodyLarge      TextStyle
	BodyMedium     TextStyle
	BodySmall      TextStyle
	LabelLarge     TextStyle
	LabelMedium    TextStyle
	LabelSmall     TextStyle
}

func DefaultTypeScale() TypeScale {
	return TypeScale{
		DisplayLarge:   TextStyle{Size: 57, LineHeight: 64, Weight: FontWeightNormal},
		DisplayMedium:  TextStyle{Size: 45, LineHeight: 52, Weight: FontWeightNormal},
		DisplaySmall:   TextStyle{Size: 36, LineHeight: 44, Weight: FontWeightNormal},
		HeadlineLarge:  TextStyle{Size: 32, LineHeight: 40, Weight: FontWeightNormal},
		HeadlineMedium: TextStyle{Size: 28, LineHeight: 36, Weight: FontWeightNormal},
		HeadlineSmall:  TextStyle{Size: 24, LineHeight: 32, Weight: FontWeightNormal},
		TitleLarge:     TextStyle{Size: 22, LineHeight: 28, Weight: FontWeightNormal},
		TitleMedium:    TextStyle{Size: 16, LineHeight: 24, Weight: FontWeightMedium},
		TitleSmall:     TextStyle{Size: 14, LineHeight: 20, Weight: FontWeightMedium},
		BodyLarge:      TextStyle{Size: 16, LineHeight: 24, Weight: FontWeightNormal},
		BodyMedium:     TextStyle{Size: 14, LineHeight: 20, Weight: FontWeightNormal},
		BodySmall:      TextStyle{Size: 12, LineHeight: 16, Weight: FontWeightNormal},
		LabelLarge:     TextStyle{Size: 14, LineHeight: 20, Weight: FontWeightMedium},
		LabelMedium:    TextStyle{Size: 12, LineHeight: 16, Weight: FontWeightMedium},
		LabelSmall:     TextStyle{Size: 11, LineHeight: 16, Weight: FontWeightMedium},
	}
}
