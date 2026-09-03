<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="compute-resource-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :columns="columns"
          :filters="filterBy"
          primary-btn-label="new compute resource"
          :primary-btn-disable="disablePrimaryBtn"
          @filters-changed="setFilterBy"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          @create-entity="createComputeResource"
          hide-export-csv
        >
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper
                v-if="isEditAssetEnabled"
                :display-tooltip="disableEdit"
                :tooltip-text="noActionPermissionText"
              >
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.ComputeResources, action: Action.Update }"
                  aid="edit-compute-btn"
                  btn-action="edit"
                  :disable="disableEdit"
                  @click="editComputeResource"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper v-else :display-tooltip="disableRename" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.ComputeResources, action: Action.Update }"
                  aid="rename-compute-btn"
                  btn-action="rename"
                  @click="openRenameModal"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableClone" :tooltip-text="cloneBtnTooltip" max-width="400px">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.ComputeResources, action: Action.Create }"
                  btn-action="duplicate"
                  aid="copy-compute-btn"
                  @click="cloneComputeResource"
                  :disable="disableClone"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="deleteBtnTooltip">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.ComputeResources, action: Action.Delete }"
                  btn-action="delete"
                  aid="delete-compute-btn"
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
          class="cr-table"
          :rows="computeResources"
          :columns="columns"
          :loading="loadingTableData"
          :filter-by="filterBy"
          @row-select="onRowSelected"
          @update:selected="onRowSelected"
          @update-filters="setFilterBy"
          v-model:selected="selectedRows"
          :loading-error="loadingError"
          :get-row-key="getRowKey"
          :top-row="lastCreatedComputeResource"
          sticky-columns
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedComputeResource"
              :show-error="loadingError"
              :filter-by="filterBy"
              entity-name="compute resource"
              ph-icon-name="ph-bold ph-hard-drives"
              icon-size="44"
              @create-new="createComputeResource"
              @clear-filters="clearFilters"
              :disable-create-btn="!canCreateComputeResource"
              ignore-cluster-filter
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      aid="delete-compute-modal"
      v-if="isDeleteModalOpen && selectedComputeResource"
      :entity-name="selectedComputeResource.meta.name"
      entity-type="compute resource"
      @cancel="isDeleteModalOpen = false"
      @delete="removeComputeResource"
      :deleting="deleting"
    >
      <template #subheader>
        <div v-if="isAssetBeingUsedByWorkloads" class="usage-warning-text q-mb-md">
          At least one workload is using this compute resource. <br />
          Deleting the compute resource won’t affect those workloads, but only partial workload copying will be possible
          later on.
        </div>
      </template>
    </runai-delete-modal>

    <runai-rename-modal
      aid="rename-compute-modal"
      v-if="isRenameModalOpen"
      asset-type="compute"
      :model-value="isRenameModalOpen"
      :existing-name="selectedComputeName"
      :existing-description="selectedComputeDescription"
      :scope-model="selectedScopeModel"
      :renaming="renaming"
      @close="isRenameModalOpen = false"
      @rename="renameComputeResource"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiRenameModal } from "@/components/common/runai-rename-modal";
import { RunaiDeleteModal } from "@/components/common";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// stores
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";

// models
import {
  type ComputeAsset,
  type ComputeUpdateRequest,
  Scope,
  type AssetUpdateRequest,
} from "@/swagger-models/assets-service-client";
import { ResourceType, Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IAssetsFilter } from "@/models/filter.model";
import { computeResourceIndexColumns } from "@/table-models/compute-resource.table-model";

// utils
import { alertUtil } from "@/utils/alert.util";
import { assetsUtil } from "@/utils/assets.util/assets.util";

// constants
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { IScopeModel } from "@/models/global.model";
// composables
import { useTableData } from "@/composables/use-table-data.composable";
import { EIndexPages } from "@/composables/helpers/use-table-data.mapper";

