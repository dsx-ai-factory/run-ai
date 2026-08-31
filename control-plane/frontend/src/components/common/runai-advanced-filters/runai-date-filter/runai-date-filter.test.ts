/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, test } from "vitest";
import type { IDateFilterModel } from "@/models/filter.model";
import { EColumnFilterType, EDateFilterOperator } from "@/models/filter.model";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import RunaiDateFilter from "@/components/common/runai-advanced-filters/runai-date-filter/runai-date-filter.vue";
import { Quasar } from "quasar";
import { unitTestUtil } from "@/utils/unit-test.util";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

enum EDateFilterSelectors {
  DateFilterSelect = "[aid=date-filter-select]",
  SubmitFilterBtn = '[aid="submit-filter-btn"]',
}

describe("RunaiDateFilter", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiDateFilter>>;
  const mountComponent = async (filter: IDateFilterModel): Promise<void> => {
    wrapper = mount(RunaiDateFilter, {
      global: {
        plugins: [Quasar],
      },
      props: {
        filter,
      },
    });
    await wrapper.vm.$nextTick();
  };

  afterEach(() => {
    wrapper.unmount();
  });

  test("renders the component correctly", async () => {
    const filter: IDateFilterModel = {
      type: EColumnFilterType.Date,
      name: "date",
      label: "Date",
      date: "2024/09/05",
      value: EDateFilterOperator.On,
    };
    await mountComponent(filter);
    const menuElement = unitTestUtil.getMenuElement();
    expect(wrapper.exists()).toBe(true);
    expect(menuElement.find(EDateFilterSelectors.DateFilterSelect).text()).toBe(EDateFilterOperator.On);
    expect(wrapper.vm.selectedDate).toEqual(filter.date);
  });

  test("emits 'save' event when saveFilter method is called with a valid date", async () => {
    const filter: IDateFilterModel = {
      label: "Date",
      date: "2024/09/05",
      value: EDateFilterOperator.Before,
      name: "date",
      type: EColumnFilterType.Date,
    };
    await mountComponent(filter);

    const menuElement = unitTestUtil.getMenuElement();

    expect(menuElement.find(EDateFilterSelectors.DateFilterSelect).text()).toBe(EDateFilterOperator.Before);
    await menuElement.find(EDateFilterSelectors.SubmitFilterBtn).trigger("click");
    expect(wrapper.emitted("save")?.[0][0]).toEqual({
      name: "date",
      date: expect.anything(),
      label: "Date",
      type: EColumnFilterType.Date,
      value: EDateFilterOperator.Before,
    });
  });

  test("should disabled apply button when selected date is invalid", async () => {
    const filter: IDateFilterModel = {
      label: "Date",
      date: "2024/09/05",
      value: EDateFilterOperator.Before,
      name: "date",
      type: EColumnFilterType.Date,
    };
    await mountComponent(filter);

    const menuElement = unitTestUtil.getMenuElement();
    // make the date invalid - hard to do with quasar date picker
    wrapper.vm.selectedDate = null;
    await wrapper.vm.$nextTick();
    expect(unitTestUtil.isElementDisabled(menuElement, EDateFilterSelectors.SubmitFilterBtn)).toBe(true);
  });

  test("should select today's date as default", async () => {
    const filter: IDateFilterModel = {
      label: "Date",
      date: "",
      value: EDateFilterOperator.Before,
      name: "date",
      type: EColumnFilterType.Date,
    };
    await mountComponent(filter);

    const menuElement = unitTestUtil.getMenuElement();
    expect(menuElement.find(EDateFilterSelectors.DateFilterSelect).text()).toBe(EDateFilterOperator.Before);
    expect(wrapper.vm.selectedDate).toEqual(dateUtil.dateFormat(new Date(), DATE_FORMATS.STANDARD_DATE));
  });
});
