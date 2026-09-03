/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import WorkloadDetailsPanelView from "./workload-details-panel-view.vue";
import type { WorkloadDetailed } from "@/swagger-models/workloads-client";

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn() }),
}));

const PASSTHROUGH_STUB = { template: "<div><slot /></div>" };

describe("WorkloadDetailsPanelView", () => {
  const baseWorkload = {
    id: "workload-123",
    name: "test-workload",
    type: "Workspace",
    clusterId: "cluster-123",
    projectName: "p-1",
    departmentName: "default",
    category: "Build",
    priorityClassName: "high",
    preemptible: false,
    createdAt: "2026-01-13T10:32:00Z",
    submittedBy: "test@run.ai",
    additionalFields: { cliCommand: "runai submit test-workload" },
  } as unknown as WorkloadDetailed;

  const mountView = (overrides: Record<string, unknown> = {}) =>
    mount(WorkloadDetailsPanelView, {
      props: {
        workloadDetailed: baseWorkload,
        workloadCrud: null,
        clusterName: "test-cluster",
        topologyLevels: [],
        nodeAffinities: [],
        selectedSpec: "",
        showSpecSelect: false,
        specOptions: [],
        ...overrides,
      },
      global: {
        plugins: [
          [Quasar, { plugins: { Notify: vi.fn() } }],
          createTestingPinia({ createSpy: vi.fn, stubActions: false }),
        ],
        stubs: {
          RunaiDetailsPanel: PASSTHROUGH_STUB,
          RunaiSelect: true,
          WorkloadCliCommandSection: true,
          WorkloadNodeAffinitySection: true,
          WorkloadTopologySection: true,
          WorkloadGeneralInfoSection: true,
          WorkloadEnvironmentSection: true,
          WorkloadNetworkSection: true,
          WorkloadComputeSection: true,
          WorkloadStorageSection: true,
        },
      },
    });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("builds the General information section with the expected labels and groups", () => {
    const wrapper = mountView();

    const general = wrapper.vm.accordionSections.find((s) => s.id === "general");
    expect(general).toBeDefined();
    const generalData = general?.data as { groups: { label: string }[][] };
    expect(generalData.groups).toHaveLength(3);

    const labels = generalData.groups.flat().map((i) => i.label);
    expect(labels).toEqual(
      expect.arrayContaining([
        "Type",
        "Category",
        "Priority",
        "Preemptible",
        "Project",
        "Cluster",
        "Department",
        "Creation time",
        "Created by",
      ]),
    );
  });

  it("includes the CLI Command section when cliCommand is present", () => {
    const wrapper = mountView();
    expect(wrapper.vm.accordionSections.some((s) => s.id === "cli-command")).toBe(true);
  });

  it("omits the CLI Command section when cliCommand is missing", () => {
    const wrapper = mountView({
      workloadDetailed: { ...baseWorkload, additionalFields: {} } as unknown as WorkloadDetailed,
    });
    expect(wrapper.vm.accordionSections.some((s) => s.id === "cli-command")).toBe(false);
  });

  it("emits update:selectedSpec when the select value changes", async () => {
    const wrapper = mountView({ showSpecSelect: true, specOptions: ["Master", "Worker"], selectedSpec: "Master" });

    // simulate the inner runai-select emitting model-value
    wrapper.vm.$emit("update:selectedSpec", "Worker");
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted("update:selectedSpec")).toBeTruthy();
  });
});
