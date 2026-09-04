{{- define "common.serviceAccount.metadata" -}}
{{- $top := first . }}
{{- $serviceAccount := index . 1 }}
name: {{ include "common.serviceAccountName" . }}
{{- with $serviceAccount.annotations }}
annotations:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{- define "common.serviceAccount.tpl" -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  {{- include "common.metadata" (append . "common.serviceAccount.metadata") | nindent 2 }}
{{- end }}

{{- define "common.serviceAccount" -}}
{{- $top := first . }}
{{- $serviceAccount := index . 1 }}
{{- if $serviceAccount.create }}
  {{- include "common.utils.merge" (append . "common.serviceAccount.tpl") }}
{{- end }}
{{- end }}
