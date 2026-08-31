<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          ref="pageActions"
          :filters="filterBy"
          :columns="columns"
          @filters-changed="updateFilterBy"
          @selected-actions-close="resetSelectedRows"
          :show-primary-btn="!isDeletedViewType"
          @export-csv="exportCsv"
          support-refresh
          @refresh-data="loadWorkloads"
          show-details-btn
          empty-message="Select a workload to view its details"
          :selected-rows-amount="selectedRowsAmount"
          @drawer-toggled="isDrawerOpen = $event"
          @drawer-resized="drawerWidth = $event"
          advanced-filters
          hide-all-cluster-filter-option
        >
          <template v-slot:primary-btn>
            <create-workload-dropdown
              class="q-mr-md"
              :no-existing-projects="noExistingProjects"
              :is-cluster-connected="isClusterConnected"
              :is-yaml-submission-supported="isYamlSubmissionSupported"
            />
          </template>
          <template v-slot:selected-rows-actions>
            <workload-page-actions
              v-if="selectedWorkload"
              @action="handleWorkloadAction"
              :loading="isInitialLoader"
              :is-cluster-connected="isClusterConnected"
              :selected-workloads="selectedWorkloads"
              :workload="selectedWorkload"
              :is-deleted-view-type="isDeletedViewType"
              :selected-workloads-amount="selectedRowsAmount"
            />
          </template>
          <template v-slot:drawer-content>
            <workload-drawer-content
              :drawer-width="drawerWidth"
              :selected-workloads-amount="selectedRowsAmount"
              :workload="selectedWorkload"
              :is-deleted-view-type="isDeletedViewType"
              :hide-logs-tab="isDeletedViewType"
              :hide-terminal-tab="isDeletedViewType || !isVersionSupportTerminal"
              :hide-yaml-tab="isDeletedViewType || !isVersionSupportYamlTab"
              :hide-details-tab="!isVersionSupportWorkloadDetails"
              :hide-status="isDeletedViewType"
              :hide-architecture-tab="hideArchitectureTab"
              :navigation-payload="drawerNavigationPayload"
              @navigation-handled="clearDrawerNavigationPayload"
            />
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :style="{ marginRight: tableMargin }"
          :columns="columns"
          :rows="workloads"
          :loading="isInitialLoader"
          :filter-by="filterBy"
          v-model:selected="selectedWorkloads"
          selection="multiple"
          @update-filters="updateFilterBy"
          @pods-clicked="displayPodsModal"
          @connections-clicked="displayConnectionsModal"
          @data-source-clicked="displayDataSourcesModal"
          @project-clicked="displayNodePoolsModal"
          :get-row-key="getRowKey"
          :get-row-icon="getWorkloadIcon"
          :top-row="lastCreatedWorkload"
          sticky-columns
          ref="runaiTable"
          is-server-side-pagination
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25, 50, 100]"
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!isInitialLoader"
              :filter-by="filterBy"
              entity-name="workload"
              :ph-icon-name="isDeletedViewType ? 'ph-bold ph-trash-simple' : 'runai-workload'"
              icon-size="44"
              @clear-filters="clearFilters"
              :show-error="loadingError"
              :show-create-btn="!isDeletedViewType && canCreateAnyWorkloadType"
              :custom-message="noDeletedWorkloadsText"
              :show-filter-icon-and-button="!isDeletedViewType"
              ignore-cluster-filter
            >
              <template #primary-btn>
                <create-workload-dropdown
                  flat
                  :no-existing-projects="noExistingProjects"
                  :is-cluster-connected="isClusterConnected"
                  :is-yaml-submission-supported="isYamlSubmissionSupported"
                  :cluster-id="activeClusterFilterId"
                />
              </template>
              <template #secondary-text>
                <runai-table-no-projects
                  v-if="noExistingProjects && !isDeletedViewType"
                  entity-type="workload"
                  :has-entity-create-permission="canCreateWorkload"
                  :has-project-create-permission="canCreateProject"
                  :cluster-id="activeClusterFilterId"
                />
                <span v-else-if="isDeletedViewType">
                  Delete workloads you don't need anymore. Workloads will be deleted forever after 14 days
                </span>
              </template>
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>
    <template v-if="isDeleteModalOpen">
      <runai-delete-modal
        v-if="selectedWorkloads?.length"
        entity-type="workload"
        :custom-title="
          selectedWorkloads.length === 1
            ? `Permanently delete workload ${selectedWorkloads[0].name} and its logs?`
            : `Permanently delete ${deletableWorkloads.length} workloads and their logs?`
        "
        :entity-name="selectedWorkloads[0].name"
        @cancel="isDeleteModalOpen = false"
        @delete="deleteWorkloads"
        :delete-button-label="
          selectedWorkloads.length === 1 ? 'Delete' : `Delete ${deletableWorkloads.length} workloads`
        "
      >
        <template #subheader v-if="selectedWorkloads.length !== deletableWorkloads.length">
          <div class="row items-center q-gutter-sm q-mb-md">
            <span>
              {{ selectedWorkloads.length - deletableWorkloads.length }} out of {{ selectedWorkloads.length }} workloads
              can't be deleted
            </span>
            <runai-tooltip
              tooltip-text="At least one selected workload can't be deleted. You're either not authorized to delete them or need to go to the service account in which they were created to delete them from there."
            />
          </div>
        </template>
      </runai-delete-modal>
    </template>
    <workload-pods-modal
      v-if="isPodsModalOpen"
      :modal-options="podsModalOptions"
      @close="closePodsModal"
      @open-drawer-tab="handleOpenDrawerTab"
    />
    <data-source-list-modal
      v-if="isDataSourceModalOpen"
      @close="closeDataSourceModal"
      :modal-options="dataSourceModalOptions"
    />
    <workload-copy-cli-modal
      v-if="isCopyCliModalOpen && selectedWorkload"
      :workload-name="selectedWorkload.name"
      :cli-command="selectedWorkload?.additionalFields?.cliCommand"
      @cancel="isCopyCliModalOpen = false"
    />
    <workload-connections-modal
      v-if="isConnectionsModalOpen"
      @close="closeConnectionsModal"
      :modal-options="connectionsModalOptions"
    />
    <node-pools-modal
      :modal-options="nodePoolsModalOptions"
      v-if="isNodePoolsModalOpen"
      @close="closeNodePoolsModal"
      @redirect="redirectToProjects"
    />

    <workload-stop-modal
      v-if="isStopModalOpen"
      :selected-workloads-amount="selectedWorkloads.length"
      :stoppable-workloads-amount="selectedStoppableWorkloads.length"
      @close="isStopModalOpen = false"
      @stop-workloads="stopMultipleWorkloads"
    />

    <workload-run-modal
      v-if="isRunModalOpen"
      :selected-workloads-amount="selectedWorkloads.length"
      :runnable-workloads-amount="selectedRunnableWorkloads.length"
      @close="isRunModalOpen = false"
      @run-workloads="runMultipleWorkloads"
    />
  </section>
