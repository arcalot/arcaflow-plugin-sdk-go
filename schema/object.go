package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Object holds the definition for objects comprised of defined fields.
type Object interface {
	Type
	ID() string
	Properties() map[string]*PropertySchema
	// GetDefaults returns the defaults in a serialized form.
	GetDefaults() map[string]any
}

// NewObjectSchema creates a new object definition.
// If you need it tied to a struct, use NewStructMappedObjectSchema instead.
func NewObjectSchema(id string, properties map[string]*PropertySchema) *ObjectSchema {
	var anyValue any
	return &ObjectSchema{
		id,
		properties,

		extractObjectDefaultValues(properties),
		nil,
		reflect.TypeOf(anyValue),
		nil,
	}
}

// ObjectSchema is the implementation of the object schema type.
type ObjectSchema struct {
	IDValue         string                     `json:"id"`
	PropertiesValue map[string]*PropertySchema `json:"properties"`

	defaultValues map[string]any // Key: Object field name, value: The default value

	defaultValue     any
	defaultValueType reflect.Type
	fieldCache       map[string]reflect.StructField
}

func (o *ObjectSchema) ReflectedType() reflect.Type {
	if o.fieldCache != nil {
		return o.defaultValueType
	}
	return reflect.TypeOf(map[string]any{})
}

func (o *ObjectSchema) GetDefaults() map[string]any {
	if o.defaultValues == nil {
		o.defaultValues = extractObjectDefaultValues(o.PropertiesValue)
	}
	return o.defaultValues
}

func (o *ObjectSchema) ApplyScope(scope Scope) {
	for _, property := range o.PropertiesValue {
		property.ApplyScope(scope)
	}
}

func (o *ObjectSchema) TypeID() TypeID {
	return TypeIDObject
}

func (o *ObjectSchema) ID() string {
	return o.IDValue
}

func (o *ObjectSchema) Properties() map[string]*PropertySchema {
	return o.PropertiesValue
}

func (o *ObjectSchema) Unserialize(data any) (result any, err error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Map {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must be a map, %T given", data),
		}
	}
	rawData, err := o.convertData(v)
	if err != nil {
		return nil, err
	}
	if err := o.validateFieldInterdependencies(rawData); err != nil {
		return nil, err
	}

	if o.fieldCache != nil {
		return o.unserializeToStruct(rawData)
	}
	return rawData, nil
}

func (o *ObjectSchema) unserializeToStruct(rawData map[string]any) (any, error) {
	reflectType := reflect.TypeOf(o.defaultValue)
	var reflectedValue reflect.Value
	if reflectType.Kind() != reflect.Pointer {
		reflectedValue = reflect.New(reflectType)
	} else {
		reflectedValue = reflect.New(reflectType.Elem())
	}
	for key, value := range rawData {
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
				f.Elem().Set(v.Convert(f.Elem().Type()))
				field.Set(f)
			} else {
				f.Set(v.Convert(f.Type()))
			}
		}()
		if recoveredError != nil {
			return nil, &ConstraintError{
				"Field cannot be set",
				[]string{key},
				recoveredError,
			}
		}
	}
	reflectType = reflect.TypeOf(o.defaultValue)
	var result any
	if reflectType.Kind() != reflect.Pointer {
		result = reflectedValue.Elem().Interface()
	} else {
		result = reflectedValue.Interface()
	}
	return result, nil
}

func (o *ObjectSchema) serializeMap(data map[string]any) (any, error) {
	if err := o.validateFieldInterdependencies(data); err != nil {
		return nil, err
	}

	rawSerializedData := map[string]any{}
	for k, v := range data {
		property, ok := o.PropertiesValue[k]
		if !ok {
			return nil, o.invalidKeyError(k)
		}
		serializedValue, err := property.Serialize(v)
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, k)
		}
		rawSerializedData[k] = serializedValue
	}
	return rawSerializedData, nil
}

func (o *ObjectSchema) serializeStruct(data any) (any, error) {
	if reflect.TypeOf(data) != o.ReflectedType() {
		return o.defaultValue, &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type, expected %s.", data, o.ReflectedType().String()),
		}
	}

	rawData := map[string]any{}
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Nil value passed instead of %T", o.defaultValue),
		}
	}
	for propertyID, property := range o.PropertiesValue {
		a, err := o.extractPropertyValue(propertyID, v, property)
		if err != nil {
			return nil, err
		}
		if a != nil {
			rawData[propertyID] = *a
		}
	}

	if err := o.validateFieldInterdependencies(rawData); err != nil {
		return nil, err
	}

	return rawData, nil
}

