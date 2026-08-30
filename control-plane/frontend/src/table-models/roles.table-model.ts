import { EColumnFilterType, type IFilterBy } from "@/models/filter.model";
import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import type { Role } from "@/swagger-models/authorization-client";

import { dateUtil } from "@/utils/date.util";

export const allRoleColumnsMap: Record<string, ITableColumn> = {
  role: {
    name: "role",
    label: "Role",
    field: (row: Role) => row.name,
    sortable: true,
    align: "left",
    customCell: ECustomCell.ROLE_WITH_BADGE_COL,
    customCellEvent: { emitName: "role-clicked" },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.roles((r) => r.name),
    },
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row: Role) => row.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.roles((r) => r.createdBy),
    },
  },
  creationTime: {
    name: "creationTime",
    label: "Creation time",
    field: (row: Role) => row.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => createdAt && dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
};

export const allRoleColumns: Array<ITableColumn> = [
  allRoleColumnsMap.role,
  allRoleColumnsMap.createdBy,
  allRoleColumnsMap.creationTime,
];

export const roleIndexColumns: Array<ITableColumn> = [
  { ...allRoleColumnsMap.role, display: true, mandatory: true },
  { ...allRoleColumnsMap.createdBy, display: true },
  { ...allRoleColumnsMap.creationTime, display: false },
];

export const defaultRolesFilterBy: IFilterBy = {
  sortBy: "role",
  descending: false,
  page: 1,
  rowsPerPage: 20,
  columnFilters: [],
  searchTerm: "",
  displayedColumns: [allRoleColumnsMap.role.name, allRoleColumnsMap.createdBy.name],
};
