package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// OneOfSchema is the root interface for one-of types. It should not be used directly but is provided for convenience.
type OneOfSchema[KeyType int64 | string, RefSchemaType RefSchema] interface {
	AbstractSchema
	Types() map[KeyType]RefSchemaType
	DiscriminatorFieldName() string
}

type oneOfSchema[KeyType int64 | string, RefSchemaType RefSchema] struct {
	TypesValue                  map[KeyType]RefSchemaType `json:"types"`
	DiscriminatorFieldNameValue string                    `json:"discriminator_field_name"`
}

func (o oneOfSchema[KeyType, RefSchemaType]) TypeID() TypeID {
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

func (o oneOfSchema[KeyType, RefSchemaType]) Types() map[KeyType]RefSchemaType {
	return o.TypesValue
}

func (o oneOfSchema[KeyType, RefSchemaType]) DiscriminatorFieldName() string {
	return o.DiscriminatorFieldNameValue
}

type oneOfType[KeyType int64 | string] struct {
	oneOfSchema[KeyType, RefType[any]] `json:",inline"`
}

func (o oneOfType[KeyType]) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
	for _, t := range o.TypesValue {
		t.ApplyScope(s)
	}
}

func (o oneOfType[KeyType]) UnderlyingType() any {
	return nil
}

func (o oneOfType[KeyType]) Unserialize(data any) (any, error) {
	reflectedValue := reflect.ValueOf(data)
	if reflectedValue.Kind() != reflect.Map {
		return nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type for one-of type: '%s'",
				reflect.TypeOf(data).Name(),
			),
		}
	}

	discriminatorValue := reflectedValue.MapIndex(reflect.ValueOf(o.DiscriminatorFieldNameValue))
	if !discriminatorValue.IsValid() {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Missing discriminator field '%s' in '%v'", o.DiscriminatorFieldNameValue, data),
		}
	}
	discriminator := discriminatorValue.Interface()
	var typedDiscriminator KeyType
	switch any(typedDiscriminator).(type) {
	case int64:
		intDiscriminator, err := intInputMapper(discriminator, nil)
		if err != nil {
			return nil, &ConstraintError{
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
			return nil, &ConstraintError{
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
	typedData := data.(map[string]interface{})

	selectedType, ok := o.TypesValue[typedDiscriminator]
	if !ok {
		validDiscriminators := make([]string, len(o.TypesValue))
		i := 0
		for k := range o.TypesValue {
			validDiscriminators[i] = fmt.Sprintf("%v", k)
			i++
		}
		return nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid value for '%s', expected one of: %s",
				o.DiscriminatorFieldNameValue,
				strings.Join(validDiscriminators, ", "),
			),
		}
	}

	if !selectedType.HasProperty(o.DiscriminatorFieldNameValue) {
		delete(typedData, o.DiscriminatorFieldNameValue)
	}

	return selectedType.Unserialize(typedData)
}

func (o oneOfType[KeyType]) Validate(data any) error {
	discriminatorValue, underlyingType, err := o.findUnderlyingType(data)
	if err != nil {
		return err
	}
	if err := underlyingType.Validate(data); err != nil {
		return ConstraintErrorAddPathSegment(err, fmt.Sprintf("{oneof[%v]}", discriminatorValue))
	}
	return nil
}

func (o oneOfType[KeyType]) Serialize(data any) (any, error) {
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

func (o oneOfType[KeyType]) findUnderlyingType(data any) (KeyType, RefType[any], error) {
	reflectedType := reflect.TypeOf(data)
	if reflectedType.Kind() != reflect.Struct {
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
		underlyingType := ref.UnderlyingType()
		if reflect.TypeOf(underlyingType).Kind() == reflectedType.Kind() {
			keyValue := key
			foundKey = &keyValue
		}
	}
	if foundKey == nil {
		var defaultValue KeyType
		return defaultValue, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type: '%s'",
				reflect.TypeOf(data).Name(),
			),
		}
	}
	return *foundKey, o.TypesValue[*foundKey], nil
}
