package schema_test

import (
	"context"
	"fmt"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type stepTestInputData struct {
	Name string `json:"name"`
}

type stepTestSuccessOutput struct {
	Message string `json:"message"`
}

type stepTestErrorOutput struct {
	Error string `json:"message"`
}

var testStepSchema = schema.NewCallableStep(
	"hello",
	schema.NewScopeSchema(
		schema.NewStructMappedObjectSchema[stepTestInputData](
			"input",
			map[string]*schema.PropertySchema{
				"name": schema.NewPropertySchema(
					schema.NewStringSchema(schema.IntPointer(1), nil, nil),
					nil,
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		),
	),
	map[string]*schema.StepOutputSchema{
		"success": schema.NewStepOutputSchema(
			schema.NewScopeSchema(
				schema.NewStructMappedObjectSchema[stepTestSuccessOutput](
					"output",
					map[string]*schema.PropertySchema{
						"message": schema.NewPropertySchema(
							schema.NewStringSchema(schema.IntPointer(1), nil, nil),
							nil,
							true,
							nil,
							nil,
							nil,
							nil,
							nil,
						),
					},
				),
			),
			nil,
			false,
		),
		"error": schema.NewStepOutputSchema(
			schema.NewScopeSchema(
				schema.NewStructMappedObjectSchema[stepTestErrorOutput](
					"output",
					map[string]*schema.PropertySchema{
						"message": schema.NewPropertySchema(
							schema.NewStringSchema(schema.IntPointer(1), nil, nil),
							nil,
							true,
							nil,
							nil,
							nil,
							nil,
							nil,
						),
					},
				),
			),
			nil,
			true,
		),
	},
	nil,
	stepTestHandler,
)

func stepTestHandler(_ context.Context, input stepTestInputData) (string, any) {
	return "success", stepTestSuccessOutput{
		Message: fmt.Sprintf("Hello, %s!", input.Name),
	}
}

func TestStepExecution(t *testing.T) {
	ctx := context.Background()
	outputID, outputData, err := testStepSchema.Call(ctx, t.Name(), stepTestInputData{Name: "Arca Lot"})
	assert.NoError(t, err)
	assert.Equals(t, outputID, "success")
	assert.Equals(t, outputData.(stepTestSuccessOutput).Message, "Hello, Arca Lot!")
}
