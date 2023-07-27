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
