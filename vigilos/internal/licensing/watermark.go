package licensing

// Watermark placeholder for runtime watermark generation.
type Watermark struct{}

func NewWatermark() *Watermark { return &Watermark{} }

func (w *Watermark) Apply(input []byte) ([]byte, error) {
	_ = input
	return input, nil
}

