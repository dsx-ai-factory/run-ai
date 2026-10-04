/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/dao"
)

type OnSetDeps struct {
	OrgUnitSettingsUpdater  OrgUnitSettingsUpdater
	IdentityManagerUpdater  IdentityManagerUpdater
	BffWorkloadsUpdater     BffWorkloadsUpdater
	BffNotificationsUpdater BffNotificationsUpdater
	WorkloadsUpdater        WorkloadsUpdater
	CatalogUpdater          CatalogUpdater
	DatavolumesUpdater      DatavolumesUpdater
	AuthorizationUpdater    AuthorizationUpdater
	ClusterUpdater          ClusterUpdater
	SettingsDao             *dao.TenantSettingsDAO
}

type OrgUnitSettingsUpdater interface {
	UpdateOrgUnitSettings(ctx context.Context, tenantID int, name string, value string) error
	UpdateLimitQuotaOverSubscription(ctx context.Context, tenantID int, value interface{}) error
}

type IdentityManagerUpdater interface {
	UpdateLogoutRedirectUri(ctx context.Context, tenantID int, uri interface{}) error
	UpdateBrowserSessionTimeout(ctx context.Context, tenantID int, timeout interface{}) error
}

type BffWorkloadsUpdater interface {
	UpdateDeletedWorkloadsView(ctx context.Context, tenantID int, enabled interface{}) error
	UpdateAllowGPT2Model(ctx context.Context, tenantID int, enabled interface{}) error
	UpdateEnableModelCatalog(ctx context.Context, tenantID int, enabled interface{}) error
	UpdateEnableNIM(ctx context.Context, tenantID int, enabled interface{}) error
}

type BffNotificationsUpdater interface {
	UpdateAdminMessage(ctx context.Context, tenantID int, msg interface{}) error
}

type CatalogUpdater interface {
	UpdateEnableNgcIntegration(ctx context.Context, tenantID int, value interface{}) error
}

type AuthorizationUpdater interface {
	UpdateEnableScopedAuthorization(ctx context.Context, tenantID int, value interface{}) error
}

type ClusterUpdater interface {
	UpdateEnableAdvancedGPUMetrics(ctx context.Context, tenantID int, value interface{}) error
	UpdateProjectManualOverQuota(ctx context.Context, tenantID int, value interface{}) error
	UpdateCPUResourcesQuota(ctx context.Context, tenantID int, value interface{}) error
}

type DatavolumesUpdater interface {
	UpdateEnableDataVolumes(ctx context.Context, tenantID int, value interface{}) error
}

type WorkloadsUpdater interface {
	UpdateSetting(ctx context.Context, tenantID int, key string, value interface{}) error
}

