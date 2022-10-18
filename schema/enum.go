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

	ValidValues() map[T]*DisplayValue
}

type EnumSchema[T enumValue] struct {
	ValidValuesMap map[T]*DisplayValue `json:"values"`
}

func (e EnumSchema[T]) ValidValues() map[T]*DisplayValue {
	return e.ValidValuesMap
}

func (e EnumSchema[T]) ApplyScope(scope Scope) {
}

func (e EnumSchema[T]) ReflectedType() reflect.Type {
	var defaultValue T
	return reflect.TypeOf(defaultValue)
}

func (e EnumSchema[T]) Validate(d any) error {
	data, err := e.asType(d)
	if err != nil {
		return err
	}
	return e.ValidateType(data)
}

func (e EnumSchema[T]) Serialize(d any) (any, error) {
	data, err := e.asType(d)
	if err != nil {
		return data, err
	}
	return data, e.Validate(data)
}

func (e EnumSchema[T]) ValidateType(data T) error {
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

func (e EnumSchema[T]) SerializeType(data T) (any, error) {
	return data, e.Validate(data)
}

func (e EnumSchema[T]) asType(d any) (T, error) {
	data, ok := d.(T)
	if !ok {
		var defaultValue T
		tType := reflect.TypeOf(defaultValue)
		dValue := reflect.ValueOf(d)
		if !dValue.CanConvert(tType) {
			return defaultValue, &ConstraintError{
				Message: fmt.Sprintf("%T is not a valid data type for an int schema.", d),
			}
		}
		data = dValue.Convert(tType).Interface().(T)
	}
	return data, nil
}
