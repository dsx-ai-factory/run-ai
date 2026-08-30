<template>
  <runai-base-dialog model-value class="workload-stop-modal" size="width-sm" @close="$emit('close')">
    <template #header>Stop</template>

    <template #body>
      <div class="modal-info row items-center q-gutter-sm q-mb-md">
        <span>
          {{ selectedWorkloadsAmount - stoppableWorkloadsAmount }} out of {{ selectedWorkloadsAmount }} workloads can't
          be stopped
        </span>
        <runai-tooltip
          tooltip-text="At least one selected workload can't be stopped. You're either not authorized to perform this action or need to go to the application from which they were created to perform the action."
        />
      </div>

      <div class="q-mt-lg modal-msg">
        <span class="text-weight-medium">
          Stop {{ stoppableWorkloadsAmount }} workload{{ stoppableWorkloadsAmount === 1 ? "" : "s" }}?
        </span>
      </div>
    </template>

    <template #footer>
      <q-btn color="primary" flat @click="$emit('close')">cancel</q-btn>
      <q-btn color="primary" @click="$emit('stop-workloads')">
        {{
          selectedWorkloadsAmount === 1
            ? "Stop"
            : `Stop ${stoppableWorkloadsAmount} workload${stoppableWorkloadsAmount === 1 ? "" : "s"}`
        }}
      </q-btn>
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// Components
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export default defineComponent({
  name: "workload-stop-modal",
  components: {
    RunaiBaseDialog,
    RunaiTooltip,
  },
  emits: {
    close: () => true,
    "stop-workloads": () => true,
  },
  props: {
    selectedWorkloadsAmount: {
      type: Number as PropType<number>,
      required: true,
    },
    stoppableWorkloadsAmount: {
      type: Number as PropType<number>,
      required: true,
    },
  },
  data() {
    return {};
  },
});
</script>

<style lang="scss" scoped>
.workload-stop-modal {
  .modal-info {
    font-size: 14px;
  }
  .modal-msg {
    font-size: 16px;
  }
}
</style>
