package events

// Engine evaluates rules and triggers actions.
type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Evaluate(event interface{}) error {
	_ = event
	return nil
}

