/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IService } from "@/models/service.model";
import type { ITableColumn } from "@/models/table.model";

export const allServiceColumnsMap: Record<string, ITableColumn> = {
  name: {
    name: "name",
    label: "Service",
    field: (row: IService) => row.name,
    sortable: true,
    align: "left",
  },
};

export const serviceMiniTableColumns: Array<ITableColumn> = [{ ...allServiceColumnsMap.name, display: true }];
