package schema

// StepOutputSchema holds the possible outputs of a step and the metadata information related to these outputs.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use StepOutputType.
type StepOutputSchema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] interface {
	Schema() S
	Display() DisplayValue
	Error() bool
}

// NewStepOutputSchema defines a new output for a step.
func NewStepOutputSchema(
	schema ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	display DisplayValue,
	error bool,
) StepOutputSchema[
	PropertySchema,
	ObjectSchema[PropertySchema],
	ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
] {
	return &abstractStepOutputSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	]{
		schema,
		display,
		error,
	}
}

type abstractStepOutputSchema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] struct {
	SchemaValue  S            `json:"schema"`
	DisplayValue DisplayValue `json:"display"`
	ErrorValue   bool         `json:"error"`
}

//nolint:unused
type stepOutputSchema struct {
	abstractStepOutputSchema[*propertySchema, *objectSchema, *scopeSchema] `json:",inline"`
}

func (s abstractStepOutputSchema[P, O, S]) Schema() S {
	return s.SchemaValue
}

func (s abstractStepOutputSchema[P, O, S]) Display() DisplayValue {
	return s.DisplayValue
}

func (s abstractStepOutputSchema[P, O, S]) Error() bool {
	return s.ErrorValue
}

// NewStepOutputType defines a typed step output.
func NewStepOutputType[T any](
	schema ScopeType[T],
	display DisplayValue,
	error bool,
) StepOutputType[T] {
	return &stepOutputType[T]{
		abstractStepOutputSchema[PropertyType, ObjectType[any], ScopeType[T]]{
			SchemaValue:  schema,
			DisplayValue: display,
			ErrorValue:   error,
		},
	}
}

// StepOutputType defines a typed step output.
type StepOutputType[T any] interface {
	StepOutputSchema[PropertyType, ObjectType[any], ScopeType[T]]

	Any() StepOutputType[any]
}

type stepOutputType[T any] struct {
	abstractStepOutputSchema[PropertyType, ObjectType[any], ScopeType[T]]
}

func (s stepOutputType[T]) Schema() ScopeType[T] {
	return s.SchemaValue
}

func (s stepOutputType[T]) Any() StepOutputType[any] {
	return &stepOutputType[any]{
		abstractStepOutputSchema[PropertyType, ObjectType[any], ScopeType[any]]{
			SchemaValue:  s.SchemaValue.Any(),
			DisplayValue: s.DisplayValue,
			ErrorValue:   s.ErrorValue,
		},
	}
}
