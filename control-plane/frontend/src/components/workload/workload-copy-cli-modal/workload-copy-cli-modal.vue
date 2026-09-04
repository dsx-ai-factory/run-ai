<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-copy-cli-modal">
    <runai-base-dialog :model-value="true" @close="onCancel" size="width-lg" no-padding>
      <template #header
        ><span class="">Make a Copy of {{ workloadName }}</span></template
      >
      <template #body>
        <div class="row">
          <div class="col-10">
            <div class="q-px-lg q-pt-md column">
              <div>The selected workload was created via the CLI.</div>
              <div class="q-mb-md">
                To submit a new workload based on this, copy the following command line to the terminal
              </div>
              <runai-copyable-code-block class="q-mb-lg" :code="cliCommand" />
            </div>
          </div>
        </div>
      </template>
      <q-separator />
      <template #footer-left> </template>
      <template #footer>
        <div class="workload-copy-cli-actions q-px-md">
          <span class="italic inline-block link-text">
            To learn how to submit a workload via CLI, see the CLI
            <a href="https://docshub.run.ai/guides/reference/cli/runai" target="_blank">v2</a> and
            <a href="https://docs.run.ai/latest/Researcher/cli-reference/runai-submit/" target="_blank">v1</a> reference
          </span>
          <q-btn label="Close" color="primary" @click="onCancel" aid="close-workload-copy-cli-modal" />
        </div>
      </template>
    </runai-base-dialog>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiCopyableCodeBlock } from "@/components/common/runai-copyable-code-block/";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "workload-copy-cli-modal",
  components: { RunaiCopyableCodeBlock, RunaiBaseDialog },
  emits: ["cancel"],
  props: {
    workloadName: {
      type: String as PropType<string>,
      required: true,
    },
    cliCommand: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      DOCS,
    };
  },
  methods: {
    onCancel(): void {
      this.$emit("cancel");
    },
  },
});
</script>

<style scoped lang="scss">
.workload-copy-cli-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex: 1;
}
</style>
