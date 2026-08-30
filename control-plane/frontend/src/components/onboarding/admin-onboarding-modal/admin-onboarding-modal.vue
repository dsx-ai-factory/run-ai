<template>
  <q-dialog model-value persistent max-width="1250" class="admin-onboarding-modal" @keydown="handleKeydown">
    <div class="onboarding-container">
      <!-- CLOSE BUTTON -->
      <q-btn
        v-if="showCloseButton"
        flat
        round
        icon="ph-bold ph-x"
        class="close-button runai-icon-btn--large"
        size="md"
        @click="handleClose()"
        aria-label="Close onboarding"
      />

      <template v-if="!showDelegatedTeamSetup">
        <!-- SIDE BAR -->
        <onboarding-sidebar
          :current-step="currentStep"
          :current-sub-step="currentSubStep"
          :completed-steps="completedSteps"
          :visited-sub-steps="visitedSubSteps"
          :steps-to-show="stepsToShow"
          :hidden-substeps="hiddenSubsteps"
          :is-navigation-disabled="isBackwardNavigationDisabled()"
          @navigate-to-step="handleNavigateToStep"
        />

        <!-- Regular onboarding stepper -->
        <div class="content-wrapper">
          <onboarding-content>
            <q-form ref="formRef" @submit.prevent>
              <!-- STEPPER CONTENT -->
              <onboarding-stepper
                :step="currentStep"
                :sub-step="currentSubStep"
                :cluster-data="clusterData"
                :sso-data="ssoData"
                :team-data="teamData"
                :provisioning-data="provisioningData"
                :email-server-data="emailServerData"
                @update:cluster-data="handleClusterDataUpdate"
                @update:sso-data="handleSsoDataUpdate"
                @update:team-data="handleTeamDataUpdate"
                @update:email-server-data="handleEmailServerDataUpdate"
                @close="handleClose"
                @retry="handleRetry"
              />
            </q-form>
          </onboarding-content>

          <!-- NAVIGATION -->
          <onboarding-navigation
            v-if="showNavigation"
            :continue-button-label="continueButtonLabel"
            :is-loading="isLoading"
            :show-skip-button="showSkipButton"
            @next="handleNext"
            @skip="handleSkip"
          />
        </div>
      </template>

      <!-- Delegated team setup for users who skip cluster setup -->
      <delegated-team-setup v-else-if="showDelegatedTeamSetup" :cluster-id="initialClusterId" @close="handleClose" />
    </div>
  </q-dialog>
