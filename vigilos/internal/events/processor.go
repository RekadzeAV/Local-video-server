package events

// Processor consumes incoming events and normalizes them.
type Processor struct{}

func NewProcessor() *Processor { return &Processor{} }

func (p *Processor) Process(raw interface{}) (interface{}, error) {
	_ = raw
	return nil, nil
}

