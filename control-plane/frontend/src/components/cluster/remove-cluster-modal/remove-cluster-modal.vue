<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="true" @close="onCancel" size="width-md" :no-padding="true">
    <template #header>Remove</template>

    <template #body>
      <div class="remove-cluster-body q-px-lg q-py-md column">
        <p class="text-b-h6-md q-mt-md">Remove: {{ clusterName }}?</p>
        <div class="main-content row no-wrap">
          <q-icon class="q-mr-lg q-mt-xs" color="red" size="20px" name="ph-bold ph-warning"></q-icon>
          <div class="remove-modal-info">
            <div class="q-mb-lg">
              <div>Once a cluster is removed, it cannot be restored.</div>
              <div>
                All the associated departments, projects, workloads, environments, data sources, compute resources,
                templates, credentials, event history, node pools and dashboards will be permanently deleted.
              </div>
            </div>
            <div class="q-mb-md">The Run:ai cluster installation will remain on the Kubernetes cluster.</div>
            <div class="q-mb-md italic">
              For more information, see
              <a :href="DOCS.INSTALLATION.UNINSTALL" target="_blank">Uninstall guide</a>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #footer-left> </template>
    <template #footer>
      <runai-btn label="Cancel" color="primary" flat @click="onCancel" aid="remove-modal-cancel-button" />
      <runai-btn label="Remove" color="primary" :loading="deleting" @click="onRemove" aid="remove-modal-remove-button" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { DOCS } from "@/constants/docs";

// Components
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiBtn } from "@/components/common/runai-btn";

export enum Events {
  cancel = "cancel",
  remove = "remove",
}
export default defineComponent({
  name: "remove-cluster-modal",
  components: {
    RunaiBaseDialog,
    RunaiBtn,
  },
  emits: [Events.cancel, Events.remove],
  props: {
    clusterName: {
      type: String as PropType<string>,
      required: true,
    },
    deleting: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      DOCS,
    };
  },
  methods: {
    onRemove(): void {
      this.$emit(Events.remove);
    },
    onCancel(): void {
      this.$emit(Events.cancel);
    },
  },
});
</script>

<style lang="scss" scoped>
.remove-cluster-body {
  a:hover {
    text-decoration: underline;
  }
  .remove-modal-info {
    font-size: 14px;
    line-height: 20px;
  }
}
</style>
