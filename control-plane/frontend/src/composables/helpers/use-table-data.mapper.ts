/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ETableFilters, type ITableColumn } from "@/models/table.model";
import type { IPolicy } from "@/stores/policy.store";
import type {
  ComputeAsset,
  CredentialsListResponseEntry,
  DatasourceListResponseEntry,
  EnvironmentAsset,
  WorkloadTemplate,
} from "@/swagger-models/assets-service-client";
import type { IDepartmentTable } from "@/models/department.model";
import { EDepartmentColumnName } from "@/models/department.model";
import { EProjectColumnName, type IProjectTable } from "@/models/project.model";
import { EIntervalLabels } from "@/models/interval.model";
import { departmentIndexColumns } from "@/table-models/department.table-model";
import { projectIndexColumns } from "@/table-models/project.table-model";
import { auditIndexColumns } from "@/table-models/audit.table-model";
import type { AuditLog } from "@/swagger-models/audit-service-client";
import type { IUser } from "@/models/user.model";
import { allUserColumns } from "@/table-models/user.table-model";
import { type ServiceAccount, UsersFilterSortFields } from "@/swagger-models/identity-manager-client";
import { type INodePoolTable, nodePoolIndexColumns } from "@/table-models/node-pool.table-model";
import { type DisplayedCluster, NodeFilterSortFields } from "@/swagger-models/cluster-service-client";
import { allClusterColumns } from "@/table-models/cluster.table-model";
import { ReportFilterAndSortFields } from "@/swagger-models/metrics-service-client";
import type { IReportTable } from "@/table-models/report.table-model";
import { reportIndexColumns } from "@/table-models/report.table-model";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { type AccessRule, AccessRulesSortFilterFields } from "@/swagger-models/authorization-client";
import { accessRulesIndexColumns } from "@/table-models/access-rule.table-model";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { NodepoolSortFilterFields } from "@/swagger-models/cluster-service-client";
import type { DatavolumeNoSharedScopes } from "@/swagger-models/datavolumes-service-client";
import { dataVolumeIndexColumns } from "@/table-models/data-volume.table-model";
import { nodeIndexColumns, type INodeTable } from "@/table-models/node.table-model";
import { dataVolumeService } from "@/services/control-plane/data-volume.service/data-volume.service";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { templateIndexColumns } from "@/table-models/templates.table-model";
import { aiApplicationIndexColumns } from "@/table-models/ai-application.table-model";
import type { IApplicationTable } from "@/models/ai-application.model";
import { ApplicationSortFilterFields } from "@/models/ai-application.model";
import { aiApplicationService } from "@/services/cluster/ai-application.service/ai-application.service";

interface IIndexTableInfo {
  storageFilterKey: ETableFilters;
  sortByDefault: string;
  /**
   * Text representing the entity type for error messages and notifications.
   */
  entityText: string;
  intervalLabel?: EIntervalLabels;
  intervalLabelPostCreation?: EIntervalLabels;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getEntityId: (e: any) => string;
  /**
   * Flag to indicate if the table is in tenant scope. if true the default filers exclude the cluster filter.
   */
  isTenantScope?: boolean;
}

