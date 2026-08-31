/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi, beforeEach } from "vitest";
import { inferenceTemplateUtil } from "./inference-template.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import type { InferenceTemplate } from "@/swagger-models/workloads-client";

// Mock the workloadHybridUtil
vi.mock("@/utils/workload-hybrid.util/workload-hybrid.util", () => ({
  workloadHybridUtil: {
    getConnectionsField: vi.fn(),
    getComputeField: vi.fn(),
    getStorageField: vi.fn(),
    getSecurityField: vi.fn(),
  },
}));

describe("inference-template.util", () => {
  const mockWorkloadHybridUtil = vi.mocked(workloadHybridUtil);

  beforeEach(() => {
    vi.clearAllMocks();

    // Setup default mock implementations
    mockWorkloadHybridUtil.getConnectionsField.mockReturnValue([]);
    mockWorkloadHybridUtil.getComputeField.mockReturnValue({});
    mockWorkloadHybridUtil.getStorageField.mockReturnValue({});
    mockWorkloadHybridUtil.getSecurityField.mockReturnValue({});
  });

  describe("getTemplateModel", () => {
    it("should transform a complete InferenceTemplate to IUISuperWorkloadCreationRequest", () => {
      const mockTemplate: InferenceTemplate = {
        meta: {
          id: "template-123",
          name: "test-inference-template",
          projectId: 123,
          scopeType: "project",
          scopeId: "456",
          departmentId: "789",
          clusterId: "cluster-1",
          description: "Test inference template",
          workloadType: "Inference",
          createdBy: "user1",
          createdAt: "2023-01-01T00:00:00Z",
        },
        spec: {
          command: "python",
          args: "app.py",
          image: "test-image:latest",
          imagePullPolicy: "Always",
          workingDir: "/app",
          createHomeDir: true,
          nodeType: "gpu",
          nodePools: ["gpu-pool"],
          podAffinity: { type: "Preferred", key: "affinity" },
          environmentVariables: [{ name: "ENV_VAR", value: "test" }],
          annotations: [{ name: "annotation", value: "test" }],
          labels: [{ name: "label", value: "test" }],
          tolerations: [{ key: "toleration", value: "test" }],
          priorityClass: "high",
          ports: [{ container: 8080, serviceType: "LoadBalancer" }],
          exposedUrls: [{ url: "http://test.com" }],
          relatedUrls: [{ url: "http://related.com" }],
          compute: {},
          storage: {},
          security: {},
          servingPort: {
            container: 8080,
            protocol: "http",
            authorizationType: "public",
            clusterLocalAccessOnly: false,
          },
          autoscaling: {
            minReplicas: 1,
            maxReplicas: 5,
            initialReplicas: 2,
            activationReplicas: 1,
            concurrencyHardLimit: 100,
            scaleToZeroRetentionSeconds: 60,
            scaleDownDelaySeconds: 30,
            metric: "throughput",
            metricThreshold: 100,
            metricThresholdPercentage: 70,
          },
          servingConfiguration: {
            initializationTimeoutSeconds: 300,
          },
        },
      };

      const result = inferenceTemplateUtil.getTemplateModel(mockTemplate);

      expect(result).toEqual({
        name: mockTemplate.meta?.name,
        projectId: mockTemplate.meta?.projectId?.toString(),
        scopeType: mockTemplate.meta?.scopeType,
        scopeId: mockTemplate.meta?.scopeId,
        departmentId: mockTemplate.meta?.departmentId,
        clusterId: mockTemplate.meta?.clusterId,
        description: mockTemplate.meta?.description,
        spec: {
          command: mockTemplate.spec?.command,
          args: mockTemplate.spec?.args,
          image: mockTemplate.spec?.image,
          imagePullPolicy: mockTemplate.spec?.imagePullPolicy,
          workingDir: mockTemplate.spec?.workingDir,
          createHomeDir: mockTemplate.spec?.createHomeDir,
          nodeType: mockTemplate.spec?.nodeType,
          nodePools: mockTemplate.spec?.nodePools,
          podAffinity: mockTemplate.spec?.podAffinity,
          environmentVariables: mockTemplate.spec?.environmentVariables,
          annotations: mockTemplate.spec?.annotations,
          labels: mockTemplate.spec?.labels,
          tolerations: mockTemplate.spec?.tolerations,
          priorityClass: mockTemplate.spec?.priorityClass,
          ports: mockTemplate.spec?.ports,
          exposedUrls: mockTemplate.spec?.exposedUrls,
          relatedUrls: mockTemplate.spec?.relatedUrls,
          compute: mockTemplate.spec?.compute,
          storage: mockTemplate.spec?.storage,
          security: mockTemplate.spec?.security,
          servingPort: mockTemplate.spec?.servingPort,
          autoscaling: mockTemplate.spec?.autoscaling,
          servingConfiguration: {
            initializationTimeoutSeconds: 300,
          },
          connections: [],
        },
      });

      expect(result.spec.servingPort).toBeDefined();
      expect(result.spec.autoscaling).toBeDefined();
      expect(mockWorkloadHybridUtil.getConnectionsField).toHaveBeenCalled();
      expect(mockWorkloadHybridUtil.getComputeField).toHaveBeenCalled();
      expect(mockWorkloadHybridUtil.getStorageField).toHaveBeenCalled();
      expect(mockWorkloadHybridUtil.getSecurityField).toHaveBeenCalled();
    });

    it("should handle template with minimal data and default values", () => {
      const mockTemplate: InferenceTemplate = {
        meta: {
          id: "template-123",
          name: "",
          projectId: null,
          scopeType: "project",
          scopeId: "456",
          departmentId: null,
          clusterId: "",
          description: "",
          workloadType: "Inference",
          createdBy: "user1",
          createdAt: "2023-01-01T00:00:00Z",
        },
        spec: undefined,
      };

      const result = inferenceTemplateUtil.getTemplateModel(mockTemplate);

      expect(result).toEqual({
        name: "",
        projectId: "",
        scopeType: "project",
        scopeId: "456",
        departmentId: undefined,
        clusterId: "",
        description: "",
        spec: {
          connections: [],
          autoscaling: {
            maxReplicas: 1,
            minReplicas: 1,
          },
          compute: {},
          storage: {},
          security: {},
          servingPort: {
            protocol: "http",
          },
        },
      });

      // Verify that workloadHybridUtil methods were called with empty arrays/undefined
      expect(mockWorkloadHybridUtil.getConnectionsField).toHaveBeenCalledWith([], [], []);
      expect(mockWorkloadHybridUtil.getComputeField).toHaveBeenCalledWith(undefined);
      expect(mockWorkloadHybridUtil.getStorageField).toHaveBeenCalledWith(undefined);
      expect(mockWorkloadHybridUtil.getSecurityField).toHaveBeenCalledWith(undefined);
    });
  });
});
