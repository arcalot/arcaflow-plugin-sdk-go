package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestUnitsParseInt(t *testing.T) {
	testMatrix := map[string]struct {
		input    string
		units    *schema.UnitsDefinition
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
		// When executed in parallel, referencing testData from the
		// outer scope will not produce the proper value, so we need
		// to bind it to a variable, localTestData, scoped inside
		// the loop body.
		localTestData := testData
		t.Run(testCase, func(t *testing.T) {
			result, err := localTestData.units.ParseInt(localTestData.input)
			if err != nil {
				t.Fatal(err)
			}
			if result != localTestData.expected {
				t.Fatalf("Result mismatch, expected: %d, got: %d", localTestData.expected, result)
			}
			formatted := localTestData.units.FormatShortInt(result)
			if formatted != localTestData.input {
				t.Fatalf("Formatted result doesn't match input, expected: %s, got: %s", localTestData.input, formatted)
			}
		})
	}
}

func TestUnitsParseFloat(t *testing.T) {
	testMatrix := map[string]struct {
		input    string
		units    *schema.UnitsDefinition
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
		// When executed in parallel, referencing testData from the
		// outer scope will not produce the proper value, so we need
		// to bind it to a variable, localTestData, scoped inside
		// the loop body.
		localTestData := testData
		t.Run(testCase, func(t *testing.T) {
			result, err := localTestData.units.ParseFloat(localTestData.input)
			if err != nil {
				t.Fatal(err)
			}
			if result != localTestData.expected {
				t.Fatalf("Result mismatch, expected: %f, got: %f", localTestData.expected, result)
			}
			formatted := localTestData.units.FormatShortFloat(result)
			if formatted != localTestData.input {
				t.Fatalf("Formatted result doesn't match input, expected: %s, got: %s", localTestData.input, formatted)
			}
		})
	}
}
