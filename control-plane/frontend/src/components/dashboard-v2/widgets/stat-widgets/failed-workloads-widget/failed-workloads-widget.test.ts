/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { FailedWorkloadsWidget } from "@/components/dashboard-v2/widgets/stat-widgets/failed-workloads-widget/index";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

vi.mock("@/services/cluster/workload.service/workload.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof FailedWorkloadsWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "34234-322-sdfs323";

describe("FailedWorkloadsWidget component", () => {
  const createComponent = (props?: { nodePoolName: string }) => {
    wrapper = mount(FailedWorkloadsWidget, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props: {
        ...props,
        clusterId,
      },
    });

    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  beforeEach(async () => {
    vi.clearAllMocks();
  });

  const getStatData = () => {
    return wrapper.find(".stat-data").text();
  };

  const clickOnGridLink = async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
  };

  it("should render FailedWorkloadsWidget component", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(144));
    createComponent();
    await flushPromises();
    expect(getStatData()).toBe("144");
  });

  it("should render FailedWorkloadsWidget component with filter", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent({ nodePoolName: "Node pool" });
    await flushPromises();
    expect(getStatData()).toBe("16");
  });

  it("should link to workloads table when clicked", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent();
    await flushPromises();
    await clickOnGridLink();
    expect(mockRouter.push).toHaveBeenCalledWith({ name: "workload-index" });
  });

  it("should trigger get telemetry data on filter change", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent();
    await flushPromises();
    await wrapper.setProps({ nodePoolName: "Node pool" });
    expect(workloadService.getWorkloadsCount).toHaveBeenCalledTimes(2);
  });

  it("should display loading skeleton when loading", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
    await flushPromises();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
  });

  it("should display error message when telemetry fetch fails", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(workloadService.getWorkloadsCount).mockRejectedValueOnce({ response: { status: 500 } });
    createComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toEqual(true);
    expect(wrapper.find(".error-section").text()).eql("Data is not available right now");
    errorSpy.mockRestore();
  });
});
