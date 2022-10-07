package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// OneOf is the root interface for one-of types. It should not be used directly but is provided for convenience.
type OneOf[KeyType int64 | string] interface {
	Type

	Types() map[KeyType]Object
	DiscriminatorFieldName() string
}

type OneOfSchema[KeyType int64 | string] struct {
	interfaceType               reflect.Type
	TypesValue                  map[KeyType]Object `json:"types"`
	DiscriminatorFieldNameValue string             `json:"discriminator_field_name"`
}

func (o OneOfSchema[KeyType]) TypeID() TypeID {
	var defaultValue KeyType
	switch any(defaultValue).(type) {
	case int64:
		return TypeIDOneOfInt
	case string:
		return TypeIDOneOfString
	default:
		panic(BadArgumentError{Message: fmt.Sprintf("Unexpected key type: %T", defaultValue)})
	}
}

func (o OneOfSchema[KeyType]) Types() map[KeyType]Object {
	return o.TypesValue
}

func (o OneOfSchema[KeyType]) DiscriminatorFieldName() string {
	return o.DiscriminatorFieldNameValue
}

func (o OneOfSchema[KeyType]) ApplyScope(scope Scope) {
	for _, t := range o.TypesValue {
		t.ApplyScope(scope)
	}
}

func (o OneOfSchema[KeyType]) ReflectedType() reflect.Type {
	if o.interfaceType == nil {
		var defaultValue any
		return reflect.TypeOf(defaultValue)
	}
	return o.interfaceType
}

//nolint:funlen
func (o OneOfSchema[KeyType]) UnserializeType(data any) (result any, err error) {
	reflectedValue := reflect.ValueOf(data)
	if reflectedValue.Kind() != reflect.Map {
		return result, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type for one-of type: '%s'",
				reflect.TypeOf(data).Name(),
			),
		}
	}

	discriminatorValue := reflectedValue.MapIndex(reflect.ValueOf(o.DiscriminatorFieldNameValue))
	if !discriminatorValue.IsValid() {
		return result, &ConstraintError{
			Message: fmt.Sprintf("Missing discriminator field '%s' in '%v'", o.DiscriminatorFieldNameValue, data),
		}
	}
	discriminator := discriminatorValue.Interface()
	typedDiscriminator, err := o.getTypedDiscriminator(discriminator)
	if err != nil {
		return result, err
	}
	typedData := make(map[string]any, reflectedValue.Len())
	for _, k := range reflectedValue.MapKeys() {
		v := reflectedValue.MapIndex(k)
		keyString, ok := k.Interface().(string)
		if !ok {
			return result, &ConstraintError{
				Message: fmt.Sprintf(
					"Invalid key type for one-of: '%T'",
					k.Interface(),
				),
			}
		}
		typedData[keyString] = v.Interface()
	}

	selectedType, ok := o.TypesValue[typedDiscriminator]
	if !ok {
		validDiscriminators := make([]string, len(o.TypesValue))
		i := 0
		for k := range o.TypesValue {
			validDiscriminators[i] = fmt.Sprintf("%v", k)
			i++
		}
		return result, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid value for '%s', expected one of: %s",
				o.DiscriminatorFieldNameValue,
				strings.Join(validDiscriminators, ", "),
			),
		}
	}

	if _, ok := selectedType.Properties()[o.DiscriminatorFieldNameValue]; !ok {
		delete(typedData, o.DiscriminatorFieldNameValue)
	}

	unserializedData, err := selectedType.Unserialize(typedData)
	if err != nil {
		return result, err
	}
	if o.interfaceType == nil {
		return unserializedData, nil
	}
	return saveConvertTo(unserializedData, o.interfaceType)
}

func (o OneOfSchema[KeyType]) ValidateType(data any) error {
	discriminatorValue, underlyingType, err := o.findUnderlyingType(data)
	if err != nil {
		return err
	}
	if err := underlyingType.Validate(data); err != nil {
		return ConstraintErrorAddPathSegment(err, fmt.Sprintf("{oneof[%v]}", discriminatorValue))
	}
	return nil
}

