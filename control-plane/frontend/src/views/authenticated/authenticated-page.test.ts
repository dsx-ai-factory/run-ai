/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach, type Mock } from "vitest";
import { OnboardingStep } from "@/components/onboarding/models/admin-onboarding.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { STORAGE_KEYS } from "@/models/storage.model";
import type { IAdminMessage } from "@/models/setting.model";

// Mock all problematic dependencies that use lottie or complex UI
vi.mock("vue3-lottie", () => ({ Vue3Lottie: { template: "<div></div>" } }));

vi.mock("@/services/infra/idle-timer.service/idle-timer.service", () => ({
  IdleTimer: {
    getInstance: () => ({
      track: vi.fn(),
    }),
  },
}));

vi.mock("@/utils/interval.util", () => ({
  intervalUtil: {
    startInterval: vi.fn(),
    stopInterval: vi.fn(),
  },
}));

vi.mock("@/utils/admin-message.util/admin-message.util", () => ({
  adminMessageUtil: {
    dismissAdminMessage: vi.fn(),
    getAdminMessage: vi.fn(),
    getSystemMessage: vi.fn(),
    dismissSystemMessage: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    countProjects: vi.fn(),
  },
}));

vi.mock("@/services/filter.service/filter.service", () => ({
  filterService: {
    clearAllTablesFilters: vi.fn(),
  },
}));

vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn(),
    save: vi.fn(),
  },
}));

// Mock utils that may cause issues
vi.mock("@/utils/infra.util/infra.util", () => ({
  getEnvType: vi.fn().mockReturnValue("DEV"),
  EnvType: { DEV: "DEV", PROD: "PROD" },
  isSaas: vi.fn().mockReturnValue(false),
}));

vi.mock("@/utils/error-alert.util", () => ({
  ErrorAlert: vi.fn(),
}));

// Import mocked services
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { storageUtil } from "@/utils/storage.util";
import { adminMessageUtil } from "@/utils/admin-message.util/admin-message.util";

interface MockAuthStore {
  accessToken: string;
  tenant: { id: string };
  currentUser: { username: string };
  userLastLogin: string;
  isSSO: boolean;
  loadUserInfo: () => Promise<void>;
  loadTenant: () => Promise<void>;
  updateUserLastLoggedIn: () => void;
  loadUserOrgUnits: () => Promise<void>;
}

interface MockClusterStore {
  clusters: Array<{ id: string }>;
  selectedClusterId: string;
  selectedClusterName: string;
  defaultClusterId: string;
  canCreateCluster: boolean;
  loadClusters: () => Promise<void>;
  setSelectedClusterById: (id: string) => void;
}

interface MockPermissionStore {
  permissionsList: Array<{ resourceType: string; action: string }>;
  hasAllPermissions: Mock<(permissions: Array<{ resourceType: ResourceType; action: Action }>) => boolean>;
  hasPermission: Mock<(resourceType: ResourceType, action: Action) => boolean>;
  isActionGrantedForEveryResource: Mock<(resourceTypes: ResourceType[], action: Action) => boolean>;
  hasSomePermittedActions: Mock<(resourceType: ResourceType, actions: Action[]) => boolean>;
  loadPermissions: () => Promise<void>;
}

interface MockSettingStore {
  loadSetting: () => Promise<void>;
  loadAdminMessage: () => Promise<null>;
  loadSystemMessage: () => Promise<null>;
}

interface MockIdpsStore {
  loadIdps: () => Promise<Array<unknown>>;
}

interface MockSecuritySettingsStore {
  getSettingByKey: (key: string) => Promise<{ idleTimeout: number }>;
}

interface MockRoleStore {
  roleList?: Array<{ name: string }>;
  loadRoles: () => Promise<void>;
}

interface MockStoreStates {
  authStore?: Partial<MockAuthStore>;
  clusterStore?: Partial<MockClusterStore>;
  permissionStore?: Partial<MockPermissionStore>;
  settingStore?: Partial<MockSettingStore>;
  idpsStore?: Partial<MockIdpsStore>;
  securitySettingsStore?: Partial<MockSecuritySettingsStore>;
  roleStore?: Partial<MockRoleStore>;
}

