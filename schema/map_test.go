package schema_test

import (
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestMapMin(t *testing.T) {
	mapType := schema.NewTypedMapSchema[string, string](
		schema.NewStringSchema(nil, nil, nil),
		schema.NewStringSchema(nil, nil, nil),
		schema.IntPointer(2),
		nil,
	)

	assert.Equals(t, *mapType.Min(), int64(2))
	assert.Equals(t, mapType.Max(), nil)

	assert.ErrorR(t)(mapType.Unserialize(map[any]any{}))
	assert.ErrorR(t)(mapType.Unserialize(map[any]any{"foo": "foo"}))
	unserialized, err := mapType.UnserializeType(map[any]any{"foo": "foo", "bar": "bar"})
	assert.NoError(t, err)
	assert.Equals(t, 2, len(unserialized))
	assert.Equals(t, "foo", unserialized["foo"])
	assert.Equals(t, "bar", unserialized["bar"])

	assert.Error(t, mapType.Validate(map[string]string{}))
	assert.Error(t, mapType.Validate(map[string]string{"foo": "foo"}))
	assert.NoError(t, mapType.Validate(map[string]string{"foo": "foo", "bar": "bar"}))

	assert.ErrorR(t)(mapType.Serialize(map[string]string{}))
	assert.ErrorR(t)(mapType.Serialize(map[string]string{"foo": "foo"}))
	serialized, err := mapType.Serialize(map[string]string{"foo": "foo", "bar": "bar"})
	assert.NoError(t, err)
	serializedMap := serialized.(map[any]any)
	assert.Equals(t, 2, len(serializedMap))
	assert.Equals(t, "foo", serializedMap["foo"].(string))
	assert.Equals(t, "bar", serializedMap["bar"].(string))
}

func TestMapMax(t *testing.T) {
	mapType := schema.NewTypedMapSchema[string, string](
		schema.NewStringSchema(nil, nil, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		schema.IntPointer(2),
	)

	assert.Equals(t, mapType.Min(), nil)
	assert.Equals(t, *mapType.Max(), int64(2))

	assert.ErrorR(t)(mapType.Unserialize(map[any]any{"foo": "foo", "bar": "bar", "baz": "baz"}))
	unserialized, err := mapType.UnserializeType(map[any]any{"foo": "foo", "bar": "bar"})
	assert.NoError(t, err)
	assert.Equals(t, 2, len(unserialized))
	assert.Equals(t, "foo", unserialized["foo"])
	assert.Equals(t, "bar", unserialized["bar"])

	assert.Error(t, mapType.Validate(map[string]string{"foo": "foo", "bar": "bar", "baz": "baz"}))
	assert.NoError(t, mapType.Validate(map[string]string{"foo": "foo", "bar": "bar"}))

	assert.ErrorR(t)(mapType.Serialize(map[string]string{"foo": "foo", "bar": "bar", "baz": "baz"}))
	serialized, err := mapType.Serialize(map[string]string{"foo": "foo", "bar": "bar"})
	assert.NoError(t, err)
	serializedMap := serialized.(map[any]any)
	assert.Equals(t, 2, len(serializedMap))
	assert.Equals(t, "foo", serializedMap["foo"].(string))
	assert.Equals(t, "bar", serializedMap["bar"].(string))
}

func TestMapSchemaID(t *testing.T) {
	assert.Equals(
		t,
		(schema.NewMapSchema(
			schema.NewStringSchema(nil, nil, nil),
			schema.NewStringSchema(nil, nil, nil),
			nil,
			nil,
		)).TypeID(),
		schema.TypeIDMap,
	)
	assert.Equals(
		t,
		(schema.NewMapSchema(
			schema.NewStringSchema(nil, nil, nil),
			schema.NewStringSchema(nil, nil, nil),
			nil,
			nil,
		)).TypeID(),
		schema.TypeIDMap,
	)
}

func TestMapItemValidation(t *testing.T) {
	mapType := schema.NewTypedMapSchema[string, string](
		schema.NewStringSchema(
			schema.IntPointer(2),
			nil,
			nil,
		),
		schema.NewStringSchema(
			schema.IntPointer(1),
			nil,
			nil,
		),
		nil,
		nil,
	)

	assert.ErrorR(t)(mapType.Unserialize(map[string]string{"a": "b"}))
	assert.ErrorR(t)(mapType.Unserialize(map[string]string{"ab": ""}))
	assert.NoErrorR[any](t)(mapType.Unserialize(map[string]string{"ab": "b"}))

	assert.Error(t, mapType.Validate(map[string]string{"a": "b"}))
	assert.Error(t, mapType.Validate(map[string]string{"ab": ""}))
	assert.NoError(t, mapType.Validate(map[string]string{"ab": "b"}))

	assert.ErrorR(t)(mapType.Serialize(map[string]string{"a": "b"}))
	assert.ErrorR(t)(mapType.Serialize(map[string]string{"ab": ""}))
	assert.NoErrorR[any](t)(mapType.Serialize(map[string]string{"ab": "b"}))

	assert.Equals(t, mapType.Keys().TypeID(), schema.TypeIDString)
	assert.Equals(t, mapType.Values().TypeID(), schema.TypeIDString)
}

func TestMapSchemaHandling(t *testing.T) {
	mapType := schema.NewMapSchema(
		schema.NewStringSchema(
			nil,
			nil,
			nil,
		),
		schema.NewStringSchema(
			nil,
			nil,
			nil,
		),
		nil,
		nil,
	)

	assert.ErrorR(t)(mapType.Unserialize(struct{}{}))
	assert.ErrorR(t)(mapType.Unserialize(map[any]any{"a": struct{}{}}))
	assert.ErrorR(t)(mapType.Unserialize(map[any]any{struct{}{}: "a"}))
}

func TestMapSchemaTypesValidation(t *testing.T) {
	s := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)

	assert.Equals(t, s.Keys().TypeID(), schema.TypeIDString)
	assert.Equals(t, s.Values().TypeID(), schema.TypeIDInt)

	s2 := schema.NewMapSchema(
		schema.NewIntSchema(nil, nil, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		nil,
	)

	assert.Equals(t, s2.Keys().TypeID(), schema.TypeIDInt)
	assert.Equals(t, s2.Values().TypeID(), schema.TypeIDString)

	s3 := schema.NewMapSchema(
		schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{1024: {NameValue: schema.PointerTo("Small")}}, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		nil,
	)

	assert.Equals(t, s3.Keys().TypeID(), schema.TypeIDIntEnum)
	assert.Equals(t, s3.Values().TypeID(), schema.TypeIDString)

	s4 := schema.NewMapSchema(
		schema.NewStringEnumSchema(map[string]*schema.DisplayValue{"s": {NameValue: schema.PointerTo("Small")}}),
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)

	assert.Equals(t, s4.Keys().TypeID(), schema.TypeIDStringEnum)
	assert.Equals(t, s4.Values().TypeID(), schema.TypeIDInt)

	func() {
		defer func() {
			assert.Error(t, recover().(error))
		}()
		schema.NewMapSchema(
			schema.NewBoolSchema(),
			schema.NewIntSchema(nil, nil, nil),
			nil,
			nil,
		)
		t.Fatalf("Bool keys did not result in an error")
	}()
}

//nolint:funlen
func TestMapSchemaCompatibilityValidation(t *testing.T) {
	s1 := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)
	// These next two have the same types, but size restrictions.
	s1small := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		schema.IntPointer(0),
		schema.IntPointer(3),
	)
	s1large := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		schema.IntPointer(4),
		schema.IntPointer(6),
	)

	// Differs in incompatible key with s1.
	s2 := schema.NewMapSchema(
		schema.NewIntSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)
	// Differs in incompatible value with s1.
	s3 := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		nil,
	)
	s4 := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewAnySchema(),
		nil,
		nil,
	)

	assert.NoError(t, s1.ValidateCompatibility(s1))           // Same
	assert.NoError(t, s2.ValidateCompatibility(s2))           // Same
	assert.NoError(t, s3.ValidateCompatibility(s3))           // Same
	assert.NoError(t, s4.ValidateCompatibility(s4))           // Same
	assert.NoError(t, s4.ValidateCompatibility(s1))           // s4-any is more general, so it will allow the more specific s1
	assert.NoError(t, s1small.ValidateCompatibility(s1small)) // Same
	assert.NoError(t, s1large.ValidateCompatibility(s1large)) // Same
	assert.NoError(t, s1.ValidateCompatibility(s1large))      // Same types. Size overlap
	assert.NoError(t, s1large.ValidateCompatibility(s1))      // Same types. Size overlap
	assert.NoError(t, s1.ValidateCompatibility(s1small))      // Same types. Size overlap
	assert.NoError(t, s1small.ValidateCompatibility(s1))      // Same types. Size overlap
	assert.Error(t, s1.ValidateCompatibility(s2))             // Incompatible keys
	assert.Error(t, s2.ValidateCompatibility(s1))             // incompatible keys
	assert.Error(t, s1.ValidateCompatibility(s3))             // Incompatible values
	assert.Error(t, s1.ValidateCompatibility(s4))             // right too general
	assert.Error(t, s1small.ValidateCompatibility(s1large))   // mutually exclusive sizes
	assert.Error(t, s1large.ValidateCompatibility(s1small))   // mutually exclusive sizes

	assert.Error(t, s1.ValidateCompatibility(schema.NewAnySchema()))
	assert.Error(t, s1.ValidateCompatibility(schema.NewStringSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewIntSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewBoolSchema()))
	assert.Error(t, s1.ValidateCompatibility(schema.NewListSchema(schema.NewBoolSchema(), nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewFloatSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewStringEnumSchema(map[string]*schema.DisplayValue{})))
	assert.Error(t, s1.ValidateCompatibility(schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{}, nil)))
}

