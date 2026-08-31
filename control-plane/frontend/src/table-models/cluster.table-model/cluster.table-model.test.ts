/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi, beforeEach } from "vitest";
import { allClusterColumnsMap, clusterIndexColumns } from "./cluster.table-model";
import { ECustomCell } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import { ClusterState, type DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { exportFile } from "quasar";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("cluster.table-model - topologies column (viewLinkColumn)", () => {
  const column = allClusterColumnsMap.topologies;

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

  it('emits "topologies-clicked" when clicked', () => {
    expect(column.customCellEvent).toEqual({ emitName: "topologies-clicked" });
  });
});

describe("cluster.table-model - CSV export flow (clusterIndexColumns via tableUtil.exportTableAsCsv)", () => {
  const mockedExportFile = vi.mocked(exportFile);

  beforeEach(() => {
    mockedExportFile.mockClear();
  });

  const cluster: DisplayedCluster = {
    uuid: "cluster-1",
    name: "cluster-1",
    domain: "cluster-1.example.com",
    version: "2.20.0",
    createdAt: "2024-01-01T00:00:00Z",
    status: { state: ClusterState.Connected },
  } as DisplayedCluster;

  const getCsvHeader = (): string[] => {
    const call = mockedExportFile.mock.calls[0];
    if (!call) throw new Error("exportFile was not called");
    const [headerLine] = (call[1] as string).split("\r\n");
    return headerLine.split(",");
  };

  it("drops the topologies drill-in column from the exported CSV", () => {
    tableUtil.exportTableAsCsv("clusters", [cluster], clusterIndexColumns);

    const header = getCsvHeader();
    expect(header).not.toContain("Network topologies");
  });

  it("keeps the other columns in the exported CSV", () => {
    tableUtil.exportTableAsCsv("clusters", [cluster], clusterIndexColumns);

    const header = getCsvHeader();
    expect(header).toContain("Cluster");
    expect(header).toContain("Run:ai cluster UUID");
  });
});
