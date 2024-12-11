package schema

import (
	"fmt"
	"reflect"
)

// NewAnySchema creates an AnySchema which is a wildcard allowing maps, lists, integers, strings, bools. and floats.
func NewAnySchema() *AnySchema {
	return &AnySchema{}
}

// AnySchema is a wildcard allowing maps, lists, integers, strings, bools. and floats.
type AnySchema struct {
	ScalarType
}

func (a *AnySchema) ReflectedType() reflect.Type {
	var defaultValue any
	return reflect.TypeOf(&defaultValue).Elem()
}

func (a *AnySchema) Unserialize(data any) (any, error) {
	return a.checkAndConvert(data)
}

func (a *AnySchema) validateSchemaCompatibility(schema Type) error {
	switch schema.ReflectedType().Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.String:
		fallthrough
	case reflect.Bool:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Map:
		return nil
	default:
		// Schema is not a primitive, slice, or map type, so check the complex types
		// Explicitly allow object schemas since their reflected type can be a struct if they are struct mapped.
		switch schema.(type) {
		case *AnySchema, *OneOfSchema[int64], *OneOfSchema[string], *ObjectSchema:
			// These are the allowed values.
		default:
			// It's not an any schema or a type compatible with an any schema, so error
			return &ConstraintError{
				Message: fmt.Sprintf("schema type `%T` cannot be used as an input for an 'any' type", schema),
			}
		}
		return nil
	}
}

func (a *AnySchema) validateAnyMap(data map[any]any) error {
	// Test individual values
	var firstReflectKind reflect.Kind
	for key, value := range data {
		// Validate key
		reflectKind := reflect.ValueOf(key).Kind()
		switch reflectKind {
		// While it is possible to add more types of ints, it's likely better to keep it consistent with i64
		case reflect.Int64:
			fallthrough
		case reflect.String:
			// Valid type
		default:
			return &ConstraintError{
				Message: fmt.Sprintf("invalid key type for map passed into 'any' type (%s); should be string or i64", reflectKind),
			}
		}
		if firstReflectKind == reflect.Invalid { // First item
			firstReflectKind = reflectKind
		} else if firstReflectKind != reflectKind {
			return &ConstraintError{
				Message: fmt.Sprintf(
					"mismatched key types in map passed into 'any' type: %s != %s",
					firstReflectKind, reflectKind),
			}
		}

		// Validate value
		err := a.ValidateCompatibility(value)
		if err != nil {
			return &ConstraintError{
				Message: fmt.Sprintf("validation error while validating map item item %q of map for 'any' type (%s)", key, err.Error()),
			}
		}
	}
	return nil
}

//nolint:nestif
func (a *AnySchema) validateAnyList(data []any) error {
	if len(data) == 0 {
		return nil // No items to check, and following code assumes non-empty list
	}
	// Test list items
	for _, item := range data {
		err := a.ValidateCompatibility(item)
		if err != nil {
			return &ConstraintError{
				Message: fmt.Sprintf("validation error while validating list item of type `%T` in any type (%s)", item, err.Error()),
			}
		}
	}
	// validate that all list items are compatible with the first to make the list homogeneous.
	firstItem := data[0]
	firstItemType, firstValIsSchema := firstItem.(Type)
	if firstValIsSchema {
		for i := 1; i < len(data); i++ {
			valToTest := data[i]
			err := firstItemType.ValidateCompatibility(valToTest)
			if err != nil {
				return &ConstraintError{
					Message: fmt.Sprintf(
						"validation error while validating for homogeneous list item `%T` in any type %s",
						valToTest, err.Error()),
				}
			}
		}
	} else {
		// Loop through all items. Ensure they have the same type.
		firstItemType := reflect.ValueOf(firstItem).Kind()
		for i := 1; i < len(data); i++ {
			valToTest := data[i]
			typeToTest := reflect.ValueOf(valToTest).Kind()
			if firstItemType != typeToTest {
				// Not compatible or is a schema
				schemaType, valIsSchema := valToTest.(Type)
				if !valIsSchema {
					return &ConstraintError{
						Message: fmt.Sprintf(
							"types do not match between list items passed for any type %T != %T; "+
								"lists should have homogeneous types",
							firstItem, valToTest),
					}
				} else {
					err := schemaType.ValidateCompatibility(valToTest)
					if err != nil {
						return &ConstraintError{
							Message: fmt.Sprintf(
								"types do not match between list items passed for any type %s; "+
									"lists should have homogeneous types",
								err),
						}
					}
				}
			}
		}
	}
	return nil
}

func (a *AnySchema) ValidateCompatibility(typeOrData any) error {
	switch typeOrData := typeOrData.(type) {
	case Type:
		return a.validateSchemaCompatibility(typeOrData)
	case map[string]any:
		// Test individual values
		for key, value := range typeOrData {
			err := a.ValidateCompatibility(value)
			if err != nil {
				return &ConstraintError{
					Message: fmt.Sprintf("validation error while validating map item item %q of map for 'any' type (%s)", key, err.Error()),
				}
			}
		}
		return nil
	case map[int64]any:
		// Test individual values
		for key, value := range typeOrData {
			err := a.ValidateCompatibility(value)
			if err != nil {
				return &ConstraintError{
					Message: fmt.Sprintf("validation error while validating map item item %q of map for 'any' type (%s)", key, err.Error()),
				}
			}
		}
		return nil
	case map[any]any:
		return a.validateAnyMap(typeOrData)
	case []interface{}:
		return a.validateAnyList(typeOrData)
	default:
		_, err := a.Unserialize(typeOrData)
		return err
	}
}

func (a *AnySchema) Validate(data any) error {
	_, err := a.checkAndConvert(data)
	return err
}

func (a *AnySchema) Serialize(data any) (any, error) {
	return a.checkAndConvert(data)
}

func (a *AnySchema) TypeID() TypeID {
	return TypeIDAny
}

//nolint:funlen
func (a *AnySchema) checkAndConvert(data any) (any, error) {
	t := reflect.ValueOf(data)
	switch t.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return intInputMapper(data, nil)
	case reflect.Int64:
		return data.(int64), nil
	case reflect.Float32:
		return floatInputMapper(data, nil)
	case reflect.Float64:
		return asFloat(data)
	case reflect.String:
		return data.(string), nil
	case reflect.Bool:
		return asBool(data)
	case reflect.Slice:
		result := make([]any, t.Len())
		for i := 0; i < t.Len(); i++ {
			val, err := a.checkAndConvert(t.Index(i).Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
			}
			result[i] = val
		}
		return result, nil
	case reflect.Map:
		result := make(map[any]any, t.Len())
		for _, k := range t.MapKeys() {
			key, err := a.checkAndConvert(k.Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("{%v}", k))
			}
			v := t.MapIndex(k)
			value, err := a.checkAndConvert(v.Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%v]", key))
			}
			result[key] = value
		}
		return result, nil
	default:
		return nil, &ConstraintError{
			Message: fmt.Sprintf("unsupported data type for 'any' type: %T", data),
		}
	}
}
