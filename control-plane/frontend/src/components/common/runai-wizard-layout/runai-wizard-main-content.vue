<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="wizard-main-content">
    <div class="wizard-content-scroll">
      <slot />
    </div>

    <!-- Content Footer -->
    <div v-if="$slots['content-footer']" class="wizard-content-footer">
      <slot name="content-footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: "runai-wizard-main-content",
});
</script>

<style scoped lang="scss">
.wizard-main-content {
  display: flex;
  flex-direction: column;
  min-height: 200px;
}

.wizard-content-scroll {
  flex: 1;
}

// Content Footer
.wizard-content-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
