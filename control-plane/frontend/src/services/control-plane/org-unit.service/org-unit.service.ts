/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IDepartmentTable, IDepartmentTableFilterBy } from "@/models/department.model";
import { DEFAULT_DEPARTMENT_NAME } from "@/models/department.model";
import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import type { IProjectTable, IProjectTableFilterBy } from "@/models/project.model";
import { fileService, FileType } from "@/services/file.service/file.service";
import { filterService } from "@/services/filter.service/filter.service";
import { orgUnitServiceApi } from "@/services/infra/client-apis/org-unit-service-api/org-unit-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  DepartmentFilterSortFields,
  ProjectFilterSortFields,
  ReportFilterAndSortFields,
  Verbosity,
  type DeletionTypes,
  type Department,
  type DepartmentCreationRequest,
  type DepartmentUpdateRequest,
  type OrgUnitTelemetryType,
  type Project,
  type ProjectCreationRequest,
  type ProjectMetadata,
  type ProjectUpdateRequest,
  type Report,
  type ReportRequestFields,
  type BatchDeleteReportsResponse,
  type RankingMetricType,
  type RankingOrderType,
  type ProjectGetRanking,
  type DepartmentGetRanking,
  type NodeType,
  type GetNodeTypes200ResponseNodeTypesInner,
  type BatchPatchNodePoolResourcesRequest,
  type ProjectNodePoolResourcesFilterSortFields,
  type DepartmentNodePoolResourcesFilterSortFields,
  type OrgUnitMetricType,
  type MetricsResponse,
  type TelemetryResponse,
} from "@/swagger-models/org-unit-service-client";
import type { IReportTable } from "@/table-models/report.table-model";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { orgTreeService } from "../rbac/org-tree.service/org-tree.service";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";

export const orgUnitService = {
  getProjects,
  getProjectsMetadata,
  listProjectsMetadata,
  getProject,
  getDepartments,
  deleteProject,
  countDepartments,
  listDepartments,
  getDepartment,
  deleteDepartment,
  getDepartmentsTelemetry,
  getProjectsTelemetry,
  listProjects,
  countProjects,
  createProject,
  updateProject,
  createDepartment,
  updateDepartment,
  isDepartmentNameUnique,
  isProjectNameUnique,
  getDefaultDepartment,
  countReports,
  getReportById,
  getReports,
  listReports,
  isReportApiAvailable,
  isReportNameUnique,
  createReport,
  deleteReport,
  batchDeleteReports,
  downloadReport,
  getProjectsRanking,
  getDepartmentsRanking,
  getNodeTypes,
  createNodeType,
  batchPatchProjectNodePoolResources,
  sumDepartmentNodePoolResources,
  getProjectNodePoolResources,
  sumProjectNodePoolResources,
  countProjectsNodePoolResources,
  batchPatchDepartmentNodePoolResources,
  getDepartmentNodePoolResources,
  countDepartmentNodePoolResources,
  getDepartmentMetrics,
  getProjectMetrics,
};

//******** ProjectsApi ************
/**
 * Lists projects based on the provided filter criteria, filters the results by search term, and returns the filtered results.
 *
 * @param {IFilterBy} [filterBy={}] - The filter criteria for listing projects.
 * @returns {Promise<IProjectTable[]>} - A promise that resolves to an array of projects.
 */
async function listProjects(filterBy: IProjectTableFilterBy): Promise<IProjectTable[]> {
  const filters: IPaginationFilter = orgUnitUtil.getProjectListFilters(filterBy);

  const projects = await getProjects(
    filters.filterBy || [],
    filters.sortBy as ProjectFilterSortFields | undefined,
    filters.sortOrder,
    filters?.offset,
    filters?.limit,
    filters?.search,
  );
  return projects.map((project: Project) => {
    return {
      ...project,
      rolesNames: [],
    };
  });
}

