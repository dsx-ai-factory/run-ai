<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <onboarding-content-wrapper>
    <template #title>Invite team members to start using NVIDIA Run:ai</template>
    <template #subtitle> Download the file with credentials for each team member and use it to send invites. </template>
    <template #content>
      <div class="team-credentials-content">
        <div class="download-section">
          <q-btn
            class="download-button"
            color="grey-4"
            text-color="grey-8"
            @click="downloadCredentials"
            icon="ph-bold ph-download-simple"
            label="DOWNLOAD TEAM'S CREDENTIALS"
            no-caps
          />
        </div>
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import type { TeamData } from "@/components/onboarding/models/admin-onboarding.model";
import type { UserPostResponse } from "@/swagger-models/identity-manager-client";
import { useDownloadCredentials } from "@/composables/use-download-credentials.composable";

export default defineComponent({
  name: "team-credentials",
  components: {
    OnboardingContentWrapper,
  },
  props: {
    teamData: {
      type: Object as PropType<TeamData>,
      required: true,
    },
    createdUsers: {
      type: Array as PropType<UserPostResponse[]>,
      default: () => [],
    },
  },
  setup(props) {
    const { downloadCredentials: downloadCredentialsComposable } = useDownloadCredentials();

    const downloadCredentials = async () => {
      await downloadCredentialsComposable({
        createdUsers: props.createdUsers,
        teamName: props.teamData.teamName,
      });
    };

    return {
      downloadCredentials,
    };
  },
});
</script>

<style scoped lang="scss">
.team-credentials-content {
  max-width: 600px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  text-align: center;
  padding: 20px;
  justify-content: center;
}

.download-section {
  margin-top: 0px;
}

.download-button {
  min-width: 280px;
  height: 42px;
  font-weight: 600;
  font-size: 14px;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);

  :deep(.q-btn__content) {
    gap: 12px;
  }

  :deep(.q-icon) {
    font-size: 20px;
  }
}
</style>
