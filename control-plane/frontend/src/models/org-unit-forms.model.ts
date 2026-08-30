import type { IRunaiWizardLayoutStep } from "./runai-wizard-layout.model";

/**
 * Tooltip texts for org unit forms (Projects & Departments)
 */
export enum EOrgUnitFormTooltip {
  ProjectCannotBeModified = "This can't be modified once a project has been created.",
  StepsDisabledUntilScopeComplete = "Complete the first step to enable the steps that follow",
  GpuQuotaAvailabilityProject = "Unassigned: The number of GPUs not assigned as quota to subordinate projects. Total: The total GPU quota assigned to the department in this node pool.",
  GpuQuotaAvailabilityDepartment = "Unassigned: The number of GPUs not assigned to subordinate projects. Total: The total number of GPUs assigned to the department in this node pool.",
}

/**
 * Text labels for GPU quota legend
 */
export enum EGpuQuotaLegendText {
  Assigned = "Assigned quota",
  Unassigned = "Unassigned quota",
  SumOfOtherProjects = "Sum of other projects' assigned quota",
  SumOfOtherDepartments = "Sum of other departments' assigned quota",
}

/**
 * Interface for legend items
 */
export interface ELegendItem {
  color: string;
  text: string;
}

/**
 * Save action types for form submission
 */
export enum ESaveOrgUnitAction {
  Close = "Close",
  Continue = "Continue",
  Save = "Save",
}
/**
 * Interface for save event payload
 */
export interface ISaveOrgUnitPayload<T> {
  project?: T;
  department?: T;
  action: ESaveOrgUnitAction;
  onComplete?: () => void; // Callback to execute after save completes (success or error)
}

/**
 * Peer org unit GPU quota change payload*/
export interface IPeerOrgUnitQuotaChange {
  orgUnitId: string;
  nodePoolName: string;
  nodePoolId: string;
  gpuQuota: number | null;
  originalLimit: number | null;
  originalOverQuotaWeight: number | null;
}

export enum EStepNames {
  ScopeName = "scope-name",
  GpuQuota = "gpu-quota",
  OverQuota = "over-quota",
  CPUQuota = "cpu-quota",
  SchedulingRules = "scheduling-rules",
  AccessRules = "access-rules",
}

export const DEPARTMENT_FORM_WIZARD_STEPS: IRunaiWizardLayoutStep[] = [
  { name: EStepNames.ScopeName, title: "Cluster & name", done: false, disabled: false, editTitle: "Description" },
  { name: EStepNames.GpuQuota, title: "GPU quota", done: false, disabled: true, editTitle: "GPU quota" },
  { name: EStepNames.OverQuota, title: "GPU over quota", done: false, disabled: true, editTitle: "GPU over quota" },
  { name: EStepNames.CPUQuota, title: "CPU quota", done: false, disabled: true, editTitle: "CPU quota" },
  {
    name: EStepNames.SchedulingRules,
    title: "Scheduling rules",
    done: false,
    disabled: true,
    editTitle: "Scheduling rules",
  },
  { name: EStepNames.AccessRules, title: "Access rules", done: false, disabled: true, editTitle: "Access rules" },
];

export const PROJECT_FORM_WIZARD_STEPS: IRunaiWizardLayoutStep[] = [
  { name: EStepNames.ScopeName, title: "Department & name", done: false, disabled: false, editTitle: "Description" },
  { name: EStepNames.GpuQuota, title: "GPU quota", done: false, disabled: true, editTitle: "GPU quota" },
  { name: EStepNames.OverQuota, title: "GPU over quota", done: false, disabled: true, editTitle: "GPU over quota" },
  { name: EStepNames.CPUQuota, title: "CPU quota", done: false, disabled: true, editTitle: "CPU quota" },
  {
    name: EStepNames.SchedulingRules,
    title: "Scheduling rules",
    done: false,
    disabled: true,
    editTitle: "Scheduling rules",
  },
  { name: EStepNames.AccessRules, title: "Access rules", done: false, disabled: true, editTitle: "Access rules" },
];

export const ORG_UNIT_STACKED_BAR_CHART_COLORS = {
  blue9: "#1565C0",
  blue5: "#42A5F5",
  blue2: "#BBDEFB",
};

/**
 * GPU resource allocation states for projects/departments.
 * Used to determine summary display text and aggregation across node pools.
 *
 * Values are ordered by priority (higher = more significant state):
 * - NoGPUResourcesAreSet (0): Workloads cannot run on GPUs
 * - OpportunisticOnly (1): No guaranteed resources, uses spare GPUs only
 * - DedicatedGPUs (2): Guaranteed resources, no over-quota allowed
 * - DedicatedAndOpportunisticGPUs (3): Guaranteed resources + can use spare GPUs
 */
export enum EGpuResourceState {
  NoGPUResourcesAreSet = 0,
  OpportunisticOnly = 1,
  DedicatedGPUs = 2,
  DedicatedAndOpportunisticGPUs = 3,
}

/**
 * Node pool total resources from cluster telemetry.
 * Used for department quota validation against node pool totals.
 */
export interface INodePoolTotals {
  [nodePoolName: string]: {
    cpuCores: number;
    memoryBytes: number;
    totalGpus: number;
  };
}
