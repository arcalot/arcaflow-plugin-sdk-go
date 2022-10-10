package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

//nolint:dupl
func TestPropertySchemaParameters(t *testing.T) {
	propertySchema := schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(
			schema.PointerTo("Greeting"),
			schema.PointerTo("Hello world!"),
			schema.PointerTo("<svg></svg>"),
		),
		true,
		[]string{"somefield1"},
		[]string{"somefield2"},
		[]string{"somefield3"},
		schema.PointerTo(`"Hello world!"`),
		[]string{`"Hello world!"`},
	)

	assertEqual(t, propertySchema.Type().TypeID(), schema.TypeIDString)
	assertEqual(t, *(propertySchema.Display().Name()), "Greeting")
	assertEqual(t, propertySchema.Required(), true)
	assertEqual(t, propertySchema.RequiredIf()[0], "somefield1")
	assertEqual(t, propertySchema.RequiredIfNot()[0], "somefield2")
	assertEqual(t, propertySchema.Conflicts()[0], "somefield3")
	assertEqual(t, propertySchema.Examples()[0], `"Hello world!"`)
	assertEqual(t, *propertySchema.Default(), `"Hello world!"`)
}

//nolint:dupl
func TestPropertyTypeParameters(t *testing.T) {
	propertySchema := schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.NewDisplayValue(
			schema.PointerTo("Greeting"),
			schema.PointerTo("Hello world!"),
			schema.PointerTo("<svg></svg>"),
		),
		true,
		[]string{"somefield1"},
		[]string{"somefield2"},
		[]string{"somefield3"},
		schema.PointerTo(`"Hello world!"`),
		[]string{`"Hello world!"`},
	)

	assertEqual(t, propertySchema.Type().TypeID(), schema.TypeIDString)
	assertEqual(t, *(propertySchema.Display().Name()), "Greeting")
	assertEqual(t, propertySchema.Required(), true)
	assertEqual(t, propertySchema.RequiredIf()[0], "somefield1")
	assertEqual(t, propertySchema.RequiredIfNot()[0], "somefield2")
	assertEqual(t, propertySchema.Conflicts()[0], "somefield3")
	assertEqual(t, propertySchema.Examples()[0], `"Hello world!"`)
	assertEqual(t, *propertySchema.Default(), `"Hello world!"`)
}

func TestPropertyTypeTypeID(t *testing.T) {
	propertyType := schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		false,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	assertEqual(t, propertyType.TypeID(), schema.TypeIDString)
}

func TestPropertyTypeInvalidTypes(t *testing.T) {
	propertyType := schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		false,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	assertError(t, propertyType.Validate(struct{}{}))
	assertError2(t)(propertyType.Serialize(struct{}{}))
}

func TestPropertyEmptyAsDefault(t *testing.T) {
	type tString string
	type TestData struct {
		Foo tString `json:"foo"`
	}

	s := schema.NewStructMappedObjectSchema[TestData](
		"TestData",
		map[string]*schema.PropertySchema{
			"foo": schema.NewPropertySchema(
				// We force validation on string length.
				schema.NewStringSchema(schema.IntPointer(1), nil, nil),
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
			).TreatEmptyAsDefaultValue(),
		},
	)

	// Here we pass an empty struct, setting the string to the default value.
	data, err := s.Serialize(TestData{})
	assertNoError(t, err)
	assertEqual(t, len(data.(map[string]any)), 0)

	assertNoError(t, s.Validate(TestData{}))

	_, err = s.Unserialize(map[string]any{
		"foo": "",
	})
	assertError(t, err)
}
