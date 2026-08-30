<template>
  <div class="onboarding-content-main">
    <div ref="scrollableContainer" class="onboarding-content-container">
      <slot />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "onboarding-content",
});
</script>

<style scoped lang="scss">
.onboarding-content-main {
  flex: 1;
  overflow: hidden;
  .onboarding-content-container {
    height: 100%;
    padding: 80px 40px;
    overflow: auto;

    &:has(.completion-content-body) {
      padding-bottom: 0;

      .content-body {
        padding-top: 0;
      }
    }
  }
}
</style>
