package schema_test

import (
	"fmt"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func ExampleNewIntEnumSchema() {
	// Create a new enum type by defining its valid values:
	var payloadSize schema.IntEnum = schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		1024:    {NameValue: schema.PointerTo("Small")},
		1048576: {NameValue: schema.PointerTo("Large")},
	}, schema.UnitBytes)

	// You can now print the valid values:
	fmt.Println(*payloadSize.ValidValues()[1024].NameValue)
	// Output: Small
}

func ExampleIntEnumSchema_unserialize() {
	payloadSize := schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		1024:    {NameValue: schema.PointerTo("Small")},
		1048576: {NameValue: schema.PointerTo("Large")},
	}, schema.UnitBytes)

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

var testIntEnumSerializationDataSet = map[string]serializationTestCase[int64]{
	"validNumberInt64": {
		int64(1024),
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberInt64": {
		int64(2024),
		true,
		0,
		0,
	},
	"validNumberUInt64": {
		uint64(1024),
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberUInt64": {
		uint64(2048),
		true,
		0,
		0,
	},
	"validNumberInt": {
		1024,
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberInt": {
		2048,
		true,
		int64(1024),
		int64(1024),
	},
	"validNumberUInt": {
		uint(1024),
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberUInt": {
		uint(2048),
		true,
		0,
		0,
	},
	"validNumberInt32": {
		int32(1024),
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberInt32": {
		int32(2048),
		true,
		int64(0),
		int64(0),
	},
	"validNumberUInt32": {
		uint32(1024),
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberUInt32": {
		uint32(2048),
		true,
		int64(0),
		int64(0),
	},
	"validNumberInt16": {
		int16(1024),
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberInt16": {
		int16(2048),
		true,
		int64(0),
		int64(0),
	},
	"validNumberUInt16": {
		uint16(1024),
		false,
		int64(1024),
		int64(1024),
	},
	"invalidNumberUInt16": {
		uint16(2048),
		true,
		int64(0),
		int64(0),
	},
	"validNumberInt8": {
		int8(64),
		false,
		int64(64),
		int64(64),
	},
	"invalidNumberInt8": {
		int8(63),
		true,
		int64(0),
		int64(0),
	},
	"validNumberUInt8": {
		uint8(64),
		false,
		int64(64),
		int64(64),
	},
	"invalidNumberUInt8": {
		uint8(129),
		true,
		int64(0),
		int64(0),
	},
	"validString": {
		"1024",
		false,
		int64(1024),
		int64(1024),
	},
	"invalidString": {
		"1023",
		true,
		int64(0),
		int64(0),
	},
	"invalidType": {
		struct{}{},
		true,
		int64(0),
		int64(0),
	},
	"validUnitType": {
		"1kB",
		false,
		int64(1024),
		int64(1024),
	},
}

func TestIntEnumSerialization(t *testing.T) {
	performSerializationTest[int64](
		t,
		schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
			64:      {NameValue: schema.PointerTo("XS")},
			1024:    {NameValue: schema.PointerTo("Small")},
			1048576: {NameValue: schema.PointerTo("Large")},
		}, schema.UnitBytes),
		testIntEnumSerializationDataSet,
		func(a int64, b int64) bool {
			return a == b
		},
		func(a any, b any) bool {
			return a == b
		},
	)
}

func TestIntEnumTypedSerialization(t *testing.T) {
	type Bytes int64
	s := schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		64:      {NameValue: schema.PointerTo("XS")},
		1024:    {NameValue: schema.PointerTo("Small")},
		1048576: {NameValue: schema.PointerTo("Large")},
	}, schema.UnitBytes)
	serializedData, err := s.Serialize(Bytes(64))
	assert.NoError(t, err)
	assert.Equals(t, serializedData.(int64), 64)
}

func TestIntEnumSchema(t *testing.T) {
	assert.Equals(t, schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{}, nil).TypeID(), schema.TypeIDIntEnum)
}

func TestIntEnumCompatibilityValidation(t *testing.T) {
	payloadSize := schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		1024:    {NameValue: schema.PointerTo("Small")},
		1048576: {NameValue: schema.PointerTo("Large")},
	}, schema.UnitBytes)
	assert.NoError(t, payloadSize.ValidateCompatibility(1024))
	assert.Error(t, payloadSize.ValidateCompatibility(1025))
}

func TestIntEnumSchemaCompatibilityValidation(t *testing.T) {
	s1 := schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		1: {NameValue: schema.PointerTo("a")},
		2: {NameValue: schema.PointerTo("b")},
		3: {NameValue: schema.PointerTo("c")},
	}, nil)
	s2 := schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		4: {NameValue: schema.PointerTo("a")},
		5: {NameValue: schema.PointerTo("b")},
		6: {NameValue: schema.PointerTo("c")},
	}, nil)
	S1 := schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{
		1: {NameValue: schema.PointerTo("A")},
		2: {NameValue: schema.PointerTo("B")},
		3: {NameValue: schema.PointerTo("C")},
	}, nil)

	assert.NoError(t, s1.ValidateCompatibility(s1))
	assert.NoError(t, s2.ValidateCompatibility(s2))
	// Mismatched keys
	assert.Error(t, s1.ValidateCompatibility(s2))
	assert.Error(t, s2.ValidateCompatibility(s1))
	// Mismatched names
	assert.Error(t, s1.ValidateCompatibility(S1))
	assert.Error(t, S1.ValidateCompatibility(s1))
}
