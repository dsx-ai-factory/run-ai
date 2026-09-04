/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DistributedFramework } from "@/swagger-models/assets-service-client";

export const toolTypeIconsMap: Record<string, string> = {
  "jupyter-notebook": "jupyter",
  "visual-studio-code": "vscode2",
  tensorboard: "tensorboard",
  rstudio: "rstudio",
  mlflow: "mlflow",
  wandb: "wandb",
  custom: "custom-tool",
  comet: "comet",
  matlab: "matlab",
  "chatbot-ui": "chatbotui",
  "serving-endpoint": "serving-endpoint",
};

export const NVIDIA_WORKLOAD_ICON = "nvidia";
export const GENERIC_WORKLOAD_ICON = "custom-tool";

export const jobKindIconsMap: Record<string, string> = {
  MPIJob: "mpi",
  MPI: "mpi",
  TFJob: "tensorboard",
  TF: "tensorboard",
  PyTorchJob: "pytorch",
  PyTorch: "pytorch",
  XGBoostJob: "xgboost",
  XGBoost: "xgboost",
  RayCluster: "ray",
  RayJob: "ray",
  RayService: "ray",
  Spark: "spark",
  JAXJob: "jax",
  Unknown: "custom-tool",
  Training: GENERIC_WORKLOAD_ICON,
  Distributed: GENERIC_WORKLOAD_ICON,
  Interactive: GENERIC_WORKLOAD_ICON,
  Workspace: GENERIC_WORKLOAD_ICON,
  Inference: GENERIC_WORKLOAD_ICON,
  HuggingFace: "hugging-face",
  NvidiaNim: "nvidianim",
  Nvidia: NVIDIA_WORKLOAD_ICON,
  VLLM: "vllm",
  TGI: "tgi",
};

export const allWorkloadsIconsMap: Record<string, string> = {
  ...jobKindIconsMap,
  ...toolTypeIconsMap,
};

export enum DATA_SOURCE_ICONS_NAMES {
  s3 = "s3",
  git = "git",
  nfs = "nfs",
  hostPath = "hostpath",
  pvc = "pvc",
  default = "data-source-gray",
  "config-map" = "config-map",
  "secret-volume" = "secret-volume",
  "data-volume" = "data-volume",
  "pvc-volume" = "pvc-volume",
  "empty-dir" = "empty-dir",
}

export const distFrameworkIconsMap: Record<DistributedFramework, string> = {
  [DistributedFramework.Tf]: "tensorboard",
  [DistributedFramework.PyTorch]: "pytorch",
  [DistributedFramework.Mpi]: "mpi",
  [DistributedFramework.XgBoost]: "xgboost",
  [DistributedFramework.Jax]: "jax",
};
