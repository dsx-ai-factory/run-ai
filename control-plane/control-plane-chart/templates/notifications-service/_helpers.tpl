{{- /*
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
*/ -}}
{{- define "control-plane.notificationsService.fullname" -}}
{{- printf "%s-%s" (include "common.fullname" .) .Values.notificationsService.name }}
{{- end }}

{{- define "control-plane.notificationsService.extractedSinkSecrets" -}}
  {{- $extracted := dict -}}
  {{- $config := fromYaml (tpl (toYaml (.Values.notificationsService.config | default dict)) .) -}}
  {{- if and $config (hasKey $config "sinks") -}}
    {{- $sensitiveFields := list "password" "api_token" "user" -}}
    {{- range $sinkName, $sinkConfig := $config.sinks -}}
      {{- if not $sinkConfig.existingSecret -}}
        {{- $sinkSecrets := dict -}}
        {{- range $fieldName, $fieldValue := $sinkConfig -}}
          {{- if and (has $fieldName $sensitiveFields) (kindIs "string" $fieldValue) (not (hasPrefix "$" $fieldValue)) (not (hasPrefix "${" $fieldValue)) -}}
            {{- $_ := set $sinkSecrets $fieldName $fieldValue -}}
          {{- end -}}
        {{- end -}}
        {{- if gt (len $sinkSecrets) 0 -}}
          {{- $_ := set $extracted $sinkName $sinkSecrets -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
  {{- $extracted | toJson -}}
{{- end -}}

{{- define "control-plane.notificationsService.processedConfig" -}}
  {{- $config := fromYaml (tpl (toYaml (.Values.notificationsService.config | default dict)) .) -}}
  {{- if $config -}}
    {{- if hasKey $config "sinks" -}}
      {{- $extractedSecrets := include "control-plane.notificationsService.extractedSinkSecrets" . | fromJson -}}
      {{- range $sinkName, $sinkConfig := $config.sinks -}}
        {{- if and $sinkConfig.existingSecret (not $sinkConfig.secretKeys) -}}
          {{- fail (printf "sink %s: existingSecret requires secretKeys to be set" $sinkName) -}}
        {{- end -}}
        {{- if and $sinkConfig.existingSecret $sinkConfig.secretKeys -}}
          {{- range $fieldName, $secretKey := $sinkConfig.secretKeys -}}
            {{- $filePath := printf "$__file{/etc/notifications-service/secret/sinks/%s/%s}" $sinkName $fieldName -}}
            {{- $_ := set $sinkConfig $fieldName $filePath -}}
          {{- end -}}
          {{- $_ := unset $sinkConfig "existingSecret" -}}
          {{- $_ := unset $sinkConfig "secretKeys" -}}
        {{- else if hasKey $extractedSecrets $sinkName -}}
          {{- $sinkSecrets := index $extractedSecrets $sinkName -}}
          {{- range $fieldName, $fieldValue := $sinkSecrets -}}
            {{- $filePath := printf "$__file{/etc/notifications-service/secret/sinks/%s/%s}" $sinkName $fieldName -}}
            {{- $_ := set $sinkConfig $fieldName $filePath -}}
          {{- end -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
    {{- toYaml $config -}}
  {{- else -}}
    {}
  {{- end -}}
{{- end -}}
