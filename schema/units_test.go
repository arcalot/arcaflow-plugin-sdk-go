package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestUnitsParseInt(t *testing.T) {
	testMatrix := map[string]struct {
		input    string
		units    schema.Units
		expected int64
	}{
		"5m5s": {
			"5m5s",
			schema.UnitDurationNanoseconds,
			305000000,
		},
		"1%": {
			"1%",
			schema.UnitPercentage,
			1,
		},
	}

	for testCase, testData := range testMatrix {
		t.Run(testCase, func(t *testing.T) {
			result, err := testData.units.ParseInt(testData.input)
			if err != nil {
				t.Fatal(err)
			}
			if result != testData.expected {
				t.Fatalf("Result mismatch, expected: %d, got: %d", testData.expected, result)
			}
			formatted := testData.units.FormatShortInt(result)
			if formatted != testData.input {
				t.Fatalf("Formatted result doesn't match input, expected: %s, got: %s", testData.input, formatted)
			}
		})
	}
}
