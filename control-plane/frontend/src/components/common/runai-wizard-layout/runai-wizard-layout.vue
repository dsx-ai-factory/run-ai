<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-panels-layout>
    <template #left>
      <runai-panel-container :transparent="transparentStepsPanel">
        <runai-wizard-steps-panel
          :steps="steps"
          :active-step="activeStep"
          :disable-navigation="disableNavigation"
          @step-click="handleStepClick"
        />
      </runai-panel-container>
    </template>

    <!-- Main content area -->
    <div class="wizard-main-content">
      <div class="wizard-content-scroll">
        <slot />
      </div>

      <!-- Content Footer -->
      <div v-if="$slots['content-footer']" class="wizard-content-footer">
        <slot name="content-footer" />
      </div>
    </div>

    <template #right v-if="$slots['right-sidebar']">
      <runai-panel-container>
        <slot name="right-sidebar" />
      </runai-panel-container>
    </template>
  </runai-panels-layout>
</template>

<script setup lang="ts">
import { watch } from "vue";
import RunaiWizardStepsPanel from "./runai-wizard-steps-panel.vue";
import RunaiPanelsLayout from "@/components/common/runai-panels-layout/runai-panels-layout.vue";
import RunaiPanelContainer from "@/components/common/runai-panel-container/runai-panel-container.vue";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";

defineOptions({
  name: "runai-wizard-layout",
});

const props = defineProps<{
  steps: IRunaiWizardLayoutStep[];
  activeStep: string;
  disableNavigation?: boolean;
  transparentStepsPanel?: boolean;
}>();

const emit = defineEmits<{
  "step-click": [payload: { step: IRunaiWizardLayoutStep; name: string }];
}>();

watch(
  () => props.activeStep,
  () => {
    // Quasar uses a .scroll container for the scrollable area
    const scrollContainer = document.querySelector(".app-main .scroll");
    scrollContainer?.scrollTo({ top: 0, behavior: "smooth" });
  },
);

const handleStepClick = (payload: { step: IRunaiWizardLayoutStep; name: string }) => {
  emit("step-click", payload);
};
</script>

<style scoped lang="scss">
.wizard-main-content {
  display: flex;
  flex-direction: column;
  min-height: 200px;
}

.wizard-content-scroll {
  flex: 1;
}

// Content Footer
.wizard-content-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
