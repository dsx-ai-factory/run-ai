/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { INimAccessModelModel } from "../nim-access-model/nim-access-model.model";

export interface INimModelAndAccessModel {
  modelName: string;
  profileNameOrHash: string;
  accessModel: INimAccessModelModel;
}
