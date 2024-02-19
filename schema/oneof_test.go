package schema_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/schema"
)

type oneOfTestObjectB struct {
	Message string `json:"message"`
}

func (o oneOfTestObjectB) String() string {
	return o.Message
}

type oneOfTestObjectC struct {
	M string `json:"m"`
}

type oneOfTestObjectA struct {
	S any `json:"s"`
}

type oneOfTestInlineObjectB struct {
	Message string `json:"message"`
	Choice  string `json:"choice"`
}

func (o oneOfTestInlineObjectB) String() string {
	return o.Message
}

type oneOfTestInlineObjectC struct {
	M      string `json:"m"`
	Choice string `json:"choice"`
}

func TestOneOfTypeID(t *testing.T) {
	assert.Equals(
		t,
		oneOfStringTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assert.Equals(
		t,
		oneOfStringTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfString,
	)
	assert.Equals(
		t,
		oneOfIntTestObjectASchema.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
	assert.Equals(
		t,
		oneOfIntTestObjectAType.
			Objects()["A"].
			Properties()["s"].
			Type().
			TypeID(),
		schema.TypeIDOneOfInt,
	)
}

var oneOfTestObjectBProperties = map[string]*schema.PropertySchema{
	"message": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestObjectCProperties = map[string]*schema.PropertySchema{
	"m": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestObjectDProperties = map[string]*schema.PropertySchema{
	"K": schema.NewPropertySchema(
		schema.NewIntSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestBSchema = schema.NewObjectSchema(
	"B",
	oneOfTestObjectBProperties,
)

var oneOfTestCSchema = schema.NewObjectSchema(
	"C",
	oneOfTestObjectCProperties,
)

var oneOfTestDSchema = schema.NewObjectSchema(
	"D",
	oneOfTestObjectDProperties,
)

var oneOfTestBMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestObjectB](
	"B",
	oneOfTestObjectBProperties,
)

var oneOfTestCMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestObjectC](
	"C",
	oneOfTestObjectCProperties,
)

var oneOfTestInlineObjectBProperties = map[string]*schema.PropertySchema{
	"message": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"choice": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestInlineObjectCProperties = map[string]*schema.PropertySchema{
	"m": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
	"choice": schema.NewPropertySchema(
		schema.NewStringSchema(nil, nil, nil),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	),
}

var oneOfTestInlineBSchema = schema.NewObjectSchema(
	"B",
	oneOfTestInlineObjectBProperties,
)

var oneOfTestInlineCSchema = schema.NewObjectSchema(
	"C",
	oneOfTestInlineObjectCProperties,
)

var oneOfTestInlineBMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestInlineObjectB](
	"B",
	oneOfTestInlineObjectBProperties,
)

var oneOfTestInlineCMappedSchema = schema.NewStructMappedObjectSchema[oneOfTestInlineObjectC](
	"C",
	oneOfTestInlineObjectCProperties,
)

func Test_OneOfString_ConstructorBypass(t *testing.T) {
	data := `{
    "objects": {
      "FullName": {
        "id": "FullName",
        "properties": {
          "first_name": {
            "required": true,
            "type": {
              "type_id": "string"
            }
          },
          "last_name": {
            "required": true,
            "type": {
              "type_id": "string"
            }
          }
        }
      },
      "Nickname": {
        "id": "Nickname",
        "properties": {
          "nick": {
            "required": true,
            "type": {
              "type_id": "string"
            }
          }
        }
      },
      "InputParams": {
        "id": "InputParams",
        "properties": {
          "name": {
            "required": true,
            "type": {
              "discriminator_field_name": "_type",
              "type_id": "one_of_string",
              "types": {
                "fullname": {
                  "id": "FullName",
                  "type_id": "ref"
                },
                "nickname": {
                  "id": "Nickname",
                  "type_id": "ref"
                }
              }
            }
          }
        }
      }
    },
    "root": "InputParams"
}`
	var input any
	assert.NoError(t, json.Unmarshal([]byte(data), &input))
	fmt.Printf("%v\n", input)
	myScopeSchema := schema.DescribeScope()
	scopeAny, err := myScopeSchema.Unserialize(input)
	assert.NoError(t, err)
	scopeSchemaTyped := scopeAny.(*schema.ScopeSchema)
	scopeSchemaTyped.ApplyScope(scopeSchemaTyped)
	fmt.Printf("%v\n", scopeSchemaTyped)

	//var input_nick any = map[string]any{
	//	"name": map[string]any{
	//		"_type": "nickname",
	//		"nick":  "ArcaLot",
	//	},
	//}

	var input_full any = map[string]any{
		"name": map[string]any{
			"_type":      "fullname",
			"first_name": "Arca",
			"last_name":  "Lot",
		},
	}

	unserializedData, err := scopeSchemaTyped.Unserialize(input_full)
	assert.NoError(t, err)
	fmt.Printf("%v\n", unserializedData)

}
