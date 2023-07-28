package schema_test

import (
	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
	"testing"
)

func TestIncompatibleEnumTypeValidation(t *testing.T) {
	stringEnumSchema := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"a": {NameValue: schema.PointerTo("a")},
		"b": {NameValue: schema.PointerTo("b")},
		"c": {NameValue: schema.PointerTo("c")},
	})
	intEnumSchema := schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		1: {NameValue: schema.PointerTo("a")},
		2: {NameValue: schema.PointerTo("b")},
		3: {NameValue: schema.PointerTo("c")},
	}, nil)

	// Make sure self-validation is working correctly
	assert.NoError(t, intEnumSchema.ValidateCompatibility(intEnumSchema))
	assert.NoError(t, stringEnumSchema.ValidateCompatibility(stringEnumSchema))
	// Mismatched names
	assert.Error(t, intEnumSchema.ValidateCompatibility(stringEnumSchema))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(intEnumSchema))

	// Not enums
	assert.Error(t, stringEnumSchema.ValidateCompatibility(schema.NewAnySchema()))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(schema.NewStringSchema(nil, nil, nil)))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(schema.NewIntSchema(nil, nil, nil)))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(schema.NewBoolSchema()))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(schema.NewListSchema(schema.NewBoolSchema(), nil, nil)))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(schema.NewFloatSchema(nil, nil, nil)))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(schema.NewDisplayValue(nil, nil, nil)))
	assert.Error(t, stringEnumSchema.ValidateCompatibility("test"))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(1))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(1.5))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(true))
	assert.Error(t, stringEnumSchema.ValidateCompatibility([]string{}))
	assert.Error(t, stringEnumSchema.ValidateCompatibility(map[string]any{}))
}
