import { describe, it, expect } from "vitest";
import type { Workload, WorkloadElement, Pod, WorkloadElementResponse } from "@/swagger-models/workloads-client";
import { Phase } from "@/swagger-models/workloads-client";
import {
  buildDiagram,
  EArchitectureNodeKind,
  ENDPOINT_LEFT_HANDLE_ID,
  ENDPOINT_NODE_TYPE,
  HORIZONTAL_EDGE_TYPE,
  NODE_TYPE,
  WORKLOAD_RIGHT_HANDLE_ID,
  type ArchitectureNodeData,
  type EndpointNodeData,
  type WorkloadElementDetailsById,
} from "./workload-architecture-diagram.util";

const createWorkload = (overrides: Partial<Workload> = {}): Workload =>
  ({
    id: "workload-1",
    name: "test-workload",
    type: "Workspace",
    phase: Phase.Running,
    runningPods: 2,
    requestedPods: { number: 3 },
    allocatedResources: {
      gpu: 1,
      gpuMemory: "12582M",
      cpu: 0.5,
      cpuMemory: "512M",
    },
    ...overrides,
  }) as Workload;

const createElement = (overrides: Partial<WorkloadElement> = {}): WorkloadElement => ({
  id: "element-1",
  name: "replicaset",
  hierarchyLevel: 1,
  createdAt: "2026-03-24T12:00:00Z",
  ...overrides,
});

const createPod = (overrides: Partial<Pod> = {}): Pod =>
  ({
    id: "pod-1",
    name: "test-pod-1",
    k8sPhase: "Running",
    elements: [{ elementId: "element-1", hierarchyLevel: 1 }],
    ...overrides,
  }) as Pod;

