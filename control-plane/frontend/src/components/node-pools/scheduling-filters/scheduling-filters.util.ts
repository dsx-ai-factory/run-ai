/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import {
  EColumnFilterType,
  EFilterOperator,
  type IEnumBasedFilterModel,
  type IFreeTextFilterModel,
  type IAdvancedFilterModel,
} from "@/models/filter.model";
export function matchesSchedulingFilters(row: { name: string; rank?: string }, filters: IAdvancedFilterModel[]) {
  return filters.every((filter) => {
    if (filter.type === EColumnFilterType.EnumBased)
      return (
        !(filter as IEnumBasedFilterModel).selectedValues.length ||
        (filter as IEnumBasedFilterModel).selectedValues.includes(row.rank ?? "")
      );
    if (filter.type !== EColumnFilterType.FreeText) return true;
    const name = row.name.toLocaleLowerCase(),
      term = (filter as IFreeTextFilterModel).term.toLocaleLowerCase();
    switch ((filter as IFreeTextFilterModel).value) {
      case EFilterOperator.Contains:
        return name.includes(term);
      case EFilterOperator.DoesNotContain:
        return !name.includes(term);
      case EFilterOperator.Equals:
        return name === term;
      case EFilterOperator.NotEquals:
        return name !== term;
      default:
        return true;
    }
  });
}
