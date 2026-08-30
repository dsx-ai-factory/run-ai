import { ECustomCell, type ICustomCellConfig, type ITableColumn } from "@/models/table.model";

export function viewLinkColumn(options: {
  name: string;
  label: string;
  emitName: string;
  text?: string;
  customCellConfig?: ICustomCellConfig;
}): ITableColumn {
  const { name, label, emitName, text, customCellConfig } = options;
  return {
    name,
    label,
    field: () => text ?? "View",
    sortable: false,
    align: "left",
    hideFilter: true,
    hideExport: true,
    customCell: ECustomCell.LINK_COL,
    customCellEvent: { emitName },
    customCellConfig,
  };
}
