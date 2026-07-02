package style

import "image/color"

// ShadowLayer 是单层 CSS box-shadow 语义。
type ShadowLayer struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   color.NRGBA
}

// BoxShadow 描述盒阴影。渲染层会把它生成为可缓存的柔和模糊贴图，
// 避免用多层硬边矩形模拟阴影。
type BoxShadow struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   color.NRGBA
	Layers  []ShadowLayer
}

func (s BoxShadow) IsZero() bool {
	return len(s.EffectiveLayers()) == 0
}

func (s BoxShadow) EffectiveLayers() []ShadowLayer {
	if len(s.Layers) > 0 {
		layers := make([]ShadowLayer, 0, len(s.Layers))
		for _, layer := range s.Layers {
			if layer.Blur > 0 && layer.Color.A > 0 {
				layers = append(layers, layer)
			}
		}
		return layers
	}
	if s.Blur <= 0 || s.Color.A == 0 {
		return nil
	}
	return []ShadowLayer{{OffsetX: s.OffsetX, OffsetY: s.OffsetY, Blur: s.Blur, Spread: s.Spread, Color: s.Color}}
}

// ElevationBoxShadow 返回 Material Design 高度等级对应的阴影预设。
// level [1,5] 映射：1=按钮hover, 2=卡片, 3=浮卡/FAB,
// 4=对话框, 5=模态/抽屉。
func ElevationBoxShadow(level int) BoxShadow {
	switch level {
	case 1:
		return materialElevationShadow(1, 2, 0, 1, 3, 1)
	case 2:
		return materialElevationShadow(1, 2, 0, 2, 6, 2)
	case 3:
		return materialElevationShadow(1, 3, 0, 4, 8, 3)
	case 4:
		return materialElevationShadow(2, 3, 0, 6, 10, 4)
	case 5:
		return materialElevationShadow(4, 4, 0, 8, 12, 6)
	default:
		if level <= 0 {
			return BoxShadow{}
		}
		s := ElevationBoxShadow(5)
		s.Blur = s.Blur * float32(level) / 5.0
		return s
	}
}

func materialElevationShadow(keyY, keyBlur, keySpread, ambientY, ambientBlur, ambientSpread float32) BoxShadow {
	return BoxShadow{
		OffsetY: ambientY,
		Blur:    ambientBlur,
		Spread:  ambientSpread,
		Color:   color.NRGBA{R: 0, G: 0, B: 0, A: 38},
		Layers: []ShadowLayer{
			{OffsetY: keyY, Blur: keyBlur, Spread: keySpread, Color: color.NRGBA{R: 0, G: 0, B: 0, A: 77}},
			{OffsetY: ambientY, Blur: ambientBlur, Spread: ambientSpread, Color: color.NRGBA{R: 0, G: 0, B: 0, A: 38}},
		},
	}
}
