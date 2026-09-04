/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IScopeModel, IUIAssetCreationMeta } from "@/models/global.model";
import { Scope } from "@/swagger-models/assets-service-client";
import { ScopeType } from "@/swagger-models/authorization-client";

export const scopeUtil = {
  updateScope,
  getScopeId,
  getScopeModelFromTypeAndId,
};

function updateScope(assetCreationMeta: IUIAssetCreationMeta, updatedScope: IScopeModel): IUIAssetCreationMeta {
  assetCreationMeta.scope = updatedScope.scope;
  if (updatedScope.scope === Scope.Project) {
    assetCreationMeta.clusterId = null;
    assetCreationMeta.departmentId = null;
    assetCreationMeta.projectId = updatedScope.projectId;
  } else if (updatedScope.scope === Scope.Department) {
    assetCreationMeta.clusterId = null;
    assetCreationMeta.projectId = null;
    assetCreationMeta.departmentId = updatedScope.departmentId;
  } else if (updatedScope.scope === Scope.Cluster) {
    assetCreationMeta.projectId = null;
    assetCreationMeta.departmentId = null;
    assetCreationMeta.clusterId = updatedScope.clusterId;
  } else {
    assetCreationMeta.clusterId = null;
    assetCreationMeta.projectId = null;
    assetCreationMeta.departmentId = null;
  }
  return assetCreationMeta;
}

function getScopeId(scopeModel: IScopeModel): string {
  switch (scopeModel.scope) {
    case Scope.Cluster:
      return String(scopeModel.clusterId);
    case Scope.Department:
      return String(scopeModel.departmentId);
    case Scope.Project:
      return String(scopeModel.projectId);
    default:
      return "";
  }
}

function getScopeModelFromTypeAndId(scopeType: ScopeType, scopeId: string): IScopeModel {
  switch (scopeType) {
    case ScopeType.Project:
      return {
        scope: Scope.Project,
        projectId: +scopeId,
      };
    case ScopeType.Department:
      return {
        scope: Scope.Department,
        departmentId: scopeId,
      };
    case ScopeType.Cluster:
      return {
        scope: Scope.Cluster,
        clusterId: scopeId,
      };
    default:
      return {
        scope: scopeType,
      };
  }
}
