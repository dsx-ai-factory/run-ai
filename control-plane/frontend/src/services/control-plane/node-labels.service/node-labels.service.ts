/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { httpService } from "@/services/infra/https.service/http.service";

export interface INodeLabelQuery {
  contains: string;
  offset: number;
  limit: number;
}

// Contract from #4610. Use the generated Nodes API methods after that PR is
// merged and its client is available on this branch.
export const nodeLabelsService = {
  keys: (clusterId: string, query: INodeLabelQuery): Promise<string[]> =>
    httpService.request(
      "",
      `/api/v1/clusters/${encodeURIComponent(clusterId)}/nodes/labels/keys`,
      "GET",
      { responseType: "json" },
      {},
      query,
    ),
  values: (clusterId: string, key: string, query: INodeLabelQuery): Promise<string[]> =>
    httpService.request(
      "",
      `/api/v1/clusters/${encodeURIComponent(clusterId)}/nodes/labels/values`,
      "GET",
      { responseType: "json" },
      {},
      { ...query, key },
    ),
};
