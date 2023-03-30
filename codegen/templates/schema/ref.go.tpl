{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.RefSchema*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewRefSchema(
    {{ .IDValue | escapeString }},
    {{ prefix (partial "schema/display" .DisplayValue) "    " }},
)