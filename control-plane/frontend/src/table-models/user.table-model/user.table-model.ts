/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import type { IUser } from "@/models/user.model";
import { formatDate } from "@/utils/table-format.util";
import { UsersFilterSortFields } from "@/swagger-models/identity-manager-client";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import type { IFreeTextFilterModel } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { viewLinkColumn } from "@/table-models/common-columns";

export const allUsersColumnsMap: Record<string, ITableColumn> = {
  user: {
    name: UsersFilterSortFields.Username,
    label: "User",
    field: (row: IUser) => row.username,
    sortable: true,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.FreeText,
      excludeOperators: [EFilterOperator.DoesNotContain, EFilterOperator.Equals, EFilterOperator.NotEquals],
      autocompleteOptionsFn: autocompleteUtil.users(UsersFilterSortFields.Username, (u) => u.username),
    } as IFreeTextFilterModel,
  },
  fullName: {
    name: UsersFilterSortFields.FullName,
    label: "Full name",
    field: (row: IUser) => [row.firstName, row.lastName].filter(Boolean).join(" "),
    sortable: true,
    align: "left",
    display: true,
    format: (val: string) => val || "-",
    filter: {
      type: EColumnFilterType.FreeText,
      excludeOperators: [EFilterOperator.DoesNotContain, EFilterOperator.Equals, EFilterOperator.NotEquals],
      autocompleteOptionsFn: autocompleteUtil.users(
        UsersFilterSortFields.FullName,
        (u) => [u.firstName, u.lastName].filter(Boolean).join(" ") || undefined,
      ),
    } as IFreeTextFilterModel,
  },
  type: {
    name: "type",
    label: "Type",
    field: (row: IUser) => row.isLocal,
    align: "left",
    display: true,
    format: (isLocal) => (isLocal === null ? "-" : isLocal ? "Local" : "SSO"),
    sortable: true,
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: [
        {
          label: "Local",
          value: true,
        },
        {
          label: "SSO",
          value: false,
        },
      ],
    },
  },
  lastLogin: {
    name: UsersFilterSortFields.LastLogin,
    label: "Last login",
    field: (row: IUser) => row.lastLogin,
    sortable: true,
    align: "left",
    display: true,
    format: (lastLogin) => formatDate(lastLogin),
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  accessRules: viewLinkColumn({
    name: "accessRules",
    label: "Access rule(s)",
    emitName: "access-rules-clicked",
  }),
  createdBy: {
    name: UsersFilterSortFields.CreatedBy,
    label: "Created by",
    field: (row: IUser) => row.createdBy,
    sortable: true,
    align: "left",
    display: true,
    format: (createdBy) => createdBy || "-",
    // 230px from Figma.
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      excludeOperators: [EFilterOperator.DoesNotContain, EFilterOperator.Equals, EFilterOperator.NotEquals],
      autocompleteOptionsFn: autocompleteUtil.users(UsersFilterSortFields.CreatedBy, (u) => u.createdBy),
    } as IFreeTextFilterModel,
  },
  creationTime: {
    name: UsersFilterSortFields.CreationTime,
    label: "Creation time",
    field: (row: IUser) => row.createdAt,
    sortable: true,
    align: "left",
    display: true,
    format: (creationTime) => formatDate(creationTime),
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  lastUpdated: {
    name: UsersFilterSortFields.LastUpdated,
    label: "Last updated",
    field: (row: IUser) => row.updatedAt,
    sortable: true,
    align: "left",
    display: true,
    format: (updatedAt) => formatDate(updatedAt) || "-",
    filter: {
      type: EColumnFilterType.Date,
    },
  },
};

export const allUserColumns: Array<ITableColumn> = [
  allUsersColumnsMap.user,
  allUsersColumnsMap.fullName,
  allUsersColumnsMap.type,
  allUsersColumnsMap.lastLogin,
  allUsersColumnsMap.accessRules,
  allUsersColumnsMap.createdBy,
  allUsersColumnsMap.creationTime,
  allUsersColumnsMap.lastUpdated,
];

export const usersIndexColumns: Array<ITableColumn> = [
  { ...allUsersColumnsMap.user, display: true, mandatory: true },
  { ...allUsersColumnsMap.fullName, display: true },
  { ...allUsersColumnsMap.type, display: true },
  { ...allUsersColumnsMap.lastLogin, display: true },
  { ...allUsersColumnsMap.accessRules, display: true },
  { ...allUsersColumnsMap.createdBy, display: true },
  { ...allUsersColumnsMap.creationTime, display: true },
  { ...allUsersColumnsMap.lastUpdated, display: true },
];

export const usersDependentColumns = {
  accessRules: new Set([allUsersColumnsMap.accessRules.name]),
};
