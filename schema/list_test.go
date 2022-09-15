package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestListMin(t *testing.T) {
	listType := schema.NewListType[string](
		schema.NewStringType(
			nil,
			nil,
			nil,
		),
		schema.IntPointer(2),
		nil,
	)

	assertEqual(t, *listType.Min(), int64(2))
	assertEqual(t, listType.Max(), nil)

	assertError2(t)(listType.Unserialize([]any{}))
	assertError2(t)(listType.Unserialize([]any{"foo"}))
	unserialized, err := listType.Unserialize([]any{"foo", "bar"})
	assertNoError(t, err)
	assertEqual(t, 2, len(unserialized))
	assertEqual(t, "foo", unserialized[0])
	assertEqual(t, "bar", unserialized[1])

	assertError(t, listType.Validate([]string{}))
	assertError(t, listType.Validate([]string{"foo"}))
	assertNoError(t, listType.Validate([]string{"foo", "bar"}))

	assertError2(t)(listType.Serialize([]string{}))
	assertError2(t)(listType.Serialize([]string{"foo"}))
	serialized, err := listType.Serialize([]string{"foo", "bar"})
	assertNoError(t, err)
	serializedList := serialized.([]any)
	assertEqual(t, 2, len(serializedList))
	assertEqual(t, "foo", serializedList[0].(string))
	assertEqual(t, "bar", serializedList[1].(string))
}

func TestListMax(t *testing.T) {
	listType := schema.NewListType[string](
		schema.NewStringType(
			nil,
			nil,
			nil,
		),
		nil,
		schema.IntPointer(2),
	)

	assertEqual(t, listType.Min(), nil)
	assertEqual(t, *listType.Max(), int64(2))

	assertError2(t)(listType.Unserialize([]any{"foo", "bar", "baz"}))
	unserialized, err := listType.Unserialize([]any{"foo", "bar"})
	assertNoError(t, err)
	assertEqual(t, 2, len(unserialized))
	assertEqual(t, "foo", unserialized[0])
	assertEqual(t, "bar", unserialized[1])

	assertError(t, listType.Validate([]string{"foo", "bar", "baz"}))
	assertNoError(t, listType.Validate([]string{"foo", "bar"}))

	assertError2(t)(listType.Serialize([]string{"foo", "bar", "baz"}))
	serialized, err := listType.Serialize([]string{"foo", "bar"})
	assertNoError(t, err)
	serializedList := serialized.([]any)
	assertEqual(t, 2, len(serializedList))
	assertEqual(t, "foo", serializedList[0].(string))
	assertEqual(t, "bar", serializedList[1].(string))
}

func TestListTypeID(t *testing.T) {
	assertEqual(
		t,
		(schema.NewListSchema(schema.NewStringType(nil, nil, nil), nil, nil)).TypeID(),
		schema.TypeIDList,
	)
	assertEqual(
		t,
		(schema.NewListType[string](schema.NewStringType(nil, nil, nil), nil, nil)).TypeID(),
		schema.TypeIDList,
	)
}

func TestListItemValidation(t *testing.T) {
	listType := schema.NewListType[string](
		schema.NewStringType(
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
	assertEqual(t, listType.TypedItems().TypeID(), schema.TypeIDString)

}

func TestListTypeHandling(t *testing.T) {
	listType := schema.NewListType[string](
		schema.NewStringType(
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
