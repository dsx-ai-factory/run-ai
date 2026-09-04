/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { viewLinkColumn } from "./common-columns";
import { ECustomCell } from "@/models/table.model";

describe("viewLinkColumn", () => {
  it("builds the frozen drill-in link column shape", () => {
    const column = viewLinkColumn({
      name: "numaNodes",
      label: "NUMA nodes",
      emitName: "numa-nodes-clicked",
    });

    expect(column).toMatchObject({
      name: "numaNodes",
      label: "NUMA nodes",
      sortable: false,
      align: "left",
      hideFilter: true,
      hideExport: true,
      customCell: ECustomCell.LINK_COL,
      customCellEvent: { emitName: "numa-nodes-clicked" },
    });
  });

  it("defaults the field to the literal 'View'", () => {
    const column = viewLinkColumn({ name: "workloads", label: "Workloads", emitName: "workloads-clicked" });

    expect((column.field as () => string)()).toBe("View");
  });

  it("uses the provided text instead of 'View' when given", () => {
    const column = viewLinkColumn({
      name: "accessRules",
      label: "Access rule(s)",
      emitName: "access-rules-clicked",
      text: "Manage",
    });

    expect((column.field as () => string)()).toBe("Manage");
  });

  it("passes customCellConfig through untouched", () => {
    const customCellConfig = { [ECustomCell.LINK_COL]: { tooltipText: "Click to view" } };

    const column = viewLinkColumn({
      name: "topologies",
      label: "Network topologies",
      emitName: "topologies-clicked",
      customCellConfig,
    });

    expect(column.customCellConfig).toBe(customCellConfig);
  });

  it("leaves customCellConfig undefined when not provided", () => {
    const column = viewLinkColumn({ name: "pods", label: "Pod(s)", emitName: "pods-clicked" });

    expect(column.customCellConfig).toBeUndefined();
  });

  it("does not set exportFormat — hideExport: true is the frozen CSV-export decision, not exportFormat", () => {
    const column = viewLinkColumn({ name: "pods", label: "Pod(s)", emitName: "pods-clicked" });

    expect(column.exportFormat).toBeUndefined();
    expect(column.hideExport).toBe(true);
  });
});
