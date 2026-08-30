import { describe, vi, it, expect, afterEach, beforeEach } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import ProjectEdit from "@/views/projects/project-edit.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import type { Department, Project, ResourcesNullable } from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";
import { createTestingPinia } from "@pinia/testing";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { useAppStore } from "@/stores/app.store";
import { ESaveOrgUnitAction } from "@/models/org-unit-forms.model";
import { requestToLeave } from "@/services/infra/router.service/router.service";
// ============================================================================
// Mock Data
// ============================================================================
import {
  mockClusters,
  CLUSTER_ID,
  PROJECT_ID,
  DEPARTMENT_ID,
  mockDepartment,
  mockDepartmentWithLimitedCpu,
  mockDepartmentWithUnlimitedCpu,
  mockDepartmentWithNodeTypes,
  mockProject,
  mockCpuEnabledNodePool,
  mockCpuDisabledNodePool,
  mockDepartmentWithMultipleDefaultPools,
  mockProjectInheritingPools,
  mockProjectWithExplicitDefaultPool,
} from "./project-edit.test.mocks";

// ============================================================================
// Mock Setup
// ============================================================================

vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/cluster.service/cluster.service");
vi.mock("@/services/infra/router.service/router.service");
vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    getAccessRulesCount: vi.fn().mockResolvedValue(0),
  },
}));

const mockRouter = {
  push: vi.fn(),
  back: vi.fn(),
};

const createMockRoute = (overrides = {}) => ({
  name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
  params: { id: "" },
  query: { clusterId: "test-cluster-id" },
  meta: {
    backPageName: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
  },
  ...overrides,
});

let mockRoute = createMockRoute();

vi.mock("vue-router", () => ({
  useRoute: () => mockRoute,
  useRouter: () => mockRouter,
}));

// ============================================================================
// Test Suite
// ============================================================================

