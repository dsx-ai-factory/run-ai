<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="onboarding-stepper-container">
    <transition name="step-transition" mode="out-in">
      <!-- LANDING STEP -->
      <landing-step v-if="step === OnboardingStep.Landing" key="landing-step" />
      <!-- CLUSTER SETUP -->
      <cluster-setup-step
        v-else-if="step === OnboardingStep.ClusterSetup"
        key="cluster-setup-step"
        :sub-step="subStep"
        :cluster-data="clusterData"
        @update:cluster-data="handleClusterDataUpdate"
        @close="$emit('close')"
      />
      <!-- SSO SETUP -->
      <sso-step
        v-else-if="step === OnboardingStep.SsoSetup"
        key="sso-step"
        :sub-step="subStep"
        :sso-data="ssoData"
        @update:sso-data="handleSsoDataUpdate"
      />
      <!-- EMAIL SERVER SETUP -->
      <email-server-setup-step
        v-else-if="step === OnboardingStep.EmailServerSetup"
        key="email-server-setup-step"
        :email-server-data="emailServerData"
        @update:email-server-data="handleEmailServerDataUpdate"
      />
      <!--  TEAM SETUP -->
      <team-setup-step
        v-else-if="step === OnboardingStep.TeamSetup"
        :sub-step="subStep"
        key="team-setup-step"
        :team-data="teamData"
        :created-users="provisioningData?.createdUsers || []"
        @update:team-data="handleTeamDataUpdate"
      />

      <!-- COMPLETION STATUS -->
      <onboarding-completion-status
        v-else-if="step === OnboardingStep.Completion && provisioningData.completionState"
        :completion-state="provisioningData.completionState"
        :failed-subject-ids="provisioningData.failedSubjectIds"
        :created-users="provisioningData?.createdUsers || []"
        :team-name="teamData.teamName"
        @close="$emit('close', $event)"
        @retry="$emit('retry')"
      />
    </transition>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { ClusterSetupStep } from "@/components/onboarding/steps/cluster-setup";
import { LandingStep } from "@/components/onboarding/steps/landing";
import { SsoStep } from "@/components/onboarding/steps/sso";
import { EmailServerSetupStep } from "@/components/onboarding/steps/email-server-setup";
import { TeamSetupStep } from "@/components/onboarding/steps/team-setup";
import type {
  ClusterData,
  ClusterDataKey,
  SsoData,
  SsoDataKey,
  TeamData,
  TeamDataKey,
  ProvisioningData,
  EmailServerData,
  EmailServerDataKey,
} from "@/components/onboarding/models/admin-onboarding.model";
import { OnboardingStep } from "@/components/onboarding/models/admin-onboarding.model";
import { ssoDataUpdateEmitValidator } from "@/components/onboarding/models/admin-onboarding.model";
import type { ISettingEmailServer } from "@/models/setting.model";
import { OnboardingCompletionStatus } from "@/components/onboarding/common/onboarding-completion-status";

export default defineComponent({
  name: "onboarding-stepper",
  components: {
    OnboardingCompletionStatus,
    ClusterSetupStep,
    LandingStep,
    SsoStep,
    EmailServerSetupStep,
    TeamSetupStep,
  },
  emits: {
    "update:cluster-data": (key: ClusterDataKey, value: string | boolean) => !!key && value !== undefined,
    "update:sso-data": ssoDataUpdateEmitValidator,
    "update:team-data": (key: TeamDataKey, value: string | boolean | string[]) => !!key && value !== undefined,
    "update:email-server-data": (key: EmailServerDataKey, value: ISettingEmailServer) => !!key && value !== null,
    close: (routeName?: string) => typeof routeName === "string" || routeName === undefined,
    retry: () => true,
  },
  props: {
    step: {
      type: Number as PropType<number>,
      required: true,
    },
    subStep: {
      type: Number as PropType<number>,
      default: 1,
    },
    clusterData: {
      type: Object as PropType<ClusterData>,
      required: true,
    },
    ssoData: {
      type: Object as PropType<SsoData>,
      required: true,
    },
    teamData: {
      type: Object as PropType<TeamData>,
      required: true,
    },
    provisioningData: {
      type: Object as PropType<ProvisioningData>,
      required: true,
    },
    emailServerData: {
      type: Object as PropType<EmailServerData>,
      required: true,
    },
  },
  setup(_, { emit }) {
    const handleClusterDataUpdate = <K extends ClusterDataKey>(key: K, value: ClusterData[K]) => {
      emit("update:cluster-data", key, value);
    };

    const handleSsoDataUpdate = <K extends SsoDataKey>(key: K, value: SsoData[K]) => {
      emit("update:sso-data", key, value);
    };

    const handleTeamDataUpdate = <K extends TeamDataKey>(key: K, value: TeamData[K]) => {
      emit("update:team-data", key, value);
    };

    const handleEmailServerDataUpdate = <K extends EmailServerDataKey>(key: K, value: EmailServerData[K]) => {
      emit("update:email-server-data", key, value);
    };

    return {
      OnboardingStep,
      handleClusterDataUpdate,
      handleSsoDataUpdate,
      handleTeamDataUpdate,
      handleEmailServerDataUpdate,
    };
  },
});
</script>

<style lang="scss" scoped>
.onboarding-stepper-container {
  width: 100%;
  height: 100%;
}

// Step transition animations
.step-transition-enter-active,
.step-transition-leave-active {
  transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

.step-transition-enter-from {
  opacity: 0;
  transform: translateX(30px);
}

.step-transition-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}

.step-transition-enter-to,
.step-transition-leave-from {
  opacity: 1;
  transform: translateX(0);
}
</style>
