/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi, beforeEach } from "vitest";
import { nodeColumnName, nodeIndexColumns, type INodeTable } from "./node.table-model";
import { ECustomCell } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import { NodeStatusEnum } from "@/swagger-models/cluster-service-client";
import { exportFile } from "quasar";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("node.table-model - viewLinkColumn-based columns", () => {
  const getColumn = (name: string) => {
    const column = nodeIndexColumns.find((col) => col.name === name);
    if (!column) throw new Error(`Column "${name}" not found in nodeIndexColumns`);
    return column;
  };

  describe.each([
    { columnName: nodeColumnName.numaNodes, emitName: "numa-nodes-clicked" },
    { columnName: nodeColumnName.pods, emitName: "pods-clicked" },
  ])("$columnName column", ({ columnName, emitName }) => {
    const column = getColumn(columnName);

    it("renders the 'View' drill-in link (unsortable, unfiltered, no CSV export)", () => {
      expect(column).toMatchObject({
        sortable: false,
        align: "left",
        hideFilter: true,
        hideExport: true,
        customCell: ECustomCell.LINK_COL,
      });
      expect((column.field as () => string)()).toBe("View");
    });

    it(`emits "${emitName}" when clicked`, () => {
      expect(column.customCellEvent).toEqual({ emitName });
    });
  });

  it("hides the numaNodes column by default (only surfaced via the numa-dependent-column gate)", () => {
    expect(getColumn(nodeColumnName.numaNodes).display).toBe(false);
  });

  it("shows the pods column by default", () => {
    expect(getColumn(nodeColumnName.pods).display).toBe(true);
  });
});

describe("node.table-model - CSV export flow (nodeIndexColumns via tableUtil.exportTableAsCsv)", () => {
  const mockedExportFile = vi.mocked(exportFile);

  beforeEach(() => {
    mockedExportFile.mockClear();
  });

  const node: INodeTable = {
    name: "node-1",
    status: NodeStatusEnum.Ready,
    nodePool: "default",
    createdAt: "2024-01-01T00:00:00Z",
    clusterUuid: "cluster-1",
    nvLinkDomainUid: "domain-1",
    nvLinkCliqueId: "clique-1",
    gpuInfo: { gpuType: "A100" },
    numa: { topologyManagerPolicy: "single-numa-node" },
    additionalFields: {
      readyGpus: 3,
      totalGpus: 4,
      totalGpuMemoryBytes: 1024,
      allocatedGpus: 2,
      freeGpus: 1,
      totalCpuCores: 32,
      totalCpuMemoryBytes: 2048,
      allocatedCpuCores: 8,
      allocatedCpuMemoryBytes: 512,
    },
  } as INodeTable;

  const getCsvRow = (): { header: string[]; row: string[] } => {
    const call = mockedExportFile.mock.calls[0];
    if (!call) throw new Error("exportFile was not called");
    const [headerLine, rowLine] = (call[1] as string).split("\r\n");
    return { header: headerLine.split(","), row: rowLine.split(",") };
  };

  it("drops the numaNodes and pods drill-in columns from the exported CSV", () => {
    tableUtil.exportTableAsCsv("nodes", [node], nodeIndexColumns);

    const { header } = getCsvRow();
    expect(header).not.toContain("NUMA nodes");
    expect(header).not.toContain("Pod(s)");
  });

  it("keeps every other column, showing the value each will actually display", () => {
    tableUtil.exportTableAsCsv("nodes", [node], nodeIndexColumns);

    const { header, row } = getCsvRow();
    const cellByLabel = (label: string): string => row[header.indexOf(label)];

    expect(header).toEqual([
      "Node",
      "Status",
      "Node Pool",
      "Cluster",
      "MNNVL Domain UID",
      "MNNVL domain clique ID",
      "Topology Manager policy",
      "GPU type",
      "Ready / total GPU devices",
      "GPU memory",
      "Allocated GPUs",
      "Fully free GPU devices",
      "CPU (Cores)",
      "CPU memory",
      "Allocated CPU (Cores)",
      "Allocated CPU memory",
    ]);

    expect(cellByLabel("Node")).toBe("node-1");
    expect(cellByLabel("Status")).toBe(NodeStatusEnum.Ready); // via exportFormat, bypassing the status-cell format
    expect(cellByLabel("Node Pool")).toBe("default");
    expect(cellByLabel("MNNVL Domain UID")).toBe("domain-1");
    expect(cellByLabel("MNNVL domain clique ID")).toBe("clique-1");
    expect(cellByLabel("Topology Manager policy")).toBe("single-numa-node");
    expect(cellByLabel("GPU type")).toBe("A100");
    expect(cellByLabel("Ready / total GPU devices")).toBe("3/4"); // via exportFormat, bypassing the LIST_COL format
    expect(cellByLabel("Allocated GPUs")).toBe("2.00"); // via tableFormatUtil.getFormattedNumberOrDefault
    expect(cellByLabel("Fully free GPU devices")).toBe("1");
    expect(cellByLabel("CPU (Cores)")).toBe("32.00"); // via tableFormatUtil.getFormattedNumberOrDefault
    expect(cellByLabel("Allocated CPU (Cores)")).toBe("8.00"); // via tableFormatUtil.getFormattedNumberOrDefault
  });
});
