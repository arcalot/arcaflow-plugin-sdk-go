package schema

// AbstractType describes the common methods all types need to implement.
type AbstractType interface {
    ValidateSerialized(data any, path []string) error
}

// DisplayValue holds the data related to displaying fields.
type DisplayValue struct {
    Name        *string `json:"name" name:"Name" description:"Short text serving as a name or title for this item." examples:"[\"Fruit\"]" min:"1"`
    Description *string `json:"description" name:"Description" description:"Description for this item if needed." examples:"[\"Please select the fruit you would like.\"]" min:"1"`
    Icon        *string `json:"icon" name:"Icon" description:"SVG icon for this item. Must have the declared size of 64x64, must not include additional namespaces, and must not reference external resources." examples:"[\"<svg ...></svg>\"]" min:"1"`
}

type enumValueType interface {
    int | string
}

type enumType[T enumValueType] struct {
}

func (e enumType[T]) ValidateSerialized(data any, path []string) error {
    return nil
}

// EnumStringType is an enum type with string values.
type EnumStringType struct {
    enumType[string]
}

// EnumIntType is an enum type with integer values.
type EnumIntType struct {
    enumType[int]
}

// MapKeyType are types that can be used as map keys.
type MapKeyType interface {
    int64 | string
}

// NumberType is a type collection of number types.
type NumberType interface {
    int64 | float64
}
