/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, beforeEach, it, expect, vi, type Mock } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { defineComponent } from "vue";

import WorkloadTerminal from "./workload-terminal.vue";
import type { Workload } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { EDistributedFrameworkUrlSlug, type WorkloadServicePod } from "@/models/workload.model";

// Helper type to access component internal state for testing
type WorkloadTerminalInstance = {
  loading: boolean;
  error: string;
  websocket?: WebSocket;
  closeTerminal: () => void;
};

vi.mock("@/services/cluster/workload.service/workload.service");

// Mock RunaiTerminal component
vi.mock("@/components/common/runai-terminal/runai-terminal.vue", () => ({
  default: defineComponent({
    name: "RunaiTerminalMock",
    props: ["websocket"],
    template: '<div class="mock-runai-terminal">Terminal Mock</div>',
  }),
}));

describe("workload-terminal.vue", () => {
  let wrapper: VueWrapper;
  let pinia: TestingPinia;

  // Helper to access component internal state (setup script components don't expose by default)
  const getComponentData = () => wrapper.vm as unknown as WorkloadTerminalInstance;

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
    type: "Training",
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
      k8sPhase: "Pending",
      tenantId: 1001,
      k8sPhaseUpdatedAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    },
  ];

  const mockCluster = {
    name: "test-cluster",
    domain: "https://test-cluster.example.com",
    uuid: "test-uuid",
    version: "1.0.0",
    tenantId: 1001,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  } as DisplayedCluster;

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    setActivePinia(pinia);

    // Mock auth store
    const authStore = useAuthStore(pinia);
    vi.mocked(authStore).accessToken = "test-token";

    // Mock cluster store
    const clusterStore = useClusterStore(pinia);
    vi.mocked(clusterStore.getClusterById).mockReturnValue(mockCluster);

    // Mock workload service (still needed for getFrameworkForUrl)
    vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue(EDistributedFrameworkUrlSlug.PyTorch);

    // Mock WebSocket
    global.WebSocket = vi.fn().mockImplementation(() => ({
      readyState: WebSocket.OPEN,
      send: vi.fn(),
      close: vi.fn(),
      binaryType: "arraybuffer",
      onopen: null,
      onerror: null,
      onmessage: null,
      onclose: null,
    })) as unknown as typeof WebSocket;
  });

  it("should mount and show loader when loadingPods is true", () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: [],
        loadingPods: true,
        selectedPodName: "",
        selectedContainerName: "",
      },
    });

    // showLoader computed = loadingPods || loading, so with loadingPods=true the loader should be visible
    expect(wrapper.find(".runai-default-loader").exists()).toBe(true);
  });

  it("should use pods from props (parent provides pods)", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();

    // Component should use pods from props, not fetch them
    const header = wrapper.findComponent({ name: "workloadPodContainerSelector" });
    expect(header.props("pods")).toEqual(mockPods);
  });

  it("should use selectedPodName and selectedContainerName from props", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();

    const header = wrapper.findComponent({ name: "workloadPodContainerSelector" });
    expect(header.props("selectedPodName")).toBe("test-pod-1");
    expect(header.props("selectedContainerName")).toBe("container-1");
  });

  it("should create websocket with correct URL", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    expect(global.WebSocket).toHaveBeenCalledWith(
      expect.stringContaining(
        "wss://test-cluster.example.com/cluster-api/api/v1/test-project/workloads/training/pytorch/test-workload/exec",
      ),
      expect.arrayContaining([expect.stringContaining("auth.bearer.test-token"), "Bearer"]),
    );
  });

  it("should create websocket with correct URL for external workload", async () => {
    const externalWorkload: Workload = {
      ...mockWorkload,
      externalWorkloadName: "external-workload-123",
    };

    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: externalWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    expect(global.WebSocket).toHaveBeenCalledWith(
      expect.stringContaining(
        "wss://test-cluster.example.com/cluster-api/api/v1/test-project/workloads/external/external/external-workload-123/exec",
      ),
      expect.arrayContaining([expect.stringContaining("auth.bearer.test-token"), "Bearer"]),
    );
  });

  it("should show error when cluster is not found", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const clusterStore = useClusterStore(pinia);
    vi.mocked(clusterStore.getClusterById).mockReturnValue(null);

    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    expect(getComponentData().error).toBe("Cluster not found");
    errorSpy.mockRestore();
  });

  it("should emit pod-selected event when pod is selected in header", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();

    // Find header and trigger pod-selected event
    const header = wrapper.findComponent({ name: "workloadPodContainerSelector" });
    await header.vm.$emit("pod-selected", "test-pod-2");

    const podSelectedEvents = wrapper.emitted("pod-selected");
    expect(podSelectedEvents).toBeTruthy();
    expect(podSelectedEvents?.[0]).toEqual(["test-pod-2"]);
  });

  it("should emit container-selected event when container is selected in header", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();

    // Find header and trigger container-selected event
    const header = wrapper.findComponent({ name: "workloadPodContainerSelector" });
    await header.vm.$emit("container-selected", "container-2");

    const containerSelectedEvents = wrapper.emitted("container-selected");
    expect(containerSelectedEvents).toBeTruthy();
    expect(containerSelectedEvents?.[0]).toEqual(["container-2"]);
  });

  it("should reconnect when selectedPodName prop changes", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    const initialWsCallCount = (global.WebSocket as unknown as Mock).mock.calls.length;

    const runningPod: WorkloadServicePod = {
      ...mockPods[0],
      id: "pod-3",
      name: "test-pod-3",
      containers: [{ name: "container-4", image: "test-image4:latest" }],
    };

    // Change selected pod via props
    await wrapper.setProps({
      pods: [...mockPods, runningPod],
      selectedPodName: "test-pod-3",
      selectedContainerName: "container-4",
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Should create a new websocket connection
    expect((global.WebSocket as unknown as Mock).mock.calls.length).toBeGreaterThan(initialWsCallCount);
  });

  it("should not create websocket when selected pod is not running", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-2", // Pending pod
        selectedContainerName: "container-3",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    expect(getComponentData().error).toBe("To connect, the pod must be running");
    expect(global.WebSocket).not.toHaveBeenCalled();
  });

  it("should have websocket that can be closed", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    // Verify websocket is created
    expect(getComponentData().websocket).toBeDefined();

    // Call closeTerminal
    getComponentData().closeTerminal();

    // Verify websocket is closed
    expect(getComponentData().websocket).toBeUndefined();
  });

  it("should reset terminal when workload ID changes", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    const newWorkload = { ...mockWorkload, id: "workload-2", name: "new-workload" };

    await wrapper.setProps({ workload: newWorkload });
    await wrapper.vm.$nextTick();

    // Error should be reset when workload changes
    expect(getComponentData().error).toBe("");
  });

  it("should include correct query params in websocket URL", async () => {
    wrapper = mount(WorkloadTerminal, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        workload: mockWorkload,
        pods: mockPods,
        loadingPods: false,
        selectedPodName: "test-pod-1",
        selectedContainerName: "container-1",
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await wrapper.vm.$nextTick();

    const wsConstructor = global.WebSocket as unknown as Mock;
    const wsUrl = wsConstructor.mock.calls[0][0] as string;

    expect(wsUrl).toContain("pod=test-pod-1");
    expect(wsUrl).toContain("container=container-1");
    expect(wsUrl).toContain("stdin=true");
    expect(wsUrl).toContain("stdout=true");
    expect(wsUrl).toContain("stderr=true");
    expect(wsUrl).toContain("tty=true");
    expect(wsUrl).toContain(`command=${encodeURIComponent(btoa(JSON.stringify(["/bin/bash"])))}`);
  });

  describe("empty and zero states", () => {
    it("should connect to a running pod when workload is not running", async () => {
      const degradedWorkload: Workload = { ...mockWorkload, phase: "Degraded" };

      wrapper = mount(WorkloadTerminal, {
        global: {
          plugins: [Quasar, pinia],
        },
        props: {
          workload: degradedWorkload,
          pods: mockPods,
          loadingPods: false,
          selectedPodName: "test-pod-1",
          selectedContainerName: "container-1",
        },
      });

      await wrapper.vm.$nextTick();
      await new Promise((resolve) => setTimeout(resolve, 0));
      await wrapper.vm.$nextTick();

      expect(global.WebSocket).toHaveBeenCalled();
      expect(wrapper.find(".runai-drawer-nodata").exists()).toBe(false);
    });

    it("should show drawer no data when selected pod is not running", async () => {
      wrapper = mount(WorkloadTerminal, {
        global: {
          plugins: [Quasar, pinia],
        },
        props: {
          workload: mockWorkload,
          pods: mockPods,
          loadingPods: false,
          selectedPodName: "test-pod-2", // Pending pod
          selectedContainerName: "container-3",
        },
      });

      await wrapper.vm.$nextTick();
      await new Promise((resolve) => setTimeout(resolve, 0));
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect(wrapper.find(".runai-drawer-nodata").exists()).toBe(true);
      expect(wrapper.text()).toContain("To connect, the pod must be running");
      expect(global.WebSocket).not.toHaveBeenCalled();
    });

    it("should show connection error with troubleshooting link when websocket fails on a running pod", async () => {
      const mockWsInstance = {
        readyState: WebSocket.OPEN,
        send: vi.fn(),
        close: vi.fn(),
        binaryType: "arraybuffer",
        onopen: null,
        onerror: null,
      };

      global.WebSocket = vi.fn().mockImplementation(() => mockWsInstance) as unknown as typeof WebSocket;

      wrapper = mount(WorkloadTerminal, {
        global: {
          plugins: [Quasar, pinia],
        },
        props: {
          workload: mockWorkload,
          pods: mockPods,
          loadingPods: false,
          selectedPodName: "test-pod-1", // Running pod
          selectedContainerName: "container-1",
        },
      });

      await wrapper.vm.$nextTick();
      await new Promise((resolve) => setTimeout(resolve, 0));
      await wrapper.vm.$nextTick();

      // Trigger WebSocket error for running pod
      const wsInstance = mockWsInstance as { onerror?: ((ev: Event) => void) | null };
      wsInstance.onerror?.({} as Event);
      await wrapper.vm.$nextTick();

      expect(getComponentData().error).toBe("Failed to connect to terminal");

      // Should show connection error via runai-drawer-no-data with troubleshooting link
      expect(wrapper.find(".runai-drawer-nodata").exists()).toBe(true);
      expect(wrapper.text()).toContain("Couldn't connect to the terminal.");
      expect(wrapper.text()).toContain("troubleshooting guide");
      expect(wrapper.find(".troubleshooting-link").exists()).toBe(true);
    });

    it("should show connection error with exclamation icon when cluster is not found", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      const clusterStore = useClusterStore(pinia);
      vi.mocked(clusterStore.getClusterById).mockReturnValue(null);

      wrapper = mount(WorkloadTerminal, {
        global: {
          plugins: [Quasar, pinia],
        },
        props: {
          workload: mockWorkload,
          pods: mockPods,
          loadingPods: false,
          selectedPodName: "test-pod-1",
          selectedContainerName: "container-1",
        },
      });

      await wrapper.vm.$nextTick();
      await new Promise((resolve) => setTimeout(resolve, 0));
      await wrapper.vm.$nextTick();

      expect(wrapper.text()).toContain("Couldn't connect to the terminal.");
      expect(wrapper.find(".troubleshooting-link").exists()).toBe(true);
      expect(wrapper.find(".runai-drawer-nodata").exists()).toBe(true);
      errorSpy.mockRestore();
    });

    it("should show no pods message when pods list is empty", () => {
      wrapper = mount(WorkloadTerminal, {
        global: {
          plugins: [Quasar, pinia],
        },
        props: {
          workload: mockWorkload,
          pods: [],
          loadingPods: false,
          selectedPodName: "",
          selectedContainerName: "",
        },
      });

      expect(wrapper.find(".runai-drawer-nodata").exists()).toBe(true);
      expect(wrapper.text()).toContain("No pods available for this workload.");
    });
  });
});
