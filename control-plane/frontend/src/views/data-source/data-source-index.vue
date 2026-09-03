<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="data-source-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          @filters-changed="setFilterBy"
          :columns="columns"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          @export-csv="exportTableAsCsv"
          @refresh-data="refreshList"
          @drawer-toggled="isDrawerOpen = $event"
          @drawer-resized="drawerWidth = $event"
          support-refresh
          show-details-btn
        >
          <template v-slot:primary-btn>
            <div class="q-mr-md">
              <q-btn
                color="primary"
                class="btn-icon-sm runai-icon-btn--default"
                icon="ph-bold ph-plus"
                aid="primary-cta-btn"
                label="new data source"
                :disable="disablePrimaryBtn"
              >
                <data-source-dropdown
                  @data-source-selected="createDataSource"
                  :allowed-options="allowedDatasourceOptions"
                />
              </q-btn>
            </div>
          </template>
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper :display-tooltip="disableRename" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.HostPathAssets, action: Action.Update }"
                  aid="rename-data-source-btn"
                  btn-action="rename"
                  @click="openRenameModal"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableClone" :tooltip-text="cloneBtnTooltip">
                <runai-action-button
                  aid="copy-edit-data-source-btn"
                  btn-action="duplicate"
                  @click="redirectTo"
                  :disable="disableClone"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="deleteBtnTooltip">
                <runai-action-button
                  aid="delete-data-source-btn"
                  btn-action="delete"
                  @click="openDeleteModal"
                  :disable="disableDelete"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
          <template v-slot:drawer-content>
            <data-source-drawer-content :data-source="dataSourceForDrawer" />
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="dataSources"
          :columns="columns"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          @update:selected="onRowSelected"
          :loading="totalLoading"
          @update-filters="setFilterBy"
          :top-row="lastCreatedDataSource"
          sticky-columns
          :get-row-key="getRowKey"
          :get-row-icon="getRowIcon"
          :bordered="false"
          :style="{ marginRight: tableMargin }"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedDataSource"
              :filter-by="filterBy"
              entity-name="data source"
              ph-icon-name="ph-bold ph-database"
              icon-size="44"
              :show-error="loadingError"
              :disable-create-btn="!canCreateDataSource"
              @clear-filters="clearFilters"
              ignore-cluster-filter
            >
              <template v-slot:primary-btn="{ disabled }">
                <q-btn
                  color="primary"
                  flat
                  class="btn-icon-sm runai-icon-btn--default"
                  icon="ph-bold ph-plus"
                  aid="empty-state-add-entity-btn"
                  label="New Data Source"
                  :disable="disabled"
                >
                  <data-source-dropdown
                    v-if="!disabled"
                    @data-source-selected="createDataSource"
                    :allowed-options="allowedDatasourceOptions"
                  />
                </q-btn>
              </template>
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedRows[0]"
      :deleting="isDeleting"
      :entity-name="selectedRows[0].meta.name"
      entity-type="data source"
      @cancel="isDeleteModalOpen = false"
      @delete="deleteDataSource"
    >
      <template #subheader>
        <div v-if="isAssetBeingUsedByWorkloads" class="usage-warning-text q-mb-md">
          At least one workload is using this data source. <br />
          Deleting the data source won’t affect those workloads, but only partial workload copying will be possible later
          on.
        </div>
      </template>
    </runai-delete-modal>

    <runai-rename-modal
      v-if="isRenameModalOpen"
      :asset-type="selectedDataSourceKind"
      :scope-model="selectedScopeModel"
      :model-value="isRenameModalOpen"
      :existing-name="selectedDataSourceName"
      :existing-description="selectedDataSourceDescription"
      :renaming="renaming"
      @close="isRenameModalOpen = false"
      @rename="renameDataSource"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// models
