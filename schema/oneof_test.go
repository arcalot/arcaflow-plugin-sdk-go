package schema_test

import (
	"fmt"
	"testing"

	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
)

type oneOfTestObjectB struct {
	Message string `json:"message"`
}

func (o oneOfTestObjectB) String() string {
	return o.Message
}

type oneOfTestObjectC struct {
	M string `json:"m"`
}

type oneOfTestObjectA struct {
	S any `json:"s"`
}

func TestOneOfTypeID(t *testing.T) {
	assert.Equals(
		t,
		oneOfStringTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assert.Equals(
		t,
		oneOfStringTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assert.Equals(
		t,
		oneOfIntTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
	assert.Equals(
		t,
		oneOfIntTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
}

var oneOfTestObjectBProperties = map[string]*schema.PropertySchema{
	"message": schema.NewPropertySchema(
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

var oneOfTestObjectCProperties = map[string]*schema.PropertySchema{
	"m": schema.NewPropertySchema(
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

var oneOfTestObjectDProperties = map[string]*schema.PropertySchema{
	"K": schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestBSchema = schema.NewObjectSchema(
	"B",
	oneOfTestObjectBProperties,
)

var oneOfTestCSchema = schema.NewObjectSchema(
	"C",
	oneOfTestObjectCProperties,
)

var oneOfTestDSchema = schema.NewObjectSchema(
	"D",
	oneOfTestObjectDProperties,
)

var oneOfTestBMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestObjectB](
	"B",
	oneOfTestObjectBProperties,
)

var oneOfTestCMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestObjectC](
	"C",
	oneOfTestObjectCProperties,
)

// Test_OneOf_ConstructorBypass tests the behavior of a OneOf object created
// by the Scope ScopeSchema, a scope that contains the schema of a scope
// and an object, through unserialization of data without using a
// New* constructor function, like NewOneOfStringSchema or NewOneOfIntSchema,
// behaves as one would expect from a OneOf object created from a constructor.
func Test_OneOf_ConstructorBypass(t *testing.T) { //nolint:funlen
	discriminator_field := "_type"
	input_schema := map[string]any{
		"root": "InputParams",
		"objects": map[string]any{
			"InputParams": map[string]any{
				"id": "InputParams",
				"properties": map[string]any{
					"name": map[string]any{
						"required": true,
						"type": map[string]any{
							"discriminator_field_name": discriminator_field,
							"type_id":                  "one_of_string",
							"types": map[string]any{
								"fullname": map[string]any{
									"id":      "FullName",
									"type_id": "ref",
								},
								"nick": map[string]any{
									"id":      "Nickname",
									"type_id": "ref",
								},
							},
						},
					},
				},
			},
			"FullName": map[string]any{
				"id": "FullName",
				"properties": map[string]any{
					"first_name": map[string]any{
						"required": true,
						"type": map[string]any{
							"type_id": "string",
						},
					},
					"last_name": map[string]any{
						"required": true,
						"type": map[string]any{
							"type_id": "string",
						},
					},
				},
			},
			"Nickname": map[string]any{
				"id": "Nickname",
				"properties": map[string]any{
					"nick": map[string]any{
						"required": true,
						"type": map[string]any{
							"type_id": "string",
						},
					},
				},
			},
		},
	}
	var input_data_fullname any = map[string]any{
		"name": map[string]any{
			discriminator_field: "fullname",
			"first_name":        "Arca",
			"last_name":         "Lot",
		},
	}

	scopeAny := assert.NoErrorR[any](t)(schema.DescribeScope().Unserialize(input_schema))
	scopeSchemaTyped := scopeAny.(*schema.ScopeSchema)
	scopeSchemaTyped.ApplyScope(scopeSchemaTyped)
	assert.NoError(t, scopeSchemaTyped.Validate(input_data_fullname))
	unserialized := assert.NoErrorR[any](t)(scopeSchemaTyped.Unserialize(input_data_fullname))
	serialized := assert.NoErrorR[any](t)(scopeSchemaTyped.Serialize(unserialized))
	unserialized2 := assert.NoErrorR[any](t)(scopeSchemaTyped.Unserialize(serialized))
	assert.Equals(t, unserialized2, unserialized)

	var input_invalid_discriminator_value any = map[string]any{
		"name": map[string]any{
			discriminator_field: "robotname",
			"first_name":        "Arca",
			"last_name":         "Lot",
		},
	}
	error_msg := fmt.Sprintf("Invalid value for %q", discriminator_field)
	_, err := scopeSchemaTyped.Unserialize(input_invalid_discriminator_value)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), error_msg)
}
