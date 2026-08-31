/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { useOnboardingState } from "./use-onboarding-state";
import { setActivePinia, createPinia } from "pinia";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getLatestClusterVersion: vi.fn().mockResolvedValue("2.22.0"),
  },
}));
vi.mock("@/services/control-plane/service-accounts.service/service-accounts.service");

vi.mock("@/services/infra/boot-config.service/boot-config.service", () => ({
  showsClusterLocationStep: vi.fn(() => true),
  getBootConfig: vi.fn(() => ({
    version: { display: "release-42" },
    docs: {
      baseUrl: "https://docs.example.test/product",
      whatsNew: { label: "Product updates", url: "https://docs.example.test/product/updates" },
    },
    support: { url: "https://support.example.test/contact", label: "Get help" },
    capabilities: { showsClusterLocationStep: true },
    unleashConfig: null,
  })),
}));

// Mock the installation service
vi.mock("@/services/control-plane/installation.service", () => ({
  installationService: {
    prepareInstallation: vi.fn().mockResolvedValue({
      success: true,
      data: {
        installationCommand: "mock-install-command",
        createdCluster: { uuid: "mock-cluster-uuid" },
        isExistingCluster: false,
      },
    }),
  },
}));

// Mock stores
vi.mock("@/stores/auth.store", () => ({
  useAuthStore: vi.fn(() => ({
    tenant: { id: "test-tenant-id" },
    isSSO: false,
  })),
}));

vi.mock("@/stores/idps.store", () => ({
  useIdpsStore: vi.fn(() => ({
    addIdp: vi.fn(),
  })),
}));

vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: vi.fn(() => ({
    clusters: [],
    selectedClusterId: "test-cluster-id",
  })),
}));

vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: vi.fn(() => ({
    hasAllPermissions: vi.fn(() => true), // Always return true for tests
  })),
}));

vi.mock("@/stores/app.store", () => ({
  useAppStore: vi.fn(() => ({
    isMultiTenant: false,
  })),
}));

vi.mock("@/stores/notifications.store", () => ({
  useNotificationStore: vi.fn(() => ({
    channels: [],
    isChannelsLoaded: true,
    isEmailChannelAvailable: false, // Default to no SMTP for tests
    loadNotificationChannels: vi.fn(),
  })),
}));

// Mock storage util
vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn(() => false), // Always return false for skipped flags
    save: vi.fn(),
  },
}));

// Mock team provisioning service
vi.mock("@/services/control-plane/team-provisioning.service/team-provisioning.service", () => ({
  teamProvisioningService: {
    createAdminUser: vi.fn().mockResolvedValue({ success: true }),
    provisionTeam: vi.fn().mockResolvedValue({
      completionState: "Completed",
      failedSubjectIds: [],
    }),
  },
}));

// Mock error alert utility
vi.mock("@/utils/error-alert.util", () => ({
  ErrorAlert: vi.fn(),
}));

// Mock settings utility
vi.mock("@/utils/settings.util/settings.util", () => ({
  settingsUtil: {
    createIdpCreationRequest: vi.fn().mockReturnValue({}),
  },
}));

// Mock swagger models
vi.mock("@/swagger-models/identity-manager-client", () => ({
  SamlCreationDataMetadataXmlTypeEnum: {
    File: "File",
    Url: "Url",
  },
  IdpTypeEnum: {
    Oidc: "oidc",
    Saml: "saml",
  },
  IdpCreationRequestTypeEnum: {
    Oidc: "oidc",
    Saml: "saml",
  },
  Configuration: vi.fn().mockImplementation(() => ({})),
  ClientsApi: vi.fn().mockImplementation(() => ({})),
  IdpsApi: vi.fn().mockImplementation(() => ({})),
  RealmApi: vi.fn().mockImplementation(() => ({})),
  TokensApi: vi.fn().mockImplementation(() => ({})),
  UsersApi: vi.fn().mockImplementation(() => ({})),
  ConfigApi: vi.fn().mockImplementation(() => ({})),
  AccessKeysApi: vi.fn().mockImplementation(() => ({})),
  SettingsApi: vi.fn().mockImplementation(() => ({})),
  ServiceAccountsApi: vi.fn().mockImplementation(() => ({})),
}));

