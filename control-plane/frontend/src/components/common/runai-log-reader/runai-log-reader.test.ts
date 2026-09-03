/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, test, vi, beforeEach, afterEach } from "vitest";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

import RunaiLogReader from "./runai-log-reader.vue";
import type { ILogParams } from "@/models/logs.model";
import { EWorkloadType, EExternalWorkloadType } from "@/models/workload.model";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { ClusterState } from "@/swagger-models/cluster-service-client";
import type { IWebSocketSubscription } from "@/models/global.model";

// Mock services
vi.mock("@/services/infra/websockets.service/websockets.service", () => ({
  webSocketsService: {
    subscribe: vi.fn(),
  },
}));

// Import services and stores
import { useClusterStore } from "@/stores/cluster.store";
import { webSocketsService } from "@/services/infra/websockets.service/websockets.service";

describe("runai-log-reader.vue", () => {
  let logReader: VueWrapper<InstanceType<typeof RunaiLogReader>>;
  let pinia: TestingPinia;
  let clusterStore: ReturnType<typeof useClusterStore>;

  const mockLogParams: ILogParams = {
    podName: "test-pod",
    container: "test-container",
    jobName: "test-job",
    project: "test-project",
    type: EWorkloadType.Training,
    tail: 100,
    distributedFramework: EExternalWorkloadType.PyTorchJob,
  };

  const mockExternalLogParams: ILogParams = {
    podName: "test-pod",
    container: "test-container",
    jobName: "test-job",
    project: "test-project",
    type: EWorkloadType.Training,
    tail: 100,
    externalWorkloadName: "external-workload-123",
  };

  const mockCluster: DisplayedCluster = {
    uuid: "test-cluster-id",
    tenantId: 1,
    name: "test-cluster",
    domain: "https://test-cluster.example.com",
    version: "2.17.0",
    createdAt: "2023-01-01T00:00:00Z",
    status: {
      state: ClusterState.Connected,
    },
  };

  const mockWsSubscription: IWebSocketSubscription = {
    onMessage: vi.fn(),
    onError: vi.fn(),
    onOpen: vi.fn(),
    onClose: vi.fn(),
    unsubscribe: vi.fn(),
  };

  function factory(props = {}): VueWrapper<InstanceType<typeof RunaiLogReader>> {
    const defaultProps = {
      logParams: mockLogParams,
      clusterId: "test-cluster-id",
      supportClusterApiLogs: true,
      ...props,
    };

    const wrapper = mount(RunaiLogReader, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: defaultProps,
    });

    return wrapper;
  }

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    clusterStore = useClusterStore(pinia);
    setActivePinia(pinia);

    // Mock cluster store methods
    vi.mocked(clusterStore.getClusterById).mockReturnValue(mockCluster);

    // Mock websocket service
    vi.mocked(webSocketsService.subscribe).mockReturnValue(mockWsSubscription);

    logReader = factory();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Cluster Version Checking and Initialization", () => {
    test("Should use cluster API log reader when prop is true", () => {
      const testLogReader = factory({ supportClusterApiLogs: true });

      const initClusterApiSpy = vi.spyOn(testLogReader.vm, "initClusterApiLogReader");
      const initOldSpy = vi.spyOn(testLogReader.vm, "initOldLogReader");

      testLogReader.vm.handleInitLogReader();

      expect(initClusterApiSpy).toHaveBeenCalledWith(mockCluster);
      expect(initOldSpy).not.toHaveBeenCalled();
    });

    test("Should use old log reader when prop is false", () => {
      const testLogReader = factory({ supportClusterApiLogs: false });

      const initClusterApiSpy = vi.spyOn(testLogReader.vm, "initClusterApiLogReader");
      const initOldSpy = vi.spyOn(testLogReader.vm, "initOldLogReader");

      testLogReader.vm.handleInitLogReader();

      expect(initOldSpy).toHaveBeenCalledWith(mockCluster);
      expect(initClusterApiSpy).not.toHaveBeenCalled();
    });
  });

  describe("URL Construction", () => {
    test("Should construct correct cluster API URL with container", () => {
      const logParamsWithContainer = {
        ...mockLogParams,
        type: EWorkloadType.Training,
        container: "main-container",
      };

      const testLogReader = factory({
        logParams: logParamsWithContainer,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/cluster-api/api/v1/${
          logParamsWithContainer.project
        }/workloads/${logParamsWithContainer.type.toLowerCase()}/runai/${logParamsWithContainer.jobName}`,
        "logs",
        {
          follow: true,
          pod: logParamsWithContainer.podName,
          container: "main-container",
          tailLines: logParamsWithContainer.tail,
        },
        mockCluster.version,
        true,
      );
    });

    test("Should construct correct cluster API URL without container", () => {
      const logParamsWithoutContainer = {
        ...mockLogParams,
        container: undefined,
      };

      const testLogReader = factory({
        logParams: logParamsWithoutContainer,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/cluster-api/api/v1/${
          logParamsWithoutContainer.project
        }/workloads/${logParamsWithoutContainer.type.toLowerCase()}/runai/${logParamsWithoutContainer.jobName}`,
        "logs",
        {
          follow: true,
          pod: logParamsWithoutContainer.podName,
          tailLines: logParamsWithoutContainer.tail,
        },
        mockCluster.version,
        true,
      );
    });

    test("Should construct correct URL for distributed workload with framework", () => {
      const distributedLogParams = {
        ...mockLogParams,
        type: EWorkloadType.Distributed,
        distributedFramework: EExternalWorkloadType.PyTorchJob,
      };

      const testLogReader = factory({
        logParams: distributedLogParams,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/cluster-api/api/v1/${distributedLogParams.project}/workloads/distributed/pytorch/${distributedLogParams.jobName}`,
        "logs",
        {
          follow: true,
          pod: distributedLogParams.podName,
          container: distributedLogParams.container,
          tailLines: distributedLogParams.tail,
        },
        mockCluster.version,
        true,
      );
    });

    test("Should construct correct URL for inference workload", () => {
      const inferenceLogParams = {
        ...mockLogParams,
        type: EWorkloadType.Inference,
      };

      const testLogReader = factory({
        logParams: inferenceLogParams,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/cluster-api/api/v1/${inferenceLogParams.project}/workloads/inference/knative/${inferenceLogParams.jobName}`,
        "logs",
        {
          follow: true,
          pod: inferenceLogParams.podName,
          container: inferenceLogParams.container,
          tailLines: inferenceLogParams.tail,
        },
        mockCluster.version,
        true,
      );
    });

    test("Should construct correct old API URL", () => {
      logReader.vm.initOldLogReader(mockCluster);

      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/researcher/api/v1/`,
        "logs",
        mockLogParams,
        mockCluster.version,
        false,
      );
    });

    test("Should construct correct URL for external workload", () => {
      const testLogReader = factory({
        logParams: mockExternalLogParams,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/cluster-api/api/v1/${mockExternalLogParams.project}/workloads/external/external/${mockExternalLogParams.externalWorkloadName}`,
        "logs",
        {
          follow: true,
          pod: mockExternalLogParams.podName,
          container: mockExternalLogParams.container,
          tailLines: mockExternalLogParams.tail,
        },
        mockCluster.version,
        true,
      );
    });

    test("Should construct correct URL for external workload without container", () => {
      const externalLogParamsWithoutContainer = {
        ...mockExternalLogParams,
        container: undefined,
      };

      const testLogReader = factory({
        logParams: externalLogParamsWithoutContainer,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/cluster-api/api/v1/${externalLogParamsWithoutContainer.project}/workloads/external/external/${externalLogParamsWithoutContainer.externalWorkloadName}`,
        "logs",
        {
          follow: true,
          pod: externalLogParamsWithoutContainer.podName,
          tailLines: externalLogParamsWithoutContainer.tail,
        },
        mockCluster.version,
        true,
      );
    });

    test("Should ignore jobName for external workloads", () => {
      const externalLogParamsWithDifferentJobName = {
        ...mockExternalLogParams,
        jobName: "different-job-name",
      };

      const testLogReader = factory({
        logParams: externalLogParamsWithDifferentJobName,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      // Verify that jobName is not used in the URL for external workloads
      expect(webSocketsService.subscribe).toHaveBeenCalledWith(
        `${mockCluster.domain}/cluster-api/api/v1/${externalLogParamsWithDifferentJobName.project}/workloads/external/external/${externalLogParamsWithDifferentJobName.externalWorkloadName}`,
        "logs",
        {
          follow: true,
          pod: externalLogParamsWithDifferentJobName.podName,
          container: externalLogParamsWithDifferentJobName.container,
          tailLines: externalLogParamsWithDifferentJobName.tail,
        },
        mockCluster.version,
        true,
      );

      // Ensure the URL doesn't contain the jobName
      const subscribeCalls = vi.mocked(webSocketsService.subscribe).mock.calls;
      const lastCall = subscribeCalls[subscribeCalls.length - 1];
      expect(lastCall[0]).not.toContain("different-job-name");
    });

    test("Should include tailLines parameter when tail is provided", () => {
      const logParamsWithTail = {
        ...mockLogParams,
        tail: 1000,
      };

      const testLogReader = factory({
        logParams: logParamsWithTail,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      const subscribeCalls = vi.mocked(webSocketsService.subscribe).mock.calls;
      const lastCall = subscribeCalls[subscribeCalls.length - 1];
      expect(lastCall[2]).toHaveProperty("tailLines", 1000);
    });

    test("Should not include tailLines parameter when tail is null", () => {
      const logParamsWithNullTail = {
        ...mockLogParams,
        tail: null,
      };

      const testLogReader = factory({
        logParams: logParamsWithNullTail,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      const subscribeCalls = vi.mocked(webSocketsService.subscribe).mock.calls;
      const lastCall = subscribeCalls[subscribeCalls.length - 1];
      expect(lastCall[2]).not.toHaveProperty("tailLines");
    });

    test("Should not include tailLines parameter when tail is undefined", () => {
      const logParamsWithUndefinedTail = {
        ...mockLogParams,
        tail: undefined,
      };

      const testLogReader = factory({
        logParams: logParamsWithUndefinedTail,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      const subscribeCalls = vi.mocked(webSocketsService.subscribe).mock.calls;
      const lastCall = subscribeCalls[subscribeCalls.length - 1];
      expect(lastCall[2]).not.toHaveProperty("tailLines");
    });

    test("Should include tailLines with value 0 when tail is 0", () => {
      const logParamsWithZeroTail = {
        ...mockLogParams,
        tail: 0,
      };

      const testLogReader = factory({
        logParams: logParamsWithZeroTail,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      const subscribeCalls = vi.mocked(webSocketsService.subscribe).mock.calls;
      const lastCall = subscribeCalls[subscribeCalls.length - 1];
      expect(lastCall[2]).toHaveProperty("tailLines", 0);
    });

    test("Should include previous parameter and set follow to false when previous is true", () => {
      const logParamsWithPrevious = {
        ...mockLogParams,
        previous: true,
      };

      const testLogReader = factory({
        logParams: logParamsWithPrevious,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      const subscribeCalls = vi.mocked(webSocketsService.subscribe).mock.calls;
      const lastCall = subscribeCalls[subscribeCalls.length - 1];
      expect(lastCall[2]).toHaveProperty("previous", true);
      expect(lastCall[2]).toHaveProperty("follow", false);
    });

    test("Should not include previous parameter and set follow to true when previous is not set", () => {
      const testLogReader = factory({
        logParams: mockLogParams,
        supportClusterApiLogs: true,
      });
      testLogReader.vm.initClusterApiLogReader(mockCluster);

      const subscribeCalls = vi.mocked(webSocketsService.subscribe).mock.calls;
      const lastCall = subscribeCalls[subscribeCalls.length - 1];
      expect(lastCall[2]).not.toHaveProperty("previous");
      expect(lastCall[2]).toHaveProperty("follow", true);
    });
  });

  describe("WebSocket Event Handlers", () => {
    test("Should set up correct event handlers", () => {
      logReader.vm.handleInitLogReader();

      expect(logReader.vm.wsSubscription?.onMessage).toBe(logReader.vm.appendLog);
      expect(logReader.vm.wsSubscription?.onError).toBe(logReader.vm.onError);
      expect(logReader.vm.wsSubscription?.onClose).toBe(logReader.vm.onClose);
    });
  });

  describe("Component Lifecycle", () => {
    test("Should unsubscribe on unmount", () => {
      logReader.vm.wsSubscription = mockWsSubscription;
      logReader.unmount();

      expect(mockWsSubscription.unsubscribe).toHaveBeenCalled();
    });

    test("Should reinitialize when logParams change", async () => {
      const initSpy = vi.spyOn(logReader.vm, "handleInitLogReader");

      const newLogParams = {
        ...mockLogParams,
        podName: "new-pod",
      };

      await logReader.setProps({
        logParams: newLogParams,
        supportClusterApiLogs: true,
      });

      expect(initSpy).toHaveBeenCalled();
    });

    test("Should not reinitialize when same podName and container", async () => {
      const initSpy = vi.spyOn(logReader.vm, "handleInitLogReader");
      initSpy.mockClear(); // Clear initial call

      const sameLogParams = {
        ...mockLogParams,
        jobName: "different-job", // Change something else
      };

      await logReader.setProps({
        logParams: sameLogParams,
        supportClusterApiLogs: true,
      });

      expect(initSpy).not.toHaveBeenCalled();
    });

    test("Should reinitialize when previous flag changes", async () => {
      const initSpy = vi.spyOn(logReader.vm, "handleInitLogReader");
      initSpy.mockClear();

      const logParamsWithPrevious = {
        ...mockLogParams,
        previous: true,
      };

      await logReader.setProps({
        logParams: logParamsWithPrevious,
        supportClusterApiLogs: true,
      });

      expect(initSpy).toHaveBeenCalled();
    });
  });

  describe("WebSocket Close Code Handling", () => {
    beforeEach(() => {
      // Set up component with some mock logs
      logReader.vm.lines = ["Log line 1", "Log line 2", "Error: Process failed"];
    });

    test("Should preserve logs for normal workload closure (code 1006)", () => {
      const mockCloseEvent = new CloseEvent("close", { code: 1006, reason: "abnormal_closure" });

      logReader.vm.onClose(mockCloseEvent);

      const closeEvents = logReader.emitted("close");
      const errorEvents = logReader.emitted("error");
      expect(closeEvents?.[0]).toEqual([true]);
      expect(errorEvents).toBeUndefined();
    });

    test("Should preserve logs for normal closure (code 1000)", () => {
      const mockCloseEvent = new CloseEvent("close", { code: 1000, reason: "normal_closure" });

      logReader.vm.onClose(mockCloseEvent);

      const closeEvents = logReader.emitted("close");
      const errorEvents = logReader.emitted("error");
      expect(closeEvents?.[0]).toEqual([true]);
      expect(errorEvents).toBeUndefined();
    });

    test("Should hide logs for data corruption (code 1007)", () => {
      const mockCloseEvent = new CloseEvent("close", { code: 1007, reason: "invalid_payload" });

      logReader.vm.onClose(mockCloseEvent);

      const errorEvents = logReader.emitted("error");
      const closeEvents = logReader.emitted("close");
      expect(errorEvents?.[0]).toEqual([true]);
      expect(closeEvents).toBeUndefined();
      expect(logReader.vm.error).toBe(true);
    });

    test("Should hide logs for auth failure (code 1008)", () => {
      const mockCloseEvent = new CloseEvent("close", { code: 1008, reason: "policy_violation" });

      logReader.vm.onClose(mockCloseEvent);

      const errorEvents = logReader.emitted("error");
      const closeEvents = logReader.emitted("close");
      expect(errorEvents?.[0]).toEqual([true]);
      expect(closeEvents).toBeUndefined();
      expect(logReader.vm.error).toBe(true);
    });

    test("Should hide logs for TLS handshake failure (code 1015)", () => {
      const mockCloseEvent = new CloseEvent("close", { code: 1015, reason: "tls_handshake" });

      logReader.vm.onClose(mockCloseEvent);

      const errorEvents = logReader.emitted("error");
      const closeEvents = logReader.emitted("close");
      expect(errorEvents?.[0]).toEqual([true]);
      expect(closeEvents).toBeUndefined();
      expect(logReader.vm.error).toBe(true);
    });

    test("Should emit error when no logs and error occurred", () => {
      logReader.vm.lines = []; // No logs
      logReader.vm.errorOccurred = true;
      const mockCloseEvent = new CloseEvent("close", { code: 1006 });

      logReader.vm.onClose(mockCloseEvent);

      const errorEvents = logReader.emitted("error");
      expect(errorEvents?.[0]).toEqual([true]);
      expect(logReader.vm.error).toBe(true);
    });

    test("Should emit close(false) for normal close with no logs", () => {
      logReader.vm.lines = []; // No logs
      logReader.vm.errorOccurred = false;
      const mockCloseEvent = new CloseEvent("close", { code: 1000 });

      logReader.vm.onClose(mockCloseEvent);

      const closeEvents = logReader.emitted("close");
      expect(closeEvents?.[0]).toEqual([false]);
    });

    test("Should preserve logs for unknown close codes", () => {
      const mockCloseEvent = new CloseEvent("close", { code: 9999 }); // Unknown code

      logReader.vm.onClose(mockCloseEvent);

      const closeEvents = logReader.emitted("close");
      const errorEvents = logReader.emitted("error");
      expect(closeEvents?.[0]).toEqual([true]);
      expect(errorEvents).toBeUndefined();
    });
  });
});
