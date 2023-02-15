package schema

// Display holds the data related to displaying fields.
type Display interface {
	Name() *string
	Description() *string
	Icon() *string
}

// NewDisplayValue creates a new display from the given parameters.
func NewDisplayValue(name *string, description *string, icon *string) *DisplayValue {
	return &DisplayValue{
		NameValue:        name,
		DescriptionValue: description,
		IconValue:        icon,
	}
}

// DisplayValue holds the data related to displaying fields.
type DisplayValue struct {
	NameValue        *string `json:"name"`
	DescriptionValue *string `json:"description"`
	IconValue        *string `json:"icon"`
}

func (d DisplayValue) Name() *string {
	return d.NameValue
}

func (d DisplayValue) Description() *string {
	return d.DescriptionValue
}

func (d DisplayValue) Icon() *string {
	return d.IconValue
}
