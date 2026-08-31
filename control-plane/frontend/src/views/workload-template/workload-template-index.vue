<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-template-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          primary-btn-label="new template"
          :primary-btn-disable="disablePrimaryBtn"
          :columns="columns"
          :filters="filterBy"
          :selected-rows-amount="selectedRowsAmount"
          @filters-changed="setFilterBy"
          @selected-actions-close="resetSelectedRows"
          @create-entity="createWorkloadTemplate"
          @export-csv="exportTableAsCsv"
        >
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Templates, action: Action.Update }"
                  btn-action="rename"
                  @click="openRenameModal"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="!canCreateTemplate" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Templates, action: Action.Create }"
                  btn-action="duplicate"
                  aid="copy-template-btn"
                  @click="cloneWorkloadTemplate"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Templates, action: Action.Delete }"
                  btn-action="delete"
                  aid="delete-template-btn"
                  @click="openDeleteModal"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="workloadTemplates"
          :columns="columns"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          :loading="loading"
          @update-filters="setFilterBy"
          @data-source-clicked="displayDataSources"
          :top-row="lastCreatedWorkloadTemplate"
          :get-row-key="getRowKey"
          :get-row-icon="getRowIcon"
          sticky-columns
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading && !lastCreatedWorkloadTemplate"
              :show-error="loadingError"
              :filter-by="filterBy"
              entity-name="template"
              ph-icon-name="ph-bold ph-certificate"
              icon-size="44"
              :disable-create-btn="!canCreateTemplate"
              @clear-filters="clearFilters"
              @create-new="createWorkloadTemplate"
              ignore-cluster-filter
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <data-source-list-modal
      v-if="openDataSourceModal"
      @close="closeDataSourceModal"
      :modal-options="dataSourceModalOptions"
    ></data-source-list-modal>

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedWorkloadTemplate"
      :deleting="isDeleting"
      :entity-name="selectedWorkloadTemplate.meta.name"
      entity-type="template"
      @cancel="isDeleteModalOpen = false"
      @delete="deleteTemplate"
    />

    <runai-rename-modal
      v-if="isRenameModalOpen"
      asset-type="workload-template"
      :scope-model="selectedScopeModel"
      :model-value="isRenameModalOpen"
      :existing-name="selectedTemplateName"
      :existing-description="selectedTemplateDescription"
      :renaming="renaming"
      @close="isRenameModalOpen = false"
      @rename="renameTemplate"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { DataSourceListModal } from "@/components/data-source/data-source-list-modal";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiRenameModal } from "@/components/common/runai-rename-modal";
import { RunaiDeleteModal } from "@/components/common";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// models
import type {
  WorkloadTemplate,
  DatasourceRef,
  WorkloadTemplateUpdateRequest,
  AssetsRef,
  DistributedFramework,
  AssetUpdateRequest,
} from "@/swagger-models/assets-service-client";
import { Scope } from "@/swagger-models/assets-service-client";
import type { ITableColumn } from "@/models/table.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";
import { workloadTemplateIndexColumns } from "@/table-models/workload-template.table-model";

// stores
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";

// utils
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";

// constants
import { HttpErrorResponse } from "@/models/http-response.model";
import { distFrameworkIconsMap, toolTypeIconsMap } from "@/common/icons.constant";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import type { IScopeModel } from "@/models/global.model";
import type { IDataSourceModalOptions } from "@/models/data-source.model";
import { useTableData } from "@/composables/use-table-data.composable";
import { EIndexPages } from "@/composables/helpers/use-table-data.mapper";