func BuildSettingDefinitions(deps *OnSetDeps) map[string]SettingDefinition {
	settings := []SettingDefinition{
		{
			Key:          SettingKeyLogoutURI,
			Source:       SourceTenant,
			Category:     CategoryDashboard,
			Description:  "page to load after logout",
			Label:        "Logout uri",
			Type:         TypeString,
			Privilege:    PrivilegeAdmin,
			DefaultValue: nil,
			Validate:     ValidateNone,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.IdentityManagerUpdater != nil {
					if err := deps.IdentityManagerUpdater.UpdateLogoutRedirectUri(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update logout redirect URI in identity manager: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyProjectManualOverQuota,
			Source:   SourceTenant,
			Category: CategoryProject,
			Description: "Enable over quota weight to set what proportion of unused resources a project can get on top of its quota. " +
				"For more information, see the <a data-doc-key='overQuotaWeight' target='_blank'>Over quota weight guide</a>",
			Label:        "Over quota weight",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.ClusterUpdater != nil {
					if err := deps.ClusterUpdater.UpdateProjectManualOverQuota(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update ProjectManualOverQuota in cluster service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyLimitQuotaOverSubscription,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable to prevent projects' total quota from exceeding the department's quota or being less than " +
				"the total usage by non-preemptible workloads. For more information, see the " +
				"<a data-doc-key='overSubscription' target='_blank'>Run:ai scheduler guide</a>",
			Label:        "Limit projects from exceeding department quota",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.OrgUnitSettingsUpdater != nil {
					if err := deps.OrgUnitSettingsUpdater.UpdateLimitQuotaOverSubscription(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update %s in org-unit service: %w", SettingKeyLimitQuotaOverSubscription, err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyEnableCPUDashboards,
			Source:   SourceTenant,
			Category: CategoryDashboard,
			Description: "Enable the CPU analytics dashboard for better visibility of CPU intensive workloads. " +
				"For more information, see the <a href='https://docs.run.ai/latest/admin/admin-ui-setup/" +
				"dashboard-analysis/?h=dashboa#cpu-dashboard' target='_blank'>CPU dashboard guide</a>",
			Stage:                       StageLegacy,
			Label:                       "CPU analytics",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                false,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.13",
		},
		{
			Key:      SettingKeyCPUResourcesQuota,
			Source:   SourceTenant,
			Category: CategoryProject,
			Description: "Enable resource quota based on CPU resources.\nFor more information, see the " +
				"<a data-doc-key='departments' target='_blank'>Departments guide</a> and the <a data-doc-key='projects' target='_blank'>Projects guide</a>",
			Label:        "CPU quota",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.ClusterUpdater != nil {
					if err := deps.ClusterUpdater.UpdateCPUResourcesQuota(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update CPUResourcesQuota in cluster service: %w", err)
					}
				}
				return nil, nil
			},
			// No PostSave: disabling CPU quota wipes org-unit project/department quota via org-unit-service's
			// DisableCpuQuota, which cluster-service now triggers from UpdateNodePoolSettingByKey (reached by both
			// this OnSet's ClusterUpdater call and the frontend's direct write). A PostSave here would double-wipe.
		},
		{
			Key:          SettingKeyUseOldInference,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Use Old Inference",
			Stage:        StageBeta,
			Label:        "Use Old Inference",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Filter:       filterNever,
		},
		{
			Key:          SettingKeyDeletedWorkloadsView,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Enable to view deleted workloads in the Workloads view.",
			Label:        "Deleted workloads",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: true,
			Validate:     ValidateBoolean,
			Migrated:     true,
			Filter:       filterNever,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.BffWorkloadsUpdater != nil {
					if err := deps.BffWorkloadsUpdater.UpdateDeletedWorkloadsView(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update deleted workloads view in BFF service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyEnablePolicies,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable to impose restrictions and default values on workloads' fields. " +
				"For more information, see the <a data-doc-key='policies' target='_blank'>Policies guide</a>",
			Label:                       "Policies",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.18",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.WorkloadsUpdater != nil {
					if err := deps.WorkloadsUpdater.UpdateSetting(ctx, tenantID, SettingKeyEnablePolicies, value); err != nil {
						return nil, fmt.Errorf("failed to update EnablePolicies in workloads service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:                         SettingKeyWorkloadSubmissionV2Enabled,
			Source:                      SourceTenant,
			Category:                    CategoryTenant,
			Description:                 "Enable submission of workload with the new API",
			Stage:                       StageAlpha,
			Label:                       "Workload submission v2",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.18",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.WorkloadsUpdater != nil {
					if err := deps.WorkloadsUpdater.UpdateSetting(ctx, tenantID, SettingKeyWorkloadSubmissionV2Enabled, value); err != nil {
						return nil, fmt.Errorf("failed to update WorkloadSubmissionV2Enabled in workloads service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyEnableNewCliInstaller,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable for better data consistency with the UI, auto-upgrade, and interactive mode. " +
				"For more information, see the <a data-doc-key='cliReference' target='_blank'>CLI guide</a>",
			Label:                       "Improved command<br>line interface",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.18",
		},
		{
			Key:      SettingKeyAllowGPT2Model,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "GPT2 model - present OpenAI's GPT2 model in the catalog in favor of demos. " +
				"This model's compute requirements are relatively low (CPU resources are sufficient for it). " +
				"The underlying image of this model also consists of a Chatbot UI",
			Stage:        StageAlpha,
			Label:        "GPT2 model",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Migrated:     true,
			Filter:       filterNever,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.BffWorkloadsUpdater != nil {
					if err := deps.BffWorkloadsUpdater.UpdateAllowGPT2Model(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update allow GPT2 model in BFF service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:          SettingKeyDisableAudit,
			Source:       SourceTenant,
			Category:     CategoryAudit,
			Description:  "Disable Audit",
			Stage:        StageBeta,
			Label:        "Disable Audit",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Filter:       filterNever,
		},
		{
			Key:          SettingKeyEnableInactivityLogout,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Enable automatic logout after user inactivity",
			Label:        "Enable automatic logout after user inactivity",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: true,
			Validate:     ValidateBoolean,
			Filter:       filterNever,
		},
		{
			Key:          SettingKeyInactivityMinutesTillLogout,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Session inactivity timeout (in minutes)",
			Label:        "Session inactivity timeout (in minutes)",
			Type:         TypeNumber,
			Privilege:    PrivilegeAdmin,
			DefaultValue: float64(DefaultSessionTimeout),
			Validate:     ValidateNumber,
			Stage:        StageLegacy,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.IdentityManagerUpdater != nil {
					if err := deps.IdentityManagerUpdater.UpdateBrowserSessionTimeout(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update session timeout in identity manager: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:          SettingKeyExcludeViewerInactivityLogout,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Exclude 'viewer' role from inactivity logout",
			Stage:        StageBeta,
			Label:        "Exclude 'viewer' role from inactivity logout",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Filter:       filterNever,
		},
		{
			Key:      SettingKeyConsumptionReport,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable the consumption dashboard to view consumption by segment. " +
				"For more information, see the <a href='https://docs.run.ai/latest/admin/admin-ui-setup/" +
				"dashboard-analysis/?h=consumpt#consumption-dashboard' target='_blank'>Consumption dashboard guide</a>",
			Stage:        StageLegacy,
			Label:        "Consumption",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Migrated:     true,
			Filter:       filterNever,
		},
		{
			Key:      SettingKeyEnableAdvancedGPUMetrics,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable to view metrics in the Node and Workload metrics tab. " +
				"To configure the <a href='https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/" +
				"feature-overview.html#metrics' target='_blank'>NVIDIA exporter</a> and NVIDIA Run:ai cluster, see " +
				"<a data-doc-key='gpuProfiling' target='_blank'>GPU profiling metrics</a>.",
			Label:                       "GPU profiling metrics",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                false,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.10",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.ClusterUpdater != nil {
					if err := deps.ClusterUpdater.UpdateEnableAdvancedGPUMetrics(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update EnableAdvancedGPUMetrics in cluster service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:         SettingKeyDashboardShowOld,
			Source:      SourceTenant,
			Category:    CategoryDashboard,
			Description: "Show an older version of the Analytics dashboard",
			Label:       "Show Old Analytics Dashboard",
			Type:        TypeBoolean,
			Privilege:   PrivilegeAdmin,
			Validate:    ValidateBoolean,
			Filter:      filterShowOnlyWhenConfigured,
		},
		{
			Key:          SettingKeyAdminMessage,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Label:        "Write a message that will be displayed to all users",
			Description:  "The message will be displayed every time users sign in and immediately after it is saved to users who are already signed in.",
			Type:         TypeString,
			Privilege:    PrivilegeAdmin,
			DefaultValue: map[string]interface{}{"msg": "", "id": "", "createdAt": float64(0), "enableDismiss": false},
			Filter:       filterNever,
			Validate:     ValidateAdminMessage,
			IsJson:       true,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.BffNotificationsUpdater != nil {
					if err := deps.BffNotificationsUpdater.UpdateAdminMessage(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update admin message in BFF service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:          SettingKeyOptimizedAnalyticsQueries,
			Source:       SourceTenant,
			Category:     CategoryDashboard,
			Description:  "Improve dashboard performance with enhanced queries",
			Label:        "Dashboard performance optimization",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: true,
			Validate:     ValidateBoolean,
		},
		{
			Key:      SettingKeyEnableModelCatalog,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable instant deployment of Hugging Face models via the Models view or the workload " +
				"submission form. For more information, see the <a data-doc-key='huggingFaceInference' target='_blank'>" +
				"Deploy Inference Workloads from Hugging Face</a>",
			Label:                       "Hugging face models",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.16",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.BffWorkloadsUpdater != nil {
					if err := deps.BffWorkloadsUpdater.UpdateEnableModelCatalog(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update enable model catalog in BFF service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyEnableNIM,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable instant deployment of NIM models via the workload submission form. " +
				"For more information, see the <a data-doc-key='nimInference' target='_blank'>" +
				"Deploy Inference Workloads with NVIDIA NIM</a>",
			Label:                       "NIM models",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.19",
			Migrated:                    true,
			Filter:                      filterNever,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.BffWorkloadsUpdater != nil {
					if err := deps.BffWorkloadsUpdater.UpdateEnableNIM(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update enable NIM in BFF service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyEnableHybridSubmission,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable to submit workloads via the new flexible form. Allow selecting an existing setup " +
				"or starting from scratch, reviewing existing setups and understanding policy definitions. " +
				"The form also includes advanced features tailored for inference workloads. For more information, see " +
				"<a data-doc-key='flexibleSubmission' target='_blank'>Flexible submission.</a>",
			Label:                       "Flexible workload submission",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.18",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				boolVal, _ := value.(bool)
				if !boolVal && deps.SettingsDao != nil {
					if err := deps.SettingsDao.Set(ctx, tenantID, map[string]interface{}{
						SettingKeyEnableTemplatesV2: false,
						SettingKeyEnableDataVolumes: false,
					}); err != nil {
						return nil, fmt.Errorf("failed to update tenant settings: %w", err)
					}
				}
				if deps.WorkloadsUpdater != nil {
					if err := deps.WorkloadsUpdater.UpdateSetting(ctx, tenantID, SettingKeyEnableHybridSubmission, value); err != nil {
						return nil, fmt.Errorf("failed to update EnableHybridSubmission in workloads service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyEnableTemplatesV2,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable to create workload templates via the new flexible form. " +
				"For more information, see the <a data-doc-key='workloadTemplates' target='_blank'>Workload Templates guide</a>",
			Label:                       "Flexible workload templates",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.22",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.WorkloadsUpdater != nil {
					if err := deps.WorkloadsUpdater.UpdateSetting(ctx, tenantID, SettingKeyEnableTemplatesV2, value); err != nil {
						return nil, fmt.Errorf("failed to update EnableTemplatesV2 in workloads service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyEnableDataVolumes,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable to allow creation of data volumes to share underlying data across scopes. " +
				"For more information, see the <a data-doc-key='dataVolumes' target='_blank'>Data volumes guide</a>",
			Label:                       "Data volumes",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                true,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.19",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.DatavolumesUpdater != nil {
					if err := deps.DatavolumesUpdater.UpdateEnableDataVolumes(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update EnableDataVolumes in datavolumes service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:          SettingKeyEnableNgcIntegration,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Enable to browse and select container images from the NVIDIA NGC catalog for faster environment setup.",
			Label:        "NGC catalog",
			Type:         TypeBoolean,
			Stage:        StageBeta,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.CatalogUpdater != nil {
					if err := deps.CatalogUpdater.UpdateEnableNgcIntegration(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update EnableNgcIntegration in catalog service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:      SettingKeyGpuOverProvisioning,
			Source:   SourceTenant,
			Category: CategoryTenant,
			Description: "Enable to set a GPU memory limit for workloads. To enable this on each Run:ai cluster, see the " +
				"<a data-doc-key='dynamicFractions' target='_blank'>Dynamic GPU fractions guide</a>",
			Stage:                       StageBeta,
			Label:                       "GPU resource optimization",
			Type:                        TypeBoolean,
			Privilege:                   PrivilegeAdmin,
			DefaultValue:                false,
			Validate:                    ValidateBoolean,
			MinimumSingleClusterVersion: "2.17",
			Migrated:                    true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.WorkloadsUpdater != nil {
					if err := deps.WorkloadsUpdater.UpdateSetting(ctx, tenantID, SettingKeyGpuOverProvisioning, value); err != nil {
						return nil, fmt.Errorf("failed to update GpuOverProvisioning in workloads service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:          SettingKeyEnableScopedAuthorization,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Enable scoped Authorization",
			Stage:        StageAlpha,
			Label:        "Enable scoped Authorization",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			DefaultValue: false,
			Validate:     ValidateBoolean,
			Migrated:     true,
			OnSet: func(ctx context.Context, tenantID int, value interface{}) (*OnSetResult, error) {
				if deps.AuthorizationUpdater != nil {
					if err := deps.AuthorizationUpdater.UpdateEnableScopedAuthorization(ctx, tenantID, value); err != nil {
						return nil, fmt.Errorf("failed to update EnableScopedAuthorization in authorization service: %w", err)
					}
				}
				return nil, nil
			},
		},
		{
			Key:          SettingKeyNewNavigation,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Force the new navigation experience for all users. This setting will override the user's preference.",
			Label:        "New navigation experience",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			Stage:        StageAlpha,
			DefaultValue: false,
			Validate:     ValidateBoolean,
		},
		{
			Key:          SettingKeyEnableCustomRoles,
			Source:       SourceTenant,
			Category:     CategoryTenant,
			Description:  "Enable CRUD of custom roles.",
			Label:        "Enable custom roles",
			Type:         TypeBoolean,
			Privilege:    PrivilegeAdmin,
			Stage:        StageAlpha,
			DefaultValue: false,
			Validate:     ValidateBoolean,
		},
	}

	m := make(map[string]SettingDefinition, len(settings))
	for _, s := range settings {
		m[s.Key] = s
	}
	return m
}

func filterNever(_ interface{}) bool {
	return false
}

func filterShowOnlyWhenConfigured(value interface{}) bool {
	return value != nil
}
