<template>
  <div class="onboarding-sidebar">
    <sidebar-header class="sidebar-header" />

    <div class="steps-container">
      <sidebar-step
        v-for="(step, index) in filteredSteps"
        :key="step.id"
        :step="step"
        :display-index="index + 1"
        :current-step="currentStep"
        :current-sub-step="currentSubStep"
        :completed-steps="completedSteps"
        :visited-sub-steps="visitedSubSteps"
        :is-navigation-disabled="isNavigationDisabled"
        @navigate-to-step="handleNavigateToStep"
      />
    </div>

    <sidebar-footer />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType, computed } from "vue";
import SidebarHeader from "@/components/onboarding/common/sidebar-header";
import SidebarStep from "@/components/onboarding/common/sidebar-step";
import SidebarFooter from "@/components/onboarding/common/sidebar-footer";
import { ONBOARDING_SIDEBAR_CONFIG } from "@/components/onboarding/common/onboarding-sidebar/sidebar-config";

export default defineComponent({
  name: "onboarding-sidebar",
  components: {
    SidebarHeader,
    SidebarStep,
    SidebarFooter,
  },
  emits: {
    "navigate-to-step": (step: number, subStep: number) => step && subStep,
  },
  props: {
    currentStep: {
      type: Number as PropType<number>,
      required: true,
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
    stepsToShow: {
      type: Array as PropType<number[]>,
      default: () => [1, 2, 3, 4], // Show all steps by default
    },
    hiddenSubsteps: {
      type: Object as PropType<Record<number, number[]>>,
      default: () => ({}),
    },
  },

  setup(props, { emit }) {
    const handleNavigateToStep = (step: number, subStep: number) => {
      emit("navigate-to-step", step, subStep);
    };

    // Filter steps based on stepsToShow prop and hide specific substeps
    const filteredSteps = computed(() => {
      return ONBOARDING_SIDEBAR_CONFIG.steps
        .filter((step) => props.stepsToShow.includes(step.id))
        .map((step) => {
          const hiddenSubstepsForStep = props.hiddenSubsteps[step.id] || [];
          const filteredSubSteps = step.subSteps?.filter((subStep) => !hiddenSubstepsForStep.includes(subStep.id));
          return {
            ...step,
            subSteps: filteredSubSteps,
          };
        });
    });

    return {
      config: ONBOARDING_SIDEBAR_CONFIG,
      filteredSteps,
      handleNavigateToStep,
    };
  },
});
</script>

<style scoped lang="scss">
.onboarding-sidebar {
  width: 400px;
  background: #252531;
  color: white;
  display: flex;
  flex-direction: column;
  padding: 32px 24px;

  .steps-container {
    flex: 1;
  }

  .sidebar-header {
    margin-bottom: 51px;
  }
}
</style>
