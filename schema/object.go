package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ObjectSchema holds the definition for objects comprised of defined fields. This dataclass only has the ability to hold
// the configuration but cannot serialize, unserialize or validate. For that functionality please use
// PropertyType.
type ObjectSchema[T PropertySchema] interface {
	AbstractSchema
	ID() string
	Properties() map[string]T
}

// NewObjectSchema creates a new object definition.
func NewObjectSchema(id string, properties map[string]PropertySchema) ObjectSchema[PropertySchema] {
	return &objectSchema[PropertySchema]{
		id,
		properties,
	}
}

type objectSchema[T PropertySchema] struct {
	IDValue         string       `json:"id"`
	PropertiesValue map[string]T `json:"properties"`
}

func (o objectSchema[T]) TypeID() TypeID {
	return TypeIDObject
}

func (o objectSchema[T]) ID() string {
	return o.IDValue
}

func (o objectSchema[T]) Properties() map[string]T {
	return o.PropertiesValue
}

// ObjectType is a serializable version of ObjectSchema.
type ObjectType[T any] interface {
	ObjectSchema[PropertyType]
	AbstractType[T]

	Anonymous() ObjectType[any]
}

// NewObjectType creates a serializable representation for an object, for filling structs.
func NewObjectType[T any](id string, properties map[string]PropertyType) ObjectType[T] {
	validateObjectIsStruct[T]()

	return &objectType[T]{
		objectSchema[PropertyType]{
			id,
			properties,
		},
		extractObjectDefaultValues(properties),
		buildObjectFieldCache[T](properties),
	}
}

func validateObjectIsStruct[T any]() {
	var defaultValue T
	reflectValue := reflect.ValueOf(defaultValue)
	if reflectValue.Kind() != reflect.Struct {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"NewObjectType should only be called with a struct type, %T given",
				defaultValue,
			),
		})
	}
}

func buildObjectFieldCache[T any](properties map[string]PropertyType) map[string]reflect.StructField {
	var defaultValue T
	fieldCache := make(map[string]reflect.StructField, len(properties))
	reflectType := reflect.TypeOf(defaultValue)
	for propertyID := range properties {
		field, ok := reflectType.FieldByNameFunc(func(s string) bool {
			fieldType, _ := reflectType.FieldByName(s)
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag != "" {
				parts := strings.SplitN(jsonTag, ",", 2)
				if parts[0] == propertyID {
					return true
				}
			}
			return false
		})
		if !ok {
			field, ok = reflectType.FieldByName(propertyID)
			if !ok {
				panic(BadArgumentError{
					Message: fmt.Sprintf(
						"Cannot find a valid field to set for '%s' on '%s'. Please name a field identically "+
							"or provide a `json:\"%s\"` tag.",
						propertyID,
						reflectType.Name(),
						propertyID,
					),
				})
			}
		}
		fieldCache[propertyID] = field
	}
	return fieldCache
}

func extractObjectDefaultValues(properties map[string]PropertyType) map[string]any {
	defaultValues := map[string]any{}
	for propertyID, property := range properties {
		if property.Default() != nil {
			var value any
			if err := json.Unmarshal([]byte(*property.Default()), &value); err != nil {
				panic(BadArgumentError{
					Message: fmt.Sprintf("Default value for property %s is not a valid JSON", propertyID),
					Cause:   err,
				})
			}
			unserializedDefault, err := property.Unserialize(value)
			if err != nil {
				panic(BadArgumentError{
					Message: fmt.Sprintf("Default value for property %s is not a unserializable", propertyID),
					Cause:   err,
				})
			}
			defaultValues[propertyID] = unserializedDefault
		}
	}
	return defaultValues
}

type objectType[T any] struct {
	objectSchema[PropertyType] `json:",inline"`
	defaultValues              map[string]any
	fieldCache                 map[string]reflect.StructField
}

func (o objectType[T]) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
	for _, property := range o.PropertiesValue {
		property.ApplyScope(s)
	}
}

func (o objectType[T]) UnderlyingType() T {
	var defaultValue T
	return defaultValue
}

func (o objectType[T]) Unserialize(data any) (result T, err error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Map {
		return result, &ConstraintError{
			Message: fmt.Sprintf("Must be a map, %T given", data),
		}
	}
	tempData := make(map[string]any, v.Len())
	for _, key := range v.MapKeys() {
		stringKey, ok := key.Interface().(string)
		if !ok {
			return result, o.invalidKeyError(key.Interface())
		}
		property, ok := o.PropertiesValue[stringKey]
		if !ok {
			return result, o.invalidKeyError(stringKey)
		}
		unserializedData, err := property.Unserialize(v.MapIndex(key).Interface())
		if err != nil {
			return result, ConstraintErrorAddPathSegment(err, stringKey)
		}
		tempData[stringKey] = unserializedData
	}
	for propertyID := range o.PropertiesValue {
		_, isSet := tempData[propertyID]
		if !isSet {
			if defaultValue, ok := o.defaultValues[propertyID]; ok {
				tempData[propertyID] = defaultValue
			}
		}
	}
	if err := o.validateFieldInterdependencies(tempData); err != nil {
		return result, err
	}

	reflectValue := reflect.ValueOf(&result)
	for key, value := range tempData {
		v := reflect.ValueOf(value)
		var recoveredError error
		func() {
			defer func() {
				e := recover()
				if e != nil {
					var ok bool
					recoveredError, ok = e.(error)
					if !ok {
						recoveredError = fmt.Errorf("%v", e)
					}
				}
			}()
			reflectValue.Elem().FieldByName(o.fieldCache[key].Name).Set(v)
		}()
		if recoveredError != nil {
			return result, &ConstraintError{
				"Field cannot be set",
				[]string{key},
				recoveredError,
			}
		}
	}
	return result, nil
}

