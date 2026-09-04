<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div
    class="step-item"
    :class="{
      active: isActive,
      completed: isCompleted,
      clickable: isClickable,
      'has-sub-steps': hasSubSteps,
    }"
    role="button"
    tabindex="0"
    :aria-disabled="!isClickable"
    :aria-current="isActive ? 'step' : undefined"
    @click="handleStepClick"
    @keydown.enter.prevent="handleStepClick"
    @keydown.space.prevent="handleStepClick"
  >
    <div class="step-number">{{ displayIndex }}</div>
    <div class="step-content">
      <div class="step-title">{{ step.title }}</div>
      <transition name="sub-steps">
        <div class="sub-steps" v-if="isActive && step.subSteps?.length">
          <sidebar-sub-step
            v-for="subStep in step.subSteps"
            :key="subStep.id"
            :sub-step="subStep"
            :is-active="currentSubStep === subStep.id"
            :is-completed="currentSubStep > subStep.id"
            :is-clickable="getSubStepClickability(subStep)"
            @click="handleSubStepClick(subStep.id)"
          />
        </div>
      </transition>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { OnboardingSidebarStep, OnboardingSubStep } from "@/components/onboarding/common/onboarding-sidebar/types";
import SidebarSubStep from "@/components/onboarding/common/sidebar-sub-step";

export default defineComponent({
  name: "sidebar-step",
  components: {
    SidebarSubStep,
  },
  emits: {
    "navigate-to-step": (step: number, subStep: number) => step && subStep,
  },
  props: {
    step: {
      type: Object as PropType<OnboardingSidebarStep>,
      required: true,
    },
    displayIndex: {
      type: Number as PropType<number>,
      required: true,
    },
    currentStep: {
      type: Number as PropType<number>,
      required: false,
    },
    currentSubStep: {
      type: Number as PropType<number>,
      required: true,
    },
    completedSteps: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    visitedSubSteps: {
      type: Object as PropType<Record<number, Set<number>>>,
      required: false,
      default: () => ({}),
    },
    isNavigationDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    hasSubSteps(): boolean {
      return (this.step.subSteps?.length ?? 0) > 0;
    },
    isActive(): boolean {
      return this.currentStep === this.step.id;
    },
    isCompleted(): boolean {
      return this.completedSteps.includes(this.step.id);
    },
    isClickable(): boolean {
      // On landing step no steps should be clickable
      if (!this.currentStep || this.currentStep === 0) {
        return false;
      }

      // If navigation is disabled, only allow clicking the current active step
      if (this.isNavigationDisabled) {
        return this.isActive;
      }

      // Allow navigation to:
      // 1. Completed steps (can go back)
      // 2. Current step (already there, but still clickable for consistency)
      // 3. Next step if previous step is completed (e.g., step 2 becomes available when step 1 is completed)
      const isPreviousStepCompleted = this.step.id > 1 && this.completedSteps.includes(this.step.id - 1);
      const isNextStepAvailable = this.step.id === this.currentStep + 1 && isPreviousStepCompleted;

      return this.isCompleted || this.isActive || isNextStepAvailable;
    },
  },
  methods: {
    handleStepClick(): void {
      if (this.isClickable) {
        // When navigating to a different step, start at the first substep
        const targetSubStep = this.step.id === this.currentStep ? this.currentSubStep : 1;
        this.$emit("navigate-to-step", this.step.id, targetSubStep);
      }
    },

    handleSubStepClick(subStepId: number): void {
      // Allow navigation to any visited substep for the current step
      if (
        this.step.id === this.currentStep &&
        this.visitedSubSteps &&
        this.visitedSubSteps[this.step.id] &&
        this.visitedSubSteps[this.step.id].has(subStepId)
      ) {
        this.$emit("navigate-to-step", this.step.id, subStepId);
      }
    },

    canNavigateToStep(stepId: number): boolean {
      return stepId <= (this.currentStep || 0) || this.completedSteps.includes(stepId);
    },

    getSubStepClickability(subStep: OnboardingSubStep): boolean {
      // If navigation is disabled, don't allow any sub-step clicks
      if (this.isNavigationDisabled) {
        return false;
      }

      // Allow click if substep is in visited set for this step
      return (
        this.step.id === this.currentStep &&
        this.visitedSubSteps &&
        this.visitedSubSteps[this.step.id] &&
        this.visitedSubSteps[this.step.id].has(subStep.id) &&
        subStep.isClickable !== false
      );
    },
  },
});
</script>

<style scoped lang="scss">
.step-item {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
  transition: all 0.3s ease;
  position: relative;

  &.active {
    opacity: 1;
  }

  &.clickable {
    cursor: pointer;

    &:hover {
      .step-content .step-title {
        color: white; // White 100% for enabled/clickable steps
      }
    }
  }

  &:not(.clickable) {
    cursor: default;

    &:hover {
      .step-content .step-title {
        // No color change for non-clickable steps
      }
    }
  }

  // Add vertical line under each step dot (except the last one)
  &:not(:last-child) {
    &::after {
      content: "";
      position: absolute;
      left: 15px; // Center the line under the dot
      top: 58px; // Start below the step number
      width: 1px;
      height: calc(100% + 10px - 58px - 20px);
      background: #666666;
      transition: all 0.3s ease;
    }
  }

  .step-number {
    width: 32px;
    height: 32px;
    background: #ffffff80;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 14px;
    flex-shrink: 0;
    color: black;
    position: relative;
    z-index: 1;
    transition: all 0.3s ease;
  }

  &.active .step-number {
    background: var(--primary-color);
    color: black;
  }

  &.completed .step-number {
    background: #ffffffb2;
    color: black;
  }

  .step-content {
    flex: 1;

    .step-title {
      font-size: 14px;
      margin-bottom: 8px;
      padding-top: 5px;
      color: #ffffff80;
      transition: all 0.3s ease;
      font-weight: 700;
    }

    .sub-steps {
      overflow: hidden;
      padding-top: 15px;
    }
  }

  &.active .step-content .step-title {
    color: white;
  }

  &.active.has-sub-steps .step-content .step-title {
    margin-bottom: 16px;
  }

  &.completed .step-content .step-title {
    color: #ffffffb2;
  }

  // Hover effect for completed AND clickable steps
  &.completed.clickable:hover .step-content .step-title {
    color: white; // White 100% for completed steps on hover
  }

  // Ensure non-clickable completed steps don't change on hover
  &.completed:not(.clickable):hover .step-content .step-title {
    color: #ffffffb2; // Keep the same color as non-hover state
  }
}

// Transition animations for sub-steps
.sub-steps-enter-active,
.sub-steps-leave-active {
  transition: all 0.3s ease;
  transform-origin: top;
}

.sub-steps-enter-from {
  opacity: 0;
  transform: translateY(-10px) scaleY(0.8);
  max-height: 0;
}

.sub-steps-leave-to {
  opacity: 0;
  transform: translateY(-10px) scaleY(0.8);
  max-height: 0;
}

.sub-steps-enter-to,
.sub-steps-leave-from {
  opacity: 1;
  transform: translateY(0) scaleY(1);
  max-height: 200px;
}
</style>
