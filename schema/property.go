package schema

import (
	"reflect"
)

// Property holds the schema definition for a single object property. It is usable in conjunction with
// Object.
type Property interface {
	Type

	// Type returns the underlying type this property holds.
	Type() Type
	Display() Display
	Default() *string
	Required() bool
	RequiredIf() []string
	RequiredIfNot() []string
	Conflicts() []string
	Examples() []string
}

// NewPropertySchema creates a new object property schema.
func NewPropertySchema(
	t Type,
	displayValue Display,
	required bool,
	requiredIf []string,
	requiredIfNot []string,
	conflicts []string,
	defaultValue *string,
	examples []string,
) *PropertySchema {
	return &PropertySchema{
		t,
		displayValue,
		required,
		requiredIf,
		requiredIfNot,
		conflicts,
		defaultValue,
		examples,
		false,
	}
}

type PropertySchema struct {
	TypeValue          Type     `json:"type"`
	DisplayValue       Display  `json:"display,omitempty"`
	RequiredValue      bool     `json:"required"`
	RequiredIfValue    []string `json:"required_if,omitempty"`
	RequiredIfNotValue []string `json:"required_if_not,omitempty"`
	ConflictsValue     []string `json:"conflicts,omitempty"`
	DefaultValue       *string  `json:"default,omitempty"`
	ExamplesValue      []string `json:"examples,omitempty"`

	emptyIsDefault bool
}

// TreatEmptyAsDefaultValue triggers the property to treat an empty value (e.g. "", or 0) as the default value for
// serialization and validation only. It has no effect on objects that are not mapped to a struct.
//
// This is useful in case of third party structs where the property may not have a pointer despite being optional.
// However, to avoid ambiguity and better performance, this option should be used only when needed.
func (p *PropertySchema) TreatEmptyAsDefaultValue() *PropertySchema {
	p.emptyIsDefault = true
	return p
}

func (p *PropertySchema) Default() *string {
	return p.DefaultValue
}

func (p *PropertySchema) ReflectedType() reflect.Type {
	return p.TypeValue.ReflectedType()
}

func (p *PropertySchema) Type() Type {
	return p.TypeValue
}

func (p *PropertySchema) TypeID() TypeID {
	return p.TypeValue.TypeID()
}

func (p *PropertySchema) Display() Display {
	return p.DisplayValue
}

func (p *PropertySchema) Required() bool {
	return p.RequiredValue
}

func (p *PropertySchema) RequiredIf() []string {
	return p.RequiredIfValue
}

func (p *PropertySchema) RequiredIfNot() []string {
	return p.RequiredIfNotValue
}

func (p *PropertySchema) Conflicts() []string {
	return p.ConflictsValue
}

func (p *PropertySchema) Examples() []string {
	return p.ExamplesValue
}

func (p *PropertySchema) ApplyScope(scope Scope) {
	p.TypeValue.ApplyScope(scope)
}

func (p *PropertySchema) Unserialize(data any) (any, error) {
	return p.TypeValue.Unserialize(data)
}

func (p *PropertySchema) Validate(data any) error {
	return p.TypeValue.Validate(data)
}
func (p *PropertySchema) Serialize(data any) (any, error) {
	return p.TypeValue.Serialize(data)
}
