import { describe, beforeEach, it, expect, vi } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import type { Workload } from "@/swagger-models/workloads-client";
import { useClusterStore } from "@/stores/cluster.store";
import { defineComponent } from "vue";
import { Quasar } from "quasar";
import type { WorkloadServicePod } from "@/models/workload.model";

vi.mock("@/utils/interval.util");

vi.mock("@/components/common/runai-log-reader/runai-log-reader.vue", async () => {
  return {
    default: defineComponent({
      name: "RunaiLogReaderMock", // Give it a name to help debugging
      template: '<div class="mock-runai-log-reader"></div>',
    }),
  };
});

import WorkloadLogsTab from "./workload-logs-tab.vue";
import { ELogsSource } from "@/models/logs.model";

describe("workload-logs-tab.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadLogsTab>>;
  let pinia: TestingPinia;

  const workload: Workload = {
    id: "1",
    name: "test-workload",
    projectName: "test-project",
    clusterId: "cluster-1",
    phase: "Running",
    tenantId: 0,
    runningPods: 0,
    phaseUpdatedAt: "",
    k8sPhaseUpdatedAt: "",
    updatedAt: "",
    source: "CLI",
    deletedAt: null,
    type: "Training",
    priority: null,
    priorityClassName: "",
    projectId: "",
    departmentName: "",
    departmentId: "",
    namespace: "",
    createdAt: "",
    conditions: [],
    k8sPhase: "",
    category: "",
  };

  const externalWorkload: Workload = {
    ...workload,
    name: "regular-workload-name",
    externalWorkloadName: "external-workload-123",
  };

  const pods: WorkloadServicePod[] = [
    {
      name: "pod-1",
      priorityClassName: "",
      id: "pod-1-id",
      workloadId: "1",
      clusterId: "cluster-1",
      createdAt: "",
      containers: [{ name: "container-1", image: "test:latest" }],
      k8sPhase: "Running",
      tenantId: 0,
      k8sPhaseUpdatedAt: "",
      updatedAt: "",
    },
  ];

  beforeEach(() => {
    // Create and set up Pinia
    pinia = createTestingPinia({ createSpy: vi.fn });
    setActivePinia(pinia);

    // Mock the cluster store
    const clusterStore = useClusterStore(pinia);
    vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(true);

    wrapper = mount(WorkloadLogsTab, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload,
        pods,
        loadingPods: false,
        selectedPodName: "pod-1",
        selectedContainerName: "container-1",
      },
    });
    wrapper.vm.$q.notify = vi.fn();
  });

  it("shows the loader when loadingPods or loadingLogs is true", async () => {
    await wrapper.setProps({ loadingPods: true });
    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(true);
    await wrapper.setProps({ loadingPods: false });
    await wrapper.setData({ loadingLogs: true });
    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(true);
  });

  it("hides the loader when loadingPods and loadingLogs are false", async () => {
    await wrapper.setProps({ loadingPods: false });
    await wrapper.setData({ loadingLogs: false });
    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(false);
  });

  it("shows the no-data view when logs are not available", async () => {
    await wrapper.setData({ loadingLogs: false, noData: true });
    const noDataDrawer = wrapper.findComponent(RunaiDrawerNoData);
    expect(noDataDrawer.exists()).toBe(true);
    expect(noDataDrawer.props("iconName")).toBe("ph-bold ph-file");
    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(false);
  });

  it("shows file-minus when the workload is inactive and has no logs", async () => {
    await wrapper.setProps({ workload: { ...workload, phase: "Stopped" }, pods: [] });
    await wrapper.setData({ loadingLogs: false });

    const noDataDrawer = wrapper.findComponent(RunaiDrawerNoData);
    expect(noDataDrawer.exists()).toBe(true);
    expect(wrapper.text()).toContain("This workload has no logs because it is not active.");
    expect(noDataDrawer.props("iconName")).toBe("ph-bold ph-file-minus");
  });

  it("hides the no-data view when logs are available", async () => {
    await wrapper.setData({ loadingLogs: false, noData: false });
    expect(wrapper.findComponent(RunaiDrawerNoData).exists()).toBe(false);
    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(false);
  });

  it("hides the log reader if an error occurs", async () => {
    await wrapper.setData({ error: true, loadingLogs: false });
    expect(wrapper.findComponent({ name: "RunaiLogReaderMock" }).exists()).toBe(false);
    const noDataDrawer = wrapper.findComponent(RunaiDrawerNoData);
    expect(noDataDrawer.exists()).toBe(true);
    expect(noDataDrawer.props("iconName")).toBe("ph-bold ph-file-minus");
    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(false);
  });

  it("show the log reader if there is no error and data exist", async () => {
    await wrapper.setData({ loadingLogs: false });
    expect(wrapper.findComponent({ name: "RunaiLogReaderMock" }).exists()).toBe(true);
    expect(wrapper.findComponent(RunaiDrawerNoData).exists()).toBe(false);
    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(false);
  });

  it("disables log export when no-data drawer is shown", async () => {
    await wrapper.setData({ noData: true, loadingLogs: false });

    const header = wrapper.findComponent({ name: "WorkloadLogsHeader" });
    expect(header.props("disableExport")).toBe(true);
  });

  describe("getDefaultLogParams", () => {
    it("should generate correct log params for regular workload", () => {
      const logParams = wrapper.vm.getDefaultLogParams();

      expect(logParams).toEqual({
        project: "test-project",
        jobName: "test-workload",
        podName: "",
        tail: 1000,
        type: "Training",
        distributedFramework: undefined,
        container: undefined,
        externalWorkloadName: undefined,
      });
    });

    it("should generate correct log params for external workload", async () => {
      await wrapper.setProps({ workload: externalWorkload });

      const logParams = wrapper.vm.getDefaultLogParams();

      expect(logParams).toEqual({
        project: "test-project",
        jobName: "regular-workload-name",
        podName: "",
        tail: 1000,
        type: "Training",
        distributedFramework: undefined,
        container: undefined,
        externalWorkloadName: "external-workload-123",
      });
    });

    it("should handle workload with missing name", async () => {
      const workloadWithoutName = { ...workload, name: "" };
      await wrapper.setProps({ workload: workloadWithoutName });

      const logParams = wrapper.vm.getDefaultLogParams();

      expect(logParams.jobName).toBe("");
    });

    it("should handle workload with missing project name", async () => {
      const workloadWithoutProject = { ...workload, projectName: "" };
      await wrapper.setProps({ workload: workloadWithoutProject });

      const logParams = wrapper.vm.getDefaultLogParams();

      expect(logParams.project).toBe("");
    });
  });

  describe("Logs source selection", () => {
    it("should default to CurrentInstance", () => {
      expect(wrapper.vm.logsSource).toBe(ELogsSource.CurrentInstance);
    });

    it("should update logParams with previous=true when Previous instance is selected", () => {
      wrapper.vm.handleLogsSourceSelected(ELogsSource.PreviousInstance);

      expect(wrapper.vm.logsSource).toBe(ELogsSource.PreviousInstance);
      expect(wrapper.vm.logParams.previous).toBe(true);
    });

    it("should update logParams with previous=false when Current instance is selected", () => {
      wrapper.vm.handleLogsSourceSelected(ELogsSource.PreviousInstance);
      wrapper.vm.handleLogsSourceSelected(ELogsSource.CurrentInstance);

      expect(wrapper.vm.logsSource).toBe(ELogsSource.CurrentInstance);
      expect(wrapper.vm.logParams.previous).toBe(false);
    });

    it("should reset log state when logs source changes", () => {
      wrapper.vm.error = true;
      wrapper.vm.noData = true;

      wrapper.vm.handleLogsSourceSelected(ELogsSource.PreviousInstance);

      expect(wrapper.vm.error).toBe(false);
      expect(wrapper.vm.noData).toBe(false);
      expect(wrapper.vm.loadingLogs).toBe(true);
    });
  });

  describe("Log Close Event Handling", () => {
    it("Should preserve logs and clear error when close event has logs", async () => {
      await wrapper.setData({ error: true }); // Set initial error state

      wrapper.vm.handleLoggerClose(true); // Has logs

      expect(wrapper.vm.connectionClosed).toBe(true);
      expect(wrapper.vm.hasLogsAfterClose).toBe(true);
      expect(wrapper.vm.error).toBe(false);
      expect(wrapper.vm.loadingLogs).toBe(false);
    });

    it("Should not preserve logs when close event has no logs", async () => {
      wrapper.vm.handleLoggerClose(false); // No logs

      expect(wrapper.vm.connectionClosed).toBe(true);
      expect(wrapper.vm.hasLogsAfterClose).toBe(false);
      expect(wrapper.vm.loadingLogs).toBe(false);
    });

    it("Should show logs after close when hasLogsAfterClose is true", async () => {
      await wrapper.setData({
        connectionClosed: true,
        hasLogsAfterClose: true,
        loadingLogs: false,
        error: false,
      });

      expect(wrapper.vm.showDrawerNoData).toBe(false);
      expect(wrapper.vm.showLogger).toBe(true);
    });

    it("Should delay error setting and not override log preservation", async () => {
      vi.useFakeTimers();

      // First set logs to be preserved
      wrapper.vm.handleLoggerClose(true);
      expect(wrapper.vm.hasLogsAfterClose).toBe(true);

      // Then trigger error handler
      wrapper.vm.handleLoggerError(true);

      // Advance timer
      vi.advanceTimersByTime(100);

      // Error should not be set because logs are preserved
      expect(wrapper.vm.error).toBe(false);

      vi.useRealTimers();
    });

    it("Should set error when no logs to preserve", async () => {
      vi.useFakeTimers();

      // Set no logs preserved
      wrapper.vm.handleLoggerClose(false);
      expect(wrapper.vm.hasLogsAfterClose).toBe(false);

      // Then trigger error handler
      wrapper.vm.handleLoggerError(true);

      // Advance timer
      vi.advanceTimersByTime(100);

      // Error should be set because no logs to preserve
      expect(wrapper.vm.error).toBe(true);

      vi.useRealTimers();
    });
  });
});
