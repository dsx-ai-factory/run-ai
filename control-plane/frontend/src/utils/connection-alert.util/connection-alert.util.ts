/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { alertUtil } from "@/utils/alert.util";
import type { QNotifyCreateOptions } from "quasar";

export const connectionAlertUtil = {
  getAlert,
};

function getAlert(message: string, options = {}): QNotifyCreateOptions {
  return alertUtil.getSuccess(message, {
    timeout: 0,
    actions: [{ icon: "ph-bold ph-x", color: "white", flat: true, round: true, size: "sm" }],
    ...options,
  });
}
