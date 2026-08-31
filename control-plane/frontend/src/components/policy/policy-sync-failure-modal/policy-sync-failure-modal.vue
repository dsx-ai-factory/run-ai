<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="true" size="width-sm" @close="$emit('close')" is-modal-persistent>
    <template #header> Policy Couldn't Be Saved </template>

    <template #body>
      <section class="q-my-sm fs-16" aid="policy-sync-failure-modal">
        <div>The policy couldn't be saved due to a network or other unknown issue.</div>
        <div>
          <a
            class="link no-underline"
            aid="policy-sync-failure-modal-download"
            role="button"
            tabindex="0"
            @click="$emit('download')"
            @keyup.enter="$emit('download')"
          >
            Download YAML as a draft
          </a>
          and try pasting and saving it again later.
        </div>
      </section>
    </template>

    <template #footer>
      <q-btn label="GOT IT" aid="policy-sync-failure-modal-ok-button" color="primary" @click="$emit('close')" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

export default defineComponent({
  components: {
    RunaiBaseDialog,
  },
  emits: ["download", "close"],
});
</script>
<style lang="scss" scoped></style>
