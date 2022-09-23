package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestDisplayValue(t *testing.T) {
	var dv schema.Display = schema.NewDisplayValue(
		schema.PointerTo("Greeting"),
		schema.PointerTo("Hello world!"),
		schema.PointerTo("<svg ...></svg>"),
	)
	assertEqual(t, *dv.Name(), "Greeting")
	assertEqual(t, *dv.Description(), "Hello world!")
	assertEqual(t, *dv.Icon(), "<svg ...></svg>")
}
