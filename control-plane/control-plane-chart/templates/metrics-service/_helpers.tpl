{{- /*
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
*/ -}}
# Datasource Provider Auth Secret (grafana labs access token)
{{- define "control-plane.metricsService.config.datasourceProvider.auth.secret.name" -}}
{{ .Values.metricsService.config.datasourceProvider.auth.existingSecret | default (printf "%s-metrics-service" .Values.fullnameOverride) }}
{{- end -}}

{{- define "control-plane.metricsService.config.datasourceProvider.auth.secret.accessTokenKey" -}}
{{ .Values.metricsService.config.datasourceProvider.auth.secretKeys.accessTokenKey }}
{{- end -}}

# Static Datasource Secret (read/write basic-auth credentials)
{{- define "control-plane.metricsService.config.datasourceProvider.staticDatasource.secret.name" -}}
{{ .Values.metricsService.config.datasourceProvider.staticDatasource.existingSecret | default (printf "%s-metrics-service" .Values.fullnameOverride) }}
{{- end -}}

{{- define "control-plane.metricsService.config.datasourceProvider.staticDatasource.secret.readUserKey" -}}
{{ .Values.metricsService.config.datasourceProvider.staticDatasource.secretKeys.readUserKey }}
{{- end -}}

{{- define "control-plane.metricsService.config.datasourceProvider.staticDatasource.secret.readPasswordKey" -}}
{{ .Values.metricsService.config.datasourceProvider.staticDatasource.secretKeys.readPasswordKey }}
{{- end -}}

{{- define "control-plane.metricsService.config.datasourceProvider.staticDatasource.secret.writeUserKey" -}}
{{ .Values.metricsService.config.datasourceProvider.staticDatasource.secretKeys.writeUserKey }}
{{- end -}}

{{- define "control-plane.metricsService.config.datasourceProvider.staticDatasource.secret.writePasswordKey" -}}
{{ .Values.metricsService.config.datasourceProvider.staticDatasource.secretKeys.writePasswordKey }}
{{- end -}}
