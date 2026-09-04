<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="access-key-box q-pa-md row items-center">
    <div class="col-10">
      <div class="row">
        <q-input
          aid="user-app-name"
          class="col-3 q-mr-md"
          label="Access key name"
          stack-label
          :model-value="accessKey.name"
          no-error-icon
          :disable="true"
        />
        <q-input
          aid="user-app-creation-time"
          class="col-4 q-mr-md"
          label="Creation time"
          stack-label
          :model-value="createdAt"
          no-error-icon
          :disable="true"
        />
        <q-input
          aid="user-app-last-login"
          class="col-4"
          label="Last login"
          stack-label
          :model-value="lastLogin"
          no-error-icon
          :disable="true"
        />
      </div>

      <div class="row q-mt-md">
        <q-input
          aid="user-app-client-id"
          class="col-3"
          label="Client ID"
          stack-label
          :model-value="clientId"
          no-error-icon
          :disable="true"
        />
        <div class="col-1 flex items-center"><runai-copy-to-clipboard-btn :text-to-copy="clientId" /></div>
      </div>
    </div>

    <div class="col-2 row items-center justify-center">
      <div class="row items-center">
        <runai-btn
          aid="regenerate-password-user-app-button"
          icon="ph-fill ph-lock"
          flat
          round
          @click="regeneratePassword"
          class="icon-button runai-icon-btn--default"
          color="black-54"
          tooltip="Regenerate client secret"
        />
        <runai-btn
          aid="remove-user-app-button"
          icon="ph-fill ph-trash"
          flat
          round
          @click="removeAccessKey"
          class="icon-button runai-icon-btn--default"
          color="black-54"
          tooltip="Delete access key"
        />
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
// Models
// Utils
import type { AccessKey } from "@/swagger-models/identity-manager-client";
import { dateUtil } from "@/utils/date.util";
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn";
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  name: "access-key-box",
  components: { RunaiBtn, RunaiCopyToClipboardBtn },
  emits: ["remove-access-key", "regenerate-password"],
  props: {
    accessKey: {
      type: Object as PropType<AccessKey>,
      required: true,
    },
  },
  computed: {
    createdAt(): string {
      return this.accessKey.createdAt ? dateUtil.dateAndTimeFormat(new Date(this.accessKey.createdAt)) : "";
    },
    lastLogin(): string {
      return this.accessKey.lastLogin
        ? dateUtil.dateAndTimeFormat(new Date(this.accessKey.lastLogin as string))
        : "Never";
    },
    clientId(): string {
      return this.accessKey.clientId || "";
    },
  },
  methods: {
    removeAccessKey(): void {
      this.$emit("remove-access-key", this.accessKey);
    },
    regeneratePassword(): void {
      this.$emit("regenerate-password", this.accessKey);
    },
  },
});
</script>

<style lang="scss" scoped>
.access-key-box {
  background-color: $body-background-color;
}
</style>
