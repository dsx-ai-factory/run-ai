/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Training } from "@/swagger-models/assets-service-client";
import type { IExtendJobValues } from "./workspace.model";
import type { IToolItem } from "./workload.model";

export interface ITrainingList extends Training {
  connections?: Array<IToolItem>;
  extendJobValues?: IExtendJobValues;
}