// services
import { filterService } from "@/services/filter.service/filter.service";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
export default defineComponent({
  components: {
    RunaiTableWrapper,
    RunaiPageActions,
    DataSourceListModal,
    RunaiActionButton,
    RunaiRenameModal,
    RunaiDeleteModal,
    RunaiTable,
    RunaiTableNoData,
    RunaiTooltipWrapper,
  },
  data() {
    return {
      workloadTemplateStore: useWorkloadTemplateStore(),
      appStore: useAppStore(),
      permissionStore: usePermissionStore(),
      unleashService,
      selectedRows: [] as Array<WorkloadTemplate>,
      isDeleteModalOpen: false as boolean,
      isDeleting: false as boolean,
      openDataSourceModal: false as boolean,
      dataSourceModalOptions: { dataSourceIds: [], header: "" } as IDataSourceModalOptions,
      workloadTemplateLoading: false,
      isRenameModalOpen: false as boolean,
      renaming: false as boolean,
    };
  },
  setup() {
    const columns = workloadTemplateIndexColumns;

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
    } = useTableData<WorkloadTemplate>(
      useWorkloadTemplateStore().loadWorkloadTemplates,
      columns,
      EIndexPages.WORKLOAD_TEMPLATE,
    );

    initTableFilter();

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
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    disablePrimaryBtn(): boolean {
      if (this.loading) return true;
      return !this.canCreateTemplate;
    },
    disableDelete(): boolean {
      return !this.canDeleteTemplate;
    },
    canDeleteTemplate(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Templates, Action.Delete);
    },
    canCreateTemplate(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Templates, Action.Create);
    },
    loading(): boolean {
      return this.loadingTableData;
    },
    workloadTemplates(): Array<WorkloadTemplate> {
      return this.tableData;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    selectedWorkloadTemplate(): WorkloadTemplate | null {
      return this.selectedRows[0] || null;
    },
    selectedTemplateName(): string {
      return this.selectedRows[0]?.meta.name || "";
    },
    selectedTemplateDescription(): string {
      return this.selectedRows[0]?.meta.description || "";
    },
    lastCreatedWorkloadTemplate(): WorkloadTemplate | null {
      return this.lastCreatedEntity;
    },
    noActionPermissionText(): string {
      return "Action for the selected template is not authorized";
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
  },
  methods: {
    getRowIcon(workloadTemplate: WorkloadTemplate): string {
      const assetsRef: AssetsRef = workloadTemplate.spec.assets;
      const defaultIcon = "environment-gray-new";
      const distFramework: DistributedFramework | undefined =
        workloadTemplate.meta.workloadSupportedTypes?.distFramework;
      if (distFramework) return distFrameworkIconsMap[distFramework];
      if (!assetsRef.environment.toolTypes || assetsRef.environment.toolTypes.length === 0) {
        return defaultIcon;
      }
      const toolType: string = assetsRef.environment.toolTypes[0];
      return toolTypeIconsMap[toolType] || defaultIcon;
    },
    getRowKey(workloadTemplate: WorkloadTemplate): string {
      return workloadTemplate.meta.id;
    },
    cloneWorkloadTemplate(): void {
      this.$router.push({
        name: "workload-template-new",
        query: { fromCopyId: this.selectedWorkloadTemplate?.meta.id },
      });
    },
    createWorkloadTemplate(templateType: string): void {
      const query = filterService.getClusterUuidQueryFromColumnFilter(this.filterBy);
      if (templateType === "training") {
        this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW, query });
      } else if (templateType === "workspace") {
        this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW, query });
      } else {
        this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW, query });
      }
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    displayDataSources(workloadTemplate: WorkloadTemplate): void {
      if (!workloadTemplate.spec.assets.datasources) return;

      const datasources: Array<DatasourceRef> = workloadTemplate.spec.assets.datasources;
      this.dataSourceModalOptions.header = `Data Sources Associated with Template ${workloadTemplate.meta.name}`;
      this.dataSourceModalOptions.dataSourceIds = datasources.map((dataSource: DatasourceRef) => dataSource.id);
      this.openDataSourceModal = Boolean(workloadTemplate);
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    openRenameModal(): void {
      this.isRenameModalOpen = true;
    },
    closeDeleteModal(): void {
      this.isDeleteModalOpen = false;
    },
    async deleteTemplate(): Promise<void> {
      if (!this.selectedWorkloadTemplate) return;
      try {
        this.isDeleting = true;
        await this.workloadTemplateStore.deleteWorkloadTemplate(this.selectedWorkloadTemplate);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Template ${this.selectedWorkloadTemplate.meta.name} deleted`));
        this.isDeleteModalOpen = false;
        await this.workloadTemplateStore.loadWorkloadTemplates(this.filterBy);
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify(alertUtil.getError(`Template ${this.selectedWorkloadTemplate.meta.name} couldn't be deleted`));
      } finally {
        this.resetSelectedRows();
        this.isDeleting = false;
      }
    },
    async renameTemplate(changes: AssetUpdateRequest): Promise<void> {
      if (!this.selectedRows.length) return;
      try {
        this.renaming = true;
        const templateToUpdate: WorkloadTemplateUpdateRequest = this.readyModelForRenameRequest(
          changes.name,
          changes.description,
        );
        await this.workloadTemplateStore.update(this.selectedRows[0].meta.id, templateToUpdate);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Template ${this.selectedTemplateName} updated`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Template ${this.selectedTemplateName} couldn't be updated`));
        }
      } finally {
        this.resetSelectedRows();
        this.isRenameModalOpen = false;
        this.renaming = false;
      }
    },
    readyModelForRenameRequest(newName: string, newDescription?: string | null): WorkloadTemplateUpdateRequest {
      return {
        meta: {
          name: newName,
          description: newDescription,
        },
        spec: {
          assets: {
            environment: this.selectedRows[0].spec.assets.environment.id,
            compute: this.selectedRows[0].spec.assets.compute?.id || undefined,
            datasources: this.selectedRows[0].spec.assets.datasources?.map((ds: DatasourceRef) => {
              return {
                id: ds.id,
                kind: ds.kind,
              };
            }),
          },
          specificEnv: this.selectedRows[0].spec.specificEnv,
        },
      };
    },
    closeDataSourceModal(): void {
      this.openDataSourceModal = false;
      this.resetDataSourceModalOptions();
    },
    resetDataSourceModalOptions(): void {
      this.dataSourceModalOptions = { dataSourceIds: [], header: "" };
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));

      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.Template, this.workloadTemplates, columns);
    },
  },
});
</script>
