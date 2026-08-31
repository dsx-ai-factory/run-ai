/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Policy Field Iterator Utility
 *
 * This utility provides a generic way to iterate over policy rules and apply operations
 * to corresponding workload model fields. It handles:
 * - Simple fields (e.g., rules.command -> spec.command)
 * - Nested fields (e.g., rules.compute.cpuCoreRequest -> spec.compute.cpuCoreRequest)
 * - Array item attributes (e.g., rules.tolerations.attributes.key -> spec.tolerations[*].key)
 *
 * The utility uses recursive traversal to find all blocked fields and clear them.
 */

import type { IUISuperWorkloadSpec } from "@/models/workload.model";

/**
 * Generic policy type that can accept both UnifiedPolicyInfoPerReplicaType and
 * InferencePolicyDefaultsAndRulesV2 (or any policy with a rules object).
 */
export type TPolicyWithRules = {
  rules?: Record<string, unknown> | null;
  defaults?: Record<string, unknown> | null;
};

// ============================================================================
// Types
// ============================================================================

/**
 * Represents a rule object that may have a `blocked` property.
 * This is a common interface for StringRules, NumberRules, BooleanRules, etc.
 */
export interface IBlockableRule {
  blocked?: boolean | null;
  canEdit?: boolean | null;
  required?: boolean | null;
  sourceOfRule?: unknown;
}

/**
 * Information about a blocked field found during traversal.
 */
export interface IBlockedFieldInfo {
  /** Full path in the policy rules (e.g., "tolerations.attributes.key") */
  policyPath: string;
  /** Path in the model, may include array marker (e.g., "tolerations[*].key") */
  modelPath: string;
  /** Whether this is an array item attribute (affects how we clear it) */
  isArrayItemAttribute: boolean;
  /** The array field name if isArrayItemAttribute is true */
  arrayFieldName?: string;
  /** The attribute path within the array item */
  attributePath?: string;
}

// ============================================================================
// Path Utilities
// ============================================================================

/**
 * Safely gets a nested value from an object using a dot-notation path.
 */
export function getValueByPath<T = unknown>(
  obj: Record<string, unknown> | null | undefined,
  path: string,
): T | undefined {
  if (!obj || !path) return undefined;

  const keys = path.split(".");
  let current: unknown = obj;

  for (const key of keys) {
    if (current === null || current === undefined) {
      return undefined;
    }
    if (typeof current !== "object") {
      return undefined;
    }
    current = (current as Record<string, unknown>)[key];
  }

  return current as T;
}

/**
 * Safely deletes a nested value from an object.
 */
export function deleteValueByPath(obj: Record<string, unknown>, path: string): void {
  if (!obj || !path) return;

  const keys = path.split(".");
  let current: Record<string, unknown> = obj;

  for (let i = 0; i < keys.length - 1; i++) {
    const key = keys[i];
    if (current[key] === undefined || current[key] === null) {
      return;
    }
    if (typeof current[key] !== "object") {
      return;
    }
    current = current[key] as Record<string, unknown>;
  }

  const lastKey = keys[keys.length - 1];
  delete current[lastKey];
}

// ============================================================================
// Policy Rule Detection Helpers
// ============================================================================

/**
 * Checks if a value is a non-null object (not an array).
 */
function isObject(val: unknown): val is Record<string, unknown> {
  return val !== null && typeof val === "object" && !Array.isArray(val);
}

/**
 * Checks if an object is a policy rule object.
 * Rule objects are identified by having a 'sourceOfRule' property.
 */
function isRuleObject(val: unknown): val is IBlockableRule {
  return isObject(val) && "sourceOfRule" in val;
}

/**
 * Checks if an object has an 'attributes' property (used for array item rules).
 */
function hasAttributes(val: unknown): val is { attributes: Record<string, unknown> } {
  return isObject(val) && "attributes" in val && isObject((val as Record<string, unknown>).attributes);
}

/**
 * Checks if a rule has blocked: true.
 */
function isBlocked(rule: IBlockableRule | null | undefined): boolean {
  return rule?.blocked === true;
}

// ============================================================================
// Recursive Blocked Field Discovery
// ============================================================================

/**
 * Recursively finds all blocked fields in the policy rules object.
 * Handles nested objects and the `attributes` pattern for array items.
 *
 * @param rules The policy rules object to traverse
 * @param currentPath Current path in the traversal (for building full paths)
 * @param isInsideAttributes Whether we're currently inside an `attributes` object
 * @param arrayFieldPath The full path to the array field if we're inside attributes (e.g., "storage.pvc")
 * @returns Array of blocked field information
 */
