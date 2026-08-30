import { describe, expect, it } from "vitest";

import {
  completeChartDrilldown,
  getChartDrillUpNavigation,
  getNextChartDrilldownBreadcrumbs,
} from "./chart-drilldown.util";

interface TestState {
  id: string;
  meta: { entityId: string } | null;
}

describe("chart drilldown navigation", () => {
  it("starts breadcrumbs at the root without mutating an existing path", () => {
    const current = ["cluster-a", "Departments"];

    expect(getNextChartDrilldownBreadcrumbs([], "cluster-a")).toEqual(["cluster-a"]);
    expect(getNextChartDrilldownBreadcrumbs(current, "cluster-a")).toEqual(current);
    expect(getNextChartDrilldownBreadcrumbs(current, "cluster-a")).not.toBe(current);
  });

  it("commits a completed level with its history and announcement", () => {
    const stack: TestState[] = [{ id: "root", meta: null }];
    const previousState: TestState = { id: "department", meta: { entityId: "department-a" } };

    const result = completeChartDrilldown(stack, previousState, ["cluster-a", "Departments"], "Projects", 4);

    expect(result).toEqual({
      stack: [stack[0], previousState],
      breadcrumbs: ["cluster-a", "Departments", "Projects"],
      announcement: "Projects details loaded. 4 categories shown.",
    });
    expect(stack).toHaveLength(1);
  });

  it("restores an earlier level and trims newer history", () => {
    const stack: TestState[] = [
      { id: "root", meta: null },
      { id: "department", meta: { entityId: "department-a" } },
    ];

    expect(getChartDrillUpNavigation(stack, ["cluster-a", "Departments", "Projects"], 1)).toEqual({
      targetState: stack[1],
      stack: [stack[0]],
      breadcrumbs: ["cluster-a", "Departments"],
      announcement: "Previous drilldown level loaded.",
    });
  });

  it("restores the top level and rejects an invalid index", () => {
    const stack: TestState[] = [{ id: "root", meta: null }];

    expect(getChartDrillUpNavigation(stack, ["cluster-a", "Departments"], 0)).toEqual({
      targetState: stack[0],
      stack: [],
      breadcrumbs: [],
      announcement: "Top chart level loaded.",
    });
    expect(getChartDrillUpNavigation(stack, ["cluster-a"], 1)).toBeNull();
    expect(getChartDrillUpNavigation(stack, ["cluster-a"], -1)).toBeNull();
  });
});
