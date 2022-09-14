package schema

// Schema is a collection of steps supported by a plugin.
type Schema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] interface {
	Steps() map[string]StepSchema[P, O, S]
}

// NewSchema builds a new schema with the specified steps.
func NewSchema(
	steps map[string]StepSchema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	],
) Schema[PropertySchema, ObjectSchema[PropertySchema], ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]]] {
	return &schema[
		PropertySchema,
		ObjectSchema[PropertySchema],
		ScopeSchema[PropertySchema, ObjectSchema[PropertySchema]],
	]{
		steps,
	}
}

type schema[P PropertySchema, O ObjectSchema[P], S ScopeSchema[P, O]] struct {
	StepsValue map[string]StepSchema[P, O, S] `json:"steps"`
}

func (s schema[P, O, S]) Steps() map[string]StepSchema[P, O, S] {
	return s.StepsValue
}
