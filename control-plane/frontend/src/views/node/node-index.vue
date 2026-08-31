<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-table-wrapper :filters-object="filterBy" sticky>
    <template v-slot:actions>
      <runai-page-actions
        :columns="columns"
        :filters="filterBy"
        @filters-changed="updateFilters"
        disable-actions
        show-details-btn
        @drawer-resized="drawerWidth = $event"
        @drawer-toggled="isDrawerOpen = $event"
        @export-csv="exportTableAsCsv"
        advanced-filters
        :show-primary-btn="false"
      >
        <template v-slot:drawer-content>
          <runai-drawer-content v-if="!selectedRow">
            <template v-slot:body>
              <runai-drawer-no-data
                :messages="['Select a node to view its details.']"
                icon-name="ph-bold ph-file"
                icon-class="text-icon-default"
              />
            </template>
          </runai-drawer-content>
          <node-drawer-content v-else :filter-by="filterBy" :cluster-id="clusterId" :node="selectedRow" />
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
        @update-filters="updateFilters"
        @clear-filters="clearFilters"
        @pods-clicked="displayPodsModal"
        @gpus-clicked="displayGpusModal"
        @numa-nodes-clicked="displayNumaNodesModal"
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
            entity-name="node"
            ph-icon-name="ph-bold ph-cube"
            icon-size="44"
            :show-error="loadingError"
            :show-create-btn="false"
            @clear-filters="clearFilters"
          >
          </runai-table-no-data>
        </template>
      </runai-table>
      <workload-pods-modal
        v-if="isPodsModalOpen"
        :modal-options="podsModalOptions"
        @close="closePodsModal"
        @workload-clicked="onPodWorkloadClicked"
        @project-clicked="onPodProjectClicked"
        @open-drawer-tab="handleOpenDrawerTab"
      />
      <gpus-modal v-if="nodeForGpusModal" :node="nodeForGpusModal" @close="nodeForGpusModal = null" />
      <numa-nodes-modal
        v-if="nodeForNumaNodesModal"
        :node="nodeForNumaNodesModal"
        @close="nodeForNumaNodesModal = null"
      />
    </template>
  </runai-table-wrapper>
</template>
<script lang="ts">
import { computed, defineComponent, ref, type ComputedRef } from "vue";
import { useAppStore } from "@/stores/app.store";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { ETableExportCsvFilesNames, type ITableColumn } from "@/models/table.model";
import { nodeDependentColumns, nodeIndexColumns } from "@/table-models/node.table-model";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiTable } from "../../components/common/runai-table/";
import { type Node, NodeFilterSortFields, NodeTelemetryType } from "@/swagger-models/cluster-service-client";
import { tableUtil } from "@/utils/table.util";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { EFilterOperator, type IFilterBy, type IPaginationFilter } from "@/models/filter.model";
import type { INodeTable } from "@/table-models/node.table-model";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { useSettingStore } from "@/stores/setting.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { useClusterStore } from "@/stores/cluster.store";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { PodSortFilterFields, PodVerbosity, WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import type {
  IWorkloadPodsModalOptions,
  WorkloadServicePod,
  IOpenWorkloadDrawerTabPayload,
} from "@/models/workload.model";
import { useWorkloadDrawerNavigation } from "@/composables/use-workload-drawer-navigation.composable";
import { nodePodsModalTableColumns } from "@/table-models/workload.table-model";
import { WorkloadPodsModal } from "../../components/workload/workload-pods-modal/";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { useRouter } from "vue-router";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { filterService } from "@/services/filter.service/filter.service";
import { RunaiDrawerContent } from "@/components/common/runai-drawer-content";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import { NodeDrawerContent } from "@/components/node/node-drawer-content/";
import GpusModal from "@/components/node/gpus-modal/gpus-modal.vue";
import NumaNodesModal from "@/components/node/numa-nodes-modal/numa-nodes-modal.vue";
import { MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING } from "@/common/version.constant";

