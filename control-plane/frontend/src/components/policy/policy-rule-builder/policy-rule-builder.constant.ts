/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ISelectOption } from "@/models/global.model";
import { EEnforcementType, type IEnforcementDraft } from "@/models/gvk-policy.model";
import { errorMessages } from "@/common/error-message.constant";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";

export const ENFORCEMENT_META: Record<EEnforcementType, { label: string; description: string }> = {
  [EEnforcementType.Set]: {
    label: "Set",
    description: "Enforces the value of the field.",
  },
  [EEnforcementType.Default]: {
    label: "Default",
    description: "Sets the field to the given value unless the user supplies one.",
  },
  [EEnforcementType.Required]: {
    label: "Required",
    description: "Requires the field to have a value, supplied either in the submission or by a policy default.",
  },
  [EEnforcementType.Blocked]: {
    label: "Blocked",
    description: 'Prevents the field from being assigned any value. Inverse of "required".',
  },
  [EEnforcementType.Length]: {
    label: "Length",
    description:
      "Enforces the field's length (character count for strings, element count for slices, arrays, and maps) to fall within the inclusive range [min, max]. At least one of min/max must be specified.",
  },
  [EEnforcementType.Options]: {
    label: "Options",
    description: "Restricts the value of the field to one of the listed options.",
  },
  [EEnforcementType.Match]: {
    label: "Match",
    description: "Requires the value of the field to match the regular expression.",
  },
  [EEnforcementType.Range]: {
    label: "Range",
    description:
      "Enforces a numeric value to fall within the inclusive range [min, max], optionally restricted to increments of step starting from min.",
  },
  [EEnforcementType.QuantityRange]: {
    label: "Quantity range",
    description:
      "Enforces a Kubernetes resource quantity to fall within the inclusive range [min, max], optionally restricted to increments of step starting from min. All bounds must be valid Kubernetes quantities such as 100m, 1Gi, or 500Mi.",
  },
  [EEnforcementType.Elements]: {
    label: "Elements",
    description:
      "Enforces the number of elements in an array field to fall within the inclusive range [min, max]. At least one of min/max must be specified.",
  },
  [EEnforcementType.Items]: {
    label: "Items",
    description:
      "Enforces the number of items produced by a field's itemSelector to fall within the inclusive range [min, max]. At least one of min/max must be specified. Applies only to fields that declare itemSelectors.",
  },
};

export function buildEnforcementOptions(
  allowed: EEnforcementType[],
  usedEnforcements: IEnforcementDraft[],
  setDefaultGroupingArg: string | null,
): ISelectOption[] {
  const usedEnforcementTypes = usedEnforcements
    .map((enforcement) => enforcement.type)
    .filter((type): type is EEnforcementType => type !== null);
  const used = new Set(usedEnforcementTypes);
  return allowed.map((type) => {
    const meta = ENFORCEMENT_META[type];
    const isDuplicateType = used.has(type) && !gvkPolicyUtil.enforcementMayRepeat(type);
    const isFlatSetDefaultConflict =
      setDefaultGroupingArg === null &&
      gvkPolicyUtil.isSetOrDefault(type) &&
      usedEnforcementTypes.some(gvkPolicyUtil.isSetOrDefault);
    const isIncompatible = usedEnforcements.some((enforcement) =>
      gvkPolicyUtil.enforcementTypeUnavailableWith(type, enforcement),
    );
    const disable = isDuplicateType || isFlatSetDefaultConflict || isIncompatible;
    return {
      label: meta.label,
      value: type,
      aid: `enforcement-type-option-${type}`,
      toolTip: meta.description,
      disable,
      disabledTooltip: isDuplicateType
        ? errorMessages.DUPLICATE_ENFORCEMENT
        : isFlatSetDefaultConflict
          ? errorMessages.DUPLICATE_SET_DEFAULT_TARGET
          : isIncompatible
            ? errorMessages.INCOMPATIBLE_ENFORCEMENTS
            : undefined,
    };
  });
}

export const BOOLEAN_VALUE_OPTIONS: ISelectOption[] = [
  { label: "True", value: "true" },
  { label: "False", value: "false" },
];

// Suffix subset accepted by both backend resource.ParseQuantity and frontend isValidK8sQuantity; "" is unitless.
export const K8S_QUANTITY_UNIT_OPTIONS: ISelectOption[] = [
  { label: "None", value: "" },
  { label: "m", value: "m" },
  { label: "Ki", value: "Ki" },
  { label: "Mi", value: "Mi" },
  { label: "Gi", value: "Gi" },
  { label: "Ti", value: "Ti" },
  { label: "Pi", value: "Pi" },
  { label: "Ei", value: "Ei" },
  { label: "k", value: "k" },
  { label: "M", value: "M" },
  { label: "G", value: "G" },
  { label: "T", value: "T" },
  { label: "P", value: "P" },
  { label: "E", value: "E" },
];

export const NO_RESULTS_FOUND_LABEL = "No results found";
