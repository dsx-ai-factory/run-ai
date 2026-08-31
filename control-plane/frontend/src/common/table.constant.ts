/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export type TTableDefaultProps = Record<string, string | boolean>;

export const tableDefaultProps: TTableDefaultProps = {
  "binary-state-sort": true,
  "hide-selected-banner": true,
  "table-class": "runai-table",
  "rows-per-page-label": "Rows per page",
};
