# Admin Password — materialized by the secret-generator hook into
# runai-backend-gen-tenants-manager unless the user supplies an existingSecret
# that contains ADMIN_PASSWORD itself.
{{- define "control-plane.tenantsManager.config.adminPassword.secret.name" -}}
{{- $tm := .Values.tenantsManager | default .Values -}}
{{- if $tm.config.existingSecret -}}
{{ $tm.config.existingSecret }}
{{- else -}}
{{ printf "%s-gen-tenants-manager" .Values.fullnameOverride }}
{{- end -}}
{{- end -}}

{{- define "control-plane.tenantsManager.config.adminPassword.secret.key" -}}
{{- $tm := .Values.tenantsManager | default .Values -}}
{{- if $tm.config.existingSecret -}}
{{ $tm.config.secretKeys.adminPasswordKey | default "ADMIN_PASSWORD" }}
{{- else -}}
ADMIN_PASSWORD
{{- end -}}
{{- end -}}
