/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useWorkloadStore } from "./workload.store";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import type { GetWorkloadActionsResponse, WorkloadActionRecord } from "@/swagger-models/workloads-client";

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: { getWorkloadActions: vi.fn() },
}));

const row = (id: string): WorkloadActionRecord => ({
  id,
  workloadId: "w1",
  workloadName: "training",
  subject: "person@example.com",
  subjectType: "user",
  action: "create",
  acceptedAt: "2026-11-02T08:00:00Z",
});
const deferred = () => {
  let resolve!: (value: GetWorkloadActionsResponse) => void;
  const promise = new Promise<GetWorkloadActionsResponse>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};

beforeEach(() => {
  setActivePinia(createPinia());
  vi.clearAllMocks();
});

describe("workload activity state", () => {
  it("keeps only the requested page and the server next offset", async () => {
    const store = useWorkloadStore();
    vi.mocked(workloadService.getWorkloadActions)
      .mockResolvedValueOnce({ actions: [row("first")], next: 50 })
      .mockResolvedValueOnce({ actions: [row("second")] });
    await store.loadWorkloadActions("w1");
    expect(store.workloadActionRecords).toEqual([row("first")]);
    expect(store.workloadActionsNext).toBe(50);
    await store.loadWorkloadActions("w1", 50, 50);
    expect(store.workloadActionRecords).toEqual([row("second")]);
    expect(store.workloadActionsNext).toBeNull();
  });

  it("retains rows and next while refresh is pending or fails", async () => {
    const store = useWorkloadStore();
    vi.mocked(workloadService.getWorkloadActions).mockResolvedValueOnce({ actions: [row("first")], next: 50 });
    await store.loadWorkloadActions("w1");
    const existing = store.workloadActionRecords;
    const pending = deferred();
    vi.mocked(workloadService.getWorkloadActions).mockReturnValueOnce(pending.promise);
    const refresh = store.loadWorkloadActions("w1");
    expect(store.workloadActionRecords).toBe(existing);
    pending.resolve({ actions: [row("updated")], next: 50 });
    await refresh;
    vi.mocked(workloadService.getWorkloadActions).mockRejectedValueOnce(new Error("offline"));
    await expect(store.loadWorkloadActions("w1")).rejects.toThrow("offline");
    expect(store.workloadActionRecords).toEqual([row("updated")]);
    expect(store.workloadActionsNext).toBe(50);
  });

  it("ignores an old workload response arriving after a new workload", async () => {
    const store = useWorkloadStore();
    const old = deferred();
    vi.mocked(workloadService.getWorkloadActions)
      .mockReturnValueOnce(old.promise)
      .mockResolvedValueOnce({ actions: [row("new")] });
    const previous = store.loadWorkloadActions("old");
    await store.loadWorkloadActions("new");
    old.resolve({ actions: [row("old")], next: 50 });
    await previous;
    expect(store.workloadActionRecords).toEqual([row("new")]);
    expect(store.workloadActionsNext).toBeNull();
  });

  it("invalidates in-flight requests when activity is cleared", async () => {
    const store = useWorkloadStore();
    const pending = deferred();
    vi.mocked(workloadService.getWorkloadActions).mockReturnValueOnce(pending.promise);
    const load = store.loadWorkloadActions("w1");
    store.clearWorkloadActions();
    pending.resolve({ actions: [row("old")], next: 50 });
    await load;
    expect(store.workloadActionRecords).toEqual([]);
    expect(store.workloadActionsNext).toBeNull();
  });

  it("treats zero next as the last page", async () => {
    const store = useWorkloadStore();
    vi.mocked(workloadService.getWorkloadActions).mockResolvedValueOnce({ actions: [row("first")], next: 0 });
    await store.loadWorkloadActions("w1");
    expect(store.workloadActionsNext).toBeNull();
  });
});
