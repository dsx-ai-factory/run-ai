import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar, type QField } from "quasar";
import { describe, expect, test, vi, beforeEach, afterEach } from "vitest";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";

import { RunaiScopeSection } from ".";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { RunaiOrgTreeWithPopOver } from "@/components/common/runai-org-tree/with-popover";
import { NotificationAlertBar } from "@/components/common/notification-alert-bar";
import { runaiScopeSectionMock } from "./runai-scope-section.mock";

const {
  MOCK_IDS,
  MOCK_NAMES,
  mockTreeData,
  mockOrgUnits,
  mockTenant,
  mockClusters,
  mockScopeModels,
  mockAllowedScopes,
  mockExpandedPermittedScopes,
  mockScopesSupportedVersions,
  mockForcedDisabledScopes,
  mockNodePathMappings,
  mockDefaultFactoryProps,
  mockInformationText,
} = runaiScopeSectionMock;

const SELECTORS = {
  SCOPE_SECTION: "[aid=runai-scope-section]" as const,
  SCOPE_FIELD: "[aid=scope-field]" as const,
  SCOPE_TREE_BTN: "[aid=scope-tree-btn]" as const,
} as const;

vi.mock("@/services/control-plane/rbac/org-tree.service/org-tree.service.ts");

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const mockGetExpandedPermittedScopes = vi.spyOn(orgTreeService, "getExpandedPermittedScopes");
const mockSortOrgTree = vi.spyOn(orgTreeService, "sortOrgTree");
const mockFilterHiddenScopes = vi.spyOn(orgTreeService, "filterHiddenScopes");
const mockFindMatchingNodePaths = vi.spyOn(orgTreeService, "findMatchingNodePaths");