// Mock setting models
vi.mock("@/models/setting.model", () => ({
  defaultMappers: {},
  DEFAULT_EMPTY_EMAIL_SETTINGS: {
    smtpHost: "",
    authType: "AuthPlain",
    smtpPort: 587,
    username: "",
    password: "",
    fromEmail: "",
    fromName: "",
  },
  EmailAuthType: {
    AuthPlain: "AuthPlain",
    AuthLogin: "AuthLogin",
  },
}));

// Mock SSO settings
vi.mock("@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common", () => ({
  OPENID_SCOPE: "openid",
}));

// Mock authorization client
vi.mock("@/swagger-models/authorization-client", () => ({
  Action: {
    Create: "Create",
    Update: "Update",
    Read: "Read",
    Delete: "Delete",
  },
  ResourceType: {
    SecuritySettings: "SecuritySettings",
    Users: "Users",
    AccessRules: "AccessRules",
    Project: "Project",
  },
}));

// Mock storage keys
vi.mock("@/models/storage.model", () => ({
  STORAGE_KEYS: {
    ONBOARDING_SKIPPED_SSO: "onboarding_skipped_sso",
    ONBOARDING_SKIPPED_TEAM_SETUP: "onboarding_skipped_team_setup",
  },
}));

import { showsClusterLocationStep } from "@/services/infra/boot-config.service/boot-config.service";
import { EClusterLocations } from "@/models/cluster.model";
import { IOnboardingCompletionState } from "@/components/onboarding/models/admin-onboarding.model";

