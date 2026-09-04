/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import WorkloadTopologySection from "./workload-topology-section.vue";
import type { WorkloadTopology } from "@/swagger-models/workloads-client";
function createWrapper(topology: WorkloadTopology, levels: string[] = []) {
  return mount(WorkloadTopologySection, {
    global: {
      plugins: [Quasar],
    },
    props: {
      topology,
      levels,
    },
  });
}

describe("WorkloadTopologySection", () => {
  const baseLevels = [
    "network.topology.nvidia.com/region",
    "network.topology.nvidia.com/zone",
    "network.topology.nvidia.com/rack",
    "kubernetes.io/hostname",
  ];

  it("should display topology name", () => {
    const wrapper = createWrapper({ name: "my-gb200-topology" });
    expect(wrapper.text()).toContain("Topology name:");
    expect(wrapper.text()).toContain("my-gb200-topology");
  });

  it("should display actual placement when available", () => {
    const wrapper = createWrapper({
      name: "my-topology",
      actualPlacement: { level: "network.topology.nvidia.com/zone", value: "us-central" },
    });
    expect(wrapper.text()).toContain("Actual placement:");
    expect(wrapper.text()).toContain("network.topology.nvidia.com/zone");
    expect(wrapper.text()).toContain("[us-central]");
  });

  it("should not display actual placement when not available", () => {
    const wrapper = createWrapper({ name: "my-topology" });
    expect(wrapper.text()).not.toContain("Actual placement:");
  });

  it("should display topology levels", () => {
    const wrapper = createWrapper({ name: "my-topology" }, baseLevels);
    expect(wrapper.text()).toContain("Topology levels:");
    baseLevels.forEach((level) => {
      expect(wrapper.text()).toContain(level);
    });
  });

  it("should not display topology levels section when no levels provided", () => {
    const wrapper = createWrapper({ name: "my-topology" }, []);
    expect(wrapper.text()).not.toContain("Topology levels:");
  });

  it("should show Required badge on the correct level", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        requiredPlacement: "network.topology.nvidia.com/region",
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const regionRow = levelRows[0];
    expect(regionRow.text()).toContain("Required");

    const zoneRow = levelRows[1];
    expect(zoneRow.text()).not.toContain("Required");
  });

  it("should show Preferred badge on the correct level", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        preferredPlacement: "network.topology.nvidia.com/zone",
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const zoneRow = levelRows[1];
    expect(zoneRow.text()).toContain("Preferred");

    const regionRow = levelRows[0];
    expect(regionRow.text()).not.toContain("Preferred");
  });

  it("should show Actual badge on the correct level", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        actualPlacement: { level: "network.topology.nvidia.com/region", value: "us-east" },
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const regionRow = levelRows[0];
    expect(regionRow.text()).toContain("Actual");

    const zoneRow = levelRows[1];
    expect(zoneRow.text()).not.toContain("Actual");
  });

  it("should show green check icon when required constraint is met", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        requiredPlacement: "network.topology.nvidia.com/region",
        actualPlacement: { level: "network.topology.nvidia.com/region", value: "us-east" },
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const regionRow = levelRows[0];
    const icon = regionRow.find(".constraint-icon .q-icon");
    expect(icon.exists()).toBe(true);
    expect(icon.attributes("class")).toContain("constraint-icon--success");
  });

  it("should show warning icon when preferred constraint is not met", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        preferredPlacement: "network.topology.nvidia.com/zone",
        actualPlacement: { level: "network.topology.nvidia.com/region", value: "us-east" },
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const zoneRow = levelRows[1];
    const icon = zoneRow.find(".constraint-icon .q-icon");
    expect(icon.exists()).toBe(true);
    expect(icon.attributes("class")).toContain("constraint-icon--warning");
  });

  it("should show green check when preferred constraint is met", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        preferredPlacement: "network.topology.nvidia.com/zone",
        actualPlacement: { level: "network.topology.nvidia.com/zone", value: "us-central" },
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const zoneRow = levelRows[1];
    const icon = zoneRow.find(".constraint-icon .q-icon");
    expect(icon.exists()).toBe(true);
    expect(icon.attributes("class")).toContain("constraint-icon--success");
  });

  it("should not show icons when there is no actual placement", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        requiredPlacement: "network.topology.nvidia.com/region",
        preferredPlacement: "network.topology.nvidia.com/zone",
      },
      baseLevels,
    );
    const icons = wrapper.findAll(".constraint-icon");
    expect(icons.length).toBe(0);
  });

  it("should handle both required and preferred on different levels", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        requiredPlacement: "network.topology.nvidia.com/region",
        preferredPlacement: "network.topology.nvidia.com/zone",
        actualPlacement: { level: "network.topology.nvidia.com/region", value: "us-east" },
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");

    const regionRow = levelRows[0];
    expect(regionRow.text()).toContain("Required");
    expect(regionRow.text()).toContain("Actual");

    const zoneRow = levelRows[1];
    expect(zoneRow.text()).toContain("Preferred");
  });

  it("should show green check when actual is more specific than constraint level", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        preferredPlacement: "network.topology.nvidia.com/region",
        actualPlacement: { level: "kubernetes.io/hostname", value: "worker-0" },
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const regionRow = levelRows[0];
    const icon = regionRow.find(".constraint-icon .q-icon");
    expect(icon.exists()).toBe(true);
    expect(icon.attributes("class")).toContain("constraint-icon--success");
  });

  it("should show warning when actual is less specific than constraint level", () => {
    const wrapper = createWrapper(
      {
        name: "my-topology",
        requiredPlacement: "network.topology.nvidia.com/rack",
        actualPlacement: { level: "network.topology.nvidia.com/zone", value: "zone-a" },
      },
      baseLevels,
    );
    const levelRows = wrapper.findAll(".topology-level-row");
    const rackRow = levelRows[2];
    const icon = rackRow.find(".constraint-icon .q-icon");
    expect(icon.exists()).toBe(true);
    expect(icon.attributes("class")).toContain("constraint-icon--warning");
  });
});
