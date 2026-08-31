<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <email-server-setup-error v-if="emailServerData.saveError" />
  <onboarding-content-wrapper v-else>
    <template #title>Set an email server</template>
    <template #subtitle>
      Configure an SMTP server to send
      <a :href="DOCS.INSTALLATION.USER_INVITATIONS" target="_blank">user invitations</a>
      and enable
      <a :href="DOCS.INSTALLATION.EMAIL_NOTIFICATIONS" target="_blank">email notifications</a>.
    </template>
    <template #content>
      <email-server-form
        :email-server-settings="emailServerSettings"
        :read-only="emailServerData.isConfigured"
        :display-read-only-tooltip="false"
        @update="updateSettings"
        ref="emailServerForm"
      />
      <div class="q-mt-lg">
        <verify-recipient
          :model-value="verifyRecipient"
          @update:model-value="updateVerifyRecipient"
          ref="verifyRecipientInput"
          :read-only="emailServerData.isConfigured"
          :display-read-only-tooltip="false"
        />
      </div>
      <div class="q-mt-lg">
        <verify-button
          :verify-status="verifyStatus"
          :read-only="emailServerData.isConfigured"
          :display-read-only-tooltip="false"
          @verify="verifyEmailServer"
          :validation-message="emailServerData.isConfigured ? '' : emailServerValidation"
        />
      </div>
      <!--  HIDDEN FIELD FOR VALIDATION -->
      <q-input
        v-model="hiddenValidationField"
        :rules="[validateEmailServerSetup]"
        class="hidden-field"
        style="display: none"
      />
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent, ref, computed, type PropType } from "vue";
import { DOCS } from "@/constants/docs";
import { EmailServerForm } from "@/components/settings/sections/notifications/email-notifications/email-server-form";
import { VerifyButton } from "@/components/settings/sections/notifications/email-notifications/verify-button";
import { VerifyRecipient } from "@/components/settings/sections/notifications/email-notifications/verify-recipient";
import { OnboardingContentWrapper } from "@/components/onboarding/common/onboarding-content-wrapper";
import { ERunaiVerifyButtonStatus } from "@/components/common/runai-verify-button";
import type { ISettingEmailServer } from "@/models/setting.model";
import type { EmailServerData } from "@/components/onboarding/models/admin-onboarding.model";
import { EmailServerSetupError } from "@/components/onboarding/steps/email-server-setup/errors-state/email-server-setup-error";
import { useEmailServerVerification } from "@/components/settings/sections/notifications/email-notifications/composables";
import { isEqual } from "@/utils/common.util";

export default defineComponent({
  name: "email-server-setup-step",
  components: {
    EmailServerForm,
    VerifyButton,
    VerifyRecipient,
    OnboardingContentWrapper,
    EmailServerSetupError,
  },
  emits: {
    "update:email-server-data": (key: "settings", value: ISettingEmailServer) => !!key && !!value,
  },
  props: {
    emailServerData: {
      type: Object as PropType<EmailServerData>,
      required: true,
    },
  },
  setup(props, { emit }) {
    const {
      verifyStatus,
      emailServerValidation,
      verifyEmailServer: verifyEmailServerComposable,
      resetVerificationStatus,
    } = useEmailServerVerification();
    const emailServerForm = ref<typeof EmailServerForm | null>(null);
    const verifyRecipientInput = ref<typeof VerifyRecipient | null>(null);
    const verifyRecipient = ref("");
    const hiddenValidationField = ref("");

    const updateSettings = (settings: ISettingEmailServer): void => {
      const previousSettings = { ...props.emailServerData.settings };
      emit("update:email-server-data", "settings", settings);
      if (!isEqual(settings, previousSettings)) {
        resetVerificationStatus();
      }
    };

    const validateMandatoryFields = async (): Promise<boolean> => {
      if (!emailServerForm.value || !verifyRecipientInput.value) return false;
      const [isFormValid, isRecipientValid] = await Promise.all([
        emailServerForm.value.validate(),
        verifyRecipientInput.value.validate(),
      ]);
      return isFormValid && isRecipientValid;
    };

    const updateVerifyRecipient = (recipient: string): void => {
      verifyRecipient.value = recipient;
      resetVerificationStatus();
    };

    const validateEmailServerSetup = async (): Promise<boolean | string> => {
      if (props.emailServerData.isConfigured) {
        return true;
      }
      const isFormValid = await validateMandatoryFields();
      if (!isFormValid) {
        return "Please fill in all required email server fields";
      }
      if (verifyStatus.value !== ERunaiVerifyButtonStatus.Verified) {
        return "To continue, complete the verification by clicking above";
      }
      return true;
    };

    const verifyEmailServer = async (): Promise<void> => {
      await verifyEmailServerComposable(props.emailServerData.settings, validateMandatoryFields, verifyRecipient.value);
    };

    return {
      emailServerSettings: computed(() => props.emailServerData.settings),
      verifyStatus,
      emailServerValidation,
      emailServerForm,
      verifyRecipient,
      verifyRecipientInput,
      hiddenValidationField,
      updateSettings,
      updateVerifyRecipient,
      verifyEmailServer,
      validateEmailServerSetup,
      DOCS,
    };
  },
});
</script>
