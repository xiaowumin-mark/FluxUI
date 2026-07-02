package theme

import "testing"

func TestInteractionQualityParsing(t *testing.T) {
	tests := map[string]InteractionQuality{
		"full":     InteractionQualityFull,
		"balanced": InteractionQualityBalanced,
		"low_cpu":  InteractionQualityLowCPU,
		"low-cpu":  InteractionQualityLowCPU,
		"lowcpu":   InteractionQualityLowCPU,
	}
	for input, want := range tests {
		got, ok := ParseInteractionQuality(input)
		if !ok {
			t.Fatalf("expected %q to parse", input)
		}
		if got != want {
			t.Fatalf("ParseInteractionQuality(%q) = %v, want %v", input, got, want)
		}
	}

	if _, ok := ParseInteractionQuality("unknown"); ok {
		t.Fatal("expected invalid quality to be rejected")
	}
}

func TestThemeInteractionQualityOptionAndEnv(t *testing.T) {
	t.Setenv(InteractionQualityEnv, "low_cpu")
	if got := Default().EffectiveInteractionQuality(); got != InteractionQualityLowCPU {
		t.Fatalf("env quality = %v, want low_cpu", got)
	}

	th := Default(WithInteractionQuality(InteractionQualityBalanced))
	if got := th.EffectiveInteractionQuality(); got != InteractionQualityBalanced {
		t.Fatalf("theme option quality = %v, want balanced", got)
	}

	th.SetInteractionQuality(InteractionQuality(99))
	if got := th.EffectiveInteractionQuality(); got != InteractionQualityFull {
		t.Fatalf("invalid explicit quality = %v, want full", got)
	}
}
