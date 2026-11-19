<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <section class="scheduling-department-name-fields" aria-label="Department name and description">
    <div class="department-name-title">Department name and description</div>
    <runai-name-validation
      class="department-name-input"
      aid="department-name-input"
      auto-focus
      label=""
      placeholder="Name"
      aria-label="Name"
      :model-value="departmentName"
      @update:model-value="emit('update:departmentName', $event)"
      :rules="nameRules"
    />
    <q-input
      class="department-description-input"
      aid="department-description-input"
      :model-value="description"
      @update:model-value="emit('update:description', String($event ?? ''))"
      :maxlength="250"
      counter
      placeholder="Description"
      aria-label="Description"
      autogrow
    />
  </section>
</template>
<script setup lang="ts">
import { useQuasar, type ValidationRule } from "quasar";
import { RunaiNameValidation } from "@/components/common/runai-name-validation/index";
import { isValidEntityName } from "@/common/form.validators/index";
import { errorMessages } from "@/common/error-message.constant";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ErrorAlert } from "@/utils/error-alert.util/index";

const props = defineProps<{ departmentName: string; description?: string; clusterId: string }>();
const emit = defineEmits<{
  "update:departmentName": [value: string];
  "update:description": [value: string];
}>();
const $q = useQuasar();
async function isNameUnique(name: string): Promise<boolean | string> {
  try {
    return (await orgUnitService.isDepartmentNameUnique(props.clusterId, name)) || errorMessages.NAME_ALREADY_EXIST;
  } catch (error: unknown) {
    const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedToValidateMessage("department name") });
    $q.notify(errorAlert.getNotification(error));
    console.error(error);
    return false;
  }
}
// Keep async validation stable while description and administrator fields change.
const nameRules: ValidationRule[] = [
  (name: string) => isValidEntityName(name) || errorMessages.VALID_FORMAT,
  isNameUnique,
];
</script>
<style scoped lang="scss">
.scheduling-department-name-fields {
  margin-bottom: 26px;
}
.department-name-title {
  display: flex;
  align-items: center;
  height: 32px;
  line-height: 21px;
}
.department-name-input {
  width: 510px;
  max-width: 100%;
  margin-top: 26px;
}
.department-description-input {
  margin-top: 26px;
}
.scheduling-department-name-fields :deep(.q-field__control) {
  height: auto;
  min-height: 31px;
}
.scheduling-department-name-fields :deep(.q-field__marginal) {
  height: 31px;
}
.scheduling-department-name-fields :deep(.q-field__native) {
  min-height: 0;
  font-size: 14px;
  line-height: 21px;
  padding: 0 0 10px;
}
.scheduling-department-name-fields :deep(.q-field__native::placeholder) {
  color: var(--b-text-sub);
  opacity: 1;
}
.scheduling-department-name-fields :deep(.q-field--error .q-field__native::placeholder) {
  color: var(--q-negative);
}
.scheduling-department-name-fields :deep(.q-field--with-bottom) {
  padding-bottom: 0;
}
.scheduling-department-name-fields :deep(.q-field__bottom) {
  position: static;
  transform: none;
  bottom: auto;
  min-height: 0;
  padding: 7px 0 0;
}
.scheduling-department-name-fields :deep(.q-field__messages),
.scheduling-department-name-fields :deep(.q-field__counter) {
  font-size: 12px;
  line-height: 16.8px;
}
</style>
