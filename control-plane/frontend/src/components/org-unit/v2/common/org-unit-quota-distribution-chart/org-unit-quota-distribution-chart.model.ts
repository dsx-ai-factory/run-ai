/**
 * Props interface for the OrgUnitQuotaDistributionChart component
 */
export interface IOrgUnitQuotaDistributionChartProps {
  /** Entity name (project or department) to display as tooltip title */
  entityName: string;
  /** Entity type (e.g., 'project', 'department') used in labels */
  entityType: string;
  /** Assigned quota value for this entity */
  assignedQuota: number;
  /** Parent entity name (e.g., department name for projects, cluster name for departments) for the tooltip summary */
  parentName: string;
  /** Total available quota from the parent entity (used as maxValue) */
  parentAvailableQuota: number;
  /** Sum of assigned quota from other sibling entities */
  otherEntitiesAssignedQuota: number;
  /** Height of the chart in pixels (default: 20) */
  height?: number;
  /** Whether to show value labels on segments (default: true) */
  showLabels?: boolean;
  /** Whether to show the scale labels below the chart (default: true) */
  showScale?: boolean;
}
