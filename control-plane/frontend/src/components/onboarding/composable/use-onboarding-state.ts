import { reactive, computed, watch } from "vue";
import {
  type OnboardingState,
  type IStepConfig,
  type ClusterData,
  type ClusterDataKey,
  type SsoData,
  type SsoDataKey,
  type TeamData,
  type TeamDataKey,
  type ProvisioningData,
  type ProvisioningDataKey,
  type EmailServerData,
  type EmailServerDataKey,
  InstallationErrorType,
  SsoSubstep,
  SsoErrorType,
  OnboardingStep,
} from "@/components/onboarding/models/admin-onboarding.model";
import {
  DEFAULT_CLUSTER_VERSION,
  IOnboardingCompletionState,
  TeamIdentificationMethod,
} from "@/components/onboarding/models/admin-onboarding.model";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { notificationService } from "@/services/control-plane/notifications.service/notification.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { teamProvisioningService } from "@/services/control-plane/team-provisioning.service/team-provisioning.service";
import { installationService } from "@/services/control-plane/installation.service/installation.service";
import { showsClusterLocationStep } from "@/services/infra/boot-config.service/boot-config.service";
import { useAuthStore } from "@/stores/auth.store";
import { EClusterLocations } from "@/models/cluster.model";
import type { UserPostResponse } from "@/swagger-models/identity-manager-client";
import { OPENID_SCOPE } from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";
import { DEFAULT_EMPTY_EMAIL_SETTINGS, defaultMappers } from "@/models/setting.model";
import { settingsUtil } from "@/utils/settings.util/settings.util";
import {
  SamlCreationDataMetadataXmlTypeEnum,
  IdpTypeEnum,
  IdpCreationRequestTypeEnum,
  type OidcCreationData,
  type SamlCreationData,
} from "@/swagger-models/identity-manager-client";
import { useIdpsStore } from "@/stores/idps.store";
import { useClusterStore } from "@/stores/cluster.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { storageUtil } from "@/utils/storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";
import { stepsRequiredPermissions } from "@/components/onboarding/models/admin-onboarding.model";
import { useNotificationStore } from "@/stores/notifications.store";
import { useAppStore } from "@/stores/app.store";

export const ONBOARDING_TITLES = {
  TEAM_SETUP: "Create your first research team",
} as const;

const _getEmptySsoData = (): SsoData => {
  return {
    protocol: "",
    idpMetadata: {},
    runaiMetadata: "",
    isDelegated: false,
    delegateEmail: "",
    hasIdpCreated: false,
    oidcData: {
      clientId: "",
      clientSecret: "",
      discoverDocumentUrl: "",
      scopes: [OPENID_SCOPE],
    },
    samlData: {
      entityId: "",
      metadataXmlUrl: "",
      metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.File,
      metadataXmlFile: "",
      fileName: "",
    },
    mappers: defaultMappers,
    idp: undefined,
    ssoError: null,
    ssoErrorType: null,
  };
};

const _getStepSubSteps = (stepId: number): number => {
  const step = ONBOARDING_STEPS.find((s) => s.id === stepId);
  if (!step) return 1;
  return step.getSubSteps ? step.getSubSteps() : step.subSteps;
};
// Resolve the cluster setup length from the configured capabilities.
export const getClusterSetupSubStepIndex = (): number => {
  return showsClusterLocationStep() ? 4 : 3;
};

export const getTeamSetupSubSteps = (): number => {
  const notificationStore = useNotificationStore();

  if (!notificationStore.isChannelsLoaded) {
    notificationStore.loadNotificationChannels();
  }

  return notificationStore.isEmailChannelAvailable ? 1 : 2;
};

