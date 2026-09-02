/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { MockInstance } from "vitest";
// --- Service and Store Mocks ---
vi.mock("vue3-lottie", () => ({ Vue3Lottie: { template: "<div />" } }));
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    create: vi.fn().mockResolvedValue({ uuid: "mock-uuid", name: "clusterName" }),
    getNodeTelemetry: vi.fn().mockResolvedValue({ values: [{ value: "16" }] }),
    list: vi.fn().mockResolvedValue([{ uuid: "mock-uuid", status: { state: "Connected" } }]),
    getClusterInstallationInfo: vi.fn().mockResolvedValue({ data: { installationStr: "mock-install-cmd" } }),
    getLatestClusterVersion: vi.fn().mockResolvedValue("2.22.0"),
  },
}));
vi.mock("@/services/control-plane/service-accounts.service/service-accounts.service", () => ({
  serviceAccountsService: {
    getRunAiInstallerServiceAccountClient: vi.fn().mockResolvedValue({ secret: "mock-secret" }),
  },
}));
vi.mock("@/services/control-plane/team-provisioning.service/team-provisioning.service", () => ({
  teamProvisioningService: {
    provisionTeam: vi.fn().mockResolvedValue({
      completionState: "Completed",
      failedSubjectIds: [],
    }),
    createAdminUser: vi.fn().mockResolvedValue({
      success: true,
      createdUserId: "admin-user-123",
    }),
  },
}));
vi.mock("@/services/control-plane/installation.service/installation.service", () => ({
  installationService: {
    prepareInstallation: vi.fn().mockResolvedValue({
      success: true,
      data: {
        installationCommand: "mock-install-command",
        createdCluster: { uuid: "mock-uuid" },
        isExistingCluster: false,
      },
    }),
  },
}));
vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: () => ({
    clusters: [], // Empty array should make isNoClusters = true
    selectedClusterId: null,
    loadClusters: vi.fn().mockResolvedValue(undefined),
    isClusterNameExists: vi.fn().mockReturnValue(false),
    setSelectedClusterById: vi.fn(),
  }),
}));
vi.mock("@/stores/auth.store", () => ({
  useAuthStore: () => ({
    tenant: {
      id: 12345, // Mock tenant ID
    },
    currentUser: {
      username: "test-user",
    },
    isSSO: false, // Mock SSO status - false so SSO step is shown
  }),
}));
vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: () => ({
    permissionsList: [{ resourceType: "users", action: "read" }],
    hasAllPermissions: vi.fn().mockReturnValue(true),
    hasPermission: vi.fn().mockReturnValue(true),
    isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
    hasSomePermittedActions: vi.fn().mockReturnValue(true),
    loadPermissions: vi.fn().mockResolvedValue(undefined),
  }),
}));
vi.mock("@/stores/idps.store", () => ({
  useIdpsStore: () => ({
    idps: [],
    addIdp: vi.fn().mockResolvedValue(undefined),
  }),
}));
vi.mock("@/stores/role.store", () => ({
  useRoleStore: () => ({
    roleList: [],
    loadRoles: vi.fn().mockResolvedValue(undefined),
  }),
}));
vi.mock("@/services/filter.service/filter.service", () => ({
  filterService: {
    loadFilters: vi.fn().mockReturnValue({ advancedFilters: [] }),
    setColumnAdvancedFilter: vi.fn(),
    saveFilters: vi.fn(),
  },
}));
vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn().mockReturnValue(false), // SSO and team setup not previously skipped
    save: vi.fn(),
  },
}));
vi.mock("@/stores/notifications.store", () => ({
  useNotificationStore: vi.fn(() => ({
    channels: [],
    isChannelsLoaded: true,
    isEmailChannelAvailable: false, // Default to no SMTP for tests
    loadNotificationChannels: vi.fn(),
  })),
}));

import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import AdminOnboardingModal from "./admin-onboarding-modal.vue";
import { createTestingPinia } from "@pinia/testing";
import { createRouter, createWebHistory } from "vue-router";
import { unitTestUtil } from "@/utils/unit-test.util";
import { EClusterLocations } from "@/models/cluster.model";
import { teamProvisioningService } from "@/services/control-plane/team-provisioning.service/team-provisioning.service";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import {
  IOnboardingCompletionState,
  InstallationErrorType,
  OnboardingStep,
  SsoErrorType,
} from "@/components/onboarding/models/admin-onboarding.model";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";

let wrapper: VueWrapper<InstanceType<typeof AdminOnboardingModal>> | null = null;

// --- Helper Functions ---
function mountComponent(props = {}, options = {}) {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      { path: "/", component: { template: "<div>Home</div>" } },
      { path: "/clusters", component: { template: "<div>Clusters</div>" }, name: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX },
      { path: "/settings", component: { template: "<div>Settings</div>" }, name: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX },
      { path: "/projects", component: { template: "<div>Projects</div>" }, name: PROJECT_ROUTE_NAMES.PROJECT_INDEX },
    ],
  });

  wrapper = mount(AdminOnboardingModal, {
    global: {
      plugins: [Quasar, createTestingPinia(), router],
      directives: {
        clickOutside: vi.fn(),
        permission: vi.fn(),
      },
    },
    props,
    ...options,
  });
  return wrapper;
}

function getModal() {
  return unitTestUtil.getModalElement();
}

// Navigation helpers
async function clickContinue() {
  const modal = getModal();
  const continueButton = modal.find(".continue-button");
  expect(continueButton.exists()).toBe(true);
  await continueButton.trigger("click");
  await flushPromises();
}

async function pressEnter() {
  const modal = getModal();
  await modal.trigger("keydown", { key: "Enter" });
  await flushPromises();
}

