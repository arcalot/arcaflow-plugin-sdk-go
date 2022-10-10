package schema_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func ExampleNewStringEnumSchema() {
	// Create a new enum type by defining its valid values:
	portionSize := schema.NewStringEnumSchema(map[string]string{
		"small": "Small",
		"large": "Large",
	})

	// You can now print the valid values:
	fmt.Println(portionSize.ValidValues())
	// Output: map[large:Large small:Small]
}

func ExampleStringEnumSchema_Unserialize() {
	portionSize := schema.NewStringEnumSchema(map[string]string{
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
		schema.NewStringEnumSchema(map[string]string{
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

func TestStringEnumTypedSerialization(t *testing.T) {
	type Size string
	s := schema.NewStringEnumSchema(map[string]string{
		"small": "Small",
		"large": "Large",
	})
	serializedData, err := s.Serialize(Size("small"))
	assertNoError(t, err)
	assertEqual(t, serializedData.(string), "small")
}

func TestStringEnumJSONMarshal(t *testing.T) {
	typeUnderTest := schema.NewStringEnumSchema(map[string]string{
		"small": "Small",
		"large": "Large",
	})

	marshalled, err := json.Marshal(typeUnderTest)
	if err != nil {
		t.Fatal(err)
	}
	if string(marshalled) != `{"values":{"large":"Large","small":"Small"}}` {
		t.Fatalf("Invalid marshalled JSON output: %s", marshalled)
	}
	typeUnderTest = schema.NewStringEnumSchema(map[string]string{})
	if err := json.Unmarshal(marshalled, &typeUnderTest); err != nil {
		t.Fatal(err)
	}
	if typeUnderTest.ValidValues()["small"] != "Small" {
		t.Fatalf("Unmarshalling failed.")
	}
}

func TestStringEnumType(t *testing.T) {
	assertEqual(t, schema.NewStringEnumSchema(map[string]string{}).TypeID(), schema.TypeIDStringEnum)
}
