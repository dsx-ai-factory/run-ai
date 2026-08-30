<template>
  <section class="cluster-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          advanced-filters
          primary-btn-label="new cluster"
          :columns="columns"
          :filters="filterBy"
          @filters-changed="setFilterBy"
          @create-entity="createCluster"
          :selected-rows-amount="selectedRowsAmount"
          :show-primary-btn="true"
          hide-cluster-filter
          @selected-actions-close="resetSelectedRows"
          @export-csv="exportTableAsCsv"
        >
          <template v-slot:selected-rows-actions>
            <div class="row items-center justify-start">
              <runai-action-button
                v-permission="{ resourceType: ResourceType.Cluster, action: Action.Create }"
                aid="install-cluster-btn"
                btn-action="install"
                @click="installCluster"
              />
              <runai-action-button
                v-permission="{ resourceType: ResourceType.NetworkTopologies, action: Action.Read }"
                aid="network-topology-btn"
                btn-action="topology"
                @click="openTopologyManagementModal"
              />
              <runai-tooltip-wrapper
                :display-tooltip="disableDelete"
                tooltip-text="This is the only cluster and therefore can't be removed"
              >
                <runai-action-button
                  v-permission="{
                    resourceType: ResourceType.Cluster,
                    action: Action.Delete,
                  }"
                  :disable="disableDelete"
                  aid="delete-cluster-btn"
                  btn-action="remove"
                  @click="openDeleteModal"
                />
              </runai-tooltip-wrapper>
            </div>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :disable-selection="false"
          :rows="tableData"
          :columns="columns"
          :loading="loadingTableData"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          :top-row="lastCreatedEntity"
          :get-row-key="getRowKey"
          sticky-columns
          @tooltip-click="handleToolTipClick"
          @update-filters="setFilterBy"
          @topologies-clicked="openTopologyPreviewModal"
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedEntity"
              :filter-by="filterBy"
              entity-name="cluster"
              ph-icon-name="runai-cluster"
              icon-size="44"
              @clear-filters="clearFilters"
              @create-new="createCluster"
              :show-error="loadingError"
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <remove-cluster-modal
      v-if="isDeleteModalOpen && selectedRows.length >= 1"
      :cluster-name="selectedRows[0].name"
      :deleting="deleting"
      @cancel="isDeleteModalOpen = false"
      @remove="deleteCluster"
    ></remove-cluster-modal>
    <nonfunctioning-services-table-modal
      :modal-options="nonfunctioningServicesTableModalOptions"
      v-if="isNonfunctioningModalOpen"
      @close="isNonfunctioningModalOpen = false"
    />
    <cluster-network-topologies-modal
      v-if="isTopologyManagementModalOpen"
      :cluster-name="selectedRows[0].name"
      :cluster-uuid="selectedRows[0].uuid"
      @cancel="isTopologyManagementModalOpen = false"
    />
    <topologies-preview-modal
      v-if="isTopologyPreviewModalOpen"
      :modal-options="topologyPreviewModalOptions"
      @close="isTopologyPreviewModalOpen = false"
    />
  </section>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";

// Stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
// components
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RemoveClusterModal } from "@/components/cluster/remove-cluster-modal";
import { ClusterNetworkTopologiesModal } from "@/components/cluster/cluster-network-topologies-modal";
import { TopologiesPreviewModal } from "@/components/cluster/topologies-preview-modal";
// models
import { allClusterColumns, clusterIndexColumns } from "@/table-models/cluster.table-model";
import { filterClustersByAdvancedFilters } from "@/views/cluster/cluster-filter.util";
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

// Services
import { alertUtil } from "@/utils/alert.util";
import { tableUtil } from "@/utils/table.util";
import { filterService } from "@/services/filter.service/filter.service";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import NonfunctioningServicesTableModal from "@/components/cluster/nonfunctioning-services-table-modal/nonfunctioning-services-table-modal.vue";
import type { IService, IServiceTableModalOptions } from "@/models/service.model";
import { serviceMiniTableColumns } from "@/table-models/services.table-model";
import { clusterUtil } from "@/utils/cluster.util";
import { EIntervalLabels } from "@/models/interval.model";
import { intervalUtil } from "@/utils/interval.util";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import type { ITopologyPreviewModalOptions } from "@/table-models/topology.table-model";
import { ETopologyPreviewModalEntityType, topologyPreviewColumns } from "@/table-models/topology.table-model";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";

