import { fileService } from "@/services/file.service/file.service";
import { ESortOrder, type IPaginationFilter } from "@/models/filter.model";
import { describe, vi, test, expect } from "vitest";
import { tableUtil } from "@/utils/table.util";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { API_LIMIT_MAX_RECORDS } from "@/composables/helpers/use-export-file.mapper";
import type { ETableExportCsvFilesNames, ITableColumn } from "@/models/table.model";

vi.mock("@/utils/table.util");

describe("fileService", () => {
  describe("csvExporter", () => {
    test("should fetch twice ", async () => {
      const RECORDS_COUNT = 1000;
      const getRecords = vi
        .fn()
        .mockResolvedValueOnce([{ id: 1 }])
        .mockResolvedValueOnce([{ id: 2 }]);
      const getCount = vi.fn().mockResolvedValue(RECORDS_COUNT);
      const filter: IPaginationFilter = {
        filterBy: [],
        sortBy: "",
        sortOrder: ESortOrder.Asc,
        offset: 0,
        limit: 100,
        search: "",
      };
      const clusters: DisplayedCluster[] = [];
      const columns: ITableColumn[] = [];
      const csvName = "test.csv" as ETableExportCsvFilesNames;

      await fileService.csvExporter(getRecords, getCount, filter, columns, csvName, clusters);

      expect(getRecords).toHaveBeenCalledTimes(RECORDS_COUNT / API_LIMIT_MAX_RECORDS);
      expect(tableUtil.exportTableAsCsv).toHaveBeenCalledWith(csvName, [{ id: 1 }, { id: 2 }], columns);
    });

    test("exports no records when there are no records", async () => {
      const getRecords = vi.fn().mockResolvedValue({ records: [], next: null });
      const getCount = vi.fn().mockResolvedValue(0);
      const filter: IPaginationFilter = {
        filterBy: [],
        sortBy: "",
        sortOrder: ESortOrder.Asc,
        offset: 0,
        limit: 100,
        search: "",
      };
      const clusters: DisplayedCluster[] = [];
      const columns: ITableColumn[] = [];
      const csvName = "test.csv" as ETableExportCsvFilesNames;

      await fileService.csvExporter(getRecords, getCount, filter, columns, csvName, clusters);

      expect(getRecords).toHaveBeenCalledTimes(0);
      expect(tableUtil.exportTableAsCsv).toHaveBeenCalledWith(csvName, [], columns);
    });

    test("exports records when there is only one page", async () => {
      const getRecords = vi.fn().mockResolvedValue([{ id: 1 }]);
      const getCount = vi.fn().mockResolvedValue(API_LIMIT_MAX_RECORDS);
      const filter: IPaginationFilter = {
        filterBy: [],
        sortBy: "",
        sortOrder: ESortOrder.Asc,
        offset: 0,
        limit: 100,
        search: "",
      };
      const clusters: DisplayedCluster[] = [];
      const columns: ITableColumn[] = [];
      const csvName = "test.csv" as ETableExportCsvFilesNames;

      await fileService.csvExporter(getRecords, getCount, filter, columns, csvName, clusters);

      expect(getRecords).toHaveBeenCalledTimes(1);
      expect(tableUtil.exportTableAsCsv).toHaveBeenCalledWith(csvName, [{ id: 1 }], columns);
    });
  });
});
