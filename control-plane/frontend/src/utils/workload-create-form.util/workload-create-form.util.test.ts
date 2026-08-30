import { describe, it, expect } from "vitest";
import { workloadCreateFormUtil } from "@/utils/workload-create-form.util";
import type { IUIWorkloadCreation } from "@/models/workload.model";
import { ENodePoolsListOrigin } from "@/components/section/compute-resource-section";
import { EAccessOptions } from "@/components/environment/connection-access";
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";

describe("workload-create-form.util", () => {
  describe("getComputeResourceData", () => {
    it("should return default values when workload and projectNodepools are empty", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {},
      };
      const result = workloadCreateFormUtil.getComputeResourceData(workload);

      expect(result).toEqual({
        computeResourceId: null,
        nodeType: null,
        nodePools: null,
        workers: null,
        autoScaleData: undefined,
        tolerations: undefined,
        podAffinity: undefined,
      });
    });

    it("should return values from the workload object", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: { environment: "", compute: "compute-id" },
        specificEnv: {
          nodeType: "large",
          tolerations: [{ name: "example" }],
          podAffinity: { key: "soft" },
          nodePools: ["pool1", "pool2"],
        },
        distributed: { numWorkers: 5 },
      };
      const result = workloadCreateFormUtil.getComputeResourceData(workload);

      expect(result).toEqual({
        computeResourceId: "compute-id",
        nodeType: "large",
        nodePools: {
          defaultNodePools: ["pool1", "pool2"],
          allNodePoolsOptions: [],
          nodePoolsListOrigin: ENodePoolsListOrigin.Project,
        },
        workers: 5,
        tolerations: [{ name: "example" }],
        podAffinity: { key: "soft" },
      });
    });

    it("should handle projectNodepools array correctly", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          nodePools: ["pool1"],
        },
      };

      const projectNodepools = [
        {
          nodePool: {
            id: "100000",
            name: "default",
          },
          gpu: {
            deserved: 100,
            limit: -1,
            overQuotaWeight: 2,
          },
        },
        {
          nodePool: {
            id: "100001",
            name: "pool1",
          },
          gpu: {
            deserved: 100,
            limit: -1,
            overQuotaWeight: 2,
          },
        },
      ];
      const result = workloadCreateFormUtil.getComputeResourceData(workload, projectNodepools);

      expect(result.nodePools).toEqual({
        defaultNodePools: ["pool1"],
        allNodePoolsOptions: ["default", "pool1"],
        nodePoolsListOrigin: ENodePoolsListOrigin.Project,
      });
    });

    it("should return null nodePools when specificEnv.nodePools is undefined", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          nodeType: "small",
        },
      };
      const result = workloadCreateFormUtil.getComputeResourceData(workload);

      expect(result.nodePools).toBeNull();
    });

    it("should handle null projectNodepools", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: { nodePools: ["pool1"] },
      };
      const result = workloadCreateFormUtil.getComputeResourceData(workload, null);

      expect(result.nodePools).toEqual({
        defaultNodePools: ["pool1"],
        allNodePoolsOptions: [],
        nodePoolsListOrigin: ENodePoolsListOrigin.Project,
      });
    });

    it("should handle empty projectNodepools array", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: { nodePools: ["pool1"] },
      };
      const result = workloadCreateFormUtil.getComputeResourceData(workload, []);

      expect(result.nodePools).toEqual({
        defaultNodePools: ["pool1"],
        allNodePoolsOptions: [],
        nodePoolsListOrigin: ENodePoolsListOrigin.Project,
      });
    });
  });

  describe("getEnvironmentRunParamsData", () => {
    it("should return default values when specificEnv is undefined", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {},
      };
      const result = workloadCreateFormUtil.getEnvironmentRunParamsData(workload);

      expect(result).toEqual({
        command: "",
        args: "",
        environmentVariables: [],
        connections: [],
        runAsGid: null,
        runAsUid: null,
        supplementalGroups: null,
        servingPortAccess: null,
      });
    });

    it("should return values provided in specificEnv", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          command: "echo",
          args: "hello",
          environmentVariables: [{ name: "ENV_VAR", value: "value" }],
          connections: [{ name: "DB_CONNECTION" }],
          runAsGid: 1000,
          runAsUid: 1001,
          supplementalGroups: "2000",
        },
      };
      const result = workloadCreateFormUtil.getEnvironmentRunParamsData(workload);

      expect(result).toEqual({
        command: "echo",
        args: "hello",
        environmentVariables: [{ name: "ENV_VAR", value: "value" }],
        connections: [{ name: "DB_CONNECTION" }],
        runAsGid: 1000,
        runAsUid: 1001,
        supplementalGroups: "2000",
        servingPortAccess: null,
      });
    });

    it("should handle empty environmentVariables and connections arrays", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          environmentVariables: [],
          connections: [],
        },
      };
      const result = workloadCreateFormUtil.getEnvironmentRunParamsData(workload);

      expect(result.environmentVariables).toEqual([]);
      expect(result.connections).toEqual([]);
    });

    it("should set runAsGid, runAsUid, and supplementalGroups to null when they are undefined", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          runAsGid: undefined,
          runAsUid: undefined,
          supplementalGroups: undefined,
        },
      };
      const result = workloadCreateFormUtil.getEnvironmentRunParamsData(workload);

      expect(result.runAsGid).toBeNull();
      expect(result.runAsUid).toBeNull();
      expect(result.supplementalGroups).toBeNull();
    });

    it("should return default values when specificEnv is an empty object", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {},
      };
      const result = workloadCreateFormUtil.getEnvironmentRunParamsData(workload);

      expect(result).toEqual({
        command: "",
        args: "",
        environmentVariables: [],
        connections: [],
        runAsGid: null,
        runAsUid: null,
        supplementalGroups: null,
        servingPortAccess: null,
      });
    });

    it("should return servingPortAccess values when he provided", () => {
      const servingPortAccess: IServingEndpointAccessData = {
        access: { accessOption: EAccessOptions.GROUPS, authorizedUsers: null, authorizedGroups: ["rnd"] },
      };
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          servingPortAccess,
        },
      };
      const result = workloadCreateFormUtil.getEnvironmentRunParamsData(workload);

      expect(result).toEqual({
        command: "",
        args: "",
        environmentVariables: [],
        connections: [],
        runAsGid: null,
        runAsUid: null,
        supplementalGroups: null,
        servingPortAccess: servingPortAccess,
      });
    });
  });

  describe("getGeneralSettingsData", () => {
    it("should return default values when specificEnv is undefined", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {},
      };

      const result = workloadCreateFormUtil.getGeneralSettingsData(workload);

      expect(result).toEqual({
        allowOverQuota: false,
        autoDeletionTimeAfterCompletionSeconds: null,
        annotations: undefined,
        labels: undefined,
        backoffLimit: null,
      });
    });

    it("should return values as provided in specificEnv", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          allowOverQuota: true,
          autoDeletionTimeAfterCompletionSeconds: 3600,
          annotations: [{ name: "ann-name", value: "ann1" }],
          labels: [{ name: "lbl-name", value: "lbl1" }],
          backoffLimit: 3,
        },
      };
      const result = workloadCreateFormUtil.getGeneralSettingsData(workload);

      expect(result).toEqual({
        allowOverQuota: true,
        autoDeletionTimeAfterCompletionSeconds: 3600,
        annotations: [{ name: "ann-name", value: "ann1" }],
        labels: [{ name: "lbl-name", value: "lbl1" }],
        backoffLimit: 3,
      });
    });

    it("should set autoDeletionTimeAfterCompletionSeconds to null when it is not an integer", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          autoDeletionTimeAfterCompletionSeconds: 3600.5,
        },
      };
      const result = workloadCreateFormUtil.getGeneralSettingsData(workload);

      expect(result.autoDeletionTimeAfterCompletionSeconds).toBeNull();
    });

    it("should handle empty annotations and labels objects", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          annotations: [],
          labels: [],
        },
      };
      const result = workloadCreateFormUtil.getGeneralSettingsData(workload);

      expect(result.annotations).toEqual([]);
      expect(result.labels).toEqual([]);
    });

    it("should set backoffLimit to null when undefined", () => {
      const workload: IUIWorkloadCreation = {
        name: "my-workload",
        projectId: 123,
        namespace: "",
        clusterId: "cluster-id",
        assets: {
          environment: "",
          compute: "",
        },
        specificEnv: {
          backoffLimit: undefined,
        },
      };
      const result = workloadCreateFormUtil.getGeneralSettingsData(workload);

      expect(result.backoffLimit).toBeNull();
    });
  });
});
