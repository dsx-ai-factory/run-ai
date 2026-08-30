import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiTableTooltip } from "./index";

const columns = [
  { key: "nodePool", label: "Node pool" },
  { key: "reason", label: "Reason" },
  { key: "details", label: "Details" },
];

const rows = [
  { nodePool: "default", reason: "PodPending", details: "Insufficient GPUs available." },
  { nodePool: "high-mem", reason: "Throttled", details: "Quota exhausted for the project." },
];

describe("runai-table-tooltip", () => {
  it("renders all column headers", () => {
    const cmp = mount(RunaiTableTooltip, {
      global: { plugins: [Quasar] },
      props: { columns, rows },
    });

    const headers = cmp.findAll("th").map((th) => th.text());
    expect(headers).toEqual(["Node pool", "Reason", "Details"]);
  });

  it("renders one body row per data row, with cells in column order", () => {
    const cmp = mount(RunaiTableTooltip, {
      global: { plugins: [Quasar] },
      props: { columns, rows },
    });

    const bodyRows = cmp.findAll("tbody tr");
    expect(bodyRows).toHaveLength(2);
    expect(bodyRows[0]?.findAll("td").map((td) => td.text())).toEqual([
      "default",
      "PodPending",
      "Insufficient GPUs available.",
    ]);
  });

  it("renders the heading when provided", () => {
    const cmp = mount(RunaiTableTooltip, {
      global: { plugins: [Quasar] },
      props: {
        heading: "The workload is in queue and waiting to be scheduled, due to the following reasons:",
        columns,
        rows,
      },
    });

    expect(cmp.find(".runai-table-tooltip__heading").text()).toContain("workload is in queue");
  });

  it("hides the heading when not provided", () => {
    const cmp = mount(RunaiTableTooltip, {
      global: { plugins: [Quasar] },
      props: { columns, rows },
    });

    expect(cmp.find(".runai-table-tooltip__heading").exists()).toBe(false);
  });

  it("applies width style on header cells when column.width is set", () => {
    const cmp = mount(RunaiTableTooltip, {
      global: { plugins: [Quasar] },
      props: {
        columns: [
          { key: "a", label: "A", width: "120px" },
          { key: "b", label: "B" },
        ],
        rows: [{ a: "1", b: "2" }],
      },
    });

    const headers = cmp.findAll("th");
    expect(headers[0]?.attributes("style")).toContain("width: 120px");
    expect(headers[1]?.attributes("style") || "").not.toContain("width");
  });

  it("renders a spinner instead of rows when loading is true", () => {
    const cmp = mount(RunaiTableTooltip, {
      global: { plugins: [Quasar] },
      props: { columns, rows, loading: true },
    });

    expect(cmp.find(".runai-table-tooltip__cell--loading").exists()).toBe(true);
    expect(cmp.find(".q-spinner").exists()).toBe(true);
    // No data rows should render while loading.
    expect(cmp.findAll("tbody tr").length).toBe(1);
  });

  it("renders custom cell content via the cell-{key} slot", () => {
    const cmp = mount(RunaiTableTooltip, {
      global: { plugins: [Quasar] },
      props: { columns, rows },
      slots: {
        "cell-details": '<span class="custom-cell">CUSTOM: {{ params.row.details }}</span>',
      },
    });

    const customCells = cmp.findAll(".custom-cell");
    expect(customCells.length).toBe(2);
    expect(customCells[0]?.text()).toBe("CUSTOM: Insufficient GPUs available.");
  });
});
