{{- define "controlPlane.url" -}}
{{- if .Values.controlPlane -}}
{{- if .Values.controlPlane.url -}}
{{- if (or (hasPrefix "https://" .Values.controlPlane.url) (hasPrefix "http://" .Values.controlPlane.url)) -}}
{{- printf "%s" .Values.controlPlane.url -}}
{{- else -}}
{{- printf "https://%s" .Values.controlPlane.url -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "controlPlane.clientID" -}}
{{- if .Values.controlPlane -}}
{{- if .Values.controlPlane.clientID -}}
{{- printf "%s" .Values.controlPlane.clientID -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "cluster.uid" -}}
{{- if .Values.cluster -}}
{{- if .Values.cluster.uid -}}
{{- printf "%s" .Values.cluster.uid -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "cluster.url" -}}
{{- if .Values.cluster -}}
{{- if .Values.cluster.url -}}
{{- if (or (hasPrefix "https://" .Values.cluster.url) (hasPrefix "http://" .Values.cluster.url)) -}}
{{- printf "%s" .Values.cluster.url -}}
{{- else -}}
{{- printf "https://%s" .Values.cluster.url -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "controlPlane.secret.name" -}}
{{ .Values.controlPlane.existingSecret | default "runai-hooks-secret" }}
{{- end -}}

{{- define "controlPlane.secret.clientSecretKey" -}}
{{- if .Values.controlPlane.existingSecret -}}
{{ .Values.controlPlane.secretKeys.clientSecret }}
{{- else -}}
client-secret
{{- end -}}
{{- end -}}

{{- define "controlPlane.secret.volume" -}}
- secret:
    name: {{ include "controlPlane.secret.name" . }}
    items:
      - key: {{ include "controlPlane.secret.clientSecretKey" . }}
        path: client_secret
{{- end -}}

{{ define "jobLabels" -}}
    {{- if .Values.global.additionalJobLabels -}}
        {{- tpl (toYaml .Values.global.additionalJobLabels) . -}}
    {{- end -}}
{{- end -}}

{{ define "jobAnnotations" -}}
    {{- if .Values.global.additionalJobAnnotations -}}
        {{- tpl (toYaml .Values.global.additionalJobAnnotations) . -}}
    {{- end -}}
{{- end -}}

{{- define "nmcServiceAccountArgs" -}}
{{- $nmcServiceAccounts := .Values.global.serviceAccountName -}}
{{- if .Values.global.nmc.allowedServiceAccounts -}}
{{- $nmcServiceAccounts = printf "%s,%s" $nmcServiceAccounts (join "," .Values.global.nmc.allowedServiceAccounts) -}}
{{- end -}}
{{- $nmcServiceAccounts -}}
{{- end -}}

