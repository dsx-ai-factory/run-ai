import { describe, expect, test } from "vitest";

import { getRoundedDownTime } from "./use-generic-metric-widget-options.composable";

describe("useGenericMetricWidgetOptions time bounds", () => {
  test("floors the minimum to the previous interval boundary", () => {
    const interval = 10 * 60 * 1000;
    const timestamp = new Date("2026-07-07T12:07:00.000Z").getTime();

    expect(getRoundedDownTime(timestamp, interval)).toBe(new Date("2026-07-07T12:00:00.000Z").getTime());
  });

  test("preserves an already aligned minimum", () => {
    const interval = 10 * 60 * 1000;
    const timestamp = new Date("2026-07-07T12:10:00.000Z").getTime();

    expect(getRoundedDownTime(timestamp, interval)).toBe(timestamp);
  });

  test("rejects invalid timestamps", () => {
    expect(getRoundedDownTime("not-a-date", 10 * 60 * 1000)).toBeUndefined();
  });
});
