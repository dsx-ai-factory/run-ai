/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ISelectOption } from "@/models/global.model";

export enum ENimImagePullCredentialType {
  NgcApiKey = "ngcApiKey",
  DockerRegistry = "dockerRegistry",
}

export enum ENimImagePullCredentialSelectorEvents {
  UpdateModelValue = "update:model-value",
}

export interface INimImagePullCredentialSelectorModel {
  credentialType: ENimImagePullCredentialType;
  credentialName: string;
}

export const NIM_IMAGE_PULL_CREDENTIAL_TYPE_OPTIONS: ISelectOption[] = [
  { value: ENimImagePullCredentialType.NgcApiKey, label: "NGC API key" },
  { value: ENimImagePullCredentialType.DockerRegistry, label: "Docker registry" },
];
