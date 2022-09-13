package schema_test

import (
	"fmt"

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
