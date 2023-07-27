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
}