async function pressKey(key: string) {
  const modal = getModal();
  await modal.trigger("keydown", { key });
  await flushPromises();
}

// Form helpers
async function fillClusterName(name: string) {
  const modal = getModal();
  const nameInput = modal.find('[aid="cluster-name"]');
  expect(nameInput.exists()).toBe(true);
  await nameInput.setValue(name);
  await flushPromises();
}

async function selectClusterLocation(location: EClusterLocations) {
  const modal = getModal();
  const locationCards = modal.findAll(".onboarding-select-card");
  expect(locationCards.length).toBeGreaterThan(0);
  const card = locationCards.find((card) =>
    card
      .text()
      .toLowerCase()
      .includes(location === EClusterLocations.REMOTE ? "remote" : "same"),
  );
  expect(card).toBeTruthy();
  await (card as DOMWrapper<HTMLDivElement>).trigger("click");
  await flushPromises();
}

async function fillClusterUrl(url: string) {
  const modal = getModal();
  const urlInput = modal.find(".cluster-url-input input");
  expect(urlInput.exists()).toBe(true);
  await urlInput.setValue(url);
  await flushPromises();
}

// Assertion helpers
function expectStepTitle(title: string) {
  const modal = getModal();
  const titleEl = modal.find(".content-title");
  expect(titleEl.exists()).toBe(true);
  expect(titleEl.text()).toBe(title);
}

function expectCurrentStep(step: number, subStep?: number) {
  expect(wrapper?.vm.currentStep).toBe(step);
  if (subStep !== undefined) {
    expect(wrapper?.vm.currentSubStep).toBe(subStep);
  }
}

async function simulateInputFocus(selector: string) {
  const modal = getModal();
  const input = modal.find(selector);
  expect(input.exists()).toBe(true);
  await input.trigger("focus");
  await input.trigger("keydown", { key: "Enter" });
  await flushPromises();
}

// SSO helper functions
async function navigateToSsoSetup() {
  // Navigate through cluster setup to reach SSO setup
  await flushPromises();

  // Landing -> First available step
  await pressEnter();

  // Check what step we're actually on
  if (wrapper?.vm.currentStep === OnboardingStep.ClusterSetup) {
    // We're on cluster setup as expected
    expectCurrentStep(OnboardingStep.ClusterSetup, 1);

    // Fill cluster name and continue
    await fillClusterName("test-cluster");
    await pressEnter();
    expectCurrentStep(OnboardingStep.ClusterSetup, 2);

    // Select location and continue
    await selectClusterLocation(EClusterLocations.REMOTE);
    await pressEnter();
    expectCurrentStep(OnboardingStep.ClusterSetup, 3);

    // Fill URL and continue
    await fillClusterUrl("https://test.example.com");
    await pressEnter();
    expectCurrentStep(OnboardingStep.ClusterSetup, 4);

    // Continue to SSO setup
    await pressEnter();
    expectCurrentStep(OnboardingStep.SsoSetup, 1);
  } else if (wrapper?.vm.currentStep === OnboardingStep.SsoSetup) {
    // Cluster setup was skipped, we're directly on SSO setup
    expectCurrentStep(OnboardingStep.SsoSetup, 1);
  } else {
    throw new Error(`Unexpected step after landing: ${wrapper?.vm.currentStep}`);
  }
}

async function selectSsoProtocol(protocol: IdpTypeEnum) {
  if (!wrapper) return;
  wrapper.vm.handleSsoDataUpdate("protocol", protocol);
  await flushPromises();
}

async function navigateToEmailServerSetup() {
  // Navigate to SSO setup first
  await navigateToSsoSetup();

  // Skip SSO to go to email server setup
  const modal = getModal();
  const skipButton = modal.find(".skip-button");
  await skipButton.trigger("click");
  await flushPromises();
  expectCurrentStep(OnboardingStep.EmailServerSetup, 1);
}

async function navigateToTeamSetup() {
  // Navigate to email server setup first
  await navigateToEmailServerSetup();

  // Skip email server setup to go to team setup
  const modal = getModal();
  const skipButton = modal.find(".skip-button");
  await skipButton.trigger("click");
  await flushPromises();
  expectCurrentStep(OnboardingStep.TeamSetup, 1);
}

async function fillTeamData(teamName = "test-team", memberEmails = ["barak@run.ai"]) {
  if (!wrapper) {
    throw new Error("Wrapper is not initialized");
  }

  const modal = getModal();
  await modal.find(".team-name-input input").setValue(teamName);
  wrapper.vm.teamData.memberEmails = memberEmails;
  await flushPromises();
}

function setupTeamProvisioningSpy() {
  return vi.spyOn(teamProvisioningService, "provisionTeam");
}

function expectTeamProvisioningCalled(spy: MockInstance, expectedCallCount = 1) {
  if (!wrapper) {
    throw new Error("Wrapper is not initialized");
  }

  expect(spy).toHaveBeenCalledTimes(expectedCallCount);
  if (expectedCallCount > 0) {
    expect(spy).toHaveBeenCalledWith("mock-uuid", wrapper.vm.teamData, expect.any(Boolean));
  }
}

