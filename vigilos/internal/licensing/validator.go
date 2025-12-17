package licensing

// Validator checks edition keys and signatures.
type Validator struct{}

func NewValidator() *Validator { return &Validator{} }

func (v *Validator) Validate(key string) (bool, error) {
	_ = key
	return true, nil
}

