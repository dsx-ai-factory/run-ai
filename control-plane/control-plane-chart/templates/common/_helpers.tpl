{{- define "control-plane.image-pull-secret.name"  -}}
{{- $imagePullSecret := .Values.global.imagePullSecrets | default list | first -}}
{{- if $imagePullSecret -}}
{{ $imagePullSecret.name | default "runai-reg-creds" }}
{{- else -}}
runai-reg-creds
{{- end -}}
{{- end -}}

{{- define "control-plane.image-pull-secret.enabled" -}}
{{- $globalImagePullSecrets := .Values.global.imagePullSecrets | default list -}}
{{- and .Values.clusterService.config.imagePullSecret.enabled (gt (len $globalImagePullSecrets) 0) -}}
{{- end -}}

{{/*
Effective enabled state for the telemetryService (call-home) support-entitled
feature. Mirrors the Chart.yaml dependency condition
"telemetryService.enabled,supportEntitled": an explicit boolean
telemetryService.enabled wins (opt-in/opt-out); otherwise it derives from
supportEntitled. Only a real boolean is honored (matching Helm's condition,
which ignores non-bool values), so a stringified "false" can't leave the
subchart and this gate out of sync. Returns the string "true"/"false"; compare
with (eq ... "true").
*/}}
{{- define "control-plane.telemetry.enabled" -}}
{{- if kindIs "bool" .Values.telemetryService.enabled -}}
{{- .Values.telemetryService.enabled -}}
{{- else -}}
{{- .Values.supportEntitled -}}
{{- end -}}
{{- end -}}

{{/* ClusterService helpers for application metrics write secret wiring */}}
{{- define "control-plane.clusterService.secret.name" -}}
{{ printf "%s-%s" .Values.fullnameOverride .Values.clusterService.name }}
{{- end -}}

{{- define "control-plane.clusterService.applicationMetricsWrite.password.secret.name" -}}
{{- if .Values.clusterService.config.applicationMetricsWrite.existingSecret -}}
{{ .Values.clusterService.config.applicationMetricsWrite.existingSecret }}
{{- else -}}
{{ include "control-plane.clusterService.secret.name" . }}
{{- end -}}
{{- end -}}

{{/* ClusterService helpers for Cloudflare token secret wiring */}}
{{- define "control-plane.clusterService.cloudflare.token.secret.name" -}}
{{- $cf := .Values.clusterService.config.cloudflare | default dict -}}
{{- if $cf.existingSecret -}}
{{ $cf.existingSecret }}
{{- else -}}
{{ include "control-plane.clusterService.secret.name" . }}
{{- end -}}
{{- end -}}

{{- define "control-plane.tenantsManager.secret.name" -}}
{{ printf "%s-%s" .Values.fullnameOverride .Values.tenantsManager.name }}
{{- end -}}

{{- define "control-plane.clusterService.cloudflare.token.secret.key" -}}
{{- $cf := .Values.clusterService.config.cloudflare | default dict -}}
{{- if $cf.existingSecret -}}
{{ $cf.secretKeys.tokenKey | default "CF_TOKEN" }}
{{- else -}}
CF_TOKEN
{{- end -}}
{{- end -}}

{{- define "control-plane.clusterService.applicationMetricsWrite.password.secret.key" -}}
{{ .Values.clusterService.config.applicationMetricsWrite.secretKeys.passwordKey | default "APPLICATION_METRIC_STORE_WRITE_PASSWORD" }}
{{- end -}}

{{- define "control-plane.postgresql.password.secret.name" -}}
{{ .Values.global.postgresql.auth.existingSecret }}
{{- end -}}

{{- define "control-plane.postgresql.password.secret.key" -}}
{{- if .Values.global.postgresql.auth.existingSecret -}} 
{{ .Values.global.postgresql.auth.secretKeys.userPasswordKey }} 
{{- else -}}
password
{{- end -}}
{{- end -}}

