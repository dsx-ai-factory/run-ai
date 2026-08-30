import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import { Quasar } from "quasar";

import type { INodepoolSchedulingCapabilities } from "@/composables/org-unit/use-nodepool-scheduling-capabilities.composable";
import { useNodepoolSchedulingCapabilities } from "@/composables/org-unit/use-nodepool-scheduling-capabilities.composable";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { NodepoolMinimal, SchedulingCapabilities } from "@/swagger-models/cluster-service-client";

// ============================================================================
// Mock Setup
// ============================================================================

vi.mock("@/services/control-plane/cluster.service/cluster.service");

const mockNotify = vi.fn();
vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    useQuasar: () => ({
      notify: mockNotify,
    }),
  };
});

// ============================================================================
// Test Component
// ============================================================================

interface ComposableResult {
  loading: boolean;
  capabilitiesByClusterId: Record<string, INodepoolSchedulingCapabilities>;
  loadCapabilities: (clusterId: string) => Promise<void>;
}

const ComposableHost = defineComponent({
  setup() {
    const { loading, capabilitiesByClusterId, loadCapabilities } = useNodepoolSchedulingCapabilities();
    return { loading, capabilitiesByClusterId, loadCapabilities };
  },
  template: "<div />",
});

const mountComposable = (): ComposableResult => {
  const wrapper = mount(ComposableHost, {
    global: {
      plugins: [Quasar],
    },
  });
  return wrapper.vm as unknown as ComposableResult;
};

// ============================================================================
// Mock Data Helpers
// ============================================================================

const CLUSTER_ID_1 = "cluster-1";
const CLUSTER_ID_2 = "cluster-2";

const createSchedulingCapabilities = (quota: boolean, cpuQuota: boolean): SchedulingCapabilities => ({
  quota,
  cpuQuota,
  adjustableWeights: false,
  ranks: false,
});

const createNodePool = (
  id: string,
  name: string,
  clusterId: string,
  quota: boolean,
  cpuQuota: boolean,
): NodepoolMinimal => ({
  id,
  name,
  clusterId,
  schedulingCapabilities: createSchedulingCapabilities(quota, cpuQuota),
});

const mockNodePoolWithCpuAndGpuQuota: NodepoolMinimal[] = [
  createNodePool("pool-1", "default", CLUSTER_ID_1, true, true),
];

const mockNodePoolWithOnlyGpuQuota: NodepoolMinimal[] = [
  createNodePool("pool-2", "gpu-only", CLUSTER_ID_1, true, false),
];

const mockNodePoolWithNeitherQuota: NodepoolMinimal[] = [
  createNodePool("pool-3", "no-quota", CLUSTER_ID_1, false, false),
];

const mockMixedNodePools: NodepoolMinimal[] = [
  createNodePool("pool-1", "with-cpu", CLUSTER_ID_1, true, true),
  createNodePool("pool-2", "no-cpu", CLUSTER_ID_1, false, false),
];

// ============================================================================
// Test Suite
// ============================================================================

describe("useNodepoolSchedulingCapabilities", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNotify.mockClear();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  // ==========================================================================
  // Initial State Tests
  // ==========================================================================

  describe("Initial State", () => {
    it("initializes with loading as false", () => {
      const vm = mountComposable();
      expect(vm.loading).toBe(false);
    });

    it("initializes with empty capabilities map", () => {
      const vm = mountComposable();
      expect(Object.keys(vm.capabilitiesByClusterId)).toHaveLength(0);
    });
  });

  // ==========================================================================
  // Loading Capabilities Tests
  // ==========================================================================

  describe("Loading Capabilities", () => {
    it("sets loading to true while fetching", async () => {
      let resolvePromise: (value: NodepoolMinimal[]) => void = () => {};
      vi.mocked(clusterService.getNodepoolsMinimal).mockImplementation(
        () =>
          new Promise((resolve) => {
            resolvePromise = resolve;
          }),
      );

      const vm = mountComposable();
      const loadPromise = vm.loadCapabilities(CLUSTER_ID_1);

      expect(vm.loading).toBe(true);

      resolvePromise(mockNodePoolWithCpuAndGpuQuota);
      await loadPromise;
      await flushPromises();

      expect(vm.loading).toBe(false);
    });

    it("correctly identifies cpuQuota capability when present", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePoolWithCpuAndGpuQuota);

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(vm.capabilitiesByClusterId[CLUSTER_ID_1]).toEqual({
        cpuQuota: true,
        quota: true,
      });
    });

    it("correctly identifies when cpuQuota is disabled", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePoolWithOnlyGpuQuota);

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(vm.capabilitiesByClusterId[CLUSTER_ID_1]).toEqual({
        cpuQuota: false,
        quota: true,
      });
    });

    it("correctly identifies when neither quota is enabled", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePoolWithNeitherQuota);

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(vm.capabilitiesByClusterId[CLUSTER_ID_1]).toEqual({
        cpuQuota: false,
        quota: false,
      });
    });

    it("returns true if any node pool has the capability", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockMixedNodePools);

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(vm.capabilitiesByClusterId[CLUSTER_ID_1]).toEqual({
        cpuQuota: true,
        quota: true,
      });
    });
  });

  // ==========================================================================
  // Caching Tests
  // ==========================================================================

  describe("Caching", () => {
    it("caches capabilities and does not refetch for same cluster", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePoolWithCpuAndGpuQuota);

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledTimes(1);

      // Load again for same cluster
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      // Should not have fetched again
      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledTimes(1);
    });

    it("fetches for different clusters separately", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePoolWithCpuAndGpuQuota);

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      await vm.loadCapabilities(CLUSTER_ID_2);
      await flushPromises();

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledTimes(2);
      expect(vm.capabilitiesByClusterId[CLUSTER_ID_1]).toBeDefined();
      expect(vm.capabilitiesByClusterId[CLUSTER_ID_2]).toBeDefined();
    });
  });

  // ==========================================================================
  // Error Handling Tests
  // ==========================================================================

  describe("Error Handling", () => {
    it("sets capabilities to false on error", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(clusterService.getNodepoolsMinimal).mockRejectedValue(new Error("Network error"));

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(vm.capabilitiesByClusterId[CLUSTER_ID_1]).toEqual({
        cpuQuota: false,
        quota: false,
      });
      errorSpy.mockRestore();
    });

    it("sets loading to false after error", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(clusterService.getNodepoolsMinimal).mockRejectedValue(new Error("Network error"));

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(vm.loading).toBe(false);
      errorSpy.mockRestore();
    });

    it("shows error notification on failure", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(clusterService.getNodepoolsMinimal).mockRejectedValue(new Error("Network error"));

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(mockNotify).toHaveBeenCalled();
      errorSpy.mockRestore();
    });
  });

  // ==========================================================================
  // API Call Tests
  // ==========================================================================

  describe("API Calls", () => {
    it("calls getNodepoolsMinimal with correct cluster filter", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([]);

      const vm = mountComposable();
      await vm.loadCapabilities(CLUSTER_ID_1);
      await flushPromises();

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledWith([`clusterId==${CLUSTER_ID_1}`]);
    });
  });
});
