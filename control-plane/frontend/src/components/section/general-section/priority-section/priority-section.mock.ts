import type { ClusterPriority } from "@/swagger-models/workloads-client";
import type { StringRules, StringOption } from "@/swagger-models/policy-service-client";
import { ScopeType } from "@/swagger-models/authorization-client";
import type {
  Project,
  TotalResources,
  ProjectEffective,
  Units,
  ProjectPhase,
} from "@/swagger-models/org-unit-service-client";
import type { IWorkloadPriorityData } from "../general-section.model";

export const MOCK_PRIORITY_IDS = {
  VERY_LOW: "90c87fb9-9e49-40db-bddf-962ff169ae0e",
  LOW: "98be7a8c-2eb7-4f31-aa9d-89901682ff24",
  MEDIUM_LOW: "46af985f-4340-4158-86a0-aac74f1b3a88",
  MEDIUM: "13a7592a-bf49-48e1-b40b-6939bffa7c4e",
  MEDIUM_HIGH: "84346616-ec99-4754-985b-06daf1f9a0ce",
  HIGH: "224ae4a1-a960-4c83-9b2f-17cac6c2d0e8",
  VERY_HIGH: "518d8dcf-3e17-436d-9f03-260d98f2c6f6",
} as const;

export const MOCK_PRIORITY_NAMES = {
  VERY_LOW: "very-low",
  LOW: "low",
  MEDIUM_LOW: "medium-low",
  MEDIUM: "medium",
  MEDIUM_HIGH: "medium-high",
  HIGH: "high",
  VERY_HIGH: "very-high",
} as const;

export const mockWorkloadPriorities: ClusterPriority[] = [
  {
    id: MOCK_PRIORITY_IDS.LOW,
    source: "Run.ai",
    name: MOCK_PRIORITY_NAMES.LOW,
    value: 40,
    createdAt: "0001-01-01T00:00:00Z",
    updatedAt: "0001-01-01T00:00:00Z",
  },
  {
    id: MOCK_PRIORITY_IDS.VERY_LOW,
    source: "Run.ai",
    name: MOCK_PRIORITY_NAMES.VERY_LOW,
    value: 25,
    createdAt: "0001-01-01T00:00:00Z",
    updatedAt: "0001-01-01T00:00:00Z",
  },
  {
    id: MOCK_PRIORITY_IDS.MEDIUM_HIGH,
    source: "Run.ai",
    name: MOCK_PRIORITY_NAMES.MEDIUM_HIGH,
    value: 90,
    createdAt: "0001-01-01T00:00:00Z",
    updatedAt: "0001-01-01T00:00:00Z",
  },
  {
    id: MOCK_PRIORITY_IDS.VERY_HIGH,
    source: "Run.ai",
    name: MOCK_PRIORITY_NAMES.VERY_HIGH,
    value: 150,
    createdAt: "0001-01-01T00:00:00Z",
    updatedAt: "0001-01-01T00:00:00Z",
  },
  {
    id: MOCK_PRIORITY_IDS.MEDIUM_LOW,
    source: "Run.ai",
    name: MOCK_PRIORITY_NAMES.MEDIUM_LOW,
    value: 65,
    createdAt: "0001-01-01T00:00:00Z",
    updatedAt: "0001-01-01T00:00:00Z",
  },
  {
    id: MOCK_PRIORITY_IDS.HIGH,
    source: "Run.ai",
    name: MOCK_PRIORITY_NAMES.HIGH,
    value: 125,
    createdAt: "0001-01-01T00:00:00Z",
    updatedAt: "0001-01-01T00:00:00Z",
  },
  {
    id: MOCK_PRIORITY_IDS.MEDIUM,
    source: "Run.ai",
    name: MOCK_PRIORITY_NAMES.MEDIUM,
    value: 80,
    createdAt: "0001-01-01T00:00:00Z",
    updatedAt: "0001-01-01T00:00:00Z",
  },
];

