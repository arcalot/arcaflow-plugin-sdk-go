package schema_test

import (
	"go.arcalot.io/assert"
	"strconv"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type testStruct struct {
	Field1 int64
	Field2 string `json:"field3"`
}

var fieldKey1 = "Field1"
var fieldKey2 = "field3"

var testSchemaProperties = map[string]*schema.PropertySchema{
	fieldKey1: schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil, true, nil, nil, nil,
		nil, nil,
	),
	fieldKey2: schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil, true, nil, nil, nil,
		nil, nil,
	),
}

var testStructSchema = schema.NewTypedObject[testStruct](
	"testStruct", testSchemaProperties)

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
		assert.NoError(t, err)
		assert.InstanceOf[testStruct](t, unserializedData)
		assert.Equals(t, unserializedData.Field1, int64(42))
		assert.Equals(t, unserializedData.Field2, "Hello world!")
	})

	t.Run("ptr", func(t *testing.T) {
		unserializedDataPtr, err := testStructSchemaPtr.UnserializeType(data)
		assert.NoError(t, err)
		assert.InstanceOf[*testStructPtr](t, unserializedDataPtr)
		assert.NotNil(t, unserializedDataPtr.Field1)
		assert.NotNil(t, unserializedDataPtr.Field2)
		assert.Equals(t, *unserializedDataPtr.Field1, int64(42))
		assert.Equals(t, *unserializedDataPtr.Field2, "Hello world!")
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
	assert.NoError(t, err)
	assert.Equals(t, unserializedData.Field1, int64(42))
	assert.Equals(t, unserializedData.Field2, "Hello world!")
}

func TestObjectSerialization(t *testing.T) {
	testData := testStruct{
		Field1: 42,
		Field2: "Hello world!",
	}

	serializedData, err := testStructSchema.Serialize(testData)
	assert.NoError(t, err)

	typedData := serializedData.(map[string]any)

	assert.Equals(t, len(typedData), 2)
	assert.Equals(t, typedData["Field1"].(int64), int64(42))
	assert.Equals(t, typedData["field3"].(string), "Hello world!")
}

func TestObjectSerializationEmbedded(t *testing.T) {
	testData := testStructWithEmbed{
		embeddedTestStruct{
			Field1: 42,
		},
		"Hello world!",
	}

	serializedData, err := testStructWithEmbedSchema.Serialize(testData)
	assert.NoError(t, err)

	typedData := serializedData.(map[string]any)

	assert.Equals(t, len(typedData), 2)
	assert.Equals(t, typedData["Field1"].(int64), int64(42))
	assert.Equals(t, typedData["field3"].(string), "Hello world!")
}

func TestObjectValidation(t *testing.T) {
	testData := testStruct{
		Field1: 42,
		Field2: "Hello world!",
	}

	assert.NoError(t, testStructSchema.Validate(testData))
}

func TestObjectValidationEmbedded(t *testing.T) {
	testData := testStructWithEmbed{
		embeddedTestStruct{
			Field1: 42,
		},
		"Hello world!",
	}

	assert.NoError(t, testStructWithEmbedSchema.Validate(testData))
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
	assert.NoError(t, err)
	if data.A != nil {
		t.Fatalf("Unexpected value: %s", *data.A)
	}
}

//nolint:funlen
func TestObjectNestedDefaults(t *testing.T) {
	type nested struct {
		A string `json:"a"`
	}
	nestedProperty := schema.NewPropertySchema(
		schema.NewRefSchema("nested", nil),
		nil,
		false,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	t.Run("nested-nopointer", func(t *testing.T) {
		type root1 struct {
			Nested nested `json:"nested"`
		}
		scope1 := schema.NewTypedScopeSchema[root1](
			schema.NewStructMappedObjectSchema[root1]("root1", map[string]*schema.PropertySchema{
				"nested": nestedProperty,
			}),
			schema.NewStructMappedObjectSchema[nested](
				"nested",
				map[string]*schema.PropertySchema{
					"a": schema.NewPropertySchema(
						schema.NewStringSchema(nil, nil, nil),
						nil,
						false,
						nil,
						nil,
						nil,
						schema.PointerTo("\"Hello world!\""),
						nil,
					),
				},
			),
		)
		unserialized1, err := scope1.UnserializeType(map[string]any{})
		assert.NoError(t, err)
		assert.Equals(t, unserialized1.Nested.A, "Hello world!")
	})

	t.Run("nested-pointer", func(t *testing.T) {
		type root2 struct {
			Nested *nested `json:"nested"`
		}
		scope2 := schema.NewTypedScopeSchema[root2](
			schema.NewStructMappedObjectSchema[root2]("root2", map[string]*schema.PropertySchema{
				"nested": nestedProperty,
			}),
			schema.NewStructMappedObjectSchema[*nested](
				"nested",
				map[string]*schema.PropertySchema{
					"a": schema.NewPropertySchema(
						schema.NewStringSchema(nil, nil, nil),
						nil,
						false,
						nil,
						nil,
						nil,
						schema.PointerTo("\"Hello world!\""),
						nil,
					),
				},
			),
		)
		unserialized2, err := scope2.UnserializeType(map[string]any{})
		assert.NoError(t, err)
		assert.Nil(t, unserialized2.Nested)
	})

	t.Run("nested-nopointer-double", func(t *testing.T) {
		type nested2 struct {
			Nested nested `json:"nested"`
		}
		type root3 struct {
			Nested nested2 `json:"nested"`
		}
		scope3 := schema.NewTypedScopeSchema[root3](
			schema.NewStructMappedObjectSchema[root3]("root3", map[string]*schema.PropertySchema{
				"nested": schema.NewPropertySchema(
					schema.NewRefSchema("nested2", nil),
					nil,
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			}),
			schema.NewStructMappedObjectSchema[nested2](
				"nested2",
				map[string]*schema.PropertySchema{
					"nested": nestedProperty,
				},
			),
			schema.NewStructMappedObjectSchema[nested](
				"nested",
				map[string]*schema.PropertySchema{
					"a": schema.NewPropertySchema(
						schema.NewStringSchema(nil, nil, nil),
						nil,
						false,
						nil,
						nil,
						nil,
						schema.PointerTo("\"Hello world!\""),
						nil,
					),
				},
			),
		)
		unserialized3, err := scope3.UnserializeType(map[string]any{})
		assert.NoError(t, err)
		assert.Equals(t, unserialized3.Nested.Nested.A, "Hello world!")
	})
}

func TestTypedString(t *testing.T) {
	type testEnum string
	type testStruct struct {
		T1 testEnum  `json:"t1"`
		T2 *testEnum `json:"t2"`
	}
	o := schema.NewStructMappedObjectSchema[testStruct](
		"testStruct",
		map[string]*schema.PropertySchema{
			"t1": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
			"t2": schema.NewPropertySchema(
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
	result, err := o.Unserialize(map[string]any{"t1": "Hello world!"})
	assert.NoError(t, err)
	assert.Equals(t, result.(testStruct).T1, "Hello world!")
	serialized, err := o.Serialize(result)
	assert.NoError(t, err)
	unserialized2, err := o.Unserialize(serialized)
	assert.NoError(t, err)
	// test reversiblity
	assert.Equals(t, unserialized2, result)

	result, err = o.Unserialize(map[string]any{"t2": "Hello world!"})
	assert.NoError(t, err)
	assert.Equals(t, *result.(testStruct).T2, "Hello world!")
	serialized, err = o.Serialize(result)
	assert.NoError(t, err)
	unserialized2, err = o.Unserialize(serialized)
	assert.NoError(t, err)
	// test reversiblity
	assert.Equals(t, unserialized2, result)
}

func TestNonDefaultSerialization(t *testing.T) {
	type TestData struct {
		Foo *string `json:"foo"`
	}
	s := schema.NewStructMappedObjectSchema[TestData](
		"TestData",
		map[string]*schema.PropertySchema{
			"foo": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				false,
				nil,
				nil,
				nil,
				schema.PointerTo(`"Hello world!"`),
				nil,
			),
		},
	)
	text := "Hello Arca Lot!"
	serializedData, err := s.Serialize(TestData{&text})
	assert.NoError(t, err)
	assert.Equals(t, serializedData.(map[string]any)["foo"].(string), text)

	unserialized, err := s.Unserialize(serializedData)
	assert.NoError(t, err)
	serialized2, err := s.Serialize(unserialized)
	assert.NoError(t, err)
	// test reversiblity
	assert.Equals(t, serialized2, serializedData)
}

func TestTypedObjectSchema_Any(t *testing.T) {
	type TestData struct {
		Foo *string `json:"foo"`
	}
	s := schema.NewTypedObject[TestData](
		"TestData",
		map[string]*schema.PropertySchema{
			"foo": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				false,
				nil,
				nil,
				nil,
				schema.PointerTo(`"Hello world!"`),
				nil,
			),
		},
	)
	anyObject := s.Any()
	text := "Hello Arca Lot!"
	serializedData, err := anyObject.SerializeType(TestData{&text})
	assert.NoError(t, err)
	assert.Equals(t, serializedData.(map[string]any)["foo"].(string), text)

	_, err = anyObject.SerializeType(text)
	assert.Error(t, err)

	unserialized, err := s.Unserialize(serializedData)
	assert.NoError(t, err)
	serialized2, err := s.Serialize(unserialized)
	assert.NoError(t, err)
	// test reversibility
	assert.Equals(t, serialized2, serializedData)
}

func TestDefaultsStructSerialization(t *testing.T) {
	type TestData struct {
		Foo *string `json:"foo"`
	}
	default_foo_value := "abc"
	s := schema.NewTypedObject[TestData](
		"TestData",
		map[string]*schema.PropertySchema{
			"foo": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				false,
				nil,
				nil,
				nil,
				schema.PointerTo(strconv.Quote(default_foo_value)),
				nil,
			),
		},
	)
	// First, unserialization
	unserialized, err := s.Unserialize(map[string]any{})
	assert.NoError(t, err)
	assert.NotNil(t, unserialized)
	assert.InstanceOf[TestData](t, unserialized)
	assert.NotNil(t, unserialized.(TestData).Foo)
	// Validate that default is included
	assert.Equals(t, *unserialized.(TestData).Foo, default_foo_value)

	// Next, serialization.
	serialized, err := s.Serialize(unserialized)
	assert.NoError(t, err)
	assert.NotNil(t, serialized)
	assert.InstanceOf[map[string]any](t, serialized)
	actual_value := assert.MapContainsKey[string](
		t, "foo", serialized.(map[string]any))
	assert.Equals(t,
		actual_value.(string),
		default_foo_value)

	unserialized2, err := s.Unserialize(serialized)
	assert.NoError(t, err)
	// test unserialize and serialize are reversible
	assert.Equals(t, unserialized2, unserialized)
}

func TestDefaultsObjectSerialization(t *testing.T) {
	foo_key := "foo"
	default_foo_value := "abc"

	s := schema.NewObjectSchema(
		"TestData",
		map[string]*schema.PropertySchema{
			foo_key: schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				false,
				nil,
				nil,
				nil,
				schema.PointerTo(strconv.Quote(default_foo_value)),
				nil,
			),
		},
	)
	// First, unserialization
	unserialized, err := s.Unserialize(map[string]any{})
	assert.NoError(t, err)
	assert.NotNil(t, unserialized)
	assert.InstanceOf[map[string]any](t, unserialized)
	assert.MapContainsKey[string](t, foo_key, unserialized.(map[string]any))
	// Validate that default is included
	assert.Equals(t,
		unserialized.(map[string]any)[foo_key].(string), default_foo_value)

	// Next, serialization.
	serialized, err := s.Serialize(unserialized)
	assert.NoError(t, err)
	assert.NotNil(t, serialized)
	assert.InstanceOf[map[string]any](t, serialized)
	actual_value := assert.MapContainsKey[string](
		t, foo_key, serialized.(map[string]any))
	assert.Equals(t, actual_value.(string), default_foo_value)

	unserialized2, err := s.Unserialize(serialized)
	assert.NoError(t, err)
	// test unserialize and serialize are reversible
	assert.Equals(t, unserialized2, unserialized)
}

var testStructScope = schema.NewScopeSchema(&testStructSchema.ObjectSchema)

func TestObjectSchema_ValidateCompatibility(t *testing.T) {
	// Schema validation
	assert.NoError(t, testStructSchema.ValidateCompatibility(testStructSchema))
	assert.Error(t, testStructSchema.ValidateCompatibility(testOptionalFieldSchema)) // Not the same
	// Schema validation with ref
	objectTestRef := schema.NewRefSchema("testStruct", nil)
	objectTestRef.ApplyScope(testStructScope)
	assert.NoError(t, objectTestRef.ValidateCompatibility(testStructSchema))
	assert.NoError(t, testStructSchema.ValidateCompatibility(objectTestRef))
	// Schema validation with scope
	testStructScopeSchema := schema.NewScopeSchema(&testStructSchema.ObjectSchema)
	assert.NoError(t, objectTestRef.ValidateCompatibility(testStructScopeSchema))

	// map verification
	validData := map[string]any{
		"Field1": 42,
		"field3": "Hello world!",
	}
	invalidData := map[string]any{
		"Field1": "notanint",
		"field3": "Hello world!",
	}
	validDataAndSchema := map[string]any{
		"Field1": schema.NewIntSchema(nil, nil, nil),
		"field3": schema.NewStringSchema(nil, nil, nil),
	}
	invalidDataAndSchema := map[string]any{
		"Field1": schema.NewStringSchema(nil, nil, nil),
		"field3": schema.NewStringSchema(nil, nil, nil),
	}
	assert.NoError(t, testStructSchema.ValidateCompatibility(validData))
	assert.NoError(t, testStructSchema.ValidateCompatibility(validDataAndSchema))
	assert.Error(t, testStructSchema.ValidateCompatibility(invalidData))
	assert.Error(t, testStructSchema.ValidateCompatibility(invalidDataAndSchema))

	// Two objects with the same set of property schemas are compatible.
	compatibleObjSchema := schema.NewObjectSchema(
		"random_id_string",
		testSchemaProperties,
	)
	assert.NoError(t, testStructSchema.ValidateCompatibility(compatibleObjSchema))

	// Test non-object types
	s1 := testStructSchema
	assert.Error(t, s1.ValidateCompatibility(schema.NewAnySchema()))
	assert.Error(t, s1.ValidateCompatibility(schema.NewStringSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewIntSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewBoolSchema()))
	assert.Error(t, s1.ValidateCompatibility(schema.NewListSchema(schema.NewBoolSchema(), nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewFloatSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewDisplayValue(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility("test"))
	assert.Error(t, s1.ValidateCompatibility(1))
	assert.Error(t, s1.ValidateCompatibility(1.5))
	assert.Error(t, s1.ValidateCompatibility(true))
	assert.Error(t, s1.ValidateCompatibility([]string{}))
	assert.Error(t, s1.ValidateCompatibility(map[string]any{}))
	assert.Error(t, s1.ValidateCompatibility(schema.NewStringEnumSchema(map[string]*schema.DisplayValue{})))
	assert.Error(t, s1.ValidateCompatibility(schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{}, nil)))
}
