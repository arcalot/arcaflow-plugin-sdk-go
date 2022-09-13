package schema

// PropertySchema holds the schema definition for a single object property. It is usable in conjunction with
// ObjectSchema.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use PropertyType.
type PropertySchema interface {
	Type() AbstractSchema
	Display() *DisplayValue
	Default() *string
	Required() bool
	RequiredIf() []string
	RequiredIfNot() []string
	Conflicts() []string
	Examples() []string
}

// NewPropertySchema creates a new object property schema.
func NewPropertySchema(
	t AbstractSchema,
	displayValue *DisplayValue,
	required bool,
	requiredIf []string,
	requiredIfNot []string,
	conflicts []string,
	defaultValue *string,
	examples []string,
) PropertySchema {
	return &propertySchema{
		t,
		displayValue,
		required,
		requiredIf,
		requiredIfNot,
		conflicts,
		defaultValue,
		examples,
	}
}

type propertySchema struct {
	TypeValue          AbstractSchema `json:"type"`
	DisplayValue       *DisplayValue  `json:"display,omitempty"`
	RequiredValue      bool           `json:"required"`
	RequiredIfValue    []string       `json:"required_if,omitempty"`
	RequiredIfNotValue []string       `json:"required_if_not,omitempty"`
	ConflictsValue     []string       `json:"conflicts,omitempty"`
	DefaultValue       *string        `json:"default,omitempty"`
	ExamplesValue      []string       `json:"examples,omitempty"`
}

func (p propertySchema) Default() *string {
	return p.DefaultValue
}

func (p propertySchema) Type() AbstractSchema {
	return p.TypeValue
}

func (p propertySchema) Display() *DisplayValue {
	return p.DisplayValue
}

func (p propertySchema) Required() bool {
	return p.RequiredValue
}

func (p propertySchema) RequiredIf() []string {
	return p.RequiredIfValue
}

func (p propertySchema) RequiredIfNot() []string {
	return p.RequiredIfNotValue
}

func (p propertySchema) Conflicts() []string {
	return p.ConflictsValue
}

func (p propertySchema) Examples() []string {
	return p.ExamplesValue
}