import {
  type DatasourceListResponseEntry,
  AssetKind,
  type GitUpdateRequest,
  type NFSUpdateRequest,
  type PVCUpdateRequest,
  type HostPathUpdateRequest,
  type S3UpdateRequest,
  type GitCreationRequest,
  type HostPathCreationRequest,
  type NFSCreationRequest,
  type S3CreationRequest,
  type ConfigMapCreationRequest,
  Scope,
  type AssetUpdateRequest,
  type SecretAssetUpdateRequest,
  type DatasourceListResponseAssetSpec,
  type PVCAsset,
  type ConfigMapUpdateRequest,
} from "@/swagger-models/assets-service-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { ITableColumn } from "@/models/table.model";
import type { IAssetsFilter } from "@/models/filter.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";

// stores
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAppStore } from "@/stores/app.store";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { DataSourceDropdown } from "@/components/data-source/data-source-dropdown";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiRenameModal } from "@/components/common/runai-rename-modal";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiDeleteModal } from "@/components/common";
import { DataSourceDrawerContent } from "@/components/data-source/data-source-drawer-content";

// utils
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";

// constants
import type { TUIDataSourceInstancesKinds } from "@/models/data-source.model";
import { DS_KIND_TO_DS_LIST_KEY, type TDataSourceKinds } from "@/models/data-source.model";
import {
  DATA_SOURCE_CREATION_ROUTE_MAP,
  DATA_SOURCE_ROUTE_NAMES,
} from "@/router/data-source.routes/data-source.routes.names";
import { dataSourceIndexColumns } from "@/table-models/data-source.table-model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";
import { usePermissionStore } from "@/stores/permissions.store";
import type { IScopeModel } from "@/models/global.model";
import { useTableData } from "@/composables/use-table-data.composable";
import { EIndexPages } from "@/composables/helpers/use-table-data.mapper";

// services
import { filterService } from "@/services/filter.service/filter.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";

// private types
const DSRoutes = {
  [AssetKind.Nfs]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_NFS_NEW,
  [AssetKind.S3]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_S3_NEW,
  [AssetKind.Git]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_GIT_NEW,
  [AssetKind.Pvc]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_PVC_NEW,
  [AssetKind.HostPath]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_HOST_PATH_NEW,
  [AssetKind.ConfigMap]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_CONFIG_MAP_NEW,
  [AssetKind.SecretVolume]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_SECRET_VOLUME_NEW,
} as Record<string, string>;

export type TDataSourceCreationRequests =
  | GitUpdateRequest
  | NFSUpdateRequest
  | PVCUpdateRequest
  | HostPathUpdateRequest
  | S3UpdateRequest
  | ConfigMapCreationRequest;

