package schema_test

import (
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

type oneOfTestInlineObjectB struct {
	Message string `json:"message"`
	Choice  string `json:"choice"`
}

func (o oneOfTestInlineObjectB) String() string {
	return o.Message
}

type oneOfTestInlineObjectC struct {
	M      string `json:"m"`
	Choice string `json:"choice"`
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

var oneOfTestInlineObjectBProperties = map[string]*schema.PropertySchema{
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
	"choice": schema.NewPropertySchema(
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

var oneOfTestInlineObjectCProperties = map[string]*schema.PropertySchema{
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
	"choice": schema.NewPropertySchema(
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

var oneOfTestInlineBSchema = schema.NewObjectSchema(
	"B",
	oneOfTestInlineObjectBProperties,
)

var oneOfTestInlineCSchema = schema.NewObjectSchema(
	"C",
	oneOfTestInlineObjectCProperties,
)

var oneOfTestInlineBMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestInlineObjectB](
	"B",
	oneOfTestInlineObjectBProperties,
)

var oneOfTestInlineCMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestInlineObjectC](
	"C",
	oneOfTestInlineObjectCProperties,
)

func Test_OneOfString_ConstructorBypass(t *testing.T) {
	input_schema := map[string]any{
		"root": "InputParams",
		"objects": map[string]any{
			"InputParams": map[string]any{
				"id": "InputParams",
				"properties": map[string]any{
					"name": map[string]any{
						"required": true,
						"type": map[string]any{
							"discriminator_field_name": "_type",
							"discriminator_inlined":    false,
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
			"_type":      "fullname",
			"first_name": "Arca",
			"last_name":  "Lot",
		},
	}
	scopeAny := assert.NoErrorR[any](t)(schema.DescribeScope().Unserialize(input_schema))
	scopeSchemaTyped := scopeAny.(*schema.ScopeSchema)
	scopeSchemaTyped.ApplyScope(scopeSchemaTyped)
	unserialized := assert.NoErrorR[any](t)(scopeSchemaTyped.Unserialize(input_data_fullname))
	serialized := assert.NoErrorR[any](t)(scopeSchemaTyped.Serialize(unserialized))
	unserialized2 := assert.NoErrorR[any](t)(scopeSchemaTyped.Unserialize(serialized))
	assert.Equals(t, unserialized2, unserialized)
}
