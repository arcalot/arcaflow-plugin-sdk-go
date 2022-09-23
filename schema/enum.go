package schema

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type enumValue interface {
	int64 | string
}

// Enum is an abstract schema for enumerated types.
type Enum[T enumValue] interface {
	TypedType[T]

	ValidValues() map[T]string
}

type EnumSchema[T enumValue] struct {
	ValidValuesMap map[T]string `json:"values"`
}

func (e EnumSchema[T]) ValidValues() map[T]string {
	return e.ValidValuesMap
}

func (e EnumSchema[T]) ApplyScope(scope Scope) {
}

func (e EnumSchema[T]) ReflectedType() reflect.Type {
	var defaultValue T
	return reflect.TypeOf(defaultValue)
}

func (e EnumSchema[T]) Validate(data any) error {
	if _, ok := data.(T); !ok {
		return &ConstraintError{
			Message: fmt.Sprintf(
				"%T is not a valid for an enum of %s",
				data,
				e.ReflectedType().Name(),
			),
		}
	}
	for validValue := range e.ValidValuesMap {
		if validValue == data {
			return nil
		}
	}
	validValues := make([]string, len(e.ValidValuesMap))
	i := 0
	for validValue := range e.ValidValuesMap {
		validValues[i] = fmt.Sprintf("%v", validValue)
		i++
	}
	sort.SliceStable(validValues, func(i, j int) bool {
		return validValues[i] < validValues[j]
	})
	return &ConstraintError{
		Message: fmt.Sprintf(
			"'%v' is not a valid value, must be one of: '%s'",
			data,
			strings.Join(validValues, "', '"),
		),
	}
}

func (e EnumSchema[T]) Serialize(data any) (any, error) {
	return data, e.Validate(data)
}

func (e EnumSchema[T]) ValidateType(data T) error {
	return e.Validate(data)
}

func (e EnumSchema[T]) SerializeType(data T) (any, error) {
	return e.Serialize(data)
}
