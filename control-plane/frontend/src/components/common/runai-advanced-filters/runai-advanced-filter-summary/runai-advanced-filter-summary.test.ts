import { afterEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { RunaiAdvancedFilterSummary } from "@/components/common/runai-advanced-filters/runai-advanced-filter-summary";
import type { IAdvancedFilterModel, IDateFilterModel } from "@/models/filter.model";
import {
  EColumnFilterType,
  EDateFilterOperator,
  EFilterOperator,
  type IFreeTextFilterModel,
} from "@/models/filter.model";
import { Quasar } from "quasar";

describe("RunaiAdvancedFilterChip", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiAdvancedFilterSummary>>;

  const mountComponent = async (filters: IAdvancedFilterModel[]): Promise<void> => {
    wrapper = mount(RunaiAdvancedFilterSummary, {
      global: {
        plugins: [Quasar],
      },
      props: {
        filters,
      },
    });
    await wrapper.vm.$nextTick();
  };

  afterEach(() => {
    wrapper.unmount();
  });

  test("renders the chip component correctly", async () => {
    const filters: IAdvancedFilterModel[] = [
      {
        type: EColumnFilterType.FreeText,
        name: "name",
        label: "Workload",
        term: "workspace-ds2423e",
        value: EFilterOperator.Contains,
      } as IFreeTextFilterModel,
    ];

    await mountComponent(filters);

    expect(wrapper.exists()).toBe(true);
    expect(wrapper.findAll(".filter-chip").length).toBe(1);
    const chipContent = wrapper.find(".q-chip").text();
    expect(chipContent).includes("Workload: contains workspace-ds2423e");
  });

  test("renders date filter chip correctly with DD/MM/YYYY format", async () => {
    // Test with DD/MM/YYYY format as it comes from runai-date-filter
    const filters: IDateFilterModel[] = [
      {
        type: EColumnFilterType.Date,
        name: "date",
        label: "Date",
        date: "26/11/2025", // DD/MM/YYYY format from the date picker
        value: EDateFilterOperator.Before,
      },
    ];

    await mountComponent(filters);

    expect(wrapper.exists()).toBe(true);
    expect(wrapper.findAll(".filter-chip").length).toBe(1);
    const chipContent = wrapper.find(".q-chip").text();
    // Should display the date as-is without parsing/reformatting
    expect(chipContent).includes("Date: before 26/11/2025");
  });

  test("renders date filter chip with 'on' operator correctly", async () => {
    const filters: IDateFilterModel[] = [
      {
        type: EColumnFilterType.Date,
        name: "createdAt",
        label: "Created At",
        date: "15/06/2024",
        value: EDateFilterOperator.On,
      },
    ];

    await mountComponent(filters);

    const chipContent = wrapper.find(".q-chip").text();
    expect(chipContent).includes("Created At: on 15/06/2024");
  });

  test("renders date filter chip with 'after' operator correctly", async () => {
    const filters: IDateFilterModel[] = [
      {
        type: EColumnFilterType.Date,
        name: "deletedAt",
        label: "Deleted At",
        date: "01/01/2024",
        value: EDateFilterOperator.After,
      },
    ];

    await mountComponent(filters);

    const chipContent = wrapper.find(".q-chip").text();
    expect(chipContent).includes("Deleted At: after 01/01/2024");
  });

  test("emits 'remove-filter' event when a filter chip is removed", async () => {
    const filters: IAdvancedFilterModel[] = [
      {
        type: EColumnFilterType.FreeText,
        name: "name",
        label: "Workload",
        term: "workspace-ds2423e",
        value: EFilterOperator.Contains,
      } as IFreeTextFilterModel,
    ];

    await mountComponent(filters);

    await wrapper.find("[aria-label='Remove']").trigger("click");

    expect(wrapper.emitted("remove")).toBeTruthy();
    expect(wrapper.emitted("remove")?.[0]?.[0]).toBe("name");
  });

  test("emits 'select-filter' event when a filter chip is clicked", async () => {
    const filters: IAdvancedFilterModel[] = [
      {
        type: EColumnFilterType.FreeText,
        name: "name",
        label: "Workload",
        term: "workspace-ds2423e",
        value: EFilterOperator.Contains,
      } as IFreeTextFilterModel,
    ];

    await mountComponent(filters);

    await wrapper.find(".filter-chip").trigger("click");

    expect(wrapper.emitted("select")).toBeTruthy();
    expect(wrapper.emitted("select")?.[0]?.[0]).toEqual(filters[0]);
  });

  test("should render radio enum based summary correctly", async () => {
    const filters: IAdvancedFilterModel[] = [
      {
        selectOptions: [
          { label: "Automatic", value: "Auto" },
          { label: "Manual", value: "Use" },
        ],
        selectedValues: ["Auto"],
        inputType: "radio",
        label: "MNNVL",
        name: "gpuNetworkAccelerationDetection",
        type: EColumnFilterType.EnumBased,
        maxWidth: "290px",
      },
      {
        selectOptions: [
          { label: "Detected", value: true },
          { label: "Not detected", value: false },
        ],
        selectedValues: [true],
        inputType: "radio",
        label: "MNNVL nodes",
        name: "gpuNetworkAccelerationDetected",
        type: EColumnFilterType.EnumBased,
      },
    ];

    await mountComponent(filters);
    expect(wrapper.findAll(".filter-chip").length).toBe(2);
    const chipContent = wrapper.findAll(".q-chip");
    expect(chipContent[0].text()).includes("MNNVL: Automatic");
    expect(chipContent[1].text()).includes("MNNVL nodes: Detected");
  });
});
