package schema

// StepSchema holds the definition for a single step, it's input and output definitions.
type StepSchema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] interface {
	ID() string
	Input() S
	Outputs() map[string]StepOutputSchema[P, O, S]
	Display() *DisplayValue
}

// NewStepSchema defines a new step.
func NewStepSchema(
	id string,
	input ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	outputs map[string]StepOutputSchema[
		PropertySchema,
		ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	],
	display *DisplayValue,
) StepSchema[
	PropertySchema,
	ObjectSchema[PropertySchema],
	ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
] {
	return &stepSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	]{
		id,
		input,
		outputs,
		display,
	}
}

type stepSchema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] struct {
	IDValue      string                               `json:"id"`
	InputValue   S                                    `json:"input"`
	OutputsValue map[string]StepOutputSchema[P, O, S] `json:"outputs"`
	DisplayValue *DisplayValue                        `json:"display,omitempty"`
}

func (s stepSchema[P, O, S]) ID() string {
	return s.IDValue
}

func (s stepSchema[P, O, S]) Input() S {
	return s.InputValue
}

func (s stepSchema[P, O, S]) Outputs() map[string]StepOutputSchema[P, O, S] {
	return s.OutputsValue
}

func (s stepSchema[P, O, S]) Display() *DisplayValue {
	return s.DisplayValue
}
