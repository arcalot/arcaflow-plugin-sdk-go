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

var oneOfStringTestObjectASchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"A",
		map[string]*schema.PropertySchema{
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
			"M": schema.NewPropertySchema(
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

var oneOfStringTestObjectAType = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[oneOfTestObjectA](
		"A",
		map[string]*schema.PropertySchema{
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
