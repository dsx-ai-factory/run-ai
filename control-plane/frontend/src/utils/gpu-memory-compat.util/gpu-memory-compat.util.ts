/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Existing forms still write the deprecated names. Keep one value in their models so an
// edit or deletion cannot resubmit a stale alias from a dual-field API response.
export function legacyComputeForForm<T>(source: T): T {
  if (!source || typeof source !== "object") return source;
  const result = { ...source } as Record<string, unknown>;
  if (!("gpuPortionRequest" in result) && "gpuMemoryPortionRequest" in result) {
    result.gpuPortionRequest = result.gpuMemoryPortionRequest;
  }
  if (!("gpuPortionLimit" in result) && "gpuMemoryPortionLimit" in result) {
    result.gpuPortionLimit = result.gpuMemoryPortionLimit;
  }
  delete result.gpuMemoryPortionRequest;
  delete result.gpuMemoryPortionLimit;
  return result as T;
}

export function legacyPolicyForForm<T>(source: T): T {
  if (!source || typeof source !== "object" || Array.isArray(source)) return source;
  const result = { ...source } as Record<string, unknown>;
  if ("policy" in result) result.policy = legacyPolicyForForm(result.policy);
  if ("effective" in result) result.effective = legacyPolicyForForm(result.effective);
  if ("defaults" in result) result.defaults = legacyPolicyForForm(result.defaults);
  if ("rules" in result) result.rules = legacyPolicyForForm(result.rules);
  if ("worker" in result) result.worker = legacyPolicyForForm(result.worker);
  if ("master" in result) result.master = legacyPolicyForForm(result.master);
  if ("leader" in result) result.leader = legacyPolicyForForm(result.leader);
  if ("compute" in result) result.compute = legacyComputeForForm(result.compute);
  return result as T;
}
