<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-dropdown
    custom-classes="btn-icon-sm runai-icon-btn--default min-w-173px"
    color="primary"
    :options="newReportOptions"
    icon="ph-bold ph-plus"
    label="new report"
    :flat="flat"
    @on-selected="$emit('create', $event)"
  />
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RunaiDropdown } from "@/components/common";
//model
import type { ISelectOption } from "@/models/global.model";
import { EReportType } from "@/models/report.model";

export default defineComponent({
  name: "create-report-dropdown",
  components: { RunaiDropdown },
  emits: ["create"],
  props: {
    flat: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    newReportOptions(): ISelectOption[] {
      return [{ label: "Consumption", value: EReportType.Consumption }];
    },
  },
});
</script>

<style lang="scss"></style>
