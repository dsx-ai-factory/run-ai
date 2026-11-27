/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { isProportional, PROPORTIONAL_CAPABILITIES } from "@/models/node-pool.model";

describe("node pool creation fields", () => {
  it("initializes independent drafts without sending omitted version-gated defaults", () => {
    const first = nodePoolUtil.createNodePoolDraft("cluster");
    const second = nodePoolUtil.createNodePoolDraft("cluster");
    expect(first.gpuNetworkAccelerationDetection).toBeUndefined();
    expect(first.schedulingConfiguration?.numaAwareScheduling).toBeUndefined();
    expect(first.schedulingConfiguration?.gpuPodsPlacementStrategy?.device).toBeUndefined();
    if (first.schedulingCapabilities) first.schedulingCapabilities.quota = false;
    expect(second.schedulingCapabilities).toMatchObject({ quota: true, ranks: true, adjustableWeights: false });
  });
  it("matches the backend UTF-8 description limit", () => {
    expect(nodePoolUtil.validateDescription("")).toBe(true);
    expect(nodePoolUtil.validateDescription("a".repeat(250))).toBe(true);
    expect(nodePoolUtil.validateDescription("א".repeat(125))).toBe(true);
    expect(nodePoolUtil.validateDescription("א".repeat(126))).not.toBe(true);
    expect(nodePoolUtil.validateDescription("😀".repeat(62) + "ab")).toBe(true);
    expect(nodePoolUtil.validateDescription("😀".repeat(63))).not.toBe(true);
  });
  it("validates required fields and name format", () => {
    expect(nodePoolUtil.validateName("")).toBe("Enter a name");
    expect(nodePoolUtil.validateName("1pool")).not.toBe(true);
    expect(nodePoolUtil.validateName("Pool")).not.toBe(true);
    expect(nodePoolUtil.validateName("gpu-pool-1")).toBe(true);
    expect(nodePoolUtil.validateName("a".repeat(41))).not.toBe(true);
  });
  it("accepts qualified Kubernetes keys and rejects malformed keys", () => {
    expect(nodePoolUtil.validateLabelKey("nvidia.com/gpu.product")).toBe(true);
    expect(nodePoolUtil.validateLabelKey("node-type")).toBe(true);
    expect(nodePoolUtil.validateLabelKey("/node-type")).toBe(true);
    expect(nodePoolUtil.validateLabelKey("a/b/c")).not.toBe(true);
    expect(nodePoolUtil.validateLabelKey("a/-b")).not.toBe(true);
    expect(nodePoolUtil.validateLabelKey("")).toBe("Enter a key");
    expect(nodePoolUtil.validateLabelKey("test@run.ai")).not.toBe(true);
    expect(nodePoolUtil.validateLabelKey("a".repeat(254))).not.toBe(true);
  });
  it("requires a value and accepts dots, underscores and hyphens", () => {
    expect(nodePoolUtil.validateLabelValue("")).toBe("Select a value");
    expect(nodePoolUtil.validateLabelValue("H100_SXM-80GB")).toBe(true);
    expect(nodePoolUtil.validateLabelValue("-H100")).not.toBe(true);
    expect(nodePoolUtil.validateLabelValue("a".repeat(64))).not.toBe(true);
  });
  it("keeps unavailable options disabled even when their exact value is typed", () => {
    const options = nodePoolUtil.labelOptions(["T4", "T5", "T4"], "T5", ["T5"]);
    expect(options).toHaveLength(1);
    expect(options[0]).toMatchObject({ value: "T5", disable: true });
    expect(nodePoolUtil.labelOptions(["T4"], "H100", [])).toEqual([
      { label: "H100 (new value)", value: "H100", disable: false },
    ]);
  });
  it("recognizes Proportional with limits enabled or disabled", () => {
    expect(isProportional({ ...PROPORTIONAL_CAPABILITIES, limits: true })).toBe(true);
    expect(isProportional(PROPORTIONAL_CAPABILITIES)).toBe(true);
  });
  it("defaults new Proportional pools to weights only", () => {
    expect(PROPORTIONAL_CAPABILITIES).toEqual({
      adjustableWeights: true,
      quota: false,
      ranks: false,
      cpuQuota: false,
      limits: false,
    });
  });
});
