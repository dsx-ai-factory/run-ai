<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="team-setup-step">
    <sub-step-transition>
      <project-creation
        v-if="subStep === TeamSetupSubstep.CreateTeam"
        :team-data="teamData"
        @update:team-data="handleTeamDataUpdate"
      />
      <team-credentials
        v-if="subStep === TeamSetupSubstep.TeamCredentials && !isEmailChannelAvailable && !isSSOEnabled"
        :team-data="teamData"
        :created-users="createdUsers"
        @update:team-data="handleTeamDataUpdate"
      />
    </sub-step-transition>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { SubStepTransition } from "@/components/onboarding/common/sub-step-transition";
import { ProjectCreation } from "@/components/onboarding/steps/team-setup/substeps/project-creation";
import { TeamCredentials } from "@/components/onboarding/steps/team-setup/substeps/team-credentials";
import type { TeamData, TeamDataKey } from "@/components/onboarding/models/admin-onboarding.model";
import type { UserPostResponse } from "@/swagger-models/identity-manager-client";
import { TeamSetupSubstep } from "@/components/onboarding/models/admin-onboarding.model";
import { useNotificationStore } from "@/stores/notifications.store";
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  name: "team-setup-step",
  components: { ProjectCreation, SubStepTransition, TeamCredentials },
  emits: {
    "update:team-data": (key: TeamDataKey, value: string | boolean | string[]) => !!key && value !== undefined,
  },
  props: {
    subStep: {
      type: Number as PropType<number>,
      required: true,
    },
    teamData: {
      type: Object as PropType<TeamData>,
      required: true,
    },
    createdUsers: {
      type: Array as PropType<UserPostResponse[]>,
      default: () => [],
    },
  },
  setup(_, { emit }) {
    const handleTeamDataUpdate = <K extends TeamDataKey>(key: K, value: TeamData[K]) => {
      emit("update:team-data", key, value);
    };
    const notificationStore = useNotificationStore();
    if (!notificationStore.isChannelsLoaded) {
      notificationStore.loadNotificationChannels();
    }

    const isSSOEnabled = useAuthStore().isSSO;

    return {
      TeamSetupSubstep,
      handleTeamDataUpdate,
      isEmailChannelAvailable: notificationStore.isEmailChannelAvailable,
      isSSOEnabled,
    };
  },
});
</script>

<style scoped lang="scss">
.team-setup-step {
  width: 100%;
  height: 100%;
}
</style>
