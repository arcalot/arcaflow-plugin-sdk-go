package schema

// StepSchema holds the definition for a single step, it's input and output definitions.
type StepSchema interface {
	ID() string
	Input() ScopeSchema
	Outputs() map[string]StepOutputSchema
	Display() *DisplayValue
}

// NewStepSchema defines a new step.
func NewStepSchema(
	id string,
	input ScopeSchema,
	outputs map[string]StepOutputSchema,
	display *DisplayValue,
) StepSchema {
	return &stepSchema{
		id,
		input,
		outputs,
		display,
	}
}

type stepSchema struct {
	IDValue      string                      `json:"id"`
	InputValue   ScopeSchema                 `json:"input"`
	OutputsValue map[string]StepOutputSchema `json:"outputs"`
	DisplayValue *DisplayValue               `json:"display,omitempty"`
}

func (s stepSchema) ID() string {
	return s.IDValue
}

func (s stepSchema) Input() ScopeSchema {
	return s.InputValue
}

func (s stepSchema) Outputs() map[string]StepOutputSchema {
	return s.OutputsValue
}

func (s stepSchema) Display() *DisplayValue {
	return s.DisplayValue
}
