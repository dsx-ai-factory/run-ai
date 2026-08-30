import type { IAdvancedFilterModel, IEnumBasedFilterModel, IFreeTextFilterModel } from "@/models/filter.model";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { arrayToObjectMap } from "@/utils/common.util";

function getCellValue<T>(col: ITableColumn, entity: T): unknown {
  const raw = typeof col.field === "function" ? col.field(entity) : entity[col.field as keyof T];
  if (!col.format) return raw;
  const formatted = col.format(raw, entity);
  return col.filterKey ? (formatted as Record<string, unknown>)[col.filterKey] : formatted;
}

function matchesFreeText(cellValue: unknown, filter: IFreeTextFilterModel): boolean {
  const term = filter.term?.trim();
  if (!term) return true;

  if (Array.isArray(cellValue)) {
    return cellValue.some((item) => String(item).includes(term));
  }

  const haystack = String(cellValue ?? "").toLowerCase();
  const needle = term.toLowerCase();

  switch (filter.value) {
    case EFilterOperator.Equals:
      return haystack === needle;
    case EFilterOperator.NotEquals:
      return haystack !== needle;
    case EFilterOperator.DoesNotContain:
      return !haystack.includes(needle);
    case EFilterOperator.StartsWith:
      return haystack.startsWith(needle);
    case EFilterOperator.EndsWith:
      return haystack.endsWith(needle);
    case EFilterOperator.Contains:
    default:
      return haystack.includes(needle);
  }
}

function matchesEnumBased(cellValue: unknown, filter: IEnumBasedFilterModel): boolean {
  if (!filter.selectedValues?.length) return true;
  return filter.selectedValues.some((selected) => selected === cellValue || String(selected) === String(cellValue));
}

export function filterClustersByAdvancedFilters<T>(
  list: T[],
  filters: IAdvancedFilterModel[] | undefined,
  columns: ITableColumn[],
): T[] {
  if (!filters?.length) return list;

  const columnsMap = arrayToObjectMap(columns, "name");

  return list.filter((entity) =>
    filters.every((filter) => {
      const col = columnsMap[filter.name];
      if (!col) return true;
      if (col.ignoreFilter?.(entity)) return true;

      const cellValue = getCellValue(col, entity);

      switch (filter.type) {
        case EColumnFilterType.FreeText:
          return matchesFreeText(cellValue, filter as IFreeTextFilterModel);
        case EColumnFilterType.EnumBased:
          return matchesEnumBased(cellValue, filter as IEnumBasedFilterModel);
        default:
          return true;
      }
    }),
  );
}
