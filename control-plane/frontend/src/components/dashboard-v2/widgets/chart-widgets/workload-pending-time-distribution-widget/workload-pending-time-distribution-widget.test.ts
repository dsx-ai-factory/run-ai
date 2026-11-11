/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { WorkloadPendingTimeDistributionWidget } from "@/components/dashboard-v2/widgets/chart-widgets/workload-pending-time-distribution-widget";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { filterService } from "@/services/filter.service/filter.service";
import { EColumnFilterType } from "@/models/filter.model";
import { Phase, WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

vi.mock("@/services/cluster/workload.service/workload.service");

const mockRouter = { push: vi.fn() };

function mountWidget(nodePoolName?: string) {
  vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue({ values: [] } as never);
  return mount(WorkloadPendingTimeDistributionWidget, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      mocks: { $route: {}, $router: mockRouter },
      stubs: { VChart: true },
    },
    props: { clusterId: "cluster-1", nodePoolName, hasDepartments: true },
  });
}

describe("workload-pending-time-distribution-widget", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("links to the workloads page filtered by Pending status", async () => {
    const setFilterSpy = vi.spyOn(filterService, "setColumnAdvancedFilter").mockImplementation(() => undefined);
    const wrapper = mountWidget();
    await flushPromises();

    wrapper.vm.redirectToWorkloads();

    expect(setFilterSpy).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        name: WorkloadSortFilterFields.Phase,
        label: "Status",
        type: EColumnFilterType.EnumBased,
        selectedValues: [Phase.Pending],
      }),
      expect.anything(),
    );
    expect(mockRouter.push).toHaveBeenCalledWith({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
  });

  it("also filters by the selected node pool", async () => {
    const setFilterSpy = vi.spyOn(filterService, "setColumnAdvancedFilter").mockImplementation(() => undefined);
    const wrapper = mountWidget("pool-a");
    await flushPromises();

    wrapper.vm.redirectToWorkloads();

    const filterNames = setFilterSpy.mock.calls.map(([, filter]) => filter.name);
    expect(filterNames).toEqual([WorkloadSortFilterFields.Phase, WorkloadSortFilterFields.Nodepool]);
  });
});
