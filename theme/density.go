package theme

type DensityLevel int

const (
	DensityDefault DensityLevel = iota
	DensityCompact
)

type DensityScale struct {
	Level DensityLevel
}

func Density(level DensityLevel) DensityScale {
	return DensityScale{Level: level}.Normalize()
}

func DefaultDensityScale() DensityScale {
	return Density(DensityDefault)
}

func CompactDensityScale() DensityScale {
	return Density(DensityCompact)
}

func (d DensityScale) Normalize() DensityScale {
	switch d.Level {
	case DensityCompact:
		return d
	default:
		return DensityScale{Level: DensityDefault}
	}
}

func (d DensityScale) IsCompact() bool {
	return d.Normalize().Level == DensityCompact
}

func (d DensityScale) ComponentHeight(defaultHeight, compactHeight float32) float32 {
	if d.Normalize().IsCompact() && compactHeight > 0 {
		return compactHeight
	}
	return defaultHeight
}

func (d DensityScale) Metric(defaultValue, compactValue float32) float32 {
	if d.Normalize().IsCompact() {
		return compactValue
	}
	return defaultValue
}

func (t *Theme) SetDensity(density DensityScale) {
	if t == nil {
		return
	}
	t.Density = density.Normalize()
}
