/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { SERVICE_ACCOUNTS_ROUTE_NAMES } from "@/router/service-accounts.routes/service-accounts.routes.names";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { POLICIES_ROUTE_NAMES } from "@/router/policy.routes/policy.routes.names";
import { USER_ROUTES_NAMES } from "@/router/users.routes/users.routes.names";
import { DASHBOARD_ROUTE_NAMES } from "@/router/dashboard.routes/dashboard.routes.names";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { DATA_SOURCE_ROUTE_NAMES } from "@/router/data-source.routes/data-source.routes.names";
import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { AUDIT_ROUTE_NAMES } from "@/router/audit.routes/audit.routes.names";
import { AI_APPLICATION_ROUTE_NAMES } from "@/router/ai-application.routes/ai-application.routes.names";

// models
import type { EBadgeLabel } from "@/models/setting.model";
import { SettingKeys } from "@/models/setting.model";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { ReportRoutesName } from "@/router/report.routes/report.routes.names";
import { type EFeatureToggle } from "@/services/infra/unleash.service/unleash.model";
import { EFeatureToggle as FeatureToggleEnum } from "@/services/infra/unleash.service/unleash.model";
import { DATA_VOLUME_ROUTE_NAMES } from "@/router/data-volume.routes/data-volume.routes.names";

// Single source of truth for nav icon glyphs (mirrors Figma "Master icons → Navigation icons").
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

// IMPORTANT!
// when you add a new route, make sure to add parentSubject prop to the route.meta

export interface ISubjectRoutes {
  name: string;
  label: string;
  aid: string;
  subtitle?: string;
  stageBadge?: EBadgeLabel;
  subItems?: ISubjectRoutes[];
  resourceType?: ResourceType | ResourceType[];
  minPermittedActions?: Action[];
  // Use this to check if user has the action on ANY of the resource types (OR logic)
  anyOfResourceTypes?: ResourceType[];
  anyOfResourceTypesAction?: Action;
  featureFlagName?: SettingKeys; // settings feature flag
  featureToggleName?: EFeatureToggle; // unleash feature toggle
  hiddenForBranded?: boolean;
}

export interface IMenuItem {
  label: string;
  icon: string;
  value: EMenuSubjectNames;
  aid: EMenuSubjectNames;
  items: ISubjectRoutes[];
  isActive?: boolean;
  subtitle?: string;
}

export enum EMenuSubjectNames {
  ANALYTICS = "analytics",
  WORKLOADS_MANAGER = "workloads-manager",
  RESOURCES = "resources",
  ORGANIZATION = "organization",
  ACCESS = "access",
  POLICIES = "policies",
  ADMIN = "admin",
}

export const SIDE_NAV_SUBJECT_TEXTS: Record<EMenuSubjectNames, string> = {
  [EMenuSubjectNames.ANALYTICS]: "Analytics",
  [EMenuSubjectNames.WORKLOADS_MANAGER]: "Workload manager",
  [EMenuSubjectNames.RESOURCES]: "Resources",
  [EMenuSubjectNames.ORGANIZATION]: "Organization",
  [EMenuSubjectNames.ACCESS]: "Access",
  [EMenuSubjectNames.POLICIES]: "Policies",
  [EMenuSubjectNames.ADMIN]: "Admin",
};

const analytics: IMenuItem = {
  label: SIDE_NAV_SUBJECT_TEXTS[EMenuSubjectNames.ANALYTICS],
  icon: toQuasarIconName(runaiIcons.nav.analytics),
  value: EMenuSubjectNames.ANALYTICS,
  aid: EMenuSubjectNames.ANALYTICS,
  items: [
    {
      name: DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX,
      label: "Dashboard",
      aid: "menu-item-dashboard",
      anyOfResourceTypes: [
        ResourceType.Cluster,
        ResourceType.Nodes,
        ResourceType.Department,
        ResourceType.Project,
        ResourceType.Workloads,
        ResourceType.Nodepools,
      ],
      anyOfResourceTypesAction: Action.Read,
      subItems: [
        {
          name: DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX,
          label: "Overview",
          aid: "menu-item-overview",
          anyOfResourceTypes: [
            ResourceType.Cluster,
            ResourceType.Nodes,
            ResourceType.Department,
            ResourceType.Project,
            ResourceType.Workloads,
            ResourceType.Nodepools,
          ],
          anyOfResourceTypesAction: Action.Read,
          hiddenForBranded: false,
        },
        {
          name: DASHBOARD_ROUTE_NAMES.RESOURCES_INDEX,
          label: "Quota management",
          aid: "menu-item-resources",
          resourceType: ResourceType.Cluster,
          minPermittedActions: [Action.Read],
          hiddenForBranded: false,
        },
      ],
    },
    {
      name: ReportRoutesName.ReportIndex,
      label: "Reports",
      aid: "menu-item-report",
      resourceType: ResourceType.Project,
      minPermittedActions: [Action.Read],
      hiddenForBranded: false,
    },
  ],
};

