<template>
  <q-form ref="formRef" class="delegated-form-wrapper">
    <!-- LOGO -->
    <div :class="{ 'logo-wrapper': !completionState }">
      <runai-distribution-logo v-bind="DISTRIBUTION_LOGO_PRESETS.delegatedOnboarding" />
    </div>
    <div class="delegated-team-setup" :style="{ paddingTop: completionState ? '0' : '20px' }">
      <!-- TEAM SETUP FORM -->
      <template v-if="!completionState">
        <div class="row items-center justify-center">
          <div class="setup-title">{{ ONBOARDING_TITLES.TEAM_SETUP }}</div>
          <div class="setup-subtitle">
            <team-setup-subtitle />
          </div>
          <div class="inputs-wrapper q-mt-xl">
            <!-- Team Name -->
            <team-name-input :team-data="teamData" @update:team-data="handleTeamDataUpdate" class="full-width" />

            <!-- Team Identification Method (only show if IDP is available) -->
            <div v-if="hasIdpConfigured" class="text-left q-mb-lg full-width">
              <runai-radio-options
                :model-value="teamData.identificationMethod"
                @update:model-value="handleIdentificationMethodChange"
                :options="identificationOptions"
                title="Set how to identify your team"
                class="identification-radio-options"
              />
            </div>

            <!-- Team Member Emails (show when member-emails is selected or no IDP) -->
            <div v-if="shouldShowMemberEmails" class="full-width">
              <team-member-emails-input
                :team-data="teamData"
                @update:team-data="handleTeamDataUpdate"
                class="full-width"
              />
            </div>

            <!-- Group Name Input (show when group-name is selected) -->
            <div v-if="shouldShowGroupName" class="group-name-section full-width">
              <group-name-input :team-data="teamData" @update:team-data="handleTeamDataUpdate" />
            </div>
          </div>
        </div>
      </template>
      <!-- COMPLETION STATUS -->
      <onboarding-completion-status
        v-else
        center-align
        :completion-state="completionState"
        :failed-subject-ids="failedSubjectIds"
        content-padding-top="0px"
        @close="$emit('close', $event)"
        @retry="retry"
        class="q-pa-xl"
      />
      <!-- SUBMIT BUTTON -->
      <q-btn
        color="primary"
        :label="submitBtnLabel"
        @click="submit"
        class="complete-btn q-mb-lg"
        aid="complete-setup-btn"
        :loading="isLoading"
        :disable="isLoading"
      />
    </div>
    <div class="help-section q-mb-md text-center" :class="{ 'support-unavailable': !DOCS.SUPPORT.URL }">
      <template v-if="DOCS.SUPPORT.URL">
        <span class="text-grey-6">Having trouble? </span>
        <a
          aid="delegated-team-get-help-link"
          :href="DOCS.SUPPORT.URL"
          target="_blank"
          rel="noopener noreferrer"
          class="help-link q-pl-xs"
        >
          Get help
        </a>
      </template>
    </div>
  </q-form>
</template>

