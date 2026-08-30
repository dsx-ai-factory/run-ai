import { defineStore } from "pinia";

// models
import type {
  PVCAsset,
  WorkloadTemplate,
  WorkloadTemplateCreationRequest,
  WorkloadTemplateUpdateRequest,
} from "@/swagger-models/assets-service-client";
import type { Action } from "@/swagger-models/authorization-client";
import { ResourceType, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IAssetsFilter } from "@/models/filter.model";

// services
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import type { IUIVolume } from "@/models/data-source.model";
import { dataSourceUtil } from "@/utils/data-source.util";
import { permissionService } from "@/services/authorization/permission.service/permission.service";

export const useWorkloadTemplateStore = defineStore("WorkloadTemplate", {
  state: () => ({
    templates: [] as Array<WorkloadTemplate>,
    selectedWorkloadTemplate: null as WorkloadTemplate | null,
    workloadTemplateCreation: null as WorkloadTemplateCreationRequest | null,
    workloadVolumes: [] as Array<IUIVolume>,
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,
  }),
  getters: {
    workloadTemplateList(): Array<WorkloadTemplate> {
      return this.templates;
    },
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Templates);
    },
    async loadVolumesFromExisting(volumeIds: string[]): Promise<void> {
      const volumes: PVCAsset[] = await Promise.all(volumeIds.map((volumeId) => dataSourceService.getPvcById(volumeId)));
      this.workloadVolumes = dataSourceUtil.mapPvcsToUiVolumes(volumes);
    },
    setWorkloadVolumes(volumes: Array<IUIVolume>): void {
      this.workloadVolumes = volumes;
    },
    updateWorkloadTemplateCreation(workloadTemplateCreation: WorkloadTemplateCreationRequest | null): void {
      this.workloadTemplateCreation = workloadTemplateCreation;
    },
    async loadById(id: string): Promise<WorkloadTemplate> {
      this.selectedWorkloadTemplate = await workloadTemplateService.getById(id);
      return this.selectedWorkloadTemplate;
    },
    async loadWorkloadTemplates(filterBy: IAssetsFilter = {}): Promise<Array<WorkloadTemplate>> {
      this.templates = await workloadTemplateService.list(filterBy);
      return this.templates;
    },
    async deleteWorkloadTemplate(workloadTemplate: WorkloadTemplate): Promise<void> {
      await workloadTemplateService.remove(workloadTemplate);
    },
    async update(templateId: string, template: WorkloadTemplateUpdateRequest): Promise<void> {
      await workloadTemplateService.update(templateId, template);
    },
    setSelectedWorkloadTemplate(workloadTemplate: WorkloadTemplate): void {
      this.selectedWorkloadTemplate = workloadTemplate;
    },
    resetCreateWorkloadTemplateCreationData(): void {
      this.workloadTemplateCreation = null;
      this.workloadVolumes = [];
    },
  },
});
