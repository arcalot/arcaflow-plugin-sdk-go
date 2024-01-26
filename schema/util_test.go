package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type serializationTestCase[T any] struct {
	SerializedValue         any
	ExpectError             bool
	ExpectUnserializedValue T
	ExpectedSerializedValue any
}

func performSerializationTest[T any](
	t *testing.T,
	typeUnderTest schema.TypedType[T],
	testCases map[string]serializationTestCase[T],
	compareUnserialized func(a T, b T) bool,
	compareSerialized func(a any, b any) bool,
) {
	t.Helper()
	for name, tc := range testCases {
		// The call to t.Parallel() means that referencing the tc
		// from the outer scope won't produce the proper value, so
		// we need to place it in a variable, localTC, scoped inside
		// the loop body.
		localTC := tc
		t.Run(name, func(t *testing.T) {
			t.Helper()
			unserialized, err := typeUnderTest.UnserializeType(localTC.SerializedValue)
			if err != nil {
				if localTC.ExpectError {
					return
				}
				t.Fatal(err)
			}
			if err := typeUnderTest.ValidateType(unserialized); err != nil {
				t.Fatal(err)
			}
			if !compareUnserialized(unserialized, localTC.ExpectUnserializedValue) {
				t.Fatalf(
					"Unexpected unserialized value, expected: %v, got: %v",
					localTC.ExpectUnserializedValue,
					unserialized,
				)
			}
			serialized, err := typeUnderTest.SerializeType(unserialized)
			if err != nil {
				t.Fatal(err)
			}
			if !compareSerialized(serialized, localTC.ExpectedSerializedValue) {
				t.Fatalf(
					"Serialized value mismatch, expected: %v (%T), got: %v (%T)",
					localTC.ExpectedSerializedValue,
					localTC.ExpectedSerializedValue,
					serialized,
					serialized,
				)
			}
		})
	}
}
