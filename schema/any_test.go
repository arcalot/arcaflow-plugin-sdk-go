package schema_test

import (
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

//nolint:funlen
func TestAny(t *testing.T) {
	validValues := map[string]struct {
		input        any
		unserialized any
		serialized   any
	}{
		"bool": {
			true,
			true,
			true,
		},
		"int": {
			1,
			int64(1),
			int64(1),
		},
		"uint": {
			uint(1),
			int64(1),
			int64(1),
		},
		"int8": {
			int8(1),
			int64(1),
			int64(1),
		},
		"uint8": {
			uint8(1),
			int64(1),
			int64(1),
		},
		"int16": {
			int16(1),
			int64(1),
			int64(1),
		},
		"uint16": {
			uint16(1),
			int64(1),
			int64(1),
		},
		"int32": {
			int32(1),
			int64(1),
			int64(1),
		},
		"uint32": {
			uint32(1),
			int64(1),
			int64(1),
		},
		"int64": {
			int64(1),
			int64(1),
			int64(1),
		},
		"uint64": {
			uint64(1),
			int64(1),
			int64(1),
		},
		"float32": {
			float32(1),
			float64(1),
			float64(1),
		},
		"float64": {
			float64(1),
			float64(1),
			float64(1),
		},
		"map": {
			map[any]any{
				1:      "test",
				"test": 1,
			},
			map[any]any{
				int64(1): "test",
				"test":   int64(1),
			},
			map[any]any{
				int64(1): "test",
				"test":   int64(1),
			},
		},
		"slice": {
			[]any{
				"test",
				1,
			},
			[]any{
				"test",
				int64(1),
			},
			[]any{
				"test",
				int64(1),
			},
		},
	}

	anyType := schema.NewAnySchema()
	for name, val := range validValues {
		testCase := val
		t.Run(name, func(t *testing.T) {
			unserialized, err := anyType.Unserialize(testCase.input)
			assert.NoError(t, err)
			assert.Equals(t, unserialized, testCase.unserialized)
			err = anyType.Validate(testCase.unserialized)
			assert.NoError(t, err)
			serialized, err := anyType.Serialize(testCase.unserialized)
			assert.NoError(t, err)
			assert.Equals(t, serialized, testCase.serialized)

			unserialized2, err := anyType.Unserialize(serialized)
			assert.NoError(t, err)
			// test unserialize and serialize are reversible
			assert.Equals(t, unserialized2, unserialized)
		})
	}

	invalidValues := map[string]any{
		"struct": struct{}{},
		"map of struct": map[string]struct{}{
			"test": {},
		},
	}
	for name, val := range invalidValues {
		t.Run(name, func(t *testing.T) {
			_, err := anyType.Unserialize(val)
			assert.Error(t, err)
			err = anyType.Validate(val)
			assert.Error(t, err)
			_, err = anyType.Serialize(val)
			assert.Error(t, err)
		})
	}
}

func TestAnyTypeReflectedType(t *testing.T) {
	a := schema.NewAnySchema()
	assert.NotNil(t, a.ReflectedType())
}

var properties = map[string]*schema.PropertySchema{
	"field1": schema.NewPropertySchema(
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

type someStruct struct {
	field1 int
}

var objectSchema = schema.NewObjectSchema("some-id", properties)
var structMappedObjectSchema = schema.NewStructMappedObjectSchema[someStruct]("some-id", properties)

func TestAnyValidateCompatibilitySimple(t *testing.T) {
	s1 := schema.NewAnySchema()
	assert.NoError(t, s1.ValidateCompatibility(schema.NewAnySchema()))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewStringSchema(nil, nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewIntSchema(nil, nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewBoolSchema()))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewFloatSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewDisplayValue(nil, nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility("test"))
	assert.NoError(t, s1.ValidateCompatibility(1))
	assert.NoError(t, s1.ValidateCompatibility(1.5))
	assert.NoError(t, s1.ValidateCompatibility(true))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewStringEnumSchema(map[string]*schema.DisplayValue{})))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{}, nil)))
	assert.NoError(t, s1.ValidateCompatibility(objectSchema))
	// Test struct mapped since it may have a different reflected type.
	assert.NoError(t, s1.ValidateCompatibility(structMappedObjectSchema))
	assert.NoError(t, s1.ValidateCompatibility(
		schema.NewOneOfStringSchema[string](map[string]schema.Object{}, "id", false),
	))
	assert.NoError(t, s1.ValidateCompatibility(
		schema.NewOneOfIntSchema[int64](map[int64]schema.Object{}, "id", false),
	))

}

func TestAnyValidateCompatibilityLists(t *testing.T) {
	s1 := schema.NewAnySchema()
	assert.NoError(t, s1.ValidateCompatibility(schema.NewListSchema(schema.NewBoolSchema(), nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility([]string{}))

	// Test non-homogeneous list
	err := s1.ValidateCompatibility([]any{
		int64(5),
		"5",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "homogeneous")
	// Test a list of object schemas
	assert.NoError(t, s1.ValidateCompatibility([]any{
		structMappedObjectSchema,
	}))
}

func TestAnyValidateCompatibilityMaps(t *testing.T) {
	// Test custom maps with schemas and data
	s1 := schema.NewAnySchema()
	assert.NoError(t, s1.ValidateCompatibility(map[string]any{}))
	// Include invalid item within an any map
	err := s1.ValidateCompatibility(map[any]any{
		"b": someStruct{field1: 1},
	})
	assert.Error(t, err)
	// Include invalid item within a string map
	err = s1.ValidateCompatibility(map[string]any{
		"b": someStruct{field1: 1},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `"b"`)        // Identifies the problematic key
	assert.Contains(t, err.Error(), "someStruct") // Identifies the problematic type
	// String key type
	assert.NoError(t, s1.ValidateCompatibility(map[string]any{
		"a": true,
		"b": "test",
		"c": []any{
			structMappedObjectSchema,
		},
		"d": structMappedObjectSchema,
		"e": schema.NewStringSchema(nil, nil, nil),
	}))
	// int key type
	assert.NoError(t, s1.ValidateCompatibility(map[int64]any{
		1: true,
		2: "test",
		3: []any{
			structMappedObjectSchema,
		},
		4: structMappedObjectSchema,
		5: schema.NewStringSchema(nil, nil, nil),
	}))
	// any key type with string key values
	assert.NoError(t, s1.ValidateCompatibility(map[any]any{
		"a": true,
		"b": "test",
		"c": []any{
			structMappedObjectSchema,
		},
		"d": structMappedObjectSchema,
		"e": schema.NewStringSchema(nil, nil, nil),
	}))
	// any key type with integer key values
	assert.NoError(t, s1.ValidateCompatibility(map[any]any{
		int64(1): true,
		int64(2): "test",
		int64(3): []any{
			structMappedObjectSchema,
		},
		int64(4): structMappedObjectSchema,
		int64(5): schema.NewStringSchema(nil, nil, nil),
	}))
	// any key type with mixed key values
	err = s1.ValidateCompatibility(map[any]any{
		"a":      true,
		int64(2): "test",
		int64(3): []any{
			structMappedObjectSchema,
		},
		int64(4): structMappedObjectSchema,
		int64(5): schema.NewStringSchema(nil, nil, nil),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched")
	assert.Contains(t, err.Error(), "string")
	assert.Contains(t, err.Error(), "int64")
}
