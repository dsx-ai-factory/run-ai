/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type {
  InferenceTemplate,
  InferenceTemplateUpdateRequest,
  TemplateAssetsIds,
} from "@/swagger-models/workloads-client";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { inferenceWorkloadUtil } from "../inference-workload.util/inference-workload.util";

export const inferenceTemplateUtil = {
  getTemplateModel,
  prepareTemplateRequestToUpdate,
};

function getTemplateModel(template: InferenceTemplate): IUISuperWorkloadCreationRequest {
  const model: IUISuperWorkloadCreationRequest = {
    name: template.meta?.name || "",
    projectId: template.meta?.projectId?.toString() || "",
    scopeType: template.meta?.scopeType,
    scopeId: template.meta?.scopeId,
    departmentId: template.meta?.departmentId || undefined,
    clusterId: template.meta?.clusterId || "",
    description: template.meta?.description || "",
    spec: {
      ...(template.spec || {}),
      connections: workloadHybridUtil.getConnectionsField(
        template.spec?.ports || [],
        template.spec?.exposedUrls || [],
        template.spec?.relatedUrls || [],
      ),
      compute: workloadHybridUtil.getComputeField(template.spec?.compute),
      storage: workloadHybridUtil.getStorageField(template.spec?.storage),
      security: workloadHybridUtil.getSecurityField(template.spec?.security),
    },
  };

  if (!model.spec.servingPort) {
    model.spec.servingPort = {
      protocol: "http",
    };
  }

  if (!model.spec.autoscaling) {
    model.spec.autoscaling = {
      minReplicas: 1,
      maxReplicas: 1,
    };
  }

  return model;
}

function prepareTemplateRequestToUpdate(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: TemplateAssetsIds,
): InferenceTemplateUpdateRequest {
  model.spec.storage = inferenceWorkloadUtil.removeEmptyStorageArrays(model.spec.storage);
  return workloadHybridUtil.prepareSingleReplicaWorkloadTemplateFlatUpdateRequest<InferenceTemplateUpdateRequest>(
    model,
    linkedAssetsIds,
  );
}
