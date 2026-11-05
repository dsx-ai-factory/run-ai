/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import {
  labelOptions,
  createNodePoolDraft,
  validateName,
  validateDescription,
  validateLabelKey,
  validateLabelValue,
  PROPORTIONAL_CAPABILITIES,
} from "./node-pool-create-flow.model";

describe("node pool creation fields", () => {
  it("initializes independent drafts without sending omitted version-gated defaults", () => {
    const first = createNodePoolDraft("cluster");
    const second = createNodePoolDraft("cluster");
    expect(first.gpuNetworkAccelerationDetection).toBeUndefined();
    expect(first.schedulingConfiguration?.numaAwareScheduling).toBeUndefined();
    expect(first.schedulingConfiguration?.gpuPodsPlacementStrategy?.device).toBeUndefined();
    if (first.schedulingCapabilities) first.schedulingCapabilities.quota = true;
    expect(second.schedulingCapabilities?.quota).toBe(false);
  });
  it("matches the backend UTF-8 description limit", () => {
    expect(validateDescription("")).toBe(true);
    expect(validateDescription("a".repeat(250))).toBe(true);
    expect(validateDescription("א".repeat(125))).toBe(true);
    expect(validateDescription("א".repeat(126))).not.toBe(true);
    expect(validateDescription("😀".repeat(62) + "ab")).toBe(true);
    expect(validateDescription("😀".repeat(63))).not.toBe(true);
  });
  it("validates required fields and name format", () => {
    expect(validateName("")).toBe("Enter a name");
    expect(validateName("1pool")).not.toBe(true);
    expect(validateName("Pool")).not.toBe(true);
    expect(validateName("gpu-pool-1")).toBe(true);
    expect(validateName("a".repeat(41))).not.toBe(true);
  });
  it("accepts qualified Kubernetes keys and rejects malformed keys", () => {
    expect(validateLabelKey("nvidia.com/gpu.product")).toBe(true);
    expect(validateLabelKey("node-type")).toBe(true);
    expect(validateLabelKey("")).toBe("Select a key");
    expect(validateLabelKey("test@run.ai")).not.toBe(true);
    expect(validateLabelKey("a".repeat(254))).not.toBe(true);
  });
  it("requires a value and accepts dots, underscores and hyphens", () => {
    expect(validateLabelValue("")).toBe("Select a value");
    expect(validateLabelValue("H100_SXM-80GB")).toBe(true);
    expect(validateLabelValue("-H100")).not.toBe(true);
    expect(validateLabelValue("a".repeat(64))).not.toBe(true);
  });
  it("keeps unavailable options disabled even when their exact value is typed", () => {
    const options = labelOptions(["T4", "T5", "T4"], "T5", ["T5"]);
    expect(options).toHaveLength(1);
    expect(options[0]).toMatchObject({ value: "T5", disable: true });
    expect(labelOptions(["T4"], "H100", [])).toEqual([{ label: "H100 (new value)", value: "H100", disable: false }]);
  });
  it("defines Proportional as weights only", () => {
    expect(PROPORTIONAL_CAPABILITIES).toEqual({ adjustableWeights: true, quota: false, ranks: false, cpuQuota: false });
  });
});
