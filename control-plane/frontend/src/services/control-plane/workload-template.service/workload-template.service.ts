/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// services
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { httpService } from "@/services/infra/https.service/http.service";

// models
import { WorkloadType } from "@/swagger-models/workloads-client";
import { EFilterOperator } from "@/models/filter.model";
import type {
  CountResponse,
  DistributedTemplate,
  DistributedTemplateCreationRequest,
  DistributedTemplatesListResponse,
  DistributedTemplateUpdateRequest,
  SupersetTemplate,
  TrainingTemplate,
  TrainingTemplateCreationRequest,
  TrainingTemplatesListResponse,
  TrainingTemplateUpdateRequest,
  WorkspaceTemplate,
  WorkspaceTemplateCreationRequest,
  WorkspaceTemplatesListResponse,
  WorkspaceTemplateUpdateRequest,
  InferenceTemplate,
  InferenceTemplateCreationRequest,
  InferenceTemplateUpdateRequest,
  InferenceTemplatesListResponse,
} from "@/swagger-models/workloads-client";
import { HttpErrorResponse } from "@/models/http-response.model";

export const workloadTemplateService = {
  createWorkspaceTemplateFlat,
  updateWorkspaceTemplateFlat,
  createTrainingTemplateFlat,
  updateTrainingTemplateFlat,
  createDistributedTemplateFlat,
  updateDistributedTemplateFlat,
  createInferenceTemplateFlat,
  updateInferenceTemplateFlat,
  getWorkspaceTemplateById,
  getTrainingTemplateById,
  getDistributedTemplateById,
  getInferenceTemplateById,
  getTemplateNameCount,
  templatesCount,
  templatesList,
  getFlatTemplateById,
  deleteFlatTemplate,
  workspaceTemplatesList,
  trainingTemplatesList,
  distributedTemplatesList,
  inferenceTemplatesList,
  templateNameAvailability,
};

// api calls