describe("ProjectEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof ProjectEdit>>;
  let notifyMock: ReturnType<typeof vi.fn>;

  const mountComponent = async (routeOverrides = {}) => {
    mockRoute = createMockRoute(routeOverrides);

    const pinia = createTestingPinia({ createSpy: vi.fn });
    const clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ clusters: mockClusters });
    clusterStore.$patch({ selectedCluster: mockClusters[0] });

    wrapper = mount(ProjectEdit, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: mockRoute,
          $router: mockRouter,
        },
      },
    });

    // Mock notify
    notifyMock = vi.fn();
    if (wrapper.vm.$q) {
      wrapper.vm.$q.notify = notifyMock;
    }

    await flushPromises();
  };

  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock implementations
    vi.mocked(orgUnitService.getDefaultDepartment).mockResolvedValue(mockDepartment as Department);
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment as Department);
    vi.mocked(orgUnitService.getProject).mockResolvedValue(mockProject as Project);
    vi.mocked(orgUnitService.createProject).mockResolvedValue(mockProject as Project);
    vi.mocked(orgUnitService.updateProject).mockResolvedValue(mockProject as Project);
    vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([
      mockCpuEnabledNodePool,
      mockCpuDisabledNodePool,
    ] as NodepoolMinimal[]);
    vi.mocked(requestToLeave).mockResolvedValue(true);
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
  });

  // ==========================================================================
  // Initial Loading Tests
  // ==========================================================================

  describe("Initial Loading", () => {
    it("loads default department for new project", async () => {
      await mountComponent();

      expect(orgUnitService.getDefaultDepartment).toHaveBeenCalledWith(CLUSTER_ID);
    });

    it("passes department to ProjectEditFormV2", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      expect(form.exists()).toBe(true);
      expect(form.props("department")).toEqual(mockDepartment);
    });

    it("sets isNewProject to true for new project route", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      expect(form.props("isNewProject")).toBe(true);
    });

    it("loads existing project for edit route", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      expect(orgUnitService.getProject).toHaveBeenCalledWith(PROJECT_ID);
    });

    it("sets isNewProject to false for edit route", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      expect(form.props("isNewProject")).toBe(false);
    });

    it("loads department for existing project with parentId", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      expect(orgUnitService.getDepartment).toHaveBeenCalledWith(DEPARTMENT_ID);
    });
  });

  // ==========================================================================
  // Node Pools Loading Tests
  // ==========================================================================

  describe("Node Pools Loading", () => {
    it("loads node pools metadata on mount", async () => {
      await mountComponent();

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalled();
    });

    it("passes all node pools metadata to form", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const nodePoolsMinimal = form.props("nodePoolsMinimal") as NodepoolMinimal[];

      // Should include all node pools (filtering happens in child component)
      expect(nodePoolsMinimal).toHaveLength(2);
      expect(nodePoolsMinimal[0].name).toBe("default");
      expect(nodePoolsMinimal[1].name).toBe("gpu-only");
    });

    it("handles node pools loading error gracefully", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockRejectedValue(new Error("Network error"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const nodePoolsMinimal = form.props("nodePoolsMinimal") as NodepoolMinimal[];

      expect(nodePoolsMinimal).toEqual([]);
      errorSpy.mockRestore();
    });

    it("loads node pools for existing project using project clusterId", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalled();
    });
  });

  // ==========================================================================
  // Scope Update Tests
  // ==========================================================================

  describe("Scope Update", () => {
    it("reloads project and node pools when scope changes", async () => {
      await mountComponent();

      vi.clearAllMocks();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const newClusterId = "new-cluster-id";

      form.vm.$emit("update-scope", { clusterId: newClusterId, departmentId: "new-dept" });
      await flushPromises();

      expect(orgUnitService.getDepartment).toHaveBeenCalledWith("new-dept");
      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalled();
    });

    it("loads default department when no departmentId provided", async () => {
      await mountComponent();

      vi.clearAllMocks();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const newClusterId = "new-cluster-id";

      form.vm.$emit("update-scope", { clusterId: newClusterId });
      await flushPromises();

      expect(orgUnitService.getDefaultDepartment).toHaveBeenCalledWith(newClusterId);
    });
  });

  // ==========================================================================
  // Cancel Tests
  // ==========================================================================

  describe("Cancel", () => {
    it("requests leave confirmation on cancel", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      form.vm.$emit("on-cancel");
      await flushPromises();

      expect(requestToLeave).toHaveBeenCalled();
    });

    it("navigates back when leave is allowed", async () => {
      vi.mocked(requestToLeave).mockResolvedValue(true);

      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      form.vm.$emit("on-cancel");
      await flushPromises();

      expect(mockRouter.back).toHaveBeenCalled();
    });

    it("does not navigate when leave is denied", async () => {
      vi.mocked(requestToLeave).mockResolvedValue(false);

      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      form.vm.$emit("on-cancel");
      await flushPromises();

      expect(mockRouter.back).not.toHaveBeenCalled();
    });
  });

  // ==========================================================================
  // Save Tests - Create
  // ==========================================================================

  describe("Save - Create Project", () => {
    it("creates project and shows success notification", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        project: { name: "New Project", clusterId: CLUSTER_ID, parentId: DEPARTMENT_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: onCompleteMock,
      });
      await flushPromises();

      expect(orgUnitService.createProject).toHaveBeenCalled();
      expect(notifyMock).toHaveBeenCalled();
      expect(onCompleteMock).toHaveBeenCalled();
    });

    it("navigates to project index after create with Close action", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });

      form.vm.$emit("on-save", {
        project: { name: "New Project", clusterId: CLUSTER_ID, parentId: DEPARTMENT_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(mockRouter.push).toHaveBeenCalledWith({
        name: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
        query: { createdEntityId: PROJECT_ID },
      });
    });

    it("does not navigate after create with Continue action", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });

      form.vm.$emit("on-save", {
        project: { name: "New Project", clusterId: CLUSTER_ID, parentId: DEPARTMENT_ID },
        action: ESaveOrgUnitAction.Continue,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(mockRouter.push).not.toHaveBeenCalled();
    });

    it("shows error notification on create failure", async () => {
      vi.mocked(orgUnitService.createProject).mockRejectedValue(new Error("Create failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });

      form.vm.$emit("on-save", {
        project: { name: "New Project", clusterId: CLUSTER_ID, parentId: DEPARTMENT_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(notifyMock).toHaveBeenCalled();
      expect(mockRouter.push).not.toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("does NOT call onComplete callback when create fails", async () => {
      vi.mocked(orgUnitService.createProject).mockRejectedValue(new Error("Create failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        project: { name: "New Project", clusterId: CLUSTER_ID, parentId: DEPARTMENT_ID },
        action: ESaveOrgUnitAction.Continue,
        onComplete: onCompleteMock,
      });
      await flushPromises();

      expect(onCompleteMock).not.toHaveBeenCalled();
      errorSpy.mockRestore();
    });
  });

  // ==========================================================================
  // Save Tests - Update
  // ==========================================================================

  describe("Save - Update Project", () => {
    it("updates project and shows success notification", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        project: { ...mockProject, name: "Updated Project" },
        action: ESaveOrgUnitAction.Close,
        onComplete: onCompleteMock,
      });
      await flushPromises();

      expect(orgUnitService.updateProject).toHaveBeenCalled();
      expect(notifyMock).toHaveBeenCalled();
      expect(onCompleteMock).toHaveBeenCalled();
    });

    it("navigates after update with Close action", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });

      form.vm.$emit("on-save", {
        project: { ...mockProject, name: "Updated Project" },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(mockRouter.push).toHaveBeenCalled();
    });

    it("shows error notification on update failure", async () => {
      vi.mocked(orgUnitService.updateProject).mockRejectedValue(new Error("Update failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });

      form.vm.$emit("on-save", {
        project: { ...mockProject, name: "Updated Project" },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(notifyMock).toHaveBeenCalled();
      expect(mockRouter.push).not.toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("does NOT call onComplete callback when update fails", async () => {
      vi.mocked(orgUnitService.updateProject).mockRejectedValue(new Error("Update failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        project: { ...mockProject, name: "Updated Project" },
        action: ESaveOrgUnitAction.Continue,
        onComplete: onCompleteMock,
      });
      await flushPromises();

      expect(onCompleteMock).not.toHaveBeenCalled();
      errorSpy.mockRestore();
    });
  });

  // ==========================================================================
  // Loading State Tests
  // ==========================================================================

  describe("Loading States", () => {
    it("passes isProjectSubmitting to form during save", async () => {
      // Make createProject hang to test loading state
      let resolveCreate: (value: Project) => void = () => {};
      vi.mocked(orgUnitService.createProject).mockImplementation(
        () =>
          new Promise((resolve) => {
            resolveCreate = resolve;
          }),
      );

      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      expect(form.props("isProjectSubmitting")).toBe(false);

      form.vm.$emit("on-save", {
        project: { name: "New Project", clusterId: CLUSTER_ID, parentId: DEPARTMENT_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(form.props("isProjectSubmitting")).toBe(true);

      // Resolve the promise
      resolveCreate(mockProject as Project);
      await flushPromises();

      expect(form.props("isProjectSubmitting")).toBe(false);
    });
  });

  // ==========================================================================
  // Error Handling Tests
  // ==========================================================================

  describe("Error Handling", () => {
    it("handles department loading error", async () => {
      vi.mocked(orgUnitService.getDefaultDepartment).mockRejectedValue(new Error("Load failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent();

      expect(notifyMock).toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("handles project loading error for edit mode", async () => {
      vi.mocked(orgUnitService.getProject).mockRejectedValue(new Error("Project not found"));
      const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // This should set fallback to true in app store
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      // Component should still exist even with error
      expect(wrapper.exists()).toBe(true);
      logSpy.mockRestore();
    });
  });

  // ==========================================================================
  // Props Passing Tests
  // ==========================================================================

  describe("Props Passing", () => {
    it("passes all required props to ProjectEditFormV2", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });

      expect(form.props()).toMatchObject({
        isNewProject: true,
        isProjectSubmitting: false,
      });
      expect(form.props("project")).toBeDefined();
      expect(form.props("department")).toBeDefined();
      expect(form.props("nodePoolsMinimal")).toBeDefined();
    });
  });

  // ==========================================================================
  // CPU Defaults Adjustment at Save Time
  // ==========================================================================

  describe("CPU Defaults Adjustment by Department (prepareProjectForSave)", () => {
    const mountWithCpuEnabled = async (departmentMock: Record<string, unknown>, routeOverrides = {}) => {
      mockRoute = createMockRoute(routeOverrides);

      const pinia = createTestingPinia({ createSpy: vi.fn });
      const clusterStore = useClusterStore(pinia);
      clusterStore.$patch({ clusters: mockClusters });
      clusterStore.$patch({ selectedCluster: mockClusters[0] });

      const settingStore = useSettingStore(pinia);
      settingStore.cpuResourcesQuota = { value: true };

      vi.mocked(orgUnitService.getDefaultDepartment).mockResolvedValue(departmentMock as unknown as Department);
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(departmentMock as unknown as Department);

      wrapper = mount(ProjectEdit, {
        global: {
          plugins: [Quasar, pinia],
          mocks: {
            $route: mockRoute,
            $router: mockRouter,
          },
        },
      });

      notifyMock = vi.fn();
      if (wrapper.vm.$q) {
        wrapper.vm.$q.notify = notifyMock;
      }

      await flushPromises();
    };

    const getResourceFromCreateCall = (nodePoolName: string) => {
      const createCall = vi.mocked(orgUnitService.createProject).mock.calls[0][0];
      return createCall.resources.find((r: ResourcesNullable) => r.nodePool?.name === nodePoolName);
    };

    const triggerSave = async () => {
      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      form.vm.$emit("on-save", {
        project: form.props("project"),
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();
    };

    it("sends cpu.deserved as 0 when department has limited CPU quota", async () => {
      await mountWithCpuEnabled(mockDepartmentWithLimitedCpu);
      await triggerSave();

      const resource = getResourceFromCreateCall("default");
      expect(resource?.cpu?.deserved).toBe(0);
    });

    it("sends memory.deserved as 0 when department has limited memory quota", async () => {
      await mountWithCpuEnabled(mockDepartmentWithLimitedCpu);
      await triggerSave();

      const resource = getResourceFromCreateCall("default");
      expect(resource?.memory?.deserved).toBe(0);
    });

    it("sends cpu.deserved as null when department has unlimited CPU quota", async () => {
      await mountWithCpuEnabled(mockDepartmentWithUnlimitedCpu);
      await triggerSave();

      const resource = getResourceFromCreateCall("default");
      expect(resource?.cpu?.deserved).toBeNull();
    });

    it("sends memory.deserved as null when department has unlimited memory quota", async () => {
      await mountWithCpuEnabled(mockDepartmentWithUnlimitedCpu);
      await triggerSave();

      const resource = getResourceFromCreateCall("default");
      expect(resource?.memory?.deserved).toBeNull();
    });

    it("does not adjust CPU defaults when CPU quota setting is disabled", async () => {
      vi.mocked(orgUnitService.getDefaultDepartment).mockResolvedValue(
        mockDepartmentWithLimitedCpu as unknown as Department,
      );
      await mountComponent();
      await triggerSave();

      const resource = getResourceFromCreateCall("default");
      // CPU should not be present when CPU quota is disabled
      expect(resource?.cpu).toBeNull();
    });
  });

  // ==========================================================================
  // Page Title Tests
  // ==========================================================================

  describe("Page Title", () => {
    it("sets pageTitleOverride with entity name for edit route", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const appStore = useAppStore(wrapper.vm.$pinia);
      expect(appStore.setPageTitleOverride).toHaveBeenCalledWith(`Edit project: ${mockProject.name}`);
    });

    it("does not set pageTitleOverride for new project route", async () => {
      await mountComponent();

      const appStore = useAppStore(wrapper.vm.$pinia);
      expect(appStore.setPageTitleOverride).not.toHaveBeenCalledWith(expect.stringContaining("Edit project:"));
    });

    it("clears pageTitleOverride on unmount", async () => {
      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const appStore = useAppStore(wrapper.vm.$pinia);
      wrapper.unmount();

      expect(appStore.setPageTitleOverride).toHaveBeenLastCalledWith(null);
    });
  });

  // ==========================================================================
  // Default Node Pools — dirty-tracking at Save Time (RUN-41143)
  // ==========================================================================

  describe("Default Node Pools (prepareProjectForSave)", () => {
    const editRoute = {
      name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
      params: { id: PROJECT_ID },
    };

    const getUpdatePayload = () => vi.mocked(orgUnitService.updateProject).mock.calls[0][1];

    const triggerSaveWithPools = async (pools: string[] | null) => {
      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      const baseProject = form.props("project") as Project;
      form.vm.$emit("on-save", {
        project: { ...baseProject, defaultNodePools: pools },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();
    };

    const triggerSaveUnchanged = async () => {
      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      form.vm.$emit("on-save", {
        project: form.props("project"),
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();
    };

    describe("existing project that inherits from department (API defaultNodePools: null)", () => {
      beforeEach(async () => {
        vi.mocked(orgUnitService.getProject).mockResolvedValue(mockProjectInheritingPools as unknown as Project);
        vi.mocked(orgUnitService.getDepartment).mockResolvedValue(
          mockDepartmentWithMultipleDefaultPools as unknown as Department,
        );
        vi.mocked(orgUnitService.updateProject).mockResolvedValue(mockProjectInheritingPools as unknown as Project);
        await mountComponent(editRoute);
      });

      it("sends null when user saves without changing the pre-filled list (inherit unchanged)", async () => {
        await triggerSaveUnchanged();
        expect(getUpdatePayload().defaultNodePools).toBeNull();
      });

      it("sends the explicit list when user reduces to only 'default' (the customer bug scenario)", async () => {
        await triggerSaveWithPools(["default"]);
        expect(getUpdatePayload().defaultNodePools).toEqual(["default"]);
      });

      it("sends the explicit list when user removes some pools from the department list", async () => {
        await triggerSaveWithPools(["default", "dist"]);
        expect(getUpdatePayload().defaultNodePools).toEqual(["default", "dist"]);
      });

      it("sends null when user sets pools identical to the department list (no real change)", async () => {
        await triggerSaveWithPools(["default", "dist", "driver-test"]);
        expect(getUpdatePayload().defaultNodePools).toBeNull();
      });
    });

    describe("existing project with an explicit single-pool override already set", () => {
      beforeEach(async () => {
        vi.mocked(orgUnitService.getProject).mockResolvedValue(mockProjectWithExplicitDefaultPool as unknown as Project);
        vi.mocked(orgUnitService.getDepartment).mockResolvedValue(
          mockDepartmentWithMultipleDefaultPools as unknown as Department,
        );
        vi.mocked(orgUnitService.updateProject).mockResolvedValue(
          mockProjectWithExplicitDefaultPool as unknown as Project,
        );
        await mountComponent(editRoute);
      });

      it("preserves the existing ['default'] override when user saves without changing", async () => {
        await triggerSaveUnchanged();
        expect(getUpdatePayload().defaultNodePools).toEqual(["default"]);
      });

      it("sends the new value when user changes the explicit override", async () => {
        await triggerSaveWithPools(["default", "dist"]);
        expect(getUpdatePayload().defaultNodePools).toEqual(["default", "dist"]);
      });
    });

    describe("new project default node pools", () => {
      it("sends null when user creates a project without changing the pre-filled pools", async () => {
        vi.mocked(orgUnitService.getDepartment).mockResolvedValue(
          mockDepartmentWithMultipleDefaultPools as unknown as Department,
        );
        vi.mocked(orgUnitService.getDefaultDepartment).mockResolvedValue(
          mockDepartmentWithMultipleDefaultPools as unknown as Department,
        );
        await mountComponent();

        const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
        form.vm.$emit("on-save", {
          project: form.props("project"),
          action: ESaveOrgUnitAction.Close,
          onComplete: vi.fn(),
        });
        await flushPromises();

        const createPayload = vi.mocked(orgUnitService.createProject).mock.calls[0][0];
        expect(createPayload.defaultNodePools).toBeNull();
      });
    });
  });

  // ==========================================================================
  // Node Types Inheritance at Save Time
  // ==========================================================================

  describe("Node Types Inheritance (prepareProjectForSave)", () => {
    const triggerSave = async () => {
      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      form.vm.$emit("on-save", {
        project: form.props("project"),
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();
    };

    it("includes parent department nodeTypes in creation payload for new project", async () => {
      vi.mocked(orgUnitService.getDefaultDepartment).mockResolvedValue(
        mockDepartmentWithNodeTypes as unknown as Department,
      );
      await mountComponent();
      await triggerSave();

      const createCall = vi.mocked(orgUnitService.createProject).mock.calls[0][0];
      expect(createCall.nodeTypes).toEqual(mockDepartmentWithNodeTypes.nodeTypes);
    });

    it("does not include nodeTypes when department has no nodeTypes", async () => {
      await mountComponent();
      await triggerSave();

      const createCall = vi.mocked(orgUnitService.createProject).mock.calls[0][0];
      expect(createCall.nodeTypes).toBeUndefined();
    });

    it("does not override nodeTypes for existing project update", async () => {
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartmentWithNodeTypes as unknown as Department);

      await mountComponent({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: PROJECT_ID },
      });

      const form = wrapper.findComponent({ name: "ProjectEditFormV2" });
      form.vm.$emit("on-save", {
        project: { ...mockProject, name: "Updated" },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(orgUnitService.updateProject).toHaveBeenCalled();
      expect(orgUnitService.createProject).not.toHaveBeenCalled();
    });
  });
});
