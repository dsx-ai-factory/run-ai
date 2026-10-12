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

function createWrapper(authenticatedOnly = true) {
  return shallowMount(FormNgcCatalogImageSection, {
    props: { authenticatedOnly, supportAuthenticatedNgc: true, ngcCatalogRegistryId: "ngc-registry" },
  });
}

describe("Distributed catalog authentication", () => {
  it("requires a credential before showing the catalog image picker and offers no guest option", async () => {
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
    expect(imageSelector.props()).toMatchObject({ credentialId: "personal-key-id" });
    expect(wrapper.emitted("credential-changed")?.at(-1)?.[0]).toEqual({
      credentialName: "personal-key",
      credentialSource: ERunaiCredentialSource.MyCredentials,
    });

    credentialSelector.vm.$emit("update:model-value", {
      credentialName: "",
      source: ERunaiCredentialSource.MyCredentials,
      type: ERunaiCredentialType.NgcApiKey,
    });
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent({ name: "ngc-image-tag-selector" }).exists()).toBe(false);
    expect(wrapper.emitted("update:model-value")?.at(-1)).toEqual([null]);
    expect(wrapper.emitted("credential-changed")?.at(-1)?.[0]).toEqual({
      credentialName: "",
      credentialSource: ERunaiCredentialSource.MyCredentials,
    });
  });

  it("uses the selected project scope for shared-credential browsing", () => {
    const wrapper = shallowMount(FormNgcCatalogImageSection, {
      props: { authenticatedOnly: true, scopeId: "42", scopeType: "project", clusterId: "test-cluster" },
    });
    expect(wrapper.findComponent({ name: "runai-credential-selector-by-source" }).props("scopeModel")).toEqual({
      scope: null,
      projectId: 42,
      clusterId: "test-cluster",
      departmentId: undefined,
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
