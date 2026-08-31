/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  LLM_IMAGE_FOR_HUGGING_FACE,
  LLM_IMAGE_FOR_TGI,
  NIM_IMAGE_PREFIX,
  NIM_WORKLOAD_LABEL,
} from "@/models/inference.model";
import { allWorkloadsIconsMap } from "../icons.constant";
import { EWorkloadType } from "@/models/workload.model";
import type { Label } from "@/swagger-models/assets-service-client";

export function isNvidiaImage(image: string | null | undefined): boolean {
  if (!image) return false;
  return image.includes("nvcr.io/nvidia");
}

export function isNvidiaPytorch(image: string | null | undefined): boolean {
  if (!image) return false;
  return image.includes("nvcr.io/nvidia/pytorch");
}

export function getNvidiaImageBasedIcon(image: string | null | undefined): string | null {
  if (!image) return null;

  if (isNvidiaPytorch(image)) return "pytorch";

  if (isNvidiaImage(image)) return "nvidia";

  return null;
}

export function getNimIconIfNimImage(
  image: string | null | undefined,
  labels?: Label[] | null | undefined,
): string | undefined {
  if (labels?.some((label) => label.name === NIM_WORKLOAD_LABEL)) {
    return allWorkloadsIconsMap.NvidiaNim;
  }

  return image && image.includes(NIM_IMAGE_PREFIX) ? allWorkloadsIconsMap.NvidiaNim : undefined;
}

/**
 * @description - This function is used to get the icon of the workload.
 * Note that this logic supports the flat inference workloads/ templates only.
 * For the non-flat inference workloads use the workload.util.getWorkloadIcon.
 * @param type - The type of the workload
 * @param image - The image of the workload
 * @param environmentVariables - The environment variables of the workload
 * @returns Hugging Face icon name if it is fit the hugging face params, otherwise undefined
 */
export function getHuggingFaceIcon(type: EWorkloadType, image: string | null | undefined): string | undefined {
  if (type !== EWorkloadType.Inference || !image) return undefined;

  if (image.includes(LLM_IMAGE_FOR_HUGGING_FACE)) {
    return allWorkloadsIconsMap.VLLM;
  }

  if (image.includes(LLM_IMAGE_FOR_TGI)) {
    return allWorkloadsIconsMap.TGI;
  }

  return undefined;
}
