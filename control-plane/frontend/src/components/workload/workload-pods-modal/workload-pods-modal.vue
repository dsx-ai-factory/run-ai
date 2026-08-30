<template>
  <runai-base-dialog model-value @close="$emit('close')" size="dynamic" no-padding>
    <template #header>{{ modalOptions.header }}</template>
    <template #body>
      <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search pods" :debounce="300" />
      <q-separator />
      <runai-table
        :rows="pods"
        :columns="modalOptions.columns"
        :loading="loading"
        :filter-by="tableFilter"
        :get-row-key="getRowKey"
        @update-filters="updateFilter"
        modal-view
        disable-selection
        @project-clicked="$emit('project-clicked', $event)"
        @workload-clicked="$emit('workload-clicked', $event)"
        :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25]"
        @view-logs="handleLogsClicked"
        @connect-terminal="handleTerminalClicked"
      >
        <template #no-data>
          <runai-table-no-data
            v-if="!loading"
            :filter-by="tableFilter"
            entity-name="pod"
            :show-create-btn="false"
            :show-filter-icon-and-button="false"
            :show-error="shouldDisplayError"
          />
        </template>
      </runai-table>
    </template>
    <template #footer-left v-if="modalOptions.displayRedirection">
      <q-btn
        flat
        :ripple="false"
        color="primary"
        class="btn-link"
        label="Go to nodes to view more details"
        @click="redirectToNodes"
      />
    </template>
    <template #footer>
      <q-btn label="Close" color="primary" @click="$emit('close')" />
    </template>
  </runai-base-dialog>
</template>
<script lang="ts">
import { defineComponent, type PropType, ref } from "vue";

//service
import { filterService } from "@/services/filter.service/filter.service";
//cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiTable } from "@/components/common/runai-table";
//model
import { allWorkloadPodsModalColumnsMap } from "@/table-models/workload.table-model";
import {
  EWorkloadErrorMessage,
  type IWorkloadPodsModalOptions,
  type IOpenWorkloadDrawerTabPayload,
} from "@/models/workload.model";
import {
  CLUSTER_COLUMN_FILTER_NAME,
  type IFilterBy,
  type IFilterModel,
  type IPaginationFilter,
} from "@/models/filter.model";
import { ETableFilters } from "@/models/table.model";
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
import { type Pod as WorkloadServicePod, PodSortFilterFields } from "@/swagger-models/workloads-client";
import { EIntervalLabels } from "@/models/interval.model";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useRefresh } from "@/composables/use-refresh.composable";
import { useQuasar } from "quasar";
import { nodeColumnName, nodeIndexColumns } from "@/table-models/node.table-model";

export default defineComponent({
  name: "workload-pods-modal",
  components: { RunaiTable, RunaiTableNoData, RunaiListModalSearch, RunaiBaseDialog },
  emits: ["close", "project-clicked", "workload-clicked", "open-drawer-tab"],
  props: {
    modalOptions: {
      type: Object as PropType<IWorkloadPodsModalOptions>,
      required: true,
    },
  },
  setup(props) {
    const $q = useQuasar();
    const shouldDisplayError = ref(false);
    const loading = ref(true);
    const pods = ref<WorkloadServicePod[]>([]);
    const tableFilter = ref<IFilterBy>({
      rowsNumber: 10,
      sortBy: PodSortFilterFields.CreatedAt,
    });

    const getPods = async () => {
      try {
        const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(
          tableFilter.value,
          Object.values(PodSortFilterFields),
          true,
        );

        pods.value = await props.modalOptions.getPods(filters);
        tableFilter.value.rowsNumber = await props.modalOptions.getPodsCount(filters);
        shouldDisplayError.value = false;
      } catch (error: unknown) {
        shouldDisplayError.value = true;
        const errorAlert = new ErrorAlert({ generalMessage: EWorkloadErrorMessage.FailedToLoadPads });
        $q.notify(errorAlert.getNotification(error));
      } finally {
        loading.value = false;
      }
    };
    useRefresh(EIntervalLabels.PodsModal, getPods);

    return { shouldDisplayError, loading, pods, tableFilter, getPods };
  },
  created() {
    this.loadFilters();
  },
  methods: {
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        allWorkloadPodsModalColumnsMap.name.name,
        this.modalOptions.columns,
      );
      this.tableFilter = filterService.getEmptyFilterByModel(defaultFilters);
    },
    updateFilter(filter: IFilterBy): void {
      this.loading = true;
      this.tableFilter = filter;
      this.getPods();
    },
    getRowKey(pod: WorkloadServicePod): string {
      return pod.id;
    },
    getNodesFilterBy(): IFilterBy {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(nodeColumnName.name, nodeIndexColumns);
      return filterService.loadFilters(window.location, ETableFilters.NODE, defaultFilters);
    },
    getNodeName(): string | null {
      return this.pods.find((pod) => pod.nodeName)?.nodeName || null;
    },
    addNodeFilter(filterBy: IFilterBy): void {
      const nodeName: null | string = this.getNodeName();
      const isFilteredByNode: boolean | undefined = filterBy.columnFilters?.some(
        (colFilter: IFilterModel) => colFilter.name === "node",
      );
      if (nodeName === null) return;
      if (
        filterBy.columnFilters &&
        !filterBy.columnFilters.some((colFilter: IFilterModel) => colFilter.term === nodeName)
      ) {
        if (!isFilteredByNode) {
          filterBy.columnFilters.push({
            term: nodeName,
            name: "node",
            field: () => "node",
          });
        } else {
          filterBy.columnFilters = [
            {
              term: nodeName,
              name: "node",
              field: () => "node",
            },
          ];
        }
      }
    },
    saveAndNavigateToNodes(filterBy: IFilterBy): void {
      if (this.modalOptions.clusterId) {
        filterService.setColumnFilter(
          filterBy,
          this.modalOptions.clusterId,
          CLUSTER_COLUMN_FILTER_NAME,
          ETableFilters.NODE,
        );
      }
      filterService.saveFilters(ETableFilters.NODE, {
        ...filterBy,
        clusterUuid: this.modalOptions.clusterId,
        sortBy: "node",
      });

      this.$router.push({ name: NODE_ROUTE_NAMES.NODE_INDEX });
    },
    redirectToNodes(): void {
      const nodesFilterBy: IFilterBy = this.getNodesFilterBy();
      this.addNodeFilter(nodesFilterBy);
      this.saveAndNavigateToNodes(nodesFilterBy);
    },
    handleLogsClicked(pod: WorkloadServicePod): void {
      const payload: IOpenWorkloadDrawerTabPayload = {
        workloadId: pod.workloadId,
        podName: pod.name,
        tab: "logs",
      };
      this.$emit("open-drawer-tab", payload);
      this.$emit("close");
    },
    handleTerminalClicked(pod: WorkloadServicePod): void {
      const payload: IOpenWorkloadDrawerTabPayload = {
        workloadId: pod.workloadId,
        podName: pod.name,
        tab: "terminal",
      };
      this.$emit("open-drawer-tab", payload);
      this.$emit("close");
    },
  },
});
</script>

<style scoped lang="scss"></style>
