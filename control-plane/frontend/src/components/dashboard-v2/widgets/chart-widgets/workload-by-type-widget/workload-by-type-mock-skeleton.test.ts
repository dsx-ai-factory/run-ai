import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import WorkloadByTypeMockSkeleton from "./workload-by-type-mock-skeleton.vue";

describe("workload-by-type-mock-skeleton", () => {
  it("exposes the same guided-tour target as the populated chart", () => {
    const wrapper = mount(WorkloadByTypeMockSkeleton, {
      global: { plugins: [Quasar] },
      props: { clusterName: "test-cluster" },
    });

    expect(wrapper.find('[aid="workload-by-type-drilldown-label"]').text()).toBe("test-cluster");
  });
});
