import type { Project, ProjectEffective } from "@/swagger-models/org-unit-service-client";

export const projectsMock: Project[] = [
  {
    description: "",
    enforceRunaiScheduler: false,
    effective: {} as ProjectEffective,
    defaultNodePools: ["default"],
    nodeTypes: { names: {} },
    resources: [
      {
        nodePool: { id: "3", name: "default" },
        gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
        cpu: { limit: -1 },
        memory: { limit: -1, units: "Mib" },
      },
    ],
    name: "suricata",
    clusterId: "5852ddb1-e94f-4879-9e35-522bf87bda8f",
    id: "4500031",
    parentId: "4500015",
    requestedNamespace: "",
    status: {
      quotaStatus: { allocated: {}, allocatedNonPreemptible: {}, requested: {} },
      additionalStatusData: { conditions: null },
    },
    totalResources: { gpuQuota: 0 },
    createdAt: "2024-08-29T07:51:46.682408Z",
    updatedAt: "2024-08-29T07:51:46.688687Z",
    createdBy: "test@run.ai",
    parent: { id: "4500015", name: "default" },
  },
];
