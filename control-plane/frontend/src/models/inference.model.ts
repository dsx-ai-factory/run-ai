/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IToolItem } from "@/models/workload.model";
import type { Inference } from "@/swagger-models/assets-service-client";

export enum EInferenceType {
  RunaiModel = "runai-model",
  HuggingFace = "hugging-face",
  AssetsBased = "assets-based",
  Nim = "nim",
}

export enum EInferenceArchitecture {
  Standard = "Standard",
  Distributed = "Distributed",
}

export interface IInferenceArchitectureClusterFeatures {
  isKnativeInstalled: boolean;
  isLwsInstalled: boolean;
}

export enum EHuggingFaceEnvVariableNames {
  RUNAI_MODEL = "RUNAI_MODEL",
  HF_TOKEN = "HF_TOKEN",
  RUNAI_MODEL_NAME = "RUNAI_MODEL_NAME",
  HF_HOME = "HF_HOME",
  HOME = "HOME",
}

export const HF_HOME_DEFAULT_PATH = "/tmp/huggingface";
export const HOME_DEFAULT_PATH = "/tmp/home";

export enum ENimEnvVariableNames {
  NimModelName = "NIM_MODEL_NAME",
  NimModelProfile = "NIM_MODEL_PROFILE",
  NgcApiKey = "NGC_API_KEY",
  NimCachePath = "NIM_CACHE_PATH",
  OutlinesCacheDir = "OUTLINES_CACHE_DIR",
  NimServerPort = "NIM_SERVER_PORT",
  NimHttpApiPort = "NIM_HTTP_API_PORT",
  NimJsonlLogging = "NIM_JSONL_LOGGING",
  NimLogLevel = "NIM_LOG_LEVEL",
}

export const SERVING_PORT_SERVER_PORT = 8000;

export const LLM_IMAGE_FOR_HUGGING_FACE = "vllm/vllm-openai:latest";
export const LLM_IMAGE_FOR_TGI = "ghcr.io/huggingface/text-generation-inference:latest";
export const NIM_IMAGE_PREFIX = "nvcr.io/nim";
export const NIM_WORKLOAD_LABEL = "run.ai/nim-workload";
export const VLLM_WORKLOAD_LABEL = "run.ai/vllm-workload";
export const NIM_IMAGE_FOR_HF_MODELS = "nvcr.io/nim/nvidia/llm-nim:1.12";

export const ARGS_PREFIX_FOR_TGI = "--model-id";
export const ARGS_PREFIX_FOR_VLLM = "--model";

export enum EServerType {
  VLLM = "vLLM",
  TGI = "TGI",
  Nim = "NVIDIA NIM",
  Custom = "Custom",
}

export type TInferenceServerSelection = EServerType.Custom | EServerType.Nim | EServerType.VLLM | EServerType.TGI;

// old values
export const OLD_LLM_ENVIRONMENT_FOR_HUGGING_FACE = "llm-server";
export const OLD_LLM_IMAGE_FOR_HUGGING_FACE = "runai-vllm";

export interface IInferenceList extends Inference {
  connections?: Array<IToolItem>;
  extendJobValues?: IExtendJobValues;
}
export interface IExtendJobValues {
  usedMemory?: string;
  usedCPUs?: string;
  gpusUtilization?: string;
  usedGpuMemory?: string;
  currentAllocatedGPUsMemoryBytes?: number;
}

export const MIN_REPLICAS_DEFAULT_VALUE = 1;
export const MAX_REPLICAS_DEFAULT_VALUE = 1;
