/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { NodepoolCreateFields, SchedulingCapabilitiesNullable } from "@/swagger-models/cluster-service-client";

// Temporary frontend gate while the new flow and its backend dependencies are developed.
export const ENABLE_NODE_POOL_CREATE_FLOW = false;
export const PROPORTIONAL_CAPABILITIES = Object.freeze({
  adjustableWeights: true,
  quota: false,
  ranks: false,
  cpuQuota: false,
} satisfies SchedulingCapabilitiesNullable);

// Keep description non-null in the form draft; the API accepts an optional description.
export interface NodePoolCreationDraft extends NodepoolCreateFields {
  description: string;
}

export const USED_LABEL_MESSAGE = "This key-value pair is already in use by another node pool.";
const LABEL_FORMAT = "(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?";
function validLabelSegment(value: string): boolean {
  return /^[A-Za-z0-9_.-]+$/.test(value) && /^[A-Za-z0-9]/.test(value) && /[A-Za-z0-9]$/.test(value);
}
export function validateName(value: string): true | string {
  if (!value) return "Enter a name";
  return (
    (value.length <= 40 && /^[a-z][a-z0-9-]*$/.test(value) && !value.endsWith("-")) ||
    "Enter a name starting with a lowercase letter and including only a-z, 0-9, -"
  );
}
export function validateLabelKey(value: string): true | string {
  if (!value) return "Select a key";
  const parts = value.split("/");
  return (
    (value.length <= 253 && parts.length <= 2 && parts.every(validLabelSegment)) ||
    `Key must fit the following regular expression ${LABEL_FORMAT}`
  );
}
export function validateLabelValue(value: string): true | string {
  if (!value) return "Select a value";
  return (
    (value.length <= 63 && validLabelSegment(value)) || `Value must fit the following regular expression ${LABEL_FORMAT}`
  );
}
export function labelOptions(values: string[], search: string, unavailable: string[], kind = "value") {
  const options = [...new Set([...values, ...unavailable])]
    .filter((value) => value.toLowerCase().includes(search.toLowerCase()))
    .map((value) => ({ label: value, value, disable: unavailable.includes(value) }));
  if (search && ![...values, ...unavailable].includes(search)) {
    options.push({ label: `${search} (new ${kind})`, value: search, disable: false });
  }
  return options;
}
export function isProportional(capabilities?: SchedulingCapabilitiesNullable): boolean {
  return (
    capabilities?.adjustableWeights === true &&
    capabilities.quota === false &&
    capabilities.ranks === false &&
    capabilities.cpuQuota === false
  );
}
