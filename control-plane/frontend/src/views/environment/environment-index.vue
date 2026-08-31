<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="environment-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :columns="columns"
          :filters="filterBy"
          primary-btn-label="new environment"
          :primary-btn-disable="disablePrimaryBtn"
          @filters-changed="setFilterBy"
          @create-entity="createEnvironment"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          @export-csv="exportTableAsCsv"
        >
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper
                v-if="isEditAssetEnabled"
                :display-tooltip="disableEdit"
                :tooltip-text="disableEditTooltip"
              >
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Environments, action: Action.Update }"
                  :disable="disableEdit"
                  aid="edit-environment-btn"
                  btn-action="edit"
                  @click="editEnvironment"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper v-else :display-tooltip="disableRename" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Environments, action: Action.Update }"
                  aid="rename-environment-btn"
                  btn-action="rename"
                  @click="openRenameModal"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableClone" :tooltip-text="cloneBtnTooltip">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Environments, action: Action.Create }"
                  aid="copy-environment-btn"
                  btn-action="duplicate"
                  @click="cloneEnvironment"
                  :disable="disableClone"
                />
              </runai-tooltip-wrapper>

              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="deleteBtnTooltip">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Environments, action: Action.Delete }"
                  aid="delete-environment-btn"
                  btn-action="delete"
                  @click="openDeleteModal"
                  :disable="disableDelete"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="environments"
          :columns="columns"
          :loading="loadingTableData"
          :filter-by="filterBy"
          v-model:selected="selectedEnvironments"
          @update-filters="setFilterBy"
          @tool-clicked="displayToolsModal"
          :top-row="lastCreatedEnvironment"
          :get-row-key="getRowKey"
          :get-row-icon="getRowIcon"
          sticky-columns
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading && !lastCreatedEnvironment"
              :filter-by="filterBy"
              entity-name="environment"
              ph-icon-name="ph-bold ph-x-square"
              icon-size="44"
              @clear-filters="clearFilters"
              @create-new="createEnvironment"
              :show-error="loadingError || loadingModelsError"
              :disable-create-btn="!canCreateEnvironment"
              ignore-cluster-filter
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <tool-modal
      v-if="showToolModal"
      :environment-name="toolsModalEnvName"
      :tools="toolsToDisplay"
      @close="closeToolModal"
    ></tool-modal>

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedEnvironments[0]"
      :entity-name="selectedEnvName"
      entity-type="environment"
      @cancel="isDeleteModalOpen = false"
      @delete="removeEnvironment"
      :deleting="deleting"
    >
      <template #subheader>
        <div v-if="isAssetBeingUsedByWorkloads" class="usage-warning-text q-mb-md">
          At least one workload is using this environment. <br />
          Deleting the environment won’t affect those workloads, but only partial workload copying will be possible later
          on.
        </div>
      </template>
    </runai-delete-modal>

    <runai-rename-modal
      v-if="isRenameModalOpen"
      asset-type="environment"
      :scope-model="selectedScopeModel"
      :model-value="isRenameModalOpen"
      :existing-name="selectedEnvName"
      :existing-description="selectedEnvDescription"
      :renaming="renaming"
      @close="isRenameModalOpen = false"
      @rename="renameEnvironment"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// store
import { useEnvironmentStore } from "@/stores/environment.store";
import { useAuthStore } from "@/stores/auth.store";

// models
import { type ITableColumn, ETableExportCsvFilesNames } from "@/models/table.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { DistributedFramework } from "@/swagger-models/assets-service-client";
import {
  type EnvironmentAsset,
  type WorkloadRefAndStatus,
  type Connection,
  type EnvironmentUpdateRequest,
  Scope,
  type AssetUpdateRequest,
} from "@/swagger-models/assets-service-client";
import { ResourceType, Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { environmentIndexColumns } from "@/table-models/environment.table-model";

// component
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { ToolModal } from "@/components/environment/tool-modal";
import { RunaiRenameModal } from "@/components/common/runai-rename-modal";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiDeleteModal } from "@/components/common";

// utils
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";
import { assetsUtil } from "@/utils/assets.util/assets.util";

// composable
import { useTableData } from "@/composables/use-table-data.composable";
import { EIndexPages } from "@/composables/helpers/use-table-data.mapper";

// constants
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { useAppStore } from "@/stores/app.store";
import { HttpErrorResponse } from "@/models/http-response.model";
import { useSettingStore } from "@/stores/setting.store";
import { distFrameworkIconsMap, toolTypeIconsMap } from "@/common/icons.constant";
import { getNvidiaImageBasedIcon } from "@/common/icon.util";
import { usePermissionStore } from "@/stores/permissions.store";
import type { IScopeModel } from "@/models/global.model";

// services
import { filterService } from "@/services/filter.service/filter.service";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { GPT2_ASSET_ID } from "@/models/gpt2-model-spec.model";

