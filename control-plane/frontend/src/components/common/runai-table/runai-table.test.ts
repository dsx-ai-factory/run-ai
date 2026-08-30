import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, beforeEach, expect, test, vi } from "vitest";
import { RunaiTable } from ".";
import { tableModels } from "./runai-table.mock";
// Models
import type {
  AssetRef,
  ComputeAsset,
  CredentialsListResponseEntry,
  DatasourceListResponseEntry,
  EnvironmentAsset,
  WorkloadTemplate,
} from "@/swagger-models/assets-service-client";
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { credentialIndexColumns } from "@/table-models/credential.table-model";
import { workloadTemplateIndexColumns } from "@/table-models/workload-template.table-model";
import { DATA_SOURCE_TYPE, type TDataSourceKinds } from "@/models/data-source.model";
// Utils
import { dateUtil } from "@/utils/date.util";
import { getMemoryUnitsForUI } from "@/utils/table-format.util";
import { previewStrings } from "@/utils/common.util";
import { credentialTextForUI, type TCredentialKinds } from "@/models/credential.model";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

const SORT_ICON = "arrow_upward";
const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);
describe("runai-table.vue", () => {
  let runaiTableCmp: VueWrapper<InstanceType<typeof RunaiTable>>;

  beforeEach(() => {
    runaiTableCmp = mount(RunaiTable, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        filterBy: {} as IFilterBy,
        loading: false as boolean,
        rows: [] as any[], //eslint-disable-line
        columns: [] as ITableColumn[],
      },
    });
  });

  function getCellsByColumnName(columnName: string): DOMWrapper<HTMLTableCellElement>[] {
    return runaiTableCmp.findAll(`td[aid="${columnName}"]`);
  }

  test("component is loading properly", () => {
    expect(runaiTableCmp.html()).toContain("runai-table");
  });

  describe("Table style behavior", () => {
    // Loading behavior
    test("table should not show the linear progress when 'loading' prop is false", async () => {
      expect(runaiTableCmp.find(".q-linear-progress").exists()).toBe(false);
    });
    test("table should show the linear progress when 'loading' prop is true", async () => {
      await runaiTableCmp.setProps({ loading: true });
      expect(runaiTableCmp.find(".q-linear-progress").exists()).toBe(true);
    });

    // Shadow behavior
    test("table should have shadow by default if the 'noShadow' prop isn't passed", async () => {
      const elTable: DOMWrapper<Element> = runaiTableCmp.find(".runai-table") as DOMWrapper<Element>;
      expect(elTable.classes()).not.toContain("q-table--flat");
    });
    test("table should not have shadow if the 'noShadow' prop is true", async () => {
      await runaiTableCmp.setProps({ noShadow: true });
      const elTable: DOMWrapper<Element> = runaiTableCmp.find(".runai-table") as DOMWrapper<Element>;
      expect(elTable.classes()).toContain("q-table--flat");
    });

    // Pagination display behavior
    test("table should show the pagination by default if the 'hidePagination' prop isn't passed", async () => {
      await runaiTableCmp.setProps({ rows: tableModels.environmentModel.rows });
      await runaiTableCmp.vm.$nextTick();
      expect(runaiTableCmp.find(".q-table__bottom .q-table__control").exists()).toBe(true);
    });
    test("table should not show the pagination by default if the 'hidePagination' prop is true", async () => {
      await runaiTableCmp.setProps({ hidePagination: true });
      expect(runaiTableCmp.find(".q-table__control").exists()).toBe(false);
    });
  });

  describe("real runai table models behavior", () => {
    function getExpectedColumns(columns: Array<ITableColumn>): Array<string> {
      const expectedColumns = columns
        .filter((column) => column.display && column.name !== "scope")
        .map((column) => column.label + SORT_ICON);

      return ["", ...expectedColumns]; //adding empty string to represent the checkbox header that display by default
    }

    describe("environment model", () => {
      beforeEach(async () => {
        runaiTableCmp.setProps({
          columns: tableModels.environmentColumns,
          rows: tableModels.environmentModel.rows,
          filterBy: tableModels.environmentModel.filterBy,
        });
      });

      test("should show the correct order of columns", () => {
        const expectedColumns: Array<string> = getExpectedColumns(tableModels.environmentColumns);
        const actualColumns: Array<string> = runaiTableCmp
          .findAll("th")
          .map((th: DOMWrapper<HTMLTableCellElement>) => th.text());
        expect(actualColumns).toEqual(expectedColumns);
      });

      describe("cells content", () => {
        test("should show the correct content in the 'environment' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("environment");
          const expectedCellsContent: Array<string> = tableModels.environmentModel.rows.map(
            (env: EnvironmentAsset) => env.meta.name,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'image' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("image");
          const expectedCellsContent: Array<string> = tableModels.environmentModel.rows.map(
            (env: EnvironmentAsset) => env.spec.image || "",
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'tools' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("tools");
          const expectedCellsContent: Array<string> = tableModels.environmentModel.rows.map((env: EnvironmentAsset) =>
            previewStrings(env.spec.connections?.map((c) => c.name) || []),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'created-at' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("created-at");
          const expectedCellsContent: Array<string> = tableModels.environmentModel.rows.map((env: EnvironmentAsset) =>
            dateUtil.dateAndTimeFormat(new Date(env.meta.createdAt)),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });
      });
    });

    describe("compute resource model", () => {
      beforeEach(async () => {
        runaiTableCmp.setProps({
          columns: tableModels.computeColumns,
          rows: tableModels.computeResourceModel.rows,
          filterBy: tableModels.computeResourceModel.filterBy,
        });
      });

      test("should show the correct order of columns", () => {
        const expectedColumns = getExpectedColumns(tableModels.computeColumns);
        const actualColumns = runaiTableCmp.findAll("th").map((th: DOMWrapper<HTMLTableCellElement>) => th.text());
        expect(actualColumns).toEqual(expectedColumns);
      });

      describe("cells content", () => {
        test("should show the correct content in the 'compute-resource' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("compute-resource");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => computeResource.meta.name,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'gpu-memory-request' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("gpu-memory-request");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => {
              const { gpuPortionRequest, gpuMemoryRequest } = computeResource.spec;
              return (
                (gpuMemoryRequest && getMemoryUnitsForUI(gpuMemoryRequest)) ||
                (gpuPortionRequest && `${gpuPortionRequest * 100}%`) ||
                "-"
              );
            },
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'cpu-cores-request' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("cpu-cores-request");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => {
              if (typeof computeResource.spec.cpuCoreRequest !== "number") return "-";
              let postfix = "cores";
              let amount = computeResource.spec.cpuCoreRequest;
              if (computeResource.spec.cpuCoreRequest < 1 && computeResource.spec.cpuCoreRequest !== 0) {
                amount *= 1000;
                postfix = "millicores";
              }
              return `${amount} ${postfix}`;
            },
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'cpu-cores-limit' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("cpu-cores-limit");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => {
              if (typeof computeResource.spec.cpuCoreLimit !== "number") return "-";
              let postfix = "cores";
              let amount = computeResource.spec.cpuCoreLimit;
              if (computeResource.spec.cpuCoreLimit < 1 && computeResource.spec.cpuCoreLimit !== 0) {
                amount *= 1000;
                postfix = "millicores";
              }
              return `${amount} ${postfix}`;
            },
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'cpu-memory-request' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("cpu-memory-request");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => {
              return getMemoryUnitsForUI(computeResource.spec.cpuMemoryRequest || "") || "-";
            },
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'cpu-memory-limit' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("cpu-memory-limit");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => {
              return getMemoryUnitsForUI(computeResource.spec.cpuMemoryLimit || "") || "-";
            },
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'created-by' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("created-by");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => computeResource.meta.createdBy,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'created-at' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("created-at");
          const expectedCellsContent: Array<string> = tableModels.computeResourceModel.rows.map(
            (computeResource: ComputeAsset) => dateUtil.dateAndTimeFormat(new Date(computeResource.meta.createdAt)),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });
      });
    });

    describe("data source model", () => {
      beforeEach(async () => {
        runaiTableCmp.setProps({
          columns: tableModels.dataSourceColumns,
          rows: tableModels.dataSourceModel.rows,
          filterBy: tableModels.dataSourceModel.filterBy,
        });
      });

      test("should show the correct order of columns", () => {
        const expectedColumns = getExpectedColumns(tableModels.dataSourceColumns);
        const actualColumns = runaiTableCmp.findAll("th").map((th) => th.text());
        expect(actualColumns).toEqual(expectedColumns);
      });

      describe("cells content", () => {
        test("should show the correct content in the 'data-source' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("data-source");
          const expectedCellsContent: Array<string> = tableModels.dataSourceModel.rows.map(
            (dataSource: DatasourceListResponseEntry) => dataSource.meta.name,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'type' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("type");
          const expectedCellsContent: Array<string> = tableModels.dataSourceModel.rows.map(
            (dataSource: DatasourceListResponseEntry) => DATA_SOURCE_TYPE[dataSource.meta.kind as TDataSourceKinds],
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'created-at' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("created-at");
          const expectedCellsContent: Array<string> = tableModels.dataSourceModel.rows.map(
            (dataSource: DatasourceListResponseEntry) => dateUtil.dateAndTimeFormat(new Date(dataSource.meta.createdAt)),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });
      });
    });

    describe("credential model", () => {
      beforeEach(async () => {
        runaiTableCmp.setProps({
          columns: credentialIndexColumns,
          rows: tableModels.credentialModel.rows,
          filterBy: tableModels.credentialModel.filterBy,
        });
      });

      test("should show the correct order of columns", () => {
        const expectedColumns = getExpectedColumns(credentialIndexColumns);
        const actualColumns = runaiTableCmp.findAll("th").map((th) => th.text());
        expect(actualColumns).toEqual(expectedColumns);
      });

      describe("cells content", () => {
        test("should show the correct content in the 'credentials' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("credentials");
          const expectedCellsContent: Array<string> = tableModels.credentialModel.rows.map(
            (credential: CredentialsListResponseEntry) => credential.meta.name,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'type' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("type");
          const expectedCellsContent: Array<string> = tableModels.credentialModel.rows.map(
            (credential: CredentialsListResponseEntry) => credentialTextForUI[credential.meta.kind as TCredentialKinds],
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'data-sources' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("data-sources");
          const expectedCellsContent: Array<string> = tableModels.credentialModel.rows.map(
            (credential: CredentialsListResponseEntry) =>
              previewStrings(credential.usedBy?.assets?.datasources?.map((ds: AssetRef) => ds.name) || []),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'created-at' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("created-at");
          const expectedCellsContent: Array<string> = tableModels.credentialModel.rows.map(
            (credential: CredentialsListResponseEntry) =>
              dateUtil.dateAndTimeFormat(new Date(credential.meta.createdAt)),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });
      });
    });

    describe("template model", () => {
      beforeEach(async () => {
        runaiTableCmp.setProps({
          columns: workloadTemplateIndexColumns,
          rows: tableModels.templateModel.rows,
          filterBy: tableModels.templateModel.filterBy,
        });
      });

      test("should show the correct order of columns", () => {
        const expectedColumns = getExpectedColumns(workloadTemplateIndexColumns);
        const actualColumns = runaiTableCmp.findAll("th").map((th) => th.text());
        expect(actualColumns).toEqual(expectedColumns);
      });

      describe("cells content", () => {
        test("should show the correct content in the 'template' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("template");
          const expectedCellsContent: Array<string> = tableModels.templateModel.rows.map(
            (template: WorkloadTemplate) => template.meta.name,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'environment' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("environment");
          const expectedCellsContent: Array<string> = tableModels.templateModel.rows.map(
            (template: WorkloadTemplate) => template.spec.assets.environment.name,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'compute-resource' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("compute-resource");
          const expectedCellsContent: Array<string> = tableModels.templateModel.rows.map(
            (template: WorkloadTemplate) => template.spec.assets.compute?.name || "-",
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'datasources' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("datasources");
          const expectedCellsContent: Array<string> = tableModels.templateModel.rows.map((template: WorkloadTemplate) =>
            previewStrings(template.spec.assets.datasources?.map((ds) => ds.name) || []),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'created-by' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("created-by");
          const expectedCellsContent: Array<string> = tableModels.templateModel.rows.map(
            (template: WorkloadTemplate) => template.meta.createdBy,
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });

        test("should show the correct content in the 'created-at' column", () => {
          const cells: DOMWrapper<HTMLTableCellElement>[] = getCellsByColumnName("created-at");
          const expectedCellsContent: Array<string> = tableModels.templateModel.rows.map((template: WorkloadTemplate) =>
            dateUtil.dateAndTimeFormat(new Date(template.meta.createdAt)),
          );
          const actualCellsContent: Array<string> = cells.map((cell: DOMWrapper<HTMLTableCellElement>) => cell.text());
          expect(actualCellsContent).toEqual(expectedCellsContent);
        });
      });
    });
  });
});