func (o *ObjectSchema) extractPropertyValue(propertyID string, v reflect.Value, property *PropertySchema) (*any, error) {
	valPtr := o.getFieldReflection(propertyID, v, property)
	if valPtr == nil {
		return nil, nil
	}
	value := valPtr.Interface()

	if property.emptyIsDefault {
		// Handle the case where the empty value corresponds to the default value.
		defaultValue := reflect.New(property.ReflectedType()).Elem().Convert(valPtr.Type()).Interface()
		if defaultValue == value {
			return nil, nil
		}
	}

	serializedData, err := property.Serialize(value)
	if err != nil {
		return nil, ConstraintErrorAddPathSegment(err, propertyID)
	}
	return &serializedData, nil
}

func (o *ObjectSchema) getFieldReflection(propertyID string, v reflect.Value, property *PropertySchema) *reflect.Value {
	field := o.fieldCache[propertyID]
	var val reflect.Value
	if v.Kind() == reflect.Pointer {
		val = v.Elem().FieldByName(field.Name)
	} else {
		val = v.FieldByName(field.Name)
	}
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil
		}
		if property.ReflectedType().Kind() != reflect.Pointer {
			val = val.Elem()
		}
	}
	if val.Interface() == nil {
		return nil
	}
	return &val
}

func (o *ObjectSchema) Serialize(data any) (any, error) {
	if o.fieldCache != nil {
		return o.serializeStruct(data)
	}
	d, ok := data.(map[string]any)
	if !ok {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for an object schema.", d),
		}
	}
	return o.serializeMap(d)
}

func (o *ObjectSchema) validateMap(data map[string]any) error {
	if err := o.validateFieldInterdependencies(data); err != nil {
		return err
	}
	for k, v := range data {
		property, ok := o.PropertiesValue[k]
		if !ok {
			return o.invalidKeyError(k)
		}
		if err := property.Validate(v); err != nil {
			return ConstraintErrorAddPathSegment(err, k)
		}
	}
	return nil
}
func (o *ObjectSchema) validateMapTypesCompatibility(data map[string]any) error {
	// Note: Interdependencies are not validated here yet.

	// Verify that all present fields match the self schema
	for k, v := range data {
		property, ok := o.PropertiesValue[k]
		if !ok {
			return o.invalidKeyError(k)
		}
		if err := property.ValidateCompatibility(v); err != nil {
			return ConstraintErrorAddPathSegment(err, k)
		}
	}
	// Verify that all required fields are present
	for k, property := range o.PropertiesValue {
		if property.Required() && data[k] == nil {
			return &ConstraintError{
				Message: fmt.Sprintf("error while validating fields of objects %s, could not find required field %s", o.ReflectedType().String(), k),
			}
		}
	}
	return nil
}

func (o *ObjectSchema) validateStruct(data any) error {
	if reflect.TypeOf(data) != o.ReflectedType() {
		return &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type, expected %s.", data, o.ReflectedType().String()),
		}
	}

	rawData := map[string]any{}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return &ConstraintError{
			Message: fmt.Sprintf("Nil value passed instead of %T", o.defaultValue),
		}
	}
	for propertyID, property := range o.PropertiesValue {
		valPtr := o.getFieldReflection(propertyID, v, property)
		if valPtr == nil {
			continue
		}
		value := valPtr.Interface()
		if property.emptyIsDefault {
			// Handle the case where the empty value corresponds to the default value.
			defaultValue := reflect.New(property.ReflectedType()).Elem().Convert(valPtr.Type()).Interface()
			if reflect.DeepEqual(defaultValue, value) {
				continue
			}
		}
		if err := property.Validate(value); err != nil {
			return ConstraintErrorAddPathSegment(err, propertyID)
		}
		rawData[propertyID] = value
	}

	return o.validateFieldInterdependencies(rawData)
}

func (o *ObjectSchema) convertToObjectSchema(typeOrData any) (Object, bool) {
	// Try plain object schema
	objectSchemaType, ok := typeOrData.(*ObjectSchema)
	if ok {
		return objectSchemaType, true
	}
	// Next, try ref schema
	refSchemaType, ok := typeOrData.(*RefSchema)
	if ok {
		return refSchemaType.referencedObjectCache, true
	}
	// Next, try scope schema.
	scopeSchemaType, ok := typeOrData.(*ScopeSchema)
	if ok {
		return scopeSchemaType.Objects()[scopeSchemaType.Root()], true
	}
	// Try getting the inlined ObjectSchema for objects, like TypedObjectSchema, that do that.
	value := reflect.ValueOf(typeOrData)
	if reflect.Indirect(value).Kind() == reflect.Struct {
		field := reflect.Indirect(value).FieldByName("ObjectSchema")
		if field.IsValid() {
			fieldAsInterface := field.Interface()
			objectType, ok2 := fieldAsInterface.(ObjectSchema)
			if ok2 {
				objectSchemaType = &objectType
				ok = true
			}
		}
	}
	return objectSchemaType, ok
}