describe("workload-architecture-diagram.util", () => {
  describe("buildDiagram", () => {
    describe("workload root node", () => {
      it("should create a root node from the workload", () => {
        const workload = createWorkload();
        const { nodes } = buildDiagram(workload, [], []);

        expect(nodes).toHaveLength(1);
        const root = nodes[0];
        expect(root.id).toBe("workload-1");
        expect(root.type).toBe(NODE_TYPE);

        const data = root.data as ArchitectureNodeData;
        expect(data.name).toBe("test-workload");
        expect(data.kind).toBe(EArchitectureNodeKind.Workload);
        expect(data.rowIndex).toBe(0);
        expect(data.parentId).toBeNull();
      });

      it("should include workload phase and type", () => {
        const workload = createWorkload({ phase: Phase.Running, type: "Training" });
        const { nodes } = buildDiagram(workload, [], []);

        const data = nodes[0].data as ArchitectureNodeData;
        expect(data.phase).toBe("Running");
        expect(data.workloadType).toBe("Training");
      });

      it("should include allocated resources", () => {
        const workload = createWorkload();
        const { nodes } = buildDiagram(workload, [], []);

        const data = nodes[0].data as ArchitectureNodeData;
        expect(data.resources).toEqual({
          gpuComputeAllocation: 1,
          gpuMemoryAllocation: "12582M",
          cpuComputeAllocation: 0.5,
          cpuMemoryAllocation: "512M",
        });
      });

      it("should include pod display using running/requested format", () => {
        const workload = createWorkload({ runningPods: 2, requestedPods: { number: 3 } });
        const { nodes } = buildDiagram(workload, [], []);

        const data = nodes[0].data as ArchitectureNodeData;
        expect(data.podsDisplay).toBe("2/3");
      });
    });

    describe("element nodes", () => {
      it("should create nodes for elements as children of workload", () => {
        const workload = createWorkload();
        const elements = [
          createElement({ id: "el-1", name: "group-0", hierarchyLevel: 1 }),
          createElement({ id: "el-2", name: "group-1", hierarchyLevel: 1 }),
        ];

        const { nodes } = buildDiagram(workload, elements, []);

        expect(nodes).toHaveLength(3);
        const elementNodes = nodes.filter(
          (n) => (n.data as ArchitectureNodeData).kind === EArchitectureNodeKind.Element,
        );
        expect(elementNodes).toHaveLength(2);
      });

      it("should set element kind from kubernetesRef", () => {
        const element = createElement({
          kubernetesRef: { uid: "uid-1", name: "rs-name", gvk: { group: "apps", version: "v1", kind: "ReplicaSet" } },
        });

        const { nodes } = buildDiagram(createWorkload(), [element], []);

        const elNode = nodes.find((n) => n.id === element.id);
        const data = elNode?.data as ArchitectureNodeData;
        expect(data.elementKind).toBe("ReplicaSet");
        expect(data.kubernetesName).toBe("rs-name");
      });

      it("should set rowIndex from hierarchyLevel", () => {
        const elements = [
          createElement({ id: "el-1", hierarchyLevel: 1 }),
          createElement({ id: "el-2", hierarchyLevel: 2, parentId: "el-1" }),
        ];

        const { nodes } = buildDiagram(createWorkload(), elements, []);

        const el1 = nodes.find((n) => n.id === "el-1");
        const el2 = nodes.find((n) => n.id === "el-2");
        expect((el1?.data as ArchitectureNodeData).rowIndex).toBe(1);
        expect((el2?.data as ArchitectureNodeData).rowIndex).toBe(2);
      });
    });

    describe("tree structure and edges", () => {
      it("should create edges from workload to level-1 elements", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];

        const { edges } = buildDiagram(createWorkload(), elements, []);

        expect(edges).toHaveLength(1);
        expect(edges[0].source).toBe("workload-1");
        expect(edges[0].target).toBe("el-1");
        expect(edges[0].type).toBe("circle");
      });

      it("should create edges from parent elements to child elements", () => {
        const elements = [
          createElement({ id: "el-1", hierarchyLevel: 1 }),
          createElement({ id: "el-2", hierarchyLevel: 2, parentId: "el-1" }),
        ];

        const { edges } = buildDiagram(createWorkload(), elements, []);

        expect(edges).toHaveLength(2);
        expect(edges).toContainEqual(expect.objectContaining({ source: "workload-1", target: "el-1" }));
        expect(edges).toContainEqual(expect.objectContaining({ source: "el-1", target: "el-2" }));
      });

      it("should handle elements without parentId as children of workload", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1, parentId: undefined })];

        const { edges } = buildDiagram(createWorkload(), elements, []);

        expect(edges[0].source).toBe("workload-1");
        expect(edges[0].target).toBe("el-1");
      });

      it("should produce no edges for workload-only diagram", () => {
        const { edges } = buildDiagram(createWorkload(), [], []);
        expect(edges).toHaveLength(0);
      });
    });

    describe("pod counting per element", () => {
      it("should count pods linked to elements", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const pods = [
          createPod({ id: "pod-1", elements: [{ elementId: "el-1", hierarchyLevel: 1 }] }),
          createPod({ id: "pod-2", elements: [{ elementId: "el-1", hierarchyLevel: 1 }] }),
        ];

        const { nodes } = buildDiagram(createWorkload(), elements, pods);

        const elNode = nodes.find((n) => n.id === "el-1");
        const data = elNode?.data as ArchitectureNodeData;
        expect(data.podsDisplay).toBe("2");
      });

      it("should assign pod to deepest parent element", () => {
        const elements = [
          createElement({ id: "el-1", hierarchyLevel: 1 }),
          createElement({ id: "el-2", hierarchyLevel: 2, parentId: "el-1" }),
        ];
        const pods = [
          createPod({
            id: "pod-1",
            elements: [
              { elementId: "el-1", hierarchyLevel: 1 },
              { elementId: "el-2", hierarchyLevel: 2 },
            ],
          }),
        ];

        const { nodes } = buildDiagram(createWorkload(), elements, pods);

        const el1 = nodes.find((n) => n.id === "el-1");
        const el2 = nodes.find((n) => n.id === "el-2");
        expect((el1?.data as ArchitectureNodeData).podsDisplay).toBeNull();
        expect((el2?.data as ArchitectureNodeData).podsDisplay).toBe("1");
      });

      it("should return null pod count for elements with no pods", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];

        const { nodes } = buildDiagram(createWorkload(), elements, []);

        const elNode = nodes.find((n) => n.id === "el-1");
        expect((elNode?.data as ArchitectureNodeData).podsDisplay).toBeNull();
      });

      it("should skip pods without elements array", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const pods = [createPod({ id: "pod-1", elements: undefined })];

        const { nodes } = buildDiagram(createWorkload(), elements, pods);

        const elNode = nodes.find((n) => n.id === "el-1");
        expect((elNode?.data as ArchitectureNodeData).podsDisplay).toBeNull();
      });
    });

    describe("top-down layout positioning", () => {
      it("should position workload root at row 0 (y=0)", () => {
        const { nodes } = buildDiagram(createWorkload(), [], []);
        expect(nodes[0].position.y).toBe(0);
      });

      it("should position level-1 elements below the workload", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const { nodes } = buildDiagram(createWorkload(), elements, []);

        const root = nodes.find((n) => n.id === "workload-1");
        const el1 = nodes.find((n) => n.id === "el-1");
        expect(el1?.position.y).toBeGreaterThan(root?.position.y ?? 0);
      });

      it("should position level-2 elements below level-1", () => {
        const elements = [
          createElement({ id: "el-1", hierarchyLevel: 1 }),
          createElement({ id: "el-2", hierarchyLevel: 2, parentId: "el-1" }),
        ];
        const { nodes } = buildDiagram(createWorkload(), elements, []);

        const el1 = nodes.find((n) => n.id === "el-1");
        const el2 = nodes.find((n) => n.id === "el-2");
        expect(el2?.position.y).toBeGreaterThan(el1?.position.y ?? 0);
      });

      it("should center parent over its children horizontally", () => {
        const elements = [
          createElement({ id: "el-1", hierarchyLevel: 1 }),
          createElement({ id: "el-2", hierarchyLevel: 1 }),
        ];
        const { nodes } = buildDiagram(createWorkload(), elements, []);

        const root = nodes.find((n) => n.id === "workload-1");
        const el1 = nodes.find((n) => n.id === "el-1");
        const el2 = nodes.find((n) => n.id === "el-2");

        if (root && el1 && el2) {
          const childrenCenter = (el1.position.x + el2.position.x + 200) / 2;
          const rootCenter = root.position.x + 100;
          expect(rootCenter).toBeCloseTo(childrenCenter, 0);
        }
      });

      it("should spread sibling elements horizontally", () => {
        const elements = [
          createElement({ id: "el-1", hierarchyLevel: 1 }),
          createElement({ id: "el-2", hierarchyLevel: 1 }),
        ];
        const { nodes } = buildDiagram(createWorkload(), elements, []);

        const el1 = nodes.find((n) => n.id === "el-1");
        const el2 = nodes.find((n) => n.id === "el-2");
        expect(el1?.position.x).not.toBe(el2?.position.x);
        expect(el1?.position.y).toBe(el2?.position.y);
      });

      it("should account for card heights when positioning rows", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const tallHeights = { "workload-1": 150 };
        const defaultHeights = {};

        const tallResult = buildDiagram(createWorkload(), elements, [], tallHeights);
        const defaultResult = buildDiagram(createWorkload(), elements, [], defaultHeights);

        const tallEl = tallResult.nodes.find((n) => n.id === "el-1");
        const defaultEl = defaultResult.nodes.find((n) => n.id === "el-1");

        expect(tallEl?.position.y).toBeGreaterThan(defaultEl?.position.y ?? 0);
      });
    });

    describe("element resource enrichment via elementDetailsById", () => {
      const createElementResponse = (
        id: string,
        allocatedResources: WorkloadElementResponse["allocatedResources"],
      ): WorkloadElementResponse =>
        ({
          id,
          name: `name-${id}`,
          hierarchyLevel: 1,
          createdAt: "2026-04-25T00:00:00Z",
          allocatedResources,
        }) as unknown as WorkloadElementResponse;

      it("should map allocatedResources from elementDetailsById onto element nodes", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const elementDetailsById: WorkloadElementDetailsById = {
          "el-1": createElementResponse("el-1", {
            gpu: 2,
            gpuMemory: "24576M",
            cpu: 1.5,
            cpuMemory: "1024M",
          }),
        };

        const { nodes } = buildDiagram(createWorkload(), elements, [], {}, elementDetailsById);

        const data = nodes.find((n) => n.id === "el-1")?.data as ArchitectureNodeData;
        expect(data.resources).toEqual({
          gpuComputeAllocation: 2,
          gpuMemoryAllocation: "24576M",
          cpuComputeAllocation: 1.5,
          cpuMemoryAllocation: "1024M",
        });
      });

      it("should leave resources null for elements without details", () => {
        const elements = [
          createElement({ id: "el-1", hierarchyLevel: 1 }),
          createElement({ id: "el-2", hierarchyLevel: 1 }),
        ];
        const elementDetailsById: WorkloadElementDetailsById = {
          "el-1": createElementResponse("el-1", { gpu: 1, gpuMemory: null, cpu: null, cpuMemory: null }),
        };

        const { nodes } = buildDiagram(createWorkload(), elements, [], {}, elementDetailsById);

        const el1 = nodes.find((n) => n.id === "el-1")?.data as ArchitectureNodeData;
        const el2 = nodes.find((n) => n.id === "el-2")?.data as ArchitectureNodeData;

        expect(el1.resources?.gpuComputeAllocation).toBe(1);
        expect(el2.resources).toBeNull();
      });

      it("should normalize undefined resource fields to null", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const elementDetailsById: WorkloadElementDetailsById = {
          "el-1": createElementResponse("el-1", { gpu: undefined, gpuMemory: undefined, cpu: 0, cpuMemory: "0" }),
        };

        const { nodes } = buildDiagram(createWorkload(), elements, [], {}, elementDetailsById);

        const data = nodes.find((n) => n.id === "el-1")?.data as ArchitectureNodeData;
        expect(data.resources).toEqual({
          gpuComputeAllocation: null,
          gpuMemoryAllocation: null,
          cpuComputeAllocation: 0,
          cpuMemoryAllocation: "0",
        });
      });
    });

    describe("complex tree structure", () => {
      it("should handle the real-world distributed workload structure", () => {
        const workload = createWorkload({ id: "wk-root", name: "dynamo-workload" });
        const elements = [
          createElement({ id: "g0", name: "group-0", hierarchyLevel: 1 }),
          createElement({ id: "g1", name: "group-1", hierarchyLevel: 1 }),
          createElement({ id: "l0", name: "leader-0", hierarchyLevel: 2, parentId: "g0" }),
          createElement({ id: "l1", name: "leader-1", hierarchyLevel: 2, parentId: "g1" }),
          createElement({ id: "w0", name: "worker-0", hierarchyLevel: 2, parentId: "g0" }),
          createElement({ id: "w1", name: "worker-1", hierarchyLevel: 2, parentId: "g1" }),
        ];

        const { nodes, edges } = buildDiagram(workload, elements, []);

        expect(nodes).toHaveLength(7);
        expect(edges).toHaveLength(6);

        expect(edges).toContainEqual(expect.objectContaining({ source: "wk-root", target: "g0" }));
        expect(edges).toContainEqual(expect.objectContaining({ source: "wk-root", target: "g1" }));
        expect(edges).toContainEqual(expect.objectContaining({ source: "g0", target: "l0" }));
        expect(edges).toContainEqual(expect.objectContaining({ source: "g0", target: "w0" }));
        expect(edges).toContainEqual(expect.objectContaining({ source: "g1", target: "l1" }));
        expect(edges).toContainEqual(expect.objectContaining({ source: "g1", target: "w1" }));
      });

      it("should handle workload with no elements", () => {
        const { nodes, edges } = buildDiagram(createWorkload(), [], []);
        expect(nodes).toHaveLength(1);
        expect(edges).toHaveLength(0);
      });

      it("should handle single element chain", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const { nodes, edges } = buildDiagram(createWorkload(), elements, []);

        expect(nodes).toHaveLength(2);
        expect(edges).toHaveLength(1);
      });
    });

    describe("external endpoints", () => {
      it("should create one endpoint node per workload endpoint with a url", () => {
        const workload = createWorkload({
          endpoints: [
            { url: "https://inference-api.example.com/v1/completions" },
            { url: "https://inference-api.example.com/v1/embeddings" },
          ],
        });

        const { nodes } = buildDiagram(workload, [], []);

        const endpointNodes = nodes.filter((n) => n.type === ENDPOINT_NODE_TYPE);
        expect(endpointNodes).toHaveLength(2);
        expect((endpointNodes[0].data as EndpointNodeData).url).toBe("https://inference-api.example.com/v1/completions");
        expect((endpointNodes[1].data as EndpointNodeData).url).toBe("https://inference-api.example.com/v1/embeddings");
      });

      it("should connect each endpoint to the workload root via a horizontal edge with right/left handles", () => {
        const workload = createWorkload({
          endpoints: [{ url: "https://a.example.com" }, { url: "https://b.example.com" }],
        });

        const { edges } = buildDiagram(workload, [], []);

        const endpointEdges = edges.filter((e) => e.type === HORIZONTAL_EDGE_TYPE);
        expect(endpointEdges).toHaveLength(2);
        for (const edge of endpointEdges) {
          expect(edge.source).toBe("workload-1");
          expect(edge.sourceHandle).toBe(WORKLOAD_RIGHT_HANDLE_ID);
          expect(edge.targetHandle).toBe(ENDPOINT_LEFT_HANDLE_ID);
        }
        expect(endpointEdges.map((e) => e.target)).toEqual(["workload-1-endpoint-0", "workload-1-endpoint-1"]);
      });

      it("should position endpoint nodes to the right of the workload root", () => {
        const workload = createWorkload({ endpoints: [{ url: "https://a.example.com" }] });

        const { nodes } = buildDiagram(workload, [], []);

        const root = nodes.find((n) => n.id === "workload-1");
        const endpoint = nodes.find((n) => n.type === ENDPOINT_NODE_TYPE);

        expect(endpoint?.position.x).toBeGreaterThan(root?.position.x ?? 0);
        expect(endpoint?.position.y).toBe(root?.position.y);
      });

      it("should stack multiple endpoint nodes vertically", () => {
        const workload = createWorkload({
          endpoints: [{ url: "https://a.example.com" }, { url: "https://b.example.com" }],
        });

        const { nodes } = buildDiagram(workload, [], []);

        const endpoints = nodes.filter((n) => n.type === ENDPOINT_NODE_TYPE);
        expect(endpoints[1].position.y).toBeGreaterThan(endpoints[0].position.y);
        expect(endpoints[0].position.x).toBe(endpoints[1].position.x);
      });

      it("should set targetPosition to 'left' on endpoint nodes", () => {
        const workload = createWorkload({ endpoints: [{ url: "https://a.example.com" }] });

        const { nodes } = buildDiagram(workload, [], []);

        const endpoint = nodes.find((n) => n.type === ENDPOINT_NODE_TYPE);
        expect(endpoint?.targetPosition).toBe("left");
      });

      it("should skip endpoints without a url", () => {
        const workload = createWorkload({
          endpoints: [{ url: "https://valid.example.com" }, { host: "no-url.example.com" }, { url: "" }],
        });

        const { nodes } = buildDiagram(workload, [], []);

        const endpoints = nodes.filter((n) => n.type === ENDPOINT_NODE_TYPE);
        expect(endpoints).toHaveLength(1);
        expect((endpoints[0].data as EndpointNodeData).url).toBe("https://valid.example.com");
      });

      it("should not create endpoint nodes when endpoints is undefined or empty", () => {
        const noEndpoints = buildDiagram(createWorkload(), [], []);
        expect(noEndpoints.nodes.filter((n) => n.type === ENDPOINT_NODE_TYPE)).toHaveLength(0);
        expect(noEndpoints.edges.filter((e) => e.type === HORIZONTAL_EDGE_TYPE)).toHaveLength(0);

        const emptyEndpoints = buildDiagram(createWorkload({ endpoints: [] }), [], []);
        expect(emptyEndpoints.nodes.filter((n) => n.type === ENDPOINT_NODE_TYPE)).toHaveLength(0);
      });

      it("should not affect existing element nodes or edges", () => {
        const elements = [createElement({ id: "el-1", hierarchyLevel: 1 })];
        const workload = createWorkload({ endpoints: [{ url: "https://a.example.com" }] });

        const { nodes, edges } = buildDiagram(workload, elements, []);

        const elementNodes = nodes.filter((n) => n.type === NODE_TYPE);
        expect(elementNodes).toHaveLength(2);
        const childEdges = edges.filter((e) => e.type === "circle");
        expect(childEdges).toHaveLength(1);
        expect(childEdges[0].source).toBe("workload-1");
        expect(childEdges[0].target).toBe("el-1");
      });
    });
  });
});
