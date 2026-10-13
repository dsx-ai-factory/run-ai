/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { workloadHybridUtil } from "./workload-hybrid.util";
import type {
  IUIEnvironmentFields,
  IUIStorageFields,
  IUISuperWorkloadCreationRequest,
  IUISuperWorkloadSpec,
} from "@/models/workload.model";
import {
  DistributedFramework,
  InternalConnectionType,
  ServingPortAccessAuthorizationTypeEnum,
  ToolType,
} from "@/swagger-models/assets-service-client";
import {
  GpuRequestType,
  GpuComputeMode,
  ScopeType,
  UidGidSource,
  WorkloadType,
  type UnifiedWorkload,
  type WorkspaceSpecSpecAllOfCompute,
  type WorkspaceSpecSpecAllOfSecurity,
  AuthorizationType,
} from "@/swagger-models/workloads-client";
import type { IUIConnection } from "@/models/environment.model";
import { EAccessOptions } from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import { UI_EMPTY_DIR_KIND, UI_VOLUME_KIND } from "@/models/data-source.model";

describe("workload-hybrid.util", () => {
  const mockSpec: IUISuperWorkloadSpec = {
    command: "python main.py",
    args: "--batch-size",
    image: "my-container-image:latest",
    imagePullPolicy: "Always",
    workingDir: "/app",
    createHomeDir: true,
    environmentVariables: [
      { name: "value1", value: "value1" },
      { name: "value2", value: "value2" },
    ],
    connections: [
      {
        id: "123",
        name: "my-con",
        toolType: ToolType.Pycharm,
      },
    ],
    security: { runAsNonRoot: true },
    backoffLimit: 100,
    priorityClass: "build",
    servingPort: {
      protocol: "http",
      container: 8080,
      authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
      authorizedGroups: undefined,
      authorizedUsers: undefined,
    },
  };

  describe("getBaseEnvironmentFields", () => {
    it("should return the correct base environment fields from spec and remove non based fields", () => {
      const expectedFields: IUIEnvironmentFields = {
        command: "python main.py",
        args: "--batch-size",
        image: "my-container-image:latest",
        imagePullPolicy: "Always",
        workingDir: "/app",
        createHomeDir: true,
        environmentVariables: [
          { name: "value1", value: "value1" },
          { name: "value2", value: "value2" },
        ],
        connections: [
          {
            id: "123",
            name: "my-con",
            toolType: ToolType.Pycharm,
          },
        ],
        security: { runAsNonRoot: true },
      };

      expect(workloadHybridUtil.getBaseEnvironmentFields(mockSpec)).toEqual(expectedFields);
    });
  });

  describe("getRelevantEnvironmentFields", () => {
    it("should return the correct environment fields with priorityClass", () => {
      const mockModel: IUISuperWorkloadCreationRequest = {
        name: "",
        projectId: "",
        clusterId: "",
        spec: mockSpec,
      };
      const expectedFields: IUIEnvironmentFields = {
        command: "python main.py",
        args: "--batch-size",
        image: "my-container-image:latest",
        imagePullPolicy: "Always",
        workingDir: "/app",
        createHomeDir: true,
        environmentVariables: [
          { name: "value1", value: "value1" },
          { name: "value2", value: "value2" },
        ],
        connections: [
          {
            id: "123",
            name: "my-con",
            toolType: ToolType.Pycharm,
          },
        ],
        security: { runAsNonRoot: true },
      };

      const res: IUIEnvironmentFields = workloadHybridUtil.getRelevantEnvironmentFields(mockModel);
      expect(res).toEqual(expectedFields);
      expect(res.servingPort).toBeUndefined();
    });
  });

  describe("getInferenceRelevantEnvironmentFields", () => {
    it("should return the correct environment fields with servingPort", () => {
      const mockModel: IUISuperWorkloadCreationRequest = {
        name: "",
        projectId: "",
        clusterId: "",
        spec: mockSpec,
      };

      const expectedFields: IUIEnvironmentFields = {
        command: "python main.py",
        args: "--batch-size",
        image: "my-container-image:latest",
        imagePullPolicy: "Always",
        workingDir: "/app",
        createHomeDir: true,
        environmentVariables: [
          { name: "value1", value: "value1" },
          { name: "value2", value: "value2" },
        ],
        connections: [
          {
            id: "123",
            name: "my-con",
            toolType: ToolType.Pycharm,
          },
        ],
        security: { runAsNonRoot: true },
        servingPort: {
          protocol: "http",
          container: 8080,
          authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
          authorizedGroups: undefined,
          authorizedUsers: undefined,
        },
      };

      const res: IUIEnvironmentFields = workloadHybridUtil.getInferenceRelevantEnvironmentFields(mockModel);
      expect(res).toEqual(expectedFields);
      expect(res.servingPort).toBeDefined();
      expect(res.priorityClass).toBeUndefined();
    });
  });

  describe("_getCleanedStorageFieldsFromUI", () => {
    it("returns undefined if storage is empty", () => {
      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI({} as IUIStorageFields)).toBeUndefined();
    });

    it("removes id, type, createdAt fields from PVC", () => {
      const storage: IUIStorageFields = {
        pvc: [{ id: "123", type: "pvc", createdAt: 123, path: "/data", claimInfo: { size: "1G" } }],
      };

      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI(storage)).toEqual({
        pvc: [{ path: "/data", claimInfo: { size: "1G" } }],
      });
    });

    it("removes id, type, createdAt fields from multiple storage types", () => {
      const storage: IUIStorageFields = {
        pvc: [{ id: "1", type: "pvc", createdAt: 123, path: "/app", claimInfo: { size: "5G" } }],
        nfs: [{ id: "2", type: "nfs", createdAt: 1234, path: "/mnt/nfs" }],
      };

      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI(storage)).toEqual({
        pvc: [{ path: "/app", claimInfo: { size: "5G" } }],
        nfs: [{ path: "/mnt/nfs" }],
      });
    });

    it("removes id, type, createdAt fields from emptyDirVolume", () => {
      const storage: IUIStorageFields = {
        emptyDirVolume: [
          {
            id: "e1",
            type: UI_EMPTY_DIR_KIND,
            createdAt: 123,
            path: "/cache",
            sizeLimit: "2G",
          },
        ],
      };

      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI(storage)).toEqual({
        emptyDirVolume: [{ path: "/cache", sizeLimit: "2G" }],
      });
    });

    it("returns undefined if all storage types are empty", () => {
      const storage: IUIStorageFields = {
        pvc: [],
        nfs: [],
        git: [],
        dataVolume: [],
      };

      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI(storage)).toBeUndefined();
    });

    it("merges volumes into pvc correctly", () => {
      const storage: IUIStorageFields = {
        pvc: [{ id: "1", type: "pvc", createdAt: 1234, path: "/app", claimInfo: { size: "2G" } }],
        volumes: [
          {
            id: "2",
            type: UI_VOLUME_KIND,
            path: "/data",
            claimInfo: { size: "10G" },
          },
        ],
      };

      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI(storage)).toEqual({
        pvc: [
          { path: "/app", claimInfo: { size: "2G" } },
          { path: "/data", claimInfo: { size: "10G" } },
        ],
      });
    });

    it("correctly transforms dataVolume structure", () => {
      const storage: IUIStorageFields = {
        dataVolume: [
          {
            dataVolumeId: "dv1",
            mountPath: "/mnt/data",
            id: "123",
            type: "pvc",
          },
        ],
      };

      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI(storage)).toEqual({
        dataVolume: [{ id: "dv1", mountPath: "/mnt/data" }],
      });
    });

    it("removes empty fields after processing", () => {
      const storage: IUIStorageFields = {
        pvc: [],
        nfs: [{ id: "2", type: "nfs", createdAt: 1234, path: "/mnt/nfs" }],
        git: [],
      };

      expect(workloadHybridUtil._getCleanedStorageFieldsFromUI(storage)).toEqual({
        nfs: [{ path: "/mnt/nfs" }],
      });
    });

    it("removes empty name field from PVC during copy/edit", () => {
      const storage: IUIStorageFields = {
        pvc: [
          {
            id: "123",
            type: "pvc",
            createdAt: 123,
            name: "",
            path: "/home",
            claimInfo: { size: "1G", storageClass: "nfs-csi" },
          },
        ],
      };

      const result = workloadHybridUtil._getCleanedStorageFieldsFromUI(storage);

      expect(result?.pvc?.[0]).not.toHaveProperty("name");
    });
  });

  describe("_getCleanedComputeFieldsFromUI", () => {
    it("preserves the configured mode through load and edit", () => {
      const loaded = workloadHybridUtil.getComputeField({
        gpuDevicesRequest: 1,
        gpuComputeMode: GpuComputeMode.SmSharing,
      });
      expect(loaded?.gpuComputeMode).toBe(GpuComputeMode.SmSharing);
      const result = workloadHybridUtil._getCleanedComputeFieldsFromUI({
        ...loaded,
      });
      expect(result).toEqual({ gpuDevicesRequest: 1, gpuComputeMode: GpuComputeMode.SmSharing });
    });

    it("preserves mode omission through load and edit", () => {
      const loaded = workloadHybridUtil.getComputeField({ gpuDevicesRequest: 1 });
      const result = workloadHybridUtil._getCleanedComputeFieldsFromUI(loaded || {});
      expect(result).toEqual({ gpuDevicesRequest: 1 });
      expect(result).not.toHaveProperty("gpuComputeMode");
    });

    it("does not restore cleared memory portions from response aliases", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuDevicesRequest: 2,
        gpuPortionRequest: undefined,
        gpuPortionLimit: undefined,
        gpuMemoryPortionRequest: 0.5,
        gpuMemoryPortionLimit: 0.75,
      };

      expect(workloadHybridUtil._getCleanedComputeFieldsFromUI(compute)).toEqual({ gpuDevicesRequest: 2 });
      expect(compute.gpuMemoryPortionRequest).toBe(0.5);
    });

    it("removes gpuPortionRequest and gpuRequestType when gpuPortionRequest is 1", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuPortionRequest: 1,
        gpuRequestType: GpuRequestType.Portion,
        cpuCoreRequest: 2,
        cpuMemoryRequest: "4",
      };

      const result = workloadHybridUtil._getCleanedComputeFieldsFromUI(compute);
      expect(result).toEqual({
        cpuCoreRequest: 2,
        cpuMemoryRequest: "4",
      });
    });

    it("does not remove fields when gpuPortionRequest is not 1", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuPortionRequest: 0.5,
        gpuRequestType: GpuRequestType.Portion,
        cpuCoreRequest: 2,
        cpuMemoryRequest: "4",
      };

      const result = workloadHybridUtil._getCleanedComputeFieldsFromUI(compute);
      expect(result).toEqual({
        gpuPortionRequest: 0.5,
        gpuRequestType: GpuRequestType.Portion,
        cpuCoreRequest: 2,
        cpuMemoryRequest: "4",
      });
    });

    it("returns undefined if no valid fields are left after cleanup", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuPortionRequest: 1,
        gpuRequestType: GpuRequestType.Portion,
      };

      const result = workloadHybridUtil._getCleanedComputeFieldsFromUI(compute);
      expect(result).toBeUndefined();
    });

    it("removes empty fields from the result", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuPortionRequest: 0.5,
        gpuRequestType: undefined,
        cpuCoreRequest: 2,
        cpuMemoryRequest: undefined,
      };

      const result = workloadHybridUtil._getCleanedComputeFieldsFromUI(compute);
      expect(result).toEqual({
        gpuPortionRequest: 0.5,
        cpuCoreRequest: 2,
      });
    });
  });

  describe("_getCleanedSecurityFieldsFromUI", () => {
    it("returns undefined if security is empty after cleanup", () => {
      const result = workloadHybridUtil._getCleanedSecurityFieldsFromUI(null);

      expect(result).toBeUndefined();
    });

    it("preserves the IdP source when identity fields are already absent", () => {
      const security: WorkspaceSpecSpecAllOfSecurity = { uidGidSource: UidGidSource.FromIdpToken };

      const result = workloadHybridUtil._getCleanedSecurityFieldsFromUI(security);

      expect(result).toEqual({
        uidGidSource: UidGidSource.FromIdpToken,
        runAsGid: undefined,
        runAsUid: undefined,
        supplementalGroups: undefined,
      });
    });
  });

  describe("_getCleanedConnectionsFieldsFromUI", () => {
    it("returns map connections with exposedUrls, ports, and relatedUrls based on connection types", () => {
      const connections: IUIConnection[] = [
        {
          connectionType: InternalConnectionType.NodePort,
          containerPort: 8080,
          nodePort: 3000,
          id: "1234",
          name: "node-port-1",
          toolType: ToolType.ChatbotUi,
          externalUrl: "",
        },
        {
          connectionType: InternalConnectionType.ExternalUrl,
          containerPort: 8080,
          nodePort: 3000,
          id: "1234",
          name: "external-url-1",
          toolType: ToolType.Comet,
          externalUrl: "http://example.com",
        },
        {
          connectionType: InternalConnectionType.ExternalUrl,
          containerPort: 8080,
          nodePort: 3000,
          id: "1234",
          name: "external-url-1",
          toolType: ToolType.Wandb,
          externalUrl: "http://example.com",
        },
        {
          connectionType: InternalConnectionType.ExternalUrl,
          containerPort: 8080,
          nodePort: 3000,
          id: "1234",
          name: "external-url-1",
          toolType: ToolType.Pycharm,
          externalUrl: "http://example.com",
        },
      ];

      const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

      expect(result).toEqual({
        exposedUrls: [
          {
            container: 8080,
            toolName: "external-url-1",
            toolType: ToolType.Pycharm,
            url: "http://example.com",
            authorizedGroups: undefined,
            authorizedUsers: undefined,
          },
        ],
        ports: [
          {
            container: 8080,
            external: 3000,
            toolName: "node-port-1",
            toolType: ToolType.ChatbotUi,
            serviceType: InternalConnectionType.NodePort,
          },
        ],
        relatedUrls: [
          {
            name: "external-url-1",
            type: ToolType.Comet,
            url: "http://example.com",
          },
          {
            name: "external-url-1",
            type: ToolType.Wandb,
            url: "http://example.com",
          },
        ],
      });
    });

    it("sets fields to undefined if there are no connections", () => {
      const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI([]);

      expect(result).toEqual({
        exposedUrls: undefined,
        ports: undefined,
        relatedUrls: undefined,
      });
    });

    it("sets exposedUrls to undefined if no external URLs are provided", () => {
      const connections: IUIConnection[] = [
        {
          connectionType: InternalConnectionType.NodePort,
          containerPort: 8080,
          nodePort: 3000,
          name: "node-port-1",
          id: "id123",
          toolType: ToolType.Wandb,
          externalUrl: "",
          authorizedUsers: ["user1"],
          authorizedGroups: ["group1"],
        },
      ];

      const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

      expect(result).toEqual({
        exposedUrls: undefined,
        ports: [
          {
            container: 8080,
            external: 3000,
            toolName: "node-port-1",
            toolType: ToolType.Wandb,
            serviceType: InternalConnectionType.NodePort,
          },
        ],
        relatedUrls: undefined,
      });
    });

    it("sets relatedUrls to undefined if no Wandb or Comet tool types are provided", () => {
      const connections: IUIConnection[] = [
        {
          connectionType: InternalConnectionType.ExternalUrl,
          containerPort: 8080,
          nodePort: 3000,
          name: "external-url-1",
          toolType: ToolType.JupyterNotebook, // not Wandb or Comet
          externalUrl: "http://example.com",
          authorizedUsers: ["user1"],
          authorizedGroups: ["group1"],
          id: "",
        },
      ];

      const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

      expect(result).toEqual({
        exposedUrls: [
          {
            toolName: "external-url-1",
            toolType: ToolType.JupyterNotebook,
            url: "http://example.com",
            container: 8080,
            authorizedUsers: ["user1"],
            authorizedGroups: ["group1"],
          },
        ],
        ports: undefined,
        relatedUrls: undefined,
      });
    });

    describe("authorizationType conversion", () => {
      it("should set authorizationType to authenticatedUsers when accessOption is EVERYONE", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.ExternalUrl,
            containerPort: 8080,
            name: "test-tool",
            toolType: ToolType.JupyterNotebook,
            externalUrl: "http://example.com",
            accessOption: EAccessOptions.EVERYONE,
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.exposedUrls?.[0]?.authorizationType).toBe(AuthorizationType.AuthenticatedUsers);
      });

      it("should set authorizationType to authorizedUsers when accessOption is SPECIFIC_USERS", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.ExternalUrl,
            containerPort: 8080,
            name: "test-tool",
            toolType: ToolType.JupyterNotebook,
            externalUrl: "http://example.com",
            accessOption: EAccessOptions.SPECIFIC_USERS,
            authorizedUsers: ["user1", "user2"],
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.exposedUrls?.[0]?.authorizationType).toBe(AuthorizationType.AuthorizedUsers);
      });

      it("should set authorizationType to authorizedUsers when accessOption is PRIVATE", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.ExternalUrl,
            containerPort: 8080,
            name: "test-tool",
            toolType: ToolType.JupyterNotebook,
            externalUrl: "http://example.com",
            accessOption: EAccessOptions.PRIVATE,
            authorizedUsers: ["user1"],
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.exposedUrls?.[0]?.authorizationType).toBe(AuthorizationType.AuthorizedUsers);
      });

      it("should set authorizationType to authorizedGroups when accessOption is GROUPS", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.ExternalUrl,
            containerPort: 8080,
            name: "test-tool",
            toolType: ToolType.JupyterNotebook,
            externalUrl: "http://example.com",
            accessOption: EAccessOptions.GROUPS,
            authorizedGroups: ["group1", "group2"],
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.exposedUrls?.[0]?.authorizationType).toBe(AuthorizationType.AuthorizedGroups);
      });

      it("should set authorizationType to undefined when accessOption is not provided", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.ExternalUrl,
            containerPort: 8080,
            name: "test-tool",
            toolType: ToolType.JupyterNotebook,
            externalUrl: "http://example.com",
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.exposedUrls?.[0]?.authorizationType).toBeUndefined();
      });

      it("should set authorizationType to undefined when accessOption is PUBLIC (not supported for exposedUrls)", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.ExternalUrl,
            containerPort: 8080,
            name: "test-tool",
            toolType: ToolType.JupyterNotebook,
            externalUrl: "http://example.com",
            accessOption: EAccessOptions.PUBLIC,
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.exposedUrls?.[0]?.authorizationType).toBeUndefined();
      });

      it("should not set authorizationType for NodePort connections", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.NodePort,
            containerPort: 8080,
            nodePort: 30000,
            name: "test-tool",
            toolType: ToolType.JupyterNotebook,
            accessOption: EAccessOptions.SPECIFIC_USERS,
            authorizedUsers: ["user1"],
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.ports?.[0]).not.toHaveProperty("authorizationType");
      });

      it("should not set authorizationType for RelatedUrl connections (Wandb/Comet)", () => {
        const connections: IUIConnection[] = [
          {
            connectionType: InternalConnectionType.ExternalUrl,
            containerPort: 8080,
            name: "test-tool",
            toolType: ToolType.Wandb,
            externalUrl: "http://example.com",
            accessOption: EAccessOptions.SPECIFIC_USERS,
            id: "test-id",
          },
        ];

        const result = workloadHybridUtil._getCleanedConnectionsFieldsFromUI(connections);

        expect(result.relatedUrls?.[0]).not.toHaveProperty("authorizationType");
      });
    });

    describe("authorizationType conversion from backend to UI", () => {
      it("should convert authenticatedUsers to EVERYONE accessOption when loading from backend", () => {
        const connections = workloadHybridUtil.getConnectionsField(
          [],
          [
            {
              toolName: "Jupyter",
              toolType: ToolType.JupyterNotebook,
              container: 8888,
              authorizationType: AuthorizationType.AuthenticatedUsers,
            },
          ],
          [],
        );

        expect(connections[0].accessOption).toBe(EAccessOptions.EVERYONE);
      });

      it("should convert authorizedUsers to SPECIFIC_USERS accessOption when loading from backend", () => {
        const connections = workloadHybridUtil.getConnectionsField(
          [],
          [
            {
              toolName: "Jupyter",
              toolType: ToolType.JupyterNotebook,
              container: 8888,
              authorizationType: AuthorizationType.AuthorizedUsers,
              authorizedUsers: ["user1", "user2"],
            },
          ],
          [],
        );

        expect(connections[0].accessOption).toBe(EAccessOptions.SPECIFIC_USERS);
        expect(connections[0].authorizedUsers).toEqual(["user1", "user2"]);
      });

      it("should convert authorizedGroups to GROUPS accessOption when loading from backend", () => {
        const connections = workloadHybridUtil.getConnectionsField(
          [],
          [
            {
              toolName: "Jupyter",
              toolType: ToolType.JupyterNotebook,
              container: 8888,
              authorizationType: AuthorizationType.AuthorizedGroups,
              authorizedGroups: ["group1", "group2"],
            },
          ],
          [],
        );

        expect(connections[0].accessOption).toBe(EAccessOptions.GROUPS);
        expect(connections[0].authorizedGroups).toEqual(["group1", "group2"]);
      });

      it("should set accessOption to undefined when authorizationType is not provided", () => {
        const connections = workloadHybridUtil.getConnectionsField(
          [],
          [
            {
              toolName: "Jupyter",
              toolType: ToolType.JupyterNotebook,
              container: 8888,
            },
          ],
          [],
        );

        expect(connections[0].accessOption).toBeUndefined();
      });

      it("should handle the scenario from the user's bug report correctly", () => {
        // Simulating the exact scenario: policy has authorizedGroups with authorizationType,
        // but backend returns authenticatedUsers
        const connections = workloadHybridUtil.getConnectionsField(
          [],
          [
            {
              toolName: "Jupyter",
              toolType: ToolType.JupyterNotebook,
              container: 8888,
              authorizationType: AuthorizationType.AuthenticatedUsers,
              authorizedUsers: ["user1", "user2", "user3"],
              authorizedGroups: ["rtee21"],
            },
          ],
          [],
        );

        // Should convert to EVERYONE, not SPECIFIC_USERS or GROUPS
        expect(connections[0].accessOption).toBe(EAccessOptions.EVERYONE);
        // Should preserve both lists from backend
        expect(connections[0].authorizedUsers).toEqual(["user1", "user2", "user3"]);
        expect(connections[0].authorizedGroups).toEqual(["rtee21"]);
      });
    });
  });

  describe("_getBaseFlatWorkloadSpecData", () => {
    it("throws an error when image is not provided in workloadSpec", () => {
      const workloadSpec: IUISuperWorkloadSpec = { image: undefined };

      expect(() => workloadHybridUtil._getBaseFlatWorkloadSpecData(workloadSpec)).toThrowError(
        "Invalid workloadSpec configuration: image is required but was not provided.",
      );
    });

    it("returns the correct SupersetSpec object when valid data is provided", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        command: "start",
        args: "--verbose",
        imagePullPolicy: "Always",
        workingDir: "/home",
        createHomeDir: true,
        probes: {},
        nodeType: "cpu",
        nodePools: ["pool1"],
        environmentVariables: [{ name: "VAR1", value: "value1" }],
        annotations: [{ name: "KEY1", value: "value1" }],
        labels: [{ name: "LBL1", value: "valueLbl1" }],
        tolerations: [],
        terminateAfterPreemption: true,
        restartPolicy: "Always",
        priorityClass: null,
        autoDeletionTimeAfterCompletionSeconds: 300,
        storage: {},
        compute: {},
        security: {},
        connections: [],
      };
      const result = workloadHybridUtil._getBaseFlatWorkloadSpecData(workloadSpec);

      expect(result).toEqual({
        command: "start",
        args: "--verbose",
        image: "nginx:latest",
        imagePullPolicy: "Always",
        workingDir: "/home",
        createHomeDir: true,
        probes: undefined,
        nodeType: "cpu",
        nodePools: ["pool1"],
        podAffinity: undefined,
        tty: undefined,
        stdin: undefined,
        environmentVariables: [{ name: "VAR1", value: "value1" }],
        annotations: [{ name: "KEY1", value: "value1" }],
        labels: [{ name: "LBL1", value: "valueLbl1" }],
        tolerations: undefined,
        restartPolicy: "Always",
        priorityClass: undefined,
        storage: undefined,
        compute: undefined,
        security: undefined,
        exposedUrls: undefined,
        ports: undefined,
        relatedUrls: undefined,
      });
    });

    it("excludes fields with empty arrays or undefined values", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        nodePools: [],
        environmentVariables: [],
        annotations: undefined,
        labels: undefined,
        tolerations: [],
      };

      const result = workloadHybridUtil._getBaseFlatWorkloadSpecData(workloadSpec);

      expect(result).toEqual({
        image: "nginx:latest",
        nodePools: undefined,
        environmentVariables: undefined,
        annotations: undefined,
        labels: undefined,
        tolerations: undefined,
        storage: undefined,
        compute: undefined,
        security: undefined,
        exposedUrls: undefined,
        ports: undefined,
        relatedUrls: undefined,
      });
    });
  });

  describe("getHybridDistributedModel", () => {
    it("should add sshAuthMountPath from the workerSpec to the masterSpec ", () => {
      const sshAuthMountPath = "/home/local";
      const model = workloadHybridUtil.getHybridDistributedModel({
        spec: {
          worker: {
            sshAuthMountPath,
          },
        },
      });

      expect(model.spec?.sshAuthMountPath).toBe(sshAuthMountPath);
      expect(model.masterSpec?.sshAuthMountPath).toBe(sshAuthMountPath);
    });

    it("should not add sshAuthMountPath from the workerSpec to the masterSpec if no masterSpec is provided", () => {
      const sshAuthMountPath = "/home/local";
      const model = workloadHybridUtil.getHybridDistributedModel(
        {
          spec: { worker: { sshAuthMountPath } },
        },
        true,
      );

      expect(model.spec?.sshAuthMountPath).toBe(sshAuthMountPath);
      expect(model.masterSpec).toBeUndefined();
    });

    it("should preserve segmentSize from the worker specification", () => {
      const model = workloadHybridUtil.getHybridDistributedModel(
        {
          spec: { worker: { segmentSize: 2 } },
        },
        true,
      );

      expect(model.spec.segmentSize).toBe(2);
    });
  });

  describe("getHybridDistributedTemplateModel", () => {
    it("should preserve segmentSize from the template worker specification", () => {
      const model = workloadHybridUtil.getHybridDistributedTemplateModel(
        {
          meta: {
            id: "template-id",
            name: "template-name",
            workloadType: WorkloadType.Distributed,
            scopeType: ScopeType.Project,
            scopeId: "project-id",
            createdBy: "test-user",
            createdAt: "2026-08-02T00:00:00Z",
          },
          spec: { segmentSize: 2 },
        },
        true,
      );

      expect(model.spec.segmentSize).toBe(2);
    });
  });

  describe("prepareWorkloadFlatRequest", () => {
    it("should convert empty string nodeType to undefined", () => {
      const model = workloadHybridUtil.prepareWorkloadFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          nodeType: "",
        },
      });

      expect(model.spec?.nodeType).toBeUndefined();
    });

    it("should preserve valid nodeType value", () => {
      const model = workloadHybridUtil.prepareWorkloadFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          nodeType: "gpu-node",
        },
      });

      expect(model.spec?.nodeType).toBe("gpu-node");
    });
  });

  describe("prepareTrainingFlatRequest", () => {
    it("should convert empty string nodeType to undefined", () => {
      const model = workloadHybridUtil.prepareTrainingFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          nodeType: "",
        },
      });

      expect(model.spec?.nodeType).toBeUndefined();
    });

    it("should preserve valid nodeType value", () => {
      const model = workloadHybridUtil.prepareTrainingFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          nodeType: "gpu-node",
        },
      });

      expect(model.spec?.nodeType).toBe("gpu-node");
    });
  });

  describe("prepareDistributedFlatRequest", () => {
    it("should remove sshAuthMountPath from the masterSpec", () => {
      const sshAuthMountPath = "/home/local";
      const model = workloadHybridUtil.prepareDistributedFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          sshAuthMountPath,
        },
        masterSpec: {
          sshAuthMountPath,
        },
        masterSpecSameAsWorker: false,
      });

      expect(model.spec?.sshAuthMountPath).toBe(sshAuthMountPath);
      expect(model.masterSpec).not.ownProperty("sshAuthMountPath");
    });

    it("should convert empty string nodeType to undefined in worker spec", () => {
      const model = workloadHybridUtil.prepareDistributedFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          nodeType: "",
        },
      });

      expect(model.spec?.nodeType).toBeUndefined();
    });

    it("should convert empty string nodeType to undefined in master spec", () => {
      const model = workloadHybridUtil.prepareDistributedFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          nodeType: "valid-type",
        },
        masterSpec: {
          image: "ubuntu",
          nodeType: "",
        },
        masterSpecSameAsWorker: false,
      });

      expect(model.masterSpec?.nodeType).toBeUndefined();
    });

    it("should preserve valid nodeType value in both specs", () => {
      const model = workloadHybridUtil.prepareDistributedFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          nodeType: "gpu-node",
        },
        masterSpec: {
          image: "ubuntu",
          nodeType: "cpu-node",
        },
        masterSpecSameAsWorker: false,
      });

      expect(model.spec?.nodeType).toBe("gpu-node");
      expect(model.masterSpec?.nodeType).toBe("cpu-node");
    });

    it("should preserve segmentSize in the worker spec and remove it from the master spec", () => {
      const model = workloadHybridUtil.prepareDistributedFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          distributedFramework: DistributedFramework.PyTorch,
          numWorkers: 4,
          segmentSize: 2,
        },
        masterSpec: {
          segmentSize: 2,
        },
        masterSpecSameAsWorker: false,
      });

      expect(model.spec?.segmentSize).toBe(2);
      expect(model.masterSpec).not.ownProperty("segmentSize");
    });

    it("should strip GPU fields from the master spec for MPI workloads", () => {
      const model = workloadHybridUtil.prepareDistributedFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          distributedFramework: DistributedFramework.Mpi,
          compute: { gpuDevicesRequest: 2 },
        },
        masterSpec: {
          image: "ubuntu",
          compute: {
            gpuDevicesRequest: 2,
            gpuRequestType: GpuRequestType.Portion,
            gpuPortionRequest: 0.5,
            gpuPortionLimit: 1,
            gpuMemoryRequest: "100M",
            gpuMemoryLimit: "200M",
            cpuCoreRequest: 1,
          },
        },
        masterSpecSameAsWorker: false,
      });

      expect(model.spec?.compute?.gpuDevicesRequest).toBe(2);
      expect(model.masterSpec?.compute).not.ownProperty("gpuDevicesRequest");
      expect(model.masterSpec?.compute).not.ownProperty("gpuRequestType");
      expect(model.masterSpec?.compute).not.ownProperty("gpuPortionRequest");
      expect(model.masterSpec?.compute).not.ownProperty("gpuPortionLimit");
      expect(model.masterSpec?.compute).not.ownProperty("gpuMemoryRequest");
      expect(model.masterSpec?.compute).not.ownProperty("gpuMemoryLimit");
      expect(model.masterSpec?.compute?.cpuCoreRequest).toBe(1);
    });

    it("should keep GPU fields on the master spec for non-MPI workloads", () => {
      const model = workloadHybridUtil.prepareDistributedFlatRequest({
        name: "test",
        projectId: "test",
        clusterId: "test",
        spec: {
          image: "ubuntu",
          distributedFramework: DistributedFramework.XgBoost,
          compute: { gpuDevicesRequest: 2 },
        },
        masterSpec: {
          image: "ubuntu",
          compute: { gpuDevicesRequest: 2 },
        },
        masterSpecSameAsWorker: false,
      });

      expect(model.masterSpec?.compute?.gpuDevicesRequest).toBe(2);
    });
  });

  describe("setWorkspaceUIDefaults", () => {
    const mockUIModel: IUISuperWorkloadCreationRequest = {
      name: "test-workspace",
      projectId: "project-123",
      clusterId: "cluster-123",
      spec: {
        image: "nginx:latest",
      },
    };

    it("should set terminationGracePeriodSeconds to 30 when cluster version supports termination grace period", () => {
      const result = workloadHybridUtil.setWorkspaceUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBe(30);
    });

    it("should set terminationGracePeriodSeconds to null when cluster version does not support termination grace period", () => {
      const result = workloadHybridUtil.setWorkspaceUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.19", // clusterVersion - does not support termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBeNull();
    });

    it("should preserve existing terminationGracePeriodSeconds value when cluster version supports termination grace period", () => {
      const mockUIModelWithValue = {
        ...mockUIModel,
        spec: {
          ...mockUIModel.spec,
          terminationGracePeriodSeconds: 60,
        },
      };

      const result = workloadHybridUtil.setWorkspaceUIDefaults(
        mockUIModelWithValue,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBe(60);
    });

    it("should omit preemptibility when cluster version does not support it", () => {
      const result = workloadHybridUtil.setWorkspaceUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "non-preemptible",
          },
        },
        false,
        null,
        "2.23",
      );

      expect(result.spec.preemptibility).toBeUndefined();
    });

    it("should preserve preemptibility when cluster version supports it", () => {
      const result = workloadHybridUtil.setWorkspaceUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "non-preemptible",
          },
        },
        false,
        null,
        "2.24",
      );

      expect(result.spec.preemptibility).toBe("non-preemptible");
    });
  });

  describe("setTrainingUIDefaults", () => {
    const mockUIModel: IUISuperWorkloadCreationRequest = {
      name: "test-training",
      projectId: "project-123",
      clusterId: "cluster-123",
      spec: {
        image: "pytorch:latest",
      },
    };

    it("should set terminationGracePeriodSeconds to 30 when cluster version supports termination grace period", () => {
      const result = workloadHybridUtil.setTrainingUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBe(30);
    });

    it("should set terminationGracePeriodSeconds to null when cluster version does not support termination grace period", () => {
      const result = workloadHybridUtil.setTrainingUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.19", // clusterVersion - does not support termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBeNull();
    });

    it("should set completions to 1 and parallelism to null when cluster version supports completions and parallelism", () => {
      const result = workloadHybridUtil.setTrainingUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports completions and parallelism
      );

      expect(result.spec.completions).toBe(1);
      expect(result.spec.parallelism).toBeNull();
    });

    it("should set completions and parallelism to null when cluster version does not support them", () => {
      const result = workloadHybridUtil.setTrainingUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.19", // clusterVersion - does not support completions and parallelism
      );

      expect(result.spec.completions).toBeNull();
      expect(result.spec.parallelism).toBeNull();
    });

    it("should preserve existing completions and parallelism values when cluster version supports them", () => {
      const mockUIModelWithValues = {
        ...mockUIModel,
        spec: {
          ...mockUIModel.spec,
          completions: 5,
          parallelism: 3,
        },
      };

      const result = workloadHybridUtil.setTrainingUIDefaults(
        mockUIModelWithValues,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports completions and parallelism
      );

      expect(result.spec.completions).toBe(5);
      expect(result.spec.parallelism).toBe(3);
    });

    it("should omit preemptibility when cluster version does not support it", () => {
      const result = workloadHybridUtil.setTrainingUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "preemptible",
          },
        },
        false,
        null,
        "2.23",
      );

      expect(result.spec.preemptibility).toBeUndefined();
    });

    it("should preserve preemptibility when cluster version supports it", () => {
      const result = workloadHybridUtil.setTrainingUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "preemptible",
          },
        },
        false,
        null,
        "2.24",
      );

      expect(result.spec.preemptibility).toBe("preemptible");
    });

    it("should omit semi-preemptibility before cluster version 3.0", () => {
      const result = workloadHybridUtil.setDistributedUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "semi-preemptible",
          },
        },
        false,
        null,
        "2.99",
      );

      expect(result.spec.preemptibility).toBeUndefined();
    });

    it("should preserve semi-preemptibility from cluster version 3.0", () => {
      const result = workloadHybridUtil.setDistributedUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "semi-preemptible",
          },
        },
        false,
        null,
        "3.0",
      );

      expect(result.spec.preemptibility).toBe("semi-preemptible");
    });
  });

  describe("setDistributedUIDefaults", () => {
    const mockUIModel: IUISuperWorkloadCreationRequest = {
      name: "test-distributed",
      projectId: "project-123",
      clusterId: "cluster-123",
      spec: {
        image: "tensorflow:latest",
      },
    };

    it("should set terminationGracePeriodSeconds to 30 when cluster version supports termination grace period", () => {
      const result = workloadHybridUtil.setDistributedUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBe(30);
    });

    it("should set terminationGracePeriodSeconds to null when cluster version does not support termination grace period", () => {
      const result = workloadHybridUtil.setDistributedUIDefaults(
        mockUIModel,
        false, // isSSO
        null, // policy
        "2.19", // clusterVersion - does not support termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBeNull();
    });

    it("should override existing terminationGracePeriodSeconds value when cluster version does not support it", () => {
      const mockUIModelWithValue = {
        ...mockUIModel,
        spec: {
          ...mockUIModel.spec,
          terminationGracePeriodSeconds: 90,
        },
      };

      const result = workloadHybridUtil.setDistributedUIDefaults(
        mockUIModelWithValue,
        false, // isSSO
        null, // policy
        "2.19", // clusterVersion - does not support termination grace period
      );

      expect(result.spec.terminationGracePeriodSeconds).toBe(null);
    });

    it("should omit preemptibility when cluster version does not support it", () => {
      const result = workloadHybridUtil.setDistributedUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "preemptible",
          },
        },
        false,
        null,
        "2.23",
      );

      expect(result.spec.preemptibility).toBeUndefined();
    });

    it("should preserve preemptibility when cluster version supports it", () => {
      const result = workloadHybridUtil.setDistributedUIDefaults(
        {
          ...mockUIModel,
          spec: {
            ...mockUIModel.spec,
            preemptibility: "preemptible",
          },
        },
        false,
        null,
        "2.24",
      );

      expect(result.spec.preemptibility).toBe("preemptible");
    });
  });

  describe("setDistributedMasterSpecUIDefaults", () => {
    const mockMasterSpec: IUISuperWorkloadSpec = {
      image: "tensorflow:latest",
    };

    it("should set terminationGracePeriodSeconds to 30 when cluster version supports termination grace period", () => {
      const result = workloadHybridUtil.setDistributedMasterSpecUIDefaults(
        mockMasterSpec,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports termination grace period
      );

      expect(result.terminationGracePeriodSeconds).toBe(30);
    });

    it("should set terminationGracePeriodSeconds to null when cluster version does not support termination grace period", () => {
      const result = workloadHybridUtil.setDistributedMasterSpecUIDefaults(
        mockMasterSpec,
        false, // isSSO
        null, // policy
        "2.19", // clusterVersion - does not support termination grace period
      );

      expect(result.terminationGracePeriodSeconds).toBeNull();
    });

    it("should preserve existing terminationGracePeriodSeconds value when cluster version supports termination grace period", () => {
      const mockMasterSpecWithValue = {
        ...mockMasterSpec,
        terminationGracePeriodSeconds: 120,
      };

      const result = workloadHybridUtil.setDistributedMasterSpecUIDefaults(
        mockMasterSpecWithValue,
        false, // isSSO
        null, // policy
        "2.20", // clusterVersion - supports termination grace period
      );

      expect(result.terminationGracePeriodSeconds).toBe(120);
    });

    it("should omit preemptibility when cluster version does not support it", () => {
      const result = workloadHybridUtil.setDistributedMasterSpecUIDefaults(
        {
          ...mockMasterSpec,
          preemptibility: "preemptible",
        },
        false,
        null,
        "2.23",
      );

      expect(result.preemptibility).toBeUndefined();
    });

    it("should preserve preemptibility when cluster version supports it", () => {
      const result = workloadHybridUtil.setDistributedMasterSpecUIDefaults(
        {
          ...mockMasterSpec,
          preemptibility: "preemptible",
        },
        false,
        null,
        "2.24",
      );

      expect(result.preemptibility).toBe("preemptible");
    });
  });

  describe("combineWorkloadModelWithStorageDetails", () => {
    const mockWorkloadModel: IUISuperWorkloadCreationRequest = {
      name: "test-workload",
      projectId: "project-123",
      clusterId: "cluster-123",
      spec: {
        image: "nginx:latest",
        storage: {
          pvc: [{ id: "existing-pvc", name: "existing-pvc", path: "/data", type: "pvc" }],
        },
      },
    };

    it("should combine existing storage with new storage fields", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
        nfs: [{ id: "new-nfs", name: "new-nfs", path: "/mnt/nfs", type: "nfs" }],
      };

      const result = workloadHybridUtil.combineWorkloadModelWithStorageDetails(mockWorkloadModel, newStorageFields);

      expect(result.spec.storage?.pvc).toHaveLength(2);
      expect(result.spec.storage?.pvc?.[0]).toMatchObject({ name: "existing-pvc", path: "/data" });
      expect(result.spec.storage?.pvc?.[1]).toMatchObject({ name: "new-pvc", path: "/app" });
      expect(result.spec.storage?.nfs).toHaveLength(1);
      expect(result.spec.storage?.nfs?.[0]).toMatchObject({ name: "new-nfs", path: "/mnt/nfs" });
    });

    it("should preserve the original workload model properties", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineWorkloadModelWithStorageDetails(mockWorkloadModel, newStorageFields);

      expect(result.name).toBe("test-workload");
      expect(result.projectId).toBe("project-123");
      expect(result.clusterId).toBe("cluster-123");
      expect(result.spec.image).toBe("nginx:latest");
    });

    it("should handle undefined storage fields", () => {
      const result = workloadHybridUtil.combineWorkloadModelWithStorageDetails(mockWorkloadModel, undefined);

      expect(result.spec.storage?.pvc).toHaveLength(1);
      expect(result.spec.storage?.pvc?.[0]).toMatchObject({ name: "existing-pvc", path: "/data" });
    });

    it("should not duplicate storage items with the same name", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [
          { id: "dup-pvc", name: "existing-pvc", path: "/new-path", type: "pvc" },
          { id: "new-pvc", name: "unique-pvc", path: "/app", type: "pvc" },
        ],
      };

      const result = workloadHybridUtil.combineWorkloadModelWithStorageDetails(mockWorkloadModel, newStorageFields);

      expect(result.spec.storage?.pvc).toHaveLength(2);
      expect(result.spec.storage?.pvc?.[0]).toMatchObject({ name: "existing-pvc", path: "/data" });
      expect(result.spec.storage?.pvc?.[1]).toMatchObject({ name: "unique-pvc", path: "/app" });
    });

    it("should handle workload model with no existing storage", () => {
      const modelWithoutStorage: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
        },
      };

      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineWorkloadModelWithStorageDetails(modelWithoutStorage, newStorageFields);

      expect(result.spec.storage?.pvc).toHaveLength(1);
      expect(result.spec.storage?.pvc?.[0]).toMatchObject({ name: "new-pvc", path: "/app" });
    });

    it("should preserve volumes from initial storage", () => {
      const modelWithVolumes: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
          storage: {
            volumes: [{ id: "vol-1", name: "my-volume", path: "/vol", type: "pvc" }],
          },
        },
      };

      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineWorkloadModelWithStorageDetails(modelWithVolumes, newStorageFields);

      expect(result.spec.storage?.volumes).toHaveLength(1);
      expect(result.spec.storage?.volumes?.[0]).toMatchObject({ name: "my-volume", path: "/vol" });
    });
  });

  describe("combineDistributedInferenceLeaderWorkloadModelWithComputeDetails", () => {
    it("should preserve existing leader environment fields when updating compute", () => {
      const mockWorkloadModel: IUISuperWorkloadCreationRequest = {
        name: "test-distributed-inference",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
          leader: {
            image: "leader-image:latest",
            command: "python train.py",
            environmentVariables: [{ name: "ENV_VAR", value: "test-value" }],
            compute: {
              cpuCoreRequest: 1,
              cpuMemoryRequest: "1GB",
            },
          },
          worker: {
            image: "worker-image:latest",
            command: "worker command",
          },
        },
      };

      const newComputeFields = {
        cpuCoreRequest: 4,
        memoryRequest: "8GB",
        gpuDevicesRequest: 2,
      };

      const result = workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithComputeDetails(
        mockWorkloadModel,
        newComputeFields,
      );

      // Verify compute was updated
      expect(result.spec.leader?.compute?.cpuCoreRequest).toBe(4);
      expect(result.spec.leader?.compute?.gpuDevicesRequest).toBe(2);

      // Verify leader environment fields are preserved (not replaced by worker fields)
      expect(result.spec.leader?.image).toBe("leader-image:latest");
      expect(result.spec.leader?.command).toBe("python train.py");
      expect(result.spec.leader?.environmentVariables).toEqual([{ name: "ENV_VAR", value: "test-value" }]);

      // Verify worker fields are not affected
      expect(result.spec.worker?.image).toBe("worker-image:latest");
    });

    it("should not use worker fields when updating leader compute", () => {
      const mockWorkloadModel: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
          leader: {
            image: "leader-specific-image:latest",
            environmentVariables: [{ name: "LEADER_VAR", value: "leader-value" }],
          },
          worker: {
            image: "worker-specific-image:latest",
            environmentVariables: [{ name: "WORKER_VAR", value: "worker-value" }],
          },
        },
      };

      const newComputeFields = {
        cpuCoreRequest: 2,
      };

      const result = workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithComputeDetails(
        mockWorkloadModel,
        newComputeFields,
      );

      // Verify leader fields are preserved (not replaced with worker fields)
      expect(result.spec.leader?.image).toBe("leader-specific-image:latest");
      expect(result.spec.leader?.environmentVariables).toEqual([{ name: "LEADER_VAR", value: "leader-value" }]);

      // Verify the image is NOT the worker image (regression test for the bug)
      expect(result.spec.leader?.image).not.toBe("worker-specific-image:latest");
    });
  });

  describe("combineDistributedInferenceLeaderWorkloadModelWithStorageDetails", () => {
    const mockWorkloadModel: IUISuperWorkloadCreationRequest = {
      name: "test-distributed-inference",
      projectId: "project-123",
      clusterId: "cluster-123",
      spec: {
        image: "nginx:latest",
        leader: {
          image: "leader-image:latest",
          storage: {
            pvc: [{ id: "existing-leader-pvc", name: "leader-pvc", path: "/leader-data", type: "pvc" }],
          },
        },
      },
    };

    it("should combine existing leader storage with new storage fields", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
        nfs: [{ id: "new-nfs", name: "new-nfs", path: "/mnt/nfs", type: "nfs" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithStorageDetails(
        mockWorkloadModel,
        newStorageFields,
      );

      expect(result.spec.leader?.storage?.pvc).toHaveLength(2);
      expect(result.spec.leader?.storage?.pvc?.[0]).toMatchObject({ name: "leader-pvc", path: "/leader-data" });
      expect(result.spec.leader?.storage?.pvc?.[1]).toMatchObject({ name: "new-pvc", path: "/app" });
      expect(result.spec.leader?.storage?.nfs).toHaveLength(1);
    });

    it("should preserve the original workload model properties", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithStorageDetails(
        mockWorkloadModel,
        newStorageFields,
      );

      expect(result.name).toBe("test-distributed-inference");
      expect(result.spec.image).toBe("nginx:latest");
      expect(result.spec.leader?.image).toBe("leader-image:latest");
    });

    it("should handle workload model with no existing leader storage", () => {
      const modelWithoutLeaderStorage: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
          leader: {
            image: "leader-image:latest",
          },
        },
      };

      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithStorageDetails(
        modelWithoutLeaderStorage,
        newStorageFields,
      );

      expect(result.spec.leader?.storage?.pvc).toHaveLength(1);
      expect(result.spec.leader?.storage?.pvc?.[0]).toMatchObject({ name: "new-pvc", path: "/app" });
    });

    it("should handle workload model with no leader spec", () => {
      const modelWithoutLeader: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
        },
      };

      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithStorageDetails(
        modelWithoutLeader,
        newStorageFields,
      );

      expect(result.spec.leader?.storage?.pvc).toHaveLength(1);
      expect(result.spec.leader?.storage?.pvc?.[0]).toMatchObject({ name: "new-pvc", path: "/app" });
    });

    it("should not duplicate storage items with the same name in leader", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [
          { id: "dup-pvc", name: "leader-pvc", path: "/new-path", type: "pvc" },
          { id: "new-pvc", name: "unique-pvc", path: "/app", type: "pvc" },
        ],
      };

      const result = workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithStorageDetails(
        mockWorkloadModel,
        newStorageFields,
      );

      expect(result.spec.leader?.storage?.pvc).toHaveLength(2);
      expect(result.spec.leader?.storage?.pvc?.[0]).toMatchObject({ name: "leader-pvc", path: "/leader-data" });
      expect(result.spec.leader?.storage?.pvc?.[1]).toMatchObject({ name: "unique-pvc", path: "/app" });
    });
  });

  describe("combineDistributedInferenceWorkerWorkloadModelWithStorageDetails", () => {
    const mockWorkloadModel: IUISuperWorkloadCreationRequest = {
      name: "test-distributed-inference",
      projectId: "project-123",
      clusterId: "cluster-123",
      spec: {
        image: "nginx:latest",
        worker: {
          image: "worker-image:latest",
          storage: {
            pvc: [{ id: "existing-worker-pvc", name: "worker-pvc", path: "/worker-data", type: "pvc" }],
          },
        },
      },
    };

    it("should combine existing worker storage with new storage fields", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
        nfs: [{ id: "new-nfs", name: "new-nfs", path: "/mnt/nfs", type: "nfs" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithStorageDetails(
        mockWorkloadModel,
        newStorageFields,
      );

      expect(result.spec.worker?.storage?.pvc).toHaveLength(2);
      expect(result.spec.worker?.storage?.pvc?.[0]).toMatchObject({ name: "worker-pvc", path: "/worker-data" });
      expect(result.spec.worker?.storage?.pvc?.[1]).toMatchObject({ name: "new-pvc", path: "/app" });
      expect(result.spec.worker?.storage?.nfs).toHaveLength(1);
    });

    it("should preserve the original workload model properties", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithStorageDetails(
        mockWorkloadModel,
        newStorageFields,
      );

      expect(result.name).toBe("test-distributed-inference");
      expect(result.spec.image).toBe("nginx:latest");
      expect(result.spec.worker?.image).toBe("worker-image:latest");
    });

    it("should handle workload model with no existing worker storage", () => {
      const modelWithoutWorkerStorage: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
          worker: {
            image: "worker-image:latest",
          },
        },
      };

      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithStorageDetails(
        modelWithoutWorkerStorage,
        newStorageFields,
      );

      expect(result.spec.worker?.storage?.pvc).toHaveLength(1);
      expect(result.spec.worker?.storage?.pvc?.[0]).toMatchObject({ name: "new-pvc", path: "/app" });
    });

    it("should handle workload model with no worker spec", () => {
      const modelWithoutWorker: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
        },
      };

      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "new-pvc", name: "new-pvc", path: "/app", type: "pvc" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithStorageDetails(
        modelWithoutWorker,
        newStorageFields,
      );

      expect(result.spec.worker?.storage?.pvc).toHaveLength(1);
      expect(result.spec.worker?.storage?.pvc?.[0]).toMatchObject({ name: "new-pvc", path: "/app" });
    });

    it("should not duplicate storage items with the same name in worker", () => {
      const newStorageFields: IUIStorageFields = {
        pvc: [
          { id: "dup-pvc", name: "worker-pvc", path: "/new-path", type: "pvc" },
          { id: "new-pvc", name: "unique-pvc", path: "/app", type: "pvc" },
        ],
      };

      const result = workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithStorageDetails(
        mockWorkloadModel,
        newStorageFields,
      );

      expect(result.spec.worker?.storage?.pvc).toHaveLength(2);
      expect(result.spec.worker?.storage?.pvc?.[0]).toMatchObject({ name: "worker-pvc", path: "/worker-data" });
      expect(result.spec.worker?.storage?.pvc?.[1]).toMatchObject({ name: "unique-pvc", path: "/app" });
    });

    it("should combine multiple storage types correctly", () => {
      const modelWithMultipleStorage: IUISuperWorkloadCreationRequest = {
        name: "test-workload",
        projectId: "project-123",
        clusterId: "cluster-123",
        spec: {
          image: "nginx:latest",
          worker: {
            image: "worker-image:latest",
            storage: {
              pvc: [{ id: "pvc-1", name: "pvc-1", path: "/data", type: "pvc" }],
              nfs: [{ id: "nfs-1", name: "nfs-1", path: "/nfs", type: "nfs" }],
            },
          },
        },
      };

      const newStorageFields: IUIStorageFields = {
        pvc: [{ id: "pvc-2", name: "pvc-2", path: "/app", type: "pvc" }],
        git: [{ id: "git-1", name: "git-1", path: "/repo", type: "git" }],
        s3: [{ id: "s3-1", name: "s3-1", path: "/bucket", type: "s3" }],
      };

      const result = workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithStorageDetails(
        modelWithMultipleStorage,
        newStorageFields,
      );

      expect(result.spec.worker?.storage?.pvc).toHaveLength(2);
      expect(result.spec.worker?.storage?.nfs).toHaveLength(1);
      expect(result.spec.worker?.storage?.git).toHaveLength(1);
      expect(result.spec.worker?.storage?.s3).toHaveLength(1);
    });
  });
});

