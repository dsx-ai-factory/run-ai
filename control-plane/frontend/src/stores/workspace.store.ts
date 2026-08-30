import { defineStore } from "pinia";

// models
import type {
  AssetsDefaults,
  PolicyInfo,
  WorkspaceCreationRequestV2,
  WorkspaceSpecificRunParams,
  WorkspaceV2,
} from "@/swagger-models/assets-service-client";
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IUIWorkloadCreation } from "@/models/workload.model";
// services
import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";
import { permissionService } from "@/services/authorization/permission.service/permission.service";

// utils
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { deepCopy } from "@/utils/common.util/common.util";
import type { THybridFormState } from "@/models/workload-hybrid.model";

export const useWorkspaceStore = defineStore("Workspace", {
  state: () => ({
    workspace: workloadUtil.getEmptyUIWorkloadCreation() as IUIWorkloadCreation,
    templateId: null as string | null,
    selectedWorkspace: null as WorkspaceV2 | null,
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,

    // these details are treat as read only after setting (saving the origin model with policies set)
    loadedWorkspaceModelWithDefaultsAndPolicy: null as WorkspaceV2 | null,

    // flat submission data ONLY
    flatFormState: null as THybridFormState | null,
  }),
  getters: {
    workspaceName(): string {
      return this.workspace.name || "";
    },
    loadedPolicyInfo(): PolicyInfo | null {
      return this.loadedWorkspaceModelWithDefaultsAndPolicy?.policyInfo || null;
    },
    // these details are treated as read only after setting (saving the origin model with policies set)
    loadedSpecificRunParamsWithDefaults(): WorkspaceSpecificRunParams | null {
      return deepCopy(this.loadedWorkspaceModelWithDefaultsAndPolicy?.specificRunParams || null);
    },
    loadedAssetsAttributesDefaults(): AssetsDefaults | null {
      return this.loadedWorkspaceModelWithDefaultsAndPolicy?.policyInfo?.attributeDefaults?.assets || null;
    },
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Workspaces);
    },
    setLoadedWorkspaceModelWithDefaults(model: WorkspaceV2): void {
      this.loadedWorkspaceModelWithDefaultsAndPolicy = model;
    },
    setWorkspace(workspace: IUIWorkloadCreation): void {
      this.workspace = workspace;
    },
    setTemplate(templateId: string | null): void {
      this.templateId = templateId;
    },
    resetCreateWorkspaceData(): void {
      this.templateId = null;
      this.loadedWorkspaceModelWithDefaultsAndPolicy = null;
      this.workspace = workloadUtil.getEmptyUIWorkloadCreation();
    },
    async createWorkspaceV2(workspace: WorkspaceCreationRequestV2): Promise<WorkspaceV2> {
      return workspaceService.createWorkspaceV2(workspace);
    },
    setFlatFormState(flatData: THybridFormState | null): void {
      this.flatFormState = flatData;
    },
    getFlatFormState(): THybridFormState | null {
      return this.flatFormState;
    },
    clearFlatFormState(): void {
      this.flatFormState = null;
    },
  },
});
