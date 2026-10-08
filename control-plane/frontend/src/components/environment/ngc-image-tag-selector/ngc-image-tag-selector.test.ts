/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test, vi } from "vitest";
import { NgcImageTagSelector } from ".";

const { getImages } = vi.hoisted(() => ({
  getImages: vi.fn(),
}));

vi.mock("@/services/control-plane/catalog.service/catalog.service", () => ({
  catalogService: {
    getImages,
    getImageTags: vi.fn(),
  },
}));

describe("ngc-image-tag-selector", () => {
  beforeEach(() => {
    getImages.mockReset();
    getImages.mockResolvedValue({
      images: ["nvidia/pytorch"],
      prefix: "nvcr.io/",
    });
  });

  test("passes the image search query to the catalog service", async () => {
    const wrapper = mount(NgcImageTagSelector, {
      global: {
        plugins: [Quasar],
      },
      props: {
        registryId: "registry-id",
        credentialId: "credential-id",
      },
    });
    await flushPromises();
    getImages.mockClear();

    wrapper.vm.onImageFilter("nvidia/pytorch", (update) => update());
    await flushPromises();

    expect(getImages).toHaveBeenCalledWith("registry-id", "nvidia/pytorch", "credential-id");
    expect(wrapper.vm.imageOptions).toEqual([{ label: "nvidia/pytorch", value: "nvidia/pytorch" }]);
  });
  test("limits NIM catalog browsing to the NIM namespace", async () => {
    getImages.mockResolvedValue({
      images: ["nvidia/pytorch", "nim/nvidia/llama", "nim/other/model", "not-nim/model"],
      prefix: "nvcr.io/",
    });
    const wrapper = mount(NgcImageTagSelector, {
      global: { plugins: [Quasar] },
      props: { registryId: "registry-id", credentialId: "credential-id", nimOnly: true },
    });
    await flushPromises();
    expect(getImages).toHaveBeenCalledWith("registry-id", "nim/", "credential-id");
    expect(wrapper.vm.imageOptions).toEqual([
      { label: "nim/nvidia/llama", value: "nim/nvidia/llama" },
      { label: "nim/other/model", value: "nim/other/model" },
    ]);
    wrapper.vm.onImageFilter("llama", (update) => update());
    await flushPromises();
    expect(getImages).toHaveBeenLastCalledWith("registry-id", "llama", "credential-id");
    expect(wrapper.vm.imageOptions.some((option) => option.value === "nvidia/pytorch")).toBe(false);
  });
});