export const mockPolicyRules = {
  editable: {
    sourceOfRule: null,
    required: false,
    canEdit: true,
    options: [
      {
        value: MOCK_PRIORITY_NAMES.VERY_LOW,
        displayed: "Trim Priority",
      },
      {
        value: MOCK_PRIORITY_NAMES.LOW,
        displayed: "Build Priority",
      },
      {
        value: MOCK_PRIORITY_NAMES.MEDIUM_LOW,
      },
      {
        value: MOCK_PRIORITY_NAMES.HIGH,
        displayed: "Train Priority",
      },
      {
        value: MOCK_PRIORITY_NAMES.VERY_HIGH,
        displayed: "Train Priority High",
      },
    ] as StringOption[],
  } as StringRules,

  nonEditable: {
    sourceOfRule: {
      scope: ScopeType.Cluster,
      clusterId: "test-cluster-id",
      isFallback: false,
    },
    required: true,
    canEdit: false,
    options: [
      {
        value: MOCK_PRIORITY_NAMES.MEDIUM,
        displayed: "Fixed Priority",
      },
    ] as StringOption[],
  } as StringRules,

  unrestricted: {
    sourceOfRule: null,
    required: false,
    canEdit: true,
    options: [],
  } as StringRules,
};

export const mockProject: Project = {
  id: "test-project-id",
  name: "test-project",
  clusterId: "test-cluster-id",
  parentId: "test-department-id",
  description: "Test project for priority section",
  enforceRunaiScheduler: false,
  status: {
    namespace: "runai-test-project",
    phase: "Ready" as ProjectPhase,
    phaseMessage: "",
    nodePoolQuotaStatuses: [
      {
        allocated: { gpu: 5, cpu: 500, memory: 512 },
        allocatedNonPreemptible: { gpu: 10, cpu: 1000, memory: 1024 },
        requested: { gpu: 5, cpu: 500, memory: 512 },
        nodePoolName: "default",
        nodePoolId: "100000",
      },
    ],
  },
  totalResources: {} as TotalResources,
  createdAt: "2023-01-01T00:00:00Z",
  updatedAt: "2023-01-01T00:00:00Z",
  parent: null,
  effective: {} as ProjectEffective,
  resources: [
    {
      nodePool: {
        id: "100000",
        name: "default",
      },
      gpu: {
        deserved: 15,
        limit: -1,
        overQuotaWeight: 5,
      },
      cpu: {
        deserved: 1500,
        limit: -1,
        overQuotaWeight: 5,
      },
      memory: {
        deserved: 1536,
        limit: -1,
        overQuotaWeight: 5,
        units: "Mib" as Units,
      },
      priority: "MediumLow",
    },
  ],
};

export const mockWorkloadPriorityData: IWorkloadPriorityData = {
  currProject: mockProject,
  selectedNodePoolsResources: ["default"],
  totalGpuRequest: 2,
  workloadPriorities: mockWorkloadPriorities,
  supportPriorityClass: true,
};

export const mockWorkloadPriorityDataPending: IWorkloadPriorityData = {
  currProject: mockProject,
  selectedNodePoolsResources: ["default"],
  totalGpuRequest: 10, // This will cause pending since quota(15) < gpuRequest(10) + allocatedNonPreemptible(10)
  workloadPriorities: mockWorkloadPriorities,
  supportPriorityClass: true,
};

export const mockDefaultFactoryProps = {
  modelValue: MOCK_PRIORITY_NAMES.MEDIUM_LOW,
  disabled: false,
  policyRules: mockPolicyRules.editable,
  workloadPriorityData: {
    ...mockWorkloadPriorityData,
    workloadPriorities: mockWorkloadPriorities,
    supportPriorityClass: true,
  },
};

export const prioritySectionMock = {
  MOCK_PRIORITY_IDS,
  MOCK_PRIORITY_NAMES,
  mockWorkloadPriorities,
  mockPolicyRules,
  mockDefaultFactoryProps,
  mockProject,
  mockWorkloadPriorityData,
  mockWorkloadPriorityDataPending,
};
