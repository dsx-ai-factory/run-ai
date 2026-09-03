/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ECustomCell, type ITableColumn } from "@/models/table.model";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";

const LegacyAccessRuleTransitionColumns: Record<string, ITableColumn> = {
  accessRuleCount: {
    name: "accessRuleCount",
    label: "Access rule(s)",
    field: (row: ILegacyRoleTransitionRow) => row.accessRuleCount,
    align: "left",
    style: "width: 125px",
  },
  legacyRole: {
    name: "legacyRole",
    label: "Legacy role",
    field: (row: ILegacyRoleTransitionRow) => row.legacyRole,
    align: "left",
    style: "width: 263px; max-width: 263px",
    customCell: ECustomCell.LEGACY_ROLE_COL,
  },
  newRole: {
    name: "newRole",
    label: "New role",
    field: (row: ILegacyRoleTransitionRow) => row.newRole,
    align: "left",
    style: "width: 280px",
    customCell: ECustomCell.NEW_ROLE_COL,
  },
};

export const legacyAccessRuleTransitionColumns: Array<ITableColumn> = [
  LegacyAccessRuleTransitionColumns.accessRuleCount,
  LegacyAccessRuleTransitionColumns.legacyRole,
  LegacyAccessRuleTransitionColumns.newRole,
];
