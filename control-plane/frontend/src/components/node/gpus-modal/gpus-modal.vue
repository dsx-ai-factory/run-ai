<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="true" @close="$emit('close')" size="dynamic" no-padding :loading="loading">
    <template #header> GPU Devices for Node {{ node.name }} </template>
    <template #body>
      <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search devices" />
      <runai-table
        :rows="gpuRows"
        :columns="columns"
        :loading="loading"
        :filter-by="tableFilter"
        modal-view
        disable-selection
        @update-filters="updateFilter"
      >
        <template #no-data>
          <runai-table-no-data
            v-if="!loading"
            :filter-by="tableFilter"
            :show-error="displayError"
            entity-name="GPU"
            custom-no-filter-results-message="No GPUs match your filters"
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
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, ref, computed, onMounted } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { type Node, NodeMetricType } from "@/swagger-models/cluster-service-client";
import { gpuModalColumns, type IGpuModalRow } from "@/table-models/node.table-model";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";
import RunaiTableNoData from "@/components/common/runai-table/runai-table-no-data/runai-table-no-data.vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import { nodeUtil } from "@/utils/node.util";
import type { IFilterBy } from "@/models/filter.model";
import { alertUtil } from "@/utils/alert.util";
import { useQuasar } from "quasar";

export default defineComponent({
  name: "gpus-modal",
  components: { RunaiListModalSearch, RunaiTable, RunaiTableNoData, RunaiBaseDialog },
  props: {
    node: {
      type: Object as PropType<Node>,
      required: true,
    },
  },
  setup(props) {
    const gpuRows = ref<IGpuModalRow[]>([]);
    const q = useQuasar();
    const searchTerm = ref("");
    const loading = ref(true);
    const displayError = ref(false);
    const tableFilter = ref<IFilterBy>({
      searchTerm: "",
      rowsPerPage: 5,
    });

    const columns = computed(() => gpuModalColumns);

    const fetchGpuMetrics = async () => {
      if (!props.node.id) return;
      loading.value = true;
      try {
        const response = await fetchMetricsFromApi();
        displayError.value = false;
        if (!response) return;
        gpuRows.value = nodeUtil.parseNodeGpuMeasurements(response);
      } catch (error) {
        displayError.value = true;
        console.error("Failed to get GPU metrics:", error);
        q.notify(alertUtil.getError("Failed to get GPU metrics"));
      } finally {
        loading.value = false;
      }
    };

    const fetchMetricsFromApi = async () => {
      if (!props.node.id) return;
      const metricTypes: NodeMetricType[] = [
        NodeMetricType.GpuAllocationPerGpu,
        NodeMetricType.GpuMemoryAllocationBytesPerGpu,
        NodeMetricType.GpuMemoryUsageBytesPerGpu,
        NodeMetricType.GpuMemoryUtilizationPerGpu,
        NodeMetricType.GpuUtilizationPerGpu,
        NodeMetricType.GpuIdleTimeSecondsPerGpu,
      ];
      const endTime = new Date().toISOString();
      const startTime = new Date(Date.now() - 1).toISOString();
      return await clusterService.getNodeMetrics(props.node.id, metricTypes, startTime, endTime, 1);
    };

    const updateFilter = (filter: IFilterBy): void => {
      tableFilter.value = filter;
    };

    onMounted(() => {
      fetchGpuMetrics();
    });

    return {
      gpuRows,
      searchTerm,
      loading,
      tableFilter,
      columns,
      fetchGpuMetrics,
      updateFilter,
      displayError,
    };
  },
});
</script>
