/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AuditLogsApi, Configuration, type ConfigurationParameters } from "@/swagger-models/audit-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class AuditServiceApi {
  configuration: Configuration;
  auditLogsApi: AuditLogsApi;

  constructor() {
    this.configuration = new Configuration();

    this.auditLogsApi = new AuditLogsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }
  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const auditServiceApi = new AuditServiceApi();
