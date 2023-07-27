package schema_test

import (
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

func TestDisplayValue(t *testing.T) {
	var dv schema.Display = schema.NewDisplayValue(
		schema.PointerTo("Greeting"),
		schema.PointerTo("Hello world!"),
		schema.PointerTo("<svg ...></svg>"),
	)
	assert.Equals(t, *dv.Name(), "Greeting")
	assert.Equals(t, *dv.Description(), "Hello world!")
	assert.Equals(t, *dv.Icon(), "<svg ...></svg>")
}
