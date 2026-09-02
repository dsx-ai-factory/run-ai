/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import type { DepartmentsMetricData } from "@/models/metrics.model";
/**
 * @deprecated This service is deprecated and will be removed in future versions.
 */
export const departmentService = {
  getDepartmentsMetric,
};

async function getDepartmentsMetric(clusterId: string, nodePoolFilter?: string): Promise<DepartmentsMetricData> {
  const endpoint = `v1/k8s/clusters/${clusterId}/departments/metrics`;
  if (nodePoolFilter) {
    return await controlPlaneService.get(`${endpoint}?nodepoolName=${nodePoolFilter}`);
  }
  return await controlPlaneService.get(endpoint);
}
