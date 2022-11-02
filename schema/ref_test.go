package schema_test

import (
	"testing"

	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestMissingObjectCache(t *testing.T) {
	var refSchema = schema.NewRefSchema("id", nil)

	assert.Panics(t, func() {
		refSchema.Properties()
	})
	assert.Panics(t, func() {
		refSchema.GetDefaults()
	})
	assert.Panics(t, func() {
		refSchema.GetObject()
	})
	assert.Panics(t, func() {
		refSchema.ReflectedType()
	})
	assert.Panics(t, func() {
		// The cache check comes before any use of the param
		refSchema.Unserialize(nil)
	})
	assert.Panics(t, func() {
		// The cache check comes before any use of the param
		refSchema.Validate(nil)
	})
	assert.Panics(t, func() {
		// The cache check comes before any use of the param
		refSchema.Serialize(nil)
	})
}

func TestMissingObject(t *testing.T) {

}
