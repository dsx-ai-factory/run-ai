<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    :class="{ 'department-name-section--compact': compact }"
    :label="compact ? 'Department name and description' : 'Department name & description'"
    default-opened
    :section-invalid="sectionInvalid"
    :body-no-padding="compact"
    :disable-closing="compact"
  >
    <template #subheader>
      <span>
        {{ departmentName || "&lt;Department name&gt;" }}
      </span>
    </template>
    <div class="row">
      <div class="department-name-column" :class="compact ? 'col-8' : 'col-7'">
        <runai-tooltip-wrapper :display-tooltip="isInputDisabled" tooltip-text="This can't be modified">
          <runai-name-validation
            aid="department-name-input"
            :auto-focus="autoFocus"
            :label="compact ? '' : 'Name'"
            :placeholder="compact ? 'Name' : undefined"
            :aria-label="compact ? 'Name' : undefined"
            :disable="isInputDisabled"
            :model-value="departmentName"
            @update:model-value="$emit('update:departmentName', $event)"
            :rules="compact ? compactNameRules : [isValidEntityName, isNameUnique]"
          />
        </runai-tooltip-wrapper>
      </div>
    </div>

    <div class="row" :class="compact ? 'compact-description-row' : 'q-mt-md'">
      <div :class="compact ? 'col-12' : 'col-11'">
        <q-input
          aid="department-description-input"
          :model-value="description"
          :disable="readOnly"
          @update:model-value="$emit('update:description', $event)"
          :maxlength="250"
          counter
          :label="compact ? undefined : 'Description'"
          :placeholder="compact ? 'Description' : undefined"
          :aria-label="compact ? 'Description' : undefined"
          autogrow
        />
      </div>
    </div>
  </runai-expansion-item>
</template>
<script lang="ts">
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { defineComponent, type PropType } from "vue";
import type { ValidationRule } from "quasar";
import { isValidEntityName } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

export default defineComponent({
  name: "department-name-section",
  components: {
    RunaiNameValidation,
    RunaiExpansionItem,
    RunaiTooltipWrapper,
  },
  emits: ["update:departmentName", "is-section-invalid", "update:description"],
  props: {
    compact: { type: Boolean, default: false },
    autoFocus: { type: Boolean, default: false },
    departmentName: {
      type: String as PropType<string>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    description: {
      type: String as PropType<string>,
      required: false,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      isDepartmentNameUnique: true as boolean,
    };
  },
  computed: {
    compactNameRules(): ValidationRule[] {
      // Keep async name validation stable while description and role fields change.
      return [this.isValidEntityName, this.isNameUnique];
    },
    sectionInvalid(): boolean {
      return !isValidEntityName(this.departmentName) || !this.isDepartmentNameUnique;
    },
    isInputDisabled(): boolean {
      return this.readOnly || this.isEditMode;
    },
  },
  methods: {
    isValidEntityName(name: string): boolean | string {
      return isValidEntityName(name) || errorMessages.VALID_FORMAT;
    },
    async isNameUnique(name: string): Promise<boolean | string> {
      try {
        const isNameUnique = await orgUnitService.isDepartmentNameUnique(this.clusterId, name);
        this.isDepartmentNameUnique = isNameUnique;
        if (!isNameUnique) {
          return errorMessages.NAME_ALREADY_EXIST;
        }
        return true;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedToValidateMessage("department name"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error(error);
        return false;
      }
    },
  },
  watch: {
    sectionInvalid: {
      handler(isSectionInvalid: boolean): void {
        this.$emit("is-section-invalid", isSectionInvalid);
      },
      immediate: true,
    },
  },
});
</script>

<style scoped lang="scss">
.department-name-section--compact {
  padding: 0 !important;
}
.department-name-section--compact :deep(> .q-expansion-item__container) {
  border: 0;
  box-shadow: none;
}
.department-name-section--compact :deep(.runai-expansion-header) {
  height: 32px;
  min-height: 32px;
  padding: 0;
  font-weight: 400;
  line-height: 20px;
}
.department-name-section--compact :deep(.q-item__section--side),
.department-name-section--compact :deep(.runai-expansion-divider),
.department-name-section--compact :deep(.q-expansion-item__border) {
  display: none;
}
.department-name-section--compact :deep(.runai-expansion-body) {
  padding: 26px 0 0 !important;
}
.department-name-section--compact .department-name-column {
  width: 510px;
  max-width: 100%;
}
.department-name-section--compact .compact-description-row {
  margin-top: 26px;
}
.department-name-section--compact :deep(.q-field__control) {
  height: auto;
  min-height: 31px;
}
.department-name-section--compact :deep(.q-field__marginal) {
  height: 31px;
}
.department-name-section--compact :deep(.q-field__native) {
  min-height: 0;
  font-size: 14px;
  line-height: 19.5px;
  padding: 0 0 10px;
}
.department-name-section--compact :deep(.q-field__native::placeholder) {
  color: var(--b-text-sub);
  opacity: 1;
}
.department-name-section--compact :deep(.q-field--error .q-field__native::placeholder) {
  color: var(--q-negative);
}
.department-name-section--compact :deep(.q-field--with-bottom) {
  padding-bottom: 0;
}
.department-name-section--compact :deep(.q-field__bottom) {
  position: static;
  transform: none;
  bottom: auto;
  min-height: 0;
  padding: 7px 0 0;
}
.department-name-section--compact :deep(.q-field__messages),
.department-name-section--compact :deep(.q-field__counter) {
  font-size: 12px;
  line-height: 16.8px;
}
</style>
