type {{ .ID | structName }} struct {
{{- range $propertyID, $property := .Properties -}}
{{- if and $property.Required (ne $property.Type.TypeID "pattern")}}
    {{ $propertyID | property }} {{ partial "code/type" $property.Type}} `json:"{{ $propertyID }}"`
{{ else }}
    {{ $propertyID | property }} *{{ partial "code/type" $property.Type}} `json:"{{ $propertyID }}"`
{{ end }}{{ end -}}
}
