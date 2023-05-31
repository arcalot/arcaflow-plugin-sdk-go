package schema_test

import (
	"context"
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
	outputID, outputData, err := schemaTestSchema.Call(ctx, "hello", data)
	assertNoError(t, err)
	assertEqual(t, outputID, "success")
	typedData := outputData.(map[string]any)
	assertEqual(t, typedData["message"].(string), "Hello, Arca Lot!")
}
