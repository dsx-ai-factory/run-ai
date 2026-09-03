/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { scaleDownUtil } from "./scale-down.util";
import { SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE } from "./scale-down-section.model";
import type { AutoScaling } from "@/swagger-models/workloads-client";
import type { AutoScalingRules } from "@/swagger-models/assets-service-client";
import { describe, expect, it } from "vitest";

describe("getScaleToZeroRetentionSecondsDefaultValue", () => {
  it("returns autoscalingDefaults.scaleToZeroRetentionSeconds if set", () => {
    const autoscalingDefaults: AutoScaling = { scaleToZeroRetentionSeconds: 42 };
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(undefined, autoscalingDefaults, undefined);
    expect(result).toBe(42);
  });

  it("returns minRule if minRule > default", () => {
    const autoscalingRules: AutoScalingRules = {
      scaleToZeroRetentionSeconds: { min: (SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE + 5) * 60 },
    };
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(undefined, undefined, autoscalingRules);
    expect(result).toBe((SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE + 5) * 60);
  });

  it("returns default if minRule <= default", () => {
    const autoscalingRules: AutoScalingRules = {
      scaleToZeroRetentionSeconds: { min: (SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE - 1) * 60 },
    };
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(undefined, undefined, autoscalingRules);
    expect(result).toBe(SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE * 60);
  });

  it("returns default if no autoscalingDefaults or autoscalingRules", () => {
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(undefined, undefined, undefined);
    expect(result).toBe(SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE * 60);
  });

  it("returns default if autoscalingDefaults.scaleToZeroRetentionSeconds is 0", () => {
    const autoscalingDefaults: AutoScaling = { scaleToZeroRetentionSeconds: 0 };
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(undefined, autoscalingDefaults, undefined);
    expect(result).toBe(SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE * 60);
  });

  it("returns default if autoscalingRules.scaleToZeroRetentionSeconds is undefined", () => {
    const autoscalingRules: AutoScalingRules = {};
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(undefined, undefined, autoscalingRules);
    expect(result).toBe(SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE * 60);
  });

  it("returns scaleToZeroRetentionSeconds if set (positive value)", () => {
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(
      123,
      { scaleToZeroRetentionSeconds: 456 },
      { scaleToZeroRetentionSeconds: { min: 789 } },
    );
    expect(result).toBe(123);
  });

  it("returns default if scaleToZeroRetentionSeconds is 0 and no other values", () => {
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(0, undefined, undefined);
    expect(result).toBe(SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE * 60);
  });

  it("returns autoscalingDefaults if scaleToZeroRetentionSeconds is undefined and autoscalingDefaults is set", () => {
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(
      undefined,
      { scaleToZeroRetentionSeconds: 321 },
      undefined,
    );
    expect(result).toBe(321);
  });

  it("returns minRule if scaleToZeroRetentionSeconds and autoscalingDefaults are undefined, but minRule is set and greater than default", () => {
    const minRule = (SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE + 10) * 60;
    const result = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(undefined, undefined, {
      scaleToZeroRetentionSeconds: { min: minRule },
    });
    expect(result).toBe(minRule);
  });
});
