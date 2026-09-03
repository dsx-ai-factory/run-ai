/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IResourcesSectionModel, IGpuRequestModel } from "@/models/compute-resource.model";
import type { WorkspaceSpecSpecAllOfCompute } from "@/swagger-models/workloads-client/api";

export interface IGpuResourcesSummary {
  totalGpus: string;
  tooltipText: string;
}

export const computeResourceUtil = {
  computeResourceUISummary,
  totalGpuResources,
  totalGpuResourcesNumeric,
};

function computeResourceUISummary(resourcesModel: IResourcesSectionModel): string {
  return `GPU Devices: ${resourcesModel.gpuDevicesRequest || 0} / GPU Memory: ${_gpuSummary(
    resourcesModel,
  )} / CPU Compute: ${resourcesModel.cpuCore.request} Cores / CPU Memory: ${resourcesModel.cpuMemory.request}`;
}

function _gpuSummary(resourcesModel: IResourcesSectionModel): string | number {
  const { gpuPortionRequest, gpuMemoryRequest } = resourcesModel;
  if (gpuPortionRequest != null && gpuPortionRequest != undefined) {
    return `${gpuPortionRequest * 100}%`;
  }
  if (gpuMemoryRequest != null && gpuMemoryRequest != undefined) {
    return `${gpuMemoryRequest}B`;
  }
  return "None";
}

function totalGpuResources(
  gpuResources: Partial<IGpuRequestModel>,
  workers?: number,
  replicas?: { min?: number | null; max?: number | null },
): IGpuResourcesSummary {
  const totalGpus = gpuResources.gpuDevicesRequest || 0;
  if (!gpuResources.gpuRequestType) {
    return totalGpuDevices(gpuResources, workers, replicas);
  }

  if (gpuResources.gpuRequestType === "portion" && gpuResources.gpuPortionRequest) {
    return totalGpuPortion(gpuResources, gpuResources.gpuPortionRequest, workers, replicas);
  }

  if (gpuResources.gpuRequestType === "memory" && gpuResources.gpuMemoryRequest) {
    return totalGpuMemory(gpuResources, gpuResources.gpuMemoryRequest, workers, replicas);
  }

  return { totalGpus: `${totalGpus.toFixed(2)} GPUs`, tooltipText: `(${totalGpus.toFixed(2)} GPUs)` };
}

function totalGpuResourcesNumeric(
  gpuResources: WorkspaceSpecSpecAllOfCompute,
  workers?: number,
  replicas?: { min?: number | null; max?: number | null },
): number {
  const totalGpus = gpuResources.gpuDevicesRequest || 0;
  if (!gpuResources.gpuRequestType) {
    return totalGpuDevicesNumeric(gpuResources, workers, replicas);
  }

  if (gpuResources.gpuRequestType === "portion" && gpuResources.gpuPortionRequest) {
    return totalGpuPortionNumeric(gpuResources, gpuResources.gpuPortionRequest, workers, replicas);
  }

  return totalGpus;
}

function totalGpuDevices(
  gpuResources: Partial<IGpuRequestModel>,
  workers?: number,
  replicas?: { min?: number | null; max?: number | null },
): IGpuResourcesSummary {
  const replicasMin = replicas?.min || 0;
  const replicasMax = replicas?.max || 0;
  const gpuDevices = gpuResources.gpuDevicesRequest || 0;

  // Special handling for workers with elasticity
  if (workers && replicas) {
    const hasMin = replicas.min !== undefined && replicas.min !== null;
    const hasMax = replicas.max !== undefined && replicas.max !== null;

    // Workers with both min and max
    if (hasMin && hasMax) {
      if (replicasMin === replicasMax) {
        return {
          totalGpus: `${gpuDevices * replicasMin} GPUs`,
          tooltipText: `(${gpuDevices} GPUs) x (${replicas.min} workers)`,
        };
      }
      return {
        totalGpus: `${gpuDevices * replicasMin}-${gpuDevices * replicasMax} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${replicas.min}-${replicas.max} workers)`,
      };
    }

    // Workers with only max (e.g., workers: 1, max: 5 → show 1-5)
    if (hasMax) {
      return {
        totalGpus: `${gpuDevices * workers}-${gpuDevices * replicasMax} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${workers}-${replicas.max} workers)`,
      };
    }

    // Workers with only min (e.g., workers: 5, min: 2 → show 2-5)
    if (hasMin) {
      return {
        totalGpus: `${gpuDevices * replicasMin}-${gpuDevices * workers} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${replicas.min}-${workers} workers)`,
      };
    }
  }

  // If replicas are provided and show a range, prioritize showing the range
  if (
    replicas &&
    replicas.min !== undefined &&
    replicas.max !== undefined &&
    replicas.min !== null &&
    replicas.max !== null
  ) {
    const label = workers ? "Workers" : "Replicas";
    if (replicasMin === replicasMax) {
      return {
        totalGpus: `${gpuDevices * replicasMin} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${replicas.min} ${label})`,
      };
    }
    return {
      totalGpus: `${gpuDevices * replicasMin}-${gpuDevices * replicasMax} GPUs`,
      tooltipText: `(${gpuDevices} GPUs) x (${replicas.min}-${replicas.max} ${label})`,
    };
  }

  if (workers) {
    return {
      totalGpus: `${gpuDevices * workers} GPUs`,
      tooltipText: `(${gpuDevices} GPUs) x (${workers} Workers)`,
    };
  }

  return {
    totalGpus: `${gpuDevices} GPUs`,
    tooltipText: `(${gpuDevices} GPUs)`,
  };
}

