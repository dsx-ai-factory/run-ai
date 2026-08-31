<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="user-review-screen">
    <p class="text-b-subtitle2 credential-header">User credentials:</p>
    <div class="row items-center no-wrap credential-row">
      <span class="field-type">Email address: </span>
      <div class="row items-center flex-1 credential-value">
        <span class="flex-1 text-b-subtitle2">{{ name }}</span>
        <runai-copy-to-clipboard-btn :text-to-copy="name" />
      </div>
    </div>
    <div v-if="fullName" class="row items-center no-wrap credential-row">
      <span class="field-type">Full name: </span>
      <div class="row items-center flex-1 credential-value">
        <span class="flex-1 text-b-subtitle2">{{ fullName }}</span>
        <runai-copy-to-clipboard-btn :text-to-copy="fullName" />
      </div>
    </div>
    <div class="row items-center no-wrap credential-row">
      <span class="field-type">Temporary password: </span>
      <div class="row items-center flex-1 credential-value">
        <span class="flex-1 text-b-subtitle2">{{ isPasswordHidden ? "•".repeat(password.length) : password }}</span>
        <runai-btn
          class="runai-icon-btn--default"
          flat
          round
          color="black-54"
          :icon="isPasswordHidden ? 'ph-fill ph-eye-slash' : 'ph-fill ph-eye'"
          @click="handleTogglePasswordVisibility"
          aid="toggle-password"
          :tooltip="isPasswordHidden ? 'Show password' : 'Hide password'"
        />
        <runai-copy-to-clipboard-btn :text-to-copy="password" />
      </div>
    </div>
    <!-- <q-checkbox class="check-box" v-model="isSendToEmail" :label="getSendEmailLabel" />  // Will enable this action once notification service is ready to use -->
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { type TActionTypes } from "../action-type.model";
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn/";
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  components: { RunaiBtn, RunaiCopyToClipboardBtn },
  props: {
    actionType: {
      type: String as PropType<TActionTypes>,
      required: true,
    },
    name: {
      type: String,
      required: true,
    },
    password: {
      type: String,
      required: true,
    },
    firstName: {
      type: String,
      default: "",
    },
    lastName: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      isPasswordHidden: true as boolean,
      isSendToEmail: false as boolean,
    };
  },
  computed: {
    fullName(): string {
      return [this.firstName, this.lastName].filter(Boolean).join(" ");
    },
  },
  methods: {
    handleTogglePasswordVisibility(): void {
      this.isPasswordHidden = !this.isPasswordHidden;
    },
  },
});
</script>
<style lang="scss" scoped>
.user-review-screen {
  margin: -24px;

  .credential-header {
    height: 60px;
    padding: 16px;
    display: flex;
    align-items: center;
  }
  .credential-row {
    height: 60px;
    padding: 16px;
    gap: 20px;
  }
  .field-type {
    flex: 1 0 0;
    color: $black-54;
    white-space: nowrap;
  }
  .credential-value {
    gap: 16px;
    min-width: 0;

    span {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
      min-width: 0;
    }
  }
}
</style>
