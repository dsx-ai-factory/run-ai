/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import "strings"

var tenantSettingKeys = map[string]string{
	"DashboardShowOld":              "dashboard.show_old",
	"OptimizedAnalyticsQueries":     "dashboard.optimized_analytics_queries",
	"ProjectManualOverQuota":        "project.quota.manual_over_quota",
	"CPUResourcesQuota":             "project.quota.cpu_resources",
	"LogoutURI":                     "tenant.logout_uri",
	"UseOldInference":               "tenant.use_old_inference",
	"DisableAudit":                  "audit.disable",
	"EnableInactivityLogout":        "tenant.enable_inactivity_logout",
	"EnableScopedAuthorization":     "tenant.enable_scoped_authorization",
	"InactivityMinuetsTillLogout":   "tenant.inactivity_minutes_till_logout",
	"ExcludeViewerInactivityLogout": "tenant.exclude_viewers_from_inactivity_logout",
	"ConsumptionReport":             "tenant.show_consumption_report",
	"EnableAdvancedGPUMetrics":      "tenant.enable_advanced_gpu_metrics",
	"AllowGPT2Model":                "tenant.allow_gpt2_model",
	"LimitQuotaOverSubscription":    "tenant.limit_quota_over_subscription",
	"EnableCPUDashboards":           "tenant.enable_cpu_dashboards",
	"EnablePolicies":                "tenant.enable_policies",
	"EnableModelCatalog":            "tenant.enable_model_catalog",
	"EnableNIM":                     "tenant.enable_nim",
	"AdminMessage":                  "tenant.admin_message",
	"GpuOverProvisioning":           "tenant.gpu_over_provisioning",
	"WorkloadSubmissionV2Enabled":   "tenant.workload_submission_v2_enabled",
	"TenantIsCspManaged":            "tenant.is_csp_managed",
	"EnableNewCliInstaller":         "tenant.is_new_cli_installer_enabled",
	"DeletedWorkloadsView":          "tenant.deleted_workloads_view",
	"NewNavigation":                 "tenant.new_navigation",
	"EnableHybridSubmission":        "tenant.enable_hybrid_submission",
	"EnableCustomRoles":             "tenant.enable_custom_roles",
	"EnableDataVolumes":             "tenant.enable_data_volumes",
	"EnableTemplatesV2":             "tenant.enable_templates_v2",
	"EnableNgcIntegration":          "tenant.enable_ngc_integration",
}

type tenantSettingEntry struct {
	dbKey, originalKey string
}

var tenantSettingKeysLookup map[string]tenantSettingEntry

func init() {
	tenantSettingKeysLookup = make(map[string]tenantSettingEntry, len(tenantSettingKeys))
	for k, v := range tenantSettingKeys {
		tenantSettingKeysLookup[strings.ToLower(k)] = tenantSettingEntry{v, k}
	}
}

func LookupTenantSettingKey(key string) (dbKey string, originalKey string, ok bool) {
	e, ok := tenantSettingKeysLookup[strings.ToLower(key)]
	return e.dbKey, e.originalKey, ok
}