export default defineComponent({
  components: {
    RunaiTableWrapper,
    RunaiPageActions,
    DataSourceDropdown,
    RunaiActionButton,
    RunaiTooltipWrapper,
    RunaiRenameModal,
    RunaiTable,
    RunaiTableNoData,
    RunaiDeleteModal,
    DataSourceDrawerContent,
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      dataSourceStore: useDataSourceStore(),
      appStore: useAppStore(),
      selectedRows: [] as Array<DatasourceListResponseEntry>,
      isDeleteModalOpen: false as boolean,
      isRenameModalOpen: false as boolean,
      isDeleting: false as boolean,
      renaming: false as boolean,
      allowedDatasourceOptions: new Set([]) as Set<TUIDataSourceInstancesKinds>,
      isPvcDeletable: null as boolean | null,
      isDrawerOpen: false as boolean,
      drawerWidth: "0px" as string,
    };
  },
  setup() {
    const columns: ITableColumn[] = dataSourceIndexColumns;

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
    } = useTableData<DatasourceListResponseEntry>(
      useDataSourceStore().loadDataSources,
      columns,
      EIndexPages.DATA_SOURCE,
    );

    const indexSpecificFilter: IAssetsFilter = {
      usageInfo: true,
      statusInfo: true,
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
  async created() {
    this.allowedDatasourceOptions = await this.dataSourceStore.getAvailableDataSources();
    if (this.dataSourceStore.fallbackPolicyLoadingError) this.appStore.setFallback(true);

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
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    selectedDataSourceKind(): TDataSourceKinds {
      return this.selectedRows[0].meta.kind as TDataSourceKinds;
    },
    disablePrimaryBtn(): boolean {
      if (this.loadingTableData) return true;
      return !this.canCreateDataSource;
    },
    canCreateDataSource(): boolean {
      return this.permissionStore.hasPermission(
        [
          ResourceType.HostPathAssets,
          ResourceType.PvcAssets,
          ResourceType.CmVolumeAssets,
          ResourceType.S3Assets,
          ResourceType.NfsAssets,
          ResourceType.GitAssets,
          ResourceType.SecretVolumeAssets,
        ],
        Action.Create,
      );
    },
    canDeleteDataSource(): boolean {
      return this.permissionStore.hasPermission(ResourceType.PvcAssets, Action.Delete);
    },
    noActionPermissionText(): string {
      return "Action for the selected data source is not authorized";
    },
    disableDelete(): boolean {
      if (!this.canDeleteDataSource || this.isPvcDeletable === false) return true;
      const usedByTemplates = !!this.selectedRows[0]?.usedBy?.templates?.length;
      return usedByTemplates;
    },
    deleteBtnTooltip(): string {
      if (!this.canCreateDataSource) {
        return this.noActionPermissionText;
      } else if (this.isPvcDeletable === false) {
        return "The selected PVC can’t be deleted because it is being used by a data volume. To perform the action, first delete the data volume.";
      } else {
        return "The selected data source is being used by at least one template and therefore can't be deleted";
      }
    },
    disableClone(): boolean {
      return !this.canCreateDataSource;
    },
    disableRename(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.HostPathAssets, Action.Update);
    },
    cloneBtnTooltip(): string {
      return this.noActionPermissionText;
    },
    totalLoading(): boolean {
      return this.loadingTableData;
    },
    dataSources(): Array<DatasourceListResponseEntry> {
      return this.tableData;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    lastCreatedDataSource(): DatasourceListResponseEntry | null {
      return this.lastCreatedEntity;
    },
    selectedDataSourceName(): string {
      return this.selectedRows[0]?.meta.name || "";
    },
    selectedDataSourceDescription(): string {
      return this.selectedRows[0]?.meta.description || "";
    },
    selectedScopeModel(): IScopeModel {
      if (!this.selectedRows[0]) return { scope: Scope.Tenant };
      return {
        scope: this.selectedRows[0].meta.scope,
        projectId: this.selectedRows[0].meta.projectId,
        departmentId: this.selectedRows[0].meta.departmentId,
        clusterId: this.selectedRows[0].meta.clusterId,
      };
    },
    dataSourceForDrawer(): DatasourceListResponseEntry | undefined {
      if (!this.selectedRows[0] || this.selectedRows[0].meta.kind !== AssetKind.Pvc) return;

      if (this.lastCreatedDataSource !== null && this.lastCreatedDataSource.meta.id === this.selectedRows[0].meta.id) {
        return this.lastCreatedDataSource;
      }

      return this.dataSources.find(
        (dataSource: DatasourceListResponseEntry) => dataSource.meta.id === this.selectedRows[0].meta.id,
      );
    },
    tableMargin(): string {
      return this.isDrawerOpen ? this.drawerWidth : "0px";
    },
  },
  methods: {
    async onRowSelected(selectedRows: Array<DatasourceListResponseEntry>): Promise<void> {
      if (selectedRows[0] && selectedRows[0].meta.kind === AssetKind.Pvc) {
        this.isPvcDeletable = false;
        const pvcAsset = await dataSourceService.getPvcById(selectedRows[0].meta.id, {
          dvUsageInfo: true,
          statusInfo: true,
        });
        this.isPvcDeletable = pvcAsset.usedBy?.usedByDataVolumes !== true;
      } else {
        this.isPvcDeletable = null;
      }
    },
    getRowKey(dataSource: DatasourceListResponseEntry): string {
      return dataSource.meta.id;
    },
    getRowIcon(dataSource: DatasourceListResponseEntry): string {
      const kind: AssetKind = dataSource.meta.kind;
      return (DATA_SOURCE_ICONS_NAMES[kind as keyof object] as string) || DATA_SOURCE_ICONS_NAMES.default;
    },
    openRenameModal(): void {
      this.isRenameModalOpen = true;
    },
    createDataSource(type: TDataSourceKinds): void {
      const query = filterService.getClusterUuidQueryFromColumnFilter(this.filterBy);
      this.$router.push({ name: DATA_SOURCE_CREATION_ROUTE_MAP[type], query });
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    closeDeleteModal(): void {
      this.isDeleteModalOpen = false;
    },
    async deleteDataSource(): Promise<void> {
      if (!this.selectedRowsAmount) return;
      try {
        this.isDeleting = true;
        await this.dataSourceStore.deleteDataSource(this.selectedRows[0]);
        this.clearLastCreated();
        this.isDeleteModalOpen = false;
        this.$q.notify(alertUtil.getSuccess(`Data source ${this.selectedRows[0].meta.name} deleted`));
        await this.refreshList();
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify(alertUtil.getError(`Data source ${this.selectedRows[0].meta.name} couldn't be deleted`));
      } finally {
        this.resetSelectedRows();
        this.isDeleting = false;
      }
    },
    redirectTo(): void {
      const type: AssetKind | undefined = this.selectedRows[0]?.meta.kind;
      if (!type || !DSRoutes[type]) return;

      this.$router.push({
        name: DSRoutes[type],
        query: { fromCopyId: this.selectedRows[0].meta.id },
      });
    },
    async renameDataSource(changes: AssetUpdateRequest): Promise<void> {
      if (!this.selectedRows.length) return;
      try {
        this.renaming = true;

        const specKey: keyof DatasourceListResponseAssetSpec = DS_KIND_TO_DS_LIST_KEY[this.selectedDataSourceKind];

        const actualSpec = this.selectedRows[0].spec[specKey];
        const dataSourceToUpdate = {
          meta: {
            name: changes.name,
            description: changes.description,
          },
          spec: actualSpec,
        } as TDataSourceCreationRequests;
        await this.updateDataSource(dataSourceToUpdate);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Data source ${this.selectedDataSourceName} updated`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Data source ${this.selectedDataSourceName} couldn't be updated`));
        }
      } finally {
        this.resetSelectedRows();
        this.isRenameModalOpen = false;
        this.renaming = false;
      }
    },
    async updateDataSource(dataSourceToUpdate: TDataSourceCreationRequests): Promise<void> {
      const selectedDsId: string = this.selectedRows[0].meta.id;
      const dsKind: TDataSourceKinds = this.selectedDataSourceKind;
      switch (dsKind) {
        case "git":
          return this.dataSourceStore.updateGit(selectedDsId, dataSourceToUpdate as GitCreationRequest);
        case "hostPath":
          return this.dataSourceStore.updateHostPath(selectedDsId, dataSourceToUpdate as HostPathCreationRequest);
        case "nfs":
          return this.dataSourceStore.updateNfs(selectedDsId, dataSourceToUpdate as NFSCreationRequest);
        case "pvc": {
          const pvcToUpdate: PVCUpdateRequest = {
            meta: dataSourceToUpdate.meta,
            spec: { path: (dataSourceToUpdate as PVCAsset).spec.path },
          };
          return this.dataSourceStore.updatePVC(selectedDsId, pvcToUpdate);
        }
        case "s3":
          return this.dataSourceStore.updateS3(selectedDsId, dataSourceToUpdate as S3CreationRequest);
        case "config-map": {
          const configMapToUpdate: ConfigMapUpdateRequest = {
            meta: dataSourceToUpdate.meta,
          };
          return this.dataSourceStore.updateConfigMap(selectedDsId, configMapToUpdate);
        }
        case "secret-volume": {
          const secretToUpdate: SecretAssetUpdateRequest = {
            meta: dataSourceToUpdate.meta,
          };
          return this.dataSourceStore.updateSecretVolume(selectedDsId, secretToUpdate);
        }
      }
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));

      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.DataSource, this.dataSources, columns);
    },
  },
});
</script>