// Step configuration - easy to maintain and extend
const ONBOARDING_STEPS: IStepConfig[] = [
  { id: OnboardingStep.Landing, name: "landing", subSteps: 1, continueLabel: "LET'S GO" },
  {
    id: OnboardingStep.ClusterSetup,
    name: "cluster-setup",
    subSteps: 4,
    getSubSteps: getClusterSetupSubStepIndex,
  },
  { id: OnboardingStep.SsoSetup, name: "sso", subSteps: 3 },
  { id: OnboardingStep.EmailServerSetup, name: "email-server-setup", subSteps: 0 },
  { id: OnboardingStep.TeamSetup, name: "team-setup", subSteps: 2, getSubSteps: getTeamSetupSubSteps },
];

export function useOnboardingState() {
  const authStore = useAuthStore();
  const tenantId = computed(() => authStore.tenant.id);
  const idpStore = useIdpsStore();
  const permissionStore = usePermissionStore();
  const clusterStore = useClusterStore();
  const appStore = useAppStore();
  const isNoClusters = computed(() => clusterStore.clusters.length === 0);
  const notificationStore = useNotificationStore();

  const enum StepName {
    CLUSTER_SETUP = "Install your cluster",
    SSO_SETUP = "Set up SSO (Single Sign-On)",
    EMAIL_SERVER_SETUP = "Set an email server",
    TEAM_SETUP = "Create your first research team",
  }

  const stepEligibilityConfigurations = [
    {
      stepId: OnboardingStep.ClusterSetup,
      stepName: StepName.CLUSTER_SETUP,
      permissionChecks: stepsRequiredPermissions[OnboardingStep.ClusterSetup],
      additionalConditions: [isNoClusters],
    },
    {
      stepId: OnboardingStep.SsoSetup,
      stepName: StepName.SSO_SETUP,
      permissionChecks: stepsRequiredPermissions[OnboardingStep.SsoSetup],
      additionalConditions: [
        !storageUtil.get<boolean>(STORAGE_KEYS.ONBOARDING_SKIPPED_SSO), // Don't show if SSO was previously skipped
        !authStore.isSSO,
      ],
    },
    {
      stepId: OnboardingStep.EmailServerSetup,
      stepName: StepName.EMAIL_SERVER_SETUP,
      permissionChecks: stepsRequiredPermissions[OnboardingStep.EmailServerSetup],
      additionalConditions: [
        !appStore.isMultiTenant,
        !storageUtil.get<boolean>(STORAGE_KEYS.ONBOARDING_SKIPPED_EMAIL_SERVER_SETUP),
        !notificationStore.isEmailChannelAvailable,
      ],
    },
    {
      stepId: OnboardingStep.TeamSetup,
      stepName: StepName.TEAM_SETUP,
      permissionChecks: stepsRequiredPermissions[OnboardingStep.TeamSetup],
      additionalConditions: [
        !storageUtil.get<boolean>(STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP), // Don't show if team setup was previously skipped
      ],
    },
  ];

  const isSsoConfigured = computed(() => authStore.isSSO);

  // Evaluate which steps to display ONCE at initialization
  const stepsToShow: number[] = [];

  for (const stepConfig of stepEligibilityConfigurations) {
    const hasRequiredPermissions = permissionStore.hasAllPermissions(stepConfig.permissionChecks || []);
    const hasAdditionalConditions = stepConfig.additionalConditions.every((condition) => {
      try {
        // Handle computed refs
        if (condition && typeof condition === "object" && "value" in condition) {
          return Boolean(condition.value);
        }
        // Handle plain booleans
        if (typeof condition === "boolean") {
          return condition;
        }
        // Fallback to boolean conversion
        return Boolean(condition);
      } catch (error) {
        return false;
      }
    });
    if (hasRequiredPermissions && hasAdditionalConditions) {
      stepsToShow.push(stepConfig.stepId);
    }
  }

  // Reactive state management
  const state = reactive<OnboardingState>({
    currentStep: OnboardingStep.Landing,
    currentSubStep: 0,
    completedSteps: [],
    canContinue: true,
    isLoading: false,
    clusterData: {
      clusterName: "",
      installationOption: "",
      clusterLocation: "",
      clusterUrl: "",
      hasClusterValues: false,
      clusterVersion: DEFAULT_CLUSTER_VERSION,
      isClusterConnected: false,
      installationCommand: "",
      installationError: null,
      installationErrorType: null,
      isExistingCluster: false,
    },
    ssoData: _getEmptySsoData(),
    teamData: {
      letSomeoneElseFill: false,
      delegateEmail: "",
      teamName: "",
      memberEmails: [],
      identificationMethod: TeamIdentificationMethod.MemberEmails,
      groupName: "",
    },
    visitedSubSteps: {},
    provisioningData: {
      completionState: undefined,
      failedSubjectIds: [],
    },
    emailServerData: {
      settings: DEFAULT_EMPTY_EMAIL_SETTINGS,
      isConfigured: false,
      saveError: false,
    },
  });

  // Helper function to check if we're transitioning from URL to Installation step
  const isUrlToInstallationTransition = (nextStep: number, nextSubStep: number): boolean => {
    if (state.currentStep !== OnboardingStep.ClusterSetup) return false; // Only applies to cluster setup step

    const urlSubStepIndex = showsClusterLocationStep() ? 3 : 2;
    const installationSubStepIndex = getClusterSetupSubStepIndex(); // Installation step index

    return (
      state.currentSubStep === urlSubStepIndex &&
      nextStep === OnboardingStep.ClusterSetup &&
      nextSubStep === installationSubStepIndex
    );
  };

  // Helper function to check if we're transitioning from IdpMetadata to RunaiMetadata substep
  const isIdpMetadataToRunaiMetadataTransition = (nextStep: number, nextSubStep: number): boolean => {
    if (state.currentStep !== OnboardingStep.SsoSetup) return false; // Only applies to SSO step

    return (
      state.currentSubStep === SsoSubstep.IdpMetadata &&
      nextStep === OnboardingStep.SsoSetup &&
      nextSubStep === SsoSubstep.RunaiMetadata
    );
  };

  // Installation preparation logic
  const prepareInstallationData = async (): Promise<boolean> => {
    try {
      state.isLoading = true;

      // // FAKE ERROR RETURN FOR TESTING
      // state.clusterData.installationError = "Failed to create cluster: Connection timeout";
      // state.clusterData.installationErrorType = InstallationErrorType.ClusterCreation;
      // return false;

      // Clear any previous errors
      state.clusterData.installationError = null;
      state.clusterData.installationErrorType = null;

      const result = await installationService.prepareInstallation(
        state.clusterData.clusterName,
        state.clusterData.clusterVersion,
        state.clusterData.clusterUrl,
      );

      if (result.success && result.data) {
        // Store the installation data
        state.clusterData.installationCommand = result.data.installationCommand;
        state.clusterData.clusterId = result.data.createdCluster?.uuid;
        state.clusterData.isExistingCluster = result.data.isExistingCluster;

        return true;
      } else if (result.error) {
        // Store the error information
        state.clusterData.installationError = result.error.message;
        state.clusterData.installationErrorType = result.error.type;

        return false;
      }

      return false;
    } catch (error) {
      // Fallback error handling
      state.clusterData.installationError = error instanceof Error ? error.message : "Unknown error occurred";
      state.clusterData.installationErrorType = InstallationErrorType.ClusterCreation;

      return false;
    } finally {
      state.isLoading = false;
    }
  };

  // Send IDP creation request to the backend
  const sendIdpCreationRequest = async (): Promise<boolean> => {
    try {
      state.isLoading = true;

      // Clear any previous errors
      state.ssoData.ssoError = null;
      state.ssoData.ssoErrorType = null;

      let idpData: OidcCreationData | SamlCreationData;
      let mappers = {};
      let protocolType: string;

      if (state.ssoData.protocol === IdpTypeEnum.Oidc && state.ssoData.oidcData) {
        idpData = state.ssoData.oidcData;
        mappers = state.ssoData.mappers || {};
        protocolType = IdpCreationRequestTypeEnum.Oidc;
      } else if (state.ssoData.protocol === IdpTypeEnum.Saml && state.ssoData.samlData) {
        idpData = state.ssoData.samlData;
        mappers = state.ssoData.mappers || {};
        protocolType = IdpCreationRequestTypeEnum.Saml;
      } else {
        // Store error for missing data
        state.ssoData.ssoError = "Required SSO data is missing";
        state.ssoData.ssoErrorType = SsoErrorType.MissingData;
        return false;
      }

      // Create IDP creation request
      const idpCreationRequest = settingsUtil.createIdpCreationRequest(protocolType, idpData, mappers);

      if (!idpCreationRequest) {
        state.ssoData.ssoError = "Failed to create IDP request";
        state.ssoData.ssoErrorType = SsoErrorType.IdpCreation;
        return false;
      }

      // Send request to backend
      await idpStore.addIdp(idpCreationRequest);

      // Set the hasIdpCreated flag on success
      state.ssoData.hasIdpCreated = true;

      return true;
    } catch (error: unknown) {
      // Store error information based on protocol
      state.ssoData.ssoError = "Failed to create IDP request";
      state.ssoData.ssoErrorType = SsoErrorType.IdpCreation;
      return false;
    } finally {
      state.isLoading = false;
    }
  };

  const startTeamProvisioning = async () => {
    //if there is not clusterData.clusterI then check if there are in the store
    if (!state.clusterData.clusterId) {
      state.clusterData.clusterId = clusterStore.selectedClusterId;
    }

    try {
      state.isLoading = true;

      // Check for delegation in either SSO or team setup
      const isDelegated = state.ssoData.isDelegated || state.teamData.letSomeoneElseFill;
      const delegateEmail = state.ssoData.isDelegated ? state.ssoData.delegateEmail : state.teamData.delegateEmail;

      if (isDelegated) {
        await handleAdminUserProvisioning(delegateEmail);
      } else {
        await handleTeamProvisioning();
      }
    } catch (error) {
      handleProvisioningError();
    } finally {
      state.isLoading = false;
    }
  };

  const skipSsoSetup = () => {
    // Save flag indicating SSO was skipped
    storageUtil.save(STORAGE_KEYS.ONBOARDING_SKIPPED_SSO, true);

    // Clear any SSO data when skipping
    state.ssoData = _getEmptySsoData();

    // Reset SSO step tracking - remove SSO step from completed steps and clear visited substeps
    const ssoStepIndex = state.completedSteps.indexOf(OnboardingStep.SsoSetup);
    if (ssoStepIndex > -1) {
      state.completedSteps.splice(ssoStepIndex, 1);
    }

    // Clear visited substeps for SSO step so user can start fresh if they come back
    if (state.visitedSubSteps[OnboardingStep.SsoSetup]) {
      delete state.visitedSubSteps[OnboardingStep.SsoSetup];
    }

    // Move to next step (team setup)
    moveToNextStep();
  };

  const skipEmailServerSetup = () => {
    storageUtil.save(STORAGE_KEYS.ONBOARDING_SKIPPED_EMAIL_SERVER_SETUP, true);
    moveToNextStep();
  };

  const saveEmailServer = async () => {
    try {
      if (!state.emailServerData.settings) return;
      state.isLoading = true;

      const notificationChannelConfig = settingsUtil.emailSettingsToSmtpConfig(state.emailServerData.settings);
      await notificationService.createEmail({
        config: notificationChannelConfig,
      });
      await enableEmailNotifications();
      state.emailServerData.isConfigured = true;
    } catch (error) {
      console.error("Failed to save email server", error);
      state.emailServerData.saveError = true;
    } finally {
      state.isLoading = false;
    }
  };

  const enableEmailNotifications = async () => {
    try {
      await notificationService.updateEmailNotificationState({ enabled: true });
    } catch (error) {
      console.error("Failed to enable email notifications", error);
    }
  };

  const skipTeamProvisioning = () => {
    // Set completion state for skipped team setup
    storageUtil.save(STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP, true);
    updateProvisioningState(IOnboardingCompletionState.CompletedWithoutProject, []);

    // Move to completion step
    if (isLastStep.value) {
      state.currentStep = OnboardingStep.Completion;
      state.currentSubStep = 1;
    } else {
      moveToNextStep();
    }
  };

  const handleAdminUserProvisioning = async (delegateEmail: string) => {
    // Check if SSO is enabled to determine whether to create local user
    const result = await teamProvisioningService.createAdminUser(delegateEmail, tenantId.value, isSsoConfigured.value);

    updateProvisioningState(
      result.success
        ? IOnboardingCompletionState.CompletedWithoutProject
        : IOnboardingCompletionState.AdminUserCreationFailed,
      result.success ? [] : [delegateEmail],
      result.success && result.createdUser ? [result.createdUser] : undefined,
    );
  };
  const loadUserOrgUnits = async () => {
    try {
      await authStore.loadUserOrgUnits();
    } catch {
      console.error("Failed to load org units");
    }
  };

  const handleTeamProvisioning = async () => {
    if (!state.clusterData.clusterId) return; // for ts warning, (cluster is existing in this case)
    const result = await teamProvisioningService.provisionTeam(
      state.clusterData.clusterId,
      state.teamData,
      isSsoConfigured.value,
    );

    updateProvisioningState(result.completionState, result.failedSubjectIds, result.createdUsers);

    // After project creation, refresh org units so permitted scopes include the new project
    await loadUserOrgUnits();
  };

  const updateProvisioningState = (
    completionState: IOnboardingCompletionState,
    failedSubjectIds: string[],
    createdUsers?: UserPostResponse[],
  ) => {
    state.provisioningData.completionState = completionState;
    state.provisioningData.failedSubjectIds = failedSubjectIds;
    state.provisioningData.createdUsers = createdUsers;
  };

  const handleProvisioningError = () => {
    const isDelegated = state.ssoData.isDelegated || state.teamData.letSomeoneElseFill;
    const delegateEmail = state.ssoData.isDelegated ? state.ssoData.delegateEmail : state.teamData.delegateEmail;
    const failedIds = isDelegated ? [delegateEmail] : [];

    updateProvisioningState(IOnboardingCompletionState.ProjectCreationFailed, failedIds);
  };

  const resetTeamProvisioning = () => {
    // Reset to the team setup step
    moveToPreviousStep();
    // Reset provisioning state for retry
    state.currentSubStep = 1;
    state.provisioningData.completionState = undefined;
    state.provisioningData.failedSubjectIds = [];
  };

  const updateProvisioningData = <K extends ProvisioningDataKey>(key: K, value: ProvisioningData[K]) => {
    state.provisioningData[key] = value;
  };

  // Computed properties for better readability
  const currentStepConfig = computed(() => ONBOARDING_STEPS.find((step) => step.id === state.currentStep));

  const continueButtonLabel = computed(() => {
    // Show "START EXPLORING" when we're on the last step and have completed the onboarding
    if (isLastStep.value && state.provisioningData.completionState) {
      return "START EXPLORING";
    }

    // Show "Skip SSO Setup" when there's an SSO creation error
    if (state.currentStep === OnboardingStep.SsoSetup && state.ssoData.ssoErrorType === SsoErrorType.IdpCreation) {
      return "Skip SSO Setup";
    }

    // Show "Skip Email Server Setup" when there's a email server setup error
    if (state.currentStep === OnboardingStep.EmailServerSetup && state.emailServerData.saveError) {
      return "Skip Email Server Setup";
    }

    if (state.currentStep === OnboardingStep.TeamSetup) {
      const totalSubSteps = _getStepSubSteps(state.currentStep);
      if (state.currentSubStep >= totalSubSteps) {
        return "Done";
      }
      return "Continue";
    }

    return currentStepConfig.value?.continueLabel || "Continue";
  });

  const isLastSubStep = computed(() => {
    if (!currentStepConfig.value) return false;
    const subSteps = _getStepSubSteps(state.currentStep);
    return state.currentSubStep >= subSteps;
  });

  const isLastStep = computed(() => {
    return state.currentStep === OnboardingStep.Completion;
  });

  const clusterCreationCompleted = computed(() => {
    const optionsStepIndex = getClusterSetupSubStepIndex();
    return state.currentSubStep >= optionsStepIndex;
  });

  // Watch for when we have cluster values to mark them as complete
  watch(state, (newState) => {
    if (
      newState.currentStep === OnboardingStep.ClusterSetup &&
      newState.currentSubStep === getClusterSetupSubStepIndex()
    ) {
      state.clusterData.hasClusterValues = true;
    }
  });

  // Helper to mark a substep as visited
  const markSubStepAsVisited = (stepId: number, subStepId: number) => {
    if (!state.visitedSubSteps[stepId]) {
      state.visitedSubSteps[stepId] = new Set();
    }
    state.visitedSubSteps[stepId].add(subStepId);
  };

  // Navigation helper functions
  const markStepAsCompleted = (stepId: number) => {
    if (!state.completedSteps.includes(stepId)) {
      state.completedSteps.push(stepId);
    }
  };

  const moveToNextSubStep = () => {
    markSubStepAsVisited(state.currentStep, state.currentSubStep);
    const subSteps = _getStepSubSteps(state.currentStep);
    if (state.currentSubStep < subSteps) {
      state.currentSubStep++;
    }
  };

  const moveToNextStep = () => {
    if (!isLastStep.value) {
      markStepAsCompleted(state.currentStep);

      // Find the next step to show
      const currentIndex = stepsToShow.indexOf(state.currentStep);
      const nextStep = stepsToShow[currentIndex + 1];

      if (nextStep) {
        state.currentStep = nextStep;
        state.currentSubStep = 1;
      } else {
        // No more steps to show, go to completion
        state.currentStep = OnboardingStep.Completion;
        state.currentSubStep = 1;
      }
    }
  };

  const moveToPreviousSubStep = () => {
    if (state.currentSubStep > 1) {
      state.currentSubStep--;
    }
  };

  const moveToPreviousStep = () => {
    if (state.currentStep > OnboardingStep.Landing) {
      state.currentStep--;
      state.currentSubStep = _getStepSubSteps(state.currentStep);
    }
  };

  // Event handlers
  const handleNavigateToStep = (step: number, subStep: number) => {
    // Prevent navigation to previous steps if backward navigation is disabled
    if (isBackwardNavigationDisabled()) {
      // Only allow navigation to current step or forward
      const currentStepNumber = state.currentStep;
      const currentSubStepNumber = state.currentSubStep;

      if (step < currentStepNumber || (step === currentStepNumber && subStep < currentSubStepNumber)) {
        return; // Prevent backward navigation
      }
    }

    state.currentStep = step;
    state.currentSubStep = subStep;
  };

  const handleCanContinueChange = (canContinue: boolean) => {
    state.canContinue = canContinue;
  };

  const handleClusterDataUpdate = <K extends ClusterDataKey>(key: K, value: ClusterData[K]) => {
    state.clusterData[key] = value;

    // Autofill cluster URL when "Same as control plane" is selected
    if (key === "clusterLocation" && value === EClusterLocations.CONTROL_PLANE) {
      // Generate control plane URL from current location
      const controlPlaneUrl = window.location.origin;
      state.clusterData.clusterUrl = controlPlaneUrl;
    }
    // Clear URL when switching to remote (let user fill it)
    else if (key === "clusterLocation" && value === EClusterLocations.REMOTE) {
      state.clusterData.clusterUrl = "";
    }
  };

  const handleSsoDataUpdate = <K extends SsoDataKey>(key: K, value: SsoData[K]) => {
    if (key === "protocol" && state.ssoData.protocol && state.ssoData.protocol !== value) {
      state.ssoData = _getEmptySsoData();
    }
    state.ssoData[key] = value;
  };

  const handleTeamDataUpdate = <K extends TeamDataKey>(key: K, value: TeamData[K]) => {
    state.teamData[key] = value;
  };

  const handleEmailServerDataUpdate = <K extends EmailServerDataKey>(key: K, value: EmailServerData[K]) => {
    state.emailServerData[key] = value;
  };

  const handleLoadingChange = (isLoading: boolean) => {
    state.isLoading = isLoading;
  };

  // Helper function to check if backward navigation should be prevented
  const isBackwardNavigationDisabled = (): boolean => {
    // Prevent going back if on completion step
    if (state.currentStep >= OnboardingStep.Completion) {
      return true;
    }

    // Prevent going back if team provisioning has been completed successfully
    const completionState = state.provisioningData.completionState;
    return (
      completionState === IOnboardingCompletionState.Completed ||
      completionState === IOnboardingCompletionState.CompletedWithoutProject
    );
  };

  // Main navigation functions
  const goBack = () => {
    // Prevent backward navigation after team completion
    if (isBackwardNavigationDisabled()) {
      return;
    }

    if (state.currentSubStep > 1) {
      moveToPreviousSubStep();
    } else if (state.currentStep > OnboardingStep.Landing) {
      moveToPreviousStep();
    }
  };

  const goNext = async (onComplete: () => void = () => "") => {
    // Handle landing step (step 0) - go to first step to show
    if (state.currentStep === OnboardingStep.Landing) {
      const firstStep = stepsToShow[0] || OnboardingStep.ClusterSetup;
      state.currentStep = firstStep;
      state.currentSubStep = 1;
      return;
    }

    if (state.currentStep === OnboardingStep.SsoSetup && state.ssoData.isDelegated) {
      state.currentStep = OnboardingStep.Completion;
      return;
    }

    // Handle completion of final step
    if (isLastStep.value && isLastSubStep.value) {
      markStepAsCompleted(state.currentStep);
      state.currentStep = OnboardingStep.Completion;
      state.currentSubStep = 1;
      onComplete();
      return;
    }

    // Check if we need to prepare installation before moving to installation step
    const nextSubStep = isLastSubStep.value ? 1 : state.currentSubStep + 1;
    const nextStep = isLastSubStep.value ? state.currentStep + 1 : state.currentStep;

    if (isUrlToInstallationTransition(nextStep, nextSubStep)) {
      // Prepare installation before moving to the next step
      const success = await prepareInstallationData();

      // Only proceed if installation preparation was successful or if there's an error to display
      if (success || state.clusterData.installationErrorType) {
        // Proceed with navigation
        if (!isLastSubStep.value) {
          moveToNextSubStep();
        } else {
          moveToNextStep();
        }
        markSubStepAsVisited(state.currentStep, state.currentSubStep);
      }
      // If installation preparation failed and no error type is set, don't navigate
      return;
    }

    // Check if we need to send IDP creation request before moving to RunaiMetadata step
    if (isIdpMetadataToRunaiMetadataTransition(nextStep, nextSubStep) && !state.ssoData.hasIdpCreated) {
      // Send IDP creation request before moving to the next step
      const success = await sendIdpCreationRequest();

      // Only proceed if IDP creation was successful or if there's an error to display
      if (success) {
        // Proceed with navigation
        if (!isLastSubStep.value) {
          moveToNextSubStep();
        } else {
          moveToNextStep();
        }
        markSubStepAsVisited(state.currentStep, state.currentSubStep);
      }
      // If IDP creation failed and no error type is set, don't navigate
      return;
    }

    // Handle normal sub-step navigation
    if (!isLastSubStep.value) {
      moveToNextSubStep();
    } else {
      moveToNextStep();
    }
    // Mark current substep as visited
    markSubStepAsVisited(state.currentStep, state.currentSubStep);
  };

  // Reset state (useful for testing or restarting the flow)
  const resetState = () => {
    state.currentStep = OnboardingStep.Landing;
    state.currentSubStep = 0;
    state.completedSteps = [];
    state.canContinue = true;
    state.isLoading = false;
    state.clusterData = {
      clusterName: "",
      installationOption: "",
      clusterLocation: "",
      clusterUrl: "",
      clusterVersion: DEFAULT_CLUSTER_VERSION,
      hasClusterValues: false,
      isClusterConnected: false,
      installationCommand: "",
      installationError: null,
      installationErrorType: null,
      isExistingCluster: false,
    };
    state.ssoData = _getEmptySsoData();
    state.teamData = {
      letSomeoneElseFill: false,
      delegateEmail: "",
      teamName: "",
      memberEmails: [],
      identificationMethod: TeamIdentificationMethod.MemberEmails,
      groupName: "",
    };
    state.visitedSubSteps = {};
    state.provisioningData = {
      completionState: undefined,
      failedSubjectIds: [],
    };
  };

  //Load functions
  const fetchLatestClusterVersion = async (): Promise<void> => {
    try {
      state.clusterData.clusterVersion = await clusterService.getLatestClusterVersion();
    } catch (error) {
      new ErrorAlert({
        generalMessage: "Failed to fetch cluster versions. Using default version.",
      });
    }
  };

  fetchLatestClusterVersion();

  return {
    // State (as computed refs for reactivity)
    currentStep: computed(() => state.currentStep),
    currentSubStep: computed(() => state.currentSubStep),
    completedSteps: computed(() => state.completedSteps),
    isLoading: computed(() => state.isLoading),
    canContinue: computed(() => state.canContinue),
    clusterData: computed(() => state.clusterData),
    ssoData: computed(() => state.ssoData),
    teamData: computed(() => state.teamData),
    // Computed properties
    continueButtonLabel,
    currentStepConfig,
    isLastStep,
    isLastSubStep,
    clusterCreationCompleted: computed(() => clusterCreationCompleted.value),
    isLandingStep: computed(() => state.currentStep === OnboardingStep.Landing),
    isClusterSetupStep: computed(() => state.currentStep === OnboardingStep.ClusterSetup),
    isSsoSetupStep: computed(() => state.currentStep === OnboardingStep.SsoSetup),
    isEmailServerSetupStep: computed(() => state.currentStep === OnboardingStep.EmailServerSetup),
    isTeamSetupStep: computed(() => state.currentStep === OnboardingStep.TeamSetup),
    visitedSubSteps: computed(() => state.visitedSubSteps),
    provisioningData: computed(() => state.provisioningData),
    emailServerData: computed(() => state.emailServerData),

    // Event handlers
    handleNavigateToStep,
    handleCanContinueChange,
    handleLoadingChange,
    handleClusterDataUpdate,
    handleSsoDataUpdate,
    handleTeamDataUpdate,
    handleEmailServerDataUpdate,
    startTeamProvisioning,
    skipSsoSetup,
    skipEmailServerSetup,
    skipTeamProvisioning,
    resetTeamProvisioning,
    updateProvisioningData,
    saveEmailServer,

    // Navigation
    goBack,
    goNext,
    resetState,
    isBackwardNavigationDisabled,

    // Steps to show (evaluated once at init)
    stepsToShow: computed(() => stepsToShow),

    notificationStore,
  };
}
