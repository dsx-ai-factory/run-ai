/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import CpuResourceBox, { CONVERSION_VALUE } from "./cpu-resource-box.vue";
import type { TTestEmit } from "@/models/unit-test.model";
import type { ICpuCore } from "@/models/compute-resource.model";

describe("cpu-resource-box.vue", () => {
  let cpuResourceBoxCmp: VueWrapper<InstanceType<typeof CpuResourceBox>>;

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("shoud load the component properly", () => {
    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore: {
          request: 0,
          limit: null,
        },
      },
    });

    expect(cpuResourceBoxCmp.html()).toContain("CPU");
    expect(cpuResourceBoxCmp.classes()).toContain("cpu-box");
  });

  test("should emit the request with value and limit as undefined when limit is off", () => {
    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore: {
          request: 0,
          limit: undefined,
        },
      },
    });

    const cpuRequestVal = 2;
    cpuResourceBoxCmp.find(".cpu-request-input").setValue(cpuRequestVal);
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual(cpuRequestVal);
    expect(emittedVal[0][0].limit).toEqual(undefined);
  });

  test("should emit the limit as null when limit toggle is on and request is zero", () => {
    const cpuCore: ICpuCore = { request: 0, limit: null };

    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    cpuResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];

    expect(emittedVal[0][0].request).toEqual(cpuCore.request);
    expect(emittedVal[0][0].limit).toEqual(null);
  });

  test("should emit the request and limit as the same value when limit is on and request is not a zero", () => {
    const cpuCore: ICpuCore = { request: 5, limit: null };

    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    cpuResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual(cpuCore.request);
    expect(emittedVal[0][0].limit).toEqual(cpuCore.request);
  });

  test("should update the limit when user enter number in the request that is bigger then the limit", () => {
    const cpuCore: ICpuCore = { request: 5, limit: 5 };

    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    cpuResourceBoxCmp.find(".cpu-request-input").setValue((cpuCore.request || 0) + 1);
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual((cpuCore.request || 0) + 1);
    expect(emittedVal[0][0].limit).toEqual((cpuCore.request || 0) + 1);
  });

  test("should emit the limit value when user change the limit value", () => {
    const cpuCore: ICpuCore = { request: 5, limit: 5 };

    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    cpuResourceBoxCmp.find(".cpu-limit-input").setValue((cpuCore.request || 0) + 1);
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual(cpuCore.request);
    expect(emittedVal[0][0].limit).toEqual((cpuCore.request || 0) + 1);
  });

  test("should set the limit as undefined when turning the toggle limit off", () => {
    const cpuCore: ICpuCore = { request: 5, limit: 5 };

    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    cpuResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual(cpuCore.request);
    expect(emittedVal[0][0].limit).toEqual(undefined);
  });

  test("should set the limit to null when the toggle limit is on and there is no value", () => {
    const cpuCore: ICpuCore = { request: 0, limit: undefined };

    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    cpuResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual(cpuCore.request);
    expect(emittedVal[0][0].limit).toEqual(null);
  });

  test("should emit is-section-invalid as true when limit value is lower then the request", () => {
    const cpuCore: ICpuCore = { request: 7, limit: 4 };

    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    const emittedVal: TTestEmit<boolean> = cpuResourceBoxCmp.emitted<Array<boolean>>("is-section-invalid") || [[]];
    expect(emittedVal[0][0]).toEqual(true);
  });

  test("should emit is-section-invalid as true when limit is on and no value is set", async () => {
    const cpuCore: ICpuCore = { request: 7, limit: null };
    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    cpuResourceBoxCmp.find(".limit-toggle").trigger("click");
    await cpuResourceBoxCmp.vm.$nextTick();
    cpuResourceBoxCmp.find(".cpu-limit-input").setValue(null);

    // test the cpu-core-changed emit
    const emittedVal: TTestEmit<ICpuCore> = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    const secondEmit = emittedVal[1][0];
    expect(secondEmit.request).toEqual(cpuCore.request);
    expect(secondEmit.limit).toEqual(null);

    // test the is-secion-valid emit
    await cpuResourceBoxCmp.vm.$nextTick();
    const emittedInvalidVal: TTestEmit<boolean> = cpuResourceBoxCmp.emitted<Array<boolean>>("is-section-invalid") || [
      [],
    ];
    const secondInvalidEmit = emittedInvalidVal[1][0];
    expect(secondInvalidEmit).toEqual(true);
  });

  test("should emit 'update-cpu-request' with a converted value unit upon a selection of a new unit", async () => {
    const cpuCore: ICpuCore = { request: 5, limit: 7 };
    cpuResourceBoxCmp = mount(CpuResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuCore,
      },
    });

    const selectedOption = { label: "Millicores", value: CONVERSION_VALUE };
    await (cpuResourceBoxCmp.getComponent(".unit-select") as VueWrapper).vm.$emit("update:model-value", selectedOption);
    await cpuResourceBoxCmp.vm.$nextTick();
    const emittedVal = cpuResourceBoxCmp.emitted<Array<ICpuCore>>("cpu-core-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual((cpuCore.request || 0) / CONVERSION_VALUE);
    expect(emittedVal[0][0].limit).toEqual(Number(cpuCore.limit) / CONVERSION_VALUE);
  });
});
