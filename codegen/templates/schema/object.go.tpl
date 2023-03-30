{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.ObjectSchema*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewStructMappedObjectSchema[{{ .ID | structName }}](
    {{ .ID | escapeString }},
    map[string]*schema.PropertySchema{{"{"}}{{- range $propertyID, $property := .Properties }}
        {{ $propertyID | escapeString}}: {{ prefix (partial "schema/property" $property) "        "}},
    {{ end -}}
    },
)