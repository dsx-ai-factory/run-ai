import type { Pod } from "@/models/pods.model";
import { K8S_API } from "@/common/api.constant";
import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { PodMetricType, MetricsResponse } from "@/swagger-models/workloads-client";
import type { AxiosResponse } from "axios";

export const podService = {
  getPods,
  getPodMetrics,
};

const clusterEndpoint = (clusterUuid: string): string => `${K8S_API.v1}/clusters/${clusterUuid}`;

async function getPods(jobId: string, clusterUuid: string, deploymentId?: string): Promise<Pod[]> {
  const podPath = deploymentId
    ? `${clusterEndpoint(clusterUuid)}/deployments/${deploymentId}/pods`
    : `${clusterEndpoint(clusterUuid)}/jobs/${jobId}/pods`;
  return controlPlaneService.get(podPath);
}

// new metrics apis
async function getPodMetrics(
  workloadId: string,
  podId: string,
  start: string,
  end: string,
  metricType: Array<PodMetricType>,
  numberOfSamples?: number,
) {
  try {
    const res: AxiosResponse<MetricsResponse> = await workloadServiceApi.podsApi.getWorkloadPodMetrics(
      workloadId,
      podId,
      metricType,
      start,
      end,
      numberOfSamples,
    );
    return res.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
