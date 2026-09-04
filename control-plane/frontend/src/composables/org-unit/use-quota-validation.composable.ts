/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { computed, unref, type Ref, type ComputedRef } from "vue";
import { EQuotaInputState } from "@/models/org-unit.model";
import { formValidators } from "@/common/form.validators";
import { EQuotaEntity } from "@/models/resource.model";

export interface UseQuotaValidationOptions {
  quotaValue: Ref<number | null>;
  parentLimit: Ref<number | null>;
  allocatedNonPreemptible: Ref<number>;
  isLimitQuotaOverSubscriptionEnabled: Ref<boolean>;
  resourceName: string; // e.g., "CPU", "GPU", "Memory"
  resourceUnit: string | Ref<string> | ComputedRef<string>; // e.g., "cores", "GPUs", "GB"
  requireNumberGreaterOrEqualToZero?: boolean; // If true, null and values < 0 are invalid - probably only for CPU and Memory
  entityType?: EQuotaEntity; // "project" or "department" - used for messages
  subordinateQuota?: Ref<number | null | undefined>; // For departments: sum of child projects' quotas (undefined = no data, null = unlimited)
  showDefaultMessage?: boolean; // If false, only show error/warning messages, not informational hints (default: true)
  fieldLabel?: string; // Custom label for error messages (e.g., "Maximum GPUs"). Defaults to "${resourceName} quota"
  availabilityTotal: Ref<number | null>; // Total available resource in the node pool (for departments) , for project the total deserved resource in the department
  isDefaultDepartmentUnlimited?: Ref<boolean>; // When true (default dept has -1 quota), show exceeded warning without amounts for all departments
}

// New clusters default department GPU deserved value is -1
const INITIAL_DEFAULT_NODEPOOL_GPU_DESERVED_VALUE = -1;

