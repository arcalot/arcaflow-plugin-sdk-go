{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.PropertySchema*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewPropertySchema(
    {{ prefix (partial "schema/type" .Type) "    " }},
    {{ prefix (partial "schema/display" .DisplayValue) "    " }},
    {{ if .Required }}true{{ else }}false{{end}},
    {{ .RequiredIfValue | stringList }},
    {{ .RequiredIfNotValue | stringList }},
    {{ .ConflictsValue | stringList }},
    {{ if .DefaultValue }}{{ .DefaultValue | escapeStringPtr }}{{ else }}nil{{ end }},
    {{ .ExamplesValue | stringList }},
)