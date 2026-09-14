{{- /*
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
*/ -}}
{{- define "crdUpgrader.image" -}}
    {{- $globalRegistry := .Values.global.image.registry -}}
    {{- $crdUpgraderImage := .Values.crdUpgrader.image -}}
    {{- with .Values.crdUpgrader.image -}}
        {{- if $globalRegistry -}}
            {{- printf "%s/%s:%s" $globalRegistry $crdUpgraderImage.name .tag -}}
        {{- else -}}
          {{- printf "%s/%s:%s" .registry .name .tag -}}
        {{- end -}}
    {{- end -}}
{{- end -}}