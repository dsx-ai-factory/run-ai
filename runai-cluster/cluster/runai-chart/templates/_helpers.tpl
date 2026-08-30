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

{{- define "shouldApplyFreshInstallLegacyDefaults" -}}
    {{- $appVersion := trimPrefix "v" .Chart.AppVersion -}}
    {{- if or (eq $appVersion "") (eq $appVersion "CHART_APP_VERSION") (hasPrefix "0.0.0" $appVersion) -}}
        {{- true -}}
    {{- else -}}
        {{- semverCompare ">=3.0.0-0" $appVersion -}}
    {{- end -}}
{{- end -}}

{{- define "legacyControllerRBACEnabled" -}}
    {{- $root := index . 0 -}}
    {{- $controller := index . 1 -}}
    {{- if $root.Release.IsInstall -}}
        {{- false -}}
    {{- else -}}
        {{- $enabled := true -}}
        {{- if eq (include "runaiConfigCRDExists" $root) "true" -}}
            {{- $runaiConfig := lookup "run.ai/v1" "RunaiConfig" $root.Release.Namespace "runai" -}}
            {{- if $runaiConfig -}}
                {{- $controllerConfig := get $runaiConfig.spec $controller -}}
                {{- if and $controllerConfig (hasKey $controllerConfig "enabled") -}}
                    {{- $enabled = get $controllerConfig "enabled" -}}
                {{- end -}}
            {{- end -}}
        {{- end -}}
        {{- $enabled -}}
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
{{- if kindIs "bool" $fipsMode -}}
{{- $fipsMode = ternary "on" "off" $fipsMode -}}
{{- end -}}
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
