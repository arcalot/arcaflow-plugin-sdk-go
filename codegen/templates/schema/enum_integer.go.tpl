{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.EnumSchema[int64]*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewIntEnumSchema(
    map[int64]*schema.DisplayValue{
        {{- range $value, $display := .ValidValuesMap }}
            {{ $value }}: {{ prefix (partial "schema/display" $display) "        " }},
        {{- end }}
    },
    {{ with .IntUnits }}{{ prefix (partial "schema/units" .IntUnits) "    " }}{{ else }}nil{{ end }},
)