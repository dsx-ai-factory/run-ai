<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="distributed-selection q-mb-sm">
    <section class="row">
      <workload-architecture-select
        :is-distributed="distributedModel.isDistributed"
        :form-type="formType"
        :read-only="readOnly"
        :disable-selection="disableSelection"
        @update="updateDistributed"
      />
    </section>
    <template v-if="distributedModel.isDistributed && formType !== 'environment'">
      <distributed-framework-selection-section
        class="q-mt-lg"
        :dist-framework="distFramework"
        :cluster-id="clusterId"
        :read-only="readOnly"
        :disable-selection="disableSelection"
        @update="updateDistributedFramework"
      />
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { WorkloadArchitectureSelect } from "./workload-architecture-select";
import { DistributedFrameworkSelectionSection } from "./distributed-framework-selection-section";

// Models
import type { DistributedFramework } from "@/swagger-models/assets-service-client";
import type { IDistributedSelectionModel } from "./distributed-selection.model";

export enum EDistributedFormType {
  ENVIRONMENT = "environment",
  TRAINING = "training",
}

export default defineComponent({
  name: "distributed-selection",
  components: {
    WorkloadArchitectureSelect,
    DistributedFrameworkSelectionSection,
  },
  emits: ["update-distributed-model"],
  props: {
    distributedModel: {
      type: Object as PropType<IDistributedSelectionModel>,
      required: true,
    },
    disableSelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    formType: {
      type: String as PropType<EDistributedFormType>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  computed: {
    distFramework(): DistributedFramework | null {
      return this.distributedModel?.distFramework || null;
    },
  },
  methods: {
    updateDistributed(isDistributed: boolean): void {
      const distFramework: DistributedFramework | null = isDistributed ? this.distributedModel.distFramework : null;
      this.$emit("update-distributed-model", { isDistributed, distFramework });
    },
    updateDistributedFramework(distFramework: DistributedFramework): void {
      if (!this.distributedModel.isDistributed) return;
      this.$emit("update-distributed-model", { ...this.distributedModel, distFramework });
    },
  },
});
</script>
