/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { legacyPolicyForForm } from "@/utils/gpu-memory-compat.util/gpu-memory-compat.util";

import {
  ScopeType,
  type PolicyListEntry,
  type PolicyMeta,
  type WorkspacePolicyV2,
  type TrainingPolicyV2,
  PolicyType,
  type TrainingPolicyChangeRequestV2,
  type WorkspacePolicyChangeRequestV2,
  type DistributedPolicyV2,
  type DistributedPolicyChangeRequestV2,
  type PolicyCreationFields,
  type InferencePolicyV2,
  type InferencePolicyChangeRequestV2,
  type DistributedInferencePolicyV2,
  type FallbackPolicy,
} from "@/swagger-models/policy-service-client";

import { policyUtil, type IListPolicyQuery, type IPolicyScopeProperties } from "@/utils/policy.util/policy.util";
import { policyServiceApi } from "@/services/infra/client-apis/policy-service-api/policy-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { TPolicyChangeRequest, TPolicyScopes, TWorkloadPolicy } from "@/models/policy.model";

export const policyService = {
  listPolicies,
  deleteWorkspacePolicy,
  deleteTrainingPolicy,
  deleteDistributedPolicy,
  deleteInferencePolicy,
  deleteDistributedInferencePolicy,
  loadPolicyByMetaAndType,
  getEmptyPolicyModel,
  validatePolicy,
  updatePolicy,
  prepareDataForUpdateRequest,
  loadPolicyById,
  deletePolicyByType,
  getWorkspacePolicyByScope,
  getTrainingPolicyByScope,
  getInferencePolicyByScope,
  getDistributedPolicyByScope,
  getTenantFallbackPolicy,
  prepareDistributedDataForUpdateRequest,
};

