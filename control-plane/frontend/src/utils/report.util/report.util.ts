import type { IStatusColOptions } from "@/models/table.model";
import type { Report } from "@/swagger-models/org-unit-service-client";
import { ReportStatusMap } from "@/models/report.model";
import { tableUtil } from "@/utils/table.util";

export const reportUtil = {
  getStatusColOptions,
};

function getStatusColOptions(report: Report): IStatusColOptions {
  const statusOptions: IStatusColOptions = ReportStatusMap[report.status];

  return tableUtil.getStatusColOptions(statusOptions, report.statusMessage);
}
