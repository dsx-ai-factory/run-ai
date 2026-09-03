/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { GpuResourceBox } from ".";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import type { IGpuRequestModel } from "@/models/compute-resource.model";
import { GpuRequestType } from "@/swagger-models/assets-service-client";
import type { TTestEmit } from "@/models/unit-test.model";
import { useSettingStore } from "@/stores/setting.store";
import { SettingKeys, type TSettingValue } from "@/models/setting.model";

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("gpu-resource-box.vue", () => {
  let gpuResourceBoxCmp: VueWrapper<InstanceType<typeof GpuResourceBox>>;

  beforeEach(() => {
    const gpuRequest: IGpuRequestModel = {
      gpuDevicesRequest: 1,
      gpuPortionRequest: 0,
      gpuRequestType: GpuRequestType.Portion,
    };
    const settingStore = useSettingStore();
    settingStore.$patch({
      settings: new Map([[SettingKeys.GpuOverProvisioning, { value: true } as unknown as TSettingValue]]),
      gpuOverProvisioning: { value: true },
    });

    gpuResourceBoxCmp = mount(GpuResourceBox, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        gpuRequest,
      },
      data() {
        return {
          settingStore,
        };
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(gpuResourceBoxCmp.html()).toContain("GPU memory");
    expect(gpuResourceBoxCmp.classes()).toContain("gpu-resource-box");
  });

  test("selecting anything will reveal the the request input and the limit input", async () => {
    await gpuResourceBoxCmp.setProps({
      gpuRequest: { gpuDevicesRequest: 1, gpuRequestType: GpuRequestType.Portion, gpuPortionRequest: 0 },
    });
    expect(gpuResourceBoxCmp.find("[aid=gpu-request-input]").exists()).toBe(true);
    expect(gpuResourceBoxCmp.find("[aid=gpu-limit-input]").exists()).toBe(true);
  });

  test("updating the portion request will not change the limit if it is not enabled", async () => {
    await gpuResourceBoxCmp.find('[aid="gpu-request-input"]').setValue(1);
    const expectedEmit: TTestEmit<IGpuRequestModel> = [
      [{ gpuDevicesRequest: 1, gpuPortionRequest: 0.01, gpuRequestType: GpuRequestType.Portion }],
    ];
    expect(gpuResourceBoxCmp.emitted("update-gpu-request")).toEqual(expectedEmit);
  });

  test("updating the portion request will change the limit if it is enabled", async () => {
    await gpuResourceBoxCmp.setProps({
      gpuRequest: {
        gpuDevicesRequest: 1,
        gpuRequestType: GpuRequestType.Portion,
        gpuPortionRequest: 0,
        gpuPortionLimit: 0,
      },
    });
    await gpuResourceBoxCmp.find('[aid="gpu-request-input"]').setValue(1);
    const expectedEmit: TTestEmit<IGpuRequestModel> = [
      [{ gpuDevicesRequest: 1, gpuPortionRequest: 0.01, gpuPortionLimit: 0.01, gpuRequestType: GpuRequestType.Portion }],
    ];
    expect(gpuResourceBoxCmp.emitted("update-gpu-request")).toEqual(expectedEmit);
  });

  test("updating the memory request will not change the limit if it is not enabled", async () => {
    await gpuResourceBoxCmp.setProps({
      gpuRequest: { gpuDevicesRequest: 1, gpuRequestType: GpuRequestType.Memory, gpuMemoryRequest: "0M" },
    });
    await gpuResourceBoxCmp.find('[aid="gpu-request-input"]').setValue(1);
    const expectedEmit: TTestEmit<IGpuRequestModel> = [
      [{ gpuDevicesRequest: 1, gpuMemoryRequest: "1M", gpuRequestType: GpuRequestType.Memory }],
    ];
    expect(gpuResourceBoxCmp.emitted("update-gpu-request")).toEqual(expectedEmit);
  });

  test("updating the memory request will change the limit if it is enabled", async () => {
    await gpuResourceBoxCmp.setProps({
      gpuRequest: {
        gpuDevicesRequest: 1,
        gpuRequestType: GpuRequestType.Memory,
        gpuMemoryRequest: "0M",
        gpuMemoryLimit: "0M",
      },
    });
    await gpuResourceBoxCmp.find('[aid="gpu-request-input"]').setValue(1);
    const expectedEmit: TTestEmit<IGpuRequestModel> = [
      [{ gpuDevicesRequest: 1, gpuMemoryRequest: "1M", gpuMemoryLimit: "1M", gpuRequestType: GpuRequestType.Memory }],
    ];
    expect(gpuResourceBoxCmp.emitted("update-gpu-request")).toEqual(expectedEmit);
  });
});
