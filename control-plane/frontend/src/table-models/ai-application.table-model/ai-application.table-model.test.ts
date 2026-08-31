/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, test } from "vitest";
import { allAIApplicationColumnsMap } from "./ai-application.table-model";
import type { IApplicationTable } from "@/models/ai-application.model";
import type { IStatusColOptions } from "@/models/table.model";

describe("ai-application table model - status column", () => {
  const statusFormat = allAIApplicationColumnsMap.status.format as (app: IApplicationTable) => IStatusColOptions;

  const makeApp = (overrides: Partial<IApplicationTable> = {}): IApplicationTable => ({
    id: "app-1",
    name: "test-app",
    ...overrides,
  });

  test("Deployed phase uses success color", () => {
    const result = statusFormat(makeApp({ phase: "Deployed" }));
    expect(result.status).toBe("Deployed");
    expect(result.color).toBe("success");
  });

  test("Failed phase uses negative color", () => {
    const result = statusFormat(makeApp({ phase: "Failed" }));
    expect(result.status).toBe("Failed");
    expect(result.color).toBe("negative");
  });

  test("unknown phase has no color", () => {
    const result = statusFormat(makeApp({ phase: "Pending" }));
    expect(result.color).toBeUndefined();
  });

  test("falls back to actionStatus when phase is absent", () => {
    const result = statusFormat(makeApp({ phase: undefined, actionStatus: "Failed" }));
    expect(result.status).toBe("Failed");
    expect(result.color).toBe("negative");
  });

  test("lowercase phase from API is normalized before color lookup", () => {
    const result = statusFormat(makeApp({ phase: "deployed" }));
    expect(result.status).toBe("Deployed");
    expect(result.color).toBe("success");
  });

  test("lowercase failed phase from API is normalized before color lookup", () => {
    const result = statusFormat(makeApp({ phase: "failed" }));
    expect(result.status).toBe("Failed");
    expect(result.color).toBe("negative");
  });

  test("falls back to Unknown when phase and actionStatus are absent", () => {
    const result = statusFormat(makeApp({ phase: undefined, actionStatus: undefined }));
    expect(result.status).toBe("Unknown");
    expect(result.color).toBeUndefined();
  });

  test("populates statusUpdatedTimeInMs from updatedAt", () => {
    const updatedAt = "2026-06-10T12:00:00Z";
    const result = statusFormat(makeApp({ phase: "Deployed", updatedAt }));
    expect(result.statusUpdatedTimeInMs).toEqual(new Date(updatedAt));
  });

  test("statusUpdatedTimeInMs is undefined when updatedAt is absent", () => {
    const result = statusFormat(makeApp({ phase: "Deployed", updatedAt: undefined }));
    expect(result.statusUpdatedTimeInMs).toBeUndefined();
  });

  test("tooltipText is set from actionMessage", () => {
    const result = statusFormat(makeApp({ phase: "Failed", actionMessage: "Deployment failed: OOMKilled" }));
    expect(result.tooltipText).toBe("Deployment failed: OOMKilled");
  });

  test("tooltipText is undefined when actionMessage is absent", () => {
    const result = statusFormat(makeApp({ phase: "Deployed", actionMessage: null }));
    expect(result.tooltipText).toBeUndefined();
  });

  test("displayAnimation is always false", () => {
    const result = statusFormat(makeApp({ phase: "Deployed" }));
    expect(result.displayAnimation).toBe(false);
  });
});
