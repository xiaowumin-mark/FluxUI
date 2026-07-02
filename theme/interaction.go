package theme

import (
	"os"
	"strings"
)

const InteractionQualityEnv = "FLUXUI_INTERACTION_QUALITY"

type InteractionQuality int

const (
	InteractionQualityFull InteractionQuality = iota
	InteractionQualityBalanced
	InteractionQualityLowCPU
)

type ThemeOption func(*Theme)

func WithInteractionQuality(quality InteractionQuality) ThemeOption {
	return func(t *Theme) {
		if t != nil {
			t.SetInteractionQuality(quality)
		}
	}
}

func (q InteractionQuality) String() string {
	switch NormalizeInteractionQuality(q) {
	case InteractionQualityBalanced:
		return "balanced"
	case InteractionQualityLowCPU:
		return "low_cpu"
	default:
		return "full"
	}
}

func NormalizeInteractionQuality(quality InteractionQuality) InteractionQuality {
	switch quality {
	case InteractionQualityFull, InteractionQualityBalanced, InteractionQualityLowCPU:
		return quality
	default:
		return InteractionQualityFull
	}
}

func ParseInteractionQuality(value string) (InteractionQuality, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "":
		return InteractionQualityFull, false
	case "full", "default":
		return InteractionQualityFull, true
	case "balanced", "balance":
		return InteractionQualityBalanced, true
	case "low_cpu", "low-cpu", "lowcpu", "low", "minimal":
		return InteractionQualityLowCPU, true
	default:
		return InteractionQualityFull, false
	}
}

func InteractionQualityFromEnv() (InteractionQuality, bool) {
	return ParseInteractionQuality(os.Getenv(InteractionQualityEnv))
}

func defaultInteractionQuality() InteractionQuality {
	if quality, ok := InteractionQualityFromEnv(); ok {
		return quality
	}
	return InteractionQualityFull
}

func (t *Theme) SetInteractionQuality(quality InteractionQuality) {
	if t == nil {
		return
	}
	t.InteractionQuality = NormalizeInteractionQuality(quality)
}

func (t *Theme) EffectiveInteractionQuality() InteractionQuality {
	if t == nil {
		return defaultInteractionQuality()
	}
	return NormalizeInteractionQuality(t.InteractionQuality)
}
