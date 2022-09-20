package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var schemaTestSchema = schema.NewSchemaType(
	map[string]schema.StepType[any]{
		"hello": testStepSchema.Any(),
	},
)

func TestSchemaCall(t *testing.T) {
	data := map[string]any{
		"name": "Arca Lot",
	}

	outputID, outputData, err := schemaTestSchema.Call("hello", data)
	assertNoError(t, err)
	assertEqual(t, outputID, "success")
	typedData := outputData.(map[string]any)
	assertEqual(t, typedData["message"].(string), "Hello, Arca Lot!")
}
