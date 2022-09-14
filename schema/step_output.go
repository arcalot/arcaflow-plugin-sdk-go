package schema

// StepOutputSchema holds the possible outputs of a step and the metadata information related to these outputs.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use StepOutputType.
type StepOutputSchema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] interface {
	Schema() S
	Display() *DisplayValue
	Error() bool
}

// NewStepOutputSchema defines a new output for a step.
func NewStepOutputSchema(
	schema ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	display *DisplayValue,
	error bool,
) StepOutputSchema[
	PropertySchema,
	ObjectSchema[PropertySchema],
	ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
] {
	return &stepOutputSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	]{
		schema,
		display,
		error,
	}
}

type stepOutputSchema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] struct {
	schema  S
	display *DisplayValue
	error   bool
}

func (s stepOutputSchema[P, O, S]) Schema() S {
	return s.schema
}

func (s stepOutputSchema[P, O, S]) Display() *DisplayValue {
	return s.display
}

func (s stepOutputSchema[P, O, S]) Error() bool {
	return s.error
}
