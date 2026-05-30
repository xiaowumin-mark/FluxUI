package theme

type DensityLevel int

const (
	DensityDefault DensityLevel = iota
	DensityCompact
)

type DensityScale struct {
	Level DensityLevel
}

func DefaultDensityScale() DensityScale {
	return DensityScale{Level: DensityDefault}
}

func CompactDensityScale() DensityScale {
	return DensityScale{Level: DensityCompact}
}

func (d DensityScale) IsCompact() bool {
	return d.Level == DensityCompact
}

func (d DensityScale) ComponentHeight(defaultHeight, compactHeight float32) float32 {
	if d.IsCompact() && compactHeight > 0 {
		return compactHeight
	}
	return defaultHeight
}

func (d DensityScale) Metric(defaultValue, compactValue float32) float32 {
	if d.IsCompact() {
		return compactValue
	}
	return defaultValue
}
