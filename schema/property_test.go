package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestPropertySchemaParameters(t *testing.T) {
	propertySchema := schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		schema.PointerTo(schema.NewDisplayValue(
			schema.PointerTo("Greeting"),
			schema.PointerTo("Hello world!"),
			schema.PointerTo("<svg></svg>"),
		)),
		true,
		[]string{"somefield1"},
		[]string{"somefield2"},
		[]string{"somefield3"},
		schema.PointerTo(`"Hello world!"`),
		[]string{`"Hello world!"`},
	)

	assertEqual(t, propertySchema.Type().TypeID(), schema.TypeIDString)
	assertEqual(t, *(*propertySchema.Display()).Name(), "Greeting")
	assertEqual(t, propertySchema.Required(), true)
	assertEqual(t, propertySchema.RequiredIf()[0], "somefield1")
	assertEqual(t, propertySchema.RequiredIfNot()[0], "somefield2")
	assertEqual(t, propertySchema.Conflicts()[0], "somefield3")
	assertEqual(t, propertySchema.Examples()[0], `"Hello world!"`)
	assertEqual(t, *propertySchema.Default(), `"Hello world!"`)
}

func TestPropertyTypeParameters(t *testing.T) {
	propertySchema := schema.NewPropertyType[string](
		schema.NewStringType(nil, nil, nil),
		schema.PointerTo(schema.NewDisplayValue(
			schema.PointerTo("Greeting"),
			schema.PointerTo("Hello world!"),
			schema.PointerTo("<svg></svg>"),
		)),
		true,
		[]string{"somefield1"},
		[]string{"somefield2"},
		[]string{"somefield3"},
		schema.PointerTo(`"Hello world!"`),
		[]string{`"Hello world!"`},
	)

	assertEqual(t, propertySchema.Type().TypeID(), schema.TypeIDString)
	assertEqual(t, *(*propertySchema.Display()).Name(), "Greeting")
	assertEqual(t, propertySchema.Required(), true)
	assertEqual(t, propertySchema.RequiredIf()[0], "somefield1")
	assertEqual(t, propertySchema.RequiredIfNot()[0], "somefield2")
	assertEqual(t, propertySchema.Conflicts()[0], "somefield3")
	assertEqual(t, propertySchema.Examples()[0], `"Hello world!"`)
	assertEqual(t, *propertySchema.Default(), `"Hello world!"`)
}

func TestPropertyTypeTypeID(t *testing.T) {
	propertyType := schema.NewPropertyType[string](
		schema.NewStringType(nil, nil, nil),
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
	propertyType := schema.NewPropertyType[string](
		schema.NewStringType(nil, nil, nil),
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

func TestPropertyTypeSerialization(t *testing.T) {
	performSerializationTest[any](
		t,
		schema.NewPropertyType[string](
			schema.NewStringType(nil, nil, nil),
			nil,
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
		),
		map[string]serializationTestCase[any]{
			"emptyString": {
				SerializedValue:         "",
				ExpectUnserializedValue: "",
				ExpectedSerializedValue: "",
			},
			"incorrectType": {
				SerializedValue: struct{}{},
				ExpectError:     true,
			},
		},
		func(a any, b any) bool {
			return a == b
		},
		func(a any, b any) bool {
			return a == b
		},
	)
}