</template>

<script lang="ts">
import { computed, defineComponent, type PropType, ref, type DefineComponent } from "vue";
//store
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { useClusterStore } from "@/stores/cluster.store";

//models
import { EColumnFilterType, type IFilterBy, type IPaginationFilter } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { ETableFilters } from "@/models/table.model";
import {
  type BatchResponse,
  type ClusterPriority,
  type Connection,
  type Datasource,
  Phase,
  PodSortFilterFields,
  Source,
  type Workload,
  type WorkloadCategory,
  WorkloadSortFilterFields,
} from "@/swagger-models/workloads-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import {
  EWorkloadAction,
  EWorkloadErrorMessage,
  EWorkloadType,
  ExtendedPhase,
  type ICanSubmitWorkload,
  type IWorkloadConnectionsModalOptions,
  type IWorkloadPodsModalOptions,
  type IWorkloadDrawerNavigationPayload,
  type IOpenWorkloadDrawerTabPayload,
} from "@/models/workload.model";
import { useWorkloadDrawerNavigation } from "@/composables/use-workload-drawer-navigation.composable";
import {
  deletedWorkloadIndexColumns,
  workloadDependentColumns,
  workloadIndexColumns,
  workloadPodsModalDependentColumns,
  workloadPodsModalTableColumns,
} from "@/table-models/workload.table-model";

//components
import { WorkloadConnectionsModal } from "@/components/workload/workload-connections-modal";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiTable, RunaiTableNoProjects } from "@/components/common/runai-table/";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import RunaiDeleteModal from "@/components/common/runai-delete-modal/runai-delete-modal.vue";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { WorkloadDrawerContent } from "@/components/workload/workload-drawer-content";
import { WorkloadPodsModal } from "@/components/workload/workload-pods-modal";
import { WorkloadPageActions } from "@/components/workload/workload-page-actions";
import { CreateWorkloadDropdown } from "@/components/workload/workload-create/create-workload-dropdown/";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { WorkloadStopModal } from "@/components/workload/workload-stop-modal";
import { WorkloadRunModal } from "@/components/workload/workload-run-modal";

//utils
import { ErrorAlert } from "@/utils/error-alert.util";
import { alertUtil } from "@/utils/alert.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";

//services
import { filterService } from "@/services/filter.service/filter.service";
import { useAuthStore } from "@/stores/auth.store";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