func (o objectType[T]) validateFieldInterdependencies(rawData map[string]any) error {
	for propertyID, property := range o.PropertiesValue {
		if _, isSet := rawData[propertyID]; isSet {
			if err := o.validatePropertyInterdependenciesIfSet(rawData, propertyID, property); err != nil {
				return err
			}
		} else {
			err := o.validatePropertyInterdependenciesIfUnset(rawData, propertyID, property)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o objectType[T]) validatePropertyInterdependenciesIfUnset(
	rawData map[string]any,
	propertyID string,
	property PropertyType,
) error {
	if property.Required() {
		return &ConstraintError{
			Message: "This field is required",
			Path:    []string{propertyID},
		}
	}
	for _, requiredIf := range property.RequiredIf() {
		if _, set := rawData[requiredIf]; set {
			return &ConstraintError{
				Message: fmt.Sprintf(
					"This field is required because '%s' is set",
					requiredIf,
				),
				Path: []string{propertyID},
			}
		}
	}
	if len(property.RequiredIfNot()) > 0 {
		foundSet := false
		for _, requiredIfNot := range property.RequiredIfNot() {
			if _, set := rawData[requiredIfNot]; set {
				foundSet = true
				break
			}
		}
		if !foundSet {
			if len(property.RequiredIfNot()) == 1 {
				return &ConstraintError{
					Message: fmt.Sprintf(
						"This field is required because '%s' is not set",
						property.RequiredIfNot()[0],
					),
					Path: []string{propertyID},
				}
			}
			return &ConstraintError{
				Message: fmt.Sprintf(
					"This field is required because none of '%s' are set",
					strings.Join(property.RequiredIfNot(), "', '"),
				),
				Path: []string{propertyID},
			}
		}
	}
	return nil
}

func (o objectType[T]) validatePropertyInterdependenciesIfSet(
	rawData map[string]any,
	propertyID string,
	property PropertyType,
) error {
	for _, conflict := range property.Conflicts() {
		if _, set := rawData[conflict]; set {
			return &ConstraintError{
				Message: fmt.Sprintf(
					"Field conflicts '%s', set one of the two, not both",
					conflict,
				),
				Path: []string{propertyID},
			}
		}
	}
	return nil
}

func (o objectType[T]) invalidKeyError(value any) error {
	validKeys := make([]string, len(o.PropertiesValue))
	i := 0
	for k := range o.PropertiesValue {
		validKeys[i] = k
		i++
	}
	return &ConstraintError{
		Message: fmt.Sprintf(
			"Invalid parameter '%v', expected one of: %s",
			value,
			strings.Join(validKeys, ", "),
		),
	}
}

func (o objectType[T]) Validate(data T) error {
	rawData := map[string]any{}

	v := reflect.ValueOf(data)
	for propertyID, property := range o.PropertiesValue {
		field := o.fieldCache[propertyID]
		val := v.FieldByName(field.Name)
		if err := property.Validate(val.Interface()); err != nil {
			return ConstraintErrorAddPathSegment(err, propertyID)
		}
		rawData[propertyID] = val.Interface()
	}

	return o.validateFieldInterdependencies(rawData)
}

func (o objectType[T]) Serialize(data T) (any, error) {
	rawData := map[string]any{}

	v := reflect.ValueOf(data)
	for propertyID, property := range o.PropertiesValue {
		field := o.fieldCache[propertyID]
		val := v.FieldByName(field.Name)
		serializedData, err := property.Serialize(val.Interface())
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, propertyID)
		}

		if defaultValue, ok := o.defaultValues[propertyID]; !ok || defaultValue == serializedData {
			rawData[propertyID] = serializedData
		}
	}

	return rawData, o.validateFieldInterdependencies(rawData)
}

func (o objectType[T]) Anonymous() ObjectType[any] {
	return &objectTypeAnonymous[T]{
		o,
	}
}

type objectTypeAnonymous[T any] struct {
	objectType[T] `json:",inline"`
}

func (o objectTypeAnonymous[T]) UnderlyingType() any {
	return any(o.objectType.UnderlyingType())
}

func (o objectTypeAnonymous[T]) Unserialize(data any) (any, error) {
	result, err := o.objectType.Unserialize(data)
	return any(result), err
}

func (o objectTypeAnonymous[T]) Validate(data any) error {
	typedData, ok := data.(T)
	if !ok {
		underlyingType := o.objectType.UnderlyingType()
		return &ConstraintError{
			Message: fmt.Sprintf("Invalid type %T for %T", data, underlyingType),
		}
	}
	return o.objectType.Validate(typedData)
}

func (o objectTypeAnonymous[T]) Serialize(data any) (any, error) {
	typedData, ok := data.(T)
	if !ok {
		underlyingType := o.objectType.UnderlyingType()
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Invalid type %T for %T", data, underlyingType),
		}
	}
	return o.objectType.Serialize(typedData)
}

func (o objectTypeAnonymous[T]) Anonymous() ObjectType[any] {
	return o
}
