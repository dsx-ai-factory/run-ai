<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <router-view :key="`${String($route.name)}:${$route.params.id ?? ''}`" />
</template>

<style scoped lang="scss">
:deep([aria-current="step"] .wizard-step-title) {
  color: var(--interactive-stepper-completed-text, #4a4a4a) !important;
  line-height: 20px !important;
}
</style>
