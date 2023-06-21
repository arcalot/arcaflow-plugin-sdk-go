package schema

import (
	"context"
)

// Signal holds the definition for a single signal. This is universal for emitted or received.
type Signal interface {
	ID() string
	DataSchema() Scope
	Display() Display
}

// CallableSignal is a signal that can be directly called.
type CallableSignal interface {
	Signal
	ToSignalSchema() *SignalSchema
	Call(ctx context.Context, data any) (err error)
}

// NewSignalSchema defines a new signal.
func NewSignalSchema(
	id string,
	dataSchema Scope,
	display Display,
) *SignalSchema {
	return &SignalSchema{
		id,
		dataSchema,
		display,
	}
}

// SignalSchema describes a single signal in a schema to execute one task. It has a fixed data input or output,
// which is either input or output depending on whether it's receiving or emitting the signal.
type SignalSchema struct {
	IDValue         string  `json:"id"`
	DataSchemaValue Scope   `json:"input"`
	DisplayValue    Display `json:"display"`
}

func (s SignalSchema) ID() string {
	return s.IDValue
}

func (s SignalSchema) DataSchema() Scope {
	return s.DataSchemaValue
}

func (s SignalSchema) Display() Display {
	return s.DisplayValue
}

// NewCallableSignal creates a callable signal definition.
func NewCallableSignal[InputType any](
	id string,
	input *ScopeSchema,
	display Display,
	handler func(context.Context, InputType),
) CallableSignal {
	return &CallableSignalSchema[InputType]{
		IDValue:      id,
		InputValue:   input,
		DisplayValue: display,
		handler:      handler,
	}
}

// CallableSignalSchema is a signal that can be directly called and is typed to a specific input type.
// This is an input-only representation of the signal.
type CallableSignalSchema[InputType any] struct {
	IDValue      string       `json:"id"`
	InputValue   *ScopeSchema `json:"data_input_schema"`
	DisplayValue Display      `json:"display"`
	handler      func(context.Context, InputType)
}

func (s CallableSignalSchema[InputType]) ID() string {
	return s.IDValue
}

func (s CallableSignalSchema[InputType]) DataSchema() Scope {
	return s.InputValue
}

func (s CallableSignalSchema[InputType]) Display() Display {
	return s.DisplayValue
}

func (s CallableSignalSchema[InputType]) ToSignalSchema() *SignalSchema {
	return &SignalSchema{
		IDValue:         s.IDValue,
		DataSchemaValue: s.InputValue,
		DisplayValue:    s.DisplayValue,
	}
}

func (s CallableSignalSchema[InputType]) Call(ctx context.Context, input any) error {
	if err := s.InputValue.Validate(input); err != nil {
		return InvalidInputError{err}
	}

	s.handler(ctx, input.(InputType))
	return nil
}
