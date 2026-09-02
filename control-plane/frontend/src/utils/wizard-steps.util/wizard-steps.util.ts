/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import type { ISelectOption } from "@/models/global.model";
import { EStepNames } from "@/models/org-unit-forms.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

export interface INodepoolCapabilities {
  cpuQuota?: boolean;
  quota?: boolean;
}

export interface IStepFilterContext {
  nodepoolCapabilities?: INodepoolCapabilities;
  hasPermission?: (resourceType: ResourceType, action: Action) => boolean;
}

/**
 * Filters org unit wizard steps based on nodepool capabilities and permissions
 *
 * @param steps - Array of wizard steps to filter
 * @param context - Filtering context containing capabilities and permission checker
 * @returns Filtered array of steps
 *
 * @example
 * ```ts
 * const filtered = filterOrgUnitSteps(DEPARTMENT_FORM_WIZARD_STEPS, {
 *   nodepoolCapabilities: { cpuQuota: true, quota: false },
 *   hasPermission: (r, a) => permissionStore.hasPermission(r, a),
 * });
 * ```
 */
export function filterOrgUnitSteps(
  steps: IRunaiWizardLayoutStep[],
  context: IStepFilterContext,
): IRunaiWizardLayoutStep[] {
  return steps.filter((step) => {
    if (step.name === EStepNames.CPUQuota && !context.nodepoolCapabilities?.cpuQuota) {
      return false;
    }
    if (step.name === EStepNames.GpuQuota && !context.nodepoolCapabilities?.quota) {
      return false;
    }
    if (step.name === EStepNames.AccessRules && context.hasPermission) {
      return context.hasPermission(ResourceType.AccessRules, Action.Read);
    }
    return true;
  });
}

/**
 * Converts wizard steps to select options format (label/value)
 *
 * @param steps - Array of wizard steps
 * @returns Array of select options
 */
export function stepsToSelectOptions(steps: IRunaiWizardLayoutStep[]): ISelectOption[] {
  return steps.map((step) => ({
    label: step.editTitle ?? step.title,
    value: step.name,
  }));
}
