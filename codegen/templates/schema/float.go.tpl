{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.FloatSchema*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewFloatSchema(
    {{with .MinValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{with .MaxValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{ prefix (partial "schema/units" .Units) "    " }},
)