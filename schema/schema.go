package schema

import "fmt"

// Schema is a collection of steps supported by a plugin.
type Schema[P PropertySchema, O ObjectSchema[P], IS ScopeSchema[P, O], OSC ScopeSchema[P, O], OS StepOutputSchema[P, O, OSC], ST StepSchema[P, O, IS, OSC, OS]] interface {
	Steps() map[string]ST
}

// NewSchema builds a new schema with the specified steps.
func NewSchema(
	steps map[string]StepSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
		StepOutputSchema[PropertySchema, ObjectSchema[PropertySchema],
			ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]]],
	],
) Schema[PropertySchema, ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]], StepOutputSchema[PropertySchema, ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]]], StepSchema[PropertySchema, ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]], StepOutputSchema[PropertySchema, ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]]]]] {
	return &abstractSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
		StepOutputSchema[PropertySchema, ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]]],
		StepSchema[
			PropertySchema,
			ObjectSchema[PropertySchema],
			ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
			ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
			StepOutputSchema[PropertySchema, ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]]],
		],
	]{
		steps,
	}
}

type abstractSchema[P PropertySchema, O ObjectSchema[P], IS ScopeSchema[P, O], OSC ScopeSchema[P, O], OS StepOutputSchema[P, O, OSC], ST StepSchema[P, O, IS, OSC, OS]] struct {
	StepsValue map[string]ST `json:"steps"`
}

//nolint:unused
type schema struct {
	abstractSchema[*propertySchema, *objectSchema, *scopeSchema, *scopeSchema, *stepOutputSchema, *stepSchema] `json:",inline"`
}

func (s abstractSchema[P, O, IS, OSC, OS, ST]) Steps() map[string]ST {
	return s.StepsValue
}

// SchemaType defines a complete callable schema.
//
// Disable linting, this is intentional:
//nolint:revive
//goland:noinspection GoNameStartsWithPackageName
type SchemaType interface {
	Schema[PropertyType, ObjectType[any], ScopeType[any], ScopeType[any], StepOutputType[any], StepType[any]]

	Call(stepID string, serializedInputData any) (outputID string, serializedOutputData any, err error)
}

// NewSchemaType defines a callable schema.
func NewSchemaType(
	steps map[string]StepType[any],
) SchemaType {
	return &schemaType{
		abstractSchema: abstractSchema[PropertyType, ObjectType[any], ScopeType[any], ScopeType[any], StepOutputType[any], StepType[any]]{
			StepsValue: steps,
		},
	}
}

type schemaType struct {
	abstractSchema[PropertyType, ObjectType[any], ScopeType[any], ScopeType[any], StepOutputType[any], StepType[any]] `json:",inline"`
}

func (s schemaType) Call(
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
	outputID, unserializedOutput := step.Call(unserializedInputData)
	output := step.Outputs()[outputID]
	serializedData, err := output.Schema().Serialize(unserializedOutput)
	if err != nil {
		return "", nil, InvalidOutputError{err}
	}
	return outputID, serializedData, nil
}
