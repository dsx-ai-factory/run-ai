/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { clusterServiceApi } from "@/services/infra/client-apis/cluster-service-api/cluster-service-api";
import { httpService } from "@/services/infra/https.service/http.service";

export interface INodeLabelQuery {
  search: string;
  offset: number;
  limit: number;
}

export const nodeLabelsService = { keys, values };

async function keys(clusterId: string, query: INodeLabelQuery): Promise<string[]> {
  try {
    const response = await clusterServiceApi.nodesApi.getNodeLabelKeys(
      clusterId,
      query.offset,
      query.limit,
      query.search,
    );
    return response.data;
  } catch (error) {
    throw httpService.handleHttpError(error);
  }
}

async function values(clusterId: string, key: string, query: INodeLabelQuery): Promise<string[]> {
  try {
    const response = await clusterServiceApi.nodesApi.getNodeLabelValues(
      clusterId,
      key,
      query.offset,
      query.limit,
      query.search,
    );
    return response.data;
  } catch (error) {
    throw httpService.handleHttpError(error);
  }
}
