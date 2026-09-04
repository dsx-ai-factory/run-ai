/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { DeptAllocationRatioWidget } from "./index";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { RankingMetricType, RankingOrderType } from "@/swagger-models/org-unit-service-client";

vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "cluster-123";

const mockRanking = [
  { id: "1", name: "dept-a", allocationValue: 0.25 },
  { id: "2", name: "dept-b", allocationValue: 0.5 },
];

describe("DeptAllocationRatioWidget", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeptAllocationRatioWidget>>;

  const createComponent = () => {
    wrapper = mount(DeptAllocationRatioWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId, hasDepartments: true },
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fetches departments ranking with gpuAllocationRatio on mount", async () => {
    vi.mocked(orgUnitService.getDepartmentsRanking).mockResolvedValue(mockRanking as never);
    createComponent();
    await flushPromises();

    expect(orgUnitService.getDepartmentsRanking).toHaveBeenCalledWith(
      RankingMetricType.GpuAllocationRatio,
      RankingOrderType.Bottom,
      clusterId,
      undefined,
    );
  });

  it("shows empty state when no data", async () => {
    vi.mocked(orgUnitService.getDepartmentsRanking).mockResolvedValue([]);
    createComponent();
    await flushPromises();

    expect(wrapper.find('[aid="empty-widget-message"]').exists()).toBe(true);
  });

  it("shows error state on failure", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(orgUnitService.getDepartmentsRanking).mockRejectedValue({ response: { status: 500 } });
    createComponent();
    await flushPromises();

    expect(wrapper.find(".error-section").exists()).toBe(true);
    errorSpy.mockRestore();
  });
});
