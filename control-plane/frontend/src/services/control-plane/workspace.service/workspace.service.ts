// services
import { storageUtil } from "@/utils/storage.util";
import { dateUtil } from "@/utils/date.util";
import { type Workspace as WorkspaceFlat } from "@/swagger-models/workloads-client";
// models
import type { WorkspaceCreationRequestV2, WorkspaceV2 } from "@/swagger-models/assets-service-client";
import type { WorkspacePolicyV2 } from "@/swagger-models/policy-service-client";
import type { ILocalStatus } from "@/models/global.model";

// constants
import { policyService } from "../policy.service/policy.service";
import { ScopeType } from "@/swagger-models/authorization-client";
import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import type { HttpResponse, WorkspaceCreationRequest } from "@/swagger-models/workloads-client";

const WORKSPACE_STATUS = "workspaceStatus";

export const workspaceService = {
  resumeV2,
  stopV2,
  getById,
  stop,
  getPolicy,
  createWorkspaceV2,
  getWorkspaceMergedWithPolicy,
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

async function createWorkspaceV2(wsCreationRequest: WorkspaceCreationRequestV2): Promise<WorkspaceV2> {
  try {
    const response = await assetsServiceApi.workspaceApi.createWorkspaceV2(wsCreationRequest);
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

async function getById(workspaceId: string): Promise<WorkspaceV2> {
  try {
    const response = await assetsServiceApi.workspaceApi.getWorkspaceByIdV2(workspaceId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getWorkspaceMergedWithPolicy(
  projectId?: string,
  workloadId?: string,
  templateId?: string,
): Promise<WorkspaceV2> {
  try {
    const response = await assetsServiceApi.workspaceApi.getWorkspaceMerged(projectId, workloadId, templateId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getPolicy(projectId: number): Promise<WorkspacePolicyV2> {
  return policyService.getWorkspacePolicyByScope({ scope: ScopeType.Project, projectId });
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
