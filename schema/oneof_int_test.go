// golangci-lint does not accurately detect changes in type parameters.
//
//nolint:dupl
package schema_test

import (
	"encoding/json"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var oneOfIntTestObjectASchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		map[string]*schema.PropertySchema{
			"s": schema.NewPropertySchema(
				schema.NewOneOfIntSchema[any](
					map[int64]schema.Object{
						1: schema.NewRefSchema("B", nil),
						2: schema.NewRefSchema("C", nil),
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
		},
	),
	schema.NewObjectSchema(
		"B",
		map[string]*schema.PropertySchema{
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
		},
	),
	schema.NewObjectSchema(
		"C",
		map[string]*schema.PropertySchema{
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
		},
	),
)

var oneOfIntTestObjectAType = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[oneOfTestObjectA](
		"A",
		map[string]*schema.PropertySchema{
			"s": schema.NewPropertySchema(
				schema.NewOneOfIntSchema[any](
					map[int64]schema.Object{
						1: schema.NewRefSchema("B", nil),
						2: schema.NewRefSchema("C", nil),
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
		},
	),
	schema.NewStructMappedObjectSchema[oneOfTestObjectB](
		"B",
		map[string]*schema.PropertySchema{
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
		},
	),
	schema.NewStructMappedObjectSchema[oneOfTestObjectC](
		"C",
		map[string]*schema.PropertySchema{
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
		},
	),
)

func TestOneOfIntUnserialization(t *testing.T) {
	data := `{
	"s": {
		"_type": 1,
		"message": "Hello world!"
	}
}`
	var input any
	assertNoError(t, json.Unmarshal([]byte(data), &input))
	unserializedData, err := oneOfIntTestObjectAType.Unserialize(input)
	assertNoError(t, err)
	assertEqual(t, unserializedData.(oneOfTestObjectA).S.(oneOfTestObjectB).Message, "Hello world!")
}