export default defineComponent({
  components: {
    RunaiPageActions,
    RunaiActionButton,
    RunaiTooltipWrapper,
    RunaiTableWrapper,
    ToolModal,
    RunaiRenameModal,
    RunaiTable,
    RunaiTableNoData,
    RunaiDeleteModal,
  },
  data() {
    return {
      settingStore: useSettingStore(),
      permissionStore: usePermissionStore(),
      environmentStore: useEnvironmentStore(),
      appStore: useAppStore(),
      authStore: useAuthStore(),
      workspaces: null as Array<WorkloadRefAndStatus> | null,
      selectedEnvironments: [] as Array<EnvironmentAsset>,
      isDeleteModalOpen: false as boolean,
      isRenameModalOpen: false as boolean,
      deleting: false as boolean,
      renaming: false as boolean,
      showToolModal: false as boolean,
      toolsToDisplay: [] as Array<Connection>,
      toolsModalEnvName: "" as string,
      loadingModelsError: false as boolean,
    };
  },
  setup() {
    const columns: Array<ITableColumn> = environmentIndexColumns;

    const {
      initTableFilter,
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      clearLastCreated,
    } = useTableData<EnvironmentAsset>(useEnvironmentStore().loadEnvironments, columns, EIndexPages.ENVIRONMENT);

    const indexSpecificFilter: IAssetsFilter = {
      usageInfo: true,
      statusInfo: true,
      includeDescendants: true,
    };
    initTableFilter(indexSpecificFilter);

    return {
      initTableFilter,
      lastCreatedEntity,
      clearLastCreated,
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      columns,
    };
  },
  async created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    isAssetBeingUsedByWorkloads(): boolean {
      return (
        !!this.selectedEnvironments[0]?.usedBy?.workspaces?.length ||
        !!this.selectedEnvironments[0]?.usedBy?.trainings?.length ||
        !!this.selectedEnvironments[0]?.usedBy?.inferences?.length ||
        !!this.selectedEnvironments[0].usedBy?.distributed?.length
      );
    },
    disableEdit(): boolean {
      if (!this.selectedEnvironments[0]) return true;
      if (this.selectedEnvironments[0].meta.scope === Scope.System) return true;
      return !assetsUtil.isPermittedAsset(this.selectedEnvironments[0].meta, this.updateEnvironmentPermissions);
    },
    updateEnvironmentPermissions(): PermittedScopes {
      return orgTreeService.getExpandedPermittedScopes(
        this.environmentStore.actionPermissionsByScopes.update || {},
        this.authStore.orgUnitList,
      );
    },
    isEditAssetEnabled(): boolean {
      return unleashService.isUpdateAssetEnabled();
    },
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    disablePrimaryBtn(): boolean {
      if (this.loading) return true;
      return !this.canCreateEnvironment;
    },
    canDeleteEnvironment(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Environments, Action.Delete);
    },
    disableDelete(): boolean {
      if (!this.canDeleteEnvironment) return true;
      if (this.selectedEnvironments[0]?.meta.scope === Scope.System) return true;
      const usedByTemplates = !!this.selectedEnvironments[0]?.usedBy?.templates?.length;
      return usedByTemplates;
    },
    disableRename(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.Environments, Action.Update);
    },
    canCreateEnvironment(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Environments, Action.Create);
    },
    disableEditTooltip(): string {
      if (this.selectedEnvironments[0].meta.scope === Scope.System) return "The selected environment can’t be edited";
      return this.noActionPermissionText;
    },
    noActionPermissionText(): string {
      return "Action for the selected environment is not authorized";
    },
    deleteBtnTooltip(): string {
      if (!this.canCreateEnvironment) {
        return this.noActionPermissionText;
      } else if (this.selectedEnvironments[0]?.meta.scope === Scope.System) {
        return "The selected environment can’t be deleted";
      } else {
        return "The selected environment is being used by at least one template and therefore can't be deleted.";
      }
    },
    disableClone(): boolean {
      return !this.canCreateEnvironment;
    },
    cloneBtnTooltip(): string {
      return this.noActionPermissionText;
    },
    loading(): boolean {
      return this.loadingTableData;
    },
    allowDeployingGPT2ModelEnabled(): boolean {
      return this.settingStore.isAllowDeployingGPT2ModelEnabled;
    },
    environments(): Array<EnvironmentAsset> {
      if (this.loadingError || this.loadingModelsError) return [];
      return this.environmentList;
    },
    selectedRowsAmount(): number {
      return this.selectedEnvironments.length;
    },
    selectedEnvName(): string {
      return this.selectedEnvironments[0]?.meta.name || "";
    },
    selectedEnvDescription(): string {
      return this.selectedEnvironments[0]?.meta.description || "";
    },
    lastCreatedEnvironment(): EnvironmentAsset | null {
      return this.lastCreatedEntity;
    },
    selectedScopeModel(): IScopeModel {
      if (!this.selectedEnvironments[0]) return { scope: Scope.Tenant };
      return {
        scope: this.selectedEnvironments[0].meta.scope,
        projectId: this.selectedEnvironments[0].meta.projectId,
        departmentId: this.selectedEnvironments[0].meta.departmentId,
        clusterId: this.selectedEnvironments[0].meta.clusterId,
      };
    },
    environmentList(): Array<EnvironmentAsset> {
      return this.tableData.filter((env) =>
        this.allowDeployingGPT2ModelEnabled ? true : env.meta.id !== GPT2_ASSET_ID,
      );
    },
  },
  methods: {
    getRowIcon(environment: EnvironmentAsset): string {
      // Check for image-based icons first (NVIDIA PyTorch, NVIDIA general, etc.)
      const image = environment.spec.image;
      const imageBasedIcon = getNvidiaImageBasedIcon(image);
      if (imageBasedIcon) {
        return imageBasedIcon;
      }

      const distFramework: DistributedFramework | null = environment.meta.workloadSupportedTypes?.distFramework || null;
      if (distFramework) return distFrameworkIconsMap[distFramework];
      const connections: Array<Connection> = environment.spec.connections || [];
      if (!connections?.length) return "environment-gray-new";

      const defaultIcon = "custom-tool";
      const firstConnection = connections[0];
      return (
        (firstConnection.internalToolInfo && toolTypeIconsMap[firstConnection.internalToolInfo.toolType]) ||
        (firstConnection.externalToolInfo && toolTypeIconsMap[firstConnection.externalToolInfo.toolType]) ||
        defaultIcon
      );
    },
    getRowKey(env: EnvironmentAsset): string {
      return env.meta.id;
    },
    editEnvironment(): void {
      const selectedEnvironmentId: string = this.selectedEnvironments[0].meta.id;
      this.$router.push({
        name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_EDIT,
        params: { id: selectedEnvironmentId },
      });
    },
    cloneEnvironment(): void {
      const selectedEnvironmentId: string = this.selectedEnvironments[0].meta.id;
      this.$router.push({
        name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_NEW,
        query: { fromCopyId: selectedEnvironmentId },
      });
    },
    // THIS FUNCTION IS NOT USED YET, but should be here for future use
    // async copyEnvironment(): Promise<void> {
    //   const environmentCreationRequest: EnvironmentCreationRequest = {
    //     meta: {
    //       name: this.selectedEnvironments[0].meta.name + "-copy",
    //       scope: this.selectedEnvironments[0].meta.scope,
    //       projectId: this.selectedEnvironments[0].meta.projectId,
    //       departmentId: this.selectedEnvironments[0].meta.departmentId,
    //       clusterId: this.selectedEnvironments[0].meta.clusterId,
    //       description: this.selectedEnvironments[0].meta.description,
    //       workloadSupportedTypes: this.selectedEnvironments[0].meta.workloadSupportedTypes,
    //     },
    //     spec: this.selectedEnvironments[0].spec,
    //   };
    //   try {
    //     const createdEnvironment = await this.environmentStore.createEnvironment(environmentCreationRequest);
    //     this.$q.notify(alertUtil.getSuccess(`Environment ${createdEnvironment.meta.name} created`));
    //     this.clearLastCreated();
    //     this.lastCreatedEntity = createdEnvironment;
    //     this.resetSelectedRows();
    //     await this.refreshList();
    //   } catch (e: unknown) {
    //     console.error(e);
    //     this.$q.notify(alertUtil.getError(`Failed to copy environment ${this.selectedEnvironments[0].meta.name}`));
    //   }
    // },
    resetSelectedRows(): void {
      this.selectedEnvironments = [];
    },
    createEnvironment(): void {
      const query = filterService.getClusterUuidQueryFromColumnFilter(this.filterBy);
      this.$router.push({ name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_NEW, query });
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    openRenameModal(): void {
      this.isRenameModalOpen = true;
    },
    displayToolsModal(environment: EnvironmentAsset): void {
      if (!environment.spec.connections) return;
      this.showToolModal = true;
      this.toolsModalEnvName = environment.meta.name || "";
      this.toolsToDisplay = environment.spec.connections;
    },
    async removeEnvironment(): Promise<void> {
      if (!this.selectedEnvironments.length) return;
      try {
        this.deleting = true;
        await this.environmentStore.remove(this.selectedEnvironments[0].meta.id);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Environment ${this.selectedEnvName} deleted`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Environment ${this.selectedEnvName} couldn't be deleted`));
        }
      } finally {
        this.resetSelectedRows();
        this.isDeleteModalOpen = false;
        this.deleting = false;
      }
    },
    async renameEnvironment(changes: AssetUpdateRequest): Promise<void> {
      if (!this.selectedEnvironments.length) return;
      try {
        this.renaming = true;
        const environmentToUpdate: EnvironmentUpdateRequest = {
          meta: {
            name: changes.name,
            description: changes.description,
          },
          spec: this.selectedEnvironments[0].spec,
        };
        await this.environmentStore.update(this.selectedEnvironments[0].meta.id, environmentToUpdate);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Environment ${this.selectedEnvName} updated`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Environment ${this.selectedEnvName} couldn't be updated`));
        }
      } finally {
        this.resetSelectedRows();
        this.isRenameModalOpen = false;
        this.renaming = false;
      }
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));
      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.Environment, this.environments, columns);
    },

    closeToolModal(): void {
      this.toolsModalEnvName = "";
      this.showToolModal = false;
      this.toolsToDisplay = [];
    },
  },
});
</script>

<style lang="scss" scoped>
.usage-warning-text {
  font-size: 12px;
}
</style>
