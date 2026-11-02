/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { useWorkloadDrawerNavigation } from "./use-workload-drawer-navigation.composable";
const { route, push, replace } = vi.hoisted(() => ({
  route: { query: {} as Record<string, string> },
  push: vi.fn(),
  replace: vi.fn(),
}));
vi.mock("vue-router", () => ({ useRoute: () => route, useRouter: () => ({ push, replace }) }));

beforeEach(() => {
  route.query = {};
  vi.clearAllMocks();
});

describe("workload activity navigation", () => {
  it("reads a direct activity link", () => {
    route.query = { openDrawer: "true", workloadId: "deleted-id", drawerTab: "activity" };
    expect(useWorkloadDrawerNavigation().getDrawerParamsFromUrl()).toEqual({
      workloadId: "deleted-id",
      tab: "activity",
      podName: "",
    });
  });
  it("rejects an unrecognized tab instead of casting arbitrary URL input", () => {
    route.query = { openDrawer: "true", workloadId: "id", drawerTab: "unknown" };
    expect(useWorkloadDrawerNavigation().getDrawerParamsFromUrl()).toBeNull();
  });
  it("creates an activity link", () => {
    useWorkloadDrawerNavigation().navigateToWorkloadDrawer("id", "activity");
    expect(push).toHaveBeenCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ workloadId: "id", drawerTab: "activity" }) }),
    );
  });
});
