package schema_test

import (
	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
	"testing"
)

type oneOfTestObjectB struct {
	Message string `json:"message"`
}

func (o oneOfTestObjectB) String() string {
	return o.Message
}

type oneOfTestObjectC struct {
	M string `json:"m"`
}

type oneOfTestObjectA struct {
	S any `json:"s"`
}

func TestOneOfTypeID(t *testing.T) {
	assert.Equals(
		t,
		oneOfStringTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assert.Equals(
		t,
		oneOfStringTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assert.Equals(
		t,
		oneOfIntTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
	assert.Equals(
		t,
		oneOfIntTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
}

var oneOfTestObjectBProperties = map[string]*schema.PropertySchema{
	"message": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestObjectCProperties = map[string]*schema.PropertySchema{
	"m": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestObjectDProperties = map[string]*schema.PropertySchema{
	"K": schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestBSchema = schema.NewObjectSchema(
	"B",
	oneOfTestObjectBProperties,
)

var oneOfTestCSchema = schema.NewObjectSchema(
	"C",
	oneOfTestObjectCProperties,
)

var oneOfTestDSchema = schema.NewObjectSchema(
	"D",
	oneOfTestObjectDProperties,
)

var oneOfTestBMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestObjectB](
	"B",
	oneOfTestObjectBProperties,
)

var oneOfTestCMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestObjectC](
	"C",
	oneOfTestObjectCProperties,
)