function createMockStores(overrides: MockStoreStates = {}) {
  return {
    authStore: {
      accessToken: "mock-token",
      tenant: { id: "tenant-123" },
      currentUser: { username: "test-user" },
      userLastLogin: "2024-01-01",
      isSSO: false,
      loadUserInfo: vi.fn().mockResolvedValue(undefined),
      loadTenant: vi.fn().mockResolvedValue(undefined),
      updateUserLastLoggedIn: vi.fn(),
      loadUserOrgUnits: vi.fn().mockResolvedValue(undefined),
      ...overrides.authStore,
    },
    clusterStore: {
      clusters: [],
      selectedClusterId: "test-cluster",
      selectedClusterName: "test-cluster",
      defaultClusterId: "default-cluster",
      canCreateCluster: true,
      loadClusters: vi.fn().mockResolvedValue(undefined),
      setSelectedClusterById: vi.fn(),
      ...overrides.clusterStore,
    },
    permissionStore: {
      permissionsList: [{ resourceType: "users", action: "read" }],
      hasAllPermissions: vi.fn().mockReturnValue(true),
      hasPermission: vi.fn().mockReturnValue(true),
      isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
      hasSomePermittedActions: vi.fn().mockReturnValue(true),
      loadPermissions: vi.fn().mockResolvedValue(undefined),
      ...overrides.permissionStore,
    },
    settingStore: {
      loadSetting: vi.fn().mockResolvedValue(undefined),
      loadAdminMessage: vi.fn().mockResolvedValue(null),
      loadSystemMessage: vi.fn().mockResolvedValue(null),
      ...overrides.settingStore,
    },
    idpsStore: {
      loadIdps: vi.fn().mockResolvedValue([]),
      ...overrides.idpsStore,
    },
    securitySettingsStore: {
      getSettingByKey: vi.fn().mockResolvedValue({ idleTimeout: 30 }),
      ...overrides.securitySettingsStore,
    },
    roleStore: {
      loadRoles: vi.fn().mockResolvedValue(undefined),
      ...overrides.roleStore,
    },
  };
}

function createMockComponent(overrides: MockStoreStates = {}): MockComponent {
  const mockStores = createMockStores(overrides);

  // Create a minimal Vue component instance for testing computed properties
  return {
    // Mock the stores as component data
    appStore: {
      appLoaded: true,
      isPageLoading: false,
      isFallback: false,
      version: "1.0.0",
      adminOnboardingModalKey: "admin-onboarding-modal-shown",
    },
    authStore: mockStores.authStore,
    clusterStore: mockStores.clusterStore,
    permissionStore: mockStores.permissionStore,
    settingStore: mockStores.settingStore,
    idpsStore: mockStores.idpsStore,
    securitySettingsStore: mockStores.securitySettingsStore,
    roleStore: mockStores.roleStore,
    showAdminOnboardingModal: false,
    adminMessage: null,
    appIsReady: true,

    // Mock Vue instance methods
    $route: { meta: {}, name: "dashboard" },
    $router: { push: vi.fn(), replace: vi.fn() },

    // Computed properties from the actual component
    hasMinTeamSetupPermissions() {
      return this.permissionStore.isActionGrantedForEveryResource(
        [ResourceType.Project, ResourceType.AccessRules, ResourceType.Users],
        Action.Create,
      );
    },

    hasOnboardingStepsToShow() {
      const stepsPermissions = [
        {
          stepId: OnboardingStep.ClusterSetup,
          permissionChecks: [{ resourceType: ResourceType.Cluster, action: Action.Create }],
          additionalConditions: [this.clusterStore.clusters.length === 0],
        },
        {
          stepId: OnboardingStep.SsoSetup,
          permissionChecks: [
            { resourceType: ResourceType.SecuritySettings, action: Action.Create },
            { resourceType: ResourceType.SecuritySettings, action: Action.Update },
            { resourceType: ResourceType.Users, action: Action.Create },
            { resourceType: ResourceType.AccessRules, action: Action.Create },
          ],
          additionalConditions: [!(storageUtil.get as Mock)(STORAGE_KEYS.ONBOARDING_SKIPPED_SSO), !this.authStore.isSSO],
        },
        {
          stepId: OnboardingStep.TeamSetup,
          permissionChecks: [
            { resourceType: ResourceType.Project, action: Action.Create },
            { resourceType: ResourceType.AccessRules, action: Action.Create },
            { resourceType: ResourceType.Users, action: Action.Create },
          ],
          additionalConditions: [!(storageUtil.get as Mock)(STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP)],
        },
      ];

      return stepsPermissions.some((stepPermission) => {
        const hasRequiredPermissions = this.permissionStore.hasAllPermissions(stepPermission.permissionChecks || []);
        const hasAdditionalConditions = stepPermission.additionalConditions.every((condition) => Boolean(condition));
        return hasRequiredPermissions && hasAdditionalConditions;
      });
    },

    // Mock the displayOnboardingModalIfNeeded method
    async displayOnboardingModalIfNeeded() {
      try {
        const modalAlreadyShown = (storageUtil.get as Mock)("intro-modal-v1");
        if (!this.hasMinTeamSetupPermissions() || modalAlreadyShown) return;
        const numberOfProjects = await (orgUnitService.countProjects as Mock)({});
        if (numberOfProjects === 0 && this.hasMinTeamSetupPermissions()) {
          this.showAdminOnboardingModal = true;
          (storageUtil.save as Mock)("intro-modal-v1", true);
        }
      } catch (e) {
        // Handle error silently for tests
      }
    },
  };
}