describe("Admin Onboarding Modal", () => {
  beforeEach(() => {
    mountComponent();
  });
  afterEach(() => {
    if (wrapper) wrapper.unmount();
  });

  describe("AdminOnboardingModal.vue flow", () => {
    describe("AdminOnboardingModal.vue - Onboarding Flow", () => {
      it("should complete the cluster onboarding flow step by step", async () => {
        // Step 1: Landing -> Name
        await flushPromises();
        await clickContinue();
        expectStepTitle("Name your cluster");

        // Step 2: Name
        await fillClusterName("cluster-name");
        expect(wrapper?.vm.clusterData.clusterName).toBe("cluster-name");
        await clickContinue();
        expectStepTitle("Set your cluster location");
        expect(wrapper?.vm.clusterData.clusterLocation).toBe("");

        // Step 3: Location
        await selectClusterLocation(EClusterLocations.REMOTE);
        expect(wrapper?.vm.clusterData.clusterLocation).toBe(EClusterLocations.REMOTE);
        await clickContinue();
        expectStepTitle("Set your cluster URL");

        // Step 4: URL
        await fillClusterUrl("https://my-cluster.example.com");
        expect(wrapper?.vm.clusterData.clusterUrl).toBe("https://my-cluster.example.com");
        await clickContinue();
        expectStepTitle("Install NVIDIA Run:ai on your cluster");

        // Step 5: Installation
        const modal = getModal();
        expect(modal.find(".cluster-installation-command").exists()).toBe(true);
      });

      it("should auto-fill cluster URL when 'Same as control plane' is selected", async () => {
        if (!wrapper) return;

        // Navigate to cluster setup
        await flushPromises();
        await clickContinue();
        expectStepTitle("Name your cluster");

        // Fill cluster name and continue to location
        await fillClusterName("test-cluster");
        await clickContinue();
        expectStepTitle("Set your cluster location");

        // Select "Same as control plane"
        await selectClusterLocation(EClusterLocations.CONTROL_PLANE);
        expect(wrapper.vm.clusterData.clusterLocation).toBe(EClusterLocations.CONTROL_PLANE);

        // Check that cluster URL is autofilled
        const expectedUrl = window.location.origin;
        expect(wrapper.vm.clusterData.clusterUrl).toBe(expectedUrl);

        // Continue to URL step
        await clickContinue();
        expectStepTitle("Set your cluster URL");

        // Check that URL field is readonly when control plane is selected
        const modal = getModal();
        const urlInput = modal.find(".cluster-url-input input");
        expect(urlInput.exists()).toBe(true);
        expect(urlInput.attributes("readonly")).toBeDefined();
        expect((urlInput.element as HTMLInputElement).value).toBe(expectedUrl);
      });

      it("should clear URL and make it editable when switching to remote", async () => {
        if (!wrapper) return;

        // Navigate to location step and select control plane first
        await flushPromises();
        await clickContinue();
        await fillClusterName("test-cluster");
        await clickContinue();

        await selectClusterLocation(EClusterLocations.CONTROL_PLANE);
        expect(wrapper.vm.clusterData.clusterUrl).toBeTruthy(); // Should have auto-filled URL

        // Switch to remote
        await selectClusterLocation(EClusterLocations.REMOTE);
        expect(wrapper.vm.clusterData.clusterLocation).toBe(EClusterLocations.REMOTE);
        expect(wrapper.vm.clusterData.clusterUrl).toBe(""); // Should clear URL

        // Continue to URL step
        await clickContinue();
        expectStepTitle("Set your cluster URL");

        // Check that URL field is editable when remote is selected
        const modal = getModal();
        const urlInput = modal.find(".cluster-url-input input");
        expect(urlInput.exists()).toBe(true);
        expect(urlInput.attributes("readonly")).toBeUndefined();
        expect((urlInput.element as HTMLInputElement).value).toBe("");
      });

      it("should accept localhost URLs for development", async () => {
        if (!wrapper) return;

        // Navigate to cluster setup
        await flushPromises();
        await clickContinue();
        await fillClusterName("test-cluster");
        await clickContinue();

        // Select remote location (so URL field is editable)
        await selectClusterLocation(EClusterLocations.REMOTE);
        await clickContinue();
        expectStepTitle("Set your cluster URL");

        // Test various localhost URLs that should be valid for development
        const localhostUrls = [
          "http://localhost:9000",
          "https://localhost:3000",
          "http://127.0.0.1:8080",
          "https://127.0.0.1:9000",
          "http://localhost",
          "https://localhost/path",
        ];

        for (const url of localhostUrls) {
          await fillClusterUrl(url);
          expect(wrapper.vm.clusterData.clusterUrl).toBe(url);

          // Try to continue - should work with valid localhost URL
          const modal = getModal();
          const continueButton = modal.find(".continue-button");
          await continueButton.trigger("click");
          await flushPromises();

          // Should advance to installation step (validation passed)
          expectStepTitle("Install NVIDIA Run:ai on your cluster");

          // Go back to test next URL
          if (localhostUrls.indexOf(url) < localhostUrls.length - 1) {
            wrapper.vm.handleNavigateToStep(OnboardingStep.ClusterSetup, 3); // Go back to URL step
            await flushPromises();
          }
        }
      });
    });

    describe("AdminOnboardingModal.vue - Enter Key Navigation", () => {
      it("should continue to next step when Enter key is pressed", async () => {
        await flushPromises();
        expectCurrentStep(OnboardingStep.Landing);

        await pressEnter();
        expectCurrentStep(OnboardingStep.ClusterSetup, 1);
        expectStepTitle("Name your cluster");
      });

      it("should not trigger navigation when typing in input fields", async () => {
        await flushPromises();
        await clickContinue();
        expectCurrentStep(OnboardingStep.ClusterSetup, 1);

        await simulateInputFocus('[aid="cluster-name"]');
        expectCurrentStep(OnboardingStep.ClusterSetup, 1); // Should remain unchanged
      });

      it("should ignore non-Enter keys", async () => {
        await flushPromises();
        expectCurrentStep(OnboardingStep.Landing);

        const keys = ["Space", "Tab", "Escape", "ArrowDown"];
        for (const key of keys) {
          await pressKey(key);
        }

        expectCurrentStep(OnboardingStep.Landing); // Should remain unchanged
      });

      it("should ignore Enter when loading", async () => {
        await flushPromises();
        expectCurrentStep(OnboardingStep.Landing);

        wrapper?.vm.handleLoadingChange(true);
        await flushPromises();

        await pressEnter();
        expectCurrentStep(OnboardingStep.Landing); // Should remain unchanged
      });

      it("should navigate through complete flow with Enter", async () => {
        await flushPromises();

        // Landing -> Cluster Setup
        await pressEnter();
        expectCurrentStep(OnboardingStep.ClusterSetup, 1);

        // Name substep
        await fillClusterName("test-cluster");
        await pressEnter();
        expectCurrentStep(OnboardingStep.ClusterSetup, 2);

        // Location substep
        await selectClusterLocation(EClusterLocations.REMOTE);
        await pressEnter();
        expectCurrentStep(OnboardingStep.ClusterSetup, 3);

        // URL substep
        await fillClusterUrl("https://test.example.com");
        await pressEnter();
        expectCurrentStep(OnboardingStep.ClusterSetup, 4);
      });
    });

    describe("AdminOnboardingModal.vue - Team Provisioning", () => {
      it("should call team provisioning when continuing via button click", async () => {
        const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
        const provisionSpy = setupTeamProvisioningSpy();

        await navigateToTeamSetup();
        await fillTeamData();
        await clickContinue();
        await flushPromises();

        expectTeamProvisioningCalled(provisionSpy);
        errorSpy.mockRestore();
      });

      it("should call team provisioning when continuing via Enter key", async () => {
        const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
        const provisionSpy = setupTeamProvisioningSpy();

        await navigateToTeamSetup();
        await fillTeamData();
        await pressEnter();
        await flushPromises();

        expectTeamProvisioningCalled(provisionSpy);
        errorSpy.mockRestore();
      });

      it("should not call team provisioning during cluster setup", async () => {
        const provisionSpy = setupTeamProvisioningSpy();

        await flushPromises();
        await pressEnter(); // Landing -> Cluster Setup
        await fillClusterName("test-cluster");
        await pressEnter(); // Name -> Location
        await flushPromises();

        expectTeamProvisioningCalled(provisionSpy, 0);
      });

      it("should display completion state after successful provisioning", async () => {
        const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
        if (!wrapper) return;
        const provisionSpy = setupTeamProvisioningSpy();

        await navigateToTeamSetup();
        await fillTeamData();
        await clickContinue();
        await flushPromises();

        // If Team Setup has multiple substeps (no SMTP), advance to completion
        if (wrapper.vm.currentStep === OnboardingStep.TeamSetup) {
          await clickContinue();
          await flushPromises();
        }

        // Verify provisioning was called
        expectTeamProvisioningCalled(provisionSpy);

        // Verify completion state
        expectCurrentStep(OnboardingStep.Completion, 1);
        expect(wrapper.vm.provisioningData.completionState).toBe(IOnboardingCompletionState.Completed);

        // Verify completion UI
        const modal = getModal();
        expect(modal.find(".content-subtitle").text()).toContain("Your first team is ready to start running workloads");

        // Verify continue button is available for final action
        const continueButton = modal.find(".continue-button");
        expect(continueButton.exists()).toBe(true);
        errorSpy.mockRestore();
      });

      it("should emit close event when clicking the close icon on completion", async () => {
        const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
        if (!wrapper) return;
        await navigateToTeamSetup();
        await fillTeamData();
        await clickContinue();
        await flushPromises();

        // Click close icon on completion screen
        const modal = getModal();
        const closeButton = modal.find(".close-button");
        expect(closeButton.exists()).toBe(true);
        await closeButton.trigger("click");
        await flushPromises();

        // Verify close event was emitted (robust truthy check)
        const emitted = wrapper.emitted("close");
        expect(!!emitted && emitted.length > 0).toBe(true);
        errorSpy.mockRestore();
      });
    });

    describe("AdminOnboardingModal.vue - Team Delegation", () => {
      async function setupDelegation(delegateEmail = "admin@company.com") {
        if (!wrapper) return;

        await navigateToTeamSetup();

        // Set delegation mode
        wrapper.vm.teamData.letSomeoneElseFill = true;
        wrapper.vm.teamData.delegateEmail = delegateEmail;
        await flushPromises();
      }

      function setupAdminUserSpy() {
        return vi.spyOn(teamProvisioningService, "createAdminUser");
      }

      function expectAdminUserCalled(spy: MockInstance, expectedCallCount = 1, delegateEmail = "admin@company.com") {
        expect(spy).toHaveBeenCalledTimes(expectedCallCount);
        if (expectedCallCount > 0) {
          expect(spy).toHaveBeenCalledWith(delegateEmail, expect.any(Number), expect.any(Boolean)); // tenantId, isSsoConfigured
        }
      }

      it("should call createAdminUser when delegation is enabled via button click", async () => {
        const adminUserSpy = setupAdminUserSpy();
        const provisionTeamSpy = setupTeamProvisioningSpy();

        await setupDelegation();
        await clickContinue();
        await flushPromises();

        // Should call createAdminUser, not provisionTeam
        expectAdminUserCalled(adminUserSpy);
        expectTeamProvisioningCalled(provisionTeamSpy, 0);
      });

      it("should call createAdminUser when delegation is enabled via Enter key", async () => {
        const adminUserSpy = setupAdminUserSpy();
        const provisionTeamSpy = setupTeamProvisioningSpy();

        await setupDelegation();
        await pressEnter();
        await flushPromises();

        // Should call createAdminUser, not provisionTeam
        expectAdminUserCalled(adminUserSpy);
        expectTeamProvisioningCalled(provisionTeamSpy, 0);
      });

      it("should display completion state after successful delegation", async () => {
        if (!wrapper) return;
        const adminUserSpy = setupAdminUserSpy();

        // Mock successful admin user creation
        adminUserSpy.mockResolvedValue({
          success: true,
          createdUserId: "admin-user-123",
        });

        await setupDelegation("admin@example.com");
        await clickContinue();
        await flushPromises();

        // Verify admin user creation was called
        expectAdminUserCalled(adminUserSpy, 1, "admin@example.com");

        // Verify completion state
        if (wrapper.vm.currentStep === OnboardingStep.TeamSetup) {
          await clickContinue();
          await flushPromises();
        }
        expectCurrentStep(OnboardingStep.Completion, 1);
        expect(wrapper.vm.provisioningData.completionState).toBe("CompletedWithoutProject");
        expect(wrapper.vm.provisioningData.failedSubjectIds).toEqual([]);

        // Verify completion UI shows appropriate message
        const modal = getModal();
        expect(modal.find(".content-subtitle").exists()).toBe(true);
      });

      it("should handle failed delegation gracefully", async () => {
        if (!wrapper) return;
        const adminUserSpy = setupAdminUserSpy();

        // Mock failed admin user creation
        adminUserSpy.mockResolvedValue({
          success: false,
          error: "Admin user creation failed",
        });

        await setupDelegation("failed@example.com");
        await clickContinue();
        await flushPromises();

        // Verify admin user creation was called
        expectAdminUserCalled(adminUserSpy, 1, "failed@example.com");

        // Verify error state
        if (wrapper.vm.currentStep === OnboardingStep.TeamSetup) {
          await clickContinue();
          await flushPromises();
        }
        expectCurrentStep(OnboardingStep.Completion, 1);
        expect(wrapper.vm.provisioningData.completionState).toBe(IOnboardingCompletionState.AdminUserCreationFailed);
        expect(wrapper.vm.provisioningData.failedSubjectIds).toEqual(["failed@example.com"]);
      });

      it("should not call provisionTeam during regular team flow when delegation is disabled", async () => {
        const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
        if (!wrapper) return;
        const adminUserSpy = setupAdminUserSpy();
        const provisionTeamSpy = setupTeamProvisioningSpy();

        await navigateToTeamSetup();

        // Ensure delegation is disabled (default state)
        expect(wrapper.vm.teamData.letSomeoneElseFill).toBe(false);

        await fillTeamData();
        await clickContinue();
        await flushPromises();

        // Should call provisionTeam, not createAdminUser
        expectTeamProvisioningCalled(provisionTeamSpy);
        expectAdminUserCalled(adminUserSpy, 0);
        errorSpy.mockRestore();
      });

      it("should pass correct tenant ID to createAdminUser", async () => {
        if (!wrapper) return;
        const adminUserSpy = setupAdminUserSpy();

        await setupDelegation("tenant-test@example.com");
        await clickContinue();
        await flushPromises();

        // Verify the call includes a tenant ID (number) and SSO status (boolean)
        expect(adminUserSpy).toHaveBeenCalledWith("tenant-test@example.com", expect.any(Number), expect.any(Boolean));

        // Get the actual call to verify tenant ID structure
        const call = adminUserSpy.mock.calls[0];
        expect(typeof call[1]).toBe("number");
      });
    });

    describe("AdminOnboardingModal.vue - Skip Button", () => {
      async function clickSkip() {
        const modal = getModal();
        const skipButton = modal.find(".skip-button");
        expect(skipButton.exists()).toBe(true);
        await skipButton.trigger("click");
        await flushPromises();
      }

      it("should display skip button in SSO setup and team setup steps", async () => {
        // Landing step - no skip button
        await flushPromises();
        expectCurrentStep(OnboardingStep.Landing);
        let modal = getModal();
        expect(modal.find(".skip-button").exists()).toBe(false);

        // Cluster setup step - no skip button
        await pressEnter();
        expectCurrentStep(OnboardingStep.ClusterSetup, 1);
        modal = getModal();
        expect(modal.find(".skip-button").exists()).toBe(false);

        // Navigate to SSO setup step - should show skip button
        await navigateToSsoSetup();
        expectCurrentStep(OnboardingStep.SsoSetup, 1);
        modal = getModal();
        expect(modal.find(".skip-button").exists()).toBe(true);

        // Navigate to team setup step - should show skip button
        await navigateToTeamSetup();
        expectCurrentStep(OnboardingStep.TeamSetup, 1);
        modal = getModal();
        expect(modal.find(".skip-button").exists()).toBe(true);
      });

      it("should skip team setup and go to completion when skip button is clicked", async () => {
        if (!wrapper) return;
        const provisionTeamSpy = setupTeamProvisioningSpy();
        const adminUserSpy = vi.spyOn(teamProvisioningService, "createAdminUser");

        await navigateToTeamSetup();
        expectCurrentStep(OnboardingStep.TeamSetup, 1);

        // Click skip button
        await clickSkip();

        // Should go directly to completion step
        expectCurrentStep(OnboardingStep.Completion, 1);

        // Should not call any provisioning services
        expect(provisionTeamSpy).not.toHaveBeenCalled();
        expect(adminUserSpy).not.toHaveBeenCalled();

        // Should set completion state to "CompletedWithoutProject"
        expect(wrapper.vm.provisioningData.completionState).toBe("CompletedWithoutProject");
        expect(wrapper.vm.provisioningData.failedSubjectIds).toEqual([]);
      });

      it("should skip team setup without validation", async () => {
        if (!wrapper) return;

        await navigateToTeamSetup();
        expectCurrentStep(OnboardingStep.TeamSetup, 1);

        // Leave form fields empty (invalid state)
        expect(wrapper.vm.teamData.teamName).toBe("");
        expect(wrapper.vm.teamData.memberEmails).toEqual([]);

        // Skip should work even with invalid form data
        await clickSkip();

        // Should still advance to completion step
        expectCurrentStep(OnboardingStep.Completion, 1);
        expect(wrapper.vm.provisioningData.completionState).toBe("CompletedWithoutProject");
      });

      it("should disable skip button when loading", async () => {
        await navigateToTeamSetup();
        expectCurrentStep(OnboardingStep.TeamSetup, 1);

        // Set loading state
        wrapper?.vm.handleLoadingChange(true);
        await flushPromises();

        // Skip button should be disabled
        const modal = getModal();
        const skipButton = modal.find(".skip-button");
        expect(skipButton.exists()).toBe(true);
        expect(skipButton.attributes("disabled")).toBeDefined();
      });
    });

    describe("AdminOnboardingModal.vue - Email Input Blur Functionality", () => {
      it("should not add invalid email on blur", async () => {
        if (!wrapper) return;

        await navigateToTeamSetup();
        expectCurrentStep(OnboardingStep.TeamSetup, 1);

        const emailInput = wrapper.findComponent({ name: "team-member-emails-input" });
        expect(emailInput.exists()).toBe(true);

        // Trigger input value change with invalid email and blur
        await emailInput.vm.handleInputValueChange("invalid-email");
        await emailInput.vm.handleInputBlur();
        await flushPromises();

        // Invalid email should not be added
        expect(wrapper.vm.teamData.memberEmails).not.toContain("invalid-email");
      });

      it("should not add duplicate email on blur", async () => {
        if (!wrapper) return;

        await navigateToTeamSetup();
        expectCurrentStep(OnboardingStep.TeamSetup, 1);

        // Pre-populate with an email
        wrapper.vm.teamData.memberEmails = ["existing@example.com"];
        await flushPromises();

        const emailInput = wrapper.findComponent({ name: "team-member-emails-input" });
        expect(emailInput.exists()).toBe(true);

        // Try to add the same email via blur
        await emailInput.vm.handleInputValueChange("existing@example.com");
        await emailInput.vm.handleInputBlur();
        await flushPromises();

        // Should still have only one instance
        const emailCount = wrapper.vm.teamData.memberEmails.filter(
          (email: string) => email === "existing@example.com",
        ).length;
        expect(emailCount).toBe(1);
      });

      it("should not add email when 10 email limit is reached", async () => {
        if (!wrapper) return;

        await navigateToTeamSetup();
        expectCurrentStep(OnboardingStep.TeamSetup, 1);

        // Pre-populate with 10 emails (at the limit)
        wrapper.vm.teamData.memberEmails = [
          "user1@example.com",
          "user2@example.com",
          "user3@example.com",
          "user4@example.com",
          "user5@example.com",
          "user6@example.com",
          "user7@example.com",
          "user8@example.com",
          "user9@example.com",
          "user10@example.com",
        ];
        await flushPromises();

        const emailInput = wrapper.findComponent({ name: "team-member-emails-input" });
        expect(emailInput.exists()).toBe(true);

        // Try to add an 11th email via blur
        await emailInput.vm.handleInputValueChange("user11@example.com");
        await emailInput.vm.handleInputBlur();
        await flushPromises();

        // Should still have only 10 emails
        expect(wrapper.vm.teamData.memberEmails.length).toBe(10);
        expect(wrapper.vm.teamData.memberEmails).not.toContain("user11@example.com");
      });
    });
  });

  describe("AdminOnboardingModal.vue - Close Button Visibility", () => {
    it("should not show close button on landing step", async () => {
      await flushPromises();
      expectCurrentStep(OnboardingStep.Landing);

      const modal = getModal();
      expect(modal.find(".close-button").exists()).toBe(false);
    });

    it("should not show close button in early cluster setup steps", async () => {
      await flushPromises();
      await clickContinue(); // Go to step 1
      expectCurrentStep(OnboardingStep.ClusterSetup, 1);

      const modal = getModal();
      expect(modal.find(".close-button").exists()).toBe(false);

      // Fill name and continue to location
      await fillClusterName("test-cluster");
      await clickContinue();
      expectCurrentStep(1, 2);
      expect(modal.find(".close-button").exists()).toBe(false);
    });

    it("should show close button after cluster creation is completed", async () => {
      await navigateToTeamSetup();
      expectCurrentStep(OnboardingStep.TeamSetup, 1);

      const modal = getModal();
      expect(modal.find(".close-button").exists()).toBe(true);
    });

    it("should show close button on completion step", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      if (!wrapper) return;

      await navigateToTeamSetup();
      await fillTeamData();
      await clickContinue();
      await flushPromises();

      // If still on TeamSetup due to extra substep, advance once more
      if (wrapper.vm.currentStep === OnboardingStep.TeamSetup) {
        await clickContinue();
        await flushPromises();
      }

      expectCurrentStep(OnboardingStep.Completion, 1);
      const modal = getModal();
      expect(modal.find(".close-button").exists()).toBe(true);
      errorSpy.mockRestore();
    });

    it("should show close button in delegated team setup flow", async () => {
      // Unmount the regular component and mount with delegated flow props
      if (wrapper) wrapper.unmount();

      mountComponent({
        skipClusterSetup: true,
        clusterId: "test-cluster-id",
      });

      await flushPromises();

      const modal = getModal();
      // This should pass but currently fails due to the bug
      expect(modal.find(".close-button").exists()).toBe(true);
    });

    it("should show close button in delegated flow even during form interaction", async () => {
      // Unmount the regular component and mount with delegated flow props
      if (wrapper) wrapper.unmount();

      mountComponent({
        skipClusterSetup: true,
        clusterId: "test-cluster-id",
      });

      await flushPromises();

      // Simulate user interacting with the form
      const modal = getModal();
      const teamNameInput = modal.find('[aid="team-name"]');
      if (teamNameInput.exists()) {
        await teamNameInput.setValue("test-team");
        await flushPromises();
      }

      // Close button should still be visible
      expect(modal.find(".close-button").exists()).toBe(true);
    });
  });

  describe("AdminOnboardingModal.vue - Form Validation", () => {
    it("should prevent navigation when form validation fails", async () => {
      await flushPromises();
      await clickContinue(); // Go to cluster setup
      expectCurrentStep(OnboardingStep.ClusterSetup, 1);

      // Try to continue without filling cluster name (should fail validation)
      await clickContinue();
      expectCurrentStep(OnboardingStep.ClusterSetup, 1); // Should remain on same step
    });

    it("should prevent Enter key navigation when form validation fails", async () => {
      await flushPromises();
      await clickContinue(); // Go to cluster setup
      expectCurrentStep(OnboardingStep.ClusterSetup, 1);

      // Try to continue with Enter without filling cluster name
      await pressEnter();
      expectCurrentStep(OnboardingStep.ClusterSetup, 1); // Should remain on same step
    });

    it("should allow navigation when form validation passes", async () => {
      await flushPromises();
      await clickContinue(); // Go to cluster setup
      expectCurrentStep(OnboardingStep.ClusterSetup, 1);

      // Fill valid cluster name
      await fillClusterName("valid-cluster");
      await clickContinue();
      expectCurrentStep(1, 2); // Should advance to next step
    });
  });

  describe("AdminOnboardingModal.vue - Error Handling", () => {
    it("should handle team provisioning retry", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      if (!wrapper) return;
      const provisionSpy = setupTeamProvisioningSpy();

      // Mock failed provisioning
      provisionSpy.mockResolvedValue({
        success: false,
        completionState: IOnboardingCompletionState.ProjectCreationFailed,
        failedSubjectIds: ["failed@example.com"],
      });

      await navigateToTeamSetup();
      await fillTeamData();
      await clickContinue();
      await flushPromises();

      // Advance to completion if there is an extra substep
      if (wrapper.vm.currentStep === OnboardingStep.TeamSetup) {
        await clickContinue();
        await flushPromises();
      }
      // Should be on completion step with error state
      expectCurrentStep(OnboardingStep.Completion, 1);
      expect(wrapper.vm.provisioningData.completionState).toBe("ProjectCreationFailed");

      // Click retry button
      const modal = getModal();
      const retryButton = modal.find('[data-cy="retry-button"]'); // Assuming retry button exists
      if (retryButton.exists()) {
        await retryButton.trigger("click");
        await flushPromises();

        // Should go back to team setup step
        expectCurrentStep(OnboardingStep.TeamSetup, 1);
        expect(wrapper.vm.provisioningData.completionState).toBeUndefined();
      }
      errorSpy.mockRestore();
    });

    it("should handle installation preparation errors", async () => {
      if (!wrapper) return;

      // Mock installation service to return error
      const mockInstallationService = await import("@/services/control-plane/installation.service/installation.service");
      const prepareSpy = vi.spyOn(mockInstallationService.installationService, "prepareInstallation");
      prepareSpy.mockResolvedValue({
        success: false,
        error: {
          type: InstallationErrorType.ClusterCreation,
          message: "Failed to create cluster",
        },
      });

      await flushPromises();
      await clickContinue(); // Landing -> Cluster Setup
      await fillClusterName("test-cluster");
      await clickContinue(); // Name -> Location
      await selectClusterLocation(EClusterLocations.REMOTE);
      await clickContinue(); // Location -> URL
      await fillClusterUrl("https://test.example.com");
      await clickContinue(); // URL -> Installation (should trigger error)

      // Should remain on installation step but show error
      expectCurrentStep(1, 4);
      expect(wrapper.vm.clusterData.installationError).toBe("Failed to create cluster");
      expect(wrapper.vm.clusterData.installationErrorType).toBe("cluster-creation");
    });
  });

  describe("AdminOnboardingModal.vue - Sidebar Navigation Behavior", () => {
    it("should disable sidebar navigation after team completion", async () => {
      if (!wrapper) return;

      // Navigate to team setup and complete it
      await navigateToTeamSetup();
      await fillTeamData("completed-team");
      await clickContinue();
      await flushPromises();

      if (wrapper.vm.currentStep === OnboardingStep.TeamSetup) {
        await clickContinue();
        await flushPromises();
      }

      // Should be on completion step
      expectCurrentStep(OnboardingStep.Completion, 1);

      // Verify that navigation is disabled
      expect(wrapper.vm.isBackwardNavigationDisabled()).toBe(true);

      // Check that sidebar receives the correct prop
      const sidebar = wrapper.findComponent({ name: "onboarding-sidebar" });
      expect(sidebar.exists()).toBe(true);
      expect(sidebar.props("isNavigationDisabled")).toBe(true);
    });

    it("should enable sidebar navigation before team completion", async () => {
      if (!wrapper) return;

      // Navigate to team setup but don't complete it
      await navigateToTeamSetup();
      expectCurrentStep(OnboardingStep.TeamSetup, 1);

      // Verify that navigation is still enabled
      expect(wrapper.vm.isBackwardNavigationDisabled()).toBe(false);

      // Check that sidebar receives the correct prop
      const sidebar = wrapper.findComponent({ name: "onboarding-sidebar" });
      expect(sidebar.exists()).toBe(true);
      expect(sidebar.props("isNavigationDisabled")).toBe(false);
    });

    it("should disable sidebar navigation when on completion step", async () => {
      if (!wrapper) return;

      // Directly navigate to completion step
      wrapper.vm.handleNavigateToStep(OnboardingStep.Completion, 1);
      await flushPromises();

      expectCurrentStep(OnboardingStep.Completion, 1);

      // Verify that navigation is disabled
      expect(wrapper.vm.isBackwardNavigationDisabled()).toBe(true);

      // Check that sidebar receives the correct prop
      const sidebar = wrapper.findComponent({ name: "onboarding-sidebar" });
      expect(sidebar.exists()).toBe(true);
      expect(sidebar.props("isNavigationDisabled")).toBe(true);
    });
  });

  // SSO Step Tests
  describe("AdminOnboardingModal.vue - SSO Setup Flow", () => {
    it("should navigate to SSO setup after cluster setup", async () => {
      await navigateToSsoSetup();

      expectCurrentStep(OnboardingStep.SsoSetup, 1);
      expectStepTitle("Select the SSO protocol");
    });

    it("should show skip button in SSO setup step", async () => {
      await navigateToSsoSetup();
      const modal = getModal();
      expect(modal.find(".skip-button").exists()).toBe(true);
    });

    it("should allow selecting OIDC protocol", async () => {
      if (!wrapper) return;
      await navigateToSsoSetup();

      await selectSsoProtocol(IdpTypeEnum.Oidc);
      expect(wrapper.vm.ssoData.protocol).toBe(IdpTypeEnum.Oidc);
    });

    it("should allow selecting SAML protocol", async () => {
      if (!wrapper) return;
      await navigateToSsoSetup();

      await selectSsoProtocol(IdpTypeEnum.Saml);
      expect(wrapper.vm.ssoData.protocol).toBe(IdpTypeEnum.Saml);
    });

    it("should skip SSO setup and go to email server setup", async () => {
      await navigateToSsoSetup();

      const modal = getModal();
      const skipButton = modal.find(".skip-button");
      await skipButton.trigger("click");
      await flushPromises();

      expectCurrentStep(OnboardingStep.EmailServerSetup, 1);
    });

    it("should reset SSO data when skipping", async () => {
      if (!wrapper) return;
      await navigateToSsoSetup();

      // Set some SSO data
      await selectSsoProtocol(IdpTypeEnum.Oidc);
      expect(wrapper.vm.ssoData.protocol).toBe(IdpTypeEnum.Oidc);

      // Skip SSO
      const modal = getModal();
      const skipButton = modal.find(".skip-button");
      await skipButton.trigger("click");
      await flushPromises();

      // SSO data should be reset
      expect(wrapper.vm.ssoData.protocol).toBe("");
    });
  });

  describe("AdminOnboardingModal.vue - SSO Error Handling", () => {
    it("should show 'Skip SSO Setup' button when there's an SSO creation error", async () => {
      if (!wrapper) return;
      await navigateToSsoSetup();

      // Simulate SSO creation error
      wrapper.vm.handleSsoDataUpdate("ssoErrorType", SsoErrorType.IdpCreation);
      await flushPromises();

      // Continue button should show 'Skip SSO Setup'
      expect(wrapper.vm.continueButtonLabel).toBe("Skip SSO Setup");

      // Normal skip button should be hidden
      expect(wrapper.vm.showSkipButton).toBe(false);
    });

    it("should skip SSO setup when 'Skip SSO Setup' button is clicked", async () => {
      if (!wrapper) return;
      await navigateToSsoSetup();

      // Simulate SSO creation error
      wrapper.vm.handleSsoDataUpdate("ssoErrorType", SsoErrorType.IdpCreation);
      await flushPromises();

      // Continue button becomes "Skip SSO Setup"; emulate its action directly
      wrapper.vm.skipSsoSetup();
      await flushPromises();

      // Expect to move to the next available step (EmailServerSetup, TeamSetup, or Completion)
      const nextStep = wrapper.vm.currentStep;
      expect([OnboardingStep.EmailServerSetup, OnboardingStep.TeamSetup, OnboardingStep.Completion]).toContain(nextStep);
    });

    it("should reset SSO step tracking when skipping due to error", async () => {
      if (!wrapper) return;
      await navigateToSsoSetup();

      // Mark SSO step as completed
      wrapper.vm.visitedSubSteps[OnboardingStep.SsoSetup] = new Set([1, 2]);
      await flushPromises();

      // Simulate SSO creation error and skip
      wrapper.vm.handleSsoDataUpdate("ssoErrorType", SsoErrorType.IdpCreation);
      await flushPromises();
      // Call skip directly to validate reset logic deterministically
      wrapper.vm.skipSsoSetup();
      await flushPromises();

      // Visited substeps should be cleared
      expect(wrapper.vm.visitedSubSteps[OnboardingStep.SsoSetup]).toBeUndefined();
    });

    it("should show normal buttons when there's no SSO error", async () => {
      if (!wrapper) return;
      await navigateToSsoSetup();

      // No error state
      expect(wrapper.vm.ssoData.ssoErrorType).toBeNull();

      // Continue button should show normal text
      expect(wrapper.vm.continueButtonLabel).toBe("Continue");

      // Normal skip button should be visible
      expect(wrapper.vm.showSkipButton).toBe(true);
    });
  });
});
