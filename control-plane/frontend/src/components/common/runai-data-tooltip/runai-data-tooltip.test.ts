import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiDataTooltip } from "./index";

const baseRows = [
  { label: "Allocated quota:", value: 20, dotColor: "#2caffe" },
  { label: "Non-allocated quota:", value: 66, dotColor: "#544fc5" },
];

describe("runai-data-tooltip", () => {
  it("renders all legend rows with label and value", () => {
    const cmp = mount(RunaiDataTooltip, {
      global: { plugins: [Quasar] },
      props: { rows: baseRows },
    });

    const html = cmp.html();
    expect(html).toContain("Allocated quota:");
    expect(html).toContain("20");
    expect(html).toContain("Non-allocated quota:");
    expect(html).toContain("66");
    expect(cmp.findAll(".runai-data-tooltip__row")).toHaveLength(2);
  });

  it("renders the heading when provided", () => {
    const cmp = mount(RunaiDataTooltip, {
      global: { plugins: [Quasar] },
      props: { heading: "{project-name}", rows: baseRows },
    });

    expect(cmp.find(".runai-data-tooltip__heading").text()).toBe("{project-name}");
  });

  it("hides the heading when not provided", () => {
    const cmp = mount(RunaiDataTooltip, {
      global: { plugins: [Quasar] },
      props: { rows: baseRows },
    });

    expect(cmp.find(".runai-data-tooltip__heading").exists()).toBe(false);
  });

  it("renders the summary row when provided", () => {
    const cmp = mount(RunaiDataTooltip, {
      global: { plugins: [Quasar] },
      props: {
        rows: baseRows,
        summary: { label: "Department available quota:", value: 100 },
      },
    });

    const summary = cmp.find(".runai-data-tooltip__summary");
    expect(summary.exists()).toBe(true);
    expect(summary.text()).toContain("Department available quota:");
    expect(summary.text()).toContain("100");
  });

  it("renders a dot when dotColor is provided and skips it otherwise", () => {
    const cmp = mount(RunaiDataTooltip, {
      global: { plugins: [Quasar] },
      props: {
        rows: [
          { label: "with dot", value: 1, dotColor: "#2caffe" },
          { label: "without dot", value: 2 },
        ],
      },
    });

    const dots = cmp.findAll(".runai-data-tooltip__dot");
    expect(dots).toHaveLength(1);
    expect(dots[0]?.attributes("style")).toContain("background-color: rgb(44, 175, 254)");
  });
});
