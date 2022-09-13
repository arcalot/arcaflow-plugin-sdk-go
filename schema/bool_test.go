package schema_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func ExampleBoolType() {
	boolType := schema.NewBoolType()

	// Unserialize a bool:
	unserializedValue, err := boolType.Unserialize(true)
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Validate:
	if err := boolType.Validate(unserializedValue); err != nil {
		panic(err)
	}

	// Serialize:
	serializedValue, err := boolType.Serialize(unserializedValue)
	if err != nil {
		panic(err)
	}
	fmt.Println(serializedValue)

	// Print protocol type ID
	fmt.Println(boolType.TypeID())

	// Output: true
	// true
	// bool
}

func ExampleBoolType_unserialize() {
	boolType := schema.NewBoolType()

	// Unserialize a bool:
	unserializedValue, err := boolType.Unserialize(true)
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Unserialize an int:
	unserializedValue, err = boolType.Unserialize(1)
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Unserialize a string like this:
	unserializedValue, err = boolType.Unserialize("true")
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Or like this:
	unserializedValue, err = boolType.Unserialize("yes")
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Or like this:
	unserializedValue, err = boolType.Unserialize("y")
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Or like this:
	unserializedValue, err = boolType.Unserialize("enable")
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Or like this:
	unserializedValue, err = boolType.Unserialize("enable")
	if err != nil {
		panic(err)
	}
	fmt.Println(unserializedValue)

	// Output: true
	// true
	// true
	// true
	// true
	// true
	// true
}

var boolTestSerializationCases = map[string]struct {
	input         interface{}
	expectedError bool
	output        bool
}{
	"true":     {input: "true", output: true},
	"false":    {input: "false", output: false},
	"2":        {input: "2", expectedError: true},
	"int-2":    {input: 2, expectedError: true},
	"1":        {input: "1", output: true},
	"int-1":    {input: 1, output: true},
	"uint-1":   {input: uint(1), output: true},
	"int64-1":  {input: int64(1), output: true},
	"int32-1":  {input: int32(1), output: true},
	"int16-1":  {input: int16(1), output: true},
	"int8-1":   {input: int8(1), output: true},
	"uint64-1": {input: uint64(1), output: true},
	"uint32-1": {input: uint32(1), output: true},
	"uint16-1": {input: uint16(1), output: true},
	"uint8-1":  {input: uint8(1), output: true},
	"0":        {input: "0", output: false},
	"int-0":    {input: 0, output: false},
	"uint-0":   {input: uint(0), output: false},
	"int64-0":  {input: int64(0), output: false},
	"int32-0":  {input: int32(0), output: false},
	"int16-0":  {input: int16(0), output: false},
	"int8-0":   {input: int8(0), output: false},
	"uint64-0": {input: uint64(0), output: false},
	"uint32-0": {input: uint32(0), output: false},
	"uint16-0": {input: uint16(0), output: false},
	"uint8-0":  {input: uint8(0), output: false},
	"yes":      {input: "yes", output: true},
	"no":       {input: "no", output: false},
	"y":        {input: "y", output: true},
	"n":        {input: "n", output: false},
	"enable":   {input: "enable", output: true},
	"disable":  {input: "disable", output: false},
	"enabled":  {input: "enabled", output: true},
	"disabled": {input: "disabled", output: false},
}

func TestBoolSerializationCycle(t *testing.T) {
	for name, tc := range boolTestSerializationCases {
		t.Run(name, func(t *testing.T) {
			boolType := schema.NewBoolType()
			output, err := boolType.Unserialize(tc.input)
			if err != nil {
				if tc.expectedError {
					return
				}
				t.Fatalf("Failed to unserialize %v: %v", tc.input, err)
			}
			if output != tc.output {
				t.Fatalf("Unexpected unserialize output for %v: %v", tc.input, output)
			}

			if err := boolType.Validate(output); err != nil {
				t.Fatalf("Failed to validate %v: %v", output, err)
			}

			serialized, err := boolType.Serialize(output)
			if err != nil {
				t.Fatalf("Failed to serialize %v: %v", output, err)
			}
			if serialized != output {
				t.Fatalf("Invalid value after serialization: %v", serialized)
			}
		})
	}
}

func TestBoolJSONMarshal(t *testing.T) {
	j, err := json.Marshal(schema.NewBoolType())
	if err != nil {
		t.Fatal(err)
	}
	if string(j) != "{}" {
		t.Fatalf("Unexpected JSON output: %s", j)
	}
	boolType := schema.NewBoolType()
	if err := json.Unmarshal(j, &boolType); err != nil {
		t.Fatal(err)
	}
}

func TestBoolType(t *testing.T) {
	assertEqual(t, schema.NewBoolSchema().TypeID(), schema.TypeIDBool)
	assertEqual(t, schema.NewBoolType().TypeID(), schema.TypeIDBool)
}
