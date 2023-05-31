package schema

import (
	"context"
	"fmt"
)

// Step holds the definition for a single step, it's input and output definitions.
type Step interface {
	ID() string
	Input() Scope
	Outputs() map[string]*StepOutputSchema
	Display() Display
}

// CallableStep is a step that can be directly called.
type CallableStep interface {
	Step
	ToStepSchema() *StepSchema
	Call(ctx context.Context, data any) (outputID string, outputData any, err error)
}

// NewStepSchema defines a new step.
func NewStepSchema(
	id string,
	input Scope,
	outputs map[string]*StepOutputSchema,
	display Display,
) *StepSchema {
	return &StepSchema{
		id,
		input,
		outputs,
		display,
	}
}

// StepSchema describes a single step in a schema to execute one task. It has a fixed input and one or more outputs,
// denominated by a string output ID.
type StepSchema struct {
	IDValue      string                       `json:"id"`
	InputValue   Scope                        `json:"input"`
	OutputsValue map[string]*StepOutputSchema `json:"outputs"`
	DisplayValue Display                      `json:"display"`
}

func (s StepSchema) ID() string {
	return s.IDValue
}

func (s StepSchema) Input() Scope {
	return s.InputValue
}

func (s StepSchema) Outputs() map[string]*StepOutputSchema {
	return s.OutputsValue
}

func (s StepSchema) Display() Display {
	return s.DisplayValue
}

// NewCallableStep creates a callable step definition.
func NewCallableStep[InputType any](
	id string,
	input *ScopeSchema,
	outputs map[string]*StepOutputSchema,
	display Display,
	handler func(context.Context, InputType) (string, any),
) CallableStep {
	return &CallableStepSchema[InputType]{
		IDValue:      id,
		InputValue:   input,
		OutputsValue: outputs,
		DisplayValue: display,
		handler:      handler,
	}
}

// CallableStepSchema is a step that can be directly called and is typed to a specific input type.
type CallableStepSchema[InputType any] struct {
	IDValue      string                       `json:"id"`
	InputValue   *ScopeSchema                 `json:"input"`
	OutputsValue map[string]*StepOutputSchema `json:"outputs"`
	DisplayValue Display                      `json:"display"`
	handler      func(context.Context, InputType) (string, any)
}

func (s CallableStepSchema[InputType]) ID() string {
	return s.IDValue
}

func (s CallableStepSchema[InputType]) Input() Scope {
	return s.InputValue
}

func (s CallableStepSchema[InputType]) Outputs() map[string]*StepOutputSchema {
	return s.OutputsValue
}

func (s CallableStepSchema[InputType]) Display() Display {
	return s.DisplayValue
}

func (s CallableStepSchema[InputType]) ToStepSchema() *StepSchema {
	return &StepSchema{
		IDValue:      s.IDValue,
		InputValue:   s.InputValue,
		OutputsValue: s.OutputsValue,
		DisplayValue: s.DisplayValue,
	}
}

func (s CallableStepSchema[InputType]) Call(ctx context.Context, input any) (string, any, error) {
	if err := s.InputValue.Validate(input); err != nil {
		return "", nil, InvalidInputError{err}
	}

	outputID, outputData := s.handler(ctx, input.(InputType))
	output, ok := s.OutputsValue[outputID]
	if !ok {
		return "", nil, InvalidOutputError{
			fmt.Errorf("undeclared output ID: %s", outputID),
		}
	}
	return outputID, outputData, output.Validate(outputData)
}
