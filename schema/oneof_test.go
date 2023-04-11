package schema_test

import (
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
	assertEqual(
		t,
		oneOfStringTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assertEqual(
		t,
		oneOfStringTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assertEqual(
		t,
		oneOfIntTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
	assertEqual(
		t,
		oneOfIntTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
}