function findBlockedFields(
  rules: Record<string, unknown> | null | undefined,
  currentPath = "",
  isInsideAttributes = false,
  arrayFieldPath?: string,
): IBlockedFieldInfo[] {
  if (!rules || !isObject(rules)) {
    return [];
  }

  const blockedFields: IBlockedFieldInfo[] = [];

  for (const [key, value] of Object.entries(rules)) {
    if (key === "instances") continue; // handled separately for locked items

    const fullPolicyPath = currentPath ? `${currentPath}.${key}` : key;

    if (isRuleObject(value)) {
      if (isBlocked(value)) {
        if (isInsideAttributes && arrayFieldPath) {
          blockedFields.push({
            policyPath: fullPolicyPath,
            modelPath: `${arrayFieldPath}[*].${key}`,
            isArrayItemAttribute: true,
            arrayFieldName: arrayFieldPath,
            attributePath: key,
          });
        } else {
          blockedFields.push({
            policyPath: fullPolicyPath,
            modelPath: fullPolicyPath,
            isArrayItemAttribute: false,
          });
        }
      }
    } else if (isObject(value)) {
      if (hasAttributes(value)) {
        const nestedBlocked = findBlockedFields(
          value.attributes as Record<string, unknown>,
          fullPolicyPath,
          true,
          fullPolicyPath,
        );
        blockedFields.push(...nestedBlocked);
      } else if (key === "attributes") {
        const nestedBlocked = findBlockedFields(value, currentPath, true, currentPath);
        blockedFields.push(...nestedBlocked);
      } else {
        const nestedBlocked = findBlockedFields(value, fullPolicyPath, isInsideAttributes, arrayFieldPath);
        blockedFields.push(...nestedBlocked);
      }
    }
  }

  return blockedFields;
}

// ============================================================================
// Policy-to-UI Field Path Mappings
// ============================================================================

/**
 * Maps policy paths to additional UI model paths that should also be cleared.
 * Handles cases where UI model field names differ from policy/backend names.
 */
const POLICY_TO_UI_PATH_MAPPINGS: Record<string, string[]> = {
  "storage.pvc": ["storage.volumes"], // volumes is UI-only, merges into pvc on submission
};

// ============================================================================
// Value Clearing Functions
// ============================================================================

/**
 * Clears an attribute from each item in an array.
 */
function clearArrayItemAttribute(specRecord: Record<string, unknown>, arrayPath: string, attributePath: string): void {
  const array = getValueByPath<Array<Record<string, unknown>>>(specRecord, arrayPath);
  if (Array.isArray(array)) {
    for (const item of array) {
      if (isObject(item)) {
        deleteValueByPath(item, attributePath);
      }
    }
  }
}

function clearBlockedField(spec: IUISuperWorkloadSpec, fieldInfo: IBlockedFieldInfo): void {
  const specRecord = spec as Record<string, unknown>;

  if (fieldInfo.isArrayItemAttribute && fieldInfo.arrayFieldName && fieldInfo.attributePath) {
    clearArrayItemAttribute(specRecord, fieldInfo.arrayFieldName, fieldInfo.attributePath);

    const additionalPaths = POLICY_TO_UI_PATH_MAPPINGS[fieldInfo.arrayFieldName];
    if (additionalPaths) {
      for (const uiPath of additionalPaths) {
        clearArrayItemAttribute(specRecord, uiPath, fieldInfo.attributePath);
      }
    }
  } else {
    deleteValueByPath(specRecord, fieldInfo.modelPath);
  }
}

// ============================================================================
// Main Functions
// ============================================================================

/**
 * Clears all blocked field values from the spec based on policy rules.
 * Called at model init time to prevent blocked values from triggering UI side effects.
 */
export function clearWorkloadBlockedFields<T extends IUISuperWorkloadSpec | null | undefined>(
  spec: T,
  policy: TPolicyWithRules | null | undefined,
): T {
  if (!spec || !policy?.rules) {
    return spec;
  }

  const blockedFields = findBlockedFields(policy.rules);

  for (const fieldInfo of blockedFields) {
    clearBlockedField(spec, fieldInfo);
  }

  return spec;
}

/**
 * Gets a list of all blocked fields from the policy.
 * Useful for debugging or UI display.
 *
 * @param policy The policy containing rules
 * @returns Array of blocked field information
 */
export function getBlockedFields(policy: TPolicyWithRules | null | undefined): IBlockedFieldInfo[] {
  if (!policy?.rules) {
    return [];
  }

  return findBlockedFields(policy.rules);
}

/**
 * Checks if a specific field path is blocked by policy.
 *
 * @param policy The policy containing rules
 * @param policyPath The path to check (e.g., "command", "compute.cpuCoreRequest")
 * @returns true if the field is blocked, false otherwise
 */
export function isFieldBlocked(policy: TPolicyWithRules | null | undefined, policyPath: string): boolean {
  const rule = getValueByPath<IBlockableRule>(policy?.rules as Record<string, unknown> | null | undefined, policyPath);
  return rule?.blocked === true;
}

/**
 * Gets just the policy paths of all blocked fields.
 *
 * @param policy The policy containing rules
 * @returns Array of policy paths that are blocked
 */
export function getBlockedFieldPaths(policy: TPolicyWithRules | null | undefined): string[] {
  return getBlockedFields(policy).map((f) => f.policyPath);
}
