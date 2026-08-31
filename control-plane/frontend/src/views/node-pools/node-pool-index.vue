<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :columns="columns"
          :filters="filterBy"
          @filters-changed="updateFilters"
          @selected-actions-close="resetSelectedRows"
          primary-btn-label="New node pool"
          empty-message="Select a node pool to view its details"
          @create-entity="createNodePool"
          :selected-rows-amount="selectedRowsAmount"
          show-details-btn
          @drawer-toggled="isDrawerOpen = $event"
          @export-csv="exportTableAsCsv"
          advanced-filters
        >
          <template v-slot:drawer-content>
            <runai-drawer-content v-if="!selectedNodePool">
              <template v-slot:body>
                <runai-drawer-no-data
                  :messages="['Select a node pool to view its details.']"
                  icon-name="ph-bold ph-file"
                  icon-class="text-icon-default"
                />
              </template>
            </runai-drawer-content>
            <runai-drawer-content v-else-if="!isSelectedNodePoolReady">
              <template v-slot:body>
                <runai-drawer-no-data
                  icon-name="ph-bold ph-file"
                  icon-class="text-icon-default"
                  :messages="['This node pool doesn\'t have any details because it isn\'t ready.']"
                />
              </template>
            </runai-drawer-content>
            <node-pool-drawer-content v-else :node-pool="selectedNodePool" />
          </template>

          <template v-slot:selected-rows-actions>
            <div class="row items-center justify-start">
              <runai-tooltip-wrapper :display-tooltip="!canEditNodePool" :tooltip-text="editDisableTooltip">
                <runai-action-button
                  aid="node-pool-edit-btn"
                  btn-action="edit"
                  @click="editNodePool"
                  :disable="!canEditNodePool"
                  v-permission="{ resourceType: ResourceType.Nodepools, action: Action.Update }"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="!canDeleteNodePool" :tooltip-text="deleteDisableTooltip">
                <runai-action-button
                  aid="node-pool-delete-btn"
                  btn-action="delete"
                  @click="isDeleteModalOpen = true"
                  :disable="!canDeleteNodePool"
                  v-permission="{ resourceType: ResourceType.Nodepools, action: Action.Delete }"
                />
              </runai-tooltip-wrapper>
            </div>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="tableData"
          :columns="columns"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          :loading="loadingTableData"
          @update-filters="setFilterBy"
          @clear-filters="clearFilters"
          @workloads-clicked="displayWorkloadsModal"
          @nodes-clicked="openNodes"
          @topology-clicked="openTopologyModal"
          :top-row="lastCreatedEntity"
          :get-row-key="getRowKey"
          sticky-columns
          :bordered="false"
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25, 50, 100]"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedEntity"
              :filter-by="filterBy"
              entity-name="node-pool"
              ph-icon-name="ph-bold ph-cube"
              icon-size="44"
              :show-error="loadingError"
              @clear-filters="clearFilters"
              @create-new="isAddModalOpen = true"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
        <node-list-modal v-if="isNodesModalOpen" :options="nodesModalOptions" @close="isNodesModalOpen = false" />
        <runai-delete-modal
          v-if="isDeleteModalOpen && selectedNodePool"
          :entity-name="selectedNodePool.name"
          entity-type="node pool"
          custom-description="The pool's nodes and workloads will be moved to the default node pool automatically. This may affect the NVIDIA Run:ai scheduler decisions."
          :deleting="deleting"
          @cancel="isDeleteModalOpen = false"
          @delete="deleteNodepool"
        />
      </template>
    </runai-table-wrapper>
    <workload-list-modal
      v-if="isWorkloadsModalOpen"
      :modal-options="workloadsModalOptions"
      @close="isWorkloadsModalOpen = false"
    />
    <topology-labels-modal
      v-if="isTopologyModalOpen && topologyModalOptions"
      :modal-options="topologyModalOptions"
      @close="closeTopologyModal"
    />
  </section>
</template>

<script lang="ts">
import { computed, defineComponent, type ComputedRef, type DefineComponent } from "vue";

// stores
import { useNodePoolStore } from "@/stores/node-pool.store";

// models
import { nodePoolWorkloadListModalColumns } from "@/table-models/node-pool.table-model";
import { type ITableColumn } from "@/models/table.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";
import {
  CLUSTER_COLUMN_FILTER_NAME,
  EFilterOperator,
  type IFilterBy,
  type IPaginationFilter,
} from "@/models/filter.model";

