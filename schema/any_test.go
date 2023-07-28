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
		t.Run(name, func(t *testing.T) {
			unserialized, err := anyType.Unserialize(val.input)
			assert.NoError(t, err)
			assert.Equals(t, unserialized, val.unserialized)
			err = anyType.Validate(val.unserialized)
			assert.NoError(t, err)
			serialized, err := anyType.Serialize(val.unserialized)
			assert.NoError(t, err)
			assert.Equals(t, serialized, val.serialized)
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

func TestAnyValidateCompatibility(t *testing.T) {
	s1 := schema.NewAnySchema()

	assert.NoError(t, s1.ValidateCompatibility(schema.NewAnySchema()))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewStringSchema(nil, nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewIntSchema(nil, nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewBoolSchema()))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewListSchema(schema.NewBoolSchema(), nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewFloatSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewDisplayValue(nil, nil, nil)))
	assert.NoError(t, s1.ValidateCompatibility("test"))
	assert.NoError(t, s1.ValidateCompatibility(1))
	assert.NoError(t, s1.ValidateCompatibility(1.5))
	assert.NoError(t, s1.ValidateCompatibility(true))
	assert.NoError(t, s1.ValidateCompatibility([]string{}))
	assert.NoError(t, s1.ValidateCompatibility(map[string]any{}))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewStringEnumSchema(map[string]*schema.DisplayValue{})))
	assert.NoError(t, s1.ValidateCompatibility(schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{}, nil)))

}
