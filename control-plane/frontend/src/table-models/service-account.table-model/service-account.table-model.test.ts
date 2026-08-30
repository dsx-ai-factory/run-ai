import { describe, expect, it, vi, beforeEach } from "vitest";
import { allServiceAccountsColumnsMap, serviceAccountsIndexColumns } from "./service-account.table-model";
import { ECustomCell } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";
import { exportFile } from "quasar";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("service-account.table-model - accessRules column (viewLinkColumn)", () => {
  const column = allServiceAccountsColumnsMap.accessRules;

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

  it("is no longer sortable (fixes sorting on the constant 'View' value)", () => {
    expect(column.sortable).toBe(false);
  });
});

describe("service-account.table-model - CSV export flow (serviceAccountsIndexColumns via tableUtil.exportTableAsCsv)", () => {
  const mockedExportFile = vi.mocked(exportFile);

  beforeEach(() => {
    mockedExportFile.mockClear();
  });

  const serviceAccount = { name: "service-account-1" } as ServiceAccount;

  const getCsvHeader = (): string[] => {
    const call = mockedExportFile.mock.calls[0];
    if (!call) throw new Error("exportFile was not called");
    const [headerLine] = (call[1] as string).split("\r\n");
    return headerLine.split(",");
  };

  it("drops the accessRules drill-in column from the exported CSV (previously exported the literal 'View' string)", () => {
    tableUtil.exportTableAsCsv("service-accounts", [serviceAccount], serviceAccountsIndexColumns);

    const header = getCsvHeader();
    expect(header).not.toContain("Access rule(s)");
  });

  it("keeps the other columns in the exported CSV", () => {
    tableUtil.exportTableAsCsv("service-accounts", [serviceAccount], serviceAccountsIndexColumns);

    const header = getCsvHeader();
    expect(header).toContain("Service account");
  });
});
