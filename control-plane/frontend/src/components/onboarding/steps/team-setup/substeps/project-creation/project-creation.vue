<template>
  <onboarding-content-wrapper>
    <template #title>{{ ONBOARDING_TITLES.TEAM_SETUP }}</template>
    <template #subtitle>
      <team-setup-subtitle />
    </template>
    <template #content>
      <div class="team-setup-content">
        <!-- Delegation Option -->
        <delegation-option
          :is-delegated="teamData.letSomeoneElseFill"
          :delegate-email="teamData.delegateEmail"
          @update:is-delegated="handleLetSomeoneElseFillChange"
          @update:delegate-email="handleDelegateEmailChange"
          class="q-mb-lg"
        />

        <!-- Team Setup Content (when not delegated) -->
        <transition name="input-mode-transition" mode="out-in">
          <div
            v-if="!teamData.letSomeoneElseFill"
            key="user-fills-mode"
            class="input-mode-container user-fills-mode q-pl-sm"
          >
            <!-- Team Name -->
            <div class="row q-mb-lg">
              <div class="col-12">
                <team-name-input :team-data="teamData" @update:team-data="handleTeamDataUpdate" />
              </div>
            </div>

            <!-- Team Identification Method (only show if IDP is available) -->
            <div v-if="hasIdpConfigured" class="team-identification-section q-mb-lg">
              <runai-radio-options
                :model-value="teamData.identificationMethod"
                @update:model-value="handleIdentificationMethodChange"
                :options="identificationOptions"
                title="Set how to identify your team"
                class="identification-radio-options"
              />
            </div>

            <!-- Team Member Emails (show when member-emails is selected or no IDP) -->
            <div v-if="shouldShowMemberEmails" class="team-members-section">
              <div class="row">
                <div class="col-12">
                  <team-member-emails-input :team-data="teamData" @update:team-data="handleTeamDataUpdate" />
                </div>
              </div>
            </div>

            <!-- Group Name Input (show when group-name is selected) -->
            <div v-if="shouldShowGroupName" class="group-name-section">
              <div class="row">
                <div class="col-12">
                  <group-name-input :team-data="teamData" @update:team-data="handleTeamDataUpdate" />
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType, computed } from "vue";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { DelegationOption } from "@/components/onboarding/common/delegation-option";
import { TeamSetupSubtitle } from "@/components/onboarding/common/team-setup-subtitle";
import { ONBOARDING_TITLES } from "@/components/onboarding/composable/use-onboarding-state";
import type { TeamData, TeamDataKey } from "@/components/onboarding/models/admin-onboarding.model";
import { TeamIdentificationMethod } from "@/components/onboarding/models/admin-onboarding.model";
import { TeamNameInput } from "@/components/onboarding/common/team-name-input";
import { TeamMemberEmailsInput } from "@/components/onboarding/common/team-member-emails-input";
import { GroupNameInput } from "@/components/onboarding/common/group-name-input";
import RunaiRadioOptions from "@/components/common/runai-radio-options/runai-radio-options.vue";
import { useIdpsStore } from "@/stores/idps.store";
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  name: "project-creation",
  components: {
    OnboardingContentWrapper,
    DelegationOption,
    TeamSetupSubtitle,
    TeamNameInput,
    TeamMemberEmailsInput,
    GroupNameInput,
    RunaiRadioOptions,
  },
  emits: {
    "update:team-data": (key: TeamDataKey, value: string | boolean | string[]) => !!key && value !== undefined,
  },
  props: {
    teamData: {
      type: Object as PropType<TeamData>,
      required: true,
    },
  },
  setup(props, { emit }) {
    const idpStore = useIdpsStore();

    // Check if IDP is configured
    const hasIdpConfigured = computed(() => idpStore.idps.length > 0);

    // Radio button options for identification method
    const identificationOptions: ISelectOption[] = [
      { label: "Member emails", value: TeamIdentificationMethod.MemberEmails },
      { label: "Group name", value: TeamIdentificationMethod.GroupName },
    ];

    // Computed properties for conditional rendering
    const shouldShowMemberEmails = computed(() => {
      return !hasIdpConfigured.value || props.teamData.identificationMethod === TeamIdentificationMethod.MemberEmails;
    });

    const shouldShowGroupName = computed(() => {
      return hasIdpConfigured.value && props.teamData.identificationMethod === TeamIdentificationMethod.GroupName;
    });

    // Event handlers
    const handleLetSomeoneElseFillChange = (value: boolean) => {
      emit("update:team-data", "letSomeoneElseFill", value);
      // Clear other fields when switching modes
      if (value) {
        emit("update:team-data", "teamName", "");
        emit("update:team-data", "memberEmails", []);
        emit("update:team-data", "groupName", "");
      } else {
        emit("update:team-data", "delegateEmail", "");
      }
    };

    const handleDelegateEmailChange = (value: string) => {
      emit("update:team-data", "delegateEmail", value);
    };

    const handleIdentificationMethodChange = (value: string) => {
      emit("update:team-data", "identificationMethod", value as TeamIdentificationMethod);
      // Clear the fields from the method we're switching away from
      if (value === TeamIdentificationMethod.MemberEmails) {
        emit("update:team-data", "groupName", "");
      } else {
        emit("update:team-data", "memberEmails", []);
      }
    };

    const handleTeamDataUpdate = <K extends TeamDataKey>(key: K, value: TeamData[K]) => {
      emit("update:team-data", key, value);
    };

    return {
      ONBOARDING_TITLES,
      hasIdpConfigured,
      identificationOptions,
      shouldShowMemberEmails,
      shouldShowGroupName,
      handleLetSomeoneElseFillChange,
      handleDelegateEmailChange,
      handleIdentificationMethodChange,
      handleTeamDataUpdate,
    };
  },
});
</script>

<style scoped lang="scss">
.team-setup-content {
  max-width: 600px;
}

// Delegation styles are now in the delegation-option component

// Transition animations for input mode switching
.input-mode-transition-enter-active,
.input-mode-transition-leave-active {
  transition: all 0.45s cubic-bezier(0.4, 0, 0.2, 1);
}

.input-mode-transition-enter-from {
  opacity: 0;
  transform: translateX(-20px) scale(0.95);
}

.input-mode-transition-leave-to {
  opacity: 0;
  transform: translateX(20px) scale(0.95);
}

.input-mode-transition-enter-to,
.input-mode-transition-leave-from {
  opacity: 1;
  transform: translateX(0) scale(1);
}

// Add smooth container transition
.input-mode-container {
  transition: min-height 0.4s ease-in-out;
  will-change: transform, opacity;
}

// Delegation transitions are now in the delegation-option component
</style>
