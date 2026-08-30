import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import { AutoDeleteTimeSection } from ".";

describe("auto-delete-time-section.vue", () => {
  let autoDeleteTimeSectionCmp: VueWrapper<InstanceType<typeof AutoDeleteTimeSection>>;

  test("component is loading properly", () => {
    autoDeleteTimeSectionCmp = mount(AutoDeleteTimeSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: null,
        workloadType: "workload",
      },
    });
    expect(autoDeleteTimeSectionCmp.html()).toContain(
      "Set the timeframe for auto-deletion after workload completion or failure",
    );
  });

  test("component should show the time inputs by default", () => {
    autoDeleteTimeSectionCmp = mount(AutoDeleteTimeSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: null,
        workloadType: "workload",
      },
    });
    expect(autoDeleteTimeSectionCmp.find(".runai-duration-inputs").exists()).toBe(true);
  });
  test("component should set no auto delete time to 30 days by default if its prop value is not defined", async () => {
    autoDeleteTimeSectionCmp = mount(AutoDeleteTimeSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: null,
        workloadType: "workload",
      },
    });
    expect(autoDeleteTimeSectionCmp.vm.modelValue).toBe(null);
    expect(autoDeleteTimeSectionCmp.emitted("update:model-value")).toEqual([[60 * 60 * 24 * 30]]);
  });
});
