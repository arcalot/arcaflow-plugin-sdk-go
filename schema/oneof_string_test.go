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
	assert.NoError(t, json.Unmarshal([]byte(data), &input))
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

var oneOfNamePropertiesNoRefs = map[string]*schema.PropertySchema{
	"name": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"fullname": fullnameSchema,
				"nickname": nicknameSchema,
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

var fullnameProperties = map[string]*schema.PropertySchema{
	"first_name": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("first_name"), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"last_name": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("last_name"), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"middle": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("middle"), nil, nil),
		false,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var nicknameProperties = map[string]*schema.PropertySchema{
	"nick": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("nick"), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var fullnameSchema = schema.NewObjectSchema(
	"FullName",
	fullnameProperties,
)

var nicknameSchema = schema.NewObjectSchema(
	"Nickname",
	nicknameProperties,
)

var oneOfNameNoRefsRootScope = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"RootObject",
		oneOfNamePropertiesNoRefs,
	),
	fullnameSchema,
	nicknameSchema,
)

func TestOneOfString_Nickname(t *testing.T) {
	var input any = map[string]any{
		"name": map[string]any{
			"_type": "nickname",
			"nick":  "ArcaLot",
		},
	}
	unserialized, err := oneOfNameNoRefsRootScope.Unserialize(input)
	assert.NoError(t, err)
	serialized, err := oneOfNameNoRefsRootScope.Serialize(unserialized)
	assert.NoError(t, err)
	unserialized2, err := oneOfNameNoRefsRootScope.Unserialize(serialized)
	assert.NoError(t, err)
	assert.Equals(t, unserialized2, unserialized)
}

func TestOneOfString_Fullname(t *testing.T) {
	var input_full any = map[string]any{
		"name": map[string]any{
			"_type":      "fullname",
			"first_name": "Arca",
			"last_name":  "Lot",
		},
	}
	unserialized, err := oneOfNameNoRefsRootScope.Unserialize(input_full)
	assert.NoError(t, err)
	serialized, err := oneOfNameNoRefsRootScope.Serialize(unserialized)
	assert.NoError(t, err)
	unserialized2, err := oneOfNameNoRefsRootScope.Unserialize(serialized)
	assert.Equals(t, unserialized2, unserialized)
}

var discriminatorInline = "_type"

var fullnameInlineProperties = map[string]*schema.PropertySchema{
	discriminatorInline: schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo(discriminatorInline), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"first_name": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("first_name"), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"last_name": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("last_name"), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"middle": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("middle"), nil, nil),
		false,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var nicknameInlineProperties = map[string]*schema.PropertySchema{
	discriminatorInline: schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo(discriminatorInline), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"nick": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(schema.PointerTo("nick"), nil, nil),
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var fullnameInlineSchema = schema.NewObjectSchema(
	"FullName",
	fullnameInlineProperties,
)

var nicknameInlineSchema = schema.NewObjectSchema(
	"Nickname",
	nicknameInlineProperties,
)

var oneOfNameInlineProperties = map[string]*schema.PropertySchema{
	"name": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"fullname": fullnameInlineSchema,
				"nickname": nicknameInlineSchema,
			},
			"_type",
			true,
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

var oneOfNameInlineRootScope = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"RootObject",
		oneOfNameInlineProperties,
	),
	fullnameInlineSchema,
	nicknameInlineSchema,
)

func TestOneOfString_FullnameInline(t *testing.T) {
	var input_full any = map[string]any{
		"name": map[string]any{
			"_type":      "fullname",
			"first_name": "Arca",
			"last_name":  "Lot",
		},
	}
	unserialized, err := oneOfNameInlineRootScope.Unserialize(input_full)
	assert.NoError(t, err)
	serialized, err := oneOfNameInlineRootScope.Serialize(unserialized)
	assert.NoError(t, err)
	unserialized2, err := oneOfNameInlineRootScope.Unserialize(serialized)
	assert.Equals(t, unserialized2, unserialized)
}

var oneOfStringTestInlineObjectAProperties = map[string]*schema.PropertySchema{
	"s": schema.NewPropertySchema(
		schema.NewOneOfStringSchema[any](
			map[string]schema.Object{
				"B": schema.NewRefSchema("B", nil),
				"C": schema.NewRefSchema("C", nil),
			},
			"choice",
			true,
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

var oneOfStringTestInlineObjectAType = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[oneOfTestObjectA](
		"A",
		oneOfStringTestInlineObjectAProperties,
	),
	oneOfTestInlineBMappedSchema,
	oneOfTestInlineCMappedSchema,
)

var oneOfStringTestInlineObjectASchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		oneOfStringTestInlineObjectAProperties,
	),
	oneOfTestInlineBSchema,
	oneOfTestInlineCSchema,
)

