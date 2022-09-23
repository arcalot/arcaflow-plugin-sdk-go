package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestListMin(t *testing.T) {
	listType := schema.NewTypedListSchema[string](
		schema.NewStringSchema(
			nil,
			nil,
			nil,
		),
		schema.IntPointer(2),
		nil,
	)

	assertEqual(t, *listType.Min(), int64(2))
	assertEqual(t, listType.Max(), nil)

	assertError2(t)(listType.UnserializeType([]any{}))
	assertError2(t)(listType.UnserializeType([]any{"foo"}))
	unserialized, err := listType.UnserializeType([]any{"foo", "bar"})
	assertNoError(t, err)
	assertEqual(t, 2, len(unserialized))
	assertEqual(t, "foo", unserialized[0])
	assertEqual(t, "bar", unserialized[1])

	assertError(t, listType.ValidateType([]string{}))
	assertError(t, listType.ValidateType([]string{"foo"}))
	assertNoError(t, listType.ValidateType([]string{"foo", "bar"}))

	assertError2(t)(listType.SerializeType([]string{}))
	assertError2(t)(listType.SerializeType([]string{"foo"}))
	serialized, err := listType.SerializeType([]string{"foo", "bar"})
	assertNoError(t, err)
	serializedList := serialized.([]any)
	assertEqual(t, 2, len(serializedList))
	assertEqual(t, "foo", serializedList[0].(string))
	assertEqual(t, "bar", serializedList[1].(string))
}

func TestListMax(t *testing.T) {
	listType := schema.NewTypedListSchema[string](
		schema.NewStringSchema(
			nil,
			nil,
			nil,
		),
		nil,
		schema.IntPointer(2),
	)

	assertEqual(t, listType.Min(), nil)
	assertEqual(t, *listType.Max(), int64(2))

	assertError2(t)(listType.UnserializeType([]any{"foo", "bar", "baz"}))
	unserialized, err := listType.UnserializeType([]any{"foo", "bar"})
	assertNoError(t, err)
	assertEqual(t, 2, len(unserialized))
	assertEqual(t, "foo", unserialized[0])
	assertEqual(t, "bar", unserialized[1])

	assertError(t, listType.ValidateType([]string{"foo", "bar", "baz"}))
	assertNoError(t, listType.ValidateType([]string{"foo", "bar"}))

	assertError2(t)(listType.SerializeType([]string{"foo", "bar", "baz"}))
	serialized, err := listType.SerializeType([]string{"foo", "bar"})
	assertNoError(t, err)
	serializedList := serialized.([]any)
	assertEqual(t, 2, len(serializedList))
	assertEqual(t, "foo", serializedList[0].(string))
	assertEqual(t, "bar", serializedList[1].(string))
}

func TestListTypeID(t *testing.T) {
	assertEqual(
		t,
		(schema.NewListSchema(schema.NewStringSchema(nil, nil, nil), nil, nil)).TypeID(),
		schema.TypeIDList,
	)
	assertEqual(
		t,
		(schema.NewTypedListSchema[string](schema.NewStringSchema(nil, nil, nil), nil, nil)).TypeID(),
		schema.TypeIDList,
	)
}

func TestListItemValidation(t *testing.T) {
	listType := schema.NewTypedListSchema[string](
		schema.NewStringSchema(
			schema.IntPointer(1),
			nil,
			nil,
		),
		nil,
		nil,
	)

	assertError2(t)(listType.Unserialize([]string{""}))
	assertNoError2(t)(listType.Unserialize([]string{"a"}))

	assertError(t, listType.Validate([]string{""}))
	assertNoError(t, listType.Validate([]string{"a"}))

	assertError2(t)(listType.Serialize([]string{""}))
	assertNoError2(t)(listType.Serialize([]string{"a"}))

	assertEqual(t, listType.Items().TypeID(), schema.TypeIDString)

}

func TestListTypeHandling(t *testing.T) {
	listType := schema.NewTypedListSchema[string](
		schema.NewStringSchema(
			nil,
			nil,
			nil,
		),
		nil,
		nil,
	)

	assertError2(t)(listType.Unserialize(struct{}{}))
	assertError2(t)(listType.Unserialize([]any{struct{}{}}))
}