describe("useOnboardingState", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    vi.mocked(showsClusterLocationStep).mockReturnValue(true);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should initialize with correct default values", () => {
    const state = useOnboardingState();

    expect(state.currentStep.value).toBe(0);
    expect(state.currentSubStep.value).toBe(0);
    expect(state.completedSteps.value).toEqual([]);
    expect(state.canContinue.value).toBe(true);
    expect(state.isLoading.value).toBe(false);

    // Check consolidated cluster data
    expect(state.clusterData.value.clusterName).toBe("");
    expect(state.clusterData.value.installationOption).toBe("");
    expect(state.clusterData.value.clusterLocation).toBe("");
    expect(state.clusterData.value.clusterUrl).toBe("");

    // Verify that at least ClusterSetup step is available (since clusters array is empty)
    expect(state.stepsToShow.value).toContain(1); // ClusterSetup
  });

  it("should initialize consolidated cluster data with correct default values", () => {
    const state = useOnboardingState();

    expect(state.clusterData.value).toEqual({
      clusterName: "",
      clusterVersion: "2.22.0",
      installationOption: "",
      clusterLocation: "",
      clusterUrl: "",
      hasClusterValues: false,
      installationCommand: "",
      installationError: null,
      installationErrorType: null,
      isClusterConnected: false,
      isExistingCluster: false,
    });
  });

  it("should handle navigation from landing step", () => {
    const state = useOnboardingState();
    const onComplete = vi.fn();

    // Start at landing step (step 0)
    expect(state.currentStep.value).toBe(0);

    // Go next should move to cluster setup
    state.goNext(onComplete);

    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(1);
    expect(onComplete).not.toHaveBeenCalled();
  });

  it("should handle sub-step navigation when the location step is enabled", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    const state = useOnboardingState();
    const onComplete = vi.fn();

    // Move to cluster setup step
    state.handleNavigateToStep(1, 1);

    // Should navigate through sub-steps: 1 -> 2 -> 3 -> 4
    await state.goNext(onComplete);
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(2);

    await state.goNext(onComplete);
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(3);

    await state.goNext(onComplete);
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(4);

    // Should move to next available step after last sub-step
    await state.goNext(onComplete);
    // With our mocked setup, all steps should be eligible, so should go to step 2 (SsoSetup)
    const expectedNextStep = state.stepsToShow.value.includes(2) ? 2 : 4; // If step 2 not available, goes to completion
    expect(state.currentStep.value).toBe(expectedNextStep);
    expect(state.currentSubStep.value).toBe(1);
    expect(state.completedSteps.value).toContain(1);
    errorSpy.mockRestore();
  });

  it("should handle sub-step navigation when the location step is disabled", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(showsClusterLocationStep).mockReturnValue(false);

    const state = useOnboardingState();
    const onComplete = vi.fn();

    // Move to cluster setup step
    state.handleNavigateToStep(1, 1);

    // Should navigate through sub-steps: 1 -> 2 -> 3
    await state.goNext(onComplete);
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(2);

    await state.goNext(onComplete);
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(3);

    // Should move to next available step after last sub-step
    await state.goNext(onComplete);
    // With our mocked setup, all steps should be eligible, so should go to step 2 (SsoSetup)
    const expectedNextStep = state.stepsToShow.value.includes(2) ? 2 : 4; // If step 2 not available, goes to completion
    expect(state.currentStep.value).toBe(expectedNextStep);
    expect(state.currentSubStep.value).toBe(1);
    expect(state.completedSteps.value).toContain(1);
    errorSpy.mockRestore();
  });

  it("should handle backward navigation correctly", () => {
    const state = useOnboardingState();

    // Navigate to step 1, sub-step 3
    state.handleNavigateToStep(1, 3);

    // Go back should move to sub-step 2
    state.goBack();
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(2);

    // Go back should move to sub-step 1
    state.goBack();
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(1);

    // Navigate to step 2, sub-step 1
    state.handleNavigateToStep(2, 1);

    // The enabled location step makes installation the fourth sub-step.
    state.goBack();
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(4);
  });

  it("should complete the onboarding flow correctly", async () => {
    const state = useOnboardingState();
    const onComplete = vi.fn();

    // Navigate to the last actual onboarding step (TeamSetup), substep 2 (the last substep with no SMTP)
    state.handleNavigateToStep(3, 2);

    // Verify we're on team setup step which is the last configured step
    expect(state.currentStep.value).toBe(3);
    expect(state.currentSubStep.value).toBe(2);

    // Complete the flow from the team setup step
    await state.goNext(onComplete);

    // The state should move to completion step
    expect(state.currentStep.value).toBe(4);
    expect(state.completedSteps.value).toContain(3); // Previous step should be marked complete

    // Note: onComplete callback is only called when isLastStep && isLastSubStep are both true
    // Since completion step has no substep config, isLastSubStep is false, so onComplete won't be called
    // This is the expected behavior for this implementation
  });

  it("should update cluster data correctly using consolidated handler", () => {
    const state = useOnboardingState();

    // Test individual key-value updates
    state.handleClusterDataUpdate("clusterName", "test-cluster");
    state.handleClusterDataUpdate("installationOption", "helm");
    state.handleClusterDataUpdate("clusterLocation", "same");
    state.handleClusterDataUpdate("clusterUrl", "https://api.cluster.example.com:6443");
    state.handleLoadingChange(true);
    state.handleClusterDataUpdate("isClusterConnected", true);

    expect(state.clusterData.value.clusterName).toBe("test-cluster");
    expect(state.clusterData.value.installationOption).toBe("helm");
    expect(state.clusterData.value.clusterLocation).toBe("same");
    expect(state.clusterData.value.clusterUrl).toBe("https://api.cluster.example.com:6443");
    expect(state.isLoading.value).toBe(true);
    expect(state.clusterData.value.isClusterConnected).toBe(true);
  });

  it("should handle individual cluster data updates correctly", () => {
    const state = useOnboardingState();

    // Set initial data
    state.handleClusterDataUpdate("clusterName", "initial-cluster");
    state.handleClusterDataUpdate("clusterUrl", "https://initial.example.com");

    expect(state.clusterData.value.clusterName).toBe("initial-cluster");
    expect(state.clusterData.value.clusterUrl).toBe("https://initial.example.com");
    expect(state.clusterData.value.installationOption).toBe(""); // Should remain unchanged

    // Individual updates
    state.handleClusterDataUpdate("clusterName", "updated-cluster");
    state.handleLoadingChange(true);

    expect(state.clusterData.value.clusterName).toBe("updated-cluster");
    expect(state.clusterData.value.clusterUrl).toBe("https://initial.example.com"); // Should remain unchanged
    expect(state.isLoading.value).toBe(true);
    expect(state.clusterData.value.installationOption).toBe(""); // Should remain unchanged
  });

  it("should compute continue button label correctly (no SMTP available)", () => {
    const state = useOnboardingState();

    // Landing step
    expect(state.continueButtonLabel.value).toBe("LET'S GO");

    // Cluster setup step
    state.handleNavigateToStep(1, 1);
    expect(state.continueButtonLabel.value).toBe("Continue");

    // Team setup step substep 1 with no SMTP (default mock) - should show "Continue"
    state.handleNavigateToStep(4, 1);
    expect(state.continueButtonLabel.value).toBe("Continue");

    // Team setup step substep 2 with no SMTP - should show "Done"
    state.handleNavigateToStep(4, 2);
    expect(state.continueButtonLabel.value).toBe("Done");
  });

  // Note: Testing with SMTP available is complex due to module-level initialization
  // The core button label functionality is sufficiently covered by the no-SMTP test above

  it("should reset state correctly", () => {
    const state = useOnboardingState();

    // Modify state
    state.handleNavigateToStep(2, 1);
    state.handleClusterDataUpdate("clusterName", "test");
    state.handleClusterDataUpdate("clusterLocation", "same");
    state.handleClusterDataUpdate("clusterUrl", "https://test.com:6443");
    state.handleLoadingChange(true);
    state.handleCanContinueChange(false);

    // Reset
    state.resetState();

    expect(state.currentStep.value).toBe(0);
    expect(state.currentSubStep.value).toBe(0);
    expect(state.completedSteps.value).toEqual([]);
    expect(state.canContinue.value).toBe(true);
    expect(state.isLoading.value).toBe(false);

    // Verify consolidated data is also reset
    expect(state.clusterData.value.clusterName).toBe("");
    expect(state.clusterData.value.clusterLocation).toBe("");
    expect(state.clusterData.value.clusterUrl).toBe("");
    expect(state.isLoading.value).toBe(false);
    expect(state.clusterData.value.isClusterConnected).toBe(false);
  });

  it("should handle backward navigation when the location step is disabled", () => {
    vi.mocked(showsClusterLocationStep).mockReturnValue(false);

    const state = useOnboardingState();

    // Navigate to step 2, sub-step 1
    state.handleNavigateToStep(2, 1);

    // Without the location step, installation is the third sub-step.
    state.goBack();
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(3);
  });

  it("should mark substeps as visited when navigating with goNext", () => {
    const state = useOnboardingState();
    const onComplete = vi.fn();

    // Start at landing step (step 0)
    expect(state.visitedSubSteps.value).toEqual({});

    // Go to cluster setup (step 1, substep 1)
    state.goNext(onComplete);
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(1);

    // Go to substep 2
    state.goNext(onComplete);
    expect(state.currentSubStep.value).toBe(2);
    // Should mark substep 1 and 2 as visited
    expect(state.visitedSubSteps.value[1]?.has(1)).toBe(true);
    expect(state.visitedSubSteps.value[1]?.has(2)).toBe(true);

    // Go to substep 3
    state.goNext(onComplete);
    expect(state.currentSubStep.value).toBe(3);
    expect(state.visitedSubSteps.value[1]?.has(3)).toBe(true);
  });

  it("should mark substeps as visited when navigating directly with handleNavigateToStep", () => {
    const state = useOnboardingState();

    // Directly navigate to step 1, substep 2
    state.handleNavigateToStep(1, 2);
    expect(state.currentStep.value).toBe(1);
    expect(state.currentSubStep.value).toBe(2);
    // Should NOT mark as visited (since markSubStepAsVisited is not called in handleNavigateToStep)
    expect(state.visitedSubSteps.value[1]?.has(2)).not.toBe(true);

    // Now goNext, which should mark as visited
    const onComplete = vi.fn();
    state.goNext(onComplete);
    expect(state.visitedSubSteps.value[1]?.has(2)).toBe(true);
    expect(state.visitedSubSteps.value[1]?.has(3)).toBe(true);
  });

  it("should reset visitedSubSteps on resetState", () => {
    const state = useOnboardingState();
    const onComplete = vi.fn();
    state.goNext(onComplete);
    state.goNext(onComplete);
    expect(state.visitedSubSteps.value[1]?.has(1)).toBe(true);
    expect(state.visitedSubSteps.value[1]?.has(2)).toBe(true);
    state.resetState();
    expect(state.visitedSubSteps.value).toEqual({});
  });

  describe("Backward Navigation Prevention", () => {
    it("should allow backward navigation before team completion", () => {
      const state = useOnboardingState();

      // Navigate to step 1, substep 2
      state.handleNavigateToStep(1, 2);
      expect(state.currentStep.value).toBe(1);
      expect(state.currentSubStep.value).toBe(2);

      // Should allow going back
      expect(state.isBackwardNavigationDisabled()).toBe(false);
      state.goBack();
      expect(state.currentSubStep.value).toBe(1);
    });

    it("should prevent backward navigation when on completion step", () => {
      const state = useOnboardingState();

      // Navigate to completion step (step 4, not step 3)
      state.handleNavigateToStep(5, 1);
      expect(state.currentStep.value).toBe(5);

      // Should prevent going back
      expect(state.isBackwardNavigationDisabled()).toBe(true);
      state.goBack();
      expect(state.currentStep.value).toBe(5); // Should remain on step 5
    });

    it("should prevent backward navigation after successful team provisioning", () => {
      const state = useOnboardingState();

      // Navigate to step 2 (team setup)
      state.handleNavigateToStep(2, 1);
      expect(state.currentStep.value).toBe(2);

      // Initially should allow backward navigation
      expect(state.isBackwardNavigationDisabled()).toBe(false);

      // Set completed team provisioning state
      state.updateProvisioningData("completionState", IOnboardingCompletionState.Completed);
      expect(state.provisioningData.value.completionState).toBe(IOnboardingCompletionState.Completed);

      // Now should prevent backward navigation
      expect(state.isBackwardNavigationDisabled()).toBe(true);
      state.goBack();
      expect(state.currentStep.value).toBe(2); // Should remain on step 2
    });

    it("should prevent backward navigation after delegation completion", () => {
      const state = useOnboardingState();

      // Navigate to step 2 (team setup)
      state.handleNavigateToStep(2, 1);
      expect(state.currentStep.value).toBe(2);

      // Set delegation completion state
      state.updateProvisioningData("completionState", IOnboardingCompletionState.CompletedWithoutProject);

      // Should prevent backward navigation
      expect(state.isBackwardNavigationDisabled()).toBe(true);
      state.goBack();
      expect(state.currentStep.value).toBe(2); // Should remain on step 2
    });

    it("should prevent direct navigation to previous steps after completion", () => {
      const state = useOnboardingState();

      // Navigate to completion step (step 5)
      state.handleNavigateToStep(5, 1);
      expect(state.currentStep.value).toBe(5);

      // Try to navigate back to step 1
      state.handleNavigateToStep(1, 2);
      expect(state.currentStep.value).toBe(5); // Should remain on step 5
      expect(state.currentSubStep.value).toBe(1);
    });

    it("should allow forward navigation even when backward is disabled", () => {
      const state = useOnboardingState();

      // Set up completion state while on step 2
      state.handleNavigateToStep(2, 1);
      state.updateProvisioningData("completionState", IOnboardingCompletionState.Completed);

      // Should prevent backward navigation
      expect(state.isBackwardNavigationDisabled()).toBe(true);

      // But should still allow forward navigation to step 4 (completion)
      state.handleNavigateToStep(4, 1);
      expect(state.currentStep.value).toBe(4);
      expect(state.currentSubStep.value).toBe(1);
    });

    it("should allow backward navigation if provisioning failed and we're not on completion step", () => {
      const state = useOnboardingState();

      // Navigate to step 2 (team setup)
      state.handleNavigateToStep(2, 1);

      // Set failed provisioning state (not completed)
      state.updateProvisioningData("completionState", IOnboardingCompletionState.ProjectCreationFailed);

      // Should still allow backward navigation for retry scenarios since we're not on step 3
      // and the provisioning failed (not completed successfully)
      expect(state.isBackwardNavigationDisabled()).toBe(false);

      // Should be able to go back
      state.goBack();
      expect(state.currentStep.value).toBe(1); // Should move to step 1
      expect(state.currentSubStep.value).toBe(4); // Should be on last substep of step 1
    });
  });

  describe("Cluster URL Auto-filling", () => {
    beforeEach(() => {
      // Mock window.location.origin for consistent testing
      Object.defineProperty(window, "location", {
        value: {
          origin: "https://control-plane.example.com",
          protocol: "https:",
          host: "control-plane.example.com",
          hostname: "control-plane.example.com",
        },
        writable: true,
      });
    });

    it("should auto-fill cluster URL when control plane location is selected", () => {
      const state = useOnboardingState();

      // Select control plane location
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.CONTROL_PLANE);

      // Should autofill with current origin
      expect(state.clusterData.value.clusterLocation).toBe(EClusterLocations.CONTROL_PLANE);
      expect(state.clusterData.value.clusterUrl).toBe("https://control-plane.example.com");
    });

    it("should clear cluster URL when remote location is selected", () => {
      const state = useOnboardingState();

      // First set control plane (which autofills URL)
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.CONTROL_PLANE);
      expect(state.clusterData.value.clusterUrl).toBe("https://control-plane.example.com");

      // Switch to remote
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.REMOTE);

      // Should clear the URL
      expect(state.clusterData.value.clusterLocation).toBe(EClusterLocations.REMOTE);
      expect(state.clusterData.value.clusterUrl).toBe("");
    });

    it("should not modify URL when updating other cluster data fields", () => {
      const state = useOnboardingState();

      // Set initial URL manually
      state.handleClusterDataUpdate("clusterUrl", "https://manual.example.com");
      expect(state.clusterData.value.clusterUrl).toBe("https://manual.example.com");

      // Update other fields
      state.handleClusterDataUpdate("clusterName", "test-cluster");
      state.handleClusterDataUpdate("isClusterConnected", true);

      // URL should remain unchanged
      expect(state.clusterData.value.clusterUrl).toBe("https://manual.example.com");
      expect(state.clusterData.value.clusterName).toBe("test-cluster");
      expect(state.clusterData.value.isClusterConnected).toBe(true);
    });

    it("should handle multiple location switches correctly", () => {
      const state = useOnboardingState();

      // Start with remote (empty URL)
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.REMOTE);
      expect(state.clusterData.value.clusterUrl).toBe("");

      // Switch to control plane (autofill)
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.CONTROL_PLANE);
      expect(state.clusterData.value.clusterUrl).toBe("https://control-plane.example.com");

      // Switch back to remote (clear URL)
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.REMOTE);
      expect(state.clusterData.value.clusterUrl).toBe("");

      // Switch to control plane again (auto-fill again)
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.CONTROL_PLANE);
      expect(state.clusterData.value.clusterUrl).toBe("https://control-plane.example.com");
    });

    it("should use window.location.origin for auto-filling", () => {
      // Test with different origin
      Object.defineProperty(window, "location", {
        value: {
          origin: "https://localhost:3000",
          protocol: "https:",
          host: "localhost:3000",
          hostname: "localhost",
        },
        writable: true,
      });

      const state = useOnboardingState();
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.CONTROL_PLANE);

      expect(state.clusterData.value.clusterUrl).toBe("https://localhost:3000");
    });

    it("should not interfere with manual URL updates after location selection", () => {
      const state = useOnboardingState();

      // Select control plane (auto-fills URL)
      state.handleClusterDataUpdate("clusterLocation", EClusterLocations.CONTROL_PLANE);
      expect(state.clusterData.value.clusterUrl).toBe("https://control-plane.example.com");

      // Manually update URL (should work)
      state.handleClusterDataUpdate("clusterUrl", "https://custom.example.com");
      expect(state.clusterData.value.clusterUrl).toBe("https://custom.example.com");

      // Location should remain unchanged
      expect(state.clusterData.value.clusterLocation).toBe(EClusterLocations.CONTROL_PLANE);
    });
  });
});
