<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section :title="title">
    <div class="row">
      <policy-number-input
        class="col-3"
        type="number"
        :model-value="workers"
        @update:model-value="$emit('update', $event)"
        stack-label
        :min-value="1"
        :policy-rules="policyRules"
        :disable="sectionDisabled"
        custom-ref="workersInput"
        aid="workers-input"
      />
    </div>
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </runai-section>
</template>

<script lang="ts">
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { RunaiSection } from "@/components/common/runai-section";
import type { IntegerRules } from "@/swagger-models/assets-service-client";
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "workers-input",
  components: {
    PolicyNumberInput,
    RunaiSection,
  },
  emits: ["update"],
  props: {
    workers: {
      type: Number as PropType<number>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<IntegerRules | null>,
      required: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    title: {
      type: String as PropType<string>,
      required: false,
      default: "Set the number of workers for your workload",
    },
  },
});
</script>