export default defineComponent({
  components: {
    NonfunctioningServicesTableModal,
    RunaiTableWrapper,
    RunaiPageActions,
    RunaiActionButton,
    RunaiTable,
    RunaiTableNoData,
    RunaiTooltipWrapper,
    RemoveClusterModal,
    ClusterNetworkTopologiesModal,
    TopologiesPreviewModal,
  },
  data() {
    return {
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      loading: false as boolean,
      selectedRows: [] as Array<DisplayedCluster>,
      currentTimeoutId: 0 as number,
      isInitialLoader: true as boolean,
      deleting: false as boolean,
      isDeleteModalOpen: false as boolean,
      isNonfunctioningModalOpen: false as boolean,
      isTopologyManagementModalOpen: false as boolean,
      isTopologyPreviewModalOpen: false as boolean,
      nonfunctioningServicesTableModalOptions: {
        services: [] as Array<IService>,
        header: "Nonfunctioning Services",
        columns: serviceMiniTableColumns,
        loading: false,
      } as IServiceTableModalOptions,
      topologyPreviewModalOptions: {
        sortByDefault: "name",
        entityName: "",
        entityType: ETopologyPreviewModalEntityType.Cluster,
        columns: topologyPreviewColumns,
        clusterId: "",
      } as ITopologyPreviewModalOptions,
    };
  },
  setup() {
    const loadClusters = async (filterBy: IFilterBy): Promise<Array<DisplayedCluster>> => {
      const clusterList = await clusterService.list();
      const filtered = filterService.filterListByTableFilters(clusterList, filterBy, allClusterColumns);
      return filterClustersByAdvancedFilters(filtered, filterBy.advancedFilters, allClusterColumns);
    };
    const columns = computed(() => clusterIndexColumns);
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
    } = useTableDataAdvancedFilters<DisplayedCluster>(loadClusters, columns, EAdvancedIndexPages.CLUSTER);

    initTableFilter();

    return {
      setFilterBy,
      lastCreatedEntity,
      clearLastCreated,
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
    clusters(): Array<ClustersMinimalData> {
      return this.clusterStore.clusterList;
    },
    isSingleCluster(): boolean {
      return this.clusters.length === 1;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    disableDelete(): boolean {
      return this.isSingleCluster || this.selectedRows.length === 0;
    },
  },
  methods: {
    handleToolTipClick(row: DisplayedCluster): void {
      const services = clusterUtil.getNonFunctioningServices(row.status?.operands || {});
      this.nonfunctioningServicesTableModalOptions.services = services.map((service) => ({ name: service }));
      this.isNonfunctioningModalOpen = true;
    },
    getRowKey(cluster: DisplayedCluster): string {
      return cluster.uuid;
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    async createCluster(): Promise<void> {
      this.$router.push({
        name: CLUSTER_ROUTE_NAMES.CLUSTER_NEW,
      });
    },
    stopRefreshClusters(): void {
      intervalUtil.stopInterval(EIntervalLabels.ClusterIndex);
    },
    startRefreshClusters(): void {
      intervalUtil.startInterval(EIntervalLabels.ClusterIndex, this.refreshList);
    },
    openTopologyManagementModal(): void {
      this.isTopologyManagementModalOpen = true;
    },
    openTopologyPreviewModal(cluster: DisplayedCluster): void {
      this.topologyPreviewModalOptions = {
        sortByDefault: "name",
        entityName: cluster.name,
        entityType: ETopologyPreviewModalEntityType.Cluster,
        columns: topologyPreviewColumns,
        clusterId: cluster.uuid,
      };
      this.isTopologyPreviewModalOpen = true;
    },
    openDeleteModal(): void {
      this.stopRefreshClusters();
      this.isDeleteModalOpen = true;
    },
    async deleteCluster(): Promise<void> {
      const cluster: DisplayedCluster | undefined = this.selectedRows[0];
      if (!cluster || this.isSingleCluster) return;

      try {
        this.deleting = true;
        await this.clusterStore.deleteCluster(cluster.uuid);
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Cluster ${cluster.name} removed`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to delete cluster"));
        }
      } finally {
        this.startRefreshClusters();
        this.resetSelectedRows();
        this.clearLastCreated();
        this.isDeleteModalOpen = false;
        this.deleting = false;
      }
    },
    installCluster(): void {
      const selectedClusterId: string = this.selectedRows[0].uuid;
      if (!selectedClusterId) return;
      this.$router.push({
        name: CLUSTER_ROUTE_NAMES.CLUSTER_INSTALL,
        params: { id: selectedClusterId },
      });
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));
      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.Cluster, this.tableData, columns);
    },
  },
});
</script>
