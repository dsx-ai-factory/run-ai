/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import type { VueWrapper } from "@vue/test-utils";

// Components
import { NodePoolsModal } from "@/components/node-pools/node-pools-modal";
import { flushPromises, mount } from "@vue/test-utils";
import { unitTestUtil } from "@/utils/unit-test.util";

// Mock data and services
import type { INodePoolModalOptions } from "@/models/project.model";
import { EQuotaEntity } from "@/models/resource.model";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });

const mockNodePools = [
  {
    nodePool: { name: "default", id: "1" },
    allocatedGpus: 2,
    allocatedCpu: 4,
    allocatedMemory: 8,
  },
  {
    nodePool: { name: "high-performance", id: "2" },
    allocatedGpus: 4,
    allocatedCpu: 8,
    allocatedMemory: 16,
  },
];

const modalOptions: INodePoolModalOptions = {
  resources: [
    {
      nodePool: { id: "100001", name: "node-pool-2" },
      gpu: { deserved: 150, limit: -1, overQuotaWeight: 2 },
      priority: "Normal",
    },
    {
      nodePool: { id: "100000", name: "default" },
      gpu: { deserved: 150, limit: -1, overQuotaWeight: 2 },
      priority: "Normal",
    },
  ],
  entityType: EQuotaEntity.department,
  header: "Node Pools Associated with department default",
  nodePoolQuotaStatuses: [
    { allocated: {}, allocatedNonPreemptible: {}, requested: {}, nodePoolName: "default", nodePoolId: "100000" },
    {
      allocated: {},
      allocatedNonPreemptible: {},
      requested: {},
      nodePoolName: "node-pool-2",
      nodePoolId: "100001",
    },
  ],
  loading: false,
  defaultNodePools: ["node-pool-2", "default"],
};

describe("node-pools-modal component", () => {
  let wrapper: VueWrapper<InstanceType<typeof NodePoolsModal>>;

  const createComponent = (props: { modalOptions: INodePoolModalOptions }): void => {
    wrapper = mount(NodePoolsModal, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
      },
      props: props,
    });
    wrapper.vm.$q.notify = vi.fn(); // Disable notify to prevent test failures
  };

  beforeEach(() => {
    createComponent({
      modalOptions,
    });
  });

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  test("Modal loads correctly", () => {
    expect(wrapper.findComponent(NodePoolsModal).exists()).toBe(true);
  });

  test("Table renders with correct data", () => {
    const modalEl = unitTestUtil.getModalElement();
    expect(modalEl.find(".runai-table").exists()).toBe(true);
    expect(modalEl.findAll(".runai-table-row").length).toBe(mockNodePools.length);
  });

  test("Search functionality works", async () => {
    const modalEl = unitTestUtil.getModalElement();
    await modalEl.find("[aid=runai-list-modal-search]").setValue("test");
    await flushPromises();
    expect(modalEl.findAll(".runai-table-row").length).toBe(0);
  });

  test("Close button emits close event", async () => {
    const modalEl = unitTestUtil.getModalElement();
    await modalEl.find("[aid=close-node-pools-modal-btn]").trigger("click");
    expect(wrapper.emitted("close")).toBeTruthy();
  });

  test("Filtering renders only prioritized node pools", async () => {
    const modalEl = unitTestUtil.getModalElement();
    // Simulate setting a filter that prioritizes certain node pools
    wrapper.vm.tableFilter = {
      searchTerm: "",
      columnFilters: [
        {
          field: "priorityOrder",
          name: "Priority Order",
          term: "1",
        },
      ],
    };
    await flushPromises();

    // Verify that only the prioritized node pools are rendered
    const visibleRows = modalEl.findAll(".runai-table-row");
    expect(visibleRows.length).toBe(2);
    expect(visibleRows.at(0)?.text()).toContain("node-pool-2");
  });
});
