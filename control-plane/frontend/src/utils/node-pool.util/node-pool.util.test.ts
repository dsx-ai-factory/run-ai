import { NodepoolPhase, type TelemetryResponse } from "@/swagger-models/cluster-service-client";
import { describe, expect, it, vi } from "vitest";
import type { INodePoolTable } from "@/table-models/node-pool.table-model";
import type { INodeTable } from "@/table-models/node.table-model";

vi.mock("@/utils/docs.util/docs.util", () => ({
  docsUtil: {
    getNumaAwareSchedulingHTMLLink: () =>
      'For more information see the <a href="https://docs.example.test/numa">Configuring NUMA-Aware Scheduling guide.</a>',
  },
}));

const { nodePoolUtil } = await import("@/utils/node-pool.util/node-pool.util");

describe("node-pool.util", () => {
  describe("getEmptyNodePool", () => {
    it("should not set gpuResourceOptimization", () => {
      expect(nodePoolUtil.getEmptyNodePool("cluster-id")).not.toHaveProperty("gpuResourceOptimization");
    });
  });

  describe("getStatusColOptions", () => {
    it("appends the NUMA guide link to an NRT-health missing-prerequisites message", () => {
      const nodePool = {
        phase: NodepoolPhase.MissingPrerequisites,
        phaseMessage:
          "The following prerequisites are not met on one or more nodes in the node pool: " +
          "NodeResourceTopology custom resource missing, Kubelet Topology Manager Policy misconfigured. " +
          "As a result, NUMA-aware scheduling may be impacted.",
      } as INodePoolTable;

      const result = nodePoolUtil.getStatusColOptions(nodePool);

      expect(result.status).toBe("Missing Prerequisites");
      expect(result.tooltipText).toBe(
        "The following prerequisites are not met on one or more nodes in the node pool: " +
          "NodeResourceTopology custom resource missing, Kubelet Topology Manager Policy misconfigured. " +
          "As a result, NUMA-aware scheduling may be impacted. " +
          'For more information see the <a href="https://docs.example.test/numa">Configuring NUMA-Aware Scheduling guide.</a>',
      );
    });

    it("still applies existing phase-message rewrites for unrelated messages", () => {
      const nodePool = {
        phase: NodepoolPhase.Unschedulable,
        phaseMessage: "Unschedulable Nodes: node-1, node-2",
      } as INodePoolTable;

      const result = nodePoolUtil.getStatusColOptions(nodePool);

      expect(result.tooltipText).toBe("The following nodes within the node pool aren't ready:<br> node-1, node-2");
    });
  });

  describe("enrichNodePoolsWithTelemetry", () => {
    it("should enrich node pools with telemetry data", () => {
      const nodePools = [
        {
          name: "default",
          clusterId: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
        },
        {
          name: "nodepool2",
          clusterId: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
        },
      ];

      const telemetryData: TelemetryResponse[] = [
        {
          type: "TOTAL_GPUS",
          timestamp: "2024-10-30T12:26:37.136768251Z",
          values: [
            {
              value: "64",
              groups: [
                {
                  key: "Nodepool",
                  value: "default",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
            {
              value: "65",
              groups: [
                {
                  key: "Nodepool",
                  value: "nodepool2",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
                },
              ],
            },
            {
              value: "66",
              groups: [
                {
                  key: "Nodepool",
                  value: "nodepool3",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
                },
              ],
            },
          ],
        },
        {
          type: "TOTAL_GPU_MEMORY_BYTES",
          timestamp: "2024-10-30T12:26:37.132056172Z",
          values: [
            {
              value: "805306368000",
              groups: [
                {
                  key: "Nodepool",
                  value: "default",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "ALLOCATED_GPUS",
          timestamp: "2024-10-30T12:26:37.134838955Z",
          values: [
            {
              value: "2",
              groups: [
                {
                  key: "Nodepool",
                  value: "default",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "TOTAL_CPU_CORES",
          timestamp: "2024-10-30T12:26:37.136717561Z",
          values: [
            {
              value: "40",
              groups: [
                {
                  key: "Nodepool",
                  value: "default",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "TOTAL_CPU_MEMORY_BYTES",
          timestamp: "2024-10-30T12:26:37.561343179Z",
          values: [
            {
              value: "166319124480",
              groups: [
                {
                  key: "Nodepool",
                  value: "default",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "ALLOCATED_CPU_CORES",
          timestamp: "2024-10-30T12:26:37.604667157Z",
          values: [
            {
              value: "21.982000000000003",
              groups: [
                {
                  key: "Nodepool",
                  value: "default",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "ALLOCATED_CPU_MEMORY_BYTES",
          timestamp: "2024-10-30T12:26:37.609337253Z",
          values: [
            {
              value: "21.982000000000003",
              groups: [
                {
                  key: "Nodepool",
                  value: "default",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
      ];

      const expected = [
        {
          name: "default",
          clusterId: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
          additionalFields: {
            totalGpus: 64,
            totalGpuMemoryBytes: 805306368000,
            allocatedGpus: 2,
            totalCpuCores: 40,
            totalCpuMemoryBytes: 166319124480,
            allocatedCpuCores: 21.982000000000003,
            allocatedCpuMemoryBytes: 21.982000000000003,
          },
        },
        {
          name: "nodepool2",
          clusterId: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
          additionalFields: {
            totalGpus: 65,
          },
        },
      ];

      const result = nodePoolUtil.enrichNodePoolsWithTelemetry(nodePools as INodePoolTable[], telemetryData);
      expect(result).toEqual(expected);
    });
    it("should return empty array if nodePools is empty", () => {
      const nodePools: INodePoolTable[] = [];
      const telemetryData: TelemetryResponse[] = [];
      const expected: INodePoolTable[] = [];
      const result = nodePoolUtil.enrichNodePoolsWithTelemetry(nodePools, telemetryData);
      expect(result).toEqual(expected);
    });
    it("should not mutate node pools when telemetry data is empty", () => {
      const nodePools = [
        { name: "nodepool1", clusterId: "cluster1" },
        { name: "nodepool2", clusterId: "cluster2" },
      ];

      const telemetryData: TelemetryResponse[] = [];
      const expected = [
        { name: "nodepool1", clusterId: "cluster1" },
        { name: "nodepool2", clusterId: "cluster2" },
      ];

      const result = nodePoolUtil.enrichNodePoolsWithTelemetry(nodePools as INodePoolTable[], telemetryData);
      expect(result).toEqual(expected);
    });
    it("should not mutate the original node pools", () => {
      const nodePools = [
        { name: "nodepool1", clusterId: "cluster1" },
        { name: "nodepool2", clusterId: "cluster2" },
      ];

      const telemetryData: TelemetryResponse[] = [
        {
          type: "TOTAL_GPUS",
          timestamp: "2024-10-30T12:26:37.136768251Z",
          values: [
            {
              value: "64",
              groups: [
                {
                  key: "Nodepool",
                  value: "nodepool1",
                },
                {
                  key: "ClusterId",
                  value: "cluster1",
                },
              ],
            },
          ],
        },
      ];

      nodePoolUtil.enrichNodePoolsWithTelemetry(nodePools as INodePoolTable[], telemetryData);
      expect(nodePools).toEqual([
        { name: "nodepool1", clusterId: "cluster1" },
        { name: "nodepool2", clusterId: "cluster2" },
      ]);
    });
  });
  describe("enrichNodesWithTelemetry", () => {
    it("should enrich nodes with telemetry data", () => {
      const nodes = [
        {
          name: "runai-cluster-worker-1",
          clusterUuid: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
        },
        {
          name: "runai-cluster-worker-2",
          clusterUuid: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
        },
      ];

      const telemetryData: TelemetryResponse[] = [
        {
          type: "TOTAL_GPUS",
          timestamp: "2024-10-30T12:26:37.136768251Z",
          values: [
            {
              value: "64",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
            {
              value: "65",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-2",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
                },
              ],
            },
            {
              value: "66",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
                },
              ],
            },
          ],
        },
        {
          type: "TOTAL_GPU_MEMORY_BYTES",
          timestamp: "2024-10-30T12:26:37.132056172Z",
          values: [
            {
              value: "805306368000",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "ALLOCATED_GPUS",
          timestamp: "2024-10-30T12:26:37.134838955Z",
          values: [
            {
              value: "2",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "TOTAL_CPU_CORES",
          timestamp: "2024-10-30T12:26:37.136717561Z",
          values: [
            {
              value: "40",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "TOTAL_CPU_MEMORY_BYTES",
          timestamp: "2024-10-30T12:26:37.561343179Z",
          values: [
            {
              value: "166319124480",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "ALLOCATED_CPU_CORES",
          timestamp: "2024-10-30T12:26:37.604667157Z",
          values: [
            {
              value: "21.982000000000003",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
        {
          type: "ALLOCATED_CPU_MEMORY_BYTES",
          timestamp: "2024-10-30T12:26:37.609337253Z",
          values: [
            {
              value: "21.982000000000003",
              groups: [
                {
                  key: "Node",
                  value: "runai-cluster-worker-1",
                },
                {
                  key: "ClusterId",
                  value: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
                },
              ],
            },
          ],
        },
      ];

      const expected = [
        {
          name: "runai-cluster-worker-1",
          clusterUuid: "9929b131-bba2-49b9-ad3a-42817fa8bdf4",
          additionalFields: {
            totalGpus: 64,
            totalGpuMemoryBytes: 805306368000,
            allocatedGpus: 2,
            totalCpuCores: 40,
            totalCpuMemoryBytes: 166319124480,
            allocatedCpuCores: 21.982000000000003,
            allocatedCpuMemoryBytes: 21.982000000000003,
          },
        },
        {
          name: "runai-cluster-worker-2",
          clusterUuid: "9929b131-bba2-49b9-ad3a-42817fa8bdf5",
          additionalFields: {
            totalGpus: 65,
          },
        },
      ];

      const result = nodePoolUtil.enrichNodesWithTelemetry(nodes as INodeTable[], telemetryData);
      expect(result).toEqual(expected);
    });
    it("should return empty array if nodePools is empty", () => {
      const nodes: INodeTable[] = [];
      const telemetryData: TelemetryResponse[] = [];
      const expected: INodePoolTable[] = [];
      const result = nodePoolUtil.enrichNodesWithTelemetry(nodes, telemetryData);
      expect(result).toEqual(expected);
    });
    it("should not mutate node pools when telemetry data is empty", () => {
      const nodes = [
        { name: "nodepool1", clusterUuid: "cluster1" },
        { name: "nodepool2", clusterUuid: "cluster2" },
      ];

      const telemetryData: TelemetryResponse[] = [];
      const expected = [
        { name: "nodepool1", clusterUuid: "cluster1" },
        { name: "nodepool2", clusterUuid: "cluster2" },
      ];

      const result = nodePoolUtil.enrichNodesWithTelemetry(nodes as INodeTable[], telemetryData);
      expect(result).toEqual(expected);
    });
  });
});
