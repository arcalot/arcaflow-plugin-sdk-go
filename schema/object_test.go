package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type testStruct struct {
	Field1 int64
	Field2 string `json:"field3"`
}

var testStructSchema = schema.NewObjectType[testStruct]("testStruct", map[string]schema.PropertyType{
	"Field1": schema.NewPropertyType[int64](
		schema.NewIntType(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"field3": schema.NewPropertyType[string](
		schema.NewStringType(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
})

func TestObjectUnserialization(t *testing.T) {
	unserializedData, err := testStructSchema.Unserialize(map[string]any{
		"Field1": 42,
		"field3": "Hello world!",
	})
	assertNoError(t, err)
	assertEqual(t, unserializedData.Field1, int64(42))
	assertEqual(t, unserializedData.Field2, "Hello world!")
}

type embeddedTestStruct struct {
	Field1 int64
}

type testStructWithEmbed struct {
	embeddedTestStruct `json:",inline"`
	Field2             string `json:"field3"`
}

var testStructWithEmbedSchema = schema.NewObjectType[testStructWithEmbed]("testStruct", map[string]schema.PropertyType{
	"Field1": schema.NewPropertyType[int64](
		schema.NewIntType(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"field3": schema.NewPropertyType[string](
		schema.NewStringType(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
})

func TestObjectUnserializationEmbeddedStruct(t *testing.T) {
	unserializedData, err := testStructWithEmbedSchema.Unserialize(map[string]any{
		"Field1": 42,
		"field3": "Hello world!",
	})
	assertNoError(t, err)
	assertEqual(t, unserializedData.Field1, int64(42))
	assertEqual(t, unserializedData.Field2, "Hello world!")
}

func TestObjectSerialization(t *testing.T) {
	testData := testStruct{
		Field1: 42,
		Field2: "Hello world!",
	}

	serializedData, err := testStructSchema.Serialize(testData)
	assertNoError(t, err)

	typedData := serializedData.(map[string]any)

	assertEqual(t, len(typedData), 2)
	assertEqual(t, typedData["Field1"].(int64), int64(42))
	assertEqual(t, typedData["field3"].(string), "Hello world!")
}

func TestObjectSerializationEmbedded(t *testing.T) {
	testData := testStructWithEmbed{
		embeddedTestStruct{
			Field1: 42,
		},
		"Hello world!",
	}

	serializedData, err := testStructWithEmbedSchema.Serialize(testData)
	assertNoError(t, err)

	typedData := serializedData.(map[string]any)

	assertEqual(t, len(typedData), 2)
	assertEqual(t, typedData["Field1"].(int64), int64(42))
	assertEqual(t, typedData["field3"].(string), "Hello world!")
}

func TestObjectValidation(t *testing.T) {
	testData := testStruct{
		Field1: 42,
		Field2: "Hello world!",
	}

	assertNoError(t, testStructSchema.Validate(testData))
}

func TestObjectValidationEmbedded(t *testing.T) {
	testData := testStructWithEmbed{
		embeddedTestStruct{
			Field1: 42,
		},
		"Hello world!",
	}

	assertNoError(t, testStructWithEmbedSchema.Validate(testData))
}
