/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { PREDEFINED_HOURS, PREDEFINED_MINUTES } from "@/models/date.model";
import { getOverviewShortRangeAxisInterval, ONE_MINUTE_IN_MS, TEN_MINUTES_IN_MS } from "./overview-time-axis.util";

describe("overview time axis", () => {
  it("uses one-minute ticks for the Last 15 minutes preset", () => {
    expect(getOverviewShortRangeAxisInterval(PREDEFINED_MINUTES.LAST_15_MINUTES)).toBe(ONE_MINUTE_IN_MS);
  });

  it("keeps the existing ten-minute ticks for other time ranges", () => {
    expect(getOverviewShortRangeAxisInterval(PREDEFINED_HOURS.LAST_3_HOURS)).toBe(TEN_MINUTES_IN_MS);
  });
});