const workloadManager = {
  label: SIDE_NAV_SUBJECT_TEXTS[EMenuSubjectNames.WORKLOADS_MANAGER],
  // Figma Foundations → Navigation row uses entity/workload (custom runai-workload glyph)
  // for the Workloads-manager slot, not a separate nav-specific icon.
  icon: toQuasarIconName(runaiIcons.entity.workload),
  value: EMenuSubjectNames.WORKLOADS_MANAGER,
  aid: EMenuSubjectNames.WORKLOADS_MANAGER,
  items: [
    {
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      label: "Workloads",
      aid: "menu-item-workloads",
      resourceType: ResourceType.Workloads,
      minPermittedActions: [Action.Read],
    },
    {
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_DELETED_INDEX,
      label: "Deleted workloads",
      aid: "menu-item-deleted-workloads",
      resourceType: ResourceType.Workloads,
      minPermittedActions: [Action.Read],
      featureFlagName: SettingKeys.DeletedWorkloadsView,
    },
    {
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
      label: "Templates",
      aid: "menu-item-templates",
      resourceType: ResourceType.Templates,
      minPermittedActions: [Action.Read],
    },
    {
      name: "assets",
      label: "Assets",
      aid: "menu-item-assets",
      subItems: [
        {
          name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX,
          label: "Environments",
          aid: "menu-item-environments",
          resourceType: ResourceType.Environments,
          minPermittedActions: [Action.Read],
        },
        {
          name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX,
          label: "Compute resources",
          aid: "menu-item-compute-resource",
          resourceType: ResourceType.ComputeResources,
          minPermittedActions: [Action.Read],
        },
        {
          name: "data-and-storage",
          label: "Data & storage",
          aid: "menu-item-data-and-storage",
          featureFlagName: SettingKeys.EnableDataVolumes,
          subItems: [
            {
              name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX,
              label: "Data sources",
              aid: "menu-item-data-sources",
              featureFlagName: SettingKeys.EnableDataVolumes,
              resourceType: [
                ResourceType.NfsAssets,
                ResourceType.S3Assets,
                ResourceType.GitAssets,
                ResourceType.PvcAssets,
                ResourceType.HostPathAssets,
              ],
              minPermittedActions: [Action.Read],
            },
            {
              name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX,
              featureFlagName: SettingKeys.EnableDataVolumes,
              label: "Data volumes",
              aid: "menu-item-data-volumes",
              resourceType: [ResourceType.Datavolumes],
              minPermittedActions: [Action.Read],
            },
          ],
        },
        {
          name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX,
          label: "Data sources",
          aid: "menu-item-data-sources",
          resourceType: [
            ResourceType.NfsAssets,
            ResourceType.S3Assets,
            ResourceType.GitAssets,
            ResourceType.PvcAssets,
            ResourceType.HostPathAssets,
          ],
          minPermittedActions: [Action.Read],
        },
        {
          name: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX,
          label: "Credentials",
          aid: "menu-item-credentials",
          resourceType: ResourceType.Credentials,
          minPermittedActions: [Action.Read],
        },
      ],
    },
    {
      name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX,
      label: "AI applications",
      aid: "menu-item-ai-applications",
      resourceType: ResourceType.Workloads,
      minPermittedActions: [Action.Read],
      featureToggleName: FeatureToggleEnum.ShowAIApplications,
    },
  ],
};

const resources = {
  label: SIDE_NAV_SUBJECT_TEXTS[EMenuSubjectNames.RESOURCES],
  // Figma Foundations → Navigation row uses entity/resource for the Resources slot.
  icon: toQuasarIconName(runaiIcons.entity.resource),
  value: EMenuSubjectNames.RESOURCES,
  aid: EMenuSubjectNames.RESOURCES,
  items: [
    {
      name: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
      label: "Clusters",
      aid: "menu-item-clusters",
      resourceType: ResourceType.Cluster,
      minPermittedActions: [Action.Create],
    },
    {
      name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
      label: "Node pools",
      aid: "menu-item-node-pools",
      resourceType: ResourceType.Nodes,
      minPermittedActions: [Action.Read],
    },
    {
      name: NODE_ROUTE_NAMES.NODE_INDEX,
      label: "Nodes",
      aid: "menu-item-nodes",
      resourceType: ResourceType.Nodes,
      minPermittedActions: [Action.Read],
    },
  ],
};

