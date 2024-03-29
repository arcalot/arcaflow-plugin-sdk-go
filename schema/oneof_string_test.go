// golangci-lint does not accurately detect changes in type parameters.
//
//nolint:dupl
package schema_test

import (
	"encoding/json"
	"fmt"
	"go.arcalot.io/assert"
	"reflect"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

const discriminatorFieldName = "d_type"

var oneOfStringTestObjectAProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"C": schema.NewRefSchema("C", nil),
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
var oneOfStringTestObjectAbProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"C": schema.NewRefSchema("C", nil),
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
var oneOfStringTestObjectAcProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"D": schema.NewRefSchema("C", nil),
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
var oneOfStringTestObjectAdProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"C": schema.NewRefSchema("D", nil),
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
	serialized, err := oneOfStringTestObjectAType.Serialize(unserializedData)
	assert.NoError(t, err)
	unserialized2, err := oneOfStringTestObjectAType.Unserialize(serialized)
	assert.NoError(t, err)
	assert.Equals(t, unserialized2, unserializedData)

	// Not explicitly using a struct mapped object, but the type is inferred
	// by the compiler when the oneOfTestBMappedSchema is in the test suite.
	unserializedData, err = oneOfStringTestObjectASchema.Unserialize(input)
	assert.NoError(t, err)
	assert.Equals(t, unserializedData.(map[string]any)["s"].(oneOfTestObjectB).Message, "Hello world!")
	serialized, err = oneOfStringTestObjectASchema.Serialize(unserializedData)
	assert.NoError(t, err)
	unserialized2, err = oneOfStringTestObjectASchema.Unserialize(serialized)
	assert.NoError(t, err)
	assert.Equals(t, unserialized2, unserializedData)
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

type inlinedTestObjectA struct {
	DType       string `json:"d_type"`
	OtherFieldA string `json:"other_field_a"`
}

type inlinedTestObjectB struct {
	DType       string `json:"d_type"`
	OtherFieldB string `json:"other_field_b"`
}

type nonInlinedTestObjectA struct {
	OtherFieldA string `json:"other_field_a"`
}

type nonInlinedTestObjectB struct {
	OtherFieldB string `json:"other_field_b"`
}

var inlinedTestObjectAProperties = map[string]*schema.PropertySchema{
	discriminatorFieldName: schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"other_field_a": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var inlinedTestObjectBProperties = map[string]*schema.PropertySchema{
	discriminatorFieldName: schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"other_field_b": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var nonInlinedTestObjectAProperties = map[string]*schema.PropertySchema{
	"other_field_a": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var nonInlinedTestObjectBProperties = map[string]*schema.PropertySchema{
	"other_field_b": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var inlinedTestObjectAMappedSchema = schema.NewStructMappedObjectSchema[inlinedTestObjectA](
	"inlined_A",
	inlinedTestObjectAProperties,
)

var inlinedTestObjectBMappedSchema = schema.NewStructMappedObjectSchema[inlinedTestObjectB](
	"inlined_B",
	inlinedTestObjectBProperties,
)

var nonInlinedTestObjectAMappedSchema = schema.NewStructMappedObjectSchema[nonInlinedTestObjectA](
	"non_inlined_A",
	nonInlinedTestObjectAProperties,
)

var nonInlinedTestObjectBMappedSchema = schema.NewStructMappedObjectSchema[nonInlinedTestObjectB](
	"non_inlined_B",
	nonInlinedTestObjectBProperties,
)

var inlinedTestObjectASchema = schema.NewObjectSchema(
	"inlined_A",
	inlinedTestObjectAProperties,
)

var inlinedTestObjectBSchema = schema.NewObjectSchema(
	"inlined_B",
	inlinedTestObjectBProperties,
)

var nonInlinedTestObjectASchema = schema.NewObjectSchema(
	"non_inlined_A",
	nonInlinedTestObjectAProperties,
)

var nonInlinedTestObjectBSchema = schema.NewObjectSchema(
	"non_inlined_B",
	nonInlinedTestObjectBProperties,
)

func TestOneOf_InlinedStructMapped(t *testing.T) {
	oneofSchema := schema.NewOneOfStringSchema[any](map[string]schema.Object{
		"A": inlinedTestObjectAMappedSchema,
		"B": inlinedTestObjectBMappedSchema,
	}, discriminatorFieldName, true)
	serializedData := map[string]any{
		discriminatorFieldName: "A",
		"other_field_a":        "test",
	}
	// Since this is struct-mapped, unserializedData is a struct.
	unserializedData := assert.NoErrorR[any](t)(oneofSchema.Unserialize(serializedData))
	reserializedData := assert.NoErrorR[any](t)(oneofSchema.Serialize(unserializedData))
	assert.Equals[any](t, reserializedData, serializedData)
}

func TestOneOf_NonInlinedStructMapped(t *testing.T) {
	oneofSchema := schema.NewOneOfStringSchema[any](map[string]schema.Object{
		"A": nonInlinedTestObjectAMappedSchema,
		"B": nonInlinedTestObjectBMappedSchema,
	}, discriminatorFieldName, false)
	serializedData := map[string]any{
		discriminatorFieldName: "A",
		"other_field_a":        "test",
	}
	// Since this is struct-mapped, unserializedData is a struct.
	unserializedData := assert.NoErrorR[any](t)(oneofSchema.Unserialize(serializedData))
	reserializedData := assert.NoErrorR[any](t)(oneofSchema.Serialize(unserializedData))
	assert.Equals[any](t, reserializedData, serializedData)
}

func TestOneOf_InlinedNonStructMapped(t *testing.T) {
	oneofSchema := schema.NewOneOfStringSchema[any](map[string]schema.Object{
		"A": inlinedTestObjectASchema,
		"B": inlinedTestObjectBSchema,
	}, discriminatorFieldName, true)
	serializedData := map[string]any{
		discriminatorFieldName: "A",
		"other_field_a":        "test",
	}
	// Since this is not struct-mapped, unserializedData is a map.
	unserializedData := assert.NoErrorR[any](t)(oneofSchema.Unserialize(serializedData))
	reserializedData := assert.NoErrorR[any](t)(oneofSchema.Serialize(unserializedData))
	assert.Equals[any](t, reserializedData, serializedData)
}

func TestOneOf_NonInlinedNonStructMapped(t *testing.T) {
	oneofSchema := schema.NewOneOfStringSchema[any](map[string]schema.Object{
		"A": nonInlinedTestObjectASchema,
		"B": nonInlinedTestObjectBSchema,
	}, discriminatorFieldName, false)
	serializedData := map[string]any{
		discriminatorFieldName: "A",
		"other_field_a":        "test",
	}
	// Since this is not struct-mapped, unserializedData is a map.
	unserializedData := assert.NoErrorR[any](t)(oneofSchema.Unserialize(serializedData))
	reserializedData := assert.NoErrorR[any](t)(oneofSchema.Serialize(unserializedData))
	assert.Equals[any](t, reserializedData, serializedData)

	var input_mismatched_type any = struct{}{}
	err := oneofSchema.Validate(input_mismatched_type)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid type for one-of schema")

	var input_invalid_type any = true
	error_msg := fmt.Sprintf("Invalid type for one-of type: %q. Expected map.", reflect.TypeOf(input_invalid_type).Kind())
	_, err = oneofSchema.Unserialize(input_invalid_type)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), error_msg)
	error_msg = fmt.Sprintf("Invalid type for one-of type: %q expected struct or map.", reflect.TypeOf(input_invalid_type).Kind())
	err = oneofSchema.Validate(input_invalid_type)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), error_msg)

	var input_nil any = nil
	_, err = oneofSchema.Unserialize(input_nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bug: data is nil")
}

type inlinedTestIntDiscriminatorA struct {
	DType       int    `json:"d_type"`
	OtherFieldA string `json:"other_field_a"`
}

var inlinedTestIntDiscriminatorAProperties = map[string]*schema.PropertySchema{
	discriminatorFieldName: schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil, true, nil, nil, nil,
		nil, nil,
	),
	"other_field_a": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil, true, nil, nil, nil,
		nil, nil,
	),
}

