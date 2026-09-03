<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="runai-verify-button">
    <q-btn
      aid="verify-button"
      class="verify-button"
      :class="{ verified: verified, failed: verifyingFailed, pending: noneVerify || verifying }"
      :loading="verifying"
      color="grey-4"
      text-color="b-main"
      @click="verify"
      :disable="disabled"
    >
      <div v-if="!verifying">
        <q-icon name="ph-fill ph-shield-check" size="20px" class="q-mr-sm" />
        {{ verifyButtonLabel }}
      </div>
      <template v-slot:loading>
        <q-spinner class="on-left" />
        VERIFYING
      </template>
    </q-btn>
    <div class="validation q-mt-md" :class="{ 'verified-message': verified, 'non-verified-message': !verified }">
      {{ validationMessage }}
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Models
import { ERunaiVerifyButtonStatus } from "./runai-verify-button.model";

export default defineComponent({
  name: "runai-verify-button",
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
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  computed: {
    verified(): boolean {
      return this.verifyStatus === ERunaiVerifyButtonStatus.Verified;
    },
    verifying(): boolean {
      return this.verifyStatus === ERunaiVerifyButtonStatus.Verifying;
    },
    noneVerify(): boolean {
      return this.verifyStatus === ERunaiVerifyButtonStatus.None;
    },
    verifyingFailed(): boolean {
      return this.verifyStatus === ERunaiVerifyButtonStatus.Failed;
    },
    verifyButtonLabel(): string {
      switch (this.verifyStatus) {
        case ERunaiVerifyButtonStatus.Verified:
          return "Verified";
        case ERunaiVerifyButtonStatus.Verifying:
          return "Verifying";
        case ERunaiVerifyButtonStatus.Failed:
          return "Failed";
        default:
          return "Verify";
      }
    },
  },
  methods: {
    verify(): void {
      this.$emit("verify");
    },
  },
});
</script>

<style scoped lang="scss">
.runai-verify-button {
  .verify-button {
    font-size: 14px;
    width: 130px;
    height: 36px;
    display: flex;
  }
}
</style>
<style lang="scss">
.runai-verify-button {
  .verified .q-icon {
    color: $success;
  }
  .failed .q-icon {
    color: $negative;
  }
  .noneVerify .q-icon {
    color: $black-87;
  }
  .validation {
    font-size: 12px;
    height: 18px;
  }
  .verified-message {
    color: $success;
  }
  .non-verified-message {
    color: $negative;
  }
}
</style>
