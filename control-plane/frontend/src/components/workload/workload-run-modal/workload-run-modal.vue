<template>
  <runai-base-dialog model-value class="workload-run-modal" size="width-sm" @close="$emit('close')">
    <template #header>Run</template>

    <template #body>
      <div class="modal-info row items-center q-gutter-sm q-mb-md">
        <span>
          {{ selectedWorkloadsAmount - runnableWorkloadsAmount }} out of {{ selectedWorkloadsAmount }} workloads can't
          run
        </span>
        <runai-tooltip
          tooltip-text="At least one selected workload can't be set to run. You're either not authorized to perform this action or need to go to the application from which they were created to perform the action."
        />
      </div>

      <div class="q-mt-lg modal-msg">
        <span class="text-weight-medium">
          Run {{ runnableWorkloadsAmount }} workload{{ runnableWorkloadsAmount === 1 ? "" : "s" }}?
        </span>
      </div>
    </template>

    <template #footer>
      <q-btn color="primary" flat @click="$emit('close')">cancel</q-btn>
      <q-btn color="primary" @click="$emit('run-workloads')">
        {{
          selectedWorkloadsAmount === 1
            ? "Run"
            : `Run ${runnableWorkloadsAmount} workload${runnableWorkloadsAmount === 1 ? "" : "s"}`
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
  name: "workload-run-modal",
  components: {
    RunaiBaseDialog,
    RunaiTooltip,
  },
  emits: {
    close: () => true,
    "run-workloads": () => true,
  },
  props: {
    selectedWorkloadsAmount: {
      type: Number as PropType<number>,
      required: true,
    },
    runnableWorkloadsAmount: {
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
.workload-run-modal {
  .modal-info {
    font-size: 14px;
  }
  .modal-msg {
    font-size: 16px;
  }
}
</style>
