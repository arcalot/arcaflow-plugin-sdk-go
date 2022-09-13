package schema_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func ExampleNewStringEnumType() {
	// Create a new enum type by defining its valid values:
	portionSize := schema.NewStringEnumType(map[string]string{
		"small": "Small",
		"large": "Large",
	})

	// You can now print the valid values:
	fmt.Println(portionSize.ValidValues())
	// Output: map[large:Large small:Small]
}

func ExampleStringEnumType_unserialize() {
	portionSize := schema.NewStringEnumType(map[string]string{
		"small": "Small",
		"large": "Large",
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
	performSerializationTest[string](
		t,
		schema.NewStringEnumType(map[string]string{
			"small": "Small",
			"large": "Large",
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

func TestStringEnumJSONMarshal(t *testing.T) {
	typeUnderTest := schema.NewStringEnumType(map[string]string{
		"small": "Small",
		"large": "Large",
	})

	marshalled, err := json.Marshal(typeUnderTest)
	if err != nil {
		t.Fatal(err)
	}
	if string(marshalled) != `{"valid_values":{"large":"Large","small":"Small"}}` {
		t.Fatalf("Invalid marshalled JSON output: %s", marshalled)
	}
	typeUnderTest = schema.NewStringEnumType(map[string]string{})
	if err := json.Unmarshal(marshalled, &typeUnderTest); err != nil {
		t.Fatal(err)
	}
	if typeUnderTest.ValidValues()["small"] != "Small" {
		t.Fatalf("Unmarshalling failed.")
	}
}

func TestStringEnumType(t *testing.T) {
	assertEqual(t, schema.NewStringEnumSchema(map[string]string{}).TypeID(), schema.TypeIDStringEnum)
	assertEqual(t, schema.NewStringEnumType(map[string]string{}).TypeID(), schema.TypeIDStringEnum)
}
