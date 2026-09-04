/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, test } from "vitest";
import type { INumericFilterModel } from "@/models/filter.model";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { unitTestUtil } from "@/utils/unit-test.util";
import { RunaiNumericFilter } from "@/components/common/runai-advanced-filters/runai-numeric-filter/index";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";

describe("RunaiFreeTextFilter", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiNumericFilter>>;
  let menuElement: DOMWrapper<Element>;
  const mountComponent = async (filter: INumericFilterModel): Promise<void> => {
    wrapper = mount(RunaiNumericFilter, {
      global: {
        plugins: [Quasar],
      },
      props: {
        filter,
      },
    });
    await wrapper.vm.$nextTick();
    menuElement = unitTestUtil.getMenuElement();
  };

  afterEach(() => {
    wrapper.unmount();
  });
  const getNumericInputValue = (selector: "min" | "max"): string => {
    return (menuElement.find(`[aid=numeric-filter-input-${selector}]`).element as HTMLInputElement).value;
  };
  const getNumericOperatorValue = (): string => {
    return menuElement.find("[aid=numeric-filter-select]").text();
  };
  const clickOnClose = async () => {
    await menuElement.find('[aid="close-filter-btn"]').trigger("click");
  };
  const clickOnSave = async () => {
    await menuElement.find('[aid="submit-filter-btn"]').trigger("click");
  };
  test("renders the component correctly with Range operator", async () => {
    const filter: INumericFilterModel = {
      type: EColumnFilterType.Numeric,
      name: WorkloadSortFilterFields.AllocatedGpu,
      label: "GPU compute allocation",
      operator: EFilterOperator.Range,
      minValue: 0,
      maxValue: 10,
    };
    await mountComponent(filter);
    const minInputValue: string = getNumericInputValue("min");
    const maxInputValue: string = getNumericInputValue("max");
    const selectedOperator: string = getNumericOperatorValue();
    expect(minInputValue).toBe("0");
    expect(maxInputValue).toBe("10");
    expect(selectedOperator).toBe("range");
  });

  test("renders the component correctly with Equals operator", async () => {
    const filter: INumericFilterModel = {
      type: EColumnFilterType.Numeric,
      name: WorkloadSortFilterFields.AllocatedGpu,
      label: "GPU compute allocation",
      operator: EFilterOperator.Equals,
      minValue: 5,
      maxValue: 0,
    };
    await mountComponent(filter);
    const minInputValue: string = getNumericInputValue("min");
    const selectedOperator: string = getNumericOperatorValue();
    expect(minInputValue).toBe("5");
    expect(selectedOperator).toBe("=");
  });

  test("renders the component correctly with LessThanOrEqual operator", async () => {
    const filter: INumericFilterModel = {
      type: EColumnFilterType.Numeric,
      name: WorkloadSortFilterFields.AllocatedGpu,
      label: "GPU compute allocation",
      operator: EFilterOperator.LessThanOrEqual,
      minValue: 5,
      maxValue: 0,
    };
    await mountComponent(filter);
    const minInputValue: string = getNumericInputValue("min");
    const selectedOperator: string = getNumericOperatorValue();
    expect(minInputValue).toBe("5");
    expect(selectedOperator).toBe("<=");
  });

  test("emits 'save' event when save btn is called", async () => {
    const filter: INumericFilterModel = {
      type: EColumnFilterType.Numeric,
      name: WorkloadSortFilterFields.AllocatedGpu,
      label: "GPU compute allocation",
      operator: EFilterOperator.LessThanOrEqual,
      minValue: 5,
      maxValue: 0,
    };
    await mountComponent(filter);
    await clickOnSave();

    expect(wrapper.emitted("save")?.[0][0]).toEqual({
      label: "GPU compute allocation",
      maxValue: 0,
      minValue: 5,
      name: "allocatedGPU",
      operator: "<=",
      type: "Numeric",
    });
  });

  test("emits 'close' event when saveFilter method is called with an empty term", async () => {
    const filter: INumericFilterModel = {
      type: EColumnFilterType.Numeric,
      name: WorkloadSortFilterFields.AllocatedGpu,
      label: "GPU compute allocation",
      operator: EFilterOperator.LessThanOrEqual,
      minValue: 5,
      maxValue: 0,
    };
    await mountComponent(filter);
    await clickOnClose();

    expect(wrapper.emitted("close")).toBeTruthy();
  });
});