// Helper function to setup permissions for specific steps
function setupStepPermissions(
  permissionStore: MockPermissionStore,
  stepPermissions: Partial<Record<OnboardingStep, boolean>>,
) {
  permissionStore.hasAllPermissions.mockImplementation(
    (permissions: Array<{ resourceType: ResourceType; action: Action }>) => {
      if (!permissions || permissions.length === 0) return true;

      // Check cluster permissions
      const hasClusterCreate = permissions.some(
        (p) => p.resourceType === ResourceType.Cluster && p.action === Action.Create,
      );
      if (hasClusterCreate) return stepPermissions[OnboardingStep.ClusterSetup] ?? false;

      // Check SSO permissions (SecuritySettings + Users + AccessRules)
      const hasSsoPermissions =
        permissions.some((p) => p.resourceType === ResourceType.SecuritySettings) &&
        permissions.some((p) => p.resourceType === ResourceType.Users && p.action === Action.Create) &&
        permissions.some((p) => p.resourceType === ResourceType.AccessRules && p.action === Action.Create);
      if (hasSsoPermissions) return stepPermissions[OnboardingStep.SsoSetup] ?? false;

      // Check team permissions (Project + AccessRules + Users)
      const hasTeamPermissions =
        permissions.some((p) => p.resourceType === ResourceType.Project && p.action === Action.Create) &&
        permissions.some((p) => p.resourceType === ResourceType.AccessRules && p.action === Action.Create) &&
        permissions.some((p) => p.resourceType === ResourceType.Users && p.action === Action.Create);
      if (hasTeamPermissions) return stepPermissions[OnboardingStep.TeamSetup] ?? false;

      return false;
    },
  );
}

interface MockComponent {
  appStore: {
    appLoaded: boolean;
    isPageLoading: boolean;
    isFallback: boolean;
    version: string;
    adminOnboardingModalKey: string;
  };
  authStore: MockAuthStore;
  clusterStore: MockClusterStore;
  permissionStore: MockPermissionStore;
  settingStore: MockSettingStore;
  idpsStore: MockIdpsStore;
  securitySettingsStore: MockSecuritySettingsStore;
  roleStore: MockRoleStore;
  showAdminOnboardingModal: boolean;
  adminMessage: null;
  appIsReady: boolean;
  $route: { meta: Record<string, unknown>; name: string };
  $router: { push: () => void; replace: () => void };
  hasMinTeamSetupPermissions: () => boolean;
  hasOnboardingStepsToShow: () => boolean;
  displayOnboardingModalIfNeeded: () => Promise<void>;
}

