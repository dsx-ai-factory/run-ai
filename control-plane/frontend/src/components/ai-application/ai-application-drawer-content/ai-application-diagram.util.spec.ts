import { describe, expect, it } from "vitest";
import type {
  AIApplicationResource,
  AiApplicationEndpoint,
  AiApplicationResourceConnection,
  Workload,
} from "@/swagger-models/workloads-client";
import { ResourceRefKind } from "@/swagger-models/workloads-client";
import {
  CONNECTION_EDGE_TYPE,
  ENDPOINT_NODE_TYPE,
  EResourceKind,
  ISLAND_NODE_TYPE,
  buildApplicationGraph,
  buildDiagram,
  createWorkloadMap,
  filterApplicationGraph,
  getGraphStructureFingerprint,
} from "./ai-application-diagram.util";

const resources: AIApplicationResource[] = [
  { id: "w1", name: "workload-1", kind: "ExternalWorkload" },
  { id: "w2", name: "workload-2", kind: "Workload" },
  { id: "pvc-uid", name: "shared-pvc", kind: "PersistentVolumeClaim" },
  { id: "orphan-uid", name: "orphan-config", kind: "ConfigMap" },
];

const connections: AiApplicationResourceConnection[] = [
  { workloadId: "w1", uid: "pvc-uid", name: "shared-pvc", kind: ResourceRefKind.PersistentVolumeClaim },
  { workloadId: "w2", uid: "pvc-uid", name: "shared-pvc", kind: ResourceRefKind.PersistentVolumeClaim },
];

const endpoints: AiApplicationEndpoint[] = [
  { workloadId: "w1", url: "https://example.com" },
  { workloadId: "w1", url: "https://example.com" },
  { workloadId: "w2", host: "model.local", port: 8080, protocol: "HTTP", path: "/v1" },
];

const workloads = [
  { id: "w1", name: "workload-1", phase: "Running" },
  { id: "w2", name: "workload-2", phase: "Pending" },
] as Workload[];

