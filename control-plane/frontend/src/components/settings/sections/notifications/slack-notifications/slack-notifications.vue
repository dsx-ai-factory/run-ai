<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="slack-notifications">
    <div class="row items-center q-mb-md">
      <span class="q-mr-md title">Slack notifications</span>
    </div>
    <div class="row items-center q-mb-md">
      <span class="q-mr-md">Authorize and connect your Slack workspace to NVIDIA Run:ai</span>
      <runai-tooltip
        aid="tool-tip-slack-notification"
        tooltip-text="To allow users to receive Slack notifications and set their own preferences via the user settings, first connect your Slack workspace."
        width="400px"
        tooltip-position="right"
      />
    </div>
    <setting-readonly-input resource-type="NotificationSettings">
      <template v-slot:default="{ readonly }">
        <template v-if="!hasSlackIntegration">
          <q-input
            v-if="needsAccessToken"
            v-model="accessToken"
            label="Access token"
            placeholder="e.g., xoxe.xoxp-1234567890"
            input-class="placeholder-italic"
            stack-label
            no-error-icon
            :rules="[isValidToken]"
            class="access-token-input q-mb-md"
            aid="slack-access-token-input"
            ref="accessTokenInput"
            :disable="readonly"
          />
          <runai-btn
            class="runai-icon-btn--default"
            icon="ph-bold ph-slack-logo"
            label="Connect Slack Workspace"
            @click="handleConnectClick"
            color="grey-4"
            text-color="b-main"
            :disable="isConnectDisabled || readonly"
            :disabled-tooltip="connectDisabledTooltip"
            aid="add-slack-app-btn"
          />
          <div v-if="needsAccessToken" class="q-mt-md italic">
            For more information on how to get the access token, see
            <a :href="DOCS.SLACK_NOTIFICATIONS.CONNECT_WORKSPACE" target="_blank">Slack Notifications</a>
          </div>
        </template>
        <slack-integration-card
          v-else
          :workspace-name="workspaceName"
          :readonly="readonly"
          @disconnect="disconnectSlackWorkspace"
        />
      </template>
    </setting-readonly-input>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { QInput } from "quasar";

import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiBtn } from "@/components/common/runai-btn";
import { SlackIntegrationCard } from "./slack-integration-card";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import { DOCS } from "@/constants/docs";
import { useAppStore } from "@/stores/app.store";
import { isEmpty } from "@/common/form.validators";

const SLACK_ACCESS_TOKEN_REGEX = /^(xoxe\.)?xoxp-[a-zA-Z0-9-]+$/;

export default defineComponent({
  name: "slack-notifications",
  components: { RunaiTooltip, RunaiBtn, SlackIntegrationCard, SettingReadonlyInput },
  emits: ["connect-slack-workspace", "disconnect-slack-workspace"],
  props: {
    hasSlackIntegration: {
      type: Boolean,
      default: false,
    },
    workspaceName: {
      type: String,
      default: "",
    },
    appCredentialsExist: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      accessToken: "",
    };
  },
  setup() {
    return { DOCS };
  },
  computed: {
    isSingleTenant(): boolean {
      return !this.appStore.isMultiTenant;
    },
    needsAccessToken(): boolean {
      return this.isSingleTenant && !this.appCredentialsExist;
    },
    isConnectDisabled(): boolean {
      return !this.isSingleTenant && !this.appCredentialsExist;
    },
    connectDisabledTooltip(): string | undefined {
      if (this.isConnectDisabled) return "To enable Slack integration, contact your cloud provider";
      return undefined;
    },
  },
  methods: {
    isValidToken(value: string): boolean | string {
      if (isEmpty(value)) return "Enter token";
      return SLACK_ACCESS_TOKEN_REGEX.test(value) || "Enter a valid token";
    },
    async handleConnectClick(): Promise<void> {
      if (!this.needsAccessToken) {
        this.$emit("connect-slack-workspace");
        return;
      }

      const isValid = await (this.$refs.accessTokenInput as typeof QInput).validate();
      if (!isValid) return;
      this.$emit("connect-slack-workspace", this.accessToken);
    },
    disconnectSlackWorkspace(): void {
      this.$emit("disconnect-slack-workspace");
    },
  },
});
</script>

<style scoped lang="scss">
.slack-notifications {
  .title {
    font-size: 16px;
  }
}
</style>
