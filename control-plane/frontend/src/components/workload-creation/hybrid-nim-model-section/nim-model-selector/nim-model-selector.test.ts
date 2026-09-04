/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import NimModelSelector from "./nim-model-selector.vue";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { Quasar } from "quasar";
import type { TTestEmit } from "@/models/unit-test.model";
import { errorMessages } from "@/common/error-message.constant";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";

vi.mock("@/services/control-plane/inference.service/inference.service");

const mockNotify = vi.fn();
const $q = {
  notify: mockNotify,
  platform: {
    has: { touch: false },
  },
};

const modelListResponse = {
  resources: [{ resourceId: "nvidia/parakeet-ctc-1.1b-asr" }, { resourceId: "meta/llama-3.1-8b-base" }],
  isStaticData: false,
};

describe("nim-model-selector.vue", () => {
  const createWrapper = function (props = { modelName: "" }) {
    return mount(NimModelSelector, {
      global: {
        plugins: [Quasar],
        mocks: {
          $q,
        },
      },
      props,
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders correctly", async () => {
    vi.mocked(inferenceService.getNimCatalogModels).mockResolvedValue(modelListResponse);
    const wrapper = createWrapper();
    await flushPromises();
    expect(wrapper.find(".nim-model-selector").exists()).toBe(true);
    expect(wrapper.html()).toContain("Select or type the model name");
    expect(wrapper.find("[aid='nim-autocomplete-select']").exists()).toBe(true);
  });

  it("validates empty model name", async () => {
    vi.mocked(inferenceService.getNimCatalogModels).mockResolvedValue(modelListResponse);
    const wrapper = createWrapper();
    await flushPromises();
    const isNameEmpty = wrapper.vm.isNameEmpty("");
    expect(isNameEmpty).toBe(errorMessages.SELECT_OR_TYPE_A_NAME);
  });

  it("validates incorrect model name format", async () => {
    vi.mocked(inferenceService.getNimCatalogModels).mockResolvedValue(modelListResponse);
    const wrapper = createWrapper();
    await flushPromises();
    const nimFormatValidator = wrapper.vm.nimFormatValidator("invalid@model");
    expect(nimFormatValidator).toBe(errorMessages.ENTER_A_VALID_NAME);
  });

  it("calls getNimModels and returns filtered models from API", async () => {
    vi.mocked(inferenceService.getNimCatalogModels).mockResolvedValue(modelListResponse);
    const wrapper = createWrapper();
    await flushPromises();
    const filteredModels = await wrapper.vm.getNimModels("parakeet-ctc-1.1b-asr");
    expect(filteredModels).toEqual([
      { label: "nvidia/parakeet-ctc-1.1b-asr", value: "nvidia/parakeet-ctc-1.1b-asr", subLabel: "" },
      { label: "meta/llama-3.1-8b-base", value: "meta/llama-3.1-8b-base", subLabel: "" },
    ]);
    expect(inferenceService.getNimCatalogModels).toHaveBeenCalledWith("parakeet-ctc-1.1b-asr", 0, 500);
  });

  it("filters models from staticModelsList if available", async () => {
    vi.mocked(inferenceService.getNimCatalogModels).mockResolvedValue({
      ...modelListResponse,
      isStaticData: true,
    });
    const wrapper = createWrapper();
    await flushPromises();
    const filtered = await wrapper.vm.getNimModels("llama");
    expect(filtered).toEqual([{ label: "meta/llama-3.1-8b-base", value: "meta/llama-3.1-8b-base", subLabel: "" }]);
    expect(wrapper.vm.apiModelsList).toBeNull();
  });

  it("emits 'nim-model-name-changed' when model name is updated and not exists in list", async () => {
    const wrapper = createWrapper();
    await flushPromises();
    wrapper.vm.updateModelName("google/t5-large");
    const emitResult: TTestEmit<string[]> = wrapper.emitted("nim-model-name-changed") || [];
    expect(emitResult).toHaveLength(1);
    expect(emitResult[0]).toEqual(["google/t5-large"]);
  });

  it("emits 'is-valid' when modelName changes", async () => {
    const wrapper = createWrapper();
    await flushPromises();
    let emitResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
    expect(emitResult).toHaveLength(1);
    expect(emitResult[0]).toEqual([false]);
    await wrapper.setProps({ modelName: "google/t5-large" });
    emitResult = wrapper.emitted("is-valid") || [];
    expect(emitResult).toHaveLength(2);
    expect(emitResult[1]).toEqual([true]);
  });

  it("emits 'is-valid' with false for an invalid model name", async () => {
    const wrapper = createWrapper({ modelName: "invalid_model" });
    await flushPromises();
    await wrapper.setProps({ modelName: "another_invalid&model" });
    const emitResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
    expect(emitResult).toHaveLength(2);
    expect(emitResult[1]).toEqual([false]);
  });

  it("filters models from apiModelsList", async () => {
    const term = "llama";
    const filterResources = modelListResponse.resources.filter((model) => model.resourceId.includes(term));
    vi.mocked(inferenceService.getNimCatalogModels)
      .mockResolvedValueOnce(modelListResponse)
      .mockResolvedValue({
        ...modelListResponse,
        resources: filterResources,
      });
    const wrapper = createWrapper();
    await flushPromises();
    const filtered = await wrapper.vm.getNimModels(term);
    expect(wrapper.vm.apiModelsList).toEqual([
      { label: "meta/llama-3.1-8b-base", value: "meta/llama-3.1-8b-base", subLabel: "" },
    ]);
    expect(filtered).toEqual([{ label: "meta/llama-3.1-8b-base", value: "meta/llama-3.1-8b-base", subLabel: "" }]);
    expect(wrapper.vm.staticModelsList).toBeNull();
  });

  it("handles API error gracefully and resets lists", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(inferenceService.getNimCatalogModels)
      .mockResolvedValueOnce(modelListResponse)
      .mockRejectedValue(new Error("API error"));
    const wrapper = createWrapper();
    await flushPromises();
    expect(wrapper.vm.staticModelsList).toBeNull();
    expect(wrapper.vm.apiModelsList).not.toBeNull();
    await wrapper.vm.listModelsFromApi("test");
    expect(wrapper.vm.staticModelsList).toBeNull();
    expect(wrapper.vm.apiModelsList).toBeNull();
    expect(wrapper.vm.$q.notify).toHaveBeenCalled();
    errorSpy.mockRestore();
  });
});
