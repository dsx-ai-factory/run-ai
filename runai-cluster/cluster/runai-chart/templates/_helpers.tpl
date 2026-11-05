{{- /*
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
*/ -}}
{{- define "createRunaiConfig" -}}
    {{- if not (index .Values "runai-operator" "enabled") -}}
        {{- print false -}}
    {{- else if not (and .Values.controlPlane .Values.cluster) -}}
        {{- print true -}}
    {{- else if (and
            .Values.controlPlane.url
            (or .Values.cluster.uid (and .Values.controlPlane.existingSecret .Values.controlPlane.secretKeys.clusterUid)))
    -}}
        {{- print false -}}
    {{- else -}}
        {{- print true -}}
    {{- end -}}
{{- end -}}

{{- define "imageRegistry" -}}
    {{- $globalRegistry := .Values.global.image.registry -}}
    {{- with index .Values "runai-operator" "image" -}}
        {{- if $globalRegistry -}}
            {{- printf "%s" $globalRegistry -}}
        {{- else -}}
          {{- printf "%s" .registry -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- /*
hooks.render and hooks.phase split the chart into the release and its lifecycle
hooks so they can be deployed separately. See values.yaml.
*/ -}}
{{- define "runai.hooksRender" -}}
    {{- $render := .Values.hooks.render | default "all" -}}
    {{- if not (has $render (list "all" "none" "only")) -}}
        {{- fail (printf "hooks.render must be one of all, none or only, got %q" $render) -}}
    {{- end -}}
    {{- $render -}}
{{- end -}}

{{- define "runai.hooksPhase" -}}
    {{- $phase := .Values.hooks.phase | default "all" -}}
    {{- if not (has $phase (list "all" "pre" "post" "delete")) -}}
        {{- fail (printf "hooks.phase must be one of all, pre, post or delete, got %q" $phase) -}}
    {{- end -}}
    {{- $phase -}}
{{- end -}}

{{- /* "true" when the resources that are not hooks should be rendered. */ -}}
{{- define "runai.renderRelease" -}}
    {{- ternary "false" "true" (eq (include "runai.hooksRender" .) "only") -}}
{{- end -}}

{{- /*
"true" when a hook template that belongs to the given phases should be rendered.
Call with (list . (list "pre" "post")).
*/ -}}
{{- define "runai.renderHook" -}}
    {{- $ctx := index . 0 -}}
    {{- $phases := index . 1 -}}
    {{- $phase := include "runai.hooksPhase" $ctx -}}
    {{- if and (ne (include "runai.hooksRender" $ctx) "none") (or (eq $phase "all") (has $phase $phases)) -}}
        {{- print true -}}
    {{- else -}}
        {{- print false -}}
    {{- end -}}
{{- end -}}

{{- define "fullImageName" -}}
    {{- $origContext := . -}}
    {{- with index .Values "runai-operator" "image" -}}
        {{- printf "%s/%s:%s" (include "imageRegistry" $origContext) .name .tag -}}
    {{- end -}}
{{- end -}}

{{- define "imagePullSecretNames" -}}
    {{- $names := list -}}
    {{- range (.Values.global.additionalImagePullSecrets | default list) -}}
        {{- if . -}}
            {{- $names = append $names . -}}
        {{- end -}}
    {{- end -}}
    {{- $names | toJson -}}
{{- end -}}

{{- define "imagePullSecretList" -}}
    {{- $names := include "imagePullSecretNames" . | fromJsonArray -}}
    {{- if eq (len $names) 0 -}}
        {{- print "[]" -}}
    {{- else -}}
    {{- range $names }}
        {{- println "- name:" . -}}
    {{- end }}
    {{- end -}}
{{- end -}}

{{- /*
additionalImagePullSecretsArg builds the value for the helm-hooks-handler
--additional-image-pull-secrets flag. Control-plane downloaded credentials are
appended by helm-hooks-handler at runtime.
*/ -}}
{{- define "additionalImagePullSecretsArg" -}}
    {{- $names := include "imagePullSecretNames" . | fromJsonArray -}}
    {{- $names | join "," -}}
{{- end -}}

{{- define "helmHooksHandler.image" -}}
    {{ $registry := .Values.global.image.registry | default .Values.helmHooksHandler.image.registry }}
    {{- printf "%s/%s:%s" $registry .Values.helmHooksHandler.image.name .Values.helmHooksHandler.image.tag -}}
{{- end -}}

{{- define "kartaMigrate.image" -}}
    {{ $registry := .Values.global.image.registry | default .Values.kartaMigrate.image.registry }}
    {{- printf "%s/%s:%s" $registry .Values.kartaMigrate.image.name .Values.kartaMigrate.image.tag -}}
{{- end -}}

{{- define "runaiConfigCRDExists" -}}
    {{- $exists := false -}}
    {{- $crd := (lookup "apiextensions.k8s.io/v1" "CustomResourceDefinition" "" "runaiconfigs.run.ai") -}}
    {{- if $crd -}}
        {{- $exists = true -}}
    {{- end -}}
    {{- $exists -}}
{{- end -}}

{{- define "defaultAffinity" -}}
nodeAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 1
      preference:
        matchExpressions:
          - key: node-role.kubernetes.io/runai-system
            operator: Exists
{{- end -}}

