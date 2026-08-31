/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import ClusterNameSection from "./cluster-name-section.vue";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import type { DefineComponent } from "vue";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";

createTestingPinia({ createSpy: vi.fn });

describe("cluster-name-section.vue", () => {
  let clusterNameSectionCmp: VueWrapper<InstanceType<typeof ClusterNameSection>>;

  beforeEach(() => {
    clusterNameSectionCmp = mount(ClusterNameSection, {
      global: {
        plugins: [Quasar],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        modelValue: "",
        projectId: 1,
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(clusterNameSectionCmp.html()).toContain("Cluster name");
    expect(clusterNameSectionCmp.classes()).toContain("cluster-name");
  });

  test("emits 'update:modelValue' event when 'modelValue' changes", async () => {
    const clusterName = "new-value";
    await clusterNameSectionCmp.find("input").setValue(clusterName);
    const expectedEmits: string[][] = [[clusterName]];
    expect(clusterNameSectionCmp.emitted("update:modelValue")).toEqual(expectedEmits);
  });

  test("emits 'is-section-invalid' events when 'modelValue' changed to invalid and vice-versa", async () => {
    const clusterName = "new-test-name-for-model";
    await clusterNameSectionCmp.setProps({ modelValue: clusterName });
    const expectedEmits: boolean[][] = [[true], [false]]; // firstly emits true when component created with defaullt data, afterwards emits when data changes
    expect(clusterNameSectionCmp.emitted("is-section-invalid")).toEqual(expectedEmits);
  });

  test("sets nameAlreadyExist to true when 'invalid' event is emitted by runai-name-validation", () => {
    expect(clusterNameSectionCmp.vm.nameAlreadyExist).toBe(false);

    const runaiNameValidation = clusterNameSectionCmp.findComponent<DefineComponent>(".runai-name-validation");
    runaiNameValidation.vm.$emit("invalid", true);
    expect(clusterNameSectionCmp.vm.nameAlreadyExist).toBe(true);
  });
});
