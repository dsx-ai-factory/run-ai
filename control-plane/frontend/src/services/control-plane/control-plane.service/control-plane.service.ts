/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable */
import { httpService } from "@/services/infra/https.service/http.service";
import type { IHTTPOptions } from "@/models/http-response.model";
import { urlService } from "@/services/url.service/url.service";

function getBaseUrl(): string {
  return urlService.getBackendBaseUrl();
}

export const controlPlaneService = {
  get(endpoint: string, query?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(getBaseUrl(), endpoint, "GET", { responseType: "json", ...options }, {}, query);
  },
  getFullData(endpoint: string, query?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.requestFullData(getBaseUrl(), endpoint, "GET", { responseType: "json", ...options }, {}, query);
  },
  post(endpoint: string, data?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(getBaseUrl(), endpoint, "POST", { responseType: "json", ...options }, data);
  },
  put(endpoint: string, data?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(getBaseUrl(), endpoint, "PUT", { responseType: "json", ...options }, data);
  },
  delete(endpoint: string, data?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(getBaseUrl(), endpoint, "DELETE", { responseType: "json", ...options }, data);
  },
  patch(endpoint: string, data?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(getBaseUrl(), endpoint, "PATCH", { responseType: "json", ...options }, data);
  },
};