<script lang="ts">
import { computed, defineComponent, type PropType, reactive, ref, type Ref } from "vue";
import DOCS from "@/constants/docs";
import { ONBOARDING_TITLES } from "@/components/onboarding/composable/use-onboarding-state";
import {
  IOnboardingCompletionState,
  type TeamData,
  type TeamDataKey,
  TeamIdentificationMethod,
} from "@/components/onboarding/models/admin-onboarding.model";
import { DISTRIBUTION_LOGO_PRESETS, RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";
import { TeamSetupSubtitle } from "@/components/onboarding/common/team-setup-subtitle";
import { TeamNameInput } from "@/components/onboarding/common/team-name-input";
import { TeamMemberEmailsInput } from "@/components/onboarding/common/team-member-emails-input";
import { GroupNameInput } from "@/components/onboarding/common/group-name-input";
import RunaiRadioOptions from "@/components/common/runai-radio-options/runai-radio-options.vue";
import type { QForm } from "quasar";
import { teamProvisioningService } from "@/services/control-plane/team-provisioning.service/team-provisioning.service";
import { OnboardingCompletionStatus } from "@/components/onboarding/common/onboarding-completion-status";
import { useIdpsStore } from "@/stores/idps.store";
import type { ISelectOption } from "@/models/global.model";
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  name: "delegated-team-setup",
  components: {
    OnboardingCompletionStatus,
    RunaiDistributionLogo,
    TeamSetupSubtitle,
    TeamNameInput,
    TeamMemberEmailsInput,
    GroupNameInput,
    RunaiRadioOptions,
  },
  emits: {
    close: (routeName?: string) => typeof routeName === "string" || routeName === undefined,
  },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
  },
  setup(props, { emit }) {
    const formRef: Ref<QForm | null> = ref(null);
    const isLoading = ref(false);
    const completionState = ref<IOnboardingCompletionState | null>(null);
    const failedSubjectIds = ref<string[] | undefined>(undefined);
    const idpStore = useIdpsStore();

    const teamData = reactive<TeamData>({
      letSomeoneElseFill: false,
      delegateEmail: "",
      teamName: "",
      memberEmails: [],
      identificationMethod: TeamIdentificationMethod.MemberEmails,
      groupName: "",
    });

    // Check if IDP is configured
    const hasIdpConfigured = computed(() => idpStore.idps.length > 0);

    // Radio button options for identification method
    const identificationOptions: ISelectOption[] = [
      { label: "Member emails", value: TeamIdentificationMethod.MemberEmails },
      { label: "Group name", value: TeamIdentificationMethod.GroupName },
    ];

    // Computed properties for conditional rendering
    const shouldShowMemberEmails = computed(() => {
      return !hasIdpConfigured.value || teamData.identificationMethod === TeamIdentificationMethod.MemberEmails;
    });

    const shouldShowGroupName = computed(() => {
      return hasIdpConfigured.value && teamData.identificationMethod === TeamIdentificationMethod.GroupName;
    });

    const validateForm = async (): Promise<boolean> => {
      if (!formRef.value) return false;
      return await formRef.value.validate();
    };

    const submitBtnLabel = computed(() => {
      return completionState.value ? "Done" : "Continue";
    });

    // Event handlers
    const handleTeamDataUpdate = <K extends TeamDataKey>(key: K, value: TeamData[K]) => {
      teamData[key] = value;
    };

    const handleIdentificationMethodChange = (value: string) => {
      teamData.identificationMethod = value as TeamIdentificationMethod;
      // Clear the fields from the method we're switching away from
      if (value === TeamIdentificationMethod.MemberEmails) {
        teamData.groupName = "";
      } else {
        teamData.memberEmails = [];
      }
    };

    const retry = () => {
      completionState.value = null;
      failedSubjectIds.value = undefined;
    };

    const loadUserOrgUnits = async () => {
      try {
        await useAuthStore().loadUserOrgUnits();
      } catch {
        console.error("Failed to load org units");
      }
    };

    const submit = async () => {
      if (completionState.value) {
        emit("close");
        return;
      }
      try {
        if (await validateForm()) {
          isLoading.value = true;
          const response = await teamProvisioningService.provisionTeam(
            props.clusterId,
            teamData,
            hasIdpConfigured.value,
          );
          completionState.value = response.completionState;
          failedSubjectIds.value = response.failedSubjectIds || undefined;

          // After project creation, refresh org units so permitted scopes include the new project
          await loadUserOrgUnits();
        }
      } catch (error) {
        completionState.value = IOnboardingCompletionState.ProjectCreationFailed;
        failedSubjectIds.value = undefined;
      } finally {
        isLoading.value = false;
      }
    };
    return {
      //State
      teamData,
      isLoading,
      formRef,
      completionState,
      failedSubjectIds,
      hasIdpConfigured,
      identificationOptions,
      shouldShowMemberEmails,
      shouldShowGroupName,
      //computed
      submitBtnLabel,
      //Handlers
      handleTeamDataUpdate,
      handleIdentificationMethodChange,
      submit,
      retry,
      DOCS,
      ONBOARDING_TITLES,
      DISTRIBUTION_LOGO_PRESETS,
    };
  },
});
</script>

<style scoped lang="scss">
.delegated-form-wrapper {
  width: 90%;
  height: 100%;
  padding-top: 80px;
  padding-bottom: 10px;
  margin: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;

  .help-section.support-unavailable::before {
    content: "\00a0";
    visibility: hidden;
  }

  .delegated-team-setup {
    width: 100%;
    height: 100%;
    text-align: center;

    display: flex;
    flex-direction: column;
    align-items: center;

    justify-content: space-around;
    .logo-wrapper {
      margin-bottom: 80px;
    }
    .setup-title {
      font-size: 24px;
      font-weight: 500;
      color: $black-87;
      line-height: 20px;
      margin-bottom: 20px;
    }
    .setup-subtitle {
      font-weight: 400;
      font-size: 16px;
      animation: slideInSubtitle 0.3s ease-out 0.15s both;
      width: 100%;
      line-height: 20px;
    }
    .inputs-wrapper {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: space-around;
      width: 100%;
      max-width: 500px;
      min-height: 300px;
    }
    .help-text {
      font-size: 12px;

      .help-link {
        text-decoration: none;
        &:hover {
          text-decoration: underline;
        }
      }
    }
  }
}
</style>
