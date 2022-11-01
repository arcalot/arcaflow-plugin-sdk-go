package schema_test

import (
	"testing"

	"go.arcalot.io/assert"
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

	assert.Equals(t, *listType.Min(), int64(2))
	assert.Equals(t, listType.Max(), nil)

	assert.ErrorR[any](t)(listType.UnserializeType([]any{}))
	assert.ErrorR[any](t)(listType.UnserializeType([]any{"foo"}))
	unserialized, err := listType.UnserializeType([]any{"foo", "bar"})
	assert.NoError(t, err)
	assert.Equals(t, 2, len(unserialized))
	assert.Equals(t, "foo", unserialized[0])
	assert.Equals(t, "bar", unserialized[1])

	assert.Error(t, listType.ValidateType([]string{}))
	assert.Error(t, listType.ValidateType([]string{"foo"}))
	assert.NoError(t, listType.ValidateType([]string{"foo", "bar"}))

	assert.ErrorR[any](t)(listType.SerializeType([]string{}))
	assert.ErrorR[any](t)(listType.SerializeType([]string{"foo"}))
	serialized, err := listType.SerializeType([]string{"foo", "bar"})
	assert.NoError(t, err)
	serializedList := serialized.([]any)
	assert.Equals(t, 2, len(serializedList))
	assert.Equals(t, "foo", serializedList[0].(string))
	assert.Equals(t, "bar", serializedList[1].(string))
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

	assert.Equals(t, listType.Min(), nil)
	assert.Equals(t, *listType.Max(), int64(2))

	assert.ErrorR[any](t)(listType.UnserializeType([]any{"foo", "bar", "baz"}))
	unserialized, err := listType.UnserializeType([]any{"foo", "bar"})
	assert.NoError(t, err)
	assert.Equals(t, 2, len(unserialized))
	assert.Equals(t, "foo", unserialized[0])
	assert.Equals(t, "bar", unserialized[1])

	assert.Error(t, listType.ValidateType([]string{"foo", "bar", "baz"}))
	assert.NoError(t, listType.ValidateType([]string{"foo", "bar"}))

	assert.ErrorR[any](t)(listType.SerializeType([]string{"foo", "bar", "baz"}))
	serialized, err := listType.SerializeType([]string{"foo", "bar"})
	assert.NoError(t, err)
	serializedList := serialized.([]any)
	assert.Equals(t, 2, len(serializedList))
	assert.Equals(t, "foo", serializedList[0].(string))
	assert.Equals(t, "bar", serializedList[1].(string))
}

func TestListTypeID(t *testing.T) {
	assert.Equals(
		t,
		(schema.NewListSchema(schema.NewStringSchema(nil, nil, nil), nil, nil)).TypeID(),
		schema.TypeIDList,
	)
	assert.Equals(
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

	assert.ErrorR[any](t)(listType.Unserialize([]string{""}))
	assert.NoErrorR[any](t)(listType.Unserialize([]string{"a"}))

	assert.Error(t, listType.Validate([]string{""}))
	assert.NoError(t, listType.Validate([]string{"a"}))

	assert.ErrorR[any](t)(listType.Serialize([]string{""}))
	assert.NoErrorR[any](t)(listType.Serialize([]string{"a"}))

	assert.Equals(t, listType.Items().TypeID(), schema.TypeIDString)

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

	assert.ErrorR[any](t)(listType.Unserialize(struct{}{}))
	assert.ErrorR[any](t)(listType.Unserialize([]any{struct{}{}}))
}
