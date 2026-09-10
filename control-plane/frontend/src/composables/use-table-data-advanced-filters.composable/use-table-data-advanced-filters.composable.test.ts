/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { computed, defineComponent } from "vue";
import { mount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type * as QuasarModule from "quasar";

import { useTableDataAdvancedFilters } from "./use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "../helpers/use-table-data.mapper";
import { reportIndexColumns, type IReportTable } from "@/table-models/report.table-model";

vi.mock("vue-router", () => ({
  useRoute: vi.fn().mockReturnValue({ name: "", query: {}, matched: [], params: {}, path: "", fullPath: "", hash: "" }),
  useRouter: vi.fn().mockReturnValue({ push: vi.fn(), replace: vi.fn() }),
}));

vi.mock("quasar", async (importOriginal) => ({
  ...(await importOriginal<typeof QuasarModule>()),
  useQuasar: vi.fn().mockReturnValue({ notify: vi.fn() }),
}));

vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    countReports: vi.fn().mockResolvedValue(0),
    listReports: vi.fn().mockResolvedValue([]),
  },
}));

function makeReport(id: string, name: string): IReportTable {
  return { id, name } as IReportTable;
}

const reports: IReportTable[] = [
  makeReport("id-1", "report-one"),
  makeReport("id-2", "report-two"),
  makeReport("id-3", "report-three"),
  makeReport("id-4", "report-four"),
];

type PinningApi = ReturnType<typeof useTableDataAdvancedFilters<IReportTable>>;

function mountComposable(
  rows: IReportTable[] = reports,
  getEntityById?: (id: string) => Promise<IReportTable>,
): PinningApi {
  const loadList = vi.fn().mockResolvedValue(rows);
  let api: PinningApi | null = null;

  const Host = defineComponent({
    setup() {
      api = useTableDataAdvancedFilters<IReportTable>(
        loadList,
        computed(() => reportIndexColumns),
        EAdvancedIndexPages.REPORT,
        getEntityById,
      );
      return () => null;
    },
  });

  mount(Host, { global: { plugins: [createTestingPinia({ createSpy: vi.fn })] } });
  return api as unknown as PinningApi;
}

describe("useTableDataAdvancedFilters - pinned rows", () => {
  beforeEach(() => {
    setActivePinia(createTestingPinia({ createSpy: vi.fn }));
  });

  it("pins nothing before any action is registered", async () => {
    const api = mountComposable();
    await api.refreshList();

    expect(api.pinnedRows.value).toEqual([]);
  });

  it("pins the entities registered via pinEntities", async () => {
    const api = mountComposable();
    await api.refreshList();

    api.pinEntities(["id-2", "id-4"]);

    expect(api.pinnedRows.value.map((row: IReportTable) => row.id)).toEqual(["id-2", "id-4"]);
  });

  it("keeps pinned rows in table order regardless of the order they were registered", async () => {
    const api = mountComposable();
    await api.refreshList();

    api.pinEntities(["id-4", "id-1"]);

    expect(api.pinnedRows.value.map((row: IReportTable) => row.id)).toEqual(["id-1", "id-4"]);
  });

  it("accumulates across separate pinEntities calls", async () => {
    const api = mountComposable();
    await api.refreshList();

    api.pinEntities(["id-1"]);
    api.pinEntities(["id-3"]);

    expect(api.pinnedRows.value.map((row: IReportTable) => row.id)).toEqual(["id-1", "id-3"]);
  });

  it("ignores ids that are not in the loaded table data", async () => {
    const api = mountComposable();
    await api.refreshList();

    api.pinEntities(["id-2", "does-not-exist"]);

    expect(api.pinnedRows.value.map((row: IReportTable) => row.id)).toEqual(["id-2"]);
  });

  it("does not pin rows just because of their state - only registered ids pin", async () => {
    const api = mountComposable();
    await api.refreshList();

    expect(api.tableData.value).toHaveLength(reports.length);
    expect(api.pinnedRows.value).toEqual([]);
  });

  it("clearing the last created entity also drops the pins", async () => {
    const api = mountComposable();
    await api.refreshList();
    api.pinEntities(["id-1", "id-2"]);

    api.clearLastCreated();

    expect(api.pinnedRows.value).toEqual([]);
  });

  it("changing filters drops the pins", async () => {
    const api = mountComposable();
    await api.refreshList();
    api.pinEntities(["id-1"]);

    await api.setFilterBy({ sortBy: "name" });

    expect(api.pinnedRows.value).toEqual([]);
  });

  it("keeps a pinned row that the active filter no longer returns", async () => {
    // A report leaves the filtered list the moment it starts deleting. It has to stay
    // pinned and visible, not vanish while the delete is still running.
    const deleting = { ...makeReport("id-2", "report-two"), status: "Deleting" } as IReportTable;
    const loadList = vi
      .fn()
      .mockResolvedValueOnce(reports)
      .mockResolvedValue(reports.filter((row: IReportTable) => row.id !== "id-2"));
    const getEntityById = vi.fn().mockResolvedValue(deleting);

    let api: PinningApi | null = null;
    const Host = defineComponent({
      setup() {
        api = useTableDataAdvancedFilters<IReportTable>(
          loadList,
          computed(() => reportIndexColumns),
          EAdvancedIndexPages.REPORT,
          getEntityById,
        );
        return () => null;
      },
    });
    mount(Host, { global: { plugins: [createTestingPinia({ createSpy: vi.fn })] } });
    const pinning = api as unknown as PinningApi;

    await pinning.refreshList();
    pinning.pinEntities(["id-2"]);
    await pinning.refreshList();

    expect(pinning.tableData.value.map((row: IReportTable) => row.id)).not.toContain("id-2");
    expect(pinning.pinnedRows.value.map((row: IReportTable) => row.id)).toEqual(["id-2"]);
    expect(getEntityById).toHaveBeenCalledWith("id-2");
  });

  it("unpins a row once the backend no longer has it", async () => {
    const loadList = vi
      .fn()
      .mockResolvedValueOnce(reports)
      .mockResolvedValue(reports.filter((row: IReportTable) => row.id !== "id-2"));
    const getEntityById = vi.fn().mockRejectedValue(new Error("not found"));

    let api: PinningApi | null = null;
    const Host = defineComponent({
      setup() {
        api = useTableDataAdvancedFilters<IReportTable>(
          loadList,
          computed(() => reportIndexColumns),
          EAdvancedIndexPages.REPORT,
          getEntityById,
        );
        return () => null;
      },
    });
    mount(Host, { global: { plugins: [createTestingPinia({ createSpy: vi.fn })] } });
    const pinning = api as unknown as PinningApi;

    await pinning.refreshList();
    pinning.pinEntities(["id-2"]);
    await pinning.refreshList();

    expect(pinning.pinnedRows.value).toEqual([]);
  });
});
