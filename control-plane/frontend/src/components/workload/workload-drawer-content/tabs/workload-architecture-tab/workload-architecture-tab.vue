<template>
  <section class="workload-architecture-tab">
    <runai-default-loader v-if="loading" />
    <workload-architecture-diagram
      v-else
      :workload="workload"
      :elements="elements"
      :pods="pods"
      :selected-element-id="selectedElementId"
      :show-expand-actions="showExpandActions"
      :selectable="selectable"
      :element-details-by-id="elementDetailsById"
      @node-selected="$emit('element-selected', $event)"
      @expand-node="$emit('expand-node', $event)"
    />
  </section>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from "vue";
import { useQuasar } from "quasar";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import type { Workload, WorkloadElement, Pod, WorkloadElementResponse } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import WorkloadArchitectureDiagram from "./workload-architecture-diagram.vue";
import type { WorkloadElementDetailsById } from "./workload-architecture-diagram.util";

const $q = useQuasar();

export interface Props {
  workload: Workload;
  selectedElementId?: string | null;
  showExpandActions?: boolean;
  selectable?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  selectedElementId: null,
  showExpandActions: false,
  selectable: false,
});

defineEmits<{
  "element-selected": [elementId: string | null];
  "expand-node": [elementId: string | null];
}>();

const elements = ref<WorkloadElement[]>([]);
const pods = ref<Pod[]>([]);
const elementDetailsById = ref<WorkloadElementDetailsById>({});
const loading = ref(false);
const currentWorkloadId = ref<string | null>(null);

const fetchElementDetails = async (
  workloadId: string,
  fetchedElements: WorkloadElement[],
): Promise<WorkloadElementDetailsById> => {
  const elementDetailsEntries = await Promise.all(
    fetchedElements.map(async (element): Promise<[string, WorkloadElementResponse | null]> => {
      try {
        return [element.id, await workloadService.getWorkloadElement(workloadId, element.id)];
      } catch {
        return [element.id, null];
      }
    }),
  );

  return elementDetailsEntries.reduce<WorkloadElementDetailsById>((acc, [elementId, elementDetails]) => {
    if (elementDetails) {
      acc[elementId] = elementDetails;
    }
    return acc;
  }, {});
};

const fetchData = async (workloadId: string): Promise<void> => {
  try {
    loading.value = true;
    const [fetchedElements, fetchedPods] = await Promise.all([
      workloadService.getWorkloadElements(workloadId),
      workloadService.getWorkloadPods(workloadId, false, undefined, undefined, undefined, true),
    ]);
    elements.value = fetchedElements;
    pods.value = fetchedPods as Pod[];
    elementDetailsById.value = await fetchElementDetails(workloadId, fetchedElements);
  } catch (error: unknown) {
    const errorAlert = new ErrorAlert({ generalMessage: "Failed to load workload structure" });
    $q.notify(errorAlert.getNotification(error));
    elements.value = [];
    pods.value = [];
    elementDetailsById.value = {};
  } finally {
    loading.value = false;
  }
};

const resetState = (): void => {
  currentWorkloadId.value = null;
  elements.value = [];
  pods.value = [];
  elementDetailsById.value = {};
};

watch(
  () => props.workload,
  (newWorkload) => {
    if (newWorkload?.id && newWorkload.id !== currentWorkloadId.value) {
      currentWorkloadId.value = newWorkload.id;
      fetchData(newWorkload.id);
    } else if (!newWorkload?.id) {
      resetState();
    }
  },
  { immediate: true },
);

onUnmounted(() => {
  resetState();
});
</script>

<style lang="scss" scoped>
.workload-architecture-tab {
  padding: 4px;
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