describe("runai-scope-section.vue", () => {
  let runaiScopeSection: VueWrapper<InstanceType<typeof RunaiScopeSection>>;
  const authStore = useAuthStore(pinia);
  const clusterStore = useClusterStore(pinia);

  const mockGetOrgTree = vi.spyOn(orgTreeService, "getOrgTree");
  const mockGetNodePathById = vi.spyOn(orgTreeService, "getNodePathById");
  const mockGetParentClusterUuid = vi.spyOn(orgTreeService, "getParentClusterUuid");

  function factory(props = {}): VueWrapper<InstanceType<typeof RunaiScopeSection>> {
    const defaultProps = {
      ...mockDefaultFactoryProps,
      ...props,
    };

    const wrapper = mount(RunaiScopeSection, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: { ...defaultProps, allowedScopes: defaultProps.allowedScopes as unknown as PermittedScopes },
    });

    wrapper.vm.$q.notify = vi.fn();
    return wrapper;
  }

  beforeEach(() => {
    authStore.$patch({
      tenant: mockTenant,
      orgUnits: mockOrgUnits,
    });

    clusterStore.$patch({
      clusters: mockClusters,
    });

    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);

    mockGetOrgTree.mockImplementation(() => {
      return mockTreeData[0];
    });

    mockGetNodePathById.mockImplementation((tree, nodeId) => {
      return mockNodePathMappings[nodeId?.id as keyof typeof mockNodePathMappings] || null;
    });

    mockGetParentClusterUuid.mockReturnValue(MOCK_IDS.CLUSTER_ID);

    mockGetExpandedPermittedScopes.mockReturnValue(mockExpandedPermittedScopes);

    mockSortOrgTree.mockImplementation((nodes) => nodes);
    mockFilterHiddenScopes.mockImplementation((nodes) => nodes);
    mockFindMatchingNodePaths.mockReturnValue([]);

    runaiScopeSection = factory();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
    // Clear all orgTreeService mocks
    mockGetOrgTree.mockClear();
    mockGetNodePathById.mockClear();
    mockGetParentClusterUuid.mockClear();
    mockGetExpandedPermittedScopes.mockClear();
    mockSortOrgTree.mockClear();
    mockFilterHiddenScopes.mockClear();
    mockFindMatchingNodePaths.mockClear();
  });

  describe("Component Initialization", () => {
    test("Should load component properly", () => {
      expect(runaiScopeSection.html()).toContain("runai-scope-section");
      expect(runaiScopeSection.find(SELECTORS.SCOPE_SECTION).exists()).toBeTruthy();
    });

    test("Should initialize with default props", () => {
      expect(runaiScopeSection.html()).toContain("runai-scope-section");
      expect(runaiScopeSection.find(SELECTORS.SCOPE_SECTION).exists()).toBeTruthy();
    });
  });

  describe("Scope Selection & Validation", () => {
    test("Should validate scope field correctly", async () => {
      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.empty,
      });

      const scopeField = runaiScopeSection.findComponent<QField>(SELECTORS.SCOPE_FIELD);
      expect(scopeField.exists()).toBeTruthy();

      await (scopeField.vm.validate as () => Promise<boolean>)();
      expect(scopeField.vm.hasError).toBeTruthy();

      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.project,
      });

      await (scopeField.vm.validate as () => Promise<boolean>)();
      expect(scopeField.vm.hasError).toBeFalsy();
    });

    test("Should display correct scope text for project", async () => {
      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.project,
      });

      const scopeField = runaiScopeSection.findComponent<QField>(SELECTORS.SCOPE_FIELD);
      expect(scopeField.exists()).toBeTruthy();
      expect(scopeField.vm.modelValue).toContain(MOCK_NAMES.PROJECT_ONE_NAME);
    });

    test("Should display department scope", async () => {
      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.department,
      });

      const scopeField = runaiScopeSection.findComponent<QField>(SELECTORS.SCOPE_FIELD);
      expect(scopeField.vm.modelValue).toContain(MOCK_NAMES.DEPARTMENT_NAME);
    });

    test("Should display cluster scope", async () => {
      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.cluster,
      });

      const scopeField = runaiScopeSection.findComponent<QField>(SELECTORS.SCOPE_FIELD);
      expect(scopeField.vm.modelValue).toContain(MOCK_NAMES.CLUSTER_NAME);
    });

    test("Should return correct scope computed value", async () => {
      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.project,
      });

      expect(runaiScopeSection.vm.scope).toContain(MOCK_NAMES.PROJECT_ONE_NAME);
    });

    test("Should compute summary correctly based on scope", async () => {
      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.empty,
      });

      expect(runaiScopeSection.vm.summary).toBe("None");

      await runaiScopeSection.setProps({
        scopeModel: mockScopeModels.project,
      });
      await flushPromises();

      expect(runaiScopeSection.vm.summary).toContain(MOCK_NAMES.PROJECT_ONE_NAME);
    });
  });

  describe("Modal Interactions", () => {
    test("Should open modal on field focus", async () => {
      const scopeField = runaiScopeSection.findComponent<QField>(SELECTORS.SCOPE_FIELD);
      expect(scopeField.exists()).toBeTruthy();
      expect(runaiScopeSection.vm.isOrgTreeModalOpen).toBeFalsy();

      await scopeField.vm.$emit("focus");
      await flushPromises();

      expect(runaiScopeSection.vm.isOrgTreeModalOpen).toBeTruthy();
    });

    test("Should open modal on button click", async () => {
      const scopeTreeBtn = runaiScopeSection.find(SELECTORS.SCOPE_TREE_BTN);
      expect(scopeTreeBtn.exists()).toBeTruthy();
      expect(runaiScopeSection.vm.isOrgTreeModalOpen).toBeFalsy();

      await scopeTreeBtn.trigger("click");
      await flushPromises();

      expect(runaiScopeSection.vm.isOrgTreeModalOpen).toBeTruthy();
    });

    test("Should render modal when state is true", async () => {
      runaiScopeSection.vm.isOrgTreeModalOpen = true;
      await flushPromises();

      expect(runaiScopeSection.vm.isOrgTreeModalOpen).toBeTruthy();
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.exists()).toBeTruthy();
    });

    test("Should hide modal when state is false", async () => {
      runaiScopeSection.vm.isOrgTreeModalOpen = false;

      expect(runaiScopeSection.vm.isOrgTreeModalOpen).toBeFalsy();
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.exists()).toBeFalsy();
    });
  });

  describe("Props Configuration", () => {
    test("Should control interactivity with readOnly prop", async () => {
      await runaiScopeSection.setProps({ readOnly: false });

      expect(runaiScopeSection.find(SELECTORS.SCOPE_FIELD).exists()).toBeTruthy();
      expect(runaiScopeSection.find(SELECTORS.SCOPE_TREE_BTN).exists()).toBeTruthy();
    });

    test("Should display asset description with kindOfAsset prop", async () => {
      await runaiScopeSection.setProps({ kindOfAsset: "environment" });
      await flushPromises();

      expect(runaiScopeSection.html()).toContain("Set the scope for this environment");

      await runaiScopeSection.setProps({ kindOfAsset: "workload" });

      expect(runaiScopeSection.html()).toContain("Set the scope for this workload");
    });

    test("Should affect component state with scopeModel prop", async () => {
      const testScopeModel = mockScopeModels.project;

      await runaiScopeSection.setProps({ scopeModel: testScopeModel });

      expect(runaiScopeSection.vm.scopeModel).toEqual(testScopeModel);
      expect(runaiScopeSection.vm.selectedScope?.type).toBe(ScopeType.Project);
      expect(runaiScopeSection.vm.selectedScope?.id).toBe(MOCK_IDS.PROJECT_ONE_ID);
    });

    test("Should restrict scopes with allowedScopes prop", async () => {
      const testAllowedScopes = mockAllowedScopes.limited;

      await runaiScopeSection.setProps({ allowedScopes: testAllowedScopes as unknown as PermittedScopes });

      expect(runaiScopeSection.vm.allowedScopes).toEqual(testAllowedScopes);
    });

    test("Should affect scope availability with scopesSupportedVersions prop", async () => {
      await runaiScopeSection.setProps({ scopesSupportedVersions: mockScopesSupportedVersions });

      expect(runaiScopeSection.vm.scopesSupportedVersions).toEqual(mockScopesSupportedVersions);
    });

    test("Should disable scopes with forcedDisabledScopes prop", async () => {
      await runaiScopeSection.setProps({ forcedDisabledScopes: mockForcedDisabledScopes });

      expect(runaiScopeSection.vm.forcedDisabledScopes).toEqual(mockForcedDisabledScopes);
    });

    test("Should set cluster context with selectedClusterId prop", async () => {
      const testClusterId = MOCK_IDS.CLUSTER_ID;

      await runaiScopeSection.setProps({ selectedClusterId: testClusterId });

      expect(runaiScopeSection.vm.selectedClusterId).toBe(testClusterId);
    });

    test("Should display information bar with informationText prop", async () => {
      await runaiScopeSection.setProps({ informationText: undefined });
      expect(runaiScopeSection.findComponent(NotificationAlertBar).exists()).toBeFalsy();

      await runaiScopeSection.setProps({ informationText: mockInformationText });

      expect(runaiScopeSection.html()).toContain(mockInformationText);
      expect(runaiScopeSection.findComponent(NotificationAlertBar).exists()).toBeTruthy();
    });
  });

  describe("Child Component Integration", () => {
    beforeEach(async () => {
      runaiScopeSection.vm.isOrgTreeModalOpen = true;
      await flushPromises();
    });

    test("Should pass allowedScopes prop correctly", () => {
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.props("allowedScopes")).toEqual(runaiScopeSection.props("allowedScopes"));
    });

    test("Should pass scopesSupportedVersions prop correctly", () => {
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.props("scopesSupportedVersions")).toEqual(
        runaiScopeSection.props("scopesSupportedVersions"),
      );
    });

    test("Should pass forcedDisabledScopes prop correctly", () => {
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.props("forcedDisabledScopes")).toEqual(runaiScopeSection.props("forcedDisabledScopes"));
    });

    test("Should pass selectedClusterId prop correctly", () => {
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.props("selectedClusterId")).toEqual(runaiScopeSection.props("selectedClusterId"));
    });

    test("Should pass disableTenant computed correctly", () => {
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.props("disableTenant")).toEqual(true);
    });

    test("Should pass disableCluster computed correctly", () => {
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      expect(orgTreePopover.props("disableCluster")).toEqual(false);
    });

    test("Should derive selected prop correctly", () => {
      const orgTreePopover = runaiScopeSection.findComponent(RunaiOrgTreeWithPopOver);
      const expectedSelected = runaiScopeSection.vm.selectedScope ? [runaiScopeSection.vm.selectedScope] : [];
      expect(orgTreePopover.props("selected")).toEqual(expectedSelected);
    });
  });
});
