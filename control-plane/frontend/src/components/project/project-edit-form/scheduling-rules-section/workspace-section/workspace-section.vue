<template>
  <wrapper-card
    header="Workspace duration"
    subheader="Set a time limit for workspaces regardless of their activity (e.g., Stop the workspace after 1 day of work)"
  >
    <template #body>
      <gray-card hide-close-button>
        <runai-duration-input
          hide-seconds
          @update:model-value="updateDuration"
          :model-value="duration"
          :rules="[durationValidation]"
        />
      </gray-card>
    </template>
  </wrapper-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// cmps
import { RunaiDurationInput } from "@/components/common/runai-duration-input/";
import { WrapperCard } from "@/components/project/project-edit-form/cards/";
import { GrayCard } from "@/components/project/project-edit-form/cards/";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  components: { GrayCard, WrapperCard, RunaiDurationInput },
  emits: ["update:workspace-duration"],
  props: {
    workspaceDuration: {
      type: [Number, null] as PropType<number | null>,
      required: false,
    },
  },
  computed: {
    duration(): number {
      return this.workspaceDuration || 0;
    },
  },
  methods: {
    updateDuration(duration: number): void {
      if (duration > 0) {
        this.$emit("update:workspace-duration", duration);
      } else {
        this.$emit("update:workspace-duration", null);
      }
    },
    durationValidation(val: number): boolean | string {
      return val <= 0 ? errorMessages.DURATION_GREATER_THAN_ZERO : true;
    },
  },
});
</script>
