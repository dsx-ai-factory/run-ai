/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import WorkloadPageActions from "./workload-page-actions.vue";
import { Phase, Source, SourceApi } from "@/swagger-models/workloads-client";
import type { Workload, Connection } from "@/swagger-models/workloads-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { EWorkloadCategory, EWorkloadErrorMessage, EWorkloadType } from "@/models/workload.model";

const makeConnection = (overrides: Partial<Connection> = {}): Connection => ({
  name: "Jupyter",
  toolType: "jupyter-notebook" as Connection["toolType"],
  connectionType: "ExternalUrl" as Connection["connectionType"],
  url: "https://example.com/project/workspace",
  containerPort: 8888,
  ...overrides,
});

const makeWorkload = (overrides: Partial<Workload> = {}): Workload => ({
  tenantId: 1,
  phase: Phase.Running,
  type: EWorkloadType.Workspace as Workload["type"],
  urls: ["https://example.com/project/workspace"],
  clusterId: "cluster-1",
  projectId: "project-1",
  source: Source.Other,
  actionsSupport: { delete: true, suspend: true },
  runningPods: 0,
  phaseUpdatedAt: "2024-01-01T00:00:00Z",
  k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
  updatedAt: "2024-01-01T00:00:00Z",
  deletedAt: null,
  name: "test-workload",
  priorityClassName: "",
  projectName: "test-project",
  departmentName: "test-department",
  departmentId: "test-department-id",
  namespace: "test-namespace",
  createdAt: "2024-01-01T00:00:00Z",
  conditions: [],
  k8sPhase: "",
  category: EWorkloadCategory.Workspace,
  id: "workload-1",
  priority: null,
  ...overrides,
});

const mountComponent = (workload: Workload, orgUnitProjectIds: string[] = [workload.projectId]) => {
  return shallowMount(WorkloadPageActions, {
    global: {
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          stubActions: false,
          initialState: {
            Auth: {
              orgUnits: orgUnitProjectIds.map((id) => ({ type: ResourceType.Project, id })),
            },
            Permission: {
              resourceTypeMap: {
                [ResourceType.Workloads]: {
                  resourceType: ResourceType.Workloads,
                  actions: [Action.Create, Action.Read, Action.Delete],
                },
              },
            },
          },
        }),
      ],
      stubs: {
        RunaiTooltipWrapper: { template: "<div><slot /></div>" },
        RunaiActionButton: { template: "<div><slot /></div>" },
        ToolsSelect: true,
      },
    },
    props: {
      selectedWorkloads: [workload],
      loading: false,
      isClusterConnected: true,
      isDeletedViewType: false,
      selectedWorkloadsAmount: 1,
      workload,
    },
  });
};

describe("workload-page-actions - handleConnectClick", () => {
  beforeEach(() => {
    vi.stubGlobal("open", vi.fn());
  });

  describe("YAML-submitted workload (no environments, no externalConnections)", () => {
    it("opens the URL from the urls field via openFirstConnectionIfOldWorkloadModel", () => {
      const workload = makeWorkload({
        urls: ["https://example.com/project/workspace"],
      });
      const wrapper = mountComponent(workload);

      const openFirstConnectionSpy = vi.spyOn(wrapper.vm, "openFirstConnectionIfOldWorkloadModel");

      wrapper.vm.handleConnectClick();

      expect(openFirstConnectionSpy).toHaveBeenCalledTimes(1);
      expect(window.open).toHaveBeenCalledWith("https://example.com/project/workspace", "_blank", "noopener,noreferrer");
    });

    it("rejects non-HTTP URL schemes", () => {
      const workload = makeWorkload({
        urls: ["javascript:alert(document.domain)"],
      });
      const wrapper = mountComponent(workload);

      wrapper.vm.handleConnectClick();

      expect(window.open).not.toHaveBeenCalled();
    });

    it("does not crash when toolItems is empty (regression: !hasMultipleTools includes length 0)", () => {
      const workload = makeWorkload({ urls: ["https://example.com/project/workspace"] });
      const wrapper = mountComponent(workload);

      expect(wrapper.vm.toolItems).toHaveLength(0);

      expect(() => wrapper.vm.handleConnectClick()).not.toThrow();
    });
  });

  describe("UI-submitted workload with a single tool", () => {
    it("opens the tool directly without showing the dropdown", () => {
      const workload = makeWorkload({
        source: Source.ControlPlane as Workload["source"],
        urls: ["https://example.com/project/workspace"],
        environments: [{ connections: [makeConnection()], name: "", id: "" }],
      });
      const wrapper = mountComponent(workload);
      const openToolSpy = vi.spyOn(wrapper.vm, "openTool");

      expect(wrapper.vm.toolItems).toHaveLength(1);

      wrapper.vm.handleConnectClick();

      expect(openToolSpy).toHaveBeenCalledWith("https://example.com/project/workspace");
    });

    it("does not call openFirstConnectionIfOldWorkloadModel when there is exactly one tool", () => {
      const workload = makeWorkload({
        source: Source.ControlPlane as Workload["source"],
        urls: ["https://example.com/project/workspace"],
        environments: [{ connections: [makeConnection()], name: "", id: "" }],
      });
      const wrapper = mountComponent(workload);
      const openFirstConnectionSpy = vi.spyOn(wrapper.vm, "openFirstConnectionIfOldWorkloadModel");

      wrapper.vm.handleConnectClick();

      expect(openFirstConnectionSpy).not.toHaveBeenCalled();
    });
  });

  describe("UI-submitted workload with multiple tools", () => {
    it("defers to openFirstConnectionIfOldWorkloadModel so the dropdown handles selection", () => {
      const workload = makeWorkload({
        source: Source.ControlPlane as Workload["source"],
        urls: ["https://example.com/project/workspace", "https://example.com/project/workspace/tensorboard"],
        environments: [
          {
            connections: [
              makeConnection({ name: "Jupyter", url: "https://example.com/project/workspace" }),
              makeConnection({ name: "TensorBoard", url: "https://example.com/project/workspace/tensorboard" }),
            ],
            name: "",
            id: "",
          },
        ],
      });
      const wrapper = mountComponent(workload);
      const openFirstConnectionSpy = vi.spyOn(wrapper.vm, "openFirstConnectionIfOldWorkloadModel");

      expect(wrapper.vm.toolItems).toHaveLength(2);

      wrapper.vm.handleConnectClick();

      expect(openFirstConnectionSpy).toHaveBeenCalledTimes(1);
    });
  });

  describe("hasMultipleTools computed", () => {
    it("returns false when toolItems is empty", () => {
      const workload = makeWorkload();
      const wrapper = mountComponent(workload);

      expect(wrapper.vm.hasMultipleTools).toBe(false);
    });

    it("returns false when there is exactly one tool", () => {
      const workload = makeWorkload({
        environments: [{ connections: [makeConnection()], name: "", id: "" }],
      });
      const wrapper = mountComponent(workload);

      expect(wrapper.vm.hasMultipleTools).toBe(false);
    });

    it("returns true when there are two or more tools", () => {
      const workload = makeWorkload({
        environments: [
          {
            connections: [makeConnection({ name: "Jupyter" }), makeConnection({ name: "TensorBoard" })],
            name: "",
            id: "",
          },
        ],
      });
      const wrapper = mountComponent(workload);

      expect(wrapper.vm.hasMultipleTools).toBe(true);
    });
  });
});

