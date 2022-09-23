package schema_test

import (
	"math"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var testIntSerializationDataSet = map[string]serializationTestCase[int64]{
	"tooSmallInt": {
		SerializedValue: int(4),
		ExpectError:     true,
	},
	"tooLargeInt": {
		SerializedValue: int(11),
		ExpectError:     true,
	},
	"tooSmallInt64": {
		SerializedValue: int64(4),
		ExpectError:     true,
	},
	"tooLargeInt64": {
		SerializedValue: int64(11),
		ExpectError:     true,
	},
	"tooSmallUInt": {
		SerializedValue: uint(4),
		ExpectError:     true,
	},
	"tooLargeUInt": {
		SerializedValue: uint(11),
		ExpectError:     true,
	},
	"tooSmallUInt64": {
		SerializedValue: uint64(4),
		ExpectError:     true,
	},
	"tooLargeUInt64": {
		SerializedValue: uint64(11),
		ExpectError:     true,
	},
	"tooSmallInt32": {
		SerializedValue: int32(4),
		ExpectError:     true,
	},
	"tooLargeInt32": {
		SerializedValue: int32(11),
		ExpectError:     true,
	},
	"tooSmallUInt32": {
		SerializedValue: uint32(4),
		ExpectError:     true,
	},
	"tooLargeUInt32": {
		SerializedValue: uint32(11),
		ExpectError:     true,
	},
	"tooSmallInt16": {
		SerializedValue: int16(4),
		ExpectError:     true,
	},
	"tooLargeInt16": {
		SerializedValue: int16(11),
		ExpectError:     true,
	},
	"tooSmallUInt16": {
		SerializedValue: uint16(4),
		ExpectError:     true,
	},
	"tooLargeUInt16": {
		SerializedValue: uint16(11),
		ExpectError:     true,
	},
	"tooSmallInt8": {
		SerializedValue: int8(4),
		ExpectError:     true,
	},
	"tooLargeInt8": {
		SerializedValue: int8(11),
		ExpectError:     true,
	},
	"tooSmallUInt8": {
		SerializedValue: uint8(4),
		ExpectError:     true,
	},
	"tooLargeUInt8": {
		SerializedValue: uint8(11),
		ExpectError:     true,
	},
	"tooSmallString": {
		SerializedValue: "4",
		ExpectError:     true,
	},
	"tooLargeString": {
		SerializedValue: "11",
		ExpectError:     true,
	},
	"tooSmallStringUnit": {
		SerializedValue: "4B",
		ExpectError:     true,
	},
	"tooLargeStringUnit": {
		SerializedValue: "1kB",
		ExpectError:     true,
	},
	"validInt": {
		SerializedValue:         int(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validUInt": {
		SerializedValue:         uint(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validInt64": {
		SerializedValue:         int64(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validUInt64": {
		SerializedValue:         uint64(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validInt32": {
		SerializedValue:         int32(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validUInt32": {
		SerializedValue:         uint32(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validInt16": {
		SerializedValue:         int16(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validUInt16": {
		SerializedValue:         uint16(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validInt8": {
		SerializedValue:         int8(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validUInt8": {
		SerializedValue:         uint8(5),
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validString": {
		SerializedValue:         "5",
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
	"validStringUnit": {
		SerializedValue:         "5B",
		ExpectUnserializedValue: int64(5),
		ExpectedSerializedValue: int64(5),
	},
}

func TestIntSerialization(t *testing.T) {
	performSerializationTest[int64](
		t,
		schema.NewIntSchema(schema.IntPointer(5), schema.IntPointer(10), schema.PointerTo(schema.UnitBytes)),
		testIntSerializationDataSet,
		func(a int64, b int64) bool {
			return a == b
		},
		func(a any, b any) bool {
			return a == b
		},
	)
}

func TestIntSerializationNoValidation(t *testing.T) {
	performSerializationTest[int64](
		t,
		schema.NewIntSchema(nil, nil, nil),
		map[string]serializationTestCase[int64]{
			"maxUInt": {
				SerializedValue: uint64(math.MaxUint64),
				ExpectError:     true,
			},
			"stringInvalid": {
				SerializedValue: "a",
				ExpectError:     true,
			},
			"stringValid": {
				SerializedValue:         "1",
				ExpectedSerializedValue: int64(1),
				ExpectUnserializedValue: int64(1),
			},
			"stringValidNegative": {
				SerializedValue:         "-1",
				ExpectedSerializedValue: int64(-1),
				ExpectUnserializedValue: int64(-1),
			},
			"boolTrue": {
				SerializedValue:         true,
				ExpectedSerializedValue: int64(1),
				ExpectUnserializedValue: int64(1),
			},
			"boolFalse": {
				SerializedValue:         false,
				ExpectedSerializedValue: int64(0),
				ExpectUnserializedValue: int64(0),
			},
		},
		func(a int64, b int64) bool {
			return a == b
		},
		func(a any, b any) bool {
			return a == b
		},
	)
}

func TestIntParameters(t *testing.T) {
	intType := schema.NewIntSchema(schema.IntPointer(1), schema.IntPointer(2), schema.PointerTo(schema.UnitBytes))
	assertEqual(t, 1, *intType.Min())
	assertEqual(t, 2, *intType.Max())
	assertEqual(t, schema.UnitBytes.BaseUnit().NameShortSingular(), (*intType.Units()).BaseUnit().NameShortSingular())
}

func TestIntType(t *testing.T) {
	assertEqual(t, schema.NewIntSchema(nil, nil, nil).TypeID(), schema.TypeIDInt)
}
