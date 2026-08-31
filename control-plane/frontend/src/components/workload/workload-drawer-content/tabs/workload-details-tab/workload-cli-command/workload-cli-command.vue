<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <workload-details-sub-section-wrapper
    title="CLI command"
    tooltip-text="To submit a new workload based on the selected workload, copy the following command line to the terminal"
  >
    <p v-if="!cliCommand">CLI command is available only for workloads submitted via the CLI</p>
    <p v-else class="q-mb-md">{{ cliCommand }}</p>
    <span class="italic inline-block link-text"
      >To learn how to submit a workload via CLI, see the
      <a target="_blank" :href="DOCS.CLI.SUBMIT">CLI Commands Reference</a></span
    >
  </workload-details-sub-section-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import WorkloadDetailsSubSectionWrapper from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/workload-details-sub-section-wrapper/workload-details-sub-section-wrapper.vue";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "workload-cli-command",
  components: { WorkloadDetailsSubSectionWrapper },
  props: {
    cliCommand: {
      type: String as PropType<string>,
      default: "",
    },
  },
  data() {
    return {
      DOCS,
    };
  },
});
</script>

<style scoped lang="scss"></style>
