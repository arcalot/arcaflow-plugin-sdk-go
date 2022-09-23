package schema_test

import (
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

	assertEqual(t, *mapType.Min(), int64(2))
	assertEqual(t, mapType.Max(), nil)

	assertError2(t)(mapType.Unserialize(map[any]any{}))
	assertError2(t)(mapType.Unserialize(map[any]any{"foo": "foo"}))
	unserialized, err := mapType.UnserializeType(map[any]any{"foo": "foo", "bar": "bar"})
	assertNoError(t, err)
	assertEqual(t, 2, len(unserialized))
	assertEqual(t, "foo", unserialized["foo"])
	assertEqual(t, "bar", unserialized["bar"])

	assertError(t, mapType.Validate(map[string]string{}))
	assertError(t, mapType.Validate(map[string]string{"foo": "foo"}))
	assertNoError(t, mapType.Validate(map[string]string{"foo": "foo", "bar": "bar"}))

	assertError2(t)(mapType.Serialize(map[string]string{}))
	assertError2(t)(mapType.Serialize(map[string]string{"foo": "foo"}))
	serialized, err := mapType.Serialize(map[string]string{"foo": "foo", "bar": "bar"})
	assertNoError(t, err)
	serializedMap := serialized.(map[any]any)
	assertEqual(t, 2, len(serializedMap))
	assertEqual(t, "foo", serializedMap["foo"].(string))
	assertEqual(t, "bar", serializedMap["bar"].(string))
}

func TestMapMax(t *testing.T) {
	mapType := schema.NewTypedMapSchema[string, string](
		schema.NewStringSchema(nil, nil, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		schema.IntPointer(2),
	)

	assertEqual(t, mapType.Min(), nil)
	assertEqual(t, *mapType.Max(), int64(2))

	assertError2(t)(mapType.Unserialize(map[any]any{"foo": "foo", "bar": "bar", "baz": "baz"}))
	unserialized, err := mapType.UnserializeType(map[any]any{"foo": "foo", "bar": "bar"})
	assertNoError(t, err)
	assertEqual(t, 2, len(unserialized))
	assertEqual(t, "foo", unserialized["foo"])
	assertEqual(t, "bar", unserialized["bar"])

	assertError(t, mapType.Validate(map[string]string{"foo": "foo", "bar": "bar", "baz": "baz"}))
	assertNoError(t, mapType.Validate(map[string]string{"foo": "foo", "bar": "bar"}))

	assertError2(t)(mapType.Serialize(map[string]string{"foo": "foo", "bar": "bar", "baz": "baz"}))
	serialized, err := mapType.Serialize(map[string]string{"foo": "foo", "bar": "bar"})
	assertNoError(t, err)
	serializedMap := serialized.(map[any]any)
	assertEqual(t, 2, len(serializedMap))
	assertEqual(t, "foo", serializedMap["foo"].(string))
	assertEqual(t, "bar", serializedMap["bar"].(string))
}

func TestMapSchemaID(t *testing.T) {
	assertEqual(
		t,
		(schema.NewMapSchema(
			schema.NewStringSchema(nil, nil, nil),
			schema.NewStringSchema(nil, nil, nil),
			nil,
			nil,
		)).TypeID(),
		schema.TypeIDMap,
	)
	assertEqual(
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

	assertError2(t)(mapType.Unserialize(map[string]string{"a": "b"}))
	assertError2(t)(mapType.Unserialize(map[string]string{"ab": ""}))
	assertNoError2(t)(mapType.Unserialize(map[string]string{"ab": "b"}))

	assertError(t, mapType.Validate(map[string]string{"a": "b"}))
	assertError(t, mapType.Validate(map[string]string{"ab": ""}))
	assertNoError(t, mapType.Validate(map[string]string{"ab": "b"}))

	assertError2(t)(mapType.Serialize(map[string]string{"a": "b"}))
	assertError2(t)(mapType.Serialize(map[string]string{"ab": ""}))
	assertNoError2(t)(mapType.Serialize(map[string]string{"ab": "b"}))

	assertEqual(t, mapType.Keys().TypeID(), schema.TypeIDString)
	assertEqual(t, mapType.Values().TypeID(), schema.TypeIDString)
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

	assertError2(t)(mapType.Unserialize(struct{}{}))
	assertError2(t)(mapType.Unserialize(map[any]any{"a": struct{}{}}))
	assertError2(t)(mapType.Unserialize(map[any]any{struct{}{}: "a"}))
}

func TestMapSchemaTypesValidation(t *testing.T) {
	s := schema.NewMapSchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)

	assertEqual(t, s.Keys().TypeID(), schema.TypeIDString)
	assertEqual(t, s.Values().TypeID(), schema.TypeIDInt)

	s2 := schema.NewMapSchema(
		schema.NewIntSchema(nil, nil, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		nil,
	)

	assertEqual(t, s2.Keys().TypeID(), schema.TypeIDInt)
	assertEqual(t, s2.Values().TypeID(), schema.TypeIDString)

	s3 := schema.NewMapSchema(
		schema.NewIntEnumSchema(map[int64]string{1024: "Small"}, nil),
		schema.NewStringSchema(nil, nil, nil),
		nil,
		nil,
	)

	assertEqual(t, s3.Keys().TypeID(), schema.TypeIDIntEnum)
	assertEqual(t, s3.Values().TypeID(), schema.TypeIDString)

	s4 := schema.NewMapSchema(
		schema.NewStringEnumSchema(map[string]string{"s": "Small"}),
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)

	assertEqual(t, s4.Keys().TypeID(), schema.TypeIDStringEnum)
	assertEqual(t, s4.Values().TypeID(), schema.TypeIDInt)

	func() {
		defer func() {
			assertError(t, recover().(error))
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
