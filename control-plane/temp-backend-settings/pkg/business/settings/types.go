/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"time"
)

const (
	SettingKeyDashboardShowOld              = "dashboard.show_old"
	SettingKeyOptimizedAnalyticsQueries     = "dashboard.optimized_analytics_queries"
	SettingKeyProjectManualOverQuota        = "project.quota.manual_over_quota"
	SettingKeyCPUResourcesQuota             = "project.quota.cpu_resources"
	SettingKeyLogoutURI                     = "tenant.logout_uri"
	SettingKeyUseOldInference               = "tenant.use_old_inference"
	SettingKeyDisableAudit                  = "audit.disable"
	SettingKeyEnableInactivityLogout        = "tenant.enable_inactivity_logout"
	SettingKeyEnableScopedAuthorization     = "tenant.enable_scoped_authorization"
	SettingKeyInactivityMinutesTillLogout   = "tenant.inactivity_minutes_till_logout"
	SettingKeyExcludeViewerInactivityLogout = "tenant.exclude_viewers_from_inactivity_logout"
	SettingKeyConsumptionReport             = "tenant.show_consumption_report"
	SettingKeyEnableAdvancedGPUMetrics      = "tenant.enable_advanced_gpu_metrics"
	SettingKeyAllowGPT2Model                = "tenant.allow_gpt2_model"
	SettingKeyLimitQuotaOverSubscription    = "tenant.limit_quota_over_subscription"
	SettingKeyEnableCPUDashboards           = "tenant.enable_cpu_dashboards"
	SettingKeyEnablePolicies                = "tenant.enable_policies"
	SettingKeyEnableModelCatalog            = "tenant.enable_model_catalog"
	SettingKeyEnableNIM                     = "tenant.enable_nim"
	SettingKeyAdminMessage                  = "tenant.admin_message"
	SettingKeyGpuOverProvisioning           = "tenant.gpu_over_provisioning"
	SettingKeyWorkloadSubmissionV2Enabled   = "tenant.workload_submission_v2_enabled"
	SettingKeyTenantIsCspManaged            = "tenant.is_csp_managed"
	SettingKeyEnableNewCliInstaller         = "tenant.is_new_cli_installer_enabled"
	SettingKeyDeletedWorkloadsView          = "tenant.deleted_workloads_view"
	SettingKeyNewNavigation                 = "tenant.new_navigation"
	SettingKeyEnableHybridSubmission        = "tenant.enable_hybrid_submission"
	SettingKeyEnableCustomRoles             = "tenant.enable_custom_roles"
	SettingKeyEnableDataVolumes             = "tenant.enable_data_volumes"
	SettingKeyEnableTemplatesV2             = "tenant.enable_templates_v2"
	SettingKeyEnableNgcIntegration          = "tenant.enable_ngc_integration"
)

const (
	PrivilegeTest   = "Test"
	PrivilegeAdmin  = "Admin"
	PrivilegeReader = "Reader"
	PrivilegeWriter = "Writer"
)

const (
	StageAlpha  = "Alpha"
	StageBeta   = "Beta"
	StageGA     = "GA"
	StageLegacy = "Legacy"
)

const (
	SourceTenant  = "Tenant"
	SourceProfile = "Profile"
)

const (
	CategoryTenant     = "Tenant"
	CategoryDashboard  = "Dashboard"
	CategoryDepartment = "Department"
	CategoryProject    = "Project"
	CategoryAudit      = "Audit"
)

const (
	TypeBoolean = "Boolean"
	TypeNumber  = "Number"
	TypeString  = "String"
)

var (
	OptimizedAnalyticsDefaultFalseUntil = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	NewCliDefaultFalseUntil             = time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)
	NewNavigationDefaultFalseUntil      = time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)
)

const DefaultSessionTimeout = 60

type TenantStaticData struct {
	CreatedAt time.Time
}

type OnSetResult struct {
	SkipDaoSave bool
	Value       interface{}
	HasNewValue bool
}

type OnSetHandler func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error)

type PostSaveHandler func(ctx context.Context, tenantID int, value interface{}) (revertValue interface{}, err error)

type ValidateFunc func(value interface{}) string

type SettingDefinition struct {
	Key                         string
	Source                      string
	Category                    string
	Description                 string
	Label                       string
	Type                        string
	Privilege                   string
	DefaultValue                interface{}
	Stage                       string
	Filter                      func(value interface{}) bool
	Validate                    ValidateFunc
	MinimumSingleClusterVersion string
	IsJson                      bool
	Migrated                    bool
	OnSet                       OnSetHandler
	PostSave                    PostSaveHandler
}

type SettingOverride struct {
	Value       interface{}
	Description string
	DisplayName string
	Stage       string
}

type SettingProvider interface {
	FetchSettings(ctx context.Context, tenantID int) (map[string]SettingOverride, error)
}
