<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-file
    dense
    aid="chart-file-input"
    label="Chart file"
    accept=".tgz"
    :max-file-size="MAX_CHART_UPLOAD_BYTES"
    stack-label
    :clearable="!uploading"
    :disable="uploading"
    no-error-icon
    reactive-rules
    v-model="selectedFile"
    :hint="hintText"
    :rules="[chartFileRequiredRule, chartFileUploadedRule]"
    @rejected="onRejected"
  >
    <template v-slot:prepend>
      <q-icon name="ph-bold ph-cloud-arrow-up" size="20px" />
    </template>
    <template v-slot:append>
      <q-spinner v-if="uploading" color="primary" size="20px" aid="chart-uploading-spinner" />
    </template>
  </q-file>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { errorMessages } from "@/common/error-message.constant";

const MAX_CHART_UPLOAD_BYTES = 10 * 1024 * 1024;

const selectedFile = defineModel<File | null>("selectedFile", { default: null });

const props = withDefaults(
  defineProps<{
    chartId: string;
    uploading: boolean;
    uploadError: string;
  }>(),
  {
    chartId: "",
    uploading: false,
    uploadError: "",
  },
);

const emit = defineEmits<{
  rejected: [rejections: { failedPropValidation: string }[]];
}>();

const hintText = computed(() =>
  props.uploading ? errorMessages.CHART_UPLOAD_IN_PROGRESS : "Upload a file no larger than 10 MB",
);

function onRejected(rejections: { failedPropValidation: string }[]): void {
  emit("rejected", rejections);
}

function chartFileRequiredRule(): string | boolean {
  if (selectedFile.value) return true;
  return props.uploadError || errorMessages.CHART_FILE_REQUIRED;
}

function chartFileUploadedRule(): string | boolean {
  if (!selectedFile.value || props.uploading) return true;
  return !!props.chartId || props.uploadError || errorMessages.CHART_UPLOAD_FAILED;
}
</script>
