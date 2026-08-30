<template>
  <runai-base-dialog :model-value="true" @close="$emit('close')" size="dynamic" no-padding :loading="loading">
    <template #header> NUMA Nodes for Node {{ node.name }} </template>
    <template #body>
      <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search nodes" />
      <runai-table
        :rows="numaRows"
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
            entity-name="NUMA node"
            custom-no-filter-results-message="No nodes match your filters"
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

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { type Node, type NumaNodeDetails } from "@/swagger-models/cluster-service-client";
import { numaModalColumns } from "@/table-models/node.table-model";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";
import RunaiTableNoData from "@/components/common/runai-table/runai-table-no-data/runai-table-no-data.vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import type { IFilterBy } from "@/models/filter.model";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useQuasar } from "quasar";

const props = defineProps<{
  node: Node;
}>();

defineEmits<{
  close: [];
}>();

const q = useQuasar();
const numaRows = ref<NumaNodeDetails[]>([]);
const loading = ref(true);
const displayError = ref(false);
const tableFilter = ref<IFilterBy>({
  searchTerm: "",
  rowsPerPage: 5,
});

const columns = computed(() => numaModalColumns);

const fetchNumaNodes = async (): Promise<void> => {
  if (!props.node.id) {
    loading.value = false;
    return;
  }
  loading.value = true;
  try {
    const nodeExtended = await clusterService.getNode(props.node.id);
    displayError.value = false;
    numaRows.value = nodeExtended.numa?.nodes || [];
  } catch (error) {
    displayError.value = true;
    const errorAlert = new ErrorAlert({ generalMessage: "Failed to get NUMA nodes" });
    q.notify(errorAlert.getNotification(error));
  } finally {
    loading.value = false;
  }
};

const updateFilter = (filter: IFilterBy): void => {
  tableFilter.value = filter;
};

onMounted(() => {
  fetchNumaNodes();
});
</script>
