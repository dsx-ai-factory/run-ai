/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ITableColumn } from "@/models/table.model";

export interface IService {
  name: string;
}

export interface IServiceTableModalOptions {
  services: IService[];
  missingPrerequisites: string[];
  columns: Array<ITableColumn>;
  header: string;
  loading: boolean;
}
