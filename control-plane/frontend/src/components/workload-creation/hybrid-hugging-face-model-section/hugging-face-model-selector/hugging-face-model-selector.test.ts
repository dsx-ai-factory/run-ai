/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, expect, vi, it, beforeEach } from "vitest";
import HuggingFaceModelSelector from "./hugging-face-model-selector.vue";
import { Quasar } from "quasar";

// services
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { catalogService } from "@/services/control-plane/catalog.service/catalog.service";

import { errorMessages } from "@/common/error-message.constant";
import { BadRequestError } from "@/models/http-response.model";
import { BACKEND_MESSAGES, EHuggingFaceModelEvents } from "./hugging-face-model-selector.model";

vi.mock("@/services/control-plane/inference.service/inference.service");
vi.mock("@/services/control-plane/catalog.service/catalog.service");

describe("hugging-face-model-selector.vue", () => {
  const createWrapper = function (modelName = "") {
    return mount(HuggingFaceModelSelector, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: modelName,
      },
    });
  };

  describe("General", () => {
    beforeEach(() => {
      vi.resetAllMocks();
    });

    it("renders the component with initial props", () => {
      const wrapper = createWrapper();
      expect(wrapper.find(".hf-model-selector").exists()).toBe(true);
      expect(wrapper.text()).toContain("Select or type a language model name as displayed in Hugging Face");
    });

    it("initializes selectedModel with the provided modelValue prop", () => {
      const wrapper = createWrapper("some/model");
      expect(wrapper.vm.selectedModel).toBe("some/model");
    });
  });

  describe("Methods", () => {
    beforeEach(() => {
      vi.resetAllMocks();
    });

    describe("getHfModels", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("filters models in the autocomplete select", async () => {
        const modelId = "google/gemma-2-2b-it";
        vi.mocked(catalogService.searchHuggingFaceCatalog).mockResolvedValue([
          {
            id: modelId,
            gated: false,
            downloads: 89400,
          },
        ]);

        const wrapper = createWrapper();
        const filteredModels = await wrapper.vm.getHfModels(modelId);

        expect(catalogService.searchHuggingFaceCatalog).toHaveBeenCalledWith(modelId);
        expect(filteredModels).toEqual([{ label: modelId, value: modelId }]);
        expect(wrapper.vm.modelsMap[modelId]).toEqual({
          id: modelId,
          gated: false,
          downloads: 89400,
          formattedDownloads: "89.4k",
        });
      });

      it("handles gated models correctly", async () => {
        const modelId = "meta/llama-3.1-8b";
        vi.mocked(catalogService.searchHuggingFaceCatalog).mockResolvedValue([
          {
            id: modelId,
            gated: true,
            downloads: 1835100,
          },
        ]);

        const wrapper = createWrapper();
        const filteredModels = await wrapper.vm.getHfModels(modelId);

        expect(filteredModels).toEqual([{ label: modelId, value: modelId }]);
        expect(wrapper.vm.modelsMap[modelId].gated).toBe(true);
        expect(wrapper.vm.modelsMap[modelId].formattedDownloads).toBe("1.8m");
      });

      it("handles errors and returns empty array", async () => {
        vi.mocked(catalogService.searchHuggingFaceCatalog).mockRejectedValue(new Error("Network error"));

        const wrapper = createWrapper();
        wrapper.vm.$q.notify = vi.fn();

        const filteredModels = await wrapper.vm.getHfModels("some-model");

        expect(filteredModels).toEqual([]);
        expect(wrapper.vm.modelsMap).toEqual({});
        expect(wrapper.vm.$q.notify).toHaveBeenCalled();
      });
    });

    describe("onModelNameChanged", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("updates selectedModel and emits the correct event when the model name changes", async () => {
        const model = "google/gemma-2-2b-it";
        const wrapper = createWrapper();
        wrapper.vm.onModelNameChanged(model);
        expect(wrapper.vm.selectedModel).toBe(model);
        expect(wrapper.emitted()[EHuggingFaceModelEvents.UpdateModelValue][0]).toEqual([{ modelName: model }]);
      });

      it("updates selectedModel when it is null and emits the correct event when the model name changes", async () => {
        const model = null;
        const wrapper = createWrapper();
        wrapper.vm.onModelNameChanged(model);
        expect(wrapper.vm.selectedModel).toBe("");
        expect(wrapper.emitted()[EHuggingFaceModelEvents.UpdateModelValue][0]).toEqual([{ modelName: "" }]);
      });

      it("updates selectedModel when an ISelectOption object is passed (keyboard selection)", async () => {
        const modelOption = { label: "google/gemma-2-2b-it", value: "google/gemma-2-2b-it" };
        const wrapper = createWrapper();
        wrapper.vm.onModelNameChanged(modelOption);
        expect(wrapper.vm.selectedModel).toBe(modelOption.value);
        expect(wrapper.emitted()[EHuggingFaceModelEvents.UpdateModelValue][0]).toEqual([
          { modelName: modelOption.value },
        ]);
      });

      it("clears warningMessage when model name changes", async () => {
        const wrapper = createWrapper();
        wrapper.vm.warningMessage = "Some warning";
        wrapper.vm.onModelNameChanged("new-model");
        expect(wrapper.vm.warningMessage).toBe("");
      });
    });

    describe("validateModel", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("should return true and invoke validationChanged with true when no error message exists", async () => {
        const modelName = "valid-model";

        vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
          modelName: modelName,
          language: true,
        });

        const wrapper = createWrapper();
        const result = await wrapper.vm.validateModel(modelName);

        expect(result).toBe(true);
        expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
        expect(wrapper.emitted()[EHuggingFaceModelEvents.IsValid][0]).toEqual([{ valid: true, gated: false }]);
      });

      it("should return the error message and invoke validationChanged with false when there is an error message", async () => {
        const modelName = "valid-model";

        vi.mocked(inferenceService.getHuggingFaceModelInfo).mockRejectedValue(
          new BadRequestError(BACKEND_MESSAGES.MODEL_WAS_NOT_FOUND_ON_HUGGING_FACE),
        );

        const wrapper = createWrapper();
        const result = await wrapper.vm.validateModel(modelName);
        expect(result).toBe(errorMessages.HUGGING_FACE_MODEL_NOT_FOUND);
        expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
        expect(wrapper.emitted()[EHuggingFaceModelEvents.IsValid][0]).toEqual([{ valid: false, gated: false }]);
      });

      it("should return true and invoke validationChanged with true for valid and gated", async () => {
        const modelName = "valid-model";
        vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
          modelName: modelName,
          language: true,
          gated: true,
        });

        const wrapper = createWrapper();
        const result = await wrapper.vm.validateModel(modelName);

        expect(result).toBe(true);
        expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
        expect(wrapper.emitted()[EHuggingFaceModelEvents.IsValid][0]).toEqual([{ valid: true, gated: true }]);
      });

      it("should display warning message in DOM when model is not a language model", async () => {
        const modelName = "non-language-model";
        vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
          modelName: modelName,
          language: false,
        });

        const wrapper = createWrapper();
        await wrapper.vm.validateModel(modelName);

        expect(wrapper.vm.warningMessage).toBe(errorMessages.MODEL_IS_NOT_A_VLLM);
        expect(wrapper.findComponent({ name: "NotificationAlertBar" }).exists()).toBe(true);
        expect(wrapper.html()).toContain("This isn't identified as a language model.");
        expect(wrapper.html()).toContain("Non-language models won't run.");
      });
    });

    describe("verifyModelName", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("validates empty model name", async () => {
        vi.mocked(inferenceService.getHuggingFaceModelInfo);
        const wrapper = createWrapper();

        const result = await wrapper.vm.verifyModelName("");
        expect(result).toEqual({
          modelName: "",
          errorMessage: errorMessages.SELECT_OR_TYPE_A_NAME,
        });
        expect(inferenceService.getHuggingFaceModelInfo).not.toHaveBeenCalled();
      });

      it("return error when model name has incorrect format", async () => {
        vi.mocked(inferenceService.getHuggingFaceModelInfo);
        const wrapper = createWrapper();
        const modelName = "invalid@model";
        const result = await wrapper.vm.verifyModelName(modelName);
        expect(result).toEqual({
          modelName,
          errorMessage: errorMessages.ENTER_A_VALID_NAME,
        });
        expect(inferenceService.getHuggingFaceModelInfo).not.toHaveBeenCalled();
      });

      it("return warning message when the model is not a vllm", async () => {
        const modelName = "meta/llama-3.1-8b-base";
        vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
          modelName: modelName,
          language: false,
        });

        const wrapper = createWrapper();
        const result = await wrapper.vm.verifyModelName(modelName);
        expect(result).toEqual({
          modelName,
          warningMessage: errorMessages.MODEL_IS_NOT_A_VLLM,
        });
        expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
      });

      it("return error when the model is not found in hugging face", async () => {
        const modelName = "meta/llama-3.1-8b-base";
        vi.mocked(inferenceService.getHuggingFaceModelInfo).mockRejectedValue(
          new BadRequestError(BACKEND_MESSAGES.MODEL_WAS_NOT_FOUND_ON_HUGGING_FACE),
        );

        const wrapper = createWrapper();
        const result = await wrapper.vm.verifyModelName(modelName);
        expect(result).toEqual({
          modelName,
          errorMessage: errorMessages.HUGGING_FACE_MODEL_NOT_FOUND,
        });
        expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
      });

      it("return error when getHuggingFaceModelInfo fails", async () => {
        const modelName = "meta/llama-3.1-8b-base";
        vi.mocked(inferenceService.getHuggingFaceModelInfo).mockRejectedValue(new Error("network error"));

        const wrapper = createWrapper();
        const result = await wrapper.vm.verifyModelName(modelName);
        expect(result).toEqual({
          modelName,
          errorMessage: errorMessages.FAILED_TO_GET_MODEL_INFO,
        });
        expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
      });
    });

    describe("buildModelsMap", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("builds modelsMap with formatted downloads", () => {
        const wrapper = createWrapper();
        const models = [
          { id: "model-1", gated: false, downloads: 89400 },
          { id: "model-2", gated: true, downloads: 1835100 },
        ];

        wrapper.vm.buildModelsMap(models);

        expect(wrapper.vm.modelsMap["model-1"]).toEqual({
          id: "model-1",
          gated: false,
          downloads: 89400,
          formattedDownloads: "89.4k",
        });
        expect(wrapper.vm.modelsMap["model-2"]).toEqual({
          id: "model-2",
          gated: true,
          downloads: 1835100,
          formattedDownloads: "1.8m",
        });
      });

      it("ignores models without an ID", () => {
        const wrapper = createWrapper();
        const models = [
          { id: "model-1", gated: false, downloads: 89400 },
          { gated: true, downloads: 1000 }, // no id
        ];

        wrapper.vm.buildModelsMap(models);

        expect(Object.keys(wrapper.vm.modelsMap)).toHaveLength(1);
        expect(wrapper.vm.modelsMap["model-1"]).toBeDefined();
      });

      it("handles models with no downloads", () => {
        const wrapper = createWrapper();
        const models = [{ id: "model-1", gated: false }];

        wrapper.vm.buildModelsMap(models);

        expect(wrapper.vm.modelsMap["model-1"]).toEqual({
          id: "model-1",
          gated: false,
          formattedDownloads: "",
        });
      });

      it("handles empty array", () => {
        const wrapper = createWrapper();
        wrapper.vm.buildModelsMap([]);
        expect(wrapper.vm.modelsMap).toEqual({});
      });
    });

    describe("createSelectOptions", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("creates select options from models", () => {
        const wrapper = createWrapper();
        const models = [
          { id: "model-1", gated: false, downloads: 89400 },
          { id: "model-2", gated: true, downloads: 1835100 },
        ];

        const options = wrapper.vm.createSelectOptions(models);

        expect(options).toEqual([
          { label: "model-1", value: "model-1" },
          { label: "model-2", value: "model-2" },
        ]);
      });

      it("handles models with empty id", () => {
        const wrapper = createWrapper();
        const models = [{ id: "", gated: false }];

        const options = wrapper.vm.createSelectOptions(models);

        expect(options).toEqual([{ label: "", value: "" }]);
      });

      it("handles empty array", () => {
        const wrapper = createWrapper();
        const options = wrapper.vm.createSelectOptions([]);
        expect(options).toEqual([]);
      });
    });

    describe("handleFetchError", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("clears modelsMap on error", () => {
        const wrapper = createWrapper();
        wrapper.vm.modelsMap = { "model-1": { id: "model-1", gated: false } };
        wrapper.vm.$q.notify = vi.fn();

        wrapper.vm.handleFetchError(new Error("Network error"));

        expect(wrapper.vm.modelsMap).toEqual({});
      });

      it("calls $q.notify with error notification", () => {
        const wrapper = createWrapper();
        wrapper.vm.$q.notify = vi.fn();

        wrapper.vm.handleFetchError(new Error("Network error"));

        expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
          expect.objectContaining({
            message: expect.any(String),
          }),
        );
      });
    });

    describe("getHfModels - Edge Cases", () => {
      beforeEach(() => {
        vi.resetAllMocks();
      });

      it("handles empty search results", async () => {
        vi.mocked(catalogService.searchHuggingFaceCatalog).mockResolvedValue([]);

        const wrapper = createWrapper();
        const models = await wrapper.vm.getHfModels("nonexistent");

        expect(models).toEqual([]);
        expect(wrapper.vm.modelsMap).toEqual({});
      });

      it("handles multiple models correctly", async () => {
        const mockModels = [
          { id: "openai/gpt-oss-20b", gated: true, downloads: 89400 },
          { id: "openai/gpt-oss-120b", gated: false, downloads: 1835100 },
          { id: "rednote-hilab/dots.ocr", gated: false, downloads: 7745400 },
        ];

        vi.mocked(catalogService.searchHuggingFaceCatalog).mockResolvedValue(mockModels);

        const wrapper = createWrapper();
        const models = await wrapper.vm.getHfModels("openai");

        expect(models).toHaveLength(3);
        expect(wrapper.vm.modelsMap["openai/gpt-oss-20b"].formattedDownloads).toBe("89.4k");
        expect(wrapper.vm.modelsMap["openai/gpt-oss-120b"].formattedDownloads).toBe("1.8m");
        expect(wrapper.vm.modelsMap["rednote-hilab/dots.ocr"].formattedDownloads).toBe("7.7m");
      });

      it("updates modelsMap when searching multiple times", async () => {
        // First search
        vi.mocked(catalogService.searchHuggingFaceCatalog).mockResolvedValueOnce([
          { id: "model-1", gated: false, downloads: 1000 },
        ]);

        const wrapper = createWrapper();
        await wrapper.vm.getHfModels("model-1");
        expect(Object.keys(wrapper.vm.modelsMap)).toHaveLength(1);

        // Second search
        vi.mocked(catalogService.searchHuggingFaceCatalog).mockResolvedValueOnce([
          { id: "model-2", gated: true, downloads: 2000 },
          { id: "model-3", gated: false, downloads: 3000 },
        ]);

        await wrapper.vm.getHfModels("model-2");
        expect(Object.keys(wrapper.vm.modelsMap)).toHaveLength(2);
        expect(wrapper.vm.modelsMap["model-1"]).toBeUndefined();
        expect(wrapper.vm.modelsMap["model-2"]).toBeDefined();
        expect(wrapper.vm.modelsMap["model-3"]).toBeDefined();
      });
    });
  });

  describe("Emits", () => {
    beforeEach(() => {
      vi.resetAllMocks();
    });

    it("emits hf-model-name-changed when model name changes", async () => {
      const newModelName = "meta/llama-3.1-8b-base";

      const wrapper = createWrapper();

      vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
        modelName: newModelName,
        authorized: true,
        language: true,
      });

      await wrapper.vm.onModelNameChanged(newModelName);

      expect(wrapper.emitted(EHuggingFaceModelEvents.UpdateModelValue)).toBeTruthy();
      expect(wrapper.emitted()[EHuggingFaceModelEvents.UpdateModelValue][0]).toEqual([{ modelName: newModelName }]);
      expect(inferenceService.getHuggingFaceModelInfo).not.toHaveBeenCalled();
    });

    it("emits 'is-valid' as true for valid model name that is not gated", async () => {
      const validModelName = "meta/llama-3.1-8b-base";
      vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
        modelName: validModelName,
        language: true,
      });
      const wrapper = createWrapper();
      await wrapper.vm.validateModel(validModelName);
      expect(wrapper.emitted(EHuggingFaceModelEvents.IsValid)).toBeTruthy();
      expect(wrapper.emitted()[EHuggingFaceModelEvents.IsValid][0]).toEqual([{ valid: true, gated: false }]);
      expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
    });

    it("emits 'is-valid' as true for a gated model", async () => {
      const validModelName = "meta/llama-3.1-8b-base";
      vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
        modelName: validModelName,
        language: true,
        gated: true,
      });

      const wrapper = createWrapper();

      await wrapper.vm.validateModel(validModelName);

      expect(wrapper.emitted(EHuggingFaceModelEvents.IsValid)).toBeTruthy();
      expect(wrapper.emitted()[EHuggingFaceModelEvents.IsValid][0]).toEqual([{ valid: true, gated: true }]);
      expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
    });

    it("emits 'is-valid' as true when model is not a vllm", async () => {
      const validModelName = "meta/llama-3.1-8b-base";
      vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
        modelName: validModelName,
        language: false,
      });

      const wrapper = createWrapper();

      await wrapper.vm.validateModel(validModelName);

      expect(wrapper.emitted(EHuggingFaceModelEvents.IsValid)).toBeTruthy();
      expect(wrapper.emitted()[EHuggingFaceModelEvents.IsValid][0]).toEqual([{ valid: true, gated: false }]);
      expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
    });
  });
});