const organization = {
  label: SIDE_NAV_SUBJECT_TEXTS[EMenuSubjectNames.ORGANIZATION],
  icon: toQuasarIconName(runaiIcons.nav.organization),
  value: EMenuSubjectNames.ORGANIZATION,
  aid: EMenuSubjectNames.ORGANIZATION,
  items: [
    {
      name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
      label: "Departments",
      aid: "menu-item-departments",
      resourceType: ResourceType.Department,
      minPermittedActions: [Action.Read],
    },
    {
      name: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
      label: "Projects",
      aid: "menu-item-projects",
      resourceType: ResourceType.Project,
      minPermittedActions: [Action.Read],
    },
  ],
};

const access = {
  label: SIDE_NAV_SUBJECT_TEXTS[EMenuSubjectNames.ACCESS],
  // Custom monochrome glyph (Figma "Custom icons for Dima" → nav/access). Renders via the
  // runai-glyph SVG sprite, not Phosphor — there's no Phosphor equivalent for this mark.
  icon: toQuasarIconName(runaiIcons.nav.access),
  value: EMenuSubjectNames.ACCESS,
  aid: EMenuSubjectNames.ACCESS,
  items: [
    {
      name: USER_ROUTES_NAMES.USER_INDEX,
      label: "Users",
      aid: "menu-item-users",
      resourceType: ResourceType.Users,
      minPermittedActions: [Action.Read],
    },
    {
      name: SERVICE_ACCOUNTS_ROUTE_NAMES.SERVICE_ACCOUNTS_INDEX,
      label: "Service accounts",
      subtitle: "(formerly applications)",
      aid: "menu-item-service-accounts",
      resourceType: [ResourceType.ServiceAccount, ResourceType.Apps],
      minPermittedActions: [Action.Read],
    },
    {
      name: ROLE_ROUTE_NAMES.ROLE_INDEX,
      label: "Roles",
      aid: "menu-item-roles",
      resourceType: ResourceType.Roles,
      minPermittedActions: [Action.Read],
    },
    {
      name: ROLE_ROUTE_NAMES.ACCESS_RULES_INDEX,
      label: "Access rules",
      aid: "menu-item-access-rules",
      resourceType: ResourceType.AccessRules,
      minPermittedActions: [Action.Read],
    },
  ],
};

const policies = {
  label: SIDE_NAV_SUBJECT_TEXTS[EMenuSubjectNames.POLICIES],
  // Navigation subject icons are bold in Figma. Keep entity.policy regular for policy empty/YAML states.
  icon: toQuasarIconName(runaiIcons.nav.policy),
  value: EMenuSubjectNames.POLICIES,
  aid: EMenuSubjectNames.POLICIES,
  items: [
    {
      name: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
      label: "Workload policies",
      aid: "menu-item-policies",
      resourceType: ResourceType.Policies,
      minPermittedActions: [Action.Read],
      featureFlagName: SettingKeys.EnablePolicies,
    },
  ],
};

const admin = {
  label: SIDE_NAV_SUBJECT_TEXTS[EMenuSubjectNames.ADMIN],
  icon: toQuasarIconName(runaiIcons.nav.admin),
  value: EMenuSubjectNames.ADMIN,
  aid: EMenuSubjectNames.ADMIN,
  items: [
    {
      name: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
      label: "General settings",
      aid: "menu-item-settings",
      resourceType: [
        ResourceType.Settings,
        ResourceType.SecuritySettings,
        ResourceType.BrandingSettings,
        ResourceType.NotificationSettings,
        ResourceType.Registries,
      ],
      minPermittedActions: [Action.Create, Action.Update, Action.Delete],
    },
    {
      name: AUDIT_ROUTE_NAMES.AUDIT_INDEX,
      label: "Event history",
      aid: "menu-item-event-history",
      resourceType: ResourceType.EventsHistory,
      minPermittedActions: [Action.Read],
    },
  ],
};

export const menuSubjects: Record<string, IMenuItem> = {
  analytics,
  workloadManager,
  resources,
  organization,
  access,
  policies,
  admin,
};
