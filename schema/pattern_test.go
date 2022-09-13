package schema_test

import (
	"fmt"
	"regexp"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func ExamplePatternType() {
	patternType := schema.NewPatternType()

	// Unserialize a string
	pattern, err := patternType.Unserialize("^[a-z]+$")
	if err != nil {
		panic(err)
	}
	if pattern.MatchString("asdf") {
		fmt.Println("The pattern matches!")
	}

	// Output: The pattern matches!
}

func TestPatternType(t *testing.T) {
	performSerializationTest[*regexp.Regexp](
		t,
		schema.NewPatternType(),
		map[string]serializationTestCase[*regexp.Regexp]{
			"valid": {
				"^[a-z]+$",
				false,
				regexp.MustCompile("^[a-z]+$"),
				"^[a-z]+$",
			},
			"invalidPattern": {
				"^[a-z",
				true,
				nil,
				nil,
			},
			"invalidType": {
				struct{}{},
				true,
				nil,
				nil,
			},
		},
		func(a *regexp.Regexp, b *regexp.Regexp) bool {
			if a == nil || b == nil {
				return false
			}
			return a.String() == b.String()
		},
		func(a any, b any) bool {
			return a.(string) == b.(string)
		},
	)
}

func TestPatternValidateInvalid(t *testing.T) {
	patternType := schema.NewPatternType()
	if err := patternType.Validate(nil); err == nil {
		t.Fatalf("Validating nil did not result in an error.")
	}
}

func TestPatternSerializeInvalid(t *testing.T) {
	patternType := schema.NewPatternType()
	if _, err := patternType.Serialize(nil); err == nil {
		t.Fatalf("Serializing nil did not result in an error.")
	}
}

func TestPatternID(t *testing.T) {
	assertEqual(t, schema.NewPatternSchema().TypeID(), schema.TypeIDPattern)
	assertEqual(t, schema.NewPatternType().TypeID(), schema.TypeIDPattern)
}
