package schema

import (
	"fmt"
	"reflect"
)

// List holds the schema definition for lists.
type List[ItemType Type] interface {
	Type
	Items() ItemType
	Min() *int64
	Max() *int64
}

// TypedList extends List by providing typed unserialization.
type TypedList[UnserializedType any, ItemType TypedType[UnserializedType]] interface {
	List[ItemType]
	TypedType[[]UnserializedType]
}

// NewListSchema creates a new list schema from the specified values.
func NewListSchema(items Type, min *int64, max *int64) *ListSchema {
	return &ListSchema{
		AbstractListSchema[Type]{
			items,
			min,
			max,
		},
	}
}

// NewTypedListSchema creates a new list schema from the specified values with typed unserialization.
func NewTypedListSchema[UnserializedType any](
	items TypedType[UnserializedType],
	min *int64,
	max *int64,
) *TypedListSchema[UnserializedType, TypedType[UnserializedType]] {
	return &TypedListSchema[UnserializedType, TypedType[UnserializedType]]{
		AbstractListSchema[TypedType[UnserializedType]]{
			items,
			min,
			max,
		},
	}
}

// ListSchema is the untyped representation of a list.
type ListSchema struct {
	AbstractListSchema[Type] `json:",inline"`
}

// TypedListSchema is the typed variant of the list.
type TypedListSchema[UnserializedType any, ItemType TypedType[UnserializedType]] struct {
	AbstractListSchema[ItemType] `json:",inline"`
}

// AbstractListSchema is a root type for both the untyped and the typed lists.
type AbstractListSchema[ItemType Type] struct {
	ItemsValue ItemType `json:"items"`
	MinValue   *int64   `json:"min"`
	MaxValue   *int64   `json:"max"`
}

func (l AbstractListSchema[ItemType]) TypeID() TypeID {
	return TypeIDList
}

func (l AbstractListSchema[ItemType]) Items() ItemType {
	return l.ItemsValue
}

func (l AbstractListSchema[ItemType]) Min() *int64 {
	return l.MinValue
}

func (l AbstractListSchema[ItemType]) Max() *int64 {
	return l.MaxValue
}

func (l AbstractListSchema[ItemType]) ApplyScope(scope Scope) {
	l.ItemsValue.ApplyScope(scope)
}

func (l AbstractListSchema[ItemType]) ReflectedType() reflect.Type {
	elementType := l.ItemsValue.ReflectedType()
	return reflect.SliceOf(elementType)
}

func (l AbstractListSchema[ItemType]) Unserialize(data any) (any, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice:
		if l.MinValue != nil && *l.MinValue > int64(v.Len()) {
			return nil, &ConstraintError{
				Message: fmt.Sprintf("Must have at least %d items, %d given", *l.MinValue, v.Len()),
			}
		}
		if l.MaxValue != nil && *l.MaxValue < int64(v.Len()) {
			return nil, &ConstraintError{
				Message: fmt.Sprintf("Must have at most %d items, %d given", *l.MaxValue, v.Len()),
			}
		}

		result := reflect.MakeSlice(reflect.SliceOf(l.ItemsValue.ReflectedType()), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			unserializedV, err := l.ItemsValue.Unserialize(v.Index(i).Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
			}
			result.Index(i).Set(reflect.ValueOf(unserializedV))
		}
		return result.Interface(), nil
	default:
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must be a slice, %T given", data),
		}
	}
}

func (l AbstractListSchema[ItemType]) Validate(data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for a slice schema.", data),
		}
	}
	if l.MinValue != nil && *l.MinValue > int64(v.Len()) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at least %d items, %d given", *l.MinValue, v.Len()),
		}
	}
	if l.MaxValue != nil && *l.MaxValue < int64(v.Len()) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at most %d items, %d given", *l.MaxValue, v.Len()),
		}
	}

	for i := 0; i < v.Len(); i++ {
		if err := l.ItemsValue.Validate(v.Index(i).Interface()); err != nil {
			return ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
		}
	}
	return nil
}

func (l AbstractListSchema[ItemType]) Serialize(data any) (any, error) {
	if err := l.Validate(data); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(data)
	result := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		serialized, err := l.ItemsValue.Serialize(v.Index(i).Interface())
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
		}
		result[i] = serialized
	}
	return result, nil
}

func (t TypedListSchema[UnserializedType, ItemType]) UnserializeType(data any) (result []UnserializedType, err error) {
	unserialized, err := t.Unserialize(data)
	if err != nil {
		return result, err
	}
	return unserialized.([]UnserializedType), nil
}

func (t TypedListSchema[UnserializedType, ItemType]) ValidateType(data []UnserializedType) error {
	return t.Validate(data)
}

func (t TypedListSchema[UnserializedType, ItemType]) SerializeType(data []UnserializedType) (any, error) {
	return t.Serialize(data)
}