import { useNodePoolStore } from "@/stores/node-pool.store";
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { DataSourceListModal } from "@/components/data-source/data-source-list-modal";
import { useSettingStore } from "@/stores/setting.store";
import type { IDataSourceModalOptions } from "@/models/data-source.model";
import WorkloadCopyCliModal from "@/components/workload/workload-copy-cli-modal/workload-copy-cli-modal.vue";
import {
  MIN_CLUSTER_VERSION_FOR_WORKLOAD_DETAILS_TAB,
  MIN_CLUSTER_VERSION_FOR_YAML_SUBMISSION,
  MIN_CLUSTER_VERSION_FOR_TERMINAL,
  MIN_CLUSTER_VERSION_FOR_K8S_OBJECT_API,
  MIN_CLUSTER_VERSION_FOR_ARCHITECTURE_TAB,
} from "@/common/version.constant";
import type { INodePoolModalOptions } from "@/models/project.model";
import { EIntervalLabels } from "@/models/interval.model";
import { intervalUtil } from "@/utils/interval.util";
import { EWorkloadViewType } from "@/router/workloads.routes/workloads.routes.names";
import { TRAINING_DISTRIBUTED_MIN_VERSION } from "@/models/cluster.model";
import { connectionAlertUtil } from "@/utils/connection-alert.util";
import NodePoolsModal from "@/components/node-pools/node-pools-modal/node-pools-modal.vue";
import { EQuotaEntity } from "@/models/resource.model";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { type OrgUnit, ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import type { LocationQueryRaw } from "vue-router";
import { useExportFileComposable } from "@/composables/use-export-file-composable/use-export-file-composable";
import { EAdvancedExportEntity } from "@/composables/helpers/use-export-file.mapper";
import type { ISelectOption } from "@/models/global.model";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
import { storageUtil } from "@/utils/storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";

const NODE_POOLS_MODAL_DEFAULT_CONFIG: INodePoolModalOptions = {
  header: "",
  entityType: EQuotaEntity.project,
  resources: [],
  nodePoolQuotaStatuses: [],
  showRedirectBtn: true,
  defaultNodePools: [],
};
const WorkloadIndex = defineComponent({
  name: "workload-index",
  components: {
    NodePoolsModal,
    WorkloadCopyCliModal,
    DataSourceListModal,
    WorkloadPageActions,
    RunaiTableNoProjects,
    CreateWorkloadDropdown,
    RunaiDeleteModal,
    WorkloadDrawerContent,
    RunaiPageActions,
    RunaiTableNoData,
    RunaiTable,
    RunaiTableWrapper,
    WorkloadPodsModal,
    WorkloadConnectionsModal,
    RunaiTooltip,
    WorkloadStopModal,
    WorkloadRunModal,
  },
  props: {
    viewType: {
      type: String as PropType<EWorkloadViewType>,
      required: true,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      workloadStore: useWorkloadStore(),
      permissionStore: usePermissionStore(),
      authStore: useAuthStore(),
      nodePoolStore: useNodePoolStore(),
      guidedTourStore: useGuidedTourStore(),
      loading: false as boolean,
      loadingError: false as boolean,
      isPodsModalOpen: false as boolean,
      isInitialLoader: true as boolean,
      selectedWorkloads: [] as Array<Workload>,
      loadingPolicies: false as boolean,
      isDrawerOpen: false as boolean,
      isClusterConnected: true as boolean,
      drawerWidth: "0px" as string,
      projectsCount: 0 as number,
      deleting: false as boolean,
      stopping: false as boolean,
      /** Workload priorities **/
      workloadPriorities: [] as ClusterPriority[],
      /** Modals **/
      isNodePoolsModalOpen: false as boolean,
      isDeleteModalOpen: false as boolean,
      isStopModalOpen: false as boolean,
      isRunModalOpen: false as boolean,
      isCopyCliModalOpen: false as boolean,
      isWorkloadsIntroModalOpen: true as boolean,
      isConnectionsModalOpen: false as boolean,
      isDataSourceModalOpen: false as boolean,
      projectNameForNodePoolsModal: "" as string,
      clusterIdForNodePoolsModal: "" as string,
      dataSourceModalOptions: { dataSourceIds: [], header: "" } as IDataSourceModalOptions,
      podsModalOptions: {
        header: "",
        displayRedirection: true,
        columns: [] as ITableColumn[],
        getPods: async () => [],
        getPodsCount: async () => 0,
      } as IWorkloadPodsModalOptions,
      connectionsModalOptions: {
        connections: [],
        header: "",
      } as IWorkloadConnectionsModalOptions,
      nodePoolsModalOptions: {
        ...NODE_POOLS_MODAL_DEFAULT_CONFIG,
      } as INodePoolModalOptions,
      /** Drawer Navigation **/
      drawerNavigationPayload: null as IWorkloadDrawerNavigationPayload | null,
      pendingDrawerNavigation: false as boolean,
    };
  },
  setup(props: { viewType: EWorkloadViewType }) {
    const clusterStore = useClusterStore();
    const settingStore = useSettingStore();
    const filterBy = ref<IFilterBy>({});
    const isOnlyDefaultNodePool = ref(false);
    const hideNodePoolsColumn = ref(false);

    const activeClusterFilterId = computed(() => clusterStore.selectedClusterId);
    const isDeletedViewType = computed(() => props.viewType === EWorkloadViewType.Deleted);
    const columns = computed(() => {
      const baseColumns = isDeletedViewType.value ? deletedWorkloadIndexColumns : workloadIndexColumns;
      return baseColumns.filter((col: ITableColumn) => {
        if (workloadDependentColumns.multiNode.has(col.name)) {
          return clusterStore.isClusterVersionSufficient(activeClusterFilterId.value, TRAINING_DISTRIBUTED_MIN_VERSION);
        }
        if (workloadDependentColumns.nodePools.has(col.name)) {
          return !isOnlyDefaultNodePool.value && !hideNodePoolsColumn.value && !isDeletedViewType.value;
        }
        if (workloadDependentColumns.hiddenOnDeletedView.has(col.name)) {
          return !isDeletedViewType.value;
        }

        if (workloadDependentColumns.cluster.has(col.name)) {
          return clusterStore.isMultiCluster;
        }

        return true;
      });
    });

    const { exportCsv } = useExportFileComposable<Workload>(
      filterBy,
      columns,
      EAdvancedExportEntity.Workloads,
      (filterBy, sortBy, sortOrder, offset, limit, search) =>
        workloadService.getWorkloads(
          isDeletedViewType.value,
          offset,
          limit,
          sortOrder,
          sortBy as WorkloadSortFilterFields,
          filterBy,
          search,
        ),
      (filter: IPaginationFilter) =>
        workloadService.getWorkloadsCount(filter.filterBy, isDeletedViewType.value, filter.search),
    );

    // Drawer navigation composable for cross-page navigation
    const { getDrawerParamsFromUrl, clearDrawerParams } = useWorkloadDrawerNavigation();

    return {
      clusterStore,
      settingStore,
      activeClusterFilterId,
      filterBy,
      isOnlyDefaultNodePool,
      hideNodePoolsColumn,
      isDeletedViewType,
      exportCsv,
      getDrawerParamsFromUrl,
      clearDrawerParams,
    };
  },
  created() {
    this.appStore.setPageLoading(false);
    this.loadFilters();
    this.loadProjectsCount();
    this.startRefreshWorkloads();
    this.loadIsOnlyDefaultNodePool();

    this.checkConnectivityStatus();
    this.loadWorkloadCategories();

    this.triggerWorkloadTourIfNeeded();

    // Check for cross-page drawer navigation params
    const drawerParams = this.getDrawerParamsFromUrl();
    if (drawerParams) {
      this.pendingDrawerNavigation = true;
    }
  },
  computed: {
    isYamlSubmissionSupported(): boolean {
      return (
        this.settingStore.isYamlSubmissionEnabled &&
        this.clusterStore.isClusterVersionSufficient(this.activeClusterFilterId, MIN_CLUSTER_VERSION_FOR_YAML_SUBMISSION)
      );
    },
    noExistingProjects(): boolean {
      return this.projectsCount === 0;
    },
    workloadCategoriesOptions(): ISelectOption[] {
      return this.workloadStore.workloadCategories.map((category: WorkloadCategory) => ({
        label: category.name,
        value: category.name,
      }));
    },
    filteredColumns(): ITableColumn[] {
      const baseColumns = this.isDeletedViewType ? deletedWorkloadIndexColumns : workloadIndexColumns;
      return baseColumns.filter((col: ITableColumn) => {
        if (workloadDependentColumns.multiNode.has(col.name)) {
          return this.clusterStore.isClusterVersionSufficient(
            this.activeClusterFilterId,
            TRAINING_DISTRIBUTED_MIN_VERSION,
          );
        }
        if (workloadDependentColumns.nodePools.has(col.name)) {
          return !this.isOnlyDefaultNodePool && !this.hideNodePoolsColumn && !this.isDeletedViewType;
        }
        if (workloadDependentColumns.hiddenOnDeletedView.has(col.name)) {
          return !this.isDeletedViewType;
        }

        if (workloadDependentColumns.cluster.has(col.name)) {
          return this.clusterStore.isMultiCluster;
        }

        return true;
      });
    },
    columns(): ITableColumn[] {
      return this.filteredColumns.map((col) => {
        if (col.name === WorkloadSortFilterFields.Category) {
          return {
            ...col,
            hideFilter: this.workloadCategoriesOptions.length === 0,
            filter: {
              type: EColumnFilterType.EnumBased,
              selectOptions: this.workloadCategoriesOptions,
            },
          };
        }
        if (col.name === WorkloadSortFilterFields.Priority) {
          return {
            ...col,
            hideFilter: this.workloadPriorities.length === 0,
            filter: {
              type: EColumnFilterType.EnumBased,
              selectOptions: this.workloadPriorities.map((priority: ClusterPriority) => ({
                label: priority.name,
                value: priority.name,
              })),
            },
          };
        }
        return col;
      });
    },
    podsModalColumns(): ITableColumn[] {
      return workloadPodsModalTableColumns.filter((col: ITableColumn) => {
        if (workloadPodsModalDependentColumns.nodePools.has(col.name)) {
          return !this.isOnlyDefaultNodePool;
        }
        return true;
      });
    },
    workloads(): Array<Workload> {
      if (this.isDeletedViewType) {
        return this.workloadStore.workloadList.map((workload: Workload) => {
          return {
            ...workload,
            phase: ExtendedPhase.Deleted as Phase,
          };
        });
      }
      return this.workloadStore.workloadList;
    },
    selectedRowsAmount(): number {
      return this.selectedWorkloads.length;
    },
    selectedWorkload(): Workload | null {
      const selectedWorkload = this.selectedWorkloads[0];
      if (!selectedWorkload) return null;

      if (this.lastCreatedWorkload !== null && this.lastCreatedWorkload.id === selectedWorkload.id) {
        return this.lastCreatedWorkload;
      }
      return this.workloadStore.allWorkloads.find((workload: Workload) => workload.id === selectedWorkload.id) || null;
    },
    tableMargin(): string {
      return this.isDrawerOpen ? this.drawerWidth : "0px";
    },
    canCreateWorkload(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Workloads, Action.Create);
    },
    canCreateProject(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Project, Action.Create);
    },
    canCreateAnyWorkloadType(): boolean {
      // Matches the dropdown's own gate so the empty-state CTA is hidden
      // for view-only users (no workload AND no inference create permission).
      return this.permissionStore.hasSomePermittedActions(
        [ResourceType.Workloads, ResourceType.Inferences],
        [Action.Create],
      );
    },
    canReadNodePools(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Nodepools, Action.Read);
    },
    lastCreatedWorkload(): Workload | null {
      return this.workloadStore.lastCreatedWorkload;
    },
    isVersionSupportWorkloadDetails(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.activeClusterFilterId,
        MIN_CLUSTER_VERSION_FOR_WORKLOAD_DETAILS_TAB,
      );
    },
    isVersionSupportTerminal(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.activeClusterFilterId, MIN_CLUSTER_VERSION_FOR_TERMINAL);
    },
    isVersionSupportYamlTab(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.activeClusterFilterId,
        MIN_CLUSTER_VERSION_FOR_K8S_OBJECT_API,
      );
    },
    hideArchitectureTab(): boolean {
      // Always show the tab when there's no selected workload so the empty state stays visible
      // alongside the other tabs. We only hide it for non-external workloads on supported clusters.
      if (!this.selectedWorkload) return false;
      const isExternalWorkload = !!this.selectedWorkload.externalWorkloadName;
      const isVersionSupported = this.clusterStore.isClusterVersionSufficient(
        this.activeClusterFilterId,
        MIN_CLUSTER_VERSION_FOR_ARCHITECTURE_TAB,
      );
      return !isExternalWorkload || !isVersionSupported;
    },
    deletableWorkloads(): Array<Workload> {
      return this.selectedWorkloads.filter((workload: Workload) => {
        if (workload.type === EWorkloadType.Inference) {
          return this.permissionStore.hasPermission(ResourceType.Inferences, Action.Create);
        }

        return (
          this.permissionStore.hasPermission(ResourceType.Workloads, Action.Delete) && !!workload.actionsSupport?.delete
        );
      });
    },
    /*Deleted view computed */
    tableFilterName(): ETableFilters.WORKLOAD_DELETED | ETableFilters.WORKLOAD {
      return this.isDeletedViewType ? ETableFilters.WORKLOAD_DELETED : ETableFilters.WORKLOAD;
    },
    noDeletedWorkloadsText(): string | undefined {
      return this.isDeletedViewType ? "No deleted workloads" : undefined;
    },
    selectedStoppableWorkloads(): Array<Workload> {
      return this.selectedWorkloads.filter(
        (workload: Workload) =>
          this.isWorkloadStoppable(workload) &&
          this.isWorkloadSupportSuspend(workload) &&
          this.isWorkloadProjectExists(workload),
      );
    },
    selectedRunnableWorkloads(): Array<Workload> {
      return this.selectedWorkloads.filter(
        (workload: Workload) =>
          workload.phase === Phase.Stopped &&
          this.isWorkloadSupportSuspend(workload) &&
          this.isWorkloadProjectExists(workload),
      );
    },
    isTourActive(): boolean {
      return this.guidedTourStore.isActive;
    },
    isWhiteLabel(): boolean {
      return this.appStore.isWhiteLabel;
    },
  },
  methods: {
    isWorkloadStoppable(workload: Workload): boolean {
      const nonStoppablePhases: Phase[] = [
        Phase.Creating,
        Phase.Stopping,
        Phase.Deleting,
        Phase.Stopped,
        Phase.Failed,
        Phase.Completed,
        Phase.Terminating,
      ];
      return !nonStoppablePhases.includes(workload?.phase);
    },
    isWorkloadSupportSuspend(workload: Workload): boolean {
      return !!workload.actionsSupport?.suspend;
    },
    isWorkloadProjectExists(workload: Workload): boolean {
      return !!this.authStore.orgUnits.find(
        (orgUnit: OrgUnit) => orgUnit.type === ResourceType.Project && orgUnit.id === workload.projectId,
      );
    },
    async loadIsOnlyDefaultNodePool(): Promise<void> {
      if (this.canReadNodePools) {
        const clusterId = this.clusterStore.selectedClusterId;
        try {
          this.isOnlyDefaultNodePool = await this.nodePoolStore.isOnlyDefaultNodePoolByClusterId(clusterId);
        } catch (e) {
          console.log(`Failed to get node pools for cluster ${clusterId}`, e);
          this.hideNodePoolsColumn = true;
        }
      }
    },
    async loadWorkloadsCount(): Promise<void> {
      try {
        this.setLoading(true);
        const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(this.filterBy, undefined, true);
        this.filterBy.rowsNumber = await workloadService.getWorkloadsCount(
          filters.filterBy,
          this.isDeletedViewType,
          filters.search,
        );
      } catch (e) {
        this.handleLoadWorkloadsError(e);
      } finally {
        this.finalizeWorkloadLoading();
      }
    },
    async loadWorkloads(): Promise<void> {
      try {
        this.setLoading(true);
        await this.workloadStore.loadWorkloads(this.filterBy, this.isDeletedViewType);
        // Making sure that the selected workloads are still in the list
        this.validateSelectedWorkloads();
        this.loadingError = false;
        // Check for cross-page drawer navigation after workloads are loaded
        this.checkPendingDrawerNavigation();
      } catch (e: unknown) {
        this.handleLoadWorkloadsError(e);
      } finally {
        this.finalizeWorkloadLoading();
      }
    },
    handleLoadWorkloadsError(error: unknown): void {
      console.error(error, EWorkloadErrorMessage.FailedToLoadWorkloads);
      this.loadingError = true;
      this.$q.notify(alertUtil.getError(EWorkloadErrorMessage.FailedToLoadWorkloads));
    },
    finalizeWorkloadLoading(): void {
      this.setLoading(false);
    },
    setLoading(isLoading: boolean): void {
      this.loading = isLoading;
    },
    async deleteWorkloads(): Promise<void> {
      if (this.selectedWorkloads.length === 0) return;
      if (this.selectedWorkloads.length === 1) {
        await this.deleteSingleWorkload();
        return;
      }
      if (this.selectedWorkloads.length > 1) {
        await this.deleteMultipleWorkloads();
        return;
      }
    },
    async deleteMultipleWorkloads(): Promise<void> {
      try {
        this.deleting = true;
        this.setLoading(true);
        const res: BatchResponse = await this.workloadStore.deleteWorkloads(
          this.deletableWorkloads.map((workload: Workload) => workload.id),
        );
        if (this.deletableWorkloads.map((w) => w.id).includes(this.lastCreatedWorkload?.id || "")) {
          if (res.succeeded?.includes(this.lastCreatedWorkload?.id || "")) {
            this.workloadStore.removeWorkloadAdded();
          }
        }

        if (res.failed?.length && res.failed[0].id) {
          const firstFailedName = this.workloadStore.allWorkloadMapById.get(res.failed[0].id)?.name;
          let errMsg = "";
          if (firstFailedName) {
            errMsg = `Workload ${firstFailedName}`;
            if (res.failed.length > 1) {
              errMsg += ` +${res.failed.length - 1} couldn't be deleted`;
            } else {
              errMsg += " couldn't be deleted";
            }
          } else {
            errMsg = `${res.failed.length} workloads couldn't be deleted`;
          }
          // TODO - add error message
          this.$q.notify(alertUtil.getError(errMsg));
        } else if (res.succeeded?.length && res.succeeded[0]) {
          const firstSucceededName = this.workloadStore.allWorkloadMapById.get(res.succeeded[0])?.name;
          let successMsg = "";
          if (firstSucceededName) {
            successMsg = `Workload ${firstSucceededName}`;
            if (res.succeeded.length > 1) {
              successMsg += ` +${res.succeeded.length - 1} are being deleted`;
            } else {
              successMsg += " is being deleted";
            }
          } else {
            successMsg = `${res.succeeded.length} workloads are being deleted`;
          }
          this.$q.notify(alertUtil.getSuccess(successMsg));
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: EWorkloadErrorMessage.FailedToDeleteWorkloads,
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.isDeleteModalOpen = false;
        this.resetSelectedRows();
        this.deleting = false;
        this.setLoading(false);
      }
    },
    async deleteSingleWorkload(): Promise<void> {
      if (this.selectedWorkload === null) return;

      const workload: Workload | null = this.selectedWorkload;

      try {
        this.deleting = true;
        this.setLoading(true);
        await this.workloadStore.deleteWorkload(this.activeClusterFilterId, workload);
        if (this.lastCreatedWorkload?.id === workload.id) {
          this.workloadStore.removeWorkloadAdded();
        }
        this.$q.notify(alertUtil.getSuccess(`Workload ${workload.name} is being deleted`));
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: EWorkloadErrorMessage.FailedToDeleteWorkloads,
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.isDeleteModalOpen = false;
        this.resetSelectedRows();
        this.deleting = false;
        this.setLoading(false);
      }
    },
    handleWorkloadAction(action: EWorkloadAction): void {
      switch (action) {
        case EWorkloadAction.run:
          this.runWorkloads();
          break;
        case EWorkloadAction.clone:
          this.cloneWorkload();
          break;
        case EWorkloadAction.stop:
          this.stopWorkloads();
          break;
        case EWorkloadAction.delete:
          this.openDeleteModal();
          break;
        case EWorkloadAction.update:
          this.updateWorkload();
          break;
        default:
          console.error(`Unhandled workload action: ${action}`);
      }
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    openStopModal(): void {
      this.isStopModalOpen = true;
    },
    stopRefreshWorkloads(): void {
      intervalUtil.stopInterval(EIntervalLabels.WorkloadIndex);
      intervalUtil.stopInterval(EIntervalLabels.WorkloadIndexCount);
    },
    startRefreshWorkloads(): void {
      intervalUtil.startInterval(EIntervalLabels.WorkloadIndex, this.loadWorkloads);
      intervalUtil.startInterval(EIntervalLabels.WorkloadIndexCount, this.loadWorkloadsCount);
    },
    clearFilters(): void {
      this.updateFilterBy(
        {
          ...this.filterBy,
          advancedFilters: [],
          searchTerm: "",
        },
        true,
      );
    },
    resetSelectedRows(): void {
      this.selectedWorkloads = [];
    },
    loadFilters(): void {
      const tableFilterName = this.isDeletedViewType ? ETableFilters.WORKLOAD_DELETED : ETableFilters.WORKLOAD;
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(WorkloadSortFilterFields.Name, this.columns);
      const filterBy: IFilterBy = filterService.loadFilters(window.location, tableFilterName, defaultFilters);
      this.updateFilterBy({ ...filterBy }, false);
    },
    async loadProjectsCount(): Promise<void> {
      try {
        this.projectsCount = await orgUnitService.countProjects({
          filterBy: [filterUtil.getEqualsFilterString(ProjectFilterSortFields.ClusterId, this.activeClusterFilterId)],
        });
      } catch (e) {
        console.error(e);
      }
    },
    async loadWorkloadCategories(): Promise<void> {
      if (this.workloadStore.workloadCategories.length > 0) return;
      try {
        await this.workloadStore.loadWorkloadCategories();
      } catch (error: unknown) {
        console.error(error);
      }
    },
    async updateFilterBy(filterBy: IFilterBy, removeTopRow = true): Promise<void> {
      this.filterBy = { ...filterBy, clusterUuid: this.activeClusterFilterId };

      if (!this.isDeletedViewType) {
        if (removeTopRow) {
          //no top row for deleted view
          this.workloadStore.removeWorkloadAdded();
        }
        //phase filter is hidden for deleted view
        this.adjustWorkloadPhaseFilter();
      }

      filterService.saveFilters(this.tableFilterName, this.filterBy);
      this.isInitialLoader = true;
      await Promise.all([this.loadWorkloads(), this.loadWorkloadsCount()]);
      // removing the selected workloads who don't exists in the current list
      this.validateSelectedWorkloads();
      this.isInitialLoader = false;
    },
    validateSelectedWorkloads(): void {
      const workloadsMapById = this.workloads.reduce((acc: Record<string, Workload>, workload: Workload) => {
        acc[workload.id] = workload;
        return acc;
      }, {});
      this.selectedWorkloads = this.selectedWorkloads.filter((workload: Workload) => workloadsMapById[workload.id]);
    },
    adjustWorkloadPhaseFilter(): void {
      const advancedFilters = this.filterBy.advancedFilters;
      if (advancedFilters) {
        this.filterBy.advancedFilters = workloadUtil.adjustWorkloadPhaseFilter(advancedFilters);
      }
    },
    getRowKey(row: Workload): string {
      return row.id;
    },
    getWorkloadIcon(workload: Workload): string {
      return workloadUtil.getWorkloadIcon(workload);
    },
    async runWorkloads(): Promise<void> {
      if (this.selectedWorkloads.length === 0) return;

      if (this.selectedWorkloads.length > 1) {
        if (this.selectedWorkloads.length !== this.selectedRunnableWorkloads.length) {
          this.isRunModalOpen = true;
        } else {
          this.runMultipleWorkloads();
        }
        return;
      }

      if (this.selectedWorkloads.length === 1) {
        await this.runSingleWorkload();
        return;
      }
    },
    async runSingleWorkload(): Promise<void> {
      if (this.selectedWorkload === null) return;
      this.setLoading(true);
      try {
        this.workloadStore.setWorkloadPhase(this.selectedWorkload, Phase.Resuming);

        await workloadService.activateWorkloadV2(this.selectedWorkload);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: EWorkloadErrorMessage.FailedToRunWorkload,
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        await this.workloadStore.loadWorkloads(this.filterBy);
        this.setLoading(false);
      }
    },
    async runMultipleWorkloads(): Promise<void> {
      this.setLoading(true);
      try {
        const resp = await workloadService.activateMultipleWorkloads(
          this.selectedRunnableWorkloads.map((workload: Workload) => workload.id),
        );

        resp.succeeded?.forEach((workloadId: string) => {
          const workload = this.workloadStore.allWorkloadMapById.get(workloadId);
          if (workload) {
            this.workloadStore.setWorkloadPhase(workload, Phase.Resuming);
          }
        });

        if (resp.failed?.length && resp.failed[0].id) {
          const firstFailedName = this.workloadStore.allWorkloadMapById.get(resp.failed[0].id)?.name;
          let errMsg = "";
          if (firstFailedName) {
            errMsg = `Workload ${firstFailedName}`;
            if (resp.failed.length > 1) {
              errMsg += ` +${resp.failed.length - 1} couldn't be running`;
            } else {
              errMsg += " couldn't be running";
            }
          } else {
            errMsg = `${resp.failed.length} workloads couldn't be running`;
          }
          this.$q.notify(alertUtil.getError(errMsg));
        } else if (resp.succeeded?.length && resp.succeeded[0]) {
          const firstSucceededName = this.workloadStore.allWorkloadMapById.get(resp.succeeded[0])?.name;
          let successMsg = "";
          if (firstSucceededName) {
            successMsg = `Workload ${firstSucceededName}`;
            if (resp.succeeded.length > 1) {
              successMsg += ` +${resp.succeeded.length - 1} running`;
            } else {
              successMsg += " running";
            }
          } else {
            successMsg = `${resp.succeeded.length} workloads running`;
          }
          this.$q.notify(alertUtil.getSuccess(successMsg));
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: EWorkloadErrorMessage.FailedToRunWorkload,
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.isRunModalOpen = false;
        this.resetSelectedRows();
        this.setLoading(false);
      }
    },
    async stopWorkloads(): Promise<void> {
      if (this.selectedWorkloads.length === 0) return;

      if (this.selectedWorkloads.length > 1) {
        if (this.selectedWorkloads.length !== this.selectedStoppableWorkloads.length) {
          this.isStopModalOpen = true;
          return;
        } else {
          this.stopMultipleWorkloads();
        }
      }

      if (this.selectedWorkloads.length === 1) {
        await this.stopSingleWorkload();
        return;
      }
    },
    async stopSingleWorkload(): Promise<void> {
      if (this.selectedWorkload === null) return;
      const prevPhase: Phase = this.selectedWorkload?.phase;
      this.setLoading(true);
      try {
        this.workloadStore.setWorkloadPhase(this.selectedWorkload, Phase.Stopping);

        await workloadService.stopWorkloadV2(this.selectedWorkload);
      } catch (error: unknown) {
        this.workloadStore.setWorkloadPhase(this.selectedWorkload, prevPhase);
        const errorAlert = new ErrorAlert({
          generalMessage: EWorkloadErrorMessage.FailedToStopWorkloads,
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        await this.workloadStore.loadWorkloads(this.filterBy);
        this.setLoading(false);
      }
    },
    async stopMultipleWorkloads(): Promise<void> {
      this.setLoading(true);
      this.stopping = true;
      try {
        const resp = await workloadService.stopMultipleWorkloads(
          this.selectedStoppableWorkloads.map((workload: Workload) => workload.id),
        );

        resp.succeeded?.forEach((workloadId: string) => {
          const workload = this.workloadStore.allWorkloadMapById.get(workloadId);
          if (workload) {
            this.workloadStore.setWorkloadPhase(workload, Phase.Stopping);
          }
        });

        if (resp.failed?.length && resp.failed[0].id) {
          const firstFailedName = this.workloadStore.allWorkloadMapById.get(resp.failed[0].id)?.name;
          let errMsg = "";
          if (firstFailedName) {
            errMsg = `Workload ${firstFailedName}`;
            if (resp.failed.length > 1) {
              errMsg += ` +${resp.failed.length - 1} couldn't be stopped`;
            } else {
              errMsg += " couldn't be stopped";
            }
          } else {
            errMsg = `${resp.failed.length} workloads couldn't be stopped`;
          }
          this.$q.notify(alertUtil.getError(errMsg));
        } else if (resp.succeeded?.length && resp.succeeded[0]) {
          const firstSucceededName = this.workloadStore.allWorkloadMapById.get(resp.succeeded[0])?.name;
          let successMsg = "";
          if (firstSucceededName) {
            successMsg = `Workload ${firstSucceededName}`;
            if (resp.succeeded.length > 1) {
              successMsg += ` +${resp.succeeded.length - 1} are being stopped`;
            } else {
              successMsg += " is being stopped";
            }
          } else {
            successMsg = `${resp.succeeded.length} workloads are being stopped`;
          }
          this.$q.notify(alertUtil.getSuccess(successMsg));
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: EWorkloadErrorMessage.FailedToStopWorkloads,
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.isStopModalOpen = false;
        this.resetSelectedRows();
        this.stopping = false;
        this.setLoading(false);
      }
    },
    closeDataSourceModal(): void {
      this.isDataSourceModalOpen = false;
      this.resetDataSourceModalOptions();
    },
    resetDataSourceModalOptions(): void {
      this.dataSourceModalOptions = { dataSourceIds: [], header: "" };
    },
    cloneWorkload(): void {
      if (this.selectedWorkload === null) return;

      switch (this.selectedWorkload?.source) {
        case Source.ControlPlane:
          this.cloneControlPlaneSourceWorkload(this.selectedWorkload);
          break;
        case Source.Cli:
          this.cloneCLIWorkload();
          break;
        default:
          console.error("workload source is neither control plane or cli");
      }
    },
    updateWorkload(): void {
      if (this.selectedWorkload === null) return;

      switch (this.selectedWorkload?.source) {
        case Source.ControlPlane:
          this.updateHybridWorkload(this.selectedWorkload);
          break;
        case Source.Cli:
          this.updateCLIWorkload(this.selectedWorkload);
          break;
        default:
          console.error("workload source is neither control plane or cli");
      }
    },
    updateCLIWorkload(workload: Workload): void {
      const isFlatWorkload = !workload?.environments?.[0]?.id;
      const supportsHybridCLI =
        [EWorkloadType.Inference].includes(workload?.type as EWorkloadType) &&
        !!workload?.fullySaasWorkload &&
        isFlatWorkload &&
        this.settingStore.isHybridSubmissionEnabled;

      if (supportsHybridCLI) this.updateHybridWorkload(workload);
      else this.isCopyCliModalOpen = true;
    },
    cloneCLIWorkload(): void {
      this.isCopyCliModalOpen = true;

      // Due to a mismatch between the API and the UI,
      // we currently want to prevent users from opening the UI
      // while a workload created via the CLI is being copied.

      // const isFlatWorkload = !workload?.environments?.[0]?.id;
      // const supportsHybridCLI =
      //   [EWorkloadType.Workspace, EWorkloadType.Training, EWorkloadType.Distributed, EWorkloadType.Inference].includes(
      //     workload?.type as EWorkloadType,
      //   ) &&
      //   !!workload?.fullySaasWorkload &&
      //   isFlatWorkload &&
      //   this.settingStore.isHybridSubmissionEnabled;

      // if (supportsHybridCLI) this.cloneHybridWorkload(workload);
      // else this.isCopyCliModalOpen = true;
    },
    cloneControlPlaneSourceWorkload(workload: Workload): void {
      const isFlatWorkload = !workload?.environments?.[0]?.id;
      const isHybridSubmission = isFlatWorkload && this.settingStore.isHybridSubmissionEnabled;
      const isFlatInference =
        (isFlatWorkload && workload.type === EWorkloadType.Inference) ||
        workload.type === EWorkloadType.DistributedInference;

      if (isHybridSubmission || isFlatInference) {
        this.cloneHybridWorkload(workload);
      } else {
        this.cloneAssetBasedWorkload(workload);
      }
    },
    updateHybridWorkload(workload: Workload): void {
      const isFlatWorkload = !workload?.environments?.[0]?.id;
      if (!isFlatWorkload || !workload) return;

      const query: LocationQueryRaw = {
        updateWorkloadId: workload.id,
      };

      this.$router.push({
        name: INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_UPDATE,
        query,
      });
    },
    cloneAssetBasedWorkload(workload: Workload): void {
      let routeName = "";

      switch (workload.type) {
        case EWorkloadType.Training:
          routeName = TRAINING_ROUTE_NAMES.TRAINING_ASSETS_EDIT;
          break;
        case EWorkloadType.Distributed:
        case EWorkloadType.DistributedTraining:
        case EWorkloadType.TFJob:
          routeName = TRAINING_ROUTE_NAMES.DISTRIBUTED_ASSETS_EDIT;
          break;
        case EWorkloadType.Workspace:
          routeName = WORKSPACE_ROUTE_NAMES.WORKSPACE_ASSETS_EDIT;
          break;
        case EWorkloadType.Inference:
          routeName = INFERENCE_ROUTE_NAMES.INFERENCE_ASSETS_EDIT;
          break;
        default:
          console.error(`Type ${workload.type} not found`);
      }

      if (!routeName) return;

      const query: LocationQueryRaw = {
        fromCopyId: workload?.id,
      };

      this.$router.push({
        name: routeName,
        query,
      });
    },
    cloneHybridWorkload(workload: Workload): void {
      const isFlatWorkload = !workload?.environments?.[0]?.id;
      const isHybridSubmission = isFlatWorkload && this.settingStore.isHybridSubmissionEnabled;
      const isFlatInference = isFlatWorkload && workload.type === EWorkloadType.Inference;
      if (!isHybridSubmission && !isFlatInference) return;

      let routeName = "";
      switch (workload.type) {
        case EWorkloadType.Training:
          routeName = TRAINING_ROUTE_NAMES.TRAINING_HYBRID;
          break;
        case EWorkloadType.Distributed:
          routeName = TRAINING_ROUTE_NAMES.DISTRIBUTED_HYBRID;
          break;
        case EWorkloadType.Workspace:
          routeName = WORKSPACE_ROUTE_NAMES.WORKSPACE_HYBRID;
          break;
        case EWorkloadType.Inference:
          routeName = INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID;
          break;
        case EWorkloadType.DistributedInference:
          routeName = INFERENCE_ROUTE_NAMES.INFERENCE_DISTRIBUTED_EDIT;
          break;
        default:
          console.error(`Type ${workload.type} not found`);
      }

      if (!routeName) return;

      const query: LocationQueryRaw = {
        fromCopyId: workload?.id,
        projectId: workload?.projectId,
        clusterId: workload?.clusterId,
      };

      this.$router.push({
        name: routeName,
        query,
      });
    },
    checkConnectivityStatus(): void {
      const cluster = this.clusterStore.getClusterById(this.activeClusterFilterId);
      if (!cluster) {
        this.$q.notify(connectionAlertUtil.getAlert(workloadUtil.getMissingDomainMessage()));
        return;
      }

      const canSubmitWorkload: ICanSubmitWorkload | null = workloadService.canSubmitWorkload(cluster);
      this.isClusterConnected = canSubmitWorkload === null;
      if (this.isClusterConnected) return;

      if (!cluster.domain) {
        this.$q.notify(connectionAlertUtil.getAlert(workloadUtil.getMissingDomainMessage()));
      } else {
        this.$q.notify(connectionAlertUtil.getAlert(workloadUtil.getConnectivityMessage()));
      }
    },
    /*** Modals ***/
    async displayPodsModal(workload: Workload): Promise<void> {
      this.podsModalOptions.columns = this.podsModalColumns;
      this.podsModalOptions.header = `Pods Associated with Workload ${workload.name}`;
      this.isPodsModalOpen = true;
      this.podsModalOptions.clusterId = this.activeClusterFilterId;
      this.podsModalOptions.getPods = async (filters: IPaginationFilter) => {
        return await workloadService.getWorkloadPods(
          workload.id,
          false,
          filters.offset,
          filters.limit,
          filters.search,
          false,
          undefined,
          filters.sortOrder,
          filters.sortBy as PodSortFilterFields,
        );
      };
      this.podsModalOptions.getPodsCount = async (filters: IPaginationFilter) => {
        return await workloadService.countPods(
          false,
          [filterUtil.getEqualsFilterString(PodSortFilterFields.WorkloadId, workload.id)],
          filters.search,
        );
      };
    },
    closePodsModal(): void {
      this.isPodsModalOpen = false;
      this.resetPodsModalOptions();
    },
    handleOpenDrawerTab(payload: IOpenWorkloadDrawerTabPayload): void {
      const workload = this.workloadStore.allWorkloads.find((w) => w.id === payload.workloadId);
      if (!workload) return;

      this.selectedWorkloads = [workload];
      this.openPageActionsDrawer();

      // Delay setting the navigation payload to ensure drawer is open and component is mounted
      this.$nextTick(() => {
        this.drawerNavigationPayload = { tab: payload.tab, podName: payload.podName };
      });
    },
    clearDrawerNavigationPayload(): void {
      this.drawerNavigationPayload = null;
    },
    openPageActionsDrawer(): void {
      const pageActions = this.$refs.pageActions as InstanceType<typeof RunaiPageActions> | undefined;
      pageActions?.openDrawer();
    },
    checkPendingDrawerNavigation(): void {
      if (!this.pendingDrawerNavigation) return;

      const params = this.getDrawerParamsFromUrl();
      if (!params) {
        this.pendingDrawerNavigation = false;
        return;
      }

      const workload = this.workloadStore.allWorkloads.find((w) => w.id === params.workloadId);
      if (!workload) {
        this.pendingDrawerNavigation = false;
        this.clearDrawerParams();
        return;
      }

      this.selectedWorkloads = [workload];
      this.openPageActionsDrawer();
      this.pendingDrawerNavigation = false;
      this.clearDrawerParams();

      // Delay setting the navigation payload to ensure drawer is open and component is mounted
      this.$nextTick(() => {
        this.drawerNavigationPayload = { tab: params.tab, podName: params.podName };
      });
    },
    resetPodsModalOptions(): void {
      this.podsModalOptions = {
        header: "",
        displayRedirection: true,
        columns: this.podsModalColumns,
        getPods: async () => [],
        getPodsCount: async () => 0,
      };
    },
    displayConnectionsModal(workload: Workload): void {
      const cluster = this.clusterStore.getClusterById(this.activeClusterFilterId);
      if (!cluster) {
        this.$q.notify(connectionAlertUtil.getAlert(workloadUtil.getMissingDomainMessage()));
        return;
      }

      const connections: Connection[] = workloadUtil.getWorkloadConnectionsForGrid(workload, cluster.domain || "");

      if (!connections.length) return;

      this.connectionsModalOptions.connections = connections;
      this.connectionsModalOptions.header = `Connections Associated with Workload ${workload.name}`;
      this.connectionsModalOptions.workloadDetails = {
        currentUser: this.authStore.userEmail,
        isRunning: workload.phase === Phase.Running,
        name: workload.name,
        type: workload.type,
        projectName: workload.projectName,
      };
      this.isConnectionsModalOpen = true;
    },
    closeConnectionsModal(): void {
      this.isConnectionsModalOpen = false;
      this.connectionsModalOptions = { connections: [], header: "" };
    },
    displayDataSourcesModal(workload: Workload): void {
      if (!workload.datasources?.length) return;
      this.dataSourceModalOptions.header = `Data Sources Associated with Workload ${workload.name}`;
      this.dataSourceModalOptions.dataSourceIds = workload.datasources.map((dataSource: Datasource) => dataSource.id);
      this.isDataSourceModalOpen = true;
    },
    async displayNodePoolsModal(workload: Workload): Promise<void> {
      this.nodePoolsModalOptions.loading = true;
      this.isNodePoolsModalOpen = true;
      this.nodePoolsModalOptions.header = `Project ${workload.projectName}`;
      this.projectNameForNodePoolsModal = workload.projectName;
      this.clusterIdForNodePoolsModal = workload.clusterId;
      try {
        const project = await orgUnitService.getProject(workload.projectId);
        this.nodePoolsModalOptions.resources = project.resources;
        this.nodePoolsModalOptions.nodePoolQuotaStatuses = project?.status?.nodePoolQuotaStatuses || [];
        this.nodePoolsModalOptions.defaultNodePools = project.effective?.defaultNodePools || [];
      } catch (error: unknown) {
        this.isNodePoolsModalOpen = false;
        this.handleNodePoolsModalError(error);
      } finally {
        this.nodePoolsModalOptions.loading = false;
      }
    },
    handleNodePoolsModalError(error: unknown): void {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedLoadingMessage("project"),
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
    closeNodePoolsModal(): void {
      this.isNodePoolsModalOpen = false;
      this.resetNodePoolsModalOptions();
    },
    resetNodePoolsModalOptions(): void {
      this.nodePoolsModalOptions = {
        ...NODE_POOLS_MODAL_DEFAULT_CONFIG,
      };
    },
    redirectToProjects(): void {
      this.$router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
        query: {
          [ProjectFilterSortFields.Name]: this.projectNameForNodePoolsModal,
          [ProjectFilterSortFields.ClusterId]: this.activeClusterFilterId,
        },
      });
    },
    /** Workload priorities **/
    async loadWorkloadPriorities(clusterId: string): Promise<void> {
      try {
        this.workloadPriorities = await workloadService.getWorkloadClusterPriorities(clusterId);
      } catch (e: unknown) {
        new ErrorAlert({
          generalMessage: EWorkloadErrorMessage.FailedToLoadWorkloadPriorities,
        });
        this.workloadPriorities = [];
      }
    },
    triggerWorkloadTourIfNeeded(): void {
      if (!this.guidedTourStore.workloadTourAllowed || this.isWhiteLabel) return;

      const seenTourAlready = storageUtil.get(STORAGE_KEYS.WORKLOAD_TOUR_SEEN);
      if (seenTourAlready) return;

      this.guidedTourStore.openWorkloadsGuidedTour();
      storageUtil.save(STORAGE_KEYS.WORKLOAD_TOUR_SEEN, true);
    },
  },
  watch: {
    isPodsModalOpen(isOpen: boolean) {
      if (!isOpen) {
        this.podsModalOptions.getPods = async () => [];
      }
    },
    isDeletedViewType() {
      this.resetSelectedRows();
      this.workloadStore.removeWorkloadAdded();
      this.loadFilters();
      this.appStore.setPageLoading(false);
    },
    activeClusterFilterId() {
      this.loadProjectsCount();
      this.loadIsOnlyDefaultNodePool();
      this.checkConnectivityStatus();
    },
    "filterBy.clusterUuid"(newClusterUuid?: string, oldClusterUuid?: string) {
      if (oldClusterUuid) {
        filterService.removeAdvancedFilter(this.filterBy, WorkloadSortFilterFields.Priority, ETableFilters.WORKLOAD);
      }
      if (newClusterUuid) {
        this.loadWorkloadPriorities(newClusterUuid);
      }
    },
    isTourActive: {
      handler(newVal: boolean, oldVal: boolean) {
        // if the tour is triggered we have to remove any selected workloads to ensure the NEW WORKLOAD button is visible
        if (newVal && newVal !== oldVal) {
          this.selectedWorkloads = [];
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    this.stopRefreshWorkloads();
    this.workloadStore.removeWorkloadAdded();
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default WorkloadIndex as typeof WorkloadIndex & DefineComponent;
</script>
