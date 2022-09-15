package schema_test

import (
	"encoding/json"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type scopeTestObjectB struct {
	C string `json:"c"`
}

type scopeTestObjectA struct {
	B scopeTestObjectB `json:"b"`
}

var scopeTestObjectASchema = schema.NewScopeSchema[schema.PropertySchema, schema.ObjectSchema[schema.PropertySchema]](
	map[string]schema.ObjectSchema[schema.PropertySchema]{
		"scopeTestObjectA": schema.NewObjectSchema(
			"scopeTestObjectA",
			map[string]schema.PropertySchema{
				"b": schema.NewPropertySchema(
					schema.NewRefSchema("scopeTestObjectB", nil),
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
		"scopeTestObjectB": schema.NewObjectSchema(
			"scopeTestObjectB",
			map[string]schema.PropertySchema{
				"c": schema.NewPropertySchema(
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
	"scopeTestObjectA",
)

var scopeTestObjectAType = schema.NewScopeType[scopeTestObjectA](
	map[string]schema.ObjectType[any]{
		"scopeTestObjectA": schema.NewObjectType[scopeTestObjectA](
			"scopeTestObjectA",
			map[string]schema.PropertyType{
				"b": schema.NewPropertyType[scopeTestObjectB](
					schema.NewRefType[scopeTestObjectB]("scopeTestObjectB", nil),
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
		"scopeTestObjectB": schema.NewObjectType[scopeTestObjectB](
			"scopeTestObjectB",
			map[string]schema.PropertyType{
				"c": schema.NewPropertyType[string](
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
	"scopeTestObjectA",
)

func TestScopeConstructor(t *testing.T) {
	assertEqual(t, scopeTestObjectASchema.TypeID(), schema.TypeIDScope)
	assertEqual(t, scopeTestObjectAType.TypeID(), schema.TypeIDScope)
}

func TestUnserialization(t *testing.T) {
	data := `{"b":{"c": "Hello world!"}}`
	var input any
	assertNoError(t, json.Unmarshal([]byte(data), &input))
	result, err := scopeTestObjectAType.Unserialize(input)
	assertNoError(t, err)
	assertEqual(t, result.B.C, "Hello world!")
}

func TestValidation(t *testing.T) {
	err := scopeTestObjectAType.Validate(scopeTestObjectA{
		scopeTestObjectB{
			"Hello world!",
		},
	})
	assertNoError(t, err)
}

func TestSerialization(t *testing.T) {
	serialized, err := scopeTestObjectAType.Serialize(scopeTestObjectA{
		scopeTestObjectB{
			"Hello world!",
		},
	})
	assertNoError(t, err)
	assertEqual(t, serialized.(map[string]any)["b"].(map[string]any)["c"].(string), "Hello world!")
}
