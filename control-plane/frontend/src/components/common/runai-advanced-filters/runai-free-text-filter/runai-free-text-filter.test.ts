/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, test } from "vitest";
import type { IFreeTextFilterModel } from "@/models/filter.model";
import { EColumnFilterType, EFilterOperator, EFilterOperatorLabel, filterOperatorMap } from "@/models/filter.model";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import RunaiFreeTextFilter from "@/components/common/runai-advanced-filters/runai-free-text-filter/runai-free-text-filter.vue";
import { Quasar } from "quasar";
import { unitTestUtil } from "@/utils/unit-test.util";

describe("RunaiFreeTextFilter", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiFreeTextFilter>>;
  const mountComponent = async (filter: IFreeTextFilterModel): Promise<void> => {
    wrapper = mount(RunaiFreeTextFilter, {
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
    const filter: IFreeTextFilterModel = {
      type: EColumnFilterType.FreeText,
      name: "name",
      label: "Workload",
      term: "workspace-ds2423e",
      value: EFilterOperator.Contains,
    };
    await mountComponent(filter);
    const menuElement = unitTestUtil.getMenuElement();
    expect(wrapper.exists()).toBe(true);
    expect(menuElement.find("[aid=free-text-filter-select]").text()).toBe(filterOperatorMap[EFilterOperator.Contains]);
    expect(menuElement.find("[aid=free-text-filter-input]").exists()).toBe(true);
    expect(wrapper.vm.term).eql(filter.term);
  });

  test("it should validate that initial term is empty and default operator label is contains  ", async () => {
    const filter: IFreeTextFilterModel = {
      type: EColumnFilterType.FreeText,
      name: "name",
      label: "Workload",
      term: "",
      value: EFilterOperator.Contains,
    };
    await mountComponent(filter);
    const menuElement = unitTestUtil.getMenuElement();
    expect(menuElement.find("[aid=free-text-filter-select]").text()).toBe(EFilterOperatorLabel.Contains);
    expect(wrapper.vm.term).toBe("");
  });

  test.each([
    [EFilterOperator.Contains, EFilterOperatorLabel.Contains],
    [EFilterOperator.DoesNotContain, EFilterOperatorLabel.DoesNotContain],
    [EFilterOperator.Equals, EFilterOperatorLabel.Equals],
    [EFilterOperator.NotEquals, EFilterOperatorLabel.NotEquals],
  ])(
    "it should validate that select display correct label for operator %s",
    async (operator: EFilterOperator, expectedLabel: EFilterOperatorLabel) => {
      const filter: IFreeTextFilterModel = {
        type: EColumnFilterType.FreeText,
        name: "name",
        label: "Workload",
        term: "workspace-ds2423e",
        value: operator,
      };
      await mountComponent(filter);

      const menuElement = unitTestUtil.getMenuElement();
      expect(menuElement.find("[aid=free-text-filter-select]").text()).toBe(expectedLabel);
    },
  );

  test("emits 'save' event when saveFilter method is called with a non-empty term", async () => {
    const filter: IFreeTextFilterModel = {
      label: "Workload",
      term: "",
      value: EFilterOperator.Contains,
      name: "name",
      type: EColumnFilterType.FreeText,
    };
    await mountComponent(filter);

    wrapper.vm.saveFilter();

    expect(wrapper.emitted("save")).toBeFalsy();

    await wrapper.setData({ term: "training" });

    wrapper.vm.saveFilter();

    expect(wrapper.emitted("save")).toBeTruthy();
    expect(wrapper.emitted("save")?.[0][0]).toEqual({
      name: "name",
      term: "training",
      label: "Workload",
      type: EColumnFilterType.FreeText,
      value: EFilterOperator.Contains,
    });
  });

  test("disables the operator select and uses the only remaining operator when excludeOperators leaves a single option", async () => {
    const filter: IFreeTextFilterModel = {
      type: EColumnFilterType.FreeText,
      name: "project",
      label: "Project",
      term: "",
      value: EFilterOperator.Contains,
      excludeOperators: [EFilterOperator.DoesNotContain, EFilterOperator.NotEquals, EFilterOperator.Contains],
    };
    await mountComponent(filter);
    const qSelect = wrapper.findComponent({ name: "QSelect" });
    expect(qSelect.props("disable")).toBe(true);
    expect(wrapper.vm.selectedOption.value).toBe(EFilterOperator.Equals);
    expect(wrapper.vm.isOperatorSelectDisabled).toBe(true);
  });

  test("emits 'close' event when saveFilter method is called with an empty term", async () => {
    const filter: IFreeTextFilterModel = {
      label: "Workload",
      term: "",
      value: EFilterOperator.Contains,
      name: "name",
      type: EColumnFilterType.FreeText,
    };
    await mountComponent(filter);

    wrapper.vm.saveFilter();

    expect(wrapper.emitted("close")).toBeTruthy();
  });
});
