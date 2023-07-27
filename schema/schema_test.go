package schema_test

import (
	"context"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var schemaTestSchema = schema.NewCallableSchema(
	testStepSchema,
)

func TestSchemaCall(t *testing.T) {
	data := map[string]any{
		"name": "Arca Lot",
	}

	ctx := context.Background()
	outputID, outputData, err := schemaTestSchema.CallStep(ctx, "hello", data)
	assert.NoError(t, err)
	assert.Equals(t, outputID, "success")
	typedData := outputData.(map[string]any)
	assert.Equals(t, typedData["message"].(string), "Hello, Arca Lot!")
}
