/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import FormNgcCatalogImageSection from "./form-ngc-catalog-image-section.vue";
import {
  ERunaiCredentialSource,
  ERunaiCredentialType,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { AssetKind, type CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";

function createWrapper(nimOnly = true) {
  return shallowMount(FormNgcCatalogImageSection, {
    props: { nimOnly, supportAuthenticatedNgc: true, ngcCatalogRegistryId: "ngc-registry" },
  });
}

describe("NIM catalog authentication", () => {
  it("requires a credential before showing the NIM image picker and offers no guest option", async () => {
    const wrapper = createWrapper();
    expect(wrapper.findComponent({ name: "runai-radio-options" }).exists()).toBe(false);
    const credentialSelector = wrapper.findComponent({ name: "runai-credential-selector-by-source" });
    expect(credentialSelector.props("credentialSelectConfig").required).toBe(true);
    expect(wrapper.findComponent({ name: "ngc-image-tag-selector" }).exists()).toBe(false);

    credentialSelector.vm.$emit("update:model-value", {
      id: "personal-key-id",
      credentialName: "personal-key",
      source: ERunaiCredentialSource.MyCredentials,
      type: ERunaiCredentialType.NgcApiKey,
    });
    await wrapper.vm.$nextTick();

    const imageSelector = wrapper.findComponent({ name: "ngc-image-tag-selector" });
    expect(imageSelector.props()).toMatchObject({ credentialId: "personal-key-id", nimOnly: true });
    expect(wrapper.emitted("credential-changed")?.at(-1)?.[0]).toEqual({
      credentialName: "personal-key",
      credentialSource: ERunaiCredentialSource.MyCredentials,
      runtimeCredentialName: "personal-key",
    });

    credentialSelector.vm.$emit("update:model-value", {
      credentialName: "",
      source: ERunaiCredentialSource.MyCredentials,
      type: ERunaiCredentialType.NgcApiKey,
    });
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent({ name: "ngc-image-tag-selector" }).exists()).toBe(false);
    expect(wrapper.emitted("update:model-value")?.at(-1)).toEqual([null]);
    expect(wrapper.emitted("credential-changed")?.at(-1)?.[0]).toMatchObject({ runtimeCredentialName: "" });
  });

  it("resolves separate image-pull and runtime secrets for a shared NGC credential", () => {
    const wrapper = createWrapper();
    const credential = {
      meta: { id: "shared-key-id", name: "shared-key", kind: AssetKind.NgcApiKey, private: false },
      clusterInfo: {
        resources: [{ name: "shared-key-dockerregistry-123" }, { name: "shared-key-genericsecret-456" }],
      },
    } as CredentialsListResponseEntry;
    const imagePullName = wrapper.vm.resolveCredentialName(credential);
    expect(imagePullName).toBe("shared-key-dockerregistry-123");
    wrapper.vm.onCredentialSelectorChanged({
      id: credential.meta.id,
      credentialName: imagePullName,
      source: ERunaiCredentialSource.SharedSecret,
      type: ERunaiCredentialType.NgcApiKey,
    });
    expect(wrapper.emitted("credential-changed")?.at(-1)?.[0]).toEqual({
      credentialName: "shared-key-dockerregistry-123",
      credentialSource: ERunaiCredentialSource.SharedSecret,
      runtimeCredentialName: "shared-key-genericsecret-456",
    });
  });

  it("preserves guest catalog browsing for other workload forms", () => {
    const wrapper = createWrapper(false);
    const options = wrapper.findComponent({ name: "runai-radio-options" }).props("options");
    expect(options).toContainEqual({ value: "guest", label: "As a guest" });
    expect(wrapper.findComponent({ name: "ngc-image-tag-selector" }).exists()).toBe(true);
    expect(wrapper.findComponent({ name: "runai-credential-selector-by-source" }).exists()).toBe(false);
  });
});
