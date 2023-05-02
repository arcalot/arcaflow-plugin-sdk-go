{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.IntSchema*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewIntSchema(
    {{with .MinValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{with .MaxValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{ prefix (partial "schema/units" .Units) "    " }},
)