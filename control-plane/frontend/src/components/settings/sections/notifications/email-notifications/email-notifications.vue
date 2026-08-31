<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="email-notifications">
    <div class="row items-center q-mb-md">
      <span class="q-mr-md title">Email notifications</span>
    </div>
    <div class="row items-center q-mb-md">
      <span class="q-mr-md">Set the email server through which the email notifications will be sent</span>
      <runai-tooltip
        aid="tool-tip-email-notification"
        tooltip-text="To allow users to receive email notifications and set their own preferences via the user settings, first configure the details of the email server through which the emails will be sent"
        width="400px"
        tooltip-position="right"
      />
    </div>
    <setting-readonly-input resource-type="NotificationSettings">
      <template v-slot:default="{ readonly }">
        <q-btn
          aid="add-email-botifications-btn"
          icon="ph-bold ph-plus"
          label="email server"
          class="btn-icon-sm"
          flat
          color="primary"
          v-if="showAddButton"
          @click="edit"
          :disable="readonly"
        />
      </template>
    </setting-readonly-input>
    <setting-editable-field
      v-if="!editMode && hasSmtpServer"
      resource-type="NotificationSettings"
      delete-tool-tip="Remove email server"
      edit-tool-tip="Edit email server"
      label="Email server"
      delete-sub-title="Email notifications will be disabled."
      @edit="edit"
      @delete="deleteSmtpServer"
    />
    <setting-wrapper-box
      v-if="editMode"
      primary-button-label="save"
      :secondary-button-label="isReadOnly ? 'Close' : 'Cancel'"
      @primary-button-clicked="saveSmtpServer"
      @secondary-button-clicked="cancel"
      has-footer
      :validation-message="validationMessage"
      :show-primary="!isReadOnly"
    >
      <email-server-form
        :email-server-settings="emailServerSettings"
        @update="updateSettings"
        ref="emailServerForm"
        :read-only="isReadOnly"
      />
      <div class="q-mt-lg">
        <verify-recipient
          :model-value="verifyRecipient"
          @update:model-value="updateVerifyRecipient"
          ref="verifyRecipientInput"
          :read-only="isReadOnly"
        />
      </div>
      <div class="q-mt-lg">
        <verify-button
          :verify-status="verifyStatus"
          @verify="verifyEmailServer"
          :validation-message="emailServerValidation"
          :read-only="isReadOnly"
        />
      </div>
    </setting-wrapper-box>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import { EmailErrorMessages, type ISettingEmailServer } from "@/models/setting.model";
import { DEFAULT_EMPTY_EMAIL_SETTINGS } from "@/models/setting.model";
import { ResourceType } from "@/swagger-models/authorization-client";
// components
import { ERunaiVerifyButtonStatus } from "@/components/common/runai-verify-button";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import SettingWrapperBox from "@/components/settings/setting-wrapper-box/setting-wrapper-box.vue";
import SettingEditableField from "@/components/settings/setting-editable-field/setting-editable-field.vue";
import { VerifyButton } from "@/components/settings/sections/notifications/email-notifications/verify-button";
import { VerifyRecipient } from "@/components/settings/sections/notifications/email-notifications/verify-recipient";
import { EmailServerForm } from "./email-server-form";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
// stores
import { useSettingStore } from "@/stores/setting.store";
import { usePermissionStore } from "@/stores/permissions.store";
// utils
import { alertUtil } from "@/utils/alert.util";
import { isEqual } from "@/utils/common.util";
import { settingsUtil } from "@/utils/settings.util";
// services
import { notificationService } from "@/services/control-plane/notifications.service/notification.service";
// composables
import { useEmailServerVerification } from "@/components/settings/sections/notifications/email-notifications/composables";

