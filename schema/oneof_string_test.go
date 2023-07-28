// golangci-lint does not accurately detect changes in type parameters.
//
//nolint:dupl
package schema_test

import (
	"encoding/json"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var oneOfStringTestObjectAProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"C": schema.NewRefSchema("C", nil),
			},
			"_type",
		),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

// Differs in that the discriminator field is different.
var oneOfStringTestObjectAbProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"C": schema.NewRefSchema("C", nil),
			},
			"_difftype",
		),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

// Differs in that a key doesn't match.
var oneOfStringTestObjectAcProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"D": schema.NewRefSchema("C", nil),
			},
			"_type",
		),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

// Differs in that a oneof schema doesn't match.
var oneOfStringTestObjectAdProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"C": schema.NewRefSchema("D", nil),
			},
			"_type",
		),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfStringTestObjectASchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfStringTestObjectAProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
)

var oneOfStringTestObjectAbSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfStringTestObjectAbProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
	oneOfTestDSchema,
)
var oneOfStringTestObjectAcSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfStringTestObjectAcProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
	oneOfTestDSchema,
)
var oneOfStringTestObjectAdSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfStringTestObjectAdProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
	oneOfTestDSchema,
)

var oneOfStringTestObjectAType = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[oneOfTestObjectA](
		"A",
		oneOfStringTestObjectAProperties,
	),
	oneOfTestBMappedSchema,
	oneOfTestCMappedSchema,
)

func TestOneOfStringUnserialization(t *testing.T) {
	data := `{
	"s": {
		"_type": "B",
		"message": "Hello world!"
	}
}`
	var input any
	assert.NoError(t, json.Unmarshal([]byte(data), &input))
	unserializedData, err := oneOfStringTestObjectAType.Unserialize(input)
	assert.NoError(t, err)
	assert.Equals(t, unserializedData.(oneOfTestObjectA).S.(oneOfTestObjectB).Message, "Hello world!")
}

func TestOneOfStringCompatibilityValidation(t *testing.T) {
	// The ones with NoError are matching schemas
	// All others have one thing that differs, so should error.
	assert.NoError(t, oneOfStringTestObjectASchema.ValidateCompatibility(oneOfStringTestObjectASchema))
	assert.Error(t, oneOfStringTestObjectASchema.ValidateCompatibility(oneOfStringTestObjectAbSchema))
	assert.Error(t, oneOfStringTestObjectASchema.ValidateCompatibility(oneOfStringTestObjectAcSchema))
	assert.Error(t, oneOfStringTestObjectASchema.ValidateCompatibility(oneOfStringTestObjectAdSchema))
	assert.Error(t, oneOfStringTestObjectAbSchema.ValidateCompatibility(oneOfStringTestObjectASchema))
	assert.NoError(t, oneOfStringTestObjectAbSchema.ValidateCompatibility(oneOfStringTestObjectAbSchema))
	assert.Error(t, oneOfStringTestObjectAbSchema.ValidateCompatibility(oneOfStringTestObjectAcSchema))
	assert.Error(t, oneOfStringTestObjectAbSchema.ValidateCompatibility(oneOfStringTestObjectAdSchema))
	assert.Error(t, oneOfStringTestObjectAcSchema.ValidateCompatibility(oneOfStringTestObjectASchema))
	assert.Error(t, oneOfStringTestObjectAcSchema.ValidateCompatibility(oneOfStringTestObjectAbSchema))
	assert.NoError(t, oneOfStringTestObjectAcSchema.ValidateCompatibility(oneOfStringTestObjectAcSchema))
	assert.Error(t, oneOfStringTestObjectAcSchema.ValidateCompatibility(oneOfStringTestObjectAdSchema))
	assert.Error(t, oneOfStringTestObjectAdSchema.ValidateCompatibility(oneOfStringTestObjectASchema))
	assert.Error(t, oneOfStringTestObjectAdSchema.ValidateCompatibility(oneOfStringTestObjectAbSchema))
	assert.Error(t, oneOfStringTestObjectAdSchema.ValidateCompatibility(oneOfStringTestObjectAcSchema))
	assert.NoError(t, oneOfStringTestObjectAdSchema.ValidateCompatibility(oneOfStringTestObjectAdSchema))
}

func TestOneOfStringCompatibilityMapValidation(t *testing.T) {
	validWithObjectB := map[string]any{
		"s": map[string]any{
			"_type": "B",
			// object B fields
			"message": "test",
		},
	}
	validWithObjectC := map[string]any{
		"s": map[string]any{
			"_type": "C",
			// object B fields
			"m": "test",
		},
	}
	invalidDiscriminatorType := map[string]any{
		"s": map[string]any{
			"_type": 1,
			// object B fields
			"message": "test",
		},
	}
	invalidDiscriminator := map[string]any{
		"s": map[string]any{
			"wrongdiscriminator": "B",
			// object B fields
			"message": "test",
		},
	}

	combinedMapAndSchema := map[string]any{
		"s": map[string]any{
			"_type": "B",
			// object B fields, but with a schema instead of a value
			"message": schema.NewStringSchema(nil, nil, nil),
		},
	}
	combinedMapAndInvalidSchema := map[string]any{
		"s": map[string]any{
			"_type": "B",
			// object B fields, but with a schema instead of a value
			"message": schema.NewIntSchema(nil, nil, nil),
		},
	}
	assert.NoError(t, oneOfStringTestObjectASchema.ValidateCompatibility(validWithObjectB))
	assert.NoError(t, oneOfStringTestObjectASchema.ValidateCompatibility(validWithObjectC))
	assert.Error(t, oneOfStringTestObjectASchema.ValidateCompatibility(invalidDiscriminatorType))
	assert.Error(t, oneOfStringTestObjectASchema.ValidateCompatibility(invalidDiscriminator))
	assert.NoError(t, oneOfStringTestObjectASchema.ValidateCompatibility(combinedMapAndSchema))
	assert.Error(t, oneOfStringTestObjectASchema.ValidateCompatibility(combinedMapAndInvalidSchema))
}
