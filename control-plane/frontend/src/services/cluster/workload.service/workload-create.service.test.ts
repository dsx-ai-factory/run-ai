import { describe, it, expect, vi } from "vitest";
import { workloadCreateService } from "./workload-create.service";
import { workloadService } from "./workload.service";

// models
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";

// services
import { errorMessages } from "@/common/error-message.constant";

// utils
import { clusterUtil } from "@/utils/cluster.util";
import type { ICanSubmitWorkload } from "@/models/workload.model";

vi.mock("@/utils/cluster.util");
vi.mock("./workload.service");

describe("workload-create.service", () => {
  const defaultCluster: ClustersMinimalData = {
    uuid: "12234",
    name: "my-cluster",
    tenantId: 0,
  };

  describe("isSupportedKnative", () => {
    it("should return undefined if Knative is installed in the cluster", () => {
      vi.mocked(clusterUtil.isKnativeInstalledInCluster).mockReturnValue(true);
      const result = workloadCreateService.isSupportedKnative(defaultCluster);

      expect(result).toBeUndefined();
    });

    it("should return an error message if Knative is not installed in the cluster", () => {
      vi.mocked(clusterUtil.isKnativeInstalledInCluster).mockReturnValue(false);
      const result = workloadCreateService.isSupportedKnative(defaultCluster);

      expect(result).toBe(errorMessages.KNATIVE_IS_NOT_INSTALLED_SELECTED_CLUSTER);
    });
  });

  describe("isClusterConnected", () => {
    it("should return an error message if the cluster is not connected", () => {
      const res: ICanSubmitWorkload = {
        info: "some error message",
      };
      vi.mocked(workloadService.canSubmitWorkload).mockReturnValue(res);
      const result = workloadCreateService.isClusterConnected(defaultCluster);

      expect(result).toBe(errorMessages.CLUSTER_NOT_CONNECTED);
    });

    it("should return undefined if the cluster is connected", () => {
      vi.mocked(workloadService.canSubmitWorkload).mockReturnValue(null);
      const result = workloadCreateService.isClusterConnected(defaultCluster);

      expect(result).toBeUndefined();
    });
  });

  describe("getDisabledClusterReasons", () => {
    it("should return reasons when Knative is not supported", async () => {
      vi.mocked(clusterUtil.isKnativeInstalledInCluster).mockReturnValue(false);
      vi.mocked(workloadService.canSubmitWorkload).mockReturnValue(null);

      const clusters = [defaultCluster];
      const result = workloadCreateService.getDisabledClusterReasons(clusters, true);

      expect(result).toEqual({ [defaultCluster.uuid]: errorMessages.KNATIVE_IS_NOT_INSTALLED_SELECTED_CLUSTER });
    });

    it("should return an empty object when all clusters are enabled", async () => {
      vi.mocked(clusterUtil.isKnativeInstalledInCluster).mockReturnValue(false);
      vi.mocked(workloadService.canSubmitWorkload).mockReturnValue(null);

      const clusters = [defaultCluster];
      const result = workloadCreateService.getDisabledClusterReasons(clusters);

      expect(result).toEqual({});
    });
  });
});