func (o *ObjectSchema) validateSchemaCompatibility(schemaType Object) error {
	fieldData := map[string]any{}
	// Validate IDs. This is important because the IDs should match.
	if schemaType.ID() != o.ID() {
		return &ConstraintError{
			Message: fmt.Sprintf("validation failed for object schema ID %s. ID %s does not match.",
				o.ID(), schemaType.ID()),
		}
	}
	// Copy all properties to the variable for validating later.
	for key, value := range schemaType.Properties() {
		fieldData[key] = value
	}
	// Now validate object fields
	return o.validateMapTypesCompatibility(fieldData)
}

func (o *ObjectSchema) validateRawCompatibility(typeOrData any) error {
	// Check if it's just a string->interface map. If so, pass it into validateMapTypes
	// Can't validate IDs, but that's acceptable. The only thing that matters in those cases is that the properties match.
	// The reason for that is because we're checking if fields conform to the requirements of the object in this else section.
	if fieldData, ok := typeOrData.(map[string]any); ok {
		// Validate object fields
		return o.validateMapTypesCompatibility(fieldData)
	}
	// Try validating as data
	_, err := o.Unserialize(typeOrData)
	if err != nil {
		return &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type or schema for an object schema (%s)", typeOrData, err),
		}
	} else {
		return nil
	}
}

func (o *ObjectSchema) ValidateCompatibility(typeOrData any) error {
	// Check if it's a schema. If it is, verify it. If not, verify it as data.
	schemaType, ok := o.convertToObjectSchema(typeOrData)
	if ok {
		// It's a schema, so see if the schema matches
		return o.validateSchemaCompatibility(schemaType)
	} else {
		// It's not a schema, so it's ether a map of fields or raw data
		return o.validateRawCompatibility(typeOrData)
	}
}

func (o *ObjectSchema) Validate(data any) error {
	if o.fieldCache != nil {
		return o.validateStruct(data)
	}
	d, ok := data.(map[string]any)
	if !ok {
		return &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for an object schema", d),
		}
	}
	return o.validateMap(d)
}

func (o *ObjectSchema) applySubObjectDefaultValues(propertyID string, property *PropertySchema, rawData map[string]any) {
	reflectedType := property.ReflectedType()
	if reflectedType.Kind() == reflect.Pointer {
		return
	}
	var subObject Object
	switch property.TypeID() {
	case TypeIDRef:
		subObject = property.Type().(Ref).GetObject()
	case TypeIDObject:
		subObject = property.Type().(Object)
	default:
		return
	}
	data := map[string]any{}
	if _, ok := rawData[propertyID]; ok {
		data = rawData[propertyID].(map[string]any)
	}
	subObjectDefaults := subObject.GetDefaults()
	for k, v := range subObjectDefaults {
		data[k] = v
	}
	for subPropertyID, subProperty := range subObject.Properties() {
		o.applySubObjectDefaultValues(subPropertyID, subProperty, data)
	}
	if len(data) != 0 {
		rawData[propertyID] = data
	}
}

func (o *ObjectSchema) convertData(v reflect.Value) (map[string]any, error) {
	rawData := make(map[string]any, v.Len())
	for _, key := range v.MapKeys() {
		stringKey, ok := key.Interface().(string)
		if !ok {
			return nil, o.invalidKeyError(key.Interface())
		}
		if _, ok := o.PropertiesValue[stringKey]; !ok {
			return nil, o.invalidKeyError(stringKey)
		}
		rawData[stringKey] = v.MapIndex(key).Interface()
	}
	for propertyID := range o.PropertiesValue {
		_, isSet := rawData[propertyID]
		if !isSet {
			if defaultValue, ok := o.GetDefaults()[propertyID]; ok {
				rawData[propertyID] = defaultValue
			}
			if o.fieldCache != nil {
				o.applySubObjectDefaultValues(propertyID, o.PropertiesValue[propertyID], rawData)
			}
		}
	}
	for propertyID, property := range o.PropertiesValue {
		if d, ok := rawData[propertyID]; ok {
			unserializedData, err := property.Unserialize(d)
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, propertyID)
			}
			rawData[propertyID] = unserializedData
		}
	}

	return rawData, nil
}

