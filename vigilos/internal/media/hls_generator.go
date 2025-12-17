package media

// HLSGenerator creates HLS segments and playlists.
type HLSGenerator struct{}

func NewHLSGenerator() *HLSGenerator {
	return &HLSGenerator{}
}

func (g *HLSGenerator) GenerateSegment() error {
	// TODO: segment size 2s, H.264/H.265 profiles.
	return nil
}

