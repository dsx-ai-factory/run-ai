<template>
  <div
    v-if="bannerConfig?.enabled"
    class="custom-banner q-pa-md"
    :style="{
      backgroundColor: bannerConfig.backgroundColor || undefined,
      color: bannerConfig.fontColor || undefined,
      fontSize: `${Math.min(bannerConfig.fontSize ?? DEFAULT_FONT_SIZE, MAX_FONT_SIZE)}px`,
    }"
  >
    <p class="custom-banner-text">
      {{ bannerConfig.text }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from "pinia";
import { useSettingStore } from "@/stores/setting.store";

const DEFAULT_FONT_SIZE = 14;
const MAX_FONT_SIZE = 20;

const { bannerConfig } = storeToRefs(useSettingStore());
</script>

<style lang="scss" scoped>
.custom-banner {
  height: 60px;
  overflow-y: auto;
  font-family: Roboto, sans-serif;
  font-weight: 600;
  z-index: 100;
  &-text {
    text-align: center;
    text-transform: uppercase;
    word-break: break-word;
    margin: 0;
  }
}
</style>
