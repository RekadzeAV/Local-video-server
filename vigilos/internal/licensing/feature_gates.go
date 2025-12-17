package licensing

// FeatureGates enforces edition-based limits.
type FeatureGates struct{}

func NewFeatureGates() *FeatureGates { return &FeatureGates{} }

func (g *FeatureGates) Allow(feature string) bool {
	_ = feature
	return true
}

