package schema_test

import (
	"fmt"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func ExampleNewIntEnumType() {
	// Create a new enum type by defining its valid values:
	payloadSize := schema.NewIntEnumType(map[int64]string{
		1024:    "Small",
		1048576: "Large",
	}, schema.Optional(schema.UnitBytes))

	// You can now print the valid values:
	fmt.Println(payloadSize.ValidValues())
	// Output: map[1024:Small 1048576:Large]
}

func ExampleIntEnumType_unserialize() {
	payloadSize := schema.NewIntEnumType(map[int64]string{
		1024:    "Small",
		1048576: "Large",
	}, schema.Optional(schema.UnitBytes))

	// Try to unserialize an invalid value:
	_, err := payloadSize.Unserialize(2048)
	fmt.Println(err)

	// Unserialize a valid value:
	val, err := payloadSize.Unserialize(1024)
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	// Unserialize a formatted value:
	val, err = payloadSize.Unserialize("1MB")
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	// Output: Validation failed: '2048' is not a valid value, must be one of: '1024', '1048576'
	// 1024
	// 1048576
}
