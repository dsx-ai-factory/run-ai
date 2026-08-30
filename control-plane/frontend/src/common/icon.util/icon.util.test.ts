import { describe, it, expect } from "vitest";
import { getNimIconIfNimImage, getHuggingFaceIcon } from "./icon.util";
import {
  NIM_WORKLOAD_LABEL,
  NIM_IMAGE_PREFIX,
  LLM_IMAGE_FOR_HUGGING_FACE,
  LLM_IMAGE_FOR_TGI,
} from "@/models/inference.model";
import { allWorkloadsIconsMap } from "../icons.constant";
import { EWorkloadType } from "@/models/workload.model";
import type { Label } from "@/swagger-models/assets-service-client";

describe("getNimIconIfNimImage", () => {
  describe("when labels contain NIM workload label", () => {
    it("should return NvidiaNim icon when NIM workload label is present", () => {
      const labels: Label[] = [
        { name: NIM_WORKLOAD_LABEL, value: "true" },
        { name: "other-label", value: "other-value" },
      ];

      const result = getNimIconIfNimImage("some-image", labels);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return NvidiaNim icon when NIM workload label is present with different value", () => {
      const labels: Label[] = [{ name: NIM_WORKLOAD_LABEL, value: "false" }];

      const result = getNimIconIfNimImage("some-image", labels);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return NvidiaNim icon when NIM workload label is present with empty value", () => {
      const labels: Label[] = [{ name: NIM_WORKLOAD_LABEL, value: "" }];

      const result = getNimIconIfNimImage("some-image", labels);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return NvidiaNim icon when NIM workload label is present and image is null", () => {
      const labels: Label[] = [{ name: NIM_WORKLOAD_LABEL, value: "true" }];

      const result = getNimIconIfNimImage(null, labels);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return NvidiaNim icon when NIM workload label is present and image is undefined", () => {
      const labels: Label[] = [{ name: NIM_WORKLOAD_LABEL, value: "true" }];

      const result = getNimIconIfNimImage(undefined, labels);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return NvidiaNim icon when NIM workload label is present and image contains NIM prefix", () => {
      const labels: Label[] = [{ name: NIM_WORKLOAD_LABEL, value: "true" }];

      const result = getNimIconIfNimImage(`${NIM_IMAGE_PREFIX}/some-model:latest`, labels);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });
  });

  describe("when labels do not contain NIM workload label", () => {
    it("should return NvidiaNim icon when image contains NIM prefix", () => {
      const labels: Label[] = [{ name: "other-label", value: "other-value" }];

      const result = getNimIconIfNimImage(`${NIM_IMAGE_PREFIX}/some-model:latest`, labels);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return NvidiaNim icon when image contains NIM prefix with tag", () => {
      const result = getNimIconIfNimImage(`${NIM_IMAGE_PREFIX}/llama2:latest`);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return NvidiaNim icon when image contains NIM prefix with specific version", () => {
      const result = getNimIconIfNimImage(`${NIM_IMAGE_PREFIX}/gpt2:v1.0.0`);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return undefined when image does not contain NIM prefix", () => {
      const labels: Label[] = [{ name: "other-label", value: "other-value" }];

      const result = getNimIconIfNimImage("nvcr.io/nvidia/pytorch:latest", labels);

      expect(result).toBeUndefined();
    });

    it("should return undefined when image is null", () => {
      const labels: Label[] = [{ name: "other-label", value: "other-value" }];

      const result = getNimIconIfNimImage(null, labels);

      expect(result).toBeUndefined();
    });

    it("should return undefined when image is undefined", () => {
      const labels: Label[] = [{ name: "other-label", value: "other-value" }];

      const result = getNimIconIfNimImage(undefined, labels);

      expect(result).toBeUndefined();
    });

    it("should return undefined when image is empty string", () => {
      const labels: Label[] = [{ name: "other-label", value: "other-value" }];

      const result = getNimIconIfNimImage("", labels);

      expect(result).toBeUndefined();
    });
  });

  describe("when labels parameter is not provided", () => {
    it("should return NvidiaNim icon when image contains NIM prefix", () => {
      const result = getNimIconIfNimImage(`${NIM_IMAGE_PREFIX}/some-model:latest`);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return undefined when image does not contain NIM prefix", () => {
      const result = getNimIconIfNimImage("nvcr.io/nvidia/pytorch:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined when image is null", () => {
      const result = getNimIconIfNimImage(null);

      expect(result).toBeUndefined();
    });

    it("should return undefined when image is undefined", () => {
      const result = getNimIconIfNimImage(undefined);

      expect(result).toBeUndefined();
    });
  });

  describe("when labels parameter is null or undefined", () => {
    it("should return NvidiaNim icon when labels is null and image contains NIM prefix", () => {
      const result = getNimIconIfNimImage(`${NIM_IMAGE_PREFIX}/some-model:latest`, null);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return undefined when labels is null and image does not contain NIM prefix", () => {
      const result = getNimIconIfNimImage("nvcr.io/nvidia/pytorch:latest", null);

      expect(result).toBeUndefined();
    });

    it("should return NvidiaNim icon when labels is undefined and image contains NIM prefix", () => {
      const result = getNimIconIfNimImage(`${NIM_IMAGE_PREFIX}/some-model:latest`, undefined);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return undefined when labels is undefined and image does not contain NIM prefix", () => {
      const result = getNimIconIfNimImage("nvcr.io/nvidia/pytorch:latest", undefined);

      expect(result).toBeUndefined();
    });
  });

  describe("edge cases", () => {
    it("should return NvidiaNim icon when image contains NIM prefix as substring", () => {
      const result = getNimIconIfNimImage(`my-registry.com/${NIM_IMAGE_PREFIX}-custom:latest`);

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return undefined when image contains NIM prefix with different casing", () => {
      const result = getNimIconIfNimImage("NVCR.IO/NIM/some-model:latest");

      expect(result).toBeUndefined();
    });

    it("should return NvidiaNim icon when image contains NIM prefix as part of path", () => {
      const result = getNimIconIfNimImage("nvcr.io/nim-custom/some-model:latest");

      expect(result).toBe(allWorkloadsIconsMap.NvidiaNim);
    });

    it("should return undefined when image contains NIM as part of different registry", () => {
      const result = getNimIconIfNimImage("my-registry.com/nim-custom/some-model:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined when image contains NIM in different context", () => {
      const result = getNimIconIfNimImage("my-registry.com/some-nim-image:latest");

      expect(result).toBeUndefined();
    });
  });
});

describe("getHuggingFaceIcon", () => {
  describe("when workload type is not Inference", () => {
    it("should return undefined for Workspace workload type", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Workspace, "vllm/vllm-openai:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for Training workload type", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Training, "vllm/vllm-openai:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for Distributed workload type", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Distributed, "vllm/vllm-openai:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for DistributedInference workload type", () => {
      const result = getHuggingFaceIcon(EWorkloadType.DistributedInference, "vllm/vllm-openai:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for DistributedTraining workload type", () => {
      const result = getHuggingFaceIcon(EWorkloadType.DistributedTraining, "vllm/vllm-openai:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for TFJob workload type", () => {
      const result = getHuggingFaceIcon(EWorkloadType.TFJob, "vllm/vllm-openai:latest");

      expect(result).toBeUndefined();
    });
  });

  describe("when workload type is Inference but image is null or undefined", () => {
    it("should return undefined when image is null", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, null);

      expect(result).toBeUndefined();
    });

    it("should return undefined when image is undefined", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, undefined);

      expect(result).toBeUndefined();
    });

    it("should return undefined when image is empty string", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "");

      expect(result).toBeUndefined();
    });
  });

  describe("when workload type is Inference and image contains HuggingFace images", () => {
    it("should return VLLM icon for vllm image", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, LLM_IMAGE_FOR_HUGGING_FACE);

      expect(result).toBe(allWorkloadsIconsMap.VLLM);
    });

    it("should return TGI icon for TGI image", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, LLM_IMAGE_FOR_TGI);

      expect(result).toBe(allWorkloadsIconsMap.TGI);
    });

    it("should return undefined for vllm image with different tag", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "vllm/vllm-openai:v1.0.0");

      expect(result).toBeUndefined();
    });

    it("should return undefined for TGI image with different tag", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "ghcr.io/huggingface/text-generation-inference:v0.9.0");

      expect(result).toBeUndefined();
    });

    it("should return VLLM icon for vllm image with custom registry", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "my-registry.com/vllm/vllm-openai:latest");

      expect(result).toBe(allWorkloadsIconsMap.VLLM);
    });

    it("should return TGI icon for TGI image with custom registry", () => {
      const result = getHuggingFaceIcon(
        EWorkloadType.Inference,
        "my-registry.com/ghcr.io/huggingface/text-generation-inference:latest",
      );

      expect(result).toBe(allWorkloadsIconsMap.TGI);
    });
  });

  describe("when workload type is Inference but image does not contain HuggingFace images", () => {
    it("should return undefined for nvidia pytorch image", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "nvcr.io/nvidia/pytorch:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for nvidia base image", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "nvcr.io/nvidia/cuda:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for custom image", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "my-custom-image:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for image with similar name but different path", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "my-registry.com/vllm-custom:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined for image with different casing", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "VLLM/VLLM-OPENAI:latest");

      expect(result).toBeUndefined();
    });
  });

  describe("edge cases", () => {
    it("should return VLLM icon when image contains exact vllm path", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "my-registry.com/vllm/vllm-openai:latest");

      expect(result).toBe(allWorkloadsIconsMap.VLLM);
    });

    it("should return TGI icon when image contains exact huggingface path", () => {
      const result = getHuggingFaceIcon(
        EWorkloadType.Inference,
        "my-registry.com/ghcr.io/huggingface/text-generation-inference:latest",
      );

      expect(result).toBe(allWorkloadsIconsMap.TGI);
    });

    it("should return undefined when image contains vllm as substring but not exact path", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "my-registry.com/vllm-custom:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined when image contains huggingface as substring but not exact path", () => {
      const result = getHuggingFaceIcon(EWorkloadType.Inference, "my-registry.com/huggingface-custom:latest");

      expect(result).toBeUndefined();
    });

    it("should return undefined when image contains text-generation-inference as substring but not exact path", () => {
      const result = getHuggingFaceIcon(
        EWorkloadType.Inference,
        "my-registry.com/text-generation-inference-custom:latest",
      );

      expect(result).toBeUndefined();
    });
  });
});
