import { defineComponent, h } from "vue";
import { vi } from "vitest";
import type { Project, Department, ProjectCreationRequest } from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";

// ============================================================================
// Project Mocks
// ============================================================================

export const mockProject = {
  id: "project-1",
  name: "test-project",
  description: "test description",
  clusterId: "cluster-1",
  parentId: "dept-1",
  defaultNodePools: ["default"],
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: 1000 },
    },
  ],
  status: {
    nodePoolQuotaStatuses: [{ nodePoolId: "pool-1", nodePoolName: "default", allocatedNonPreemptible: { cpu: 500 } }],
  },
} as Project;

export const mockProjectCreationRequest = {
  name: "new-project",
  description: "new project description",
  clusterId: "cluster-1",
  parentId: "dept-1",
  resources: [],
} as ProjectCreationRequest;

export const mockMultiNodePoolProject = {
  ...mockProject,
  defaultNodePools: ["pool1", "pool2"],
  resources: [
    { nodePool: { name: "pool1", id: "1" }, gpu: { deserved: 0 }, cpu: { deserved: 1000 } },
    { nodePool: { name: "pool2", id: "2" }, gpu: { deserved: 0 }, cpu: { deserved: 2000 } },
  ],
  status: {
    nodePoolQuotaStatuses: [
      { nodePoolId: "1", nodePoolName: "pool1", allocatedNonPreemptible: { cpu: 500 } },
      { nodePoolId: "2", nodePoolName: "pool2", allocatedNonPreemptible: { cpu: 500 } },
    ],
  },
} as Project;

// ============================================================================
// Department Mocks
// ============================================================================

export const mockDepartment = {
  id: "dept-1",
  name: "Test Department",
  description: "Test Department Description",
  clusterId: "cluster-1",
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: 2000 },
    },
  ],
} as Department;

export const mockMultiNodePoolDepartment = {
  ...mockDepartment,
  resources: [
    { nodePool: { name: "pool1", id: "1" }, gpu: { deserved: 0 }, cpu: { deserved: 2000 } },
    { nodePool: { name: "pool2", id: "2" }, gpu: { deserved: 0 }, cpu: { deserved: 4000 } },
  ],
} as Department;

// ============================================================================
// CPU Quota State Scenario Mocks
// ============================================================================

/**
 * Project with CPU quota exceeding department limit
 * Project quota: 3000 (3 cores), Department limit: 2000 (2 cores)
 * Should show Warning (or Error if over-subscription enabled)
 */
export const mockProjectExceedingQuota = {
  ...mockProject,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: 3000 }, // 3 cores - exceeds department's 2 cores
    },
  ],
} as Project;

/**
 * Project with CPU quota under allocated non-preemptible usage
 * Project quota: 500 (0.5 cores), Allocated: 1000 (1 core)
 * Should show Warning (or Error if over-subscription enabled)
 */
export const mockProjectUnderAllocated = {
  ...mockProject,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: 500 }, // 0.5 cores - under allocated 1 core
    },
  ],
  status: {
    nodePoolQuotaStatuses: [
      {
        nodePoolId: "pool-1",
        nodePoolName: "default",
        allocatedNonPreemptible: { cpu: 1000 }, // 1 core allocated
      },
    ],
  },
} as Project;

/**
 * Project with CPU quota within limits
 * Project quota: 1000 (1 core), Department limit: 2000 (2 cores), Allocated: 500 (0.5 cores)
 * Should show Default state
 */
export const mockProjectWithinLimits = {
  ...mockProject,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: 1000 }, // 1 core - within department's 2 cores
    },
  ],
  status: {
    nodePoolQuotaStatuses: [
      {
        nodePoolId: "pool-1",
        nodePoolName: "default",
        allocatedNonPreemptible: { cpu: 500 }, // 0.5 cores allocated
      },
    ],
  },
} as Project;

/**
 * Project with unlimited CPU quota (null)
 * Parent department has a limit, so this should show Warning/Error
 */
export const mockProjectUnlimitedQuota = {
  ...mockProject,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: null }, // Unlimited
    },
  ],
} as Project;

/**
 * Department with limited CPU quota
 */
export const mockDepartmentWithLimit = {
  ...mockDepartment,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: 2000 }, // 2 cores limit
    },
  ],
} as Department;

/**
 * Department with unlimited CPU quota
 */
export const mockDepartmentUnlimited = {
  ...mockDepartment,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 0 },
      cpu: { deserved: null }, // Unlimited
    },
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
export const mockGpuQuotaDisabledNodePool2: NodepoolMinimal = {
  id: "pool-4",
  name: "no-gpu-quota-pool4",
  clusterId: "cluster-1",
  schedulingCapabilities: {
    adjustableWeights: true,
    ranks: true,
    quota: false,
    cpuQuota: true,
  },
};

/**
 * Array of node pools with CPU quota enabled - use as nodePoolsMinimal prop
 * to test scenarios where CPU quota step should be visible
 */
export const mockCpuEnabledNodePools: NodepoolMinimal[] = [mockCpuEnabledNodePool, mockGpuQuotaDisabledNodePool2];

/**
 * Empty array for testing when no node pools exist - use as nodePoolsMinimal prop
 * to test scenarios where CPU quota step should be hidden
 */
export const mockNoCpuEnabledNodePools: NodepoolMinimal[] = [];

// Aliases for clearer naming when used as nodePoolsMinimal prop
export const mockNodePoolsMinimalWithCpuQuota = mockCpuEnabledNodePools;
export const mockNodePoolsMinimalEmpty = mockNoCpuEnabledNodePools;

/**
 * Array of node pools with GPU quota disabled - use as nodePoolsMinimal prop
 * to test scenarios where GPU quota step should be hidden
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
  name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
  params: { id: "" },
  query: { clusterId: "test-cluster-id" },
  meta: {
    backPageName: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
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
  OverQuota = "over-quota",
  CPUQuota = "cpu-quota",
  SchedulingRules = "scheduling-rules",
  AccessRules = "access-rules",
}
