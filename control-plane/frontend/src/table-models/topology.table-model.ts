import { ECustomCell, type ITableColumn } from "@/models/table.model";
import type { NetworkTopology } from "@/swagger-models/cluster-service-client";
import { dateUtil } from "@/utils/date.util";
import { tableFormatUtil } from "@/utils/table-format.util/table-format.util";

export const topologyColumnName = {
  name: "name",
  labels: "labels",
  createdBy: "created-by",
  createdAt: "created-at",
  nodePools: "node-pools",
};

enum ETopologyColumnLabel {
  Name = "Topology",
  Labels = "Labels",
  CreatedBy = "Created by",
  CreatedAt = "Creation time",
  NodePools = "Node pools",
}

export const allTopologyColumnsMap: Record<keyof typeof topologyColumnName, ITableColumn> = {
  name: {
    name: topologyColumnName.name,
    label: ETopologyColumnLabel.Name,
    field: (row: NetworkTopology) => row.name,
    sortable: false,
    align: "left",
  },
  labels: {
    name: topologyColumnName.labels,
    label: ETopologyColumnLabel.Labels,
    field: (row: NetworkTopology) => row.levels,
    sortable: false,
    align: "left",
    format: (labels: string[]) => labels.join(", ") || "-",
    customCell: ECustomCell.DESCRIPTION_COL,
    sanitizeCell: true,
  },
  createdBy: {
    name: topologyColumnName.createdBy,
    label: ETopologyColumnLabel.CreatedBy,
    field: (row: NetworkTopology) => row.createdBy,
    sortable: false,
    align: "left",
    format: (createdBy: string) => createdBy || "-",
  },
  createdAt: {
    name: topologyColumnName.createdAt,
    label: ETopologyColumnLabel.CreatedAt,
    field: (row: NetworkTopology) => row.createdAt,
    sortable: false,
    align: "left",
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  nodePools: {
    name: topologyColumnName.nodePools,
    label: ETopologyColumnLabel.NodePools,
    field: (row: NetworkTopology) => row.nodePools,
    sortable: false,
    align: "left",
    format: (nodePools: string[] | undefined): string[] | string => {
      if (!nodePools || nodePools.length === 0) {
        return "-";
      }
      return tableFormatUtil.formatArrayStrings(nodePools);
    },
  },
};

export const allTopologyColumns: Array<ITableColumn> = [
  allTopologyColumnsMap.name,
  allTopologyColumnsMap.labels,
  allTopologyColumnsMap.nodePools,
  allTopologyColumnsMap.createdBy,
  allTopologyColumnsMap.createdAt,
];

export const topologyPreviewColumns: Array<ITableColumn> = [
  { ...allTopologyColumnsMap.name, display: true, mandatory: true },
  { ...allTopologyColumnsMap.labels, display: true },
  { ...allTopologyColumnsMap.nodePools, display: true },
  { ...allTopologyColumnsMap.createdBy, display: true },
  { ...allTopologyColumnsMap.createdAt, display: true },
];

export interface ITopologyPreviewModalOptions {
  sortByDefault: string;
  entityName: string;
  entityType: ETopologyPreviewModalEntityType;
  columns: ITableColumn[];
  clusterId: string;
}

export enum ETopologyPreviewModalEntityType {
  Cluster = "cluster",
  NodePool = "nodepool",
}

export interface ITopologyLabelsModalOptions {
  topologyId: string;
  topologyName?: string;
}

export type ITopologyLabelPreview = {
  name: string;
};

export const nodePoolTopologyLabelsTableColumnsMap: Record<string, ITableColumn> = {
  labels: {
    name: "labels",
    label: "Labels",
    field: (row: ITopologyLabelPreview) => row.name,
    sortable: false,
    align: "left",
    sanitizeCell: true,
  },
};

export const nodePoolTopologyLabelsTableColumns: Array<ITableColumn> = [
  { ...nodePoolTopologyLabelsTableColumnsMap.labels, display: true, mandatory: true },
];
