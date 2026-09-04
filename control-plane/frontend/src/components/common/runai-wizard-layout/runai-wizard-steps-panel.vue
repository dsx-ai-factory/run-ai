<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div
    class="wizard-steps-panel q-py-md"
    :class="{
      'wizard-steps-panel--nav-disabled': disableNavigation,
      'wizard-steps-panel--horizontal': orientation === 'horizontal',
    }"
    role="list"
    aria-label="Form steps"
  >
    <div
      v-for="(step, index) in steps"
      :key="step.name"
      class="wizard-step"
      :class="getStepClasses(step)"
      role="listitem"
    >
      <div class="wizard-step-wrapper">
        <q-item
          :clickable="!disableNavigation && !step.disabled"
          :disable="step.disabled"
          role="button"
          :aria-current="step.name === activeStep ? 'step' : undefined"
          :aria-label="step.title"
          :aria-disabled="step.disabled || disableNavigation ? 'true' : undefined"
          @click="handleStepClick(step)"
          class="wizard-step-item q-pa-sm q-px-md"
          :aid="getStepAid(step.name)"
        >
          <q-item-section avatar class="wizard-step-avatar">
            <div class="wizard-step-indicator" :class="getIndicatorClasses(step)">
              <q-icon
                class="wizard-step-icon"
                v-if="shouldShowIcon(step)"
                :name="getStepIcon(step)"
                :size="getStepIconSize(step)"
                color="white"
              />
              <span v-else class="wizard-step-number text-weight-bolder">{{ index + 1 }}</span>
            </div>
          </q-item-section>
          <q-item-section>
            <q-item-label class="wizard-step-title text-weight-medium" :class="getTitleClasses(step)">
              {{ step.title }}
            </q-item-label>
          </q-item-section>
        </q-item>

        <!-- Tooltip placed outside q-item so it works even when step is disabled -->
        <q-tooltip v-if="step.tooltip" anchor="center right">
          {{ step.tooltip }}
        </q-tooltip>
      </div>

      <!-- Divider line between steps -->
      <div v-if="index < steps.length - 1" class="wizard-step-divider"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { IRunaiWizardLayoutStep, WizardOrientation } from "@/models/runai-wizard-layout.model";
import { stringUtil } from "@/utils/string.util";

defineOptions({
  name: "runai-wizard-steps-panel",
});

const props = withDefaults(
  defineProps<{
    steps: IRunaiWizardLayoutStep[];
    activeStep: string;
    disableNavigation?: boolean;
    orientation?: WizardOrientation;
  }>(),
  {
    orientation: "vertical",
  },
);

const emit = defineEmits<{
  "step-click": [payload: { step: IRunaiWizardLayoutStep; name: string }];
}>();

const getStepAid = (stepName: string): string => {
  return `wizard-step-item-${stringUtil.slug(stepName)}`;
};

const handleStepClick = (step: IRunaiWizardLayoutStep) => {
  if (!step.disabled && !props.disableNavigation) {
    emit("step-click", { step, name: step.name });
  }
};

const shouldShowIcon = (step: IRunaiWizardLayoutStep): boolean => {
  // Show system icons for active/error/done states, or custom icon if defined
  return step.name === props.activeStep || step.hasError || step.done || !!step.icon;
};

const getStepIcon = (step: IRunaiWizardLayoutStep): string => {
  if (step.name === props.activeStep) return "ph-fill ph-pencil-simple";
  if (step.hasError) return "ph-bold ph-warning-circle";
  if (step.done) return "ph-bold ph-check";
  if (step.icon) return step.icon; // Custom icon when not active/done/error
  return "";
};

const getStepIconSize = (step: IRunaiWizardLayoutStep): string => {
  if (step.hasError) return "20px";
  return "16px";
};

const getStepClasses = (step: IRunaiWizardLayoutStep) => {
  const isSelected = step.name === props.activeStep;
  return {
    "wizard-step--completed": step.done && !isSelected,
    "wizard-step--selected": isSelected,
    "wizard-step--disabled": step.disabled,
    "wizard-step--error": step.hasError,
    "wizard-step--default": !step.done && !isSelected && !step.hasError,
  };
};

const getIndicatorClasses = (step: IRunaiWizardLayoutStep) => {
  const isSelected = step.name === props.activeStep;
  if (isSelected || step.done) return "bg-primary";
  if (step.hasError) return "bg-negative";
  if (step.disabled) return "bg-grey-5";
  return "bg-grey-6";
};

const getTitleClasses = (step: IRunaiWizardLayoutStep) => {
  const isSelected = step.name === props.activeStep;
  if (isSelected) return "text-grey-9";
  if (step.hasError) return "text-negative";
  if (step.disabled) return "text-grey-5";
  return "text-grey-6";
};

defineExpose({
  getStepAid,
  handleStepClick,
  shouldShowIcon,
  getStepIcon,
  getStepIconSize,
  getStepClasses,
  getIndicatorClasses,
  getTitleClasses,
});
</script>

<style scoped lang="scss">
.wizard-step {
  position: relative;
}

.wizard-step-item {
  min-height: 38px;
  transition: background-color 0.2s ease;
}

.wizard-step-avatar {
  min-width: 24px;
  padding-right: 10px;
}

.wizard-step-indicator {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  position: relative;
  transition: all 0.2s ease;
}

.wizard-step-number {
  font-size: 12px;
  color: white;
  position: absolute;
  top: 50%;
  left: 51%; // 51%? odd I know, I am not sure why 50% would not place it perfectly in the center
  transform: translate(-50%, -50%);
}

// Icon alignment - absolute positioning for perfect centering
:deep(.wizard-step-icon) {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.wizard-step-title {
  font-size: 14px;
}

.wizard-step-divider {
  width: 1px;
  height: 30px;
  background-color: $grey-4;
  margin-left: 28px; // 16px (q-px-md) + 12px (center of 24px indicator)
}

// Hover states
.wizard-step--default .wizard-step-item:not([disable]):hover,
.wizard-step--completed .wizard-step-item:not([disable]):hover,
.wizard-step--selected .wizard-step-item:not([disable]):hover {
  background-color: $grey-3;
}

// Disabled state
.wizard-step--disabled {
  cursor: not-allowed;

  .wizard-step-item {
    pointer-events: none;
  }
}

// Navigation disabled state - remove interactivity but keep normal appearance
.wizard-steps-panel--nav-disabled {
  .wizard-step-item {
    cursor: default;
    pointer-events: none; // Prevents clicks and hover effects

    // Override any hover styles
    &:hover {
      background-color: transparent;
    }
  }

  // Re-enable pointer events on the wrapper to allow tooltips to work
  .wizard-step-wrapper {
    pointer-events: auto;
  }
}

.wizard-steps-panel--horizontal {
  display: flex;
  align-items: flex-start;
  width: 100%;

  .wizard-step {
    display: flex;
    align-items: center;
    flex: 1;
  }

  .wizard-step:last-child {
    flex: 0 0 auto;
  }

  .wizard-step-divider {
    width: auto;
    height: 1px;
    flex: 1;
    margin-left: 12px;
    margin-top: 0;
  }
}
</style>
