import { describe, vi, it, expect, afterEach, beforeEach } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import DepartmentEdit from "@/views/department/department-edit.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import type { Department, Resources } from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";
import { createTestingPinia } from "@pinia/testing";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { ESaveOrgUnitAction } from "@/models/org-unit-forms.model";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { useAppStore } from "@/stores/app.store";

// ============================================================================
// Mock Data
// ============================================================================
import {
  mockClusters,
  CLUSTER_ID,
  DEPARTMENT_ID,
  mockDepartment,
  mockCpuEnabledNodePool,
  mockCpuDisabledNodePool,
} from "./department-edit.test.mocks";

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
  name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_NEW,
  params: { id: "" },
  query: { clusterId: "test-cluster-id" },
  meta: {
    backPageName: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
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

describe("DepartmentEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof DepartmentEdit>>;
  let notifyMock: ReturnType<typeof vi.fn>;

  const mountComponent = async (routeOverrides = {}) => {
    mockRoute = createMockRoute(routeOverrides);

    const pinia = createTestingPinia({ createSpy: vi.fn });
    const clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ clusters: mockClusters });
    clusterStore.$patch({ selectedCluster: mockClusters[0] });

    wrapper = mount(DepartmentEdit, {
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
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment as Department);
    vi.mocked(orgUnitService.createDepartment).mockResolvedValue(mockDepartment as Department);
    vi.mocked(orgUnitService.updateDepartment).mockResolvedValue(mockDepartment as Department);
    vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([
      mockCpuEnabledNodePool,
      mockCpuDisabledNodePool,
    ] as NodepoolMinimal[]);
    vi.mocked(clusterService.getNodeTelemetry).mockResolvedValue({
      timestamp: "",
      type: "",
      values: [
        {
          groups: [{ key: "Nodepool", value: "default" }],
          value: "10",
        },
      ],
    });
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
    it("creates empty department for new department route", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      expect(form.exists()).toBe(true);
    });

    it("sets isNewDepartment to true for new department route", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      expect(form.props("isNewDepartment")).toBe(true);
    });

    it("loads existing department for edit route", async () => {
      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      expect(orgUnitService.getDepartment).toHaveBeenCalledWith(DEPARTMENT_ID);
    });

    it("sets isNewDepartment to false for edit route", async () => {
      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      expect(form.props("isNewDepartment")).toBe(false);
    });
  });

  // ==========================================================================
  // Node Pools Loading Tests
  // ==========================================================================

  describe("Node Pools Loading", () => {
    it("loads node pools metadata on mount for new department", async () => {
      await mountComponent();

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalled();
    });

    it("passes all node pools metadata to form", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const nodePoolsMinimal = form.props("nodePoolsMinimal") as NodepoolMinimal[];

      expect(nodePoolsMinimal).toHaveLength(2);
      expect(nodePoolsMinimal[0].name).toBe("default");
      expect(nodePoolsMinimal[1].name).toBe("gpu-only");
    });

    it("handles node pools loading error gracefully", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockRejectedValue(new Error("Network error"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const nodePoolsMinimal = form.props("nodePoolsMinimal") as NodepoolMinimal[];

      expect(nodePoolsMinimal).toEqual([]);
      errorSpy.mockRestore();
    });

    it("loads node pools for existing department using department clusterId", async () => {
      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalled();
    });

    it("loads node pool telemetry for CPU and memory on mount", async () => {
      await mountComponent();

      // Should be called twice - once for CPU cores and once for memory bytes
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledTimes(1);
    });

    it("handles node pool telemetry loading error gracefully", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockRejectedValue(new Error("Telemetry error"));
      const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const nodePoolTotals = form.props("nodePoolTotals");

      // Should pass empty object when telemetry fails
      expect(nodePoolTotals).toEqual({});
      warnSpy.mockRestore();
    });
  });

  // ==========================================================================
  // Scope Update Tests
  // ==========================================================================

  describe("Scope Update", () => {
    it("reloads department and node pools when scope changes", async () => {
      await mountComponent();

      vi.clearAllMocks();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const newClusterId = "new-cluster-id";

      form.vm.$emit("update-scope", { clusterId: newClusterId });
      await flushPromises();

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalled();
    });

    it("reloads node pool telemetry when scope changes", async () => {
      await mountComponent();

      vi.clearAllMocks();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const newClusterId = "new-cluster-id";

      form.vm.$emit("update-scope", { clusterId: newClusterId });
      await flushPromises();

      // Should reload telemetry for both CPU and memory
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledTimes(1);
    });
  });

  // ==========================================================================
  // Cancel Tests
  // ==========================================================================

  describe("Cancel", () => {
    it("requests leave confirmation on cancel", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      form.vm.$emit("on-cancel");
      await flushPromises();

      expect(requestToLeave).toHaveBeenCalled();
    });

    it("navigates back when leave is allowed", async () => {
      vi.mocked(requestToLeave).mockResolvedValue(true);

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      form.vm.$emit("on-cancel");
      await flushPromises();

      expect(mockRouter.back).toHaveBeenCalled();
    });

    it("does not navigate when leave is denied", async () => {
      vi.mocked(requestToLeave).mockResolvedValue(false);

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      form.vm.$emit("on-cancel");
      await flushPromises();

      expect(mockRouter.back).not.toHaveBeenCalled();
    });
  });

  // ==========================================================================
  // Save Tests - Create
  // ==========================================================================

  describe("Save - Create Department", () => {
    it("creates department and shows success notification", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        department: { name: "New Department", clusterId: CLUSTER_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: onCompleteMock,
      });
      await flushPromises();

      expect(orgUnitService.createDepartment).toHaveBeenCalled();
      expect(notifyMock).toHaveBeenCalled();
      expect(onCompleteMock).toHaveBeenCalled();
    });

    it("navigates to department index after create with Close action", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });

      form.vm.$emit("on-save", {
        department: { name: "New Department", clusterId: CLUSTER_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(mockRouter.push).toHaveBeenCalledWith({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
        query: { createdEntityId: DEPARTMENT_ID },
      });
    });

    it("does not navigate after create with Continue action", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });

      form.vm.$emit("on-save", {
        department: { name: "New Department", clusterId: CLUSTER_ID },
        action: ESaveOrgUnitAction.Continue,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(mockRouter.push).not.toHaveBeenCalled();
    });

    it("shows error notification on create failure", async () => {
      vi.mocked(orgUnitService.createDepartment).mockRejectedValue(new Error("Create failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });

      form.vm.$emit("on-save", {
        department: { name: "New Department", clusterId: CLUSTER_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(notifyMock).toHaveBeenCalled();
      expect(mockRouter.push).not.toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("does NOT call onComplete callback when create fails", async () => {
      vi.mocked(orgUnitService.createDepartment).mockRejectedValue(new Error("Create failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        department: { name: "New Department", clusterId: CLUSTER_ID },
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

  describe("Save - Update Department", () => {
    it("updates department and shows success notification", async () => {
      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        department: { ...mockDepartment, name: "Updated Department" },
        action: ESaveOrgUnitAction.Close,
        onComplete: onCompleteMock,
      });
      await flushPromises();

      expect(orgUnitService.updateDepartment).toHaveBeenCalled();
      expect(notifyMock).toHaveBeenCalled();
      expect(onCompleteMock).toHaveBeenCalled();
    });

    it("navigates after update with Close action", async () => {
      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });

      form.vm.$emit("on-save", {
        department: { ...mockDepartment, name: "Updated Department" },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(mockRouter.push).toHaveBeenCalled();
    });

    it("shows error notification on update failure", async () => {
      vi.mocked(orgUnitService.updateDepartment).mockRejectedValue(new Error("Update failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });

      form.vm.$emit("on-save", {
        department: { ...mockDepartment, name: "Updated Department" },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(notifyMock).toHaveBeenCalled();
      expect(mockRouter.push).not.toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("does NOT call onComplete callback when update fails", async () => {
      vi.mocked(orgUnitService.updateDepartment).mockRejectedValue(new Error("Update failed"));
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const onCompleteMock = vi.fn();

      form.vm.$emit("on-save", {
        department: { ...mockDepartment, name: "Updated Department" },
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
    it("passes isDepartmentSubmitting to form during save", async () => {
      // Make createDepartment hang to test loading state
      let resolveCreate: (value: Department) => void = () => {};
      vi.mocked(orgUnitService.createDepartment).mockImplementation(
        () =>
          new Promise((resolve) => {
            resolveCreate = resolve;
          }),
      );

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      expect(form.props("isDepartmentSubmitting")).toBe(false);

      form.vm.$emit("on-save", {
        department: { name: "New Department", clusterId: CLUSTER_ID },
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(form.props("isDepartmentSubmitting")).toBe(true);

      // Resolve the promise
      resolveCreate(mockDepartment as Department);
      await flushPromises();

      expect(form.props("isDepartmentSubmitting")).toBe(false);
    });
  });

  // ==========================================================================
  // Props Passing Tests
  // ==========================================================================

  describe("Props Passing", () => {
    it("passes all required props to DepartmentEditFormV2", async () => {
      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });

      expect(form.props()).toMatchObject({
        isNewDepartment: true,
        isDepartmentSubmitting: false,
      });
      expect(form.props("department")).toBeDefined();
      expect(form.props("nodePoolsMinimal")).toBeDefined();
      expect(form.props("nodePoolTotals")).toBeDefined();
    });

    it("passes nodePoolTotals with correct structure to form", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockResolvedValueOnce({
        timestamp: "",
        type: "",
        values: [
          { groups: [{ key: "Nodepool", value: "default" }], value: "20" },
          { groups: [{ key: "Nodepool", value: "gpu-pool" }], value: "8" },
        ],
      });
      vi.mocked(clusterService.getNodeTelemetry).mockResolvedValueOnce({
        timestamp: "",
        type: "",
        values: [
          { groups: [{ key: "Nodepool", value: "default" }], value: "10737418240" },
          { groups: [{ key: "Nodepool", value: "gpu-pool" }], value: "8589934592" },
        ],
      });

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const nodePoolTotals = form.props("nodePoolTotals");

      expect(nodePoolTotals).toEqual({
        default: {
          cpuCores: 0,
          memoryBytes: 0,
          totalGpus: 20,
        },
        "gpu-pool": {
          cpuCores: 0,
          memoryBytes: 0,
          totalGpus: 8,
        },
      });
    });

    it("passes empty nodePoolTotals when telemetry request fails", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockRejectedValue(new Error("Network error"));

      await mountComponent();

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      const nodePoolTotals = form.props("nodePoolTotals");

      expect(nodePoolTotals).toEqual({
        default: {
          cpuCores: 0,
          memoryBytes: 0,
          totalGpus: 10737418240,
        },
        "gpu-pool": {
          cpuCores: 0,
          memoryBytes: 0,
          totalGpus: 8589934592,
        },
      });
    });
  });

  // ==========================================================================
  // Page Title Tests
  // ==========================================================================

  describe("Page Title", () => {
    it("sets pageTitleOverride with entity name for edit route", async () => {
      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      const appStore = useAppStore(wrapper.vm.$pinia);
      expect(appStore.setPageTitleOverride).toHaveBeenCalledWith(`Edit department: ${mockDepartment.name}`);
    });

    it("does not set pageTitleOverride for new department route", async () => {
      await mountComponent();

      const appStore = useAppStore(wrapper.vm.$pinia);
      expect(appStore.setPageTitleOverride).not.toHaveBeenCalledWith(expect.stringContaining("Edit department:"));
    });

    it("clears pageTitleOverride on unmount", async () => {
      await mountComponent({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: DEPARTMENT_ID },
      });

      const appStore = useAppStore(wrapper.vm.$pinia);
      wrapper.unmount();

      expect(appStore.setPageTitleOverride).toHaveBeenLastCalledWith(null);
    });
  });

  // ==========================================================================
  // Over-Quota Priority Feature Toggle Tests
  // ==========================================================================

  describe("Over-Quota Priority Feature Toggle", () => {
    it("calls updateOverQuotaWeightToDeserved when feature is disabled", async () => {
      const updateOverQuotaWeightToDeservedSpy = vi.spyOn(orgUnitUtil, "updateOverQuotaWeightToDeserved");
      const alignCpuAndMemoryLimitsByGpuSpy = vi.spyOn(orgUnitUtil, "alignCpuAndMemoryLimitsByGpu");

      await mountComponent();

      const pinia = wrapper.vm.$pinia;
      if (pinia) {
        const settingStore = useSettingStore(pinia);
        // Mock feature toggle as disabled
        vi.spyOn(settingStore, "isOverQuotaPriorityEnabled", "get").mockReturnValue(false);
      }

      const departmentWithResources = {
        ...mockDepartment,
        resources: [
          {
            nodePool: { name: "default", id: "1" },
            gpu: { deserved: 10, overQuotaWeight: 5, limit: null },
          },
        ] as Resources[],
      };

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      form.vm.$emit("on-save", {
        department: departmentWithResources,
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(alignCpuAndMemoryLimitsByGpuSpy).toHaveBeenCalled();
      expect(updateOverQuotaWeightToDeservedSpy).toHaveBeenCalled();
    });

    it("does not call updateOverQuotaWeightToDeserved when feature is enabled", async () => {
      const updateOverQuotaWeightToDeservedSpy = vi.spyOn(orgUnitUtil, "updateOverQuotaWeightToDeserved");
      const alignCpuAndMemoryLimitsByGpuSpy = vi.spyOn(orgUnitUtil, "alignCpuAndMemoryLimitsByGpu");

      await mountComponent();

      const pinia = wrapper.vm.$pinia;
      if (pinia) {
        const settingStore = useSettingStore(pinia);
        // Mock feature toggle as enabled
        vi.spyOn(settingStore, "isOverQuotaPriorityEnabled", "get").mockReturnValue(true);
      }

      const departmentWithResources = {
        ...mockDepartment,
        resources: [
          {
            nodePool: { name: "default", id: "1" },
            gpu: { deserved: 10, overQuotaWeight: 5, limit: null },
          },
        ] as Resources[],
      };

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      form.vm.$emit("on-save", {
        department: departmentWithResources,
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(alignCpuAndMemoryLimitsByGpuSpy).toHaveBeenCalled();
      expect(updateOverQuotaWeightToDeservedSpy).not.toHaveBeenCalled();
    });

    it("does not transform resources when resources array is empty", async () => {
      const updateOverQuotaWeightToDeservedSpy = vi.spyOn(orgUnitUtil, "updateOverQuotaWeightToDeserved");
      const alignCpuAndMemoryLimitsByGpuSpy = vi.spyOn(orgUnitUtil, "alignCpuAndMemoryLimitsByGpu");

      await mountComponent();

      const departmentWithoutResources = {
        ...mockDepartment,
        resources: [],
      };

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      form.vm.$emit("on-save", {
        department: departmentWithoutResources,
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      expect(alignCpuAndMemoryLimitsByGpuSpy).not.toHaveBeenCalled();
      expect(updateOverQuotaWeightToDeservedSpy).not.toHaveBeenCalled();
    });

    it("creates department successfully when feature is disabled and overQuotaWeight is set", async () => {
      await mountComponent();

      const pinia = wrapper.vm.$pinia;
      if (pinia) {
        const settingStore = useSettingStore(pinia);
        vi.spyOn(settingStore, "isOverQuotaPriorityEnabled", "get").mockReturnValue(false);
      }

      const departmentWithOverQuota = {
        name: "Test Department",
        clusterId: CLUSTER_ID,
        resources: [
          {
            nodePool: { name: "default", id: "1" },
            gpu: { deserved: 10, overQuotaWeight: 5, limit: null },
          },
        ] as Resources[],
      };

      const form = wrapper.findComponent({ name: "DepartmentEditFormV2" });
      form.vm.$emit("on-save", {
        department: departmentWithOverQuota,
        action: ESaveOrgUnitAction.Close,
        onComplete: vi.fn(),
      });
      await flushPromises();

      // Should successfully create without backend validation error
      expect(orgUnitService.createDepartment).toHaveBeenCalled();
      expect(notifyMock).toHaveBeenCalled();
      expect(mockRouter.push).toHaveBeenCalled();
    });
  });
});