async function getWorkspaceTemplateById(id: string): Promise<WorkspaceTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.getWorkspaceTemplate(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getTrainingTemplateById(id: string): Promise<TrainingTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.getTrainingTemplate(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDistributedTemplateById(id: string): Promise<DistributedTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.getDistributedTemplate(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getInferenceTemplateById(id: string): Promise<InferenceTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.getInferenceTemplate(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createWorkspaceTemplateFlat(template: WorkspaceTemplateCreationRequest): Promise<WorkspaceTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.createWorkspaceTemplate(template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateWorkspaceTemplateFlat(
  id: string,
  template: WorkspaceTemplateUpdateRequest,
): Promise<WorkspaceTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.updateWorkspaceTemplate(id, template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateTrainingTemplateFlat(
  id: string,
  template: TrainingTemplateUpdateRequest,
): Promise<TrainingTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.updateTrainingTemplate(id, template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateDistributedTemplateFlat(
  id: string,
  template: DistributedTemplateUpdateRequest,
): Promise<DistributedTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.updateDistributedTemplate(id, template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateInferenceTemplateFlat(
  id: string,
  template: InferenceTemplateUpdateRequest,
): Promise<InferenceTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.updateInferenceTemplate(id, template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createTrainingTemplateFlat(template: TrainingTemplateCreationRequest): Promise<TrainingTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.createTrainingTemplate(template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createDistributedTemplateFlat(
  template: DistributedTemplateCreationRequest,
): Promise<DistributedTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.createDistributedTemplate(template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createInferenceTemplateFlat(template: InferenceTemplateCreationRequest): Promise<InferenceTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.createInferenceTemplate(template);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getTemplateNameCount(templateName: string, operator = EFilterOperator.Equals): Promise<CountResponse> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.countTemplates([`name${operator}${templateName}`]);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function templatesCount(filterBy?: Array<string>, search?: string, includeAncestors = false): Promise<number> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.countTemplates(filterBy, search, includeAncestors);
    return response.data.count || 0;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function templateNameAvailability(templateName: string, scopeType: string, scopeId: string): Promise<boolean> {
  try {
    await workloadServiceApi.workloadTemplatesApi.templateNameAvailability(
      templateName,
      scopeType as "tenant" | "cluster" | "department" | "project",
      scopeId,
    );
    return true;
  } catch (error: unknown) {
    // 409 if its taken
    // other error status are other issues
    if (error instanceof HttpErrorResponse) {
      if (error.statusCode === 409) return true;
      console.error(error.serialize());
    }
    throw error;
  }
}

export type TTemplateSortBy =
  | "name"
  | "workloadType"
  | "runnable"
  | "scopeType"
  | "clusterId"
  | "projectId"
  | "departmentId"
  | "createdAt"
  | "updatedAt"
  | "createdBy"
  | "updatedBy";

async function templatesList(
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc",
  sortBy?: TTemplateSortBy,
  filterBy?: Array<string>,
  search?: string,
  includeAncestors?: boolean,
): Promise<SupersetTemplate[]> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.listTemplates(
      offset,
      limit,
      sortOrder,
      sortBy,
      filterBy,
      search,
      includeAncestors,
    );
    // Note: currently .Entries would show a ts error because the generated client expect this to be .entries
    // but the backend returns .Entries
    return response.data.entries || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getFlatTemplateById(templateId: string): Promise<SupersetTemplate> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.getTemplateById(templateId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function workspaceTemplatesList(
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc",
  sortBy?: TTemplateSortBy,
  filterBy?: Array<string>,
  search?: string,
  includeAncestors = false,
  complyToProject?: number,
): Promise<WorkspaceTemplatesListResponse> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.listWorkspaceTemplates(
      offset,
      limit,
      sortOrder,
      sortBy,
      filterBy,
      search,
      includeAncestors,
      complyToProject,
    );

    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function trainingTemplatesList(
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc",
  sortBy?: TTemplateSortBy,
  filterBy?: Array<string>,
  search?: string,
  includeAncestors = false,
  complyToProject?: number,
): Promise<TrainingTemplatesListResponse> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.listTrainingTemplates(
      offset,
      limit,
      sortOrder,
      sortBy,
      filterBy,
      search,
      includeAncestors,
      complyToProject,
    );
    return response.data || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function distributedTemplatesList(
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc",
  sortBy?: TTemplateSortBy,
  filterBy?: Array<string>,
  search?: string,
  includeAncestors = false,
  complyToProject?: number,
): Promise<DistributedTemplatesListResponse> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.listDistributedTemplates(
      offset,
      limit,
      sortOrder,
      sortBy,
      filterBy,
      search,
      includeAncestors,
      complyToProject,
    );
    return response.data || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function inferenceTemplatesList(
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc",
  sortBy?: TTemplateSortBy,
  filterBy?: Array<string>,
  search?: string,
  includeAncestors = false,
  complyToProject?: number,
): Promise<InferenceTemplatesListResponse> {
  try {
    const response = await workloadServiceApi.workloadTemplatesApi.listInferenceTemplates(
      offset,
      limit,
      sortOrder,
      sortBy,
      filterBy,
      search,
      includeAncestors,
      complyToProject,
    );

    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteFlatTemplate(template: SupersetTemplate): Promise<void> {
  if (template.meta.workloadType === WorkloadType.Workspace) {
    return deleteFlatWorkspaceTemplate(template.meta.id);
  } else if (template.meta.workloadType === WorkloadType.Training) {
    return deleteFlatTrainingTemplate(template.meta.id);
  } else if (template.meta.workloadType === WorkloadType.Distributed) {
    return deleteFlatDistributedTemplate(template.meta.id);
  } else if (template.meta.workloadType === WorkloadType.Inference) {
    return deleteFlatInferenceTemplate(template.meta.id);
  } else {
    console.warn(`Unsupported workload type: ${template.meta.workloadType}`);
  }
}

async function deleteFlatWorkspaceTemplate(id: string): Promise<void> {
  try {
    await workloadServiceApi.workloadTemplatesApi.deleteWorkspaceTemplate(id);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteFlatTrainingTemplate(id: string): Promise<void> {
  try {
    await workloadServiceApi.workloadTemplatesApi.deleteTrainingTemplate(id);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteFlatDistributedTemplate(id: string): Promise<void> {
  try {
    await workloadServiceApi.workloadTemplatesApi.deleteDistributedTemplate(id);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteFlatInferenceTemplate(id: string): Promise<void> {
  try {
    await workloadServiceApi.workloadTemplatesApi.deleteInferenceTemplate(id);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
