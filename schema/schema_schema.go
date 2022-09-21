package schema

import "regexp"

var unitsProperty = NewPropertyType[*units](
	NewRefType[*units]("Units", nil),
	NewDisplayValue(
		PointerTo("Units"),
		PointerTo("Units this number represents."),
		nil,
	),
	false,
	nil,
	nil,
	nil,
	nil,
	[]string{
		"{" +
			"   \"base_unit\": {" +
			"       \"name_short_singular\": \"%\"," +
			"       \"name_short_plural\": \"%\"," +
			"       \"name_long_singular\": \"percent\"," +
			"       \"name_long_plural\": \"percent\"" +
			"   }" +
			"}",
	},
)
var idType = NewStringType(
	IntPointer(1),
	IntPointer(255),
	regexp.MustCompile("^[$@a-zA-Z0-9-_]+$"),
)
var mapKeyType = NewOneOfStringType(
	map[string]RefType[any]{
		"integer": NewRefType[*intSchema](
			"IntSchema",
			NewDisplayValue(
				PointerTo("Integer"),
				nil,
				nil,
			),
		).Any(),
		"string": NewRefType[*stringSchema](
			"StringSchema",
			NewDisplayValue(
				PointerTo("String"),
				nil,
				nil,
			),
		).Any(),
	},
	"type_id",
)
var displayType = NewDisplayValue(
	PointerTo("DisplayValue"),
	PointerTo(
		"Name, description and icon.",
	),
	nil,
)
var displayProperty = NewPropertyType[*displayValue](
	NewRefType[*displayValue](
		"DisplayValue",
		nil,
	),
	displayType,
	false,
	nil,
	nil,
	nil,
	nil,
	nil,
)
var valueType = NewOneOfStringType(
	map[string]RefType[any]{
		"bool": NewRefType[*boolSchema](
			"BoolSchema",
			NewDisplayValue(
				PointerTo("Bool"),
				nil,
				nil,
			),
		).Any(),
		"enum_integer": NewRefType[*intEnumSchema](
			"IntEnumSchema",
			NewDisplayValue(
				PointerTo("Integer enum"),
				nil,
				nil,
			),
		).Any(),
		"enum_string": NewRefType[*stringEnumSchema](
			"StringEnumSchema",
			NewDisplayValue(
				PointerTo("String enum"),
				nil,
				nil,
			),
		).Any(),
		"float": NewRefType[*floatSchema](
			"FloatSchema",
			NewDisplayValue(
				PointerTo("Float"),
				nil,
				nil,
			),
		).Any(),
		"integer": NewRefType[*intSchema](
			"IntSchema",
			NewDisplayValue(
				PointerTo("Integer"),
				nil,
				nil,
			),
		).Any(),
		"list": NewRefType[*abstractListSchema[AbstractSchema]](
			"ListSchema",
			NewDisplayValue(
				PointerTo("List"),
				nil,
				nil,
			),
		).Any(),
		"map": NewRefType[*mapSchema](
			"MapSchema",
			NewDisplayValue(
				PointerTo("Map"),
				nil,
				nil,
			),
		).Any(),
		"object": NewRefType[*objectSchema](
			"ObjectSchema",
			NewDisplayValue(
				PointerTo("Object"),
				nil,
				nil,
			),
		).Any(),
		"one_of_int": NewRefType[*oneOfSchema[int64, *refSchema]](
			"OneOfIntSchema",
			NewDisplayValue(
				PointerTo("Multiple with int key"),
				nil,
				nil,
			),
		).Any(),
		"one_of_string": NewRefType[*oneOfSchema[string, *refSchema]](
			"OneOfStringSchema",
			NewDisplayValue(
				PointerTo("Multiple with string key"),
				nil,
				nil,
			),
		).Any(),
		"pattern": NewRefType[*patternSchema](
			"PatternSchema",
			NewDisplayValue(
				PointerTo("Pattern"),
				nil,
				nil,
			),
		).Any(),
		"ref": NewRefType[*refSchema](
			"RefSchema",
			NewDisplayValue(
				PointerTo("Object reference"),
				nil,
				nil,
			),
		).Any(),
		"scope": NewRefType[*scopeSchema](
			"ScopeSchema",
			NewDisplayValue(
				PointerTo("Scope"),
				nil,
				nil,
			),
		).Any(),
		"string": NewRefType[*stringSchema](
			"StringSchema",
			NewDisplayValue(
				PointerTo("String"),
				nil,
				nil,
			),
		).Any(),
	},
	"type_id",
)

