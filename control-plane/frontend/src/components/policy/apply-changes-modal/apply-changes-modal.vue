<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    :model-value="true"
    size="width-sm"
    @close="$emit('discard')"
    :on-enter-callback="() => $emit('apply')"
  >
    <template #header> Apply Changes? </template>

    <template #body>
      <div>Changing the scope and/or workload type will reset the policy YAML.</div>
    </template>

    <template #footer>
      <q-btn label="Discard changes" aid="apply-changes-modal-stay-btn" flat color="primary" @click="$emit('discard')" />
      <q-btn label="Apply changes" aid="apply-changes-modal-leave-btn" color="primary" @click="$emit('apply')" />
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
  emits: ["apply", "discard"],
});
</script>
<style lang="scss" scoped></style>
