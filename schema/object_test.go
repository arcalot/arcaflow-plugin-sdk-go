package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type testStruct struct {
	Field1 int64
	Field2 string `json:"field3"`
}

var testStructSchema = schema.NewTypedObject[testStruct]("testStruct", map[string]*schema.PropertySchema{
	"Field1": schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"field3": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
})

type testStructPtr struct {
	Field1 *int64
	Field2 *string `json:"field3"`
}

var testStructSchemaPtr = schema.NewTypedObject[*testStructPtr]("testStruct", map[string]*schema.PropertySchema{
	"Field1": schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"field3": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
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
	data := map[string]any{
		"Field1": 42,
		"field3": "Hello world!",
	}

	t.Run("noptr", func(t *testing.T) {
		unserializedData, err := testStructSchema.UnserializeType(data)
		assertNoError(t, err)
		assertInstanceOf[testStruct](t, unserializedData)
		assertEqual(t, unserializedData.Field1, int64(42))
		assertEqual(t, unserializedData.Field2, "Hello world!")
	})

	t.Run("ptr", func(t *testing.T) {
		unserializedDataPtr, err := testStructSchemaPtr.UnserializeType(data)
		assertNoError(t, err)
		assertInstanceOf[*testStructPtr](t, unserializedDataPtr)
		assertNotNil(t, unserializedDataPtr.Field1)
		assertNotNil(t, unserializedDataPtr.Field2)
		assertEqual(t, *unserializedDataPtr.Field1, int64(42))
		assertEqual(t, *unserializedDataPtr.Field2, "Hello world!")
	})
}

type embeddedTestStruct struct {
	Field1 int64
}

type testStructWithEmbed struct {
	embeddedTestStruct `json:",inline"`
	Field2             string `json:"field3"`
}

var testStructWithEmbedSchema = schema.NewTypedObject[testStructWithEmbed]("testStruct", map[string]*schema.PropertySchema{
	"Field1": schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"field3": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
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
	unserializedData, err := testStructWithEmbedSchema.UnserializeType(map[string]any{
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

type testOptionalFieldStruct struct {
	A *string `json:"a"`
}

var testOptionalFieldSchema = schema.NewTypedObject[testOptionalFieldStruct](
	"testOptionalFieldStruct",
	map[string]*schema.PropertySchema{
		"a": schema.NewPropertySchema(
			schema.NewStringSchema(nil, nil, nil),
			nil,
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
		),
	},
)

func TestOptionalField(t *testing.T) {
	data, err := testOptionalFieldSchema.UnserializeType(map[string]any{})
	assertNoError(t, err)
	if data.A != nil {
		t.Fatalf("Unexpected value: %s", *data.A)
	}
}
