package schema

import "fmt"

// PropertySchema holds the schema definition for a single object property. It is usable in conjunction with
// ObjectSchema.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use PropertyType.
type PropertySchema interface {
	Type() AbstractSchema
	Display() DisplayValue
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
	displayValue DisplayValue,
	required bool,
	requiredIf []string,
	requiredIfNot []string,
	conflicts []string,
	defaultValue *string,
	examples []string,
) PropertySchema {
	return &abstractPropertySchema[AbstractSchema]{
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

type abstractPropertySchema[T AbstractSchema] struct {
	TypeValue          T            `json:"type"`
	DisplayValue       DisplayValue `json:"display,omitempty"`
	RequiredValue      bool         `json:"required"`
	RequiredIfValue    []string     `json:"required_if,omitempty"`
	RequiredIfNotValue []string     `json:"required_if_not,omitempty"`
	ConflictsValue     []string     `json:"conflicts,omitempty"`
	DefaultValue       *string      `json:"default,omitempty"`
	ExamplesValue      []string     `json:"examples,omitempty"`
}

//nolint:unused
type propertySchema struct {
	abstractPropertySchema[AbstractSchema] `json:",inline"`
}

func (p abstractPropertySchema[T]) Default() *string {
	return p.DefaultValue
}

func (p abstractPropertySchema[T]) Type() AbstractSchema {
	return p.TypeValue
}

func (p abstractPropertySchema[T]) Display() DisplayValue {
	return p.DisplayValue
}

func (p abstractPropertySchema[T]) Required() bool {
	return p.RequiredValue
}

func (p abstractPropertySchema[T]) RequiredIf() []string {
	return p.RequiredIfValue
}

func (p abstractPropertySchema[T]) RequiredIfNot() []string {
	return p.RequiredIfNotValue
}

func (p abstractPropertySchema[T]) Conflicts() []string {
	return p.ConflictsValue
}

func (p abstractPropertySchema[T]) Examples() []string {
	return p.ExamplesValue
}

// PropertyType is a typed version of PropertySchema.
type PropertyType interface {
	PropertySchema
	AbstractType[any]
}

// NewPropertyType defines a new property to be used in an object.
func NewPropertyType[T any](
	t AbstractType[T],
	displayValue DisplayValue,
	required bool,
	requiredIf []string,
	requiredIfNot []string,
	conflicts []string,
	defaultValue *string,
	examples []string,
) PropertyType {
	return &propertyType[T]{
		abstractPropertySchema[AbstractType[T]]{
			t,
			displayValue,
			required,
			requiredIf,
			requiredIfNot,
			conflicts,
			defaultValue,
			examples,
		},
	}
}

type propertyType[T any] struct {
	abstractPropertySchema[AbstractType[T]] `json:",inline"`
}

func (p propertyType[T]) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
	p.TypeValue.ApplyScope(s)
}

func (p propertyType[T]) UnderlyingType() any {
	return p.TypeValue.UnderlyingType()
}

func (p propertyType[T]) TypeID() TypeID {
	return p.TypeValue.TypeID()
}

func (p propertyType[T]) Unserialize(data any) (any, error) {
	return p.TypeValue.Unserialize(data)
}

func (p propertyType[T]) Validate(data any) error {
	typedData, err := p.typeData(data)
	if err != nil {
		return err
	}
	return p.TypeValue.Validate(typedData)
}

func (p propertyType[T]) typeData(data any) (T, error) {
	var typedData T
	var ok bool
	typedData, ok = data.(T)
	if !ok {
		return typedData, &ConstraintError{
			Message: fmt.Sprintf(
				"Type error: cannot use %T as %T",
				data,
				typedData,
			),
		}
	}
	return typedData, nil
}

func (p propertyType[T]) Serialize(data any) (any, error) {
	typedData, err := p.typeData(data)
	if err != nil {
		return typedData, err
	}
	return p.TypeValue.Serialize(typedData)
}
