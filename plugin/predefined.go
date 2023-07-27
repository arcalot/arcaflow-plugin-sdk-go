package plugin

import "go.flow.arcalot.io/pluginsdk/schema"

type CancelInput struct {
	// Possibly add a time limit input
}

var CancellationSignalSchema = schema.NewSignalSchema(
	"cancel",
	schema.NewScopeSchema(
		schema.NewStructMappedObjectSchema[CancelInput](
			"cancelInput",
			map[string]*schema.PropertySchema{},
		),
	),
	schema.NewDisplayValue(
		schema.PointerTo("Cancel"),
		schema.PointerTo("Cancels the running step."),
		nil,
	),
)
