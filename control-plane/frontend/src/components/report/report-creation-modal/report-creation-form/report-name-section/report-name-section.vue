<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="row items-center">
    <div class="col-4">
      <runai-name-validation
        aid="report-name-input"
        :model-value="reportName"
        placeholder="Enter a name"
        :rules="[isEmpty, isReportNameUnique]"
        stack-label
        label="Report name"
        @update:model-value="$emit('update:reportName', $event)"
      />
    </div>
  </div>
  <div class="row">
    <div class="col-7">
      <q-input
        aid="report-description-input"
        :model-value="description"
        @update:model-value="$emit('update:description', $event)"
        :maxlength="250"
        counter
        autogrow
        label="Description"
      />
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import RunaiNameValidation from "@/components/common/runai-name-validation/runai-name-validation.vue";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ErrorAlert } from "@/utils/error-alert.util";
//cmps

//model

export default defineComponent({
  name: "report-name-section",
  components: { RunaiNameValidation },
  emits: ["update:reportName", "update:description"],
  props: {
    reportName: {
      type: String as PropType<string>,
      required: true,
    },
    description: {
      type: [String, null] as PropType<string | null>,
      required: false,
    },
  },
  methods: {
    isEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    async isReportNameUnique(name: string): Promise<boolean | string> {
      try {
        const isNameUnique = await orgUnitService.isReportNameUnique(name);
        if (!isNameUnique) {
          return errorMessages.NAME_ALREADY_EXIST;
        }
        return true;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedToValidateMessage("report name"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        return false;
      }
    },
  },
});
</script>

<style scoped lang="scss"></style>
