{{/*
Merge one or more YAML templates and output the result.
This takes an list of values:
- the top context
- [optional] zero or more template args
- [optional] the template name of the overrides (destination)
- the template name of the base (source)
*/}}
{{- define "common.utils.merge" -}}
{{- $top := first . }}
{{- $tplName := last . }}
{{- $args := initial . }}
{{- if typeIs "string" (last $args) }}
  {{- $overridesName := last $args }}
  {{- $args = initial $args }}
  {{- $tpl := fromYaml (include $tplName $args) | default (dict) }}
  {{- $overrides := fromYaml (include $overridesName $args) | default (dict) }}
  {{- toYaml (merge $overrides $tpl) }}
{{- else }}
  {{- include $tplName $args }}
{{- end }}
{{- end }}