// services
import { filterService } from "@/services/filter.service/filter.service";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  components: {
    RunaiTableWrapper,
    RunaiPageActions,
    RunaiActionButton,
    RunaiTooltipWrapper,
    RunaiRenameModal,
    RunaiTable,
    RunaiTableNoData,
    RunaiDeleteModal,
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      computeResourceStore: useComputeResourceStore(),
      appStore: useAppStore(),
      settingStore: useSettingStore(),
      authStore: useAuthStore(),
      selectedRows: [] as Array<ComputeAsset>,
      isDeleteModalOpen: false as boolean,
      isRenameModalOpen: false as boolean,
      deleting: false as boolean,
      renaming: false as boolean,
    };
  },
  setup() {
    const showGpuLimitColumn = useSettingStore().isGPUOverProvisioningEnabled;
    const columns = showGpuLimitColumn
      ? computeResourceIndexColumns
      : computeResourceIndexColumns.filter((c) => c.name !== "gpu-memory-limit");

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
    } = useTableData<ComputeAsset>(
      useComputeResourceStore().loadComputeResources,
      columns,
      EIndexPages.COMPUTE_RESOURCE,
    );

    const indexSpecificFilter: IAssetsFilter = {
      usageInfo: true,
    };

    initTableFilter(indexSpecificFilter);

    return {
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
  created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    isAssetBeingUsedByWorkloads(): boolean {
      return (
        !!this.selectedRows[0]?.usedBy?.workspaces?.length ||
        !!this.selectedRows[0]?.usedBy?.trainings?.length ||
        !!this.selectedRows[0]?.usedBy?.inferences?.length ||
        !!this.selectedRows[0].usedBy?.distributed?.length
      );
    },
    disableEdit(): boolean {
      if (!this.selectedRows[0]) return true;
      return !assetsUtil.isPermittedAsset(this.selectedRows[0].meta, this.updateComputePermissions);
    },
    updateComputePermissions(): PermittedScopes {
      return orgTreeService.getExpandedPermittedScopes(
        this.computeResourceStore.actionPermissionsByScopes.update || {},
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
      if (this.loadingTableData) return true;
      return !this.canCreateComputeResource;
    },
    canCreateComputeResource(): boolean {
      return this.permissionStore.hasPermission(ResourceType.ComputeResources, Action.Create);
    },
    canDeleteCompute(): boolean {
      return this.permissionStore.hasPermission(ResourceType.ComputeResources, Action.Delete);
    },
    disableDelete(): boolean {
      if (!this.canDeleteCompute) return true;
      const usedByTemplates = !!this.selectedRows[0]?.usedBy?.templates?.length;
      return usedByTemplates;
    },
    deleteBtnTooltip(): string {
      if (!this.canDeleteCompute) {
        return this.noActionPermissionText;
      } else {
        return "The selected compute resource is being used by at least one template and therefore can't be deleted.";
      }
    },
    disableClone(): boolean {
      return !this.canCreateComputeResource;
    },
    disableRename(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.ComputeResources, Action.Update);
    },
    cloneBtnTooltip(): string {
      return this.noActionPermissionText;
    },
    noActionPermissionText(): string {
      return "Action for the selected compute resource is not authorized";
    },
    computeResources(): Array<ComputeAsset> {
      return this.tableData;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    selectedComputeResource(): ComputeAsset | undefined {
      return this.selectedRows[0];
    },
    lastCreatedComputeResource(): ComputeAsset | null {
      return this.lastCreatedEntity;
    },
    selectedComputeName(): string {
      return this.selectedRows[0]?.meta.name || "";
    },
    selectedComputeDescription(): string {
      return this.selectedRows[0]?.meta.description || "";
    },
    selectedScopeModel(): IScopeModel {
      if (!this.selectedComputeResource) return { scope: Scope.Tenant };
      return {
        scope: this.selectedComputeResource.meta.scope,
        projectId: this.selectedComputeResource.meta.projectId,
        departmentId: this.selectedComputeResource.meta.departmentId,
        clusterId: this.selectedComputeResource.meta.clusterId,
      };
    },
  },
  methods: {
    getRowKey(computeResource: ComputeAsset): string {
      return computeResource.meta.id;
    },
    editComputeResource(): void {
      this.$router.push({
        name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_EDIT,
        params: { id: this.selectedRows[0].meta.id },
      });
    },
    cloneComputeResource(): void {
      const selectedComputeResourceId: string = this.selectedRows[0].meta.id;
      this.$router.push({
        name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_NEW,
        query: { fromCopyId: selectedComputeResourceId },
      });
    },
    // // THIS FUNCTION IS NOT USED YET, but should be here for future use
    // async copyComputeResource(): Promise<void> {
    //   if (!this.selectedComputeResource) return;
    //   const computeCreationRequest: ComputeCreationRequest = {
    //     meta: {
    //       name: this.selectedComputeResource.meta.name + "-copy",
    //       scope: this.selectedComputeResource.meta.scope,
    //       projectId: this.selectedComputeResource.meta.projectId,
    //       departmentId: this.selectedComputeResource.meta.departmentId,
    //       clusterId: this.selectedComputeResource.meta.clusterId,
    //       description: this.selectedComputeResource.meta.description,
    //       workloadSupportedTypes: this.selectedComputeResource.meta.workloadSupportedTypes,
    //     },
    //     spec: this.selectedComputeResource.spec,
    //   };
    //   try {
    //     const createdComputeResource = await computeResourceService.create(computeCreationRequest);
    //     this.$q.notify(alertUtil.getSuccess(`Compute resource ${createdComputeResource.meta.name} created`));
    //     this.clearLastCreated();
    //     this.lastCreatedEntity = createdComputeResource;
    //     this.resetSelectedRows();
    //     await this.refreshList();
    //   } catch (e: unknown) {
    //     console.error(e);
    //     this.$q.notify(alertUtil.getError(`Failed to copy compute resource ${this.selectedRows[0].meta.name}`));
    //   }
    // },
    onRowSelected(): void {
      this.computeResourceStore.setSelectedComputeResource(this.selectedRows[0]);
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
      this.computeResourceStore.setSelectedComputeResource(null);
    },
    createComputeResource(): void {
      const query = filterService.getClusterUuidQueryFromColumnFilter(this.filterBy);
      this.$router.push({ name: "compute-resource-new", query });
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    openRenameModal(): void {
      this.isRenameModalOpen = true;
    },
    async removeComputeResource(): Promise<void> {
      if (!this.selectedComputeResource) return;
      try {
        this.deleting = true;
        await this.computeResourceStore.remove(this.selectedComputeResource);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Compute resource ${this.selectedComputeResource.meta.name} deleted`));
      } catch (e: unknown) {
        this.$q.notify(
          alertUtil.getError(`Compute resource ${this.selectedComputeResource?.meta.name} couldn't be deleted`),
        );
      } finally {
        this.resetSelectedRows();
        this.isDeleteModalOpen = false;
        this.deleting = false;
      }
    },
    async renameComputeResource(changes: AssetUpdateRequest): Promise<void> {
      if (!this.selectedRowsAmount) return;
      try {
        this.renaming = true;
        const computeResourceToUpdate: ComputeUpdateRequest = {
          meta: {
            name: changes.name,
            description: changes.description,
          },
          spec: this.selectedRows[0].spec,
        };
        await this.computeResourceStore.update(this.selectedRows[0].meta.id, computeResourceToUpdate);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Compute resource ${this.selectedComputeName} updated`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Compute resource ${this.selectedComputeName} couldn't be updated`));
        }
      } finally {
        this.resetSelectedRows();
        this.isRenameModalOpen = false;
        this.renaming = false;
      }
    },
  },
});
</script>
