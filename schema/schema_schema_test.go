package schema_test

import (
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
	"gopkg.in/yaml.v3"
)

var examplePluginSchema = `
steps:
  hello-world:
    display:
      description: Says hello :)
      name: Hello world!
    id: hello-world
    input:
      objects:
        FullName:
          id: FullName
          properties:
            first_name:
              display:
                name: First name
              examples:
              - '"Arca"'
              required: true
              type:
                min: 1
                pattern: ^[a-zA-Z]+$
                type_id: string
            last_name:
              display:
                name: Last name
              examples:
              - '"Lot"'
              required: true
              type:
                min: 1
                pattern: ^[a-zA-Z]+$
                type_id: string
        InputParams:
          id: InputParams
          properties:
            name:
              display:
                description: Who do we say hello to?
                name: Name
              examples:
              - '{"_type": "fullname", "first_name": "Arca", "last_name": "Lot"}'
              - '{"_type": "nickname", "nick": "Arcalot"}'
              required: true
              type:
                discriminator_field_name: _type
                type_id: one_of_string
                types:
                  fullname:
                    display:
                      name: Full name
                    id: FullName
                  nickname:
                    display:
                      name: Nick
                    id: Nickname
        Nickname:
          id: Nickname
          properties:
            nick:
              display:
                name: Nickname
              examples:
              - '"Arcalot"'
              required: true
              type:
                min: 1
                pattern: ^[a-zA-Z]+$
                type_id: string
      root: InputParams
    outputs:
      error:
        error: false
        schema:
          objects:
            ErrorOutput:
              id: ErrorOutput
              properties:
                error:
                  display: {}
                  required: true
                  type:
                    type_id: string
          root: ErrorOutput
      success:
        error: false
        schema:
          objects:
            SuccessOutput:
              id: SuccessOutput
              properties:
                message:
                  display: {}
                  required: true
                  type:
                    type_id: string
          root: SuccessOutput
`

func TestSchemaUnserialization(t *testing.T) {
	data := map[string]any{}
	assertNoError(t, yaml.Unmarshal([]byte(examplePluginSchema), &data))
	unserializedData, err := schema.SchemaSchema.Unserialize(data)
	assertNoError(t, err)
	steps := assertNotNil(t, unserializedData.Steps())
	helloWorldStep := assertNotNil(t, steps["hello-world"])
	display := assertNotNil(t, helloWorldStep.Display())
	name := assertNotNil(t, display.Name())
	assertEqual(t, *name, "Hello world!")
}
