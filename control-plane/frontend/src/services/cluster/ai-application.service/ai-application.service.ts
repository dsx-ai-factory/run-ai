/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { filterService } from "@/services/filter.service/filter.service";
import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import type { AIApplication, AIApplicationCreationRequest } from "@/swagger-models/workloads-client";
import type { AIApplicationVerbosity } from "@/swagger-models/workloads-client";
import type { IApplicationTable } from "@/models/ai-application.model";

export const aiApplicationService = {
  list,
  createApplication,
  uploadChart,
  deleteApplication,
  getApplicationById,
  getApplications,
  getApplicationsCount,
};

const validFilterFields = ["clusterId", "name", "projectId", "projectName", "source", "type"];

async function list(filterBy: IFilterBy = {}): Promise<IApplicationTable[]> {
  const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy, validFilterFields, true);

  const applications = await getApplications(
    filters.offset,
    filters.limit,
    filters.sortOrder,
    filters.sortBy || undefined,
    filters.filterBy,
    filters.search,
  );

  return applications.map((app) => ({
    ...app,
  })) as IApplicationTable[];
}

async function getApplications(
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc" | undefined,
  sortBy?: string,
  filterBy?: string[],
  search?: string,
): Promise<AIApplication[]> {
  try {
    const response = await workloadServiceApi.aiApplicationsApi.listApplications(
      "full" as AIApplicationVerbosity,
      offset,
      limit,
      sortOrder,
      sortBy as "clusterId" | "name" | "projectId" | "projectName" | "source" | "type" | undefined,
      filterBy,
      search,
    );
    return response.data.applications || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getApplicationsCount(filterBy?: Array<string>, search?: string): Promise<number> {
  try {
    const response = await workloadServiceApi.aiApplicationsApi.countApplications(filterBy, search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createApplication(request: AIApplicationCreationRequest): Promise<AIApplication> {
  try {
    const response = await workloadServiceApi.aiApplicationsApi.createApplication(request);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function uploadChart(file: File): Promise<string> {
  try {
    const response = await workloadServiceApi.aiApplicationsApi.uploadApplicationChart(file);
    return response.data.chartId;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteApplication(id: string): Promise<void> {
  try {
    await workloadServiceApi.aiApplicationsApi.deleteApplication(id);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getApplicationById(id: string): Promise<AIApplication> {
  try {
    const response = await workloadServiceApi.aiApplicationsApi.getApplication(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
