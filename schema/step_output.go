package schema

// StepOutputSchema holds the possible outputs of a step and the metadata information related to these outputs.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use StepOutputType.
type StepOutputSchema interface {
	Schema() ScopeSchema
	Display() *DisplayValue
	Error() bool
}

// NewStepOutputSchema defines a new output for a step.
func NewStepOutputSchema(schema ScopeSchema, display *DisplayValue, error bool) StepOutputSchema {
	return &stepOutputSchema{
		schema,
		display,
		error,
	}
}

type stepOutputSchema struct {
	schema  ScopeSchema
	display *DisplayValue
	error   bool
}

func (s stepOutputSchema) Schema() ScopeSchema {
	return s.schema
}

func (s stepOutputSchema) Display() *DisplayValue {
	return s.display
}

func (s stepOutputSchema) Error() bool {
	return s.error
}
