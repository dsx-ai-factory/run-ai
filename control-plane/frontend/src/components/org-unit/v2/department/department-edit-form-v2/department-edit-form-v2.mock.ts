import { defineComponent, h } from "vue";
import { vi } from "vitest";
import type { Department, DepartmentCreationRequest } from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";

// ============================================================================
// Department Mocks
// ============================================================================

export const mockDepartment = {
  id: "dept-1",
  name: "test-department",
  description: "test description",
  clusterId: "cluster-1",
  defaultNodePools: ["default"],
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 10 },
      cpu: { deserved: 2000 },
    },
  ],
} as Department;

export const mockDepartmentCreationRequest = {
  name: "new-department",
  description: "new department description",
  clusterId: "cluster-1",
  resources: [],
} as DepartmentCreationRequest;

export const mockMultiNodePoolDepartment = {
  ...mockDepartment,
  defaultNodePools: ["pool1", "pool2"],
  resources: [
    { nodePool: { name: "pool1", id: "1" }, gpu: { deserved: 10 }, cpu: { deserved: 2000 } },
    { nodePool: { name: "pool2", id: "2" }, gpu: { deserved: 20 }, cpu: { deserved: 4000 } },
  ],
} as Department;

// ============================================================================
// Node Pools Metadata Mocks (used as nodePoolsMinimal prop)
// ============================================================================

/**
 * Node pool with CPU quota enabled
 */
export const mockCpuEnabledNodePool: NodepoolMinimal = {
  id: "pool-1",
  name: "default",
  clusterId: "cluster-1",
  schedulingCapabilities: {
    adjustableWeights: true,
    ranks: true,
    quota: true,
    cpuQuota: true,
  },
};

/**
 * Node pool with CPU quota disabled
 */
export const mockCpuDisabledNodePool: NodepoolMinimal = {
  id: "pool-2",
  name: "gpu-only-pool",
  clusterId: "cluster-1",
  schedulingCapabilities: {
    adjustableWeights: true,
    ranks: true,
    quota: true,
    cpuQuota: false,
  },
};

/**
 * Node pool with GPU quota disabled (schedulingCapabilities.quota: false)
 */
export const mockGpuQuotaDisabledNodePool: NodepoolMinimal = {
  id: "pool-3",
  name: "no-gpu-quota-pool",
  clusterId: "cluster-1",
  schedulingCapabilities: {
    adjustableWeights: true,
    ranks: true,
    quota: false,
    cpuQuota: true,
  },
};

/**
 * Array of node pools with CPU quota enabled
 */
export const mockCpuEnabledNodePools: NodepoolMinimal[] = [mockCpuEnabledNodePool];

/**
 * Empty array for testing when no node pools exist
 */
export const mockNoCpuEnabledNodePools: NodepoolMinimal[] = [];

/**
 * Array of node pools with GPU quota disabled
 */
export const mockNoGpuQuotaEnabledNodePools: NodepoolMinimal[] = [mockGpuQuotaDisabledNodePool];

/**
 * Array of node pools with both CPU and GPU quota enabled
 */
export const mockFullCapabilitiesNodePools: NodepoolMinimal[] = [mockCpuEnabledNodePool];

// ============================================================================
// Router Mocks
// ============================================================================

export const createMockRouter = () => ({
  push: vi.fn(),
  back: vi.fn(),
  replace: vi.fn(),
});

export const createMockRoute = (overrides = {}) => ({
  name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_NEW,
  params: { id: "" },
  query: { clusterId: "test-cluster-id" },
  meta: {
    backPageName: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
  },
  ...overrides,
});

// ============================================================================
// Component Stubs
// ============================================================================

/**
 * Stub for RunaiFormWrapper that exposes validate method
 * Always returns true for validation to allow testing form submission flows
 */
export const RunaiFormWrapperStub = defineComponent({
  name: "RunaiFormWrapper",
  setup(_, { slots, expose }) {
    const validate = () => Promise.resolve(true);
    expose({ validate });
    return () => h("div", { class: "runai-form-wrapper-stub" }, slots.default?.());
  },
});

/**
 * Stub for RunaiFormWrapper that fails validation
 * Useful for testing validation error scenarios
 */
export const RunaiFormWrapperFailingValidationStub = defineComponent({
  name: "RunaiFormWrapper",
  setup(_, { slots, expose }) {
    const validate = () => Promise.resolve(false);
    expose({ validate });
    return () => h("div", { class: "runai-form-wrapper-stub" }, slots.default?.());
  },
});

// ============================================================================
// Storage Mock
// ============================================================================

export const createStorageUtilMock = () => ({
  get: vi.fn(),
  save: vi.fn(),
  remove: vi.fn(),
});

// ============================================================================
// Test Helpers
// ============================================================================

/**
 * Helper to create step click payload
 */
export const createStepClickPayload = (stepName: string) => ({
  step: { name: stepName },
  name: stepName,
});

/**
 * Wizard step names enum for test consistency
 */
export enum ETestStepNames {
  ScopeName = "scope-name",
  GpuQuota = "gpu-quota",
  CPUQuota = "cpu-quota",
  SchedulingRules = "scheduling-rules",
  AccessRules = "access-rules",
}
