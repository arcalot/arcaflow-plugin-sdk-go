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
	// Create object with some ID, and nil scope.
	id1 := "id1"
	name := "name"
	description := "description"
	icon := "icon"
	displayVal := schema.NewDisplayValue(&name, &description, &icon)
	refSchema := schema.NewRefSchema(id1, displayVal)

	// Make sure everything else matches the original input
	assert.Equals(t, refSchema.ID(), id1)
	assert.Equals(t, refSchema.Display().Name(), displayVal.Name())
	assert.Equals(t, refSchema.Display().Icon(), displayVal.Icon())
	assert.Equals(t, refSchema.Display().Description(), displayVal.Description())

	// Create scopes for testing
	object1Schema := schema.NewObjectSchema(id1, nil)
	object2Schema := schema.NewObjectSchema("id2", nil)
	id1Scope := schema.NewScopeSchema(object1Schema)
	id2Scope := schema.NewScopeSchema(object2Schema)

	// id1Scope contains id1, which is in the refSchema.
	refSchema.ApplyScope(id1Scope)
	assert.Panics(t, func() {
		// Make sure it panics due to the id not being found in scope
		refSchema.ApplyScope(id2Scope)
	})
}
