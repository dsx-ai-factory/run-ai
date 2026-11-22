{{- /*
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
*/ -}}
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

{{- /*
The hooks read credentials from files projected out of controlPlane.existingSecret
when the matching controlPlane.secretKeys entry is set, so they can be provisioned
at runtime (e.g. by a Job) instead of being passed as Helm values.
*/ -}}
{{- define "controlPlane.secret.clientSecretEnabled" -}}
{{- if or (not .Values.controlPlane.existingSecret) .Values.controlPlane.secretKeys.clientSecret -}}true{{- end -}}
{{- end -}}

{{- define "controlPlane.secret.bootstrapTokenFromSecret" -}}
{{- if and .Values.controlPlane.existingSecret .Values.controlPlane.secretKeys.bootstrapToken -}}true{{- end -}}
{{- end -}}

{{- define "controlPlane.secret.clusterUidFromSecret" -}}
{{- if and .Values.controlPlane.existingSecret .Values.controlPlane.secretKeys.clusterUid -}}true{{- end -}}
{{- end -}}

{{- define "controlPlane.credentialArgs" -}}
{{- if include "controlPlane.secret.clientSecretEnabled" . }}
- --control-plane-client-secret=@/runai/config/client_secret
{{- end }}
{{- if include "controlPlane.secret.bootstrapTokenFromSecret" . }}
- --control-plane-bootstrap-token=@/runai/config/bootstrap_token
{{- else }}
- --control-plane-bootstrap-token={{ .Values.controlPlane.bootstrapToken }}
{{- end }}
{{- end -}}

{{- define "cluster.uidArg" -}}
{{- if include "controlPlane.secret.clusterUidFromSecret" . -}}
- --cluster-uid=@/runai/config/cluster_uid
{{- else -}}
- --cluster-uid={{ include "cluster.uid" . }}
{{- end -}}
{{- end -}}

{{- define "controlPlane.secret.volume" -}}
- secret:
    name: {{ include "controlPlane.secret.name" . }}
    items:
{{- if include "controlPlane.secret.clientSecretEnabled" . }}
      - key: {{ include "controlPlane.secret.clientSecretKey" . }}
        path: client_secret
{{- end }}
{{- if include "controlPlane.secret.bootstrapTokenFromSecret" . }}
      - key: {{ .Values.controlPlane.secretKeys.bootstrapToken }}
        path: bootstrap_token
{{- end }}
{{- if include "controlPlane.secret.clusterUidFromSecret" . }}
      - key: {{ .Values.controlPlane.secretKeys.clusterUid }}
        path: cluster_uid
{{- end }}
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