// cmps
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { NodePoolDrawerContent } from "@/components/node-pools/node-pool-drawer-content";
import { RunaiDrawerContent } from "@/components/common/runai-drawer-content";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import { RunaiDeleteModal } from "@/components/common";
import { TopologyLabelsModal } from "@/components/node-pools/topology-labels-modal";
// constants
import { tableDefaultProps } from "@/common/table.constant";
import { useClusterStore } from "@/stores/cluster.store";
import { alertUtil } from "@/utils/alert.util";
import { tableUtil } from "@/utils/table.util";
import type { INode, INodeModalOptions } from "@/models/node.model";
import NodeListModal from "@/components/node/node-list-modal/node-list-modal.vue";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { usePermissionStore } from "@/stores/permissions.store";
import { EWorkloadModalEntity, type IWorkloadListModalOptions } from "@/models/workload.model";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import WorkloadListModal from "@/components/workload/workload-list-modal/workload-list-modal.vue";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { useAppStore } from "@/stores/app.store";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import {
  type INodePoolTable,
  nodePoolsDependentColumns,
  nodePoolIndexColumns,
} from "@/table-models/node-pool.table-model";
import { filterService } from "@/services/filter.service/filter.service";
import {
  NodepoolGPUNetworkAccelerationDetection,
  NodepoolPhase,
  NodepoolSortFilterFields,
  NodeTelemetryType,
} from "@/swagger-models/cluster-service-client";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import {
  MIN_CLUSTER_VERSION_FOR_MINIMUM_GUARANTEED_RUNTIME,
  MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY,
} from "@/common/version.constant";
import { filterUtil } from "@/utils/filter.util/filter.util";
import type { ITopologyLabelsModalOptions } from "@/table-models/topology.table-model";
import { useRoute } from "vue-router";

