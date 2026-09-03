<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="data-volume-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          primary-btn-label="new data volume"
          @filters-changed="setFilterBy"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          :columns="columns"
          advanced-filters
          @refresh-data="refreshList"
          @create-entity="createDataVolume"
          support-refresh
          hide-export-csv
        >
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper :display-tooltip="disableRename" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Datavolumes, action: Action.Update }"
                  aid="edit-data-volume-btn"
                  btn-action="edit"
                  @click="editDataVolume"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="!canCreateDataVolume" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Datavolumes, action: Action.Create }"
                  aid="make-a-copy-data-volume-btn"
                  btn-action="duplicate"
                  @click="createDataVolume(selectedRows[0].id)"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="!canDeleteDataVolume" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  aid="delete-data-volume-btn"
                  btn-action="delete"
                  @click="openDeleteModal"
                  :disable="!canDeleteDataVolume"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="tableData"
          :columns="columns"
          :filter-by="filterBy"
          :loading="loadingTableData"
          v-model:selected="selectedRows"
          @update-filters="setFilterBy"
          :top-row="lastCreatedEntity"
          sticky-columns
          :get-row-key="getRowKey"
          :get-row-icon="getRowIcon"
          @scope-col-clicked="openScopesModal"
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedEntity"
              :filter-by="filterBy"
              entity-name="data volume"
              ph-icon-name="ph-bold ph-database"
              icon-size="44"
              :show-error="loadingError"
              :disable-create-btn="!canCreateDataVolume"
              @clear-filters="clearFilters"
              ignore-cluster-filter
              @create-new="createDataVolume"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-rename-modal
      v-if="isRenameModalOpen"
      asset-type="data volume"
      :scope-model="selectedScopeModel"
      :model-value="isRenameModalOpen"
      :existing-name="selectedDataVolumeName"
      :existing-description="selectedDataVolumeDescription"
      :renaming="renaming"
      @close="isRenameModalOpen = false"
      @rename="renameDataVolume"
    />

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedRows[0]"
      :deleting="deleting"
      :entity-name="selectedRows[0].name"
      entity-type="data volume"
      @cancel="isDeleteModalOpen = false"
      @delete="deleteDataVolume"
    />

    <data-volume-scope-modal
      v-if="isScopesModalOpen"
      :loading="scopesModalOptions.loading"
      :dv-name="scopesModalOptions.dvName"
      :selected-scopes="scopesModalOptions.sharedScopes"
      @close="closeScopesModal"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// Components
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiTable } from "@/components/common";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiRenameModal } from "@/components/common/runai-rename-modal";
import { RunaiDeleteModal } from "@/components/common/runai-delete-modal";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { DataVolumeScopeModal } from "@/components/data-volume/data-volume-scope-modal";
// Models
import type { ITableColumn } from "@/models/table.model";
import { dataVolumeIndexColumns } from "@/table-models/data-volume.table-model";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { dataVolumeService } from "@/services/control-plane/data-volume.service/data-volume.service";
import type { DatavolumeNoSharedScopes, SharedScope } from "@/swagger-models/datavolumes-service-client";
import { useAppStore } from "@/stores/app.store";
import { DATA_VOLUME_ROUTE_NAMES } from "@/router/data-volume.routes/data-volume.routes.names";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";
import type { IScopeModel } from "@/models/global.model";
import { alertUtil } from "@/utils/alert.util";
import type { IOrgTreeNodeId } from "@/models/org-tree.model";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { computed } from "vue";
import type { IFilterBy } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";

