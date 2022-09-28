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
			305000000000,
		},
		"1%": {
			"1%",
			schema.UnitPercentage,
			1,
		},
		"1kB": {
			"1kB",
			schema.UnitBytes,
			1024,
		},
		"1char": {
			"1char",
			schema.UnitCharacters,
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

func TestUnitsParseFloat(t *testing.T) {
	testMatrix := map[string]struct {
		input    string
		units    schema.Units
		expected float64
	}{
		"5m5s": {
			"5m5.1s",
			schema.UnitDurationSeconds,
			305.1,
		},
		"1.1%": {
			"1.1%",
			schema.UnitPercentage,
			1.1,
		},
	}

	for testCase, testData := range testMatrix {
		t.Run(testCase, func(t *testing.T) {
			result, err := testData.units.ParseFloat(testData.input)
			if err != nil {
				t.Fatal(err)
			}
			if result != testData.expected {
				t.Fatalf("Result mismatch, expected: %f, got: %f", testData.expected, result)
			}
			formatted := testData.units.FormatShortFloat(result)
			if formatted != testData.input {
				t.Fatalf("Formatted result doesn't match input, expected: %s, got: %s", testData.input, formatted)
			}
		})
	}
}