async function getProjects(
  filterBy?: Array<string>,
  sortBy?: string,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
): Promise<Project[]> {
  try {
    const response = await orgUnitServiceApi.projectsApi.getProjects(
      filterBy,
      sortBy as ProjectFilterSortFields | undefined,
      sortOrder,
      offset,
      limit,
      search,
    );
    return response.data.projects;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getProjectsMetadata(
  clusterUuid: string,
  permittedScopes?: PermittedScopes,
  orgUnits?: Array<OrgUnit>,
): Promise<Array<ProjectMetadata>> {
  function delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  try {
    const projects: Array<ProjectMetadata> = [];
    let page = 0;
    const pageSize = 500;
    const delayTime = 300; // delay time in milliseconds

    let next = true;
    while (next) {
      const response = await orgUnitServiceApi.projectsApi.getProjectsMetadata(
        [`clusterId==${clusterUuid}`],
        undefined,
        undefined,
        page,
        pageSize,
      );

      projects.push(...response.data.projects);

      if (!response.data.next) {
        next = false;
        break;
      }

      page += pageSize;
      await delay(delayTime); // Delay before the next request
    }

    if (permittedScopes && orgUnits) {
      const allowedProjectsSet: Set<string> = new Set(
        orgTreeService.getExpandedPermittedScopes(permittedScopes, orgUnits).projects,
      );
      return projects.filter((project) => allowedProjectsSet.has(project.id));
    }

    return projects;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listProjectsMetadata(
  filterBy?: Array<string>,
  sortBy?: "name" | "clusterId" | "parentId" | "phase" | "createdAt",
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
): Promise<Array<ProjectMetadata>> {
  try {
    const response = await orgUnitServiceApi.projectsApi.getProjectsMetadata(
      filterBy,
      sortBy,
      sortOrder,
      offset,
      limit,
      search,
    );
    return (response.data.projects ?? []) as Array<ProjectMetadata>;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countProjects(filters: IPaginationFilter): Promise<number> {
  try {
    const response = await orgUnitServiceApi.projectsApi.countProjects(filters.filterBy, filters.search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getProject(projectId: string): Promise<Project> {
  try {
    const response = await orgUnitServiceApi.projectsApi.getProject(projectId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteProject(projectId: string, deleteType?: DeletionTypes, force?: boolean): Promise<void> {
  try {
    await orgUnitServiceApi.projectsApi.deleteProject(projectId, deleteType, force);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createProject(projectCreationRequest: ProjectCreationRequest): Promise<Project> {
  try {
    const response = await orgUnitServiceApi.projectsApi.createProject(projectCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateProject(projectId: string, projectUpdateRequest: ProjectUpdateRequest): Promise<Project> {
  try {
    const response = await orgUnitServiceApi.projectsApi.updateProject(projectId, projectUpdateRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function isProjectNameUnique(clusterId: string, name: string): Promise<boolean> {
  try {
    const filterBy = [
      filterUtil.getEqualsFilterString(ProjectFilterSortFields.ClusterId, clusterId),
      filterUtil.getEqualsFilterString(ProjectFilterSortFields.Name, name),
    ];
    const count: number = await orgUnitService.countProjects({ filterBy });
    return count <= 0;
  } catch (error: unknown) {
    throw httpService.handleHttpError(error);
  }
}

async function getProjectsRanking(
  metricType: RankingMetricType,
  order: RankingOrderType,
  clusterId: string,
  nodepoolName?: string,
  departmentId?: string,
): Promise<ProjectGetRanking[]> {
  try {
    const response = await orgUnitServiceApi.projectsApi.getProjectsRanking(
      metricType,
      order,
      clusterId,
      nodepoolName,
      departmentId,
    );
    return response.data.projects;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

//******** DepartmentsApi ************

/**
 * Lists departments based on the provided filter criteria, filters the results by search term, and returns the filtered results.
 *
 * @param {IFilterBy} [filterBy={}] - The filter criteria for listing departments.
 * @returns {Promise<IDepartmentTable[]>} - A promise that resolves to an array of departments.
 */
async function listDepartments(filterBy: IDepartmentTableFilterBy): Promise<IDepartmentTable[]> {
  const filters: IPaginationFilter = orgUnitUtil.getDepartmentListFilters(filterBy);
  const departments = await getDepartments(
    filters.filterBy || [],
    filters.sortBy as DepartmentFilterSortFields | undefined,
    Verbosity.Verbose,
    filters.sortOrder,
    filters?.offset,
    filters?.limit,
  );

  return departments.map((department: Department) => {
    return {
      ...department,
      rolesNames: [],
    };
  });
}

async function getDepartments(
  filterBy?: Array<string>,
  sortBy?: string,
  verbosity?: Verbosity,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
): Promise<Department[]> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.getDepartments(
      filterBy,
      sortBy as DepartmentFilterSortFields | undefined,
      verbosity,
      sortOrder,
      offset,
      limit,
    );
    return response.data.departments as Department[];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countDepartments(filterBy?: Array<string>): Promise<number> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.countDepartments(filterBy);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDepartment(departmentId: string): Promise<Department> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.getDepartment(departmentId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteDepartment(departmentId: string): Promise<void> {
  try {
    await orgUnitServiceApi.departmentsApi.deleteDepartment(departmentId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createDepartment(department: DepartmentCreationRequest): Promise<Department> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.createDepartment(department);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateDepartment(departmentId: string, department: DepartmentUpdateRequest): Promise<Department> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.updateDepartment(departmentId, department);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDepartmentsTelemetry(
  telemetryType: OrgUnitTelemetryType,
  clusterId?: string,
  nodepoolName?: string,
  departmentId?: string,
  groupBy?: Array<"CLUSTER_ID">,
): Promise<void> {
  try {
    await orgUnitServiceApi.departmentsApi.getDepartmentsTelemetry(
      telemetryType,
      clusterId,
      nodepoolName,
      departmentId,
      groupBy,
    );
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getProjectsTelemetry(
  telemetryType: OrgUnitTelemetryType,
  clusterId?: string,
  nodepoolId?: string,
  departmentId?: string,
  projectId?: string,
  groupBy?: Array<"ClusterId" | "NodepoolId" | "ParentId" | "ProjectId">,
): Promise<TelemetryResponse> {
  try {
    const response = await orgUnitServiceApi.projectsApi.getProjectsTelemetry(
      telemetryType,
      clusterId,
      nodepoolId,
      departmentId,
      projectId,
      groupBy,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function isDepartmentNameUnique(clusterId: string, name: string): Promise<boolean> {
  try {
    const filterBy = [
      filterUtil.getEqualsFilterString(DepartmentFilterSortFields.ClusterId, clusterId),
      filterUtil.getEqualsFilterString(DepartmentFilterSortFields.Name, name),
    ];
    const count: number = await orgUnitService.countDepartments(filterBy);
    return count <= 0;
  } catch (error: unknown) {
    throw httpService.handleHttpError(error);
  }
}

async function getDefaultDepartment(clusterId: string): Promise<Department | null> {
  try {
    const filterBy = [
      filterUtil.getEqualsFilterString(DepartmentFilterSortFields.ClusterId, clusterId),
      filterUtil.getEqualsFilterString(DepartmentFilterSortFields.Name, DEFAULT_DEPARTMENT_NAME),
    ];
    const response = await orgUnitService.getDepartments(filterBy);
    if (response.length > 0) {
      return response[0];
    } else {
      return null;
    }
  } catch (error: unknown) {
    throw httpService.handleHttpError(error);
  }
}

async function getDepartmentsRanking(
  metricType: RankingMetricType,
  order: RankingOrderType,
  clusterId: string,
  nodepoolName?: string,
): Promise<DepartmentGetRanking[]> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.getDepartmentsRanking(
      metricType,
      order,
      clusterId,
      nodepoolName,
    );
    return response.data.departments;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

//******** ReportsApi ************

/**
 * Checks if the report API is available - usually available only in SaaS.
 */
async function isReportApiAvailable(): Promise<boolean> {
  try {
    const response = await orgUnitServiceApi.reportsApi.areReportsAvailable();
    return response.data.isAvailable;
  } catch (err: unknown) {
    return false;
  }
}

async function listReports(filters: IFilterBy): Promise<IReportTable[]> {
  const parsedFilters = filterService.mapColumnsFilterToFilterParams(
    filters,
    Object.values(ReportFilterAndSortFields),
    true,
  );

  const { filterBy, sortBy, sortOrder, offset, limit, search } = parsedFilters;
  const reports = await getReports(
    filterBy,
    sortBy as ReportFilterAndSortFields | undefined,
    sortOrder,
    offset,
    limit,
    search,
  );
  return reports as IReportTable[];
}

async function countReports(filterBy?: Array<string>, search?: string): Promise<number> {
  try {
    const response = await orgUnitServiceApi.reportsApi.countReports(filterBy, search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getReports(
  filterBy?: Array<string>,
  sortBy?: ReportFilterAndSortFields,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
): Promise<Report[]> {
  try {
    const response = await orgUnitServiceApi.reportsApi.listReports(filterBy, sortBy, sortOrder, offset, limit, search);
    return response.data.reports;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getReportById(reportId: string): Promise<Report> {
  try {
    const response = await orgUnitServiceApi.reportsApi.getReportById(reportId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function isReportNameUnique(name: string): Promise<boolean> {
  try {
    const filterBy = [filterUtil.getEqualsFilterString(ReportFilterAndSortFields.Name, name)];
    const count: number = await orgUnitService.countReports(filterBy);
    return count <= 0;
  } catch (error: unknown) {
    throw httpService.handleHttpError(error);
  }
}

async function createReport(reportRequestFields: ReportRequestFields): Promise<Report> {
  try {
    const response = await orgUnitServiceApi.reportsApi.createReport(reportRequestFields);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteReport(reportId: string): Promise<void> {
  try {
    await orgUnitServiceApi.reportsApi.deleteReportById(reportId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function batchDeleteReports(reportIds: string[]): Promise<BatchDeleteReportsResponse> {
  try {
    // The generated client types `ids` as Set<string> because the spec marks it
    // uniqueItems, but a Set serializes to {} - the wire format is a JSON array.
    // Deduping preserves that intent; the backend rejects duplicates with a 400.
    const ids = [...new Set(reportIds)] as unknown as Set<string>;
    const response = await orgUnitServiceApi.reportsApi.batchDeleteReports({ ids });
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function downloadReport(reportId: string): Promise<void> {
  try {
    await fileService.downloadFile(`api/v1/org-unit/reports/${reportId}/file`, FileType.csv);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNodeTypes(clusterId: string): Promise<Array<GetNodeTypes200ResponseNodeTypesInner>> {
  try {
    const response = await orgUnitServiceApi.nodeTypesApi.getNodeTypes(clusterId);

    return (response.data as GetNodeTypes200ResponseNodeTypesInner[]) || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createNodeType(name: string, clusterId: string): Promise<NodeType> {
  try {
    const response = await orgUnitServiceApi.nodeTypesApi.createNodeType({ name, "cluster-id": clusterId });
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * ============================================
 *  Project and Department Node Pool Resources APIs - mainly for org unit forms v2
 * ============================================
 */

async function batchPatchDepartmentNodePoolResources(
  batchPatchDepartmentResourcesRequest: BatchPatchNodePoolResourcesRequest,
) {
  try {
    const response = await orgUnitServiceApi.departmentsApi.batchPatchDepartmentNodePoolResources(
      batchPatchDepartmentResourcesRequest,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function batchPatchProjectNodePoolResources(
  batchPatchProjectResourcesRequest: BatchPatchNodePoolResourcesRequest,
) {
  try {
    const response = await orgUnitServiceApi.projectsApi.batchPatchProjectNodePoolResources(
      batchPatchProjectResourcesRequest,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function sumDepartmentNodePoolResources(nodepoolId: string, filterBy?: Array<string>) {
  try {
    const response = await orgUnitServiceApi.departmentsApi.sumDepartmentNodePoolResources(nodepoolId, filterBy);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getProjectNodePoolResources(
  filterBy?: Array<string>,
  sortBy?: ProjectNodePoolResourcesFilterSortFields,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
) {
  try {
    const response = await orgUnitServiceApi.projectsApi.getProjectNodePoolResources(
      filterBy,
      sortBy,
      sortOrder,
      offset,
      limit,
      search,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDepartmentNodePoolResources(
  filterBy?: Array<string>,
  sortBy?: DepartmentNodePoolResourcesFilterSortFields,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
) {
  try {
    const response = await orgUnitServiceApi.departmentsApi.getDepartmentNodePoolResources(
      filterBy,
      sortBy,
      sortOrder,
      offset,
      limit,
      search,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function sumProjectNodePoolResources(nodepoolId: string, filterBy?: Array<string>) {
  try {
    const response = await orgUnitServiceApi.projectsApi.sumProjectNodePoolResources(nodepoolId, filterBy);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countProjectsNodePoolResources(filterBy?: Array<string>, search?: string): Promise<number> {
  try {
    const response = await orgUnitServiceApi.projectsApi.countProjectNodePoolResources(filterBy, search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countDepartmentNodePoolResources(filterBy?: Array<string>, search?: string): Promise<number> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.countDepartmentNodePoolResources(filterBy, search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

//******** Metrics APIs ************

async function getDepartmentMetrics(
  departmentId: string,
  metricType: Array<OrgUnitMetricType>,
  start: string,
  end: string,
  numberOfSamples?: number,
  nodepoolName?: string,
  groupBy?: "Category",
): Promise<MetricsResponse> {
  try {
    const response = await orgUnitServiceApi.departmentsApi.getDepartmentMetrics(
      departmentId,
      metricType,
      start,
      end,
      numberOfSamples,
      nodepoolName,
      groupBy,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getProjectMetrics(
  projectId: string,
  metricType: Array<OrgUnitMetricType>,
  start: string,
  end: string,
  numberOfSamples?: number,
  nodepoolName?: string,
  groupBy?: "Category",
): Promise<MetricsResponse> {
  try {
    const response = await orgUnitServiceApi.projectsApi.getProjectMetrics(
      projectId,
      metricType,
      start,
      end,
      numberOfSamples,
      nodepoolName,
      groupBy,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
