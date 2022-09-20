package schema_test

import (
	"fmt"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type stepTestInputData struct {
	Name string `json:"name"`
}

type stepTestSuccessOutput struct {
	Message string `json:"message"`
}

// golangci-lint does not detect types used in type parameters.
//nolint:unused
type stepTestErrorOutput struct {
	Error string `json:"message"`
}

var testStepSchema = schema.NewStepType(
	"hello",
	schema.NewScopeType[stepTestInputData](
		map[string]schema.ObjectType[any]{
			"input": schema.NewObjectType[stepTestInputData](
				"input",
				map[string]schema.PropertyType{
					"name": schema.NewPropertyType[string](
						schema.NewStringType(schema.IntPointer(1), nil, nil),
						nil,
						true,
						nil,
						nil,
						nil,
						nil,
						nil,
					),
				},
			).Any(),
		},
		"input",
	),
	map[string]schema.StepOutputType[any]{
		"success": schema.NewStepOutputType[stepTestSuccessOutput](
			schema.NewScopeType[stepTestSuccessOutput](
				map[string]schema.ObjectType[any]{
					"output": schema.NewObjectType[stepTestSuccessOutput](
						"output",
						map[string]schema.PropertyType{
							"message": schema.NewPropertyType[string](
								schema.NewStringType(schema.IntPointer(1), nil, nil),
								nil,
								true,
								nil,
								nil,
								nil,
								nil,
								nil,
							),
						},
					).Any(),
				},
				"output",
			),
			nil,
			false,
		).Any(),
		"error": schema.NewStepOutputType[stepTestErrorOutput](
			schema.NewScopeType[stepTestErrorOutput](
				map[string]schema.ObjectType[any]{
					"output": schema.NewObjectType[stepTestSuccessOutput](
						"output",
						map[string]schema.PropertyType{
							"message": schema.NewPropertyType[string](
								schema.NewStringType(schema.IntPointer(1), nil, nil),
								nil,
								true,
								nil,
								nil,
								nil,
								nil,
								nil,
							),
						},
					).Any(),
				},
				"output",
			),
			nil,
			true,
		).Any(),
	},
	nil,
	stepTestHandler,
)

func stepTestHandler(input stepTestInputData) (string, any) {
	return "success", stepTestSuccessOutput{
		Message: fmt.Sprintf("Hello, %s!", input.Name),
	}
}

func TestStepExecution(t *testing.T) {
	outputID, outputData := testStepSchema.Call(stepTestInputData{Name: "Arca Lot"})
	assertEqual(t, outputID, "success")
	assertEqual(t, outputData.(stepTestSuccessOutput).Message, "Hello, Arca Lot!")
}