describe("GPU memory aliases in workload copy and edit", () => {
  const forms = [
    {
      name: "workspace",
      model: workloadHybridUtil.getHybridWorkspaceModel,
      prepare: workloadHybridUtil.prepareWorkloadFlatRequest,
    },
    {
      name: "training",
      model: workloadHybridUtil.getHybridTrainingModel,
      prepare: workloadHybridUtil.prepareTrainingFlatRequest,
    },
    {
      name: "distributed",
      model: workloadHybridUtil.getHybridDistributedModel,
      prepare: workloadHybridUtil.prepareDistributedFlatRequest,
    },
  ];
  for (const form of forms) {
    it(`${form.name} edits a dual-alias response without stale duplicates`, () => {
      const compute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1, gpuRequestType: GpuRequestType.Portion };
      Object.assign(compute, { gpuPortionRequest: 0.5, gpuPortionLimit: 0.5 });
      Object.assign(compute, { gpuMemoryPortionRequest: 0.5, gpuMemoryPortionLimit: 0.5 });
      const unified: UnifiedWorkload = { spec: { worker: { image: "ubuntu", compute } } };
      const original = JSON.stringify(unified);
      const model = form.model(unified);
      expect(model.spec.compute).toMatchObject({ gpuPortionRequest: 0.5, gpuPortionLimit: 0.5 });
      model.spec.compute = {
        ...model.spec.compute,
        gpuPortionRequest: 0.25,
        gpuPortionLimit: 0.75,
        gpuMemoryPortionRequest: 0.5,
        gpuMemoryPortionLimit: 0.5,
      };
      const request = form.prepare(model);
      expect(request.spec?.compute).toMatchObject({ gpuPortionRequest: 0.25, gpuPortionLimit: 0.75 });
      expect(request.spec?.compute).not.toHaveProperty("gpuMemoryPortionRequest");
      expect(request.spec?.compute).not.toHaveProperty("gpuMemoryPortionLimit");
      expect(JSON.stringify(unified)).toBe(original);
    });
  }
});