async function listPolicies(query?: IListPolicyQuery): Promise<PolicyListEntry[]> {
  try {
    return await policyServiceApi.policyApi
      .listPolicies(
        query?.policyType,
        query?.scope,
        query?.departmentId,
        query?.projectId,
        query?.clusterId,
        query?.includeFallbackPolicies,
        query?.includeGvkPolicies,
      )
      .then((res) => res.data.policies || []);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deletePolicyByType(policyType: PolicyType, scopeProperties: IPolicyScopeProperties): Promise<void> {
  switch (policyType) {
    case PolicyType.Workspace:
      return deleteWorkspacePolicy(scopeProperties);
    case PolicyType.Training:
      return deleteTrainingPolicy(scopeProperties);
    case PolicyType.Distributed:
      return deleteDistributedPolicy(scopeProperties);
    case PolicyType.Inference:
      return deleteInferencePolicy(scopeProperties);
    case PolicyType.DistributedInference:
      return deleteDistributedInferencePolicy(scopeProperties);
  }
}

async function deleteWorkspacePolicy(scopeProperties: IPolicyScopeProperties): Promise<void> {
  try {
    const response = await policyServiceApi.policyApi.deleteWorkspacePolicy(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function deleteTrainingPolicy(scopeProperties: IPolicyScopeProperties): Promise<void> {
  try {
    const response = await policyServiceApi.policyApi.deleteTrainingPolicy(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteDistributedPolicy(scopeProperties: IPolicyScopeProperties): Promise<void> {
  try {
    const response = await policyServiceApi.policyApi.deleteDistributedPolicy(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteInferencePolicy(scopeProperties: IPolicyScopeProperties): Promise<void> {
  try {
    const response = await policyServiceApi.policyApi.deleteInferencePolicy(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteDistributedInferencePolicy(scopeProperties: IPolicyScopeProperties): Promise<void> {
  try {
    const response = await policyServiceApi.policyApi.deleteDistributedInferencePolicy(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function loadPolicyByMetaAndType(
  workloadType: PolicyType,
  scopeProperties: IPolicyScopeProperties,
): Promise<TWorkloadPolicy> {
  switch (workloadType) {
    case PolicyType.Workspace:
      return getWorkspacePolicyByScope(scopeProperties);
    case PolicyType.Training:
      return getTrainingPolicyByScope(scopeProperties);
    case PolicyType.Distributed:
      return getDistributedPolicyByScope(scopeProperties);
    case PolicyType.Inference:
      return getInferencePolicyByScope(scopeProperties);
    case PolicyType.DistributedInference:
      return getDistributedInferencePolicyByScope(scopeProperties);
    default:
      throw new Error("Policy type not supported");
  }
}

async function getTenantFallbackPolicy(): Promise<FallbackPolicy> {
  try {
    const response = await policyServiceApi.policyApi.getFallbackPolicy("tenant");
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getWorkspacePolicyByScope(scopeProperties: IPolicyScopeProperties): Promise<WorkspacePolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.getWorkspacePolicyV2(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getTrainingPolicyByScope(scopeProperties: IPolicyScopeProperties): Promise<TrainingPolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.getTrainingPolicyV2(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getDistributedPolicyByScope(scopeProperties: IPolicyScopeProperties): Promise<DistributedPolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.getDistributedPolicyV2(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getInferencePolicyByScope(scopeProperties: IPolicyScopeProperties): Promise<InferencePolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.getInferencePolicyV2(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDistributedInferencePolicyByScope(
  scopeProperties: IPolicyScopeProperties,
): Promise<DistributedInferencePolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.getDistributedInferencePolicyV2(
      scopeProperties.scope as TPolicyScopes,
      scopeProperties.departmentId || undefined,
      scopeProperties.projectId?.toString() || undefined,
      scopeProperties.clusterId || undefined,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function loadPolicyById(policyId: string): Promise<TWorkloadPolicy> {
  const policyType: PolicyType = policyUtil.getPolicyTypeScopeAndScopeId(policyId).type;
  const scopeProperties = policyUtil.getPolicyScopePropertiesFromPolicyId(policyId);

  switch (policyType) {
    case PolicyType.Workspace:
      return getWorkspacePolicyByScope(scopeProperties);
    case PolicyType.Training:
      return getTrainingPolicyByScope(scopeProperties);
    case PolicyType.Distributed:
      return getDistributedPolicyByScope(scopeProperties);
    case PolicyType.Inference:
      return getInferencePolicyByScope(scopeProperties);
    case PolicyType.DistributedInference:
      return getDistributedInferencePolicyByScope(scopeProperties);
    default:
      throw new Error("Policy type not supported");
  }
}

async function validatePolicy(
  workloadType: PolicyType,
  updatedPolicy: TPolicyChangeRequest,
  approveClusterDeletion = false,
): Promise<TWorkloadPolicy> {
  switch (workloadType) {
    case PolicyType.Training:
      return updateTrainingPolicy(updatedPolicy as TrainingPolicyChangeRequestV2, true, approveClusterDeletion);
    case PolicyType.Workspace:
      return updateWorkspacePolicy(updatedPolicy as WorkspacePolicyChangeRequestV2, true, approveClusterDeletion);
    case PolicyType.Distributed:
      return updateDistributedPolicy(updatedPolicy as DistributedPolicyChangeRequestV2, true, approveClusterDeletion);
    case PolicyType.Inference:
      return updateInferencePolicy(updatedPolicy as InferencePolicyChangeRequestV2, true, approveClusterDeletion);
    default:
      throw new Error("Policy type not supported");
  }
}

async function updatePolicy(
  workloadType: PolicyType,
  updatedPolicy: TPolicyChangeRequest,
  approveClusterDeletion = false,
): Promise<TWorkloadPolicy> {
  switch (workloadType) {
    case PolicyType.Training:
      return updateTrainingPolicy(updatedPolicy as TrainingPolicyChangeRequestV2, approveClusterDeletion);
    case PolicyType.Workspace:
      return updateWorkspacePolicy(updatedPolicy as WorkspacePolicyChangeRequestV2, approveClusterDeletion);
    case PolicyType.Distributed:
      return updateDistributedPolicy(updatedPolicy as DistributedPolicyChangeRequestV2, approveClusterDeletion);
    case PolicyType.Inference:
      return updateInferencePolicy(updatedPolicy as InferencePolicyChangeRequestV2, approveClusterDeletion);
    default:
      throw new Error("Policy type not supported");
  }
}
async function updateTrainingPolicy(
  updatedPolicy: TrainingPolicyChangeRequestV2,
  validateOnly = false,
  approveClusterDeletion = false,
): Promise<TrainingPolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.overwriteTrainingPolicyV2(
      validateOnly,
      approveClusterDeletion,
      updatedPolicy,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateWorkspacePolicy(
  updatedPolicy: WorkspacePolicyChangeRequestV2,
  validateOnly = false,
  approveClusterDeletion = false,
): Promise<WorkspacePolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.overwriteWorkspacePolicyV2(
      validateOnly,
      approveClusterDeletion,
      updatedPolicy,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateDistributedPolicy(
  updatedPolicy: DistributedPolicyChangeRequestV2,
  validateOnly = false,
  approveClusterDeletion = false,
): Promise<DistributedPolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.overwriteDistributedPolicyV2(
      validateOnly,
      approveClusterDeletion,
      updatedPolicy,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateInferencePolicy(
  updatedPolicy: InferencePolicyChangeRequestV2,
  validateOnly = false,
  approveClusterDeletion = false,
): Promise<InferencePolicyV2> {
  try {
    const response = await policyServiceApi.policyApi.overwriteInferencePolicyV2(
      validateOnly,
      approveClusterDeletion,
      updatedPolicy,
    );
    return legacyPolicyForForm(response.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function getEmptyPolicyModel(): WorkspacePolicyV2 | TrainingPolicyV2 | DistributedPolicyV2 | InferencePolicyV2 {
  return {
    meta: {
      id: "",
      name: "",
      createdBy: "",
      createdAt: "",
      updatedBy: "",
      updatedAt: "",
      scope: "" as ScopeType,
    },
  };
}

function prepareDataForUpdateRequest(
  data: WorkspacePolicyV2 | TrainingPolicyV2,
): TrainingPolicyChangeRequestV2 | WorkspacePolicyChangeRequestV2 {
  return { meta: preparePolicyMetaObjectForRequest(data.meta as PolicyMeta), policy: data.policy };
}

function prepareDistributedDataForUpdateRequest(data: DistributedPolicyV2): DistributedPolicyChangeRequestV2 {
  return { meta: preparePolicyMetaObjectForRequest(data.meta as PolicyMeta), policy: data.policy };
}

function preparePolicyMetaObjectForRequest(pMeta: PolicyMeta): PolicyCreationFields {
  const { scope, departmentId, projectId, name, clusterId } = pMeta;
  switch (scope as TPolicyScopes) {
    case ScopeType.Tenant:
      return { scope, name };
    case ScopeType.Cluster:
      return { scope, name, clusterId };
    case ScopeType.Department:
      return { scope, departmentId, name };
    case ScopeType.Project:
      return { scope, projectId, name };
  }
}
