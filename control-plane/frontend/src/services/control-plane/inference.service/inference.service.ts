/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// services
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";
import { policyService } from "@/services/control-plane/policy.service/policy.service";

// models
import type {
  InferenceCreationRequestV2,
  InferenceV2,
  HuggingFaceModelInfoRequest,
  HuggingFaceModelInfo,
} from "@/swagger-models/assets-service-client";
import type {
  Catalog,
  InferenceCreationRequest,
  InferenceUpdateSpecSpec,
  InferenceWorkloadMetricType,
  InferencePodMetricType,
  MetricsResponse,
  DistributedInferenceCreationRequest,
  DistributedInference,
} from "@/swagger-models/workloads-client";
import type { AxiosResponse } from "axios";
import type { InferencePolicyV2 } from "@/swagger-models/policy-service-client";
import { ScopeType } from "@/swagger-models/authorization-client";

import type { Inference as FlatInference } from "@/swagger-models/workloads-client";

export const inferenceService = {
  getById,
  getMetrics,
  getPodMetrics,
  createInferenceV2,
  createInferenceFlatApi,
  getFlatById,
  updateFlatWorkload,
  getHuggingFaceModelInfo,
  getPolicy,
  getInferenceMergedWithPolicy,
  getNimCatalogModels,
  createDistributedInference,
};

async function createInferenceFlatApi(inference: InferenceCreationRequest): Promise<FlatInference> {
  try {
    const response: AxiosResponse = await workloadServiceApi.inferencesApi.createInference(inference);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createDistributedInference(
  inference: DistributedInferenceCreationRequest,
): Promise<DistributedInference> {
  try {
    const response: AxiosResponse =
      await workloadServiceApi.distributedInferencesApi.createDistributedInference(inference);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getFlatById(inferenceId: string): Promise<FlatInference> {
  try {
    const response: AxiosResponse = await workloadServiceApi.inferencesApi.getInference(inferenceId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateFlatWorkload(inferenceId: string, inferenceSpec: InferenceUpdateSpecSpec): Promise<FlatInference> {
  try {
    const response: AxiosResponse = await workloadServiceApi.inferencesApi.updateInferenceSpec(inferenceId, {
      spec: inferenceSpec,
    });
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createInferenceV2(inference: InferenceCreationRequestV2): Promise<InferenceV2> {
  try {
    const response: AxiosResponse = await assetsServiceApi.inferenceApi.createInferenceV2(inference);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getById(inferenceId: string): Promise<InferenceV2> {
  try {
    const response = await assetsServiceApi.inferenceApi.getInferenceByIdV2(inferenceId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getHuggingFaceModelInfo(
  huggingFaceModelInfoRequest: HuggingFaceModelInfoRequest,
): Promise<HuggingFaceModelInfo> {
  try {
    const response: AxiosResponse =
      await assetsServiceApi.inferenceApi.getHuggingFaceModelInfo(huggingFaceModelInfoRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getPolicy(projectId: number): Promise<InferencePolicyV2> {
  return policyService.getInferencePolicyByScope({ scope: ScopeType.Project, projectId });
}

async function getInferenceMergedWithPolicy(
  projectId?: string,
  workloadId?: string,
  templateId?: string,
): Promise<InferenceV2> {
  try {
    const response = await assetsServiceApi.inferenceApi.getInferenceMerged(projectId, workloadId, templateId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNimCatalogModels(modelName?: string, page?: number, pageSize?: number): Promise<Catalog> {
  try {
    const res: AxiosResponse<Catalog> = await workloadServiceApi.inferencesApi.getNimCatalog(
      undefined,
      modelName,
      modelName ? "score" : "",
      page,
      pageSize,
    );
    return res.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

// new metrics apis
async function getMetrics(
  workloadId: string,
  start: string,
  end: string,
  metricType: Array<InferenceWorkloadMetricType>,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  try {
    const res: AxiosResponse<MetricsResponse> = await workloadServiceApi.inferencesApi.getInferenceWorkloadMetrics(
      workloadId,
      metricType,
      start,
      end,
      numberOfSamples,
    );
    return res.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getPodMetrics(
  workloadId: string,
  podId: string,
  start: string,
  end: string,
  metricType: Array<InferencePodMetricType>,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  try {
    const res: AxiosResponse<MetricsResponse> = await workloadServiceApi.inferencesApi.getInferenceWorkloadPodMetrics(
      workloadId,
      podId,
      metricType,
      start,
      end,
      numberOfSamples,
    );
    return res.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