function totalGpuPortion(
  gpuResources: Partial<IGpuRequestModel>,
  gpuPortionRequest: number,
  workers?: number,
  replicas?: { min?: number | null; max?: number | null },
): IGpuResourcesSummary {
  const replicasMin = replicas?.min || 0;
  const replicasMax = replicas?.max || 0;
  const gpuDevices = gpuResources.gpuDevicesRequest || 0;

  // Special handling for workers with elasticity
  if (workers && replicas) {
    const hasMin = replicas.min !== undefined && replicas.min !== null;
    const hasMax = replicas.max !== undefined && replicas.max !== null;

    // Workers with both min and max
    if (hasMin && hasMax) {
      const minTotalGpus = gpuDevices * gpuPortionRequest * replicasMin;
      const maxTotalGpus = gpuDevices * gpuPortionRequest * replicasMax;
      if (replicasMin === replicasMax) {
        return {
          totalGpus: `${minTotalGpus.toFixed(2)} GPUs`,
          tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%) x (${replicas.min} Workers)`,
        };
      }
      return {
        totalGpus: `${minTotalGpus.toFixed(2)}-${maxTotalGpus.toFixed(2)} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%) x (${replicas.min}-${
          replicas.max
        } Workers)`,
      };
    }

    // Workers with only max (e.g., workers: 1, max: 5 → show 1-5)
    if (hasMax) {
      const minTotalGpus = gpuDevices * gpuPortionRequest * workers;
      const maxTotalGpus = gpuDevices * gpuPortionRequest * replicasMax;
      return {
        totalGpus: `${minTotalGpus.toFixed(2)}-${maxTotalGpus.toFixed(2)} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%) x (${workers}-${
          replicas.max
        } Workers)`,
      };
    }

    // Workers with only min (e.g., workers: 5, min: 2 → show 2-5)
    if (hasMin) {
      const minTotalGpus = gpuDevices * gpuPortionRequest * replicasMin;
      const maxTotalGpus = gpuDevices * gpuPortionRequest * workers;
      return {
        totalGpus: `${minTotalGpus.toFixed(2)}-${maxTotalGpus.toFixed(2)} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%) x (${
          replicas.min
        }-${workers} Workers)`,
      };
    }
  }

  // If replicas are provided and show a range, prioritize showing the range
  if (
    replicas &&
    replicas.min !== undefined &&
    replicas.max !== undefined &&
    replicas.min !== null &&
    replicas.max !== null
  ) {
    const label = workers ? "Workers" : "Replicas";
    const minTotalGpus = gpuDevices * gpuPortionRequest * replicasMin;
    const maxTotalGpus = gpuDevices * gpuPortionRequest * replicasMax;
    if (replicasMin === replicasMax) {
      return {
        totalGpus: `${minTotalGpus.toFixed(2)} GPUs`,
        tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%) x (${replicas.min} ${label})`,
      };
    }
    return {
      totalGpus: `${minTotalGpus.toFixed(2)}-${maxTotalGpus.toFixed(2)} GPUs`,
      tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%) x (${replicas.min}-${
        replicas.max
      } ${label})`,
    };
  }

  if (workers) {
    return {
      totalGpus: `${(gpuDevices * gpuPortionRequest * workers).toFixed(2)} GPUs`,
      tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%) x (${workers} Workers)`,
    };
  }

  return {
    totalGpus: `${(gpuPortionRequest * gpuDevices).toFixed(2)} GPUs`,
    tooltipText: `(${gpuDevices} GPUs) x (${(gpuPortionRequest * 100).toFixed(0)}%)`,
  };
}

