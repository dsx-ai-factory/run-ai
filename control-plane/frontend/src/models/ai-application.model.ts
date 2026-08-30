import type { AIApplication } from "@/swagger-models/workloads-client";

export enum EApplicationPhase {
  Deployed = "Deployed",
  Failed = "Failed",
}

/**
 * Extended interface for AI Application table display
 * Extends the base Application interface with additional fields for UI
 */
export interface IApplicationTable extends AIApplication {
  departmentName?: string; // Will be enriched from project data
  departmentId?: string; // Will be enriched from project data
  clusterName?: string; // Will be enriched from cluster data
  clusterId?: string; // Will be enriched from cluster data
}

/**
 * Column names for AI Application table
 */
export enum EApplicationColumnName {
  Name = "name",
  Type = "type",
  Status = "phase",
  Department = "departmentName",
  Project = "projectName",
  GpuComputeRequest = "gpuComputeRequest",
  GpuComputeAllocation = "gpuComputeAllocation",
  GpuMemoryRequest = "gpuMemoryRequest",
  GpuMemoryAllocation = "gpuMemoryAllocation",
  CpuComputeRequest = "cpuComputeRequest",
  CpuComputeAllocation = "cpuComputeAllocation",
  CpuMemoryRequest = "cpuMemoryRequest",
  CpuMemoryAllocation = "cpuMemoryAllocation",
  Connections = "connections",
  Cluster = "clusterId",
}

/**
 * Column labels for AI Application table
 */
export enum EApplicationColumnLabel {
  Name = "AI application",
  Type = "Type",
  Status = "Status",
  Department = "Department",
  Project = "Project",
  GpuComputeRequest = "GPU compute request",
  GpuComputeAllocation = "GPU compute allocation",
  GpuMemoryRequest = "GPU memory request",
  GpuMemoryAllocation = "GPU memory allocation",
  CpuComputeRequest = "CPU compute request",
  CpuComputeAllocation = "CPU compute allocation",
  CpuMemoryRequest = "CPU memory request",
  CpuMemoryAllocation = "CPU memory allocation",
  Connections = "Connection(s)",
  Cluster = "Cluster",
}

/**
 * Sort/Filter field names for AI Applications
 * These match the API's sortBy and filterBy parameters
 */
export enum ApplicationSortFilterFields {
  ClusterId = "clusterId",
  Name = "name",
  ProjectId = "projectId",
  ProjectName = "projectName",
  Source = "source",
  Type = "type",
  CreatedAt = "createdAt",
}

/**
 * Error messages for AI Application operations
 */
export enum EAiApplicationErrorMessage {
  FailedToLoadApplicationDetails = "Failed to load AI application details",
  FailedToLoadApplicationWorkloads = "Failed to load AI application workloads",
}
