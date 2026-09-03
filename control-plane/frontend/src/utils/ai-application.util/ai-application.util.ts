/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IStatusColOptions, StatusColor } from "@/models/table.model";
import { EApplicationPhase, type IApplicationTable } from "@/models/ai-application.model";
import { stringUtil } from "@/utils/string.util";

const applicationStatusToColor: Partial<Record<EApplicationPhase, StatusColor>> = {
  [EApplicationPhase.Deployed]: "success",
  [EApplicationPhase.Failed]: "negative",
};

function getStatusColOptions(app: IApplicationTable): IStatusColOptions {
  const phase = stringUtil.capitalizeString(app.phase || app.actionStatus || "Unknown");
  return {
    status: phase,
    displayAnimation: false,
    color: applicationStatusToColor[phase as EApplicationPhase],
    statusUpdatedTimeInMs: app.updatedAt ? new Date(app.updatedAt) : undefined,
    tooltipText: app.actionMessage || undefined,
  };
}

export const aiApplicationUtil = {
  getStatusColOptions,
};
