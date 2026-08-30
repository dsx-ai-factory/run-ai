import { describe, expect, it } from "vitest";
import { OVERVIEW_GUIDED_TOUR_STEPS } from "./guided-tours";

describe("overview guided tour", () => {
  it("targets the workload drilldown labels in step 3", () => {
    const drilldownStep = OVERVIEW_GUIDED_TOUR_STEPS[2];

    expect(drilldownStep.element).toBe('[aid="workload-by-type-drilldown-label"]');
    expect(drilldownStep.beforeShowStep?.waitForElement).toBe('[aid="workload-by-type-drilldown-label"]');
  });
});
