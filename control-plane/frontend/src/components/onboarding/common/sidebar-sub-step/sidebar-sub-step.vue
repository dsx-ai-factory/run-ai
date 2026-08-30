<template>
  <div
    class="sub-step"
    :class="{
      active: isActive,
      completed: isCompleted,
      clickable: isClickable,
    }"
    role="button"
    tabindex="0"
    :aria-disabled="!isClickable"
    :aria-current="isActive ? 'step' : undefined"
    @click="handleClick"
    @keydown.enter.prevent="handleClick"
    @keydown.space.prevent="handleClick"
  >
    {{ subStep.title }}
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { OnboardingSubStep } from "@/components/onboarding/common/onboarding-sidebar/types";

export default defineComponent({
  name: "sidebar-sub-step",
  emits: {
    click: () => true,
  },
  props: {
    subStep: {
      type: Object as PropType<OnboardingSubStep>,
      required: true,
    },
    isActive: {
      type: Boolean,
      default: false,
    },
    isCompleted: {
      type: Boolean,
      default: false,
    },
    isClickable: {
      type: Boolean,
      default: false,
    },
  },
  methods: {
    handleClick(): void {
      if (this.isClickable) {
        this.$emit("click");
      }
    },
  },
});
</script>

<style scoped lang="scss">
.sub-step {
  font-weight: 400;
  font-size: 14px;
  line-height: 25px;

  color: #9a9696;
  margin-bottom: 12px;
  padding-left: 0;
  position: relative;
  transition: all 0.3s ease;

  &.active {
    color: var(--primary-color);
    font-weight: 500;
  }

  &.completed {
    color: #ffffffb2;
  }

  &.clickable {
    cursor: pointer;

    &:hover {
      color: white;
      transform: translateX(2px);
    }
  }

  // Specific hover for completed clickable substeps
  &.completed.clickable:hover {
    color: white; // White 100% for completed substeps on hover
  }
}
</style>
