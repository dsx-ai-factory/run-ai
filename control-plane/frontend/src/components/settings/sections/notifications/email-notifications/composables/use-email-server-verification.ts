/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ref, computed, type Ref } from "vue";
import { ERunaiVerifyButtonStatus } from "@/components/common/runai-verify-button";
import { notificationService } from "@/services/control-plane/notifications.service/notification.service";
import type { ISettingEmailServer } from "@/models/setting.model";
import type { EmailChannelConfig } from "@/swagger-models/notifications-proxy-client";
import { settingsUtil } from "@/utils/settings.util/settings.util";

export function useEmailServerVerification() {
  const verifyStatus: Ref<ERunaiVerifyButtonStatus> = ref(ERunaiVerifyButtonStatus.None);

  const emailServerValidation = computed(() => {
    switch (verifyStatus.value) {
      case ERunaiVerifyButtonStatus.None:
        return "To continue, complete the verification by clicking above";
      case ERunaiVerifyButtonStatus.Verified:
        return "Verification successful";
      case ERunaiVerifyButtonStatus.Failed:
        return "Verification failed. Check server details.";
      default:
        return "";
    }
  });

  const verifyEmailServer = async (
    emailServerSettings: ISettingEmailServer,
    validateMandatoryFields: () => Promise<boolean>,
    recipient?: string,
  ): Promise<void> => {
    const isMandatoryFieldsFilled = await validateMandatoryFields();
    if (!isMandatoryFieldsFilled) {
      verifyStatus.value = ERunaiVerifyButtonStatus.None;
      return;
    }

    verifyStatus.value = ERunaiVerifyButtonStatus.Verifying;
    try {
      const config: EmailChannelConfig = settingsUtil.emailSettingsToSmtpConfig(emailServerSettings);
      if (recipient) {
        config.recipients = [recipient];
      }
      const isValid = await notificationService.validateEmail(config);

      if (isValid) {
        verifyStatus.value = ERunaiVerifyButtonStatus.Verified;
      } else {
        verifyStatus.value = ERunaiVerifyButtonStatus.Failed;
      }
    } catch (error) {
      verifyStatus.value = ERunaiVerifyButtonStatus.Failed;
    }
  };

  const resetVerificationStatus = () => {
    verifyStatus.value = ERunaiVerifyButtonStatus.None;
  };

  return {
    verifyStatus,
    emailServerValidation,
    verifyEmailServer,
    resetVerificationStatus,
  };
}
