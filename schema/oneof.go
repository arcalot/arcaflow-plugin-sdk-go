package schema

import (
	"fmt"
	"reflect"
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
		return TypeIDInt
	case string:
		return TypeIDString
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

type oneOfType[KeyType int64 | string, T any] struct {
	oneOfSchema[KeyType, RefType[T]] `json:",inline"`
}

func (o oneOfType[KeyType, T]) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
	for _, t := range o.TypesValue {
		t.ApplyScope(s)
	}
}

func (o oneOfType[KeyType, T]) UnderlyingType() T {
	var defaultValue T
	return defaultValue
}

func (o oneOfType[KeyType, T]) Unserialize(data any) (T, error) {
	discriminatorValue, underlyingType, err := o.findUnderlyingType(data)
	if err != nil {
		var defaultValue T
		return defaultValue, err
	}
	unserializedData, err := underlyingType.Unserialize(data)
	if err != nil {
		var defaultValue T
		return defaultValue, ConstraintErrorAddPathSegment(err, fmt.Sprintf("{oneof[%v]}", discriminatorValue))
	}
	return unserializedData, nil
}

func (o oneOfType[KeyType, T]) Validate(data T) error {
	discriminatorValue, underlyingType, err := o.findUnderlyingType(data)
	if err != nil {
		return err
	}
	if err := underlyingType.Validate(data); err != nil {
		return ConstraintErrorAddPathSegment(err, fmt.Sprintf("{oneof[%v]}", discriminatorValue))
	}
	return nil
}

func (o oneOfType[KeyType, T]) Serialize(data T) (any, error) {
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

func (o oneOfType[KeyType, T]) findUnderlyingType(data any) (KeyType, RefType[T], error) {
	reflectedType := reflect.TypeOf(data)
	if reflectedType.Kind() != reflect.Struct {
		var defaultValue KeyType
		return defaultValue, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type: '%s'",
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
