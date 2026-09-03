<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="ai-application-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          :columns="columns"
          @filters-changed="setFilterBy"
          @selected-actions-close="resetSelectedRows"
          @create-entity="createAIApplication"
          :show-primary-btn="showSubmitAIApplication"
          primary-btn-label="New AI application"
          support-refresh
          @refresh-data="refreshList"
          :selected-rows-amount="selectedRowsAmount"
          @drawer-toggled="isDrawerOpen = $event"
          @drawer-resized="drawerWidth = $event"
          @export-csv="exportCsv"
          advanced-filters
          show-details-btn
          hide-all-cluster-filter-option
        >
          <template v-slot:selected-rows-actions>
            <section v-if="showSubmitAIApplication" class="row items-center">
              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="deleteBtnTooltip">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Applications, action: Action.Delete }"
                  btn-action="delete"
                  aid="delete-ai-application-btn"
                  :disable="disableDelete"
                  @click="openDeleteModal"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
          <template v-slot:drawer-content>
            <ai-application-drawer-content
              :drawer-width="drawerWidth"
              :selected-applications-amount="selectedRowsAmount"
              :selected-application="selectedApplication"
            />
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :style="{ marginRight: tableMargin }"
          :columns="columns"
          :rows="tableData"
          :loading="loadingTableData"
          :filter-by="filterBy"
          v-model:selected="selectedApplications"
          selection="single"
          @update-filters="setFilterBy"
          @connections-clicked="displayConnectionsModal"
          :get-row-key="getRowKey"
          sticky-columns
          ref="runaiTable"
          is-server-side-pagination
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25, 50, 100]"
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData"
              :filter-by="filterBy"
              entity-name="AI application"
              ph-icon-name="ph-bold ph-app-window"
              icon-size="44"
              @clear-filters="clearFilters"
              @create-new="createAIApplication"
              :show-error="loadingError"
              :show-create-btn="showSubmitAIApplication"
              :disable-create-btn="!canCreateApplication"
              :show-filter-icon-and-button="true"
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      aid="delete-ai-application-modal"
      v-if="isDeleteModalOpen && selectedApplication"
      :entity-name="selectedApplication.name || ''"
      entity-type="AI application"
      @cancel="isDeleteModalOpen = false"
      @delete="removeApplication"
      :deleting="deleting"
    />

    <workload-connections-modal
      v-if="isConnectionsModalOpen"
      @close="closeConnectionsModal"
      :modal-options="connectionsModalOptions"
    />
  </section>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";

// Stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";

// Models
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import type { EApplicationColumnName } from "@/models/ai-application.model";
import { type IApplicationTable, EAiApplicationErrorMessage } from "@/models/ai-application.model";
import { aiApplicationIndexColumns, aiApplicationDependentColumns } from "@/table-models/ai-application.table-model";

// Components
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiTable } from "@/components/common/runai-table/";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiDeleteModal } from "@/components/common";
import { AiApplicationDrawerContent } from "@/components/ai-application/ai-application-drawer-content";
import { WorkloadConnectionsModal } from "@/components/workload/workload-connections-modal";

// Models
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { Phase } from "@/swagger-models/workloads-client";
import type { AiApplicationEndpoint } from "@/swagger-models/workloads-client";
import type { IWorkloadConnectionsModalOptions } from "@/models/workload.model";

// Utils
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert } from "@/utils/error-alert.util";

// Services
import { aiApplicationService } from "@/services/cluster/ai-application.service/ai-application.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// Composables
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { useExportFileComposable } from "@/composables/use-export-file-composable/use-export-file-composable";
import { EAdvancedExportEntity } from "@/composables/helpers/use-export-file.mapper";
import {
  MIN_CLUSTER_VERSION_FOR_CREATING_AI_APPLICATIONS,
  MIN_CLUSTER_VERSION_FOR_ENDPOINTS,
} from "@/common/version.constant";
import { AI_APPLICATION_ROUTE_NAMES } from "@/router/ai-application.routes/ai-application.routes.names";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";