{{- define "control-plane.postgresql.admin.password.secret.name" -}}
{{ .Values.global.postgresql.auth.existingSecret }}
{{- end -}}

{{- define "control-plane.postgresql.admin.password.secret.key" -}}
{{- if .Values.global.postgresql.auth.existingSecret -}} 
{{ .Values.global.postgresql.auth.secretKeys.adminPasswordKey }}
{{- else -}}
postgres-password
{{- end -}}
{{- end -}}

{{/*
  Keycloakx generated secret name.
  Hardcoded because this helper is called from both the parent chart and the
  keycloakx subchart (via tpl). In the subchart context .Values.fullnameOverride
  resolves to "keycloak" (the subchart's own override) instead of the parent's
  "runai-backend", which would produce a wrong secret name. Hardcoding keeps
  the name consistent with what the secret-generator hook creates.
*/}}
{{- define "control-plane.keycloakx.password.secret.name" -}}
{{- $keycloakxValues := .Values.keycloakx | default .Values -}}
{{ $keycloakxValues.existingSecret | default "runai-backend-gen-keycloakx" }}
{{- end -}}

{{/* NOTE - not usable in subcharts other than Keycloakx's */}}
{{- define "control-plane.keycloakx.password.secret.key" -}}
{{- $keycloakxValues := .Values.keycloakx | default .Values -}}
{{- if $keycloakxValues.existingSecret -}} 
{{ $keycloakxValues.secretKeys.adminPasswordKey }}
{{- else -}}
KEYCLOAK_ADMIN_PASSWORD
{{- end -}}
{{- end -}}

{{- define "control-plane.cache.password.secret.name" -}}
{{- $cacheValues := .Values.global.cache | default .Values -}}
{{ $cacheValues.auth.existingSecret | default (printf "%s-gen-cache-pubsub" .Values.fullnameOverride) }}
{{- end -}}

{{- define "control-plane.cache.password.secret.key" -}}
{{- $cacheValues := .Values.global.cache | default .Values -}}
{{ $cacheValues.auth.secretKeys.passwordKey }}
{{- end -}}

