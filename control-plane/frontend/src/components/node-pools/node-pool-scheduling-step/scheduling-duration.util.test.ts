/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import {
  schedulingDurationSeconds,
  schedulingDurationToFields,
  serializeSchedulingDuration,
} from "./scheduling-duration.util";

describe("saved scheduling durations", () => {
  it.each([
    ["0.5h", { Days: 0, Hours: 0, Minutes: 30, Seconds: 0 }],
    ["1.5d", { Days: 1, Hours: 12, Minutes: 0, Seconds: 0 }],
    ["2w", { Days: 14, Hours: 0, Minutes: 0, Seconds: 0 }],
    ["500ms", { Days: 0, Hours: 0, Minutes: 0, Seconds: 0.5 }],
    ["1h30m0.5s", { Days: 0, Hours: 1, Minutes: 30, Seconds: 0.5 }],
  ])("loads API-supported runtime %s without changing its duration", (value, expected) => {
    expect(schedulingDurationToFields(value)).toEqual(expected);
  });
  it("keeps an API-supported fractional minute visible when the editor has no seconds field", () => {
    expect(schedulingDurationToFields("0.5m", false)).toEqual({ Days: 0, Hours: 0, Minutes: 0.5, Seconds: 0 });
  });

  it.each(["1hgarbage", "1.2.3s", "1.h"])(
    "rejects invalid duration %s instead of partially interpreting it",
    (value) => {
      expect(() => schedulingDurationSeconds(value)).toThrow(/valid nonnegative duration/);
    },
  );
  it("serializes fractional seconds without exponent notation rejected by the API", () => {
    expect(serializeSchedulingDuration({ Seconds: 0.0000001 })).toBe("0.0000001s");
  });
});
