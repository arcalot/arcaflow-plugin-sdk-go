// golangci-lint does not accurately detect changes in type parameters.
//nolint:dupl
package schema_test

import (
	"encoding/json"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var oneOfIntTestObjectASchema = schema.NewScopeSchema[schema.PropertySchema, schema.ObjectSchema[schema.PropertySchema]](
	map[string]schema.ObjectSchema[schema.PropertySchema]{
		"A": schema.NewObjectSchema(
			"A",
			map[string]schema.PropertySchema{
				"s": schema.NewPropertySchema(
					schema.NewOneOfIntSchema(
						map[int64]schema.RefSchema{
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
		"B": schema.NewObjectSchema(
			"B",
			map[string]schema.PropertySchema{
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
		"C": schema.NewObjectSchema(
			"C",
			map[string]schema.PropertySchema{
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
	},
	"A",
)

var oneOfIntTestObjectAType = schema.NewScopeType[oneOfTestObjectA](
	map[string]schema.ObjectType[any]{
		"A": schema.NewObjectType[oneOfTestObjectA](
			"A",
			map[string]schema.PropertyType{
				"s": schema.NewPropertyType[any](
					schema.NewOneOfIntType(
						map[int64]schema.RefType[any]{
							1: schema.NewRefType[oneOfTestObjectB]("B", nil).Any(),
							2: schema.NewRefType[oneOfTestObjectC]("C", nil).Any(),
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
		).Any(),
		"B": schema.NewObjectType[oneOfTestObjectB](
			"B",
			map[string]schema.PropertyType{
				"message": schema.NewPropertyType[string](
					schema.NewStringType(nil, nil, nil),
					nil,
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"C": schema.NewObjectType[oneOfTestObjectC](
			"C",
			map[string]schema.PropertyType{
				"m": schema.NewPropertyType[string](
					schema.NewStringType(nil, nil, nil),
					nil,
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
	},
	"A",
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
	assertEqual(t, unserializedData.S.(oneOfTestObjectB).Message, "Hello world!")
}
