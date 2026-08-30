import { describe, expect, it, vi, beforeEach } from "vitest";
import { allUsersColumnsMap, usersIndexColumns } from "./user.table-model";
import { ECustomCell } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import type { IUser } from "@/models/user.model";
import { exportFile } from "quasar";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("user.table-model - accessRules column (viewLinkColumn)", () => {
  const column = allUsersColumnsMap.accessRules;

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

  it('emits "access-rules-clicked" when clicked', () => {
    expect(column.customCellEvent).toEqual({ emitName: "access-rules-clicked" });
  });
});

describe("user.table-model - CSV export flow (usersIndexColumns via tableUtil.exportTableAsCsv)", () => {
  const mockedExportFile = vi.mocked(exportFile);

  beforeEach(() => {
    mockedExportFile.mockClear();
  });

  const user = { username: "user-1" } as IUser;

  const getCsvHeader = (): string[] => {
    const call = mockedExportFile.mock.calls[0];
    if (!call) throw new Error("exportFile was not called");
    const [headerLine] = (call[1] as string).split("\r\n");
    return headerLine.split(",");
  };

  it("drops the accessRules drill-in column from the exported CSV", () => {
    tableUtil.exportTableAsCsv("users", [user], usersIndexColumns);

    const header = getCsvHeader();
    expect(header).not.toContain("Access rule(s)");
  });

  it("keeps the other columns in the exported CSV", () => {
    tableUtil.exportTableAsCsv("users", [user], usersIndexColumns);

    const header = getCsvHeader();
    expect(header).toContain("User");
  });
});
