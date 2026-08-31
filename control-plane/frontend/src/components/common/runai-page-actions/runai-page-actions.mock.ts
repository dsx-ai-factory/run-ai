/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ITableColumn } from "@/models/table.model";

const filters = {
  sortBy: "environment",
  descending: false,
  page: 1,
  rowsPerPage: 20,
  columnFilters: [],
  searchTerm: "",
  displayedColumns: ["environment", "project", "image", "tools", "workspaces", "templates", "created-at", "created-by"],
  usageInfo: true,
};

const columns = [
  {
    name: "environment",
    label: "Environment",
    sortable: true,
    align: "left",
    display: true,
    customCell: "name-col",
  },
  {
    name: "project",
    label: "Project",
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: "image",
    label: "Image",
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: "tools",
    label: "Tool(s)",
    sortable: true,
    align: "left",
    display: true,
    customCell: "list-col",
    customCellEvent: { emitName: "tool-clicked" },
  },
  {
    name: "workspaces",
    label: "Workspace(s)",
    sortable: true,
    align: "left",
    display: true,
    customCell: "list-col",
    customCellEvent: { emitName: "workspace-clicked" },
  },
  {
    name: "templates",
    label: "Template(s)",
    sortable: true,
    align: "left",
    display: false,
    customCell: "list-col",
    customCellEvent: { emitName: "workload-template-clicked" },
  },
  {
    name: "created-by",
    label: "Created by",
    sortable: true,
    align: "left",
    display: false,
  },
  {
    name: "created-at",
    label: "Creation time",
    sortable: true,
    align: "left",
    display: true,
  },
] as ITableColumn[];

export const runaiPageActionsMocks = {
  filters,
  columns,
};
