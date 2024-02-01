package schema_test

import (
    "encoding/json"
    "fmt"
    "go.arcalot.io/assert"
    "testing"

    "go.flow.arcalot.io/pluginsdk/schema"
)

func ExampleBoolSchema() {
    boolType := schema.NewBoolSchema()

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

func ExampleBoolSchema_unserialize() {
    boolType := schema.NewBoolSchema()

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

func TestBoolAliasSerialization(t *testing.T) {
    type T bool

    s := schema.NewBoolSchema()
    serializedData, err := s.Serialize(T(true))
    assert.NoError(t, err)
    assert.Equals(t, serializedData.(bool), true)
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
        // When executed in parallel, referencing tc from the
        // outer scope will not produce the proper value, so we need
        // to bind it to a variable, localTC, scoped inside
        // the loop body.
        localTC := tc
        t.Run(name, func(t *testing.T) {
            var boolType schema.Bool = schema.NewBoolSchema()
            output, err := boolType.Unserialize(localTC.input)
            if err != nil {
                if localTC.expectedError {
                    return
                }
                t.Fatalf("Failed to unserialize %v: %v", localTC.input, err)
            }
            if output != localTC.output {
                t.Fatalf("Unexpected unserialize output for %v: %v", localTC.input, output)
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
    j, err := json.Marshal(schema.NewBoolSchema())
    if err != nil {
        t.Fatal(err)
    }
    if string(j) != "{}" {
        t.Fatalf("Unexpected JSON output: %s", j)
    }
    boolType := schema.NewBoolSchema()
    if err := json.Unmarshal(j, &boolType); err != nil {
        t.Fatal(err)
    }
}

func TestBoolSchema(t *testing.T) {
    assert.Equals(t, schema.NewBoolSchema().TypeID(), schema.TypeIDBool)
}

func TestBoolSchema_ValidateCompatibility(t *testing.T) {
    s1 := schema.NewBoolSchema()
    assert.NoError(t, s1.ValidateCompatibility(s1))    // Itself
    assert.NoError(t, s1.ValidateCompatibility(true))  // a literal
    assert.NoError(t, s1.ValidateCompatibility(false)) // a literal
    assert.Error(t, s1.ValidateCompatibility(schema.NewStringSchema(nil, nil, nil)))
    assert.Error(t, s1.ValidateCompatibility(schema.NewIntSchema(nil, nil, nil)))
    assert.Error(t, s1.ValidateCompatibility(schema.NewListSchema(schema.NewBoolSchema(), nil, nil)))
    assert.Error(t, s1.ValidateCompatibility(schema.NewFloatSchema(nil, nil, nil)))
    assert.Error(t, s1.ValidateCompatibility(schema.NewDisplayValue(nil, nil, nil)))
    assert.NoError(t, s1.ValidateCompatibility(0)) // 0 and 1 are interpreted as booleans
    assert.NoError(t, s1.ValidateCompatibility(1)) // 0 and 1 are interpreted as booleans
    assert.Error(t, s1.ValidateCompatibility(2))
    assert.Error(t, s1.ValidateCompatibility(1.5))
    assert.Error(t, s1.ValidateCompatibility("test"))
    assert.Error(t, s1.ValidateCompatibility([]string{}))
    assert.Error(t, s1.ValidateCompatibility(map[string]any{}))
}
