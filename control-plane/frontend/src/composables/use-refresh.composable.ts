/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { EIntervalLabels } from "@/models/interval.model";
import { onMounted, onUnmounted } from "vue";
import { intervalUtil } from "@/utils/interval.util";

export function useRefresh(label: EIntervalLabels, callback: () => void): void {
  onMounted(() => {
    callback();
    intervalUtil.startInterval(label, callback);
  });

  onUnmounted(() => {
    intervalUtil.stopInterval(label);
  });
}
