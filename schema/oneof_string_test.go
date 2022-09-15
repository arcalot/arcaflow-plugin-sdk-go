package schema_test

import (
	"encoding/json"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var oneOfStringTestObjectASchema = schema.NewScopeSchema[schema.PropertySchema, schema.ObjectSchema[schema.PropertySchema]](
	map[string]schema.ObjectSchema[schema.PropertySchema]{
		"A": schema.NewObjectSchema(
			"A",
			map[string]schema.PropertySchema{
				"s": schema.NewPropertySchema(
					schema.NewOneOfStringSchema(
						map[string]schema.RefSchema{
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
	},
	"A",
)

var oneOfStringTestObjectAType = schema.NewScopeType[oneOfTestObjectA](
	map[string]schema.ObjectType[any]{
		"A": schema.NewObjectType[oneOfTestObjectA](
			"A",
			map[string]schema.PropertyType{
				"s": schema.NewPropertyType[any](
					schema.NewOneOfStringType(
						map[string]schema.RefType[any]{
							"B": schema.NewRefType[oneOfTestObjectB]("B", nil).Anonymous(),
							"C": schema.NewRefType[oneOfTestObjectC]("C", nil).Anonymous(),
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
		).Anonymous(),
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
		).Anonymous(),
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
		).Anonymous(),
	},
	"A",
)

func TestOneOfStringUnserialization(t *testing.T) {
	data := `{
	"s": {
		"_type": "B",
		"message": "Hello world!"
	}
}`
	var input any
	assertNoError(t, json.Unmarshal([]byte(data), &input))
	unserializedData, err := oneOfStringTestObjectAType.Unserialize(input)
	assertNoError(t, err)
	assertEqual(t, unserializedData.S.(oneOfTestObjectB).Message, "Hello world!")
}
