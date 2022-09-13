package schema

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type enumValue interface {
	int64 | string
}

// EnumSchema is an abstract schema for enumerated types.
type EnumSchema[T enumValue] interface {
	AbstractSchema
	ValidValues() map[T]string
}

type enumSchema[T enumValue] struct {
	ValidValuesMap map[T]string `json:"valid_values"`
}

func (e enumSchema[T]) ValidValues() map[T]string {
	return e.ValidValuesMap
}

// EnumType defines an abstract type for enums.
type EnumType[T enumValue] interface {
	EnumSchema[T]
	AbstractType[T]
}

type enumType[T enumValue, K EnumSchema[T]] struct {
	schemaType K
}

func (e *enumType[T, K]) TypeID() TypeID {
	return e.schemaType.TypeID()
}

func (e *enumType[T, K]) ValidValues() map[T]string {
	return e.schemaType.ValidValues()
}

func (e *enumType[T, K]) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.schemaType)
}

func (e *enumType[T, K]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &e.schemaType)
}

func (e *enumType[T, K]) Validate(data T) error {
	for validValue := range e.schemaType.ValidValues() {
		if validValue == data {
			return nil
		}
	}
	validValues := make([]string, len(e.schemaType.ValidValues()))
	i := 0
	for validValue := range e.schemaType.ValidValues() {
		validValues[i] = fmt.Sprintf("%v", validValue)
		i++
	}
	sort.SliceStable(validValues, func(i, j int) bool {
		return validValues[i] < validValues[j]
	})
	return ConstraintError{
		Message: fmt.Sprintf(
			"'%v' is not a valid value, must be one of: '%s'",
			data,
			strings.Join(validValues, "', '"),
		),
	}
}

func (e *enumType[T, K]) Serialize(data T) (any, error) {
	return data, e.Validate(data)
}
