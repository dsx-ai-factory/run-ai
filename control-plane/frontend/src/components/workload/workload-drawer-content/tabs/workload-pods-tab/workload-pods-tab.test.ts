import { describe, expect, it } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { Quasar } from "quasar";

import WorkloadPodsTab from "./workload-pods-tab.vue";
import { allWorkloadPodsModalColumnsMap } from "@/table-models/workload.table-model";
import type { WorkloadServicePod } from "@/models/workload.model";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiListModalSearch } from "@/components/common";

const buildPod = (name: string): WorkloadServicePod =>
  ({
    id: name,
    name,
    workloadId: "wid-1",
    clusterId: "cluster-1",
    tenantId: 1,
    createdAt: "2026-04-30T00:00:00Z",
    updatedAt: "2026-04-30T00:00:00Z",
    k8sPhase: "Running",
    k8sPhaseUpdatedAt: "2026-04-30T00:00:00Z",
    priorityClassName: "",
    containers: [{ name: "main", image: "image:latest" }],
  }) as WorkloadServicePod;

const mountTab = (props: { pods: WorkloadServicePod[]; loading?: boolean }) =>
  shallowMount(WorkloadPodsTab, {
    global: { plugins: [Quasar] },
    props,
  });

// RunaiTable / RunaiListModalSearch are defined via defineComponent without an
// explicit `name`, so we look them up by component reference instead of by name.
const findTable = (wrapper: ReturnType<typeof mountTab>) => wrapper.findComponent(RunaiTable);
const findSearch = (wrapper: ReturnType<typeof mountTab>) => wrapper.findComponent(RunaiListModalSearch);

describe("workload-pods-tab.vue", () => {
  describe("table contract", () => {
    it("does NOT include the logs/terminal action columns", () => {
      const wrapper = mountTab({ pods: [buildPod("pod-1")] });
      const table = findTable(wrapper);
      expect(table.exists()).toBe(true);

      const columns = table.attributes("columns") as string | undefined;
      // shallowMount stubs serialize columns prop; use props() if available, otherwise attributes
      const columnList = (table.props("columns") as Array<{ name: string }> | undefined) ?? [];
      const columnNames = columnList.map((c) => c.name);
      expect(columnNames).not.toContain(allWorkloadPodsModalColumnsMap.logs.name);
      expect(columnNames).not.toContain(allWorkloadPodsModalColumnsMap.terminal.name);
      // sanity: serialized fallback also doesn't mention them as a stringified column id
      if (typeof columns === "string") {
        expect(columns.toLowerCase()).not.toMatch(/logs/);
      }
    });

    it("includes the expected non-action columns (Pod, Status, Node, Node pool, allocations)", () => {
      const wrapper = mountTab({ pods: [buildPod("pod-1")] });
      const table = findTable(wrapper);
      expect(table.exists()).toBe(true);

      const columnList = (table.props("columns") as Array<{ name: string }> | undefined) ?? [];
      const columnNames = columnList.map((c) => c.name);

      expect(columnNames).toContain(allWorkloadPodsModalColumnsMap.name.name);
      expect(columnNames).toContain(allWorkloadPodsModalColumnsMap.status.name);
      expect(columnNames).toContain(allWorkloadPodsModalColumnsMap.nodeName.name);
      expect(columnNames).toContain(allWorkloadPodsModalColumnsMap.nodePools.name);
    });

    it("forwards the loading prop to the table", () => {
      const wrapper = mountTab({ pods: [], loading: true });
      const table = findTable(wrapper);
      expect(table.props("loading")).toBe(true);
    });
  });

  describe("filterBy contract (regression: rowsPerPage drives client-side mode, not rowsNumber)", () => {
    it("initializes filterBy with rowsPerPage=10 and an empty searchTerm", () => {
      const wrapper = mountTab({ pods: [] });
      const table = findTable(wrapper);
      const filterBy = table.props("filterBy") as { rowsPerPage?: number; rowsNumber?: number; searchTerm?: string };

      expect(filterBy.rowsPerPage).toBe(10);
      // Critical regression check: rowsNumber must NOT be set, otherwise q-table
      // switches into server-side mode and the :filter prop is ignored, breaking search.
      expect(filterBy.rowsNumber).toBeUndefined();
      expect(filterBy.searchTerm).toBe("");
    });

    it("uses modal-view on the table so the search filter is wired through", () => {
      const wrapper = mountTab({ pods: [] });
      const table = findTable(wrapper);
      expect(table.props("modalView")).toBe(true);
    });

    it("disables the runai-table border (the panel chrome owns framing)", () => {
      const wrapper = mountTab({ pods: [] });
      const table = findTable(wrapper);
      expect(table.props("bordered")).toBe(false);
    });
  });

  describe("pagination / filter updates from the table", () => {
    // Regression: q-table emits pagination changes (e.g. user picks "Rows per page: 5")
    // through runai-table's @update-filters event. If we don't listen and reassign
    // filterBy, the rows-per-page dropdown updates visually but the table keeps using
    // the original rowsPerPage — a real bug we shipped briefly.
    it("applies update-filters from the table so rows-per-page actually takes effect", async () => {
      const wrapper = mountTab({ pods: [] });
      const table = findTable(wrapper);

      await table.vm.$emit("update-filters", { rowsPerPage: 5, searchTerm: "" });

      const filterBy = table.props("filterBy") as { rowsPerPage?: number };
      expect(filterBy.rowsPerPage).toBe(5);
    });

    it("preserves the search term when the table emits update-filters", async () => {
      const wrapper = mountTab({ pods: [] });
      const search = findSearch(wrapper);
      await search.vm.$emit("update:model-value", "rag-dynamo");

      const table = findTable(wrapper);
      // q-table doesn't know about searchTerm, so it emits only its own state.
      // runai-table.updateFilters merges over the previous filterBy so searchTerm
      // is preserved, but our onFiltersUpdate must respect the merged shape.
      await table.vm.$emit("update-filters", { rowsPerPage: 5, searchTerm: "rag-dynamo" });

      const filterBy = table.props("filterBy") as { rowsPerPage?: number; searchTerm?: string };
      expect(filterBy.rowsPerPage).toBe(5);
      expect(filterBy.searchTerm).toBe("rag-dynamo");
    });
  });

  describe("search input", () => {
    it("renders the search input with a stable aid and placeholder for automation", () => {
      const wrapper = mountTab({ pods: [] });
      const search = findSearch(wrapper);
      expect(search.exists()).toBe(true);
      // `aid` is not a declared prop on RunaiListModalSearch — it falls through as
      // an HTML attribute on the rendered element so automation tooling can target it.
      expect(search.attributes("aid")).toBe("structure-pods-search");
      expect(search.props("placeholder")).toBe("Search pods");
      expect(search.props("debounce")).toBe(300);
    });

    it("propagates user typing into filterBy.searchTerm so the table filter updates", async () => {
      const wrapper = mountTab({ pods: [] });
      const search = findSearch(wrapper);
      expect(search.exists()).toBe(true);

      await search.vm.$emit("update:model-value", "rag-dynamo");

      const table = findTable(wrapper);
      const filterBy = table.props("filterBy") as { searchTerm?: string };
      expect(filterBy.searchTerm).toBe("rag-dynamo");
    });
  });
});
