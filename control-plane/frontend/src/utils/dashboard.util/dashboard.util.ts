/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TelemetryResponseValuesInner } from "@/swagger-models/workloads-client";
import type { QNotifyCreateOptions } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import { docsUtil } from "@/utils/docs.util/docs.util";
import DOCS from "@/constants/docs";
import { DATE_FORMATS } from "@/constants/date-formats";
import { dateUtil } from "@/utils/date.util";
import { stringUtil } from "@/utils/string.util/string.util";

export const dashboardUtil = { calculateTelemetrySum, getCsvErrorMessage, getCsvFileName };

function calculateTelemetrySum(values: TelemetryResponseValuesInner[]): number {
  return values.reduce((totalSum, item) => parseFloat(item?.value || "0") + totalSum, 0);
}

function getCsvErrorMessage(): QNotifyCreateOptions {
  const message = DOCS.SUPPORT.URL
    ? `CSV couldn't be downloaded. Try again. If the issue persists, ${docsUtil.getSupportAnchor("contact Run:ai support")}.`
    : "CSV couldn't be downloaded. Try again.";
  return alertUtil.getSuccess(message, { html: true });
}

function getCsvFileName(widgetTitle: string, clusterName?: string, nodePoolName?: string, date = new Date()): string {
  const parts = [widgetTitle, clusterName, nodePoolName]
    .filter((part): part is string => !!part)
    .map((part) => toFileNameSegment(part));
  parts.push(dateUtil.dateFormat(date, DATE_FORMATS.FILE_NAME_DATE_TIME));
  return `${parts.join("_")}.csv`;
}

function toFileNameSegment(text: string): string {
  return stringUtil
    .slug(text)
    .replace(/[^a-z0-9.-]+/g, "-")
    .replace(/-{2,}/g, "-")
    .replace(/^-+|-+$/g, "");
}