{{- define "control-plane.pubsub.password.secret.name" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{ $pubsubValues.auth.existingSecret | default (printf "%s-gen-cache-pubsub" .Values.fullnameOverride) }}
{{- end -}}

{{- define "control-plane.pubsub.password.secret.key" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{ $pubsubValues.auth.secretKeys.passwordKey }}
{{- end -}}

{{- define "control-plane.pubsub.additional-subscriber.password.secret.name" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{- if $pubsubValues.additionalSubscriber.auth.existingSecret -}}
{{ $pubsubValues.additionalSubscriber.auth.existingSecret }}
{{- else -}}
{{ include "control-plane.pubsub.chart-managed.secret.name" . }}
{{- end -}}
{{- end -}}

{{- define "control-plane.pubsub.additional-subscriber.password.secret.key" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{ $pubsubValues.additionalSubscriber.auth.secretKeys.passwordKey }}
{{- end -}}

{{/*
Name of the chart-managed Secret that holds non-sensitive pubsub
configuration metadata (the pubsub user and the additional-subscriber
user). Rendered by templates/pubsub/secret.yaml.
*/}}
{{- define "control-plane.pubsub.chart-managed.secret.name" -}}
{{ printf "%s-pubsub" .Values.fullnameOverride }}
{{- end -}}

{{- define "control-plane.pubsub.user.secret.name" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{ $pubsubValues.auth.existingSecret | default (printf "%s-gen-cache-pubsub" .Values.fullnameOverride) }}
{{- end -}}

{{- define "control-plane.pubsub.user.secret.key" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{ $pubsubValues.auth.secretKeys.userKey }}
{{- end -}}

{{- define "control-plane.pubsub.additional-subscriber.user.secret.name" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{- if $pubsubValues.additionalSubscriber.auth.existingSecret -}}
{{ $pubsubValues.additionalSubscriber.auth.existingSecret }}
{{- else -}}
{{ include "control-plane.pubsub.chart-managed.secret.name" . }}
{{- end -}}
{{- end -}}

{{- define "control-plane.pubsub.additional-subscriber.user.secret.key" -}}
{{- $pubsubValues := .Values.global.pubsub | default .Values -}}
{{ $pubsubValues.additionalSubscriber.auth.secretKeys.userKey }}
{{- end -}}

{{- define "control-plane.notifications-proxy.slack.secret.name" -}}
{{ .Values.notificationsProxy.config.slack.existingSecret | default (printf "%s-notifications-proxy" .Values.fullnameOverride) }}
{{- end -}}

{{- define "control-plane.notifications-proxy.slack.client.id.key" -}}
{{ .Values.notificationsProxy.config.slack.secretKeys.clientId }}
{{- end -}}

{{- define "control-plane.notifications-proxy.slack.client.secret.key" -}}
{{ .Values.notificationsProxy.config.slack.secretKeys.clientSecret }}
{{- end -}}


{{- define "control-plane.notificationsProxy.encryptionKey.secret.name" -}}
{{ printf "%s-gen-notifications-proxy-encryption-key" .Values.fullnameOverride }}
{{- end -}}

{{- define "control-plane.notificationsService.gen.secret.name" -}}
{{ printf "%s-gen-notifications-service" .Values.fullnameOverride }}
{{- end -}}

{{/*
  Management helpers for user and password secret wiring.
  TENANT_MNG_USER lives in the chart-managed runai-backend-tenants-manager Secret
  (rendered by templates/tenants-manager/secret.yaml in MT mode).
  TENANT_MNG_PASSWORD lives in the secret-generator-managed
  runai-backend-gen-tenants-manager Secret (alongside ADMIN_PASSWORD), so the
  user and password resolve to different Secrets unless the operator supplies
  an existingSecret that overrides both.
*/}}
{{- define "control-plane.management.user.secret.name" -}}
{{- $management := .Values.global.management | default dict -}}
{{ $management.existingSecret | default (include "control-plane.tenantsManager.secret.name" .) }}
{{- end -}}

{{- define "control-plane.management.password.secret.name" -}}
{{- $management := .Values.global.management | default dict -}}
{{- if $management.existingSecret -}}
{{ $management.existingSecret }}
{{- else -}}
{{ printf "%s-gen-tenants-manager" .Values.fullnameOverride }}
{{- end -}}
{{- end -}}

{{- define "control-plane.management.user.secret.key" -}}
{{- $management := .Values.global.management | default dict -}}
{{- if $management.existingSecret -}}
{{ $management.secretKeys.userKey | default "TENANT_MNG_USER" }}
{{- else -}}
TENANT_MNG_USER
{{- end -}}
{{- end -}}

{{- define "control-plane.management.password.secret.key" -}}
{{- $management := .Values.global.management | default dict -}}
{{- if $management.existingSecret -}}
{{ $management.secretKeys.passwordKey | default "TENANT_MNG_PASSWORD" }}
{{- else -}}
TENANT_MNG_PASSWORD
{{- end -}}
{{- end -}}

{{/*
Conditional ingress annotations based on ingress controller class.
Looks up controller-specific annotations from values and merges with custom annotations.
Usage: include "control-plane.ingress.annotations" (dict "ingressClass" "nginx" "customAnnotations" {} "ingressClassAnnotations" .Values.global.ingress.ingressClassAnnotations)
*/}}
{{- define "control-plane.ingress.annotations" -}}
{{- $ingressClass := .ingressClass | default "nginx" -}}
{{- $customAnnotations := .customAnnotations | default dict -}}
{{- $ingressClassAnnotations := .ingressClassAnnotations | default dict -}}

{{- $baseAnnotations := index $ingressClassAnnotations $ingressClass | default dict -}}
{{- $merged := merge $customAnnotations $baseAnnotations -}}
{{- toYaml $merged -}}
{{- end -}}
