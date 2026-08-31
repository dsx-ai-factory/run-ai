/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ETableExportCsvFilesNames } from "@/models/table.model";
import { DepartmentFilterSortFields, ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { CLUSTER_COLUMN_FILTER_NAME } from "@/models/filter.model";
import { EOrgUnitOverTimeColumnName } from "@/models/org-unit.model";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { AccessRulesSortFilterFields } from "@/swagger-models/authorization-client";
import { UsersFilterSortFields } from "@/swagger-models/identity-manager-client";
import { ApplicationSortFilterFields } from "@/models/ai-application.model";

export const API_LIMIT_MAX_RECORDS = 500;
export const MAX_RECORDS_TO_EXPORT = 20000;

export enum EAdvancedExportFileType {
  Csv = "csv",
}

export enum EAdvancedExportEntity {
  Projects = "projects",
  Departments = "departments",
  Workloads = "workloads",
  AccessRules = "accessRules",
  Users = "users",
  AIApplications = "aiApplications",
}

export type GetRecordsFunction<T> = (
  filterBy?: Array<string>,
  sortBy?: string,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
) => Promise<T[]>;

export interface IExportFileConfig {
  fileName: ETableExportCsvFilesNames;
  fileType: EAdvancedExportFileType;
  getRecordsValidFields?: string[];
  isAdvancedSearch?: boolean;
}

export const useFileExportConfigMapper: Record<EAdvancedExportEntity, IExportFileConfig> = {
  [EAdvancedExportEntity.Projects]: {
    fileName: ETableExportCsvFilesNames.Project,
    fileType: EAdvancedExportFileType.Csv,
    isAdvancedSearch: true,
    getRecordsValidFields: [
      ...Object.values(ProjectFilterSortFields),
      CLUSTER_COLUMN_FILTER_NAME,
      EOrgUnitOverTimeColumnName.AverageGpuAllocation,
      EOrgUnitOverTimeColumnName.AverageGpuUtilization,
    ],
  },
  [EAdvancedExportEntity.Departments]: {
    fileName: ETableExportCsvFilesNames.Department,
    fileType: EAdvancedExportFileType.Csv,
    getRecordsValidFields: [
      ...Object.values(DepartmentFilterSortFields),
      CLUSTER_COLUMN_FILTER_NAME,
      EOrgUnitOverTimeColumnName.AverageGpuAllocation,
      EOrgUnitOverTimeColumnName.AverageGpuUtilization,
    ],
  },
  [EAdvancedExportEntity.Workloads]: {
    fileName: ETableExportCsvFilesNames.Workload,
    fileType: EAdvancedExportFileType.Csv,
    isAdvancedSearch: true,
    getRecordsValidFields: [...Object.values(WorkloadSortFilterFields)],
  },
  [EAdvancedExportEntity.AccessRules]: {
    fileName: ETableExportCsvFilesNames.AccessRule,
    fileType: EAdvancedExportFileType.Csv,
    isAdvancedSearch: true,
    getRecordsValidFields: [...Object.values(AccessRulesSortFilterFields)],
  },
  [EAdvancedExportEntity.Users]: {
    fileName: ETableExportCsvFilesNames.User,
    fileType: EAdvancedExportFileType.Csv,
    isAdvancedSearch: true,
    getRecordsValidFields: [...Object.values(UsersFilterSortFields)],
  },
  [EAdvancedExportEntity.AIApplications]: {
    fileName: ETableExportCsvFilesNames.AIApplication,
    fileType: EAdvancedExportFileType.Csv,
    isAdvancedSearch: true,
    getRecordsValidFields: [
      ApplicationSortFilterFields.ClusterId,
      ApplicationSortFilterFields.Name,
      ApplicationSortFilterFields.ProjectId,
      ApplicationSortFilterFields.ProjectName,
      ApplicationSortFilterFields.Type,
      CLUSTER_COLUMN_FILTER_NAME,
    ],
  },
};
