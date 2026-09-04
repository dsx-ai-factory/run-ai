<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-panels-layout v-if="orientation === 'vertical'">
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

    <runai-wizard-main-content>
      <slot />
      <template v-if="$slots['content-footer']" #content-footer>
        <slot name="content-footer" />
      </template>
    </runai-wizard-main-content>

    <template #right v-if="$slots['right-sidebar']">
      <runai-panel-container>
        <slot name="right-sidebar" />
      </runai-panel-container>
    </template>
  </runai-panels-layout>

  <div v-else class="wizard-layout-horizontal">
    <runai-panel-container :transparent="transparentStepsPanel">
      <runai-wizard-steps-panel
        :steps="steps"
        :active-step="activeStep"
        :disable-navigation="disableNavigation"
        orientation="horizontal"
        @step-click="handleStepClick"
      />
    </runai-panel-container>

    <div class="wizard-layout-horizontal-body">
      <runai-wizard-main-content>
        <slot />
        <template v-if="$slots['content-footer']" #content-footer>
          <slot name="content-footer" />
        </template>
      </runai-wizard-main-content>

      <runai-panel-container v-if="$slots['right-sidebar']">
        <slot name="right-sidebar" />
      </runai-panel-container>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from "vue";
import RunaiWizardStepsPanel from "./runai-wizard-steps-panel.vue";
import RunaiWizardMainContent from "./runai-wizard-main-content.vue";
import RunaiPanelsLayout from "@/components/common/runai-panels-layout/runai-panels-layout.vue";
import RunaiPanelContainer from "@/components/common/runai-panel-container/runai-panel-container.vue";
import type { IRunaiWizardLayoutStep, WizardOrientation } from "@/models/runai-wizard-layout.model";

defineOptions({
  name: "runai-wizard-layout",
});

const props = withDefaults(
  defineProps<{
    steps: IRunaiWizardLayoutStep[];
    activeStep: string;
    disableNavigation?: boolean;
    transparentStepsPanel?: boolean;
    orientation?: WizardOrientation;
  }>(),
  {
    orientation: "vertical",
  },
);

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
.wizard-layout-horizontal {
  display: flex;
  flex-direction: column;
  gap: 24px;
  width: 100%;
}

.wizard-layout-horizontal-body {
  display: flex;
  align-items: flex-start;
  gap: 20px;
  width: 100%;

  :deep(.wizard-main-content) {
    flex: 1;
    min-width: 0;
  }
}
</style>
