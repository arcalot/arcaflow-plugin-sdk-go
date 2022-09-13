package schema

// DisplayValue holds the data related to displaying fields.
type DisplayValue interface {
	Name() *string
	Description() *string
	Icon() *string
}

// NewDisplayValue creates a new DisplayValue from the given parameters.
func NewDisplayValue(name *string, description *string, icon *string) DisplayValue {
	return &displayValue{
		NameValue:        name,
		DescriptionValue: description,
		IconValue:        icon,
	}
}

type displayValue struct {
	NameValue        *string `json:"name"`
	DescriptionValue *string `json:"description"`
	IconValue        *string `json:"icon"`
}

func (d displayValue) Name() *string {
	return d.NameValue
}

func (d displayValue) Description() *string {
	return d.DescriptionValue
}

func (d displayValue) Icon() *string {
	return d.IconValue
}
