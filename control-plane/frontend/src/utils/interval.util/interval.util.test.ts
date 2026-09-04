/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { EIntervalLabels } from "@/models/interval.model";
import type { intervalUtil as IntervalUtilExport } from "./interval.util";

let intervalUtil: typeof IntervalUtilExport;
let removeModuleListeners: () => void;

beforeEach(async () => {
  vi.useFakeTimers();
  vi.resetModules();

  // Capture the offline/online listeners the module registers at import time so we
  // can remove them in afterEach and prevent stacking across re-imports.
  const captured: [string, EventListenerOrEventListenerObject][] = [];
  const origAddEventListener = window.addEventListener.bind(window);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (window as any).addEventListener = (
    type: string,
    handler: EventListenerOrEventListenerObject,
    options?: boolean | AddEventListenerOptions,
  ) => {
    captured.push([type, handler]);
    origAddEventListener(type, handler, options);
  };

  ({ intervalUtil } = await import("./interval.util"));

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (window as any).addEventListener = origAddEventListener;
  removeModuleListeners = () =>
    captured.forEach(([type, handler]) => window.removeEventListener(type, handler as EventListener));
});

afterEach(() => {
  removeModuleListeners?.();
  vi.useRealTimers();
});

const LABEL = EIntervalLabels.CheckClusterStatus;

describe("intervalUtil.startInterval / stopInterval", () => {
  it("fires callback on interval", () => {
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);
    vi.runOnlyPendingTimers();
    expect(cb).toHaveBeenCalledTimes(1);
  });

  it("stopInterval prevents future callbacks", () => {
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);
    intervalUtil.stopInterval(LABEL);
    vi.runOnlyPendingTimers();
    expect(cb).not.toHaveBeenCalled();
  });

  it("restarting an interval clears the old one", () => {
    const cb1 = vi.fn();
    const cb2 = vi.fn();
    intervalUtil.startInterval(LABEL, cb1);
    intervalUtil.startInterval(LABEL, cb2);
    vi.runOnlyPendingTimers();
    expect(cb1).not.toHaveBeenCalled();
    expect(cb2).toHaveBeenCalledTimes(1);
  });
});

describe("offline / online event handling", () => {
  it("pauses all intervals when offline event fires", () => {
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);

    window.dispatchEvent(new Event("offline"));
    vi.runOnlyPendingTimers();

    expect(cb).not.toHaveBeenCalled();
  });

  it("resumes intervals when online event fires after offline", () => {
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);

    window.dispatchEvent(new Event("offline"));
    vi.runOnlyPendingTimers();
    expect(cb).not.toHaveBeenCalled();

    window.dispatchEvent(new Event("online"));
    vi.runOnlyPendingTimers();
    expect(cb).toHaveBeenCalledTimes(1);
  });

  it("does not resume a stopped interval after online event", () => {
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);
    intervalUtil.stopInterval(LABEL);

    window.dispatchEvent(new Event("offline"));
    window.dispatchEvent(new Event("online"));
    vi.runOnlyPendingTimers();

    expect(cb).not.toHaveBeenCalled();
  });

  it("duplicate offline events are idempotent — interval fires exactly once after single online", () => {
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);

    window.dispatchEvent(new Event("offline"));
    window.dispatchEvent(new Event("offline")); // second offline is a no-op
    window.dispatchEvent(new Event("online"));
    vi.runOnlyPendingTimers();

    expect(cb).toHaveBeenCalledTimes(1);
  });

  it("online without prior offline does not double-schedule", () => {
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);

    window.dispatchEvent(new Event("online")); // no prior offline — should be no-op
    vi.runOnlyPendingTimers();

    expect(cb).toHaveBeenCalledTimes(1); // only the original timer fires, not two
  });

  it("startInterval while paused defers the timer until online", () => {
    window.dispatchEvent(new Event("offline"));
    const cb = vi.fn();
    intervalUtil.startInterval(LABEL, cb);
    vi.runOnlyPendingTimers();
    expect(cb).not.toHaveBeenCalled();

    window.dispatchEvent(new Event("online"));
    vi.runOnlyPendingTimers();
    expect(cb).toHaveBeenCalledTimes(1);
  });
});