// SchemaSchema is the definition of a schema itself.
var SchemaSchema = NewScopeType[*schema](
	map[string]ObjectType[any]{
		"BoolSchema": NewObjectType[*boolSchema]("BoolSchema", map[string]PropertyType{}).Any(),
		"DisplayValue": NewObjectType[*displayValue]("DisplayValue", map[string]PropertyType{
			"name": NewPropertyType[string](
				NewStringType(IntPointer(1), nil, nil),
				NewDisplayValue(
					PointerTo("Name"),
					PointerTo("Short text serving as a name or title for this item."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				[]string{"\"Fruit\""},
			),
			"description": NewPropertyType[string](
				NewStringType(IntPointer(1), nil, nil),
				NewDisplayValue(
					PointerTo("Description"),
					PointerTo("Description for this item if needed."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				[]string{"\"Please select the fruit you would like.\""},
			),
			"icon": NewPropertyType[string](
				NewStringType(IntPointer(1), nil, nil),
				NewDisplayValue(
					PointerTo("Icon"),
					PointerTo("SVG icon for this item. Must have the declared size of 64x64, must not include "+
						"additional namespaces, and must not reference external resources."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				[]string{"\"<svg ...></svg>\""},
			),
		}).Any(),
		"FloatSchema": NewObjectType[*floatSchema]("FloatSchema", map[string]PropertyType{
			"min": NewPropertyType[float64](
				NewFloatType(nil, nil, nil),
				NewDisplayValue(
					PointerTo("Minimum"),
					PointerTo("Minimum value for this float (inclusive)."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				[]string{"5.0"},
			),
			"max": NewPropertyType[float64](
				NewFloatType(nil, nil, nil),
				NewDisplayValue(
					PointerTo("Maximum"),
					PointerTo("Maximum value for this float (inclusive)."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				[]string{"16.0"},
			),
			"units": unitsProperty,
		}).Any(),
		"IntEnumSchema": NewObjectType[*intEnumSchema]("IntEnumSchema", map[string]PropertyType{
			"values": NewPropertyType[map[int64]*displayValue](
				NewMapType[int64, *displayValue](
					NewIntType(nil, nil, nil),
					NewRefType[*displayValue](
						"DisplayValue",
						nil,
					),
					IntPointer(1),
					nil,
				),
				NewDisplayValue(
					PointerTo("Values"),
					PointerTo("Possible values for this field."),
					nil,
				),
				true,
				nil,
				nil,
				nil,
				nil,
				[]string{"{\"1024\": {\"name\": \"kB\"}, \"1048576\": {\"name\": \"MB\"}}"},
			),
			"units": unitsProperty,
		}).Any(),
		"IntSchema": NewObjectType[*intSchema](
			"IntSchema",
			map[string]PropertyType{
				"min": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, nil),
					NewDisplayValue(
						PointerTo("Minimum"),
						PointerTo("Minimum value for this int (inclusive)."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"5"},
				),
				"max": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, nil),
					NewDisplayValue(
						PointerTo("Maximum"),
						PointerTo("Maximum value for this int (inclusive)."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"16"},
				),
				"units": unitsProperty,
			},
		).Any(),
		"ListSchema": NewObjectType[*listSchema](
			"ListSchema",
			map[string]PropertyType{
				"items": NewPropertyType[any](
					valueType,
					NewDisplayValue(
						PointerTo("Items"),
						PointerTo("Type definition for items in this list."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"min": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, nil),
					NewDisplayValue(
						PointerTo("Minimum"),
						PointerTo("Minimum number of items in this list.."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"5"},
				),
				"max": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, nil),
					NewDisplayValue(
						PointerTo("Maximum"),
						PointerTo("Maximum value for this int (inclusive)."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"16"},
				),
			},
		).Any(),
		"MapSchema": NewObjectType[*mapSchema](
			"MapSchema",
			map[string]PropertyType{
				"keys": NewPropertyType[any](
					mapKeyType,
					NewDisplayValue(
						PointerTo("Keys"),
						PointerTo("Type definition for keys in this map."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"values": NewPropertyType[any](
					valueType,
					NewDisplayValue(
						PointerTo("Values"),
						PointerTo("Type definition for values in this map."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"min": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, nil),
					NewDisplayValue(
						PointerTo("Minimum"),
						PointerTo("Minimum number of items in this list.."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"5"},
				),
				"max": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, nil),
					NewDisplayValue(
						PointerTo("Maximum"),
						PointerTo("Maximum value for this int (inclusive)."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"16"},
				),
			},
		).Any(),
		"ObjectSchema": NewObjectType[*objectSchema](
			"ObjectSchema",
			map[string]PropertyType{
				"id": NewPropertyType[string](
					idType,
					NewDisplayValue(
						PointerTo("ID"),
						PointerTo("Unique identifier for this object within the current scope."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"properties": NewPropertyType[map[string]*propertySchema](
					NewMapType[string, *propertySchema](
						NewStringType(
							IntPointer(1),
							nil,
							nil,
						),
						NewRefType[*propertySchema](
							"PropertySchema",
							nil,
						),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Properties"),
						PointerTo("Properties of this object."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"OneOfIntSchema": NewObjectType[*oneOfSchema[int64, *refSchema]](
			"OneOfIntSchema",
			map[string]PropertyType{
				"discriminator_field_name": NewPropertyType[string](
					NewStringType(nil, nil, nil),
					NewDisplayValue(
						PointerTo("Discriminator field name"),
						PointerTo("Name of the field used to discriminate between possible values. If this "+
							"field is present on any of the component objects it must also be an int."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"\"_type\""},
				),
				"types": NewPropertyType[map[int64]*refType[any]](
					NewMapType[int64, *refType[any]](
						NewIntType(nil, nil, nil),
						NewRefType[*refType[any]]("RefSchema", nil),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Types"),
						nil,
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"OneOfStringSchema": NewObjectType[*oneOfSchema[string, *refSchema]](
			"OneOfStringSchema",
			map[string]PropertyType{
				"discriminator_field_name": NewPropertyType[string](
					NewStringType(nil, nil, nil),
					NewDisplayValue(
						PointerTo("Discriminator field name"),
						PointerTo("Name of the field used to discriminate between possible values. If this "+
							"field is present on any of the component objects it must also be an int."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"\"_type\""},
				),
				"types": NewPropertyType[map[string]*refSchema](
					NewMapType[string, *refSchema](
						NewStringType(nil, nil, nil),
						NewRefType[*refSchema]("RefSchema", nil),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Types"),
						nil,
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"PatternSchema": NewObjectType[*patternSchema](
			"PatternSchema",
			map[string]PropertyType{},
		).Any(),
		"PropertySchema": NewObjectType[*propertySchema](
			"PropertySchema",
			map[string]PropertyType{
				"type": NewPropertyType[any](
					valueType,
					NewDisplayValue(
						PointerTo("Type"),
						PointerTo(
							"Type definition for this field.",
						),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"display": displayProperty,
				"required": NewPropertyType[bool](
					NewBoolType(),
					NewDisplayValue(
						PointerTo("Required"),
						PointerTo(
							"When set to true, the value for this field must be provided under all circumstances.",
						),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					PointerTo("true"),
					nil,
				),
				"required_if_not": NewPropertyType[[]string](
					NewListType[string](
						NewStringType(nil, nil, nil),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Required if not"),
						PointerTo(
							"Sets the current property to be required if none of the properties in this list "+
								"are set.",
						),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"required_if": NewPropertyType[[]string](
					NewListType[string](
						NewStringType(nil, nil, nil),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Required if"),
						PointerTo(
							"Sets the current property to required if any of the properties in this list are set.",
						),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"conflicts": NewPropertyType[[]string](
					NewListType[string](
						NewStringType(nil, nil, nil),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Conflicts"),
						PointerTo("The current property cannot be set if any of the listed properties are set."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"default": NewPropertyType[string](
					NewStringType(nil, nil, nil),
					NewDisplayValue(
						PointerTo("Default"),
						PointerTo(
							"Default value for this property in JSON encoding. The value must be "+
								"unserializable by the type specified in the type field.",
						),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"examples": NewPropertyType[[]string](
					NewListType[string](
						NewStringType(nil, nil, nil),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Examples"),
						PointerTo("Example values for this property, encoded as JSON."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"RefSchema": NewObjectType[*refSchema](
			"RefSchema",
			map[string]PropertyType{
				"id": NewPropertyType[string](
					idType,
					NewDisplayValue(
						PointerTo("ID"),
						PointerTo("Referenced object ID."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"display": displayProperty,
			},
		).Any(),
		"Schema": NewObjectType[*schema](
			"Schema",
			map[string]PropertyType{
				"steps": NewPropertyType[map[string]*stepSchema](
					NewMapType[string, *stepSchema](
						idType,
						NewRefType[*stepSchema](
							"StepSchema",
							nil,
						),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Steps"),
						PointerTo("Steps this schema supports."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"ScopeSchema": NewObjectType[*scopeSchema](
			"ScopeSchema",
			map[string]PropertyType{
				"objects": NewPropertyType[map[string]*objectSchema](
					NewMapType[string, *objectSchema](
						idType,
						NewRefType[*objectSchema](
							"ObjectSchema",
							nil,
						),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Objects"),
						PointerTo("A set of referencable objects. These objects may contain references themselves."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"root": NewPropertyType[string](
					idType,
					NewDisplayValue(
						PointerTo("Root object"),
						PointerTo("ID of the root object of the scope."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"StepOutputSchema": NewObjectType[*stepOutputSchema](
			"StepOutputSchema",
			map[string]PropertyType{
				"display": displayProperty,
				"error": NewPropertyType[bool](
					NewBoolType(),
					NewDisplayValue(
						PointerTo("Error"),
						PointerTo("If set to true, this output will be treated as an error output."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					PointerTo("false"),
					nil,
				),
				"schema": NewPropertyType[*scopeSchema](
					NewRefType[*scopeSchema](
						"ScopeSchema",
						nil,
					),
					NewDisplayValue(
						PointerTo("Schema"),
						PointerTo("Data schema for this particular output."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"StepSchema": NewObjectType[*stepSchema](
			"StepSchema",
			map[string]PropertyType{
				"display": displayProperty,
				"id": NewPropertyType[string](
					idType,
					NewDisplayValue(
						PointerTo("ID"),
						PointerTo("Machine identifier for this step."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"input": NewPropertyType[*scopeSchema](
					NewRefType[*scopeSchema](
						"ScopeSchema",
						nil,
					),
					NewDisplayValue(
						PointerTo("Input"),
						PointerTo("Input data schema."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"outputs": NewPropertyType[map[string]*stepOutputSchema](
					NewMapType[string, *stepOutputSchema](
						idType,
						NewRefType[*stepOutputSchema](
							"StepOutputSchema",
							nil,
						),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Input"),
						PointerTo("Input data schema."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		).Any(),
		"StringEnumSchema": NewObjectType[*stringEnumSchema](
			"StringEnumSchema",
			map[string]PropertyType{
				"values": NewPropertyType[map[int64]*displayValue](
					NewMapType[int64, *displayValue](
						NewIntType(nil, nil, nil),
						NewRefType[*displayValue](
							"DisplayValue",
							nil,
						),
						IntPointer(1),
						nil,
					),
					NewDisplayValue(
						PointerTo("Values"),
						PointerTo("Mapping where the left side of the map holds the possible value and the "+
							"right side holds the display value for forms, etc."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					[]string{"{\n" +
						"  \"apple\": {\n" +
						"    \"name\": \"Apple\"\n" +
						"  },\n" +
						"  \"orange\": {\n" +
						"    \"name\": \"Orange\"\n" +
						"  }\n" +
						"}"},
				),
			},
		).Any(),
		"StringSchema": NewObjectType[*stringSchema](
			"StringSchema",
			map[string]PropertyType{
				"min": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, PointerTo(UnitCharacters)),
					NewDisplayValue(
						PointerTo("Minimum"),
						PointerTo("Minimum length for this string (inclusive)."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"5"},
				),
				"max": NewPropertyType[int64](
					NewIntType(IntPointer(0), nil, PointerTo(UnitCharacters)),
					NewDisplayValue(
						PointerTo("Maximum"),
						PointerTo("Maximum length for this string (inclusive)."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"16"},
				),
				"pattern": NewPropertyType[*regexp.Regexp](
					NewPatternType(),
					NewDisplayValue(
						PointerTo("Pattern"),
						PointerTo("Regular expression this string must match."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{"\"^[a-zA-Z]+$\""},
				),
			},
		).Any(),
		"Unit": NewObjectType[*unit](
			"Unit",
			map[string]PropertyType{
				"name_long_plural": NewPropertyType[string](
					NewStringType(nil, nil, nil),
					NewDisplayValue(
						PointerTo("Name long (plural)"),
						PointerTo("Longer name for this unit in plural form."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					[]string{"\"bytes\",\"characters\""},
				),
				"name_long_singular": NewPropertyType[string](
					NewStringType(nil, nil, nil),
					NewDisplayValue(
						PointerTo("Name long (singular)"),
						PointerTo("Longer name for this unit in singular form."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					[]string{"\"byte\",\"character\""},
				),
				"name_short_plural": NewPropertyType[string](
					NewStringType(nil, nil, nil),
					NewDisplayValue(
						PointerTo("Name short (plural)"),
						PointerTo("Shorter name for this unit in plural form."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					[]string{"\"B\",\"chars\""},
				),
				"name_short_singular": NewPropertyType[string](
					NewStringType(nil, nil, nil),
					NewDisplayValue(
						PointerTo("Name short (singular)"),
						PointerTo("Shorter name for this unit in singular form."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					[]string{"\"B\",\"char\""},
				),
			},
		).Any(),
		"Units": NewObjectType[*units](
			"Units",
			map[string]PropertyType{
				"base_unit": NewPropertyType[*unit](
					NewRefType[*unit](
						"Unit",
						nil,
					),
					NewDisplayValue(
						PointerTo("Base unit"),
						PointerTo("The base unit is the smallest unit of scale for this set of units."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					[]string{
						"{\n" +
							"  \"name_short_singular\": \"B\",\n" +
							"  \"name_short_plural\": \"B\",\n" +
							"  \"name_long_singular\": \"byte\",\n" +
							"  \"name_long_plural\": \"bytes\"\n" +
							"}",
					},
				),
				"multipliers": NewPropertyType[map[int64]*unit](
					NewMapType[int64, *unit](
						NewIntType(nil, nil, nil),
						NewRefType[*unit]("Unit", nil),
						nil,
						nil,
					),
					NewDisplayValue(
						PointerTo("Base unit"),
						PointerTo("The base unit is the smallest unit of scale for this set of units."),
						nil,
					),
					false,
					nil,
					nil,
					nil,
					nil,
					[]string{
						"{\n" +
							"  \"1024\": {\n" +
							"    \"name_short_singular\": \"kB\",\n" +
							"    \"name_short_plural\": \"kB\",\n" +
							"    \"name_long_singular\": \"kilobyte\",\n" +
							"    \"name_long_plural\": \"kilobytes\"\n" +
							"  },\n" +
							"  \"1048576\": {\n" +
							"    \"name_short_singular\": \"MB\",\n" +
							"    \"name_short_plural\": \"MB\",\n" +
							"    \"name_long_singular\": \"megabyte\",\n" +
							"    \"name_long_plural\": \"megabytes\"\n" +
							"  }\n" +
							"}",
					},
				),
			},
		).Any(),
	},
	"Schema",
)
