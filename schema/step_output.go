package schema

import (
	"reflect"
)

// StepOutput holds the possible outputs of a step and the metadata information related to these outputs.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use StepOutputType.
type StepOutput interface {
	Serializable

	Schema() Scope
	Display() *DisplayValue
	Error() bool
}

// NewStepOutputSchema defines a new output for a step.
func NewStepOutputSchema(
	schema Scope,
	display *DisplayValue,
	error bool,
) *StepOutputSchema {
	return &StepOutputSchema{
		schema,
		display,
		error,
	}
}

type StepOutputSchema struct {
	SchemaValue  Scope         `json:"schema"`
	DisplayValue *DisplayValue `json:"display"`
	ErrorValue   bool          `json:"error"`
}

func (s StepOutputSchema) ReflectedType() reflect.Type {
	return s.SchemaValue.ReflectedType()
}

func (s StepOutputSchema) Unserialize(data any) (any, error) {
	return s.SchemaValue.Unserialize(data)
}

func (s StepOutputSchema) Validate(data any) error {
	return s.SchemaValue.Validate(data)
}

func (s StepOutputSchema) Serialize(data any) (any, error) {
	return s.SchemaValue.Serialize(data)
}

func (s StepOutputSchema) ApplyScope(scope Scope) {
	s.SchemaValue.ApplyScope(scope)
}

func (s StepOutputSchema) Schema() Scope {
	return s.SchemaValue
}

func (s StepOutputSchema) Display() *DisplayValue {
	return s.DisplayValue
}

func (s StepOutputSchema) Error() bool {
	return s.ErrorValue
}
