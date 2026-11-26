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

:deep(.wizard-step--selected .wizard-step-indicator),
:deep(.wizard-step--completed .wizard-step-indicator) {
  background-color: var(--interactive-stepper-completed-icon, #689f38) !important;
}

:is(.identity-wizard, .node-pool-edit-form-v2) :deep(.q-field__label) {
  font-size: 14px;
  line-height: 19.5px;
}

:is(.identity-wizard, .node-pool-edit-form-v2) :deep(.q-field--float .q-field__label) {
  // Quasar scales a floating label to 75% of its layout size.
  font-size: 16px;
  line-height: 22.4px;
}
</style>
