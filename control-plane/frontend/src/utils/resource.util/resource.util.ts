import type { IResourceQuota } from "@/models/project.model";
import {
  CPU_VALUE_FACTOR,
  DEFAULT_NODE_POOL_NAME,
  EMemoryUnit,
  EOverQuotaLabel,
  EOverQuotaPriorityLabel,
  EResourceState,
  EResourceType,
  k8sUnitMultiplier,
} from "@/models/resource.model";
import { type Resources, type ResourcesNullable } from "@/swagger-models/org-unit-service-client";

const MIB_TO_MB_FACTOR = 1.048576;
const MIB_TO_GB_FACTOR = 953.67431640625;
const MEMORY_SIZE_THRESHOLD = 9536.7431640625;
const BYTE_TO_MIB_FACTOR = 1048576;
export const resourceUtil = {
  getResourceDisplayValue,
  convertMemoryValue,
  fromMbToMib,
  sortNodePoolNames,
  sortResourcesByNodePoolOrder,
  fromBytesToMiB,
  getNodePoolResourceOverQuotaWeightDisplayedValue,
  getOverQuotaPriorityKeyByValue,
  getOverQuotaKeyByValue,
  parseK8sResourceSizeToBytes,
  convertBytesToK8sResourceSizeUnit,
};

function getResourceDisplayValue(value: number | null | undefined, resourceType: EResourceType): string {
  if (value === null || value === undefined || value === -1) {
    if (resourceType === EResourceType.CPU || resourceType === EResourceType.MEMORY) {
      return EResourceState.Unlimited;
    }
    if (value === -1 && (resourceType === EResourceType.GPU || resourceType === EResourceType.GPU_LIMIT)) {
      return EResourceState.Unlimited;
    }
    return "0.00";
  }

  if (resourceType === EResourceType.CPU) {
    return (value / CPU_VALUE_FACTOR).toFixed(2);
  } else if (resourceType === EResourceType.MEMORY) {
    const convertedValue = convertMemoryValue(value);
    return `${convertedValue.value.toFixed(2)} ${convertedValue.selectedUnit}`;
  }
  if (resourceType === EResourceType.GPU || resourceType === EResourceType.GPU_LIMIT) {
    if (value < 0) {
      return "Unlimited";
    }
  }
  return value.toFixed(2);
}

function fromMibToMb(mib: number): number {
  return mib * MIB_TO_MB_FACTOR;
}

function fromMibToGb(mib: number): number {
  return mib / MIB_TO_GB_FACTOR;
}

function fromMbToMib(mib: number): number {
  return mib / MIB_TO_MB_FACTOR;
}

function fromBytesToMiB(bytes: number): number {
  return bytes / BYTE_TO_MIB_FACTOR;
}

function convertMemoryValue(value: number): { value: number; selectedUnit: EMemoryUnit } {
  const beforeDecimalPointPart = Math.trunc(value);
  if (beforeDecimalPointPart > MEMORY_SIZE_THRESHOLD) {
    return {
      value: fromMibToGb(value),
      selectedUnit: EMemoryUnit.GB,
    };
  } else {
    return {
      value: fromMibToMb(value),
      selectedUnit: EMemoryUnit.MB,
    };
  }
}

/**
 * Sort resources by node pool priority order.
 * Items matching priorityOrder come first (in that order), then the rest.
 */
function sortResourcesByNodePoolOrder<T extends Resources | ResourcesNullable>(
  resources: T[],
  priorityOrder: string[],
): T[] {
  return [...resources].sort((a, b) => {
    const indexA = priorityOrder.indexOf(a.nodePool?.name ?? "");
    const indexB = priorityOrder.indexOf(b.nodePool?.name ?? "");

    if (indexA !== -1 && indexB !== -1) return indexA - indexB;
    if (indexA !== -1) return -1;
    if (indexB !== -1) return 1;
    return 0;
  });
}

/**
 * Sort node pool names alphabetically with "default" always first.
 */
function sortNodePoolNames(nodePoolNames: string[]): string[] {
  const sorted = [...nodePoolNames].sort((a, b) => a.localeCompare(b));
  const defaultIndex = sorted.indexOf(DEFAULT_NODE_POOL_NAME);
  if (defaultIndex === -1) return sorted;
  const sortedWithoutDefault = sorted.filter((name) => name !== DEFAULT_NODE_POOL_NAME);
  return [DEFAULT_NODE_POOL_NAME, ...sortedWithoutDefault];
}

function getNodePoolResourceOverQuotaWeightDisplayedValue(
  resourceQuota: IResourceQuota,
  isOverQuotaPriorityEnabled: boolean,
): string {
  if (isOverQuotaPriorityEnabled) {
    return resourceUtil.getOverQuotaPriorityKeyByValue(resourceQuota?.overQuotaWeight);
  } else {
    return resourceUtil.getOverQuotaKeyByValue(resourceQuota?.overQuotaWeight);
  }
}

function parseK8sResourceSizeToBytes(size: string): number {
  const regex = /^([+-]?[0-9.]+)([eEinumkKMGTP]*i?)?$/;
  const match = size.match(regex);

  if (!match) {
    throw new Error(`Invalid Kubernetes resource size: ${size}`);
  }

  const value = parseFloat(match[1]);
  const unit = match[2] || "";

  if (!k8sUnitMultiplier[unit]) {
    throw new Error(`Unknown unit: ${unit}`);
  }

  return value * k8sUnitMultiplier[unit];
}

function getOverQuotaPriorityKeyByValue(value: number | null | undefined): string {
  switch (value) {
    case -1:
      return EOverQuotaPriorityLabel.None;
    case 0:
      return EOverQuotaPriorityLabel.Lowest;
    case 1:
      return EOverQuotaPriorityLabel.Low;
    case 2:
      return EOverQuotaPriorityLabel.Medium;
    case 3:
      return EOverQuotaPriorityLabel.High;
    default:
      return EOverQuotaPriorityLabel.Medium;
  }
}

function getOverQuotaKeyByValue(value: number | null | undefined): string {
  switch (value) {
    case 0:
      return EOverQuotaLabel.Disabled;
    case 2:
      return EOverQuotaLabel.Enabled;
    default:
      return EOverQuotaLabel.Enabled;
  }
}

function convertBytesToK8sResourceSizeUnit(sizeInBytes: number, targetUnit: string): number {
  // Ensure the target unit is valid
  if (!k8sUnitMultiplier[targetUnit]) {
    throw new Error(`Invalid target unit: ${targetUnit}`);
  }

  // Perform conversion
  return sizeInBytes / k8sUnitMultiplier[targetUnit];
}
