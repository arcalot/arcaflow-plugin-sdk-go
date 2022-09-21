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
	return &abstractObjectSchema[PropertySchema]{
		id,
		properties,
	}
}

type abstractObjectSchema[T PropertySchema] struct {
	IDValue         string       `json:"id"`
	PropertiesValue map[string]T `json:"properties"`
}

//nolint:unused
type objectSchema struct {
	abstractObjectSchema[*propertySchema]
}

func (o abstractObjectSchema[T]) TypeID() TypeID {
	return TypeIDObject
}

func (o abstractObjectSchema[T]) ID() string {
	return o.IDValue
}

func (o abstractObjectSchema[T]) Properties() map[string]T {
	return o.PropertiesValue
}

// ObjectType is a serializable version of ObjectSchema.
type ObjectType[T any] interface {
	ObjectSchema[PropertyType]
	AbstractType[T]

	Any() ObjectType[any]
}

// NewObjectType creates a serializable representation for an object, for filling structs.
func NewObjectType[T any](id string, properties map[string]PropertyType) ObjectType[T] {
	validateObjectIsStruct[T]()

	return &objectType[T]{
		abstractObjectSchema[PropertyType]{
			id,
			properties,
		},
		extractObjectDefaultValues(properties),
		buildObjectFieldCache[T](properties),
	}
}

func validateObjectIsStruct[T any]() {
	var defaultValue T
	reflectValue := reflect.TypeOf(defaultValue)
	if reflectValue.Kind() == reflect.Pointer {
		reflectValue = reflectValue.Elem()
	}
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
	if reflectType.Kind() == reflect.Pointer {
		reflectType = reflectType.Elem()
	}
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
	abstractObjectSchema[PropertyType] `json:",inline"`
	defaultValues                      map[string]any
	fieldCache                         map[string]reflect.StructField
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

//nolint:funlen
func (o objectType[T]) Unserialize(data any) (result T, err error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Map {
		return result, &ConstraintError{
			Message: fmt.Sprintf("Must be a map, %T given", data),
		}
	}
	tempData, err := o.convertData(v)
	if err != nil {
		return result, err
	}
	if err := o.validateFieldInterdependencies(tempData); err != nil {
		return result, err
	}

	reflectType := reflect.TypeOf(result)
	var reflectedValue reflect.Value
	if reflectType.Kind() != reflect.Pointer {
		reflectedValue = reflect.New(reflectType)
	} else {
		reflectedValue = reflect.New(reflectType.Elem())
	}
	for key, value := range tempData {
		val := value
		elem := reflectedValue.Elem()
		field := elem.FieldByIndex(o.fieldCache[key].Index)
		f := field
		v := reflect.ValueOf(val)
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
			if field.Kind() == reflect.Pointer && v.Kind() != reflect.Pointer {
				f = reflect.New(f.Type().Elem())
				f.Elem().Set(v)
				field.Set(f)
			} else {
				f.Set(v)
			}
		}()
		if recoveredError != nil {
			return result, &ConstraintError{
				"Field cannot be set",
				[]string{key},
				recoveredError,
			}
		}
	}
	reflectType = reflect.TypeOf(result)
	if reflectType.Kind() != reflect.Pointer {
		result = reflectedValue.Elem().Interface().(T)
	} else {
		result = reflectedValue.Interface().(T)
	}
	return result, nil
}

func (o objectType[T]) convertData(v reflect.Value) (map[string]any, error) {
	tempData := make(map[string]any, v.Len())
	for _, key := range v.MapKeys() {
		stringKey, ok := key.Interface().(string)
		if !ok {
			return nil, o.invalidKeyError(key.Interface())
		}
		property, ok := o.PropertiesValue[stringKey]
		if !ok {
			return nil, o.invalidKeyError(stringKey)
		}
		unserializedData, err := property.Unserialize(v.MapIndex(key).Interface())
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, stringKey)
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
	return tempData, nil
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

func (o objectType[T]) Any() ObjectType[any] {
	return &objectTypeAny[T]{
		o,
	}
}

type objectTypeAny[T any] struct {
	objectType[T] `json:",inline"`
}

func (o objectTypeAny[T]) UnderlyingType() any {
	return any(o.objectType.UnderlyingType())
}

func (o objectTypeAny[T]) Unserialize(data any) (any, error) {
	result, err := o.objectType.Unserialize(data)
	return any(result), err
}

func (o objectTypeAny[T]) Validate(data any) error {
	typedData, ok := data.(T)
	if !ok {
		underlyingType := o.objectType.UnderlyingType()
		return &ConstraintError{
			Message: fmt.Sprintf("Invalid type %T for %T", data, underlyingType),
		}
	}
	return o.objectType.Validate(typedData)
}

func (o objectTypeAny[T]) Serialize(data any) (any, error) {
	typedData, ok := data.(T)
	if !ok {
		underlyingType := o.objectType.UnderlyingType()
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Invalid type %T for %T", data, underlyingType),
		}
	}
	return o.objectType.Serialize(typedData)
}

func (o objectTypeAny[T]) Any() ObjectType[any] {
	return o
}