var inlinedTestNoDiscriminatorBProperties = map[string]*schema.PropertySchema{
	"other_field_b": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil, true, nil, nil, nil,
		nil, nil,
	),
}

var inlinedTestIntDiscriminatorAMappedSchema = schema.NewStructMappedObjectSchema[inlinedTestIntDiscriminatorA](
	"inlined_int_A",
	inlinedTestIntDiscriminatorAProperties,
)

var inlinedTestIntDiscriminatorASchema = schema.NewObjectSchema(
	"inlined_int_A",
	inlinedTestIntDiscriminatorAProperties,
)

var inlinedTestIntDiscriminatorBSchema = schema.NewObjectSchema(
	"inlined_int_B",
	inlinedTestNoDiscriminatorBProperties,
)

func TestOneOf_Error_SubtypeHasInvalidDiscriminatorType(t *testing.T) {
	testSchema := schema.NewOneOfStringSchema[any](map[string]schema.Object{
		"A": inlinedTestIntDiscriminatorAMappedSchema,
		"B": inlinedTestObjectBMappedSchema,
	}, discriminatorFieldName, true)
	expMsg := fmt.Sprintf(
		"discriminator field %q does not match OneOfSchema discriminator type",
		discriminatorFieldName)

	assert.PanicsContains(t, func() {
		schema.NewScopeSchema(schema.NewObjectSchema("test",
			map[string]*schema.PropertySchema{
				"test": schema.NewPropertySchema(
					testSchema,
					nil, true, nil, nil,
					nil, nil, nil),
			}))
	}, expMsg)
}

func TestOneOf_Error_InlineSubtypeMissingDiscriminator(t *testing.T) {
	testSchema := schema.NewOneOfIntSchema[any](map[int64]schema.Object{
		1: inlinedTestIntDiscriminatorASchema,
		2: inlinedTestIntDiscriminatorBSchema,
	}, discriminatorFieldName, true)
	expMsg := "needs discriminator field"

	assert.PanicsContains(t, func() {
		schema.NewScopeSchema(schema.NewObjectSchema("test",
			map[string]*schema.PropertySchema{
				"test": schema.NewPropertySchema(
					testSchema,
					nil, true, nil, nil,
					nil, nil, nil),
			}))
	}, expMsg)
}

func TestOneOf_Error_SubtypeHasDiscriminator(t *testing.T) {
	testSchema := schema.NewOneOfStringSchema[any](map[string]schema.Object{
		"A": inlinedTestIntDiscriminatorASchema,
		"B": nonInlinedTestObjectBSchema,
	}, discriminatorFieldName, false)
	expMsg := "has conflicting field"

	assert.PanicsContains(t, func() {
		schema.NewScopeSchema(schema.NewObjectSchema("test",
			map[string]*schema.PropertySchema{
				"test": schema.NewPropertySchema(
					testSchema,
					nil, true, nil, nil,
					nil, nil, nil),
			}))
	}, expMsg)
}
