import { defineStore } from "pinia";

// models
import { type ComputeAsset, type ComputeUpdateRequest } from "@/swagger-models/assets-service-client";
import type { Action } from "@/swagger-models/authorization-client";
import { ResourceType, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IAssetsFilter } from "@/models/filter.model";

// services
import { computeResourceService } from "@/services/control-plane/compute-resource.service/compute-resource.service";
// stores
import { permissionService } from "@/services/authorization/permission.service/permission.service";

export const useComputeResourceStore = defineStore("ComputeResource", {
  state: () => ({
    computeResources: [] as Array<ComputeAsset>,
    selectedComputeResource: null as ComputeAsset | null,
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,
  }),
  getters: {
    computeResourcesList: (state): Array<ComputeAsset> => state.computeResources,
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(
        ResourceType.ComputeResources,
      );
    },
    async loadComputeResources(filterBy: IAssetsFilter = {}): Promise<Array<ComputeAsset>> {
      this.computeResources = await computeResourceService.list(filterBy);
      return this.computeResources;
    },
    async loadById(computeResourceId: string): Promise<ComputeAsset> {
      let computeResource: ComputeAsset | undefined = this.computeResources.find(
        (env: ComputeAsset) => env.meta.id === computeResourceId,
      );
      if (!computeResource) computeResource = await computeResourceService.getById(computeResourceId);
      this.setSelectedComputeResource(computeResource);
      return computeResource;
    },
    setSelectedComputeResource(computeResource: ComputeAsset | null): void {
      this.selectedComputeResource = computeResource;
    },
    async remove(computeResource: ComputeAsset): Promise<void> {
      await computeResourceService.remove(computeResource.meta.id);
    },
    async update(computeId: string, computeResource: ComputeUpdateRequest): Promise<void> {
      await computeResourceService.update(computeId, computeResource);
    },
  },
});
