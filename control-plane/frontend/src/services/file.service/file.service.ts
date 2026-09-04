/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import { type IPaginationFilter } from "@/models/filter.model";
import type { ETableExportCsvFilesNames } from "@/models/table.model";
import { type ITableColumn } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import {
  API_LIMIT_MAX_RECORDS,
  type GetRecordsFunction,
  MAX_RECORDS_TO_EXPORT,
} from "@/composables/helpers/use-export-file.mapper";
import type { IFilterBy } from "@/models/filter.model";
export const fileService = {
  downloadFile,
  csvExporter,
};

export enum FileType {
  csv = 1,
}

function getFileName(requestedFilename: string | undefined, contentDispositionHeader?: string): string {
  if (requestedFilename) {
    return requestedFilename;
  }
  let fileName = "data.csv";
  if (contentDispositionHeader?.includes("attachment")) {
    const splitHeader = contentDispositionHeader.split("=");
    if (splitHeader?.length == 2) {
      fileName = splitHeader[1];
    }
  }
  return fileName;
}

async function downloadFile(
  path: string,
  fileType: FileType,
  filename?: string,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  query?: any,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  options?: any,
): Promise<void> {
  if (fileType == FileType.csv) {
    options = {
      headers: { Accept: "text/csv" },
      ...options,
    };
  }
  const response = await controlPlaneService.getFullData(path, query, { responseType: "blob", ...options });

  // create file link in browser's memory
  const href = URL.createObjectURL(response.data);

  // create "a" HTML element with href to file & click
  const link: HTMLAnchorElement = document.createElement("a");
  link.href = href;
  filename = getFileName(filename, response.headers["content-disposition"]);
  link.setAttribute("download", filename);
  document.body.appendChild(link);
  link.click();

  // clean up "a" element & remove ObjectURL
  document.body.removeChild(link);
  URL.revokeObjectURL(href);
}

function _replaceClusterIdWithClusterName<T>(records: T[], clusters: ClustersMinimalData[]): void {
  records.forEach((record) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const cluster = clusters.find((c) => c.uuid === (record as any).clusterId);
    if (cluster) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (record as any).clusterId = cluster.name;
    }
  });
}
async function csvExporter<T>(
  getRecords: GetRecordsFunction<T>,
  getCount: (filter: IPaginationFilter) => Promise<number>,
  filter: IPaginationFilter,
  columns: ITableColumn[],
  csvName: ETableExportCsvFilesNames,
  clusters: ClustersMinimalData[],
  filterByMapper?: (filterBy: IFilterBy) => IPaginationFilter,
): Promise<void> {
  const records: T[] = [];

  const mappedFilter = filterByMapper ? filterByMapper(filter) : filter;

  let totalCount = await getCount(mappedFilter);
  if (totalCount > MAX_RECORDS_TO_EXPORT) {
    totalCount = MAX_RECORDS_TO_EXPORT;
  }
  const chunkSize = API_LIMIT_MAX_RECORDS;
  const chunks = Math.ceil(totalCount / chunkSize);
  const recordsPromises = [];

  for (let i = 0; i < chunks; i++) {
    const offset = i * chunkSize;
    recordsPromises.push(
      getRecords(
        mappedFilter.filterBy,
        mappedFilter.sortBy || undefined,
        mappedFilter.sortOrder,
        offset,
        chunkSize,
        filter.search,
      ),
    );
  }

  const recordsChunks = await Promise.all(recordsPromises);
  for (const response of recordsChunks) {
    records.push(...(response as T[]));
  }

  _replaceClusterIdWithClusterName(records, clusters);

  tableUtil.exportTableAsCsv(csvName, records, columns);
}
