/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Scope } from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { SupersetTemplate, TemplateMetaFields } from "@/swagger-models/workloads-client";
import type { EWorkloadType } from "@/models/workload.model";
import { TGI_DEPRECATION_MESSAGE_EDIT, TGI_DEPRECATION_MESSAGE_READ } from "@/models/inference.model";
import { isTgiInference } from "@/common/icon.util";

export const templatesUtil = {
  isPermittedTemplate,
  isTgiTemplate,
  getTgiInfoMessage,
};

function isPermittedTemplate(
  templateMeta: TemplateMetaFields,
  permissions: PermittedScopes,
  tenantId?: string,
): boolean {
  const { scopeType, scopeId } = templateMeta;

  switch (scopeType) {
    case Scope.System:
      // System templates are validated at tenant level (backend converts system scope to tenant for auth)
      return !!tenantId && !!permissions.tenants?.includes(tenantId);
    case Scope.Tenant:
      return !!permissions.tenants?.includes(scopeId || "");
    case Scope.Cluster:
      return !!permissions.clusters?.includes(scopeId || "");
    case Scope.Department:
      return !!permissions.departments?.includes(scopeId || "");
    case Scope.Project:
      return !!permissions.projects?.includes(scopeId || "");
    default:
      return false;
  }
}

/** Whether the template is an inference template that runs on a TGI server (detected by its image). */
function isTgiTemplate(template: SupersetTemplate): boolean {
  const image = template.spec?.image || template.combinedSpec?.image;
  return isTgiInference(template.meta.workloadType as EWorkloadType, image);
}

/**
 * The deprecation notice for a TGI template. Users who can edit the template are told to edit it, everyone else
 * is told to ask an administrator.
 */
function getTgiInfoMessage(canEdit: boolean): string {
  return canEdit ? TGI_DEPRECATION_MESSAGE_EDIT : TGI_DEPRECATION_MESSAGE_READ;
}