//nolint:funlen
func TestMapCompatibilityValidation(t *testing.T) {
	s1 := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)
	s1size2 := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		schema.IntPointer(2),
		schema.IntPointer(2),
	)
	v1 := map[string]int{
		"a": 1,
	}
	v1Size2 := map[string]int{
		"a": 1,
		"b": 2,
	}
	v1Size3 := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	v2 := map[any]any{
		"a": 1,
	}
	v2bad1 := map[any]any{
		"a": 1.1,
	}
	v2bad2 := map[any]any{
		1: 1,
	}
	v3 := map[any]any{
		schema.NewStringSchema(nil, nil, nil): 1,
	}
	v3bad := map[any]any{
		schema.NewIntSchema(nil, nil, nil): 1,
	}
	v4 := map[any]any{
		schema.NewStringSchema(nil, nil, nil): schema.NewIntSchema(nil, nil, nil),
	}
	v5 := map[any]any{
		"a": schema.NewIntSchema(nil, nil, nil),
	}
	v5bad := map[any]any{
		"a": schema.NewStringSchema(nil, nil, nil),
	}

	assert.NoError(t, s1.ValidateCompatibility(v1))           // Valid map
	assert.NoError(t, s1.ValidateCompatibility(v2))           // Valid map with any type annotation
	assert.NoError(t, s1.ValidateCompatibility(v3))           // Valid value, vaid schema key
	assert.NoError(t, s1.ValidateCompatibility(v4))           // Valid schema key and value
	assert.NoError(t, s1.ValidateCompatibility(v5))           // Valid key, valid schema value
	assert.Error(t, s1.ValidateCompatibility(v2bad1))         // Invalid value
	assert.Error(t, s1.ValidateCompatibility(v2bad2))         // Invalid key
	assert.Error(t, s1.ValidateCompatibility(v3bad))          // Invalid schema key, valid value.
	assert.Error(t, s1.ValidateCompatibility(v5bad))          // Valid key, invalid schema value.
	assert.Error(t, s1size2.ValidateCompatibility(v1))        // Too small
	assert.NoError(t, s1size2.ValidateCompatibility(v1Size2)) // Matching size
	assert.Error(t, s1size2.ValidateCompatibility(v1Size3))   // Too large

	// Non-map types
	assert.Error(t, s1.ValidateCompatibility(schema.NewDisplayValue(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility("test"))
	assert.Error(t, s1.ValidateCompatibility(1))
	assert.Error(t, s1.ValidateCompatibility(1.5))
	assert.Error(t, s1.ValidateCompatibility(true))
	assert.Error(t, s1.ValidateCompatibility([]string{}))
}

func TestMap_UnserializeIdempotent(t *testing.T) {
	mapType := schema.NewTypedMapSchema[string, string](
		schema.NewStringSchema(nil, nil, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		schema.IntPointer(3),
	)
	var serializableInput any
	serializableInput = map[any]any{"foo": "foo", "bar": "bar", "baz": "baz"}
	unserialized, err := mapType.UnserializeType(serializableInput)
	assert.NoError(t, err)
	serialized, err := mapType.Serialize(unserialized)
	assert.NoError(t, err)
	assert.Equals(t, serialized, serializableInput)
}
