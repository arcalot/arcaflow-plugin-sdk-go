package schema_test

import (
	"testing"

	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
)

var schemaTestSchema = schema.NewCallableSchema(
	testStepSchema,
)

func TestSchemaCall(t *testing.T) {
	data := map[string]any{
		"name": "Arca Lot",
	}

	outputID, outputData, err := schemaTestSchema.Call("hello", data)
	assert.NoError(t, err)
	assert.Equals(t, outputID, "success")
	typedData := outputData.(map[string]any)
	assert.Equals(t, typedData["message"].(string), "Hello, Arca Lot!")
}
