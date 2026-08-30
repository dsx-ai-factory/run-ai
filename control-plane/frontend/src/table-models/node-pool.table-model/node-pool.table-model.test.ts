import { describe, expect, it, vi, beforeEach } from "vitest";
import {
  allNodePoolsColumnsMap,
  nodePoolIndexColumns,
  nodePoolColumnName,
  type INodePoolTable,
} from "./node-pool.table-model";
import { ECustomCell } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import { exportFile } from "quasar";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("node-pool.table-model - workloads column (viewLinkColumn)", () => {
  const column = allNodePoolsColumnsMap.workloads;

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

  it('emits "workloads-clicked" when clicked', () => {
    expect(column.customCellEvent).toEqual({ emitName: "workloads-clicked" });
  });

  it("preserves display: false on the base column definition", () => {
    expect(column.display).toBe(false);
  });

  it("is shown in the index columns (overridden to display: true there)", () => {
    const indexColumn = nodePoolIndexColumns.find((col) => col.name === nodePoolColumnName.workloads);
    expect(indexColumn?.display).toBe(true);
  });
});

describe("node-pool.table-model - CSV export flow (nodePoolIndexColumns via tableUtil.exportTableAsCsv)", () => {
  const mockedExportFile = vi.mocked(exportFile);

  beforeEach(() => {
    mockedExportFile.mockClear();
  });

  const nodePool = { name: "node-pool-1" } as INodePoolTable;

  const getCsvHeader = (): string[] => {
    const call = mockedExportFile.mock.calls[0];
    if (!call) throw new Error("exportFile was not called");
    const [headerLine] = (call[1] as string).split("\r\n");
    return headerLine.split(",");
  };

  it("drops the workloads drill-in column from the exported CSV (previously exported the literal 'View' string)", () => {
    tableUtil.exportTableAsCsv("node-pools", [nodePool], nodePoolIndexColumns);

    const header = getCsvHeader();
    expect(header).not.toContain("Workload(s)");
  });

  it("keeps the other columns in the exported CSV", () => {
    tableUtil.exportTableAsCsv("node-pools", [nodePool], nodePoolIndexColumns);

    const header = getCsvHeader();
    expect(header).toContain("Node pool");
  });
});
