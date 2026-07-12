package widget

import "testing"

func TestDeprecatedSelectOptionsRemainCompatibilityNoOps(t *testing.T) {
	cfg := selectConfig[string]{
		placeholder: "Choose one",
		maxHeight:   240,
		variant:     selectVariantFilled,
	}
	SelectSearchable[string](true)(&cfg)
	SelectQuick[string](true)(&cfg)
	SelectTypeaheadDelay[string](1)(&cfg)

	if cfg.placeholder != "Choose one" || cfg.maxHeight != 240 || cfg.variant != selectVariantFilled {
		t.Fatalf("compatibility options changed select config: %#v", cfg)
	}
}