export default defineComponent({
  name: "email-notifications",
  components: {
    SettingReadonlyInput,
    EmailServerForm,
    VerifyButton,
    VerifyRecipient,
    SettingEditableField,
    SettingWrapperBox,
    RunaiTooltip,
  },
  emits: ["updated", "deleted"],
  props: {
    setting: {
      type: Object as PropType<ISettingEmailServer | undefined>,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      enableDismiss: false as boolean,
      editMode: false as boolean,
      emailServerSettings: {} as ISettingEmailServer,
      verifyRecipient: "" as string,
    };
  },
  setup() {
    const { verifyStatus, emailServerValidation, verifyEmailServer, resetVerificationStatus } =
      useEmailServerVerification();

    return {
      verifyStatus,
      emailServerValidation,
      verifyEmailServerComposable: verifyEmailServer,
      resetVerificationStatus,
    };
  },
  created() {
    if (this.setting) {
      this.emailServerSettings = { ...this.setting };
    } else {
      this.emailServerSettings = DEFAULT_EMPTY_EMAIL_SETTINGS;
    }
  },
  computed: {
    hasSmtpServer(): boolean {
      return this.setting != undefined && this.setting.smtpHost != "";
    },
    showAddButton(): boolean {
      return !this.hasSmtpServer && !this.editMode;
    },
    hasChanges(): boolean {
      return this.setting != undefined && JSON.stringify(this.setting) !== JSON.stringify(this.emailServerSettings);
    },
    validationMessage(): string {
      return this.hasChanges ? "Unsaved changes" : "";
    },
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType.NotificationSettings);
    },
  },
  methods: {
    async cancel() {
      this.emailServerSettings = { ...this.setting } as ISettingEmailServer;
      this.verifyRecipient = "";
      this.resetVerificationStatus();
      this.editMode = false;
    },
    async edit() {
      this.editMode = true;
    },
    async deleteSmtpServer(): Promise<void> {
      try {
        await notificationService.deleteEmail();
        this.emailServerSettings = { ...DEFAULT_EMPTY_EMAIL_SETTINGS };
        this.editMode = false;
        this.resetVerificationStatus();
        this.$emit("deleted");
      } catch {
        this.$q.notify(alertUtil.getError(EmailErrorMessages.DELETE_SERVER_ERROR));
      }
    },
    async saveSmtpServer(): Promise<void> {
      if (!(await this.validateFields())) return;

      try {
        const config = settingsUtil.emailSettingsToSmtpConfig(this.emailServerSettings);
        if (this.hasSmtpServer) {
          await notificationService.updateEmail({ config });
        } else {
          await notificationService.createEmail({ config });
        }
        this.$emit("updated", this.emailServerSettings);
      } catch (error) {
        this.$q.notify(alertUtil.getError(EmailErrorMessages.EMAIL_SAVED_ERROR));
      } finally {
        this.editMode = false;
      }
    },
    async validateFields(): Promise<boolean> {
      const isEmailServerFormValid = await this.validateMandatoryFields();
      return isEmailServerFormValid && this.verifyStatus === ERunaiVerifyButtonStatus.Verified;
    },
    async validateMandatoryFields(): Promise<boolean> {
      const [isFormValid, isRecipientValid] = await Promise.all([
        (this.$refs.emailServerForm as typeof EmailServerForm).validate(),
        (this.$refs.verifyRecipientInput as typeof VerifyRecipient).validate(),
      ]);
      return isFormValid && isRecipientValid;
    },
    updateVerifyRecipient(recipient: string): void {
      this.verifyRecipient = recipient;
      this.resetVerificationStatus();
    },
    updateSettings(settings: ISettingEmailServer): void {
      this.emailServerSettings = { ...settings };
      if (!isEqual(this.emailServerSettings, this.setting)) {
        this.resetVerificationStatus();
      }
    },
    async verifyEmailServer(): Promise<void> {
      await this.verifyEmailServerComposable(
        this.emailServerSettings,
        this.validateMandatoryFields,
        this.verifyRecipient,
      );
    },
  },
});
</script>

<style scoped lang="scss">
.email-notifications {
  .title {
    font-size: 16px;
  }
}
</style>
