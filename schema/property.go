package schema

import (
	"fmt"
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
		false,
		nil,
	}
}

type PropertySchema struct {
	TypeValue          Type     `json:"type"`
	DisplayValue       Display  `json:"display"`
	RequiredValue      bool     `json:"required"`
	RequiredIfValue    []string `json:"required_if"`
	RequiredIfNotValue []string `json:"required_if_not"`
	ConflictsValue     []string `json:"conflicts"`
	DefaultValue       *string  `json:"default"`
	ExamplesValue      []string `json:"examples"`

	emptyIsDefault bool

	// Disabled sets whether the field can be used. Set the DisabledReason if set to true.
	Disabled bool `json:"disabled"`
	// DisabledReason explains why the property is disabled. Default nil
	DisabledReason *string `json:"disabled_reason"`
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

// Disable is a builder-pattern way of disabling the property.
func (p *PropertySchema) Disable(reason string) *PropertySchema {
	p.Disabled = true
	p.DisabledReason = &reason
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

func (p *PropertySchema) ApplyNamespace(objects map[string]*ObjectSchema, namespace string) {
	p.TypeValue.ApplyNamespace(objects, namespace)
}

func (p *PropertySchema) ValidateReferences() error {
	return p.TypeValue.ValidateReferences()
}

func (p *PropertySchema) Unserialize(data any) (any, error) {
	if !p.Disabled {
		return p.TypeValue.Unserialize(data)
	} else {
		// Note, this is last, so that actual validation errors are returned before the disabled err
		if p.DisabledReason == nil {
			return nil, &ConstraintError{
				Message: "error due to attempting to use disabled property",
			}
		} else {
			return nil, &ConstraintError{
				Message: fmt.Sprintf("error due to attempting to use disabled property: %s", *p.DisabledReason),
			}
		}
	}
}

func (p *PropertySchema) ValidateCompatibility(typeOrData any) error {
	schemaType, ok := typeOrData.(*PropertySchema)
	if ok {
		return p.TypeValue.ValidateCompatibility(schemaType.TypeValue)
	}
	err := p.TypeValue.ValidateCompatibility(typeOrData)
	if err != nil {
		if p.DisplayValue != nil && p.Display().Name() != nil {
			return &ConstraintError{
				Message: fmt.Sprintf("error while validating sub-type of property %s with type %T (%s)",
					*p.Display().Name(), p.TypeValue, err),
			}
		} else {
			return &ConstraintError{
				Message: fmt.Sprintf("error while validating sub-type of property type %T (%s)",
					p.TypeValue, err),
			}
		}
	}
	// Now just check to see if it's enabled.
	if !p.Disabled {
		return nil
	} else {
		// Note, this is last, so that actual validation errors are returned before the disabled err
		if p.DisabledReason == nil {
			return &ConstraintError{
				Message: "error due to attempting to use disabled property",
			}
		} else {
			return &ConstraintError{
				Message: fmt.Sprintf("error due to attempting to use disabled property: %s", *p.DisabledReason),
			}
		}
	}
}

func (p *PropertySchema) Validate(data any) error {
	return p.TypeValue.Validate(data)
}
func (p *PropertySchema) Serialize(data any) (any, error) {
	return p.TypeValue.Serialize(data)
}
