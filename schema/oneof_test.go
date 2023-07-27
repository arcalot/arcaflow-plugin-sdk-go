package schema_test

import (
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
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
