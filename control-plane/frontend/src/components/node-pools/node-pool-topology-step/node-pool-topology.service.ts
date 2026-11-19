/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { clusterServiceApi } from "@/services/infra/client-apis/cluster-service-api/cluster-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { Nodepool, NodepoolUpdateFields } from "@/swagger-models/cluster-service-client";

export async function saveTopologyPool(id: string, payload: NodepoolUpdateFields): Promise<Nodepool> {
  try {
    return (await clusterServiceApi.nodePoolsApi.patchNodepool(id, payload)).data;
  } catch (error) {
    throw httpService.handleHttpError(error);
  }
}
