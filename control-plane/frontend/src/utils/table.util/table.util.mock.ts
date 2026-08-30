import type { ITableColumn } from "@/models/table.model";

export const columnA: ITableColumn = {
  name: "string",
  label: "String",
  field: (row) => row.string,
  sortable: true,
  align: "left",
};
export const columnB: ITableColumn = {
  name: "number",
  label: "Number",
  field: (row) => row.number,
  sortable: true,
  align: "left",
};
export const rowA = {
  string: "a-department",
};
export const rowB = {
  string: "b-department",
};
export const rowC = {
  number: 1,
};
export const rowD = {
  number: 2,
};