func (o *ObjectSchema) validateFieldInterdependencies(rawData map[string]any) error {
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

func (o *ObjectSchema) validatePropertyInterdependenciesIfUnset(
	rawData map[string]any,
	propertyID string,
	property *PropertySchema,
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

func (o *ObjectSchema) validatePropertyInterdependenciesIfSet(
	rawData map[string]any,
	propertyID string,
	property *PropertySchema,
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

func (o *ObjectSchema) invalidKeyError(value any) error {
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

// TypedObject is a serializable version of Object.
type TypedObject[T any] interface {
	Object
	TypedType[T]

	Any() TypedObject[any]
}

// NewStructMappedObjectSchema creates an object schema that is tied to a specific struct. The values will be mapped to that struct
// when unserialized.
func NewStructMappedObjectSchema[T any](id string, properties map[string]*PropertySchema) *ObjectSchema {
	validateObjectIsStruct[T]()
	var defaultValue T
	return &ObjectSchema{
		IDValue:         id,
		PropertiesValue: properties,

		defaultValues: extractObjectDefaultValues(properties),

		defaultValue:     defaultValue,
		defaultValueType: reflect.TypeOf(&defaultValue).Elem(),
		fieldCache:       buildObjectFieldCache[T](properties),
	}
}

func NewTypedObject[T any](id string, properties map[string]*PropertySchema) *TypedObjectSchema[T] {
	objectSchema := NewStructMappedObjectSchema[T](id, properties)
	return &TypedObjectSchema[T]{
		*objectSchema,
	}
}

type TypedObjectSchema[T any] struct {
	ObjectSchema `json:",inline"`
}

func (t TypedObjectSchema[T]) UnserializeType(data any) (T, error) {
	data, err := t.ObjectSchema.Unserialize(data)
	if err != nil {
		var defaultValue T
		return defaultValue, err
	}
	return data.(T), err
}

func (t TypedObjectSchema[T]) ValidateType(data T) error {
	return t.ObjectSchema.Validate(data)
}

func (t TypedObjectSchema[T]) SerializeType(data T) (any, error) {
	return t.ObjectSchema.Serialize(data)
}

func (t TypedObjectSchema[T]) Any() TypedObject[any] {
	return &AnyTypedObject[T]{
		t.ObjectSchema,
	}
}

// AnyTypedObject is an object that pretends to be typed, but accepts any type.
type AnyTypedObject[T any] struct {
	ObjectSchema `json:",inline"`
}

func (a *AnyTypedObject[T]) UnserializeType(data any) (any, error) {
	return a.ObjectSchema.Unserialize(data)
}

func (a *AnyTypedObject[T]) ValidateType(data any) error {
	return a.ObjectSchema.Validate(data)
}

func (a *AnyTypedObject[T]) SerializeType(data any) (any, error) {
	return a.ObjectSchema.Serialize(data)
}

func (a *AnyTypedObject[T]) Any() TypedObject[any] {
	return a
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
				"NewStructMappedObjectSchema should only be called with a struct type, %T given",
				defaultValue,
			),
		})
	}
}

func extractObjectDefaultValues(properties map[string]*PropertySchema) map[string]any {
	defaultValues := map[string]any{}
	for propertyID, property := range properties {
		if property.Default() != nil {
			var value any
			defaultValue := *property.Default()
			propertyType := property.TypeID()
			err := jsonUnmarshal(defaultValue, &value, propertyType)
			if err != nil {
				panic(BadArgumentError{
					Message: fmt.Sprintf("Default value for property %s is not a valid JSON", propertyID),
					Cause:   err,
				})
			}
			defaultValues[propertyID] = value
		}
	}
	return defaultValues
}

func jsonUnmarshal(defaultValue string, value any, propertryType TypeID) error {
	err := json.Unmarshal([]byte(defaultValue), &value)
	if err != nil && propertryType == "string" {
		// attempt to fix yaml string to valid JSON
		defaultValueTypeString := ("\"" + defaultValue + "\"")
		err2 := json.Unmarshal([]byte(defaultValueTypeString), &value)
		if err2 != nil {
			return fmt.Errorf("{%s} additional attempt to format string with additional quotes failed:{%s}",
				err.Error(), err2.Error())
		} else {
			return nil
		}
	}
	return err
}

func buildObjectFieldCache[T any](properties map[string]*PropertySchema) map[string]reflect.StructField {
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
