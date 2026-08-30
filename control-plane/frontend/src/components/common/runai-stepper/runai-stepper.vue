<template>
  <section ref="stepperContainer" class="runai-stepper" :class="{ 'hide-stepper': steps.length === 1 }">
    <q-stepper
      class="stepper"
      :class="{ 'hide-stepper': steps.length === 1 }"
      v-model="step"
      active-icon="ph-fill ph-pencil-simple"
      done-icon="ph-bold ph-check"
      :header-nav="headerNavigation"
      vertical
      ref="stepper"
      color="primary"
      animated
      flat
      style="background-color: transparent"
      @transition="scrollToTop"
    >
      <q-step
        v-for="step in steps"
        :key="step.name"
        :name="step.name"
        :done="step.done"
        :title="step.title"
        :icon="step.icon"
      />
    </q-stepper>
    <div class="runai-step-content">
      <slot name="step"></slot>
    </div>

    <div class="runai-step-navigation">
      <slot name="actions"></slot>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import type { IStepsConfig } from "./runai-stepper.model";

export default defineComponent({
  emits: ["step-changed"],
  props: {
    currentStep: {
      type: String as PropType<string>,
      required: true,
    },
    steps: {
      type: Array as PropType<Array<IStepsConfig>>,
      required: true,
    },
    headerNavigation: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  data() {
    return {
      step: "" as string,
      stepsPriority: new Map<string, number>(),
    };
  },
  created() {
    this.stepsPriority = this.steps.reduce((acc: Map<string, number>, step: IStepsConfig, index: number) => {
      acc.set(step.name, index);
      return acc;
    }, new Map<string, number>());
  },
  methods: {
    scrollToTop(): void {
      const stepperContainer = this.$refs.stepperContainer as HTMLElement;
      stepperContainer.scrollIntoView({ behavior: "smooth", block: "start" });
    },
  },
  watch: {
    step() {
      // we want to prevent from the user from clicking the next steps from the navigation
      if ((this.stepsPriority.get(this.step) as number) > (this.stepsPriority.get(this.currentStep) as number)) {
        this.step = this.currentStep;
      } else {
        this.$emit("step-changed", this.step);
      }
    },
    currentStep: {
      immediate: true,
      handler() {
        this.step = this.currentStep;
      },
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-stepper {
  display: grid;
  grid-template-areas:
    "stepper step-content"
    "footer footer";
  grid-template-columns: minmax(150px, min(100%, 250px)) 1fr;
  grid-template-rows: auto;
  gap: 10px;
  &.hide-stepper {
    grid-template-areas:
      "step-content"
      "footer";
    grid-template-columns: 1fr;
  }

  .stepper {
    grid-area: stepper;
    position: fixed;

    :deep(.q-stepper__dot .q-icon) {
      font-size: 18px;
    }

    &.hide-stepper {
      display: none;
    }
  }

  .runai-step-content {
    grid-area: step-content;
    padding: 20px;
  }

  .runai-step-navigation {
    grid-area: footer;
    display: flex;
    justify-content: flex-end;
    padding: 5px 20px;
  }
}
</style>