export default defineComponent({
  name: "ai-application-index",
  components: {
    RunaiPageActions,
    RunaiActionButton,
    RunaiTooltipWrapper,
    RunaiDeleteModal,
    RunaiTableNoData,
    RunaiTable,
    RunaiTableWrapper,
    AiApplicationDrawerContent,
    WorkloadConnectionsModal,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      permissionStore: usePermissionStore(),
      selectedApplications: [] as Array<IApplicationTable>,
      isDeleteModalOpen: false as boolean,
      deleting: false as boolean,
      isDrawerOpen: false as boolean,
      drawerWidth: "0px" as string,
      freeSearchTerm: "" as string,
      isConnectionsModalOpen: false as boolean,
      connectionsModalOptions: { connections: [], header: "" } as IWorkloadConnectionsModalOptions,
    };
  },
  setup() {
    const clusterStore = useClusterStore();
    const settingStore = useSettingStore();

    const activeClusterFilterId = computed(() => clusterStore.selectedClusterId);

    const showSubmitAIApplication = computed(
      () =>
        unleashService.showSubmitAIApplication() &&
        settingStore.isApplicationSubmissionEnabled !== false &&
        clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_CREATING_AI_APPLICATIONS),
    );

    const columns = computed(() => {
      return aiApplicationIndexColumns.filter((col: ITableColumn) => {
        if (aiApplicationDependentColumns.cluster.has(col.name)) {
          return clusterStore.isMultiCluster;
        }
        if (aiApplicationDependentColumns.endpoints.has(col.name as EApplicationColumnName)) {
          return clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_ENDPOINTS);
        }
        return true;
      });
    });

    const getApplications = async (filterBy: IFilterBy): Promise<IApplicationTable[]> => {
      return aiApplicationService.list(filterBy);
    };

    const getApplication = async (id: string): Promise<IApplicationTable> => {
      const app = await aiApplicationService.getApplicationById(id);
      return app as IApplicationTable;
    };

    const {
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      clearFilters,
      refreshList,
      updateFilterByKey,
      initTableFilter,
      setFilterBy,
    } = useTableDataAdvancedFilters<IApplicationTable>(
      getApplications,
      columns,
      EAdvancedIndexPages.AI_APPLICATION,
      getApplication,
    );

    initTableFilter();

    const { exportCsv } = useExportFileComposable<IApplicationTable>(
      filterBy,
      columns,
      EAdvancedExportEntity.AIApplications,
      (filterBy, sortBy, sortOrder, offset, limit, search) =>
        aiApplicationService.getApplications(offset, limit, sortOrder, sortBy, filterBy, search),
      (filter) => aiApplicationService.getApplicationsCount(filter.filterBy, filter.search),
    );

    return {
      activeClusterFilterId,
      filterBy,
      columns,
      tableData,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      clearFilters,
      refreshList,
      updateFilterByKey,
      setFilterBy,
      exportCsv,
      showSubmitAIApplication,
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
    selectedApplication(): IApplicationTable | undefined {
      return this.selectedApplications[0];
    },
    selectedRowsAmount(): number {
      return this.selectedApplications.length;
    },
    tableMargin(): string {
      return this.isDrawerOpen ? this.drawerWidth : "0px";
    },
    isRunaiSource(): boolean {
      return this.selectedApplication?.source === "runai";
    },
    canCreateApplication(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Applications, Action.Create);
    },
    canDeleteApplication(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Applications, Action.Delete);
    },
    disableDelete(): boolean {
      if (!this.canDeleteApplication) return true;
      return !this.isRunaiSource;
    },
    deleteBtnTooltip(): string {
      if (!this.canDeleteApplication) {
        return "You don't have permission to delete AI applications";
      }
      return "This action is not supported for the selected AI application. To perform the action, go to the application you used to create the AI application.";
    },
  },
  methods: {
    getRowKey(row: IApplicationTable): string {
      return row.id || "";
    },
    createAIApplication() {
      this.$router.push({ name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_NEW });
    },
    resetSelectedRows() {
      this.selectedApplications = [];
    },
    openDeleteModal() {
      this.isDeleteModalOpen = true;
    },
    async displayConnectionsModal(application: IApplicationTable): Promise<void> {
      const endpoints = application.endpoints || [];
      if (!endpoints.length) return;

      const workloadPhaseMap = await this.fetchWorkloadPhaseMap(application.id ?? "");
      const connections = endpoints.map((endpoint: AiApplicationEndpoint) => ({
        name: endpoint.host || "",
        toolType: "",
        connectionType: "",
        url: endpoint.url,
        containerPort: endpoint.port ? endpoint.port : undefined,
        workloadPhase: workloadPhaseMap.get(endpoint.workloadId ?? ""),
      }));

      this.connectionsModalOptions.connections = connections;
      this.connectionsModalOptions.header = `Connections Associated with AI Application ${application.name}`;
      this.connectionsModalOptions.workloadDetails = {
        currentUser: this.authStore.userEmail,
        isRunning: application.phase === Phase.Running,
        name: application.name ?? "",
        type: application.type ?? "",
        projectName: application.projectName ?? "",
      };
      this.isConnectionsModalOpen = true;
    },
    async fetchWorkloadPhaseMap(applicationId: string): Promise<Map<string, string>> {
      try {
        const filterBy = [`aiApplicationId==${applicationId}`];
        const workloads = await workloadService.getWorkloads(false, 0, 100, undefined, undefined, filterBy);
        return new Map(workloads.map((w) => [w.id, w.phase]));
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: EAiApplicationErrorMessage.FailedToLoadApplicationWorkloads,
        });
        this.$q.notify(errorAlert.getNotification(error));
        return new Map();
      }
    },
    closeConnectionsModal(): void {
      this.isConnectionsModalOpen = false;
      this.connectionsModalOptions = { connections: [], header: "" };
    },
    async removeApplication(): Promise<void> {
      if (!this.selectedApplication?.id) return;
      try {
        this.deleting = true;
        await aiApplicationService.deleteApplication(this.selectedApplication.id);
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`AI application ${this.selectedApplication.name} is being deleted`));
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`AI application ${this.selectedApplication?.name} couldn't be deleted`));
      } finally {
        this.resetSelectedRows();
        this.isDeleteModalOpen = false;
        this.deleting = false;
      }
    },
  },
});
</script>

<style lang="scss" scoped>
.ai-application-index {
  height: 100%;
}
</style>
