package schema

import "fmt"

// StepSchema holds the definition for a single step, it's input and output definitions.
type StepSchema[
	P PropertySchema,
	O ObjectSchema[P],
	InputSchema ScopeSchema[P, O],
	OSC ScopeSchema[P, O],
	OutputSchema StepOutputSchema[P, O, OSC],
] interface {
	ID() string
	Input() InputSchema
	Outputs() map[string]OutputSchema
	Display() *DisplayValue
}

// NewStepSchema defines a new step.
func NewStepSchema(
	id string,
	input ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	outputs map[string]StepOutputSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	],
	display *DisplayValue,
) StepSchema[
	PropertySchema,
	ObjectSchema[PropertySchema],
	ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	StepOutputSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	],
] {
	return &stepSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
		StepOutputSchema[
			PropertySchema,
			ObjectSchema[PropertySchema],
			ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
		],
	]{
		id,
		input,
		outputs,
		display,
	}
}

type stepSchema[
	P PropertySchema,
	O ObjectSchema[P],
	InputScopeSchema ScopeSchema[P, O],
	OSC ScopeSchema[P, O],
	OutputSchema StepOutputSchema[P, O, OSC],
] struct {
	IDValue      string                  `json:"id"`
	InputValue   InputScopeSchema        `json:"input"`
	OutputsValue map[string]OutputSchema `json:"outputs"`
	DisplayValue *DisplayValue           `json:"display,omitempty"`
}

func (s stepSchema[P, O, InputScopeSchema, OSC, OutputScopeSchema]) ID() string {
	return s.IDValue
}

func (s stepSchema[P, O, InputScopeSchema, OSC, OutputScopeSchema]) Input() InputScopeSchema {
	return s.InputValue
}

func (s stepSchema[
	P,
	O,
	IS,
	OSC,
	OS,
]) Outputs() map[string]OS {
	return s.OutputsValue
}

func (s stepSchema[P, O, InputScopeSchema, OSC, OutputScopeSchema]) Display() *DisplayValue {
	return s.DisplayValue
}

// StepType defines a step that can be called with a type input.
type StepType[InputType any] interface {
	StepSchema[PropertyType, ObjectType[any], ScopeType[InputType], ScopeType[any], StepOutputType[any]]

	Call(input InputType) (outputID string, outputData any)
	Any() StepType[any]
}

// NewStepType creates a callable step definition.
func NewStepType[InputType any](
	id string,
	input ScopeType[InputType],
	outputs map[string]StepOutputType[any],
	display *DisplayValue,
	handler func(InputType) (string, any),
) StepType[InputType] {
	return &stepType[InputType]{
		stepSchema[PropertyType, ObjectType[any], ScopeType[InputType], ScopeType[any], StepOutputType[any]]{
			IDValue:      id,
			InputValue:   input,
			OutputsValue: outputs,
			DisplayValue: display,
		},
		handler,
	}
}

type stepType[InputType any] struct {
	stepSchema[PropertyType, ObjectType[any], ScopeType[InputType], ScopeType[any], StepOutputType[any]] `json:",inline"`

	handler func(InputType) (string, any)
}

func (s stepType[InputType]) Any() StepType[any] {
	return &anonymousStepType[InputType]{
		s,
	}
}

func (s stepType[InputType]) Input() ScopeType[InputType] {
	return s.InputValue
}

func (s stepType[InputType]) Outputs() map[string]StepOutputType[any] {
	return s.OutputsValue
}

func (s stepType[InputType]) Call(input InputType) (string, any) {
	return s.handler(input)
}

type anonymousStepType[T any] struct {
	stepType[T]
}

func (a anonymousStepType[T]) Input() ScopeType[any] {
	return a.InputValue.Any()
}

func (a anonymousStepType[T]) Outputs() map[string]StepOutputType[any] {
	return a.OutputsValue
}

func (a anonymousStepType[T]) Call(input any) (outputID string, outputData any) {
	typedInput, ok := input.(T)
	if !ok {
		var defaultValue T
		panic(BadArgumentError{
			fmt.Sprintf("Incorrect input argument type received: %T, expected: %T", input, defaultValue),
			nil,
		})
	}
	return a.stepType.Call(typedInput)
}
