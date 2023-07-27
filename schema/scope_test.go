package schema_test

import (
	"encoding/json"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type scopeTestObjectB struct {
	C string `json:"c"`
}

type scopeTestObjectA struct {
	B scopeTestObjectB `json:"b"`
}

type scopeTestObjectAPtr struct {
	B *scopeTestObjectB `json:"b"`
}

var scopeTestObjectASchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"scopeTestObjectA",
		map[string]*schema.PropertySchema{
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
	schema.NewObjectSchema(
		"scopeTestObjectB",
		map[string]*schema.PropertySchema{
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
)

var scopeTestObjectAType = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[scopeTestObjectA](
		"scopeTestObjectA",
		map[string]*schema.PropertySchema{
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
	schema.NewStructMappedObjectSchema[scopeTestObjectB](
		"scopeTestObjectB",
		map[string]*schema.PropertySchema{
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
)

var scopeTestObjectATypePtr = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[*scopeTestObjectAPtr](
		"scopeTestObjectA",
		map[string]*schema.PropertySchema{
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
	schema.NewStructMappedObjectSchema[*scopeTestObjectB](
		"scopeTestObjectB",
		map[string]*schema.PropertySchema{
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
	assertInstanceOf[scopeTestObjectA](t, result.(scopeTestObjectA))
	assertEqual(t, result.(scopeTestObjectA).B.C, "Hello world!")

	resultPtr, err := scopeTestObjectATypePtr.Unserialize(input)
	assertNoError(t, err)
	assertInstanceOf[*scopeTestObjectAPtr](t, resultPtr.(*scopeTestObjectAPtr))
	assertEqual(t, resultPtr.(*scopeTestObjectAPtr).B.C, "Hello world!")
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
	assert.NoError(t, err)
	assert.Equals(t, serialized.(map[string]any)["b"].(map[string]any)["c"].(string), "Hello world!")
}

func TestSelfSerialization(t *testing.T) {
	serializedScope, err := scopeTestObjectAType.SelfSerialize()
	assert.NoError(t, err)
	serializedScopeMap := serializedScope.(map[string]any)
	if serializedScopeMap["root"] != "scopeTestObjectA" {
		t.Fatalf("Unexpected root object: %s", serializedScopeMap["root"])
	}
}
