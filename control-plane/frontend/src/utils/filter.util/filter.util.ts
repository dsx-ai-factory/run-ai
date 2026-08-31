/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  EColumnFilterType,
  EDateFilterOperator,
  EFilterOperator,
  type IEnumBasedFilterSelectOption,
  type IAdvancedFilterModel,
} from "@/models/filter.model";
import { stringUtil } from "@/utils/string.util/string.util";

export const filterUtil = {
  getAdvancedFilterModelByType,
  getEqualsFilterString,
  getContainsFilterString,
  parseEnumToSelectOptions,
  getGreaterThanOrEqualFilterString,
  getNotEqualsFilterString,
  getFreeTextFilterQuery,
  getNumericFilterQuery,
};

function getAdvancedFilterModelByType(filterType: EColumnFilterType): Partial<IAdvancedFilterModel> {
  switch (filterType) {
    case EColumnFilterType.FreeText:
      return {
        term: "",
        value: EFilterOperator.Contains,
      };

    case EColumnFilterType.EnumBased:
      return {
        selectOptions: [],
        selectedValues: [],
        inputType: "checkbox",
      };

    case EColumnFilterType.Date:
      return {
        value: EDateFilterOperator.Before,
        date: "",
      };

    case EColumnFilterType.Numeric:
      return {
        minValue: 0,
        maxValue: 0,
        operator: EFilterOperator.GreaterThanOrEqual,
      };

    default:
      throw new Error(`Unknown filter type: ${filterType}`);
  }
}

function getEqualsFilterString(field: string, value: string | number): string {
  return `${field}${EFilterOperator.Equals}${value}`;
}

function getNotEqualsFilterString(field: string, value: string | number): string {
  return `${field}${EFilterOperator.NotEquals}${value}`;
}

function getContainsFilterString(field: string, value: string): string {
  return `${field}${EFilterOperator.Contains}${value}`;
}

function getGreaterThanOrEqualFilterString(field: string, value: number): string {
  return `${field}${EFilterOperator.GreaterThanOrEqual}${value}`;
}

function parseEnumToSelectOptions(enumObject: Record<string, string>): IEnumBasedFilterSelectOption[] {
  return Object.keys(enumObject).map((key) => ({
    label: stringUtil.capitalizeString(enumObject[key]),
    value: key,
  }));
}

/*Advanced filter query*/
function getFreeTextFilterQuery(name: string, term: string, label: string, operator: EFilterOperator): string {
  return `{ "name": "${name}", "term": "${term}", "label": "${label}", "value": "${operator}", "type": "${EColumnFilterType.FreeText}" }`;
}

function getNumericFilterQuery(
  name: string,
  label: string,
  operator: EFilterOperator,
  minValue: number,
  maxValue = 0,
): string {
  return `{ "name": "${name}", "minValue": "${minValue}", "maxValue": "${maxValue}", "label": "${label}", "operator": "${operator}", "type": "${EColumnFilterType.Numeric}" }`;
}
