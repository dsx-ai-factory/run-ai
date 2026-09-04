/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar, type ValidationRule } from "quasar";
import type { VueWrapper } from "@vue/test-utils";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { useClusterStore } from "@/stores/cluster.store";
import { RunaiClusterDropdown } from "@/components/common/runai-cluster-selection/runai-cluster-dropdown/";

import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { ERunaiClusterDropdownEvents } from "./runai-cluster-dropdown.vue";

vi.mock("@/services/cluster/workload.service/workload.service");

interface IClusterDropdownProps {
  clusterId: string;
  disabled?: boolean;
  rules?: ValidationRule[];
  disabledTooltipText?: string;
  disabledClusterOptions?: Record<string, string>;
}
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
let wrapper: VueWrapper<InstanceType<typeof RunaiClusterDropdown>>;

const defaultClusters: Array<DisplayedCluster> = [
  {
    uuid: "1",
    name: "Cluster One",
    tenantId: 0,
    createdAt: "",
  },
  {
    uuid: "2",
    name: "Cluster Two",
    tenantId: 0,
    createdAt: "",
  },
];

describe("runai-cluster-dropdown", () => {
  const factory = (props: IClusterDropdownProps, clusters: Array<DisplayedCluster> = defaultClusters) => {
    const clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ clusters });

    wrapper = mount(RunaiClusterDropdown, {
      global: {
        plugins: [Quasar, pinia],
      },
      props,
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders cluster options correctly", async () => {
    factory({ clusterId: "1" });
    await flushPromises();
    expect(wrapper.vm.clusterOptions).toEqual([
      { label: "Cluster One", value: "1", disable: false, disabledTooltip: undefined },
      { label: "Cluster Two", value: "2", disable: false, disabledTooltip: undefined },
    ]);
  });

  it("displays tooltip when disabled and tooltip text is provided", () => {
    factory({ clusterId: "1", disabled: true, disabledTooltipText: "Tooltip text" });
    expect(wrapper.vm.displayTooltip).toBe(true);
  });

  it("does not display tooltip when not disabled", () => {
    factory({ clusterId: "1", disabled: false, disabledTooltipText: "Tooltip text" });
    expect(wrapper.vm.displayTooltip).toBe(false);
  });

  it("emits update:cluster-id event when a cluster is selected", async () => {
    factory({ clusterId: "" });
    wrapper.vm.updateSelectedCluster({ label: "Cluster Two", value: "2" });
    expect(wrapper.emitted(ERunaiClusterDropdownEvents.UpdateClusterId)).toBeTruthy();
    expect(wrapper.emitted(ERunaiClusterDropdownEvents.UpdateClusterId)?.[0]).toEqual(["2", { disabled: false }]);
  });

  it("should disable options correctly", async () => {
    const disableText = "The cluster is not connected";
    factory({ clusterId: "", disabledClusterOptions: { [defaultClusters[1].uuid]: disableText } });

    expect(wrapper.vm.clusterOptions[1].disable).toBe(true);
    expect(wrapper.vm.clusterOptions[1].disabledTooltip).toBe(disableText);
  });
});
