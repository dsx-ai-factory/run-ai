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
            .Values.cluster.uid)
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

{{- define "ccPrecheck.name" -}}
    {{- printf "%s-cc-precheck" .Release.Name | trunc 45 | trimSuffix "-" -}}
{{- end -}}

{{- define "ccPrecheck.dispatcherName" -}}
    {{- printf "%s-dispatcher-%d" (include "ccPrecheck.name" .) .Release.Revision -}}
{{- end -}}

{{- define "ccPrecheck.nodeSelector" -}}
    {{- if .Values.ccPrecheck.nodeSelector -}}
        {{- printf "%s,kubernetes.io/os=linux" .Values.ccPrecheck.nodeSelector -}}
    {{- else -}}
        kubernetes.io/os=linux
    {{- end -}}
{{- end -}}

{{- define "ccPrecheck.probeImage" -}}
    {{- $registry := .Values.global.image.registry | default .Values.ccPrecheck.probeImage.registry -}}
    {{- printf "%s/%s:%s" $registry .Values.ccPrecheck.probeImage.name .Values.ccPrecheck.probeImage.tag -}}
{{- end -}}

{{- define "ccPrecheck.collectorImage" -}}
    {{- $registry := .Values.global.image.registry | default .Values.ccPrecheck.collectorImage.registry -}}
    {{- printf "%s/%s:%s" $registry .Values.ccPrecheck.collectorImage.name .Values.ccPrecheck.collectorImage.tag -}}
{{- end -}}

{{- define "ccPrecheck.dispatcherImage" -}}
    {{- $registry := .Values.global.image.registry | default .Values.ccPrecheck.dispatcherImage.registry -}}
    {{- printf "%s/%s:%s" $registry .Values.ccPrecheck.dispatcherImage.name .Values.ccPrecheck.dispatcherImage.tag -}}
{{- end -}}

{{- define "ccPrecheck.validate" -}}
    {{- $cpu := 0 -}}
    {{- $gpu := false -}}
    {{- range $requirement := .Values.ccPrecheck.anyOf -}}
        {{- if not (has $requirement (list "gpu-cc" "gpu-ppcie")) -}}
            {{- fail "ccPrecheck.anyOf accepts gpu-cc and gpu-ppcie" -}}
        {{- end -}}
        {{- $gpu = true -}}
    {{- end -}}
    {{- range $requirement := .Values.ccPrecheck.requirements -}}
        {{- if not (has $requirement (list "cc" "tdx" "snp" "se" "gpu-cc" "gpu-ppcie")) -}}
            {{- fail (printf "Unknown ccPrecheck requirement: %s" $requirement) -}}
        {{- end -}}
        {{- if has $requirement (list "gpu-cc" "gpu-ppcie") -}}
            {{- $gpu = true -}}
        {{- else -}}
            {{- $cpu = add $cpu 1 -}}
        {{- end -}}
    {{- end -}}
    {{- if gt (int $cpu) 1 -}}
        {{- fail "Choose one CPU TEE requirement, or cc to allow any supported CPU TEE" -}}
    {{- end -}}
    {{- if and (not .Values.ccPrecheck.requirements) (not .Values.ccPrecheck.anyOf) -}}
        {{- fail "global.confidentialComputing.enabled requires ccPrecheck.requirements or ccPrecheck.anyOf" -}}
    {{- end -}}
    {{- if and .Values.ccPrecheck.gpuModels (not $gpu) -}}
        {{- fail "ccPrecheck.gpuModels requires gpu-cc or gpu-ppcie" -}}
    {{- end -}}
{{- end -}}
