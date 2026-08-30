import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS } from "@/models/table.model";
import type { ITableColumn } from "@/models/table.model";
import { EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { formatDate } from "@/utils/table-format.util";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";
import { viewLinkColumn } from "@/table-models/common-columns";

export const allServiceAccountsColumnsMap: Record<string, ITableColumn> = {
  serviceAccount: {
    name: "serviceAccount",
    label: "Service account",
    field: (row: ServiceAccount) => row.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.serviceAccounts((sa) => sa.name),
    },
  },
  clientID: {
    name: "clientID",
    label: "Client ID",
    field: (row: ServiceAccount) => row.clientId,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.serviceAccounts((sa) => sa.clientId),
    },
  },
  status: {
    name: "status",
    label: "Status",
    field: (row: ServiceAccount) => row.enabled,
    sortable: true,
    customCell: ECustomCell.APP_STATUS_COL,
    align: "left",
    exportFormat: (row: ServiceAccount) => (row.enabled ? "Active" : "Revoked"),
  },
  accessRules: viewLinkColumn({
    name: "accessRules",
    label: "Access rule(s)",
    emitName: "access-rules-clicked",
  }),
  lastLogin: {
    name: "lastLogin",
    label: "Last login",
    field: (row: ServiceAccount) => row.lastLogin,
    sortable: true,
    align: "left",
    format: (lastLogin) => formatDate(lastLogin),
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row: ServiceAccount) => row.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.serviceAccounts((sa) => sa.createdBy),
    },
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row: ServiceAccount) => row.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => formatDate(createdAt),
  },
  lastUpdated: {
    name: "last-updated",
    label: "Last updated",
    field: (row: ServiceAccount) => row.updatedAt,
    sortable: true,
    align: "left",
    format: (updatedAt: string) => formatDate(updatedAt),
  },
};

export const serviceAccountsColumns: Array<ITableColumn> = [
  allServiceAccountsColumnsMap.serviceAccount,
  allServiceAccountsColumnsMap.clientID,
  allServiceAccountsColumnsMap.status,
  allServiceAccountsColumnsMap.accessRules,
  allServiceAccountsColumnsMap.lastLogin,
  allServiceAccountsColumnsMap.createdBy,
  allServiceAccountsColumnsMap.createdAt,
  allServiceAccountsColumnsMap.lastUpdated,
];

export const serviceAccountsIndexColumns: Array<ITableColumn> = [
  { ...allServiceAccountsColumnsMap.serviceAccount, display: true, mandatory: true },
  { ...allServiceAccountsColumnsMap.clientID, display: true },
  { ...allServiceAccountsColumnsMap.status, display: false },
  { ...allServiceAccountsColumnsMap.accessRules, display: true },
  { ...allServiceAccountsColumnsMap.lastLogin, display: true },
  { ...allServiceAccountsColumnsMap.createdBy, display: true },
  { ...allServiceAccountsColumnsMap.createdAt, display: true },
  { ...allServiceAccountsColumnsMap.lastUpdated, display: true },
];
