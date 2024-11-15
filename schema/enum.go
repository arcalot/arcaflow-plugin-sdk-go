package schema

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type serializedEnumValue interface {
	int64 | string
}
type enumValue interface {
	~int64 | ~string
}

// Enum is an abstract schema for enumerated types.
type Enum[T enumValue] interface {
	TypedType[T]

	ValidValues() map[T]*DisplayValue
}

type EnumSchema[S serializedEnumValue, T enumValue] struct {
	ScalarType
	ValidValuesMap map[T]*DisplayValue `json:"values"`
}

func (e EnumSchema[S, T]) ValidValues() map[T]*DisplayValue {
	return e.ValidValuesMap
}

func (e EnumSchema[S, T]) ReflectedType() reflect.Type {
	var defaultValue T
	return reflect.TypeOf(defaultValue)
}

func (e EnumSchema[S, T]) ValidateCompatibility(typeOrData any) error {
	// Check if it's a schema type. If it is, verify it. If not, verify it as data.
	value := reflect.ValueOf(typeOrData)
	if reflect.Indirect(value).Kind() != reflect.Struct {
		return e.Validate(typeOrData) // Validate as data
	}
	enumField := reflect.Indirect(value).FieldByName("EnumSchema")
	if !enumField.IsValid() {
		return e.Validate(typeOrData) // Validate as data
	}

	validValuesMapField := enumField.FieldByName("ValidValuesMap")
	if !validValuesMapField.IsValid() {
		return fmt.Errorf("failed to get values map in enum %T", e)
	}
	for _, reflectKey := range validValuesMapField.MapKeys() {
		var defaultValue T
		defaultType := reflect.TypeOf(defaultValue)
		if !reflectKey.CanConvert(defaultType) {
			return fmt.Errorf("invalid enum value type %s", reflectKey.Type())
		}
		keyToCompare := reflectKey.Convert(defaultType).Interface()
		// Validate that the key in the data under test is present in the self enum schema.
		selfDisplayValue, found := e.ValidValuesMap[keyToCompare.(T)]
		if !found {
			return &ConstraintError{
				Message: fmt.Sprintf("invalid enum values for type '%T' for custom enum of type %T. "+
					"Found key %v. Expected values: %v",
					e, typeOrData, keyToCompare, e.ValidValuesMap),
			}
		}
		// Validate that the displays are compatible.
		otherDisplay := validValuesMapField.MapIndex(reflectKey)
		if !otherDisplay.IsValid() {
			return fmt.Errorf("failed to get value at key in ValidateCompatibility")
		}
		otherDisplayValue := otherDisplay.Interface().(*DisplayValue)
		switch {
		case (selfDisplayValue == nil || selfDisplayValue.Name() == nil) &&
			(otherDisplayValue == nil || otherDisplayValue.Name() == nil):
			return nil
		case otherDisplayValue == nil || otherDisplayValue.Name() == nil:
			return &ConstraintError{
				Message: fmt.Sprintf("display values for key %s is missing in compared data %T",
					keyToCompare, typeOrData),
			}
		case selfDisplayValue == nil || selfDisplayValue.Name() == nil:
			return &ConstraintError{
				Message: fmt.Sprintf("display values for key %s is missing in the schema for %T, but present"+
					" in compared data %T", keyToCompare, e, typeOrData),
			}
		case *selfDisplayValue.Name() != *otherDisplayValue.Name():
			return &ConstraintError{
				Message: fmt.Sprintf(
					"invalid enum value. Mismatched name for key %v. Expected %s, got %s",
					keyToCompare, *selfDisplayValue.Name(), *otherDisplayValue.Name()),
			}
		}
	}
	return nil
}

func (e EnumSchema[S, T]) Validate(d any) error {
	_, data, err := e.asType(d)
	if err != nil {
		return err
	}
	return e.ValidateType(data)
}

func (e EnumSchema[S, T]) Serialize(d any) (any, error) {
	serializedData, data, err := e.asType(d)
	if err != nil {
		return serializedData, err
	}
	return serializedData, e.Validate(data)
}

func (e EnumSchema[S, T]) ValidateType(data T) error {
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

func (e EnumSchema[S, T]) SerializeType(data T) (any, error) {
	return data, e.Validate(data)
}

func (e EnumSchema[S, T]) asType(d any) (S, T, error) {
	var serializedDefaultValue S
	serializedType := reflect.TypeOf(serializedDefaultValue)
	dValue := reflect.ValueOf(d)
	var unserializedDefaultValue T
	unserializedType := reflect.TypeOf(unserializedDefaultValue)

	if !dValue.CanConvert(serializedType) {
		return serializedDefaultValue, unserializedDefaultValue, &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for an %T schema.", d, serializedDefaultValue),
		}
	}
	if !dValue.CanConvert(unserializedType) {
		return serializedDefaultValue, unserializedDefaultValue, &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for an %T schema's unserialized type %T", d, e, unserializedType),
		}
	}
	serializedData := dValue.Convert(serializedType).Interface().(S)
	unserializedData := dValue.Convert(unserializedType).Interface().(T)
	return serializedData, unserializedData, nil
}