export function useQuotaValidation(options: UseQuotaValidationOptions) {
  const {
    quotaValue,
    parentLimit,
    allocatedNonPreemptible,
    isLimitQuotaOverSubscriptionEnabled,
    resourceName,
    resourceUnit,
    requireNumberGreaterOrEqualToZero = false,
    entityType = EQuotaEntity.project,
    subordinateQuota,
    showDefaultMessage = true,
    fieldLabel,
    availabilityTotal,
    isDefaultDepartmentUnlimited,
  } = options;

  // Get display names based on entity type
  const entityName = entityType === EQuotaEntity.project ? "project" : "department";

  // Label used in error messages - use custom fieldLabel if provided, otherwise default to "${resourceName} quota"
  const displayLabel = fieldLabel ?? `${resourceName} quota`;

  const resolvedUnit = computed(() => unref(resourceUnit));
  const formatValue = (value: number | null): string => {
    if (value === null) {
      return "Unlimited";
    }
    const rounded = Math.round(value * 100) / 100;
    return Number.isInteger(rounded) ? rounded.toString() : rounded.toFixed(2);
  };
  const formatValueWithUnit = (value: number | null, unit: string): string => {
    const formatted = formatValue(value);
    const rounded = value !== null ? Math.round(value * 100) / 100 : null;
    const displayUnit = rounded === 1 && unit.endsWith("s") ? unit.slice(0, -1) : unit;
    return `${formatted} ${displayUnit}`;
  };

  const isNegative = computed(() => {
    return quotaValue.value !== null && quotaValue.value < -1;
  });

  const isInvalidNumberGreaterOrEqualToZero = computed(() => {
    if (!requireNumberGreaterOrEqualToZero) return false;
    if (quotaValue.value === null || quotaValue.value === undefined) return true;
    return !formValidators.isNumberGreaterOrEqualToZero(quotaValue.value);
  });

  const isExceededQuota = computed(() => {
    if (parentLimit.value === null || parentLimit.value === INITIAL_DEFAULT_NODEPOOL_GPU_DESERVED_VALUE) return false; // Parent is unlimited
    if (quotaValue.value === null || quotaValue.value === -1) return true; // Quota is unlimited, parent is limited
    return quotaValue.value > parentLimit.value;
  });

  const isUnderAllocatedNonPreemptible = computed(() => {
    if (quotaValue.value === null || quotaValue.value === -1) return false; // Unlimited
    return quotaValue.value < allocatedNonPreemptible.value;
  });

  // For departments: check if quota is less than what's assigned to subordinate projects
  const isUnderSubordinateQuota = computed(() => {
    // Skip if subordinateQuota is undefined (no data to validate)
    if (subordinateQuota?.value === undefined) return false;
    // Skip if subordinateQuota is null or 0 (unlimited or no child projects)
    if (!subordinateQuota?.value) return false;
    // Skip if department quota is unlimited
    if (quotaValue.value === null || quotaValue.value === -1) return false;
    return quotaValue.value < subordinateQuota.value;
  });

  // For departments: check if subordinate projects have unlimited quota but department has limited quota
  // If any child project is unlimited, department MUST be unlimited
  const isSubordinateUnlimitedButDepartmentLimited = computed(() => {
    // Only applicable for departments
    if (entityType !== EQuotaEntity.department) return false;

    // Skip if subordinateQuota is undefined (no data to validate - e.g., new department)
    if (subordinateQuota?.value === undefined) return false;

    // Check if subordinate quota is unlimited
    // subordinateQuota can be:
    // - undefined (no data - skip validation)
    // - null (API returns null for unlimited)
    // - -1 (unlimited in raw form)
    // - negative value close to 0 (e.g., -0.001 which is -1 after division by CPU_VALUE_FACTOR)
    const isSubordinateUnlimited =
      subordinateQuota.value === null ||
      subordinateQuota.value === -1 ||
      (subordinateQuota.value < 0 && subordinateQuota.value > -1);

    if (!isSubordinateUnlimited) return false; // Subordinate is limited, no issue

    // Check if department quota is limited (not null and not -1)
    const isDepartmentLimited = quotaValue.value !== null && quotaValue.value !== -1;

    // Error/Warning if subordinate is unlimited but department quota is limited
    return isDepartmentLimited;
  });

  const quotaState = computed(() => {
    if (isInvalidNumberGreaterOrEqualToZero.value) return EQuotaInputState.Error;
    if (isNegative.value) return EQuotaInputState.Error;

    // Check if subordinate is unlimited but department is limited
    if (isSubordinateUnlimitedButDepartmentLimited.value) {
      return isLimitQuotaOverSubscriptionEnabled.value ? EQuotaInputState.Error : EQuotaInputState.Warning;
    }

    // For departments with over-subscription enabled:
    // - Exceeding parent limit (node pool total) is a WARNING (we provide hints, not hard limits)
    // - Being under subordinate quota (child projects) is still an ERROR (must accommodate children)
    if (entityType === EQuotaEntity.department && isLimitQuotaOverSubscriptionEnabled.value) {
      // Subordinate quota check is always an error for departments
      if (isUnderSubordinateQuota.value) return EQuotaInputState.Error;

      // Parent limit exceeded or under allocated is a warning (not blocking)
      if (isExceededQuota.value) {
        return EQuotaInputState.Warning;
      }
    } else {
      // For projects (or departments without over-subscription enabled): use existing logic
      if (isExceededQuota.value || isUnderAllocatedNonPreemptible.value || isUnderSubordinateQuota.value) {
        return isLimitQuotaOverSubscriptionEnabled.value ? EQuotaInputState.Error : EQuotaInputState.Warning;
      }
    }

    // When default department is unlimited, show warning for all departments (not projects)
    if (isDefaultDepartmentUnlimited?.value && entityType === EQuotaEntity.department) {
      return EQuotaInputState.Warning;
    }

    return EQuotaInputState.Default;
  });

  const quotaMessage = computed(() => {
    const unit = resolvedUnit.value;
    if (isInvalidNumberGreaterOrEqualToZero.value) {
      return "Enter a value equal to or above 0";
    }
    if (isNegative.value) {
      return `${displayLabel} cannot be less than -1`;
    }
    if (isSubordinateUnlimitedButDepartmentLimited.value) {
      return `The department's ${resourceName} quota is limited but subordinate projects have unlimited quota. To remain within the limit, enter -1 for unlimited.`;
    }
    if (isUnderAllocatedNonPreemptible.value) {
      // If the allocated value is 0 or less, don't suggest entering a specific number
      if (allocatedNonPreemptible.value > 0) {
        return `The ${entityName}'s quota is less than the usage by non-preemptive workloads. To remain above the usage, enter a number above ${formatValueWithUnit(allocatedNonPreemptible.value, unit)}.`;
      }
      return `The ${entityName}'s quota is less than the usage by non-preemptive workloads.`;
    }
    if (isUnderSubordinateQuota.value && subordinateQuota?.value) {
      // If the subordinate quota is 0 or less, don't suggest entering a specific number
      if (subordinateQuota.value > 0) {
        return `The department's quota is less than the quota assigned to its subordinate projects. To remain within the limit, enter a number above ${formatValueWithUnit(subordinateQuota.value, unit)}.`;
      }
      return `The department's quota is less than the quota assigned to its subordinate projects.`;
    }
    if (isExceededQuota.value) {
      // For projects: exceeds department quota. For departments: exceeds node pool's total
      // If the parent limit is 0 or less, don't suggest entering a specific number
      if (parentLimit.value !== null && parentLimit.value > 0) {
        if (entityType === EQuotaEntity.project) {
          return `The project's quota exceeds the available department quota. To remain within the quota, enter a number up to ${formatValueWithUnit(parentLimit.value, unit)}.`;
        } else {
          return `The department's quota exceeds the node pool's total. To remain within the limit, enter a number up to ${formatValueWithUnit(parentLimit.value, unit)}.`;
        }
      }
      // Parent limit is 0 or null - just show the exceeded message without a target
      if (entityType === EQuotaEntity.project) {
        return `The project's quota exceeds the available department quota.`;
      } else {
        return `The department's quota exceeds the node pool's total.`;
      }
    }

    // When default department is unlimited, show warning without specific amounts for all departments
    if (isDefaultDepartmentUnlimited?.value && entityType === EQuotaEntity.department) {
      return `The total assigned quota exceeds the node pool's available GPU's.`;
    }

    // If showDefaultMessage is false, don't show informational hints
    if (!showDefaultMessage) {
      return "";
    }

    const hasUsage = allocatedNonPreemptible.value > 0;
    const hasLimit = parentLimit.value !== null && parentLimit.value !== INITIAL_DEFAULT_NODEPOOL_GPU_DESERVED_VALUE;

    // Default message (Recommended value)
    // For projects: show department's quota. For departments: show node pool's total
    const limitLabel = entityType === EQuotaEntity.project ? "Department's quota" : "Node pool's total";

    if (hasUsage && hasLimit) {
      return `Recommended: From ${formatValue(allocatedNonPreemptible.value)} to ${formatValue(parentLimit.value)}.<br> Non-preemptive workload usage: ${formatValueWithUnit(allocatedNonPreemptible.value, unit)}.<br> ${limitLabel}: ${formatValueWithUnit(parentLimit.value, unit)}`;
    }

    if (hasLimit) {
      return `Recommended: Up to ${formatValueWithUnit(parentLimit.value, unit)}.<br> ${limitLabel}: ${formatValueWithUnit(availabilityTotal.value, unit)}`;
    }

    if (hasUsage) {
      return `Non-preemptive workload usage: ${formatValueWithUnit(allocatedNonPreemptible.value, unit)}`;
    }

    return "";
  });

  return {
    isNegative,
    isInvalidNumberGreaterOrEqualToZero,
    isExceededQuota,
    isUnderAllocatedNonPreemptible,
    isUnderSubordinateQuota,
    isSubordinateUnlimitedButDepartmentLimited,
    quotaState,
    quotaMessage,
  };
}
