package schema_test

import (
	"go.arcalot.io/assert"
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

	assert.Equals(t, *listType.Min(), int64(2))
	assert.Equals(t, listType.Max(), nil)

	assert.ErrorR(t)(listType.UnserializeType([]any{}))
	assert.ErrorR(t)(listType.UnserializeType([]any{"foo"}))
	unserialized, err := listType.UnserializeType([]any{"foo", "bar"})
	assert.NoError(t, err)
	assert.Equals(t, 2, len(unserialized))
	assert.Equals(t, "foo", unserialized[0])
	assert.Equals(t, "bar", unserialized[1])

	assert.Error(t, listType.ValidateType([]string{}))
	assert.Error(t, listType.ValidateType([]string{"foo"}))
	assert.NoError(t, listType.ValidateType([]string{"foo", "bar"}))

	assert.ErrorR(t)(listType.SerializeType([]string{}))
	assert.ErrorR(t)(listType.SerializeType([]string{"foo"}))
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

	assert.ErrorR(t)(listType.UnserializeType([]any{"foo", "bar", "baz"}))
	unserialized, err := listType.UnserializeType([]any{"foo", "bar"})
	assert.NoError(t, err)
	assert.Equals(t, 2, len(unserialized))
	assert.Equals(t, "foo", unserialized[0])
	assert.Equals(t, "bar", unserialized[1])

	assert.Error(t, listType.ValidateType([]string{"foo", "bar", "baz"}))
	assert.NoError(t, listType.ValidateType([]string{"foo", "bar"}))

	assert.ErrorR(t)(listType.SerializeType([]string{"foo", "bar", "baz"}))
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

	assert.ErrorR(t)(listType.Unserialize([]string{""}))
	assert.NoErrorR[any](t)(listType.Unserialize([]string{"a"}))

	assert.Error(t, listType.Validate([]string{""}))
	assert.NoError(t, listType.Validate([]string{"a"}))

	assert.ErrorR(t)(listType.Serialize([]string{""}))
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

	assert.ErrorR(t)(listType.Unserialize(struct{}{}))
	assert.ErrorR(t)(listType.Unserialize([]any{struct{}{}}))
}

func TestListVerifyCompatibility(t *testing.T) {
	typedStrListSchema := schema.NewTypedListSchema[string](
		schema.NewStringSchema(nil, nil, nil),
		nil,
		nil,
	)
	standardStrListSchema := schema.NewListSchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		nil,
	)
	intListSchema := schema.NewListSchema(
		schema.NewIntSchema(nil, nil, nil),
		nil,
		nil,
	)
	// Verify list schemas with themselves
	assert.NoError(t, typedStrListSchema.ValidateCompatibility(typedStrListSchema))
	assert.NoError(t, standardStrListSchema.ValidateCompatibility(standardStrListSchema))
	assert.NoError(t, standardStrListSchema.ValidateCompatibility(typedStrListSchema))
	// Incompatible string instead of int
	assert.Error(t, standardStrListSchema.ValidateCompatibility(intListSchema))
	assert.Error(t, intListSchema.ValidateCompatibility(standardStrListSchema))
	assert.Error(t, typedStrListSchema.ValidateCompatibility(intListSchema))
	// Test a lot of non-list types and schemas
	s1 := standardStrListSchema
	assert.Error(t, s1.ValidateCompatibility(schema.NewAnySchema()))
	assert.Error(t, s1.ValidateCompatibility(schema.NewStringSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewIntSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewBoolSchema()))
	assert.Error(t, s1.ValidateCompatibility(schema.NewListSchema(schema.NewBoolSchema(), nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewFloatSchema(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility(schema.NewDisplayValue(nil, nil, nil)))
	assert.Error(t, s1.ValidateCompatibility("test"))
	assert.Error(t, s1.ValidateCompatibility(1))
	assert.Error(t, s1.ValidateCompatibility(1.5))
	assert.Error(t, s1.ValidateCompatibility(true))
	assert.Error(t, s1.ValidateCompatibility(map[string]any{}))
	assert.Error(t, s1.ValidateCompatibility(schema.NewStringEnumSchema(map[string]*schema.DisplayValue{})))
	assert.Error(t, s1.ValidateCompatibility(schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{}, nil)))
	// Test list literals
	// Note: It doesn't check the actual list type. Just value types.
	assert.NoError(t, standardStrListSchema.ValidateCompatibility([]string{"a"}))
	assert.NoError(t, standardStrListSchema.ValidateCompatibility([]any{"a"}))
	assert.Error(t, intListSchema.ValidateCompatibility([]any{"a"}))
	assert.NoError(t, intListSchema.ValidateCompatibility([]int{1}))
	assert.Error(t, standardStrListSchema.ValidateCompatibility([]int{1}))
	// Test list of schemas
	assert.NoError(t, standardStrListSchema.ValidateCompatibility([]any{schema.NewStringSchema(nil, nil, nil)}))
	assert.Error(t, standardStrListSchema.ValidateCompatibility([]any{schema.NewIntSchema(nil, nil, nil)}))
	assert.NoError(t, intListSchema.ValidateCompatibility([]any{schema.NewIntSchema(nil, nil, nil)}))
	assert.Error(t, intListSchema.ValidateCompatibility([]any{schema.NewStringSchema(nil, nil, nil)}))
}
