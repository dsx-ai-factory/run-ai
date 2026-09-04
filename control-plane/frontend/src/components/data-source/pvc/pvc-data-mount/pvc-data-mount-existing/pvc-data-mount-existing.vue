<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="pvc-data-mount-existing">
    <runai-select
      class="claim-select"
      standard
      :model-value="claimName"
      :options="pvcList"
      label="Claim"
      stack-label
      placeholder="Select..."
      :rules="[isClaimSelected]"
      @update:model-value="$emit('claim-name-changed', $event)"
      :loading="loading"
      no-option-text="No claims found"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSelect } from "@/components/common/runai-select";
// common
import { isNotEmpty } from "@/common/form.validators";

// Types
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  components: {
    RunaiSelect,
  },
  props: {
    claimName: {
      type: String as PropType<string>,
      required: true,
    },
    pvcList: {
      type: Array as PropType<string[]>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  methods: {
    isClaimSelected(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.CLAIM_NOT_EMPTY;
    },
  },
});
</script>
