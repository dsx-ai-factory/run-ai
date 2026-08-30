<template>
  <div class="content-footer">
    <div class="navigation-buttons">
      <q-btn
        v-if="showSkipButton"
        flat
        :disable="isLoading"
        label="SKIP"
        @click="$emit('skip')"
        class="skip-button q-mr-md"
        color="primary"
      />
      <q-btn
        :disable="isLoading"
        :loading="isLoading"
        :label="continueButtonLabel"
        @click="$emit('next')"
        color="primary"
        class="continue-button"
      />
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";

export default defineComponent({
  name: "onboarding-navigation",
  emits: {
    next: () => true,
    skip: () => true,
  },
  props: {
    continueButtonLabel: {
      type: String as PropType<string>,
      default: "Continue",
    },
    isLoading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showSkipButton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>

<style scoped lang="scss">
.content-footer {
  padding: 15px 20px;
  box-shadow: 0;
  transition: box-shadow 0.125s ease;

  .navigation-buttons {
    display: flex;
    justify-content: end;
    align-items: center;
    margin: 0 auto;
  }

  .skip-button {
    font-weight: 500;
  }
}
</style>