{{- define "runaiconfigGlobalAffinity" -}}
    {{- $runaiConfigCRDExists := (include "runaiConfigCRDExists" . ) -}}
    {{- if eq $runaiConfigCRDExists "true" -}}
        {{- $runaiconfig := (lookup "run.ai/v1" "RunaiConfig" .Release.Namespace "runai") -}}
        {{- if $runaiconfig -}}
            {{- $runaiconfigGlobal := $runaiconfig.spec.global -}}
            {{- if $runaiconfigGlobal -}}
                {{- if $runaiconfigGlobal.affinity -}}
                    {{- $runaiconfigGlobal.affinity | toYaml -}}
                {{- end -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- define "runaiconfigGlobalTolerations" -}}
    {{- $runaiConfigCRDExists := (include "runaiConfigCRDExists" . ) -}}
    {{- if eq $runaiConfigCRDExists "true" -}}
        {{- $runaiconfig := (lookup "run.ai/v1" "RunaiConfig" .Release.Namespace "runai") -}}
        {{- if $runaiconfig -}}
            {{- $runaiconfigGlobal := $runaiconfig.spec.global -}}
            {{- if $runaiconfigGlobal -}}
                {{- if $runaiconfigGlobal.tolerations -}}
                    {{- $runaiconfigGlobal.tolerations | toYaml -}}
                {{- end -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- define "runaiconfigGlobalTolerationsJson" -}}
    {{- $runaiConfigCRDExists := (include "runaiConfigCRDExists" . ) -}}
    {{- if eq $runaiConfigCRDExists "true" -}}
        {{- $runaiconfig := (lookup "run.ai/v1" "RunaiConfig" .Release.Namespace "runai") -}}
        {{- if $runaiconfig -}}
            {{- $runaiconfigGlobal := $runaiconfig.spec.global -}}
            {{- if $runaiconfigGlobal -}}
                {{- if $runaiconfigGlobal.tolerations -}}
                    {{- $runaiconfigGlobal.tolerations | toJson -}}
                {{- end -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- define "deploymentAffinity" -}}
    {{- $globalAffinity := .Values.global.affinity -}}
    {{- if $globalAffinity -}}
       {{- $globalAffinity | toYaml -}}
    {{- else -}}
        {{- $runaiconfigAffinity := (include "runaiconfigGlobalAffinity" .) -}}
        {{- if $runaiconfigAffinity -}}
            {{- $runaiconfigAffinity -}}
        {{- else -}}
            {{- include "defaultAffinity" . -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- define "deploymentAffinityJson" -}}
    {{- $affinity := include "deploymentAffinity" . -}}
    {{- $affinity | fromYaml | toJson | printf "%s" -}}
{{- end -}}

{{- define "deploymentTolerations" -}}
    {{- $globalTolerations := .Values.global.tolerations -}}
    {{- if $globalTolerations -}}
       {{- $globalTolerations | toYaml -}}
    {{- else -}}
        {{- $runaiconfigTolerations := (include "runaiconfigGlobalTolerations" .) -}}
        {{- if $runaiconfigTolerations -}}
            {{- $runaiconfigTolerations -}}
        {{- else -}}
            {{- $globalTolerations | toYaml -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- define "deploymentTolerationsJson" -}}
    {{- $globalTolerations := .Values.global.tolerations -}}
    {{- if $globalTolerations -}}
       {{- $globalTolerations | toJson | printf "%s"  -}}
    {{- else -}}
        {{- $runaiconfigTolerations := (include "runaiconfigGlobalTolerationsJson" .) -}}
        {{- if $runaiconfigTolerations -}}
            {{- $runaiconfigTolerations | printf "%s" -}}
        {{- else -}}
            {{- $globalTolerations | toJson | printf "%s"  -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- define "mgmtPrometheusIngressJson" -}}
    {{- $nmc := .Values.global.nmc -}}
    {{- if not $nmc -}}
        {{- "{}" -}}
    {{- else -}}
        {{- $mgmtPrometheusIngress := $nmc.mgmtPrometheusIngress -}}
        {{- if not $mgmtPrometheusIngress -}}
            {{- "{}" -}}
        {{- else -}}
            {{- $mgmtPrometheusIngress | toJson -}}
        {{- end -}}
    {{- end -}}
{{- end -}}


{{- define "globalFipsMode" -}}
{{- $fipsMode := .Values.global.fipsMode | default "off" -}}
{{- if not (has $fipsMode (list "off" "on" "only")) -}}
{{- fail (printf "global.fipsMode must be one of: off, on, only (got %q)" $fipsMode) -}}
{{- end -}}
{{- $fipsMode -}}
{{- end -}}

{{- define "leaderElectionArg" -}}
    {{- $replicas := int .replicas }}
    {{- if gt $replicas 1 -}}
        {{- print "- --leader-elect" -}}
    {{- else -}}
        {{- print "" -}}
    {{- end -}}
{{- end -}}

{{- define "isOpenShift" -}}
    {{- $mode := .Values.openShift.mode | default "AutoDetect" -}}
    {{- if eq $mode "Enabled" -}}
        {{- true -}}
    {{- else if eq $mode "Disabled" -}}
        {{- false -}}
    {{- else -}}
        {{- /* AutoDetect mode: check for clusterversions CRD */ -}}
        {{- $lookupResult := (lookup "apiextensions.k8s.io/v1" "CustomResourceDefinition" "" "clusterversions.config.openshift.io") -}}
        {{- not (empty $lookupResult) -}}
    {{- end -}}
{{- end -}}
