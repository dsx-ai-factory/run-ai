/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum EDockerRegistrySelectorEvents {
  UpdateModelValue = "update:model-value",
}

export interface IDockerRegistrySelectorModel {
  dockerRegistryName: string;
}