describe("workload-page-actions - runBtnTooltip", () => {
  it("returns the project-no-longer-exists message when the project is missing", () => {
    const workload = makeWorkload({ projectId: "missing-project" });
    const wrapper = mountComponent(workload, []);

    expect(wrapper.vm.runBtnTooltip).toBe("A workload whose project no longer exists cannot be activated.");
  });

  it("returns the not-authorized message when the workload doesn't support suspend", () => {
    const workload = makeWorkload({ actionsSupport: { delete: true, suspend: false } });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.runBtnTooltip).toBe(EWorkloadErrorMessage.RunActionNotAuthorized);
  });

  it("returns the failed-needs-copy message when the workload is in Failed phase", () => {
    const workload = makeWorkload({ phase: Phase.Failed });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.runBtnTooltip).toBe(EWorkloadErrorMessage.WorkloadFailedCantBeRun);
  });

  it("falls back to the already-running message for a runnable runai workload in a non-stopped phase", () => {
    const workload = makeWorkload({ phase: Phase.Running });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.runBtnTooltip).toBe("The selected workload is already running or being prepared to run");
  });
});

describe("workload-page-actions - copy action", () => {
  it("allows copying a WorkloadsV2 workload translated from a native workload", () => {
    const workload = makeWorkload({
      source: Source.ControlPlane as Workload["source"],
      sourceApi: SourceApi.WorkloadsV2,
      type: EWorkloadType.Training as Workload["type"],
    });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.isCopyWorkloadDisabled).toBe(false);
  });

  it("keeps copy disabled for direct WorkloadsV2 API workloads that are not native workload types", () => {
    const workload = makeWorkload({
      source: Source.ControlPlane as Workload["source"],
      sourceApi: SourceApi.WorkloadsV2,
      type: "Job" as Workload["type"],
    });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.isCopyWorkloadDisabled).toBe(true);
    expect(wrapper.vm.copyBtnTooltip).toBe(EWorkloadErrorMessage.ActionNotSupported);
  });

  it("keeps copy disabled for non-Run:ai workloads", () => {
    const workload = makeWorkload({
      source: Source.Other as Workload["source"],
      sourceApi: SourceApi.WorkloadsV2,
    });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.isCopyWorkloadDisabled).toBe(true);
    expect(wrapper.vm.copyBtnTooltip).toBe(EWorkloadErrorMessage.ActionNotSupported);
  });
});

describe("workload-page-actions - connectBtnTooltip", () => {
  it("surfaces the must-be-running tooltip when the workload is not in Running phase", () => {
    const workload = makeWorkload({ phase: Phase.Stopped });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.connectBtnTooltip).toBe("To connect, the selected workload must be running");
  });

  it("surfaces the must-be-running tooltip for a Pending workload (no urls yet)", () => {
    const workload = makeWorkload({ phase: Phase.Pending, urls: [] });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.connectBtnTooltip).toBe("To connect, the selected workload must be running");
  });

  it("returns the no-tools message for a Running workload with no external connections", () => {
    const workload = makeWorkload({
      phase: Phase.Running,
      environments: [{ connections: [makeConnection()], name: "", id: "" }],
      externalConnections: [],
    });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.connectBtnTooltip).toBe("There are no tools for the selected workload to connect to");
  });
});

describe("workload-page-actions - stopBtnTooltip", () => {
  it("returns the not-authorized message when the workload doesn't support suspend", () => {
    const workload = makeWorkload({ actionsSupport: { delete: true, suspend: false } });
    const wrapper = mountComponent(workload);

    expect(wrapper.vm.stopBtnTooltip).toBe(EWorkloadErrorMessage.StopActionNotAuthorized);
  });
});
