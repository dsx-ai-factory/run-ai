/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IPaginationFilter } from "@/models/filter.model";
import { UsersFilterSortFields } from "@/swagger-models/identity-manager-client";
import { stringUtil } from "@/utils/string.util";

export const userUtils = {
  getFilterByArrays,
  getFilterByWithEqualsDates,
  getUserDisplayLabel,
};

export interface IUserDisplayLabel {
  label: string;
  primaryLabel: string;
  secondaryLabel?: string;
}

function getUserDisplayLabel(user: {
  username: string;
  firstName?: string | null;
  lastName?: string | null;
}): IUserDisplayLabel {
  const fullName = [user.firstName, user.lastName].filter(Boolean).join(" ");
  if (fullName) {
    return { label: `${user.username} ${fullName}`, primaryLabel: user.username, secondaryLabel: fullName };
  }
  return { label: user.username, primaryLabel: user.username };
}

// Convert range date filters to equality filters
function getFilterByWithEqualsDates(filters: IPaginationFilter): string[] {
  let filtersArray = [...(filters.filterBy || [])];
  const datesFiled = [
    UsersFilterSortFields.LastUpdated,
    UsersFilterSortFields.CreationTime,
    UsersFilterSortFields.LastLogin,
  ];

  const equalsDatesFields = [];
  for (const dateFilter of datesFiled) {
    const dateFilterInstances: string[] = filters.filterBy?.filter((f) => f.startsWith(dateFilter)) || [];
    if (dateFilterInstances.length > 1) {
      // Get the date from the filter
      const date = dateFilterInstances[1].split(">=")[1] || dateFilterInstances[1].split("<=")[1];
      // Add the equality filter
      equalsDatesFields.push(`${dateFilter}==${date}`);
      // Remove the range filter
      filtersArray = filtersArray.filter((f) => !f.startsWith(dateFilter));
      // Combine the filters
      filtersArray = [...(filtersArray || []), ...equalsDatesFields.filter((f) => f.startsWith(dateFilter))];
    }
  }
  return filtersArray;
}

// Get the filter by array with the isLocal filter (converts the type filter to isLocal filter)
function getFilterByArrays(filters: IPaginationFilter, isLocalUsers: boolean): string[] {
  const typeFilter = filters.filterBy?.find((f) => f.startsWith("type!="));
  filters.filterBy = getFilterByWithEqualsDates(filters);
  const filtersWithoutType: string[] = filters.filterBy?.filter((f) => f.indexOf("type!=") < 0) || [];
  if (typeFilter) {
    const isLocalValue = !stringUtil.toBoolean(typeFilter.split("!=")[1]);
    return [...filtersWithoutType, `isLocal==${isLocalValue}`];
  } else if (isLocalUsers) {
    return [...filtersWithoutType, `isLocal==true`];
  }
  return [...filtersWithoutType];
}
