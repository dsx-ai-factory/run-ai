/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import type { GetAuditLogs200Response } from "@/swagger-models/audit-service-client";
import type { AuditEventSortBy } from "@/models/audit.model";
import { auditServiceApi } from "@/services/infra/client-apis/audit-service-api/audit-service-api";
import { httpService } from "@/services/infra/https.service/http.service";

export const auditService = {
  listAudits: listAudits,
  downloadAudits: downloadAudits,
};

async function listAudits(filterBy: IFilterBy = {}, start: string, end: string): Promise<GetAuditLogs200Response> {
  try {
    const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);
    const response = await auditServiceApi.auditLogsApi.getAuditLogs(
      start,
      end,
      filters?.offset,
      filters?.limit,
      filters.sortOrder,
      filters.sortBy as AuditEventSortBy,
      filters.filterBy,
    );
    return response.data;
  } catch (error) {
    throw httpService.handleHttpError(error);
  }
}

async function downloadAudits(
  filterBy: IFilterBy = {},
  start: string,
  end: string,
  offset: number,
  limit: number,
  fileType: "CSV" | "JSON",
): Promise<Blob> {
  try {
    const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy);
    const response = await auditServiceApi.auditLogsApi.downloadAuditLogs(
      start,
      end,
      fileType,
      offset,
      limit,
      filters.filterBy,
      { responseType: "blob" },
    );
    return response.data as unknown as Blob;
  } catch (error) {
    throw httpService.handleHttpError(error);
  }
}
