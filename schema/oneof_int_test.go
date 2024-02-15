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

var oneOfIntTestObjectAProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfIntSchema[any](
			map[int64]schema.Object{
				1: schema.NewRefSchema("B", nil),
				2: schema.NewRefSchema("C", nil),
			},
			"_type",
			false,
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
var oneOfIntTestObjectAbProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfIntSchema[any](
			map[int64]schema.Object{
				1: schema.NewRefSchema("B", nil),
				2: schema.NewRefSchema("C", nil),
			},
			"_difftype",
			false,
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
var oneOfIntTestObjectAcProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfIntSchema[any](
			map[int64]schema.Object{
				1: schema.NewRefSchema("B", nil),
				3: schema.NewRefSchema("C", nil),
			},
			"_type",
			false,
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
var oneOfIntTestObjectAdProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfIntSchema[any](
			map[int64]schema.Object{
				1: schema.NewRefSchema("B", nil),
				2: schema.NewRefSchema("D", nil),
			},
			"_type",
			false,
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

var oneOfIntTestObjectASchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfIntTestObjectAProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
	oneOfTestDSchema,
)

var oneOfIntTestObjectAbSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfIntTestObjectAbProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
	oneOfTestDSchema,
)
var oneOfIntTestObjectAcSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfIntTestObjectAcProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
	oneOfTestDSchema,
)
var oneOfIntTestObjectAdSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfIntTestObjectAdProperties,
	),
	oneOfTestBSchema,
	oneOfTestCSchema,
	oneOfTestDSchema,
)

var oneOfIntTestObjectAType = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[oneOfTestObjectA](
		"A",
		oneOfIntTestObjectAProperties,
	),
	oneOfTestBMappedSchema,
	oneOfTestCMappedSchema,
)

func TestOneOfIntUnserialization(t *testing.T) {
	data := `{
	"s": {
		"_type": 1,
		"message": "Hello world!"
	}
}`
	var input any
	assert.NoError(t, json.Unmarshal([]byte(data), &input))
	unserializedData, err := oneOfIntTestObjectAType.Unserialize(input)
	assert.NoError(t, err)
	assert.Equals(t, unserializedData.(oneOfTestObjectA).S.(oneOfTestObjectB).Message, "Hello world!")
}

func TestOneOfIntCompatibilityValidation(t *testing.T) {
	// The ones with NoError are matching schemas
	// All others have one thing that differs, so should error.
	assert.NoError(t, oneOfIntTestObjectASchema.ValidateCompatibility(oneOfIntTestObjectASchema))
	assert.Error(t, oneOfIntTestObjectASchema.ValidateCompatibility(oneOfIntTestObjectAbSchema))
	assert.Error(t, oneOfIntTestObjectASchema.ValidateCompatibility(oneOfIntTestObjectAcSchema))
	assert.Error(t, oneOfIntTestObjectASchema.ValidateCompatibility(oneOfIntTestObjectAdSchema))
	assert.Error(t, oneOfIntTestObjectAbSchema.ValidateCompatibility(oneOfIntTestObjectASchema))
	assert.NoError(t, oneOfIntTestObjectAbSchema.ValidateCompatibility(oneOfIntTestObjectAbSchema))
	assert.Error(t, oneOfIntTestObjectAbSchema.ValidateCompatibility(oneOfIntTestObjectAcSchema))
	assert.Error(t, oneOfIntTestObjectAbSchema.ValidateCompatibility(oneOfIntTestObjectAdSchema))
	assert.Error(t, oneOfIntTestObjectAcSchema.ValidateCompatibility(oneOfIntTestObjectASchema))
	assert.Error(t, oneOfIntTestObjectAcSchema.ValidateCompatibility(oneOfIntTestObjectAbSchema))
	assert.NoError(t, oneOfIntTestObjectAcSchema.ValidateCompatibility(oneOfIntTestObjectAcSchema))
	assert.Error(t, oneOfIntTestObjectAcSchema.ValidateCompatibility(oneOfIntTestObjectAdSchema))
	assert.Error(t, oneOfIntTestObjectAdSchema.ValidateCompatibility(oneOfIntTestObjectASchema))
	assert.Error(t, oneOfIntTestObjectAdSchema.ValidateCompatibility(oneOfIntTestObjectAbSchema))
	assert.Error(t, oneOfIntTestObjectAdSchema.ValidateCompatibility(oneOfIntTestObjectAcSchema))
	assert.NoError(t, oneOfIntTestObjectAdSchema.ValidateCompatibility(oneOfIntTestObjectAdSchema))
}

func TestOneOfIntCompatibilityMapValidation(t *testing.T) {
	validWithObjectB := map[string]any{
		"s": map[string]any{
			"_type": int64(1), // 1 references B
			// object B fields
			"message": "test",
		},
	}
	validWithObjectC := map[string]any{
		"s": map[string]any{
			"_type": int64(2), // 2 references C
			// object B fields
			"m": "test",
		},
	}
	invalidDiscriminator := map[string]any{
		"s": map[string]any{
			"wrongdiscriminator": int64(1), // 2 references B
			// object B fields
			"message": "test",
		},
	}

	combinedMapAndSchema := map[string]any{
		"s": map[string]any{
			"_type": int64(1), // 1 references B
			// object B fields, but with a schema instead of a value
			"message": schema.NewStringSchema(nil, nil, nil),
		},
	}
	combinedMapAndInvalidSchema := map[string]any{
		"s": map[string]any{
			"_type": int64(1), // 1 references B
			// object B fields, but with a schema instead of a value
			"message": schema.NewIntSchema(nil, nil, nil),
		},
	}
	assert.NoError(t, oneOfIntTestObjectASchema.ValidateCompatibility(validWithObjectB))
	assert.NoError(t, oneOfIntTestObjectASchema.ValidateCompatibility(validWithObjectC))
	assert.Error(t, oneOfIntTestObjectASchema.ValidateCompatibility(invalidDiscriminator))
	assert.NoError(t, oneOfIntTestObjectASchema.ValidateCompatibility(combinedMapAndSchema))
	assert.Error(t, oneOfIntTestObjectASchema.ValidateCompatibility(combinedMapAndInvalidSchema))
}

func TestOneOf_Error_OneOfInt_InvalidDiscriminatorType(t *testing.T) {
	assert.Panics(t, func() {
		schema.NewScopeSchema(schema.NewObjectSchema("test",
			map[string]*schema.PropertySchema{
				"test": schema.NewPropertySchema(
					schema.NewOneOfIntSchema[any](map[int64]schema.Object{
						1: inlinedTestIntDiscriminatorASchema,
						2: inlinedTestObjectBMappedSchema,
					}, "d_type", true),
					nil, true, nil, nil,
					nil, nil, nil),
			}))
	})
}
