/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test, afterEach, vi } from "vitest";
import { RunaiDatepickerWithPredefined } from ".";
import type { PredefinedOption } from "@/models/date.model";
import {
  dateUtil,
  defaultPredefinedOptions,
  extendedHoursPredefinedOptions,
  hoursGranularityPredefinedOptions,
  sinceFirstRunPredefinedOptions,
  TimeUnit,
  workloadsLifespanPredefinedOptions,
} from "@/utils/date.util";
import { PREDEFINED_DATES } from "@/models/date.model";
import { unitTestUtil } from "@/utils/unit-test.util";
import { extendedDefaultPredefinedOptions } from "@/utils/date.util/date.util";

describe("runai-datepicker-with-predefined", () => {
  let runaiDatePickerWithPredefinedCmp: VueWrapper<InstanceType<typeof RunaiDatepickerWithPredefined>>;

  async function testSelectPredefinedOption(predefinedOption: PredefinedOption) {
    await runaiDatePickerWithPredefinedCmp.find(".date-picker-dropdown").trigger("click");
    await runaiDatePickerWithPredefinedCmp.vm.$nextTick();
    const qList = unitTestUtil.getMenuElement();
    await qList.find(`[aid="${predefinedOption.label}"]`).trigger("click");

    expect(runaiDatePickerWithPredefinedCmp.html()).includes(predefinedOption.label);
  }

  function mountComponent(predefinedOptions?: Array<PredefinedOption>) {
    runaiDatePickerWithPredefinedCmp = mount(RunaiDatepickerWithPredefined, {
      global: {
        plugins: [Quasar],
      },
      props: predefinedOptions ? { predefinedOptions: predefinedOptions } : {},
    });
  }

  describe("When loading predefined options", () => {
    beforeEach(() => {
      mountComponent(extendedHoursPredefinedOptions);
      runaiDatePickerWithPredefinedCmp.vm.setDefaultDate();
    });

    afterEach(() => {
      runaiDatePickerWithPredefinedCmp.unmount();
    });

    test("component is loading properly with default", async () => {
      mountComponent();
      expect(runaiDatePickerWithPredefinedCmp.html()).toContain("Last 7 days");
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedOptionLabel).toContain("Last 7 days");
      await testSelectPredefinedOption(defaultPredefinedOptions[0]);
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.hoursCount).toBeUndefined();
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.daysCount).toEqual(0);
    });

    test("component is loading properly", async () => {
      // expect(runaiDatePickerWithPredefinedCmp.html()).toContain("Last 24 hours");
      await testSelectPredefinedOption(extendedHoursPredefinedOptions[1]);
    });

    test("component is loading with missing units", () => {
      let thrownError = null;
      try {
        mountComponent([{ label: "test" }, { label: "test2" }, { label: "test3" }] as Array<PredefinedOption>);
      } catch (error) {
        thrownError = error;
      }
      expect(thrownError).toBeInstanceOf(Error);
    });

    test("select predefined hours", async () => {
      await testSelectPredefinedOption(extendedHoursPredefinedOptions[1]);
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.hoursCount).toEqual(3);
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.daysCount).toBeUndefined();
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.minutesCount).toBeUndefined();
    });

    test("select predefined days", async () => {
      await testSelectPredefinedOption(extendedHoursPredefinedOptions[8]);
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.daysCount).toEqual(30);
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.hoursCount).toBeUndefined();
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.minutesCount).toBeUndefined();
    });
  });

  describe("Custom Date Selection", () => {
    beforeEach(() => {
      mountComponent(defaultPredefinedOptions);
    });

    afterEach(() => {
      runaiDatePickerWithPredefinedCmp.unmount();
    });

    test("select custom dates", async () => {
      const startDate = new Date(2023, 0, 1);
      const endDate = new Date(2023, 0, 10);

      await runaiDatePickerWithPredefinedCmp.vm.onCustomDateChange(startDate, true);
      await runaiDatePickerWithPredefinedCmp.vm.onCustomDateChange(endDate, false);

      expect(runaiDatePickerWithPredefinedCmp.vm.startDate).toEqual(startDate);
      expect(runaiDatePickerWithPredefinedCmp.vm.endDate).toEqual(endDate);
      expect(runaiDatePickerWithPredefinedCmp.html()).toContain("Jan 01 - Jan 10, 2023");
    });
  });

  describe("Event Emissions", () => {
    beforeEach(() => {
      mountComponent(defaultPredefinedOptions);
    });

    afterEach(() => {
      runaiDatePickerWithPredefinedCmp.unmount();
    });

    test("emits changed event on predefined date selection", async () => {
      const predefinedOption = defaultPredefinedOptions[2]; // e.g., Last 7 days
      await runaiDatePickerWithPredefinedCmp.vm.onPopupHide();
      const emitted = runaiDatePickerWithPredefinedCmp.emitted("changed");
      expect(emitted).toHaveLength(1);
      expect(emitted?.[0][1]).toBe(predefinedOption.label);
      expect(emitted?.[0][0]).toHaveLength(2);
    });

    test("emits changed event on custom date selection with less then 30 days", async () => {
      const startDate = new Date(2023, 0, 1);
      let endDate = new Date(2023, 0, 9);
      runaiDatePickerWithPredefinedCmp.vm.endDate = endDate;

      await runaiDatePickerWithPredefinedCmp.vm.onCustomDateChange(startDate, true);
      await runaiDatePickerWithPredefinedCmp.vm.onPopupHide();
      endDate = dateUtil.getEndOfDay(endDate);

      const emitted = runaiDatePickerWithPredefinedCmp.emitted("changed");
      expect(emitted).toHaveLength(1);
      expect(emitted?.[0][0]).toEqual([startDate, endDate]);
      expect(emitted?.[0][1]).toBe("Custom");
      expect(emitted?.[0][0]).toHaveLength(2);
    });

    test("emits changed event on custom date selection more then 30 days", async () => {
      const startDate = new Date(2023, 0, 1);
      let endDate = new Date(2023, 1, 9);
      runaiDatePickerWithPredefinedCmp.vm.endDate = endDate;

      await runaiDatePickerWithPredefinedCmp.vm.onCustomDateChange(startDate, true);
      await runaiDatePickerWithPredefinedCmp.vm.onPopupHide();
      endDate = dateUtil.adjustDateBy(TimeUnit.day, startDate, 30);
      endDate = dateUtil.getEndOfDay(endDate);

      const emitted = runaiDatePickerWithPredefinedCmp.emitted("changed");
      expect(emitted).toHaveLength(1);
      expect(emitted?.[0][0]).toEqual([startDate, endDate]);
      expect(emitted?.[0][1]).toBe("Custom");
    });
  });

  describe("Computed Properties", () => {
    beforeEach(() => {
      mountComponent(defaultPredefinedOptions);
    });

    afterEach(() => {
      runaiDatePickerWithPredefinedCmp.unmount();
    });

    test("title computed property", async () => {
      const startDate = new Date(2023, 0, 1);
      const endDate = new Date(2023, 0, 10);
      await runaiDatePickerWithPredefinedCmp.vm.onCustomDateChange(startDate, true);
      await runaiDatePickerWithPredefinedCmp.vm.onCustomDateChange(endDate, false);
      expect(runaiDatePickerWithPredefinedCmp.html()).toContain("Jan 01 - Jan 10, 2023");
    });

    test("selectedOptionLabel computed property", () => {
      const initialDefaultOption = defaultPredefinedOptions[2];
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedOptionLabel).toBe(initialDefaultOption.label);
      expect(runaiDatePickerWithPredefinedCmp.vm.selectedOptionLabel).toBe(defaultPredefinedOptions[2].label);
    });
  });

  describe("Selecting Predefined Date", () => {
    afterEach(() => {
      runaiDatePickerWithPredefinedCmp.unmount();
    });

    function mountComponent(props = {}) {
      runaiDatePickerWithPredefinedCmp = mount(RunaiDatepickerWithPredefined, {
        global: {
          plugins: [Quasar],
        },
        props: {
          defaultStartDate: new Date(2023, 0, 1),
          ...props,
        },
      });
    }

    test("select 'Last 24 hours' and check start and end date", async () => {
      mountComponent({ predefinedOptions: extendedHoursPredefinedOptions });
      const option = extendedHoursPredefinedOptions[3];
      await testSelectPredefinedOption(option);

      const expectedStartDate = dateUtil.adjustDateBy(TimeUnit.hour, new Date(), -24);
      const expectedEndDate = new Date();

      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.label).toBe("Last 24 hours");
      expect(runaiDatePickerWithPredefinedCmp.vm.startDate.getMinutes()).toBe(expectedStartDate.getMinutes());
      expect(runaiDatePickerWithPredefinedCmp.vm.endDate.getMinutes()).toBe(expectedEndDate.getMinutes());
    });

    test("select 'Last 7 days' and check start and end date", async () => {
      mountComponent({ predefinedOptions: extendedDefaultPredefinedOptions });
      const option = extendedDefaultPredefinedOptions[2];
      await testSelectPredefinedOption(option);

      const expectedStartDate = dateUtil.adjustDateBy(TimeUnit.day, new Date(), -7);
      const expectedEndDate = new Date();

      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.label).toBe("Last 7 days");
      expect(runaiDatePickerWithPredefinedCmp.vm.startDate.getMinutes()).toBe(expectedStartDate.getMinutes());
      expect(runaiDatePickerWithPredefinedCmp.vm.endDate.getMinutes()).toBe(expectedEndDate.getMinutes());
    });

    test("select 'Last 30 minutes' and check start and end date", async () => {
      mountComponent({ predefinedOptions: hoursGranularityPredefinedOptions });
      const option = hoursGranularityPredefinedOptions[0];
      await testSelectPredefinedOption(option);

      const expectedStartDate = dateUtil.adjustDateBy(TimeUnit.hour, new Date(), -1);
      const expectedEndDate = new Date();

      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.label).toBe("Last hour");
      expect(runaiDatePickerWithPredefinedCmp.vm.startDate.getHours()).toBe(expectedStartDate.getHours());
      expect(runaiDatePickerWithPredefinedCmp.vm.endDate.getHours()).toBe(expectedEndDate.getHours());
    });
  });

  describe("SINCE_FIRST_RUN", () => {
    function mountComponent(props = {}) {
      runaiDatePickerWithPredefinedCmp = mount(RunaiDatepickerWithPredefined, {
        global: {
          plugins: [Quasar],
        },
        props,
      });
    }

    afterEach(() => {
      runaiDatePickerWithPredefinedCmp.unmount();
    });

    test("select SINCE_FIRST_RUN should set startDate to workloadFirstRunDate", () => {
      const defaultStartDate = new Date(2023, 0, 1);
      const workloadFirstRunDate = new Date(2023, 0, 1, 6, 0, 0);
      mountComponent({
        predefinedOptions: [
          ...workloadsLifespanPredefinedOptions,
          ...sinceFirstRunPredefinedOptions,
          ...hoursGranularityPredefinedOptions,
        ],
        defaultStartDate,
        workloadFirstRunDate,
      });

      runaiDatePickerWithPredefinedCmp.vm.onSelectPredefinedDate(sinceFirstRunPredefinedOptions[0]);

      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.label).toBe(PREDEFINED_DATES.SINCE_FIRST_RUN);
      expect(runaiDatePickerWithPredefinedCmp.vm.startDate).toEqual(workloadFirstRunDate);
    });

    test("select SINCE_FIRST_RUN should use workloadLifespanEndDate as endDate", () => {
      const defaultStartDate = new Date(2023, 0, 1);
      const workloadFirstRunDate = new Date(2023, 0, 1, 6, 0, 0);
      const workloadLifespanEndDate = new Date(2023, 0, 5);
      mountComponent({
        predefinedOptions: [
          ...workloadsLifespanPredefinedOptions,
          ...sinceFirstRunPredefinedOptions,
          ...hoursGranularityPredefinedOptions,
        ],
        defaultStartDate,
        workloadFirstRunDate,
        workloadLifespanEndDate,
      });

      runaiDatePickerWithPredefinedCmp.vm.onSelectPredefinedDate(sinceFirstRunPredefinedOptions[0]);

      expect(runaiDatePickerWithPredefinedCmp.vm.endDate).toEqual(workloadLifespanEndDate);
    });

    test("SINCE_FIRST_RUN should be disabled when workloadFirstRunDate is not provided", () => {
      const defaultStartDate = new Date(2023, 0, 1);
      mountComponent({
        predefinedOptions: [
          ...workloadsLifespanPredefinedOptions,
          ...sinceFirstRunPredefinedOptions,
          ...hoursGranularityPredefinedOptions,
        ],
        defaultStartDate,
      });

      const sinceFirstRunOption = runaiDatePickerWithPredefinedCmp.vm.filteredPredefinedOptions.find(
        (opt: PredefinedOption) => opt.label === PREDEFINED_DATES.SINCE_FIRST_RUN,
      );
      expect(sinceFirstRunOption?.disabled).toBe(true);
    });

    test("SINCE_FIRST_RUN should be enabled when workloadFirstRunDate is provided", () => {
      const defaultStartDate = new Date(2023, 0, 1);
      const workloadFirstRunDate = new Date(2023, 0, 1, 6, 0, 0);
      mountComponent({
        predefinedOptions: [
          ...workloadsLifespanPredefinedOptions,
          ...sinceFirstRunPredefinedOptions,
          ...hoursGranularityPredefinedOptions,
        ],
        defaultStartDate,
        workloadFirstRunDate,
      });

      const sinceFirstRunOption = runaiDatePickerWithPredefinedCmp.vm.filteredPredefinedOptions.find(
        (opt: PredefinedOption) => opt.label === PREDEFINED_DATES.SINCE_FIRST_RUN,
      );
      expect(sinceFirstRunOption?.disabled).toBeFalsy();
    });
  });

  describe("WORKLOADS_LIFESPAN", () => {
    function mountComponent(props = {}) {
      runaiDatePickerWithPredefinedCmp = mount(RunaiDatepickerWithPredefined, {
        global: {
          plugins: [Quasar],
        },
        props,
      });
    }

    test("select WORKLOADS_LIFESPAN option", async () => {
      const defaultStartDate = new Date(2023, 0, 1);
      mountComponent({
        predefinedOptions: [...workloadsLifespanPredefinedOptions, ...hoursGranularityPredefinedOptions],
        defaultStartDate,
      });

      await testSelectPredefinedOption(workloadsLifespanPredefinedOptions[0]);

      expect(runaiDatePickerWithPredefinedCmp.vm.selectedPredefinedOption.label).toBe(
        PREDEFINED_DATES.WORKLOADS_LIFESPAN,
      );
      expect(runaiDatePickerWithPredefinedCmp.vm.startDate).toEqual(defaultStartDate);
      expect(runaiDatePickerWithPredefinedCmp.vm.endDate).toBeDefined();
    });

    test("select WORKLOADS_LIFESPAN option should throw exception when defaultStartDate not provided", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      mountComponent({
        predefinedOptions: [...workloadsLifespanPredefinedOptions, ...hoursGranularityPredefinedOptions],
      });

      expect(errorSpy).toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    test("select WORKLOADS_LIFESPAN option should set the startDate as default start date", async () => {
      const defaultStartDate = new Date(2023, 0, 1);
      mountComponent({
        predefinedOptions: [...workloadsLifespanPredefinedOptions, ...hoursGranularityPredefinedOptions],
        defaultStartDate,
      });
      runaiDatePickerWithPredefinedCmp.vm.startDate = new Date(2023, 0, 2);
      runaiDatePickerWithPredefinedCmp.vm.onSelectPredefinedDate(workloadsLifespanPredefinedOptions[0]);
      expect(runaiDatePickerWithPredefinedCmp.vm.startDate).toEqual(defaultStartDate);
    });

    test("select WORKLOADS_LIFESPAN option should set the startDate as default start date", async () => {
      const defaultStartDate = new Date(2023, 0, 1);
      mountComponent({
        predefinedOptions: [...workloadsLifespanPredefinedOptions, ...hoursGranularityPredefinedOptions],
        defaultStartDate,
        defaultOptionIndex: 0,
      });
      expect(runaiDatePickerWithPredefinedCmp.vm.getDefaultStartDate()).toEqual(defaultStartDate);
    });
  });
});
