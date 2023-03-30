{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.ListSchema[any]*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewListSchema(
    {{ prefix (partial "schema/type" .Items) "    "}},
    {{with .MinValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{with .MaxValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
)