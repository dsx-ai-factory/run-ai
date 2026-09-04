/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import MemoryResourceBox from "./memory-resource-box.vue";
import { afterEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import { basicMemoryUnitOptions } from "@/models/global.model";
import type { ICpuMemory } from "@/models/compute-resource.model";
import type { TTestEmit } from "@/models/unit-test.model";

describe("memory-resource-box.vue", () => {
  let memoryResourceBoxCmp: VueWrapper<InstanceType<typeof MemoryResourceBox>>;

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory: {
          request: "0M",
          limit: null,
        },
      },
    });

    expect(memoryResourceBoxCmp.html()).toContain("CPU memory");
    expect(memoryResourceBoxCmp.classes()).toContain("memory-box");
  });

  test("should emit the request with value and limit as null when limit is off", async () => {
    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory: {
          request: "0M",
          limit: null,
        },
      },
    });

    const cpuMemoryRequestVal = 2;
    memoryResourceBoxCmp.find(".cpu-memory-request-input").setValue(cpuMemoryRequestVal);
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];
    expect(emittedVal[0][0].request).toEqual(`${cpuMemoryRequestVal}M`);
    expect(emittedVal[0][0].limit).toEqual(null);
  });

  test("should emit the limit as null when limit toggle is on and request is zero", () => {
    const cpuMemory: ICpuMemory = { request: "0M", limit: null };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    memoryResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];

    expect(emittedVal[0][0].request).toEqual(cpuMemory.request);
    expect(emittedVal[0][0].limit).toEqual(null);
  });

  test("should emit the request and limit as the same value when limit is on and request is not a zero", () => {
    const cpuMemory: ICpuMemory = { request: "5M", limit: null };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    memoryResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];
    expect(emittedVal[0][0].request).toEqual(cpuMemory.request);
    expect(emittedVal[0][0].limit).toEqual(cpuMemory.request);
  });

  test("should update the limit when user enter number in the request that is bigger then the limit", () => {
    const initValue = 5;
    const expectedValue = initValue + 1;
    const cpuMemory: ICpuMemory = { request: `${initValue}M`, limit: `${initValue}M` };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    memoryResourceBoxCmp.find(".cpu-memory-request-input").setValue(initValue + 1);
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];
    expect(emittedVal[0][0].request).toEqual(`${expectedValue}M`);
    expect(emittedVal[0][0].limit).toEqual(`${expectedValue}M`);
  });

  test("should emit the limit value when user change the limit value", () => {
    const initValue = 5;
    const expectedValue = initValue + 1;
    const cpuMemory: ICpuMemory = { request: `${initValue}M`, limit: `${initValue}M` };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    memoryResourceBoxCmp.find(".cpu-memory-limit-input").setValue(expectedValue);
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];
    expect(emittedVal[0][0].request).toEqual(`${initValue}M`);
    expect(emittedVal[0][0].limit).toEqual(`${expectedValue}M`);
  });

  test("should set the limit to undefined when turning the toggle limit off", () => {
    const initValue = 5;
    const cpuMemory: ICpuMemory = { request: `${initValue}M`, limit: `${initValue}M` };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    memoryResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];
    expect(emittedVal[0][0].request).toEqual(`${initValue}M`);
    expect(emittedVal[0][0].limit).toEqual(undefined);
  });

  test("should set the limit to null when the toggle limit is on and there is no value", () => {
    const initValue = 0;
    const cpuMemory: ICpuMemory = { request: `${initValue}M`, limit: undefined };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    memoryResourceBoxCmp.find(".limit-toggle").trigger("click");
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];

    expect(emittedVal[0][0].request).toEqual(`${initValue}M`);
    expect(emittedVal[0][0].limit).toEqual(null);
  });

  test("should emit is-section-invalid as true when limit value is lower then the request", () => {
    const cpuMemory: ICpuMemory = { request: `7M`, limit: `4M` };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    const emittedVal: TTestEmit<boolean> = memoryResourceBoxCmp.emitted<Array<boolean>>("is-section-invalid") || [[]];
    expect(emittedVal[0][0]).toEqual(true);
  });

  test("should emit is-section-invalid as true when limit is on and no value is set", async () => {
    const cpuMemory: ICpuMemory = { request: `7M`, limit: null };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    memoryResourceBoxCmp.find(".limit-toggle").trigger("click");
    await memoryResourceBoxCmp.vm.$nextTick();
    memoryResourceBoxCmp.find(".cpu-memory-limit-input").setValue(null);

    // test the cpu-core-changed emit
    const emittedVal: TTestEmit<ICpuMemory> = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [
      [],
    ];
    const secondEmit = emittedVal[1][0];
    expect(secondEmit.request).toEqual(cpuMemory.request);
    expect(secondEmit.limit).toEqual(null);

    // test the is-secion-valid emit
    await memoryResourceBoxCmp.vm.$nextTick();
    const emittedInvalidVal: TTestEmit<boolean> = memoryResourceBoxCmp.emitted<Array<boolean>>("is-section-invalid") || [
      [],
    ];
    const secondInvalidEmit = emittedInvalidVal[1][0];
    expect(secondInvalidEmit).toEqual(true);
  });

  test("should emit 'update-cpu-request' with a converted value unit upon a selection of a new unit", async () => {
    const value = 5;
    const cpuMemory: ICpuMemory = { request: `${value}M`, limit: `${value}M` };

    memoryResourceBoxCmp = mount(MemoryResourceBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        cpuMemory,
      },
    });

    await (memoryResourceBoxCmp.getComponent(".unit-select") as VueWrapper).vm.$emit(
      "update:model-value",
      basicMemoryUnitOptions[1],
    );
    await memoryResourceBoxCmp.vm.$nextTick();
    const emittedVal = memoryResourceBoxCmp.emitted<Array<ICpuMemory>>("cpu-memory-changed") || [[]];
    expect(emittedVal[0][0].request).toEqual(`${value}G`);
    expect(emittedVal[0][0].limit).toEqual(`${value}G`);
  });
});
