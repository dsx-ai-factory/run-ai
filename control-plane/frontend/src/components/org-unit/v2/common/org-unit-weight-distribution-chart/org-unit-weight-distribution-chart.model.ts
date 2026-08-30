/**
 * Props interface for the OrgUnitWeightDistributionChart component
 */
export interface IOrgUnitWeightDistributionChartProps {
  /** Entity name (e.g., project or department name) to display as tooltip title */
  entityName: string;
  /** Entity type (e.g., 'project', 'department') used in labels */
  entityType: string;
  /** Weight value for this entity */
  weight: number;
  /** Total weight of other sibling entities */
  otherEntitiesWeight: number;
  /** Height of the chart in pixels (default: 20) */
  height?: number;
  /** Whether to show value labels on segments (default: true) */
  showLabels?: boolean;
  /** Whether to show the scale labels below the chart (default: true) */
  showScale?: boolean;
}
