<template>
  <q-btn-dropdown
    split
    color="primary"
    :label="mainLabel"
    @click="onDropdownMainClick"
    :loading="isLoading"
    class="quick-submission-button"
  >
    <q-list aid="quick-submission-dropdown-list">
      <q-item clickable v-close-popup :disable="!isTemplateRunnable" @click="onSubmit">
        <q-item-section aid="quick-submit-option">
          <q-item-label>{{ creationLabel }}</q-item-label>
          <q-tooltip v-if="!isTemplateRunnable">
            To create the workload, first complete the required fields in the Advanced setup.
          </q-tooltip>
        </q-item-section>
      </q-item>
      <q-item clickable v-close-popup @click="onContinue">
        <q-item-section aid="advanced-setup-option">
          <q-item-label
            >Advanced setup
            <runai-tooltip
              class="q-ml-md"
              tooltip-text="Fine-tune your setup with additional preferences, including environment details and data sources."
              icon="ph-bold ph-question"
              tooltip-position="right"
            />
          </q-item-label>
        </q-item-section>
      </q-item>
    </q-list>
  </q-btn-dropdown>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { WorkloadType } from "@/swagger-models/workloads-client";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export default defineComponent({
  name: "quick-submission-dropdown",
  components: {
    RunaiTooltip,
  },
  emits: ["submit", "continue"],
  props: {
    workloadType: {
      type: String as PropType<WorkloadType>,
      required: true,
    },
    isTemplateRunnable: {
      type: Boolean,
      required: true,
    },
    isLoading: {
      type: Boolean,
      required: true,
    },
  },
  mounted() {
    // Add aid to the dropdown button after component is mounted
    this.$nextTick(() => {
      const dropdownButton = this.$el.querySelector(".q-btn-dropdown__arrow-container");
      if (dropdownButton) {
        dropdownButton.setAttribute("aid", "quick-submission-dropdown-button");
      }
    });
  },
  computed: {
    mainLabel(): string {
      return this.isTemplateRunnable ? this.creationLabel : "Advanced setup";
    },
    creationLabel(): string {
      const wLabel = this.workloadType === WorkloadType.Distributed ? WorkloadType.Training : this.workloadType;
      return `Create ${wLabel.toLocaleLowerCase()}`;
    },
  },
  methods: {
    onContinue(): void {
      this.$emit("continue");
    },
    onDropdownMainClick(): void {
      this.isTemplateRunnable ? this.onSubmit() : this.onContinue();
    },
    async onSubmit(): Promise<void> {
      this.$emit("submit");
    },
  },
});
</script>
