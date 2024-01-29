package schema_test

import (
	"encoding/json"
	"fmt"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func ExampleNewStringEnumSchema() {
	// Create a new enum type by defining its valid values:
	portionSize := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"small": {NameValue: schema.PointerTo("Small")},
		"large": {NameValue: schema.PointerTo("Large")},
	})

	// You can now print the valid values:
	fmt.Println(*portionSize.ValidValues()["large"].NameValue)
	// Output: Large
}

func ExampleStringEnumSchema_Unserialize() {
	portionSize := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"small": {NameValue: schema.PointerTo("Small")},
		"large": {NameValue: schema.PointerTo("Large")},
	})

	// Try to unserialize an invalid value:
	_, err := portionSize.Unserialize("")
	fmt.Println(err)

	// Unserialize a valid value:
	val, err := portionSize.Unserialize("small")
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	// Output: Validation failed: '' is not a valid value, must be one of: 'large', 'small'
	// small
}

var testStringEnumSerializationDataSet = map[string]serializationTestCase[string]{
	"validString": {
		"small",
		false,
		"small",
		"small",
	},
	"invalidString": {
		"xs",
		true,
		"small",
		"small",
	},
	"invalidType": {
		struct{}{},
		true,
		"small",
		"small",
	},
}

func TestStringEnumSerialization(t *testing.T) {
	t.Parallel()
	performSerializationTest[string](
		t,
		schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
			"small": {NameValue: schema.PointerTo("Small")},
			"large": {NameValue: schema.PointerTo("Large")},
		}),
		testStringEnumSerializationDataSet,
		func(a string, b string) bool {
			return a == b
		},
		func(a any, b any) bool {
			return a == b
		},
	)
}

func TestStringEnumTypedSerialization(t *testing.T) {
	type Size string
	s := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"small": {NameValue: schema.PointerTo("Small")},
		"large": {NameValue: schema.PointerTo("Large")},
	})
	serializedData, err := s.Serialize(Size("small"))
	assert.NoError(t, err)
	assert.Equals(t, serializedData.(string), "small")
}

func TestStringEnumJSONMarshal(t *testing.T) {
	typeUnderTest := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"small": {NameValue: schema.PointerTo("Small")},
		"large": {NameValue: schema.PointerTo("Large")},
	})

	marshalled, err := json.Marshal(typeUnderTest)
	if err != nil {
		t.Fatal(err)
	}
	if string(marshalled) != `{"values":{"large":{"name":"Large","description":null,"icon":null},"small":{"name":"Small","description":null,"icon":null}}}` {
		t.Fatalf("Invalid marshalled JSON output: %s", marshalled)
	}
	typeUnderTest = schema.NewStringEnumSchema(map[string]*schema.DisplayValue{})
	if err := json.Unmarshal(marshalled, &typeUnderTest); err != nil {
		t.Fatal(err)
	}
	if *typeUnderTest.ValidValues()["small"].NameValue != "Small" {
		t.Fatalf("Unmarshalling failed.")
	}
}

func TestStringEnumType(t *testing.T) {
	assert.Equals(t, schema.NewStringEnumSchema(map[string]*schema.DisplayValue{}).TypeID(), schema.TypeIDStringEnum)
}

func TestStringEnumCompatibilityValidation(t *testing.T) {
	portionSize := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"small": {NameValue: schema.PointerTo("Small")},
		"large": {NameValue: schema.PointerTo("Large")},
	})
	assert.NoError(t, portionSize.ValidateCompatibility("small"))
	assert.Error(t, portionSize.ValidateCompatibility("wrong"))
}

func TestStringEnumSchemaCompatibilityValidation(t *testing.T) {
	s1 := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"a": {NameValue: schema.PointerTo("a")},
		"b": {NameValue: schema.PointerTo("b")},
		"c": {NameValue: schema.PointerTo("c")},
	})
	s2 := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"d": {NameValue: schema.PointerTo("a")},
		"e": {NameValue: schema.PointerTo("b")},
		"f": {NameValue: schema.PointerTo("c")},
	})
	S1 := schema.NewStringEnumSchema(map[string]*schema.DisplayValue{
		"a": {NameValue: schema.PointerTo("A")},
		"b": {NameValue: schema.PointerTo("B")},
		"c": {NameValue: schema.PointerTo("C")},
	})

	assert.NoError(t, s1.ValidateCompatibility(s1))
	assert.NoError(t, s2.ValidateCompatibility(s2))
	// Mismatched keys
	assert.Error(t, s1.ValidateCompatibility(s2))
	assert.Error(t, s2.ValidateCompatibility(s1))
	// Mismatched names
	assert.Error(t, s1.ValidateCompatibility(S1))
	assert.Error(t, S1.ValidateCompatibility(s1))
}
