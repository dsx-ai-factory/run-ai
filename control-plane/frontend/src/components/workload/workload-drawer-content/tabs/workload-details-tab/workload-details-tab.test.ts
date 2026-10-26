/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi, type Mock } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar, Notify } from "quasar";
import WorkloadDetailsTab from "./workload-details-tab.vue";
import { EWorkloadType } from "@/models/workload.model";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import {
  SourceApi,
  type Workload,
  type WorkloadDetailed,
  type DistributedWorkload,
} from "@/swagger-models/workloads-client";
import { ForbiddenError, InternalServerError, NotFoundError } from "@/models/http-response.model";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getWorkloadById: vi.fn(),
    getWorkloadV2ById: vi.fn(),
    getWorkloadCrudById: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNetworkTopologyLevelsByName: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn() }),
}));

const PASSTHROUGH_STUB = { template: "<div><slot /></div>" };
// Notify.create is attached dynamically when the Quasar plugin installs, so it doesn't exist yet
// for vi.spyOn to wrap — assign the mock directly instead.
const notifyMock = vi.fn();
Notify.create = notifyMock;

describe("WorkloadDetailsTab", () => {
  const mockWorkload: Workload = {
    id: "workload-123",
    name: "test-workload",
    type: EWorkloadType.Training,
    clusterId: "cluster-123",
  } as Workload;

  const mockWorkloadDetailed = {
    id: "workload-123",
    name: "test-workload",
    type: EWorkloadType.Training,
    clusterId: "cluster-123",
    images: ["test-image:latest"],
    projectName: "p-1",
    departmentName: "default",
    category: "Train",
    priorityClassName: "high",
    preemptible: false,
    createdAt: "2026-01-13T10:32:00Z",
    submittedBy: "test@run.ai",
    additionalFields: { cliCommand: "runai submit test-workload" },
  } as unknown as WorkloadDetailed;

  const mockCrudWorkload = { name: "test-workload", image: "test-image:latest" };

  const createWrapper = (props: Record<string, unknown> = {}) =>
    mount(WorkloadDetailsTab, {
      props: { workload: mockWorkload, ...props },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn, stubActions: false })],
        stubs: {
          RunaiDetailsPanel: PASSTHROUGH_STUB,
          WorkloadGeneralInfoSection: true,
          WorkloadEnvironmentSection: true,
          WorkloadNetworkSection: true,
          WorkloadComputeSection: true,
          WorkloadStorageSection: true,
        },
      },
    });

  function getDetailsPanelView(wrapper: ReturnType<typeof createWrapper>) {
    return wrapper.findComponent({ name: "details-panel-view" });
  }

  function getSections(wrapper: ReturnType<typeof createWrapper>): IDetailsPanelSectionDef[] {
    return getDetailsPanelView(wrapper).props("sections");
  }

  // Environment section renders "Image URL" / "Command" / "Arguments" rows derived from the
  // currently-selected spec — reading them back is how these tests verify which spec resolved,
  // without reaching into the component's internal currentSpec computed.
  function getEnvironmentFields(wrapper: ReturnType<typeof createWrapper>): Record<string, unknown> {
    const environment = getSections(wrapper).find((s) => s.id === "environment");
    const items = (environment?.data as { items: { label: string; value: unknown }[] } | undefined)?.items ?? [];
    return Object.fromEntries(items.map((item) => [item.label, item.value]));
  }

  async function selectSpec(wrapper: ReturnType<typeof createWrapper>, value: string): Promise<void> {
    getDetailsPanelView(wrapper).vm.$emit("update:selectedSpec", value);
    await wrapper.vm.$nextTick();
  }

  beforeEach(() => {
    vi.clearAllMocks();
    (workloadService.getWorkloadById as Mock).mockResolvedValue(mockWorkloadDetailed);
    (workloadService.getWorkloadV2ById as Mock).mockRejectedValue(new Error("not a v2 workload"));
    (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(mockCrudWorkload);
  });

  describe("data fetching + base states", () => {
    it("fetches workload details on mount and renders the container", async () => {
      const wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.classes()).toContain("workload-details");
      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      expect(workloadService.getWorkloadCrudById).toHaveBeenCalledWith("workload-123", EWorkloadType.Training);
    });

    it("shows the loader while fetching", async () => {
      const wrapper = createWrapper();
      // The loading UI itself lives in details-panel-view (tested there) — here we only need to
      // confirm this tab feeds it the right loading flag.
      expect(getDetailsPanelView(wrapper).props("loading")).toBe(true);
      await flushPromises();
      expect(getDetailsPanelView(wrapper).props("loading")).toBe(false);
    });

    it("renders details for supported (non-native) workloads from WorkloadDetailed alone", async () => {
      const wrapper = createWrapper({ workload: { ...mockWorkload, type: "Deployment" } });
      await flushPromises();

      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      // Supported workloads have no CRUD/asset representation — the CRUD endpoint must be skipped.
      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(getSections(wrapper).length).toBeGreaterThan(0);
    });

    it("shows the empty state for a supported workload when WorkloadDetailed is unavailable", async () => {
      (workloadService.getWorkloadById as Mock).mockRejectedValue(new Error("boom"));

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: "Deployment" } });
      await flushPromises();

      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      // The empty-state UI itself lives in details-panel-view (tested there) — here we only need
      // to confirm this tab feeds it an empty sections model when there's nothing to show.
      expect(getSections(wrapper)).toEqual([]);
    });

    it("loads v2 raw manifest details without calling v1 by-id endpoints", async () => {
      (workloadService.getWorkloadV2ById as Mock).mockResolvedValue({
        workload: {
          rawManifest: {
            spec: {
              template: {
                spec: {
                  containers: [
                    {
                      command: ["python", "-m", "trainer"],
                      args: ["--epochs", "10"],
                    },
                  ],
                },
              },
            },
          },
        },
      });

      const wrapper = createWrapper({ workload: { ...mockWorkload, sourceApi: SourceApi.WorkloadsV2 } });
      await flushPromises();

      expect(workloadService.getWorkloadById).not.toHaveBeenCalled();
      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(workloadService.getWorkloadV2ById).toHaveBeenCalledWith("workload-123");
      expect(getEnvironmentFields(wrapper)).toEqual({ Command: "python -m trainer", Arguments: "--epochs 10" });
      expect(notifyMock).not.toHaveBeenCalled();
    });

    it("falls back to v1 observed details when active v2 details are no longer available", async () => {
      (workloadService.getWorkloadV2ById as Mock).mockRejectedValue(new NotFoundError());

      const wrapper = createWrapper({ workload: { ...mockWorkload, sourceApi: SourceApi.WorkloadsV2 } });
      await flushPromises();

      expect(workloadService.getWorkloadV2ById).toHaveBeenCalledWith("workload-123");
      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(getSections(wrapper).length).toBeGreaterThan(0);
      expect(notifyMock).not.toHaveBeenCalled();
    });

    it.each([
      ["authorization", new ForbiddenError()],
      ["server", new InternalServerError()],
      ["network", new Error("network error")],
    ])("does not mask %s failures with a v1 fallback", async (_failureType, error) => {
      (workloadService.getWorkloadV2ById as Mock).mockRejectedValue(error);

      const wrapper = createWrapper({ workload: { ...mockWorkload, sourceApi: SourceApi.WorkloadsV2 } });
      await flushPromises();

      expect(workloadService.getWorkloadById).not.toHaveBeenCalled();
      expect(getSections(wrapper)).toEqual([]);
      expect(notifyMock).toHaveBeenCalled();
    });

    it("shows the empty state when both v2 details and the v1 fallback are unavailable", async () => {
      (workloadService.getWorkloadV2ById as Mock).mockRejectedValue(new NotFoundError());
      (workloadService.getWorkloadById as Mock).mockRejectedValue(new NotFoundError());

      const wrapper = createWrapper({ workload: { ...mockWorkload, sourceApi: SourceApi.WorkloadsV2 } });
      await flushPromises();

      expect(workloadService.getWorkloadV2ById).toHaveBeenCalledWith("workload-123");
      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      expect(getSections(wrapper)).toEqual([]);
      expect(notifyMock).toHaveBeenCalled();
    });

    it("uses v1 get-by-id for deleted v2 workloads instead of v2", async () => {
      const wrapper = createWrapper({
        workload: { ...mockWorkload, sourceApi: SourceApi.WorkloadsV2 },
        isDeletedViewType: true,
      });
      await flushPromises();

      expect(workloadService.getWorkloadV2ById).not.toHaveBeenCalled();
      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      expect(getSections(wrapper).length).toBeGreaterThan(0);
    });

    it("notifies the user when both API calls fail", async () => {
      (workloadService.getWorkloadById as Mock).mockRejectedValue(new Error("boom"));
      (workloadService.getWorkloadCrudById as Mock).mockRejectedValue(new Error("boom"));

      const wrapper = createWrapper();
      await flushPromises();

      expect(getSections(wrapper)).toEqual([]);
      expect(notifyMock).toHaveBeenCalled();
    });

    it("re-fetches when workload.id changes", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      (workloadService.getWorkloadById as Mock).mockClear();

      await wrapper.setProps({ workload: { ...mockWorkload, id: "workload-999" } });
      await flushPromises();

      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-999");
    });

    it("discards a superseded topology response so it cannot overwrite the current workload's levels", async () => {
      let resolveFirst: (levels: string[]) => void = () => {};
      (clusterService.getNetworkTopologyLevelsByName as Mock)
        .mockImplementationOnce(() => new Promise((resolve) => (resolveFirst = resolve)))
        .mockResolvedValueOnce(["rack", "host"]);

      (workloadService.getWorkloadById as Mock)
        .mockResolvedValueOnce({ ...mockWorkloadDetailed, topology: { name: "topology-a" } })
        .mockResolvedValueOnce({ ...mockWorkloadDetailed, id: "workload-999", topology: { name: "topology-b" } });

      const wrapper = createWrapper();
      await flushPromises();

      await wrapper.setProps({ workload: { ...mockWorkload, id: "workload-999" } });
      await flushPromises();

      // The stale first request resolves after the second one has already completed.
      resolveFirst(["region", "zone"]);
      await flushPromises();

      const network = getSections(wrapper).find((s) => s.id === "network");
      expect((network?.data as { topologyLevels: string[] }).topologyLevels).toEqual(["rack", "host"]);
    });
  });

  describe("rendering", () => {
    it("renders details-panel-view with the built sections", async () => {
      const wrapper = createWrapper();
      await flushPromises();

      const view = getDetailsPanelView(wrapper);
      expect(view.exists()).toBe(true);
      expect(getSections(wrapper).find((s) => s.id === "general")).toBeDefined();
    });

    it("builds the General information section from the fetched workload detail", async () => {
      const wrapper = createWrapper();
      await flushPromises();

      const general = getSections(wrapper).find((s) => s.id === "general");
      const labels = (general?.data as { groups: { label: string }[][] }).groups.flat().map((i) => i.label);
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

    it("includes the CLI Command section when cliCommand is present", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(getSections(wrapper).some((s) => s.id === "cli-command")).toBe(true);
    });

    it("omits the CLI Command section when cliCommand is missing", async () => {
      (workloadService.getWorkloadById as Mock).mockResolvedValue({ ...mockWorkloadDetailed, additionalFields: {} });

      const wrapper = createWrapper();
      await flushPromises();

      expect(getSections(wrapper).some((s) => s.id === "cli-command")).toBe(false);
    });
  });

  describe("distributed workload spec selector", () => {
    const distributedWorkload: DistributedWorkload = {
      name: "distributed-workload",
      masterSpec: { image: "master:latest" },
      spec: { image: "worker:latest" },
    } as DistributedWorkload;

    it("offers Master + Worker options for distributed workloads", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
      await flushPromises();

      expect(getDetailsPanelView(wrapper).props("showSpecSelect")).toBe(true);
      expect(getDetailsPanelView(wrapper).props("specOptions")).toEqual([
        { label: "Master", value: "Master" },
        { label: "Workers", value: "Worker" },
      ]);
    });

    it("does not show the spec selector for non-distributed workloads", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(getDetailsPanelView(wrapper).props("showSpecSelect")).toBe(false);
    });

    describe("workers-only distributed (no master)", () => {
      // TFJob-style workload: masterSpec is null, only a worker spec exists.
      const workersOnly = {
        name: "workers-only",
        spec: { image: "ubuntu" },
      } as unknown as DistributedWorkload;

      it("offers only the Workers option", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(getDetailsPanelView(wrapper).props("specOptions")).toEqual([{ label: "Workers", value: "Worker" }]);
      });

      it("hides the spec selector — a 1-item dropdown adds no value", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(getDetailsPanelView(wrapper).props("showSpecSelect")).toBe(false);
      });

      it("snaps selectedSpec to Worker after load (Master is invalid here)", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(getDetailsPanelView(wrapper).props("selectedSpec")).toBe("Worker");
      });

      it("currentSpec still resolves to the worker spec (so details still render)", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(getEnvironmentFields(wrapper)).toEqual({ "Image URL": "ubuntu" });
      });
    });

    it("returns the master pod spec (no inheritance) when Master is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
      await flushPromises();

      // Default selectedSpec is Master — the environment section should reflect the masterSpec
      // only (image=master:latest), without anything inherited from the worker spec.
      expect(getEnvironmentFields(wrapper)).toEqual({ "Image URL": "master:latest" });
    });

    it("returns the worker pod spec when Worker is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
      await flushPromises();
      await selectSpec(wrapper, "Worker");

      expect(getEnvironmentFields(wrapper)).toEqual({ "Image URL": "worker:latest" });
    });

    it("offers Master + Worker options for v2 distributed workloads from raw manifest", async () => {
      (workloadService.getWorkloadV2ById as Mock).mockResolvedValue({
        workload: {
          rawManifest: {
            spec: {
              pytorchReplicaSpecs: {
                Master: {
                  template: podTemplate({ image: "master:latest", command: ["python"], args: ["--mode", "master"] }),
                },
                Worker: {
                  template: podTemplate({ image: "worker:latest", command: ["python"], args: ["--mode", "worker"] }),
                },
              },
            },
          },
        },
      });

      const wrapper = createWrapper({
        workload: { ...mockWorkload, type: EWorkloadType.Distributed, sourceApi: SourceApi.WorkloadsV2 },
      });
      await flushPromises();

      expect(workloadService.getWorkloadById).not.toHaveBeenCalled();
      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(getDetailsPanelView(wrapper).props("showSpecSelect")).toBe(true);
      expect(getDetailsPanelView(wrapper).props("specOptions")).toEqual([
        { label: "Master", value: "Master" },
        { label: "Workers", value: "Worker" },
      ]);
      expect(getEnvironmentFields(wrapper)).toEqual({
        "Image URL": "master:latest",
        Command: "python",
        Arguments: "--mode master",
      });

      await selectSpec(wrapper, "Worker");
      expect(getEnvironmentFields(wrapper)).toEqual({
        "Image URL": "worker:latest",
        Command: "python",
        Arguments: "--mode worker",
      });
    });
  });

  describe("distributed-inference workload spec selector", () => {
    const distributedInferenceWorkload = {
      name: "distributed-inference-workload",
      spec: {
        category: "Deploy",
        leader: { image: "leader:latest", command: "python", args: "--mode leader" },
        worker: { image: "worker:latest", command: "python", args: "--mode worker" },
        workers: 1,
      },
    } as unknown as DistributedWorkload;

    it("offers Leader + Workers options for distributed-inference workloads", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedInferenceWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();

      expect(getDetailsPanelView(wrapper).props("showSpecSelect")).toBe(true);
      expect(getDetailsPanelView(wrapper).props("specOptions")).toEqual([
        { label: "Leader", value: "Leader" },
        { label: "Workers", value: "Worker" },
      ]);
    });

    it("snaps selectedSpec to Leader after load (data-level default Master is invalid for DI)", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedInferenceWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();

      expect(getDetailsPanelView(wrapper).props("selectedSpec")).toBe("Leader");
    });

    it("returns the leader nested spec when Leader is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedInferenceWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();

      expect(getEnvironmentFields(wrapper)).toEqual({
        "Image URL": "leader:latest",
        Command: "python",
        Arguments: "--mode leader",
      });
    });

    it("returns the worker nested spec when Worker is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedInferenceWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();
      await selectSpec(wrapper, "Worker");

      expect(getEnvironmentFields(wrapper)).toEqual({
        "Image URL": "worker:latest",
        Command: "python",
        Arguments: "--mode worker",
      });
    });

    it("offers only Leader and hides the selector when spec.worker is missing (single-pod DI)", async () => {
      const leaderOnly = {
        spec: { category: "Deploy", leader: { image: "leader:latest" }, workers: 0 },
      } as unknown as DistributedWorkload;
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(leaderOnly);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();

      expect(getDetailsPanelView(wrapper).props("specOptions")).toEqual([{ label: "Leader", value: "Leader" }]);
      expect(getDetailsPanelView(wrapper).props("showSpecSelect")).toBe(false);
    });
  });
});

function podTemplate({
  image,
  command,
  args,
}: {
  image?: string;
  command?: string[];
  args?: string[];
} = {}) {
  return {
    spec: {
      containers: [
        {
          image,
          command,
          args,
        },
      ],
    },
  };
}
