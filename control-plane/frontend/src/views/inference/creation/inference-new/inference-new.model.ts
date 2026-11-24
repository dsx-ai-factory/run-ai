/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EInferenceArchitecture, EServerType } from "@/models/inference.model";

export interface IInferenceNewModel {
  name: string;
  clusterId: string;
  projectId: number;
  serverType: EServerType;
  architecture: EInferenceArchitecture;
  templateId: string | null;
}

export function getEmptyInferenceNewModel(): IInferenceNewModel {
  return {
    name: "",
    clusterId: "",
    projectId: -1,
    serverType: EServerType.Nim,
    architecture: EInferenceArchitecture.Standard,
    templateId: null,
  };
}
