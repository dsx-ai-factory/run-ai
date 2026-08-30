import type { NetworkTopology } from "@/swagger-models/cluster-service-client";

export const mockNetworkTopologies: NetworkTopology[] = [
  {
    id: "topology-1",
    name: "Topology 1",
    clusterId: "123",
    levels: ["rack", "node"],
    tenantId: 1,
    createdAt: "2023-01-01T00:00:00Z",
  },
  {
    id: "topology-2",
    name: "Topology 2",
    clusterId: "123",
    levels: ["zone", "rack", "node"],
    tenantId: 1,
    createdAt: "2023-01-02T00:00:00Z",
  },
];
