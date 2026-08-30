<template>
  <section class="node-modal">
    <runai-base-dialog :model-value="true" @close="$emit('close')" size="dynamic" :no-padding="true">
      <template #header> Nodes Associated with Node pool {{ options.nodePoolName }}</template>

      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search nodes" />
        <q-separator />
        <runai-table
          :columns="columns"
          :loading="loading"
          :rows="nodes"
          modal-view
          :filter-by="tableFilter"
          disable-selection
          is-server-side-pagination
          @update-filters="updateFilter"
          @clear-filters="tableFilter.searchTerm = ''"
          :rows-per-page-options="[5, 7, 10, 15, 20, 25]"
        >
          <template #no-data>
            <runai-table-no-data
              :filter-by="tableFilter"
              entity-name="node"
              :show-create-btn="false"
              :show-filter-icon-and-button="false"
            />
          </template>
        </runai-table>
      </template>

      <template #footer>
        <q-btn label="Close" color="primary" @click="$emit('close')" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// models
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { IFilterBy } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import type { ITableColumn } from "@/models/table.model";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";
import { alertUtil } from "@/utils/alert.util";
import { NodeFilterSortFields, NodeTelemetryType } from "@/swagger-models/cluster-service-client";
import type { INodeModalOptions } from "@/models/node.model";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { type INodeTable, nodeModalColumns } from "@/table-models/node.table-model";
import { filterUtil } from "@/utils/filter.util/filter.util";

export default defineComponent({
  components: {
    RunaiTable,
    RunaiBaseDialog,
    RunaiListModalSearch,
    RunaiTableNoData,
  },
  emits: ["close"],
  props: {
    options: {
      type: Object as PropType<INodeModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      loading: false as boolean,
      tableFilter: {} as IFilterBy,
      columns: nodeModalColumns as Array<ITableColumn>,
      nodes: [] as Array<INodeTable>,
    };
  },
  created() {
    this.loadFilters();
    this.loadNodesData();
  },
  methods: {
    loadFilters(): void {
      this.tableFilter = filterService.getEmptyFilterByModel({});
    },
    getClusterFilters(): string[] {
      return [
        filterUtil.getEqualsFilterString(NodeFilterSortFields.ClusterUuid, this.options.clusterId),
        filterUtil.getEqualsFilterString(NodeFilterSortFields.NodePool, this.options.nodePoolName),
      ];
    },
    async countNodes(): Promise<void> {
      const clusterFilter = this.getClusterFilters();
      const count = await clusterService.countTenantNodes(clusterFilter, this.tableFilter.searchTerm);
      this.updateFilterByKey("rowsNumber", count);
    },
    async loadNodesData() {
      this.loading = true;
      try {
        const filters = filterService.mapColumnsFilterToFilterParams(
          this.tableFilter,
          Object.values(NodeFilterSortFields),
          true,
        );
        // Count nodes first
        await this.countNodes();

        this.nodes = await clusterService.getTenantNodes(
          this.getClusterFilters(),
          filters.sortBy as NodeFilterSortFields | undefined,
          filters.sortOrder,
          filters.offset,
          filters.limit,
          filters.search,
        );
        await this.getAndEnrichNodesWithTelemetry();
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify(alertUtil.getError(`Failed to load nodes`));
      } finally {
        this.loading = false;
      }
    },
    async getAndEnrichNodesWithTelemetry(): Promise<void> {
      const response = await clusterService.getNodeTelemetries(
        [
          NodeTelemetryType.ReadyGpus,
          NodeTelemetryType.TotalGpuMemoryBytes,
          NodeTelemetryType.AllocatedGpus,
          NodeTelemetryType.TotalCpuCores,
          NodeTelemetryType.TotalCpuMemoryBytes,
          NodeTelemetryType.AllocatedCpuCores,
          NodeTelemetryType.AllocatedCpuMemoryBytes,
        ],
        this.options.clusterId,
        this.options.nodePoolName,
        [ENodeTelemetryGroupBy.Node, ENodeTelemetryGroupBy.ClusterId],
      );
      nodePoolUtil.enrichNodesWithTelemetry(this.nodes, response);
    },
    async updateFilter(filter: IFilterBy): Promise<void> {
      this.tableFilter = filter;
      await this.loadNodesData();
    },
    updateFilterByKey<T extends keyof IFilterBy>(key: T, value: IFilterBy[T]): void {
      this.tableFilter = { ...this.tableFilter, [key]: value };
    },
  },
});
</script>
