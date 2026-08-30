import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util";
import {
  ECustomCell,
  USER_DISPLAY_NAME_COL_DEFAULTS,
  type IStatusColOptions,
  type ITableColumn,
} from "@/models/table.model";
import { AssetSyncStatusEnum } from "@/swagger-models/assets-service-client";
import {
  type DatavolumeNoSharedScopes,
  DatavolumePhase,
  type DatavolumeStatus,
} from "@/swagger-models/datavolumes-service-client";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { dateUtil } from "@/utils/date.util";
import { getFieldContent } from "@/utils/table-format.util";

// Mapping of backend phase messages to frontend display text
const PHASE_MESSAGE_MAP: Record<string, string> = {
  ProjectOfOriginalPvcNotFound: "Project of original PVC not found",
  NamespaceOfOriginalPvcNotFound: "Namespace of original PVC not found",
  ScopeNotFound: "Scope not found",
  InternalError: "Internal error",
  OriginalPvcDoesntExist: "Original PVC doesn't exist",
  OriginalPvcNotBound: "Original PVC not bound",
  PvOfOriginalPvcDoesntExist: "PV for the original PVC is missing",
  PvOfOriginalPvcNotBound: "PV for the original PVC is missing",
  FailedToCreatePvcsPvs: "Failed to create PVCs or PVs",
  FailedToUpdatePvcsPvs: "Failed to update PVCs or PVs",
  FailedToDeletePvcsPvs: "Failed to delete PVCs or PVs",
};

function _transformDataVolumePhaseMessage(phaseMessage?: string): string | undefined {
  if (phaseMessage) {
    if (PHASE_MESSAGE_MAP[phaseMessage]) {
      return PHASE_MESSAGE_MAP[phaseMessage];
    }
  }
  return phaseMessage;
}

export const allDataVolumeColumnsMap: Record<string, ITableColumn> = {
  dataVolume: {
    name: "name",
    label: "Data volume",
    field: (row: DatavolumeNoSharedScopes) => row.name,
    sortable: true,
    align: "left",
    customCell: ECustomCell.NAME_COL,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.dataVolumes("name", (dv) => dv.name),
    },
  },
  description: {
    name: "description",
    label: "Description",
    field: (row: DatavolumeNoSharedScopes) => row.description,
    format: (val: string | null | undefined) => val || "-",
    align: "left",
    display: false,
    hideFilter: true,
    customCell: ECustomCell.DESCRIPTION_COL,
    sanitizeCell: true,
  },
  status: {
    name: "status",
    label: "Status",
    field: (row: DatavolumeNoSharedScopes): DatavolumeStatus => row.status,
    format(status: DatavolumeStatus, row: DatavolumeNoSharedScopes): IStatusColOptions {
      let assetStatus: AssetSyncStatusEnum;
      switch (status.phase) {
        case DatavolumePhase.Ready:
          assetStatus = AssetSyncStatusEnum.NoIssuesFound;
          break;
        case DatavolumePhase.NotReady:
          assetStatus = AssetSyncStatusEnum.IssuesFound;
          break;
        case DatavolumePhase.Creating:
          assetStatus = AssetSyncStatusEnum.Creating;
          break;
        case DatavolumePhase.Deleting:
          assetStatus = AssetSyncStatusEnum.Deleting;
          break;
        default:
          assetStatus = AssetSyncStatusEnum.NoStatus;
          break;
      }

      const transformedMessage = _transformDataVolumePhaseMessage(status.phaseMessage);
      return assetsUtil.getStatusColOptions({ status: assetStatus, message: transformedMessage }, row.createdAt || "");
    },
    hideFilter: true,
    sort: (a, b, rowA, rowB) => {
      const valB: IStatusColOptions = getFieldContent(allDataVolumeColumnsMap.status, rowB);
      const valA: IStatusColOptions = getFieldContent(allDataVolumeColumnsMap.status, rowA);
      return valA.status.toLowerCase().localeCompare(valB.status.toLowerCase());
    },
    customCell: ECustomCell.STATUS_COL,
    align: "left",
  },
  scope: {
    name: "scope",
    label: "Scope",
    field: (row: DatavolumeNoSharedScopes): string => row.projectId,
    format: () => "View",
    align: "left",
    customCell: ECustomCell.LINK_COL,
    hideFilter: true,
    customCellEvent: { emitName: "scope-col-clicked" },
  },
  originPvc: {
    name: "origin-pvc",
    label: "Origin PVC",
    field: (row: DatavolumeNoSharedScopes) => row.originPvcName,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.dataVolumesClientSide((dv) => dv.originPvcName),
    },
  },
  originProject: {
    name: "origin-project",
    label: "Origin project",
    field: (row: DatavolumeNoSharedScopes) => row.projectName,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.dataVolumes("projectName", (dv) => dv.projectName),
    },
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row: DatavolumeNoSharedScopes) => row.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
  },
  createdBy: {
    name: "createdBy",
    label: "Created by",
    field: (row: DatavolumeNoSharedScopes) => row.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.dataVolumes("createdBy", (dv) => dv.createdBy),
    },
  },
  createdAt: {
    name: "createdAt",
    label: "Creation time",
    field: (row: DatavolumeNoSharedScopes) => row.createdAt,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Date,
    },
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  lastUpdated: {
    name: "last-updated",
    label: "Last updated",
    field: (row: DatavolumeNoSharedScopes) => row.updatedAt,
    align: "left",
    hideFilter: true,
    format: (updatedAt: string) => dateUtil.dateAndTimeFormat(new Date(updatedAt)),
  },
};

export const allDataVolumeColumns: Array<ITableColumn> = [
  allDataVolumeColumnsMap.dataVolume,
  allDataVolumeColumnsMap.description,
  allDataVolumeColumnsMap.status,
  allDataVolumeColumnsMap.scope,
  allDataVolumeColumnsMap.originPvc,
  allDataVolumeColumnsMap.originProject,
  allDataVolumeColumnsMap.createdBy,
  allDataVolumeColumnsMap.createdAt,
  allDataVolumeColumnsMap.cluster,
];

export const dataVolumeIndexColumns: Array<ITableColumn> = [
  { ...allDataVolumeColumnsMap.dataVolume, display: true, mandatory: true },
  { ...allDataVolumeColumnsMap.description, display: false },
  { ...allDataVolumeColumnsMap.scope, display: true },
  { ...allDataVolumeColumnsMap.status, display: true },
  { ...allDataVolumeColumnsMap.originPvc, display: true },
  { ...allDataVolumeColumnsMap.originProject, display: false },
  { ...allDataVolumeColumnsMap.cluster, display: true },
  { ...allDataVolumeColumnsMap.createdBy, display: true },
  { ...allDataVolumeColumnsMap.createdAt, display: true },
  { ...allDataVolumeColumnsMap.lastUpdated, display: false },
];
