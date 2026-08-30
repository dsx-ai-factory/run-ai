package auth

// Test-only: ServeHTTP writes headers via runaictxhttp.GetHeaders, not these.
const (
	HeaderTenantId           = "X-Runai-Tenant-Id"
	HeaderSubjectId          = "X-Runai-Subject-Id"
	HeaderSubjectType        = "X-Runai-Subject-Type"
	HeaderGroups             = "X-Runai-Groups"
	HeaderSupplementalGroups = "X-Runai-Supplemental-Groups"
	HeaderGID                = "X-Runai-GID"
	HeaderUID                = "X-Runai-UID"
	HeaderIsManagementTenant = "X-Runai-Is-Management-Tenant"
	HeaderClientId           = "X-Runai-Client-Id"
	HeaderSessionId          = "X-Runai-Session-Id"
)