func TestOneOfStringInline_Unserialization(t *testing.T) {
	data := `{
	"s": {
		"choice": "B",
		"message": "Hello world!"
	}
}`
	var input any
	assert.NoError(t, json.Unmarshal([]byte(data), &input))
	unserializedData, err := oneOfStringTestInlineObjectAType.Unserialize(input)
	assert.NoError(t, err)
	assert.Equals(t, unserializedData.(oneOfTestObjectA).S.(oneOfTestInlineObjectB).Message, "Hello world!")
	serialized, err := oneOfStringTestInlineObjectAType.Serialize(unserializedData)
	assert.NoError(t, err)
	unserialized2, err := oneOfStringTestInlineObjectAType.Unserialize(serialized)
	assert.NoError(t, err)
	assert.Equals(t, unserialized2, unserializedData)

	// Not explicitly using a struct mapped object, but the type is inferred
	// by the compiler when the oneOfTestBMappedSchema is in the test suite.
	assert.NoError(t, json.Unmarshal([]byte(data), &input))
	unserializedData, err = oneOfStringTestInlineObjectASchema.Unserialize(input)
	assert.NoError(t, err)
	assert.Equals(t, unserializedData.(map[string]any)["s"].(oneOfTestInlineObjectB).Message, "Hello world!")
	serialized, err = oneOfStringTestInlineObjectASchema.Serialize(unserializedData)
	assert.NoError(t, err)
	unserialized2, err = oneOfStringTestInlineObjectASchema.Unserialize(serialized)
	assert.NoError(t, err)
	assert.Equals(t, unserialized2, unserializedData)
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
	"d_type": schema.NewPropertySchema(
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
	"d_type": schema.NewPropertySchema(
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
	}, "d_type", true)
	assert.NoError(t, oneofSchema.ValidateSubtypeDiscriminatorInlineFields())
	serializedData := map[string]any{
		"d_type":        "A",
		"other_field_a": "test",
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
	}, "d_type", false)
	serializedData := map[string]any{
		"d_type":        "A",
		"other_field_a": "test",
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
	}, "d_type", true)
	serializedData := map[string]any{
		"d_type":        "A",
		"other_field_a": "test",
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
	}, "d_type", false)
	serializedData := map[string]any{
		"d_type":        "A",
		"other_field_a": "test",
	}
	// Since this is not struct-mapped, unserializedData is a map.
	unserializedData := assert.NoErrorR[any](t)(oneofSchema.Unserialize(serializedData))
	reserializedData := assert.NoErrorR[any](t)(oneofSchema.Serialize(unserializedData))
	assert.Equals[any](t, reserializedData, serializedData)

}

type inlinedTestIntDiscriminatorA struct {
	DType       int    `json:"d_type"`
	OtherFieldA string `json:"other_field_a"`
}

type inlinedTestIntDiscriminatorB struct {
	DType       int    `json:"d_type"`
	OtherFieldB string `json:"other_field_b"`
}

var inlinedTestIntDiscriminatorAProperties = map[string]*schema.PropertySchema{
	"d_type": schema.NewPropertySchema(
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

var inlinedTestIntDiscriminatorBProperties = map[string]*schema.PropertySchema{
	"d_type": schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil, true, nil, nil, nil,
		nil, nil,
	),
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

var inlinedTestIntDiscriminatorBMappedSchema = schema.NewStructMappedObjectSchema[inlinedTestIntDiscriminatorB](
	"inlined_int_B",
	inlinedTestIntDiscriminatorBProperties,
)

var inlinedTestIntDiscriminatorASchema = schema.NewObjectSchema(
	"inlined_int_A",
	inlinedTestIntDiscriminatorAProperties,
)

var inlinedTestIntDiscriminatorBSchema = schema.NewObjectSchema(
	"inlined_int_B",
	inlinedTestIntDiscriminatorBProperties,
)

func TestOneOf_Error_InvalidDiscriminatorTypeInSubtype(t *testing.T) {
	oneofSchema := schema.NewOneOfStringSchema[any](map[string]schema.Object{
		"A": inlinedTestIntDiscriminatorAMappedSchema,
		"B": inlinedTestObjectBMappedSchema,
	}, "d_type", true)
	assert.Error(t, oneofSchema.ValidateSubtypeDiscriminatorInlineFields())
	// check error message
}

func TestOneOf_Error_OneOfInt_InvalidDiscriminatorType(t *testing.T) {
	oneofSchema := schema.NewOneOfIntSchema[any](map[int64]schema.Object{
		1: inlinedTestIntDiscriminatorASchema,
		2: inlinedTestObjectBMappedSchema,
	}, "d_type", true)
	assert.Error(t, oneofSchema.ValidateSubtypeDiscriminatorInlineFields())
	//assert.Panics(t, func() {
	//	schema.NewOneOfIntSchema[any](map[int64]schema.Object{
	//		1: inlinedTestIntDiscriminatorASchema,
	//		2: inlinedTestObjectBMappedSchema,
	//	}, "d_type", true)
	//})
	// check error message
}