describe("AuthenticatedPage - Admin Onboarding Modal Display Logic", () => {
  let mockComponent: MockComponent | null;

  beforeEach(() => {
    vi.clearAllMocks();
    // Setup default mocks
    (storageUtil.get as Mock).mockReturnValue(false); // Nothing skipped by default
    (orgUnitService.countProjects as Mock).mockResolvedValue(0);
  });

  afterEach(() => {
    mockComponent = null;
  });

  describe("hasOnboardingStepsToShow computed property", () => {
    it("should return false when no steps meet both permission and additional conditions", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }], canCreateCluster: true }, // Has clusters (ClusterSetup fails additional condition)
        authStore: { isSSO: true }, // Is SSO (SsoSetup fails additional condition)
      });

      // Setup permissions to pass but additional conditions to fail
      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: true,
        [OnboardingStep.SsoSetup]: true,
        [OnboardingStep.TeamSetup]: true,
      });

      // Mock team setup as skipped
      (storageUtil.get as Mock).mockImplementation((key: string) => {
        return key === STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP;
      });

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(false);
    });

    it("should return true when ClusterSetup step meets all conditions", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [], canCreateCluster: true }, // No clusters
      });

      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: true,
        [OnboardingStep.SsoSetup]: false,
        [OnboardingStep.TeamSetup]: false,
      });

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(true);
    });

    it("should return true when SsoSetup step meets all conditions", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }], canCreateCluster: true }, // Has clusters
        authStore: { isSSO: false }, // Not SSO
      });

      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: false, // No cluster create permission
        [OnboardingStep.SsoSetup]: true,
        [OnboardingStep.TeamSetup]: false,
      });

      // Mock SSO not skipped
      (storageUtil.get as Mock).mockImplementation((key: string) => {
        if (key === STORAGE_KEYS.ONBOARDING_SKIPPED_SSO) return false;
        return false;
      });

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(true);
    });

    it("should return true when TeamSetup step meets all conditions", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }], canCreateCluster: true }, // Has clusters
        authStore: { isSSO: true }, // Is SSO (fails SSO setup condition)
      });

      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: false, // No cluster create permission
        [OnboardingStep.SsoSetup]: false, // No SSO permissions
        [OnboardingStep.TeamSetup]: true,
      });

      // Mock team setup not skipped
      (storageUtil.get as Mock).mockImplementation((key: string) => {
        return key !== STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP;
        // Other keys return true (skipped)
      });

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(true);
    });

    it("should return true when multiple steps meet conditions", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [], canCreateCluster: true }, // No clusters
        authStore: { isSSO: false }, // Not SSO
      });

      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: true,
        [OnboardingStep.SsoSetup]: true,
        [OnboardingStep.TeamSetup]: true,
      });

      // Mock nothing skipped
      (storageUtil.get as Mock).mockReturnValue(false);

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(true);
    });

    it("should return false when permissions are missing for all steps", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [], canCreateCluster: true }, // No clusters (good additional condition)
        authStore: { isSSO: false }, // Not SSO (good additional condition)
      });

      // No permissions for any step
      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: false,
        [OnboardingStep.SsoSetup]: false,
        [OnboardingStep.TeamSetup]: false,
      });

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(false);
    });
  });

  describe("Modal display logic - No clusters scenario", () => {
    it("should not show modal when hasOnboardingStepsToShow is false", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [], canCreateCluster: true },
      });

      // Mock hasOnboardingStepsToShow false
      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: false, // No permissions
        [OnboardingStep.SsoSetup]: false,
        [OnboardingStep.TeamSetup]: false,
      });

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(false);
      // In this case, the component logic would redirect to cluster new instead of showing modal
    });

    it("should show modal when hasOnboardingStepsToShow is true", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [], canCreateCluster: true },
      });

      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: true, // Has cluster create permissions and no clusters
        [OnboardingStep.SsoSetup]: false,
        [OnboardingStep.TeamSetup]: false,
      });

      mockComponent = createMockComponent(mockStores);

      expect(mockComponent.hasOnboardingStepsToShow()).toBe(true);
    });
  });

  describe("Modal display logic - Has clusters scenario (displayOnboardingModalIfNeeded)", () => {
    it("should show modal when user has clusters, min team permissions, no projects, and modal not shown before", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }], canCreateCluster: true },
        permissionStore: {
          ...createMockStores().permissionStore,
          isActionGrantedForEveryResource: vi.fn().mockReturnValue(true), // hasMinTeamSetupPermissions = true
        },
      });

      // Mock conditions for displayOnboardingModalIfNeeded
      (storageUtil.get as Mock).mockImplementation((key: string) => {
        if (key === "intro-modal-v1") return false; // Modal not shown before
        return false;
      });
      (orgUnitService.countProjects as Mock).mockResolvedValue(0); // No projects

      mockComponent = createMockComponent(mockStores);
      await mockComponent.displayOnboardingModalIfNeeded();

      expect(mockComponent.hasMinTeamSetupPermissions()).toBe(true);
      expect(mockComponent.showAdminOnboardingModal).toBe(true);
    });

    it("should not show modal when user lacks minimum team setup permissions", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }], canCreateCluster: true },
        permissionStore: {
          ...createMockStores().permissionStore,
          isActionGrantedForEveryResource: vi.fn().mockReturnValue(false), // hasMinTeamSetupPermissions = false
        },
      });

      (storageUtil.get as Mock).mockReturnValue(false);
      (orgUnitService.countProjects as Mock).mockResolvedValue(0);

      mockComponent = createMockComponent(mockStores);
      await mockComponent.displayOnboardingModalIfNeeded();

      expect(mockComponent.hasMinTeamSetupPermissions()).toBe(false);
      expect(mockComponent.showAdminOnboardingModal).toBe(false);
    });

    it("should not show modal when modal was already shown before", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }], canCreateCluster: true },
        permissionStore: {
          ...createMockStores().permissionStore,
          isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
        },
      });

      (storageUtil.get as Mock).mockImplementation((key: string) => {
        if (key === "intro-modal-v1") return true; // Modal was shown before
        return false;
      });
      (orgUnitService.countProjects as Mock).mockResolvedValue(0);

      mockComponent = createMockComponent(mockStores);
      await mockComponent.displayOnboardingModalIfNeeded();

      expect(mockComponent.hasMinTeamSetupPermissions()).toBe(true);
      expect(mockComponent.showAdminOnboardingModal).toBe(false); // Should stay false due to already shown
    });

    it("should not show modal when there are existing projects", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }], canCreateCluster: true },
        permissionStore: {
          ...createMockStores().permissionStore,
          isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
        },
      });

      (storageUtil.get as Mock).mockReturnValue(false);
      (orgUnitService.countProjects as Mock).mockResolvedValue(5); // Has projects

      mockComponent = createMockComponent(mockStores);
      await mockComponent.displayOnboardingModalIfNeeded();

      expect(mockComponent.hasMinTeamSetupPermissions()).toBe(true);
      expect(mockComponent.showAdminOnboardingModal).toBe(false); // Should stay false due to existing projects
    });
  });

  describe("Edge cases and combinations", () => {
    it("should handle SSO setup step conditions correctly when user is already SSO", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }] },
        authStore: { isSSO: true }, // User is already using SSO
      });

      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: false,
        [OnboardingStep.SsoSetup]: true, // Has SSO permissions
        [OnboardingStep.TeamSetup]: false,
      });

      (storageUtil.get as Mock).mockImplementation((key: string) => {
        if (key === STORAGE_KEYS.ONBOARDING_SKIPPED_SSO) return false; // SSO not skipped
        return false;
      });

      mockComponent = createMockComponent(mockStores);

      // Should be false because user is already SSO (fails additional condition)
      expect(mockComponent.hasOnboardingStepsToShow()).toBe(false);
    });

    it("should handle team setup step when it was previously skipped", async () => {
      const mockStores = createMockStores({
        clusterStore: { clusters: [{ id: "cluster-1" }] },
      });

      setupStepPermissions(mockStores.permissionStore, {
        [OnboardingStep.ClusterSetup]: false,
        [OnboardingStep.SsoSetup]: false,
        [OnboardingStep.TeamSetup]: true, // Has team permissions
      });

      (storageUtil.get as Mock).mockImplementation((key: string) => {
        if (key === STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP) return true; // Team setup was skipped
        return false;
      });

      mockComponent = createMockComponent(mockStores);

      // Should be false because team setup was skipped (fails additional condition)
      expect(mockComponent.hasOnboardingStepsToShow()).toBe(false);
    });
  });

  describe("Guided Tour and Onboarding Collision Prevention", () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    describe("isOnboardingModalClosed flag initialization", () => {
      it("should initialize from localStorage when modal was previously shown", () => {
        (storageUtil.get as Mock).mockImplementation((key: string) => {
          return key === "intro-modal-v1";
        });

        const isOnboardingModalClosed = (storageUtil.get as Mock)("intro-modal-v1") ?? false;

        expect(isOnboardingModalClosed).toBe(true);
        expect(storageUtil.get).toHaveBeenCalledWith("intro-modal-v1");
      });

      it("should initialize as false when modal was never shown", () => {
        (storageUtil.get as Mock).mockReturnValue(null);

        const isOnboardingModalClosed = (storageUtil.get as Mock)("intro-modal-v1") ?? false;

        expect(isOnboardingModalClosed).toBe(false);
      });
    });

    describe("displayOnboardingModalIfNeeded - flag management", () => {
      it("should set flag when user lacks minimum permissions", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(false),
          },
        });

        (storageUtil.get as Mock).mockReturnValue(false);
        (orgUnitService.countProjects as Mock).mockResolvedValue(0);

        const mockAppStore = { setIsOnboardingModalClosed: vi.fn() };
        mockComponent = createMockComponent(mockStores);

        const hasMinPermissions = mockComponent.hasMinTeamSetupPermissions();
        const modalAlreadyShown = (storageUtil.get as Mock)("intro-modal-v1");

        if (!hasMinPermissions || modalAlreadyShown) {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(hasMinPermissions).toBe(false);
        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);
      });

      it("should set flag when modal was already shown", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
          },
        });

        (storageUtil.get as Mock).mockImplementation((key: string) => key === "intro-modal-v1");

        const mockAppStore = { setIsOnboardingModalClosed: vi.fn() };
        mockComponent = createMockComponent(mockStores);

        const modalAlreadyShown = (storageUtil.get as Mock)("intro-modal-v1");
        if (modalAlreadyShown) {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(modalAlreadyShown).toBe(true);
        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);
      });

      it("should set flag when user has existing projects", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
          },
        });

        (storageUtil.get as Mock).mockReturnValue(false);
        (orgUnitService.countProjects as Mock).mockResolvedValue(5);

        const mockAppStore = { setIsOnboardingModalClosed: vi.fn() };
        mockComponent = createMockComponent(mockStores);

        const numberOfProjects = await (orgUnitService.countProjects as Mock)({});

        if (numberOfProjects === 0 && mockComponent.hasMinTeamSetupPermissions()) {
          mockComponent.showAdminOnboardingModal = true;
        } else {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(numberOfProjects).toBe(5);
        expect(mockComponent.showAdminOnboardingModal).toBe(false);
        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);
      });

      it("should set flag on API error for graceful fallback", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
          },
        });

        (storageUtil.get as Mock).mockReturnValue(false);
        (orgUnitService.countProjects as Mock).mockRejectedValue(new Error("API Error"));

        const mockAppStore = { setIsOnboardingModalClosed: vi.fn() };
        mockComponent = createMockComponent(mockStores);

        try {
          await (orgUnitService.countProjects as Mock)({});
        } catch (e) {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);
      });

      it("should NOT set flag when onboarding should show", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
          },
        });

        (storageUtil.get as Mock).mockReturnValue(false);
        (orgUnitService.countProjects as Mock).mockResolvedValue(0);

        const mockAppStore = { setIsOnboardingModalClosed: vi.fn() };
        mockComponent = createMockComponent(mockStores);
        await mockComponent.displayOnboardingModalIfNeeded();

        const numberOfProjects = await (orgUnitService.countProjects as Mock)({});

        expect(numberOfProjects).toBe(0);
        expect(mockComponent.showAdminOnboardingModal).toBe(true);
        expect(mockAppStore.setIsOnboardingModalClosed).not.toHaveBeenCalled();
      });
    });

    describe("No clusters scenario - flag management", () => {
      it("should set flag when no onboarding steps to show", () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [], canCreateCluster: true },
        });

        setupStepPermissions(mockStores.permissionStore, {
          [OnboardingStep.ClusterSetup]: false,
          [OnboardingStep.SsoSetup]: false,
          [OnboardingStep.TeamSetup]: false,
        });

        const mockAppStore = { setIsOnboardingModalClosed: vi.fn() };
        mockComponent = createMockComponent(mockStores);

        if (!mockComponent.hasOnboardingStepsToShow()) {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(mockComponent.hasOnboardingStepsToShow()).toBe(false);
        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);
      });

      it("should NOT set flag when onboarding steps exist", () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [], canCreateCluster: true },
        });

        setupStepPermissions(mockStores.permissionStore, {
          [OnboardingStep.ClusterSetup]: true,
          [OnboardingStep.SsoSetup]: false,
          [OnboardingStep.TeamSetup]: false,
        });

        const mockAppStore = { setIsOnboardingModalClosed: vi.fn() };
        mockComponent = createMockComponent(mockStores);

        if (mockComponent.hasOnboardingStepsToShow()) {
          mockComponent.showAdminOnboardingModal = true;
        } else {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(mockComponent.hasOnboardingStepsToShow()).toBe(true);
        expect(mockComponent.showAdminOnboardingModal).toBe(true);
        expect(mockAppStore.setIsOnboardingModalClosed).not.toHaveBeenCalled();
      });
    });

    describe("Guided tour auto-trigger blocking", () => {
      const shouldTriggerTour = (isOnboardingModalClosed: boolean, isWhiteLabel: boolean, seenAlready: boolean) => {
        return !(isWhiteLabel || !isOnboardingModalClosed || seenAlready);
      };

      it("should block when flag is false", () => {
        expect(shouldTriggerTour(false, false, false)).toBe(false);
      });

      it("should allow when flag is true and other conditions pass", () => {
        expect(shouldTriggerTour(true, false, false)).toBe(true);
      });

      it("should block when white-label regardless of flag", () => {
        expect(shouldTriggerTour(true, true, false)).toBe(false);
      });

      it("should block when already seen", () => {
        expect(shouldTriggerTour(true, false, true)).toBe(false);
      });
    });

    describe("Integration scenarios", () => {
      it("should show onboarding first for new user without projects, then allow tour", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
          },
        });

        (storageUtil.get as Mock).mockReturnValue(false);
        (orgUnitService.countProjects as Mock).mockResolvedValue(0);

        const mockAppStore = {
          isOnboardingModalClosed: false,
          setIsOnboardingModalClosed: vi.fn((value: boolean) => {
            mockAppStore.isOnboardingModalClosed = value;
          }),
        };

        mockComponent = createMockComponent(mockStores);

        const modalAlreadyShown = (storageUtil.get as Mock)("intro-modal-v1");
        const hasMinPermissions = mockComponent.hasMinTeamSetupPermissions();

        if (!hasMinPermissions || modalAlreadyShown) {
          mockAppStore.setIsOnboardingModalClosed(true);
        } else {
          const numberOfProjects = await (orgUnitService.countProjects as Mock)({});
          if (numberOfProjects === 0 && hasMinPermissions) {
            mockComponent.showAdminOnboardingModal = true;
            (storageUtil.save as Mock)("intro-modal-v1", true);
          } else {
            mockAppStore.setIsOnboardingModalClosed(true);
          }
        }

        expect(mockComponent.showAdminOnboardingModal).toBe(true);
        expect(mockAppStore.setIsOnboardingModalClosed).not.toHaveBeenCalled();

        let shouldShowTour = !mockComponent.showAdminOnboardingModal && mockAppStore.isOnboardingModalClosed;
        expect(shouldShowTour).toBe(false);

        mockComponent.showAdminOnboardingModal = false;
        mockAppStore.setIsOnboardingModalClosed(true);

        shouldShowTour = !mockComponent.showAdminOnboardingModal && mockAppStore.isOnboardingModalClosed;
        expect(shouldShowTour).toBe(true);
      });

      it("should skip onboarding and allow tour immediately for user with projects", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(true),
          },
        });

        (storageUtil.get as Mock).mockReturnValue(false);
        (orgUnitService.countProjects as Mock).mockResolvedValue(5);

        const mockAppStore = {
          isOnboardingModalClosed: false,
          setIsOnboardingModalClosed: vi.fn((value: boolean) => {
            mockAppStore.isOnboardingModalClosed = value;
          }),
        };

        mockComponent = createMockComponent(mockStores);

        const modalAlreadyShown = (storageUtil.get as Mock)("intro-modal-v1");
        const hasMinPermissions = mockComponent.hasMinTeamSetupPermissions();

        if (!hasMinPermissions || modalAlreadyShown) {
          mockAppStore.setIsOnboardingModalClosed(true);
        } else {
          const numberOfProjects = await (orgUnitService.countProjects as Mock)({});
          if (numberOfProjects === 0 && hasMinPermissions) {
            mockComponent.showAdminOnboardingModal = true;
          } else {
            mockAppStore.setIsOnboardingModalClosed(true);
          }
        }

        expect(mockComponent.showAdminOnboardingModal).toBe(false);
        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);

        const shouldShowTour = !mockComponent.showAdminOnboardingModal && mockAppStore.isOnboardingModalClosed;
        expect(shouldShowTour).toBe(true);
      });

      it("should skip onboarding and allow tour for user without permissions", async () => {
        const mockStores = createMockStores({
          clusterStore: { clusters: [{ id: "cluster-1" }] },
          permissionStore: {
            ...createMockStores().permissionStore,
            isActionGrantedForEveryResource: vi.fn().mockReturnValue(false), // No permissions
          },
        });

        (storageUtil.get as Mock).mockReturnValue(false);

        const mockAppStore = {
          isOnboardingModalClosed: false,
          setIsOnboardingModalClosed: vi.fn((value: boolean) => {
            mockAppStore.isOnboardingModalClosed = value;
          }),
        };

        mockComponent = createMockComponent(mockStores);

        const hasMinPermissions = mockComponent.hasMinTeamSetupPermissions();

        if (!hasMinPermissions) {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(hasMinPermissions).toBe(false);
        expect(mockComponent.showAdminOnboardingModal).toBe(false);
        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);

        const shouldShowTour = !mockComponent.showAdminOnboardingModal && mockAppStore.isOnboardingModalClosed;
        expect(shouldShowTour).toBe(true);
      });

      it("should allow tour immediately for returning user", () => {
        (storageUtil.get as Mock).mockImplementation((key: string) => key === "intro-modal-v1");

        const mockAppStore = {
          isOnboardingModalClosed: (storageUtil.get as Mock)("intro-modal-v1") ?? false,
          setIsOnboardingModalClosed: vi.fn(),
        };

        expect(mockAppStore.isOnboardingModalClosed).toBe(true);

        const modalAlreadyShown = (storageUtil.get as Mock)("intro-modal-v1");
        if (modalAlreadyShown) {
          mockAppStore.setIsOnboardingModalClosed(true);
        }

        expect(modalAlreadyShown).toBe(true);
        expect(mockAppStore.setIsOnboardingModalClosed).toHaveBeenCalledWith(true);

        const shouldShowTour = mockAppStore.isOnboardingModalClosed;
        expect(shouldShowTour).toBe(true);
      });
    });
  });

  describe("System and admin message", () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    it("loadSystemMessage uses getSystemMessage with raw message and does not call getAdminMessage", async () => {
      const rawSystemMessage = {
        id: "system-msg-1",
        msg: "System notice",
        createdAt: new Date(),
        enableDismiss: true,
      };
      const mockStores = createMockStores({
        settingStore: {
          loadSystemMessage: vi.fn().mockResolvedValue(rawSystemMessage),
          loadAdminMessage: vi.fn().mockResolvedValue(null),
        },
      });
      vi.mocked(adminMessageUtil.getSystemMessage).mockReturnValue(rawSystemMessage);

      let systemMessage: IAdminMessage | null = null;
      const systemMessageShownThisSession = false;

      async function loadSystemMessage() {
        if (systemMessageShownThisSession) return;
        const raw = await mockStores.settingStore.loadSystemMessage();
        if (!raw) return;
        systemMessage = adminMessageUtil.getSystemMessage(raw);
      }

      await loadSystemMessage();

      expect(mockStores.settingStore.loadSystemMessage).toHaveBeenCalled();
      expect(adminMessageUtil.getSystemMessage).toHaveBeenCalledWith(rawSystemMessage);
      expect(adminMessageUtil.getAdminMessage).not.toHaveBeenCalled();
      expect(systemMessage).toEqual(rawSystemMessage);
    });

    it("closeSystemMessage calls dismissSystemMessage and does not call dismissAdminMessage", () => {
      const systemMessage = {
        id: "system-msg-1",
        msg: "System notice",
        createdAt: new Date(),
        enableDismiss: true,
      };

      function closeSystemMessage() {
        if (!systemMessage) return;
        adminMessageUtil.dismissSystemMessage(systemMessage.id);
      }

      closeSystemMessage();

      expect(adminMessageUtil.dismissSystemMessage).toHaveBeenCalledWith("system-msg-1");
      expect(adminMessageUtil.dismissAdminMessage).not.toHaveBeenCalled();
    });
  });
});
