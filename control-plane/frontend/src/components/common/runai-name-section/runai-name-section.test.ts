/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import RunaiNameSection from "./runai-name-section.vue";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Scope } from "@/swagger-models/assets-service-client";

vi.mock("@/services/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("runai-name-section.vue", () => {
  let ComputeResourceNameSectionCmp: VueWrapper<InstanceType<typeof RunaiNameSection>>;

  beforeEach(() => {
    ComputeResourceNameSectionCmp = mount(RunaiNameSection, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        modelValue: "",
        label: "Compute resource",
        assetKind: "compute",
        scopeModel: {
          scope: Scope.Project,
          projectId: 1,
        },
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(ComputeResourceNameSectionCmp.html()).toContain("Compute resource name");
    expect(ComputeResourceNameSectionCmp.classes()).toContain("runai-name-section");
  });

  test("emits 'update:modelValue' event when 'modelValue' changes", async () => {
    const computeResourceName = "new-value";
    await ComputeResourceNameSectionCmp.find("input").setValue(computeResourceName);
    const expectedEmits: string[][] = [[computeResourceName]];
    expect(ComputeResourceNameSectionCmp.emitted("update:modelValue")).toEqual(expectedEmits);
  });
});
