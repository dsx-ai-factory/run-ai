/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { describe, expect, it, vi } from "vitest";
import HybridEnvironmentSection from "./hybrid-environment-section.vue";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

function createWrapper(isDistributedInference = true, supportImagePullSecrets = true) {
  return shallowMount(HybridEnvironmentSection, {
    props: {
      isDistributedInference,
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
  it.each([
    {
      source: ERunaiCredentialSource.MyCredentials,
      imagePull: { name: "personal-key", userCredential: true },
      runtime: { userCredential: { name: "runtime-key", key: "NGC_API_KEY" } },
    },
    {
      source: ERunaiCredentialSource.SharedSecret,
      imagePull: { name: "personal-key" },
      runtime: { secret: { name: "runtime-key", key: "NGC_API_KEY" } },
    },
  ])("supplies the $source credential at pull time and runtime", ({ source, imagePull, runtime }) => {
    const wrapper = createWrapper();
    const imageOrigin = wrapper.findComponent({ name: "form-image-origin-section" });
    expect(imageOrigin.props("nimOnly")).toBe(true);
    imageOrigin.vm.$emit("credential-changed", {
      credentialName: "personal-key",
      credentialSource: source,
      runtimeCredentialName: "runtime-key",
    });
    expect(wrapper.emitted("update")?.at(-1)?.[0]).toEqual({
      image: null,
      imagePullSecrets: [imagePull],
      environmentVariables: [
        { name: "EXISTING", value: "keep" },
        { name: "NGC_API_KEY", ...runtime },
      ],
    });
  });

  it("removes a cleared NGC credential from both the runtime and image-pull configuration", async () => {
    const wrapper = createWrapper();
    await wrapper.setProps({
      workloadSpec: {
        image: "nvcr.io/nim/nvidia/model:latest",
        imagePullSecrets: [{ name: "old-key", userCredential: true }],
        environmentVariables: [
          { name: "EXISTING", value: "keep" },
          { name: "NGC_API_KEY", userCredential: { name: "old-key", key: "NGC_API_KEY" } },
        ],
      },
    });
    wrapper.findComponent({ name: "form-image-origin-section" }).vm.$emit("credential-changed", {
      credentialName: "",
      credentialSource: ERunaiCredentialSource.MyCredentials,
      runtimeCredentialName: "",
    });
    expect(wrapper.emitted("update")?.at(-1)?.[0]).toEqual({
      image: null,
      imagePullSecrets: [],
      environmentVariables: [{ name: "EXISTING", value: "keep" }],
    });
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
    expect(wrapper.findComponent({ name: "form-image-origin-section" }).props("nimOnly")).toBe(false);
    wrapper.vm.onCredentialChanged({
      credentialName: "image-key",
      credentialSource: ERunaiCredentialSource.MyCredentials,
    });
    expect(wrapper.emitted("update")?.at(-1)?.[0]).toMatchObject({
      image: "mirror.example.com/nim/model:latest",
      environmentVariables: [{ name: "EXISTING", value: "keep" }],
    });
  });

  it("supplies the runtime key even when image-pull secrets are disabled", () => {
    const wrapper = createWrapper(true, false);
    wrapper.vm.onCredentialChanged({
      credentialName: "personal-key",
      credentialSource: ERunaiCredentialSource.MyCredentials,
      runtimeCredentialName: "personal-key",
    });
    expect(wrapper.emitted("update")?.at(-1)?.[0]).toMatchObject({
      environmentVariables: [
        { name: "EXISTING", value: "keep" },
        { name: "NGC_API_KEY", userCredential: { name: "personal-key", key: "NGC_API_KEY" } },
      ],
    });
  });
});
