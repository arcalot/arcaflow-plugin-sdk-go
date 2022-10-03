package schema

import "regexp"

var unitsProperty = NewPropertySchema(
	NewRefSchema("Units", nil),
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
var idType = NewStringSchema(
	IntPointer(1),
	IntPointer(255),
	regexp.MustCompile("^[$@a-zA-Z0-9-_]+$"),
)
var mapKeyType = NewOneOfStringSchema[any](
	map[string]Object{
		"integer": NewRefSchema(
			"Int",
			NewDisplayValue(
				PointerTo("Integer"),
				nil,
				nil,
			),
		),
		"string": NewRefSchema(
			"String",
			NewDisplayValue(
				PointerTo("String"),
				nil,
				nil,
			),
		),
	},
	"type_id",
)
var displayType = NewDisplayValue(
	PointerTo("Display"),
	PointerTo(
		"Name, description and icon.",
	),
	nil,
)
var displayProperty = NewPropertySchema(
	NewRefSchema(
		"Display",
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
var valueType = NewOneOfStringSchema[any](
	map[string]Object{
		"any": NewRefSchema(
			"AnySchema",
			NewDisplayValue(
				PointerTo("Any"),
				nil,
				nil,
			),
		),
		"bool": NewRefSchema(
			"BoolSchema",
			NewDisplayValue(
				PointerTo("Bool"),
				nil,
				nil,
			),
		),
		"enum_integer": NewRefSchema(
			"IntEnum",
			NewDisplayValue(
				PointerTo("Integer enum"),
				nil,
				nil,
			),
		),
		"enum_string": NewRefSchema(
			"StringEnum",
			NewDisplayValue(
				PointerTo("String enum"),
				nil,
				nil,
			),
		),
		"float": NewRefSchema(
			"Float",
			NewDisplayValue(
				PointerTo("Float"),
				nil,
				nil,
			),
		),
		"integer": NewRefSchema(
			"Int",
			NewDisplayValue(
				PointerTo("Integer"),
				nil,
				nil,
			),
		),
		"list": NewRefSchema(
			"List",
			NewDisplayValue(
				PointerTo("List"),
				nil,
				nil,
			),
		),
		"map": NewRefSchema(
			"Map",
			NewDisplayValue(
				PointerTo("Map"),
				nil,
				nil,
			),
		),
		"object": NewRefSchema(
			"Object",
			NewDisplayValue(
				PointerTo("Object"),
				nil,
				nil,
			),
		),
		"one_of_int": NewRefSchema(
			"OneOfIntSchema",
			NewDisplayValue(
				PointerTo("Multiple with int key"),
				nil,
				nil,
			),
		),
		"one_of_string": NewRefSchema(
			"OneOfStringSchema",
			NewDisplayValue(
				PointerTo("Multiple with string key"),
				nil,
				nil,
			),
		),
		"pattern": NewRefSchema(
			"Pattern",
			NewDisplayValue(
				PointerTo("Pattern"),
				nil,
				nil,
			),
		),
		"ref": NewRefSchema(
			"Ref",
			NewDisplayValue(
				PointerTo("Object reference"),
				nil,
				nil,
			),
		),
		"scope": NewRefSchema(
			"Scope",
			NewDisplayValue(
				PointerTo("Scope"),
				nil,
				nil,
			),
		),
		"string": NewRefSchema(
			"String",
			NewDisplayValue(
				PointerTo("String"),
				nil,
				nil,
			),
		),
	},
	"type_id",
)

var schemaSchema = NewScopeSchema(
	NewStructMappedObjectSchema[*SchemaSchema](
		"Schema",
		map[string]*PropertySchema{
			"steps": NewPropertySchema(
				NewMapSchema(
					idType,
					NewRefSchema(
						"Step",
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
	),
	NewStructMappedObjectSchema[*BoolSchema]("BoolSchema", map[string]*PropertySchema{}),
	NewStructMappedObjectSchema[*AnySchema]("AnySchema", map[string]*PropertySchema{}),
	NewStructMappedObjectSchema[*DisplayValue]("Display", map[string]*PropertySchema{
		"name": NewPropertySchema(
			NewStringSchema(IntPointer(1), nil, nil),
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
		"description": NewPropertySchema(
			NewStringSchema(IntPointer(1), nil, nil),
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
		"icon": NewPropertySchema(
			NewStringSchema(IntPointer(1), nil, nil),
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
	}),
	NewStructMappedObjectSchema[*FloatSchema]("Float", map[string]*PropertySchema{
		"min": NewPropertySchema(
			NewFloatSchema(nil, nil, nil),
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
		"max": NewPropertySchema(
			NewFloatSchema(nil, nil, nil),
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
	}),
	NewStructMappedObjectSchema[*IntEnumSchema]("IntEnum", map[string]*PropertySchema{
		"values": NewPropertySchema(
			NewMapSchema(
				NewIntSchema(nil, nil, nil),
				NewRefSchema(
					"Display",
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
	}),
	NewStructMappedObjectSchema[*IntSchema](
		"Int",
		map[string]*PropertySchema{
			"min": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, nil),
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
			"max": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, nil),
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
	),
	NewStructMappedObjectSchema[*ListSchema](
		"List",
		map[string]*PropertySchema{
			"items": NewPropertySchema(
				valueType,
				NewDisplayValue(
					PointerTo("Items"),
					PointerTo("ReflectedType definition for items in this list."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
			"min": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, nil),
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
			"max": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, nil),
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
	),
	NewStructMappedObjectSchema[*MapSchema[Type, Type]](
		"Map",
		map[string]*PropertySchema{
			"keys": NewPropertySchema(
				mapKeyType,
				NewDisplayValue(
					PointerTo("Keys"),
					PointerTo("ReflectedType definition for keys in this map."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
			"values": NewPropertySchema(
				valueType,
				NewDisplayValue(
					PointerTo("Values"),
					PointerTo("ReflectedType definition for values in this map."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
			"min": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, nil),
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
			"max": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, nil),
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
	),
	NewStructMappedObjectSchema[*ObjectSchema](
		"Object",
		map[string]*PropertySchema{
			"id": NewPropertySchema(
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
			"properties": NewPropertySchema(
				NewMapSchema(
					NewStringSchema(
						IntPointer(1),
						nil,
						nil,
					),
					NewRefSchema(
						"Property",
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
	),
	NewStructMappedObjectSchema[*OneOfSchema[int64, Object]](
		"OneOfIntSchema",
		map[string]*PropertySchema{
			"discriminator_field_name": NewPropertySchema(
				NewStringSchema(nil, nil, nil),
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
			"types": NewPropertySchema(
				NewMapSchema(
					NewIntSchema(nil, nil, nil),
					NewOneOfStringSchema[Object](
						map[string]Object{
							string(TypeIDRef):    NewRefSchema("Ref", nil),
							string(TypeIDScope):  NewRefSchema("Scope", nil),
							string(TypeIDObject): NewRefSchema("Object", nil),
						},
						"type_id",
					),
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
	),
	NewStructMappedObjectSchema[*OneOfSchema[string, Object]](
		"OneOfStringSchema",
		map[string]*PropertySchema{
			"discriminator_field_name": NewPropertySchema(
				NewStringSchema(nil, nil, nil),
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
			"types": NewPropertySchema(
				NewMapSchema(
					NewStringSchema(nil, nil, nil),
					NewOneOfStringSchema[Object](
						map[string]Object{
							string(TypeIDRef):    NewRefSchema("Ref", nil),
							string(TypeIDScope):  NewRefSchema("Scope", nil),
							string(TypeIDObject): NewRefSchema("Object", nil),
						},
						"type_id",
					),
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
	),
	NewStructMappedObjectSchema[*PatternSchema](
		"Pattern",
		map[string]*PropertySchema{},
	),
	NewStructMappedObjectSchema[*PropertySchema](
		"Property",
		map[string]*PropertySchema{
			"type": NewPropertySchema(
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
			"required": NewPropertySchema(
				NewBoolSchema(),
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
			"required_if_not": NewPropertySchema(
				NewListSchema(
					NewStringSchema(nil, nil, nil),
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
			"required_if": NewPropertySchema(
				NewListSchema(
					NewStringSchema(nil, nil, nil),
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
			"conflicts": NewPropertySchema(
				NewListSchema(
					NewStringSchema(nil, nil, nil),
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
			"default": NewPropertySchema(
				NewStringSchema(nil, nil, nil),
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
			"examples": NewPropertySchema(
				NewListSchema(
					NewStringSchema(nil, nil, nil),
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
	),
	NewStructMappedObjectSchema[*RefSchema](
		"Ref",
		map[string]*PropertySchema{
			"id": NewPropertySchema(
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
	),
	NewStructMappedObjectSchema[*ScopeSchema](
		"Scope",
		map[string]*PropertySchema{
			"objects": NewPropertySchema(
				NewMapSchema(
					idType,
					NewRefSchema(
						"Object",
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
			"root": NewPropertySchema(
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
	),
	NewStructMappedObjectSchema[*StepOutputSchema](
		"StepOutput",
		map[string]*PropertySchema{
			"display": displayProperty,
			"error": NewPropertySchema(
				NewBoolSchema(),
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
			"schema": NewPropertySchema(
				NewRefSchema(
					"Scope",
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
	),
	NewStructMappedObjectSchema[*StepSchema](
		"Step",
		map[string]*PropertySchema{
			"display": displayProperty,
			"id": NewPropertySchema(
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
			"input": NewPropertySchema(
				NewRefSchema(
					"Scope",
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
			"outputs": NewPropertySchema(
				NewMapSchema(
					idType,
					NewRefSchema(
						"StepOutput",
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
	),
	NewStructMappedObjectSchema[*StringEnumSchema](
		"StringEnum",
		map[string]*PropertySchema{
			"values": NewPropertySchema(
				NewMapSchema(
					NewIntSchema(nil, nil, nil),
					NewRefSchema(
						"Display",
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
	),
	NewStructMappedObjectSchema[*StringSchema](
		"String",
		map[string]*PropertySchema{
			"min": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, PointerTo(UnitCharacters)),
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
			"max": NewPropertySchema(
				NewIntSchema(IntPointer(0), nil, PointerTo(UnitCharacters)),
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
			"pattern": NewPropertySchema(
				NewPatternSchema(),
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
	),
	NewStructMappedObjectSchema[*unit](
		"Unit",
		map[string]*PropertySchema{
			"name_long_plural": NewPropertySchema(
				NewStringSchema(nil, nil, nil),
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
			"name_long_singular": NewPropertySchema(
				NewStringSchema(nil, nil, nil),
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
			"name_short_plural": NewPropertySchema(
				NewStringSchema(nil, nil, nil),
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
			"name_short_singular": NewPropertySchema(
				NewStringSchema(nil, nil, nil),
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
	),
	NewStructMappedObjectSchema[*units](
		"Units",
		map[string]*PropertySchema{
			"base_unit": NewPropertySchema(
				NewRefSchema(
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
			"multipliers": NewPropertySchema(
				NewMapSchema(
					NewIntSchema(nil, nil, nil),
					NewRefSchema("Unit", nil),
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
	),
)

func UnserializeSchema(data any) (*SchemaSchema, error) {
	s, err := schemaSchema.Unserialize(data)
	if err != nil {
		return nil, err
	}
	return s.(*SchemaSchema), nil
}
