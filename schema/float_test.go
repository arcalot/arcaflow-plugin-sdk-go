package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

var testFloatSerializationDataSet = map[string]serializationTestCase[float64]{
	"tooSmallFloat64": {
		SerializedValue: float64(4),
		ExpectError:     true,
	},
	"tooLargeFloat64": {
		SerializedValue: float64(11),
		ExpectError:     true,
	},
	"tooSmallFloat32": {
		SerializedValue: float32(4),
		ExpectError:     true,
	},
	"tooLargeFloat32": {
		SerializedValue: float32(11),
		ExpectError:     true,
	},
	"tooSmallInt": {
		SerializedValue: 4,
		ExpectError:     true,
	},
	"tooLargeInt": {
		SerializedValue: 11,
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
	"validFloat64": {
		SerializedValue:         float64(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validFloat32": {
		SerializedValue:         float32(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validInt": {
		SerializedValue:         int(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validUInt": {
		SerializedValue:         uint(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validInt64": {
		SerializedValue:         int64(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validUInt64": {
		SerializedValue:         uint64(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validInt32": {
		SerializedValue:         int32(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validUInt32": {
		SerializedValue:         uint32(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validInt16": {
		SerializedValue:         int16(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validUInt16": {
		SerializedValue:         uint16(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validInt8": {
		SerializedValue:         int8(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validUInt8": {
		SerializedValue:         uint8(5),
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validString": {
		SerializedValue:         "5",
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"validStringUnit": {
		SerializedValue:         "5B",
		ExpectUnserializedValue: float64(5),
		ExpectedSerializedValue: float64(5),
	},
	"invalidType": {
		SerializedValue: struct{}{},
		ExpectError:     true,
	},
}

func TestFloatSerialization(t *testing.T) {
	performSerializationTest[float64](
		t,
		schema.NewFloatSchema(
			schema.PointerTo(float64(5)),
			schema.PointerTo(float64(10)),
			schema.UnitBytes,
		),
		testFloatSerializationDataSet,
		func(a float64, b float64) bool {
			return a == b
		},
		func(a any, b any) bool {
			return a == b
		},
	)
}

func TestFloatSerializationNoValidation(t *testing.T) {
	performSerializationTest[float64](
		t,
		schema.NewFloatSchema(nil, nil, nil),
		map[string]serializationTestCase[float64]{
			"stringInvalid": {
				SerializedValue: "a",
				ExpectError:     true,
			},
			"stringValid": {
				SerializedValue:         "1",
				ExpectedSerializedValue: float64(1),
				ExpectUnserializedValue: float64(1),
			},
			"stringValidNegative": {
				SerializedValue:         "-1",
				ExpectedSerializedValue: float64(-1),
				ExpectUnserializedValue: float64(-1),
			},
			"boolTrue": {
				SerializedValue:         true,
				ExpectedSerializedValue: float64(1),
				ExpectUnserializedValue: float64(1),
			},
			"boolFalse": {
				SerializedValue:         false,
				ExpectedSerializedValue: float64(0),
				ExpectUnserializedValue: float64(0),
			},
		},
		func(a float64, b float64) bool {
			return a == b
		},
		func(a any, b any) bool {
			return a == b
		},
	)
}

func TestFloatParameters(t *testing.T) {
	floatType := schema.NewFloatSchema(
		schema.PointerTo(float64(1)),
		schema.PointerTo(float64(2)),
		schema.UnitPercentage,
	)
	assertEqual(t, 1, *floatType.Min())
	assertEqual(t, 2, *floatType.Max())
	assertEqual(
		t,
		schema.UnitPercentage.BaseUnit().NameShortSingular(),
		(*floatType.Units()).BaseUnit().NameShortSingular(),
	)
}

func TestFloatType(t *testing.T) {
	assertEqual(t, schema.NewFloatSchema(nil, nil, nil).TypeID(), schema.TypeIDFloat)
}
