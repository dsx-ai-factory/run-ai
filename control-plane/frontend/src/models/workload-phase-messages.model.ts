/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ITableColumn } from "@/models/table.model";

export interface Field {
  kind?: string | null;
  name?: string | null;
  description?: string | null;
  parentName?: string | null;
}

export interface Fields {
  network?: Field[];
  storage?: Field[];
  compute?: Field[];
  internal?: Field[];
}

export const allWorkloadPhaseMessageColumnsMap: ITableColumn[] = [
  {
    name: "parentName",
    label: "Resource",
    align: "left",
    field: "parentName",
  },
  {
    name: "kind",
    label: "Resource type",
    align: "left",
    field: "kind",
  },
  {
    name: "name",
    label: "Name",
    align: "left",
    field: "name",
  },
  {
    name: "description",
    label: "Description",
    align: "left",
    field: "description",
  },
];
