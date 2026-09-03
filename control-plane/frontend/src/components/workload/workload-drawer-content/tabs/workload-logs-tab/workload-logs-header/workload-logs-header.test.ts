/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, beforeEach, it, expect, vi } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import type { AxiosResponse } from "axios";

import { ELogsSource, type ILogParams } from "@/models/logs.model";
import type { WorkloadServicePod } from "@/models/workload.model";
import type { Workload } from "@/swagger-models/workloads-client";
import { researcherService } from "@/services/cluster/researcher.service/researcher.service";
import { clusterApiService } from "@/services/cluster/cluster-api.service/cluster-api.service";
import { alertUtil } from "@/utils/alert.util";
import { createDownloadLink } from "@/utils/common.util";

vi.mock("@/services/cluster/researcher.service/researcher.service");
vi.mock("@/services/cluster/cluster-api.service/cluster-api.service");
vi.mock("@/utils/alert.util");
vi.mock("@/utils/common.util");

import WorkloadLogsHeader from "./workload-logs-header.vue";

describe("workload-logs-header.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadLogsHeader>>;
  let pinia: TestingPinia;

  const mockWorkload: Workload = {
    id: "workload-1",
    name: "test-workload",
    projectName: "test-project",
    clusterId: "cluster-1",
    phase: "Running",
    tenantId: 1001,
    runningPods: 1,
    phaseUpdatedAt: "2024-01-01T00:00:00Z",
    k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
    source: "CLI",
    deletedAt: null,
    type: "Workspace",
    priority: null,
    priorityClassName: "",
    projectId: "1",
    departmentName: "test-dept",
    departmentId: "1",
    namespace: "test-namespace",
    createdAt: "2024-01-01T00:00:00Z",
    conditions: [],
    k8sPhase: "",
    category: "",
  };

  const mockPods: WorkloadServicePod[] = [
    {
      name: "test-pod-1",
      priorityClassName: "",
      id: "pod-1",
      workloadId: "workload-1",
      clusterId: "cluster-1",
      createdAt: "2024-01-01T00:00:00Z",
      containers: [
        { name: "container-1", image: "test-image:latest" },
        { name: "container-2", image: "test-image2:latest" },
      ],
      k8sPhase: "Running",
      tenantId: 1001,
      k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    },
    {
      name: "test-pod-2",
      priorityClassName: "",
      id: "pod-2",
      workloadId: "workload-1",
      clusterId: "cluster-1",
      createdAt: "2024-01-01T00:00:00Z",
      containers: [{ name: "container-3", image: "test-image3:latest" }],
      k8sPhase: "Running",
      tenantId: 1001,
      k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    },
  ];

  const mockLogParams: ILogParams = {
    project: "test-project",
    jobName: "test-workload",
    podName: "test-pod-1",
    tail: 1000,
    type: "Workspace",
    distributedFramework: undefined,
    container: "container-1",
  };

  const mockExternalLogParams: ILogParams = {
    project: "test-project",
    jobName: "test-workload",
    podName: "test-pod-1",
    tail: 1000,
    type: "Workspace",
    distributedFramework: undefined,
    container: "container-1",
    externalWorkloadName: "external-workload-123",
  };

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    setActivePinia(pinia);

    vi.clearAllMocks();
    vi.clearAllTimers();

    const mockNotify = vi.fn();

    wrapper = mount(WorkloadLogsHeader, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        logParams: mockLogParams,
        supportClusterApiLogs: true,
        disableExport: false,
      },
    });

    wrapper.vm.$q.notify = mockNotify;
  });

  describe("downloadLogs method", () => {
    it("should use cluster API when supportClusterApiLogs is true", async () => {
      const mockResponse = {
        data: new Blob(["log content"], { type: "text/plain" }),
        headers: {
          "content-disposition":
            'attachment; filename="test-workload_test-project_test-pod-1_container-1_2024-01-01T12:00:00Z.txt"',
        },
      };
      vi.mocked(clusterApiService.downloadLogsClusterApi).mockResolvedValue(mockResponse);

      await wrapper.vm.downloadLogs();

      expect(clusterApiService.downloadLogsClusterApi).toHaveBeenCalledWith(mockLogParams);
      expect(researcherService.downloadLogs).not.toHaveBeenCalled();
      expect(createDownloadLink).toHaveBeenCalledWith(
        mockResponse.data,
        "test-workload_test-project_test-pod-1_container-1_2024-01-01T12:00:00Z.txt",
      );
    });

    it("should use legacy API when supportClusterApiLogs is false", async () => {
      await wrapper.setProps({ supportClusterApiLogs: false });
      const mockBlob = new Blob(["log content"], { type: "text/plain" });
      vi.mocked(researcherService.downloadLogs).mockResolvedValue(mockBlob);

      await wrapper.vm.downloadLogs();

      expect(researcherService.downloadLogs).toHaveBeenCalledWith(mockLogParams);
      expect(clusterApiService.downloadLogsClusterApi).not.toHaveBeenCalled();
      expect(createDownloadLink).toHaveBeenCalledWith(mockBlob, "logs.txt");
    });

    it("should extract filename from content-disposition header", async () => {
      const mockResponse = {
        data: new Blob(["log content"], { type: "text/plain" }),
        headers: {
          "content-disposition": 'attachment; filename="custom-filename_2024-01-01T12:00:00Z.txt"',
        },
      };
      vi.mocked(clusterApiService.downloadLogsClusterApi).mockResolvedValue(mockResponse);

      await wrapper.vm.downloadLogs();

      expect(createDownloadLink).toHaveBeenCalledWith(mockResponse.data, "custom-filename_2024-01-01T12:00:00Z.txt");
    });

    it("should use default filename when content-disposition header is missing", async () => {
      const mockResponse = {
        data: new Blob(["log content"], { type: "text/plain" }),
        headers: {
          "Content-Type": "text/plain",
        },
      };
      vi.mocked(clusterApiService.downloadLogsClusterApi).mockResolvedValue(mockResponse);

      await wrapper.vm.downloadLogs();

      expect(createDownloadLink).toHaveBeenCalledWith(mockResponse.data, "logs.txt");
    });

    it("should handle filenames with quotes in content-disposition header", async () => {
      const mockResponse = {
        data: new Blob(["log content"], { type: "text/plain" }),
        headers: {
          "content-disposition": 'attachment; filename="quoted-filename.txt"',
        },
      };
      vi.mocked(clusterApiService.downloadLogsClusterApi).mockResolvedValue(mockResponse);

      await wrapper.vm.downloadLogs();

      expect(createDownloadLink).toHaveBeenCalledWith(mockResponse.data, "quoted-filename.txt");
    });

    it("should show error notification when cluster API download fails", async () => {
      const mockError = new Error("Download failed");
      vi.mocked(clusterApiService.downloadLogsClusterApi).mockRejectedValue(mockError);
      vi.mocked(alertUtil.getError).mockReturnValue({ message: "Failed to download logs" });

      await wrapper.vm.downloadLogs();

      expect(alertUtil.getError).toHaveBeenCalledWith("Failed to download logs");
      expect(wrapper.vm.$q.notify).toHaveBeenCalledWith({ message: "Failed to download logs" });
    });

    it("should show error notification when legacy API download fails", async () => {
      await wrapper.setProps({ supportClusterApiLogs: false });
      const mockError = new Error("Legacy download failed");
      vi.mocked(researcherService.downloadLogs).mockRejectedValue(mockError);
      vi.mocked(alertUtil.getError).mockReturnValue({ message: "Failed to download logs" });

      await wrapper.vm.downloadLogs();

      expect(alertUtil.getError).toHaveBeenCalledWith("Failed to download logs");
      expect(wrapper.vm.$q.notify).toHaveBeenCalledWith({ message: "Failed to download logs" });
    });

    it("should use external workload endpoint for external workloads with cluster API", async () => {
      await wrapper.setProps({ logParams: mockExternalLogParams });
      const mockResponse = {
        data: new Blob(["external log content"], { type: "text/plain" }),
        headers: {
          "content-disposition":
            'attachment; filename="external-workload-123_test-project_test-pod-1_container-1_2024-01-01T12:00:00Z.txt"',
        },
      };
      vi.mocked(clusterApiService.downloadLogsClusterApi).mockResolvedValue(mockResponse);

      await wrapper.vm.downloadLogs();

      expect(clusterApiService.downloadLogsClusterApi).toHaveBeenCalledWith(mockExternalLogParams);
      expect(researcherService.downloadLogs).not.toHaveBeenCalled();
      expect(createDownloadLink).toHaveBeenCalledWith(
        mockResponse.data,
        "external-workload-123_test-project_test-pod-1_container-1_2024-01-01T12:00:00Z.txt",
      );
    });

    it("should handle external workload download with missing filename", async () => {
      await wrapper.setProps({ logParams: mockExternalLogParams });
      const mockResponse = {
        data: new Blob(["external log content"], { type: "text/plain" }),
        headers: {
          "Content-Type": "text/plain",
        },
      };
      vi.mocked(clusterApiService.downloadLogsClusterApi).mockResolvedValue(mockResponse);

      await wrapper.vm.downloadLogs();

      expect(clusterApiService.downloadLogsClusterApi).toHaveBeenCalledWith(mockExternalLogParams);
      expect(createDownloadLink).toHaveBeenCalledWith(mockResponse.data, "logs.txt");
    });
  });

  describe("extractFilenameFromHeaders method", () => {
    it("should extract filename from content-disposition header", () => {
      const headers: AxiosResponse["headers"] = {
        "content-disposition": 'attachment; filename="test-filename.txt"',
      };

      const filename = wrapper.vm.extractFilenameFromHeaders(headers);

      expect(filename).toBe("test-filename.txt");
    });

    it("should handle quoted filenames", () => {
      const headers: AxiosResponse["headers"] = {
        "content-disposition": 'attachment; filename="quoted-filename.txt"',
      };

      const filename = wrapper.vm.extractFilenameFromHeaders(headers);

      expect(filename).toBeDefined();
      expect(filename).not.toBe("logs.txt");
    });

    it("should return default filename when content-disposition is missing", () => {
      const headers: AxiosResponse["headers"] = {
        "Content-Type": "text/plain",
      };

      const filename = wrapper.vm.extractFilenameFromHeaders(headers);

      expect(filename).toBe("logs.txt");
    });

    it("should return default filename when content-disposition has wrong format", () => {
      const headers: AxiosResponse["headers"] = {
        "content-disposition": "inline; no-filename-here",
      };

      const filename = wrapper.vm.extractFilenameFromHeaders(headers);

      expect(filename).toBe("logs.txt");
    });
  });

  describe("events", () => {
    it("should emit pod-selected when setSelectedPod is called", () => {
      wrapper.vm.setSelectedPod("new-pod");

      expect(wrapper.emitted("pod-selected")).toEqual([["new-pod"]]);
    });

    it("should emit container-selected when setSelectedContainer is called", () => {
      wrapper.vm.setSelectedContainer("new-container");

      expect(wrapper.emitted("container-selected")).toEqual([["new-container"]]);
    });

    it("should emit logs-source-selected when setSelectedLogsSource is called", () => {
      wrapper.vm.setSelectedLogsSource(ELogsSource.PreviousInstance);

      expect(wrapper.emitted("logs-source-selected")).toEqual([["previous"]]);
    });
  });

  describe("logs source selector", () => {
    it("should have two options: Current instance and Previous instance", () => {
      const options = wrapper.vm.logsSourceOptions;
      expect(options).toHaveLength(2);
      expect(options[0]).toEqual({ label: "Current instance", value: "current" });
      expect(options[1]).toEqual({ label: "Previous instance", value: "previous" });
    });

    it("should default to Current instance", () => {
      expect(wrapper.props("selectedLogsSource")).toBe("current");
    });
  });

  describe("virtual scroll optimization for large pod lists", () => {
    it("should render WorkloadPodContainerSelector component", () => {
      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });

      expect(selector.exists()).toBe(true);
      expect(selector.props("pods")).toEqual(mockPods);
    });

    it("should pass correct props to WorkloadPodContainerSelector", () => {
      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });

      expect(selector.props("selectedPodName")).toBe("test-pod-1");
      expect(selector.props("selectedContainerName")).toBe("container-1");
      expect(selector.props("showContainerSelect")).toBe(true);
    });

    it("should handle large list of pods efficiently", async () => {
      const largePodList: WorkloadServicePod[] = Array.from({ length: 100 }, (_, i) => ({
        name: `worker-${i}`,
        priorityClassName: "low",
        id: `pod-${i}`,
        workloadId: "workload-1",
        clusterId: "cluster-1",
        createdAt: "2024-01-01T00:00:00Z",
        containers: [{ name: `container-${i}`, image: "test-image:latest" }],
        k8sPhase: "Running",
        tenantId: 1001,
        k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
        updatedAt: "2024-01-01T00:00:00Z",
      }));

      await wrapper.setProps({ pods: largePodList });

      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });
      expect(selector.exists()).toBe(true);
      expect(selector.props("pods")).toHaveLength(100);
    });

    it("should handle distributed training workload with master and worker pods", async () => {
      const distributedPods: WorkloadServicePod[] = [
        {
          name: "training-master-0",
          priorityClassName: "high",
          id: "master-0",
          workloadId: "workload-1",
          clusterId: "cluster-1",
          createdAt: "2024-01-01T00:00:00Z",
          containers: [{ name: "pytorch", image: "pytorch:latest" }],
          k8sPhase: "Running",
          tenantId: 1001,
          k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
          updatedAt: "2024-01-01T00:00:00Z",
        },
        ...Array.from({ length: 84 }, (_, i) => ({
          name: `training-worker-${i}`,
          priorityClassName: "low",
          id: `worker-${i}`,
          workloadId: "workload-1",
          clusterId: "cluster-1",
          createdAt: "2024-01-01T00:00:00Z",
          containers: [{ name: "pytorch", image: "pytorch:latest" }],
          k8sPhase: "Running",
          tenantId: 1001,
          k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
          updatedAt: "2024-01-01T00:00:00Z",
        })),
      ];

      await wrapper.setProps({ pods: distributedPods });

      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });
      expect(selector.exists()).toBe(true);
      expect(selector.props("pods")).toHaveLength(85); // 1 master + 84 workers
    });

    it("should render selector even with 100+ pods", async () => {
      const massivePodList: WorkloadServicePod[] = Array.from({ length: 150 }, (_, i) => ({
        name: `pod-${i}`,
        priorityClassName: "low",
        id: `pod-${i}`,
        workloadId: "workload-1",
        clusterId: "cluster-1",
        createdAt: "2024-01-01T00:00:00Z",
        containers: [{ name: "container", image: "test:latest" }],
        k8sPhase: "Running",
        tenantId: 1001,
        k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
        updatedAt: "2024-01-01T00:00:00Z",
      }));

      await wrapper.setProps({ pods: massivePodList });

      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });
      expect(selector.exists()).toBe(true);
      expect(selector.props("pods")).toHaveLength(150);
    });

    it("should not show container select when supportClusterApiLogs is false", async () => {
      await wrapper.setProps({ supportClusterApiLogs: false });

      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });
      expect(selector.props("showContainerSelect")).toBe(false);
    });

    it("should show container select when supportClusterApiLogs is true", () => {
      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });
      expect(selector.props("showContainerSelect")).toBe(true);
    });
  });

  describe("computed properties with large datasets", () => {
    it("should handle 84 pods correctly", async () => {
      const largePods: WorkloadServicePod[] = Array.from({ length: 84 }, (_, i) => ({
        name: `worker-${i}`,
        priorityClassName: "low",
        id: `pod-${i}`,
        workloadId: "workload-1",
        clusterId: "cluster-1",
        createdAt: "2024-01-01T00:00:00Z",
        containers: [{ name: "container", image: "test:latest" }],
        k8sPhase: "Running",
        tenantId: 1001,
        k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
        updatedAt: "2024-01-01T00:00:00Z",
      }));

      await wrapper.setProps({ pods: largePods });

      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });
      expect(selector.props("pods")).toHaveLength(84);
    });

    it("should select correct pod from large list", async () => {
      const largePods: WorkloadServicePod[] = Array.from({ length: 100 }, (_, i) => ({
        name: `worker-${i}`,
        priorityClassName: "low",
        id: `pod-${i}`,
        workloadId: "workload-1",
        clusterId: "cluster-1",
        createdAt: "2024-01-01T00:00:00Z",
        containers: [{ name: "container", image: "test:latest" }],
        k8sPhase: "Running",
        tenantId: 1001,
        k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
        updatedAt: "2024-01-01T00:00:00Z",
      }));

      await wrapper.setProps({
        pods: largePods,
        logParams: {
          ...mockLogParams,
          podName: "worker-50",
        },
      });

      const selectedPod = wrapper.vm.selectedPod;
      expect(selectedPod).toBeDefined();
      expect(selectedPod?.name).toBe("worker-50");
      expect(selectedPod?.id).toBe("pod-50");
    });

    it("should handle container selection for pod with multiple containers", () => {
      const podWithManyContainers: WorkloadServicePod = {
        name: "multi-container-pod",
        priorityClassName: "low",
        id: "pod-multi",
        workloadId: "workload-1",
        clusterId: "cluster-1",
        createdAt: "2024-01-01T00:00:00Z",
        containers: Array.from({ length: 10 }, (_, i) => ({
          name: `container-${i}`,
          image: `image-${i}:latest`,
        })),
        k8sPhase: "Running",
        tenantId: 1001,
        k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
        updatedAt: "2024-01-01T00:00:00Z",
      };

      wrapper = mount(WorkloadLogsHeader, {
        global: {
          plugins: [Quasar, pinia],
        },
        props: {
          workload: mockWorkload,
          pods: [podWithManyContainers],
          logParams: {
            ...mockLogParams,
            podName: "multi-container-pod",
            container: "container-5",
          },
          supportClusterApiLogs: true,
          disableExport: false,
        },
      });

      expect(wrapper.vm.selectedContainerName).toBe("container-5");
      const selector = wrapper.findComponent({ name: "WorkloadPodContainerSelector" });
      expect(selector.props("selectedContainerName")).toBe("container-5");
      expect(selector.props("pods")[0].containers).toHaveLength(10);
    });
  });
});
