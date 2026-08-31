<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="email-verify">
    <setting-readonly-input :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
      <template v-slot:default="{ readonly }">
        <runai-verify-button
          :verify-status="verifyStatus"
          :validation-message="validationMessage"
          @verify="verify"
          :disabled="readonly"
        />
      </template>
    </setting-readonly-input>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiVerifyButton } from "@/components/common/runai-verify-button";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";

// Models
import type { ERunaiVerifyButtonStatus } from "@/components/common/runai-verify-button";

export default defineComponent({
  name: "verify-button",
  components: { SettingReadonlyInput, RunaiVerifyButton },
  emits: ["verify"],
  props: {
    verifyStatus: {
      type: String as PropType<ERunaiVerifyButtonStatus>,
      required: true,
    },
    validationMessage: {
      type: String as PropType<string>,
      required: false,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: undefined,
    },
    displayReadOnlyTooltip: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  methods: {
    verify(): void {
      this.$emit("verify");
    },
  },
});
</script>
