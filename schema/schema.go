package schema

import (
	"context"
	"fmt"
)

// Schema is a collection of steps supported by a plugin.
type Schema[S Step] interface {
	Steps() map[string]S

	SelfSerialize() (any, error)
}

// NewSchema builds a new schema with the specified steps.
func NewSchema(
	steps map[string]*StepSchema,
) Schema[Step] {
	return &SchemaSchema{
		steps,
	}
}

type SchemaSchema struct {
	StepsValue map[string]*StepSchema `json:"steps"`
}

func (s SchemaSchema) SelfSerialize() (any, error) {
	steps := make(map[string]*StepSchema, len(s.StepsValue))

	for id, step := range s.StepsValue {
		steps[id] = step
	}

	return schemaSchema.Serialize(&SchemaSchema{
		steps,
	})
}

func (s SchemaSchema) Steps() map[string]Step {
	result := make(map[string]Step, len(s.StepsValue))
	for k, v := range s.StepsValue {
		result[k] = v
	}
	return result
}

func (s SchemaSchema) applyScope() {
	for _, step := range s.StepsValue {
		// We can apply an empty scope because the scope does not need another scope.
		step.InputValue.ApplyScope(nil)
		for _, output := range step.OutputsValue {
			output.ApplyScope(nil)
		}
	}
}

func NewCallableSchema(
	steps ...CallableStep,
) *CallableSchema {

	stepMap := make(map[string]CallableStep, len(steps))
	for _, s := range steps {
		stepMap[s.ID()] = s
	}

	return &CallableSchema{
		stepMap,
	}
}

type CallableSchema struct {
	StepsValue map[string]CallableStep `json:"steps"`
}

func (s CallableSchema) Call(
	ctx context.Context,
	stepID string,
	serializedInputData any,
) (
	outputID string,
	serializedOutputData any,
	err error,
) {
	step, ok := s.StepsValue[stepID]
	if !ok {
		return "", nil, BadArgumentError{
			Message: fmt.Sprintf("Invalid step called: %s", stepID),
		}
	}
	unserializedInputData, err := step.Input().Unserialize(serializedInputData)
	if err != nil {
		return "", nil, InvalidInputError{err}
	}
	outputID, unserializedOutput, err := step.Call(ctx, unserializedInputData)
	if err != nil {
		return outputID, nil, err
	}
	output := step.Outputs()[outputID]
	serializedData, err := output.Schema().Serialize(unserializedOutput)
	if err != nil {
		return "", nil, InvalidOutputError{err}
	}
	return outputID, serializedData, nil
}

func (s CallableSchema) SelfSerialize() (any, error) {
	steps := make(map[string]*StepSchema, len(s.StepsValue))

	for id, step := range s.StepsValue {
		steps[id] = step.ToStepSchema()
	}

	return schemaSchema.Serialize(&SchemaSchema{
		steps,
	})
}
