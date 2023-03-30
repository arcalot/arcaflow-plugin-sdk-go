{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.StringSchema*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewStringSchema(
    {{with .MinValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{with .MaxValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{with .Pattern}}regexp.MustCompile({{ . | escapeString}}){{ else }}nil{{end}},
)