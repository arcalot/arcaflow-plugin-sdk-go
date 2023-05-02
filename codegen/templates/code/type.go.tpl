{{- if eq .TypeID "enum_string" -}}
    {{- partial "code/enum_string" . -}}
{{- else if eq .TypeID "enum_integer" -}}
    {{- partial "code/enum_integer" . -}}
{{- else if eq .TypeID "string" -}}
    {{- partial "code/string" . -}}
{{- else if eq .TypeID "pattern" -}}
    {{- partial "code/pattern" . -}}
{{- else if eq .TypeID "integer" -}}
    {{- partial "code/integer" . -}}
{{- else if eq .TypeID "float" -}}
    {{- partial "code/float" . -}}
{{- else if eq .TypeID "bool" -}}
    {{- partial "code/bool" . -}}
{{- else if eq .TypeID "list" -}}
    {{- partial "code/list" . -}}
{{- else if eq .TypeID "map" -}}
    {{- partial "code/map" . -}}
{{- else if eq .TypeID "scope" -}}
    {{- partial "code/scope" .  -}}
{{- else if eq .TypeID "object" -}}
    {{- partial "code/object" .  -}}
{{- else if eq .TypeID "one_of_string" -}}
    {{- partial "code/one_of_string" .  -}}
{{- else if eq .TypeID "one_of_int" -}}
    {{- partial "code/one_of_int" .  -}}
{{- else if eq .TypeID "ref" -}}
    {{- partial "code/ref" .  -}}
{{- end -}}