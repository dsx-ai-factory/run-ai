import { clusterApiServiceOld } from "@/services/cluster/cluster-api.service/cluster-api-old.service";
import { RESEARCHER } from "@/common/api.constant";
import type { IDownloadLogParams } from "@/models/logs.model";

export const researcherService = {
  downloadLogs,
};

async function downloadLogs(downloadLogParams: IDownloadLogParams): Promise<Blob> {
  return await clusterApiServiceOld.get(`${RESEARCHER.v1}/logsfile`, downloadLogParams, { responseType: "blob" });
}