interface IAdvancedIndexTableInfo extends Omit<IIndexTableInfo, "getEntityId"> {
  /**
   * Array of all entity columns.
   * Required to enable local search functionality.
   */
  allEntityColumns: Array<ITableColumn>;
  /**
   * Function to get the entity count.
   */
  getEntityCount?: (filterBy?: Array<string>, search?: string) => Promise<number>;
  /**
   * Array of valid filter fields.
   * Used in mapColumnsFilterToFilterParams
   */
  validFilterFields?: string[];
  /**
   * Flag to determine if the page use server side search.
   */
  isAdvancedSearch?: boolean;
  /**
   * Function to get the entity ID.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getEntityId: (e: any) => string;
}

export enum EIndexPages {
  COMPUTE_RESOURCE = "computeResource",
  DATA_SOURCE = "dataSource",
  ENVIRONMENT = "environment",
  CREDENTIAL = "credential",
  WORKLOAD_TEMPLATE = "workloadTemplate",
  POLICY = "policy",
  APPLICATION = "application",
}

export enum EAdvancedIndexPages {
  DEPARTMENT = "department",
  PROJECT = "project",
  AUDIT = "audit",
  USER = "user",
  NODE_POOL = "node-pool",
  NODE = "node",
  REPORT = "report",
  ACCESS_RULE = "access-rule",
  DATA_VOLUME = "data-volume",
  TEMPLATES = "templates",
  AI_APPLICATION = "ai-application",
  CLUSTER = "cluster",
}

export const useTableDataIndexMapper: Record<EIndexPages, IIndexTableInfo> = {
  [EIndexPages.COMPUTE_RESOURCE]: {
    storageFilterKey: ETableFilters.COMPUTE_RESOURCE,
    sortByDefault: "compute-resource",
    entityText: "compute resources",
    getEntityId: (e: ComputeAsset) => e.meta?.id || "",
  },
  [EIndexPages.DATA_SOURCE]: {
    storageFilterKey: ETableFilters.DATA_SOURCE,
    sortByDefault: "data-source",
    entityText: "data sources",
    getEntityId: (e: DatasourceListResponseEntry) => e.meta?.id || "",
    intervalLabel: EIntervalLabels.DataSourcesIndex,
    intervalLabelPostCreation: EIntervalLabels.IndexPostCreationRefreshRate,
  },
  [EIndexPages.ENVIRONMENT]: {
    storageFilterKey: ETableFilters.ENVIRONMENT,
    sortByDefault: "environment",
    entityText: "environments",
    getEntityId: (e: EnvironmentAsset) => e.meta?.id || "",
  },
  [EIndexPages.CREDENTIAL]: {
    storageFilterKey: ETableFilters.CREDENTIAL,
    sortByDefault: "credentials",
    entityText: "credentials",
    getEntityId: (e: CredentialsListResponseEntry) => e.meta?.id || "",
    intervalLabel: EIntervalLabels.CredentialsIndex,
    intervalLabelPostCreation: EIntervalLabels.IndexPostCreationRefreshRate,
  },
  [EIndexPages.WORKLOAD_TEMPLATE]: {
    storageFilterKey: ETableFilters.TEMPLATE,
    sortByDefault: "template",
    entityText: "templates",
    getEntityId: (e: WorkloadTemplate) => e.meta?.id || "",
  },
  [EIndexPages.POLICY]: {
    storageFilterKey: ETableFilters.POLICY,
    sortByDefault: "policy",
    entityText: "policies",
    getEntityId: (e: IPolicy) => e.policyId || "",
    intervalLabel: EIntervalLabels.PolicyIndex,
    intervalLabelPostCreation: EIntervalLabels.IndexPostCreationRefreshRate,
  },
  [EIndexPages.APPLICATION]: {
    storageFilterKey: ETableFilters.APPLICATION,
    sortByDefault: "name",
    entityText: "applications",
    getEntityId: (application: ServiceAccount) => application.id || "",
    isTenantScope: true,
  },
};

const advancedTableDataIndexConfig: Record<EAdvancedIndexPages, IAdvancedIndexTableInfo> = {
  [EAdvancedIndexPages.DEPARTMENT]: {
    storageFilterKey: ETableFilters.DEPARTMENT,
    sortByDefault: EDepartmentColumnName.DepartmentName,
    entityText: "departments",
    getEntityId: (department: IDepartmentTable) => department.id,
    allEntityColumns: departmentIndexColumns,
  },
  [EAdvancedIndexPages.PROJECT]: {
    storageFilterKey: ETableFilters.PROJECT,
    sortByDefault: EProjectColumnName.ProjectName,
    entityText: "projects",
    getEntityId: (e: IProjectTable) => e.id,
    intervalLabel: EIntervalLabels.ProjectsIndex,
    intervalLabelPostCreation: EIntervalLabels.IndexPostCreationRefreshRate,
    allEntityColumns: projectIndexColumns,
    isAdvancedSearch: true,
  },
  [EAdvancedIndexPages.AUDIT]: {
    storageFilterKey: ETableFilters.AUDIT,
    sortByDefault: "subject",
    entityText: "audit events",
    getEntityId: (e: AuditLog) => e.id || "",
    allEntityColumns: auditIndexColumns,
    isTenantScope: true,
  },
  [EAdvancedIndexPages.USER]: {
    storageFilterKey: ETableFilters.USERS,
    sortByDefault: UsersFilterSortFields.Username,
    entityText: "users",
    getEntityId: (e: IUser) => e.userId || "",
    allEntityColumns: allUserColumns,
    isAdvancedSearch: true,
    isTenantScope: true,
  },
  [EAdvancedIndexPages.NODE_POOL]: {
    storageFilterKey: ETableFilters.NODE_POOL_V2,
    sortByDefault: NodepoolSortFilterFields.Name,
    entityText: "node pools",
    getEntityId: (nodePool: INodePoolTable) => nodePool.id,
    allEntityColumns: nodePoolIndexColumns,
    intervalLabel: EIntervalLabels.NodePoolIndex,
  },
  [EAdvancedIndexPages.NODE]: {
    storageFilterKey: ETableFilters.NODE_V2,
    sortByDefault: NodeFilterSortFields.Name,
    entityText: "node",
    getEntityId: (node: INodeTable) => node.id || "",
    allEntityColumns: nodeIndexColumns,
    intervalLabel: EIntervalLabels.NodeIndex,
    isAdvancedSearch: true,
  },
  [EAdvancedIndexPages.REPORT]: {
    storageFilterKey: ETableFilters.REPORT,
    sortByDefault: ReportFilterAndSortFields.Name,
    entityText: "reports",
    getEntityId: (report: IReportTable) => report.id,
    allEntityColumns: reportIndexColumns,
    intervalLabel: EIntervalLabels.ReportIndex,
    getEntityCount: (filterBy?: Array<string>, search?: string) => orgUnitService.countReports(filterBy, search),
    validFilterFields: Object.values(ReportFilterAndSortFields),
    isAdvancedSearch: true,
    isTenantScope: true,
  },
  [EAdvancedIndexPages.ACCESS_RULE]: {
    storageFilterKey: ETableFilters.ACCESS_RULE,
    sortByDefault: AccessRulesSortFilterFields.SubjectId,
    entityText: "access rules",
    allEntityColumns: accessRulesIndexColumns,
    getEntityId: (accessRule: AccessRule) => accessRule.id.toString(),
    getEntityCount: (filterBy?: Array<string>, search?: string) =>
      accessRuleService.getAccessRulesCount(filterBy, search),
    validFilterFields: Object.values(AccessRulesSortFilterFields),
    isAdvancedSearch: true,
    isTenantScope: true,
    intervalLabel: EIntervalLabels.AccessRuleIndex,
  },
  [EAdvancedIndexPages.DATA_VOLUME]: {
    storageFilterKey: ETableFilters.DATA_VOLUME,
    sortByDefault: "name",
    entityText: "data volumes",
    getEntityId: (e: DatavolumeNoSharedScopes) => e.id || "",
    intervalLabel: EIntervalLabels.DataVolumeIndex,
    getEntityCount: (filterBy?: Array<string>, search?: string) => dataVolumeService.countDataVolumes(filterBy, search),
    intervalLabelPostCreation: EIntervalLabels.IndexPostCreationRefreshRate,
    allEntityColumns: dataVolumeIndexColumns,
    isAdvancedSearch: true,
  },
  [EAdvancedIndexPages.TEMPLATES]: {
    storageFilterKey: ETableFilters.TEMPLATE_V2,
    sortByDefault: "name",
    entityText: "templates",
    getEntityId: (e: WorkloadTemplate) => e.meta?.id || "",
    intervalLabel: EIntervalLabels.TemplatesIndex,
    getEntityCount: (filterBy?: Array<string>, search?: string) => {
      const distFilterStr = "workloadType!=Distributed";
      const trainingFilterStr = "workloadType!=Training";

      const distFilterExists = filterBy?.find((f) => f.includes(distFilterStr));
      const trainingFilterExists = filterBy?.find((f) => f.includes(trainingFilterStr));
      // Because Distributed is not treated as an actual type of workload in the UI but it does in the backend we need to consider it as training when it comes to filtering
      // if distFilter exists but training filter is not remove the distFilter
      if (distFilterExists && !trainingFilterExists) {
        filterBy = filterBy?.filter((f) => !f.includes(distFilterStr));
      }

      // we should not send includeAncestors flag if no clusterId filter is present
      const includeAncestors = !!filterBy?.find((f) => f.includes("clusterId=="));

      return workloadTemplateService.templatesCount(filterBy, search, includeAncestors);
    },
    allEntityColumns: templateIndexColumns,
    isAdvancedSearch: true,
  },
  [EAdvancedIndexPages.AI_APPLICATION]: {
    storageFilterKey: ETableFilters.AI_APPLICATION,
    sortByDefault: ApplicationSortFilterFields.Name,
    entityText: "AI applications",
    getEntityId: (e: IApplicationTable) => e.id || "",
    intervalLabel: EIntervalLabels.AIApplicationsIndex,
    getEntityCount: (filterBy?: Array<string>, search?: string) =>
      aiApplicationService.getApplicationsCount(filterBy, search),
    validFilterFields: Object.values(ApplicationSortFilterFields),
    allEntityColumns: aiApplicationIndexColumns,
    isAdvancedSearch: true,
  },
  [EAdvancedIndexPages.CLUSTER]: {
    storageFilterKey: ETableFilters.CLUSTER,
    sortByDefault: "name",
    entityText: "clusters",
    getEntityId: (cluster: DisplayedCluster) => cluster.uuid || "",
    allEntityColumns: allClusterColumns,
    intervalLabel: EIntervalLabels.ClusterIndex,
    isTenantScope: true,
  },
};

// Exposed as a Map so lookups use `.get()` instead of a computed object index (avoids object-injection sink).
// The config above keeps a typed Record so all EAdvancedIndexPages keys remain exhaustively enforced at compile time.
export const useAdvancedTableDataIndexMapper = new Map<EAdvancedIndexPages, IAdvancedIndexTableInfo>(
  Object.entries(advancedTableDataIndexConfig) as Array<[EAdvancedIndexPages, IAdvancedIndexTableInfo]>,
);