export enum EEmitNames {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "data-volume-index",
  components: {
    RunaiTableWrapper,
    RunaiTableNoData,
    RunaiTable,
    RunaiPageActions,
    RunaiRenameModal,
    RunaiDeleteModal,
    RunaiTooltipWrapper,
    RunaiActionButton,
    DataVolumeScopeModal,
  },
  emits: {},
  props: {},
  data() {
    return {
      appStore: useAppStore(),
      permissionStore: usePermissionStore(),
      selectedRows: [] as DatavolumeNoSharedScopes[],
      isRenameModalOpen: false as boolean,
      renaming: false as boolean,
      isDeleteModalOpen: false as boolean,
      deleting: false as boolean,
      isScopesModalOpen: false as boolean,
      scopesModalOptions: {
        loading: false,
        sharedScopes: [] as IOrgTreeNodeId[],
        dvName: "",
      },
    };
  },
  setup() {
    const columns = computed<ITableColumn[]>(() => dataVolumeIndexColumns);

    const getDataVolumes = async (filters: IFilterBy): Promise<DatavolumeNoSharedScopes[]> => {
      const { filterBy, sortBy, sortOrder, offset, limit, search } = filterService.mapColumnsFilterToFilterParams(
        filters,
        undefined,
        true,
      );

      return await dataVolumeService.listOriginatedDataVolumes(
        offset,
        limit,
        sortBy as
          | "name"
          | "projectId"
          | "createdAt"
          | "createdBy"
          | "projectName"
          | "departmentId"
          | "clusterId"
          | "phase",
        sortOrder,
        filterBy,
        search,
      );
    };

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
    } = useTableDataAdvancedFilters<DatavolumeNoSharedScopes>(
      getDataVolumes,
      columns,
      EAdvancedIndexPages.DATA_VOLUME,
      dataVolumeService.getDataVolumeById,
    );

    initTableFilter();

    return {
      columns,
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
    };
  },
  created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    canCreateDataVolume(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Datavolumes, Action.Create);
    },
    canDeleteDataVolume(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Datavolumes, Action.Delete);
    },
    disableRename(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.Datavolumes, Action.Update);
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    noActionPermissionText(): string {
      return "Action for the selected data source is not authorized";
    },
    selectedScopeModel(): IScopeModel {
      return {
        scope: "project",
        projectId: +this.selectedRows[0].projectId,
      };
    },
    selectedDataVolumeName(): string {
      return this.selectedRows[0]?.name || "";
    },
    selectedDataVolumeDescription(): string {
      return this.selectedRows[0]?.description || "";
    },
  },
  methods: {
    getRowKey(dataVolume: DatavolumeNoSharedScopes): string {
      return dataVolume.id;
    },
    getRowIcon(_: DatavolumeNoSharedScopes): string {
      return "data-volume";
    },
    createDataVolume(fromCopyId?: string): void {
      const query = fromCopyId ? { fromCopyId: this.selectedRows[0].id } : {};
      this.$router.push({ name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_NEW, query });
    },
    editDataVolume() {
      this.$router.push({ name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_EDIT, params: { id: this.selectedRows[0].id } });
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    openRenameModal(): void {
      this.isRenameModalOpen = true;
    },
    closeRenameModal(): void {
      this.isRenameModalOpen = false;
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    closeDeleteModal(): void {
      this.isDeleteModalOpen = false;
    },
    async openScopesModal(row: DatavolumeNoSharedScopes): Promise<void> {
      try {
        this.isScopesModalOpen = true;
        this.scopesModalOptions.loading = true;
        this.scopesModalOptions.dvName = row.name;
        const dvSharedScopes = await dataVolumeService.getSharedScopesById(row.id);
        this.scopesModalOptions.sharedScopes = dvSharedScopes.map(
          (scope: SharedScope): IOrgTreeNodeId => ({ id: scope.scopeId, type: scope.scopeType }),
        );
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.scopesModalOptions.loading = false;
      }
    },
    closeScopesModal(): void {
      this.isScopesModalOpen = false;
      this.scopesModalOptions = {
        loading: false,
        sharedScopes: [],
        dvName: "",
      };
    },
    async renameDataVolume(): Promise<void> {
      try {
        this.renaming = true;
        await dataVolumeService.updateDataVolume(this.selectedRows[0].id, {
          description: this.selectedDataVolumeDescription,
        });
        this.$q.notify(alertUtil.getSuccess(`Data volume ${this.selectedDataVolumeName} is being updated`));
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify(alertUtil.getError(`Data volume ${this.selectedDataVolumeName} couldn't be updated`));
      } finally {
        this.renaming = false;
        this.closeRenameModal();
        this.resetSelectedRows();
        this.refreshList();
      }
    },
    async deleteDataVolume(): Promise<void> {
      try {
        this.deleting = true;
        await dataVolumeService.deleteDataVolume(this.selectedRows[0].id);
        this.$q.notify(alertUtil.getSuccess(`Data volume ${this.selectedDataVolumeName} deleted`));
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify(alertUtil.getError(`Data volume ${this.selectedDataVolumeName} couldn't be deleted`));
      } finally {
        this.deleting = false;
        this.closeDeleteModal();
        this.resetSelectedRows();
        this.refreshList();
      }
    },
  },
});
</script>