function totalGpuMemory(
  gpuResources: Partial<IGpuRequestModel>,
  gpuMemoryRequest: string,
  workers?: number,
  replicas?: { min?: number | null; max?: number | null },
): IGpuResourcesSummary {
  const memoryUnit = gpuMemoryRequest.endsWith("G") ? "GB" : "MB";
  const memoryValue = parseFloat(gpuMemoryRequest);
  const replicasMin = replicas?.min || 0;
  const replicasMax = replicas?.max || 0;
  const gpuDevices = gpuResources.gpuDevicesRequest || 0;

  // Special handling for workers with elasticity
  if (workers && replicas) {
    const hasMin = replicas.min !== undefined && replicas.min !== null;
    const hasMax = replicas.max !== undefined && replicas.max !== null;

    // Workers with both min and max
    if (hasMin && hasMax) {
      const minTotalGpus = gpuDevices * memoryValue * replicasMin;
      const maxTotalGpus = gpuDevices * memoryValue * replicasMax;
      if (replicasMin === replicasMax) {
        return {
          totalGpus: `${minTotalGpus.toFixed(0)} ${memoryUnit}`,
          tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit}) x (${replicas.min} Workers)`,
        };
      }
      return {
        totalGpus: `${minTotalGpus.toFixed(0)}-${maxTotalGpus.toFixed(0)} ${memoryUnit}`,
        tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit}) x (${replicas.min}-${replicas.max} Workers)`,
      };
    }

    // Workers with only max (e.g., workers: 1, max: 5 → show 1-5)
    if (hasMax) {
      const minTotalGpus = gpuDevices * memoryValue * workers;
      const maxTotalGpus = gpuDevices * memoryValue * replicasMax;
      return {
        totalGpus: `${minTotalGpus.toFixed(0)}-${maxTotalGpus.toFixed(0)} ${memoryUnit}`,
        tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit}) x (${workers}-${replicas.max} Workers)`,
      };
    }

    // Workers with only min (e.g., workers: 5, min: 2 → show 2-5)
    if (hasMin) {
      const minTotalGpus = gpuDevices * memoryValue * replicasMin;
      const maxTotalGpus = gpuDevices * memoryValue * workers;
      return {
        totalGpus: `${minTotalGpus.toFixed(0)}-${maxTotalGpus.toFixed(0)} ${memoryUnit}`,
        tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit}) x (${replicas.min}-${workers} Workers)`,
      };
    }
  }

  // If replicas are provided and show a range, prioritize showing the range
  if (
    replicas &&
    replicas.min !== undefined &&
    replicas.max !== undefined &&
    replicas.min !== null &&
    replicas.max !== null
  ) {
    const label = workers ? "Workers" : "Replicas";
    const minTotalGpus = gpuDevices * memoryValue * replicasMin;
    const maxTotalGpus = gpuDevices * memoryValue * replicasMax;
    if (replicasMin === replicasMax) {
      return {
        totalGpus: `${minTotalGpus.toFixed(0)} ${memoryUnit}`,
        tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit}) x (${replicas.min} ${label})`,
      };
    }
    return {
      totalGpus: `${minTotalGpus.toFixed(0)}-${maxTotalGpus.toFixed(0)} ${memoryUnit}`,
      tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit}) x (${replicas.min}-${replicas.max} ${label})`,
    };
  }

  if (workers) {
    return {
      totalGpus: `${(gpuDevices * memoryValue * workers).toLocaleString()} ${memoryUnit}`,
      tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit}) x (${workers} Workers)`,
    };
  }

  return {
    totalGpus: `${(gpuDevices * memoryValue).toLocaleString()} ${memoryUnit}`,
    tooltipText: `(${gpuDevices} GPUs) x (${memoryValue} ${memoryUnit})`,
  };
}

function totalGpuDevicesNumeric(
  gpuResources: WorkspaceSpecSpecAllOfCompute,
  workers?: number,
  replicas?: { min?: number | null; max?: number | null },
): number {
  const replicasMin = replicas?.min || 0;
  const gpuDevices = gpuResources.gpuDevicesRequest || 0;

  if (workers) {
    return gpuDevices * workers;
  }
  if (replicas) {
    return gpuDevices * replicasMin;
  }

  return gpuDevices;
}

function totalGpuPortionNumeric(
  gpuResources: WorkspaceSpecSpecAllOfCompute,
  gpuPortionRequest: number,
  workers?: number,
  replicas?: { min?: number | null; max?: number | null },
): number {
  const replicasMin = replicas?.min || 0;
  const gpuDevices = gpuResources.gpuDevicesRequest || 0;

  if (workers) {
    return gpuDevices * gpuPortionRequest * workers;
  } else if (replicas) {
    return gpuDevices * gpuPortionRequest * replicasMin;
  } else {
    return gpuPortionRequest * gpuDevices;
  }
}