describe("ai-application-diagram.util", () => {
  describe("createWorkloadMap", () => {
    it("maps workloads by id and skips workloads without one", () => {
      const map = createWorkloadMap([...workloads, { name: "missing-id" } as Workload]);

      expect(map.size).toBe(2);
      expect(map.get("w1")?.phase).toBe("Running");
    });
  });

  describe("buildApplicationGraph", () => {
    it("builds one island per workload with workload-specific resource nodes", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);

      expect(graph.islands).toHaveLength(2);
      expect(graph.islands[0].resources[0].resource.id).toBe("pvc-uid");
      expect(graph.islands[1].resources[0].resource.id).toBe("pvc-uid");
      expect(graph.islands[0].resources[0].nodeId).not.toBe(graph.islands[1].resources[0].nodeId);
    });

    it("creates distinct stable nodes for duplicate endpoint URLs", () => {
      const first = buildApplicationGraph(resources, connections, endpoints, workloads);
      const second = buildApplicationGraph(resources, connections, endpoints, workloads);
      const firstEndpointIds = first.islands[0].endpoints.map((endpoint) => endpoint.nodeId);

      expect(new Set(firstEndpointIds).size).toBe(2);
      expect(firstEndpointIds).toEqual(second.islands[0].endpoints.map((endpoint) => endpoint.nodeId));
    });

    it("derives a readable endpoint label when url is absent", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);

      expect(graph.islands[1].endpoints[0].label).toBe("http://model.local:8080/v1");
    });

    it("places unreferenced storage resources in the unconnected island", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);

      expect(graph.unconnectedResources).toHaveLength(1);
      expect(graph.unconnectedResources[0].resource.name).toBe("orphan-config");
    });

    it("falls back to connection data when the resource inventory is missing", () => {
      const missingConnection: AiApplicationResourceConnection = {
        workloadId: "w1",
        uid: "missing-uid",
        name: "missing-secret",
        kind: ResourceRefKind.Secret,
      };

      const graph = buildApplicationGraph(resources, [missingConnection], [], workloads);

      expect(graph.islands[0].resources[0].resource).toEqual({
        id: "missing-uid",
        name: "missing-secret",
        kind: ResourceRefKind.Secret,
      });
    });

    it("creates a workload island from fetched workload data when inventory is incomplete", () => {
      const graph = buildApplicationGraph([], [], [{ workloadId: "w1", url: "https://example.com" }], [workloads[0]]);

      expect(graph.islands).toHaveLength(1);
      expect(graph.islands[0].workloadResource.id).toBe("w1");
      expect(graph.islands[0].endpoints).toHaveLength(1);
    });
  });

  describe("getGraphStructureFingerprint", () => {
    it("stays stable when polled workload status changes without structural changes", () => {
      const runningGraph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const updatedWorkloads = workloads.map((workload) => ({ ...workload, phase: "Completed" })) as Workload[];
      const completedGraph = buildApplicationGraph(resources, connections, endpoints, updatedWorkloads);

      expect(getGraphStructureFingerprint(completedGraph)).toBe(getGraphStructureFingerprint(runningGraph));
    });

    it("changes when graph resources or endpoints change", () => {
      const initialGraph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const updatedGraph = buildApplicationGraph(
        resources,
        connections,
        [...endpoints, { workloadId: "w2", url: "https://new.example.com" }],
        workloads,
      );

      expect(getGraphStructureFingerprint(updatedGraph)).not.toBe(getGraphStructureFingerprint(initialGraph));
    });
  });

  describe("filterApplicationGraph", () => {
    const allKinds = [
      EResourceKind.Workload,
      EResourceKind.Endpoint,
      EResourceKind.Secret,
      EResourceKind.PersistentVolumeClaim,
      EResourceKind.ConfigMap,
    ];

    it("keeps workload anchors and only children of selected types", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const filtered = filterApplicationGraph(graph, {
        selectedKinds: [EResourceKind.Endpoint],
        searchTerm: "",
      });

      expect(filtered.islands).toHaveLength(2);
      expect(filtered.islands.every((island) => island.resources.length === 0)).toBe(true);
      expect(filtered.islands.flatMap((island) => island.endpoints)).toHaveLength(3);
      expect(filtered.unconnectedResources).toEqual([]);
    });

    it("shows the matching child with its workload anchor", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const filtered = filterApplicationGraph(graph, {
        selectedKinds: allKinds,
        searchTerm: "model.local",
      });

      expect(filtered.islands).toHaveLength(1);
      expect(filtered.islands[0].workloadResource.id).toBe("w2");
      expect(filtered.islands[0].resources).toEqual([]);
      expect(filtered.islands[0].endpoints).toHaveLength(1);
    });

    it("shows selected children when the workload name matches", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const filtered = filterApplicationGraph(graph, {
        selectedKinds: allKinds,
        searchTerm: "workload-1",
      });

      expect(filtered.islands).toHaveLength(1);
      expect(filtered.islands[0].resources).toHaveLength(1);
      expect(filtered.islands[0].endpoints).toHaveLength(2);
    });

    it("searches unconnected resources and returns an empty graph for no matches", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const orphanResult = filterApplicationGraph(graph, {
        selectedKinds: allKinds,
        searchTerm: "orphan-config",
      });
      const emptyResult = filterApplicationGraph(graph, {
        selectedKinds: allKinds,
        searchTerm: "does-not-exist",
      });

      expect(orphanResult.islands).toEqual([]);
      expect(orphanResult.unconnectedResources).toHaveLength(1);
      expect(emptyResult).toEqual({ islands: [], unconnectedResources: [] });
    });
  });

  describe("buildDiagram", () => {
    it("builds workload islands with resource, endpoint nodes and relationship edges", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const result = buildDiagram(graph);

      expect(result.nodes).toHaveLength(10);
      expect(result.nodes.filter((node) => node.type === ISLAND_NODE_TYPE)).toHaveLength(2);
      expect(result.nodes.filter((node) => node.type === EResourceKind.Workload)).toHaveLength(2);
      expect(result.nodes.filter((node) => node.type === ENDPOINT_NODE_TYPE)).toHaveLength(3);
      expect(result.edges).toHaveLength(5);
      expect(result.edges.every((edge) => edge.type === CONNECTION_EDGE_TYPE)).toBe(true);
    });

    it("positions endpoints above and resources below their workload", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const result = buildDiagram(graph);
      const workloadNode = result.nodes.find((node) => node.data?.workloadId === "w1");
      const endpointNodes = result.nodes.filter((node) => node.id.startsWith("w1:endpoint:"));
      const resourceNodes = result.nodes.filter((node) => node.id.startsWith("w1:resource:"));

      expect(workloadNode).toBeDefined();
      expect(endpointNodes.every((node) => node.position.y < (workloadNode?.position.y || 0))).toBe(true);
      expect(resourceNodes.every((node) => node.position.y > (workloadNode?.position.y || 0))).toBe(true);
      expect(new Set(endpointNodes.map((node) => node.position.y)).size).toBe(1);
      expect(new Set(resourceNodes.map((node) => node.position.y)).size).toBe(1);
      expect(new Set(endpointNodes.map((node) => node.position.x)).size).toBe(endpointNodes.length);
      expect(new Set(resourceNodes.map((node) => node.position.x)).size).toBe(resourceNodes.length);
    });

    it("centers a single connected resource under its workload", () => {
      const workload = { id: "single-child", name: "single-child" } as Workload;
      const graph = buildApplicationGraph(
        [
          { id: workload.id, name: workload.name, kind: "Workload" },
          { id: "secret-uid", name: "secret", kind: "Secret" },
        ],
        [
          {
            workloadId: workload.id,
            uid: "secret-uid",
            name: "secret",
            kind: ResourceRefKind.Secret,
          },
        ],
        [],
        [workload],
      );
      const result = buildDiagram(graph);
      const workloadNode = result.nodes.find((node) => node.type === EResourceKind.Workload);
      const resourceNode = result.nodes.find((node) => node.type === EResourceKind.Secret);

      expect(resourceNode?.position.x).toBe(workloadNode?.position.x);
    });

    it("positions workload groups next to each other", () => {
      const graph = buildApplicationGraph(resources, connections, endpoints, workloads);
      const result = buildDiagram(graph, {}, 2000);
      const firstWorkload = result.nodes.find((node) => node.data?.workloadId === "w1");
      const secondWorkload = result.nodes.find((node) => node.data?.workloadId === "w2");

      expect(secondWorkload?.position.x).toBeGreaterThan(firstWorkload?.position.x || 0);
      expect(secondWorkload?.position.y).toBe(firstWorkload?.position.y);
    });

    it("wraps workload islands into responsive rows", () => {
      const workloadIds = ["w1", "w2", "w3", "w4", "w5"];
      const workloadResources = workloadIds.map((id) => ({
        id,
        name: id,
        kind: "Workload",
      }));
      const workloadEndpoints = workloadIds.flatMap((workloadId) =>
        [1, 2, 3].map((index) => ({
          workloadId,
          url: `https://${workloadId}-${index}.example.com`,
        })),
      );
      const workloadDetails = workloadResources.map(({ id, name }) => ({ id, name }) as Workload);
      const graph = buildApplicationGraph(workloadResources, [], workloadEndpoints, workloadDetails);
      const result = buildDiagram(graph, {}, 700);
      const firstWorkload = result.nodes.find((node) => node.data?.workloadId === "w1");
      const fifthWorkload = result.nodes.find((node) => node.data?.workloadId === "w5");

      expect(fifthWorkload?.position.y).toBeGreaterThan(firstWorkload?.position.y || 0);
    });

    it("renders a resource card under every connected workload", () => {
      const workloadIds = ["w1", "w2", "w3", "w4", "w5"];
      const workloadResources: AIApplicationResource[] = [
        ...workloadIds.map((id) => ({ id, name: id, kind: "Workload" })),
        { id: "secret-uid", name: "shared-secret", kind: "Secret" },
      ];
      const sharedConnections: AiApplicationResourceConnection[] = workloadIds.map((workloadId) => ({
        workloadId,
        uid: "secret-uid",
        name: "shared-secret",
        kind: ResourceRefKind.Secret,
      }));
      const workloadDetails = workloadIds.map((id) => ({ id, name: id }) as Workload);
      const graph = buildApplicationGraph(workloadResources, sharedConnections, [], workloadDetails);
      const result = buildDiagram(graph);
      const sharedResourceNodes = result.nodes.filter((node) => node.data?.resourceId === "secret-uid");
      const sharedResourceEdges = result.edges.filter((edge) => edge.target.includes("resource:secret-uid"));

      expect(sharedResourceNodes).toHaveLength(5);
      expect(sharedResourceEdges).toHaveLength(5);
    });

    it("does not create edges for unconnected resources", () => {
      const graph = buildApplicationGraph([{ id: "orphan", name: "orphan-secret", kind: "Secret" }], [], [], []);
      const result = buildDiagram(graph);

      expect(result.edges).toEqual([]);
      expect(result.nodes.some((node) => node.type === ISLAND_NODE_TYPE)).toBe(false);
      expect(result.nodes.some((node) => node.id.startsWith("unconnected:resource:orphan"))).toBe(true);
    });

    it("renders a workload without connections as a standalone card", () => {
      const workload = { id: "standalone", name: "standalone-workload" } as Workload;
      const graph = buildApplicationGraph(
        [{ id: workload.id, name: workload.name, kind: "Workload" }],
        [],
        [],
        [workload],
      );
      const result = buildDiagram(graph);

      expect(result.nodes.filter((node) => node.type === ISLAND_NODE_TYPE)).toEqual([]);
      expect(result.nodes.filter((node) => node.type === EResourceKind.Workload)).toHaveLength(1);
    });
  });
});
