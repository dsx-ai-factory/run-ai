<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-radio-options
    class="workload-architecture-select"
    :title="title"
    :model-value="isDistributed"
    @update:model-value="$emit('update', $event)"
    :options="distributedOptions"
    :disable="readOnly || disableSelection"
    :disable-tooltip="disableTooltip"
    tooltip-width="500px"
  />
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { EDistributedFormType } from "../distributed-selection.vue";
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  name: "workload-architecture-select",
  components: {
    RunaiRadioOptions,
  },
  emits: ["update"],
  props: {
    isDistributed: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    formType: {
      type: String as PropType<EDistributedFormType>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    disableSelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    distributedOptions(): ISelectOption[] {
      return [
        {
          value: false,
          label: "Standard",
          toolTip:
            this.formType === EDistributedFormType.ENVIRONMENT
              ? "Only standard workloads can use this environment. A standard workload consists of a single process."
              : "This workload will only be able to use environments that support standard workloads. A standard workload consists of a single process.",
        },
        {
          value: true,
          label: "Distributed",
          toolTip:
            this.formType === EDistributedFormType.ENVIRONMENT
              ? "Only distributed workloads can use this environment. A distributed workload consists of multiple processes working together. These processes can run on different nodes."
              : "This workload will only be able to use environments that support distributed workloads. A distributed workload consists of multiple processes working together. These processes can run on different nodes.",
        },
      ];
    },
    title(): string {
      return this.formType === EDistributedFormType.ENVIRONMENT
        ? "Set whether the environment supports standard or distributed workloads"
        : "Set whether the workload should run as a standard or distributed workload";
    },
    disableTooltip(): string {
      if (this.readOnly) return "This can't be modified when editing an environment";
      if (!this.disableSelection) return "";
      const selectionText = this.isDistributed ? "distributed" : "standard";
      return `This can't be modified when an environment is created via a ${selectionText} workload`;
    },
  },
});
</script>