func (o OneOfSchema[KeyType]) SerializeType(data any) (any, error) {
	discriminatorValue, underlyingType, err := o.findUnderlyingType(data)
	if err != nil {
		return nil, err
	}
	serializedData, err := underlyingType.Serialize(data)
	if err != nil {
		return nil, err
	}
	mapData := serializedData.(map[string]any)
	if _, ok := mapData[o.DiscriminatorFieldNameValue]; !ok {
		mapData[o.DiscriminatorFieldNameValue] = discriminatorValue
	}
	return mapData, nil
}

func (o OneOfSchema[KeyType]) Unserialize(data any) (any, error) {
	return o.UnserializeType(data)
}

func (o OneOfSchema[KeyType]) Validate(data any) error {
	if o.interfaceType == nil {
		return o.ValidateType(data)
	}
	d, err := saveConvertTo(data, o.interfaceType)
	if err != nil {
		return err
	}
	return o.ValidateType(d)
}

func (o OneOfSchema[KeyType]) Serialize(data any) (result any, err error) {
	if o.interfaceType == nil {
		return nil, o.ValidateType(data)
	}
	d, err := saveConvertTo(data, o.interfaceType)
	if err != nil {
		return nil, err
	}
	return o.SerializeType(d)
}

func (o OneOfSchema[KeyType]) getTypedDiscriminator(discriminator any) (KeyType, error) {
	var typedDiscriminator KeyType
	switch any(typedDiscriminator).(type) {
	case int64:
		intDiscriminator, err := intInputMapper(discriminator, nil)
		if err != nil {
			return typedDiscriminator, &ConstraintError{
				Message: fmt.Sprintf(
					"Invalid type %T for field %s, expected %T",
					discriminator,
					o.DiscriminatorFieldNameValue,
					typedDiscriminator,
				),
				Cause: err,
			}
		}
		typedDiscriminator = any(intDiscriminator).(KeyType)
	case string:
		stringDiscriminator, err := stringInputMapper(discriminator)
		if err != nil {
			return typedDiscriminator, &ConstraintError{
				Message: fmt.Sprintf(
					"Invalid type %T for field %s, expected %T",
					discriminator,
					o.DiscriminatorFieldNameValue,
					typedDiscriminator,
				),
				Cause: err,
			}
		}
		typedDiscriminator = any(stringDiscriminator).(KeyType)
	}
	return typedDiscriminator, nil
}

func (o OneOfSchema[KeyType]) findUnderlyingType(data any) (KeyType, Object, error) {
	reflectedType := reflect.TypeOf(data)
	if reflectedType.Kind() != reflect.Struct &&
		reflectedType.Kind() != reflect.Map &&
		(reflectedType.Kind() != reflect.Pointer || reflectedType.Elem().Kind() != reflect.Struct) {
		var defaultValue KeyType
		return defaultValue, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type for one-of type: '%s'",
				reflect.TypeOf(data).Name(),
			),
		}
	}

	var foundKey *KeyType
	for key, ref := range o.TypesValue {
		underlyingReflectedType := ref.ReflectedType()
		if underlyingReflectedType == reflectedType {
			keyValue := key
			foundKey = &keyValue
		}
	}
	if foundKey == nil {
		var defaultValue KeyType
		dataType := reflect.TypeOf(data)
		values := make([]string, len(o.TypesValue))
		i := 0
		for _, ref := range o.TypesValue {
			values[i] = ref.ReflectedType().String()
			if values[i] == "" {
				panic(fmt.Errorf("bug: reflected type name is empty"))
			}
			i++
		}
		return defaultValue, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type for one-of schema: '%s' (valid types are: %s)",
				dataType.String(),
				strings.Join(values, ", "),
			),
		}
	}
	return *foundKey, o.TypesValue[*foundKey], nil
}
