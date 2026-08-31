/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi, type Mock } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import WorkloadDetailsTab from "./workload-details-tab.vue";
import { EWorkloadType } from "@/models/workload.model";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import {
  SourceApi,
  type Workload,
  type WorkloadDetailed,
  type DistributedWorkload,
} from "@/swagger-models/workloads-client";
import { ForbiddenError, InternalServerError, NotFoundError } from "@/models/http-response.model";

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getWorkloadById: vi.fn(),
    getWorkloadV2ById: vi.fn(),
    getWorkloadCrudById: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNetworkTopologies: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn() }),
}));

const PASSTHROUGH_STUB = { template: "<div><slot /></div>" };

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
        plugins: [
          [Quasar, { plugins: { Notify: vi.fn() } }],
          createTestingPinia({ createSpy: vi.fn, stubActions: false }),
        ],
        stubs: {
          RunaiDefaultLoader: true,
          RunaiDrawerNoData: true,
          RunaiDetailsPanel: PASSTHROUGH_STUB,
          WorkloadGeneralInfoSection: true,
          WorkloadEnvironmentSection: true,
          WorkloadNetworkSection: true,
          WorkloadComputeSection: true,
          WorkloadStorageSection: true,
        },
        mocks: { $q: { notify: vi.fn() } },
      },
    });

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
      expect(wrapper.findComponent({ name: "RunaiDefaultLoader" }).exists()).toBe(true);
      await flushPromises();
      expect(wrapper.findComponent({ name: "RunaiDefaultLoader" }).exists()).toBe(false);
    });

    it("shows the element-empty state when showElementEmptyState is true", async () => {
      const wrapper = createWrapper({ showElementEmptyState: true });
      await flushPromises();
      expect(wrapper.findComponent({ name: "RunaiDrawerNoData" }).exists()).toBe(true);
      expect(workloadService.getWorkloadById).not.toHaveBeenCalled();
    });

    it("renders details for supported (non-native) workloads from WorkloadDetailed alone", async () => {
      const wrapper = createWrapper({ workload: { ...mockWorkload, type: "Deployment" } });
      await flushPromises();

      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      // Supported workloads have no CRUD/asset representation — the CRUD endpoint must be skipped.
      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(wrapper.vm.workloadDetailed).toMatchObject({ id: "workload-123" });
      expect(wrapper.vm.workloadCrud).toBeNull();
      expect(wrapper.findComponent({ name: "RunaiDrawerNoData" }).exists()).toBe(false);
    });

    it("shows the empty state for a supported workload when WorkloadDetailed is unavailable", async () => {
      (workloadService.getWorkloadById as Mock).mockRejectedValue(new Error("boom"));

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: "Deployment" } });
      await flushPromises();

      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(wrapper.vm.workloadDetailed).toBeNull();
      expect(wrapper.findComponent({ name: "RunaiDrawerNoData" }).exists()).toBe(true);
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
      expect(wrapper.vm.workloadDetailed).toMatchObject({ id: "workload-123" });
      expect(wrapper.vm.currentSpec).toEqual({ command: "python -m trainer", args: "--epochs 10" });
      expect(wrapper.vm.$q.notify).not.toHaveBeenCalled();
    });

    it("falls back to v1 observed details when active v2 details are no longer available", async () => {
      (workloadService.getWorkloadV2ById as Mock).mockRejectedValue(new NotFoundError());

      const wrapper = createWrapper({ workload: { ...mockWorkload, sourceApi: SourceApi.WorkloadsV2 } });
      await flushPromises();

      expect(workloadService.getWorkloadV2ById).toHaveBeenCalledWith("workload-123");
      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      expect(workloadService.getWorkloadCrudById).not.toHaveBeenCalled();
      expect(wrapper.vm.workloadDetailed).toMatchObject({ id: "workload-123" });
      expect(wrapper.vm.workloadCrud).toBeNull();
      expect(wrapper.vm.$q.notify).not.toHaveBeenCalled();
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
      expect(wrapper.vm.workloadDetailed).toBeNull();
      expect(wrapper.vm.$q.notify).toHaveBeenCalled();
    });

    it("shows the empty state when both v2 details and the v1 fallback are unavailable", async () => {
      (workloadService.getWorkloadV2ById as Mock).mockRejectedValue(new NotFoundError());
      (workloadService.getWorkloadById as Mock).mockRejectedValue(new NotFoundError());

      const wrapper = createWrapper({ workload: { ...mockWorkload, sourceApi: SourceApi.WorkloadsV2 } });
      await flushPromises();

      expect(workloadService.getWorkloadV2ById).toHaveBeenCalledWith("workload-123");
      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-123");
      expect(wrapper.vm.workloadDetailed).toBeNull();
      expect(wrapper.findComponent({ name: "RunaiDrawerNoData" }).exists()).toBe(true);
      expect(wrapper.vm.$q.notify).toHaveBeenCalled();
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
      expect(wrapper.vm.workloadDetailed).toMatchObject({ id: "workload-123" });
    });

    it("notifies the user when both API calls fail", async () => {
      (workloadService.getWorkloadById as Mock).mockRejectedValue(new Error("boom"));
      (workloadService.getWorkloadCrudById as Mock).mockRejectedValue(new Error("boom"));

      const wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.workloadDetailed).toBeNull();
      expect(wrapper.vm.workloadCrud).toBeNull();
      expect(wrapper.vm.$q.notify).toHaveBeenCalled();
    });

    it("re-fetches when workload.id changes", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      (workloadService.getWorkloadById as Mock).mockClear();

      await wrapper.setProps({ workload: { ...mockWorkload, id: "workload-999" } });
      await flushPromises();

      expect(workloadService.getWorkloadById).toHaveBeenCalledWith("workload-999");
    });
  });

  describe("rendering", () => {
    it("renders the WorkloadDetailsPanelView with the fetched data", async () => {
      const wrapper = createWrapper();
      await flushPromises();

      const view = wrapper.findComponent({ name: "workload-details-panel-view" });
      expect(view.exists()).toBe(true);
      expect(view.props("workloadDetailed")).toMatchObject({ id: "workload-123" });
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

      expect(wrapper.vm.showSpecSelect).toBe(true);
      expect(wrapper.vm.specOptions).toEqual([
        { label: "Master", value: "Master" },
        { label: "Workers", value: "Worker" },
      ]);
    });

    it("does not show the spec selector for non-distributed workloads", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(wrapper.vm.showSpecSelect).toBe(false);
    });

    describe("workers-only distributed (no master)", () => {
      // TFJob-style workload: masterSpec is null, only a worker spec exists.
      const workersOnly = {
        name: "workers-only",
        spec: { image: "ubuntu", distributedFramework: "TF" },
      } as unknown as DistributedWorkload;

      it("offers only the Workers option", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(wrapper.vm.specOptions).toEqual([{ label: "Workers", value: "Worker" }]);
      });

      it("hides the spec selector — a 1-item dropdown adds no value", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(wrapper.vm.showSpecSelect).toBe(false);
        expect(wrapper.find(".spec-selector").exists()).toBe(false);
      });

      it("snaps selectedSpec to Worker after load (Master is invalid here)", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(wrapper.vm.selectedSpec).toBe("Worker");
      });

      it("currentSpec still resolves to the worker spec (so details still render)", async () => {
        (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(workersOnly);

        const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
        await flushPromises();

        expect(wrapper.vm.currentSpec).toEqual({ image: "ubuntu", distributedFramework: "TF" });
      });
    });

    it("returns the master pod spec (no inheritance) when Master is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
      await flushPromises();

      // Default selectedSpec is Master — currentSpec should be the masterSpec only (image=master:latest),
      // without anything inherited from the worker spec.
      expect(wrapper.vm.currentSpec).toEqual({ image: "master:latest" });
    });

    it("returns the worker pod spec when Worker is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.Distributed } });
      await flushPromises();
      await wrapper.setData({ selectedSpec: "Worker" });

      expect(wrapper.vm.currentSpec).toEqual({ image: "worker:latest" });
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
      expect(wrapper.vm.showSpecSelect).toBe(true);
      expect(wrapper.vm.specOptions).toEqual([
        { label: "Master", value: "Master" },
        { label: "Workers", value: "Worker" },
      ]);
      expect(wrapper.vm.currentSpec).toEqual({ image: "master:latest", command: "python", args: "--mode master" });

      await wrapper.setData({ selectedSpec: "Worker" });
      expect(wrapper.vm.currentSpec).toEqual({ image: "worker:latest", command: "python", args: "--mode worker" });
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

      expect(wrapper.vm.showSpecSelect).toBe(true);
      expect(wrapper.vm.specOptions).toEqual([
        { label: "Leader", value: "Leader" },
        { label: "Workers", value: "Worker" },
      ]);
    });

    it("snaps selectedSpec to Leader after load (data-level default Master is invalid for DI)", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedInferenceWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();

      expect(wrapper.vm.selectedSpec).toBe("Leader");
    });

    it("returns the leader nested spec when Leader is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedInferenceWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();

      expect(wrapper.vm.currentSpec).toEqual({ image: "leader:latest", command: "python", args: "--mode leader" });
    });

    it("returns the worker nested spec when Worker is selected", async () => {
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(distributedInferenceWorkload);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();
      await wrapper.setData({ selectedSpec: "Worker" });

      expect(wrapper.vm.currentSpec).toEqual({ image: "worker:latest", command: "python", args: "--mode worker" });
    });

    it("offers only Leader and hides the selector when spec.worker is missing (single-pod DI)", async () => {
      const leaderOnly = {
        spec: { category: "Deploy", leader: { image: "leader:latest" }, workers: 0 },
      } as unknown as DistributedWorkload;
      (workloadService.getWorkloadCrudById as Mock).mockResolvedValue(leaderOnly);

      const wrapper = createWrapper({ workload: { ...mockWorkload, type: EWorkloadType.DistributedInference } });
      await flushPromises();

      expect(wrapper.vm.specOptions).toEqual([{ label: "Leader", value: "Leader" }]);
      expect(wrapper.vm.showSpecSelect).toBe(false);
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
