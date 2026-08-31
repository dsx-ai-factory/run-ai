<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-radio-options
    class="distributed-replica-configuration"
    title="Set the distributed workload configuration"
    title-tooltip="The configuration defines how distributed training workloads are divided across multiple machines or processes. Choose a configuration based on your training requirements and infrastructure."
    :options="noMasterOptions"
    :model-value="modelValue"
    @update:model-value="$emit('update:model-value', $event)"
  />
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  name: "distributed-replica-configuration",
  components: {
    RunaiRadioOptions,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      noMasterOptions: [
        { value: false, label: "Workers & master" },
        { value: true, label: "Workers only" },
      ] as Array<ISelectOption>,
    };
  },
});
</script>
