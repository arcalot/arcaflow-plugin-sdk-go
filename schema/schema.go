package schema

// Schema is a collection of steps supported by a plugin.
type Schema interface {
	Steps() map[string]StepSchema
}

// NewSchema builds a new schema with the specified steps.
func NewSchema(steps map[string]StepSchema) Schema {
	return &schema{
		steps,
	}
}

type schema struct {
	StepsValue map[string]StepSchema `json:"steps"`
}

func (s schema) Steps() map[string]StepSchema {
	return s.StepsValue
}