</template>
<script lang="ts">
import { defineComponent, type PropType, ref, type Ref, computed } from "vue";
import OnboardingStepper from "@/components/onboarding/stepper";
import OnboardingSidebar from "@/components/onboarding/common/onboarding-sidebar";
import OnboardingNavigation from "@/components/onboarding/common/onboarding-navigation";
import OnboardingContent from "@/components/onboarding/common/onboarding-content";
import DelegatedTeamSetup from "@/components/onboarding/steps/team-setup/delegated-team-setup/delegated-team-setup.vue";
import { useOnboardingState } from "@/components/onboarding/composable";
import { getClusterSetupSubStepIndex } from "@/components/onboarding/composable/use-onboarding-state";
import type { QForm } from "quasar";
import { useRouter } from "vue-router";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { formUtil } from "@/utils/form.util/form.util";
import { filterService } from "@/services/filter.service/filter.service";
import { ETableFilters } from "@/models/table.model";
import { EColumnFilterType, EFilterOperator, type IFreeTextFilterModel } from "@/models/filter.model";
import { ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { EProjectColumnLabel } from "@/table-models/project.table-model";
import { useClusterStore } from "@/stores/cluster.store";
import { ErrorAlert } from "@/utils/error-alert.util";
import { storageUtil } from "@/utils/storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";
import {
  InstallationErrorType,
  IOnboardingCompletionState,
  OnboardingStep,
  SsoErrorType,
  TeamSetupSubstep,
} from "@/components/onboarding/models/admin-onboarding.model";
import { useAuthStore } from "@/stores/auth.store";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";

export default defineComponent({
  name: "admin-onboarding-modal",
  components: {
    OnboardingContent,
    OnboardingNavigation,
    OnboardingSidebar,
    OnboardingStepper,
    DelegatedTeamSetup,
  },
  emits: {
    close: () => true,
  },
  props: {
    persistent: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    skipClusterSetup: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    clusterId: {
      //Cluster ID for delegated setup
      type: String as PropType<string>,
      required: false,
    },
  },
  setup(props, { emit }) {
    const router = useRouter();
    const onboardingState = useOnboardingState();
    const clusterStore = useClusterStore();
    const authStore = useAuthStore();
    const formRef: Ref<QForm | null> = ref(null);

    // Freeze delegated view decision at open to avoid reactive flipping mid-flow
    const wasSkipClusterSetupAtOpen = !!props.skipClusterSetup;
    const initialClusterId = props.clusterId || "";

    // Common conditions for both delegated flows
    const shouldShowDelegatedFlow =
      wasSkipClusterSetupAtOpen && !!initialClusterId && !storageUtil.get<boolean>(STORAGE_KEYS.ONBOARDING_SKIPPED_SSO);

    const showDelegatedTeamSetup = shouldShowDelegatedFlow && authStore.isSSO;
    const displayDelegatedStepper = shouldShowDelegatedFlow && !authStore.isSSO;

    const showSkipButton = computed(() => {
      if (
        onboardingState.isSsoSetupStep.value &&
        onboardingState.ssoData.value.ssoErrorType === SsoErrorType.IdpCreation
      ) {
        return false;
      }
      if (
        onboardingState.isEmailServerSetupStep.value &&
        (onboardingState.emailServerData.value.saveError || onboardingState.emailServerData.value.isConfigured)
      ) {
        return false;
      }
      if (onboardingState.isTeamSetupStep.value && onboardingState.currentSubStep.value === 2) {
        return false;
      }
      const canSkipSso = onboardingState.isSsoSetupStep.value && !onboardingState.ssoData.value.hasIdpCreated;
      const canSkipTeam = onboardingState.isTeamSetupStep.value;
      const canSkipEmailServer = onboardingState.isEmailServerSetupStep.value;

      return !showDelegatedTeamSetup && (canSkipSso || canSkipTeam || canSkipEmailServer);
    });

    // Control navigation visibility - hide when cluster creation fails in installation substep
    const showNavigation = computed(() => {
      if (onboardingState.isClusterSetupStep.value) {
        const isInstallationSubstep = onboardingState.currentSubStep.value === getClusterSetupSubStepIndex();

        if (
          isInstallationSubstep &&
          onboardingState.clusterData.value.installationErrorType === InstallationErrorType.ClusterCreation
        ) {
          return false;
        }
      }
      return true;
    });

    const showCloseButton = computed(() => {
      if (props.skipClusterSetup) return true;
      if (onboardingState.isLandingStep.value) return false;
      if (onboardingState.isClusterSetupStep.value) {
        return onboardingState.clusterCreationCompleted.value;
      }
      return true;
    });

    const defaultRedirectRoute = computed(() => {
      if (
        onboardingState.provisioningData.value.completionState === IOnboardingCompletionState.AdminUserCreationFailed
      ) {
        return CLUSTER_ROUTE_NAMES.CLUSTER_INDEX;
      }

      if (showDelegatedTeamSetup) {
        return PROJECT_ROUTE_NAMES.PROJECT_INDEX;
      }

      if (onboardingState.currentStep.value === OnboardingStep.Completion) {
        return PROJECT_ROUTE_NAMES.PROJECT_INDEX;
      }

      // If user is in team setup but hasn't started provisioning and SSO was completed, land on settings
      const hasCompletedSso =
        onboardingState.completedSteps.value.includes(OnboardingStep.SsoSetup) ||
        onboardingState.ssoData.value.hasIdpCreated;
      const teamProvisioningStartedOrFinished = Boolean(onboardingState.provisioningData.value.completionState);

      if (onboardingState.isTeamSetupStep.value && hasCompletedSso && !teamProvisioningStartedOrFinished) {
        return SETTINGS_ROUTE_NAMES.SETTINGS_INDEX;
      }

      // Default per current step
      switch (onboardingState.currentStep.value) {
        case OnboardingStep.ClusterSetup:
          return CLUSTER_ROUTE_NAMES.CLUSTER_INDEX;
        case OnboardingStep.SsoSetup:
          return hasCompletedSso ? SETTINGS_ROUTE_NAMES.SETTINGS_INDEX : CLUSTER_ROUTE_NAMES.CLUSTER_INDEX;
        case OnboardingStep.TeamSetup:
          return teamProvisioningStartedOrFinished
            ? PROJECT_ROUTE_NAMES.PROJECT_INDEX
            : hasCompletedSso
              ? SETTINGS_ROUTE_NAMES.SETTINGS_INDEX
              : CLUSTER_ROUTE_NAMES.CLUSTER_INDEX;
        default:
          return CLUSTER_ROUTE_NAMES.CLUSTER_INDEX;
      }
    });

    const handleClose = async (redirectTo?: string) => {
      const finalRedirectRoute = redirectTo || defaultRedirectRoute.value;

      if (onboardingState.isSsoSetupStep.value && !onboardingState.ssoData.value.hasIdpCreated) {
        storageUtil.save(STORAGE_KEYS.ONBOARDING_SKIPPED_SSO, true);
      }

      if (onboardingState.isEmailServerSetupStep.value) {
        storageUtil.save(STORAGE_KEYS.ONBOARDING_SKIPPED_EMAIL_SERVER_SETUP, true);
      }

      if (onboardingState.isTeamSetupStep.value) {
        storageUtil.save(STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP, true);
      }

      await setFiltersBeforeRedirect(finalRedirectRoute);

      await router.push({ name: finalRedirectRoute });
      emit("close");
    };

    const setFiltersBeforeRedirect = async (redirectTo: string) => {
      const clusterId = onboardingState.clusterData.value.clusterId;
      const teamName = onboardingState.teamData.value.teamName;

      if (redirectTo === PROJECT_ROUTE_NAMES.PROJECT_INDEX) {
        setProjectFilters(teamName);
      }
      if (clusterId) {
        await setSelectedClusterById(clusterId);
      }
    };

    const setProjectFilters = (projectName: string) => {
      const projectFilterBy = filterService.loadFilters(window.location, ETableFilters.PROJECT, {}, true);

      // Add project name filter if available
      if (projectName) {
        const projectNameFilter: IFreeTextFilterModel = {
          term: projectName,
          name: ProjectFilterSortFields.Name,
          value: EFilterOperator.Equals,
          type: EColumnFilterType.FreeText,
          label: EProjectColumnLabel.ProjectName,
        };
        filterService.setColumnAdvancedFilter(projectFilterBy, projectNameFilter, ETableFilters.PROJECT);
      }
    };

    const setSelectedClusterById = async (clusterId: string) => {
      try {
        await clusterStore.loadClusters(); // Ensure clusters are loaded with the latest data
        clusterStore.setSelectedClusterById(clusterId);
      } catch (e) {
        new ErrorAlert({
          generalMessage: "Failed to load clusters",
        });
      }
    };

    const validateForm = async (): Promise<boolean> => {
      // Standard form validation
      if (!formRef.value) return false;
      return formRef.value.validate();
    };

    const handleNext = async () => {
      if (onboardingState.currentStep.value === OnboardingStep.Completion) {
        handleClose();
        return;
      }

      const isValid = await validateForm();
      if (!isValid) return;

      if (
        onboardingState.isSsoSetupStep.value &&
        onboardingState.ssoData.value.ssoErrorType === SsoErrorType.IdpCreation
      ) {
        onboardingState.skipSsoSetup();
        return;
      }

      if (onboardingState.isEmailServerSetupStep.value) {
        if (onboardingState.emailServerData.value.saveError || onboardingState.emailServerData.value.isConfigured) {
          onboardingState.skipEmailServerSetup();
          return;
        }

        await onboardingState.saveEmailServer();
        if (!onboardingState.emailServerData.value.saveError) onboardingState.goNext();
        return;
      }

      if (
        onboardingState.isTeamSetupStep.value &&
        onboardingState.currentSubStep.value === TeamSetupSubstep.CreateTeam
      ) {
        await onboardingState.startTeamProvisioning();
        const hasCreatedUsers = (onboardingState.provisioningData.value.createdUsers?.length ?? 0) > 0;
        if (!hasCreatedUsers) {
          onboardingState.goNext();
          onboardingState.goNext();
        } else {
          onboardingState.goNext();
        }
      } else if (
        onboardingState.isTeamSetupStep.value &&
        onboardingState.currentSubStep.value === TeamSetupSubstep.TeamCredentials
      ) {
        onboardingState.goNext();
      } else if (onboardingState.isSsoSetupStep.value && onboardingState.ssoData.value.isDelegated) {
        await onboardingState.startTeamProvisioning();
        onboardingState.goNext();
      } else {
        await onboardingState.goNext(() => {
          handleClose(PROJECT_ROUTE_NAMES.PROJECT_INDEX);
        });
      }
    };

    const handleRetry = () => {
      onboardingState.resetTeamProvisioning();
    };

    const handleSkip = () => {
      if (onboardingState.isSsoSetupStep.value) {
        onboardingState.skipSsoSetup();
      } else if (onboardingState.isTeamSetupStep.value) {
        onboardingState.skipTeamProvisioning();
      } else if (onboardingState.isEmailServerSetupStep.value) {
        onboardingState.skipEmailServerSetup();
      }
    };

    const handleKeydown = async (event: KeyboardEvent) => {
      if (onboardingState.isLoading.value) return;
      await formUtil.enterKeySubmissionHandler(event, handleNext);
    };

    const hiddenSubsteps = computed((): Record<number, number[]> => {
      if (!onboardingState.notificationStore.isEmailChannelAvailable) {
        return {};
      }
      return {
        [OnboardingStep.TeamSetup]: [2],
      };
    });

    return {
      //STATE
      ...onboardingState,
      showCloseButton,
      showSkipButton,
      showNavigation,
      formRef,
      hiddenSubsteps,

      //HANDLERS
      handleNext,
      handleClose,
      handleRetry,
      handleSkip,
      handleKeydown,
      //CONSTANTS
      CLUSTER_ROUTE_NAMES,
      PROJECT_ROUTE_NAMES,
      displayDelegatedStepper,
      showDelegatedTeamSetup,
      initialClusterId,
    };
  },
});
</script>

<style lang="scss">
.admin-onboarding-modal {
  backdrop-filter: blur(6px);

  // Override Quasar's default max-width constraint
  .q-dialog__inner--minimized > div {
    max-width: 1250px;
    max-height: 780px;
    width: 100%;
  }
}

.onboarding-container {
  display: flex;
  height: 100vh;
  background: #ffffff;
  position: relative;
  overflow-x: hidden;

  .content-wrapper {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .close-button {
    position: absolute;
    top: 24px;
    right: 24px;
  }
}
</style>
