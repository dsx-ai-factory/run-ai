/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { describe, expect, it, vi } from "vitest";
import HybridEnvironmentSection from "./hybrid-environment-section.vue";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

function createWrapper(authenticatedNgcOnly = true, supportImagePullSecrets = true) {
  return shallowMount(HybridEnvironmentSection, {
    props: {
      authenticatedNgcOnly,
      supportImagePullSecrets,
      supportAuthenticatedNgc: true,
      workloadSpec: {
        image: "mirror.example.com/nim/model:latest",
        environmentVariables: [{ name: "EXISTING", value: "keep" }],
      },
    },
    global: {
      plugins: [createTestingPinia({ createSpy: vi.fn })],
      stubs: { HybridSectionWrapper: { template: "<div><slot /></div>" } },
    },
  });
}

describe("distributed inference NGC credentials", () => {
  it("requires authenticated catalog browsing for distributed inference", () => {
    const wrapper = createWrapper();
    expect(wrapper.findComponent({ name: "form-image-origin-section" }).props("authenticatedOnly")).toBe(true);
  });

  it("leaves custom mirror images and their existing runtime configuration untouched", () => {
    const wrapper = createWrapper();
    expect(wrapper.emitted("update")).toBeUndefined();
    wrapper
      .findComponent({ name: "form-image-origin-section" })
      .vm.$emit("update:model-value", "mirror.example.com/model:v2");
    expect(wrapper.emitted("update")?.at(-1)?.[0]).toEqual({
      image: "mirror.example.com/model:v2",
      environmentVariables: [{ name: "EXISTING", value: "keep" }],
    });
  });

  it("keeps ordinary workloads outside the NIM catalog requirement", () => {
    const wrapper = createWrapper(false);
    expect(wrapper.findComponent({ name: "form-image-origin-section" }).props("authenticatedOnly")).toBe(false);
    wrapper.vm.onCredentialChanged({
      credentialName: "image-key",
      credentialSource: ERunaiCredentialSource.MyCredentials,
    });
    expect(wrapper.emitted("update")?.at(-1)?.[0]).toMatchObject({
      image: "mirror.example.com/nim/model:latest",
      environmentVariables: [{ name: "EXISTING", value: "keep" }],
    });
  });
});
