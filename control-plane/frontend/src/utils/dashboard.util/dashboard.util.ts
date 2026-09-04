/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TelemetryResponseValuesInner } from "@/swagger-models/workloads-client";
import type { QNotifyCreateOptions } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import { docsUtil } from "@/utils/docs.util/docs.util";
import DOCS from "@/constants/docs";

export const dashboardUtil = { calculateTelemetrySum, getCsvErrorMessage };

function calculateTelemetrySum(values: TelemetryResponseValuesInner[]): number {
  return values.reduce((totalSum, item) => parseFloat(item?.value || "0") + totalSum, 0);
}

function getCsvErrorMessage(): QNotifyCreateOptions {
  const message = DOCS.SUPPORT.URL
    ? `CSV couldn't be downloaded. Try again. If the issue persists, ${docsUtil.getSupportAnchor("contact Run:ai support")}.`
    : "CSV couldn't be downloaded. Try again.";
  return alertUtil.getSuccess(message, { html: true });
}