const DEFAULT_WORKLOADS_MODAL_OPTIONS: IWorkloadListModalOptions = {
  entityFilter: "",
  entityName: "",
  filterName: WorkloadSortFilterFields.Nodepool,
  entityType: EWorkloadModalEntity.Nodepool,
  columns: nodePoolWorkloadListModalColumns,
  clusterId: "",
};
const DEFAULT_NODES_MODAL_OPTIONS: INodeModalOptions = {
  nodePoolName: "",
  clusterId: "",
};
const NodePoolIndex = defineComponent({
  name: "node-pool-index",
  components: {
    WorkloadListModal,
    RunaiTooltipWrapper,
    RunaiTableNoData,
    RunaiTable,
    NodeListModal,
    RunaiTableWrapper,
    RunaiPageActions,
    RunaiActionButton,
    NodePoolDrawerContent,
    RunaiDrawerContent,
    RunaiDrawerNoData,
    RunaiDeleteModal,
    TopologyLabelsModal,
  },
  data() {
    return {
      isDeleteModalOpen: false as boolean,
      isAddModalOpen: false as boolean,
      isEditModalOpen: false as boolean,
      isWorkloadsModalOpen: false as boolean,
      isNodesModalOpen: false as boolean,
      isTopologyModalOpen: false as boolean,
      deleting: false as boolean,
      selectedRows: [] as Array<INodePoolTable>,
      loading: false as boolean,
      isInitialLoader: true as boolean,
      displayWorkloads: false as boolean,
      tableDefaultProps,
      isDrawerOpen: false as boolean,
      nodes: [] as Array<INode>,
      workloadsModalOptions: {
        ...DEFAULT_WORKLOADS_MODAL_OPTIONS,
      } as IWorkloadListModalOptions,
      nodesModalOptions: {
        ...DEFAULT_NODES_MODAL_OPTIONS,
      },
      topologyModalOptions: null as null | ITopologyLabelsModalOptions,
    };
  },
  setup() {
    const appStore = useAppStore();
    const clusterStore = useClusterStore();
    const nodePoolStore = useNodePoolStore();
    const permissionStore = usePermissionStore();
    const route = useRoute();

    const columns: ComputedRef<ITableColumn[]> = computed(() => {
      return nodePoolIndexColumns.filter((col: ITableColumn) => {
        if (nodePoolsDependentColumns.nodes.has(col.name)) {
          return permissionStore.hasPermission(ResourceType.Nodes, Action.Read);
        }
        if (nodePoolsDependentColumns.cluster.has(col.name)) {
          return clusterStore.isMultiCluster;
        }
        if (nodePoolsDependentColumns.minGuaranteedRuntime.has(col.name)) {
          return isClusterSupportsMinGuaranteedRuntime.value;
        }
        if (nodePoolsDependentColumns.gpuDevicePlacementStrategy.has(col.name)) {
          return isClusterSupportsGpuDevicePlacement.value;
        }
        return true;
      });
    });
    const adjustDetectionFilter = (filters: IPaginationFilter): void => {
      if (
        filters?.filterBy?.includes(
          filterUtil.getNotEqualsFilterString(
            NodepoolSortFilterFields.GpuNetworkAccelerationDetection,
            NodepoolGPUNetworkAccelerationDetection.Use,
          ),
        )
      ) {
        filters?.filterBy.push(
          filterUtil.getNotEqualsFilterString(
            NodepoolSortFilterFields.GpuNetworkAccelerationDetection,
            NodepoolGPUNetworkAccelerationDetection.DontUse,
          ),
        );
      }
    };
    const getFilters = (filterBy: IFilterBy): IPaginationFilter => {
      const filters = filterService.mapColumnsFilterToFilterParams(filterBy, [
        ...Object.values(NodepoolSortFilterFields),
        CLUSTER_COLUMN_FILTER_NAME,
      ]);
      adjustDetectionFilter(filters);
      return filters;
    };

    const getNodePools = async (filterBy: IFilterBy): Promise<INodePoolTable[]> => {
      const filters: IPaginationFilter = getFilters(filterBy);
      const nodePools = await clusterService.getNodepools(
        filters.filterBy || [],
        filters.sortBy as NodepoolSortFilterFields | undefined,
        filters.sortOrder,
        filters?.offset,
        filters?.limit,
      );
      return nodePools;
    };

    const countNodePools = async (filterBy: IFilterBy): Promise<void> => {
      const filters: IPaginationFilter = getFilters(filterBy);
      const count = await clusterService.countNodepools(filters.filterBy || []);
      updateFilterByKey("rowsNumber", count);
    };

    const getNodePoolsWithAdditionalData = async (filterBy: IFilterBy, softLoad = false): Promise<INodePoolTable[]> => {
      let nodePools = tableData.value;
      if (!softLoad) {
        nodePools = await getNodePools(filterBy);
      }
      const response = await clusterService.getNodeTelemetries(
        [
          NodeTelemetryType.ReadyGpus,
          NodeTelemetryType.TotalGpus,
          NodeTelemetryType.TotalGpuMemoryBytes,
          NodeTelemetryType.AllocatedGpus,
          NodeTelemetryType.TotalCpuCores,
          NodeTelemetryType.TotalCpuMemoryBytes,
          NodeTelemetryType.AllocatedCpuCores,
          NodeTelemetryType.AllocatedCpuMemoryBytes,
        ],
        filterBy.clusterUuid,
        undefined,
        [ENodeTelemetryGroupBy.Nodepool, ENodeTelemetryGroupBy.ClusterId],
      );
      return nodePoolUtil.enrichNodePoolsWithTelemetry(nodePools, response);
    };

    const setClusterFilter = () => {
      const filteredClusterId = route.query[NodepoolSortFilterFields.ClusterId] as string;
      if (filteredClusterId) {
        clusterStore.setSelectedClusterById(filteredClusterId);
        filterBy.value.clusterUuid = filteredClusterId;
      }
    };

    const isClusterSupportsMinGuaranteedRuntime = computed(() =>
      clusterStore.isClusterVersionSufficient(clusterId.value, MIN_CLUSTER_VERSION_FOR_MINIMUM_GUARANTEED_RUNTIME),
    );

    const isClusterSupportsGpuDevicePlacement = computed(() =>
      clusterStore.isClusterVersionSufficient(clusterId.value, MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY),
    );

    const {
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      clearLastCreated,
      initTableFilter,
      updateFilterByKey,
      removeRow,
      clusterId,
    } = useTableDataAdvancedFilters<INodePoolTable>(
      getNodePools,
      columns,
      EAdvancedIndexPages.NODE_POOL,
      clusterService.getNodepool,
      getNodePoolsWithAdditionalData,
    );

    setClusterFilter();
    initTableFilter();
    appStore.setPageLoading(false);
    countNodePools(filterBy.value);
    return {
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      clearLastCreated,
      columns,
      appStore,
      nodePoolStore,
      permissionStore,
      clusterStore,
      countNodePools,
      clusterId,
      removeRow,
    };
  },
  computed: {
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    selectedNodePool(): INodePoolTable | null {
      if (this.selectedRows.length > 0) {
        return this.selectedRows[0];
      }
      return null;
    },
    isSelectedNodePoolReady(): boolean {
      return this.selectedNodePool?.phase === NodepoolPhase.Ready;
    },
    canEditNodePool(): boolean {
      return this.selectedNodePool?.phase !== NodepoolPhase.Deleting;
    },
    canDeleteNodePool(): boolean {
      return !this.selectedNodePool?.isDefault && this.selectedNodePool?.phase !== NodepoolPhase.Deleting;
    },
    hasDeletePermission(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Nodepools, Action.Delete);
    },
    hasUpdatePermission(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Nodepools, Action.Update);
    },
    editDisableTooltip(): string {
      if (!this.hasUpdatePermission) {
        return "You don't have permission to perform this action.";
      }
      return this.selectedNodePool?.isDefault
        ? "Default node pool cannot be modified"
        : "The selected node pool is being deleted";
    },
    deleteDisableTooltip(): string {
      if (!this.hasDeletePermission) {
        return "You don't have permission to perform this action.";
      }
      return this.selectedNodePool?.isDefault
        ? "Default node pool cannot be deleted"
        : "The selected node pool is being deleted";
    },
    isMultiCluster(): boolean {
      return this.clusterStore.isMultiCluster;
    },
  },
  methods: {
    updateFilters(filters: IFilterBy): void {
      this.setFilterBy(filters);
      this.countNodePools(filters);
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    createNodePool(): void {
      const clusterId = this.isMultiCluster ? this.filterBy.clusterUuid : this.clusterStore.selectedClusterId;
      this.$router.push({ name: NODE_POOL_ROUTE_NAMES.NODE_POOL_NEW, query: { clusterId } });
    },
    editNodePool(): void {
      this.$router.push({ name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, params: { id: this.selectedNodePool?.id } });
    },
    async deleteNodepool(): Promise<void> {
      if (!this.selectedNodePool) return;
      try {
        this.deleting = true;
        const nodePoolId: string = this.selectedNodePool?.id;
        await clusterService.deleteNodepool(nodePoolId);
        this.deleting = false;
        this.isDeleteModalOpen = false;

        this.removeRow(this.selectedNodePool);
        this.$q.notify(alertUtil.getSuccess(`Node pool ${this.selectedNodePool.name} deleted`));
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedDeleteMessage("node pool"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.resetSelectedRows();
      }
    },
    getRowKey(row: INodePoolTable): string {
      return row.id;
    },
    displayWorkloadsModal(nodePool: INodePoolTable): void {
      this.isWorkloadsModalOpen = true;
      this.workloadsModalOptions.clusterId = this.clusterId;
      this.workloadsModalOptions.entityFilter = `${WorkloadSortFilterFields.Nodepool}${EFilterOperator.Equals}${nodePool.name}`;
      this.workloadsModalOptions.entityName = nodePool.name;
    },
    openNodes(nodePool: INodePoolTable): void {
      this.isNodesModalOpen = true;
      this.nodesModalOptions.nodePoolName = nodePool.name;
      this.nodesModalOptions.clusterId = nodePool.clusterId;
    },
    openTopologyModal(nodePool: INodePoolTable): void {
      if (!nodePool.networkTopologyId) return;

      this.isTopologyModalOpen = true;
      this.topologyModalOptions = {
        topologyId: nodePool.networkTopologyId,
        topologyName: nodePool.networkTopologyName || "",
      };
    },
    closeTopologyModal(): void {
      this.isTopologyModalOpen = false;
      this.topologyModalOptions = null;
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));

      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.NodePool, this.tableData, columns);
    },
    resetWorkloadsModalOptions(): void {
      this.workloadsModalOptions = { ...DEFAULT_WORKLOADS_MODAL_OPTIONS };
    },
  },
  watch: {
    isWorkloadsModalOpen(isOpen: boolean): void {
      if (!isOpen) {
        this.resetWorkloadsModalOptions();
      }
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default NodePoolIndex as typeof NodePoolIndex & DefineComponent;
</script>

<style lang="scss" scoped>
.graph-header {
  display: flex;
  width: 100%;

  > div {
    flex: 50%;
  }
  .graph-header-label {
    color: $black-54;
    margin-right: 5px;
  }

  .graph-header-value {
    color: $black-87;
  }
}
</style>
