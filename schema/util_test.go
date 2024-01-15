package schema_test

import (
	"go.arcalot.io/assert"
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
		t.Run(name, func(t *testing.T) {
			t.Helper()
			unserialized, err := typeUnderTest.UnserializeType(tc.SerializedValue)
			if err != nil {
				if tc.ExpectError {
					return
				}
				t.Fatal(err)
			}
			if err := typeUnderTest.ValidateType(unserialized); err != nil {
				t.Fatal(err)
			}
			if !compareUnserialized(unserialized, tc.ExpectUnserializedValue) {
				t.Fatalf(
					"Unexpected unserialized value, expected: %v, got: %v",
					tc.ExpectUnserializedValue,
					unserialized,
				)
			}
			serialized, err := typeUnderTest.SerializeType(unserialized)
			if err != nil {
				t.Fatal(err)
			}
			if !compareSerialized(serialized, tc.ExpectedSerializedValue) {
				t.Fatalf(
					"Serialized value mismatch, expected: %v (%T), got: %v (%T)",
					tc.ExpectedSerializedValue,
					tc.ExpectedSerializedValue,
					serialized,
					serialized,
				)
			}
			serialized2, err := typeUnderTest.SerializeType(unserialized)
			assert.NoError(t, err)
			assert.Equals(t, serialized2, serialized)
			unserialized2, err := typeUnderTest.UnserializeType(serialized2)
			assert.NoError(t, err)
			assert.Equals(t, unserialized2, unserialized)
		})
	}
}
