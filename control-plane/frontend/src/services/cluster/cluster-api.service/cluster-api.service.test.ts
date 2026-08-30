import { describe, beforeEach, it, expect, vi } from "vitest";
import { clusterApiService } from "./cluster-api.service";
import { clusterApiServiceOld } from "./cluster-api-old.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import type { IDownloadLogClusterApiParams } from "@/models/logs.model";
import { EDistributedFrameworkUrlSlug } from "@/models/workload.model";

// Mock dependencies
vi.mock("./cluster-api-old.service");
vi.mock("@/services/cluster/workload.service/workload.service");

describe("cluster-api.service", () => {
  const mockResponse = {
    data: new Blob(["test log content"], { type: "text/plain" }),
    headers: {
      "content-disposition": 'attachment; filename="test-logs.txt"',
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(clusterApiServiceOld.getFullData).mockResolvedValue(mockResponse);
    vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("runai");
  });

  describe("downloadLogsClusterApi", () => {
    describe("External Workload Endpoints", () => {
      it("should construct correct endpoint for external workload", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "regular-job-name",
          type: "Training",
          externalWorkloadName: "external-workload-123",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/external/external/external-workload-123/logs/download",
          {},
          { responseType: "blob" },
        );

        // Verify workloadService.getFrameworkForUrl is NOT called for external workloads
        expect(workloadService.getFrameworkForUrl).not.toHaveBeenCalled();
      });

      it("should construct endpoint for external workload with container parameter", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "regular-job-name",
          type: "Training",
          externalWorkloadName: "external-workload-123",
          container: "main-container",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/external/external/external-workload-123/logs/download",
          { container: "main-container" },
          { responseType: "blob" },
        );
      });

      it("should construct endpoint for external workload with pod parameter", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "regular-job-name",
          type: "Training",
          externalWorkloadName: "external-workload-123",
          podName: "test-pod-1",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/external/external/external-workload-123/logs/download",
          { pod: "test-pod-1" },
          { responseType: "blob" },
        );
      });

      it("should construct endpoint for external workload with both container and pod parameters", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "regular-job-name",
          type: "Training",
          externalWorkloadName: "external-workload-123",
          container: "main-container",
          podName: "test-pod-1",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/external/external/external-workload-123/logs/download",
          { container: "main-container", pod: "test-pod-1" },
          { responseType: "blob" },
        );
      });

      it("should ignore jobName for external workloads", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "this-should-be-ignored",
          type: "Training",
          externalWorkloadName: "external-workload-123",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        const callArgs = vi.mocked(clusterApiServiceOld.getFullData).mock.calls[0];
        const endpoint = callArgs[0];

        expect(endpoint).toBe(
          "cluster-api/api/v1/test-project/workloads/external/external/external-workload-123/logs/download",
        );
        expect(endpoint).not.toContain("this-should-be-ignored");
      });

      it("should ignore type and distributedFramework for external workloads", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "regular-job-name",
          type: "Distributed",
          distributedFramework: "PyTorchJob",
          externalWorkloadName: "external-workload-123",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        const callArgs = vi.mocked(clusterApiServiceOld.getFullData).mock.calls[0];
        const endpoint = callArgs[0];

        expect(endpoint).toBe(
          "cluster-api/api/v1/test-project/workloads/external/external/external-workload-123/logs/download",
        );
        expect(endpoint).not.toContain("distributed");
        expect(endpoint).not.toContain(EDistributedFrameworkUrlSlug.PyTorch);
        expect(workloadService.getFrameworkForUrl).not.toHaveBeenCalled();
      });
    });

    describe("Regular Workload Endpoints", () => {
      it("should construct correct endpoint for training workload", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("runai");

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(workloadService.getFrameworkForUrl).toHaveBeenCalledWith("Training", undefined);
        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/training/runai/training-job-123/logs/download",
          {},
          { responseType: "blob" },
        );
      });

      it("should construct correct endpoint for distributed workload with framework", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue(EDistributedFrameworkUrlSlug.PyTorch);

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "distributed-job-123",
          type: "Distributed",
          distributedFramework: "PyTorchJob",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(workloadService.getFrameworkForUrl).toHaveBeenCalledWith("Distributed", "PyTorchJob");
        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          `cluster-api/api/v1/test-project/workloads/distributed/${EDistributedFrameworkUrlSlug.PyTorch}/distributed-job-123/logs/download`,
          {},
          { responseType: "blob" },
        );
      });

      it("should construct correct endpoint for inference workload", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("knative");

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "inference-job-123",
          type: "Inference",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(workloadService.getFrameworkForUrl).toHaveBeenCalledWith("Inference", undefined);
        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/inference/knative/inference-job-123/logs/download",
          {},
          { responseType: "blob" },
        );
      });

      it("should construct endpoint for workspace workload", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("runai");

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "workspace-job-123",
          type: "Workspace",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(workloadService.getFrameworkForUrl).toHaveBeenCalledWith("Workspace", undefined);
        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/workspace/runai/workspace-job-123/logs/download",
          {},
          { responseType: "blob" },
        );
      });

      it("should handle regular workload with container parameter", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("runai");

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
          container: "worker-container",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/training/runai/training-job-123/logs/download",
          { container: "worker-container" },
          { responseType: "blob" },
        );
      });

      it("should handle regular workload with pod parameter", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("runai");

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
          podName: "training-pod-1",
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/training/runai/training-job-123/logs/download",
          { pod: "training-pod-1" },
          { responseType: "blob" },
        );
      });
    });

    describe("Edge Cases", () => {
      it("should handle empty externalWorkloadName as falsy and use regular endpoint", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("runai");

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "regular-job-123",
          type: "Training",
          externalWorkloadName: "", // Empty string should be falsy
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(workloadService.getFrameworkForUrl).toHaveBeenCalledWith("Training", undefined);
        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/training/runai/regular-job-123/logs/download",
          {},
          { responseType: "blob" },
        );
      });

      it("should handle case sensitivity in workload types", async () => {
        vi.mocked(workloadService.getFrameworkForUrl).mockReturnValue("runai");

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "TRAINING", // Uppercase type
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          "cluster-api/api/v1/test-project/workloads/training/runai/training-job-123/logs/download",
          {},
          { responseType: "blob" },
        );
      });

      it("should return the response from clusterApiServiceOld.getFullData", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
        };

        const result = await clusterApiService.downloadLogsClusterApi(params);

        expect(result).toBe(mockResponse);
      });

      it("should propagate errors from clusterApiServiceOld.getFullData", async () => {
        const error = new Error("Network error");
        vi.mocked(clusterApiServiceOld.getFullData).mockRejectedValue(error);

        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
        };

        await expect(clusterApiService.downloadLogsClusterApi(params)).rejects.toThrow("Network error");
      });
    });

    describe("Query Parameters", () => {
      it("should not add query parameters when container and podName are undefined", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
          container: undefined,
          podName: undefined,
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          expect.stringMatching(
            /^cluster-api\/api\/v1\/test-project\/workloads\/training\/runai\/training-job-123\/logs\/download$/,
          ),
          {}, // Empty query params object
          { responseType: "blob" },
        );
      });

      it("should handle undefined values for container and podName", async () => {
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
          container: undefined,
          podName: undefined,
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          expect.stringMatching(
            /^cluster-api\/api\/v1\/test-project\/workloads\/training\/runai\/training-job-123\/logs\/download$/,
          ),
          {}, // Empty query params object
          { responseType: "blob" },
        );
      });

      it("should handle missing optional properties", async () => {
        // Test with only required properties
        const params: IDownloadLogClusterApiParams = {
          project: "test-project",
          jobName: "training-job-123",
          type: "Training",
          // container and podName are omitted (undefined by default)
        };

        await clusterApiService.downloadLogsClusterApi(params);

        expect(clusterApiServiceOld.getFullData).toHaveBeenCalledWith(
          expect.stringMatching(
            /^cluster-api\/api\/v1\/test-project\/workloads\/training\/runai\/training-job-123\/logs\/download$/,
          ),
          {}, // Empty query params object
          { responseType: "blob" },
        );
      });
    });
  });
});
