<template>
  <div class="runai-avatar" :class="sizeClass">
    <span :class="typographyClass">{{ initial }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

export type AvatarSize = "x-small" | "default" | "large";

const props = withDefaults(
  defineProps<{
    initial: string;
    size?: AvatarSize;
  }>(),
  { size: "default" },
);

const sizeClass = computed(() => `runai-avatar--${props.size}`);
const typographyClass = computed(() => (props.size === "x-small" ? "text-b-overline" : "text-b-h6"));
</script>

<style scoped lang="scss">
.runai-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background-color: var(--decorative-dark);
  padding: var(--space-s);

  span {
    color: var(--text-inverse);
  }

  &--x-small {
    width: 18px;
    height: 18px;
  }

  &--default {
    width: 36px;
    height: 36px;
  }

  &--large {
    width: 76px;
    height: 76px;
  }
}
</style>
