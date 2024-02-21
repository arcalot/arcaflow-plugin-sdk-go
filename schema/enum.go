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

func (e EnumSchema[T]) ApplyScope(scope Scope, namespace string) {
}

func (e EnumSchema[T]) ValidateReferences() error {
	// Not applicable
	return nil
}

func (e EnumSchema[T]) ReflectedType() reflect.Type {
	var defaultValue T
	return reflect.TypeOf(defaultValue)
}

func (e EnumSchema[T]) ValidateCompatibility(typeOrData any) error {
	// Check if it's a schema type. If it is, verify it. If not, verify it as data.
	value := reflect.ValueOf(typeOrData)
	if reflect.Indirect(value).Kind() != reflect.Struct {
		// Validate as data
		return e.Validate(typeOrData)
	}
	field := reflect.Indirect(value).FieldByName("EnumSchema")

	if !field.IsValid() {
		// Validate as data
		return e.Validate(typeOrData)
	}

	// Validate the type of EnumSchema
	fieldAsInterface := field.Interface()
	schemaType, ok := fieldAsInterface.(EnumSchema[T])
	if !ok {
		return &ConstraintError{
			Message: fmt.Sprintf(
				"validation failed for enum. Found type (%T) does not match expected type (%T)",
				fieldAsInterface, e),
		}
	}

	// Validate the valid values
	for key, display := range e.ValidValuesMap {
		matchingInputDisplay := schemaType.ValidValuesMap[key]
		if matchingInputDisplay == nil {
			foundValues := reflect.ValueOf(schemaType.ValidValuesMap).MapKeys()
			expectedValues := reflect.ValueOf(e.ValidValuesMap).MapKeys()
			return &ConstraintError{
				Message: fmt.Sprintf("invalid enum values for type '%T' for custom enum. Missing key %v (and potentially others). Expected values: %s, Has values: %s",
					typeOrData, key, expectedValues, foundValues),
			}
		} else if *display.Name() != *matchingInputDisplay.Name() {
			return &ConstraintError{
				Message: fmt.Sprintf(
					"invalid enum value. Mismatched name for key %v. Expected %s, got %s",
					key, *display.Name(), *matchingInputDisplay.Name()),
			}
		}
	}
	return nil

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
