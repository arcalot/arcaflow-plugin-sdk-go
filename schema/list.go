package schema

import (
	"fmt"
	"reflect"
)

// ListSchema holds the schema definition for lists. This dataclass only has the ability to hold the configuration but
// cannot serialize, unserialize or validate. For that functionality please use ListType.
type ListSchema interface {
	AbstractSchema
	Items() AbstractSchema
	Min() *int64
	Max() *int64
}

// NewListSchema creates a new list schema from the specified values.
func NewListSchema(items AbstractSchema, min *int64, max *int64) ListSchema {
	return &abstractListSchema[AbstractSchema]{
		items,
		min,
		max,
	}
}

type abstractListSchema[T AbstractSchema] struct {
	ItemsValue T      `json:"items"`
	MinValue   *int64 `json:"min"`
	MaxValue   *int64 `json:"max"`
}

//nolint:unused
type listSchema struct {
	abstractListSchema[AbstractSchema] `json:",inline"`
}

func (l abstractListSchema[T]) TypeID() TypeID {
	return TypeIDList
}

func (l abstractListSchema[T]) Items() AbstractSchema {
	return l.ItemsValue
}

func (l abstractListSchema[T]) Min() *int64 {
	return l.MinValue
}

func (l abstractListSchema[T]) Max() *int64 {
	return l.MaxValue
}

// ListType is the serializable instance of a ListSchema.
type ListType[T any] interface {
	ListSchema
	AbstractType[[]T]

	TypedItems() AbstractType[T]
}

// NewListType defines a serializable list.
func NewListType[T any](items AbstractType[T], min *int64, max *int64) ListType[T] {
	return &listType[T]{
		abstractListSchema[AbstractType[T]]{
			items,
			min,
			max,
		},
	}
}

type listType[T any] struct {
	abstractListSchema[AbstractType[T]] `json:",inline"`
}

func (l listType[T]) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
	l.ItemsValue.ApplyScope(s)
}

func (l listType[T]) UnderlyingType() []T {
	var defaultValue []T
	return defaultValue
}

func (l listType[T]) Unserialize(data any) ([]T, error) {
	var result []T
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

		result = make([]T, v.Len())
		for i := 0; i < v.Len(); i++ {
			unserializedV, err := l.ItemsValue.Unserialize(v.Index(i).Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
			}
			result[i] = unserializedV
		}
		return result, nil
	default:
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must be a slice, %T given", data),
		}
	}
}

func (l listType[T]) Validate(data []T) error {
	if l.MinValue != nil && *l.MinValue > int64(len(data)) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at least %d items, %d given", *l.MinValue, len(data)),
		}
	}
	if l.MaxValue != nil && *l.MaxValue < int64(len(data)) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at most %d items, %d given", *l.MaxValue, len(data)),
		}
	}

	for i := 0; i < len(data); i++ {
		if err := l.ItemsValue.Validate(data[i]); err != nil {
			return ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
		}
	}
	return nil
}

func (l listType[T]) Serialize(data []T) (any, error) {
	if l.MinValue != nil && *l.MinValue > int64(len(data)) {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must have at least %d items, %d given", *l.MinValue, len(data)),
		}
	}
	if l.MaxValue != nil && *l.MaxValue < int64(len(data)) {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must have at most %d items, %d given", *l.MaxValue, len(data)),
		}
	}

	result := make([]any, len(data))
	for i := 0; i < len(data); i++ {
		serialized, err := l.ItemsValue.Serialize(data[i])
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
		}
		result[i] = serialized
	}
	return result, nil
}

func (l listType[T]) TypedItems() AbstractType[T] {
	return l.ItemsValue
}
