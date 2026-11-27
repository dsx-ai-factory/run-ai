/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// services
import { storageUtil } from "@/utils/storage.util";
import { dateUtil } from "@/utils/date.util";
import { type Workspace as WorkspaceFlat } from "@/swagger-models/workloads-client";
// models
import type { ILocalStatus } from "@/models/global.model";

// constants
import { httpService } from "@/services/infra/https.service/http.service";
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import type { HttpResponse, WorkspaceCreationRequest } from "@/swagger-models/workloads-client";

const WORKSPACE_STATUS = "workspaceStatus";

export const workspaceService = {
  resumeV2,
  stopV2,
  stop,
  createWorkspaceFlat,
};

async function createWorkspaceFlat(wsCreationRequest: WorkspaceCreationRequest): Promise<WorkspaceFlat> {
  try {
    const response = await workloadServiceApi.workspacesApi.createWorkspace(wsCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function resumeV2(workspaceId: string): Promise<HttpResponse> {
  try {
    const response = await workloadServiceApi.workspacesApi.resumeWorkspace(workspaceId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function stopV2(workspaceId: string): Promise<HttpResponse> {
  try {
    const response = await workloadServiceApi.workspacesApi.suspendWorkspace(workspaceId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function _cleanupStatus() {
  const currentStatus: Record<string, ILocalStatus> = _getWorkspaceStatus();
  if (!currentStatus || Object.keys(currentStatus).length === 0) return;
  Object.keys(currentStatus).forEach((key: string) => {
    if (_shouldCleanLocalStatus(currentStatus[key])) {
      delete currentStatus[key];
    }
  });
  _saveWorkspaceStatus(currentStatus);
}

// this should be called only once in a while,
// because it's a heavy operation and not needed to be called on every status update
_cleanupStatus();

function _shouldCleanLocalStatus(workspaceToCheck: ILocalStatus): boolean {
  return dateUtil.moreThanHourFromNow(workspaceToCheck.creationTimestamp);
}

function _getWorkspaceStatus(): Record<string, ILocalStatus> {
  return storageUtil.get<Record<string, ILocalStatus>>(WORKSPACE_STATUS) || {};
}

function _saveWorkspaceStatus(workspaceStatus: Record<string, ILocalStatus>): void {
  storageUtil.save<Record<string, ILocalStatus>>(WORKSPACE_STATUS, workspaceStatus);
}
