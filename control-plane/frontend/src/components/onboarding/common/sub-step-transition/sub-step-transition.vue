<template>
  <transition name="substep-transition" mode="out-in">
    <slot />
  </transition>
</template>
<script lang="ts">
import { defineComponent } from "vue";
export default defineComponent({
  name: "sub-step-transition",
});
</script>

<style scoped lang="scss">
// Sub-step transition animations
.substep-transition-enter-active,
.substep-transition-leave-active {
  transition: all 0.2s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

.substep-transition-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.substep-transition-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

.substep-transition-enter-to,
.substep-transition-leave-from {
  opacity: 1;
  transform: translateX(0);
}
</style>
