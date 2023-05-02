{{- if eq .TypeID "enum_string" -}}
    {{- partial "schema/enum_string" . -}}
{{- else if eq .TypeID "enum_integer" -}}
    {{- partial "schema/enum_integer" . -}}
{{- else if eq .TypeID "string" -}}
    {{- partial "schema/string" . -}}
{{- else if eq .TypeID "pattern" -}}
    {{- partial "schema/pattern" . -}}
{{- else if eq .TypeID "integer" -}}
    {{- partial "schema/integer" . -}}
{{- else if eq .TypeID "float" -}}
    {{- partial "schema/float" . -}}
{{- else if eq .TypeID "bool" -}}
    {{- partial "schema/bool" . -}}
{{- else if eq .TypeID "list" -}}
    {{- partial "schema/list" . -}}
{{- else if eq .TypeID "map" -}}
    {{- partial "schema/map" . -}}
{{- else if eq .TypeID "scope" -}}
    {{- partial "schema/scope" .  -}}
{{- else if eq .TypeID "object" -}}
    {{- partial "schema/object" .  -}}
{{- else if eq .TypeID "one_of_string" -}}
    {{- partial "schema/one_of_string" .  -}}
{{- else if eq .TypeID "one_of_int" -}}
    {{- partial "schema/one_of_int" .  -}}
{{- else if eq .TypeID "ref" -}}
    {{- partial "schema/ref" .  -}}
{{- end -}}