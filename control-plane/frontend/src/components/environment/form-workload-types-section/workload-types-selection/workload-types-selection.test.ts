import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import { WorkloadTypesSelection } from ".";
import type { TTestEmit } from "@/models/unit-test.model";

describe("workload-types-selection.vue", () => {
  let workloadTypesSelectionCmp: VueWrapper<InstanceType<typeof WorkloadTypesSelection>>;

  beforeEach(() => {
    workloadTypesSelectionCmp = mount(WorkloadTypesSelection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        options: [
          {
            label: "workspace",
            value: "workspace",
            disable: true,
          },
          {
            label: "training",
            value: "training",
          },
          {
            label: "inference",
            value: "inference",
          },
        ],
        workloadTypes: {
          workspace: true,
          distributed: false,
          training: false,
          inference: false,
        },
      },
    });
  });

  test("component is loading properly", () => {
    expect(workloadTypesSelectionCmp.html()).toContain("Select the type of workload that can use this environment");
  });

  test("component should render the options provided in props", () => {
    const elOptions = workloadTypesSelectionCmp.findAll(".q-checkbox");
    expect(elOptions.length).toBe(3);
  });

  test("component should disable options which are disabled in props", () => {
    const elOptions = workloadTypesSelectionCmp.findAll(".q-checkbox");
    expect(elOptions[0].element.classList.contains("disabled")).toBe(true);
    expect(elOptions[1].element.classList.contains("disabled")).toBe(false);
  });

  test("component should emit an event when an option is selected", async () => {
    const elOptions = workloadTypesSelectionCmp.findAll(".q-checkbox");
    await elOptions[1].trigger("click");
    const expectedEmits: TTestEmit<Record<string, boolean>> = [[{ workspace: true, training: true }]];
    expect(workloadTypesSelectionCmp.emitted("update-workload-types")).toEqual(expectedEmits);
  });

  test("component should emit an event when an option is unselected", async () => {
    await workloadTypesSelectionCmp.setProps({
      options: [
        {
          label: "workspace",
          value: "workspace",
        },
        {
          label: "training",
          value: "training",
        },
      ],
    });
    const elOptions = workloadTypesSelectionCmp.findAll(".q-checkbox");
    await elOptions[0].trigger("click");
    const expectedEmits: TTestEmit<Record<string, boolean>> = [[{}]];
    expect(workloadTypesSelectionCmp.emitted("update-workload-types")).toEqual(expectedEmits);
  });
});