const DEFAULT_PODS_MODAL_OPTIONS: IWorkloadPodsModalOptions = {
  header: "",
  columns: [] as ITableColumn[],
  getPods: async () => [],
  getPodsCount: async () => 0,
};
export default defineComponent({
  name: "node-index",
  components: {
    NodeDrawerContent,
    RunaiDrawerContent,
    RunaiDrawerNoData,
    WorkloadPodsModal,
    RunaiTable,
    RunaiPageActions,
    RunaiTableNoData,
    RunaiTableWrapper,
    GpusModal,
    NumaNodesModal,
  },
  setup() {
    const appStore = useAppStore();
    const settingStore = useSettingStore();
    const nodePoolStore = useNodePoolStore();
    const clusterStore = useClusterStore();
    const router = useRouter();

    const isDrawerOpen = ref(false);
    const drawerWidth = ref("0px");
    const isOnlyDefaultNodePool = ref(false);
    const isGpusModalOpen = ref(false);
    const nodeForGpusModal = ref<Node | null>(null);
    const nodeForNumaNodesModal = ref<Node | null>(null);
    const isPodsModalOpen = ref(false);
    const podsModalOptions = ref<IWorkloadPodsModalOptions>({ ...DEFAULT_PODS_MODAL_OPTIONS });

    const isAdvancedGpuMetricsEnabled = settingStore.isAdvancedGpuMetricsEnabled;

    /** Api **/
    const countNodes = async (filterBy: IFilterBy): Promise<void> => {
      const filters: IPaginationFilter = getNodeFilters(filterBy);
      const count = await clusterService.countTenantNodes(filters.filterBy, filterBy.searchTerm);
      updateFilterByKey("rowsNumber", count);
    };

    const getNodes = async (filterBy: IFilterBy) => {
      const filters = getNodeFilters(filterBy);

      return await clusterService.getTenantNodes(
        filters.filterBy,
        filters.sortBy as NodeFilterSortFields | undefined,
        filters.sortOrder,
        filters.offset,
        filters.limit,
        filters.search,
      );
    };

    const getNodesWithAdditionalData = async (filterBy: IFilterBy, softLoad = false) => {
      let nodes = tableData.value;
      if (!softLoad) {
        nodes = await getNodes(filterBy);
      }

      const response = await clusterService.getNodeTelemetries(
        [
          NodeTelemetryType.ReadyGpus,
          NodeTelemetryType.TotalGpus,
          NodeTelemetryType.FreeGpus,
          NodeTelemetryType.TotalGpuMemoryBytes,
          NodeTelemetryType.AllocatedGpus,
          NodeTelemetryType.TotalCpuCores,
          NodeTelemetryType.TotalCpuMemoryBytes,
          NodeTelemetryType.AllocatedCpuCores,
          NodeTelemetryType.AllocatedCpuMemoryBytes,
        ],
        filterBy.clusterUuid,
        undefined,
        [ENodeTelemetryGroupBy.Node, ENodeTelemetryGroupBy.ClusterId],
      );

      return nodePoolUtil.enrichNodesWithTelemetry(nodes, response);
    };

    const checkIfOnlyDefaultNodePool = async (): Promise<void> => {
      if (!clusterId.value) return;
      try {
        isOnlyDefaultNodePool.value = await nodePoolStore.isOnlyDefaultNodePoolByClusterId(clusterId.value);
      } catch (e) {
        isOnlyDefaultNodePool.value = false;
      }
    };

    /** Columns **/
    const hasClusterWithNuma = computed<boolean>(() =>
      clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING),
    );

    const columns: ComputedRef<ITableColumn[]> = computed(() => {
      return nodeIndexColumns.filter((col: ITableColumn) => {
        if (nodeDependentColumns.cluster.has(col.name)) {
          return clusterStore.isMultiCluster;
        }
        if (nodeDependentColumns.nodePool.has(col.name as "nodePool")) {
          return !isOnlyDefaultNodePool.value;
        }
        if (nodeDependentColumns.numa.has(col.name)) {
          return hasClusterWithNuma.value;
        }
        return true;
      });
    });

    const podsModalColumns = computed<ITableColumn[]>(() => nodePodsModalTableColumns);

    /** Filters **/
    const getNodeFilters = (filterBy: IFilterBy): IPaginationFilter => {
      return filterService.mapColumnsFilterToFilterParams(filterBy, Object.values(NodeFilterSortFields), true);
    };

    const updateFilters = (filters: IFilterBy) => {
      setFilterBy(filters);
      countNodes(filters);
      checkIfOnlyDefaultNodePool();
    };

    /** Modals **/
    const getPodsModalFilterBy = (nodeName: string, clusterId: string): string[] => {
      return [
        filterUtil.getEqualsFilterString(PodSortFilterFields.ClusterId, clusterId),
        filterUtil.getEqualsFilterString(PodSortFilterFields.NodeName, nodeName),
      ];
    };

    const displayPodsModal = async (node: Node): Promise<void> => {
      podsModalOptions.value.columns = podsModalColumns.value;
      podsModalOptions.value.header = `Pods Associated with node ${node.name}`;
      isPodsModalOpen.value = true;
      podsModalOptions.value.clusterId = clusterId.value;
      podsModalOptions.value.getPods = async (filters: IPaginationFilter) => {
        return await workloadService.listPods(
          false,
          filters.offset,
          filters.limit,
          filters.sortOrder,
          filters.sortBy as PodSortFilterFields,
          getPodsModalFilterBy(node.name, node.clusterUuid),
          PodVerbosity.Full,
          "false",
          filters.search,
        );
      };
      podsModalOptions.value.getPodsCount = async (filters: IPaginationFilter) => {
        return await workloadService.countPods(
          false,
          getPodsModalFilterBy(node.name, node.clusterUuid),
          filters.search,
          "false",
        );
      };
    };

    const closePodsModal = (): void => {
      isPodsModalOpen.value = false;
      resetPodsModalOptions();
    };

    const resetPodsModalOptions = (): void => {
      podsModalOptions.value = { ...DEFAULT_PODS_MODAL_OPTIONS };
    };

    const displayGpusModal = (node: Node): void => {
      nodeForGpusModal.value = node;
      isGpusModalOpen.value = true;
    };

    const displayNumaNodesModal = (node: Node): void => {
      nodeForNumaNodesModal.value = node;
    };

    /** Redirection **/
    const onPodWorkloadClicked = (pod: WorkloadServicePod): void => {
      navigateTo(WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, WorkloadSortFilterFields.Name, "Workload", pod.workloadName);
    };

    const onPodProjectClicked = (pod: WorkloadServicePod): void => {
      navigateTo(PROJECT_ROUTE_NAMES.PROJECT_INDEX, ProjectFilterSortFields.Name, "Project", pod.projectName);
    };

    /** Drawer Navigation **/
    const { navigateToWorkloadDrawer } = useWorkloadDrawerNavigation();

    const handleOpenDrawerTab = (payload: IOpenWorkloadDrawerTabPayload): void => {
      closePodsModal();
      navigateToWorkloadDrawer(payload.workloadId, payload.tab, payload.podName);
    };

    const navigateTo = (routeName: string, field: string, entity: string, value?: string): void => {
      const queryParams = value
        ? { advancedFilters: `[${filterUtil.getFreeTextFilterQuery(field, value, entity, EFilterOperator.Equals)}]` }
        : undefined;
      closePodsModal();
      router.push({ name: routeName, query: queryParams });
    };

    /** Assorted **/
    const getRowKey = (row: Node) => row.id || "";
    const exportTableAsCsv = () => {
      const filteredColumns = columns.value.filter((col: ITableColumn) =>
        filterBy.value.displayedColumns?.includes(col.name),
      );

      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.Node, tableData.value, filteredColumns);
    };

    const {
      setFilterBy,
      clearFilters,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      selectedRows,
      resetSelectedRows,
      selectedRowsAmount,
      initTableFilter,
      selectedRow,
      removeRow,
      clusterId,
      updateFilterByKey,
    } = useTableDataAdvancedFilters<INodeTable>(
      getNodes,
      columns,
      EAdvancedIndexPages.NODE,
      undefined,
      getNodesWithAdditionalData,
    );

    initTableFilter();
    appStore.setPageLoading(false);
    checkIfOnlyDefaultNodePool();
    countNodes(filterBy.value);

    return {
      updateFilters,
      clearFilters,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      selectedRows,
      resetSelectedRows,
      selectedRowsAmount,
      initTableFilter,
      selectedRow,
      removeRow,
      getRowKey,
      columns,
      drawerWidth,
      isDrawerOpen,
      exportTableAsCsv,
      displayPodsModal,
      isPodsModalOpen,
      isAdvancedGpuMetricsEnabled,
      podsModalOptions,
      closePodsModal,
      onPodProjectClicked,
      onPodWorkloadClicked,
      handleOpenDrawerTab,
      displayGpusModal,
      clusterId,
      nodeForGpusModal,
      displayNumaNodesModal,
      nodeForNumaNodesModal,
    };
  },
});
</script>

<style scoped lang="scss"></style>
