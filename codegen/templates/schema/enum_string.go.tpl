{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.EnumSchema[string]*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewStringEnumSchema(
    map[string]*schema.DisplayValue{
        {{- range $value, $display := .ValidValuesMap }}
            {{ $value | escapeString }}: {{ prefix (partial "schema/display" $display) "        " }},
        {{- end }}
    },
)